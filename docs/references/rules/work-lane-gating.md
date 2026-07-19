---
kind: ruleset
slug: work-lane-gating
description: Gates new implementation lanes so agents do not mix unrelated work or force docs into PR workflow.
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

Detect whether requested work is implementation work and whether it constitutes a new lane of work, then gate before writing code.

## Applies When

- Always active after `safety-guardrails` recon.
- Runs before any implementation work.
- Runs before `github-pr-delivery`, unless the user already explicitly requested a PR end state.
- Does not gate non-implementation work and does not create issues, branches, commits, pushes, or PRs for non-implementation work unless the user explicitly asks.

## Rules

### What Counts As Implementation Work

Gate only when the work will mutate source code or production-affecting config. Concretely, gate when the task will:

- Add, edit, or delete code files, including application, library, infrastructure-as-code, schema, or migration files.
- Change build, CI, dependency, or runtime configuration that ships.
- Otherwise produce a diff intended to land on a branch and become a PR.

### Non-Implementation Work

Never ask the gate question, and never create an issue, branch, commit, push, or PR for:

- `kit` pipeline phases: `BRAINSTORM`, `SPEC`, `PLAN`, and `TASKS`.
- `REFLECT` and any retrospective or analysis pass.
- Documentation writing or editing, including `.md`, ADRs, READMEs, design docs, and notes, when done on its own outside a code-change lane.
- Read-only, planning-only, review-only, or exploratory work.
- Ad-hoc work the user is driving manually.

Documentation and spec artifacts may be written and committed manually by the user at any time, untied to a `GH-123` branch. Never force docs/specs into a branch or PR workflow on your own initiative because doing so needlessly creates conflicts.

If non-implementation work later turns into actual code changes, re-evaluate at that transition and gate then.

### Trigger and Consent

- Before beginning implementation, determine whether the requested work is implementation work according to this ruleset's definitions.
- If the user already explicitly asked for a PR end state, treat that as consent and proceed to `github-pr-delivery`.
- If the user explicitly asks to create an issue, branch, or PR for docs or any other non-code work, honor that. The non-implementation exclusions govern agent initiative, not explicit user instruction.

### New-Lane Definition

Among implementation work, a new lane is any work that is either:

- Net-new to the current thread, with no existing issue, branch, or PR covering it.
- Tangential enough that bundling it into the current branch would mix unrelated concerns, review surfaces, blast radius, or revertability.

### Gate

When implementation work in a new lane is detected, stop before writing code and ask exactly:

> It appears you are doing implementation work. Would you like to create a new issue, branch, and PR for this work, or continue on the existing branch with the existing work?

- Wait for an explicit answer.
- Do not proceed until the user chooses.

### Gate Tripwire

- A gate decision must be recorded in-thread as the user's explicit choice before the first source-code edit.
- If code has been edited or staged without a recorded gate decision for that lane, treat it as a violation:
  - Stop immediately.
  - Report the violation.
  - Do not commit or push the ungated work.
  - Preserve the working tree and use `safety-guardrails` autonomous failure recovery to establish the correct lane when it can be proven safely; otherwise request only the missing lane decision. Do not discard, commit, or push the ungated work.

### Detection Heuristics

Any one of these triggers the gate for implementation work only:

- Touches files or modules outside the current change's import graph.
- Introduces a new feature, subsystem, or dependency unrelated to the active task.
- Requires a new migration, API surface, or config not in the current scope.
- The user's phrasing pivots with terms such as "also", "while you're at it", "separately", or "new thing".
- The commit message for the current work would need "and" to describe both efforts.

### Do Not Gate

Do not gate when:

- The work is non-implementation work according to this ruleset's exclusions.
- The work is a direct sub-task of the active branch's purpose.
- The work is a fix or refactor required to complete the current task.
- The work falls within the existing issue or PR description's scope.

## Anti-Patterns

- Do not ask the gate question for `kit` pipeline phases, reflection, retrospectives, standalone docs edits, read-only work, planning-only work, review-only work, exploratory work, or ad-hoc work the user is driving manually.
- Do not auto-create an issue, branch, commit, push, or PR for non-implementation work.
- Do not force docs or specs into a `GH-123` branch or PR workflow on agent initiative.
- Do not hide tangential implementation work in the current branch.
- Do not proceed after a gate tripwire violation by committing, pushing, resetting, rebasing, or otherwise mutating the working tree.

## Verification

- Confirm `safety-guardrails` recon ran first.
- Confirm the requested work was classified as implementation or non-implementation using this ruleset's definitions.
- Confirm non-implementation work did not trigger the gate and did not auto-create issue, branch, commit, push, or PR state.
- Confirm any new-lane decision for implementation work was recorded in-thread before source-code edits.
- Confirm PR delivery only ran after explicit consent or an explicit PR request.
- Confirm documentation and spec artifacts were not forced into a branch or PR workflow on agent initiative.

## Examples

Required gate question for a new implementation lane:

```text
It appears you are doing implementation work. Would you like to create a new issue, branch, and PR for this work, or continue on the existing branch with the existing work?
```

Non-implementation work that must not gate:

```text
Update a standalone Markdown ruleset, README, ADR, design doc, note, BRAINSTORM, SPEC, PLAN, TASKS, or REFLECT artifact.
```

Direct implementation sub-task that does not require a new lane:

```text
Fix a failing test introduced by the current branch's implementation.
```

New implementation lane that requires the gate:

```text
While implementing the current docs change, also add a new deployment workflow.
```
