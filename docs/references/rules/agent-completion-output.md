---
kind: ruleset
slug: agent-completion-output
description: Defines natural conversational replies and concise three-section reports for substantial task completion and handoff.
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

- Keep ordinary conversation direct and free of completion scaffolding.
- Make substantial task completion immediately understandable by reporting only
  what happened, deviations, and next steps.
- Preserve evidence and status semantics without repeating the same facts across
  separate validation, delivery, coordination, or repository-memory sections.

## Applies When

Use the structured contract when a substantial terminal completion or handoff
must communicate blockers, incomplete scope, required operator action,
repository or external-system mutation, delivery artifacts, multiple validation
layers, formal coordination, or evidence that an operator must preserve. Also
use it when the user explicitly asks for the canonical structured report.

This rule governs human-readable output, not tool-native JSON, machine-only
protocol output, intermediate progress commentary, or focused clarification
questions.

## Rules

Classify the response with the proportionality gate. When structured reporting
applies, use the three-section contract and task-specific content requirements
below as one completion contract.

## Proportionality Gate

### Conversational Responses

Answer naturally and lead with the answer for direct questions, definitions,
confirmations, rewrites, brief explanations, small read-only lookups, concise
recommendations, and ordinary conversational exchanges.

For these responses:

- Treat this gate as the specific exception to any general instruction that
  says every final response must lead with outcome, validation, and risk.
- Do not emit status tokens, canonical section headings, synthetic None items,
  task profiles, or repository-memory reporting.
- Include a follow-up suggestion only when it is genuinely useful.
- Match detail and formatting to the request. A short question may receive a
  short answer.

### Structured Handoff Triggers

Use the structured contract when any of these conditions applies:

- Omitting structure could hide a blocker, incomplete required scope, required
  next action, unresolved failure, or meaningful risk.
- The agent mutated a repository or external system, implemented or delivered
  work, or must report artifacts, commits, pull requests, deployments, or
  recovery state.
- Completion depends on validation layers or evidence states that an operator
  must distinguish.
- The task coordinates owners, workstreams, dependencies, or a formal handoff.
- The user explicitly requests the canonical structured report.

Do not use word count, token count, elapsed time, or tool-call count as an
applicability threshold. Classify by operational consequence. When uncertain,
prefer natural prose unless structure is necessary to keep required action,
incomplete work, a blocker, or material evidence visible.

## Three-Section Completion Contract

When the structured contract applies, emit exactly these headings in order:

1. `## What happened`
2. `## Deviations`
3. `## Next steps`

Do not add a status heading or any other top-level or second-level section. A
higher-priority host wrapper, directive, or machine tag may surround the
response; the What happened heading remains the first human-readable line
inside it.

### What Happened

The first bullet must be:

```markdown
## What happened

- **Status: <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>.**
```

Then add concise bullets for material outcomes. Put each fact once, under the
outcome it supports.

- Lead with the user-visible result, not agent activity.
- Fold task-specific results, validation, delivery, runtime state, coordination,
  and repository-memory evidence into this section.
- Use one nested evidence layer only when it materially helps the operator
  verify or understand an outcome.
- Group tightly related identifiers on one nested line when that remains easy
  to scan.
- Planned exclusions, deliberately default-off behavior, and authorized scope
  boundaries are outcomes rather than deviations when they occurred as
  intended.

### Deviations

Always include this section. Use one `**None.**` bullet when there are no
deviations.

```markdown
## Deviations

- **None.**
```

Replace the None bullet with one bullet per material divergence from requested
or expected scope:

- blocker or unresolved failure;
- incomplete required scope;
- warning or degraded execution;
- pending, unknown, skipped, unavailable, or not-applicable evidence when it
  affects interpretation;
- unperformed action that an operator could otherwise mistake as complete;
- stale, mismatched, or lower-confidence evidence.

