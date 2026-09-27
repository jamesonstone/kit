---
kind: ruleset
slug: deadline-mode
description: Narrows validation and scope under an explicit user deadline without weakening any safety boundary.
status: active
registry_scope: downstream
applies_to:
  - prioritization
  - testing
  - implementation
read_policy_default: conditional
---

# Ruleset: deadline-mode

## Purpose

- Let a user trade breadth of validation for speed under a real deadline, without trading away safety.

## Applies When

- Only when the user explicitly declares a real deadline or time constraint in the conversation. Never infer or suggest it. It covers the declared work and ends when that work is delivered or the user ends it.

## Rules

- Skip work off the critical path: unrelated refactors and cleanup, broad re-runs of evidence that is already current, and speculative exploration.
- Per pull request, run the narrowest tests that prove the change plus the project's required pull-request checks. After a batch of merges or deployments, run one integration pass, and defer UI or browser walkthroughs until the whole authorized batch is delivered, then run them once.
- Never weaken merge authority and readiness, infrastructure and deployment approval, deletion safety, security, privacy, tenant isolation, or migration safety. Required post-deployment checks still run.
- State that deadline mode is active and report exactly which validation was deferred or narrowed as `SKIPPED` or `PARTIAL`, never as passing.
