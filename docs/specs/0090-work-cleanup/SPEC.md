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

Recover disk space automatically after completed work, without losing unique work. GH-262 preserves the original automatic cleanup purpose and replaces unsafe recovery assumptions with a native, fail-closed proof-and-action implementation.

## CONTEXT

- `deletion-safety` requires exact-target confirmation before any hard delete. It exempts task-owned scratch that never became authoritative state.
- `delivery` forbids deleting branches, stashing, cleaning, or force-removing to make a lane operation succeed. It does not say when completed lanes are cleaned up.
- No rule handles residue after completion: merged worktrees, squash-merged branches, abandoned issues, background processes, caches, or scratch artifacts.
- Optional rules ship through the embedded registry and install with `kit rules add`. `context-evidence` set the pattern of a word budget and Go tests that run a published shell recipe.

## REQUIREMENTS

- Automatically clean completed task-owned worktrees, local branches, safe stale refs/registrations, never-authoritative scratch and unused reproducible cache entries. Preserve primary default-branch synchronization.
- Require current owner handoff, stable recorded target identity, completion evidence, live remote recovery of relevant commits, proof of local-state disposability, and quiescent cooperative ownership. Neither filenames, age nor a cached remote-tracking ref establishes proof.
- Keep unknown or unique state: ignored/untracked content, hidden tracked changes, special modes/attributes, local-only reflog commits/configuration, active operations, unavailable/failed checks and concurrent changes.
- Proof and mutation execute together, with a repository cleanup lease and immediate complete revalidation. All agent writers must respect the lease; process checks supplement exclusive handoff and cannot prove absence of arbitrary external writers.
- Distinguish recovery of source commits from disposability of other local state.
- Retain normal no-prompt cleanup under standing authority. Never force removal or weaken hard-delete confirmation.
- User explicitly approved a small native cleanup command, superseding the original no-new-command constraint.
- Existing shared/legacy lanes need trustworthy creation ownership and current handoff. Unsupported shared-cache adapters are reported and retained, not treated as generally disposable. Task-owned unused cache entries remain eligible.
- Browser-profile policy and Mint title changes are outside GH-262.

## ACCEPTED PLAN

1. Implement a small native cleanup command backed by typed proof/action code, with receipts recorded at task creation and current-owner handoff at completion.
2. Support category-specific actions, remote-default-only sync and compare-and-swap reference changes. Keep unsupported or uncertain targets with a reason.
3. Align the canonical rule, universal contract, all three rendered entrypoints, delivery/merge hooks, registry and command-surface documentation.
4. Test actual reclamation and restoration of eligible fixtures, plus preservation of unique/unknown state and injected concurrent writes.
5. Keep fresh init, historical/managed upgrades, local-custom preservation and repeated reconcile convergence covered.
6. Independently review against original intent; update existing PR263 and complete CI. Do not merge, release or clean user work directories.

## DECISIONS

- GH-262 revision: the initial blanket-retention PR design was rejected by the human because it removed automatic disk recovery. It is superseded. Safety findings require stronger eligibility, not universal inaction.
- Native cleanup is justified because local-state, ownership and concurrency proofs cannot safely remain a reusable classifier string followed by a separate shell deletion. The human approved this command-surface exception.
- The command trusts the human/agent's authorized owner handoff, not arbitrary filesystem names. Creation receipts bind exact filesystem and Git metadata identities; a cleanup lease coordinates cooperating agents. Unknown external writers or unavailable process visibility require retention.
- Live remote refs protect commits; exact tracked-tree comparison and negative local-state checks protect files and metadata. Generated scratch/cache manifests identify never-authoritative task output; changing an output invalidates that proof.
- Scope matters: stale-ref deletion uses expected-old-value updates and retained recovery roots; stale worktree registration cleanup is scoped to the recorded target. Unbounded prune commands are not inherently safe.
- Historical decisions below remain evidence and are superseded where their recovery assumptions are insufficient.

- Classify by recoverability, not by action type. A worktree whose head a remote branch or pull-request ref holds, and whose only local residue is ignored content, can be removed without prompts because `deletion-safety` treats it as recoverable. This makes the aggressive behavior consistent with existing rules and avoids an exception to them.
- Squash merges: accept `refs/pull/*/head` as proof of retention, so `git branch -D` is allowed only in that verified case.
- Ignored real `.env*` files block removal (`KEEP_LOCAL_SECRETS`), but symlinked ones do not, because `delivery` links them from the primary checkout.
- Shared caches use a two-tier policy. Tool-marked unused entries are pruned on every sweep. Full regenerable caches are cleared only below 25% or 50 GiB free, so builds stay warm when space is not short. Docker volumes, early Git object pruning, trash, and transcripts are never pruned.
- "Clear evidence" needs two observations in the current session: remote-confirmed completion (or a human stop) and restorability (`RECOVERABLE`, or a verified remote copy of an artifact). With both, cleanup runs at once with no prompt or offer. With either missing or `UNAVAILABLE`, the item is kept and reported. This makes the automatic behavior depend on evidence, not on guesses from names or dates.
- The shipped rule is a bounded standing grant for reversible closes of lane-owned issues and pull requests. A project can narrow or remove it; reconcile keeps local edits. Slack stays read-only.
- Superseded: the first head shipped the rule as optional. The human chose default with automatic execution. The contract carries the automatic trigger because contextual rules are read only when a trigger applies.
- Historical, superseded by the human-approved native revision: rejected a `kit clean` command (conflicts with the harness-subtraction direction and needs host-specific cache knowledge); `git clean -X` for idle lanes (it removes ignored env files).

