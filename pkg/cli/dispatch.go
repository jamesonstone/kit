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
	dispatchLoop         bool
	dispatchOutputOnly   bool
	dispatchPR           string
	dispatchResolve      bool
	dispatchUseVim       bool
	dispatchWatch        bool
	dispatchYes          bool
	dispatchMaxSubagents int
)

const (
	defaultDispatchMaxSubagents = 3
	hardDispatchMaxSubagents    = 4
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

The command never launches subagents itself. It only outputs the prompt unless
--resolve --yes is explicitly supplied to resolve already-handled PR review
threads.`,
	Args: cobra.NoArgs,
	RunE: runDispatch,
}

func init() {
	addFreeTextInputFlags(dispatchCmd, &dispatchUseVim, &dispatchEditor)
	dispatchCmd.Flags().StringVar(&dispatchFile, "file", "", "read the raw task set from a file")
	dispatchCmd.Flags().StringVar(&dispatchPR, "pr", "", "fetch unresolved PR review threads from a PR URL, Markdown link, owner/repo#number, or current-repo number")
	dispatchCmd.Flags().BoolVar(&dispatchCodeRabbit, "coderabbit", false, "with --pr, include only CodeRabbit-authored review comments")
	dispatchCmd.Flags().BoolVar(&dispatchLoop, "loop", false, "route PR review feedback through the review-loop workflow")
	dispatchCmd.Flags().BoolVar(&dispatchResolve, "resolve", false, "with --pr, resolve matching unresolved review threads after fixes or no-op decisions are complete")
	dispatchCmd.Flags().BoolVar(&dispatchWatch, "watch", false, "with --loop, wait for current-head CodeRabbit review completion before collecting feedback")
	dispatchCmd.Flags().BoolVar(&dispatchYes, "yes", false, "confirm --resolve without an interactive prompt")
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
		defaultDispatchMaxSubagents,
		"maximum concurrent subagents allowed in the generated prompt; default 3, hard ceiling 4",
	)
	rootCmd.AddCommand(dispatchCmd)
}

func runDispatch(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	if dispatchWatch && !dispatchLoop {
		return fmt.Errorf("--watch requires --loop")
	}
	if dispatchYes && !dispatchResolve {
		return fmt.Errorf("--yes requires --resolve")
	}
	if dispatchLoop {
		if dispatchResolve {
			return fmt.Errorf("--resolve cannot be used with --loop")
		}
		return runDispatchReviewLoopAlias(cmd, outputOnly)
	}
	if dispatchResolve {
		return runDispatchPRResolve(cmd)
	}

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

func runDispatchReviewLoopAlias(cmd *cobra.Command, outputOnly bool) error {
	if strings.TrimSpace(dispatchFile) != "" {
		return fmt.Errorf("--file cannot be used with --loop")
	}
	if strings.TrimSpace(dispatchPR) == "" {
		return fmt.Errorf("--loop requires --pr")
	}

	return reviewLoopExecutor(cmd, reviewLoopOptions{
		PRRef:          dispatchPR,
		CodeRabbitOnly: dispatchCodeRabbit,
		Watch:          dispatchWatch,
		Copy:           dispatchCopy,
		OutputOnly:     outputOnly,
		UseVim:         dispatchUseVim,
		Editor:         dispatchEditor,
		MaxSubagents:   dispatchMaxSubagents,
		InputConfig:    newFreeTextInputConfig(dispatchUseVim, dispatchEditor, false, true),
	})
}

func loadDispatchInputForCommand(
	inputCfg freeTextInputConfig,
) (string, dispatchInputSource, dispatchPromptOptions, bool, error) {
	if strings.TrimSpace(dispatchPR) != "" {
		prInput, found, err := loadDispatchPRInput(dispatchPR, dispatchCodeRabbit, inputCfg)
		return prInput.RawTasks,
			dispatchInputSourcePR,
			dispatchPromptOptions{
				CodeRabbitOnly:          dispatchCodeRabbit,
				CommonReviewInstruction: prInput.CommonReviewInstruction,
				PRTarget:                dispatchPR,
			},
			found,
			err
	}

	rawInput, inputSource, err := loadDispatchInput(dispatchFile, inputCfg)
	return rawInput, inputSource, dispatchPromptOptions{}, true, err
}

func validateDispatchMaxSubagents(maxSubagents int) error {
	if maxSubagents < 1 {
		return fmt.Errorf("--max-subagents must be between 1 and %d", hardDispatchMaxSubagents)
	}
	if maxSubagents > hardDispatchMaxSubagents {
		return fmt.Errorf("--max-subagents must be between 1 and %d", hardDispatchMaxSubagents)
	}

	return nil
}
