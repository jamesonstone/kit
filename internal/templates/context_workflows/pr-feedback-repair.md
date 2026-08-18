---
kind: workflow
slug: pr-feedback-repair
description: Verify and repair current pull-request feedback in the exact writable PR-head lane.
dependencies:
  - implementation-delivery
rules:
  - slug: coding-agent-context-usage
    required: true
  - slug: agent-completion-output
    required: true
  - slug: deletion-safety
    required: true
  - slug: agent-team-orchestration
    required: true
  - slug: github-pr-delivery
    required: true
  - slug: work-lane-gating
    required: true
  - slug: testing-and-environment-validation
    required: true
evidence:
  - kind: routing
    path: docs/agents/TOOLING.md
    required: true
  - kind: guardrails
    path: docs/agents/GUARDRAILS.md
    required: true
---
# Workflow: PR Feedback Repair

## Purpose

- Preserve the supervisor contract produced by `kit pr fix` without Kit launching an agent.
- Repair only current, verified findings on the exact same-repository PR-head branch.

## Phases

1. Use `kit pr fix` for bounded current feedback intake and lane evidence.
2. Negotiate host-confirmed agent controls and record an Agent Team Plan that distinguishes actual agents from logical and omitted lanes, requested from effective profiles, confirmed from unconfirmed parallelism, and reused agents from replacement rebriefs; serialize shared files.
3. Verify every finding against current HEAD and fix only still-valid issues.
4. Run complete validation and use a fresh independent read-only verifier when supported; otherwise perform and disclose a distinct supervisor self-review.
5. Review the full integrated diff, push one coherent batch, verify the exact remote head, reflect, then explicitly resolve only addressed threads.

## Completion Gates

- The writable lane, expected head, push target, and dirty-change ownership are explicit.
- Stale, false-positive, out-of-scope, and human-needed findings are reported rather than silently changed.
- Kit itself did not edit source, stage, commit, push, comment, resolve, or merge from the prompt-producing path.
