package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/verify"
)

const loopReviewSchemaVersion = 1

var loopReviewCorrectnessPattern = regexp.MustCompile(`(?mi)^\s*Correctness:\s*(\d{1,3})\s*%`)

type loopReviewOptions struct {
	ProjectRoot       string
	Config            *config.Config
	Feature           *feature.Feature
	BaseRef           string
	PRRef             string
	WaitForCodeRabbit bool
	MinConfidence     int
	MaxIterations     int
	DryRun            bool
	JSON              bool
	Agent             config.LoopAgentConfig
}

type loopReviewReport struct {
	SchemaVersion int                   `json:"schema_version"`
	RunID         string                `json:"run_id,omitempty"`
	Status        string                `json:"status"`
	StopReason    string                `json:"stop_reason,omitempty"`
	Feature       string                `json:"feature,omitempty"`
	BaseRef       string                `json:"base_ref,omitempty"`
	PRRef         string                `json:"pr_ref,omitempty"`
	PRStatus      string                `json:"pr_status,omitempty"`
	Correctness   int                   `json:"correctness,omitempty"`
	MinConfidence int                   `json:"min_confidence"`
	MaxIterations int                   `json:"max_iterations"`
	ArtifactDir   string                `json:"artifact_dir,omitempty"`
	StartedAt     time.Time             `json:"started_at"`
	EndedAt       time.Time             `json:"ended_at"`
	Iterations    []loopReviewIteration `json:"iterations"`
}

