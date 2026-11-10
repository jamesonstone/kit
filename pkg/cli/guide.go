package cli

import (
	"github.com/jamesonstone/kit/v3/internal/guide"
	"github.com/spf13/cobra"
)

func newGuideCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "guide", Short: "Explain installed commands and repository artifact types for agents",
		Long:              "Read the complete functionality guide derived from this binary's command help.\nIncludes nested commands, flags, examples, and repository artifact instructions.\nWorks outside a Kit project, without network access, configuration changes, or usage recording.",
		Example:           "  kit guide\n  kit guide --json\n  kit rules view delivery",
		Args:              cobra.NoArgs,
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		RunE: func(cmd *cobra.Command, _ []string) error {
			d := guide.Build(cmd.Root())
			asJSON, _ := cmd.Flags().GetBool("json")
			if asJSON {
				return guide.WriteJSON(cmd.OutOrStdout(), d)
			}
			return guide.WriteText(cmd.OutOrStdout(), d)
		},
	}
	cmd.Flags().Bool("json", false, "Output the guide as deterministic JSON")
	return cmd
}

func init() { rootCmd.AddCommand(newGuideCommand()) }
