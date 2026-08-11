---
kind: workflow
slug: repository-maintenance
description: Resolve repository maintenance evidence without changing reconcile semantics.
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
# Workflow: Repository Maintenance

Inspect repository evidence before proposing semantic maintenance.
