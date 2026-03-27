package cli

import (
	"fmt"
	"regexp"
	"strings"
)

type dispatchTask struct {
	ID    string
	Index int
	Body  string
}

var topLevelNumberedTaskPattern = regexp.MustCompile(`^\d+[.)]\s+`)

func normalizeDispatchTasks(raw string) ([]dispatchTask, error) {
	normalizedInput := normalizeDispatchRawInput(raw)
	if normalizedInput == "" {
		return nil, fmt.Errorf("task input cannot be empty")
	}

	taskBodies := parseDispatchTaskBodies(normalizedInput)
	if len(taskBodies) == 0 {
		return nil, fmt.Errorf("task input did not contain any dispatchable tasks")
	}

	tasks := make([]dispatchTask, 0, len(taskBodies))
	for index, body := range taskBodies {
		tasks = append(tasks, dispatchTask{
			ID:    fmt.Sprintf("D%03d", index+1),
			Index: index + 1,
			Body:  body,
		})
	}

	return tasks, nil
}

func parseDispatchTaskBodies(raw string) []string {
	lines := strings.Split(raw, "\n")
	var (
		tasks          []string
		paragraphLines []string
		listLines      []string
		sawBlankInList bool
	)

	flushParagraph := func() {
		body := compactDispatchTaskBody(paragraphLines)
		if body != "" {
			tasks = append(tasks, body)
		}
		paragraphLines = nil
	}

	flushList := func() {
		body := compactDispatchTaskBody(listLines)
		if body != "" {
			tasks = append(tasks, body)
		}
		listLines = nil
		sawBlankInList = false
	}

	for _, line := range lines {
		if itemContent, isTopLevelList := topLevelDispatchListItem(line); isTopLevelList {
			flushParagraph()
			flushList()
			listLines = []string{itemContent}
			continue
		}

		if len(listLines) > 0 {
			if strings.TrimSpace(line) == "" {
				listLines = append(listLines, "")
				sawBlankInList = true
				continue
			}

			if sawBlankInList && isTopLevelDispatchParagraph(line) {
				flushList()
				paragraphLines = append(paragraphLines, strings.TrimSpace(line))
				continue
			}

			listLines = append(listLines, strings.TrimRight(line, " \t"))
			sawBlankInList = false
			continue
		}

		if strings.TrimSpace(line) == "" {
			flushParagraph()
			continue
		}

		paragraphLines = append(paragraphLines, strings.TrimSpace(line))
	}

	flushParagraph()
	flushList()

	return tasks
}

func topLevelDispatchListItem(line string) (string, bool) {
	switch {
	case strings.HasPrefix(line, "- "):
		return strings.TrimSpace(line[2:]), true
	case strings.HasPrefix(line, "* "):
		return strings.TrimSpace(line[2:]), true
	case strings.HasPrefix(line, "+ "):
		return strings.TrimSpace(line[2:]), true
	case topLevelNumberedTaskPattern.MatchString(line):
		match := topLevelNumberedTaskPattern.FindString(line)
		return strings.TrimSpace(strings.TrimPrefix(line, match)), true
	default:
		return "", false
	}
}

func isTopLevelDispatchParagraph(line string) bool {
	if strings.TrimSpace(line) == "" {
		return false
	}

	_, isTopLevelList := topLevelDispatchListItem(line)
	return !isTopLevelList && strings.TrimLeft(line, " \t") == line
}

func compactDispatchTaskBody(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	if start >= end {
		return ""
	}

	return strings.TrimSpace(strings.Join(lines[start:end], "\n"))
}
