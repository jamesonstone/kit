---
kind: ruleset
slug: work-lane-gating
description: Defaults coding-agent repository mutations to a new pull-request worklane without prompting.
status: active
registry_scope: downstream
applies_to:
  - git
  - github
  - workflow
  - coding-agent
read_policy_default: must
---

# Ruleset: work-lane-gating

## Purpose

- Default every accepted coding-agent repository mutation to a new issue,
  branch, canonical worktree, and pull request without asking for a lane choice.
- Continue an existing lane only when the user explicitly directs that outcome
  for the same unit of work and its exact ownership can be proven.
- Ensure every coding-agent change has a concrete pull-request landing plan.
- Keep the clone's primary checkout read-only and preserve every existing lane.

## Applies When

- Always active after read-only `safety-guardrails` recon and before any
  coding-agent repository mutation.
- Applies to source, tests, documentation, specs, plans, notes, generated
  artifacts, configuration, dependencies, workflows, migrations, and every
  other version-control-eligible repository file.
- Applies before issue, branch, worktree, staging, commit, push, or pull-request
  mutation. Issue, branch, and worktree creation may establish the default lane
  before file edits.
- Applies to implementation, maintenance, repository bootstrap, PR repair,
  documentation-only work, and commands such as `kit spec`, `kit init`, or
  `kit reconcile` that can write repository files.
- Does not block read-only discovery, safety recon, capability inspection,
  context resolution, review, explanation, or native planning that creates no
  repository or delivery mutation.
- Governs coding-agent actions. It does not prohibit a human from editing
  repository files manually.

## Rules

### Mutation Boundary

A repository mutation is any coding-agent action that creates, edits, deletes,
moves, formats, or generates a repository file, changes the Git index or refs,
or mutates a GitHub issue or pull request for the work.

Before the first mutation, the agent must have both:

1. the default new worklane established, or an explicitly requested existing
   lane proven for the same unit of work; and
2. a recorded Pull-Request Landing Plan that proves where the change will be
   reviewed.

An accepted request to mutate the repository activates the new-worklane
default and its issue-to-ready-PR delivery authority when the work has no
existing issue/branch/pull-request owner. Read-only requests do not allocate a
lane. The user does not need to separately select or approve the default lane.

### Required Default

After read-only recon:

- Do not ask whether to create a new lane or continue an existing lane.
- Default to a new worklane even when the current checkout is clean, dirty, on
  a protected branch, or on a feature branch. A bare issue reference or generic
  request for a pull request does not prove an existing owner.
- Search for one exact matching issue and reusable complete lane before
  creating anything. Reuse only a strong scope match whose issue, exact branch,
  canonical non-primary worktree, protected base, and pull-request route agree;
  otherwise create the new lane.
- Continue an existing lane only when the user explicitly directs continuation
  for the same unit of work, such as by saying `continue existing` or naming the
  exact branch or pull request to update. Never offer continuation as a choice.
- A generic implementation request, current feature branch, dirty state, issue
  reference, or pull-request end state is not an explicit continuation
  direction.
- Ask only when implementation intent or a user-named target is materially
  ambiguous and cannot be resolved from repository evidence. Do not ask for a
  new-versus-existing lane preference.
- One lane allocation covers the accepted work plus directly required tests,
  documentation, validation fixes, review fixes, delivery, remaining
  pull-request review, an authorized merge of that pull request, and
  post-merge primary leftover cleanup.
- Remaining in the coding-agent session to handle review, merge once
  authorized, and then clean the primary default branch is in-scope
  continuation of the recorded lane.
- Materially new or tangential scope defaults to another new worklane once its
  implementation intent is accepted. Do not create another lane for routine
  subtasks that remain inside the recorded lane and pull-request plan.

### Existing Pull-Request Lifecycle Precedence

- A task targeting one or more exact existing pull requests for review repair,
  CI repair, base refresh, conflict resolution, generated-artifact refresh,
  dependency-ordered merge coordination, or other scope-preserving remediation
  is explicit continuation of each targeted pull request's existing lane.
- Reuse each target's same-repository head branch, issue, owning non-primary
  worktree, protected base, and pull-request identity. Do not allocate a new
  coordination issue, branch, worktree, or pull request for that lifecycle work.
- For multiple pull requests, record one continuation entry per target in the
  bounded merge or program plan. That exact target set replaces the singular
  create-or-update landing target; it does not create a coordinator pull request.
- Follow `github-pr-merge` for dependency order and bounded in-place repair.
  Scope-preserving repair stays on the existing heads with ordinary commits;
  never create recursive corrective pull requests merely to make another pull
  request current or mergeable.
