---
kind: workflow
slug: pull-request-merge
description: Reconcile authority, policy, identity, current-head evidence, infrastructure effects, and dependency gates before exact pull-request merges.
rules:
  - slug: safety-guardrails
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
    path: docs/agents/README.md
    required: true
  - kind: guardrails
    path: docs/agents/GUARDRAILS.md
    required: true
  - kind: project-memory
    path: docs/CONSTITUTION.md
    required: true
---
# Workflow: Pull-Request Merge

## Purpose

- Execute only the exact pull-request set authorized by a direct request or
  accepted bounded plan.
- Fail closed on unknown actor, policy, head/base, checks, dependencies,
  approvals, or indirect deployment effects.

## Phases

1. Record the authorization source, approved PR set, expected actor, head/base,
   merge method, dependencies, and applicable infrastructure approval.
2. Classify each node as `MERGE_READY`, `BLOCKED`, or `UNKNOWN` from exact
   current-head and repository-policy evidence.
3. Reconcile the authorized ready frontier immediately before each wave;
   serialize dependencies and same-base sensitive operations.
4. When separately authorized routine remediation is required, use ordinary,
   non-history-rewriting commits to update the existing pull-request head
   between waves, invalidate its prior head evidence and merge authority, and
   return it to `UNKNOWN` pending fresh checks, review, revalidation, and
   exact-head authorization. Reserve replacement pull
   requests for material scope changes, heads that cannot be updated safely,
   or explicit repository-policy or user requirements. Do not rebase,
   force-push, retarget, or otherwise replace the branch's reviewed history.
5. Merge or queue only assigned `MERGE_READY` nodes, preserving partial state
   and isolating failures.
6. Record merge, hosted workflow, deployment/runtime, and production evidence
   as separate claims; recalculate the next frontier.

## Completion Gates

- Every merged or queued node belonged to the approved set and current ready
  frontier.
- No protection, review, required-check, merge-method, approval, or identity
  safeguard was bypassed.
- Partial failures, unknowns, dependents, corrective ownership, and next safe
  actions are exact.
- Routine scope-preserving remediation stays on its existing pull request, and
  no changed head reuses readiness, review, checks, or merge authority from its
  predecessor.
- Merge success is not reported as deployment or production proof.