type loopReviewIteration struct {
	Index      int                    `json:"index"`
	PromptPath string                 `json:"prompt_path,omitempty"`
	StdoutPath string                 `json:"stdout_path,omitempty"`
	StderrPath string                 `json:"stderr_path,omitempty"`
	Result     *loopReviewAgentResult `json:"result,omitempty"`
	ExitCode   int                    `json:"exit_code,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	EndedAt    time.Time              `json:"ended_at"`
	DurationMS int64                  `json:"duration_ms"`
	DryRun     bool                   `json:"dry_run,omitempty"`
}

type loopReviewTarget struct {
	BaseRef        string
	ChangedFiles   []string
	DiffStat       string
	WorkingStat    string
	StagedStat     string
	NoLocalChanges bool
}

type loopReviewAgentResult struct {
	Done        bool     `json:"done"`
	Correctness int      `json:"correctness"`
	Bullets     []string `json:"bullets,omitempty"`
	RawSummary  string   `json:"raw_summary,omitempty"`
}

type loopReviewPRFeedback struct {
	Status            reviewLoopCheckStatus
	StatusLabel       string
	Found             bool
	Fingerprint       string
	RenderedTasks     string
	CommonInstruction string
	Pending           bool
}

func newLoopReviewCommand() *cobra.Command {
	opts := loopReviewOptions{}
	cmd := &cobra.Command{
		Use:           "review [feature]",
		Short:         "Run a correctness review loop over changed code",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Run a coding-agent correctness loop over changes not in the remote
mainline. The loop repeats local review and repair passes until the configured
agent reports at least 95% correctness and ends its final response with done.

With --pr, CodeRabbit feedback is checked opportunistically while local review
continues. Use --watch or --wait-for-coderabbit to wait for CodeRabbit before
finalizing.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoopReviewCommand(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.BaseRef, "base", "", "base ref for changed-code review (default origin/main, then main)")
	cmd.Flags().StringVar(&opts.PRRef, "pr", "", "optionally ingest CodeRabbit feedback from a PR URL, Markdown link, owner/repo#number, or current-repo number")
	cmd.Flags().BoolVar(&opts.WaitForCodeRabbit, "watch", false, "wait for CodeRabbit completion before finalizing PR-mode review")
	cmd.Flags().BoolVar(&opts.WaitForCodeRabbit, "wait-for-coderabbit", false, "wait for CodeRabbit completion before finalizing PR-mode review")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "show the first review prompt without running the configured agent")
	cmd.Flags().IntVar(&opts.MinConfidence, "min-confidence", 0, "minimum correctness percentage required to stop (0 uses loop config, goal_percentage, then 95)")
	cmd.Flags().IntVar(&opts.MaxIterations, "max-iterations", 0, "maximum review iterations (0 uses loop config, then 10)")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output loop review report as JSON")
	return cmd
}

func runLoopReviewCommand(cmd *cobra.Command, args []string, opts loopReviewOptions) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	opts.ProjectRoot = projectRoot
	opts.Config = cfg
	opts.MinConfidence = effectiveLoopMinConfidence(cfg, opts.MinConfidence)
	opts.MaxIterations = effectiveLoopMaxIterations(cfg, opts.MaxIterations)
	opts.Agent = cfg.Loop.Agent

	if len(args) == 1 {
		feat, err := loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature %q not found for loop review: %w", args[0], err)
		}
		opts.Feature = feat
	}

	report, runErr := executeLoopReview(cmd.Context(), opts)
	outputErr := outputLoopReviewReport(cmd, report, opts.JSON)
	if outputErr != nil {
		return outputErr
	}
	if runErr != nil {
		return &silentCLIError{err: runErr}
	}
	return nil
}

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

	target, err := resolveLoopReviewTarget(opts)
	if err != nil {
		return stopLoopReview(report, err)
	}
	report.BaseRef = target.BaseRef

	var prCtx *reviewLoopPRContext
	if strings.TrimSpace(opts.PRRef) != "" {
		ctx, err := fetchReviewLoopPRContext(opts.PRRef)
		if err != nil {
			return stopLoopReview(report, err)
		}
		prCtx = &ctx
	}

	if opts.DryRun {
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
		return stopLoopReview(report, errors.New("loop agent command is not configured; set loop.agent.command in .kit.yaml or run with --dry-run"))
	}

	artifactDir, err := createLoopArtifactDir(opts.ProjectRoot, report.RunID)
	if err != nil {
		return stopLoopReview(report, err)
	}
	report.ArtifactDir = loopRelArtifactDir(report.RunID)

	seenFeedback := map[string]bool{}
	pendingFeedback := ""
	nextPRPoll := startedAt.Add(reviewLoopInitialWait)
	var lastResult *loopReviewAgentResult
	var lastPRFeedback loopReviewPRFeedback

	for i := 1; i <= opts.MaxIterations; i++ {
		iterStarted := time.Now().UTC()
		prompt := buildLoopReviewPrompt(opts, target, prCtx, pendingFeedback)
		iteration := loopReviewIteration{Index: i, StartedAt: iterStarted, ExitCode: -1}

		promptPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "prompt.md", prompt)
		if err != nil {
			return stopLoopReviewWithIteration(opts.ProjectRoot, report, iteration, err)
		}
		iteration.PromptPath = promptPath

		execResult := runLoopReviewAgent(ctx, opts, i, prompt)
		iteration.ExitCode = execResult.ExitCode
		stdoutPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "stdout.txt", execResult.Stdout)
		if err != nil {
			return stopLoopReviewWithIteration(opts.ProjectRoot, report, iteration, err)
		}
		stderrPath, err := writeLoopIterationFile(artifactDir, report.RunID, i, "stderr.txt", execResult.Stderr)
		if err != nil {
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

		if execResult.Err == nil && prCtx != nil && !iteration.EndedAt.Before(nextPRPoll) {
			feedback, err := pollLoopReviewPRFeedback(*prCtx)
			if err != nil {
				return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
			}
			lastPRFeedback = feedback
			nextPRPoll = iteration.EndedAt.Add(reviewLoopPollEvery)
			if feedback.Found && !seenFeedback[feedback.Fingerprint] {
				seenFeedback[feedback.Fingerprint] = true
				pendingFeedback = renderLoopReviewPRFeedback(feedback)
			}
		}

		if err := writeLoopReviewRunArtifact(opts.ProjectRoot, report); err != nil {
			return report, err
		}

		if pendingFeedback != "" {
			continue
		}
		if execResult.Err != nil {
			continue
		}
		if !result.Done || result.Correctness < opts.MinConfidence {
			continue
		}

		if prCtx != nil {
			feedback, err := pollLoopReviewPRFeedback(*prCtx)
			if err != nil {
				return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
			}
			lastPRFeedback = feedback
			if feedback.Found && !seenFeedback[feedback.Fingerprint] {
				seenFeedback[feedback.Fingerprint] = true
				pendingFeedback = renderLoopReviewPRFeedback(feedback)
				continue
			}
			if feedback.Pending && opts.WaitForCodeRabbit {
				if err := waitForReviewLoopCodeRabbit(*prCtx); err != nil {
					return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
				}
				feedback, err = pollLoopReviewPRFeedback(*prCtx)
				if err != nil {
					return stopLoopReviewAfterWrite(opts.ProjectRoot, report, err)
				}
				lastPRFeedback = feedback
				if feedback.Found && !seenFeedback[feedback.Fingerprint] {
					seenFeedback[feedback.Fingerprint] = true
					pendingFeedback = renderLoopReviewPRFeedback(feedback)
					continue
				}
			}
			if feedback.Pending {
				report.PRStatus = "local done, CodeRabbit pending"
				report.StopReason = fmt.Sprintf("CodeRabbit has not completed for PR #%d yet.\nRerun: kit loop review --pr %d", prCtx.Target.Number, prCtx.Target.Number)
			} else {
				report.PRStatus = feedback.StatusLabel
			}
		}

		report.Status = "complete"
		report.Correctness = result.Correctness
		report.EndedAt = time.Now().UTC()
		if err := writeLoopReviewRunArtifact(opts.ProjectRoot, report); err != nil {
			return report, err
		}
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
	_ = writeLoopReviewRunArtifact(opts.ProjectRoot, report)
	return report, errors.New(report.StopReason)
}

func resolveLoopReviewTarget(opts loopReviewOptions) (loopReviewTarget, error) {
	baseRef, err := resolveLoopReviewBase(opts.ProjectRoot, opts.BaseRef)
	if err != nil {
		return loopReviewTarget{}, err
	}

	files := map[string]bool{}
	for _, args := range [][]string{
		{"diff", "--name-only", baseRef + "...HEAD"},
		{"diff", "--name-only"},
		{"diff", "--cached", "--name-only"},
	} {
		output, err := reviewLoopRunner.Output(opts.ProjectRoot, "git", args...)
		if err != nil {
			return loopReviewTarget{}, fmt.Errorf("failed to inspect review diff with git %s: %w", strings.Join(args, " "), err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				files[line] = true
			}
		}
	}

	changedFiles := make([]string, 0, len(files))
	for path := range files {
		changedFiles = append(changedFiles, path)
	}
	sort.Strings(changedFiles)

	diffStat, _ := loopReviewOptionalGitOutput(opts.ProjectRoot, "diff", "--stat", baseRef+"...HEAD")
	workingStat, _ := loopReviewOptionalGitOutput(opts.ProjectRoot, "diff", "--stat")
	stagedStat, _ := loopReviewOptionalGitOutput(opts.ProjectRoot, "diff", "--cached", "--stat")

	return loopReviewTarget{
		BaseRef:        baseRef,
		ChangedFiles:   changedFiles,
		DiffStat:       diffStat,
		WorkingStat:    workingStat,
		StagedStat:     stagedStat,
		NoLocalChanges: len(changedFiles) == 0,
	}, nil
}

func resolveLoopReviewBase(projectRoot, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	for _, candidate := range []string{"origin/main", "main"} {
		if _, err := reviewLoopRunner.Output(projectRoot, "git", "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not resolve review base: tried origin/main then main; pass --base <ref> to choose a base")
}

func loopReviewOptionalGitOutput(projectRoot string, args ...string) (string, error) {
	output, err := reviewLoopRunner.Output(projectRoot, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func buildLoopReviewPrompt(
	opts loopReviewOptions,
	target loopReviewTarget,
	prCtx *reviewLoopPRContext,
	prFeedback string,
) string {
	var builder strings.Builder
	builder.WriteString("# Kit Loop Review\n\n")
	builder.WriteString("Run a local correctness review and repair pass over changed code.\n\n")
	builder.WriteString("## Target\n\n")
	builder.WriteString(fmt.Sprintf("- Base ref: `%s`\n", target.BaseRef))
	builder.WriteString("- Scope: changes on the current branch relative to the remote mainline, plus staged and unstaged working-tree changes.\n")
	if opts.Feature != nil {
		builder.WriteString(fmt.Sprintf("- Feature docs: `%s`\n", opts.Feature.DirName))
	}
	if prCtx != nil {
		builder.WriteString(fmt.Sprintf("- Pull request: `%s`\n", reviewLoopTargetRef(prCtx.Target)))
		if prCtx.URL != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", prCtx.URL))
		}
		builder.WriteString("\n")
	}
	if target.NoLocalChanges {
		builder.WriteString("- Changed files: none detected.\n")
	} else {
		builder.WriteString("- Changed files:\n")
		for _, path := range target.ChangedFiles {
			builder.WriteString(fmt.Sprintf("  - `%s`\n", path))
		}
	}
	builder.WriteString("\n## Diff Evidence\n\n")
	appendStatBlock(&builder, "Branch diff", target.DiffStat)
	appendStatBlock(&builder, "Unstaged diff", target.WorkingStat)
	appendStatBlock(&builder, "Staged diff", target.StagedStat)
	if strings.TrimSpace(prFeedback) != "" {
		builder.WriteString("\n## CodeRabbit Feedback To Ingest\n\n")
		builder.WriteString(strings.TrimSpace(prFeedback))
		builder.WriteString("\n")
	}
	builder.WriteString("\n## Instructions\n\n")
	builder.WriteString("- Inspect the actual diff and surrounding code before changing anything.\n")
	builder.WriteString("- Fix high, medium, and correctness-impacting issues; do not churn on low-risk style unless it affects correctness.\n")
	builder.WriteString("- Run the smallest relevant validation commands and add or update focused tests when needed.\n")
	builder.WriteString("- Do not stage, commit, push, post PR comments, resolve review threads, or mutate GitHub.\n")
	builder.WriteString("- If no blocking issues remain, report `done`; otherwise make the next minimal fix and report what changed.\n")
	builder.WriteString(fmt.Sprintf("- Do not report `done` unless correctness is at least %d%% and there are no high, medium, or correctness-impacting issues.\n", opts.MinConfidence))
	builder.WriteString("\n## Required Final Output\n\n")
	builder.WriteString("Keep the final response information dense and short:\n\n")
	builder.WriteString("```text\n")
	builder.WriteString("Correctness: 97%\n")
	builder.WriteString("Status: <short status>\n\n")
	builder.WriteString("- Issue: <short finding>; Fix: <short action>.\n")
	builder.WriteString("- Issue: <short finding>; Fix: <short action>.\n")
	if prCtx != nil {
		builder.WriteString("\nCodeRabbit has not completed for PR #<number> yet.\n")
		builder.WriteString("Rerun: kit loop review --pr <number>\n")
	}
	builder.WriteString("done\n")
	builder.WriteString("```\n")
	return builder.String()
}

