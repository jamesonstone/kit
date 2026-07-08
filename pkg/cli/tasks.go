package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var tasksCopy bool

var tasksOutputOnly bool

var tasksCmd = &cobra.Command{
	Use:   "tasks [feature]",
	Short: "Deprecated v1 staged workflow: create TASKS.md",
	Long: `Deprecated v1 staged workflow: create a new task list for a feature.

The default v2 feature workflow keeps the durable task checklist inside
SPEC.md through the kit spec supervisor prompt. Use this command only when
intentionally working in the legacy staged artifact flow.

Creates:
  - TASKS.md with required sections, task table, and placeholder comments

Prerequisites:
  - PLAN.md must exist (unless --force)

If no feature is specified, shows an interactive selection of features
that have SPEC.md and PLAN.md but no TASKS.md yet.

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTasks,
}

func init() {
	tasksCmd.Flags().Bool("force", false, "create missing SPEC.md and PLAN.md with headers if they don't exist")
	tasksCmd.Flags().BoolVar(&tasksCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	tasksCmd.Flags().BoolVar(&tasksOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(tasksCmd)
	legacyCmd.AddCommand(tasksCmd)
}

func runTasks(cmd *cobra.Command, args []string) error {
	tasksForce, _ := cmd.Flags().GetBool("force")
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	promptOnly := promptOnlyEnabled(cmd)

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	if promptOnly {
		if tasksForce {
			return fmt.Errorf("--prompt-only cannot be used with --force")
		}
		return runTasksPromptOnly(args, projectRoot, cfg, outputOnly)
	}

	var feat *feature.Feature

	if len(args) == 0 {

		feat, err = selectFeatureForTasks(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {

		featureRef := args[0]
		feat, err = loadFeatureWithState(specsDir, cfg, featureRef)
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
		}
	}

	if !outputOnly {
		fmt.Printf("📝 Creating tasks for feature: %s\n", feat.DirName)
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")

	if !document.Exists(planPath) {
		if tasksForce || cfg.AllowOutOfOrder {

			if !document.Exists(specPath) {
				content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
				if err := document.Write(specPath, content); err != nil {
					return fmt.Errorf("failed to create SPEC.md: %w", err)
				}
				if !outputOnly {
					fmt.Println("  ✓ Created SPEC.md (--force)")
				}
			}

			content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(planPath, content); err != nil {
				return fmt.Errorf("failed to create PLAN.md: %w", err)
			}
			if !outputOnly {
				fmt.Println("  ✓ Created PLAN.md (--force)")
			}
		} else {
			return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first or use --force", feat.Slug)
		}
	}

	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if !document.Exists(tasksPath) {
		content := templates.BuildTasksArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(tasksPath, content); err != nil {
			return fmt.Errorf("failed to create TASKS.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created TASKS.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ TASKS.md already exists")
	}

	wasPaused := feat.Paused
	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		if !outputOnly {
			fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		}
	} else if !outputOnly {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	if !outputOnly {
		fmt.Printf("\n✅ Tasks for '%s' ready!\n", feat.Slug)
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define atomic tasks", tasksPath),
			"Link tasks to plan items using [PLAN-XX] syntax",
			"Begin implementation",
		})
	}

	return outputTasksPrompt(feat, projectRoot, cfg, outputOnly)
}

func runTasksPromptOnly(args []string, projectRoot string, cfg *config.Config, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForTasksPromptOnly(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", args[0], args[0])
		}
	}

	if !document.Exists(filepath.Join(feat.Path, "SPEC.md")) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(filepath.Join(feat.Path, "PLAN.md")) {
		return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first", feat.Slug)
	}
	if !document.Exists(filepath.Join(feat.Path, "TASKS.md")) {
		return fmt.Errorf("TASKS.md not found. Run 'kit legacy tasks %s' first", feat.Slug)
	}

	return outputTasksPrompt(feat, projectRoot, cfg, outputOnly)
}

func outputTasksPrompt(feat *feature.Feature, projectRoot string, cfg *config.Config, outputOnly bool) error {
	prompt := buildTasksPrompt(feat, projectRoot, cfg)

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, tasksCopy); err != nil {
		return err
	}

	return nil
}
