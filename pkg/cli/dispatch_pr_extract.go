package cli

import (
	"fmt"
	"strings"
)

func buildDispatchPRInput(
	threads []dispatchGitHubReviewThread,
	coderabbitOnly bool,
) dispatchPRInput {
	tasks, sawSharedInstruction := extractDispatchReviewTasks(threads, coderabbitOnly)
	if len(tasks) == 0 {
		return dispatchPRInput{}
	}

	var commonInstruction string
	if sawSharedInstruction {
		commonInstruction = coderabbitSharedReviewInstruction
	}

	return dispatchPRInput{
		CommonReviewInstruction: commonInstruction,
		RawTasks:                renderDispatchReviewTasks(tasks),
	}
}

func extractDispatchReviewTasks(
	threads []dispatchGitHubReviewThread,
	coderabbitOnly bool,
) ([]dispatchReviewTask, bool) {
	var (
		tasks                []dispatchReviewTask
		seen                 = map[string]bool{}
		sawSharedInstruction bool
	)

	for _, thread := range threads {
		if thread.IsResolved || thread.IsOutdated {
			continue
		}

		comment, ok := selectDispatchReviewComment(thread, coderabbitOnly)
		if !ok {
			continue
		}

		body, foundPrompt := extractPromptForAIAgents(comment.Body)
		if foundPrompt {
			var stripped bool
			body, stripped = stripCoderabbitSharedInstruction(body)
			sawSharedInstruction = sawSharedInstruction || stripped
		} else {
			body = cleanDispatchReviewComment(comment.Body)
		}

		body = normalizeDispatchRawInput(body)
		if body == "" {
			continue
		}

		line := thread.Line
		if line == 0 {
			line = thread.StartLine
		}

		key := dispatchReviewTaskDedupeKey(thread.Path, line, body)
		if seen[key] {
			continue
		}
		seen[key] = true

		tasks = append(tasks, dispatchReviewTask{
			Author: comment.Author.Login,
			Body:   body,
			Line:   line,
			Path:   thread.Path,
			URL:    comment.URL,
		})
	}

	return tasks, sawSharedInstruction
}

func selectDispatchReviewComment(
	thread dispatchGitHubReviewThread,
	coderabbitOnly bool,
) (dispatchGitHubReviewComment, bool) {
	for _, comment := range thread.Comments.Nodes {
		if coderabbitOnly && !isCodeRabbitAuthor(comment.Author.Login) {
			continue
		}
		return comment, true
	}

	return dispatchGitHubReviewComment{}, false
}

func isCodeRabbitAuthor(login string) bool {
	normalized := strings.ToLower(strings.TrimSpace(login))
	return normalized == "coderabbitai" ||
		normalized == "coderabbitai[bot]" ||
		strings.Contains(normalized, "coderabbit")
}

func extractPromptForAIAgents(body string) (string, bool) {
	match := dispatchPromptDetailsPattern.FindStringSubmatch(body)
	if match == nil {
		return "", false
	}

	content := strings.TrimSpace(match[1])
	if fence := dispatchCodeFencePattern.FindStringSubmatch(content); fence != nil {
		return normalizeDispatchRawInput(fence[1]), true
	}

	return normalizeDispatchRawInput(stripHTMLTags(content)), true
}

func stripCoderabbitSharedInstruction(body string) (string, bool) {
	stripped := dispatchBoilerplatePattern.ReplaceAllString(body, "")
	return normalizeDispatchRawInput(stripped), stripped != body
}

func cleanDispatchReviewComment(body string) string {
	cleaned := dispatchSuggestionBlockPattern.ReplaceAllString(body, "")
	cleaned = dispatchDetailsPattern.ReplaceAllString(cleaned, "")
	cleaned = dispatchHTMLCommentPattern.ReplaceAllString(cleaned, "")
	cleaned = stripMarkdownSeverityLine(cleaned)
	cleaned = stripHTMLTags(cleaned)
	return normalizeDispatchRawInput(cleaned)
}

func stripMarkdownSeverityLine(body string) string {
	lines := strings.Split(body, "\n")
	start := 0
	for start < len(lines) {
		line := strings.TrimSpace(lines[start])
		if line == "" {
			start++
			continue
		}
		if strings.HasPrefix(line, "_") && strings.Contains(line, "Potential issue") {
			start++
			continue
		}
		break
	}

	return strings.Join(lines[start:], "\n")
}

func stripHTMLTags(body string) string {
	replacer := strings.NewReplacer(
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"</p>", "\n",
		"<p>", "",
		"</strong>", "",
		"<strong>", "",
		"</em>", "",
		"<em>", "",
	)
	return replacer.Replace(body)
}

func dispatchReviewTaskDedupeKey(path string, line int, body string) string {
	normalizedBody := strings.ToLower(dispatchWhitespacePattern.ReplaceAllString(body, " "))
	return fmt.Sprintf("%s:%d:%s", path, line, strings.TrimSpace(normalizedBody))
}

func renderDispatchPRInputForEditor(input dispatchPRInput) string {
	var parts []string
	if strings.TrimSpace(input.CommonReviewInstruction) != "" {
		parts = append(parts, strings.TrimSpace(input.CommonReviewInstruction))
	}
	if strings.TrimSpace(input.RawTasks) != "" {
		parts = append(parts, strings.TrimSpace(input.RawTasks))
	}

	return strings.Join(parts, "\n\n") + "\n"
}

func renderDispatchReviewTasks(tasks []dispatchReviewTask) string {
	var sb strings.Builder
	for _, task := range tasks {
		lineLabel := task.Path
		if task.Line > 0 {
			lineLabel = fmt.Sprintf("%s:%d", task.Path, task.Line)
		}

		sb.WriteString("- Source: ")
		sb.WriteString(lineLabel)
		sb.WriteString("\n")
		if strings.TrimSpace(task.Author) != "" {
			sb.WriteString("  Author: ")
			sb.WriteString(task.Author)
			sb.WriteString("\n")
		}
		if strings.TrimSpace(task.URL) != "" {
			sb.WriteString("  URL: ")
			sb.WriteString(task.URL)
			sb.WriteString("\n")
		}
		sb.WriteString("  Review task:\n")
		for _, line := range strings.Split(task.Body, "\n") {
			sb.WriteString("  ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String())
}

func splitDispatchPRInputFromEditor(
	raw string,
	defaultInstruction string,
) (string, string) {
	normalized := normalizeDispatchRawInput(raw)
	if normalized == "" || strings.TrimSpace(defaultInstruction) == "" {
		return normalized, ""
	}

	if strings.HasPrefix(normalized, defaultInstruction) {
		rest := strings.TrimSpace(strings.TrimPrefix(normalized, defaultInstruction))
		return rest, defaultInstruction
	}

	lines := strings.Split(normalized, "\n")
	for index, line := range lines {
		if _, isList := topLevelDispatchListItem(line); isList {
			prefix := strings.TrimSpace(strings.Join(lines[:index], "\n"))
			rest := strings.TrimSpace(strings.Join(lines[index:], "\n"))
			return rest, prefix
		}
	}

	return normalized, ""
}