Name the observed state and its impact. Preserve literal native states such as
`PENDING`, `UNKNOWN`, `SKIPPED`, `NOT_APPLICABLE`, and provider-specific
states. A planned exclusion is not a deviation. Do not repeat a deviation in
What happened unless the outcome would otherwise be misleading.

### Next Steps

Always include this section. Use one `**None.**` bullet when no action
remains.

```markdown
## Next steps

- **None.**
```

When action remains:

- Put one independently actionable item in each bullet.
- Start with `Required` or `Optional` and name the responsible actor when it
  is not obvious.
- State the action directly; add one nested line only when a reason, dependency,
  or copy-ready continuation is necessary.
- Every required follow-up includes a copy-ready prompt or command.
- Order required actions before optional actions.
- Never manufacture an action only to fill the section.

## Overall Status Semantics

### PASS

- Requested scope and required validation are complete or explicitly
  `NOT_APPLICABLE`.
- No required operator action remains.
- Non-blocking warnings or optional pending evidence may appear under
  Deviations without converting the outcome to PARTIAL.

### PARTIAL

- A usable result exists, but required scope or evidence remains incomplete.
- Deviations name every incomplete item and its impact.
- Next steps contain the exact action that resumes completion.

### BLOCKED

- Completion requires input, authority, credentials, capacity, approval, or
  external state the agent cannot establish within the task boundary.
- Safe unblocked work is complete.
- Deviations name the blocker and supporting evidence; Next steps name the
  smallest unblock action and copy-ready resume prompt.

### FAIL

- A required outcome or validation is known to fail, no external blocker is the
  stopping reason, and in-scope remediation did not produce a usable result.
- Deviations name the failure, attempted recovery, and remaining risk.
- Next steps name the next viable action.

Never translate pending, unavailable, skipped, or unobserved evidence into PASS.

## Task-Specific Content

Task types define which facts must be retained, not additional output sections.

- **Implementation and delivery:** What changed, material validation, issue,
  branch, commit, pull request, hosted state, and repository-memory decision
  when required. Put gaps in Deviations and operator actions in Next steps.
- **Research and discovery:** Question answered, finding, source or evidence,
  confidence, and implication. Separate sourced fact from inference.
- **Diagnosis and troubleshooting:** Symptom, confirmed cause or current
  hypothesis, evidence, confidence, and impact. Do not label a hypothesis as a
  confirmed root cause.
- **Planning and design:** Material decisions, chosen approach, rationale, and
  observable acceptance signals. Unresolved decisions are deviations with next
  steps.
- **Validation and testing:** Exact check, scope, observed state, and evidence
  or gap. Keep local, hosted, deployment, runtime, integration, physical, and
  business acceptance claims distinct.
- **Review and audit:** Findings ordered by severity, tight location or
  evidence, and required remediation. A clean review states inspected scope and
  residual limitations.
- **Operations, deployment, and monitoring:** Exact target and version, action
  or observation, literal state, evidence, and recovery boundary. Keep
  deployment, runtime health, integration behavior, and production acceptance
  distinct.
- **Coordination and handoff:** Workstream, owner, state, dependency, and exact
  handoff. Keep task outcome separate from degraded orchestration conformance.
- **Fallback:** Requested item, result, evidence, and limitation.

## Readability And Evidence Rules

- Use exactly the three canonical headings for a structured response.
- Use bullets, never Markdown pipe tables.
- Keep one concern per bullet and one nested evidence layer at most.
- Start with a short bold lead; put identifiers and evidence after it.
- Prefer exact links, identifiers, commands, timestamps, and counts over vague
  claims, but include only facts needed to understand, verify, or act.
- Do not repeat a fact across sections or restate the same success through
  multiple profile-shaped summaries.
- Put explanatory prose into the relevant bullet rather than adding another
  section.
- Redact secrets, credentials, private customer data, and signed URLs.

## Composition With Existing Contracts

- `github-pr-delivery` fields belong under What happened; pending checks or
  delivery gaps belong under Deviations; review or merge actions belong under
  Next steps.
- `testing-and-environment-validation` results belong under What happened;
  unavailable, pending, skipped, partial, or blocked evidence belongs under
  Deviations.
