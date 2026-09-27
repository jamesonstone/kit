---
kind: workflow
slug: pr-feedback-repair
description: Verify and repair current pull-request feedback in the exact writable PR-head lane.
dependencies:
  - implementation-delivery
rules:
  - slug: agent-completion-output
    required: false
  - slug: deletion-safety
    required: false
  - slug: agent-team-orchestration
    required: false
  - slug: github-pr-delivery
    required: true
  - slug: work-lane-gating
    required: true
  - slug: testing-and-environment-validation
    required: false
evidence:
---
# Workflow: PR Feedback Repair

## Purpose

- Preserve the supervisor contract produced by `kit pr fix` without Kit launching an agent.
- Repair only current, verified findings on the exact same-repository PR-head branch.

## Phases

1. Use `kit pr fix` for bounded current feedback intake and lane evidence.
2. Parallelize independent investigation when it helps; serialize writes to shared files and keep Git and GitHub mutations in the primary agent.
3. Verify every finding against current HEAD and fix only still-valid issues.
4. Run complete validation and use a fresh independent read-only verifier when supported; otherwise perform and disclose a distinct supervisor self-review.
5. Review the full integrated diff, push one coherent batch, verify the exact remote head, reflect, then explicitly resolve only addressed threads.

## Completion Gates

- The writable lane, expected head, push target, and dirty-change ownership are explicit.
- Stale, false-positive, out-of-scope, and human-needed findings are reported rather than silently changed.
- Kit itself did not edit source, stage, commit, push, comment, resolve, or merge from the prompt-producing path.
