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

- "Piece of work" covers lanes, branches, worktrees, issues, tickets, pull requests, and any artifact. "Completed" covers done, stopped, abandoned, superseded, and idle with no open pull request.
- Aggressive by default: prune at every completion, stop, new lane, session end, and low-disk event, with no prompts when the result can be recovered or regenerated.
- Unique work (uncommitted, unpushed, real ignored env files, orphans, remote published resources) goes into one batched hard-delete confirmation.
- No new command, no default installation, no change to the universal contract.

## ACCEPTED PLAN

1. Ship `work-cleanup` as an optional, conditional embedded rule with a read-only worktree classifier recipe.
2. List it in the references index.
3. Test opt-in installation, safety invariants, default footprint, and every classifier state in disposable repositories.

## DECISIONS

- Classify by recoverability, not by action type. A worktree whose head a remote branch or pull-request ref holds, and whose only local residue is ignored content, can be removed without prompts because `deletion-safety` treats it as recoverable. This makes the aggressive behavior consistent with existing rules and avoids an exception to them.
- Squash merges: accept `refs/pull/*/head` as proof of retention, so `git branch -D` is allowed only in that verified case.
- Ignored real `.env*` files block removal (`KEEP_LOCAL_SECRETS`), but symlinked ones do not, because `delivery` links them from the primary checkout.
- Shared caches use a two-tier policy. Tool-marked unused entries are pruned on every sweep. Full regenerable caches are cleared only below 25% or 50 GiB free, so builds stay warm when space is not short. Docker volumes, early Git object pruning, trash, and transcripts are never pruned.
- Installing the rule is a bounded standing grant for reversible closes of lane-owned issues and pull requests. Slack stays read-only.
- Optional, not default: the aggressive cache policy is a workstation preference, not a project invariant.
- Rejected: a `kit clean` command (conflicts with the harness-subtraction direction and needs host-specific cache knowledge); `git clean -X` for idle lanes (it removes ignored env files).

## DISCOVERIES

- A `case` pattern without a leading `(` inside `$( )` fails in bash's POSIX mode. The recipe uses `(pattern)`.
- The classifier must not `cd` into the worktree, because its own shell would then show as a process in that worktree. It uses `git -C` and a subshell `pwd -P`.

## VALIDATION

- `go test ./pkg/cli -run TestWorkCleanup -count=1`: pass, with `sh` as macOS bash and as `dash`.
- `gofmt -l .` (empty), `make vet`, `make lint` (0 issues), `make test` (all packages ok), `make build`: pass.
- `kit check --project`: pass with one advisory warning about `docs/references/README.md` that also appears on unchanged `main`.
- The recipe also parses with `dash -n`.

## OUTCOME

- Shipped `docs/references/rules/work-cleanup.md` and tests in `pkg/cli/work_cleanup_*_test.go`.
- Risk: lsof-based in-use detection sees only processes visible to the user. Without lsof, the agent must check in-use state itself.
- Risk: cache command lists name common tools only. Projects can add local entries; reconcile preserves local edits.

## REPOSITORY MEMORY

- Added `docs/references/rules/work-cleanup.md`.
- Updated `docs/references/README.md` index and usage note.
- Unchanged: universal contract, `deletion-safety`, `delivery`; the new rule defers to them.
