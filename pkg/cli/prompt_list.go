package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/promptlib"
)

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List effective prompts from built-in, global, and local prompt libraries",
	Args:  cobra.NoArgs,
	RunE:  runPromptList,
}

func init() {
	promptCmd.AddCommand(promptListCmd)
}

func runPromptList(cmd *cobra.Command, args []string) error {
	prompts, err := loadPromptLibrary()
	if err != nil {
		return err
	}
	if len(prompts) == 0 {
		return fmt.Errorf("no prompts are available")
	}

	return printPromptList(prompts)
}

func printPromptList(prompts []promptlib.EffectivePrompt) error {
	style := styleForStdout()
	fmt.Println(style.title("📚", "Prompt Library"))

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "COMMAND\tDESCRIPTION\tORIGIN\tOVERRIDES")
	for _, prompt := range prompts {
		description := prompt.Prompt.Description
		if description == "" {
			description = "none"
		}

		overrides := prompt.ShadowSummary()
		if overrides == "" {
			overrides = "none"
		}

		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\n",
			prompt.CommandName(),
			description,
			promptLibraryOrigin(prompt),
			overrides,
		)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to render prompt list: %w", err)
	}
	return nil
}
