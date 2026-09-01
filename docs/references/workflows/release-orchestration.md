---
kind: workflow
slug: release-orchestration
description: Build and execute an authority-aware release graph across scoped repositories while preserving repository, infrastructure, and production evidence boundaries.
dependencies:
  - implementation-delivery
rules:
  - slug: coding-agent-context-usage
    required: true
  - slug: agent-completion-output
    required: true
  - slug: deletion-safety
    required: true
  - slug: github-pr-merge
    required: true
  - slug: testing-and-environment-validation
    required: true
  - slug: infrastructure-change-approval
    required: false
  - slug: agent-team-orchestration
    required: false
  - slug: cross-repository-program-coordination
    required: false
evidence:
  - kind: routing
    path: docs/agents/TOOLING.md
    required: true
  - kind: guardrails
    path: docs/agents/GUARDRAILS.md
    required: true
  - kind: project-memory
    path: docs/CONSTITUTION.md
    required: true
---
# Workflow: Release Orchestration

## Purpose

- Turn bounded repository scope into a dependency-aware Global Release Graph and exact authorized merge plan.
- Keep merge, hosted workflow, deployment, runtime, production, and integrated-system evidence separate.

## Applicability Boundaries

- Load infrastructure approval to classify public-cloud, Kubernetes, or IaC
  effects. Proceed autonomously for additive and routine application
  operations. Exact confirmation is required only for delete, remove, or
  otherwise destructive effects. Image-only CD is not a destructive trigger.
- Load agent-team topology only when work safely separates into bounded lanes.
- Create or adopt a cross-repository program ledger only when the existing coordination trigger applies: multiple repositories plus dependent deliverables, staged deployment or activation, or expected agent/session handoff.

## Phases

1. Discover the bounded release set, repository policy, identities, dependencies, runtime relationships, deployment effects, and verification ownership without mutation.
2. Construct the Global Release Graph and bounded merge plan; record the authorization source and exact approved PR set before any merge.
3. Remediate review, conflict, compatibility, migration, and validation
   blockers through repository-owned delivery lanes. Keep routine,
   scope-preserving repairs as ordinary, non-history-rewriting updates on their
   existing pull requests; do not rebase, force-push, retarget, or otherwise
   replace reviewed history. Reserve replacement PRs for material scope changes,
   heads that cannot be updated safely, or explicit repository-policy or user
   requirements.
4. Resolve `pull-request-merge`, recompute the current in-scope `MERGE_READY` frontier, and execute only dependency-safe merge and deployment waves.
5. Verify the exact merged and deployed identities, runtime behavior, and final integrated system; preserve partial, blocked, unknown, and intentionally open state literally.

## Completion Gates

- Every merged or queued PR belonged to the exact in-scope set and current ready frontier.
- Every destructive infrastructure effect had applicable target, impact, recovery,
  confirmation, and post-change evidence. Additive and routine application operations record
  target and post-change evidence without a confirmation batch.
- Every changed existing head received fresh checks, review, and revalidation before merge; exceptional replacement PRs and newly
  discovered out-of-scope targets received clarification.
- Merge success was never substituted for hosted workflow, deployment, runtime, production, or integrated-system proof.
