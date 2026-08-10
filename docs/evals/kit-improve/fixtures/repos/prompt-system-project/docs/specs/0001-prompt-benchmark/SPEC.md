---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0001
  slug: prompt-benchmark
  dir: 0001-prompt-benchmark
references:
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: repository invariants
    status: active
---
# SPEC

## PURPOSE

Add a deterministic sample feature that exercises generated prompt contracts.

## CONTEXT

The fixture is intentionally small and contains no external dependencies.

## REQUIREMENTS

- Preserve the requested output contract.
- Validate behavior before completion.
- Non-goal: invoke a model or external system.
- Observable acceptance: context resolution includes this specification and its required local evidence.

## ACCEPTED PLAN

Implement the smallest change, run focused checks, and record evidence.

## DECISIONS

Use deterministic local evidence so the fixture can run offline.

## DISCOVERIES

No additional information required.

## VALIDATION

Resolve the implementation-delivery workflow and verify this path appears in the versioned JSON contract.

## OUTCOME

The fixture provides a complete living V3 specification for deterministic context evaluation.

## REPOSITORY MEMORY

- Decision: updated
- Rationale: the agent-contract benchmark must exercise the current living-spec shape.
- Artifacts: SPEC.md
