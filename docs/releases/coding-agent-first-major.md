# Kit 2.0: Coding-Agent-First

Kit 2.0 is a breaking product reset. Kit is now a provider-neutral,
repository-native contract layer for coding agents rather than a repository
memory, prompt, or agent-execution harness.

## Added

- Typed, versioned registry catalogs for `ruleset` and declarative `workflow`
  artifacts.
- Schema-v2 project provenance with versions, digests, section hashes, paths,
  source revision, and managed-state classification.
- `kit contract resolve` with stable local-only resolved-contract JSON.
- `kit registry list|view|add` as generic typed-artifact administration.
- Transactional initialization and section-aware reconciliation over one
  reusable registry core.
- Bounded provider-routing sections that preserve project-owned content.
- Canonical `kit init` bootstrap compatibility: deterministic environment,
  GitHub, repository-memory, RLM, reference, registry, and provenance starters;
  an evidence-based `repository-bootstrap` prompt; and safe user defaults at
  `~/.config/kit/.kit.yaml`.
- A `pr-feedback-repair` workflow with bounded asynchronous CodeRabbit and
  human-review intake, exact terminal states, rate-aware observation, and an
  explicit agent-team dependency.

## Protected command paths

- `kit init`
- `kit reconcile`
- `kit rules add`
- `kit rules list`
- `kit rules view`
- `kit registry status`

Protection covers names and core outcomes only, not legacy flags or output.
`kit init` is the deliberate exception where its complete historical bootstrap
duty is part of the protected core outcome.

## Removed

All other Kit 1.x commands, runtime packages, prompt templates, improvement
evaluations, command metadata, and command-specific automation were removed.
The former `kit pr fix` prompt semantics live in the declarative workflow; its
command and any agent-launching behavior are not restored.
The separate `git-wt` binary and implementation remain unchanged.

## Required migration

Existing projects must run `kit reconcile --json --diff`, review the complete
schema-v2 and artifact plan, then run `kit reconcile --apply`. Local
customization and routing text are preserved unless an exact artifact is
explicitly accepted from the registry. After migration, run `kit init` to
backfill missing bootstrap starters without reconciling registry drift.

See the [migration guide](../migrations/coding-agent-first-major.md) and
[command guide](../commands.md).
