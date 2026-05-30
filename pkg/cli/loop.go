package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/jamesonstone/kit/internal/verify"
)

type loopStage string

const (
	loopStageSpec      loopStage = "spec"
	loopStagePlan      loopStage = "plan"
	loopStageTasks     loopStage = "tasks"
	loopStageImplement loopStage = "implement"
	loopStageReflect   loopStage = "reflect"
	loopStageComplete  loopStage = "complete"
)

const loopSchemaVersion = 1

var (
	loopDryRun        bool
	loopUntil         string
	loopMinConfidence int
	loopMaxIterations int
	loopJSON          bool
)

var loopCmd = &cobra.Command{
	Use:           "loop [feature]",
	Short:         "Run the feature workflow through a confidence-gated agent loop",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Run the Kit workflow as an autonomous, confidence-gated loop.

The loop keeps existing workflow commands as prompt and artifact builders. It
selects the next strict stage, wraps that stage prompt with a machine-readable
loop contract, sends the prompt to the configured local agent command over
stdin, validates the result, and repeats until completion or a blocker.

Configure the local agent in .kit.yaml:

loop:
  min_confidence: 95
  max_iterations: 20
  agent:
    command: your-agent
    args: ["run", "--stdin"]`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLoop,
}

func init() {
	loopCmd.Flags().BoolVar(&loopDryRun, "dry-run", false, "show the next loop action without running the configured agent")
	loopCmd.Flags().StringVar(&loopUntil, "until", "complete", "run until this stage is complete: spec, plan, tasks, implement, reflect, complete")
	loopCmd.Flags().IntVar(&loopMinConfidence, "min-confidence", 0, "minimum agent confidence required to advance (0 uses loop config, goal_percentage, then 95)")
	loopCmd.Flags().IntVar(&loopMaxIterations, "max-iterations", 0, "maximum loop iterations (0 uses loop config, then 20)")
	loopCmd.Flags().BoolVar(&loopJSON, "json", false, "output loop report as JSON")
	rootCmd.AddCommand(loopCmd)
}

type loopOptions struct {
	ProjectRoot   string
	Config        *config.Config
	Feature       *feature.Feature
	Until         loopStage
	MinConfidence int
	MaxIterations int
	DryRun        bool
	JSON          bool
	Agent         config.LoopAgentConfig
}

type loopReport struct {
	SchemaVersion int             `json:"schema_version"`
	RunID         string          `json:"run_id,omitempty"`
	Feature       string          `json:"feature"`
	Status        string          `json:"status"`
	StopReason    string          `json:"stop_reason,omitempty"`
	Until         loopStage       `json:"until"`
	MinConfidence int             `json:"min_confidence"`
	MaxIterations int             `json:"max_iterations"`
	ArtifactDir   string          `json:"artifact_dir,omitempty"`
	StartedAt     time.Time       `json:"started_at"`
	EndedAt       time.Time       `json:"ended_at"`
	Iterations    []loopIteration `json:"iterations"`
}

type loopIteration struct {
	Index       int              `json:"index"`
	Stage       loopStage        `json:"stage"`
	Before      loopStageState   `json:"before"`
	After       loopStageState   `json:"after"`
	PromptPath  string           `json:"prompt_path,omitempty"`
	StdoutPath  string           `json:"stdout_path,omitempty"`
	StderrPath  string           `json:"stderr_path,omitempty"`
	Result      *loopAgentResult `json:"result,omitempty"`
	ExitCode    int              `json:"exit_code,omitempty"`
	Error       string           `json:"error,omitempty"`
	StartedAt   time.Time        `json:"started_at"`
	EndedAt     time.Time        `json:"ended_at"`
	DurationMS  int64            `json:"duration_ms"`
	DryRun      bool             `json:"dry_run,omitempty"`
	Description string           `json:"description,omitempty"`
}

type loopStageState struct {
	Stage       loopStage `json:"stage"`
	Diagnostics []string  `json:"diagnostics,omitempty"`
	TasksTotal  int       `json:"tasks_total,omitempty"`
	TasksDone   int       `json:"tasks_done,omitempty"`
}

type loopAgentResult struct {
	Stage      loopStage `json:"stage"`
	Status     string    `json:"status"`
	Confidence int       `json:"confidence"`
	Blockers   []string  `json:"blockers,omitempty"`
}

type loopAgentExecution struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

var loopResultPattern = regexp.MustCompile(`(?m)^KIT_LOOP_RESULT:\s*(\{.*\})\s*$`)

func runLoop(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	specsDir := cfg.SpecsPath(projectRoot)
	feat, err := resolveLoopFeature(specsDir, cfg, args)
	if err != nil {
		return err
	}
	until, err := parseLoopStage(loopUntil)
	if err != nil {
		return err
	}
	opts := loopOptions{
		ProjectRoot:   projectRoot,
		Config:        cfg,
		Feature:       feat,
		Until:         until,
		MinConfidence: effectiveLoopMinConfidence(cfg, loopMinConfidence),
		MaxIterations: effectiveLoopMaxIterations(cfg, loopMaxIterations),
		DryRun:        loopDryRun,
		JSON:          loopJSON,
		Agent:         cfg.Loop.Agent,
	}
	report, runErr := executeLoop(cmd.Context(), opts)
	outputErr := outputLoopReport(cmd, report, loopJSON)
	if outputErr != nil {
		return outputErr
	}
	if runErr != nil {
		return &silentCLIError{err: runErr}
	}
	return runErr
}

func resolveLoopFeature(specsDir string, cfg *config.Config, args []string) (*feature.Feature, error) {
	if len(args) == 1 {
		feat, err := loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return nil, fmt.Errorf("feature '%s' not found. Run 'kit brainstorm %s' or 'kit spec %s' first", args[0], args[0], args[0])
		}
		return feat, nil
	}
	return selectFeatureForLoop(specsDir, cfg)
}

func selectFeatureForLoop(specsDir string, cfg *config.Config) (*feature.Feature, error) {
	candidates, err := loopFeatureCandidates(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, errors.New("no loopable features available")
	}

	printSelectionHeader("Select a feature to loop:")
	for i, f := range candidates {
		label := fmt.Sprintf("%s (%s)", f.DirName, f.Phase)
		if f.Paused {
			label += ", paused"
		}
		fmt.Printf("  [%d] %s\n", i+1, label)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

func loopFeatureCandidates(specsDir string, cfg *config.Config) ([]feature.Feature, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, feat := range features {
		if feat.Phase == feature.PhaseComplete {
			continue
		}
		candidates = append(candidates, feat)
	}
	return candidates, nil
}

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

func buildLoopPromptForStage(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage, minConfidence int) (string, error) {
	if err := ensureLoopStageArtifact(projectRoot, cfg, feat, stage); err != nil {
		return "", err
	}
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	var base string
	switch stage {
	case loopStageSpec:
		base = buildSpecTemplatePrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg)
	case loopStagePlan:
		base = buildStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)
	case loopStageTasks:
		base = buildTasksPrompt(feat, projectRoot, cfg)
	case loopStageImplement:
		summary, _ := feature.ExtractSpecSummary(specPath)
		base = buildImplementationPrompt(feat, brainstormPath, specPath, planPath, tasksPath, summary, projectRoot)
	case loopStageReflect:
		base = buildReflectPrompt(
			projectRoot,
			filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
			cfg.ProgressSummaryPath(projectRoot),
			brainstormPath,
			specPath,
			planPath,
			tasksPath,
			feat.Slug,
		)
	default:
		return "", fmt.Errorf("stage %s does not produce a loop prompt", stage)
	}

	return appendLoopContract(prepareAgentPromptForFeature(base, feat.Path), stage, minConfidence), nil
}

func ensureLoopStageArtifact(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage) error {
	switch stage {
	case loopStageSpec:
		specPath := filepath.Join(feat.Path, "SPEC.md")
		if !document.Exists(specPath) {
			content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(specPath, content); err != nil {
				return fmt.Errorf("failed to create SPEC.md: %w", err)
			}
		}
		if effectivePromptProfile(feat.Path) == promptProfileFrontend {
			if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
				return err
			}
		}
	case loopStagePlan:
		planPath := filepath.Join(feat.Path, "PLAN.md")
		if !document.Exists(planPath) {
			content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(planPath, content); err != nil {
				return fmt.Errorf("failed to create PLAN.md: %w", err)
			}
		}
		if effectivePromptProfile(feat.Path) == promptProfileFrontend {
			specPath := filepath.Join(feat.Path, "SPEC.md")
			if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
				return err
			}
			if _, err := ensureFrontendProfileDependencyRows(planPath, document.TypePlan, feat.DirName); err != nil {
				return err
			}
		}
	case loopStageTasks:
		tasksPath := filepath.Join(feat.Path, "TASKS.md")
		if !document.Exists(tasksPath) {
			content := templates.BuildTasksArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(tasksPath, content); err != nil {
				return fmt.Errorf("failed to create TASKS.md: %w", err)
			}
		}
	}
	return rollup.Update(projectRoot, cfg)
}

func appendLoopContract(prompt string, stage loopStage, minConfidence int) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimRight(prompt, "\n"))
	builder.WriteString("\n\n## Kit Loop Contract\n\n")
	builder.WriteString("This run is controlled by `kit loop`. Complete the current stage, write all durable changes to repository files, and end your final output with exactly one machine-readable result line.\n\n")
	builder.WriteString("- Do not report `status: \"done\"` unless the stage artifact or implementation state is actually complete.\n")
	builder.WriteString(fmt.Sprintf("- Do not proceed with unresolved assumptions or confidence below %d.\n", minConfidence))
	builder.WriteString("- If blocked, set `status` to `blocked`, include concrete blockers, and do not guess.\n")
	builder.WriteString("- The result line must be a single line of JSON prefixed with `KIT_LOOP_RESULT:`.\n\n")
	builder.WriteString("Required result line:\n\n")
	builder.WriteString("```text\n")
	builder.WriteString(fmt.Sprintf("KIT_LOOP_RESULT: {\"stage\":\"%s\",\"status\":\"done\",\"confidence\":%d,\"blockers\":[]}\n", stage, minConfidence))
	builder.WriteString("```\n")
	return builder.String()
}

