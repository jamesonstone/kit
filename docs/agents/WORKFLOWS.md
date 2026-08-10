# Agent Workflows

## Repository bootstrap

1. Run `kit init` and review every created, merged, or preserved disposition.
2. Resolve `repository-bootstrap` and follow `RLM.md` from routing and indices
   into only relevant repository evidence.
3. Populate Constitution, progress summary, testing, tooling, integrations,
   Makefile, and bounded README sections only from verified evidence.
4. Validate safe populated commands and report unresolved evidence gaps.

The canonical declarative workflow is
[`repository-bootstrap.md`](../references/workflows/repository-bootstrap.md).

## Implementation delivery

1. Resolve the local contract with explicit workflow and task hints.
2. Read the returned routing entrypoints, workflow, mandatory rules, and
   conditional rules before mutation.
3. Inspect repository state, relevant code, tests, and durable rationale.
4. Use native agent planning when material ambiguity or architecture requires
   it; keep accepted durable rationale in the current feature spec.
5. Implement the smallest complete change and validate the affected boundaries.
6. Curate repository memory, review the complete diff, and deliver through the
   repository's GitHub rules.

The canonical declarative workflow is
[`implementation-delivery.md`](../references/workflows/implementation-delivery.md).

## Pull-request feedback repair

1. Resolve `pr-feedback-repair` for one exact writable PR-head lane.
2. Use a host wakeup or the bounded token-free `gh` await mode; use one-shot
   collect for late or human feedback.
3. Verify every active finding against current `HEAD`, repair only valid
   findings, validate the integrated diff, and push one coherent batch.
4. Refresh the exact pushed head and resolve only verified addressed threads.

The canonical declarative workflow is
[`pr-feedback-repair.md`](../references/workflows/pr-feedback-repair.md). Kit
does not execute its GitHub, waiting, agent, or delivery steps.

## Contract maintenance

1. Run `kit registry status` or `kit reconcile --json --diff`.
2. Review catalog revision, artifact states, diagnostics, and all file changes.
3. Apply conflict-free changes only with `kit reconcile --apply`.
4. Resolve same-section conflicts manually or accept one exact artifact.
5. Run `kit contract resolve` again and review the repository diff.

The canonical declarative workflow is
[`registry-maintenance.md`](../references/workflows/registry-maintenance.md).

## Repository memory

When code and tests cannot preserve material rationale, create or adopt
`docs/specs/<feature>/SPEC.md` before source edits. Keep decisions and
discoveries current, then record literal validation and outcome. Preserve
historical specifications even when their former runtime surface is removed.
