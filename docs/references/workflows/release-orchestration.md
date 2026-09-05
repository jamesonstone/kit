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

- Resolve explicit bounded standing authority into a dependency-aware Global Release Graph and exact current merge/deployment plan.
- Keep merge, hosted workflow, deployment, runtime, production, and integrated-system evidence separate.

## Applicability Boundaries

- Standing deployment authority covers only recorded existing standard
  workflows, environments, actors, and exact merged artifacts on already-
  provisioned targets. Load infrastructure approval for IAM, network, KMS,
  secrets, database schema/data-loss, infrastructure create/replace/delete,
  destructive, and nonstandard deployment effects.
- Load agent-team topology only when work safely separates into bounded lanes.
- Create or adopt a cross-repository program ledger only when the existing coordination trigger applies: multiple repositories plus dependent deliverables, staged deployment or activation, or expected agent/session handoff.

## Phases

1. Discover the bounded release set, repository policy, identities, dependencies, runtime relationships, deployment effects, and verification ownership without mutation.
2. Construct the Global Release Graph; record the standing-authority selector,
   pause/revocation state, and dynamically resolved exact current PR set before
   each merge wave.
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

- Every merged or queued PR matched active standing authority and belonged to
  the resolved exact current ready frontier.
- Every standard deployment matched the recorded workflow and environment and
  completed deployed-identity and runtime verification. Every separately gated
  infrastructure, security, data, destructive, or nonstandard effect retained
  its applicable approval and post-change evidence.
- Every changed existing head received fresh checks, review, and revalidation
  before returning to `MERGE_READY`, including heads changed by a human or
  external system. Selector-matching nodes retained standing merge authority
  without renewed merge authorization.
- Every agent-performed in-place repair had explicit blocker-repair permission;
  otherwise source, commit, and push stopped for renewed repair authority.
  Exceptional replacement PRs and material expansions received explicit
  updated authority.
- Merge success was never substituted for hosted workflow, deployment, runtime, production, or integrated-system proof.
