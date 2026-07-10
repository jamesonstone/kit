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
	ProjectRoot  string
	SuiteName    string
	DryRun       bool
	RunnerBinary string
	KitBinary    string
	KitVersion   string
	GitCommit    string
}

func Run(ctx context.Context, opts RunOptions) (RunManifest, error) {
	suite, tasks, err := LoadSuite(opts.ProjectRoot, opts.SuiteName)
	if err != nil {
		return RunManifest{}, err
	}
	provenance, err := benchmarkProvenance(opts, suite, tasks)
	if err != nil {
		return RunManifest{}, err
	}
	opts.KitBinary = provenance.KitBinaryPath
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
		Provenance:    provenance,
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
	manifest.Metrics = summarizeRun(manifest.Traces)
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
	for index, result := range run.Results {
		if result.Status == "pass" && result.ExitCode == 0 {
			continue
		}
		status = "failed"
		message := fmt.Sprintf("command %d exited %d", index, result.ExitCode)
		if result.TimedOut {
			message = fmt.Sprintf("command %d timed out", index)
		} else if strings.TrimSpace(result.Error) != "" {
			message += ": " + redactOutput(result.Error)
		}
		failed = append(failed, message)
	}
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
		FailureSignature:         failureSignature(task, run.Results, assertions, violations),
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

func summarizeRun(traces []Trace) RunMetrics {
	metrics := RunMetrics{TaskRuns: len(traces)}
	outputsByTask := map[string]map[string]struct{}{}
	for _, trace := range traces {
		if trace.Status == "passed" {
			metrics.PassedTaskRuns++
		} else {
			metrics.FailedTaskRuns++
		}
		for _, assertion := range trace.Assertions {
			metrics.Assertions++
			if assertion.Status == "passed" {
				metrics.PassedAssertions++
			} else {
				metrics.FailedAssertions++
			}
		}
		var traceOutputHashes []string
		for _, command := range trace.Commands {
			metrics.CommandDurationMS += command.DurationMS
			metrics.Stdout.Lines += command.Stdout.Lines
			metrics.Stdout.Words += command.Stdout.Words
			metrics.Stdout.Bytes += command.Stdout.Bytes
			metrics.Stdout.EstimatedTokens += command.Stdout.EstimatedTokens
			traceOutputHashes = append(traceOutputHashes, command.StdoutSHA256)
		}
		if outputsByTask[trace.TaskID] == nil {
			outputsByTask[trace.TaskID] = map[string]struct{}{}
		}
		outputsByTask[trace.TaskID][strings.Join(traceOutputHashes, ":")] = struct{}{}
	}
	repeatsByTask := map[string]int{}
	for _, trace := range traces {
		repeatsByTask[trace.TaskID]++
	}
	for taskID, repeats := range repeatsByTask {
		if repeats < 2 {
			continue
		}
		metrics.RepeatedTasks++
		if len(outputsByTask[taskID]) == 1 {
			metrics.StableRepeatedTasks++
		}
	}
	if metrics.TaskRuns > 0 {
		metrics.TaskSuccessRate = float64(metrics.PassedTaskRuns) / float64(metrics.TaskRuns)
	}
	if metrics.Assertions > 0 {
		metrics.OutputCompleteness = float64(metrics.PassedAssertions) / float64(metrics.Assertions)
	}
	if metrics.RepeatedTasks > 0 {
		metrics.DeterminismRate = float64(metrics.StableRepeatedTasks) / float64(metrics.RepeatedTasks)
	}
	return metrics
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
