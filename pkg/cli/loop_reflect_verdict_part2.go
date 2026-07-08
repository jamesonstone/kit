package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/feature"
)

func inlineCodePaths(line string) []string {
	var paths []string
	start := -1
	for i, r := range line {
		if r != '`' {
			continue
		}
		if start < 0 {
			start = i + 1
			continue
		}
		value := strings.TrimSpace(line[start:i])
		start = -1
		if looksLikeRepoPath(value) {
			paths = append(paths, value)
		}
	}
	return paths
}

func looksLikeRepoPath(value string) bool {
	if value == "" || strings.Contains(value, "://") {
		return false
	}
	if strings.ContainsAny(value, " \t\n\r") {
		return false
	}
	return strings.Contains(value, "/") || strings.Contains(value, ".")
}

func reflectTouchedFiles(ctx context.Context, projectRoot string, feat *feature.Feature, runner reflectEvidenceRunner) ([]string, error) {
	base, err := reflectMergeBase(ctx, projectRoot, runner)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, command := range []struct {
		id   string
		argv []string
	}{
		{id: "reflect-git-diff-base", argv: []string{"git", "diff", "--name-only", base + "...HEAD"}},
		{id: "reflect-git-diff-worktree", argv: []string{"git", "diff", "--name-only"}},
		{id: "reflect-git-diff-index", argv: []string{"git", "diff", "--name-only", "--cached"}},
		{id: "reflect-git-untracked", argv: []string{"git", "ls-files", "--others", "--exclude-standard"}},
	} {
		result := runner.Run(ctx, projectRoot, command.id, command.argv)
		if result.ExitCode != 0 {
			return nil, fmt.Errorf("%s evidence unavailable: %s", command.id, commandResultError(result))
		}
		files = append(files, commandOutputLines(result)...)
	}
	return filterReflectTouchedFiles(projectRoot, feat, files), nil
}

func reflectMergeBase(ctx context.Context, projectRoot string, runner reflectEvidenceRunner) (string, error) {
	result := runner.Run(ctx, projectRoot, "reflect-git-merge-base-origin", []string{"git", "merge-base", "HEAD", "origin/main"})
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) != "" {
		return strings.TrimSpace(result.Stdout), nil
	}
	result = runner.Run(ctx, projectRoot, "reflect-git-merge-base-main", []string{"git", "merge-base", "HEAD", "main"})
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) != "" {
		return strings.TrimSpace(result.Stdout), nil
	}
	return "", fmt.Errorf("git merge-base evidence unavailable: %s", commandResultError(result))
}

func filterReflectTouchedFiles(projectRoot string, feat *feature.Feature, files []string) []string {
	reflectPath := filepath.Join(feat.Path, reflectVerdictFileName)
	reflectRel := normalizeRepoPath(projectRoot, reflectPath)
	var filtered []string
	for _, file := range normalizeUniquePaths(projectRoot, files) {
		if file == "" || file == reflectRel || strings.HasPrefix(file, ".kit/") {
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func classifyReflectScopeDrift(declaredFiles []string, touchedFiles []string) (string, error) {
	declared := pathSet(normalizeUniquePaths("", declaredFiles))
	touched := pathSet(normalizeUniquePaths("", touchedFiles))
	if len(declared) == 0 {
		return "", errors.New("declared scope is empty")
	}
	var unlisted int
	for file := range touched {
		if !declared[file] {
			unlisted++
		}
	}
	for file := range declared {
		if !touched[file] {
			return "major", nil
		}
	}
	switch {
	case unlisted == 0:
		return "none", nil
	case unlisted <= 2:
		return "minor", nil
	default:
		return "major", nil
	}
}

func reflectReadyBoundaryCommit(ctx context.Context, projectRoot string, feat *feature.Feature, runner reflectEvidenceRunner) (reflectReadyBoundary, error) {
	specRel := normalizeRepoPath(projectRoot, filepath.Join(feat.Path, "SPEC.md"))
	result := runner.Run(ctx, projectRoot, "reflect-git-ready-log", []string{"git", "log", "--format=%H%x00%ct", "--", specRel})
	if result.ExitCode != 0 || strings.TrimSpace(result.Stdout) == "" {
		return reflectReadyBoundary{}, fmt.Errorf("git log evidence unavailable for %s: %s", specRel, commandResultError(result))
	}
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\x00")
		if len(parts) != 2 {
			return reflectReadyBoundary{}, fmt.Errorf("git log evidence unparseable for %s", specRel)
		}
		hash := strings.TrimSpace(parts[0])
		unix, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return reflectReadyBoundary{}, fmt.Errorf("git log timestamp unparseable for %s: %w", specRel, err)
		}
		show := runner.Run(ctx, projectRoot, "reflect-git-ready-show", []string{"git", "show", hash + ":" + specRel})
		if show.ExitCode != 0 {
			continue
		}
		if specPhaseIsReady(show.Stdout) {
			return reflectReadyBoundary{Hash: hash, Time: time.Unix(unix, 0).UTC()}, nil
		}
	}
	return reflectReadyBoundary{}, fmt.Errorf("git history does not contain a phase: ready boundary for %s", specRel)
}

func specPhaseIsReady(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "---" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "phase:") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "phase:")) == "ready"
		}
	}
	return false
}

func reflectReworkCount(ctx context.Context, projectRoot string, runner reflectEvidenceRunner, boundaryHash string, touchedFiles []string) (int, error) {
	if boundaryHash == "" {
		return 0, errors.New("ready boundary commit hash is required")
	}
	if len(touchedFiles) == 0 {
		return 0, nil
	}
	args := []string{"git", "log", "--format=%H", boundaryHash + "..HEAD", "--"}
	args = append(args, normalizeUniquePaths(projectRoot, touchedFiles)...)
	result := runner.Run(ctx, projectRoot, "reflect-git-rework-log", args)
	if result.ExitCode != 0 {
		return 0, fmt.Errorf("git rework evidence unavailable: %s", commandResultError(result))
	}
	return len(commandOutputLines(result)), nil
}

func writeReflectVerdictFile(path string, verdict ReflectVerdict) error {
	data, err := json.MarshalIndent(verdict, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".REFLECT-*.json")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
