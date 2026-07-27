package cli

import "fmt"

func docsOnlyWorkflowRule(target string) string {
	return fmt.Sprintf(
		"Only update %s; do not modify product code, tests, runtime config, generated artifacts, or implementation files.",
		target,
	)
}

func managedFileDeliveryInstructions(projectRoot string) []string {
	return []string{
		fmt.Sprintf(
			"Inspect `git status --short --branch` in `%s` and identify only the version-control-eligible unstaged or untracked files created or updated by this Kit command; preserve every unrelated change.",
			projectRoot,
		),
		"Create or reuse the human-assigned GitHub issue first, then create or reuse its exact `GH-<issue-number>` branch and canonical writable worktree; reuse the current worktree when it already owns that lane.",
		"If the command ran in the protected root checkout, additionally move the in-scope unstaged and untracked files into the writable issue worktree, verify that the destination content and diff match, then remove only the transferred source state so those files do not remain stale on the default branch.",
		"Integrate the transferred files with the rest of the change, validate the complete diff, stage explicit paths, commit on the issue branch, push it, and create or update the ready pull request.",
		"Never transfer or stage `.env`, secrets, ignored files, or machine-local configuration; never commit on the protected default branch, bulk-stage files, stash, reset, clean, overwrite destination work, or disturb unrelated root-checkout or worktree changes.",
	}
}
