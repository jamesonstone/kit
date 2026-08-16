---
kind: workflow
slug: implementation-delivery
description: Route a feature from accepted native planning through implementation, validation, memory curation, and delivery.
rules:
  - slug: coding-agent-context-usage
    required: true
  - slug: deletion-safety
    required: true
  - slug: testing-and-environment-validation
    required: true
  - slug: source-file-size
    required: true
  - slug: work-lane-gating
    required: true
  - slug: agent-team-orchestration
    required: false
  - slug: github-pr-delivery
    required: false
evidence:
  - kind: routing
    path: docs/agents/README.md
    required: true
  - kind: strategy
    path: docs/agents/WORKFLOWS.md
    required: true
  - kind: guardrails
    path: docs/agents/GUARDRAILS.md
    required: true
  - kind: project-memory
    path: docs/CONSTITUTION.md
    required: true
  - kind: implementation-patterns
    path: docs/references/tooling.md
    required: false
---
# Workflow: Implementation Delivery

## Purpose

- Turn an accepted native plan into a correct implementation with durable rationale and literal validation.
- Keep the coding agent accountable for repository truth, integration, and delivery.

## Phases

1. Resolve the feature context before source edits and adopt or create its living `SPEC.md` when material rationale exists.
2. Implement the smallest complete change while maintaining consequential decisions and discoveries.
3. Validate the integrated behavior at the narrowest useful layers and broader affected boundaries.
4. Reflect on the complete diff, curate repository memory, and satisfy the repository delivery gate.

## Completion Gates

- Code, tests, docs, resolved context, and the feature outcome agree.
- Required checks ran or have a literal blocker and risk statement.
- No known correctness, security, source-size, ownership, or delivery gap remains.
