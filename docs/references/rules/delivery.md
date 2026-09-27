---
kind: ruleset
slug: delivery
description: Lands every agent change through an isolated worktree and a human-owned pull request without asking for a lane choice.
status: active
registry_scope: downstream
applies_to:
  - git
  - github
  - worktree
  - pull-request
  - attribution
read_policy_default: conditional
---

# Ruleset: delivery

## Purpose

- Keep the human's primary checkout and history safe while agents work autonomously.
- Make every change reviewable in a human-owned pull request.
- Keep authorship, assignment, and commit history attributable to the human.

## Applies When

- Before the first issue, branch, worktree, file edit, staging, commit, push, or pull-request action for accepted work.
- When repairing an existing pull request (review feedback, CI, base refresh, conflicts).
- Merging is governed by `github-pr-merge`.

## Rules

### Lane

- Default to a new lane without asking: reuse one strongly matching open issue or create one assigned to the human, then use branch `GH-<issue>` in worktree `~/worktrees/<owner>/<repository>/GH-<issue>`, and deliver through one ready pull request.
- Continue an existing branch or pull request only when the user explicitly names it for the same work. Repair of an exact existing pull request always reuses that pull request's head branch; never open a coordinating or corrective pull request for scope-preserving work.
- Never ask the user to choose between lanes. Ask only when implementation intent or a named target is materially ambiguous. Materially new scope gets its own lane.

### Worktrees

- Discover the default branch and fetch it; never assume `main`. Create the lane with `git worktree add -b GH-<issue> <path> origin/<base>`, or reuse the worktree already registered for that branch.
- Never create worktrees inside a repository. Use a detached `PR-<number>` worktree only for read-only inspection.
- Link the primary checkout's `.env` and `.envrc` into a writable worktree as symlinks when they exist; never copy or overwrite environment files.
- Never stash, reset, clean, force-remove, or delete branches to make a worktree operation succeed. Refs, remotes, configuration, and stash are shared across worktrees.

### Primary Checkout

- The primary checkout is read-only for agent work. If it holds changes you did not make through a lane, preserve and report them.
- Exception: files a Kit command wrote to the primary checkout and that were delivered through a merged worktree pull request may be removed so the checkout can pull. Enumerate untracked candidates first, remove only verified command-owned paths, restore only tracked paths whose content still matches what the command wrote, and stop on any mismatch.

### Identity And Attribution

- Before the first commit, confirm `git config user.name`/`user.email` and the authenticated `gh` account are the human's; stop otherwise.
- Never add co-author trailers, "generated with" footers, agent signatures, or bot assignees to commits, issues, pull requests, or comments. Assign every created or reused issue and pull request to the human (`--assignee @me`).

### Commits And Pull Requests

- Stage explicit paths only (never `git add -A` or `git add .`), review the staged diff against the ask, and scan it for secrets before committing. Never commit `.env` files or secret values.
- Title format for commits and pull requests: `<type>(GH-<issue>): <gitmoji> <short title>` with `feat :sparkles:`, `fix :bug:`, `docs :memo:`, `test :white_check_mark:`, `refactor :recycle:`, `chore :wrench:`, or `ci :green_heart:`. Commit bodies are short bullet lists.
- Issues record the original ask, scope boundaries, acceptance criteria, and expected verification.
- Open pull requests ready for review, against the protected default branch, using the repository's pull-request template, and close the issue from the body.
- Never commit to the default or a protected branch, force-push, or rewrite pushed history.
- After pushing, confirm the remote head matches the local commit and report the issue, branch, and pull request.

### Recovery

- Diagnose failures from current evidence and recover without destructive commands. Missing credentials, an unexpected identity, or an ambiguous target is a blocker: report the smallest input needed.

## Examples

```bash
git fetch origin main
git worktree add -b GH-123 ~/worktrees/acme/api/GH-123 origin/main
```

```text
fix(GH-123): :bug: reject expired session tokens
```
