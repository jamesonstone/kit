---
kind: ruleset
slug: work-cleanup
description: Aggressively prunes worktrees, branches, caches, processes, and artifacts as soon as a piece of work is done, stopped, abandoned, or superseded.
status: active
registry_scope: optional
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

- Recover local disk space early and often. Leave nothing behind for a piece of work that has no value to resume.
- Prune without prompts when nothing unrecoverable is lost. Stop and confirm only when unique work would be lost.

## Applies When

Install with `kit rules add work-cleanup`. Once it is installed, read this rule when one of these sweep triggers occurs:

- A piece of work completes: a pull request merges or closes, an issue or ticket closes, a release ships, or an artifact is delivered.
- The human says stop, drop, abandon, never mind, or start over, or newer work supersedes the current work.
- Before you create a new lane, and before you end a session.
- Free space is below 25% or below 50 GiB, or a tool reports a full disk.

"Piece of work" means any unit with local or remote residue: lane, branch, worktree, issue, ticket, pull request, scratch directory, download, report, screenshot, log, container, background process, or scheduled check. "Completed" means done, stopped, abandoned, superseded, or idle for 7 days with no open pull request.

## Rules

### Boundaries

- This rule never weakens `deletion-safety`, `delivery`, or `infrastructure-change-approval`. It classifies each cleanup as recoverable, which proceeds now, or as a hard delete, which needs exact-target confirmation.
- Installing this rule is the human's bounded standing grant for the recoverable actions below. A human pause or hold stops them.
- Cleanup follows completion. Never delete a branch, stash, or worktree to unblock a new lane, and never use `--force` on any Git command.
- Never touch the primary checkout, the current session's working directory, or a worktree that a running process uses.

### Sweep worktrees and branches

- Sweep every registered worktree of the repository (`git worktree list --porcelain`), including host-created worktrees. Run `git worktree prune` and `git fetch --prune` first.
- Classify each candidate with the recipe below. Act on its state:
  - `RECOVERABLE` and completed: run `git worktree remove <path>`. Then delete the local branch with `git branch -d`. Use `git branch -D` only when the state is `RECOVERABLE` and a merged or closed pull request holds the exact head commit (squash merges).
  - `RECOVERABLE` and the remote branch belongs to a merged or closed pull request: delete the remote branch. The pull request page can restore it.
  - `KEEP_IN_USE`: skip it, and report it.
  - `KEEP_UNCOMMITTED`, `KEEP_LOCAL_SECRETS`, or `KEEP_UNPUSHED`: this is a hard delete. Put it in one confirmation list with the path, branch, head commit, and what would be lost.
  - `UNAVAILABLE`: keep it, and report why.
- Idle open work (7 days with no commits and an open pull request) is not completed. Remove only its regenerable dependency and build directories, such as `node_modules`, `.venv`, `target`, `dist`, `build`, and `.next`. Never run `git clean -X`, because it also removes ignored environment files.
- Directories under the worktree root that no repository registers are orphans. Report them for confirmation. Do not delete them.

### Close issues, tickets, and pull requests

- When work is done, close a still-open lane issue as completed if the merged change satisfies it.
- When work is abandoned or superseded, close the lane's pull request and issue as "not planned" with one line of reason and a link to the superseding work. Change external tracker tickets to their cancelled or done state.
- Close only work that this lane created or that the human assigned and declared finished. Close, never delete. Never post to Slack.

### Prune artifacts and processes

- Stop background processes, dev servers, watchers, port forwards, monitors, and scheduled wakeups or cron jobs that the work started.
- Remove scratch directories, temporary files, downloads, logs, screenshots, and local reports that the work created, after you deliver or record what matters from them.
- Remove containers, images, networks, and test-only volumes that the work created and labeled, if they hold no authoritative data.
- Published artifacts, shared documents, cloud preview environments, and other remote resources: deleting them is a hard delete or an infrastructure change. List them for confirmation.

### Prune shared caches

- On every sweep, remove the entries that each tool marks as unused: `docker builder prune -f`, `docker image prune -f`, `pnpm store prune`, `uv cache prune`, and `brew cleanup`.
- When free space is below the threshold, also clear caches that tools can fully regenerate: `go clean -cache -testcache`, `npm cache clean --force`, `pip cache purge`, `yarn cache clean`, Xcode `DerivedData`, and `xcrun simctl delete unavailable`. Clear `go clean -modcache` last.
- Use each tool's own prune command. Do not use `rm -rf` on a cache that a running tool can use.
- Never run `docker volume prune` or `docker system prune --volumes`, never prune unreachable Git objects or reflogs early, never empty the trash, and never delete transcripts, home-directory data, or other people's caches.

### Report

- Measure free space before and after (`df -h`). Report the space you recovered.
- For each removed item, report how to recover it: branch, full head commit, and pull request. Report each kept item and its state. Ask once for confirmation of the complete hard-delete list.

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

`RECOVERABLE` means that a remote branch or pull request holds the head commit, and the only local residue is ignored, regenerable content. It does not mean the work is completed. Check completion separately. Stashes are shared by all worktrees and are not removed by this recipe.
