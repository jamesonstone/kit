package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dispatchCopy         bool
	dispatchCodeRabbit   bool
	dispatchEditor       string
	dispatchFile         string
	dispatchOutputOnly   bool
	dispatchPR           string
	dispatchUseVim       bool
	dispatchMaxSubagents int
)

var dispatchCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "Output a subagent dispatch dry-run prompt",
	Long: `Output a prompt that tells a coding agent how to discover file overlap
across a task set, cluster related work, and queue subagents safely.

Input precedence:
  1. --pr
  2. --file
  3. piped stdin
  4. interactive editor-backed capture
Interactive capture opens $EDITOR by default, falling back to a vim-compatible editor when $EDITOR is unset.

The command never launches subagents itself. It only outputs the prompt.`,
	Args: cobra.NoArgs,
	RunE: runDispatch,
}

func init() {
	addFreeTextInputFlags(dispatchCmd, &dispatchUseVim, &dispatchEditor)
	dispatchCmd.Flags().StringVar(&dispatchFile, "file", "", "read the raw task set from a file")
	dispatchCmd.Flags().StringVar(&dispatchPR, "pr", "", "fetch unresolved PR review threads from a PR URL, Markdown link, owner/repo#number, or current-repo number")
	dispatchCmd.Flags().BoolVar(&dispatchCodeRabbit, "coderabbit", false, "with --pr, include only CodeRabbit-authored review comments")
	dispatchCmd.Flags().BoolVar(&dispatchCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	dispatchCmd.Flags().BoolVar(
		&dispatchOutputOnly,
		"output-only",
		false,
		"output prompt text to stdout instead of copying it to the clipboard",
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

	inputCfg := newFreeTextInputConfig(dispatchUseVim, dispatchEditor, false, true)
	rawInput, inputSource, promptOptions, foundInput, err := loadDispatchInputForCommand(inputCfg)
	if err != nil {
		return err
	}
	if !foundInput {
		fmt.Println("No actionable PR review comments found.")
		return nil
	}

	tasks, err := normalizeDispatchTasks(rawInput)
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	prompt := buildDispatchPrompt(tasks, dispatchMaxSubagents, workingDirectory, inputSource, promptOptions)
	if err := outputPromptWithoutSubagentsWithClipboardDefault(prompt, outputOnly, dispatchCopy); err != nil {
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

func loadDispatchInputForCommand(
	inputCfg freeTextInputConfig,
) (string, dispatchInputSource, dispatchPromptOptions, bool, error) {
	if strings.TrimSpace(dispatchPR) != "" {
		prInput, found, err := loadDispatchPRInput(dispatchPR, dispatchCodeRabbit, inputCfg)
		return prInput.RawTasks,
			dispatchInputSourcePR,
			dispatchPromptOptions{CommonReviewInstruction: prInput.CommonReviewInstruction},
			found,
			err
	}

	rawInput, inputSource, err := loadDispatchInput(dispatchFile, inputCfg)
	return rawInput, inputSource, dispatchPromptOptions{}, true, err
}

func validateDispatchMaxSubagents(maxSubagents int) error {
	if maxSubagents < 1 {
		return fmt.Errorf("--max-subagents must be >= 1")
	}

	return nil
}
