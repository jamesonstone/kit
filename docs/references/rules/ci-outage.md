---
kind: ruleset
slug: ci-outage
description: Keeps a hosted CI outage from blocking merges by allowing a human-approved manual bypass backed by full local validation on the exact head.
status: active
registry_scope: downstream
applies_to:
  - ci
  - github
  - pull-request
  - merge
read_policy_default: conditional
---

# Ruleset: ci-outage

## Purpose

- A hosted CI outage is not a blocker. Keep work moving when hosted CI cannot run, without hiding test failures or untruthful reports.

## Applies When

- Hosted CI cannot start or run because of a provider outage or exhausted minutes or billing, for example "The job was not started because recent account payments have failed or your spending limit needs to be increased".
- It never applies to a check that ran and failed, a check that is only slow or queued, or a check that the repository skips by policy. Those stay `BLOCKED` or `UNKNOWN` under `github-pr-merge`.

## Precedence

- This rule is an explicit exception to two merge gates in the always-loaded contract and in `github-pr-merge`: "never bypass required checks" and "merge only a current `MERGE_READY` head" with required checks passed.
- It takes precedence over those gates only for hosted checks that could not run. Every other gate, safeguard, and approval boundary stays in force: branch protection, reviews, merge queues, identity safeguards, admin-override prohibition, merge authority, and checks that ran.

## Rules

### Tell The Human

- State the outage plainly: the affected repositories and pull requests, the observed error, and that the hosted checks did not run. Offer a manual bypass.
- Never report a not-run hosted check as passed. Keep the literal state `NOT_RUN`.

### Bypass Approval

- The human's first bypass approval is a standing grant. It covers later in-scope merges while the outage lasts, with no new approval per pull request.
- It does not create merge authority. Each merge still needs explicit human authorization for the scope under `github-pr-merge`.

### Local Validation Replaces Hosted Checks

- Under the bypass, a pull request is `MERGE_READY` only when every other gate passes on the current head (base merged in, review comments addressed, required reviews, no conflicts, dependencies and repository policy satisfied).
- The repository's full local validation, as documented in `docs/references/testing.md` (for example `make check` plus any integration and end-to-end targets), must pass on the exact head being merged. A partial run, a run on another head, or a missing documented command is not passing.
- Never use admin overrides. If branch protection or a ruleset still blocks the merge, report the pull request as `BLOCKED` with the passing local evidence, and the human merges it manually.

### Reporting

- Before each bypass merge, post a pull-request comment that names the outage, the local validation commands, their results, and the head SHA.
- The final report lists each pull request merged under the bypass and its comment.

### When The Outage Ends

- When hosted CI runs again, stop using the bypass.
- Re-run CI on every pull request merged under the bypass (for example `gh run rerun` or a workflow dispatch on the merge commit), report each result, and fix failures in follow-up pull requests.
