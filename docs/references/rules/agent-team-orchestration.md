---
kind: ruleset
slug: agent-team-orchestration
description: Controls accountable supervisor, specialist subagent lanes, concurrency, and read-only verification.
status: active
applies_to:
  - coding-agent
  - workflow
  - dispatch
  - subagent
  - verification
read_policy_default: conditional
---

# Ruleset: agent-team-orchestration

## Purpose

Use one accountable supervisor agent with dynamic specialist subagents only when
that team shape improves correctness or throughput.

This ruleset controls execution topology. It does not relax workflow phase
gates, source-of-truth rules, dirty-worktree ownership, validation, reflection,
or GitHub delivery gates.

## Applies When

- A coding agent plans implementation, validation, review, or repair work.
- Work may split into backend, frontend, CLI, tests, docs, data, security,
  compatibility, validation, or repo-research lanes.
- `kit dispatch`, `kit pr fix`, a prompt-library dispatch prompt, CI dispatch,
  or a subagent-enabled loop review is used.
- A v2 `SPEC.md` workflow needs an Agent Team Plan.

Use one supervisor lane only, and record the reason, when the work is trivial,
tightly coupled, high-overlap, high-ambiguity, requires continuous design
judgment, cannot spawn subagents in the active runtime, or the user explicitly
requested single-agent execution.

## Rules

### Supervisor Contract

- The supervisor owns the user request, active durable artifact, scope,
  non-goals, assumptions, acceptance criteria, implementation plan, lane
  assignment, touched-file prediction, integration, conflict resolution,
  validation, read-only verification assignment, documentation updates,
  delivery decision, and final response.
- Do not parallelize accountability.
- Do not describe a logical lane as a spawned agent unless a separate agent
  actually ran.

### Required Agent Team Plan

Before implementation, produce an Agent Team Plan that includes:

- supervisor responsibilities
- proposed lanes
- subagents that will actually be spawned
- logical-only lanes that will not be spawned
- intentionally omitted implementation or verification lanes
- reason for each omitted implementation or verification subagent
- predicted touched files per lane
- overlap risks
- max concurrency
- serialized work
- validation and review lanes

### Lane Assignment

Each implementation subagent must receive a clear objective, relevant
acceptance criteria, relevant source-map or repo facts, expected files or
packages, areas to avoid, validation expectations, output format, and an
instruction to report blockers instead of guessing.

Subagents must not independently expand scope. Subagents must not mutate
delivery state, create branches, stage files, commit, push, open PRs, resolve
review threads, or mark the whole workflow complete unless explicitly assigned
and allowed by the supervisor.

### File Overlap

Predict file ownership before spawning subagents.

If two lanes would edit the same file, prefer serial execution, assign one lane
as implementation and another as read-only review, or split the work
differently. Do not run parallel implementation lanes against the same files
unless the overlap is intentional, low-risk, and the supervisor records the
integration plan.

### Concurrency Limits

Default maximum concurrent lanes: 3.

Hard ceiling: 4.

Use 4 lanes only when predicted file overlap is clearly low and each lane has
an independent validation surface. Never use "as many agents as possible."

### Read-Only Verification

After implementation, use at least one read-only verification subagent by
default unless the change is documentation-only, trivial, tightly coupled, the
runtime cannot spawn subagents, or the user requested single-agent execution.

Verification agents must not edit files, stage changes, commit, push, close
findings, mark acceptance criteria complete, resolve review threads, or mutate
issue, branch, PR, or review-thread state.

Verification agents review the durable spec or task artifact, acceptance
criteria, actual diff, tests and command output, runtime behavior when relevant,
documentation updates, evidence artifacts, and source-map or cited repo facts.

### Verification Findings

Each verification finding must include:

- gap ID
- related acceptance criterion ID
- related source-map or repo-fact ID, if available
- evidence inspected
- actual behavior
- expected behavior
- risk
- recommended fix area
- whether delivery is blocked

For every verification gap, use this trace shape when possible:

```text
gap id -> acceptance criterion id -> source-map id -> fix diff area -> rerun evidence -> verifier closure
```

Do not proceed to reflection or delivery with open verification gaps unless the
user explicitly accepts the residual risk.

### Final Response

Final responses must report actual subagents spawned, logical lanes not spawned,
and any single-lane exception.

If no separate agents actually ran, state exactly:

```text
single supervisor lane; no specialist or verification agents spawned
```

## Anti-Patterns

- Do not use subagents merely because they are available.
- Do not turn broad discovery into parallel execution before file overlap can
  be predicted with reasonable confidence.
- Do not hide parallel accountability behind multiple agents.
- Do not allow subagents to expand scope independently.
- Do not let verification agents mutate files, git state, GitHub state, or
  acceptance status.
- Do not claim subagents were used when only logical lanes were planned.

## Verification

- Confirm the Agent Team Plan exists before implementation unless a recorded
  single-lane exception applies.
- Confirm actual spawned agents and logical-only lanes are distinguished.
- Confirm max concurrency is 3 by default and never above 4.
- Confirm overlapping files were serialized, merged into one lane, or assigned
  as implementation plus read-only review.
- Confirm read-only verification ran by default after implementation or an
  allowed exception was recorded.
- Confirm open verification gaps block reflection and delivery unless the user
  explicitly accepts residual risk.
- Confirm the final response reports actual subagents spawned or the exact
  single-lane sentence.

## Examples

Single supervisor lane:

```text
This change touches one prompt builder and one paired test. The files are
tightly coupled, so use a single supervisor lane and record the exception.
```

Low-overlap lanes:

```text
Lane A updates CLI behavior, Lane B updates docs, and Lane C performs read-only
verification. Predicted files do not overlap, max concurrency is 3.
```

Overlapping files:

```text
Both the implementation and test-update lanes need the same prompt builder.
Run implementation first, then use a read-only verifier against the final diff.
```
