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

var removeYes bool

var removeCmd = &cobra.Command{
	Use:   "remove [feature]",
	Short: "Remove a feature and its lifecycle state",
	Long: `Remove a feature directory and its persisted lifecycle state.

If no feature is specified, shows an interactive selection of existing
features.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "skip the deletion confirmation prompt")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
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
		feat, err = selectFeatureForRemove(specsDir, cfg)
		if err != nil {
			return err
		}
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	}

	if !removeYes {
		confirmed, err := confirmFeatureRemoval(feat)
		if err != nil {
			return err
		}
		if !confirmed {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Removal canceled.")
			return err
		}
	}

	if err := os.RemoveAll(feat.Path); err != nil {
		return fmt.Errorf("failed to remove feature directory %s: %w", feat.Path, err)
	}

	if err := feature.ClearPersistedState(projectRoot, cfg, feat); err != nil {
		return fmt.Errorf("feature '%s' was removed but failed to clear persisted lifecycle state: %w", feat.Slug, err)
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("feature '%s' was removed but failed to refresh PROJECT_PROGRESS_SUMMARY.md: %w", feat.Slug, err)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "🗑️ Removed feature '%s'\n", feat.Slug)
	return err
}

func selectFeatureForRemove(specsDir string, cfg *config.Config) (*feature.Feature, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(features) == 0 {
		return nil, fmt.Errorf("no features available to remove")
	}

	printSelectionHeader("Select a feature to remove:")
	for i, feat := range features {
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
	if err != nil || num < 1 || num > len(features) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := features[num-1]
	return &selected, nil
}

func confirmFeatureRemoval(feat *feature.Feature) (bool, error) {
	fmt.Printf("Remove feature '%s' at %s? [y/N]: ", feat.DirName, feat.Path)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
}
