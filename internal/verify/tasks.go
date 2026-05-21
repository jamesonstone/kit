// package verify parses executable verification declarations and runs them.
package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const SchemaVersion = 1

type FeatureRef struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	DirName string `json:"dir_name"`
	Path    string `json:"path"`
}

type Command struct {
	ID         string   `json:"id"`
	TaskID     string   `json:"task_id,omitempty"`
	SourcePath string   `json:"source_path"`
	Raw        string   `json:"raw"`
	Argv       []string `json:"argv"`
	CWD        string   `json:"cwd,omitempty"`
	Shell      bool     `json:"shell"`
}

type TaskBundle struct {
	TaskID        string     `json:"task_id"`
	Feature       FeatureRef `json:"feature"`
	Title         string     `json:"title,omitempty"`
	Status        string     `json:"status,omitempty"`
	Dependencies  []string   `json:"dependencies,omitempty"`
	Goal          string     `json:"goal,omitempty"`
	Scope         []string   `json:"scope,omitempty"`
	Acceptance    []string   `json:"acceptance,omitempty"`
	Verify        []Command  `json:"verify,omitempty"`
	ExpectedFiles []string   `json:"expected_files,omitempty"`
	Risk          string     `json:"risk,omitempty"`
	Rollback      string     `json:"rollback,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	HandoffNeeded bool       `json:"handoff_required"`
	SourcePath    string     `json:"source_path"`
}

type taskIndexEntry struct {
	ID           string
	Title        string
	Status       string
	Dependencies []string
}

type taskDetail struct {
	ID     string
	Fields map[string][]string
}

var (
	taskListPattern       = regexp.MustCompile(`^\s*-\s*\[([ xX])\]\s*(T\d+):\s*(.+)$`)
	progressTableTaskID   = regexp.MustCompile(`^T\d+$`)
	taskHeadingPattern    = regexp.MustCompile(`(?m)^###\s+(T\d+)\s*$`)
	sectionHeadingPattern = regexp.MustCompile(`(?m)^##\s+`)
	taskFieldPattern      = regexp.MustCompile(`^\s*-\s+\*\*([A-Z][A-Z -]*)\*\*:\s*(.*)$`)
)

func FeatureRefFromDir(featurePath string) FeatureRef {
	dirName := filepath.Base(featurePath)
	id := ""
	slug := dirName
	if index := strings.Index(dirName, "-"); index > 0 {
		id = dirName[:index]
		slug = dirName[index+1:]
	}
	return FeatureRef{
		ID:      id,
		Slug:    slug,
		DirName: dirName,
		Path:    featurePath,
	}
}

func LoadTaskBundles(tasksPath string, featureRef FeatureRef, allowShell bool) ([]TaskBundle, error) {
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	index := parseTaskIndex(content)
	details := parseTaskDetails(content)
	ids := orderedTaskIDs(index, details)
	bundles := make([]TaskBundle, 0, len(ids))
	for _, id := range ids {
		entry := index[id]
		detail := details[id]
		bundle, err := buildTaskBundle(tasksPath, featureRef, entry, detail, allowShell)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func FindTaskBundle(bundles []TaskBundle, taskID string) (TaskBundle, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(taskID))
	for _, bundle := range bundles {
		if strings.ToUpper(bundle.TaskID) == normalized {
			return bundle, true
		}
	}
	return TaskBundle{}, false
}

func SelectCommands(bundles []TaskBundle, taskID string) ([]Command, []string) {
	var commands []Command
	var taskIDs []string
	if strings.TrimSpace(taskID) != "" {
		if bundle, ok := FindTaskBundle(bundles, taskID); ok {
			return append(commands, bundle.Verify...), []string{bundle.TaskID}
		}
		return nil, []string{strings.ToUpper(strings.TrimSpace(taskID))}
	}

	for _, bundle := range bundles {
		commands = append(commands, bundle.Verify...)
		if len(bundle.Verify) > 0 {
			taskIDs = append(taskIDs, bundle.TaskID)
		}
	}
	return commands, taskIDs
}

func SelectExpectedFiles(bundles []TaskBundle, taskID string) []string {
	seen := make(map[string]struct{})
	var expected []string
	add := func(paths []string) {
		for _, path := range paths {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			expected = append(expected, path)
		}
	}

	if strings.TrimSpace(taskID) != "" {
		if bundle, ok := FindTaskBundle(bundles, taskID); ok {
			add(bundle.ExpectedFiles)
		}
		return expected
	}

	for _, bundle := range bundles {
		if len(bundle.Verify) == 0 {
			continue
		}
		add(bundle.ExpectedFiles)
	}
	return expected
}

func parseTaskIndex(content string) map[string]taskIndexEntry {
	entries := make(map[string]taskIndexEntry)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if match := taskListPattern.FindStringSubmatch(line); match != nil {
			status := "todo"
			if strings.EqualFold(match[1], "x") {
				status = "done"
			}
			id := match[2]
			entry := entries[id]
			entry.ID = id
			entry.Status = status
			entry.Title = strings.TrimSpace(stripPlanLinks(match[3]))
			entries[id] = entry
			continue
		}

		cells := markdownTableCells(line)
		if len(cells) < 5 || !progressTableTaskID.MatchString(cells[0]) {
			continue
		}
		id := cells[0]
		entry := entries[id]
		entry.ID = id
		if entry.Title == "" {
			entry.Title = cells[1]
		}
		if cells[2] != "" {
			entry.Status = strings.ToLower(cells[2])
		}
		entry.Dependencies = splitDependencies(cells[4])
		entries[id] = entry
	}
	return entries
}

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
