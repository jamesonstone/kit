package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

const scaffoldPrepareMessage = "♻️ %s directory and files empty scaffolding created. Please prepare your notes, documents, images, and examples for the %s phase\n"

type scaffoldResult struct {
	Feature feature.Feature
	Paths   []string
}

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Create empty workflow document structures and supporting directories",
	Long: `Create empty document structures and supporting directories for a Kit workflow.

Scaffold commands prepare files only. They do not emit workflow prompts and do
not ask an agent to start the phase.`,
}

var scaffoldBrainstormCmd = &cobra.Command{
	Use:   "brainstorm <feature>",
	Short: "Prepare brainstorm directories and files without starting the brainstorm prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := scaffoldBrainstormWorkflow(args[0])
		if err != nil {
			return err
		}
		return printScaffoldWorkflowResult(cmd.OutOrStdout(), "brainstorm", result)
	},
}

var scaffoldSpecCmd = &cobra.Command{
	Use:   "spec <feature>",
	Short: "Create an empty SPEC.md scaffold without outputting the spec prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := scaffoldSpecWorkflow(args[0])
		if err != nil {
			return err
		}
		return printScaffoldWorkflowResult(cmd.OutOrStdout(), "spec", result)
	},
}

var scaffoldPlanCmd = &cobra.Command{
	Use:   "plan <feature>",
	Short: "Create an empty PLAN.md scaffold without outputting the plan prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := scaffoldPlanWorkflow(args[0])
		if err != nil {
			return err
		}
		return printScaffoldWorkflowResult(cmd.OutOrStdout(), "plan", result)
	},
}

var scaffoldTasksCmd = &cobra.Command{
	Use:   "tasks <feature>",
	Short: "Create an empty TASKS.md scaffold without outputting the tasks prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := scaffoldTasksWorkflow(args[0])
		if err != nil {
			return err
		}
		return printScaffoldWorkflowResult(cmd.OutOrStdout(), "tasks", result)
	},
}

func init() {
	scaffoldCmd.AddCommand(scaffoldBrainstormCmd)
	scaffoldCmd.AddCommand(scaffoldSpecCmd)
	scaffoldCmd.AddCommand(scaffoldPlanCmd)
	scaffoldCmd.AddCommand(scaffoldTasksCmd)
	rootCmd.AddCommand(scaffoldCmd)
}

