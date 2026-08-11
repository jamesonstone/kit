---
kind: workflow
slug: cross-repository-program-coordination
description: Coordinate dependent multi-repository work from one reconciled program ledger and ready frontier.
dependencies:
  - implementation-delivery
rules:
  - slug: coding-agent-context-usage
    required: true
  - slug: cross-repository-program-coordination
    required: true
  - slug: agent-team-orchestration
    required: true
  - slug: testing-and-environment-validation
    required: true
evidence:
  - kind: routing
    path: docs/agents/TOOLING.md
    required: true
  - kind: strategy
    path: docs/agents/RLM.md
    required: true
---
# Workflow: Cross-Repository Program Coordination

## Purpose

- Route a qualifying multi-repository program through one coordinator-owned ledger.
- Preserve participant repositories as the authority for their local requirements, implementation, delivery, and validation.

## Phases

1. Resolve the canonical program ledger and reconcile it against live participant evidence.
2. Compute the current ready frontier from explicit dependencies and gates.
3. Dispatch only independent ready work under agent-team concurrency limits.
4. Checkpoint every material implementation, delivery, deployment, validation, blocker, approval, or handoff transition.

## Completion Gates

- Every completion claim points to current repository, GitHub, runtime, and validation evidence as applicable.
- Stale or unobserved state is unknown, never assumed complete.
- One accountable program supervisor owns integration, handoff, and final acceptance.
