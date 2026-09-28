package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/feature"
	"github.com/spf13/cobra"
)

var reconcileOutputOnly bool
var reconcileAll bool
var reconcileIncludeFiles bool
var reconcileForce bool
var reconcileDryRun bool
var reconcileDiff bool
var reconcileRefreshFiles []string

var reconcileCmd = &cobra.Command{
	Use:   "reconcile [feature]",
	Short: "Migrate a Kit project to the current structure and audit its docs",
	Long: `Bring an existing Kit project, created by any Kit release, to the current
Kit structure, then audit project documents against the current contract.

Reconcile:
  - puts exactly one current Kit-managed contract block in AGENTS.md,
    CLAUDE.md, and .github/copilot-instructions.md, keeping project guidance
    outside the block
  - updates unedited Kit sections of docs/CONSTITUTION.md and its baseline
  - updates shipped rules Kit installed and nobody edited
  - removes retired Kit files (docs/agents/, docs/references/workflows/,
    docs/PROJECT_PROGRESS_SUMMARY.md, retired rules) only when they are exactly
    as a Kit release generated them and committed to Git, so Git can restore them
  - prunes retired rules and obsolete keys from .kit.yaml and records the
    current instruction_scaffold_version

Anything edited, ambiguous, or uncommitted is kept and reported. --force also
replaces edited Kit sections of pre-contract entry files and edited shipped
rules; sections Kit never wrote are always kept.

In the primary checkout, reconcile applies changes in a linked worktree on
branch kit-reconcile (created or reused) and prints where to review them.
Use --dry-run --diff to preview without writing anything.

With a feature argument, audits only that feature's docs.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReconcile,
}

func init() {
	reconcileCmd.Flags().BoolVar(&reconcileOutputOnly, "output-only", false, "accepted for compatibility; reconcile always prints to stdout")
	_ = reconcileCmd.Flags().MarkHidden("output-only")
	reconcileCmd.Flags().BoolVar(&reconcileAll, "all", false, "reconcile the whole project explicitly")
	reconcileCmd.Flags().BoolVar(&reconcileIncludeFiles, "include-files", false, "accepted for compatibility; whole-project reconcile always includes Kit-managed files")
	_ = reconcileCmd.Flags().MarkHidden("include-files")
	reconcileCmd.Flags().BoolVarP(&reconcileForce, "force", "f", false, "also replace edited Kit sections and edited shipped rules")
	reconcileCmd.Flags().BoolVar(&reconcileDryRun, "dry-run", false, "preview Kit-managed file changes without writing files")
	reconcileCmd.Flags().BoolVar(&reconcileDiff, "diff", false, "print planned file changes as a unified diff with --dry-run")
	reconcileCmd.Flags().StringArrayVar(&reconcileRefreshFiles, "file", nil, "limit the file migration to one Kit-managed file; repeat for multiple files")
	rootCmd.AddCommand(reconcileCmd)
}

func runReconcile(cmd *cobra.Command, args []string) error {
	if reconcileAll && len(args) > 0 {
		return fmt.Errorf("--all cannot be used with a feature argument")
	}
	if reconcileDiff && !reconcileDryRun {
		return fmt.Errorf("--diff requires --dry-run")
	}
	if len(args) > 0 && (len(reconcileRefreshFiles) > 0 || reconcileForce) {
		return fmt.Errorf("--file and --force apply to whole-project reconcile only")
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

	applyFiles := feat == nil
	var deliverySnapshot []managedFileDeliverySnapshot
	if applyFiles {
		opts := initRefreshOptions{force: reconcileForce, files: reconcileRefreshFiles}
		upToDate, err := reconcileHasNoChanges(projectRoot, opts)
		if err != nil {
			return err
		}
		target, err := resolveReconcileTarget(cmd.OutOrStdout(), projectRoot, reconcileDryRun, upToDate)
		if err != nil {
			return err
		}
		if target.current {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Kit-managed files are current; nothing to write.")
			if err != nil {
				return err
			}
		} else if deliverySnapshot, err = runInitRefreshWithSnapshot(target.projectRoot, initRefreshOptions{
			force:      reconcileForce,
			dryRun:     reconcileDryRun,
			diff:       reconcileDiff,
			files:      reconcileRefreshFiles,
			outputOnly: reconcileOutputOnly,
		}); err != nil {
			return err
		}
		if target.worktree {
			printReconcileWorktreeNextSteps(cmd.OutOrStdout(), target, reconcileDryRun)
		}
		projectRoot = target.projectRoot
		if cfg, err = config.Load(projectRoot); err != nil {
			return fmt.Errorf("failed to reload config after reconcile: %w", err)
		}
	}
	if reconcileDryRun {
		return nil
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		return err
	}
	report.DeliverySnapshot = deliverySnapshot
	if active, err := feature.FindActiveFeatureWithState(cfg.SpecsPath(projectRoot), cfg); err != nil {
		return fmt.Errorf("failed to resolve active feature: %w", err)
	} else if feat == nil || (active != nil && active.DirName == feat.DirName) {
		report.Findings = append(report.Findings, auditActiveFrontendRulesetAdvisory(projectRoot, active)...)
	}

	if len(report.Findings) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), report.cleanResult())
		return err
	}
	return writeReconcileFindings(cmd.OutOrStdout(), report)
}
