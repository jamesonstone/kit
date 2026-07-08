package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/verify"
)

func validateReflectVerdictFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var verdict ReflectVerdict
	if err := json.Unmarshal(data, &verdict); err != nil {
		return err
	}
	if verdict.LintDelta < 0 {
		return errors.New("reflect verdict lint_delta cannot be negative")
	}
	if verdict.CycleTimeMin < 0 {
		return errors.New("reflect verdict cycle_time_min cannot be negative")
	}
	if verdict.ReworkCount < 0 {
		return errors.New("reflect verdict rework_count cannot be negative")
	}
	switch verdict.ScopeDrift {
	case "none", "minor", "major":
	default:
		return fmt.Errorf("reflect verdict scope_drift %q is invalid", verdict.ScopeDrift)
	}
	if _, err := time.Parse(time.RFC3339, verdict.Timestamp); err != nil {
		return fmt.Errorf("reflect verdict timestamp is invalid: %w", err)
	}
	return nil
}

func commandOutputLines(result verify.CommandResult) []string {
	var lines []string
	for _, output := range []string{result.Stdout, result.Stderr} {
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			lines = append(lines, line)
		}
	}
	return lines
}

func commandResultError(result verify.CommandResult) string {
	if strings.TrimSpace(result.Error) != "" {
		return result.Error
	}
	output := strings.TrimSpace(result.Stderr)
	if output == "" {
		output = strings.TrimSpace(result.Stdout)
	}
	if output != "" {
		return output
	}
	return fmt.Sprintf("command exited %d", result.ExitCode)
}

func pathSet(paths []string) map[string]bool {
	set := make(map[string]bool)
	for _, path := range paths {
		if path == "" {
			continue
		}
		set[path] = true
	}
	return set
}

func normalizeUniquePaths(projectRoot string, paths []string) []string {
	seen := make(map[string]struct{})
	var normalized []string
	for _, path := range paths {
		path = normalizeRepoPath(projectRoot, path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}
	sort.Strings(normalized)
	return normalized
}

func normalizeRepoPath(projectRoot string, path string) string {
	path = strings.TrimSpace(path)
	if path == "" || strings.Contains(path, "://") {
		return ""
	}
	if projectRoot != "" && filepath.IsAbs(path) {
		if rel, err := filepath.Rel(projectRoot, path); err == nil {
			path = rel
		}
	}
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	if path == "." || strings.HasPrefix(path, "../") {
		return ""
	}
	return path
}
