package verify

import (
	"os"
	"path/filepath"
	"regexp"
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
