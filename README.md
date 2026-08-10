# Kit

> Repository-native contracts for coding agents.

Kit materializes registry-backed rules and declarative workflows into a
repository, then resolves the exact local contract a coding agent should
follow. The Markdown in the repository is canonical; Kit's JSON contract is a
deterministic projection. Kit does not launch, supervise, or select agents.

[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/kit)](https://github.com/jamesonstone/kit/commits)
[![Release](https://img.shields.io/github/v/release/jamesonstone/kit)](https://github.com/jamesonstone/kit/releases)

## Major update

This is an intentional semantic-version major reset. The `kit` CLI has been
reduced to the rules, registry, reconciliation, and contract surfaces. Existing
projects must preview the schema-v2 migration with `kit reconcile --diff`
before applying it with `kit reconcile --apply`.

See the [major-release migration guide](docs/migrations/coding-agent-first-major.md)
and [release notes](docs/releases/coding-agent-first-major.md).

## Primary flow

```bash
# Materialize every project-visible rule, workflow, routing block, and provenance record.
kit init

# Give a coding agent the stable, local-only contract for its task.
kit contract resolve --workflow implementation-delivery --path internal/example.go

# agent implementation happens in the repository

# Preview, then explicitly apply conflict-free registry and routing drift.
kit reconcile --diff
kit reconcile --apply
```

In short: `kit init` → `kit contract resolve` → agent implementation →
`kit reconcile`.

For asynchronous pull-request review repair, resolve
`kit contract resolve --workflow pr-feedback-repair`. The local workflow
defines bounded CodeRabbit and human-feedback intake while the coding agent or
host owns `gh`, waiting, repair, and delivery; no legacy `kit pr fix` command
or agent runtime is restored.

## Protected rules lifecycle

These command paths retain their names and core purpose across the major
release. Their flags and presentation may change.

| Command | Purpose |
| --- | --- |
| `kit init` | Materialize project-visible rulesets, workflows, routing, and provenance. |
| `kit reconcile` | Preview and safely apply registry, migration, and routing drift. |
| `kit rules add` | Install a registry-backed or project-local ruleset. |
| `kit rules list` | List rulesets with availability, state, and provenance. |
| `kit rules view` | Inspect local, registry, or diff views of a ruleset. |
| `kit registry status` | Report freshness, conflicts, and required actions. |

Typed artifact administration is available through `kit registry
list|view|add|status`. Coding agents consume `kit contract resolve`; it never
uses the network, writes files, runs Git, or infers intent from task prose.

See the [command guide](docs/commands.md) for flags, JSON behavior, and exit
codes.

## Install

```bash
go install github.com/jamesonstone/kit/cmd/kit@latest
go install github.com/jamesonstone/kit/cmd/git-wt@latest
```

`git-wt` remains a separate, unchanged convenience binary for native Git
worktrees. It is not part of the agent-contract runtime.

## Documentation

- [Product overview](docs/overview.md)
- [Agent and reconciliation workflows](docs/workflows.md)
- [Registry catalog](registry/catalog.yaml)
- [Project constitution](docs/CONSTITUTION.md)
- [Documentation index](docs/README.md)

## License

MIT