func runLoopAgent(ctx context.Context, opts loopOptions, stage loopStage, iteration int, prompt string) loopAgentExecution {
	cmd := exec.CommandContext(ctx, opts.Agent.Command, opts.Agent.Args...)
	cmd.Dir = opts.ProjectRoot
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(),
		"KIT_LOOP_STAGE="+string(stage),
		"KIT_LOOP_FEATURE="+opts.Feature.DirName,
		fmt.Sprintf("KIT_LOOP_MIN_CONFIDENCE=%d", opts.MinConfidence),
		fmt.Sprintf("KIT_LOOP_ITERATION=%d", iteration),
	)
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

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func parseLoopAgentResult(stdout, stderr string) (*loopAgentResult, error) {
	matches := loopResultPattern.FindAllStringSubmatch(stdout+"\n"+stderr, -1)
	if len(matches) == 0 {
		return nil, errors.New("agent output did not include KIT_LOOP_RESULT JSON")
	}
	raw := matches[len(matches)-1][1]
	var result loopAgentResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("invalid KIT_LOOP_RESULT JSON: %w", err)
	}
	return &result, nil
}

func validateLoopAgentResult(result loopAgentResult, expected loopStage, minConfidence int) error {
	if result.Stage != expected {
		return fmt.Errorf("agent reported stage %q, expected %q", result.Stage, expected)
	}
	if result.Status != "done" {
		if len(result.Blockers) > 0 {
			return fmt.Errorf("agent blocked at %s: %s", expected, strings.Join(result.Blockers, "; "))
		}
		return fmt.Errorf("agent reported status %q at %s", result.Status, expected)
	}
	if len(result.Blockers) > 0 {
		return fmt.Errorf("agent reported blockers at %s: %s", expected, strings.Join(result.Blockers, "; "))
	}
	if result.Confidence < minConfidence {
		return fmt.Errorf("agent confidence %d is below required %d", result.Confidence, minConfidence)
	}
	return nil
}

