package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

type loopStage string

const (
	loopStageClarify   loopStage = "clarify"
	loopStageReady     loopStage = "ready"
	loopStageImplement loopStage = "implement"
	loopStageValidate  loopStage = "validate"
	loopStageReflect   loopStage = "reflect"
	loopStageDeliver   loopStage = "deliver"
	loopStageComplete  loopStage = "complete"
	loopStageBlocked   loopStage = "blocked"
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
	Short:         "Run workflow and review agent loops",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Run Kit agent loops.

Use kit loop prompt [feature] to create work-to-completion prompts, kit loop
workflow [feature] to run the v2 single-SPEC workflow through a configured
local agent, and kit loop review for changed-code correctness review.

Configure the local agent in .kit.yaml:

loop:
  min_confidence: 95
  max_iterations: 20
  agent:
    command: codex
    args: ["--ask-for-approval", "never", "exec", "--model", "gpt-5.5", "--sandbox", "workspace-write", "--ignore-user-config", "--color", "never", "-"]`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLoop,
}

func init() {
	addWorkflowLoopFlags(loopCmd)
	loopCmd.AddCommand(newLoopPromptCommand())
	loopCmd.AddCommand(newLoopWorkflowCommand())
	loopCmd.AddCommand(newLoopReviewCommand())
	rootCmd.AddCommand(loopCmd)
}

func newLoopWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "workflow [feature]",
		Short:         "Run the feature workflow through a confidence-gated agent loop",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Run the Kit feature workflow as an autonomous, confidence-gated loop.

The workflow loop uses the v2 kit spec supervisor prompt and SPEC.md front
matter phases as durable state. It wraps the prompt with a machine-readable
loop contract, sends the prompt to the configured local agent command over
stdin, validates the result, and repeats until completion or a blocker.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runLoop,
	}
	addWorkflowLoopFlags(cmd)
	return cmd
}

func addWorkflowLoopFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&loopDryRun, "dry-run", false, "show the next loop action without running the configured agent")
	cmd.Flags().StringVar(&loopUntil, "until", "complete", "run until this v2 phase is complete: clarify, ready, implement, validate, reflect, deliver, complete")
	cmd.Flags().IntVar(&loopMinConfidence, "min-confidence", 0, "minimum agent confidence required to advance (0 uses loop config, goal_percentage, then 95)")
	cmd.Flags().IntVar(&loopMaxIterations, "max-iterations", 0, "maximum loop iterations (0 uses loop config, then 20)")
	cmd.Flags().BoolVar(&loopJSON, "json", false, "output loop report as JSON")
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
	ReflectRunner reflectEvidenceRunner
	ReflectNow    func() time.Time
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
			return nil, fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first for v2 work, or 'kit legacy brainstorm %s' for staged migration work", args[0], args[0], args[0])
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