- If the user or accepted plan authorizes merge but not source repair, stop for
  bounded in-place-remediation authority before changing a head. Missing repair
  authority is not a reason to allocate a new worklane or replacement pull request.
- Use a replacement pull request only when the remediation materially changes
  scope or architecture, the original head cannot be updated safely, or
  repository policy or the user explicitly requires replacement.

### Pull-Request Landing Plan

Before file mutation, record these fields in-thread:

```text
Pull-Request Landing Plan:
- Routing: default new lane | explicit continue existing
- Repository:
- Issue: create | reuse <number/link>
- Branch:
- Worktree:
- Protected base:
- Pull request: create ready PR | update <number/link>
- Scope match:
- Unknowns/blockers:
```

- Every field must be known and mutually consistent before file mutation.
- Existing multi-PR lifecycle work may record one complete continuation entry
  per target inside its bounded merge or program plan instead of inventing one
  coordinator issue, branch, worktree, and pull request.
- A plan is not permission to merge. It is the proven route from the writable
  worktree to one reviewable pull request.
- If scope, repository, issue, branch, worktree, base, or pull-request target
  changes materially, stop, refresh recon, apply the default new-lane routing
  or the user's explicit continuation direction, and record a revised plan
  before further mutation.

### Merge Authorization

- PR-delivery consent never implies merge consent. It authorizes issue, branch,
  commit, push, and ready-PR delivery only.
- A direct merge request or accepted bounded merge plan routes to
  `github-pr-merge` and the `pull-request-merge` context workflow.
- The exact existing pull-request set named by that request or plan is explicit
  continuation under this rule; do not apply the default-new route to create a
  separate coordination lane.
- The authorized set is exact. Adding a new PR, repository, base branch,
  deployment target, infrastructure effect, merge method, or actor requires
  follow-up authorization.
- Revalidating an unchanged authorized head, retrying a compatible path, or
  using a repository-required merge queue does not require another prompt when
  target, scope, intended effect, identity, and approval remain unchanged. A
  changed head invalidates prior merge authority and requires fresh exact-head
  authorization under `github-pr-merge`.
- A gate decision, issue, branch, commit, push, ready PR, approval, passing
  check, review-thread resolution, subagent assignment, or program ledger does
  not create merge authority.

### New Lane

For the default new lane:

- Search for an exact matching issue and reusable lane before creating
  anything. Do not create duplicates.
- Create or reuse one human-assigned GitHub issue for the unit of work.
- Use exact `GH-<issue-number>` for the branch.
- Create or reuse its canonical linked worktree at
  `~/worktrees/<owner>/<repository>/GH-<issue-number>` from the refreshed
  remote protected base.
- Verify the new worktree owns the issue branch and is not the primary
  checkout before editing files.
- Plan one ready pull request from that branch to the protected base. Create it
  after implementation and validation unless repository rules require an
  earlier draft or the user explicitly requests one.

### Continue Existing

When the user explicitly directs continuation of the existing lane:

- Prove the current non-protected branch, its exact owning linked worktree,
  issue scope, remote, protected base, and existing or planned pull request.
- Reuse the existing pull request when one exists. Do not create a replacement
  branch or second pull request for the same lane.
- If the additional scope needs separate issue traceability, create or reuse a
  human-assigned issue, scope its commits to that issue, and update the same
  pull request's issue references.
- A detached `PR-<number>` view, protected branch, primary checkout, branch
  owned by another worktree, missing pull-request route, or ambiguous dirty
  state is not a writable existing lane. Fail closed and request the smallest
  decision needed to establish a valid lane.

### Primary Checkout Protection

- Resolve the clone's primary checkout from `git worktree list --porcelain`.
- Treat that exact checkout as read-only for coding-agent work, regardless of
  its current branch, cleanliness, file type, or planned pull request.
- The primary checkout may be inspected and may supply exact `.env` and
  `.envrc` symlink targets to writable lanes. Do not edit files, generate
  artifacts, stage, commit, switch branches, or perform implementation there.
- Never use the primary checkout as a temporary edit location followed by a
  later copy or transfer. Establish the writable lane first and run mutating
  commands there.

### Gate Tripwire And Recovery

If a repository file was changed before the lane was established and the
Pull-Request Landing Plan recorded, or any coding-agent change appeared in the
primary checkout:

1. Stop immediately and report the exact path, checkout, branch, and observed
   state.
2. Preserve every working-tree and index change. Do not stage, commit, push,
   stash, reset, clean, switch branches, overwrite, delete, or silently
   transfer the ungated change.
