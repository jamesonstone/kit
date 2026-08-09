---
kind: workflow
slug: registry-maintenance
description: Safely previews and reconciles repository-local contract artifacts against their configured registry.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - registry
  - reconcile
  - maintenance
read_policy_default: conditional
dependencies:
  - ruleset/safety-guardrails
---

# Workflow: Registry Maintenance

## Purpose

Keep materialized rules, workflows, routing, and provenance current without
discarding intentional repository customization.

## Applies When

- Registry freshness or project schema requires inspection.
- A user or coding agent is preparing to apply contract updates.
- `kit registry status` reports changes or attention.

## Workflow

1. Resolve the configured registry source and validate its catalog and content
   digests before planning any write.
2. Compare installed provenance, current local content, and the new registry
   revision without mutating the repository.
3. Preview the complete change and classify managed, local-custom, missing,
   and conflicting artifacts.
4. Apply only conflict-free changes explicitly requested by the caller.
5. Preserve local-only sections, stop on same-section divergence, and require
   exact artifact acceptance before replacing customized content.
6. Resolve the local contract again and validate the repository-owned diff.

## Gates

- No apply occurs before the catalog, artifact digests, target paths, and full
  write plan validate successfully.
- Same-section conflicts require manual resolution or exact targeted registry
  acceptance; one acceptance never authorizes another artifact.

## Completion

- Reconciliation reports `current`, or every remaining conflict and preserved
  customization is explicit.
- Contract resolution is ready and provenance matches the applied registry
  revision.

## Verification

- Preview mode performs no writes.
- Applied writes match the reviewed plan and preserve unrelated content.
- Missing, invalid, and conflicting required artifacts keep the contract
  blocked until resolved.
