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
- During Linear setup, use the authorized Linear connection to read the selected project's teams and the configured team's current key. A Linear project can have multiple teams. If the project has one eligible team, use that team. If several are eligible, use the verified configured or explicitly selected team. Ask for the team if the choice is ambiguous. For team-only work, read the selected team directly. Pass its exact key (for example, `--linear-team-key ENG`) with the verified target. Kit stores it as optional `issues.linear.team_key` and reports the configured branch prefix.
- Read repository-local `.kit.yaml` `issues.provider`: omitted or `github` uses the existing GitHub Issues flow; `linear` uses Linear only. Provider selection does not change GitHub pull requests, reviews, protection, or release hosting. Never inherit an issue target from global configuration.
- For GitHub, use the repository's existing `gh` or connected GitHub tools and human assignment. For Linear, use the host's already-authorized Linear connector or API capability. If unavailable, stop ticket writes and report the missing Linear capability; do not install a connector, configure credentials, broaden access, or fall back to GitHub.
- Before a Linear write, verify the authenticated human, workspace URL slug (`issues.linear.workspace`), team UUID (`team_id`), and optional project UUID (`project_id`). Read the team and project to confirm workspace membership and that the project supports that team. Resolve the human's Linear assignee ID; GitHub logins are not Linear assignee IDs. Missing configuration, permission errors, target mismatch, or ambiguous lookup is a blocker. Never choose the first name match. An omitted project means explicitly team-only work.
- Search the selected repository/team/project for strongly matching work, including closed tickets and linked PRs, before creating. Reuse one clearly matching open ticket; if matches conflict, stop and report the ambiguity. Do not reopen completed work or create duplicates after an uncertain write: first read back by returned ID or search the exact title, target, human, and repository link. Retry only after confirming no ticket was created. A failed search is not an empty search.
- Both providers get the same title and Markdown substance: original ask, repository/spec context, scope and exclusions, acceptance criteria, expected verification, a filled `Verified when` section, and cross-repository dependencies when relevant. Preserve links and structure; never summarize away requirements for Linear. For Linear send the title and complete Markdown as `description`, plus the verified team ID, human assignee ID, and optional project ID. Read back the returned identifier, canonical URL, assignee, team, project, and description before creating the branch.
- Provider errors (authentication, permissions, transport, rate limits, GraphQL errors, unsuccessful mutation, malformed response) stop the ticket action. Report the exact failed action/target and whether denied, pending, or uncertain. Never infer success from HTTP status alone. Continue independent local work when safe; never invent an identifier.
- Keep lifecycle intent equal: record progress and PR links on the selected ticket; mark it completed only after observing a satisfying merge. Use the team's verified workflow states for Linear, not guessed state names or GitHub closing keywords. Abandoned/superseded work gets a reason and replacement link, with the appropriate verified canceled state. No automatic state transition is implied by opening a PR.
- Linear identifiers and URLs can reveal private metadata. Publish only commit/PR text approved for the repository's visibility; do not copy private ticket descriptions into public release notes. Mint preserves explicit issue links offline and does not fetch private content.

### Work ledger hook

