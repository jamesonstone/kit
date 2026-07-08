package verify

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func parseTaskDetails(content string) map[string]taskDetail {
	details := make(map[string]taskDetail)
	matches := taskHeadingPattern.FindAllStringSubmatchIndex(content, -1)
	for i, match := range matches {
		id := content[match[2]:match[3]]
		start := match[1]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		if sectionEnd := nextSectionHeading(content, start); sectionEnd >= 0 && sectionEnd < end {
			end = sectionEnd
		}
		details[id] = taskDetail{
			ID:     id,
			Fields: parseTaskFields(content[start:end]),
		}
	}
	return details
}

func nextSectionHeading(content string, start int) int {
	matches := sectionHeadingPattern.FindAllStringIndex(content[start:], -1)
	if len(matches) == 0 {
		return -1
	}
	return start + matches[0][0]
}

func parseTaskFields(content string) map[string][]string {
	fields := make(map[string][]string)
	current := ""
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if match := taskFieldPattern.FindStringSubmatch(line); match != nil {
			current = normalizeFieldName(match[1])
			if inline := strings.TrimSpace(match[2]); inline != "" {
				fields[current] = append(fields[current], inline)
			}
			continue
		}
		if current == "" {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields[current] = append(fields[current], trimmed)
	}
	return fields
}

func buildTaskBundle(
	tasksPath string,
	featureRef FeatureRef,
	entry taskIndexEntry,
	detail taskDetail,
	allowShell bool,
) (TaskBundle, error) {
	id := entry.ID
	if id == "" {
		id = detail.ID
	}
	bundle := TaskBundle{
		TaskID:        id,
		Feature:       featureRef,
		Title:         entry.Title,
		Status:        entry.Status,
		Dependencies:  entry.Dependencies,
		Goal:          firstText(detail.Fields["GOAL"]),
		Scope:         cleanList(detail.Fields["SCOPE"]),
		Acceptance:    cleanList(detail.Fields["ACCEPTANCE"]),
		ExpectedFiles: cleanList(detail.Fields["EXPECTED FILES"]),
		Risk:          firstText(detail.Fields["RISK"]),
		Rollback:      firstText(detail.Fields["ROLLBACK"]),
		Notes:         firstText(detail.Fields["NOTES"]),
		SourcePath:    tasksPath,
	}
	bundle.HandoffNeeded = handoffNeeded(bundle)

	rawCommands := cleanList(detail.Fields["VERIFY"])
	for i, raw := range rawCommands {
		command, err := ParseCommand(raw, id, i+1, tasksPath, allowShell)
		if err != nil {
			return TaskBundle{}, fmt.Errorf("%s %s VERIFY command %d: %w", filepath.Base(tasksPath), id, i+1, err)
		}
		bundle.Verify = append(bundle.Verify, command)
	}

	return bundle, nil
}

func ParseCommand(raw string, taskID string, index int, sourcePath string, allowShell bool) (Command, error) {
	cleaned := cleanInlineCode(strings.TrimSpace(raw))
	if cleaned == "" {
		return Command{}, fmt.Errorf("command cannot be empty")
	}

	if hasShellSyntax(cleaned) {
		if !allowShell {
			return Command{}, fmt.Errorf("shell syntax is disabled by default; rerun with --allow-shell if this is intentional")
		}
		return Command{
			ID:         fmt.Sprintf("%s-%03d", taskID, index),
			TaskID:     taskID,
			SourcePath: sourcePath,
			Raw:        cleaned,
			Argv:       shellArgv(cleaned),
			Shell:      true,
		}, nil
	}

	argv, err := splitCommandLine(cleaned)
	if err != nil {
		return Command{}, err
	}
	if len(argv) == 0 {
		return Command{}, fmt.Errorf("command cannot be empty")
	}

	return Command{
		ID:         fmt.Sprintf("%s-%03d", taskID, index),
		TaskID:     taskID,
		SourcePath: sourcePath,
		Raw:        cleaned,
		Argv:       argv,
	}, nil
}

func orderedTaskIDs(index map[string]taskIndexEntry, details map[string]taskDetail) []string {
	seen := make(map[string]struct{})
	ids := make([]string, 0, len(index)+len(details))
	for id := range index {
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	for id := range details {
		if _, ok := seen[id]; ok {
			continue
		}
		ids = append(ids, id)
	}
	sort.SliceStable(ids, func(i, j int) bool {
		return taskIDNumber(ids[i]) < taskIDNumber(ids[j])
	})
	return ids
}

func taskIDNumber(taskID string) int {
	value := 0
	for _, r := range strings.TrimPrefix(strings.ToUpper(taskID), "T") {
		if r < '0' || r > '9' {
			break
		}
		value = value*10 + int(r-'0')
	}
	return value
}

func markdownTableCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
		return nil
	}
	trimmed = strings.Trim(trimmed, "|")
	rawCells := strings.Split(trimmed, "|")
	cells := make([]string, 0, len(rawCells))
	for _, cell := range rawCells {
		cells = append(cells, strings.TrimSpace(cell))
	}
	return cells
}

func normalizeFieldName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

func cleanList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		item := cleanBullet(value)
		if item == "" {
			continue
		}
		cleaned = append(cleaned, item)
	}
	return cleaned
}

func firstText(values []string) string {
	return strings.Join(cleanList(values), " ")
}

func cleanBullet(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "- ")
	trimmed = strings.TrimPrefix(trimmed, "* ")
	return cleanInlineCode(strings.TrimSpace(trimmed))
}

func cleanInlineCode(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") && len(trimmed) >= 2 {
		return strings.TrimSpace(strings.Trim(trimmed, "`"))
	}
	return strings.ReplaceAll(trimmed, "`", "")
}
