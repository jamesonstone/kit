---
kind: ruleset
slug: agent-completion-output
description: Defines natural conversational replies and short, readable three-section reports for substantial task completion and handoff.
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
- Make a substantial completion report readable in one pass: what happened,
  what diverged, what remains.
- Carry enough evidence to trust and act on the outcome, and no more. A report
  that records everything the agent verified is a transcript, not a report.

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
applies, use the three-section contract, the density budget, and the
task-specific content requirements below as one completion contract.

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

## Density Budget

A structured report is a briefing for someone deciding what to do next. It is
not a record of the work, a defense of the work, or an index of everything
that was checked.

- Target twelve rendered lines or fewer for ordinary work. A genuinely complex
  outcome may run longer, but every line past that budget must earn its place.
- Keep What happened to five bullets or fewer. Needing more usually means the
  bullets describe activity rather than outcomes; merge them.
- Keep each bullet to one sentence, plus at most one short clause carrying its
  evidence. Three sentences in a bullet means it is two bullets, or it belongs
  in the opening prose.
- Nest at most one line under a bullet, and only when the reader needs that
  detail to verify or act. Most bullets should not nest at all.
- Leave the fuller account in the artifact that already holds it — the spec,
  the pull-request body, the issue — and point to it instead of restating it.

## Three-Section Completion Contract

When the structured contract applies, use these headings and no others:

1. `## What happened` — always present, always first.
2. `## Deviations` — only when something diverged.
3. `## Next steps` — only when action remains.

Keep that order when a section is present. Do not add a status heading or any
other top-level or second-level section. A higher-priority host wrapper,
directive, or machine tag may surround the response; the What happened heading
remains the first human-readable line inside it.

### What Happened

Open with the status line, bold, on its own line:

```markdown
## What happened

**Status: <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>.**
```

Follow it with one to three plain sentences saying what the user now has, in
the words a colleague would use out loud. Prose is the default shape for this
section: it reads faster than bullets and it preserves the causation that
bullets drop.

Add bullets only for outcomes that are genuinely separate items — distinct
deliverables, distinct systems touched, distinct findings. Do not split a
paragraph into bullets to look organized.

- Lead with the user-visible result, not agent activity.
- Fold task-specific results, validation, delivery, runtime state,
  coordination, and repository-memory evidence into this section, stating each
  fact once.
- Planned exclusions, deliberately default-off behavior, and authorized scope
  boundaries are outcomes rather than deviations when they occurred as
  intended.

### Evidence

Include an identifier when the reader needs it to act, or when they would
reasonably doubt the claim without it. Otherwise leave it out.

- Include: the pull request, issue, or branch they will open next; the exact
  target and version of anything deployed; the specific failure and where it
  lives; the command that resumes blocked work.
- Leave out: a commit SHA cited as proof that something was checked, rather
  than as the delivery identifier a composing contract requires; lists of test
  or suite names; file-and-line references for code that is correct; counts of
  things inspected; the tools and steps used to reach the answer.
- Name the validation that ran and what it showed in a few words. Do not
  enumerate suites, cases, or per-file results.
- At most one link per claim. A sentence carrying three inline references is
  reporting the search rather than the finding.

### Formatting

- Reserve bold for the status line, a blocker, and a required action. Bolding
  every bullet makes bold meaningless and the report harder to scan.
- Do not open a bullet with a bold restatement of its own content.
- Use bullets, never Markdown pipe tables.
- Write full sentences. Do not compress into fragments packed with identifiers.
- Redact secrets, credentials, private customer data, and signed URLs.

### Deviations

Omit this section entirely when nothing diverged. Do not emit an empty section
or a None bullet.

Include it — always, when the status is PARTIAL, BLOCKED, or FAIL — with one
bullet per material divergence:

- blocker or unresolved failure;
- incomplete required scope;
- warning or degraded execution;
- pending, unknown, skipped, unavailable, or not-applicable evidence when it
  changes how the outcome should be read;
- an unperformed action that an operator could otherwise mistake as complete;
- stale, mismatched, or lower-confidence evidence.

Name the observed state and its impact in one sentence. Preserve literal native
states such as `PENDING`, `UNKNOWN`, `SKIPPED`, `NOT_APPLICABLE`, and
provider-specific states. A planned exclusion is not a deviation. Do not repeat
a deviation in What happened unless the outcome would otherwise be misleading.

This section surfaces what an operator must know. It is not a place to
pre-empt every possible question: a limitation that was expected, is recorded
elsewhere, and changes nothing about what to do next does not belong here.

### Next Steps

Omit this section entirely when no action remains. Do not emit an empty section
or a None bullet; a PASS status with no Next steps already says that nothing is
needed.

When action remains:

- Put one independently actionable item in each bullet, required before
  optional.
- Start with `Required` or `Optional` and name the responsible actor when it is
  not obvious.
- State the action in one sentence. Every required follow-up includes a
  copy-ready prompt or command on a nested line.
- Never manufacture an action to fill the section. Speculative offers of work
  the user has not asked for are not next steps.

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

Task type decides which facts earn a place, not how many sections the report
has. Answer the reader's actual question and stop.

- **Implementation and delivery:** What changed and where it landed — issue,
  branch, pull request — and the validation that ran. Gaps go under Deviations
  and operator actions under Next steps.
- **Research and discovery:** The answer, what it rests on, and how confident
  it is. Keep sourced fact separate from inference.
- **Diagnosis and troubleshooting:** The confirmed cause or the current
  hypothesis, said plainly as one or the other, with its evidence and impact.
  Do not label a hypothesis as a confirmed root cause.
