package agentcli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/contract"
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

type contractOptions struct {
	workType      string
	feature       string
	paths         []string
	applicability []string
	workflows     []string
	json          bool
}

func newContractCommand() *cobra.Command {
	command := &cobra.Command{Use: "contract", Short: "Resolve the local coding-agent contract"}
	command.AddCommand(newContractResolveCommand())
	return command
}

func newContractResolveCommand() *cobra.Command {
	options := &contractOptions{}
	command := &cobra.Command{
		Use:   "resolve",
		Short: "Emit the deterministic effective contract as JSON",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runContractResolve(command, *options)
		},
	}
	command.Flags().StringVar(&options.workType, "work-type", "", "explicit work classification: feature or maintenance")
	command.Flags().StringVar(&options.feature, "feature", "", "canonical living feature directory")
	command.Flags().StringArrayVar(&options.paths, "path", nil, "task path hint; repeat as needed")
	command.Flags().StringArrayVar(&options.applicability, "applies-to", nil, "applicability tag; repeat as needed")
	command.Flags().StringArrayVar(&options.workflows, "workflow", nil, "explicit workflow slug; repeat as needed")
	command.Flags().BoolVar(&options.json, "json", true, "emit JSON (the stable default)")
	return command
}

func runContractResolve(command *cobra.Command, options contractOptions) error {
	root, err := registry.FindProjectRoot(mustGetwd())
	if err != nil {
		return err
	}
	resolved, err := contract.Resolve(root, contract.Hints{
		WorkType: options.workType, Feature: options.feature, Paths: options.paths,
		Applicability: options.applicability, Workflows: options.workflows,
	})
	if err != nil {
		return err
	}
	if err := writeJSON(command.OutOrStdout(), resolved); err != nil {
		return err
	}
	if resolved.State == "blocked" {
		return &exitError{code: 2, err: fmt.Errorf("coding-agent contract is blocked")}
	}
	return nil
}
