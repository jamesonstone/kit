---
kind: ruleset
slug: cross-repository-program-coordination
description: Coordinates multi-repository programs through one ledger and live-evidence reconciliation.
status: active
registry_scope: downstream
applies_to:
  - cross-repository
  - program
  - deployment
  - handoff
read_policy_default: conditional
---

# Ruleset: cross-repository-program-coordination

## Purpose

- Keep long-running work that spans repositories resumable and truthful across sessions and agents.

## Applies When

- An accepted plan spans multiple repositories and has dependent deliverables, staged deployment or activation, or an expected agent or session handoff. A few directly coupled repositories with one simple dependency do not need a program.

## Rules

- Designate one coordinator repository and keep one ledger at `docs/programs/<program>/PROGRAM.md` there. Participant repositories keep their own specs, issues, pull requests, and evidence; the ledger stores identities, dependencies, current state, and pointers, not copies.
- Record each workstream's repository, issue, branch, pull request, dependencies, and state; the current ready frontier (only unblocked work); gates for integration, deployment, and activation; blockers with owners; and any standing merge authority with its scope and pause state.
- Record evidence as repository and commit, pull request and check identity, environment and artifact, observed result, and UTC time. Keep merged, deployed, and validated as separate states.
- Dispatch only the ready frontier. Checkpoint the ledger after each material transition (merge, deployment, blocker, decision, redirect, or handoff).
- Before resuming or declaring completion, reconcile the ledger against live repositories, GitHub, and runtime evidence; live evidence wins.
- Never store secrets, customer data, or transcripts in the ledger. Participant repository rules, `github-pr-merge`, and `infrastructure-change-approval` still apply.
