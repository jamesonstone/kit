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

1. Resolve the local contract with `implementation-delivery`, explicit task
   hints, and `--feature <feature>` for every non-trivial feature.
2. Read the returned routing entrypoints, workflow, mandatory rules, and
   conditional rules before mutation.
3. Inspect repository state, relevant code, tests, and durable rationale.
4. Create or adopt the complete living V3 feature spec from the accepted native
   plan, then re-resolve and do not begin source edits until its permission is
   ready.
5. Implement the smallest complete change and validate the affected boundaries.
6. Keep decisions and discoveries live, then reconcile validation, actual
   outcome, delivery evidence, and repository-memory disposition against the
   complete integrated diff before delivery.

The canonical declarative workflow is
[`implementation-delivery.md`](../references/workflows/implementation-delivery.md).

## Pull-request feedback repair

1. Run `kit pr fix` or resolve `pr-feedback-repair` natively for one exact
   writable PR-head lane.
2. Use a host wakeup or the bounded token-free `gh` await mode; use one-shot
   collect for late or human feedback.
3. Verify every active finding against current `HEAD`, repair only valid
   findings, validate the integrated diff, and push one coherent batch.
4. Refresh the exact pushed head and resolve only verified addressed threads.

The canonical declarative workflow is
[`pr-feedback-repair.md`](../references/workflows/pr-feedback-repair.md). The
fallback adapter executes only bounded intake, lane preparation, prompt output,
and explicitly confirmed thread resolution; it never executes repair or
delivery.

## Contract maintenance

1. Run `kit registry status` or `kit reconcile --json --diff`.
2. Review catalog revision, artifact states, diagnostics, and all file changes.
3. Apply conflict-free changes only with `kit reconcile --apply`.
4. Resolve same-section conflicts manually or accept one exact artifact.
5. Run `kit contract resolve` again and review the repository diff.

The canonical declarative workflow is
[`registry-maintenance.md`](../references/workflows/registry-maintenance.md).

## Repository memory

Every non-trivial feature requires `docs/specs/<feature>/SPEC.md` before source
edits. The spec remains live through decisions, discoveries, validation,
outcome, delivery, and curation. Use the project progress summary, explicit
relationships, and repository evidence to load relevant historical specs
progressively; preserve them even when their former runtime surface is removed.
