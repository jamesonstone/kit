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

var resumeCopy bool
var resumeOutputOnly bool

type resumeCandidate struct {
	Feature feature.Feature
	Label   string
}

var resumeCmd = &cobra.Command{
	Use:   "resume [feature]",
	Short: "Resume work on a feature through the canonical prompt flow",
	Long: `Resume work on a feature using the canonical Kit flow.

When the target is a backlog item, ` + "`resume`" + ` reuses backlog pickup and
outputs the brainstorm planning prompt. When the target is not a backlog item,
it reuses the catch-up prompt behavior that restores context before further
work begins.

The resume prompt identifies the active feature, current phase, next canonical
artifact, recommended command, known blockers, and validation state when that
state is known from repository artifacts.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runResume,
}

func init() {
	resumeCmd.Flags().BoolVar(&resumeCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	resumeCmd.Flags().BoolVar(
		&resumeOutputOnly,
		"output-only",
		false,
		"output prompt text to stdout instead of copying it to the clipboard",
	)
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
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
		feat, err = selectFeatureForResume(specsDir, cfg)
		if err != nil {
			return err
		}
	}

	if feature.IsBacklogItem(*feat) {
		return resumeBacklogFeature(projectRoot, cfg, feat, outputOnly, resumeCopy, "", "resume backlog")
	}

	return outputCatchupPromptForFeature(feat, projectRoot, outputOnly, resumeCopy, "resume")
}

func selectFeatureForResume(specsDir string, cfg *config.Config) (*feature.Feature, error) {
	candidates, err := buildResumeCandidates(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no resumable features available")
	}

	printSelectionHeader("Select a feature to resume:")
	for i, candidate := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, candidate.Label)
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

	selected := candidates[num-1].Feature
	return &selected, nil
}

func buildResumeCandidates(specsDir string, cfg *config.Config) ([]resumeCandidate, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	active, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	candidates := make([]resumeCandidate, 0, len(features))
	for _, feat := range features {
		if feat.Phase == feature.PhaseComplete {
			continue
		}
		if feat.Paused && !feature.IsBacklogItem(feat) {
			candidates = append(candidates, resumeCandidate{
				Feature: feat,
				Label:   fmt.Sprintf("%s (paused, %s)", feat.DirName, feat.Phase),
			})
		}
	}

	if active != nil && !active.Paused && active.Phase != feature.PhaseComplete {
		candidates = append(candidates, resumeCandidate{
			Feature: *active,
			Label:   fmt.Sprintf("%s (active, %s)", active.DirName, active.Phase),
		})
	}

	backlog, err := feature.ListBacklogFeatures(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	for _, feat := range backlog {
		candidates = append(candidates, resumeCandidate{
			Feature: feat,
			Label:   fmt.Sprintf("%s (backlog, %s)", feat.DirName, feat.Phase),
		})
	}

	return candidates, nil
}
