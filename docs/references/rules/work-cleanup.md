---
kind: ruleset
slug: work-cleanup
description: Inventories completed work without deleting local worktrees, branches, caches, or primary-checkout state.
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

Preserve local work when a task ends. A remote commit restores committed source;
it does not restore ignored files, reflog-only commits, worktree configuration,
filesystem modes, index flags, or in-progress Git operations.

## Applies When

- A pull request merges or closes, the lane issue closes, or the human stops work.
- Before creating a lane or ending a session.
- Disk space is low or a tool reports a full disk.

## Rules

### Automatic completion

- Inventory the current task's lanes with read-only Git commands. Keep every
  worktree and branch as `KEEP_LOCAL_STATE`; a clean status or a live remote head
  is insufficient evidence that all local state is recoverable.
- Do not automatically remove worktrees, delete branches, prune refs or caches,
  remove scratch data, or update the primary checkout. Leave other tasks alone.
- Report the existing validation evidence by path. Write task-owned evidence
  during authorized validation; do not move or copy local files as cleanup.
- Stop only background watchers and checks that this task started and that have
  no outstanding work. Do not stop user processes or other tasks' processes.
- Close this lane's still-open issue as completed only after observing that its
  merged change satisfies the issue. Do not close unrelated issues or tasks.

### Explicit removal requests

- A separate human request must identify the worktrees, branches, or other local
  targets to remove. Follow `deletion-safety` and `delivery` for that request.
- Before removal, verify completion, absence of active users and Git operations,
  and recovery of the complete relevant local state in an approved location.
  Include ignored and untracked data, hidden tracked edits, local Git metadata,
  reflogs and filesystem state. A remote SHA is not a substitute for that proof.
- If any state or recovery evidence is unknown, retain the target without asking
  a routine cleanup question. Report the precise missing evidence.
- Never force removal, reset or clean a checkout, rewrite history, or delete a
  branch to clear a lane. This rule grants no destructive authority.

## Retention classifier

Set `worktree_path`. This read-only recipe never authorizes removal.
`UNAVAILABLE` means the path or Git evidence could not be read; keep that target.
The recipe needs no remote, `lsof`, or inference that ignored data is disposable.

```sh
top=$(cd -- "$worktree_path" 2>/dev/null && pwd -P) || { printf '%s\n' UNAVAILABLE; exit 0; }
inside=$(git -C "$top" rev-parse --is-inside-work-tree 2>/dev/null) || { printf '%s\n' UNAVAILABLE; exit 0; }
if [ "$inside" != true ]; then printf '%s\n' UNAVAILABLE; exit 0; fi
git -C "$top" rev-parse --verify HEAD >/dev/null 2>&1 || { printf '%s\n' UNAVAILABLE; exit 0; }
printf '%s\n' KEEP_LOCAL_STATE
```

## Report

Add one short cleanup line to the normal final response: the number of current
lane worktrees retained as `KEEP_LOCAL_STATE`, whether anything was removed, and
that primary checkouts were preserved. Do not imply that retained local state is
backed up, and do not turn retention into a permission question.
