---
kind: ruleset
slug: work-cleanup
description: Automatically removes completed task-owned work that is verifiably recoverable, keeps unknown local state, and safely fast-forwards the primary default branch.
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
A remote source commit does not back up ignored files or local Git state.

## Applies When

- After a pull request merges or closes, or the human stops or supersedes the work.
- Before a new lane or session end, or when free space is below 25% or 50 GiB.

## Rules

### Standing authority

- Cleanup never asks for routine approval. Automatically remove verifiably recoverable
  task-owned cruft; keep unique or unknown state and report it once.
- Use current task evidence to identify exact targets. Never infer ownership or completion
  from a name, directory, age, clean status, or an idle-time threshold. Leave unrelated work alone.
- Observe the task's exact source head and merged/closed pull request in the selected remote,
  or the current human stop instruction. A closed issue alone does not prove completion.
- Work from outside the target. Stop only task-owned idle watchers/checks with no outstanding
  work; establish that other writers have stopped. Unknown ownership or concurrent use means keep.
- This grant covers only the recoverable actions below. Never force, reset, clean, purge,
  remove stashes, or delete a branch to unblock a lane. Hard deletes remain under `deletion-safety`.

### Completed worktrees

Before removing a linked worktree, verify all of the following in the current session:

1. It is the exact completed task's registered linked worktree, never the primary checkout
   or current working directory. Its current head matches the completed source head.
2. Process inspection succeeds and shows no open files or working directories in the lane
   or its Git metadata. A missing, failed, partial or inconclusive `lsof` check is `UNAVAILABLE`,
   not evidence that the lane is idle.
3. Inventory tracked, untracked and ignored paths, including hidden files. Use
   `git status --porcelain=v1 --untracked-files=all --ignored=matching` and check each command's
   exit status. Ignored/untracked data is local state, not automatically regenerable content.
   First remove only task-created scratch/cache output you know never became authoritative.
   Keep any remaining local data. Exact task-created `.env`/`.envrc` symlinks may be unlinked
   only when their original primary-checkout regular-file sources still exist; never read,
   copy or remove the source bytes.
4. Check the physical tracked contents and modes against the commit, not just cached status.
   Special index flags from `git ls-files -v -f`, checkout/clean filters or other state that
   hides changes require separate verification; otherwise keep. Preserve custom filesystem
   modes/attributes, hard links, local configuration (`config.worktree` included), unfinished
   Git operations and commit drafts. Ordinary Git-generated bookkeeping is not user data.
5. Verify recovery against live refs with `git ls-remote` on the selected remote, including
   `refs/pull/*/head` for squash merges. A live ref must retain the exact source head;
   cached tracking refs and a prior fetch are insufficient. Check the lane's HEAD/branch
   reflogs and ORIG_HEAD for local-only commits too; unrelated primary HEAD history is not
   part of the lane's recovery proof. Disable replacement history with `GIT_NO_REPLACE_OBJECTS=1`;
   local grafts or unsupported history/ref inspection make recovery unknown.

If any proof fails, keep the lane as `KEEP_LOCAL_STATE`, `KEEP_IN_USE`, `KEEP_RECOVERY`
or `UNAVAILABLE`. These states do not disable cleanup of other eligible targets.
Immediately before acting, recheck the target, head, local state, process use and live
recovery evidence. Do not save a reusable `RECOVERABLE` classification for a later deletion.

When the checks pass, run the normal removal; if Git refuses, keep and report, without a fallback:

```sh
git -C "$primary" -c core.hooksPath=/dev/null -c core.fsmonitor=false worktree remove -- "$lane"
```

After removal, an exact task branch may also be deleted if checked out nowhere, is not
any remote's default branch, and its head and branch-reflog commits have verified recovery.
Use `git -C "$primary" update-ref -d "refs/heads/$branch" "$verified_head"` to refuse a changed head
and leave branch configuration in place. Never force branch deletion.

### Other cruft

- Remove exact task-created, never-authoritative scratch and reproducible cache output
  after producers and consumers stop. A scratch/cache name or ignored status is not proof.
- Prune an exact stale task tracking ref only after verifying its remote branch is absent
  and its ref/reflog commits remain reachable through retained local branches or tags.
  Remove a missing task worktree registration only after verifying its remaining Git state.
  No blanket `fetch --prune` or `worktree prune`: stale-looking refs can be the only recovery copy.
- Shared-tool unused-cache cleanup needs proof for the tool's actual selected entries:
  reproducibility, no unique data and no active users. Tool names or an unused label alone
  do not authorize broad Docker, Brew, pnpm or uv pruning. Keep unverified selections.
- Close a lane issue only when its merged change satisfies it. Leave other people's work,
  shared documents, containers, volumes, cloud state, transcripts, trash and Slack alone.

### Primary default-branch sync

- Discover the live selected remote's default branch; do not accept an arbitrary base.
  Fetch its objects without pruning, updating tracking refs, tags, recursive submodules or
  automatic maintenance (`--no-prune --no-prune-tags --refmap= --no-tags
  --no-write-fetch-head --recurse-submodules=no --no-auto-maintenance`).
- Require stopped writers, available process inspection, no active Git operation, and no
  unique state in the ref/control files being updated. Verify the local default is an
  ancestor of the live remote head. Disable replacement history and reject local grafts; unknown evidence means keep.
- If the primary is on the default branch, require ordinary index state, verified tracked
  bytes/modes, no effective checkout filters and no ignored/untracked collision. Then run:

```sh
git -C "$primary" -c core.hooksPath=/dev/null -c core.fsmonitor=false merge --ff-only --no-overwrite-ignore "$remote_head"
```

- If it is on another branch, leave working files untouched. Verify the default is checked
  out nowhere else, then use `git -C "$primary" update-ref "refs/heads/$base" "$remote_head" "$local_head"`.
- Recheck immediately before the action. Never switch branches, reset, stash, clean, rebase
  or create a merge commit. A refusal means keep, not permission to repair the checkout.

## Report

Add one short cleanup line: actual removals, primary sync (`UPDATED`, `CURRENT` or a kept
state), and kept targets with reasons. Report measured disk savings only when measured.
A failed operation can be partial; inspect actual state before claiming anything was kept.
Never turn retained items into a routine permission question.