- Load global Kit configuration from `$XDG_CONFIG_HOME/kit/.kit.yaml` when `XDG_CONFIG_HOME` is absolute and nonempty. Otherwise use `~/.config/kit/.kit.yaml`. Read unset global keys from `~/.kit/.kit.yaml`, then from `~/.config/kit/.kit.yaml` if that path differs from the canonical file. The user files are optional and are never committed. A missing user file changes nothing. The repository `.kit.yaml` uses the same schema and overrides the user file key by key. Mappings merge. A repository scalar, list, or explicit null replaces the global value at that key. Keep global values out of repository config and generated ledger notes. Keep secrets out of both scopes.
- Resolve `work_ledger` from that merged configuration. The keys are the same in both files: `provider`, `owner_label`, `plane.workspace`, `plane.project`, `plane.name`, `plane.states.active`, `plane.states.in_review`, and `plane.states.done`. An omitted `work_ledger`, an omitted `provider`, or `provider: none` is off. This hook does not change the selected issue provider.
- `work_ledger.provider` selects a swappable adapter. `plane` is implemented. `linear`, `jira`, and `notion` are reserved so a later adapter can use the same key; do not invent their calls. Any other provider is a blocker.
- Kit stores no ledger credentials, adds no ledger SDK, and does not call the ledger API. The native agent performs the adapter's steps through its already-authorized connector (for Plane, the Plane MCP). If that connector is missing, report the skipped ledger write and continue the ticket. Do not install a connector, configure credentials, broaden access, or put secrets in the repository.
- A `plane` adapter requires `work_ledger.plane.workspace` (URL slug) and `work_ledger.plane.project` (project identifier). Optional `work_ledger.plane.name` is the project title. Optional `work_ledger.owner_label` is the acting agent's label. State names come from `work_ledger.plane.states.active`, `in_review`, and `done`. Omitted state names mean `Active`, `In review`, and `Done`. Verify those states on the project before writing. The repository owner supplies the workspace, project, and label. Examples use placeholders such as `<workspace>` and `<project>`. Do not invent a person, workspace, project, URL, or bot label.
- Kit writes a work-ledger note into the agent contract only when the repository `.kit.yaml` sets `work_ledger.provider` and contains a complete target. It uses only repository values. A user-local target stays out of the repository.
- At ticket start, when no pull request is open: find the ledger item that already links this issue. If none, create one titled as the issue. Link the issue on that returned item, set the configured active state (default `Active`), and apply `work_ledger.owner_label` when it is set.
- At ticket start, when a pull request for this issue is already open, including a repair of that pull request: find or create the item, link the issue and the pull request, and set the configured in-review state (default `In review`). Do not move that item back to the active state.
- When the pull request opens: find the item that links this issue, link the pull request, and set the configured in-review state (default `In review`). Do not create a second item.
- When that pull request merges: find the item that links this issue and set the configured done state (default `Done`). Do not create a second item. Do not set the done state for an unmerged or closed-unmerged request.
- A failed search is not an empty project. After an uncertain write, read the item back before creating another. Never invent an identifier. Ledger errors stop the ledger action and are reported as denied, pending, or uncertain. They do not by themselves stop the ticket.

### Lane

- Default to a new lane without asking: reuse one strongly matching open issue or create one assigned to the human, then use branch `<issue-id>` in worktree `~/worktrees/<owner>/<repository>/<issue-id>`, and deliver through one ready pull request.
- Continue an existing branch or pull request only when the user explicitly names it for the same work. Repair of an exact existing pull request always reuses that pull request's head branch; never open a coordinating or corrective pull request for scope-preserving work.
- Never ask the user to choose between lanes. Ask only when implementation intent or a named target is materially ambiguous. Materially new scope gets its own lane.

- `<issue-id>` is `GH-<number>` for GitHub Issues or the exact returned Linear identifier, such as `ENG-123`. For a Linear target with a missing or stale stored team key, use the live team's key during verification and the returned issue identifier for the branch. Keep existing GitHub lane names and historical references valid. Switching providers applies to new work; finish an existing PR with its recorded original ticket/provider.

### Worktrees

- Discover the default branch and fetch it; never assume `main`. Create the lane with `git worktree add -b <issue-id> <path> origin/<base>`, or reuse the worktree already registered for that branch.
- Never create worktrees inside a repository. Use a detached `PR-<number>` worktree only for read-only inspection.
- Link the primary checkout's `.env` and `.envrc` into a writable worktree as symlinks when they exist; never copy or overwrite environment files.
- Never stash, reset, clean, force-remove, or delete branches to make a worktree operation succeed. Refs, remotes, configuration, and stash are shared across worktrees.

### Primary Checkout

- The primary checkout is read-only except for the default-branch fast-forward in `work-cleanup` and user-requested local SOPs under `standard-operating-procedures`. Preserve and report existing changes.
- Exception: files a Kit command wrote to the primary checkout and that were delivered through a merged worktree pull request may be removed so the checkout can pull. Enumerate untracked candidates first, remove only verified command-owned paths, restore only tracked paths whose content still matches what the command wrote, and stop on any mismatch.

### Identity And Attribution

- Before the first commit, confirm `git config user.name`/`user.email` and the authenticated `gh` account are the human's; stop otherwise.
- Never add co-author trailers, "generated with" footers, agent signatures, or bot assignees to commits, issues, pull requests, or comments. Assign every created or reused issue and pull request to the human (`--assignee @me` for GitHub; verified human assignee ID for Linear).

