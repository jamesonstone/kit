package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
)

var completeForce bool
var completeAll bool

var completeCmd = &cobra.Command{
	Use:   "complete [feature]",
	Short: "Mark a feature as complete",
	Long: `Mark a feature as complete by appending the REFLECTION_COMPLETE marker
to its TASKS.md file. This transitions the feature's phase from "reflect"
to "complete" in kit status.

If no feature is specified, shows an interactive selection of eligible features.

By default, all tasks in TASKS.md must be marked done (- [x]) before
the feature can be completed. Use --force to override this check.

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
		candidates, err := eligibleFeaturesForCompletion(specsDir)
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
		feat, err = selectFeatureForCompletion(specsDir)
		if err != nil {
			return err
		}
	} else {
		feat, err = feature.Resolve(specsDir, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found", args[0])
		}
	}

	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if feature.DeterminePhaseFromTasks(tasksPath) == feature.PhaseComplete {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "✓ Feature '%s' is already marked complete\n", feat.Slug)
		return err
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
func selectFeatureForCompletion(specsDir string) (*feature.Feature, error) {
	candidates, err := eligibleFeaturesForCompletion(specsDir)
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

func eligibleFeaturesForCompletion(specsDir string) ([]feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		tasksPath := filepath.Join(f.Path, "TASKS.md")
		if _, err := os.Stat(tasksPath); err != nil {
			continue
		}
		if f.Phase != feature.PhaseComplete {
			candidates = append(candidates, f)
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features eligible for completion")
	}
	return candidates, nil
}

func validateFeatureCanComplete(feat *feature.Feature, force bool) (string, error) {
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

func markFeaturesComplete(
	out io.Writer,
	errOut io.Writer,
	features []feature.Feature,
	force bool,
	projectRoot string,
	cfg *config.Config,
) error {
	tasksPaths := make([]string, len(features))
	for i := range features {
		tasksPath, err := validateFeatureCanComplete(&features[i], force)
		if err != nil {
			return fmt.Errorf("feature '%s': %w", features[i].Slug, err)
		}
		tasksPaths[i] = tasksPath
	}

	for i := range features {
		if err := appendReflectionMarker(tasksPaths[i]); err != nil {
			return fmt.Errorf("failed to mark feature '%s' complete: %w", features[i].Slug, err)
		}
		if _, err := fmt.Fprintf(out, "✅ Feature '%s' marked complete\n", features[i].Slug); err != nil {
			return err
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		_, _ = fmt.Fprintf(errOut, "  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		return nil
	}
	_, err := fmt.Fprintln(out, "  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	return err
}

// appendReflectionMarker appends the REFLECTION_COMPLETE marker to a TASKS.md file.
func appendReflectionMarker(tasksPath string) error {
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return err
	}

	content := string(data)

	// already present
	if strings.Contains(content, feature.ReflectionCompleteMarker) {
		return nil
	}

	// ensure trailing newline before marker
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += "\n" + feature.ReflectionCompleteMarker + "\n"

	return os.WriteFile(tasksPath, []byte(content), 0644)
}
