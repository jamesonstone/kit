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
	"github.com/jamesonstone/kit/internal/rollup"
)

var pauseCmd = &cobra.Command{
	Use:   "pause [feature]",
	Short: "Pause an in-flight feature",
	Long: `Pause an in-flight feature without changing its underlying workflow
phase.

If no feature is specified, shows an interactive selection of non-complete
features.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPause,
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

func runPause(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	var feat *feature.Feature
	if len(args) == 0 {
		feat, err = selectFeatureForPause(specsDir, cfg)
		if err != nil {
			return err
		}
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	}

	if feat.Phase == feature.PhaseComplete {
		return fmt.Errorf("feature '%s' is complete and cannot be paused", feat.Slug)
	}

	alreadyPaused := feat.Paused
	if !alreadyPaused {
		if err := feature.PersistPaused(projectRoot, cfg, feat, true); err != nil {
			return fmt.Errorf("failed to persist paused state: %w", err)
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("feature '%s' pause state updated but failed to refresh PROJECT_PROGRESS_SUMMARY.md: %w", feat.Slug, err)
	}

	if alreadyPaused {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "⏸️ Feature '%s' is already paused\n", feat.Slug)
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "⏸️ Feature '%s' paused\n", feat.Slug)
	return err
}

func selectFeatureForPause(specsDir string, cfg *config.Config) (*feature.Feature, error) {
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

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no in-flight features available to pause")
	}

	printSelectionHeader("Select a feature to pause:")
	for i, feat := range candidates {
		label := feat.DirName
		if feat.Paused {
			label += " (paused)"
		}
		fmt.Printf("  [%d] %s (%s)\n", i+1, label, feat.Phase)
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
