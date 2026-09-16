package cli

import (
	"encoding/json"
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/threadtitle"
	"github.com/spf13/cobra"
)

func newInstructionsTitleCommand() *cobra.Command {
	var in threadtitle.Input
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "title",
		Short: "Suggest a conversation title from agent-resolved ownership",
		Long: `Format [scope] domain / objective from semantic fields resolved by the active agent.
Reuse the repository, project, or program scope already known to that agent.
This read-only helper does not infer intent, inspect a project, or rename a session.

Default output is exactly one title when a change is needed, or empty stdout
when the current title should be kept. --json always prints title, changed, and
reason. --accurate preserves a nonempty current title when the active agent judges
that it still describes the conversation's ownership. Cosmetic changes are suppressed.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			decision, err := threadtitle.Resolve(in)
			if err != nil {
				return err
			}
			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(decision)
			}
			if !decision.Changed {
				return nil
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), decision.Title)
			return err
		},
	}
	cmd.Flags().StringVar(&in.Scope, "scope", "", "existing repository, project, or program scope (required)")
	cmd.Flags().StringVar(&in.Domain, "domain", "", "stable conceptual domain resolved by the agent (required)")
	cmd.Flags().StringVar(&in.Objective, "objective", "", "current owned outcome resolved by the agent (required)")
	cmd.Flags().StringVar(&in.Current, "current", "", "current conversation title")
	cmd.Flags().BoolVar(&in.Accurate, "accurate", false, "current title still accurately describes ownership")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print a structured decision, including unchanged titles")
	return cmd
}
