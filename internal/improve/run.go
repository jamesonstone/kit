package improve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/verify"
)

type RunOptions struct {
	ProjectRoot string
	SuiteName   string
	DryRun      bool
	KitBinary   string
	KitVersion  string
	GitCommit   string
}

func Run(ctx context.Context, opts RunOptions) (RunManifest, error) {
	suite, tasks, err := LoadSuite(opts.ProjectRoot, opts.SuiteName)
	if err != nil {
		return RunManifest{}, err
	}
	start := time.Now().UTC()
	runID := verify.NewRunID(start)
	runDir := filepath.Join(artifactRoot(opts.ProjectRoot), "runs", runID)
	manifest := RunManifest{
		SchemaVersion: SchemaVersion,
		Kind:          "improve_run",
		RunID:         runID,
		Suite:         suite.ID,
		StartedAt:     start,
		Status:        "pass",
		RunDir:        runDir,
	}
	if opts.DryRun {
		manifest.Status = "dry_run"
		manifest.EndedAt = time.Now().UTC()
		return manifest, nil
	}
	if err := os.MkdirAll(filepath.Join(runDir, "traces"), 0o755); err != nil {
		return RunManifest{}, err
	}
	for repeat := 1; repeat <= suite.Repeat; repeat++ {
		for _, task := range tasks {
			trace, err := runTask(ctx, opts, suite.ID, runDir, task, repeat)
			if err != nil {
				return RunManifest{}, err
			}
			manifest.Traces = append(manifest.Traces, trace)
			if trace.Status != "passed" {
				manifest.Status = "failed"
			}
		}
	}
	manifest.EndedAt = time.Now().UTC()
	if err := writeJSON(filepath.Join(runDir, "run.json"), manifest); err != nil {
		return RunManifest{}, err
	}
	updateLatest(artifactRoot(opts.ProjectRoot), runDir)
	return manifest, nil
}

func runTask(ctx context.Context, opts RunOptions, suiteID, runDir string, task Task, repeat int) (Trace, error) {
	start := time.Now().UTC()
	workspace := filepath.Join(runDir, "workspaces", fmt.Sprintf("%s-%d", task.ID, repeat))
	if err := copyDir(filepath.Join(opts.ProjectRoot, task.Fixture), workspace); err != nil {
		return Trace{}, err
	}
	before, err := snapshotDir(workspace)
	if err != nil {
		return Trace{}, err
	}
	commands, err := parseCommands(task, opts.KitBinary)
	if err != nil {
		return Trace{}, err
	}
	timeout := time.Duration(task.TimeoutSeconds) * time.Second
	run := verify.ExecuteRun(ctx, verify.RunOptions{
		ProjectRoot: workspace,
		Feature:     verify.FeatureRef{ID: task.ID, Slug: task.ID, DirName: task.ID, Path: workspace},
		TaskIDs:     []string{task.ID},
		Commands:    commands,
		Timeout:     timeout,
	})
	commandTraces, err := writeCommandOutput(runDir, task.ID, repeat, run.Results)
	if err != nil {
		return Trace{}, err
	}
	after, err := snapshotDir(workspace)
	if err != nil {
		return Trace{}, err
	}
	changed := changedFiles(before, after)
	violations := allowedSurfaceViolations(changed, task.AllowedSurfaces)
	assertions := evaluateAssertions(task, run.Results, changed)
	status := "passed"
	var failed []string
	for _, assertion := range assertions {
		if assertion.Status != "passed" {
			status = "failed"
			failed = append(failed, assertion.Message)
		}
	}
	if len(violations) > 0 {
		status = "failed"
		failed = append(failed, "changed files outside allowed surfaces: "+strings.Join(violations, ", "))
	}
	trace := Trace{
		SchemaVersion:            SchemaVersion,
		TaskID:                   task.ID,
		Suite:                    suiteID,
		KitVersion:               opts.KitVersion,
		GitCommit:                opts.GitCommit,
		StartedAt:                start,
		DurationMS:               time.Since(start).Milliseconds(),
		Status:                   status,
		WorkspacePath:            workspace,
		RepeatIndex:              repeat,
		Seed:                     "default",
		Commands:                 commandTraces,
		Assertions:               assertions,
		ChangedFiles:             changed,
		AllowedSurfaceViolations: violations,
		OracleResults:            []OracleResult{{Oracle: task.Oracle, Status: status, Message: strings.Join(failed, "; ")}},
		FailureSignature:         failureSignature(task, status),
	}
	if status == "passed" {
		trace.FailureSignature = ""
	}
	tracePath := filepath.Join(runDir, "traces", fmt.Sprintf("%s-%d.json", task.ID, repeat))
	if err := writeJSON(tracePath, trace); err != nil {
		return Trace{}, err
	}
	return trace, nil
}

func parseCommands(task Task, kitBinary string) ([]verify.Command, error) {
	commands := make([]verify.Command, 0, len(task.Commands))
	for i, raw := range task.Commands {
		command, err := verify.ParseCommand(raw, task.ID, i+1, task.ID, false)
		if err != nil {
			return nil, err
		}
		resolveKitPlaceholder(&command, kitBinary)
		commands = append(commands, command)
	}
	return commands, nil
}

func resolveKitPlaceholder(command *verify.Command, kitBinary string) {
	if strings.TrimSpace(kitBinary) == "" {
		kitBinary = "kit"
	}
	for i, arg := range command.Argv {
		if arg == "{{kit}}" {
			command.Argv[i] = kitBinary
		}
	}
}
