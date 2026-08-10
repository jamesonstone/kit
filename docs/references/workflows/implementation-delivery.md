---
kind: workflow
slug: implementation-delivery
description: Drives coding-agent work from repository grounding through validated pull-request delivery.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - implementation
  - delivery
read_policy_default: must
dependencies:
  - ruleset/feature-specification
  - ruleset/safety-guardrails
  - ruleset/work-lane-gating
  - ruleset/testing-and-environment-validation
  - ruleset/source-file-size
  - ruleset/constitution-curation
---

# Workflow: Implementation Delivery

## Purpose

Give coding agents one provider-neutral execution contract for production-ready
repository changes without making Kit an agent runtime.

## Applies When

- A coding agent will change implementation, tests, build behavior, or other
  version-controlled product behavior.
- The requested outcome includes validated review or pull-request delivery.

## Workflow

1. Classify the work explicitly as `feature` or `maintenance`; do not infer a
   maintenance exemption from missing hints.
2. Inspect the request, repository state, relevant code, historical-spec
   indices, and repository-local contract before mutation.
3. For every feature, resolve with `--work-type feature --feature <feature>`, create or adopt the
   complete living V3 `docs/specs/<feature>/SPEC.md` from the accepted native
   plan, and re-resolve before any source or test edit.
4. For genuinely mechanical maintenance, resolve with
   `--work-type maintenance`; this explicit resolved hint is the sole
   feature-spec exemption.
5. Establish the repository-owned issue, branch, worktree, identity, and
   mutation boundaries before implementation.
6. Implement the smallest complete production-ready change while keeping the
   spec's decisions, discoveries, validation map, tests, and canonical
   documentation current.
7. Run focused and complete validation appropriate to the changed boundaries;
   report exact results and separate baseline debt.
8. Reconcile the living spec against the integrated diff, record literal
   validation and actual outcome, curate durable repository memory, explicitly
   stage intended files, commit with the human identity, and deliver through
   the repository pull-request contract.

## Gates

- Before source implementation: the resolution contains an explicit valid work
  type. Feature work reports a structurally complete living V3 spec and permits
  source implementation; spec authoring remains allowed when this gate is
  blocked. Maintenance proceeds only with its explicit exemption recorded.
- Before mutation: relevant artifacts have been read, repository state is
  known, and the authorized work lane is clear.
- Before delivery: the spec is in `deliver` or `complete` phase, applicable
  validation is recorded literally, and the full integrated diff and actual
  outcome have been reconciled.
- Before completion: delivery evidence and repository-memory disposition are
  current, and historical specifications remain discoverable and preserved.

## Completion

- The requested behavior is implemented end-to-end or an exact external
  blocker is recorded.
- The living spec contains the actual implemented outcome, validation,
  repository-memory disposition, pushed head, and hosted-check state literally.

## Verification

- The resolved contract contains this workflow and every required dependency.
- No source mutation precedes repository grounding, lane authorization, and a
  structurally complete feature spec whose resolved phase permits source work.
- Validation, diff review, repository memory, and delivery evidence are
  complete or an exact blocker is reported.