func appendStatBlock(builder *strings.Builder, title, content string) {
	builder.WriteString(fmt.Sprintf("### %s\n\n", title))
	if strings.TrimSpace(content) == "" {
		builder.WriteString("none\n\n")
		return
	}
	builder.WriteString("```text\n")
	builder.WriteString(strings.TrimSpace(content))
	builder.WriteString("\n```\n\n")
}

func runLoopReviewAgent(ctx context.Context, opts loopReviewOptions, iteration int, prompt string) loopAgentExecution {
	cmd := exec.CommandContext(ctx, opts.Agent.Command, opts.Agent.Args...)
	cmd.Dir = opts.ProjectRoot
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(),
		"KIT_LOOP_MODE=review",
		fmt.Sprintf("KIT_LOOP_MIN_CONFIDENCE=%d", opts.MinConfidence),
		fmt.Sprintf("KIT_LOOP_ITERATION=%d", iteration),
	)
	if opts.Feature != nil {
		cmd.Env = append(cmd.Env, "KIT_LOOP_FEATURE="+opts.Feature.DirName)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return loopAgentExecution{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: commandExitCode(err),
		Err:      err,
	}
}

func parseLoopReviewAgentResult(stdout string) loopReviewAgentResult {
	result := loopReviewAgentResult{RawSummary: strings.TrimSpace(stdout)}
	for _, match := range loopReviewCorrectnessPattern.FindAllStringSubmatch(stdout, -1) {
		value, err := strconv.Atoi(match[1])
		if err == nil {
			result.Correctness = clampPercentage(value)
		}
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			result.Bullets = append(result.Bullets, trimmed)
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		result.Done = strings.TrimSpace(lines[i]) == "done"
		break
	}
	return result
}

