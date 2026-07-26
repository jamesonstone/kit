package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	reviewLoopExecutor          = runReviewLoop
	reviewLoopFetchPRContext    = fetchReviewLoopPRContext
	reviewLoopWaitForCodeRabbit = waitForReviewLoopCodeRabbit
	reviewLoopLoadReviewTasks   = loadDispatchPRReviewTasks
)

func runReviewLoop(cmd *cobra.Command, opts reviewLoopOptions) error {
	if strings.TrimSpace(opts.PRRef) == "" {
		return fmt.Errorf("--pr is required")
	}
	if err := validateDispatchMaxSubagents(opts.MaxSubagents); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	repair, err := resolvePRRepairContext(
		cmd.Context(),
		cmd.InOrStdin(),
		cmd.ErrOrStderr(),
		cwd,
		opts.PRRef,
	)
	if err != nil {
		return err
	}

	ctx, err := reviewLoopFetchPRContext(opts.PRRef)
	if err != nil {
		return err
	}
	ctx.LocalRoot = repair.WorktreePath
	ctx.Repair = repair
	if opts.Watch {
		if err := reviewLoopWaitForCodeRabbit(ctx); err != nil {
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
