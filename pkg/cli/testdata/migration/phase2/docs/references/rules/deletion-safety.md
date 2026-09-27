---
kind: ruleset
slug: deletion-safety
description: Defaults persistent state deletion to a recoverable lifecycle and requires exact manual confirmation before any irreversible hard delete.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - implementation
  - data
  - persistence
  - filesystem
  - identity
  - api
  - ui
  - automation
  - cleanup
  - retention
  - migration
  - operations
  - cloud
  - infrastructure
read_policy_default: must
---

# Ruleset: Deletion Safety

## Purpose

- Make recoverability the default whenever a Kit-managed project deletes
  persistent project, user, business, or external-system state.
- Prevent ordinary delete language, automation, or prior task approval from
  becoming implicit authority for an irreversible purge.
- Require one specific, auditable human decision for the exact current targets
  and consequences of every hard-delete batch.

## Applies When

- Designing, implementing, changing, reviewing, testing, operating, or
  documenting deletion behavior in a Kit-managed project.
- Deleting or removing covered records, files, objects, identities, artifacts,
  messages, resources, environments, backups, snapshots, or history.
- Configuring cleanup, retention, lifecycle, migration, reset, replacement,
  decommissioning, or garbage-collection behavior that can remove covered state.

This rule covers state that is persistent, authoritative, user-visible,
business-relevant, security-relevant, or managed in an external system.
Task-owned ephemeral scratch that never became authoritative state is not a
retained resource. If ownership or significance is uncertain, treat the target
as covered.

## Definitions

### Soft Delete

A soft delete is a reversible lifecycle transition with a supported and
authorized restore path. It may use a tombstone, archived or quarantined state,
disablement, retained version history, provider recovery control, or another
project-native mechanism.

Soft delete is not established merely because bytes might still exist in an
undocumented backup. The project must define how an authorized operator or user
can identify, restore, and verify the deleted target during its retention window.

### Hard Delete

A hard delete removes the supported restore path or makes recovery dependent on
unsupported forensics. It includes purge, destroy, force deletion, empty-trash
operations, destructive replacement, history rewrite, retention expiry,
backup or snapshot deletion, cryptographic erasure, and irreversible cascades.

## Rules

### Soft Delete Is The Default

- Interpret an unqualified request to delete or remove covered state as soft
  delete. Verbs such as clean up, decommission, reset, replace, recreate, or
  start over do not imply hard-delete authority.
- Prefer the project's existing recoverable lifecycle. When none exists and
  deletion behavior is being implemented or changed, add the smallest complete
  soft-delete and restore capability appropriate to that system.
- The normal API, UI, CLI, service, repository, and automation path must perform
  soft delete. Keep irreversible purge as a separate privileged action.
- A tracked-file deletion through retained Git history may be a recoverable
  equivalent. Rewriting history or removing the last supported recovery
  reference is a hard delete.
- A retention deadline makes covered state eligible for review or purge; it is
  not independent authority to hard-delete automatically.
- Soft-delete work is authorized by the accepted task when it remains in scope.
  Continue to honor stricter repository, identity, infrastructure, production,
  legal, privacy, and security gates.

### Restore And Lifecycle Contract

- Define lifecycle states, retention duration, visibility, actor permissions,
  and the restore entrypoint.
- Preserve tenant, site, owner, and authorization boundaries during delete,
  discovery, restore, and purge. A soft-deleted target must never become visible
  or restorable across an isolation boundary.
- Define uniqueness, foreign-key, cascade, search, event, audit, billing, and
  integration behavior while a target is deleted and after it is restored.
- Make delete and restore idempotent where retries are possible. Protect state
  transitions from stale confirmations and concurrent mutation.
- Preserve an audit record of actor, target, action, time, state transition, and
  retention deadline without retaining prohibited sensitive payloads.

### Separate Hard-Delete Surface

- Do not hide hard delete behind a normal-delete `force`, `hard`, or
  `permanent` flag that can bypass a confirmation boundary.
- Expose hard delete only through a separate privileged operation with
  server-side authorization and confirmation enforcement. A client-side modal
  or CLI prompt alone is insufficient when the server can still purge directly.
- Bind any confirmation token or approval record to the actor, exact targets,
  action, environment, material consequences, and a short expiry. For a
  bounded selector, bind the confirmation to materialized target IDs or an
  immutable snapshot or version token as well as the resolved count. Make the
  confirmation single-use.
- Reject absent, expired, reused, mismatched, or stale confirmation evidence.
- Immediately before execution, compare the current target set or immutable
  version with the confirmed snapshot. Abort on any difference, refresh the
  outline, and obtain new confirmation.

### Specific Manual Confirmation Before Hard Delete

Before every hard-delete batch, use read-only inspection to resolve and present:

