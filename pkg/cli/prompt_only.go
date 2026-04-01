package cli

import "github.com/spf13/cobra"

const promptOnlyFlagUsage = "regenerate the prompt for an existing feature without mutating repository docs"

func addPromptOnlyFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("prompt-only", false, promptOnlyFlagUsage)
}

func promptOnlyEnabled(cmd *cobra.Command) bool {
	enabled, _ := cmd.Flags().GetBool("prompt-only")
	return enabled
}
