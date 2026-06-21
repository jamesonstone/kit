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
