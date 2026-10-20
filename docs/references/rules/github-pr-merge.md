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
- Never bypass branch protection, reviews, required checks, merge queues, or identity safeguards, and never use admin overrides. The only exception is the human-approved CI-outage bypass below, which still never uses admin overrides.

### Readiness

- Classify each pull request from current evidence as `MERGE_READY` (required checks and reviews passed on the current head, dependencies satisfied, repository policy satisfied, no excluded risk), `BLOCKED` (a gate failed or authority is missing, paused, or exceeded), or `UNKNOWN` (evidence missing, stale, or unattributable). Keep these words literal.
- Only current `MERGE_READY` pull requests merge. Pending, missing, stale, or policy-ineligible skipped checks are not passing, and local tests do not replace required hosted checks, except under the CI-outage bypass below.
- Use one complete current-state snapshot per merge wave. Wait with bounded backoff instead of repeated polling.

### CI Outage Bypass

- Applies only when hosted CI cannot start or run because of a provider outage or exhausted minutes or billing, for example "The job was not started because recent account payments have failed or your spending limit needs to be increased". It never applies to a check that ran and failed, a check that is only slow or queued, or a check that the repository skips by policy; those stay `BLOCKED` or `UNKNOWN`.
- A hosted CI outage is not a blocker. Tell the human plainly: the affected repositories and pull requests, the observed error, and that the hosted checks are not run. Offer a manual bypass. Never report a not-run hosted check as passed; keep the literal state `NOT_RUN`.
- The human's first bypass approval is a standing grant. It covers later in-scope merges while the outage lasts, with no new approval per pull request. It does not create merge authority: each merge still needs explicit human authorization for the scope (see Authority).
- Under the bypass, a pull request is `MERGE_READY` only when every other gate passes on the current head (base merged in, review comments addressed, required reviews, no conflicts, dependencies and repository policy satisfied) and the repository's full local validation, as documented in `docs/references/testing.md` (for example `make check` plus any integration and end-to-end targets), passes on the exact head being merged. A partial run, a run on another head, or a missing documented command is not passing.
- Never use admin overrides. If branch protection or a ruleset still blocks the merge, report the pull request as `BLOCKED` with the passing local evidence, and the human merges it manually.
- Before each bypass merge, post a pull-request comment that names the outage, the local validation commands, their results, and the head SHA. The final report lists each pull request merged under the bypass and its comment.
- When hosted CI runs again, stop using the bypass. Re-run CI on every pull request merged under the bypass (for example `gh run rerun` or a workflow dispatch on the merge commit), report each result, and fix failures in follow-up pull requests.

### Waves And Repair

- Merge dependency chains in order; independent ready pull requests may merge together. Use the repository-permitted merge method or required merge queue.
- Repair in-scope blockers on the existing pull request (merge the base in, commit, push without force), then revalidate the new head; it re-enters the wave without a new prompt. Use a replacement pull request only when the scope or architecture materially changes.
- A failure stops that pull request and its dependents only; independent ready work may continue.

### Effects

- Standing authority may continue a named existing standard deployment of the merged artifact to already-provisioned resources. IAM, network, KMS, secrets, schema or data-loss changes, infrastructure creation, replacement, or deletion, and nonstandard deployments need their own approval (see `infrastructure-change-approval`). Unclassified effects make a pull request `UNKNOWN`.
- A merge is not proof of deployment, runtime health, or production correctness. Report merge result, hosted workflows, deployment, and runtime state separately, with the merge commit and observed state.
- Follow `deployment-recovery` for post-merge publication and deployment recovery.
- After each confirmed merge, run `work-cleanup` for the merged lane without asking.
