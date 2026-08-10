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
	suite     string
	kitBinary string
	dryRun    bool
	json      bool
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
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON output")
	cmd.AddCommand(newImproveRunCommand(opts))
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
			manifest, err := improve.Run(cmd.Context(), improve.RunOptions{
				ProjectRoot: root, SuiteName: opts.suite, DryRun: opts.dryRun,
				RunnerBinary: currentExecutable(), KitBinary: kitBinary,
				KitVersion: Version, GitCommit: currentGitCommit(root),
			})
			if err != nil {
				return err
			}
			if opts.json {
				if err := outputJSON(cmd.OutOrStdout(), manifest); err != nil {
					return err
				}
			} else if _, err := fmt.Fprintf(cmd.OutOrStdout(), "kit improve run %s: %s (%d traces)\n", manifest.RunID, manifest.Status, len(manifest.Traces)); err != nil {
				return err
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
