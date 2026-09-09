---
kind: ruleset
slug: agent-completion-output
description: Removes every required completion format and keeps only the facts a terminal response must not leave out.
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

- Let the agent write every response in its own natural shape, including
  terminal ones.
- Keep a short list of facts from going missing once the shape is free.
- Retire the three-section envelope, the status token, the None items, and the
  density budget. Each was added to make reports scannable and each instead
  produced uniform, ceremonial output that read worse than plain writing.

## Applies When

Applies to every human-readable terminal completion or handoff response, and
to ordinary conversation.

Does not govern tool-native JSON, machine-only protocol output, or any host
schema that specifies its own response shape.

## Rules

### Write In Your Own Shape

There is no required format. No mandatory headings, no status token, no fixed
section order, no required bullet style, no ban on tables, no length budget,
and no rule that empty things must be declared empty.

- Choose the shape the content calls for: a sentence for a small answer, a
  short paragraph for a result, a list when the items are genuinely parallel,
  a heading only when a reader needs to navigate between parts.
- Match length to consequence rather than to effort spent. Substantial work
  can warrant a short report, and often does.
- Do not carry retired scaffolding forward out of habit, and do not invent a
  fixed replacement template and apply it to every task.

### What Must Not Go Missing

Whatever shape a terminal response takes, it must not leave the reader wrong
about any of the following. State each one plainly, in whatever words fit.

- **Blockers.** Anything preventing completion, and what would clear it.
- **Incomplete scope.** Any part of the request not done, and why.
- **Required action.** Anything the reader must do next, said so they can act
  without reconstructing context. When it is a command or a prompt, give it
  exactly.
- **Failed and unobserved checks.** A failing, pending, unavailable, skipped,
  or never-run check reported as exactly that, never folded into success.
  Preserve literal provider states such as `PENDING`, `UNKNOWN`, `SKIPPED`,
  and `NOT_APPLICABLE`.
- **Mutations.** Repository, delivery, external-system, and infrastructure
  changes actually made, with the identifiers a reader needs to find or undo
  them.
- **Confidence.** A hypothesis said as a hypothesis, an inference as an
  inference, and a verified fact as verified.

Say plainly whether the work is finished, partly finished, blocked, or failed.
Say it in ordinary words; no token, label, or taxonomy is required.

### Honesty

- Never report a check as passing unless it ran and passed.
- Never describe a file, system, or state as inspected unless it was.
- Never smooth a partial or blocked result into a clean one.
- When something could not be validated, say so and say why.

### Evidence

Include an identifier when the reader needs it to act, or would reasonably
doubt the claim without it. Leave it out otherwise. A report is an account of
where things stand, not an index of everything that was checked.

- Include the pull request, issue, or branch a reader will open next; the exact
  target and version of anything deployed; the specific failure and where it
  lives; the command that resumes blocked work.
- Leave out a commit SHA cited as proof that something was checked rather than
  as an identifier a delivery contract requires, lists of test or suite names,
  file-and-line references for code that is correct, counts of things
  inspected, and the tools and steps used to reach the answer.
- Name a validation and what it showed rather than enumerating suites, cases,
  or per-file results.
- For merge or release orchestration, report state changes and the
  smallest evidence set that proves each terminal node.
- Do not include a chronological command log, repeated checks, or unchanged
  polling history.
- Redact secrets, credentials, private customer data, and signed URLs.

## Composition With Existing Contracts

- Delivery, validation, orchestration, program, and repository-memory
  contracts still bind. They name facts that must reach the reader; none of
  them dictates layout, and none may be satisfied by a heading alone.
- `github-pr-delivery` requires the issue, branch, commit, pull request, and
  assignee to be recoverable from the response. Where they appear is free.
- `testing-and-environment-validation` requires observed results and every
  non-passing or unavailable evidence state to be visible and distinct.
- `agent-team-orchestration` requires the task outcome and any degraded or
  unsatisfied conformance to be stated, never hidden behind task success.
- Cross-repository program work requires each workstream's state, unresolved
  dependencies, and exact handoffs to be identifiable.
- Repository-memory decisions, including `not required`, must be stated once.
- A higher-priority system, developer, client, tool, or host schema takes
  precedence over this rule.

## Anti-Patterns

- Reproducing the retired envelope: `## What happened`, `## Deviations`, and
  `## Next steps` headings, a `**Status: ...**` token, or a `**None.**` item.
- Replacing it with a different fixed template applied to every task.
- Adding Completed, Validation, Delivery, Feature State, Residual Notes,
  Coordination, or Repository Memory headings.
- Announcing that there are no deviations, no risks, or no next steps.
- Reporting success while a required check is failing, pending, or unobserved.
- Replacing a provider-native evidence state with an optimistic summary.
- Burying a blocker or unfinished scope among successful-sounding detail.
- Naming a required action without enough context to act on it.
- Padding a report with commit SHAs, suite names, or file-and-line references
  that support no decision.

## Examples

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

Orchestrated work where headings genuinely help the reader navigate:

```markdown
All three PRs merged and deployed.

**LabCore** — released as v0.58.0, healthy in production.
**UI and docs** — deployed behind it, both serving.

One thing to watch: the deployment workflow logged a Node 20 deprecation and
an unsupported input. Neither changed the outcome, but both will break on the
next runner upgrade.
```

## Verification

- Confirm no response contains the retired headings, status token, or None
  item.
- Confirm no single fixed template is being applied across unrelated tasks.
- Confirm blockers, incomplete scope, and required actions are stated plainly
  and are not hidden among successful detail.
- Confirm failing, pending, unavailable, skipped, and unobserved checks are
  reported as those states, with provider-native wording preserved.
- Confirm mutations are recoverable from the response.
- Confirm hypotheses and inferences are not presented as verified facts.
- Confirm every fact required by a composing contract reaches the reader.
- Confirm identifiers that support no decision were left out.
