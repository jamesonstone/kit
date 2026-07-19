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
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
)

var completeForce bool
var completeAll bool

var completeCmd = &cobra.Command{
	Use:   "complete [feature]",
	Short: "Mark a feature as complete",
	Long: `Mark a feature as complete.

For versioned living specs, this preserves the workflow version and sets SPEC.md
front matter phase to complete. For explicit
legacy staged features, this appends the REFLECTION_COMPLETE marker to TASKS.md.

If no feature is specified, shows an interactive selection of eligible features.

By default, living specs must be in deliver phase and satisfy their version-specific
completion gate; legacy staged features
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
		if isLivingSpecFeature(&f) {
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
	if isLivingSpecFeature(feat) {
		specPath := filepath.Join(feat.Path, "SPEC.md")
		doc, err := document.ParseFile(specPath, document.TypeSpec)
		if err != nil {
			return "", fmt.Errorf("parse SPEC.md: %w", err)
		}
		if doc.Metadata != nil && doc.Metadata.WorkflowVersion == document.WorkflowVersionV3 {
			if validationErrors := doc.Validate(); len(validationErrors) > 0 {
				return "", fmt.Errorf("v3 completion gate failed: %s", validationErrors[0].Error())
			}
			if doc.HasUnresolvedPlaceholders() {
				return "", fmt.Errorf("v3 completion gate failed: SPEC.md still contains pending TODO placeholders")
			}
		}
		if force || feat.Phase == feature.PhaseDeliver {
			return specPath, nil
		}
		return "", fmt.Errorf(
			"SPEC.md phase is %q; living specs must reach deliver before completion or use --force",
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
		if err := markFeatureComplete(&features[i], tasksPaths[i]); err != nil {
			return fmt.Errorf("failed to mark feature '%s' complete: %w", features[i].Slug, err)
		}
		if _, err := fmt.Fprintf(out, "✅ Feature '%s' marked complete\n", features[i].Slug); err != nil {
			return err
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		_, _ = fmt.Fprintf(errOut, "  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		return printCompletionAdvisories(out, projectRoot, cfg)
	}
	if _, err := fmt.Fprintln(out, "  ✓ Updated PROJECT_PROGRESS_SUMMARY.md"); err != nil {
		return err
	}
	return printCompletionAdvisories(out, projectRoot, cfg)
}

func printCompletionAdvisories(out io.Writer, projectRoot string, cfg *config.Config) error {
	if err := printProjectRefreshAdvisory(out, projectRoot, cfg); err != nil {
		return err
	}
	_, err := fmt.Fprintln(out, "  ℹ Managed guidance: run `kit status` and follow any Kit-managed refresh action before final delivery.")
	return err
}

func markFeatureComplete(feat *feature.Feature, path string) error {
	if isLivingSpecFeature(feat) {
		return setSpecPhase(path, feat.DirName, string(feature.PhaseComplete))
	}
	return appendReflectionMarker(path)
}

func setSpecPhase(specPath, featureDirName, phase string) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	doc := document.Parse(string(data), specPath, document.TypeSpec)
	workflowVersion := document.WorkflowVersionV2
	if doc.Metadata != nil && doc.Metadata.WorkflowVersion != 0 {
		workflowVersion = doc.Metadata.WorkflowVersion
	}
	updated, changed, err := document.UpsertMetadata(string(data), document.TypeSpec, document.MetadataUpsert{
		Feature:         document.FeatureMetadataFromDir(featureDirName),
		WorkflowVersion: workflowVersion,
		Phase:           phase,
	})
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return os.WriteFile(specPath, []byte(updated), 0644)
}

func isV2Feature(feat *feature.Feature) bool {
	return workflowVersionForFeature(feat) == document.WorkflowVersionV2
}

func isLivingSpecFeature(feat *feature.Feature) bool {
	version := workflowVersionForFeature(feat)
	return version == document.WorkflowVersionV2 || version == document.WorkflowVersionV3
}

func workflowVersionForFeature(feat *feature.Feature) int {
	if feat == nil {
		return 0
	}
	specPath := filepath.Join(feat.Path, "SPEC.md")
	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil || doc.Metadata == nil {
		return 0
	}
	return doc.Metadata.WorkflowVersion
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
