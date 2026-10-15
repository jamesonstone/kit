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

Preserve complete local work at automatic task completion (GH-262). Inventory
lanes without deleting persistent local state or updating primary checkouts.
A remote commit is evidence about committed source, not a backup of local state.

## CONTEXT

- `deletion-safety` requires exact-target confirmation before any hard delete. It exempts task-owned scratch that never became authoritative state.
- `delivery` forbids deleting branches, stashing, cleaning, or force-removing to make a lane operation succeed. It does not say when completed lanes are cleaned up.
- No rule handles residue after completion: merged worktrees, squash-merged branches, abandoned issues, background processes, caches, or scratch artifacts.
- Optional rules ship through the embedded registry and install with `kit rules add`. `context-evidence` set the pattern of a word budget and Go tests that run a published shell recipe.

## REQUIREMENTS

- Apply inventory at completion, stop, new lane, session end and low-disk events.
- Retain worktrees and branches as KEEP_LOCAL_STATE. Unknown path or Git evidence returns UNAVAILABLE and still retains the target.
- Do not automatically remove local files, prune refs or caches, or update primary checkouts.
- Removal requires a separate human request naming targets and complete recovery evidence under deletion-safety and delivery.
- Do not prompt for routine cleanup. Report retained state without implying it is backed up.
- Stop only task-owned idle watchers/checks with no outstanding work; do not copy evidence during cleanup.
- Preserve project-owned rules and guidance. Install the same consistent policy through init and reconcile.
- No new command. Browser-profile policy and Mint title changes are outside GH-262.

## ACCEPTED PLAN

1. Replace the embedded cleanup rule with inventory-only retention.
2. Remove automatic primary-sync and leftover-removal authority from the canonical universal contract and delivery; render all three instruction copies and update completion references and registry hashes.
3. Exercise the published recipe in isolated repositories and compare primary/lane file bytes, modes, symlinks and Git metadata.
4. Verify fresh init, historical migrations, a managed v3.0.44 upgrade, local-custom preservation and repeated reconciliation convergence.
5. Deliver through a pull request. Do not merge, release or modify downstream repositories under this task.

## DECISIONS

- GH-262 supersedes automatic removal and primary sync from GH-251, GH-253 and GH-257. Downstream reviews in lsmc-bio/labcore#1264 and lsmc-bio/labcore-ui#896 showed that a remote head cannot restore ignored data, hidden tracked edits, reflog-only commits, worktree configuration, filesystem modes or active Git operations. Git status and a fast-forward-only merge do not prove preservation of this state.
- Automatic completion now inventories and retains every lane as KEEP_LOCAL_STATE. It does not remove local data, prune refs or caches, or update primary checkouts. Removal requires a separate named human request and complete recovery evidence under deletion-safety and delivery. Unknown evidence means retain without a routine approval prompt.
- The canonical universal contract and delivery rule remove both the primary-sync and command-owned-leftover exceptions. Fixing only the contextual rule leaves contradictory authority in every generated contract. Project-owned overrides remain preserved; Kit's managed blocks carry the consistent default.
- Tests execute the published retention recipe only in disposable fixtures and compare bytes, modes, symlinks and Git metadata before and after. Fresh init and migration must install the policy; repeated reconcile must converge and preserve local customizations.
- The earlier decisions below are historical evidence, superseded where they authorize automatic local-state removal or primary sync.

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

- GH-262: targeted cleanup, contract, rendered-output and rule-invariant tests PASS.
- GH-262: gofmt (no output), make vet, make lint (0 issues), go test ./... (all packages PASS), make build and the built binary's check --project PASS. The references-index compatibility warning remains non-blocking.
- GH-262: real scoped reconciliation updates the registry; a subsequent dry-run --diff reports zero planned changes. Integration fixtures cover repeated convergence for fresh/historical/managed-v3.0.44/custom installations.
- Earlier validation below records the superseded policy, not proof of current cleanup safety.

- `go test ./pkg/cli -run TestWorkCleanup -count=1`: pass, with `sh` as macOS bash and as `dash`.
- `gofmt -l .` (empty), `make vet`, `make lint` (0 issues), `make test` (all packages ok), `make build`: pass.
- `kit check --project`: pass with one advisory warning about `docs/references/README.md` that also appears on unchanged `main`.
- The recipe also parses with `dash -n`.

## OUTCOME

- GH-262: canonical inventory-only policy, consistent generated contracts, delivery/merge completion references and registry state are implemented. Disposable fixtures assert preservation of ignored data, hidden edits, symlinks, executable modes, reflogs, worktree config, paused Git state and primary ignored-file collisions. Missing/failed lsof cannot grant removal; failed Git evidence returns UNAVAILABLE.
- GH-262: local validation complete; pull-request CI pending. Release and downstream rollout are outside this task.
- Historical release outcomes follow; their automatic removal/sync behavior is superseded.


- Shipped `docs/references/rules/work-cleanup.md` as a default rule, the contract trigger, and tests in `pkg/cli/work_cleanup_*_test.go`.
- Existing projects receive the rule and contract lines at their next `kit reconcile`.
- Risk: lsof-based in-use detection sees only processes visible to the user. Without lsof, the agent must check in-use state itself.
- Risk: cache command lists name common tools only. Projects can add local entries; reconcile preserves local edits.

## REPOSITORY MEMORY

- GH-262: corrected the recovery model in this spec and current canonical policy, rendered instructions, delivery/merge hooks, references index and regression fixtures. Prior policy decisions remain historical evidence.

- GH-257: added the primary checkout sync recipe and states to `work-cleanup.md`; narrowed the primary-checkout read-only invariant in the contract template, the three instruction files, and `delivery.md`; updated the references index; added `pkg/cli/work_cleanup_primary_sync_test.go`.

- GH-253: narrowed `work-cleanup.md`, the contract invariant in the template and the three instruction files, the references index, and the registry tests.

- Added `docs/references/rules/work-cleanup.md`.
- Updated `docs/references/README.md` index and usage note.
- Updated `internal/templates/universal_contract.md` and this repository's `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` with the invariant and routing line.
- Updated `delivery.md` and `github-pr-merge.md` with completion hooks.
- Unchanged: `deletion-safety`; the new rule defers to it.
