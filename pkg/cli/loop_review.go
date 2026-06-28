package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func newLoopReviewCommand() *cobra.Command {
	opts := loopReviewOptions{}
	cmd := &cobra.Command{
		Use:           "review [feature]",
		Short:         "Run a correctness review loop over changed code",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Run a coding-agent correctness loop over changes not in the remote
mainline. The loop repeats local review and repair passes until the configured
agent reports at least 95% correctness and ends its final response with done.

With --pr, CodeRabbit feedback is checked opportunistically while local review
continues. Use --watch or --wait-for-coderabbit to wait for CodeRabbit before
finalizing.

Review prompts use one agent by default. Pass --subagents to allow the parent
agent to run pre-analysis and decide whether subagents are useful. Interactive
terminals ask before rerunning when a previous loop review exists or the current
run reaches max iterations.
Human-readable runs stream emoji-marked progress and child-agent output to
stderr; --json keeps stdout machine-readable and suppresses progress output.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoopReviewCommand(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.BaseRef, "base", "", "base ref for changed-code review (default origin/main, then main)")
	cmd.Flags().StringVar(&opts.PRRef, "pr", "", "optionally ingest CodeRabbit feedback from a PR URL, Markdown link, owner/repo#number, or current-repo number")
	cmd.Flags().BoolVar(&opts.WaitForCodeRabbit, "watch", false, "wait for CodeRabbit completion before finalizing PR-mode review")
	cmd.Flags().BoolVar(&opts.WaitForCodeRabbit, "wait-for-coderabbit", false, "wait for CodeRabbit completion before finalizing PR-mode review")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "show the first review prompt without running the configured agent")
	cmd.Flags().IntVar(&opts.MinConfidence, "min-confidence", 0, "minimum correctness percentage required to stop (0 uses loop config, goal_percentage, then 95)")
	cmd.Flags().IntVar(&opts.MaxIterations, "max-iterations", 0, "maximum review iterations (0 uses loop config, then 20)")
	cmd.Flags().BoolVar(&opts.UseSubagents, "subagents", false, "allow the review agent to pre-analyze and choose subagents")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output loop review report as JSON")
	return cmd
}

func runLoopReviewCommand(cmd *cobra.Command, args []string, opts loopReviewOptions) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	opts.ProjectRoot = projectRoot
	opts.Config = cfg
	opts.MinConfidence = effectiveLoopMinConfidence(cfg, opts.MinConfidence)
	opts.MaxIterations = effectiveLoopMaxIterations(cfg, opts.MaxIterations)
	opts.Agent = cfg.Loop.Agent
	if !opts.JSON && !opts.DryRun {
		opts.Progress = &loopReviewSynchronizedWriter{writer: cmd.ErrOrStderr()}
	}

	if len(args) == 1 {
		feat, err := loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature %q not found for loop review: %w", args[0], err)
		}
		opts.Feature = feat
	}

	if shouldPromptLoopReviewRerun(cmd, opts) {
		previous, found, err := latestLoopReviewReport(projectRoot)
		if err != nil {
			return err
		}
		if found {
			again, err := promptLoopReviewRerun(
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				"Previous loop review run found.",
				previous,
			)
			if err != nil {
				return err
			}
			if !again {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "Loop review skipped.")
				return err
			}
		}
	}

	for {
		report, runErr := executeLoopReview(cmd.Context(), opts)
		outputErr := outputLoopReviewReport(cmd, report, opts.JSON)
		if outputErr != nil {
			return outputErr
		}
		if runErr == nil {
			return nil
		}
		if isLoopReviewMaxIterations(report) && shouldPromptLoopReviewRerun(cmd, opts) {
			again, err := promptLoopReviewRerun(
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				"Loop review reached max iterations.",
				report,
			)
			if err != nil {
				return err
			}
			if again {
				continue
			}
		}
		return &silentCLIError{err: runErr}
	}
}
