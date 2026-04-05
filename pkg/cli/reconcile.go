package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
)

var reconcileCopy bool
var reconcileOutputOnly bool
var reconcileAll bool

var reconcileCmd = &cobra.Command{
	Use:   "reconcile [feature]",
	Short: "Audit Kit-managed docs and output a reconciliation prompt",
	Long: `Audit Kit-managed project documents against the current Kit contract.

Without a feature argument, reconciles the whole project by default.
Use --all as an explicit alias for whole-project mode.
With a feature argument, audits only that feature's docs plus related rollup drift.

This command is prompt-only in v1. It does not edit files directly.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReconcile,
}

func init() {
	reconcileCmd.Flags().BoolVar(&reconcileCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	reconcileCmd.Flags().BoolVar(&reconcileOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	reconcileCmd.Flags().BoolVar(&reconcileAll, "all", false, "audit the whole project explicitly")
	addPromptOnlyFlag(reconcileCmd)
	rootCmd.AddCommand(reconcileCmd)
}

func runReconcile(cmd *cobra.Command, args []string) error {
	if reconcileAll && len(args) > 0 {
		return fmt.Errorf("--all cannot be used with a feature argument")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var feat *feature.Feature
	if len(args) == 1 {
		feat, err = loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		return err
	}

	if len(report.Findings) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), report.cleanResult())
		return err
	}

	outputOnly, _ := cmd.Flags().GetBool("output-only")
	if !outputOnly {
		printReconcileSummary(report)
		printWorkflowInstructions("reconcile (supporting step)", []string{
			"run the generated prompt in the current coding agent session",
			"keep changes limited to documentation reconciliation",
		})
	}

	return outputPromptWithClipboardDefault(buildReconcilePrompt(report), outputOnly, reconcileCopy)
}
