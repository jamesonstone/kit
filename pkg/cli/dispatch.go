package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dispatchCopy         bool
	dispatchEditor       string
	dispatchFile         string
	dispatchOutputOnly   bool
	dispatchUseVim       bool
	dispatchMaxSubagents int
)

var dispatchCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "Output a subagent dispatch-planning prompt",
	Long: `Output a prompt that tells a coding agent how to discover file overlap
across a task set, cluster related work, and queue subagents safely.

Input precedence:
  1. --file
  2. piped stdin
  3. interactive editor-backed capture
Interactive capture opens a vim-compatible editor by default.

The command never launches subagents itself. It only outputs the prompt.`,
	Args: cobra.NoArgs,
	RunE: runDispatch,
}

func init() {
	addFreeTextInputFlags(dispatchCmd, &dispatchUseVim, &dispatchEditor)
	dispatchCmd.Flags().StringVar(&dispatchFile, "file", "", "read the raw task set from a file")
	dispatchCmd.Flags().BoolVar(&dispatchCopy, "copy", false, "copy agent prompt to clipboard")
	dispatchCmd.Flags().BoolVar(
		&dispatchOutputOnly,
		"output-only",
		false,
		"output prompt only, suppressing status messages",
	)
	dispatchCmd.Flags().IntVar(
		&dispatchMaxSubagents,
		"max-subagents",
		10,
		"maximum concurrent subagents allowed in the generated prompt",
	)
	rootCmd.AddCommand(dispatchCmd)
}

func runDispatch(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")

	if err := validateDispatchMaxSubagents(dispatchMaxSubagents); err != nil {
		return err
	}

	rawInput, inputSource, err := loadDispatchInput(
		dispatchFile,
		newFreeTextInputConfig(true, dispatchEditor),
	)
	if err != nil {
		return err
	}

	tasks, err := normalizeDispatchTasks(rawInput)
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	prompt := buildDispatchPrompt(tasks, dispatchMaxSubagents, workingDirectory, inputSource)
	if err := outputPromptWithoutSubagents(prompt, outputOnly, dispatchCopy); err != nil {
		return err
	}

	if !outputOnly {
		printWorkflowInstructions("dispatch (supporting step)", []string{
			"review the discovery report and overlap clusters first",
			"approve subagent launch only after the queue looks safe",
		})
	}

	return nil
}

func validateDispatchMaxSubagents(maxSubagents int) error {
	if maxSubagents < 1 {
		return fmt.Errorf("--max-subagents must be >= 1")
	}

	return nil
}
