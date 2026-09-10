---
kind: ruleset
slug: agent-completion-output
description: States what a terminal response must convey and leaves its shape entirely to the agent.
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

- Leave the shape of every response, including terminal ones, to the agent.
- Name the facts that must survive whatever shape it chooses.

## Applies When

Applies to every human-readable terminal completion or handoff response, and
to ordinary conversation.

A host schema that specifies its own response shape takes precedence. This rule
does not govern tool-native JSON or machine-only protocol output.

## Rules

### Shape

Write each response in the shape its content calls for. A sentence, a
paragraph, a list, a heading, a table — whichever carries the meaning. Match
length to consequence rather than to effort spent; substantial work often
warrants a short account.

### What A Terminal Response Conveys

- What the user now has.
- What remains unfinished, and why.
- Anything blocking completion, and what would clear it.
- Anything the reader must do next, with enough context to act on it and the
  exact command or prompt when there is one.
- Whether the work is finished, partly finished, blocked, or failed, said
  plainly.
- Every repository, delivery, external-system, and infrastructure change made,
  with the identifiers a reader needs to find or undo it.

Blockers and unfinished scope belong where the reader will see them, as
prominent as the successes.

### Reporting State Truthfully

- Report each check as observed: a failing, pending, unavailable, skipped, or
  never-run check is reported as exactly that.
- Preserve literal provider states such as `PENDING`, `UNKNOWN`, `SKIPPED`, and
  `NOT_APPLICABLE`.
- Report a check as passing only when it ran and passed, and a file or system
  as inspected only when it was inspected.
- When something could not be validated, say so and say why.
- Distinguish a verified fact from an inference and from a hypothesis.

### Evidence

A response is an account of where things stand, not an index of everything
checked. Include an identifier when the reader needs it to act, or would
reasonably doubt the claim without it.

- Include the pull request, issue, or branch a reader will open next; the exact
  target and version of anything deployed; the specific failure and where it
  lives; the command that resumes blocked work.
- Name a validation and what it showed, in a few words.
- For merge or release orchestration, report state changes and the
  smallest evidence set that proves each terminal node.
- Redact secrets, credentials, private customer data, and signed URLs.

## Composition With Existing Contracts

Delivery, validation, orchestration, program, and repository-memory contracts
name facts that must reach the reader. Satisfy them on content; they say
nothing about layout, and a heading alone satisfies none of them.

- `github-pr-delivery` requires the issue, branch, commit, pull request, and
  assignee to be recoverable from the response.
- `testing-and-environment-validation` requires observed results and every
  non-passing or unavailable evidence state to be visible and distinct.
- `agent-team-orchestration` requires the task outcome, and any degraded or
  unsatisfied conformance, to be stated in its own right.
- Cross-repository program work requires each workstream's state, unresolved
  dependencies, and exact handoffs to be identifiable.
- Repository-memory decisions, including `not required`, are stated once.

## Anti-Patterns

These are failures of content, not of layout. No shape is wrong here; these
are.

- Reporting a check as passing when it did not run, or claiming an inspection
  that did not happen.
- Folding a pending, unavailable, or skipped state into a summary that reads
  as success.
- Presenting a hypothesis or an inference as a confirmed fact.
- Leaving a blocker or unfinished scope to be inferred from successful-sounding
  detail.
- Naming a required action without enough context to act on it.
- Padding a response with identifiers that support no decision the reader has
  to make.
- Turning a merge or deployment result into a chronological command log or a
  polling history.
- Carrying one response template across unrelated tasks instead of letting each
  response take the shape its content calls for.

## Examples

These differ in shape because their content differs. That is the point.

A small answer stays a small answer:

```markdown
“Refresh checks on the final commit” means rerun the required checks after the
last PR update, so the results apply to the exact revision being reviewed.
```

A clean delivery, said once:

```markdown
The metadata gate is in on both admission paths, so a run can't be listed or
created without complete metadata. Domain, full, and race suites pass, and
it's ready for review in PR #204.
```

Partial work, where the gap leads:

```markdown
The listing path is fixed and covered, but run creation still admits
incomplete metadata — `CreateManualHybridRun` never reaches the gate. I left
it alone because blocking there and failing at persistence are different
product decisions.

Tell me which you want and I'll finish it.
```

A blocker, with the unblock:

```markdown
I can't confirm the root cause. The request fails past the service boundary
and the logs stop there; everything on this side behaves correctly.

I need read-only access to the production dependency logs — once you've
granted it, say “resume diagnosis using the authorized production logs.”
```

Orchestrated work, where headings genuinely help the reader navigate:

```markdown
All three PRs merged and deployed.

**LabCore** — released as v0.58.0, healthy in production.
**UI and docs** — deployed behind it, both serving.

One thing to watch: the deployment workflow logged a Node 20 deprecation and
an unsupported input. Neither changed the outcome, but both will break on the
next runner upgrade.
```

## Verification

- Confirm each response's shape follows from its content rather than from a
  template carried across tasks.
- Confirm blockers, unfinished scope, and required actions are as visible as
  the successes.
- Confirm every check's state is reported as observed, with provider-native
  wording preserved.
- Confirm passing claims correspond to checks that ran, and inspection claims
  to inspections that happened.
- Confirm hypotheses and inferences read as such.
- Confirm every mutation is recoverable from the response.
- Confirm every fact required by a composing contract reaches the reader.
- Confirm the identifiers present support a decision the reader has to make.
