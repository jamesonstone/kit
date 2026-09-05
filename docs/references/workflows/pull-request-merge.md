---
kind: workflow
slug: pull-request-merge
description: Resolve bounded standing authority to exact current pull requests and reconcile policy, identity, readiness, dependencies, deployment, and risk gates.
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

- Execute exact current pull-request merges covered by explicit bounded standing
  authority once they are `MERGE_READY`. Later in-scope PRs and refreshed heads
  inside the selector do not require renewed merge authority after exact-current
  `MERGE_READY` evidence.
- Agent-performed in-place repair requires the standing-authority permitted
  actions to explicitly include blocker repair; otherwise stop before source,
  commit, or push and obtain renewed repair authority.
- Fail closed on absent, paused, revoked, expired, or mismatched authority;
  unknown actor, policy, head/base, checks, dependencies, environment, workflow,
  or effects; and every separately gated risk class.

## Phases

1. Record the standing-authority source and selector, current pause/revocation
   state, resolved exact PR set, expected actor, head/base, merge method,
   dependencies, deployment workflow and environment, and material effects.
2. Classify each node as `MERGE_READY`, `BLOCKED`, or `UNKNOWN` from exact
   current-head and repository-policy evidence.
3. Reconcile the in-scope ready frontier immediately before each wave;
   serialize dependencies and same-base sensitive operations.
4. Before agent-performed in-scope routine remediation, confirm the standing-
   authority permitted actions explicitly include blocker repair. Otherwise
   stop before source, commit, or push and obtain renewed repair authority. When
   authorized, use ordinary, non-history-rewriting commits to update the
   existing pull-request head between waves. Any changed or refreshed head,
   including a human or external update, returns to `UNKNOWN` pending fresh
   exact-current checks, review, and revalidation. When the selector still
   matches and the node returns to `MERGE_READY`, merge under standing merge
   authority without renewed merge authorization. Reserve replacement pull
   requests for material scope changes, heads that cannot be updated safely, or
   explicit repository-policy or user requirements. Do not rebase, force-push,
   retarget, or otherwise replace the branch's reviewed history.
5. Merge or queue only assigned `MERGE_READY` nodes, preserving partial state
   and isolating failures.
6. Record merge, hosted workflow, deployment/runtime, and production evidence
   as separate claims; recalculate the next frontier.

## Completion Gates

- Every merged or queued node matched active standing authority and belonged to
  the resolved exact current ready frontier.
- No protection, review, required-check, merge-method, or identity
  safeguard was bypassed.
- Partial failures, unknowns, dependents, corrective ownership, and next safe
  actions are exact.
- Routine scope-preserving remediation stays on its existing pull request, and
  no changed head reuses readiness, review, or checks from its
  predecessor.
- No IAM, network, KMS, secrets, database schema/data-loss, infrastructure
  create/replace/delete, destructive, or nonstandard deployment effect was
  treated as covered by standing merge/deploy authority.
- Merge success is not reported as deployment or production proof.
