---
kind: ruleset
slug: deletion-safety
description: Makes deletion recoverable by default and requires exact-target human confirmation before any hard delete.
status: active
registry_scope: downstream
applies_to:
  - data
  - persistence
  - filesystem
  - api
  - cleanup
  - retention
  - infrastructure
read_policy_default: conditional
---

# Ruleset: deletion-safety

## Purpose

- Prevent unrecoverable loss of project, user, business, or external-system state.

## Applies When

- Designing deletion behavior, or deleting persistent state in files, databases, services, accounts, or infrastructure. Task-owned scratch that never became authoritative state is exempt.

## Rules

### Soft Delete By Default

- An unqualified "delete" or "remove" means a recoverable soft delete: a reversible lifecycle state with a supported, authorized, tested restore path. Deleting a tracked file whose history Git retains is recoverable.
- Product and operational paths delete softly by default. Define retention, visibility, permissions, cascades, uniqueness, and audit records (actor, target, time, transition); make delete and restore idempotent.
- Hard delete (purge, force delete, empty trash, destructive replacement, history rewrite, backup or snapshot deletion, irreversible cascade, retention expiry) is a separate privileged, audited, server-enforced operation, never a `force` flag on the normal path.

### Confirmation Before Hard Delete

- Before any hard delete, present the exact targets (or a selector resolved to the current targets with count and IDs), environment, cascades, why soft delete is insufficient, what cannot be restored, backup state, and how you will verify the result.
- Then get a specific human confirmation for those targets. The original request, plan or merge approval, automation, retention schedules, prior soft-delete approval, and broad cleanup language do not count.
- Immediately before executing, re-resolve the targets; if they changed, outline and confirm again.

### Tests

- New or changed deletion paths include tests proving default soft delete, restore, and that the hard-delete path rejects missing or stale confirmation.
