package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var completeForce bool

var completeAll bool

var completeCmd = &cobra.Command{
	Use:   "complete [feature]",
	Short: "Mark a feature as complete",
	Long: `Mark a feature as complete.

For v2 features, this sets SPEC.md front matter phase to complete. For explicit
legacy staged features, this appends the REFLECTION_COMPLETE marker to TASKS.md.

If no feature is specified, shows an interactive selection of eligible features.

By default, v2 features must be in deliver phase and legacy staged features
must have all TASKS.md checkboxes done. Use --force to override this check.

Use --all to mark every currently eligible feature in the completion list
complete in one command.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runComplete,
}

func init() {
	completeCmd.Flags().BoolVarP(&completeForce, "force", "f", false, "mark complete even if tasks are incomplete")
	completeCmd.Flags().BoolVar(&completeAll, "all", false, "mark all eligible active features complete")
	rootCmd.AddCommand(completeCmd)
}

func runComplete(cmd *cobra.Command, args []string) error {
	if completeAll && len(args) > 0 {
		return fmt.Errorf("--all cannot be used with a specific feature")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	if completeAll {
		candidates, err := eligibleFeaturesForCompletion(specsDir, cfg)
		if err != nil {
			return err
		}
		return markFeaturesComplete(
			cmd.OutOrStdout(),
			cmd.ErrOrStderr(),
			candidates,
			completeForce,
			projectRoot,
			cfg,
		)
	}

	var feat *feature.Feature
	if len(args) == 0 {
		feat, err = selectFeatureForCompletion(specsDir, cfg)
		if err != nil {
			return err
		}
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found", args[0])
		}
	}

	if feat.Phase == feature.PhaseComplete {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "✓ Feature '%s' is already marked complete\n", feat.Slug)
		return err
	}

	if feat.Paused {
		if _, err := validateFeatureCanComplete(feat, completeForce); err != nil {
			return err
		}
		if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
			return err
		}
	}

	return markFeaturesComplete(
		cmd.OutOrStdout(),
		cmd.ErrOrStderr(),
		[]feature.Feature{*feat},
		completeForce,
		projectRoot,
		cfg,
	)
}

// selectFeatureForCompletion shows an interactive numbered list of features
// that have TASKS.md and are not yet marked complete.
func selectFeatureForCompletion(specsDir string, cfg *config.Config) (*feature.Feature, error) {
	candidates, err := eligibleFeaturesForCompletion(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	printSelectionHeader("Select a feature to mark complete:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
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

func eligibleFeaturesForCompletion(specsDir string, cfg *config.Config) ([]feature.Feature, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if f.Paused {
			continue
		}
		if isV2Feature(&f) {
			if f.Phase == feature.PhaseDeliver {
				candidates = append(candidates, f)
			}
			continue
		}
		if f.Phase != feature.PhaseComplete {
			tasksPath := filepath.Join(f.Path, "TASKS.md")
			if _, err := os.Stat(tasksPath); err == nil {
				candidates = append(candidates, f)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features eligible for completion")
	}
	return candidates, nil
}

func validateFeatureCanComplete(feat *feature.Feature, force bool) (string, error) {
	if isV2Feature(feat) {
		if force || feat.Phase == feature.PhaseDeliver {
			return filepath.Join(feat.Path, "SPEC.md"), nil
		}
		return "", fmt.Errorf(
			"SPEC.md phase is %q; v2 features must reach deliver before completion or use --force",
			feat.Phase,
		)
	}

	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		return "", fmt.Errorf("TASKS.md not found at %s — nothing to complete", tasksPath)
	}
	if force {
		return tasksPath, nil
	}

	progress, err := feature.ParseTaskProgress(tasksPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse task progress: %w", err)
	}
	if progress.HasTasks() && progress.Incomplete() > 0 {
		return "", fmt.Errorf(
			"%d/%d tasks incomplete in %s — complete all tasks or use --force to override",
			progress.Incomplete(),
			progress.Total,
			tasksPath,
		)
	}
	return tasksPath, nil
}
