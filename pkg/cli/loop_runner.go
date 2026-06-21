package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/verify"
)

func executeLoop(ctx context.Context, opts loopOptions) (loopReport, error) {
	if opts.Config == nil {
		opts.Config = config.Default()
	}
	opts.MinConfidence = effectiveLoopMinConfidence(opts.Config, opts.MinConfidence)
	opts.MaxIterations = effectiveLoopMaxIterations(opts.Config, opts.MaxIterations)
	startedAt := time.Now().UTC()
	report := loopReport{
		SchemaVersion: loopSchemaVersion,
		RunID:         verify.NewRunID(startedAt),
		Feature:       opts.Feature.DirName,
		Status:        "running",
		Until:         opts.Until,
		MinConfidence: opts.MinConfidence,
		MaxIterations: opts.MaxIterations,
		StartedAt:     startedAt,
	}

	if opts.DryRun {
		state := resolveStrictLoopStage(opts.ProjectRoot, opts.Feature)
		stage := state.Stage
		report.Status = "dry_run"
		report.StopReason = fmt.Sprintf("next stage: %s", stage)
		report.Iterations = append(report.Iterations, loopIteration{
			Index:       1,
			Stage:       stage,
			Before:      state,
			After:       state,
			StartedAt:   startedAt,
			EndedAt:     time.Now().UTC(),
			DryRun:      true,
			Description: loopDryRunDescription(opts, state),
		})
		report.EndedAt = time.Now().UTC()
		return report, nil
	}

	if opts.Agent.Command == "" {
		report.Status = "stopped"
		report.StopReason = "loop agent command is not configured"
		report.EndedAt = time.Now().UTC()
		return report, errors.New("loop agent command is not configured; set loop.agent.command in .kit.yaml or run with --dry-run")
	}

	artifactDir, err := createLoopArtifactDir(opts.ProjectRoot, report.RunID)
	if err != nil {
		report.Status = "stopped"
		report.StopReason = err.Error()
		report.EndedAt = time.Now().UTC()
		return report, err
	}
	report.ArtifactDir = loopRelArtifactDir(report.RunID)

	var lastImplementProgress feature.TaskProgress
	for i := 1; i <= opts.MaxIterations; i++ {
		before := resolveStrictLoopStage(opts.ProjectRoot, opts.Feature)
		if loopTargetComplete(before.Stage, opts.Until) {
			report.Status = "complete"
			report.StopReason = fmt.Sprintf("target stage %s complete", opts.Until)
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, nil
		}
		if before.Stage == loopStageBlocked {
			report.Status = "blocked"
			report.StopReason = "SPEC.md phase is blocked"
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, errors.New(report.StopReason)
		}
		if before.Stage == loopStageComplete {
			report.Status = "complete"
			report.StopReason = "feature complete"
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, nil
		}

		if before.Stage == loopStageImplement {
			lastImplementProgress = feature.TaskProgress{Total: before.TasksTotal, Complete: before.TasksDone}
		}

		iterStarted := time.Now().UTC()
		prompt, err := buildLoopPromptForStage(opts.ProjectRoot, opts.Config, opts.Feature, before.Stage, opts.MinConfidence)
		iteration := loopIteration{
			Index:     i,
			Stage:     before.Stage,
			Before:    before,
			StartedAt: iterStarted,
			ExitCode:  -1,
		}
		if err != nil {
			iteration.Error = err.Error()
			iteration.EndedAt = time.Now().UTC()
			iteration.DurationMS = iteration.EndedAt.Sub(iteration.StartedAt).Milliseconds()
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = err.Error()
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, err
		}
		promptPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "prompt.md", prompt)
		if err != nil {
			return stopLoopWithIterationError(opts.ProjectRoot, report, iteration, err)
		}
		iteration.PromptPath = promptPath

		execResult := runLoopAgent(ctx, opts, before.Stage, i, prompt)
		iteration.ExitCode = execResult.ExitCode
		stdoutPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "stdout.txt", execResult.Stdout)
		if err != nil {
			return stopLoopWithIterationError(opts.ProjectRoot, report, iteration, err)
		}
		stderrPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "stderr.txt", execResult.Stderr)
		if err != nil {
			return stopLoopWithIterationError(opts.ProjectRoot, report, iteration, err)
		}
		iteration.StdoutPath = stdoutPath
		iteration.StderrPath = stderrPath
		if execResult.Err != nil {
			iteration.Error = execResult.Err.Error()
		}
		result, err := parseLoopAgentResult(execResult.Stdout, execResult.Stderr)
		if err == nil {
			iteration.Result = result
		}
		iteration.EndedAt = time.Now().UTC()
		iteration.DurationMS = iteration.EndedAt.Sub(iteration.StartedAt).Milliseconds()

		if execResult.Err != nil {
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = fmt.Sprintf("agent command failed at %s: %v", before.Stage, execResult.Err)
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, fmt.Errorf("%s", report.StopReason)
		}
		if err != nil {
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = err.Error()
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, err
		}
		if err := validateLoopAgentResult(*result, before.Stage, opts.MinConfidence); err != nil {
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = err.Error()
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, err
		}
		if err := rollup.Update(opts.ProjectRoot, opts.Config); err != nil {
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = fmt.Sprintf("failed to update PROJECT_PROGRESS_SUMMARY.md: %v", err)
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, errors.New(report.StopReason)
		}
		if err := stopOnFailedVerification(opts.ProjectRoot, opts.Feature, report.StartedAt); err != nil {
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = err.Error()
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, err
		}

		after := resolveStrictLoopStage(opts.ProjectRoot, opts.Feature)
		iteration.After = after
		if validationErr := validateLoopProgress(before, after, lastImplementProgress); validationErr != nil {
			report.Iterations = append(report.Iterations, iteration)
			report.Status = "stopped"
			report.StopReason = validationErr.Error()
			report.EndedAt = time.Now().UTC()
			_ = writeLoopRunArtifact(opts.ProjectRoot, report)
			return report, validationErr
		}
		report.Iterations = append(report.Iterations, iteration)
		report.EndedAt = time.Now().UTC()
		if err := writeLoopRunArtifact(opts.ProjectRoot, report); err != nil {
			return report, err
		}
	}

	report.Status = "stopped"
	report.StopReason = fmt.Sprintf("max iterations reached: %d", opts.MaxIterations)
	report.EndedAt = time.Now().UTC()
	_ = writeLoopRunArtifact(opts.ProjectRoot, report)
	return report, errors.New(report.StopReason)
}

func stopLoopWithIterationError(projectRoot string, report loopReport, iteration loopIteration, err error) (loopReport, error) {
	iteration.Error = err.Error()
	iteration.EndedAt = time.Now().UTC()
	iteration.DurationMS = iteration.EndedAt.Sub(iteration.StartedAt).Milliseconds()
	report.Iterations = append(report.Iterations, iteration)
	report.Status = "stopped"
	report.StopReason = err.Error()
	report.EndedAt = time.Now().UTC()
	_ = writeLoopRunArtifact(projectRoot, report)
	return report, err
}
