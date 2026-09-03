---
kind: workflow
slug: pull-request-merge
description: Reconcile accepted-task scope, policy, identity, current-head evidence, destructive-effect classification, and dependency gates before exact pull-request merges.
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

- Execute in-scope pull-request merges authorized by an accepted task or
  active `/goal` once they are `MERGE_READY`. Do not stop for a separate
  merge-consent prompt.
- Fail closed on unknown actor, policy, head/base, checks, dependencies,
  explicit holds, or unresolved destructive effects. Record routine
  application operations such as image-only CD without inventing a
  confirmation batch.

## Phases

1. Record the accepted-scope source, in-scope PR set, expected actor, head/base,
   merge method, dependencies, and destructive versus non-destructive effects.
2. Classify each node as `MERGE_READY`, `BLOCKED`, or `UNKNOWN` from exact
   current-head and repository-policy evidence.
3. Reconcile the in-scope ready frontier immediately before each wave;
   serialize dependencies and same-base sensitive operations.
4. When in-scope routine remediation is required, use ordinary,
   non-history-rewriting commits to update the existing pull-request head
   between waves, invalidate its prior head evidence, and
   return it to `UNKNOWN` pending fresh checks, review, and revalidation.
   Do not request a new consent prompt for in-scope repair. Reserve replacement pull
   requests for material scope changes, heads that cannot be updated safely,
   or explicit repository-policy or user requirements. Do not rebase,
   force-push, retarget, or otherwise replace the branch's reviewed history.
5. Merge or queue only assigned `MERGE_READY` nodes, preserving partial state
   and isolating failures.
6. Record merge, hosted workflow, deployment/runtime, and production evidence
   as separate claims; recalculate the next frontier.

## Completion Gates

- Every merged or queued node belonged to the in-scope set and current ready
  frontier.
- No protection, review, required-check, merge-method, or identity
  safeguard was bypassed.
- Partial failures, unknowns, dependents, corrective ownership, and next safe
  actions are exact.
- Routine scope-preserving remediation stays on its existing pull request, and
  no changed head reuses readiness, review, or checks from its
  predecessor.
- Merge success is not reported as deployment or production proof.
