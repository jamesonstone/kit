package agentcli

import (
	"fmt"
	"os"

	"github.com/jamesonstone/kit/internal/bootstrap"
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

type initOptions struct {
	dryRun      bool
	json        bool
	copy        bool
	outputOnly  bool
	repo        string
	branch      string
	sourcePath  string
	catalogPath string
}

func newInitCommand() *cobra.Command {
	options := &initOptions{}
	command := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a repository for coding agents",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runInit(command, *options)
		},
	}
	command.Flags().BoolVar(&options.dryRun, "dry-run", false, "validate and preview without writing")
	command.Flags().BoolVar(&options.json, "json", false, "emit the write-free initialization plan as JSON")
	command.Flags().BoolVar(&options.copy, "copy", false, "copy the bootstrap prompt, including with --output-only")
	command.Flags().BoolVar(&options.outputOnly, "output-only", false, "emit only the raw bootstrap prompt")
	command.Flags().StringVar(&options.repo, "registry-repo", "", "GitHub owner/repository containing the registry")
	command.Flags().StringVar(&options.branch, "registry-branch", "", "registry branch or ref")
	command.Flags().StringVar(&options.sourcePath, "registry-path", "", "local registry root for offline development")
	command.Flags().StringVar(&options.catalogPath, "catalog", "", "catalog path within the registry")
	return command
}

func runInit(command *cobra.Command, options initOptions) error {
	if options.json && (options.outputOnly || options.copy) {
		return fmt.Errorf("--json cannot be combined with --output-only or --copy")
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	userDefaults, userPlan, err := bootstrap.PlanUserConfig()
	if err != nil {
		return err
	}
	sourceConfig := effectiveInitSource(command, options, userDefaults.Source())
	plan, err := bootstrap.BuildPlan(
		command.Context(), root, sourceFor(root, sourceConfig), sourceConfig, userDefaults, userPlan,
	)
	if err != nil {
		return err
	}
	applied := false
	if !options.dryRun && !options.json {
		if err := bootstrap.Apply(plan); err != nil {
			return err
		}
		applied = true
	}
	if options.json {
		return writeJSON(command.OutOrStdout(), struct {
			Applied bool           `json:"applied"`
			Plan    bootstrap.Plan `json:"plan"`
		}{Plan: plan})
	}
	if options.dryRun {
		if _, err := fmt.Fprint(command.OutOrStdout(), registry.RenderDiff(plan.Registry)); err != nil {
			return err
		}
		return writeBootstrapHuman(command, plan, false)
	}
	if options.outputOnly {
		if options.copy {
			if err := clipboardCopyFunc(plan.Prompt); err != nil {
				return fmt.Errorf("copy bootstrap prompt: %w", err)
			}
		}
		_, err := fmt.Fprint(command.OutOrStdout(), plan.Prompt)
		return err
	}
	if err := writeBootstrapHuman(command, plan, applied); err != nil {
		return err
	}
	shouldCopy := options.copy || userDefaults.CopyPrompt()
	if shouldCopy {
		if err := clipboardCopyFunc(plan.Prompt); err != nil {
			return fmt.Errorf("copy bootstrap prompt: %w", err)
		}
		_, err = fmt.Fprintln(command.OutOrStdout(), "bootstrap prompt copied to clipboard")
		return err
	}
	_, err = fmt.Fprint(command.OutOrStdout(), plan.Prompt)
	return err
}

func effectiveInitSource(command *cobra.Command, options initOptions, defaults registry.SourceConfig) registry.SourceConfig {
	result := defaults
	if command.Flags().Changed("registry-repo") {
		result.Repo = options.repo
		result.Path = ""
	}
	if command.Flags().Changed("registry-branch") {
		result.Branch = options.branch
	}
	if command.Flags().Changed("registry-path") {
		result.Path = options.sourcePath
		if options.sourcePath != "" {
			result.Repo = ""
		}
	}
	if command.Flags().Changed("catalog") {
		result.CatalogPath = options.catalogPath
	}
	return result
}

func writeBootstrapHuman(command *cobra.Command, plan bootstrap.Plan, applied bool) error {
	verb := "planned"
	if applied {
		verb = "applied"
	}
	if _, err := fmt.Fprintf(command.OutOrStdout(), "%s bootstrap: %s (%d file disposition(s))\n", verb, plan.State, len(plan.Files)); err != nil {
		return err
	}
	for _, file := range plan.Files {
		if _, err := fmt.Fprintf(command.OutOrStdout(), "  %-42s %-10s %s\n", file.Path, file.Action, file.Strategy); err != nil {
			return err
		}
	}
	for _, diagnostic := range plan.Diagnostics {
		if _, err := fmt.Fprintln(command.OutOrStdout(), "  attention:", diagnostic); err != nil {
			return err
		}
	}
	return nil
}
