---
kind: ruleset
slug: github-pr-merge
description: Merges only under explicit bounded human authority and only exact current MERGE_READY pull requests.
status: active
registry_scope: downstream
applies_to:
  - git
  - github
  - pull-request
  - merge
  - merge-queue
read_policy_default: conditional
---

# Ruleset: github-pr-merge

## Purpose

- Keep merge a distinct, human-authorized boundary without re-asking for every new commit.
- Merge only what current evidence proves ready, and report merge, deployment, and runtime as separate facts.

## Applies When

- Before any merge, merge-queue, or auto-merge action, including multi-PR and cross-repository waves.

## Rules

### Authority

- Merge only under explicit human authorization naming the scope: a pull request, a task, or a bounded program with its repositories, bases, environments, and permitted actions. Delivery consent, approvals, passing checks, subagent assignment, or a ledger never create merge authority.
- A bounded grant covers later in-scope pull requests and new heads. A commit SHA or head OID identifies readiness evidence only, never authorization; never request exact-head reauthorization. A changed in-scope head invalidates readiness, not standing authority.
- Adding a repository, base, environment, actor, merge method, deployment workflow, or material effect is scope expansion and needs updated authority. A human pause, hold, or revocation stops the affected merges and their dependents until the human resumes them.
- Never bypass branch protection, reviews, required checks, merge queues, or identity safeguards, and never use admin overrides.

### Readiness

- Classify each pull request from current evidence as `MERGE_READY` (required checks and reviews passed on the current head, dependencies satisfied, repository policy satisfied, no excluded risk), `BLOCKED` (a gate failed or authority is missing, paused, or exceeded), or `UNKNOWN` (evidence missing, stale, or unattributable). Keep these words literal.
- Only current `MERGE_READY` pull requests merge. Pending, missing, stale, or policy-ineligible skipped checks are not passing, and local tests do not replace required hosted checks.
- Use one complete current-state snapshot per merge wave. Wait with bounded backoff instead of repeated polling.

### Waves And Repair

- Merge dependency chains in order; independent ready pull requests may merge together. Use the repository-permitted merge method or required merge queue.
- Repair in-scope blockers on the existing pull request (merge the base in, commit, push without force), then revalidate the new head; it re-enters the wave without a new prompt. Use a replacement pull request only when the scope or architecture materially changes.
- A failure stops that pull request and its dependents only; independent ready work may continue.

### Effects

- Standing authority may continue a named existing standard deployment of the merged artifact to already-provisioned resources. IAM, network, KMS, secrets, schema or data-loss changes, infrastructure creation, replacement, or deletion, and nonstandard deployments need their own approval (see `infrastructure-change-approval`). Unclassified effects make a pull request `UNKNOWN`.
- A merge is not proof of deployment, runtime health, or production correctness. Report merge result, hosted workflows, deployment, and runtime state separately, with the merge commit and observed state.
