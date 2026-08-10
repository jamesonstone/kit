---
kind: ruleset
slug: feature-specification
description: Requires a complete living V3 feature specification before source implementation and through delivery.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - feature
  - implementation
  - planning
  - repository-memory
read_policy_default: must
---

# Ruleset: Feature Specification

## Purpose

- Preserve the accepted plan, evidence, decisions, discoveries, validation,
  outcome, and delivery history for every non-trivial feature.
- Give future coding agents a durable, progressively disclosed account that
  code and tests alone cannot reconstruct.
- Separate deterministic contract enforcement from semantic authoring: Kit
  validates structure and phase state, while the coding agent writes only
  repository-grounded truth.

## Applies When

- A coding agent implements a feature or another non-trivial behavior change.
- The task changes architecture, public behavior, cross-file responsibilities,
  migration behavior, or an accepted implementation plan.
- Genuinely trivial mechanical maintenance may omit a feature spec only when it
  has no material feature behavior, rationale, or historical consequence.

Resolve feature work with both `--workflow implementation-delivery` and
`--feature <feature-directory>`. Do not omit the feature hint to bypass this
rule.

## Rules

### Pre-source gate

- Before editing implementation/source or test files, create or adopt
  `docs/specs/<feature>/SPEC.md` from the accepted native plan and current
  repository evidence.
- Use living-spec workflow version 3 front matter with `artifact: spec`,
  `workflow_version: 3`, a matching `feature.dir`, and the current phase.
- The specification must contain non-empty sections for:
  - purpose;
  - context;
  - source evidence and historical relationships;
  - requirements;
  - non-goals;
  - observable acceptance criteria;
  - accepted plan;
  - architecture and decisions;
  - discoveries;
  - validation plan and map;
  - literal validation results;
  - implemented outcome;
  - delivery evidence;
  - repository-memory disposition.
- A missing, invalid, or incomplete spec permits spec authoring but blocks
  source implementation and delivery. Re-resolve after creating or updating
  the spec; do not treat a nonzero blocked result as permission for source work.

### Living implementation history

- Keep requirements and the accepted plan aligned with the authorized work.
- Record consequential decisions, alternatives, and superseded choices with
  rationale when they occur; do not reconstruct them only at the end.
- Record material repository discoveries while implementing, including
  conflicts between the plan and actual code or external evidence.
- Update the validation map as affected boundaries become known.
- Preserve relevant historical specifications. Link explicit relationships and
  use `docs/PROJECT_PROGRESS_SUMMARY.md`, `docs/specs/`, and repository evidence
  as RLM indices; do not load or rewrite all historical specs mechanically.

### Pre-delivery and completion gates

- Before delivery, reconcile the spec against the complete integrated diff.
  Record literal focused and complete validation, baseline debt, skipped or
  unavailable evidence, actual outcome, and exact delivery state.
- Advance the spec phase only when its documented evidence supports that phase.
  Delivery permission requires a structurally complete spec in `deliver` or
  `complete` phase.
- Before completion, curate durable project-wide invariants and reusable
  knowledge under the Constitution, references, or canonical domain docs, and
  record the exact repository-memory disposition in the feature spec.
- Preserve material superseded decisions and historical relationships while
  removing transient planning chatter and code-recoverable detail.

### Repository-memory boundaries

- Keep feature-specific history in the feature spec. Put demonstrated durable
  project-wide truth in `docs/CONSTITUTION.md`, reusable practice in
  `docs/references/`, and domain truth in its canonical documentation.
- Do not use `docs/notes` as an active Kit convention. Existing project-owned
  note directories and historical specifications are not migration deletion
  targets and must not be rewritten merely because the convention retired.
- Kit does not author the spec, infer project truth, execute the accepted plan,
  or launch or supervise an agent.

## Anti-Patterns

- Treating an empty `docs/specs/` directory as preserved feature history.
- Starting source edits from an in-memory or chat-only plan.
- Adding headings with empty bodies solely to satisfy structural validation.
- Omitting `--feature` from resolution to avoid a blocked spec state.
- Reconstructing decisions, validation, or delivery evidence after context is
  lost, or silently claiming unobserved validation passed.
- Deleting downstream notes or rewriting historical specs during migration.

## Verification

- Confirm the feature resolution reports the expected canonical spec path,
  required sections, phase permissions, history indices, and related specs.
- Confirm source implementation was not started while its permission was
  blocked and that the contract was re-resolved after spec changes.
- Confirm the final spec matches the integrated diff and records literal
  validation, outcome, delivery evidence, and repository-memory disposition.
- Confirm historical specifications remain intact and discoverable.

## Examples

```bash
kit contract resolve \
  --workflow implementation-delivery \
  --feature 0058-coding-agent-first \
  --json
```

If `feature_spec.phase_permissions.source_implementation` is false, author or
repair the reported spec first, then run the same local read-only resolution
again before touching source.