func pollLoopReviewPRFeedback(ctx reviewLoopPRContext) (loopReviewPRFeedback, error) {
	tasks, commonInstruction, found, err := reviewLoopLoadReviewTasks(reviewLoopTargetRef(ctx.Target), true)
	if err != nil {
		return loopReviewPRFeedback{}, err
	}

	checks, err := fetchReviewLoopChecks(ctx)
	status := reviewLoopCheckUnavailable
	statusLabel := "CodeRabbit unavailable"
	if err == nil {
		status = summarizeReviewLoopCodeRabbitChecks(checks)
		statusLabel = loopReviewPRStatusLabel(status)
	}

	feedback := loopReviewPRFeedback{
		Status:            status,
		StatusLabel:       statusLabel,
		Found:             found,
		CommonInstruction: commonInstruction,
		Pending:           status == reviewLoopCheckPending,
	}
	if !found {
		return feedback, nil
	}
	feedback.Fingerprint = loopReviewFeedbackFingerprint(tasks)
	feedback.RenderedTasks = renderDispatchReviewTasks(tasks)
	return feedback, nil
}

func loopReviewPRStatusLabel(status reviewLoopCheckStatus) string {
	switch status {
	case reviewLoopCheckPending:
		return "CodeRabbit pending"
	case reviewLoopCheckComplete:
		return "CodeRabbit complete"
	default:
		return "CodeRabbit unavailable"
	}
}