3. Apply the new-worklane default without asking, unless the user already
   explicitly directed continuation for this same scope.
4. Prove exact ownership and recovery boundaries before
   recreating or transferring any command-owned change into the writable lane.
   Never infer ownership from post-change status alone.
5. Keep unrelated or ambiguous changes untouched. If exact recovery cannot be
   proven, report the blocker and the smallest user action required.

The tripwire is a fail-closed recovery boundary, not permission to normalize
root-checkout editing into the regular workflow.

After the matching worktree pull request has been merged into the protected
default branch, leftover command-owned untracked files on the primary checkout
from that command may be removed with `git clean -fd` after enumerating or
dry-running all untracked files, verifying every candidate is command-owned,
and passing only those verified paths. Leftover command-owned tracked changes
in the index or worktree of those same exact paths may be restored to HEAD in
both the index and the worktree only after revalidating that the current index
and worktree contents of those paths still match the captured command-owned
snapshot; if any path mismatches or is ambiguous, stop and report it instead
of overwriting later edits, so the primary checkout can pull the merge.
Do not use this exception before merge, for unrelated dirty or untracked
state, or to create or clear a worktree.

## Anti-Patterns

- Asking the user whether to create a new lane or continue an existing lane.
- Continuing the current feature branch because it is convenient, clean, dirty,
  already associated with an issue, or already has a pull request.
- Writing a spec, plan, README, generated file, or configuration before the
  default or explicitly continued lane is established because it is not
  application code.
- Editing the primary checkout with the intention of moving the diff later.
- Continuing on a feature branch without proving its owning worktree and
  create-or-update pull-request route.
- Creating another lane for every test or documentation subtask already inside
  the accepted scope.
- Hiding tangential work in the current lane instead of applying the default
  new-worklane routing after its intent is accepted.
- Creating a coordination or corrective pull request for review repair, CI
  repair, base refresh, ordered merge work, or another scope-preserving change
  that belongs on a targeted existing pull-request head.
- Staging, committing, pushing, resetting, cleaning, stashing, or silently
  transferring work after a tripwire violation.

## Verification

- Confirm read-only `safety-guardrails` recon ran before lane routing.
- Confirm the agent defaulted to a new worklane without asking for a lane
  choice.
- Confirm an existing lane was used only after an explicit continuation
  direction for the same scope and exact target ownership was proven.
- Confirm the Pull-Request Landing Plan was complete before the first file
  mutation.
- Confirm a new lane used one human-assigned issue, exact issue branch,
  canonical non-primary worktree, protected base, and one ready-PR plan.
- Confirm an existing lane proved its non-protected branch, owning worktree,
  issue scope, base, and create-or-update pull-request plan.
- Confirm the primary checkout received no coding-agent file, index, commit, or
  branch mutation.
- Confirm documentation, specs, configuration, and generated files followed
  the same gate as source code.
- Confirm materially new accepted scope defaulted to another new lane while
  routine in-scope completion work stayed in the recorded lane.
- Confirm exact existing-PR lifecycle work reused every targeted head branch
  and pull request, recorded per-target continuation, and created no coordinator
  or recursive corrective pull request.
- Confirm tripwire state was preserved and no ungated change was staged,
  committed, pushed, discarded, or silently transferred.
- Confirm PR-delivery consent was not treated as merge consent, and any direct
  merge request or accepted bounded plan routed to `github-pr-merge`.

## Examples

Default routing before any repository write:

```text
The request requires repository mutation. After read-only recon, create or
reuse its human-assigned issue, exact issue branch, canonical non-primary
worktree, and ready pull-request plan without asking for a lane choice.
```

Explicit existing-lane continuation:

```text
Continue this work in GH-143 and update PR #144.
```

The agent proves that branch, worktree, issue, base, and pull request all match
before mutating. Without that explicit direction, it creates a new worklane.

Valid implementation-intent clarification:

```text
The request could mean changing generated instructions only or changing both
the canonical rule and generated instructions. Clarify that implementation
scope, then apply the new-worklane default without asking about lane choice.
```

Valid new-lane plan:

```text
Routing: default new lane
Issue: #143
Branch: GH-143
Worktree: ~/worktrees/jamesonstone/kit/GH-143
Protected base: main
Pull request: create one ready PR
```

Valid in-scope continuation:

```text
The user explicitly directed GH-143 for this unit of work. A test fix required
by that same pull request stays in the recorded lane without another prompt.
```

Tripwire:

```text
An ungated README edit exists in the primary checkout. Stop, preserve it,
establish the default new worklane without asking, and do not stage, commit,
push, discard, or silently move it.
```
