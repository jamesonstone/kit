// package feature handles feature numbering, slug validation, and directory management.
package feature

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TaskProgress holds task completion statistics.
type TaskProgress struct {
	Total    int
	Complete int
}

// Incomplete returns the number of incomplete tasks.
func (p TaskProgress) Incomplete() int {
	return p.Total - p.Complete
}

// HasTasks returns true if any tasks were found.
func (p TaskProgress) HasTasks() bool {
	return p.Total > 0
}

// FindActiveFeature returns the active feature based on:
// 1. Highest prefix-number
// 2. If no features, returns nil
func FindActiveFeature(specsDir string) (*Feature, error) {
	features, err := ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	if len(features) == 0 {
		return nil, nil
	}

	// features are already sorted by number ascending, return the last one
	active := features[len(features)-1]
	return &active, nil
}

// ParseTaskProgress parses TASKS.md and counts checkbox completion.
// Looks for markdown checkboxes: - [ ] (incomplete) and - [x] (complete)
func ParseTaskProgress(tasksPath string) (TaskProgress, error) {
	progress := TaskProgress{}

	file, err := os.Open(tasksPath)
	if err != nil {
		return progress, err
	}
	defer file.Close()

	// patterns for markdown checkboxes
	incompletePattern := regexp.MustCompile(`^\s*-\s*\[\s*\]`)
	completePattern := regexp.MustCompile(`^\s*-\s*\[[xX]\]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if incompletePattern.MatchString(line) {
			progress.Total++
		} else if completePattern.MatchString(line) {
			progress.Total++
			progress.Complete++
		}
	}

	return progress, scanner.Err()
}

// ExtractSpecSummary extracts the SUMMARY section from SPEC.md.
// Returns empty string if not found or only contains TODO placeholder.
func ExtractSpecSummary(specPath string) (string, error) {
	file, err := os.Open(specPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inSummary := false
	var summaryLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// check for SUMMARY header
		if strings.HasPrefix(trimmed, "## SUMMARY") {
			inSummary = true
			continue
		}

		// check for next section header (end of summary)
		if inSummary && strings.HasPrefix(trimmed, "## ") {
			break
		}

		// collect summary content, handling inline comments
		if inSummary && trimmed != "" {
			// skip full comment lines
			if strings.HasPrefix(trimmed, "<!--") && !strings.Contains(trimmed, "-->") {
				continue
			}
			if strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			// handle inline comments: extract text after -->
			if idx := strings.Index(trimmed, "-->"); idx != -1 {
				trimmed = strings.TrimSpace(trimmed[idx+3:])
				if trimmed == "" {
					continue
				}
			}
			summaryLines = append(summaryLines, trimmed)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	summary := strings.Join(summaryLines, " ")

	// return empty if it's just a TODO placeholder
	if strings.Contains(strings.ToLower(summary), "todo") {
		return "", nil
	}

	return summary, nil
}

// FileStatus represents the existence status of a feature file.
type FileStatus struct {
	Exists bool   `json:"exists"`
	Path   string `json:"path"`
}

// FeatureStatus holds complete status information for a feature.
type FeatureStatus struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Path     string                `json:"path"`
	Summary  string                `json:"summary,omitempty"`
	Phase    Phase                 `json:"phase"`
	Files    map[string]FileStatus `json:"files"`
	Progress *TaskProgress         `json:"progress,omitempty"`
}

// GetFeatureStatus returns complete status information for a feature.
func GetFeatureStatus(feat *Feature) (*FeatureStatus, error) {
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	status := &FeatureStatus{
		ID:    formatFeatureID(feat.Number),
		Name:  feat.Slug,
		Path:  feat.Path,
		Phase: feat.Phase,
		Files: map[string]FileStatus{
			"spec": {
				Exists: fileExists(specPath),
				Path:   specPath,
			},
			"plan": {
				Exists: fileExists(planPath),
				Path:   planPath,
			},
			"tasks": {
				Exists: fileExists(tasksPath),
				Path:   tasksPath,
			},
		},
	}

	// extract summary from spec
	if status.Files["spec"].Exists {
		summary, err := ExtractSpecSummary(specPath)
		if err == nil {
			status.Summary = summary
		}
		// silently ignore errors (file read issues are non-fatal)
	}

	// parse task progress if tasks exist
	if status.Files["tasks"].Exists {
		progress, err := ParseTaskProgress(tasksPath)
		if err == nil && progress.HasTasks() {
			status.Progress = &progress
		}
	}

	return status, nil
}

// formatFeatureID formats a feature number as a zero-padded ID.
func formatFeatureID(num int) string {
	return fmt.Sprintf("%04d", num)
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
