---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0001
  slug: material-why
  dir: 0001-material-why
---
# SPEC

## PURPOSE

Preserve a product decision that future implementation work cannot infer from code alone.

## CONTEXT

Two viable approaches produce the same observable code shape but different future extension constraints.

## REQUIREMENTS

- Keep the selected extension boundary stable.
- Non-goal: preserve transient planning conversation.
- Observable acceptance: tests pass and the rationale remains discoverable in this spec.

## ACCEPTED PLAN

Use the narrower boundary, implement it, validate the behavior, and retain the rejected broader boundary with rationale.

## DECISIONS

- Accepted the narrow boundary because it minimizes public surface.
- Rejected the broad boundary because speculative extensibility adds maintenance cost.

## DISCOVERIES

No additional information required.

## VALIDATION

Focused tests and the repository document check passed.

## OUTCOME

The narrow boundary was implemented without divergence from the accepted plan.

## REPOSITORY MEMORY

- Decision: created
- Rationale: the rejected alternative and future extension constraint are not recoverable from code and tests.
- Artifacts: SPEC.md
