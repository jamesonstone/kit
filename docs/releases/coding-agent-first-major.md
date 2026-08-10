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
- The protected `kit pr fix` prompt fallback with active human/CodeRabbit
  collection, exact writable-head lane pinning, bounded wait, and explicit
  verified thread resolution.
- A mandatory registry-backed `feature-specification` contract, living V3 spec
  phase state in local resolution, and hard pre-source/pre-delivery gates for
  every feature.
- Explicit `feature|maintenance` work classification for implementation
  delivery; missing or contradictory hints now block instead of bypassing the
  feature-spec gate.

## Protected command paths

- `kit init`
- `kit reconcile`
- `kit rules add`
- `kit rules list`
- `kit rules view`
- `kit registry status`
- `kit pr fix`

Protection covers names and core outcomes only, not legacy flags or output.
`kit init` is the deliberate exception where its complete historical bootstrap
duty is part of the protected core outcome.

## Removed

All other Kit 1.x commands, runtime packages, general prompt templates, improvement
evaluations, command metadata, and command-specific automation were removed.
Only the narrow `kit pr fix` adapter is restored; dispatch, loop, the broader
legacy PR family, and agent-launching behavior remain removed.
The former feature-notes convention is also removed from active runtime
contracts. Migration does not rewrite historical specs or delete
project-owned content.
The separate `git-wt` binary and implementation remain unchanged.

## Required migration

Existing projects must run `kit reconcile --json --diff`, review the complete
schema-v2 and artifact plan, then run `kit reconcile --apply`. Local
customization and routing text are preserved unless an exact artifact is
explicitly accepted from the registry. After migration, run `kit init` to
backfill missing bootstrap starters without reconciling registry drift.

See the [migration guide](../migrations/coding-agent-first-major.md) and
[command guide](../commands.md).
