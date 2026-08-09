package agentcli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

func newRegistryStatusCommand() *cobra.Command {
	jsonOutput := false
	command := &cobra.Command{
		Use: "status", Short: "Show registry freshness and contract drift", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			root, cfg, source, err := projectRegistryContext()
			if err != nil {
				return err
			}
			plan, err := registry.BuildReconcilePlan(command.Context(), root, source, nil)
			if err != nil {
				return err
			}
			if jsonOutput {
				if err := writeJSON(command.OutOrStdout(), plan); err != nil {
					return err
				}
			} else {
				sourceName := cfg.Registry.Source.Repo
				if cfg.Registry.Source.Path != "" {
					sourceName = cfg.Registry.Source.Path
				}
				if _, err := fmt.Fprintf(command.OutOrStdout(), "%s (%d change(s), %d diagnostic(s), source %s@%s)\n",
					plan.State, len(plan.Changes), len(plan.Diagnostics), sourceName, plan.Revision); err != nil {
					return err
				}
				for _, diagnostic := range plan.Diagnostics {
					if _, err := fmt.Fprintln(command.OutOrStdout(), "attention:", diagnostic); err != nil {
						return err
					}
				}
				for _, action := range plan.NextActions {
					if _, err := fmt.Fprintln(command.OutOrStdout(), "next:", action); err != nil {
						return err
					}
				}
			}
			if plan.State == "attention-needed" {
				return &exitError{code: 2, err: fmt.Errorf("registry attention is required")}
			}
			return nil
		},
	}
	command.Flags().BoolVar(&jsonOutput, "json", false, "emit status as JSON")
	return command
}
