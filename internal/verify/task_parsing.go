// package verify parses executable verification declarations and runs them.
package verify

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

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

func stripPlanLinks(value string) string {
	parts := strings.Fields(value)
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "[PLAN-") || strings.HasPrefix(part, "[SPEC-") {
			continue
		}
		kept = append(kept, part)
	}
	return strings.Join(kept, " ")
}

func splitDependencies(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	deps := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			deps = append(deps, part)
		}
	}
	return deps
}

func handoffNeeded(bundle TaskBundle) bool {
	risk := strings.ToLower(bundle.Risk)
	return strings.Contains(risk, "medium") || strings.Contains(risk, "high") || len(bundle.Dependencies) > 1
}

func hasShellSyntax(command string) bool {
	syntax := []string{"&&", "||", ";", "|", "<", ">", "$(", "${", "\n"}
	for _, item := range syntax {
		if strings.Contains(command, item) {
			return true
		}
	}
	return false
}

func shellArgv(command string) []string {
	return []string{"sh", "-c", command}
}

func splitCommandLine(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(r)
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in command")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
}
