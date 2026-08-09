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

1. Inspect the request, repository state, relevant code, and repository-local
   contract before mutation.
2. Resolve material ambiguity through native agent planning and persist
   consequential rationale when the repository requires it.
3. Establish the repository-owned issue, branch, worktree, identity, and
   mutation boundaries before implementation.
4. Implement the smallest complete production-ready change while keeping
   affected tests and canonical documentation current.
5. Run focused and complete validation appropriate to the changed boundaries;
   report exact results and separate baseline debt.
6. Review the complete diff, curate durable repository memory, explicitly
   stage intended files, commit with the human identity, and deliver through
   the repository pull-request contract.

## Gates

- Before mutation: the resolved contract is ready, relevant artifacts have
  been read, repository state is known, and the authorized work lane is clear.
- Before delivery: applicable validation passes, the full diff is reviewed,
  and material repository rationale is current.

## Completion

- The requested behavior is implemented end-to-end or an exact external
  blocker is recorded.
- Validation, repository-memory disposition, pushed head, and hosted-check
  state are reported literally.

## Verification

- The resolved contract contains this workflow and every required dependency.
- No source mutation precedes repository grounding and lane authorization.
- Validation, diff review, repository memory, and delivery evidence are
  complete or an exact blocker is reported.
