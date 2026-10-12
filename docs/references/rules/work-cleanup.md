---
kind: ruleset
slug: work-cleanup
description: Quietly removes completed, remote-restorable worktrees, merged branches, and the work's own scratch, and never asks the human to approve a cleanup.
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

- Recover local disk space after work completes, without adding approvals, prompts, or alarming commands to the human's day.
- Remove only what a remote can clearly restore. Keep everything else, and mention it once in the report.

## Applies When

Read this rule and run a sweep, without being asked, when one of these triggers occurs:

- A pull request merges or closes, an issue or ticket closes, or the human says stop, abandon, or superseded.
- Before you create a new lane, and before you end a session.
- Free space is below 25% or below 50 GiB, or a tool reports a full disk.

## Rules

### Never ask

- Cleanup never asks the human to approve, confirm, or choose anything. If an action needs approval, or the evidence is not clear, do not do it: keep the item and list it in the report.
- Use only the commands in this rule. Do not use Git `--force` or `-D`, `rm -rf`, `sudo`, or any purge, reset, or clean command for cleanup.
- Hard deletes stay with `deletion-safety`, and only when the human asks for them. This rule never weakens `deletion-safety`, `delivery`, or `infrastructure-change-approval`.
- Never touch the primary checkout, the current session's working directory, or a worktree that a running process uses. Never delete a branch, stash, or worktree to unblock a new lane.
- Kit ships this rule as the human's bounded standing grant for the actions below. A human pause, hold, or project edit that removes this rule stops them.

### Clear evidence

- Act only when you observe both parts in the current session:
  1. Completion: the remote shows the pull request as merged or closed, or the issue as closed; or the human said stop, abandon, or superseded; or the lane is idle for 7 days with no open pull request.
  2. Restorability: the classifier below prints `RECOVERABLE`.
- Never infer completion or restorability from names, dates alone, or memory of an earlier session.

### Automatic actions

- Run `git worktree prune` and `git fetch --prune`. Both only remove stale references.
- For each completed worktree that is `RECOVERABLE`, run `git worktree remove <path>`. Then run `git branch -d <branch>`. If Git refuses, keep the branch; a branch uses almost no disk space.
- Stop background processes, dev servers, watchers, and scheduled checks that the work started. Remove scratch files and directories that the work created in its own scratch location.
- Close GitHub issues and pull requests whose completion you observed. Close, never delete; the GitHub UI can reopen them.
  - Done: close a still-open lane issue as completed when the merged change satisfies it (`gh issue close <n> --reason completed`).
  - Abandoned or superseded: close the lane's open pull request and issue with one line of reason and a link to the superseding work, if any (`gh pr close <n> --comment "<reason>"` and `gh issue close <n> --reason "not planned" --comment "<reason>"`).
  - Close only issues and pull requests that this lane created or that the human assigned and declared finished. Never post to Slack.
- When free space is below the threshold, also run the unused-entry prune of each installed tool: `docker builder prune -f`, `docker image prune -f` (dangling images only), `pnpm store prune`, `uv cache prune`, and `brew cleanup`. Do not clear whole caches.

### Leave alone

- `KEEP_IN_USE`, `KEEP_UNCOMMITTED`, `KEEP_LOCAL_SECRETS`, `KEEP_UNPUSHED`, and `UNAVAILABLE` worktrees; orphan directories; stashes.
- Remote branches and external tracker tickets. GitHub can delete merged branches.
- Published artifacts, shared documents, containers, volumes, cloud resources, transcripts, the trash, and other people's files.

### Report

- At the end of your normal response, add one short cleanup line: what you removed, the space recovered (`df -h` before and after), and the count of kept items with their states. Example: `Cleanup: removed 2 worktrees (GH-133, GH-140), freed 1.2 GiB; kept 1 (GH-202: KEEP_UNCOMMITTED).`
- Do not turn kept items into a question. The human can ask for them by name.

## Worktree classifier

Set `worktree_path`, and optionally `remote` (default `origin`). The recipe prints one state. It changes only remote-tracking refs.

```sh
remote=${remote:-origin}
top=$(cd -- "$worktree_path" 2>/dev/null && pwd -P) || { printf '%s\n' UNAVAILABLE; exit 0; }
head=$(git -C "$top" rev-parse --verify HEAD 2>/dev/null) || { printf '%s\n' UNAVAILABLE; exit 0; }
if command -v lsof >/dev/null 2>&1 &&
  lsof -d cwd -Fn 2>/dev/null | awk -v p="n$top" '$0 == p || index($0, p "/") == 1 { f = 1 } END { exit !f }'; then
  printf '%s\n' KEEP_IN_USE; exit 0
fi
if [ -n "$(git -C "$top" status --porcelain --untracked-files=all)" ]; then
  printf '%s\n' KEEP_UNCOMMITTED; exit 0
fi
secrets=$(git -C "$top" ls-files --others --ignored --exclude-standard --directory |
  while IFS= read -r f; do
    case $f in (.env*|*/.env*) [ -L "$top/${f%/}" ] || printf '%s\n' "$f" ;; esac
  done)
if [ -n "$secrets" ]; then printf '%s\n' KEEP_LOCAL_SECRETS; exit 0; fi
git -C "$top" fetch --quiet "$remote" 2>/dev/null || { printf '%s\n' UNAVAILABLE; exit 0; }
if [ -n "$(git -C "$top" branch -r --contains "$head" 2>/dev/null)" ] ||
  git -C "$top" ls-remote "$remote" 'refs/pull/*/head' | awk -v h="$head" '$1 == h { f = 1 } END { exit !f }'; then
  printf '%s\n' RECOVERABLE
else
  printf '%s\n' KEEP_UNPUSHED
fi
```

`RECOVERABLE` means that a remote branch or pull request holds the head commit, and the only local residue is ignored, regenerable content. It does not mean the work is completed. Check completion separately.
