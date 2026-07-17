package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/instructions"
)

var instructionsCmd = newInstructionsCommand()

func init() {
	rootCmd.AddCommand(instructionsCmd)
}

func newInstructionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instructions",
		Short: "Print versioned coding-agent instructions",
		Long: `Print the provider-neutral instructions used directly by coding agents.

The current version is printed by default. Pass an exact vN identifier to
reproduce an earlier immutable version. Output is raw Markdown written only to
stdout; the command does not require a Kit project or use the clipboard.`,
		Args: cobra.NoArgs,
		RunE: runInstructions,
	}
	cmd.Flags().String("version", "", "print an exact instructions version, such as v1 (default: current)")
	return cmd
}

func runInstructions(cmd *cobra.Command, args []string) error {
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("version") && version == "" {
		return fmt.Errorf(
			"--version cannot be empty; available versions: %s",
			strings.Join(instructions.AgentInstructionVersions(), ", "),
		)
	}

	content, err := instructions.AgentInstructions(version)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), content)
	return err
}
