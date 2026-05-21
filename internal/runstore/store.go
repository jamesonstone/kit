// package runstore persists local verification run artifacts.
package runstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/verify"
)

const (
	runsDirName    = ".kit/runs"
	indexFileName  = "index.json"
	maxOutputBytes = 256 * 1024
)

type Index struct {
	SchemaVersion int          `json:"schema_version"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Runs          []IndexEntry `json:"runs"`
}

type IndexEntry struct {
	RunID       string           `json:"run_id"`
	ParentRunID string           `json:"parent_run_id,omitempty"`
	FeatureID   string           `json:"feature_id,omitempty"`
	FeatureSlug string           `json:"feature_slug"`
	FeatureDir  string           `json:"feature_dir"`
	TaskIDs     []string         `json:"task_ids,omitempty"`
	Status      verify.RunStatus `json:"status"`
	StartedAt   time.Time        `json:"started_at"`
	EndedAt     time.Time        `json:"ended_at,omitempty"`
	ArtifactDir string           `json:"artifact_dir"`
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password)\s*[:=]\s*['"]?[^'"\s]+`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]+`),
}

func RunsRoot(projectRoot string) string {
	return filepath.Join(projectRoot, filepath.FromSlash(runsDirName))
}

func Write(projectRoot string, run *verify.Run) error {
	if run == nil {
		return fmt.Errorf("run cannot be nil")
	}
	relDir := filepath.ToSlash(filepath.Join(runsDirName, run.RunID))
	absDir := filepath.Join(projectRoot, filepath.FromSlash(relDir))
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("failed to create run directory: %w", err)
	}

	for i := range run.Results {
		if err := writeCommandOutput(absDir, relDir, &run.Results[i]); err != nil {
			return err
		}
	}
	run.ArtifactDir = relDir

	runData, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal run: %w", err)
	}
	if err := os.WriteFile(filepath.Join(absDir, "run.json"), append(runData, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write run.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(absDir, "summary.md"), []byte(summaryMarkdown(*run)), 0644); err != nil {
		return fmt.Errorf("failed to write summary.md: %w", err)
	}
	return updateIndex(projectRoot, *run)
}

func Load(projectRoot, runID string) (verify.Run, error) {
	runPath := filepath.Join(RunsRoot(projectRoot), runID, "run.json")
	data, err := os.ReadFile(runPath)
	if err != nil {
		return verify.Run{}, err
	}
	var run verify.Run
	if err := json.Unmarshal(data, &run); err != nil {
		return verify.Run{}, err
	}
	return run, nil
}

func List(projectRoot string) ([]IndexEntry, error) {
	index, err := readIndex(projectRoot)
	if err != nil {
		return nil, err
	}
	return index.Runs, nil
}

func LatestForFeature(projectRoot, featureDir string) (verify.Run, bool, error) {
	entries, err := List(projectRoot)
	if err != nil {
		return verify.Run{}, false, err
	}
	bestIndex := -1
	for i, entry := range entries {
		if entry.FeatureDir != featureDir {
			continue
		}
		if bestIndex < 0 || entryCompletionTime(entry).After(entryCompletionTime(entries[bestIndex])) {
			bestIndex = i
		}
	}
	if bestIndex < 0 {
		return verify.Run{}, false, nil
	}
	run, err := Load(projectRoot, entries[bestIndex].RunID)
	if err != nil {
		return verify.Run{}, false, err
	}
	return run, true, nil
}

func readIndex(projectRoot string) (Index, error) {
	indexPath := filepath.Join(RunsRoot(projectRoot), indexFileName)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return rebuildIndex(projectRoot)
		}
		return Index{}, err
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return Index{}, err
	}
	sortIndex(index.Runs)
	return index, nil
}

func rebuildIndex(projectRoot string) (Index, error) {
	root := RunsRoot(projectRoot)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return Index{SchemaVersion: verify.SchemaVersion, UpdatedAt: time.Now().UTC()}, nil
		}
		return Index{}, err
	}
	index := Index{SchemaVersion: verify.SchemaVersion, UpdatedAt: time.Now().UTC()}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		run, err := Load(projectRoot, entry.Name())
		if err != nil {
			continue
		}
		index.Runs = append(index.Runs, indexEntryForRun(run))
	}
	sortIndex(index.Runs)
	return index, nil
}

func updateIndex(projectRoot string, run verify.Run) error {
	index, err := readIndex(projectRoot)
	if err != nil {
		return err
	}
	entry := indexEntryForRun(run)
	replaced := false
	for i := range index.Runs {
		if index.Runs[i].RunID == entry.RunID {
			index.Runs[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		index.Runs = append(index.Runs, entry)
	}
	index.SchemaVersion = verify.SchemaVersion
	index.UpdatedAt = time.Now().UTC()
	sortIndex(index.Runs)

	if err := os.MkdirAll(RunsRoot(projectRoot), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(RunsRoot(projectRoot), indexFileName), append(data, '\n'), 0644)
}

func indexEntryForRun(run verify.Run) IndexEntry {
	return IndexEntry{
		RunID:       run.RunID,
		ParentRunID: run.ParentRunID,
		FeatureID:   run.Feature.ID,
		FeatureSlug: run.Feature.Slug,
		FeatureDir:  run.Feature.DirName,
		TaskIDs:     append([]string(nil), run.TaskIDs...),
		Status:      run.Status,
		StartedAt:   run.StartedAt,
		EndedAt:     run.EndedAt,
		ArtifactDir: run.ArtifactDir,
	}
}

func sortIndex(entries []IndexEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].StartedAt.Before(entries[j].StartedAt)
	})
}

func entryCompletionTime(entry IndexEntry) time.Time {
	if !entry.EndedAt.IsZero() {
		return entry.EndedAt
	}
	return entry.StartedAt
}

func writeCommandOutput(absDir, relDir string, result *verify.CommandResult) error {
	stdout, stdoutRedacted := redactAndBound(result.Stdout)
	stderr, stderrRedacted := redactAndBound(result.Stderr)
	if stdout != "" {
		name := result.CommandID + "-stdout.txt"
		if err := os.WriteFile(filepath.Join(absDir, name), []byte(stdout), 0644); err != nil {
			return fmt.Errorf("failed to write stdout artifact for %s: %w", result.CommandID, err)
		}
		result.StdoutPath = filepath.ToSlash(filepath.Join(relDir, name))
	}
	if stderr != "" {
		name := result.CommandID + "-stderr.txt"
		if err := os.WriteFile(filepath.Join(absDir, name), []byte(stderr), 0644); err != nil {
			return fmt.Errorf("failed to write stderr artifact for %s: %w", result.CommandID, err)
		}
		result.StderrPath = filepath.ToSlash(filepath.Join(relDir, name))
	}
	result.Redacted = stdoutRedacted || stderrRedacted
	return nil
}

func redactAndBound(value string) (string, bool) {
	redacted := false
	for _, pattern := range secretPatterns {
		next := pattern.ReplaceAllStringFunc(value, func(match string) string {
			redacted = true
			if index := strings.IndexAny(match, ":="); index >= 0 {
				return match[:index+1] + " [REDACTED]"
			}
			return "[REDACTED]"
		})
		value = next
	}
	if len(value) > maxOutputBytes {
		value = value[:maxOutputBytes] + "\n[TRUNCATED]\n"
		redacted = true
	}
	return value, redacted
}

func summaryMarkdown(run verify.Run) string {
	var builder strings.Builder
	builder.WriteString("# Verification Run\n\n")
	builder.WriteString(fmt.Sprintf("- Run: `%s`\n", run.RunID))
	if run.ParentRunID != "" {
		builder.WriteString(fmt.Sprintf("- Parent run: `%s`\n", run.ParentRunID))
	}
	builder.WriteString(fmt.Sprintf("- Feature: `%s`\n", run.Feature.DirName))
	builder.WriteString(fmt.Sprintf("- Status: `%s`\n", run.Status))
	if len(run.TaskIDs) > 0 {
		builder.WriteString(fmt.Sprintf("- Tasks: `%s`\n", strings.Join(run.TaskIDs, ", ")))
	}
	builder.WriteString("\n## Commands\n\n")
	if len(run.Commands) == 0 {
		builder.WriteString("- none\n")
		return builder.String()
	}
	for _, result := range run.Results {
		builder.WriteString(fmt.Sprintf("- `%s`: %s", result.Raw, result.Status))
		if result.ExitCode != 0 {
			builder.WriteString(fmt.Sprintf(" (exit %d)", result.ExitCode))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