func scaffoldBrainstormWorkflow(featureRef string) (scaffoldResult, error) {
	projectRoot, cfg, specsDir, err := scaffoldWorkflowContext()
	if err != nil {
		return scaffoldResult{}, err
	}

	feat, _, err := feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
	if err != nil {
		return scaffoldResult{}, err
	}
	feature.ApplyLifecycleState(feat, cfg)

	notesPath, notesRelPath, err := ensureFeatureNotesDir(projectRoot, feat.DirName)
	if err != nil {
		return scaffoldResult{}, err
	}

	paths := []string{feat.Path, notesPath}
	frontendProfileActive := effectivePromptProfile(feat.Path) == promptProfileFrontend
	if frontendProfileActive {
		designPath, _, err := ensureFeatureDesignMaterialsDirs(projectRoot, feat.DirName)
		if err != nil {
			return scaffoldResult{}, err
		}
		paths = append(paths, designPath)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !document.Exists(brainstormPath) {
		content := templates.BuildBrainstormArtifactForFeature(
			"",
			document.FeatureMetadataFromDir(feat.DirName),
			[]document.MetadataDependency{featureNotesDependency(notesRelPath)},
		)
		if frontendProfileActive {
			content = seedFrontendProfileDependencyRows(content, document.TypeBrainstorm, feat.DirName)
		}
		if err := document.Write(brainstormPath, content); err != nil {
			return scaffoldResult{}, fmt.Errorf("failed to create BRAINSTORM.md: %w", err)
		}
	} else {
		if _, err := ensureBrainstormNotesDependency(brainstormPath, notesRelPath); err != nil {
			return scaffoldResult{}, err
		}
		if frontendProfileActive {
			if _, err := ensureFrontendProfileDependencyRows(brainstormPath, document.TypeBrainstorm, feat.DirName); err != nil {
				return scaffoldResult{}, err
			}
		}
	}
	paths = append(paths, brainstormPath)

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return scaffoldResult{}, fmt.Errorf("failed to update PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	return scaffoldResult{Feature: *feat, Paths: paths}, nil
}

func scaffoldSpecWorkflow(featureRef string) (scaffoldResult, error) {
	projectRoot, cfg, specsDir, err := scaffoldWorkflowContext()
	if err != nil {
		return scaffoldResult{}, err
	}

	feat, _, err := feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
	if err != nil {
		return scaffoldResult{}, err
	}
	feature.ApplyLifecycleState(feat, cfg)

	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(specPath, content); err != nil {
			return scaffoldResult{}, fmt.Errorf("failed to create SPEC.md: %w", err)
		}
	}
	if effectivePromptProfile(feat.Path) == promptProfileFrontend {
		if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
			return scaffoldResult{}, err
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return scaffoldResult{}, fmt.Errorf("failed to update PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	return scaffoldResult{Feature: *feat, Paths: []string{feat.Path, specPath}}, nil
}

func scaffoldPlanWorkflow(featureRef string) (scaffoldResult, error) {
	projectRoot, cfg, specsDir, err := scaffoldWorkflowContext()
	if err != nil {
		return scaffoldResult{}, err
	}

	feat, err := loadFeatureWithState(specsDir, cfg, featureRef)
	if err != nil {
		return scaffoldResult{}, fmt.Errorf("feature '%s' not found. Run 'kit scaffold spec %s' first", featureRef, featureRef)
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		return scaffoldResult{}, fmt.Errorf("SPEC.md not found. Run 'kit scaffold spec %s' first", feat.Slug)
	}

	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(planPath, content); err != nil {
			return scaffoldResult{}, fmt.Errorf("failed to create PLAN.md: %w", err)
		}
	}
	if effectivePromptProfile(feat.Path) == promptProfileFrontend {
		if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
			return scaffoldResult{}, err
		}
		if _, err := ensureFrontendProfileDependencyRows(planPath, document.TypePlan, feat.DirName); err != nil {
			return scaffoldResult{}, err
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return scaffoldResult{}, fmt.Errorf("failed to update PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	return scaffoldResult{Feature: *feat, Paths: []string{feat.Path, planPath}}, nil
}

func scaffoldTasksWorkflow(featureRef string) (scaffoldResult, error) {
	projectRoot, cfg, specsDir, err := scaffoldWorkflowContext()
	if err != nil {
		return scaffoldResult{}, err
	}

	feat, err := loadFeatureWithState(specsDir, cfg, featureRef)
	if err != nil {
		return scaffoldResult{}, fmt.Errorf("feature '%s' not found. Run 'kit scaffold spec %s' first", featureRef, featureRef)
	}

	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		return scaffoldResult{}, fmt.Errorf("PLAN.md not found. Run 'kit scaffold plan %s' first", feat.Slug)
	}

	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if !document.Exists(tasksPath) {
		content := templates.BuildTasksArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(tasksPath, content); err != nil {
			return scaffoldResult{}, fmt.Errorf("failed to create TASKS.md: %w", err)
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return scaffoldResult{}, fmt.Errorf("failed to update PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	return scaffoldResult{Feature: *feat, Paths: []string{feat.Path, tasksPath}}, nil
}

func scaffoldWorkflowContext() (string, *config.Config, string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return "", nil, "", err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", nil, "", err
	}

	specsDir := cfg.SpecsPath(projectRoot)
	if err := ensureDir(specsDir); err != nil {
		return "", nil, "", err
	}

	return projectRoot, cfg, specsDir, nil
}

func printScaffoldWorkflowResult(out io.Writer, workflow string, result scaffoldResult) error {
	if _, err := fmt.Fprintf(out, scaffoldPrepareMessage, workflow, workflow); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Feature: %s\n", result.Feature.DirName); err != nil {
		return err
	}
	for _, path := range result.Paths {
		if _, err := fmt.Fprintf(out, "  ✓ %s\n", path); err != nil {
			return err
		}
	}
	return nil
}
