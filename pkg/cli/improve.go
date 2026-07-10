package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/improve"
)

type improveOptions struct {
	suite         string
	kitBinary     string
	from          string
	candidate     string
	issue         string
	maxCandidates int
	dryRun        bool
	json          bool
	createPR      bool
}

func init() {
	rootCmd.AddCommand(newImproveCommand())
}

func newImproveCommand() *cobra.Command {
	opts := &improveOptions{}
	cmd := &cobra.Command{
		Use:          "improve",
		Short:        "Run benchmark-backed Kit harness improvement workflows",
		SilenceUsage: true,
		Long: `Run Kit's benchmark-backed self-improvement workflow.

The improve workflow runs deterministic fixture evals with binary and suite
provenance, mines recurring failure patterns, prepares bounded candidate
prompts, validates candidate metadata, and packages reviewable evidence without
claiming unobserved model quality or bypassing Kit delivery gates.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runImproveOverview(cmd)
		},
	}
	cmd.PersistentFlags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON output")
	cmd.AddCommand(newImproveRunCommand(opts))
	cmd.AddCommand(newImproveMineCommand(opts))
	cmd.AddCommand(newImproveProposeCommand(opts))
	cmd.AddCommand(newImproveValidateCommand(opts))
	cmd.AddCommand(newImproveReportCommand(opts))
	cmd.AddCommand(newImprovePRCommand(opts))
	return cmd
}

func newImproveRunCommand(opts *improveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "Run a Kit improvement benchmark suite",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := config.FindProjectRoot()
			if err != nil {
				return err
			}
			kitBinary := strings.TrimSpace(opts.kitBinary)
			if kitBinary == "" {
				kitBinary = currentExecutable()
			}
			manifest, err := improve.Run(context.Background(), improve.RunOptions{
				ProjectRoot:  root,
				SuiteName:    opts.suite,
				DryRun:       opts.dryRun,
				RunnerBinary: currentExecutable(),
				KitBinary:    kitBinary,
				KitVersion:   Version,
				GitCommit:    currentGitCommit(root),
			})
			if err != nil {
				return err
			}
			if opts.json {
				if err := outputJSON(cmd.OutOrStdout(), manifest); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "kit improve run %s: %s (%d traces)\n", manifest.RunID, manifest.Status, len(manifest.Traces)); err != nil {
					return err
				}
			}
			return improveRunFailure(manifest)
		},
	}
	cmd.Flags().StringVar(&opts.suite, "suite", "default", "benchmark suite name")
	cmd.Flags().StringVar(&opts.kitBinary, "kit-binary", "", "Kit executable evaluated by the suite; defaults to the current executable")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "plan the run without writing artifacts")
	return cmd
}

func improveRunFailure(manifest improve.RunManifest) error {
	if manifest.Status == "failed" {
		return fmt.Errorf("kit improve benchmark %s failed; inspect %s", manifest.RunID, manifest.RunDir)
	}
	return nil
}

func newImproveMineCommand(opts *improveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mine",
		Short: "Mine Kit improvement traces for weakness clusters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := config.FindProjectRoot()
			if err != nil {
				return err
			}
			report, err := improve.Mine(root, opts.from)
			if err != nil {
				return err
			}
			if opts.json {
				return outputJSON(cmd.OutOrStdout(), report)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "kit improve mine: %d clusters\n", len(report.Clusters))
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.from, "from", "", "artifact directory to read; defaults to .kit/improve/latest")
	return cmd
}

func newImproveProposeCommand(opts *improveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Generate bounded candidate prompts from weakness clusters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := config.FindProjectRoot()
			if err != nil {
				return err
			}
			candidates, err := improve.Propose(root, opts.from, opts.maxCandidates)
			if err != nil {
				return err
			}
			if opts.json {
				return outputJSON(cmd.OutOrStdout(), map[string]any{"schema_version": improve.SchemaVersion, "kind": "improve_candidates", "candidates": candidates})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "kit improve propose: %d candidates\n", len(candidates))
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.from, "from", "", "artifact directory to read; defaults to .kit/improve/latest")
	cmd.Flags().IntVar(&opts.maxCandidates, "max-candidates", 3, "maximum candidate count")
	return cmd
}

func newImproveValidateCommand(opts *improveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate Kit improvement candidate metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := config.FindProjectRoot()
			if err != nil {
				return err
			}
			if strings.TrimSpace(opts.candidate) == "" {
				return fmt.Errorf("--candidate is required")
			}
			scorecard, err := improve.Validate(root, opts.candidate)
			if err != nil {
				return err
			}
			if opts.json {
				return outputJSON(cmd.OutOrStdout(), scorecard)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "kit improve validate %s: %s score=%d\n", scorecard.CandidateID, scorecard.Acceptance, scorecard.Score)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.candidate, "candidate", "", "candidate JSON path")
	return cmd
}

func newImproveReportCommand(opts *improveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Render a Kit improvement report",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := config.FindProjectRoot()
			if err != nil {
				return err
			}
			report, err := improve.Report(root, opts.from)
			if err != nil {
				return err
			}
			if opts.json {
				return outputJSON(cmd.OutOrStdout(), map[string]any{"schema_version": improve.SchemaVersion, "kind": "improve_report", "markdown": report})
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), report)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.from, "from", "", "artifact directory to read; defaults to .kit/improve/latest")
	return cmd
}

func newImprovePRCommand(opts *improveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Prepare a PR body from Kit improvement run evidence",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := config.FindProjectRoot()
			if err != nil {
				return err
			}
			body, err := improve.PullRequestBody(root, opts.from, opts.issue)
			if err != nil {
				return err
			}
			if opts.createPR {
				return fmt.Errorf("--create-pr is intentionally gated; run repo-local GitHub delivery workflow with this generated body")
			}
			if opts.json {
				return outputJSON(cmd.OutOrStdout(), map[string]any{"schema_version": improve.SchemaVersion, "kind": "improve_pr_body", "body": body})
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), body)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.from, "from", "", "artifact directory to read; defaults to .kit/improve/latest")
	cmd.Flags().StringVar(&opts.issue, "issue", "", "issue reference such as #46")
	cmd.Flags().BoolVar(&opts.createPR, "create-pr", false, "stop and require repo-local GitHub delivery workflow before creating a PR")
	return cmd
}

func runImproveOverview(cmd *cobra.Command) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Kit improve")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "1. kit improve run --suite default  # capability smoke")
	fmt.Fprintln(cmd.OutOrStdout(), "   kit improve run --suite prompt-system --kit-binary <path>  # prompt regression")
	fmt.Fprintln(cmd.OutOrStdout(), "2. kit improve mine --from .kit/improve/latest")
	fmt.Fprintln(cmd.OutOrStdout(), "3. kit improve propose --from .kit/improve/latest")
	fmt.Fprintln(cmd.OutOrStdout(), "4. kit improve validate --candidate <path>")
	fmt.Fprintln(cmd.OutOrStdout(), "5. kit improve report --from .kit/improve/latest")
	return nil
}

func currentGitCommit(projectRoot string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "git", "-C", projectRoot, "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func currentExecutable() string {
	path, err := os.Executable()
	if err != nil {
		return "kit"
	}
	return path
}