- Superseded (GH-253, after the first release candidate): the batched hard-delete confirmation, remote-branch deletion, `git branch -D`, full cache clears, and idle-lane dependency pruning. The human found approval prompts during cleanup concerning. Cleanup now never asks; it runs only `git worktree prune`, `git fetch --prune`, `git worktree remove`, `git branch -d`, the work's own process and scratch cleanup, and unused-entry cache prunes when disk space is low. It also closes (never deletes) lane-owned GitHub issues and pull requests with observed completion, because the GitHub UI can reopen them without a prompt to the human. Everything else is kept and reported in one line. The disk saving that was given up is small: the worktree directory is most of a lane's footprint, and branches cost almost nothing.

- GH-257: cleanup also fast-forwards the primary checkout's local default branch. Agents create lanes from `origin/<base>`, so nothing else updated the human's local default branch after a merge. This is the only write that the contract allows in the primary checkout. The sync recipe uses only `git merge --ff-only` (on the default branch with no tracked changes) or `git fetch <remote> <base>:<base>` (on another branch). Git refuses both when work would be lost. Any other state is kept and reported as a literal state. The recipe does not check for running processes in the primary checkout: the human's shell and editor are usually there, a fast-forward on a clean tree does not lose work, and Git's own lock stops a concurrent Git operation.
- Rejected (GH-257): `git pull` (it can merge or rebase, depending on configuration); `git branch -f` or `git reset` (they can drop local commits); syncing the primary checkout from `delivery` (the lane does not need it, and cleanup already runs at every merge).

## DISCOVERIES

- A `case` pattern without a leading `(` inside `$( )` fails in bash's POSIX mode. The recipe uses `(pattern)`.
- The classifier must not `cd` into the worktree, because its own shell would then show as a process in that worktree. It uses `git -C` and a subshell `pwd -P`.

## VALIDATION

- GH-262 native revision: full `go test ./...`, `go vet ./...`, `golangci-lint run ./...` (0 issues), formatting, `go test -race ./internal/cleanup`, and Linux/Windows cleanup cross-compilation PASS. Native cleanup positives cover worktree removal/remote restoration, branch CAS deletion, stale refs/registrations, scratch/cache bytes, nested task caches, recorded environment links and both primary sync states. Negatives cover hidden/index state, ignored/untracked work, reflogs/config/operations, filesystem metadata, ownership/identity, failed/malformed process evidence, concurrent writes, live defaults, checkout filters and configured fetch pruning.
- Strict CLI fixtures reject target overrides, malformed/duplicate/trailing/oversized receipts and missing disposable declarations. Real host process visibility retains an active target and automatically reclaims idle task output; unnamed kernel pipes/network-policy descriptors do not disable the category. Streaming tests verify a 64 MiB sparse file with bounded allocation and final-byte mutation detection.
- Scoped reconciliation of the cleanup rule, delivery/merge hooks and all rendered entrypoints converges with zero planned changes. Earlier validation entries describe the superseded retention-only head.

- GH-262: targeted cleanup, contract, rendered-output and rule-invariant tests PASS.
- GH-262: gofmt (no output), make vet, make lint (0 issues), go test ./... (all packages PASS), make build and the built binary's check --project PASS. The references-index compatibility warning remains non-blocking.
- GH-262: real scoped reconciliation updates the registry; a subsequent dry-run --diff reports zero planned changes. Integration fixtures cover repeated convergence for fresh/historical/managed-v3.0.44/custom installations.
- Earlier validation below records the superseded policy, not proof of current cleanup safety.

- `go test ./pkg/cli -run TestWorkCleanup -count=1`: pass, with `sh` as macOS bash and as `dash`.
- `gofmt -l .` (empty), `make vet`, `make lint` (0 issues), `make test` (all packages ok), `make build`: pass.
- `kit check --project`: pass with one advisory warning about `docs/references/README.md` that also appears on unchanged `main`.
- The recipe also parses with `dash -n`.

## OUTCOME

- GH-262 native revision restores automatic disk recovery through `kit cleanup`, with creation receipts, current-owner handoff, live recovery evidence, complete local-state checks, repeated proof under a cooperative lease and bounded native actions. Unknown evidence retains only the affected target; eligible artifacts proceed without routine prompts.
- Deliberate limits: legacy/shared artifacts without creation ownership remain in place; shared-tool cache pruning needs a bounded adapter. Unsupported filesystem/ref backends and unknown external writers fail closed. macOS OS provenance is the sole documented metadata exception. A native failure may be partial and is reported honestly.
- Independent read-only review completed with no remaining concrete blocker after corrections. Hosted CI/review on the revised head remains to be observed. No merge, release, downstream rollout or cleanup of user work directories is authorized.

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
