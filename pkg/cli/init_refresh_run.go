package cli

import (
	"context"
	"fmt"
	"io"
	"os"
)

func runInitRefreshWithSnapshot(
	projectRoot string,
	opts initRefreshOptions,
) ([]managedFileDeliverySnapshot, error) {
	if !opts.outputOnly && !opts.dryRun {
		fmt.Println("🔄 Refreshing Kit-managed project files...")
	}

	plan, err := buildInitRefreshPlan(context.Background(), projectRoot, opts)
	if err != nil {
		return nil, err
	}
	surfaces := issueTemplateSurfaces(projectRoot)
	org := orgIssueTemplateWarning(projectRoot, plan.changes)
	askFirst := !opts.dryRun && streamsHaveInteractiveTerminal(os.Stdin, os.Stdout)
	if askFirst {
		if err := printPrewritePlan(os.Stdout, plan.changes, nil); err != nil {
			return nil, err
		}
		if err := printSurfaceWarnings(os.Stdout, surfaces); err != nil {
			return nil, err
		}
		if err := printPrewriteWarnings(os.Stdout, prewriteWarnings(plan.changes, nil), org); err != nil {
			return nil, err
		}
	}
	var notes map[string]string
	plan.changes, notes, err = gateReconcilePlan(projectRoot, plan.changes, opts.force, opts.dryRun)
	if err != nil {
		return nil, err
	}
	plan.stats = recountInitRefreshStats(plan.changes)
	if !askFirst {
		if err := printPrewritePlan(os.Stdout, plan.changes, notes); err != nil {
			return nil, err
		}
		if err := printSurfaceWarnings(os.Stdout, surfaces); err != nil {
			return nil, err
		}
		if err := printPrewriteWarnings(os.Stdout, prewriteWarnings(plan.changes, nil), org); err != nil {
			return nil, err
		}
	}
	deliverySnapshot := managedFileDeliverySnapshotFromInitRefresh(projectRoot, plan.changes)

	if opts.dryRun {
		printInitRefreshDryRun(plan.changes, plan.stats, opts)
		printInitRefreshNotes(plan.notes, opts)
		return deliverySnapshot, nil
	}

	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		return nil, err
	}

	if !opts.outputOnly {
		fmt.Println("\n✅ Kit managed refresh complete!")
		if plan.stats.changed() == 0 {
			fmt.Println("   No Kit-managed project changes needed.")
		}
		fmt.Printf(
			"   Created: %d, Updated: %d, Merged: %d, Removed: %d, Skipped: %d\n",
			plan.stats.created,
			plan.stats.updated,
			plan.stats.merged,
			plan.stats.removed,
			plan.stats.skipped,
		)
	}
	printInitRefreshNotes(plan.notes, opts)
	return deliverySnapshot, nil
}

// reportSuppressedIssueTemplate prints the plan when the only reason reconcile
// would stay quiet is that a non-interactive run skipped the opt-in issue
// template beside an existing issue setup.
func reportSuppressedIssueTemplate(out io.Writer, root string, opts initRefreshOptions) (bool, error) {
	opts.dryRun = true
	opts.outputOnly = true
	plan, err := buildInitRefreshPlan(context.Background(), root, opts)
	if err != nil {
		return false, err
	}
	wanted := instructionFileSkipped
	for _, change := range plan.changes {
		if change.relativePath == issueTemplatePath {
			wanted = change.result
		}
	}
	if wanted != instructionFileCreated {
		return false, nil
	}
	gated, notes, err := gateReconcilePlan(root, plan.changes, opts.force, true)
	if err != nil {
		return false, err
	}
	suppressed := false
	for _, change := range gated {
		if change.relativePath == issueTemplatePath && change.result == instructionFileSkipped {
			suppressed = true
		}
	}
	if !suppressed {
		return false, nil
	}
	if err := printPrewritePlan(out, gated, notes); err != nil {
		return false, err
	}
	if err := printSurfaceWarnings(out, issueTemplateSurfaces(root)); err != nil {
		return false, err
	}
	_, err = fmt.Fprintln(out, "No files written.")
	return err == nil, err
}
