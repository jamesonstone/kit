---
kind: workflow
slug: implementation-delivery
description: Resolve the local implementation contract for deterministic evaluation.
dependencies: []
rules:
  - slug: coding-agent-context-usage
    required: true
evidence:
  - kind: routing
    path: docs/agents/README.md
    required: true
  - kind: project-memory
    path: docs/CONSTITUTION.md
    required: true
---
# Workflow: Implementation Delivery

Load the living feature specification and repository-local evidence before implementation.
