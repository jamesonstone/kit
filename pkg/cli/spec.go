package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var specCmd = &cobra.Command{
	Use:   "spec [feature]",
	Short: "Create or adopt a living V3 feature specification",
	Long: `Create or adopt repository-backed feature memory.

Native agent planning owns research, clarification, design, and implementation
planning. Kit creates the durable place where the accepted plan, material
decisions, discoveries, validation, outcome, and repository-memory curation
survive after the agent session.

New specifications use the V3 contract. Existing V1 and V2 specifications
remain readable and are never mechanically rewritten.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNativePlanSpec,
}

func init() {
	rootCmd.AddCommand(specCmd)
}

func runNativePlanSpec(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("feature name required: use `kit spec <feature>`")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("kit project not initialized: run `kit init` before `kit spec`: %w", err)
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	specsDir := cfg.SpecsPath(projectRoot)
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		return err
	}

	feat, created, err := feature.EnsureExists(cfg, projectRoot, specsDir, args[0])
	if err != nil {
		return err
	}
	feature.ApplyLifecycleState(feat, cfg)

	specPath := filepath.Join(feat.Path, "SPEC.md")
	specCreated := false
	if !document.Exists(specPath) {
		content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(specPath, content); err != nil {
			return fmt.Errorf("create SPEC.md: %w", err)
		}
		specCreated = true
	}

	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil {
		return err
	}
	workflowVersion := 0
	if doc.Metadata != nil {
		workflowVersion = doc.Metadata.WorkflowVersion
	}
	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("update PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	action := "adopted"
	if created || specCreated {
		action = "created"
	}
	out := cmd.OutOrStdout()
	relativePath, err := filepath.Rel(projectRoot, specPath)
	if err != nil {
		relativePath = specPath
	}
	if _, err := fmt.Fprintf(out, "Specification %s: %s\n", action, filepath.ToSlash(relativePath)); err != nil {
		return err
	}
	if workflowVersion != document.WorkflowVersionV3 {
		if _, err := fmt.Fprintf(out, "Compatibility: preserved workflow_version %d without rewriting it. Migration requires semantic curation.\n", workflowVersion); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(out, "Next: run `kit context resolve --feature %s --json`, use native agent planning, and keep material decisions, validation, and outcome current in SPEC.md.\n", feat.DirName); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Managed guidance: run `kit status` and follow any Kit-managed refresh action before implementation.")
	return err
}
