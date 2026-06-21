package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/verify"
)

func executeLoopReview(ctx context.Context, opts loopReviewOptions) (loopReviewReport, error) {
	if opts.Config == nil {
		opts.Config = config.Default()
	}
	opts.MinConfidence = effectiveLoopMinConfidence(opts.Config, opts.MinConfidence)
	opts.MaxIterations = effectiveLoopMaxIterations(opts.Config, opts.MaxIterations)

	startedAt := time.Now().UTC()
	report := loopReviewReport{
		SchemaVersion: loopReviewSchemaVersion,
		RunID:         verify.NewRunID(startedAt),
		Status:        "running",
		PRRef:         strings.TrimSpace(opts.PRRef),
		MinConfidence: opts.MinConfidence,
		MaxIterations: opts.MaxIterations,
		StartedAt:     startedAt,
	}
	if opts.Feature != nil {
		report.Feature = opts.Feature.DirName
	}
	loopReviewProgress(opts, "run %s started (min_confidence=%d%% max_iterations=%d)", report.RunID, opts.MinConfidence, opts.MaxIterations)
	if opts.UseSubagents {
		loopReviewProgress(opts, "subagent mode enabled; parent agent will pre-analyze lanes before using subagents")
	} else {
		loopReviewProgress(opts, "single-agent mode enabled; pass --subagents to opt into pre-analyzed subagents")
	}
	if report.Feature != "" {
		loopReviewProgress(opts, "feature context: %s", report.Feature)
	}

	loopReviewProgress(opts, "resolving review target")
	target, err := resolveLoopReviewTarget(opts)
	if err != nil {
		loopReviewProgress(opts, "stopping during target resolution: %v", err)
		return stopLoopReview(report, err)
	}
	report.BaseRef = target.BaseRef
	loopReviewProgress(opts, "target resolved: base=%s changed_files=%d", target.BaseRef, len(target.ChangedFiles))
	if target.NoLocalChanges {
		loopReviewProgress(opts, "target has no changed files; running no-change correctness review")
	}

	var prCtx *reviewLoopPRContext
	if strings.TrimSpace(opts.PRRef) != "" {
		loopReviewProgress(opts, "fetching pull request context for %s", strings.TrimSpace(opts.PRRef))
		ctx, err := fetchReviewLoopPRContext(opts.PRRef)
		if err != nil {
			loopReviewProgress(opts, "stopping during pull request lookup: %v", err)
			return stopLoopReview(report, err)
		}
		prCtx = &ctx
		loopReviewProgress(opts, "pull request context ready: %s", reviewLoopTargetRef(prCtx.Target))
	}

	if opts.DryRun {
		loopReviewProgress(opts, "building dry-run review prompt")
		prompt := buildLoopReviewPrompt(opts, target, nil, "")
		iteration := loopReviewIteration{
			Index:     1,
			StartedAt: startedAt,
			EndedAt:   time.Now().UTC(),
			DryRun:    true,
		}
		report.Iterations = append(report.Iterations, iteration)
		report.Status = "dry_run"
		report.StopReason = firstPromptLine(prompt)
		report.EndedAt = time.Now().UTC()
		return report, nil
	}

	if opts.Agent.Command == "" {
		loopReviewProgress(opts, "stopping: loop agent command is not configured")
		return stopLoopReview(report, errors.New("loop agent command is not configured; set loop.agent.command in .kit.yaml or run with --dry-run"))
	}

	loopReviewProgress(opts, "creating artifacts for run %s", report.RunID)
	artifactDir, err := createLoopArtifactDir(opts.ProjectRoot, report.RunID)
	if err != nil {
		loopReviewProgress(opts, "stopping during artifact setup: %v", err)
		return stopLoopReview(report, err)
	}
	report.ArtifactDir = loopRelArtifactDir(report.RunID)
	loopReviewProgress(opts, "artifacts: %s", report.ArtifactDir)

	seenFeedback := map[string]bool{}
	pendingFeedback := ""
	nextPRPoll := startedAt.Add(reviewLoopInitialWait)
	var lastResult *loopReviewAgentResult
	var lastPRFeedback loopReviewPRFeedback

	for i := 1; i <= opts.MaxIterations; i++ {
		iterStarted := time.Now().UTC()
		loopReviewProgress(opts, "iteration %d/%d: building prompt", i, opts.MaxIterations)
		prompt := buildLoopReviewPrompt(opts, target, prCtx, pendingFeedback)
		iteration := loopReviewIteration{Index: i, StartedAt: iterStarted, ExitCode: -1}

		promptPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "prompt.md", prompt)
		if err != nil {
			loopReviewProgress(opts, "iteration %d/%d: failed to write prompt: %v", i, opts.MaxIterations, err)
			return stopLoopReviewWithIteration(opts.ProjectRoot, report, iteration, err)
		}
		iteration.PromptPath = promptPath
		loopReviewProgress(opts, "iteration %d/%d: prompt written to %s", i, opts.MaxIterations, promptPath)
		loopReviewProgress(opts, "iteration %d/%d: running agent: %s", i, opts.MaxIterations, loopReviewAgentCommandSummary(opts.Agent))

		execResult := runLoopReviewAgent(ctx, opts, i, prompt)
		loopReviewProgress(opts, "iteration %d/%d: agent finished exit_code=%d duration=%s", i, opts.MaxIterations, execResult.ExitCode, time.Since(iterStarted).Round(time.Second))
		iteration.ExitCode = execResult.ExitCode
		stdoutPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "stdout.txt", execResult.Stdout)
		if err != nil {
			loopReviewProgress(opts, "iteration %d/%d: failed to write stdout artifact: %v", i, opts.MaxIterations, err)
			return stopLoopReviewWithIteration(opts.ProjectRoot, report, iteration, err)
		}
		stderrPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "stderr.txt", execResult.Stderr)
		if err != nil {
			loopReviewProgress(opts, "iteration %d/%d: failed to write stderr artifact: %v", i, opts.MaxIterations, err)
			return stopLoopReviewWithIteration(opts.ProjectRoot, report, iteration, err)
		}
		iteration.StdoutPath = stdoutPath
		iteration.StderrPath = stderrPath
		if execResult.Err != nil {
			iteration.Error = execResult.Err.Error()
		}
		result := parseLoopReviewAgentResult(execResult.Stdout)
		iteration.Result = &result
		iteration.EndedAt = time.Now().UTC()
		iteration.DurationMS = iteration.EndedAt.Sub(iteration.StartedAt).Milliseconds()
		report.Iterations = append(report.Iterations, iteration)
		report.EndedAt = iteration.EndedAt
		lastResult = &result
		pendingFeedback = ""
		loopReviewProgress(opts, "iteration %d/%d: parsed result done=%t correctness=%d%%", i, opts.MaxIterations, result.Done, result.Correctness)

		if execResult.Err == nil && prCtx != nil && !iteration.EndedAt.Before(nextPRPoll) {
			loopReviewProgress(opts, "iteration %d/%d: checking CodeRabbit feedback", i, opts.MaxIterations)
			feedback, err := pollLoopReviewPRFeedback(*prCtx)
			if err != nil {
				loopReviewProgress(opts, "iteration %d/%d: CodeRabbit feedback check failed: %v", i, opts.MaxIterations, err)
				return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
			}
			lastPRFeedback = feedback
			nextPRPoll = iteration.EndedAt.Add(reviewLoopPollEvery)
			if feedback.Found && !seenFeedback[feedback.Fingerprint] {
				seenFeedback[feedback.Fingerprint] = true
				pendingFeedback = renderLoopReviewPRFeedback(feedback)
				loopReviewProgress(opts, "iteration %d/%d: new CodeRabbit feedback found; continuing with another pass", i, opts.MaxIterations)
			} else {
				loopReviewProgress(opts, "iteration %d/%d: CodeRabbit status=%s feedback_found=%t pending=%t", i, opts.MaxIterations, feedback.StatusLabel, feedback.Found, feedback.Pending)
			}
		}

		if err := writeLoopReviewRunArtifact(opts.ProjectRoot, report); err != nil {
			loopReviewProgress(opts, "iteration %d/%d: failed to write run artifact: %v", i, opts.MaxIterations, err)
			return report, err
		}

		if pendingFeedback != "" {
			continue
		}
		if execResult.Err != nil {
			loopReviewProgress(opts, "iteration %d/%d: stopping because agent command failed", i, opts.MaxIterations)
			return stopLoopReviewAgentFailure(opts.ProjectRoot, report, i, execResult)
		}
		if !result.Done || result.Correctness < opts.MinConfidence {
			loopReviewProgress(opts, "iteration %d/%d: continuing; done=%t correctness=%d%% threshold=%d%%", i, opts.MaxIterations, result.Done, result.Correctness, opts.MinConfidence)
			continue
		}

		if prCtx != nil {
			loopReviewProgress(opts, "iteration %d/%d: local done; checking PR feedback before finalizing", i, opts.MaxIterations)
			feedback, err := pollLoopReviewPRFeedback(*prCtx)
			if err != nil {
				loopReviewProgress(opts, "iteration %d/%d: final PR feedback check failed: %v", i, opts.MaxIterations, err)
				return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
			}
			lastPRFeedback = feedback
			if feedback.Found && !seenFeedback[feedback.Fingerprint] {
				seenFeedback[feedback.Fingerprint] = true
				pendingFeedback = renderLoopReviewPRFeedback(feedback)
				loopReviewProgress(opts, "iteration %d/%d: new PR feedback found at final check; continuing", i, opts.MaxIterations)
				continue
			}
			if feedback.Pending && opts.WaitForCodeRabbit {
				loopReviewProgress(opts, "iteration %d/%d: waiting for CodeRabbit completion", i, opts.MaxIterations)
				if err := waitForReviewLoopCodeRabbit(*prCtx); err != nil {
					loopReviewProgress(opts, "iteration %d/%d: CodeRabbit wait failed: %v", i, opts.MaxIterations, err)
					return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
				}
				feedback, err = pollLoopReviewPRFeedback(*prCtx)
				if err != nil {
					loopReviewProgress(opts, "iteration %d/%d: post-wait CodeRabbit check failed: %v", i, opts.MaxIterations, err)
					return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
				}
				lastPRFeedback = feedback
				if feedback.Found && !seenFeedback[feedback.Fingerprint] {
					seenFeedback[feedback.Fingerprint] = true
					pendingFeedback = renderLoopReviewPRFeedback(feedback)
					loopReviewProgress(opts, "iteration %d/%d: new CodeRabbit feedback found after wait; continuing", i, opts.MaxIterations)
					continue
				}
			}
			if feedback.Pending {
				report.PRStatus = "local done, CodeRabbit pending"
				report.StopReason = fmt.Sprintf("CodeRabbit has not completed for PR #%d yet.\nRerun: kit loop review --pr %d", prCtx.Target.Number, prCtx.Target.Number)
				loopReviewProgress(opts, "iteration %d/%d: local done; CodeRabbit still pending; exiting provisionally", i, opts.MaxIterations)
			} else {
				report.PRStatus = feedback.StatusLabel
				loopReviewProgress(opts, "iteration %d/%d: PR status=%s", i, opts.MaxIterations, feedback.StatusLabel)
			}
		}

		report.Status = "complete"
		report.Correctness = result.Correctness
		report.EndedAt = time.Now().UTC()
		if err := writeLoopReviewRunArtifact(opts.ProjectRoot, report); err != nil {
			loopReviewProgress(opts, "failed to write final run artifact: %v", err)
			return report, err
		}
		loopReviewProgress(opts, "run %s complete: correctness=%d%%", report.RunID, report.Correctness)
		return report, nil
	}

	report.Status = "stopped"
	report.EndedAt = time.Now().UTC()
	if lastResult != nil {
		report.Correctness = lastResult.Correctness
	}
	if lastPRFeedback.Pending {
		report.PRStatus = "CodeRabbit pending"
	}
	report.StopReason = fmt.Sprintf("max iterations reached: %d", opts.MaxIterations)
	loopReviewProgress(opts, "run %s stopped: %s", report.RunID, report.StopReason)
	_ = writeLoopReviewRunArtifact(opts.ProjectRoot, report)
	return report, errors.New(report.StopReason)
}
