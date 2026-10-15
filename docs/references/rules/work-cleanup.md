---
kind: ruleset
slug: work-cleanup
description: Automatically reclaims completed, proven-recoverable task artifacts with native ownership, local-state and concurrency checks.
status: active
registry_scope: downstream
applies_to:
  - worktree
  - git
  - github
  - cache
  - filesystem
  - artifacts
  - cleanup
read_policy_default: conditional
---

# Ruleset: work-cleanup

## Purpose

Recover disk space automatically after completed work without losing unique work.
A remote commit proves source recovery, not recovery of every local file or Git state.
Use the native `kit cleanup` proof-and-action command; never use a classifier result
as permission for a later shell deletion.

## Applies When

- Before creating a lane; after its pull request merges or closes or the human stops it.
- Before session end or a new lane, and when free space is below 25% or 50 GiB.

## Rules

### Ownership and handoff

- Record exact task-owned artifacts when the task creates them. Never infer ownership
  from a branch name, age, clean status or directory location. Do not adopt other tasks'
  or legacy/shared artifacts merely to make them eligible.
- Save receipts outside their targets. They bind task owner, physical directory and
  Git metadata identities. Record scratch/cache output immediately after its task-owned
  producer finishes, before any later edits; `--disposable` explicitly declares that
  the task produced it and it never became authoritative input, evidence or user work.
- At completion, the same owner must hand off exclusive control after all writers stop.
  Supply the exact completed source head and current merged/closed PR, or the reference
  to a current human stop instruction. Do not fabricate a stop instruction.
- `--writers-stopped` is this owner handoff, not a claim that process inspection alone
  proves exclusivity. If ownership, process visibility, external writers or recovery is
  unknown, retain the target. Running a command from another directory is not a handoff.
- Run cleanup from outside the target and its Git metadata. Cooperating agents must not
  write while `<git-common-dir>/kit-cleanup.lock` exists. Never remove another owner's
  lease or infer that it is stale from its age. The lease is not an OS sandbox for
  arbitrary external writers.

### Automatic categories

| Category | Automatic action and proof |
| --- | --- |
| Linked worktree | Remove only the exact recorded registration after fresh completion, live selected-remote exact-head recovery, physical tracked-byte/mode comparison, ordinary index state and reconstructible Git metadata. Ignored/untracked data, nested repos/submodules, hidden edits, special index flags, local-only reflog commits, local configuration and active operations retain it. |
| Local branch | Protect the live remote default. When checked out nowhere, verify source and reflog commit recovery, then delete the exact ref with an expected-old-value update. Keep branch configuration rather than discarding it; custom configuration blocks eligibility. Never force or delete to unblock a lane. |
| Stale references | Prune only an exact recorded selected-remote tracking ref that is absent remotely and whose ref/reflog commits remain reachable through retained local branches or tags. No blanket `fetch --prune`. Remove a missing task worktree's registration only after the same metadata/recovery proofs; no blanket `worktree prune`. |
| Task scratch | Remove unchanged, recorded, never-authoritative task output after its producer and consumers stop. A scratch-looking filename is insufficient. Changed data, nested Git state or overlaps with primary/tracked state retain it. |
| Unused caches | Reclaim unchanged recorded task-produced disposable cache entries, including untracked output inside its linked lane. Clean those proven entries before re-assessing their containing worktree. Shared-tool pruning needs an exact supported recovery/disposability adapter; no unbounded Docker, Brew, pnpm or uv prune commands are authorized by their names alone. Report unsupported shared selections as `KEEP_UNSUPPORTED`; do not disable eligible task-cache recovery. |
| Primary sync | Keep the live remote default branch current under a bounded sync receipt and exclusive handoff. Reject arbitrary bases, divergence, unknown state, active Git operations/users and defaults checked out elsewhere. On the default branch, compare physical state and effective incoming checkout filters, then use `--ff-only --no-overwrite-ignore`. On another branch, use an ancestor-checked expected-old-value ref update without touching working files. |

- Physical recovery proof rejects nonstandard permissions, foreign ownership, hardlinks,
  user xattrs, ACLs and file flags. On macOS, OS-generated `com.apple.provenance` is an
  explicit exception; its bytes are fingerprinted for changes, and file restoration
  regenerates OS provenance rather than preserving that token. Other attributes remain protected.
- Creation-recorded `.env`/`.envrc` symlinks may be removed with their worktree only
  when their exact primary-checkout sources remain regular files. Never copy secret bytes.
- `lsof`, Git or remote failures, malformed evidence and unsupported filesystem inspection
  return `UNAVAILABLE` and retain the target. A cached remote-tracking ref never proves recovery.
- A native operation error returns `REFUSED_OR_PARTIAL`; inspect actual state and retained
  recovery evidence. Do not claim that a failed removal left every entry in place.
- The command repeats all proofs immediately before mutation under its cleanup lease.
  Changed identity, contents, refs, completion or recovery evidence stops that action.
- Stop only task-owned idle watchers/checks with no outstanding work. Close this lane's
  satisfied issue after observing its merged result. Leave unrelated work and Slack alone.
- Cleanup never asks for routine approval. Clear evidence under this standing grant
  permits automatic action; missing evidence means keep and report. Hard deletes stay
  under `deletion-safety`; never force removal, reset, clean, purge or bypass its confirmation.

## Use

Record a new task lane after creation and environment linking:

```sh
kit cleanup --record --kind worktree --path "$lane" --repository "$primary" --owner "$task_id" > "$receipt"
```

After current-owner handoff, run from outside that lane:

```sh
kit cleanup --receipt "$receipt" --owner "$task_id" --head "$completed_head" --pr "$pr" --writers-stopped --apply
```

For never-authoritative generated scratch/cache, use `--record --kind scratch|cache
--disposable` after the producer finishes. For branch/ref receipts use the repository
as `--path` and an exact `--ref refs/heads/...` or `refs/remotes/origin/...`.
Use `--kind primary` to record the bounded default-sync target; an optional `--ref`
must match the live remote default. A current human stop can replace `--pr` with
`--stop-evidence "<current instruction reference>"`.

Omit `--apply` for a read-only preview. `ELIGIBLE` is not reusable removal authority.
`REMOVED`, `PRUNED`, `UPDATED` and `CURRENT` describe actual outcomes.
`KEEP_LOCAL_STATE`, other `KEEP_*` states and `UNAVAILABLE` retain the target and explain
the missing evidence. Unsupported legacy ownership or shared-cache adapters are
specific limitations, not a reason to retain every artifact.

## Report

Add one short line: actual removals/prunes, logical file bytes removed (not an invented
disk-space saving), primary sync result and kept states. Retention does not imply a
backup. Never turn kept items into a routine permission question.
