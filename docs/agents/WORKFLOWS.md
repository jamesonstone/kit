# Agent Workflows

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
