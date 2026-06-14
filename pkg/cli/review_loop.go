package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	reviewLoopExecutor        = runReviewLoop
	reviewLoopLoadReviewTasks = loadDispatchPRReviewTasks
)

func init() {
	rootCmd.AddCommand(newReviewLoopCommand())
}

func newReviewLoopCommand() *cobra.Command {
	opts := reviewLoopOptions{MaxSubagents: 10}
	cmd := &cobra.Command{
		Use:   "review-loop --pr <url|markdown-link|owner/repo#number|number>",
		Short: "Prepare a dispatch prompt from current PR review feedback",
		Long: `Prepare a human-reviewed dispatch prompt from current unresolved PR review
threads, optionally waiting for CodeRabbit to finish reviewing the current PR
head before collecting feedback.

The command reads GitHub through gh, opens the editor only when actionable
findings remain, and does not stage, commit, push, resolve review threads, or
post PR comments.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InputConfig = newFreeTextInputConfig(opts.UseVim, opts.Editor, false, true)
			return reviewLoopExecutor(cmd, opts)
		},
	}

	addReviewLoopFlags(cmd, &opts)
	return cmd
}

func addReviewLoopFlags(cmd *cobra.Command, opts *reviewLoopOptions) {
	addFreeTextInputFlags(cmd, &opts.UseVim, &opts.Editor)
	cmd.Flags().StringVar(&opts.PRRef, "pr", "", "GitHub PR URL, Markdown link, owner/repo#number, or current-repo number")
	cmd.Flags().BoolVar(&opts.CodeRabbitOnly, "coderabbit", false, "include only CodeRabbit-authored review comments")
	cmd.Flags().BoolVar(&opts.Watch, "watch", false, "wait for current-head CodeRabbit review completion before collecting feedback")
	cmd.Flags().BoolVar(&opts.Copy, "copy", false, "copy generated prompt to clipboard even with --output-only")
	cmd.Flags().BoolVar(&opts.OutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	cmd.Flags().IntVar(&opts.MaxSubagents, "max-subagents", 10, "maximum concurrent subagents allowed in the generated prompt")
}

func runReviewLoop(cmd *cobra.Command, opts reviewLoopOptions) error {
	if strings.TrimSpace(opts.PRRef) == "" {
		return fmt.Errorf("--pr is required")
	}
	if err := validateDispatchMaxSubagents(opts.MaxSubagents); err != nil {
		return err
	}

	ctx, err := fetchReviewLoopPRContext(opts.PRRef)
	if err != nil {
		return err
	}
	if opts.Watch {
		if err := waitForReviewLoopCodeRabbit(ctx); err != nil {
			return err
		}
	}

	tasks, commonInstruction, found, err := reviewLoopLoadReviewTasks(opts.PRRef, opts.CodeRabbitOnly)
	if err != nil {
		return err
	}
	if !found {
		renderReviewLoopSummary(cmd.OutOrStdout(), ctx, nil)
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No actionable current review feedback found.")
		return err
	}

	classified := classifyReviewLoopFindings(ctx, tasks)
	return runReviewLoopPrompt(cmd.OutOrStdout(), opts, ctx, classified, commonInstruction)
}