### Commits And Pull Requests

- Stage explicit paths only (never `git add -A` or `git add .`), review the staged diff against the ask, and scan it for secrets before committing. Never commit `.env` files or secret values.
- Title format for commits and pull requests: `<type>(<issue-id>): <gitmoji> <short title>` with `feat :sparkles:`, `fix :bug:`, `docs :memo:`, `test :white_check_mark:`, `refactor :recycle:`, `chore :wrench:`, or `ci :green_heart:`. Commit bodies are short bullet lists.
- Linear example: `feat(ENG-123): :sparkles: add invitations`. End the commit description with `Refs: https://linear.app/<workspace>/issue/ENG-123` using the canonical returned URL (including its slug when returned). Preserve it in the squash commit description. Never generate `Closes #123` for Linear; the PR Ticket section uses the full Linear URL. GitHub closing references remain unchanged.
- Issues record the original ask, scope boundaries, acceptance criteria, and a `Verified when` section filled in before an agent starts.
- Open pull requests ready for review, against the protected default branch, using the repository's pull-request template. Include the `Evidence` section and the fresh-context review outcome. Close the GitHub issue from the body (`Closes #123`), or link the Linear ticket and use its verified lifecycle after merge.
- Never commit to the default or a protected branch, force-push, or rewrite pushed history.
- After pushing, confirm the remote head matches the local commit and report the issue, branch, and pull request.
- When the pull request merges or closes, or the human stops the work, run `work-cleanup` for the lane. Cleanup after completion is not deleting a branch to clear a lane.

### Verification

- Fill `Verified when` on the issue before implementation starts. State the definition of done, what is out of scope, the concrete evidence to show (a passing test, a screenshot, a log or trace), and an optional hypothesis or expected answer from the requester. Use the repository issue template when the provider is GitHub.
- Fill the pull request `Evidence` section from the work that ran. Record the commands and their real output or excerpts. Record decisions and assumptions, including defaults chosen, scope calls, and conflicting requirements reinterpreted. For each check, name the property it verified. Name at least one thing that was not checked.
- Before opening the pull request, run a fresh-context review. A second agent or subagent with no prior context on the work reads the complete diff cold against the issue's `Verified when` section and answers "did I miss anything?". Do not open the pull request without this review.
- In the pull request's Evidence section, record the reviewed commit or identifiable diff snapshot, its scope, and the outcome. Compare the current diff with the recorded review coverage before opening the pull request or pushing another change to it.
- Require a focused independent follow-up for substantive changes after review. These include changes to behavior, permission boundaries, configuration precedence, managed-file ownership, acceptance criteria, or what a test considers a pass. Include affected behavior at the base when a rebase or conflict resolution changes it. The original independent reviewer may review the delta and its affected dependencies. Record the follow-up snapshot, scope, and outcome.
- For minor edits, record the lead's diff check. Typos, formatting, and mechanical reviewer-requested fixes qualify when they leave meaning and behavior intact. For a base refresh, use the lead check only if the patch and relevant context are equivalent, and rerun the affected checks.

### Recovery

- Diagnose failures from current evidence and recover without destructive commands. Missing credentials, an unexpected identity, or an ambiguous target is a blocker: report the smallest input needed.

## Work ledger

Track longer-running work, anything the repository owner names, or anything you recommend tracking (with the owner's OK) in the repository's configured work ledger. Before you ping the owner about a task, find its item; if it's done or owned by another agent, don't ping. Create the item if it is missing. After you act, update the item's owner label, status (Active / In review / Done), and a one-line note. Tag it with your own agent label (`agent:<you>`). If you hand work to another agent, relabel the item and add a note. The ledger holds the current state.

Use the same schema in global and repository `.kit.yaml` files. Repository keys override global defaults. If the effective settings, including legacy global fallbacks, have no provider or set it to `none`, this rule is a no-op. Use the work ledger hook above to resolve the effective settings.

```yaml
work_ledger:
  provider: plane
  owner_label: <owner-label>
  plane:
    workspace: <workspace>
    project: <project>
```

## Examples

```bash
git fetch origin main
git worktree add -b GH-123 ~/worktrees/acme/api/GH-123 origin/main
```

```text
fix(GH-123): :bug: reject expired session tokens
```
