package cli

func utilityCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability("upgrade", "Utilities", "Upgrade the Kit CLI installation.", mutationNetwork, withNetwork("downloads release metadata or binaries"), withFileWrites("writes the installed Kit binary or related install files"), withFlags(flag("--check", "check for an upgrade without installing when supported", "prefer for read-only inspection")), withRelated(related("version", "shows current installed version"))),
		capability("version", "Utilities", "Print the Kit CLI version.", mutationNone, withRelated(related("upgrade", "updates the installed version"))),
		capability("completion", "Utilities", "Generate shell completion scripts.", mutationNone, withFileWrites("none by default", "the shell may redirect output to a completion file outside Kit"), withRelated(related("help", "shows command syntax"))),
		capability(
			"git wt list",
			"Utilities",
			"Select or print registered worktrees with the primary checkout pinned above newest-first lanes.",
			mutationNone,
			withFileWrites("none"),
			withGitMutation("none; reads registered worktree metadata, status, and commit dates only"),
			withFlags(
				flag("--sort", "order by updated (default), state, head, or path"),
				flag("--root-position", "pin the primary worktree at top (default) or bottom"),
				flag("--reverse", "reverse the selected sort order"),
				flag("--plain", "print the table instead of opening the terminal selector"),
			),
			withRelated(related("git wt home", "opens a child shell in the primary checkout"), related("git wt cd", "opens a child shell for an exact lane"), related("git wt path", "prints an exact lane path for parent-shell navigation")),
			withWhenToUse("Use interactively to choose a registered worktree with arrow keys or Tab; press h to open the primary checkout immediately.", "Use with --plain or redirected output when a script or terminal needs the table."),
			withWhenNotToUse("Do not expect the selected child shell to change the parent shell's directory.", "Do not use as a policy dependency; native git worktree commands remain authoritative."),
			withExamples("git wt list", "git wt list --root-position bottom", "git wt list --plain --sort path", "git wt list --sort state --reverse"),
			withCaveats("Terminal selection uses color and opens the configured shell in the chosen worktree; the primary checkout and every main branch row remain bright green in every repository.", "Non-terminal input or output automatically uses the stable plain table.", "The primary checkout is pinned first by default; remaining entries are newest first by full commit timestamp, while --sort state, --sort head, and --sort path select alternate ordering.", "LAST UPDATED is display-only, converted to the running user's local timezone, and shown at calendar-day plus HH:MM precision without seconds."),
		),
		capability(
			"git wt sync",
			"Utilities",
			"Reconcile origin's default branch and remove only exact clean worktree lanes proven merged by GitHub.",
			mutationGit,
			withNetwork(
				"reads origin's live default branch and GitHub pull-request metadata",
				"ordinary sync fetches and prunes origin only; --dry-run does not fetch",
			),
			withFileWrites(
				"ordinary sync may unlink a verified managed .env symlink and remove a proven-safe worktree directory",
				"--dry-run performs no filesystem write",
			),
			withGitMutation(
				"ordinary sync may fast-forward the clean local default branch, remove proven merged worktrees, delete their local branches with git branch -d, and prune stale metadata",
				"--dry-run performs no fetch, ref update, removal, branch deletion, or metadata pruning",
			),
			withFlags(
				flag("--dry-run", "read live origin/GitHub state and report exact decisions without local mutation"),
				flag("--json", "emit the same typed report as deterministic indented JSON"),
			),
			withRelated(
				related("git wt list", "selects or prints worktrees without network access or mutation"),
				related("git wt path", "prints one exact registered lane path"),
			),
			withWhenToUse(
				"Use explicitly when refreshing origin/main and retiring merged same-repository PR lanes.",
				"Use --dry-run before applying when you want a strictly non-mutating preview.",
			),
			withWhenNotToUse(
				"Do not use as proof that ancestry or a deleted remote branch means a PR merged.",
				"Do not expect dirty, fork-backed, ambiguous, detached, legacy, primary, or current worktrees to be removed.",
			),
			withExamples("git wt sync --dry-run", "git wt sync", "git wt sync --json"),
			withCaveats(
				"Removal requires an exact canonical path, one merged same-repository PR into origin's default branch, exact PR-head OID equality, and no material except the verified managed .env symlink.",
				"The command never stashes, resets, cleans, force-removes, force-deletes, force-pushes, or deletes remote branches.",
				"Manual git wt remove retains its upstream/ahead proof and preserves the branch; sync uses merged-PR plus exact-head proof and then attempts ordinary git branch -d.",
				"Failures preserve ambiguous lanes, independent candidates may continue, and any operation failure makes the overall command nonzero.",
			),
		),
		capability(
			"git wt home",
			"Utilities",
			"Open an interactive shell in Git's primary worktree.",
			mutationNone,
			withGitMutation("none; reads registered worktree metadata only"),
			withRelated(related("git wt list", "offers the same primary checkout as a pinned row and h hotkey"), related("git wt path", "prints an exact lane path for parent-shell navigation")),
			withWhenToUse("Use from any linked worktree when you want a shell rooted in the clone's primary checkout."),
			withWhenNotToUse("Do not expect this child shell to change the parent shell's directory."),
			withExamples("git wt home"),
			withCaveats("The command starts the configured `$SHELL` (or `/bin/sh`) in Git's primary worktree and returns when that shell exits."),
		),
		capability(
			"git wt <branch>",
			"Utilities",
			"Open an exact branch worktree, prompting to create it when absent.",
			mutationGit,
			withGitMutation("reads registered worktrees and may create a branch and worktree after interactive confirmation"),
			withWhenToUse("Use for quick interactive navigation to a branch such as GH-93."),
			withExamples("git wt GH-93"),
			withCaveats("Existing lanes open a child shell; the command cannot change the parent shell's directory.", "Missing lanes prompt exactly `do you want to create this worktree? (y/n)`. A local or origin branch is attached after `y`; origin-default creation happens only when the branch exists in neither the local repository nor origin."),
		),
		capability(
			"git wt cd",
			"Utilities",
			"Open an interactive shell in an exact registered worktree lane.",
			mutationNone,
			withGitMutation("none; reads registered worktree metadata only"),
			withWhenToUse("Use for manual testing when you want a shell rooted in an existing lane."),
			withWhenNotToUse("Do not expect this child shell to change the parent shell's directory; use `git wt path` with `cd` for that."),
			withExamples("git wt cd GH-101"),
			withCaveats("The command starts the configured `$SHELL` (or `/bin/sh`) with the lane as its working directory and returns when that shell exits.", "The lane must already be an exact registered worktree."),
		),
		capability(
			"git wt path",
			"Utilities",
			"Print the exact registered worktree path for a lane.",
			mutationNone,
			withGitMutation("none; reads registered worktree metadata only"),
			withWhenToUse(
				"Use in shell command substitution to navigate to an existing lane.",
				"Use when scripts need the canonical path of an exact registered lane.",
			),
			withWhenNotToUse(
				"Do not expect an external command to change the parent shell's directory.",
				"Do not use for fuzzy lane discovery; use `git wt list` to inspect registered worktrees.",
			),
			withExamples(`cd "$(git wt path GH-101)"`, `git -C "$(git wt path GH-101)" status`),
			withCaveats(
				"The command is read-only and rejects unregistered lanes or traversal instead of guessing.",
				"`git wt` is an optional manual convenience; Kit-managed rules and reconciled guidance use native `git worktree` commands.",
			),
		),
		capability("help", "Utilities", "Show command help and flag syntax.", mutationNone, withRelated(related("capabilities", "adds behavior and safety metadata"))),
	}
}
