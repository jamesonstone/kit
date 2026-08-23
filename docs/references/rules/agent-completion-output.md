---
kind: ruleset
slug: agent-completion-output
description: Defines status-first, action-first terminal task responses with required goal-specific evidence profiles.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - conversation
  - task
  - completion
  - reporting
  - implementation
  - research
  - diagnosis
  - planning
  - validation
  - testing
  - review
  - operations
  - deployment
  - monitoring
  - coordination
  - handoff
read_policy_default: must
---

# Ruleset: Agent Completion Output

## Purpose

- Make every terminal task response immediately scannable and actionable.
- Put the overall outcome, blockers, unfinished work, and exact continuation
  action before supporting detail.
- Preserve task-appropriate evidence without forcing every task into an
  implementation-shaped report.

## Applies When

- A coding agent returns a terminal response that completes, partially
  completes, fails, or hands off a task because it is blocked.
- The requested deliverable is implementation, research, diagnosis, planning,
  validation, review, operations, coordination, or another bounded outcome.

This rule does not apply to intermediate progress commentary or a focused
clarification question while the task remains active. It governs the terminal
human-readable response, not tool-native JSON or machine-only protocol output.

## Rules

Follow the universal envelope, status semantics, primary-profile selection,
readability, and composition requirements below as one terminal-response
contract.

## Universal Terminal Envelope

### Status Heading

The first human-readable line must be:

```text
# <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>
```

- Use one literal uppercase status and an em dash.
- State the user-visible result, not the work performed by the agent.
- Put no greeting, preamble, recap, or narrative before the heading.
- A higher-priority host wrapper, directive, or machine tag may surround the
  response. In that case, this heading is the first human-readable line inside
  the required wrapper.

### Operator Action List

Immediately after the status heading, emit a prioritized bullet list:

```markdown
## Next actions

- **None — No action required.**
  Why: Requested scope and required validation are complete.
  Continue with: `Optional: <copy-ready prompt when useful>`
```

- Order items as `Blocker`, `Incomplete`, `Next`, `Optional`, then `None`.
- Omit types that do not apply, except every PASS response includes one
  `None` item.
- Put one independently actionable concern in each item.
- Start the bold lead with the type, then the responsible actor when
  ownership is not obvious, such as `User:`, `Agent:`, or `External system:`.
- Include indented `Why:` and `Continue with:` lines on every item.
- Make every required `Continue with` value a copy-ready prompt or command.
  Do not write only “continue,” “retry,” “follow up,” or another vague action.
- Use `None` rather than omitting the list. Never hide a blocker or incomplete
  requirement below completed work.
- Do not use a Markdown pipe table for operator actions.

## Overall Status Semantics

### PASS

- The requested outcome is complete.
- Every required acceptance condition and validation layer passed or is
  explicitly `NOT_APPLICABLE`.
- No required operator action remains. Optional review, merge, deployment, or
  future enhancement may still be named as optional.

### PARTIAL

- The response contains a usable result, but required scope or evidence
  remains incomplete.
- No unresolved known failure is being presented as success.
- The action list names every incomplete item, why it remains, and the exact
  prompt or command that resumes completion.

### BLOCKED

- Completion requires input, authority, credentials, capacity, approval, or
  another external state the agent cannot establish within the task boundary.
- The action list names the blocker, its supporting evidence, the smallest
  unblock action, and the exact resume prompt.
- Complete all safe unblocked work before reporting BLOCKED.

### FAIL

- A required outcome or validation is known to fail, no external blocker is
  the stopping reason, and in-scope remediation did not produce a usable
  completion.
- The action list names the failure, attempted recovery, remaining risk, and
  the next viable action.

Evidence items preserve native observed states such as `PENDING`, `UNKNOWN`,
`SKIPPED`, `NOT_APPLICABLE`, `QUEUED`, or provider-specific failure states.
Never translate an unobserved or pending state into PASS.

## Primary Profile Selection

