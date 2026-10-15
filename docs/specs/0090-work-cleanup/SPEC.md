---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0090"
  slug: "work-cleanup"
  dir: "0090-work-cleanup"
relationships:
  - type: builds_on
    target: 0087-investigation-evidence
    note: Reuses the optional embedded-rule distribution and the tested inline-recipe pattern.
---
# SPEC

## PURPOSE

Give agents one rule that removes local and remote residue as soon as a piece of work has no value to resume (GH-251). The human's SSD fills quickly. The rule must prune early and often, but it must not lose unique work.

## CONTEXT

- `deletion-safety` requires exact-target confirmation before any hard delete. It exempts task-owned scratch that never became authoritative state.
- `delivery` forbids deleting branches, stashing, cleaning, or force-removing to make a lane operation succeed. It does not say when completed lanes are cleaned up.
- No rule handles residue after completion: merged worktrees, squash-merged branches, abandoned issues, background processes, caches, or scratch artifacts.
- Optional rules ship through the embedded registry and install with `kit rules add`. `context-evidence` set the pattern of a word budget and Go tests that run a published shell recipe.

## REQUIREMENTS

- Keep automatic post-merge recovery of verifiably recoverable task-owned cruft and safe primary default-branch synchronization. Keep unique or unknown state without routine prompts.
- Require current completion, task ownership, successful process inspection, local-state inventory and live selected-remote recovery evidence. A source commit does not restore ignored files, hidden changes, local Git state or custom filesystem metadata.
- Recheck immediately before normal bounded Git/filesystem actions; no reusable classifier result, blanket prune, force/reset/clean fallback or receipt/lease protocol.
- Ship this as an agent rule with aligned generated contracts and focused regressions. No new command or cleanup subsystem. Preserve project customizations during reconciliation.
- Browser profiles, Mint titles, release and downstream rollout are outside scope.

## ACCEPTED PLAN

1. Strengthen the canonical rule's recovery gates while keeping automatic eligible cleanup.
2. Align the universal contract and rendered entrypoints; existing delivery/merge routing stays in place; preserve project-owned overrides.
3. Verify the concrete Git hazards, positive remove/restore and primary sync, and init/upgrade/reconcile convergence in disposable fixtures.

## DECISIONS

- GH-262: fix the rule rather than introduce a cleanup platform. The rejected implementation added a CLI, ownership receipts, leases and platform-specific inspection. Those mechanisms are removed. Current task evidence identifies bounded targets; ambiguous artifacts stay in place.
- Ignored files, failed process checks, stale refs and primary overwrite collisions are real defects. Explicit fail-closed gates and focused Git/distribution tests address them without a new runtime surface.
- Earlier decisions below are historical where they conflict with these current requirements.


- Classify by recoverability, not by action type. A worktree whose head a remote branch or pull-request ref holds, and whose only local residue is ignored content, can be removed without prompts because `deletion-safety` treats it as recoverable. This makes the aggressive behavior consistent with existing rules and avoids an exception to them.
- Squash merges: accept `refs/pull/*/head` as proof of retention, so `git branch -D` is allowed only in that verified case.
- Ignored real `.env*` files block removal (`KEEP_LOCAL_SECRETS`), but symlinked ones do not, because `delivery` links them from the primary checkout.
- Shared caches use a two-tier policy. Tool-marked unused entries are pruned on every sweep. Full regenerable caches are cleared only below 25% or 50 GiB free, so builds stay warm when space is not short. Docker volumes, early Git object pruning, trash, and transcripts are never pruned.
- "Clear evidence" needs two observations in the current session: remote-confirmed completion (or a human stop) and restorability (`RECOVERABLE`, or a verified remote copy of an artifact). With both, cleanup runs at once with no prompt or offer. With either missing or `UNAVAILABLE`, the item is kept and reported. This makes the automatic behavior depend on evidence, not on guesses from names or dates.
- The shipped rule is a bounded standing grant for reversible closes of lane-owned issues and pull requests. A project can narrow or remove it; reconcile keeps local edits. Slack stays read-only.
- Superseded: the first head shipped the rule as optional. The human chose default with automatic execution. The contract carries the automatic trigger because contextual rules are read only when a trigger applies.
- Rejected: a `kit clean` command (conflicts with the harness-subtraction direction and needs host-specific cache knowledge); `git clean -X` for idle lanes (it removes ignored env files).

