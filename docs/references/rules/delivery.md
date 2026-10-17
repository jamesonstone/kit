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

### Issue provider and ticket parity

- Configure with `kit config set --issue-provider github|linear` (Linear target flags match init), preview with `--dry-run --json`, and execute the reported canonical reconcile steps in the reported target. The setter changes only issue configuration; it does not migrate managed files or authenticate.
- Read repository-local `.kit.yaml` `issues.provider`: omitted or `github` uses the existing GitHub Issues flow; `linear` uses Linear only. Provider selection does not change GitHub pull requests, reviews, protection, or release hosting. Never inherit an issue target from global configuration.
- For GitHub, use the repository's existing `gh` or connected GitHub tools and human assignment. For Linear, use the host's already-authorized Linear connector or API capability. If unavailable, stop ticket writes and report the missing Linear capability; do not install a connector, configure credentials, broaden access, or fall back to GitHub.
- Before a Linear write, verify the authenticated human, workspace URL slug (`issues.linear.workspace`), team UUID (`team_id`), and optional project UUID (`project_id`). Read the team and project to confirm workspace membership and that the project supports that team. Resolve the human's Linear assignee ID; GitHub logins are not Linear assignee IDs. Missing configuration, permission errors, target mismatch, or ambiguous lookup is a blocker. Never choose the first name match. An omitted project means explicitly team-only work.
- Search the selected repository/team/project for strongly matching work, including closed tickets and linked PRs, before creating. Reuse one clearly matching open ticket; if matches conflict, stop and report the ambiguity. Do not reopen completed work or create duplicates after an uncertain write: first read back by returned ID or search the exact title, target, human, and repository link. Retry only after confirming no ticket was created. A failed search is not an empty search.
- Both providers get the same title and Markdown substance: original ask, repository/spec context, scope and exclusions, acceptance criteria, expected verification, and cross-repository dependencies when relevant. Preserve links and structure; never summarize away requirements for Linear. For Linear send the title and complete Markdown as `description`, plus the verified team ID, human assignee ID, and optional project ID. Read back the returned identifier, canonical URL, assignee, team, project, and description before creating the branch.
- Provider errors (authentication, permissions, transport, rate limits, GraphQL errors, unsuccessful mutation, malformed response) stop the ticket action. Report the exact failed action/target and whether denied, pending, or uncertain. Never infer success from HTTP status alone. Continue independent local work when safe; never invent an identifier.
- Keep lifecycle intent equal: record progress and PR links on the selected ticket; mark it completed only after observing a satisfying merge. Use the team's verified workflow states for Linear, not guessed state names or GitHub closing keywords. Abandoned/superseded work gets a reason and replacement link, with the appropriate verified canceled state. No automatic state transition is implied by opening a PR.
- Linear identifiers and URLs can reveal private metadata. Publish only commit/PR text approved for the repository's visibility; do not copy private ticket descriptions into public release notes. Mint preserves explicit issue links offline and does not fetch private content.

### Lane

- Default to a new lane without asking: reuse one strongly matching open issue or create one assigned to the human, then use branch `<issue-id>` in worktree `~/worktrees/<owner>/<repository>/<issue-id>`, and deliver through one ready pull request.
- Continue an existing branch or pull request only when the user explicitly names it for the same work. Repair of an exact existing pull request always reuses that pull request's head branch; never open a coordinating or corrective pull request for scope-preserving work.
- Never ask the user to choose between lanes. Ask only when implementation intent or a named target is materially ambiguous. Materially new scope gets its own lane.

- `<issue-id>` is `GH-<number>` for GitHub Issues or the exact returned Linear identifier, such as `ENG-123`. Keep existing GitHub lane names and historical references valid. Switching providers applies to new work; finish an existing PR with its recorded original ticket/provider.

### Worktrees

- Discover the default branch and fetch it; never assume `main`. Create the lane with `git worktree add -b <issue-id> <path> origin/<base>`, or reuse the worktree already registered for that branch.
- Never create worktrees inside a repository. Use a detached `PR-<number>` worktree only for read-only inspection.
- Link the primary checkout's `.env` and `.envrc` into a writable worktree as symlinks when they exist; never copy or overwrite environment files.
- Never stash, reset, clean, force-remove, or delete branches to make a worktree operation succeed. Refs, remotes, configuration, and stash are shared across worktrees.

### Primary Checkout

- The primary checkout is read-only for agent work, except the fast-forward-only default-branch update in `work-cleanup`. If it holds changes you did not make through a lane, preserve and report them.
- Exception: files a Kit command wrote to the primary checkout and that were delivered through a merged worktree pull request may be removed so the checkout can pull. Enumerate untracked candidates first, remove only verified command-owned paths, restore only tracked paths whose content still matches what the command wrote, and stop on any mismatch.

### Identity And Attribution

- Before the first commit, confirm `git config user.name`/`user.email` and the authenticated `gh` account are the human's; stop otherwise.
- Never add co-author trailers, "generated with" footers, agent signatures, or bot assignees to commits, issues, pull requests, or comments. Assign every created or reused issue and pull request to the human (`--assignee @me` for GitHub; verified human assignee ID for Linear).

### Commits And Pull Requests

- Stage explicit paths only (never `git add -A` or `git add .`), review the staged diff against the ask, and scan it for secrets before committing. Never commit `.env` files or secret values.
- Title format for commits and pull requests: `<type>(<issue-id>): <gitmoji> <short title>` with `feat :sparkles:`, `fix :bug:`, `docs :memo:`, `test :white_check_mark:`, `refactor :recycle:`, `chore :wrench:`, or `ci :green_heart:`. Commit bodies are short bullet lists.
- Linear example: `feat(ENG-123): :sparkles: add invitations`. End the commit description with `Refs: https://linear.app/<workspace>/issue/ENG-123` using the canonical returned URL (including its slug when returned). Preserve it in the squash commit description. Never generate `Closes #123` for Linear; the PR Ticket section uses the full Linear URL. GitHub closing references remain unchanged.
- Issues record the original ask, scope boundaries, acceptance criteria, and expected verification.
- Open pull requests ready for review, against the protected default branch, using the repository's pull-request template, and close the GitHub issue from the body (`Closes #123`), or link the Linear ticket and use its verified lifecycle after merge.
- Never commit to the default or a protected branch, force-push, or rewrite pushed history.
- After pushing, confirm the remote head matches the local commit and report the issue, branch, and pull request.
- When the pull request merges or closes, or the human stops the work, run `work-cleanup` for the lane. Cleanup after completion is not deleting a branch to clear a lane.

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