Select exactly one primary profile from the requested deliverable, not from
incidental activities used to produce it:

- A requested code change remains implementation/delivery even when research,
  diagnosis, review, and testing occurred during the work.
- A root-cause request remains diagnosis when no fix was requested.
- A request to prove behavior remains validation even when a defect is found.
- A request for an implementation-ready plan remains planning even when
  repository research was necessary.
- Use the fallback profile only when none of the named profiles represents the
  requested result.

Add a supplemental structure only when another active contract requires its
data or it materially improves the operator's next decision.

## Left-Aligned Detail Contract

- Do not use a Markdown pipe table in the normal terminal response.
- After the action list, use left-aligned section headings and CommonMark list
  or key/value blocks for the selected profile and every required supplement.
- Start each result with a short bold lead label. Put supporting evidence,
  rationale, scope, or limitations on indented continuation lines.
- A higher-priority output schema may still require a table or structured
  object. Preserve its required shape while keeping this rule's semantic order.

## Required Profiles

### Implementation And Delivery

```markdown
## Completed

- **`<literal result>` — `<scope item>`**
  Evidence: `<path, identifier, or observed behavior>`

## Validation

- **`<command or layer>`: `<observed state>`**
  Evidence: `<concise output or artifact>`
```

When Git or GitHub delivery occurred, include:

```markdown
## Delivery

- **Issue:** `<state and human assignee>` — `<number and URL>`
- **Branch:** `<state>` — `<exact branch>`
- **Commit:** `<state and human identity>` — `<full or short SHA>`
- **Pull request:** `<ready or draft and assignee>` — `<number and URL>`
- **Hosted checks:** `<literal observed state>` — `<check or run URLs>`
```

Every implementation response includes the existing repository-memory
contract in a left-aligned key/value block:

```markdown
## Repository Memory

- **Decision:** `created|updated|refactored|deleted|not required`
- **Rationale:** `<why>`
- **Artifacts:** `<paths or none>`
```

### Research And Discovery

```markdown
## Findings

- **Question:** `<question>`
  - **Finding:** `<answer>`
  - **Evidence and confidence:** `<source; confirmed, likely, or unknown>`
  - **Implication:** `<decision or effect>`
```

Distinguish sourced facts, agent inference, and unresolved unknowns.

### Diagnosis And Troubleshooting

```markdown
## Diagnosis

- **Symptom:** `<observed symptom>`
  - **Root cause or hypothesis:** `<confirmed cause or current hypothesis>`
  - **Evidence:** `<diagnostic evidence>`
  - **Confidence and impact:** `<confidence; affected scope>`
```

Do not label a hypothesis as a confirmed root cause. If the request included a
fix, use the implementation profile and include diagnosis as evidence.

### Planning And Design

```markdown
## Decisions

- **Decision:** `<material decision>`
  - **Chosen approach:** `<decision-complete choice>`
  - **Rationale:** `<tradeoff and evidence>`
  - **Acceptance signal:** `<observable completion>`
```

List unresolved material decisions in the action list. A plan is PASS only
when it is decision-complete for the requested scope.

### Validation And Testing

```markdown
## Validation

- **`<command, suite, or acceptance layer>`: `<observed state>`**
  - **Scope:** `<exact target>`
  - **Evidence or gap:** `<artifact, output, or missing evidence>`
```

Keep local, hosted, deployment, runtime, integration, physical, and business
acceptance claims separate. One layer never implies another.

### Review And Audit

```markdown
## Findings

- **`<priority or none>` — `<actionable finding>`**
  - **Location or evidence:** `<tight path, line, or source>`
  - **Required action:** `<smallest complete remediation>`
```

Order findings by severity. If no actionable findings exist, include one
`None` item and state the inspected scope and residual limitations.

### Operations, Deployment, And Monitoring

```markdown
## Operational Result

- **`<environment or resource>`: `<observed state>`**
  - **Action or observation:** `<bounded action or observation>`
  - **Evidence or recovery:** `<version, run, rollback, or next check>`
```

