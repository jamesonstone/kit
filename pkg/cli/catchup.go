package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var catchupCopy bool
var catchupOutputOnly bool

var catchupCmd = &cobra.Command{
	Use:   "catchup [feature]",
	Short: "Output a feature catch-up prompt for coding agents",
	Long: `Output a feature-scoped catch-up prompt that helps a coding agent
recover the current stage and state of a feature before implementation
resumes.

This command is intentionally narrower than handoff, summarize, and implement.
It keeps the agent in plan mode, asks questions first, and requires explicit
approval before any implementation begins.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCatchup,
}

func init() {
	catchupCmd.Hidden = true
	catchupCmd.Deprecated = "use `kit resume [feature]`"
	catchupCmd.Flags().BoolVar(&catchupCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	catchupCmd.Flags().BoolVar(
		&catchupOutputOnly,
		"output-only",
		false,
		"output prompt text to stdout instead of copying it to the clipboard",
	)
	addPromptOnlyFlag(catchupCmd)
	rootCmd.AddCommand(catchupCmd)
}

func runCatchup(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)
	var feat *feature.Feature
	if len(args) == 1 {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}
	} else {
		feat, err = selectFeatureForCatchup(specsDir, cfg)
		if err != nil {
			return err
		}
	}

	return outputCatchupPromptForFeature(feat, projectRoot, outputOnly, catchupCopy, "catchup (supporting step)")
}

func outputCatchupPromptForFeature(
	feat *feature.Feature,
	projectRoot string,
	outputOnly bool,
	copy bool,
	currentStep string,
) error {
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}

	prompt := buildCatchupPrompt(feat, status, projectRoot)
	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, copy); err != nil {
		return err
	}

	if !outputOnly {
		printWorkflowInstructions(currentStep, []string{
			"answer the agent's clarification questions to restore context",
			"approve a move to implementation only when you want coding to begin",
		})
	}

	return nil
}

func selectFeatureForCatchup(specsDir string, cfg *config.Config) (*feature.Feature, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(features) == 0 {
		return nil, fmt.Errorf("no features available to catch up on")
	}

	printSelectionHeader("Select a feature to catch up on:")
	for i, f := range features {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(features) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := features[num-1]
	return &selected, nil
}
