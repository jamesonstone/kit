package cli

import (
	"github.com/spf13/cobra"
)

var legacyCmd = &cobra.Command{
	Use:   "legacy",
	Short: "List deprecated v1 staged workflow commands",
	Long: `List the deprecated v1 staged workflow command surface.

The default v2 feature workflow is ` + "`kit spec <feature>`" + `. These commands remain
available for finishing existing staged work, but they are no longer the
primary feature workflow.

Deprecated v1 staged commands:
  kit legacy brainstorm [feature]  Create BRAINSTORM.md or capture backlog research
  kit legacy plan <feature>        Create PLAN.md from a legacy staged SPEC.md
  kit legacy tasks <feature>       Create TASKS.md from PLAN.md
  kit legacy implement <feature>   Output legacy implementation context from TASKS.md
  kit legacy reflect <feature>     Output legacy reflection instructions
  kit legacy verify [feature]      Run TASKS.md verification declarations

Use ` + "`kit legacy <command> --help`" + ` for command flags during migration.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(legacyCmd)
}