1. exact target identities, or a bounded selector first resolved to the exact
   current target set with its current count and materialized target IDs or an
   immutable snapshot or version token;
2. repository, tenant, site, environment, account, project, or Region as
   applicable;
3. dependent cascades and externally affected resources;
4. why soft delete or continued quarantine is insufficient;
5. the fact that the supported restore path will be removed;
6. current backup, snapshot, retention, legal, privacy, and security effects;
7. actor, execution boundary, failure handling, and verification evidence.

After presenting that outline, ask the human to confirm irreversible deletion
of those exact targets. One response may authorize one fully enumerated batch.
Approval of a plan counts only when the human response follows the complete
outline and explicitly approves the exact current hard-delete targets.

The following never count as hard-delete confirmation:

- the initial request, even when it asked to delete or destroy;
- general task, plan, pull-request, deployment, or merge approval;
- an automated signal, scheduled job, retention deadline, or unattended policy;
- prior approval to soft-delete, quarantine, replace, or decommission;
- earlier confirmation for different targets, counts, cascades, or environments;
- a broad instruction such as delete everything, clean up, or start over.

Target-set or version drift, including changed identities with the same count,
expanded cascades, a different environment or actor, or changed recovery
consequences invalidates the confirmation. Stop, refresh the outline, and
obtain a new specific manual confirmation.

### Composition With Other Gates

- This rule never weakens stricter repository or provider controls. Prohibited
  actions remain prohibited even if a human asks to hard-delete them.
- When `infrastructure-change-approval` or another rule also requires a
  post-outline deletion confirmation, one manual response may satisfy both only
  when the combined outline contains every field required by each rule.
- Revoke or disable compromised credentials immediately through a reversible
  lifecycle action where supported. Irreversible erasure remains a hard delete
  unless a mandatory external safety control performs it outside project or
  coding-agent authority.
- A legal or privacy duty may require hard deletion, but it does not remove the
  exact-target, authorization, audit, and manual-confirmation requirements.

### Implementation And Test Expectations

- New or changed deletion paths include focused tests proving default soft
  deletion, authorized restore, isolation, idempotency, lifecycle visibility,
  and retention behavior.
- Hard-delete tests prove the separate privileged path rejects missing,
  mismatched, stale, expired, reused, and drifted confirmation evidence.
- Cover cascades, partial failure, retries, concurrency, audit evidence, and
  downstream event or integration behavior when those risks apply.
- Existing projects are not automatically migrated by Kit refresh. Apply this
  contract whenever a later task creates or changes deletion behavior and
  record genuine legacy gaps literally.

## Anti-Patterns

- Making HTTP `DELETE`, a trash icon, or a default CLI remove command purge
  immediately.
- Treating a database flag as soft delete without an authorized restore path.
- Automatically purging when retention expires without a specific manual
  confirmation for the resolved targets.
- Using a broad prefix, wildcard, account, directory, or tenant as deletion
  authority without materializing the exact bounded result or binding it to an
  immutable snapshot or version.
- Relying on a typed client phrase while leaving an unguarded server purge API.
- Calling a destructive replacement safe because a new resource will be
  created afterward.
- Retaining sensitive payloads indefinitely under the label of recoverability.
- Asking twice when one combined post-outline confirmation satisfies every
  applicable deletion gate.

## Verification

- Confirm the normal delete path is reversible and the restore path is usable,
  authorized, documented, and tested.
- Confirm retention expiry does not itself execute an irreversible purge.
- Confirm hard delete has a separate privileged server-enforced surface.
- Confirm the pre-delete outline resolves exact targets, count, cascades,
  environment, recovery loss, policy consequences, and materialized IDs or an
  immutable snapshot or version for every bounded selector.
- Confirm execution aborts when the current target set or immutable version
  differs from the confirmed snapshot, even if the target count is unchanged.
- Confirm a human manually approved those exact current targets after the
  outline and that drift invalidates the approval.
- Confirm audit evidence records actor, scope, confirmation, result, failures,
  and verification without exposing secrets or prohibited payloads.
- Confirm applicable focused, integration, migration, and end-to-end tests pass
  or report exact gaps literally.

## Examples

Recoverable application deletion:

```text
DELETE /orders/123 -> state=deleted, deleted_at=<time>, restore available
POST /orders/123/restore -> state=active after owner and conflict checks
```

Hard-delete confirmation outline:

```text
Targets: order 123 and 2 exact dependent artifacts
Environment: production / tenant acme
Recovery: restore path and final snapshot will be removed
Reason soft delete is insufficient: approved legal erasure request ER-42
Verification: read-after-delete plus audit event without payload retention

Proceed with irreversible hard deletion of these exact targets?
```