func validateLoopProgress(before, after loopStageState, previousImplement feature.TaskProgress) error {
	if loopStageRank(after.Stage) > loopStageRank(before.Stage) {
		return nil
	}
	if len(after.Diagnostics) > 0 {
		return fmt.Errorf("strict validation failed for %s: %s", after.Stage, strings.Join(after.Diagnostics, "; "))
	}
	if before.Stage == loopStageImplement && after.Stage == loopStageImplement {
		if after.TasksDone > previousImplement.Complete || after.TasksTotal != previousImplement.Total {
			return nil
		}
	}
	return fmt.Errorf("stage %s did not advance after agent reported done", before.Stage)
}

func stopOnFailedVerification(projectRoot string, feat *feature.Feature, since time.Time) error {
	run, ok, err := runstore.LatestForFeature(projectRoot, feat.DirName)
	if err != nil {
		return err
	}
	if ok && !since.IsZero() && run.StartedAt.Before(since) {
		return nil
	}
	if !ok || run.Status != verify.RunStatusFail {
		return nil
	}
	return fmt.Errorf("latest verification run failed: %s", run.RunID)
}

func resolveStrictLoopStage(projectRoot string, feat *feature.Feature) loopStageState {
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	if diagnostics := validateLoopDocument(projectRoot, feat, specPath, document.TypeSpec); len(diagnostics) > 0 {
		return loopStageState{Stage: loopStageSpec, Diagnostics: diagnostics}
	}
	if diagnostics := validateLoopDocument(projectRoot, feat, planPath, document.TypePlan); len(diagnostics) > 0 {
		return loopStageState{Stage: loopStagePlan, Diagnostics: diagnostics}
	}
	if diagnostics := validateLoopDocument(projectRoot, feat, tasksPath, document.TypeTasks); len(diagnostics) > 0 {
		progress, _ := feature.ParseTaskProgress(tasksPath)
		return loopStageState{Stage: loopStageTasks, Diagnostics: diagnostics, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	progress, err := feature.ParseTaskProgress(tasksPath)
	if err != nil {
		return loopStageState{Stage: loopStageTasks, Diagnostics: []string{err.Error()}}
	}
	if progress.Total == 0 {
		return loopStageState{Stage: loopStageTasks, Diagnostics: []string{"TASKS.md has no markdown checkbox tasks"}}
	}
	if progress.Complete < progress.Total {
		return loopStageState{Stage: loopStageImplement, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return loopStageState{Stage: loopStageReflect, Diagnostics: []string{err.Error()}, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	if strings.Contains(string(data), feature.ReflectionCompleteMarker) {
		return loopStageState{Stage: loopStageComplete, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	return loopStageState{Stage: loopStageReflect, TasksTotal: progress.Total, TasksDone: progress.Complete}
}

func validateLoopDocument(projectRoot string, feat *feature.Feature, path string, docType document.DocumentType) []string {
	if !document.Exists(path) {
		return []string{fmt.Sprintf("%s not found", filepath.Base(path))}
	}
	doc, err := document.ParseFile(path, docType)
	if err != nil {
		return []string{err.Error()}
	}
	var diagnostics []string
	for _, validationErr := range doc.Validate() {
		diagnostics = append(diagnostics, validationErr.Error())
	}
	diagnostics = append(diagnostics, featureMetadataIdentityErrors(doc, feat.DirName)...)
	diagnostics = append(diagnostics, featureRulesetReferenceErrors(projectRoot, doc)...)
	if doc.HasUnresolvedPlaceholders() {
		diagnostics = append(diagnostics, fmt.Sprintf("%s has unresolved TODO placeholders", filepath.Base(path)))
	}
	return diagnostics
}

func createLoopArtifactDir(projectRoot, runID string) (string, error) {
	abs := filepath.Join(projectRoot, filepath.FromSlash(loopRelArtifactDir(runID)))
	if err := os.MkdirAll(abs, 0755); err != nil {
		return "", err
	}
	return abs, nil
}

func loopRelArtifactDir(runID string) string {
	return filepath.ToSlash(filepath.Join(".kit", "loops", runID))
}

func writeLoopIterationFile(artifactDir, runID string, index int, name, content string) (string, error) {
	dir := filepath.Join(artifactDir, fmt.Sprintf("%03d", index))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(loopRelArtifactDir(runID), fmt.Sprintf("%03d", index), name)), nil
}

func writeLoopRunArtifact(projectRoot string, report loopReport) error {
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
	if err := os.WriteFile(filepath.Join(dir, "run.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(loopSummaryMarkdown(report)), 0644)
}

func loopSummaryMarkdown(report loopReport) string {
	var builder strings.Builder
	builder.WriteString("# Kit Loop Run\n\n")
	builder.WriteString(fmt.Sprintf("- Run: `%s`\n", report.RunID))
	builder.WriteString(fmt.Sprintf("- Feature: `%s`\n", report.Feature))
	builder.WriteString(fmt.Sprintf("- Status: `%s`\n", report.Status))
	if report.StopReason != "" {
		builder.WriteString(fmt.Sprintf("- Stop reason: %s\n", report.StopReason))
	}
	builder.WriteString("\n## Iterations\n\n")
	if len(report.Iterations) == 0 {
		builder.WriteString("none\n")
		return builder.String()
	}
	for _, iteration := range report.Iterations {
		builder.WriteString(fmt.Sprintf("- %03d `%s`", iteration.Index, iteration.Stage))
		if iteration.Result != nil {
			builder.WriteString(fmt.Sprintf(" status=%s confidence=%d", iteration.Result.Status, iteration.Result.Confidence))
		}
		if iteration.Error != "" {
			builder.WriteString(fmt.Sprintf(" error=%q", iteration.Error))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func outputLoopReport(cmd *cobra.Command, report loopReport, jsonOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	out := cmd.OutOrStdout()
	if report.Status == "dry_run" {
		if len(report.Iterations) > 0 {
			_, err := fmt.Fprintf(out, "Dry run: %s\n", report.Iterations[0].Description)
			return err
		}
		_, err := fmt.Fprintln(out, "Dry run: no action")
		return err
	}
	fmt.Fprintf(out, "Loop run: %s\n", report.RunID)
	fmt.Fprintf(out, "Feature: %s\n", report.Feature)
	fmt.Fprintf(out, "Status: %s\n", report.Status)
	if report.StopReason != "" {
		fmt.Fprintf(out, "Stop reason: %s\n", report.StopReason)
	}
	if report.ArtifactDir != "" {
		fmt.Fprintf(out, "Artifacts: %s\n", report.ArtifactDir)
	}
	if len(report.Iterations) > 0 {
		fmt.Fprintln(out, "Iterations:")
		for _, iteration := range report.Iterations {
			if iteration.Result != nil {
				fmt.Fprintf(out, "  - %03d %s: %s confidence=%d\n", iteration.Index, iteration.Stage, iteration.Result.Status, iteration.Result.Confidence)
			} else {
				fmt.Fprintf(out, "  - %03d %s\n", iteration.Index, iteration.Stage)
			}
		}
	}
	return nil
}

func loopDryRunDescription(opts loopOptions, state loopStageState) string {
	if loopTargetComplete(state.Stage, opts.Until) || state.Stage == loopStageComplete {
		return fmt.Sprintf("target stage %s is already complete for %s", opts.Until, opts.Feature.DirName)
	}
	command := opts.Agent.Command
	if command == "" {
		command = "<configured-agent>"
	}
	args := strings.Join(opts.Agent.Args, " ")
	if args != "" {
		command += " " + args
	}
	return fmt.Sprintf("would run %s stage for %s with `%s`", state.Stage, opts.Feature.DirName, command)
}

func parseLoopStage(value string) (loopStage, error) {
	switch loopStage(strings.ToLower(strings.TrimSpace(value))) {
	case loopStageSpec:
		return loopStageSpec, nil
	case loopStagePlan:
		return loopStagePlan, nil
	case loopStageTasks:
		return loopStageTasks, nil
	case loopStageImplement:
		return loopStageImplement, nil
	case loopStageReflect:
		return loopStageReflect, nil
	case loopStageComplete:
		return loopStageComplete, nil
	default:
		return "", fmt.Errorf("invalid --until stage %q", value)
	}
}

func loopTargetComplete(current, until loopStage) bool {
	return loopStageRank(current) > loopStageRank(until)
}

func loopStageRank(stage loopStage) int {
	switch stage {
	case loopStageSpec:
		return 1
	case loopStagePlan:
		return 2
	case loopStageTasks:
		return 3
	case loopStageImplement:
		return 4
	case loopStageReflect:
		return 5
	case loopStageComplete:
		return 6
	default:
		return 0
	}
}

func effectiveLoopMinConfidence(cfg *config.Config, override int) int {
	if override > 0 {
		return clampPercentage(override)
	}
	if cfg != nil && cfg.Loop.MinConfidence > 0 {
		return clampPercentage(cfg.Loop.MinConfidence)
	}
	if cfg != nil && cfg.GoalPercentage > 0 {
		return clampPercentage(cfg.GoalPercentage)
	}
	return 95
}

func effectiveLoopMaxIterations(cfg *config.Config, override int) int {
	if override > 0 {
		return override
	}
	if cfg != nil && cfg.Loop.MaxIterations > 0 {
		return cfg.Loop.MaxIterations
	}
	return 20
}

func clampPercentage(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
