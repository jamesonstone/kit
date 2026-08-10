---
kind: workflow
slug: repository-bootstrap
description: Establish repository-local evidence and durable project memory from verified repository sources.
rules:
  - slug: coding-agent-context-usage
    required: true
  - slug: constitution-curation
    required: true
  - slug: kit-capabilities-usage
    required: true
evidence:
  - kind: routing
    path: docs/agents/README.md
    required: true
  - kind: strategy
    path: docs/agents/RLM.md
    required: true
  - kind: project-memory
    path: docs/CONSTITUTION.md
    required: true
  - kind: project-index
    path: docs/PROJECT_PROGRESS_SUMMARY.md
    required: false
---
# Workflow: Repository Bootstrap

## Purpose

- Help a coding agent inspect the real repository before curating project context.
- Keep Kit deterministic: the agent establishes truth; Kit only resolves local evidence.

## Phases

1. Run `kit capabilities init --json`, then resolve this workflow.
2. Start with routing and indices. Inspect only relevant manifests, build scripts, CI, tests, docs, code boundaries, specs, and verified external-system evidence.
3. Populate durable repository memory only from demonstrated evidence.
4. Run verified safe commands and record gaps literally.

## Completion Gates

- Constitution changes contain only durable project-wide truth.
- Testing, tooling, external-system, Makefile, and README guidance is repository-backed.
- Existing project prose and local customization remain intact.
