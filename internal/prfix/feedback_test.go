package prfix

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/registry"
)

func TestNormalizeFeedbackFiltersExtractsAndDedupes(t *testing.T) {
	threads := []reviewThread{
		threadFixture("T1", "a.go", 10, false, false, "coderabbitai[bot]", coderabbitBody("Fix the race.")),
		threadFixture("T2", "a.go", 10, false, false, "coderabbitai", coderabbitBody("Fix   the race.")),
		threadFixture("T3", "b.go", 20, true, false, "human", "resolved"),
		threadFixture("T4", "c.go", 30, false, true, "human", "outdated"),
		threadFixture("T5", "d.go", 40, false, false, "octocat", "Please preserve human feedback."),
	}
	reviews := []pullRequestReview{{ID: "R1", State: "CHANGES_REQUESTED", Body: "Update the API docs.", URL: "u"}}
	reviews[0].Author.Login = "reviewer"
	comments := []issueComment{
		commentFixture("C1", "maintainer", "MEMBER", "<!-- kit:pr-feedback -->\nLate feedback."),
		commentFixture("C2", "visitor", "NONE", "general chatter"),
		commentFixture("C3", "trusted", "NONE", "Explicitly trusted feedback."),
	}
	contract := registry.PRFeedbackContract{Collection: registry.PRFeedbackCollection{TrustedCommentMarker: "<!-- kit:pr-feedback -->"}}
	items, active := normalizeFeedback(threads, reviews, comments, contract, CollectionOptions{TrustedCommentUsers: []string{"trusted"}})
	if active != 6 || len(items) != 5 {
		t.Fatalf("active=%d items=%d: %#v", active, len(items), items)
	}
	joined := RenderFeedback(items)
	for _, want := range []string{"Fix the race.", "Please preserve human feedback.", "Update the API docs.", "Late feedback.", "Explicitly trusted feedback."} {
		if !strings.Contains(joined, want) {
			t.Errorf("feedback missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "resolved") || strings.Contains(joined, "outdated") || strings.Contains(joined, "general chatter") {
		t.Fatalf("inactive or untrusted feedback leaked:\n%s", joined)
	}
	filtered, _ := normalizeFeedback(threads, reviews, comments, contract, CollectionOptions{CodeRabbitOnly: true})
	if len(filtered) != 1 || filtered[0].Author == "octocat" {
		t.Fatalf("CodeRabbit filter = %#v", filtered)
	}
}

func TestExtractTaskUsesCodeRabbitPromptOrCleanedText(t *testing.T) {
	if got := ExtractTask(coderabbitBody("Use the extracted task.")); got != "Use the extracted task." {
		t.Fatalf("CodeRabbit task = %q", got)
	}
	plain := "_⚠️ Potential issue_\n\nPlease fix this.\n<!-- fingerprint -->"
	if got := ExtractTask(plain); got != "Please fix this." {
		t.Fatalf("plain task = %q", got)
	}
}

func threadFixture(id, path string, line int, resolved, outdated bool, author, body string) reviewThread {
	thread := reviewThread{ID: id, Path: path, Line: line, IsResolved: resolved, IsOutdated: outdated}
	comment := reviewComment{ID: id + "C", Body: body, URL: "https://example.com/" + id}
	comment.Author.Login = author
	thread.Comments.Nodes = []reviewComment{comment}
	return thread
}

func commentFixture(id, author, association, body string) issueComment {
	comment := issueComment{ID: id, Body: body, URL: "https://example.com/" + id, AuthorAssociation: association}
	comment.Author.Login = author
	return comment
}

func coderabbitBody(task string) string {
	return "_⚠️ Potential issue_\n<details><summary>🤖 Prompt for AI Agents</summary>\n```\n" + task + "\n```\n</details>"
}