func loopReviewFeedbackFingerprint(tasks []dispatchReviewTask) string {
	var parts []string
	for _, task := range tasks {
		parts = append(parts, strings.Join([]string{
			task.Path,
			strconv.Itoa(task.Line),
			task.URL,
			normalizeLoopReviewFeedbackBody(task.Body),
		}, "\x00"))
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x01")
}

func normalizeLoopReviewFeedbackBody(body string) string {
	return strings.ToLower(dispatchWhitespacePattern.ReplaceAllString(strings.TrimSpace(body), " "))
}

func renderLoopReviewPRFeedback(feedback loopReviewPRFeedback) string {
	var builder strings.Builder
	if strings.TrimSpace(feedback.CommonInstruction) != "" {
		builder.WriteString(strings.TrimSpace(feedback.CommonInstruction))
		builder.WriteString("\n\n")
	}
	builder.WriteString(strings.TrimSpace(feedback.RenderedTasks))
	return builder.String()
}

func writeLoopReviewRunArtifact(projectRoot string, report loopReviewReport) error {
	if report.RunID == "" {
		return nil
	}
	dir, err := createLoopArtifactDir(projectRoot, report.RunID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "run.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(loopReviewSummaryMarkdown(report)), 0o644)
}

func loopReviewSummaryMarkdown(report loopReviewReport) string {
	var builder strings.Builder
	builder.WriteString("# Kit Loop Review Run\n\n")
	builder.WriteString(fmt.Sprintf("- Run: `%s`\n", report.RunID))
	builder.WriteString(fmt.Sprintf("- Status: `%s`\n", report.Status))
	if report.Correctness > 0 {
		builder.WriteString(fmt.Sprintf("- Correctness: `%d%%`\n", report.Correctness))
	}
	if report.BaseRef != "" {
		builder.WriteString(fmt.Sprintf("- Base ref: `%s`\n", report.BaseRef))
	}
	if report.PRRef != "" {
		builder.WriteString(fmt.Sprintf("- PR: `%s`\n", report.PRRef))
	}
	if report.PRStatus != "" {
		builder.WriteString(fmt.Sprintf("- PR status: %s\n", report.PRStatus))
	}
	if report.StopReason != "" {
		builder.WriteString(fmt.Sprintf("- Stop reason: %s\n", report.StopReason))
	}
	builder.WriteString("\n## Iterations\n\n")
	for _, iteration := range report.Iterations {
		builder.WriteString(fmt.Sprintf("- %03d", iteration.Index))
		if iteration.Result != nil {
			builder.WriteString(fmt.Sprintf(" done=%t correctness=%d", iteration.Result.Done, iteration.Result.Correctness))
		}
		if iteration.Error != "" {
			builder.WriteString(fmt.Sprintf(" error=%q", iteration.Error))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func outputLoopReviewReport(cmd *cobra.Command, report loopReviewReport, jsonOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	out := cmd.OutOrStdout()
	if report.Status == "dry_run" {
		_, err := fmt.Fprintf(out, "Dry run: %s\n", report.StopReason)
		return err
	}

	result := lastLoopReviewResult(report)
	if result != nil && result.Done {
		fmt.Fprintf(out, "Correctness: %d%%\n", result.Correctness)
		if report.PRStatus != "" {
			fmt.Fprintf(out, "Status: %s\n", report.PRStatus)
		} else {
			fmt.Fprintln(out, "Status: done")
		}
		fmt.Fprintln(out)
		bullets := result.Bullets
		if len(bullets) == 0 {
			bullets = []string{"- No high, medium, or correctness-impacting issues found."}
		}
		for _, bullet := range bullets {
			fmt.Fprintln(out, bullet)
		}
		if strings.Contains(report.PRStatus, "pending") && report.StopReason != "" {
			fmt.Fprintln(out)
			fmt.Fprintln(out, report.StopReason)
		}
		fmt.Fprintln(out, "done")
		return nil
	}

	fmt.Fprintf(out, "Loop review run: %s\n", report.RunID)
	fmt.Fprintf(out, "Status: %s\n", report.Status)
	if report.StopReason != "" {
		fmt.Fprintf(out, "Stop reason: %s\n", report.StopReason)
	}
	if report.ArtifactDir != "" {
		fmt.Fprintf(out, "Artifacts: %s\n", report.ArtifactDir)
	}
	return nil
}

func lastLoopReviewResult(report loopReviewReport) *loopReviewAgentResult {
	for i := len(report.Iterations) - 1; i >= 0; i-- {
		if report.Iterations[i].Result != nil {
			return report.Iterations[i].Result
		}
	}
	return nil
}

func stopLoopReview(report loopReviewReport, err error) (loopReviewReport, error) {
	report.Status = "stopped"
	report.StopReason = err.Error()
	report.EndedAt = time.Now().UTC()
	return report, err
}

func stopLoopReviewWithIteration(
	projectRoot string,
	report loopReviewReport,
	iteration loopReviewIteration,
	err error,
) (loopReviewReport, error) {
	iteration.Error = err.Error()
	iteration.EndedAt = time.Now().UTC()
	iteration.DurationMS = iteration.EndedAt.Sub(iteration.StartedAt).Milliseconds()
	report.Iterations = append(report.Iterations, iteration)
	return stopLoopReviewAfterWrite(projectRoot, report, err)
}

func stopLoopReviewAfterWrite(projectRoot string, report loopReviewReport, err error) (loopReviewReport, error) {
	report.Status = "stopped"
	report.StopReason = err.Error()
	report.EndedAt = time.Now().UTC()
	_ = writeLoopReviewRunArtifact(projectRoot, report)
	return report, err
}

func firstPromptLine(prompt string) string {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return "review prompt"
}
