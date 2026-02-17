package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/templates"
)

var (
	brainstormCopy       bool
	brainstormOutput     string
	brainstormOutputOnly bool
)

var brainstormCmd = &cobra.Command{
	Use:   "brainstorm [topic]",
	Short: "Generate a brainstorming scaffold document",
	Long: `Generate a markdown scaffold to help narrow down thinking for ideas,
decisions, or situations.

The output is designed for immediate context injection into coding agents
or LLM tools.

Examples:
  kit brainstorm "API redesign"
  kit brainstorm "Database migration" --copy
  kit brainstorm "Auth refactor" -o docs/brainstorm-auth.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrainstorm,
}

func init() {
	brainstormCmd.Flags().BoolVar(&brainstormCopy, "copy", false, "copy output to clipboard")
	brainstormCmd.Flags().StringVarP(&brainstormOutput, "output", "o", "", "write output to file")
	brainstormCmd.Flags().BoolVar(&brainstormOutputOnly, "output-only", false, "output text only, suppressing status messages")
	rootCmd.AddCommand(brainstormCmd)
}

func runBrainstorm(cmd *cobra.Command, args []string) error {
	topic := "[Topic]"
	if len(args) == 1 {
		topic = args[0]
	}

	content := templates.Brainstorm(topic)

	printWorkflowInstructions("brainstorm (pre-spec)", []string{
		"run kit spec <feature> to create SPEC.md",
		"then follow spec -> plan -> tasks -> implement -> reflect",
	})

	// write to file if specified
	if brainstormOutput != "" {
		if err := os.WriteFile(brainstormOutput, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("✓ Written to %s\n", brainstormOutput)
	}

	// copy to clipboard if requested
	if brainstormCopy {
		if err := copyToClipboard(content); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied to clipboard")
		return nil
	}

	// print to stdout when not copying
	if brainstormOutput == "" {
		fmt.Print(content)
	}

	return nil
}

// copyToClipboard copies text to the system clipboard.
func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