- Repository-memory decision, rationale, and artifacts become one concise What
  happened bullet when that contract requires them.
- `agent-team-orchestration` task outcome belongs in the status bullet;
  material execution facts belong under What happened and degraded
  conformance belongs under Deviations.
- Cross-repository program evidence belongs under What happened, grouped by
  workstream; unresolved dependencies belong under Deviations and Next steps.
- Higher-priority system, developer, client, tool, or host schemas take
  precedence. Preserve the same three semantic groups inside the wrapper.

## Anti-Patterns

- Manufacturing structured output for an ordinary conversational answer.
- Adding a separate PASS heading above What happened.
- Adding Completed, Validation, Delivery, Feature State, Residual Notes,
  Coordination, Repository Memory, or task-profile headings.
- Repeating a deployment or validation fact in multiple sections.
- Hiding blockers or incomplete work among successful outcome bullets.
- Reporting PASS while required validation is failing, pending, or unobserved.
- Omitting Deviations or Next steps instead of using a None bullet.
- Naming a required action without its owner and copy-ready continuation.
- Using tables, multi-level nesting, or large prose blocks.
- Replacing provider-native evidence states with an optimistic summary.

## Examples

Small conversational answer:

```markdown
“Refresh checks on the final commit” means rerun the required checks after the
last PR update so the results apply to the exact revision being reviewed.
```

Substantial implementation:

```markdown
## What happened

- **Status: PASS — completion output is proportional and ready for review.**
- **The canonical rule now reserves structured reporting for substantial work.**
  - Validation: complete Go and race suites passed.
- **Delivery is ready in PR #123.**
  - Issue #122; branch `GH-122`; commit `abc123`; hosted validation `SUCCESS`.
- **Repository memory was updated.**
  - Decision: updated; artifacts: feature spec and Constitution.

## Deviations

- **CodeRabbit remains `PENDING`.**

## Next steps

- **Optional — User:** Review PR #123 after CodeRabbit completes.
```

Complex production coordination:

```markdown
## What happened

- **Status: PASS — PRs #290, #181, and #230 merged and their configured deployments completed successfully.**
- **LabCore #290 merged, released as v0.58.0, and is healthy in production.**
  - Commit `6c7e16d`; deployment `SUCCESS`; ECS 1/1 healthy; health endpoint 200.
- **UI #181 and docs #230 merged and deployed successfully.**
  - UI and docs endpoints returned 200; CloudFront invalidations completed.
- **Production validation passed and manual upsert remained default-off as intended.**
  - Aggregate `PASS`; unauthenticated mutation returned 401; activation was not performed.
- **Repository memory was updated with source-to-runtime evidence.**

## Deviations

- **Non-fatal workflow warnings:** Node 20 deprecation and an unsupported input warning did not change the successful deployment outcome.

## Next steps

- **None.**
```

Blocked diagnosis:

```markdown
## What happened

- **Status: BLOCKED — production root cause cannot be confirmed with current evidence.**
- **Available logs end at the service boundary.**

## Deviations

- **Blocker:** Read-only production dependency logs are unavailable, so the root cause remains `UNKNOWN`.

## Next steps

- **Required — User:** Grant read-only production log access.
  - Continue with: `Resume diagnosis using the authorized production logs.`
```

## Verification

- Confirm ordinary conversation contains none of the canonical headings or
  status tokens.
- Confirm every structured response contains exactly What happened, Deviations,
  and Next steps in that order.
- Confirm the first What happened bullet contains one literal overall status.
- Confirm empty Deviations and Next steps sections contain one None bullet.
- Confirm no task-profile or evidence-specific heading is emitted.
- Confirm task-specific required facts remain present once, with at most one
  nested evidence layer.
- Confirm required follow-up is owner-specific and copy-ready.
- Confirm native pending, unknown, skipped, unavailable, and not-applicable
  states remain literal.
- Confirm no Markdown pipe table appears.
