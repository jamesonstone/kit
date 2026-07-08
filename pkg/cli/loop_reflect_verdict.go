package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
