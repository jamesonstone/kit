# Kit Overview

Kit is a provider-neutral contract layer between a repository and coding
agents. It distributes typed Markdown artifacts, preserves their provenance,
and resolves the deterministic subset applicable to a task.

## Product model

1. A versioned registry catalog indexes `ruleset` and `workflow` artifacts.
2. `kit init` materializes every downstream-visible artifact and bounded
   routing block into a project.
3. `.kit.yaml` records source identity, revision, digests, section hashes,
   paths, and artifact state.
4. `kit contract resolve` selects ordered local artifacts from explicit hints.
5. `kit reconcile` compares local, installed, and current registry state and
   previews every write before an explicit apply.

Repository-local Markdown is authoritative. Resolved-contract JSON is a stable
machine interface, not a second source of truth.

Published schema files are
[`registry-catalog-v1`](../schemas/registry-catalog-v1.schema.json),
[`project-config-v2`](../schemas/project-config-v2.schema.json), and
[`resolved-contract-v1`](../schemas/resolved-contract-v1.schema.json).

## Artifact types

- A `ruleset` states durable constraints, applicability, and verification.
- A `workflow` declares phases, gates, dependencies, validation, and completion
  expectations. It never starts or controls an agent process.

Both types carry a version and digest and may declare applicability tags, path
patterns, dependencies, visibility, and read policy.

The `pr-feedback-repair` workflow adds a structured asynchronous-review
contract. Kit resolves it locally; a coding-agent host uses native GitHub tools
or event wakeups for bounded waiting, collection, repair, and thread updates.

## Trust boundaries

- Contract resolution is local-only and read-only.
- Registry administration may read the configured source.
- Initialization validates the complete catalog and write plan before changing
  the project.
- Reconciliation is read-only by default, applies only explicit conflict-free
  plans, and requires an exact artifact acceptance to replace customization.
- Materialized paths are confined to the project and managed routing is bounded
  by marker comments.
- Provider success descriptions, head identity, timeout, and rate limits remain
  explicit workflow states; success alone is never treated as completed review.

## Source providers

GitHub repository and branch catalogs are the first remote source provider.
The registry core is provider-neutral, and `.kit.yaml` owns source
configuration. A local path source supports offline development and self-hosted
repositories.

## Product boundary

Kit does not implement features, infer a task from prose, run prompts, dispatch
subagents, execute validation, manage CI or pull requests, or supervise an
agent loop. The coding agent and repository-native tooling own those actions.
The separate `git-wt` binary remains available outside this boundary.