Identify the exact target and version. Report deployment, runtime health,
integration behavior, and production acceptance as separate observations.

### Coordination And Handoff

```markdown
## Workstreams

- **`<bounded lane>`: `<literal state>`**
  - **Owner:** `<accountable actor>`
  - **Dependency or next handoff:** `<dependency and exact handoff>`
```

When `agent-team-orchestration` applies, retain task outcome separately from
orchestration conformance and report actual agents separately from logical or
omitted lanes.

### Fallback

```markdown
## Result

- **`<requested item>`: `<result>`**
  Evidence or limitation: `<supporting evidence or boundary>`
```

## Readability Rules

- Keep the action list compact and use one concern per item.
- Use one result per list item and short, concrete lead labels.
- Put multi-paragraph explanation after the list item it supports.
- Prefer exact identifiers, links, paths, commands, timestamps, and counts over
  vague claims.
- Redact secrets, credentials, private customer data, and signed URLs.
- Do not repeat the same fact in the action list and detail blocks unless
  repetition is necessary to make a blocker actionable.

## Composition With Existing Contracts

- `github-pr-delivery` fields map into Delivery list items; none may be omitted.
- `testing-and-environment-validation` evidence maps into Validation list
  items and retains literal unavailable, pending, skipped, partial, and
  blocked states.
- Repository-memory decision, rationale, and artifacts map into the required
  Repository Memory key/value block for implementation responses.
- `agent-team-orchestration` two-axis reporting maps task success to the
  heading and orchestration conformance to the coordination or evidence list.
- Cross-repository program checkpoints retain repository, dependency,
  deployment, and acceptance evidence inside the coordination profile.
- Higher-priority system, developer, client, tool, or host output schemas take
  precedence. Preserve this rule's semantic order inside the required wrapper
  or use structurally equivalent fields when Markdown tables are prohibited.

## Anti-Patterns

- Starting with “Done,” a narrative recap, or implementation details before
  the overall status.
- Reporting PASS while required validation is failing, pending, or unobserved.
- Saying “no blockers” while required scope is incomplete.
- Naming a blocker without the smallest unblock action and resume prompt.
- Using a Markdown pipe table for operator actions in the normal terminal
  response.
- Using every profile because several activities occurred during one task.
- Replacing provider-native evidence states with an optimistic summary.

## Examples

Completed implementation:

```markdown
# PASS — canonical completion output is implemented and ready for review

## Next actions

- **None — No action required.**
  Why: Requested scope and required local validation are complete.
  Continue with: `Optional: Review pull request #123.`

## Completed

- **PASS — Canonical rule**
  Evidence: Status/action envelope plus every required task profile.

## Delivery

- **Pull request:** Ready and assigned to the human user — `#123`
```

Blocked diagnosis:

```markdown
# BLOCKED — production root cause cannot be confirmed without log access

## Next actions

- **Blocker — User: grant read-only production log access.**
  Why: Current evidence ends at the service boundary.
  Continue with: `Resume diagnosis using the authorized production logs.`

## Diagnosis

- **Symptom:** Requests fail after reaching the service boundary.
  - **Root cause or hypothesis:** Production dependency failure is unconfirmed.
  - **Evidence:** Application logs end before the dependency response.
  - **Confidence and impact:** Likely; production confirmation is blocked.
```

## Verification

- Confirm the first human-readable line uses the exact four-state heading.
- Confirm the operator action list follows immediately and has no blank Why
  or Continue with lines.
- Confirm Blocker, Incomplete, Next, Optional, and None items are ordered and
  semantically consistent with the overall status.
- Confirm one primary profile matches the requested deliverable.
- Confirm required delivery, validation, orchestration, program, and
  repository-memory evidence remains present.
- Confirm every required follow-up is copy-ready and every unavailable or
  unobserved result is reported literally.
- Confirm the normal terminal response does not use a Markdown pipe table.