- **Planning and design:** The approach chosen, why, and the observable signals
  that will show it worked. Unresolved decisions are deviations with next
  steps.
- **Validation and testing:** What was checked and what it showed. Keep local,
  hosted, deployment, runtime, integration, physical, and business acceptance
  claims distinct.
- **Review and audit:** Findings worst first, each with a tight location and
  the remediation. A clean review states what was inspected and what it could
  not cover.
- **Operations, deployment, and monitoring:** The exact target and version,
  what happened to it, its literal state, and the recovery boundary. Keep
  deployment, runtime health, integration behavior, and production acceptance
  distinct.
- **Coordination and handoff:** Who owns what next and what blocks them. Keep
  task outcome separate from degraded orchestration conformance.
- **Fallback:** What was asked, what came back, and its limits.

## Composition With Existing Contracts

- `github-pr-delivery` names issue, branch, commit, pull request, and assignee
  as required facts; they belong under What happened and the density budget
  governs only their phrasing, never whether they appear. Pending checks or
  delivery gaps belong under Deviations; review or merge actions belong under
  Next steps.
- `testing-and-environment-validation` results belong under What happened;
  unavailable, pending, skipped, partial, or blocked evidence belongs under
  Deviations.
- Repository-memory decision, rationale, and artifacts become one concise What
  happened bullet when that contract requires them.
- `agent-team-orchestration` task outcome belongs in the status line; material
  execution facts belong under What happened and degraded conformance belongs
  under Deviations.
- Cross-repository program evidence belongs under What happened, grouped by
  workstream; unresolved dependencies belong under Deviations and Next steps.
- A composing contract names which facts must survive, never how many lines
  they may occupy. When such a contract lists fields, satisfy it with the
  shortest phrasing that keeps each fact recoverable.
- Higher-priority system, developer, client, tool, or host schemas take
  precedence. Preserve the same semantic groups inside the wrapper.
- For merge or release orchestration, keep What happened to state changes and
  the smallest evidence set that proves each terminal node. Put actionable
  follow-up under Next steps. Do not include a chronological command log,
  repeated checks, unchanged polling, or routine tool details.

## Anti-Patterns

- Manufacturing structured output for an ordinary conversational answer.
- Opening every bullet with a bold restatement of itself.
- A bullet running three or more sentences, or carrying three or more links.
- Listing test names, suite names, or per-file results instead of naming the
  validation and its outcome.
- Citing a commit SHA or file-and-line reference as proof that something was
  checked, where no composing contract requires the identifier.
- Emitting `## Deviations` or `## Next steps` with a None bullet.
- Using Deviations to pre-empt every possible question rather than to surface
  what the operator must know.
- Offering speculative follow-on work as a next step.
- Restating in the report what the linked spec or pull-request body holds.
- Adding a separate status heading, or Completed, Validation, Delivery, Feature
  State, Residual Notes, Coordination, Repository Memory, or task-profile
  headings.
- Repeating a deployment or validation fact in multiple sections, or hiding
  blockers among successful outcome bullets.
- Reporting PASS while required validation is failing, pending, or unobserved,
  or replacing a provider-native evidence state with an optimistic summary.
- Naming a required action without its owner and copy-ready continuation.
- Using tables or multi-level nesting.
- Expanding a merge or deployment result into a chronological work log.

## Examples

Small conversational answer:

```markdown
“Refresh checks on the final commit” means rerun the required checks after the
last PR update, so the results apply to the exact revision being reviewed.
```

Substantial implementation, nothing diverged and nothing remains:

```markdown
## What happened

**Status: PASS — completion output is proportional and ready for review.**

The rule now reserves structured reporting for substantial work and caps how
much a structured report may carry. Go and race suites pass. Delivery is ready
in PR #123, and the feature spec records why the density limits exist.
```

Complex production coordination, where separate bullets earn their place:

```markdown
## What happened

**Status: PASS — PRs #290, #181, and #230 merged and their deployments completed.**

LabCore released as v0.58.0 and is healthy in production, with the UI and docs
deployments completing behind it. Production validation passed, and manual
upsert stayed default-off as intended.

## Deviations

- The deployment workflow logged a Node 20 deprecation and an unsupported input;
  neither changed the outcome, but both need attention before the next runner
  upgrade.
```

Blocked diagnosis:

```markdown
## What happened

**Status: BLOCKED — the production root cause cannot be confirmed with the evidence available.**

The request fails somewhere past the service boundary and the logs stop there.
Everything on this side of that boundary behaves correctly.

## Deviations

- Read-only production dependency logs are unavailable, so the root cause
  remains `UNKNOWN`.

## Next steps

- **Required — User:** Grant read-only production log access.
  - Continue with: `Resume diagnosis using the authorized production logs.`
```

## Verification

- Confirm ordinary conversation contains none of the canonical headings or
  status tokens.
- Confirm a structured response begins with What happened and uses only the
  canonical headings, in order.
- Confirm Deviations and Next steps are absent when empty rather than carrying
  a None bullet, and that a PARTIAL, BLOCKED, or FAIL status carries Deviations.
- Confirm the first line of What happened contains one literal overall status,
  and that prose follows it before any bullet.
- Confirm no bullet exceeds one sentence plus one evidence clause, that nesting
  never exceeds one level, and that bold marks only the status line, blockers,
  and required actions.
- Confirm no task-profile or evidence-specific heading is emitted, and no
  Markdown pipe table appears.
- Confirm task-specific required facts remain present once, and required
  follow-up is owner-specific and copy-ready.
- Confirm native pending, unknown, skipped, unavailable, and not-applicable
  states remain literal.
