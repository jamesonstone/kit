---
kind: workflow
slug: repository-maintenance
description: Review structural and semantic repository drift without changing kit reconcile semantics.
rules:
  - slug: coding-agent-context-usage
    required: true
  - slug: constitution-curation
    required: true
  - slug: testing-and-environment-validation
    required: true
  - slug: source-file-size
    required: true
evidence:
  - kind: routing
    path: docs/agents/README.md
    required: true
  - kind: maintenance-strategy
    path: docs/agents/WORKFLOWS.md
    required: true
  - kind: project-memory
    path: docs/CONSTITUTION.md
    required: true
  - kind: project-index
    path: docs/PROJECT_PROGRESS_SUMMARY.md
    required: false
---
# Workflow: Repository Maintenance

## Purpose

- Preserve `kit reconcile` as the structural audit and refresh interface.
- Provide agent-consumed semantic maintenance guidance without adding a second reconciliation command.

## Phases

1. Inspect `kit status`, registry status, reconcile preview, health preview, and project checks using their live capability metadata.
2. Apply only conflict-free approved structural maintenance through the existing commands.
3. Review completed work and current evidence for missed or stale project-wide invariants.
4. Curate the Constitution and reusable references only when evidence demonstrates a durable change.

## Completion Gates

- Structural command behavior remains within its existing ownership boundary.
- Semantic documentation is reviewed, evidence-backed, and not an automatic rewrite.
- Existing local customization and unrelated project content remain preserved.
