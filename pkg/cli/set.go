package cli

import "github.com/spf13/cobra"

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Configure prompt-library resources",
	Long: `Configure Kit resources.

In v0, prompt is the only configurable resource, so running kit set starts
the same prompt-setting flow as kit set prompt.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetPromptWithOptions(nil, setPromptLocal, setPromptGlobal)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
