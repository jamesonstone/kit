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
	"github.com/jamesonstone/kit/internal/rollup"
)

var completeForce bool

var completeCmd = &cobra.Command{
	Use:   "complete [feature]",
	Short: "Mark a feature as complete",
	Long: `Mark a feature as complete by appending the REFLECTION_COMPLETE marker
to its TASKS.md file. This transitions the feature's phase from "reflect"
to "complete" in kit status.

If no feature is specified, shows an interactive selection of eligible features.

By default, all tasks in TASKS.md must be marked done (- [x]) before
the feature can be completed. Use --force to override this check.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runComplete,
}

func init() {
	completeCmd.Flags().BoolVarP(&completeForce, "force", "f", false, "mark complete even if tasks are incomplete")
	rootCmd.AddCommand(completeCmd)
}

func runComplete(cmd *cobra.Command, args []string) error {
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

	// verify TASKS.md exists
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		return fmt.Errorf("TASKS.md not found at %s — nothing to complete", tasksPath)
	}

	// check current phase
	phase := feature.DeterminePhaseFromTasks(tasksPath)

	if phase == feature.PhaseComplete {
		fmt.Printf("✓ Feature '%s' is already marked complete\n", feat.Slug)
		return nil
	}

	// check that all tasks are done unless --force
	if !completeForce {
		progress, err := feature.ParseTaskProgress(tasksPath)
		if err != nil {
			return fmt.Errorf("failed to parse task progress: %w", err)
		}

		if progress.HasTasks() && progress.Incomplete() > 0 {
			return fmt.Errorf(
				"%d/%d tasks incomplete in %s — complete all tasks or use --force to override",
				progress.Incomplete(), progress.Total, tasksPath,
			)
		}
	}

	// append the reflection complete marker
	if err := appendReflectionMarker(tasksPath); err != nil {
		return fmt.Errorf("failed to mark feature complete: %w", err)
	}

	fmt.Printf("✅ Feature '%s' marked complete\n", feat.Slug)

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
	} else {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	return nil
}

// selectFeatureForCompletion shows an interactive numbered list of features
// that have TASKS.md and are not yet marked complete.
func selectFeatureForCompletion(specsDir string) (*feature.Feature, error) {
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

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to mark complete:" + reset)
	fmt.Println()
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(whiteBold + "Enter number: " + reset)

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
