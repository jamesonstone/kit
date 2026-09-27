---
kind: ruleset
slug: constitution-curation
description: Promotes only demonstrated project-wide truth into the Constitution.
status: active
registry_scope: downstream
applies_to:
  - repository-memory
  - constitution
read_policy_default: conditional
---

# Ruleset: constitution-curation

## Purpose

- Keep `docs/CONSTITUTION.md` a small, current record of project-wide invariants.

## Applies When

- After implementation and validation, when deciding what durable truth to record, or when the Constitution is stale.

## Rules

- The generated Constitution starter is a valid state; do not fill it with guesses, aspirations, or Kit scaffolding.
- Promote only demonstrated project-wide principles, constraints, non-goals, definitions, and workflow boundaries, supported by implementation, validation, or explicit human decisions.
- Keep feature rationale, rejected alternatives, and history in the relevant `SPEC.md`; leave out changelog entries, plans, and anything code and tests already communicate.
- When the implementation disproves a constitutional rule, correct or remove it and keep the historical rationale in the spec.
- Preserve the Kit-managed baseline block. When nothing project-wide changed, leave the Constitution alone.
