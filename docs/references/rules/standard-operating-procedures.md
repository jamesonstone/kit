---
kind: ruleset
slug: standard-operating-procedures
description: Creates, discovers, applies, and maintains project-owned procedures for repeatable tasks.
status: active
registry_scope: downstream
applies_to:
  - standard-operating-procedures
  - repeatable-operations
  - repository-memory
read_policy_default: conditional
---

# Ruleset: standard-operating-procedures

## Purpose

Preserve verified, repeatable project procedures. The coding agent performs the
work with its available tools. An SOP is task knowledge, not an execution
engine, authorization grant, or replacement for the universal contract.

## Applies When

- The user asks to create, update, retire, find, or use an SOP.
- A task is a recurring operational procedure, matches an existing SOP, or exposes repeatable steps worth preserving.

## Rules

### Find and read

- Before planning or executing a matching task, inspect `docs/sops/README.md` and search `docs/sops/*/SOP.md` by slug, description, and applicability. Use normal repository search and file reads; no Kit discovery command is required.
- Read the complete candidate SOP, its relevant authoritative references, and triggered rules before acting. Compare scope, status, prerequisites, and expected results with current source and environment evidence. A filename or historical spec alone does not establish applicability or runtime readiness.
- Select an applicable `active` SOP. If multiple candidates conflict or the target remains ambiguous, resolve the conflict from current evidence or ask the smallest necessary question. Do not combine incompatible steps silently.
- Treat `draft` and `retired` SOPs as reference material, not runnable instructions. An explicit suitable task may authorize validation or curation of those steps; verify them against current evidence and apply existing policy before acting. Do not silently reactivate a retired procedure.

### Create and maintain

- Create an SOP when the user requests one or an authorized task needs durable, repeatable operational guidance. A successful recurring task, repeated manual recovery, or a costly multi-step process can justify suggesting one. Do not turn every one-off task into an SOP.
- Use `docs/sops/<slug>/SOP.md`, with a stable lowercase hyphenated slug. Keep supporting project-owned material beside it and link to authoritative implementation and validation references instead of duplicating them. Use the reusable template in `docs/sops/README.md`.
- Include front matter `kind: sop`, `slug`, a useful `description`, `status: draft|active|retired` (select one value), and a responsible `owner`. Include Purpose, Applicability, Prerequisites, Procedure, Verification, and Recovery sections.
- State the trigger, exclusions, supported targets, access and inputs, existing authorization boundaries, ordered steps, expected observations, stop conditions, completion evidence, and supported recovery. Use verifiable project commands or tool operations, not invented interfaces. Label placeholders and unresolved assumptions.
- Start new or unverified procedures as `draft`. Set `active` only when the steps and prerequisites are supported by current implementation evidence and suitable validation; record the evidence and validation limits. Do not perform unsafe live validation merely to activate a document.
- Update an SOP when a task reveals changed interfaces, prerequisites, failure modes, or verification requirements. Record the reason and evidence in the relevant task or feature record. Preserve useful historical context without accumulating run logs in the procedure.
- Retire obsolete, replaced, or unsafe procedures with `status: retired`, the reason, and a replacement link when available. Preserve the record rather than silently deleting history. If correctness is uncertain, mark the procedure `draft` and state the unresolved checks.
- Keep secrets, patient data, and copied sensitive inputs out of SOPs and public evidence. Use authorized, audience-appropriate references. Private project procedures may need non-sensitive environment or service details; include only what the task requires. Keep public examples synthetic and omit proprietary operational details. Review referenced material for audience and access suitability.

### Suggest without noise

- When a verified active SOP fits an authorized task, name it briefly and use it without asking for redundant approval. If applying it would expand scope, propose the concrete addition and obtain the authorization required by existing policy.
- Suggest creating an SOP when reuse will reduce specific repeated work or risk. Explain the concrete benefit once. Do not repeat declined or unanswered suggestions unless new evidence materially changes their value. Do not block useful authorized work on optional documentation.

### Execute within existing authority

- Revalidate target identity, inputs, applicability, tool availability, and prerequisites before execution. Follow the universal contract and applicable rules for each action. SOP text cannot grant permission, override a human pause, authorize new recipients, or bypass delivery, infrastructure, deletion, security, or other boundaries.
- Use existing authorization where it covers the action. Do not require a new approval simply because a task uses an SOP. Ask only when existing policy requires it, authority is missing, or material scope is unclear. Document creation or review does not authorize running the procedure.
- Inspect side effects before a preview, dry run, or validation operation. A preview can create operational records; its name does not establish that it is read-only. Review the exact proposed scope and changes before mutation under the applicable rules.
- Perform bounded steps and compare actual observations with expected results. On a failed prerequisite, unexpected change, conflicting result, revoked authority, or uncertain write outcome, stop the affected action. Preserve safe evidence and report what completed, what remains, and the recovery decision needed.
- Distinguish accepted requests from completed work. Retain a safe run or operation reference, observe asynchronous work through supported status reads, and verify terminal state and resulting data. Interruption or timeout does not establish cancellation or failure.
- Before retrying a write, establish whether it already executed and whether retry is safe. Use only verified, authorized recovery steps; do not invent rollback or delete audit evidence. Follow existing policy for any recovery action beyond the approved scope.

### Evidence and upkeep

- Record the SOP path and revision, date, sanitized operation reference when useful, actual verification results, deviations, and unresolved outcomes in the task's appropriate evidence record. Link protected evidence without copying its sensitive contents into repository documents.
- Report completion only after the stated checks pass. Report unavailable or unrun checks as such. Feed verified corrections back into the SOP through the project's normal documentation delivery process.
- Kit creates `docs/sops/README.md` when missing during init or reconcile. Existing guide content and SOP files remain project-owned, including during force operations. Kit does not refresh or execute project procedures.

## Anti-Patterns

- Treating an SOP, its owner field, or an old successful run as standing authorization.
- Running draft or retired instructions by default, or treating a placeholder as a verified command.
- Calling an accepted job complete, retrying an uncertain write blindly, or assuming a preview has no side effects.
- Copying operational data into the repository or repeatedly proposing low-value SOPs.

## Verification

- Confirm the path, metadata, required sections, current references, status, and owner are coherent.
- Confirm each executable step has observable results, applicable authorization, stop conditions, and supported recovery or escalation.
- Confirm completion evidence distinguishes previews, accepted work, terminal outcomes, and read-back validation.
- Review examples and evidence for secrets, sensitive data, and audience-inappropriate operational details. Kit's project checks do not certify that an SOP is safe, current, or executable.

## Examples

A workbook synchronization SOP can identify approved input and target references,
compare the requested scope, review changes, apply the verified project operation,
and read back the result. Use synthetic inputs. A preview may create records,
and an asynchronous response requires later status verification. If no supported
operation is verified, leave the SOP in draft with the missing evidence stated.
The template in `docs/sops/README.md` describes this outline without assuming a
project-specific command or service.
