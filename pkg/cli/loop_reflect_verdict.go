package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/verify"
)

const reflectVerdictFileName = "REFLECT.json"

type ReflectVerdict struct {
	TestsPass     bool   `json:"tests_pass"`
	LintDelta     int    `json:"lint_delta"`
	ScopeDrift    string `json:"scope_drift"`
	CycleTimeMin  int    `json:"cycle_time_min"`
	ReworkCount   int    `json:"rework_count"`
	PromptVersion string `json:"prompt_version"`
	Timestamp     string `json:"timestamp"`
}

type reflectVerdictOptions struct {
	ProjectRoot string
	Feature     *feature.Feature
	Runner      reflectEvidenceRunner
	Now         time.Time
}

type reflectEvidenceRunner interface {
	Run(ctx context.Context, projectRoot string, commandID string, argv []string) verify.CommandResult
}

type defaultReflectEvidenceRunner struct{}

type reflectReadyBoundary struct {
	Hash string
	Time time.Time
}

var lintIssueLinePattern = regexp.MustCompile(`(?m)^[^\s:][^:\n]*:\d+:(?:\d+:)?\s+\S.*$`)

func (defaultReflectEvidenceRunner) Run(ctx context.Context, projectRoot string, commandID string, argv []string) verify.CommandResult {
	run := verify.ExecuteRun(ctx, verify.RunOptions{
		ProjectRoot: projectRoot,
		Feature:     verify.FeatureRef{},
		Commands: []verify.Command{
			{
				ID:   commandID,
				Raw:  strings.Join(argv, " "),
				Argv: append([]string(nil), argv...),
				CWD:  projectRoot,
			},
		},
	})
	if len(run.Results) == 0 {
		return verify.CommandResult{
			CommandID: commandID,
			Argv:      append([]string(nil), argv...),
			Raw:       strings.Join(argv, " "),
			CWD:       projectRoot,
			ExitCode:  -1,
			Status:    "fail",
			Error:     "command produced no result",
		}
	}
	return run.Results[0]
}

func writeLoopReflectVerdict(ctx context.Context, opts reflectVerdictOptions) (ReflectVerdict, error) {
	verdict, err := buildLoopReflectVerdict(ctx, opts)
	if err != nil {
		return ReflectVerdict{}, err
	}
	path := filepath.Join(opts.Feature.Path, reflectVerdictFileName)
	if err := writeReflectVerdictFile(path, verdict); err != nil {
		return ReflectVerdict{}, err
	}
	if err := validateReflectVerdictFile(path); err != nil {
		return ReflectVerdict{}, err
	}
	return verdict, nil
}

func buildLoopReflectVerdict(ctx context.Context, opts reflectVerdictOptions) (ReflectVerdict, error) {
	if opts.ProjectRoot == "" {
		return ReflectVerdict{}, errors.New("project root is required for reflect verdict")
	}
	if opts.Feature == nil {
		return ReflectVerdict{}, errors.New("feature is required for reflect verdict")
	}
	if opts.Runner == nil {
		opts.Runner = defaultReflectEvidenceRunner{}
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	} else {
		opts.Now = opts.Now.UTC()
	}

	testsPass, err := reflectTestsPass(ctx, opts.ProjectRoot, opts.Runner)
	if err != nil {
		return ReflectVerdict{}, err
	}
	lintDelta, err := reflectLintDelta(ctx, opts.ProjectRoot, opts.Runner)
	if err != nil {
		return ReflectVerdict{}, err
	}
	declaredFiles, err := reflectDeclaredFiles(opts.Feature)
	if err != nil {
		return ReflectVerdict{}, err
	}
	touchedFiles, err := reflectTouchedFiles(ctx, opts.ProjectRoot, opts.Feature, opts.Runner)
	if err != nil {
		return ReflectVerdict{}, err
	}
	scopeDrift, err := classifyReflectScopeDrift(declaredFiles, touchedFiles)
	if err != nil {
		return ReflectVerdict{}, err
	}
	boundary, err := reflectReadyBoundaryCommit(ctx, opts.ProjectRoot, opts.Feature, opts.Runner)
	if err != nil {
		return ReflectVerdict{}, err
	}
	reworkCount, err := reflectReworkCount(ctx, opts.ProjectRoot, opts.Runner, boundary.Hash, touchedFiles)
	if err != nil {
		return ReflectVerdict{}, err
	}
	cycleTimeMin := int(opts.Now.Sub(boundary.Time).Minutes())
	if cycleTimeMin < 0 {
		cycleTimeMin = 0
	}

	return ReflectVerdict{
		TestsPass:    testsPass,
		LintDelta:    lintDelta,
		ScopeDrift:   scopeDrift,
		CycleTimeMin: cycleTimeMin,
		ReworkCount:  reworkCount,
		Timestamp:    opts.Now.Format(time.RFC3339),
	}, nil
}

func reflectTestsPass(ctx context.Context, projectRoot string, runner reflectEvidenceRunner) (bool, error) {
	result := runner.Run(ctx, projectRoot, "reflect-tests", []string{"make", "test"})
	if result.ExitCode < 0 {
		return false, fmt.Errorf("test evidence unavailable: %s", commandResultError(result))
	}
	return result.ExitCode == 0, nil
}

func reflectLintDelta(ctx context.Context, projectRoot string, runner reflectEvidenceRunner) (int, error) {
	result := runner.Run(ctx, projectRoot, "reflect-lint", []string{"make", "lint"})
	if result.ExitCode < 0 {
		return 0, fmt.Errorf("lint evidence unavailable: %s", commandResultError(result))
	}
	if result.ExitCode == 0 {
		return 0, nil
	}
	count := parseLintIssueCount(result.Stdout + "\n" + result.Stderr)
	if count == 0 {
		return 0, fmt.Errorf("lint evidence unparseable: command exited %d without recognizable findings", result.ExitCode)
	}
	return count, nil
}

func parseLintIssueCount(output string) int {
	seen := make(map[string]struct{})
	for _, line := range lintIssueLinePattern.FindAllString(output, -1) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		seen[line] = struct{}{}
	}
	return len(seen)
}

func reflectDeclaredFiles(feat *feature.Feature) ([]string, error) {
	specPath := filepath.Join(feat.Path, "SPEC.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("read SPEC.md for declared scope: %w", err)
	}
	files := declaredFilesFromSpec(string(data))
	if len(files) == 0 {
		return nil, errors.New("SPEC.md does not declare expected files for reflect scope scoring")
	}
	return files, nil
}

func declaredFilesFromSpec(content string) []string {
	var files []string
	for _, line := range strings.Split(content, "\n") {
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "expected file") && !strings.Contains(lower, "expected-file") {
			continue
		}
		files = append(files, inlineCodePaths(line)...)
	}
	return normalizeUniquePaths("", files)
}

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