- Superseded (GH-253, after the first release candidate): the batched hard-delete confirmation, remote-branch deletion, `git branch -D`, full cache clears, and idle-lane dependency pruning. The human found approval prompts during cleanup concerning. Cleanup now never asks; it runs only `git worktree prune`, `git fetch --prune`, `git worktree remove`, `git branch -d`, the work's own process and scratch cleanup, and unused-entry cache prunes when disk space is low. It also closes (never deletes) lane-owned GitHub issues and pull requests with observed completion, because the GitHub UI can reopen them without a prompt to the human. Everything else is kept and reported in one line. The disk saving that was given up is small: the worktree directory is most of a lane's footprint, and branches cost almost nothing.

- GH-257: cleanup also fast-forwards the primary checkout's local default branch. Agents create lanes from `origin/<base>`, so nothing else updated the human's local default branch after a merge. This is the only write that the contract allows in the primary checkout. The sync recipe uses only `git merge --ff-only` (on the default branch with no tracked changes) or `git fetch <remote> <base>:<base>` (on another branch). Git refuses both when work would be lost. Any other state is kept and reported as a literal state. The recipe does not check for running processes in the primary checkout: the human's shell and editor are usually there, a fast-forward on a clean tree does not lose work, and Git's own lock stops a concurrent Git operation.
- Rejected (GH-257): `git pull` (it can merge or rebase, depending on configuration); `git branch -f` or `git reset` (they can drop local commits); syncing the primary checkout from `delivery` (the lane does not need it, and cleanup already runs at every merge).

## DISCOVERIES

- A `case` pattern without a leading `(` inside `$( )` fails in bash's POSIX mode. The recipe uses `(pattern)`.
- The classifier must not `cd` into the worktree, because its own shell would then show as a process in that worktree. It uses `git -C` and a subshell `pwd -P`.

## VALIDATION

- GH-262 rule-only refactor: full Go tests, focused Git/action and install/upgrade/custom/convergence regressions, vet, lint (0 issues), build, formatting and project check PASS. Scoped real reconciliation converges with zero planned changes. Native cleanup checks from rejected PR heads are not evidence for this revision.
- Earlier validation below records the released shell recipes.

- `go test ./pkg/cli -run TestWorkCleanup -count=1`: pass, with `sh` as macOS bash and as `dash`.
- `gofmt -l .` (empty), `make vet`, `make lint` (0 issues), `make test` (all packages ok), `make build`: pass.
- `kit check --project`: pass with one advisory warning about `docs/references/README.md` that also appears on unchanged `main`.
- The recipe also parses with `dash -n`.

## OUTCOME

- GH-262: the final PR changes only the rule, its contract/distribution, current spec and focused tests/fixtures. There is no new command, production Go change, receipt/lease protocol or platform inspection subsystem. Automatic eligible cleanup and primary sync remain. No merge, release, downstream rollout or cleanup of user directories is authorized.

- Shipped `docs/references/rules/work-cleanup.md` as a default rule, the contract trigger, and tests in `pkg/cli/work_cleanup_*_test.go`.
- Existing projects receive the rule and contract lines at their next `kit reconcile`.
- Risk: lsof-based in-use detection sees only processes visible to the user. Without lsof, the agent must check in-use state itself.
- Risk: cache command lists name common tools only. Projects can add local entries; reconcile preserves local edits.

## REPOSITORY MEMORY

- GH-257: added the primary checkout sync recipe and states to `work-cleanup.md`; narrowed the primary-checkout read-only invariant in the contract template, the three instruction files, and `delivery.md`; updated the references index; added `pkg/cli/work_cleanup_primary_sync_test.go`.

- GH-253: narrowed `work-cleanup.md`, the contract invariant in the template and the three instruction files, the references index, and the registry tests.

- Added `docs/references/rules/work-cleanup.md`.
- Updated `docs/references/README.md` index and usage note.
- Updated `internal/templates/universal_contract.md` and this repository's `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` with the invariant and routing line.
- Updated `delivery.md` and `github-pr-merge.md` with completion hooks.
- Unchanged: `deletion-safety`; the new rule defers to it.
