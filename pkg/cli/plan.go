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

var planCopy bool

var planOutputOnly bool

var planCmd = &cobra.Command{
	Use:   "plan [feature]",
	Short: "Deprecated v1 staged workflow: create PLAN.md",
	Long: `Deprecated v1 staged workflow: create a new implementation plan for a feature.

The default v2 feature workflow keeps the implementation plan inside SPEC.md
through the kit spec supervisor prompt. Use this command only when intentionally
working in the legacy staged artifact flow.

Creates:
  - PLAN.md with required sections and placeholder comments

Prerequisites:
  - SPEC.md must exist (unless --force)

If no feature is specified, shows an interactive selection of features
that have SPEC.md but no PLAN.md yet.

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPlan,
}

func init() {
	planCmd.Flags().Bool("force", false, "create missing SPEC.md with headers if it doesn't exist")
	planCmd.Flags().Bool("warp", false, "output prompt for Warp coding agent to fill PLAN.md from Warp plan")
	planCmd.Flags().BoolVar(&planCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	planCmd.Flags().BoolVar(&planOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(planCmd)
	legacyCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	planForce, _ := cmd.Flags().GetBool("force")
	warpMode, _ := cmd.Flags().GetBool("warp")
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
		if planForce {
			return fmt.Errorf("--prompt-only cannot be used with --force")
		}
		return runPlanPromptOnly(args, projectRoot, cfg, warpMode, outputOnly)
	}

	var feat *feature.Feature

	if len(args) == 0 {

		feat, err = selectFeatureForPlan(specsDir)
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
		fmt.Printf("📋 Creating plan for feature: %s\n", feat.DirName)
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if planForce || cfg.AllowOutOfOrder {

			content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(specPath, content); err != nil {
				return fmt.Errorf("failed to create SPEC.md: %w", err)
			}
			if !outputOnly {
				fmt.Println("  ✓ Created SPEC.md (--force)")
			}
		} else {
			return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first or use --force", feat.Slug)
		}
	}

	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(planPath, content); err != nil {
			return fmt.Errorf("failed to create PLAN.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created PLAN.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ PLAN.md already exists")
	}
	if effectivePromptProfile(feat.Path) == promptProfileFrontend {
		if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
			return err
		}
		if _, err := ensureFrontendProfileDependencyRows(planPath, document.TypePlan, feat.DirName); err != nil {
			return err
		}
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
		fmt.Printf("\n✅ Plan for '%s' ready!\n", feat.Slug)
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define the implementation approach", planPath),
			fmt.Sprintf("Run 'kit legacy tasks %s' to create executable tasks", feat.Slug),
		})
	}
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if warpMode {
		return outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
}

func runPlanPromptOnly(args []string, projectRoot string, cfg *config.Config, warpMode, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForPlanPromptOnly(specsDir)
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

	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(planPath) {
		return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first", feat.Slug)
	}

	if warpMode {
		return outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
}
