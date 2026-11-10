# Standard Operating Procedures

Project-owned SOPs describe repeatable tasks for coding agents. Store each at
`docs/sops/<slug>/SOP.md`. Start with the template below and replace every
placeholder before activation. Use `draft`, `active`, or `retired` for `status`.

Read [the SOP rules](../references/rules/standard-operating-procedures.md) for
creation, discovery, execution, authorization, and maintenance requirements.
Kit creates this guide once during init or reconcile and preserves existing
content, even with force. Agents create and maintain SOP files directly;
Kit supplies no SOP creation command or execution engine.

## Reusable template

```markdown
---
kind: sop
slug: workbook-sync
description: Synchronize an approved workbook scope with its target system.
status: draft
owner: <responsible project role or team>
---

# Workbook synchronization

## Purpose

<Expected outcome and authoritative source references.>

## Applicability

<Trigger, supported scope and environment, exclusions, and when not to use.>

## Prerequisites

<Required access, approved inputs, current implementation references,
identity checks, validation baseline, and applicable authorization boundaries.
Use placeholders and audience-appropriate references, not credentials or
copied sensitive data. Keep public examples synthetic.>

## Procedure

1. Resolve <input reference>, <target reference>, and <scope> against current evidence.
2. Inspect the supported preview or comparison mechanism and its side effects.
3. Review the exact changes and satisfy existing policy before any mutation.
4. Run the verified project operation within the authorized scope.
5. Retain a safe operation reference; if asynchronous, read back its terminal state.

Replace these example steps with verified project instructions, expected
observations, and stop conditions. No workbook command is assumed to exist.

## Verification

<Read-back checks, expected counts or invariants, and evidence that proves
completion. An accepted request alone is not proof of success. Record date,
SOP revision, sanitized operation reference, checks, and unresolved results
in the appropriate task record; keep live data out of this document.>

## Recovery

<Stop conditions, partial-write checks, safe retry conditions, supported
recovery, and escalation owner. Resolve uncertain outcomes before retrying.
Interruption does not prove cancellation. Do not invent a rollback.>
```
