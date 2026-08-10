package prfix

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
)

var (
	promptDetailsPattern = regexp.MustCompile(`(?is)<details>\s*<summary>[^<]*Prompt for AI Agents?[^<]*</summary>(.*?)</details>`)
	codeFencePattern     = regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_-]+)?\\s*\\n(.*?)\\n```")
	detailsPattern       = regexp.MustCompile(`(?is)<details>.*?</details>`)
	suggestionPattern    = regexp.MustCompile(`(?is)<!--\s*suggestion_start\s*-->.*?<!--\s*suggestion_end\s*-->`)
	htmlCommentPattern   = regexp.MustCompile(`(?is)<!--.*?-->`)
	htmlTagPattern       = regexp.MustCompile(`(?is)</?(?:p|strong|em)>`)
	whitespacePattern    = regexp.MustCompile(`\s+`)
)

func normalizeFeedback(
	threads []reviewThread,
	reviews []pullRequestReview,
	comments []issueComment,
	contract registry.PRFeedbackContract,
	options CollectionOptions,
) ([]Feedback, int) {
	var candidates []Feedback
	for _, thread := range threads {
		if thread.IsResolved || thread.IsOutdated {
			continue
		}
		comment, found := selectThreadComment(thread, options.CodeRabbitOnly)
		if !found {
			continue
		}
		line := thread.Line
		if line == 0 {
			line = thread.StartLine
		}
		candidates = append(candidates, newFeedback(
			"review-thread", comment.ID, thread.ID, thread.Path, line,
			comment.Author.Login, comment.URL, comment.Body,
		))
	}
	for _, review := range reviews {
		if review.State != "CHANGES_REQUESTED" || strings.TrimSpace(review.Body) == "" ||
			(options.CodeRabbitOnly && !IsCodeRabbit(review.Author.Login)) {
			continue
		}
		candidates = append(candidates, newFeedback(
			"requested-change-review", review.ID, "", "", 0,
			review.Author.Login, review.URL, review.Body,
		))
	}
	trusted := trustedUsers(options.TrustedCommentUsers)
	for _, comment := range comments {
		if !trustedTopLevelComment(comment, contract.Collection.TrustedCommentMarker, trusted) ||
			(options.CodeRabbitOnly && !IsCodeRabbit(comment.Author.Login)) {
			continue
		}
		candidates = append(candidates, newFeedback(
			"trusted-pr-comment", comment.ID, "", "", 0,
			comment.Author.Login, comment.URL, comment.Body,
		))
	}
	activeCount := len(candidates)
	seen := map[string]bool{}
	items := make([]Feedback, 0, len(candidates))
	for _, item := range candidates {
		if item.Task == "" || seen[item.Fingerprint] {
			continue
		}
		seen[item.Fingerprint] = true
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Path < items[j].Path ||
			(items[i].Path == items[j].Path && items[i].Line < items[j].Line)
	})
	return items, activeCount
}

func newFeedback(kind, nodeID, threadID, path string, line int, author, url, body string) Feedback {
	task := ExtractTask(body)
	normalized := strings.ToLower(strings.TrimSpace(whitespacePattern.ReplaceAllString(task, " ")))
	key := fmt.Sprintf("%s\x00%d\x00%s", path, line, normalized)
	return Feedback{
		Kind: kind, NodeID: nodeID, ThreadID: threadID, Path: path, Line: line,
		Author: author, URL: url, Body: body, Task: task,
		Fingerprint: fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(key))),
	}
}

func selectThreadComment(thread reviewThread, coderabbitOnly bool) (reviewComment, bool) {
	for _, comment := range thread.Comments.Nodes {
		if coderabbitOnly && !IsCodeRabbit(comment.Author.Login) {
			continue
		}
		return comment, true
	}
	return reviewComment{}, false
}

func IsCodeRabbit(author string) bool {
	normalized := strings.ToLower(strings.TrimSpace(author))
	return normalized == "coderabbitai" || normalized == "coderabbitai[bot]" ||
		strings.Contains(normalized, "coderabbit")
}

func ExtractTask(body string) string {
	if match := promptDetailsPattern.FindStringSubmatch(body); match != nil {
		content := strings.TrimSpace(match[1])
		if fence := codeFencePattern.FindStringSubmatch(content); fence != nil {
			return cleanTask(fence[1])
		}
		return cleanTask(content)
	}
	cleaned := suggestionPattern.ReplaceAllString(body, "")
	cleaned = detailsPattern.ReplaceAllString(cleaned, "")
	cleaned = htmlCommentPattern.ReplaceAllString(cleaned, "")
	return cleanTask(cleaned)
}

func cleanTask(body string) string {
	body = strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n", "</p>", "\n").Replace(body)
	body = htmlTagPattern.ReplaceAllString(body, "")
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for len(lines) > 0 {
		line := strings.TrimSpace(lines[0])
		if line == "" || (strings.HasPrefix(line, "_") && strings.Contains(line, "Potential issue")) {
			lines = lines[1:]
			continue
		}
		break
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func trustedUsers(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "@")))
		if value != "" {
			result[value] = true
		}
	}
	return result
}

func trustedTopLevelComment(comment issueComment, marker string, users map[string]bool) bool {
	if users[strings.ToLower(comment.Author.Login)] {
		return strings.TrimSpace(comment.Body) != ""
	}
	association := strings.ToUpper(comment.AuthorAssociation)
	trustedAssociation := association == "OWNER" || association == "MEMBER" || association == "COLLABORATOR"
	return trustedAssociation && marker != "" && strings.Contains(comment.Body, marker)
}
