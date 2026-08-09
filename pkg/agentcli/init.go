package agentcli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

type initOptions struct {
	dryRun      bool
	json        bool
	repo        string
	branch      string
	sourcePath  string
	catalogPath string
}

func newInitCommand() *cobra.Command {
	options := &initOptions{}
	command := &cobra.Command{
		Use:   "init",
		Short: "Materialize the coding-agent contract in a repository",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runInit(command, *options)
		},
	}
	command.Flags().BoolVar(&options.dryRun, "dry-run", false, "validate and preview without writing")
	command.Flags().BoolVar(&options.json, "json", false, "emit the initialization plan as JSON")
	command.Flags().StringVar(&options.repo, "registry-repo", "jamesonstone/kit", "GitHub owner/repository containing the registry")
	command.Flags().StringVar(&options.branch, "registry-branch", "main", "registry branch or ref")
	command.Flags().StringVar(&options.sourcePath, "registry-path", "", "local registry root for offline development")
	command.Flags().StringVar(&options.catalogPath, "catalog", "registry/catalog.yaml", "catalog path within the registry")
	return command
}

func runInit(command *cobra.Command, options initOptions) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(filepath.Join(root, registry.ProjectFile)); statErr == nil {
		cfg, _, loadErr := registry.LoadProject(root)
		if loadErr != nil {
			return loadErr
		}
		plan, planErr := registry.BuildReconcilePlan(command.Context(), root, sourceFor(root, cfg.Registry.Source), nil)
		if planErr != nil {
			return planErr
		}
		if plan.State != "current" {
			return fmt.Errorf("project is already initialized and has contract drift; use `kit reconcile` to preview it")
		}
		if options.json {
			return writeJSON(command.OutOrStdout(), struct {
				Applied bool          `json:"applied"`
				Plan    registry.Plan `json:"plan"`
			}{Plan: plan})
		}
		return writePlanHuman(command.OutOrStdout(), plan, false)
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	sourceConfig := registry.SourceConfig{
		Repo: options.repo, Branch: options.branch, Path: options.sourcePath, CatalogPath: options.catalogPath,
	}
	plan, err := registry.BuildInitPlan(command.Context(), root, sourceFor(root, sourceConfig), sourceConfig)
	if err != nil {
		return err
	}
	applied := false
	if !options.dryRun {
		if err := registry.ApplyPlan(root, plan); err != nil {
			return err
		}
		applied = true
	}
	if options.json {
		return writeJSON(command.OutOrStdout(), struct {
			Applied bool          `json:"applied"`
			Plan    registry.Plan `json:"plan"`
		}{Applied: applied, Plan: plan})
	}
	if options.dryRun {
		if _, err := fmt.Fprint(command.OutOrStdout(), registry.RenderDiff(plan)); err != nil {
			return err
		}
	}
	return writePlanHuman(command.OutOrStdout(), plan, applied)
}
