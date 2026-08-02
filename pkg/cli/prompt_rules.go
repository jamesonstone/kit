package cli

import (
	"fmt"
	"sort"
	"strings"
)

func docsOnlyWorkflowRule(target string) string {
	return fmt.Sprintf(
		"Only update %s; do not modify product code, tests, runtime config, generated artifacts, or implementation files.",
		target,
	)
}

func managedFileDeliveryInstructions(
	projectRoot string,
	snapshots ...[]managedFileDeliverySnapshot,
) []string {
	var snapshot []managedFileDeliverySnapshot
	if len(snapshots) > 0 {
		for _, change := range snapshots[0] {
			change.Path = normalizeManagedFileDeliveryPath(change.Path)
			if managedFileDeliveryPathEligible(projectRoot, change.Path) {
				snapshot = append(snapshot, change)
			}
		}
	}
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].Path < snapshot[j].Path
	})

	if len(snapshot) == 0 {
		return managedFileDeliveryInstructionsWithoutSnapshot(projectRoot)
	}

	entries := make([]string, 0, len(snapshot))
	for _, change := range snapshot {
		entries = append(entries, fmt.Sprintf(
			"`%s` (%s; pre-command %s; expected %s)",
			change.Path,
			change.Action,
			change.PreCommandState,
			change.ResultState,
		))
	}
	boundary := "Treat only this exact command-owned snapshot as transferable: " + strings.Join(entries, "; ") + "."

	return []string{
		boundary,
		fmt.Sprintf(
			"Inspect `git status --short --branch` and exact-path diffs in `%s` only to verify the captured snapshot; never expand the command-owned boundary from post-command status, and preserve every unrelated change.",
			projectRoot,
		),
		"Create or reuse the human-assigned GitHub issue first, then create or reuse its exact `GH-<issue-number>` branch and canonical writable worktree; reuse the current worktree when it already owns that lane.",
		"Before transfer, verify every captured source path matches its expected state and every destination path matches its pre-command state; abort if a captured destination path has staged, working-tree, or untracked changes, if any state differs, or if the destination index is not empty.",
		"If the command ran in the protected root checkout, apply only the captured command-owned delta in the writable issue worktree: create or update captured files, remove captured deleted paths, and preserve every other destination path.",
		"Verify every destination path matches its expected state, explicitly stage only the captured paths (including deleted paths), and require `git diff --cached --name-only` plus the staged patch to contain exactly the captured command-owned change.",
		"Only after destination content and index verification succeeds, restore each captured root path to its exact pre-command state—removing command-created files and restoring command-updated or command-removed files—then verify that no captured command-owned state remains in the protected root checkout.",
		"Integrate the verified staged files with the rest of the issue change, validate the complete diff, commit on the issue branch, push it, and create or update the ready pull request.",
		"Never transfer or stage `.env`, secrets, ignored files, or machine-local configuration; never commit on the protected default branch, bulk-stage files, stash, reset, clean, overwrite destination work, or disturb unrelated root-checkout or worktree changes.",
	}
}

func managedFileDeliveryInstructionsWithoutSnapshot(projectRoot string) []string {
	return []string{
		fmt.Sprintf("No exact command-owned path snapshot is present. Do not infer or transfer a managed-file delta from post-command Git status in `%s`; apply only the listed manual findings until a fresh snapshot exists.", projectRoot),
		"Create or reuse the human-assigned GitHub issue first, then create or reuse its exact `GH-<issue-number>` branch and canonical writable worktree; reuse the current worktree when it already owns that lane.",
		"In that exact writable lane, rerun the write-capable Kit command. Before staging, require the rerun to emit a new exact command-owned snapshot containing every version-control-eligible path, action, pre-command state, and expected result state; if it cannot, do not adopt managed-file changes and report the blocker.",
		"Adopt only paths in the new snapshot, verify each path against its captured states, explicitly stage only those paths, and require `git diff --cached --name-only` plus the staged patch to match the snapshot exactly.",
		"Validate the complete diff, commit on the issue branch, push it, and create or update the ready pull request.",
		"Never transfer or stage `.env`, secrets, ignored files, or machine-local configuration; never commit on the protected default branch, bulk-stage files, stash, reset, clean, overwrite destination work, or disturb unrelated root-checkout or worktree changes.",
	}
}
