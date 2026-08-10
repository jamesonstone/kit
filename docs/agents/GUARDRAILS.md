# Agent Guardrails

## Hard rules

- Preserve user-owned work and inspect branch, worktree, identity, and diff
  state before mutation.
- Do not stash, reset, clean, force-push, merge, or alter repository settings
  unless explicitly authorized by the applicable repository rule.
- Keep handwritten source and test files at 300 physical lines or less.
- Run the applicable code-level validation and report exact evidence.
- Do not treat a blocked contract as implementation-ready.
- Do not start implementation delivery without an explicit `feature` or
  `maintenance` work type; omitted classification is not an exemption.
- A blocked feature spec permits only spec authoring; source edits wait for a
  structurally complete V3 spec and a successful re-resolution.

## GitHub delivery gate

Before issue, branch, staging, commit, push, or pull-request mutations, load
`docs/references/rules/github-pr-delivery.md`,
`docs/references/rules/work-lane-gating.md`, and
`docs/references/rules/safety-guardrails.md`. Use the human's configured Git
identity, explicitly stage intended paths, and verify the pushed head and
hosted checks.

## Infrastructure approval gate

Before changing cloud, Kubernetes, infrastructure-as-code source, config, or
state, load `docs/references/rules/infrastructure-change-approval.md`. Present
one bounded target/action/impact/recovery/validation outline and obtain the
required confirmation. Cloud identity verification belongs to the native cloud
tooling and repository configuration; Kit provides no cloud command surface.

## Validation gate

Before implementation or validation, load
`docs/references/rules/testing-and-environment-validation.md` and
`docs/references/testing.md`. Code-level tests and pull-request checks remain
required; live or end-to-end evidence supplements them.

## Completion bar

- Requested behavior is implemented end-to-end or an exact external blocker is
  evidenced.
- Focused and complete applicable validation passes, with baseline debt clearly
  separated.
- The full diff has been reviewed and no obsolete runtime or documentation
  remains in the changed boundary.
- Material rationale, decisions, discoveries, validation, and outcome are
  current in canonical repository memory.
- Contract resolution is ready and reconciliation is current or its remaining
  diagnostics are explicitly explained.
