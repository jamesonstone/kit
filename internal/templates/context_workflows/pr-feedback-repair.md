---
kind: workflow
slug: pr-feedback-repair
description: Verify and repair current pull-request feedback in the exact writable PR-head lane.
dependencies:
  - implementation-delivery
rules:
  - slug: coding-agent-context-usage
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
2. Record an Agent Team Plan before spawning; use at most three independent low-overlap lanes, never more than four, and serialize shared files.
3. Verify every finding against current HEAD and fix only still-valid issues.
4. Run complete validation and a read-only verification lane after nontrivial repair.
5. Review the full integrated diff, push one coherent batch, verify the exact remote head, reflect, then explicitly resolve only addressed threads.

## Completion Gates

- The writable lane, expected head, push target, and dirty-change ownership are explicit.
- Stale, false-positive, out-of-scope, and human-needed findings are reported rather than silently changed.
- Kit itself did not edit source, stage, commit, push, comment, resolve, or merge from the prompt-producing path.
