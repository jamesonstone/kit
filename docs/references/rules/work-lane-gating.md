---
kind: ruleset
slug: work-lane-gating
description: Requires an explicit user-selected pull-request lane before any coding-agent repository mutation.
status: active
applies_to:
  - git
  - github
  - workflow
  - coding-agent
read_policy_default: must
---

# Ruleset: work-lane-gating

## Purpose

- Require the user to choose the delivery lane before a coding agent changes
  repository files or delivery state.
- Ensure every coding-agent change has a concrete pull-request landing plan.
- Keep the clone's primary checkout read-only and preserve every existing lane.

## Applies When

- Always active after read-only `safety-guardrails` recon and before any
  coding-agent repository mutation.
- Applies to source, tests, documentation, specs, plans, notes, generated
  artifacts, configuration, dependencies, workflows, migrations, and every
  other version-control-eligible repository file.
- Applies before issue, branch, worktree, staging, commit, push, or pull-request
  mutation. After the user chooses a new lane, issue, branch, and worktree
  creation may establish that lane before file edits.
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

1. the user's explicit lane choice for the current unit of work; and
2. a recorded Pull-Request Landing Plan that proves where the change will be
   reviewed.

A generic request to implement, commit, push, fix a PR, or produce a pull
request is not a substitute for the binary lane choice. A prior choice counts
only when it clearly covers the same unit of work in the current task.

### Required User Choice

After read-only recon, stop and ask:

> Before I make any repository changes, should I create a new GitHub issue, `GH-<issue-number>` branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?

- Wait for an explicit answer unless the user already answered this exact
  choice for the same scope, for example `new worktree` or `continue existing`.
- Do not infer the choice from a clean default branch, a dirty feature branch,
  an issue reference, a generic pull-request end state, or an agent's opinion
  about the most convenient lane.
- One recorded choice covers the accepted work plus directly required tests,
  documentation, validation fixes, review fixes, and delivery.
- Ask again before materially new or tangential scope. Do not repeatedly ask
  for routine subtasks that remain inside the recorded lane and pull-request
  plan.

### Pull-Request Landing Plan

Before file mutation, record these fields in-thread:

```text
Pull-Request Landing Plan:
- Choice: new lane | continue existing
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
- A plan is not permission to merge. It is the proven route from the writable
  worktree to one reviewable pull request.
- If scope, repository, issue, branch, worktree, base, or pull-request target
  changes materially, stop, refresh recon, and record a revised user choice
  and plan before further mutation.

### New Lane

When the user chooses a new lane:

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

When the user chooses to continue the existing lane:

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

If a repository file was changed before the lane choice and Pull-Request
Landing Plan, or any coding-agent change appeared in the primary checkout:

1. Stop immediately and report the exact path, checkout, branch, and observed
   state.
2. Preserve every working-tree and index change. Do not stage, commit, push,
   stash, reset, clean, switch branches, overwrite, delete, or silently
   transfer the ungated change.
3. Ask the required lane question if it has not been answered.
4. After the user chooses, prove exact ownership and recovery boundaries before
   recreating or transferring any command-owned change into the writable lane.
   Never infer ownership from post-change status alone.
5. Keep unrelated or ambiguous changes untouched. If exact recovery cannot be
   proven, report the blocker and the smallest user action required.

The tripwire is a fail-closed recovery boundary, not permission to normalize
root-checkout editing into the regular workflow.

## Anti-Patterns

- Automatically allocating a lane because the default branch is clean or
  current.
- Treating a request for a pull request as the user's new-versus-existing lane
  choice.
- Writing a spec, plan, README, generated file, or configuration before the
  lane choice because it is not application code.
- Editing the primary checkout with the intention of moving the diff later.
- Continuing on a feature branch without proving its owning worktree and
  create-or-update pull-request route.
- Asking again for every test or documentation subtask already inside the
  accepted scope.
- Hiding tangential work in the current lane without a new choice.
- Staging, committing, pushing, resetting, cleaning, stashing, or silently
  transferring work after a tripwire violation.

## Verification

- Confirm read-only `safety-guardrails` recon ran before the lane decision.
- Confirm the user explicitly chose new or existing for the current scope.
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
- Confirm materially new scope re-opened the choice while routine in-scope
  completion work did not.
- Confirm tripwire state was preserved and no ungated change was staged,
  committed, pushed, discarded, or silently transferred.

## Examples

Required choice before any repository write:

```text
Before I make any repository changes, should I create a new GitHub issue, `GH-<issue-number>` branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?
```

Valid new-lane plan:

```text
Choice: new lane
Issue: #143
Branch: GH-143
Worktree: ~/worktrees/jamesonstone/kit/GH-143
Protected base: main
Pull request: create one ready PR
```

Valid in-scope continuation:

```text
The user already chose GH-143 for this unit of work. A test fix required by
that same pull request does not require another lane question.
```

Tripwire:

```text
An ungated README edit exists in the primary checkout. Stop, preserve it, ask
the lane question, and do not stage, commit, push, discard, or silently move it.
```
