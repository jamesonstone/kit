# Kit

```text
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║
██╔═██╗ ██║   ██║
██║  ██╗██║   ██║
╚═╝  ╚═╝╚═╝   ╚═╝

              coding-agent context from repository evidence
```

Kit is a provider-neutral, repository-local evidence and contract harness for
coding agents. It materializes durable rules and workflow contracts, preserves
living specifications and project references, and deterministically resolves
the smallest ordered evidence set an agent needs. Kit does not infer project
truth, choose a model, or launch or supervise agents.

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/kit)](https://github.com/jamesonstone/kit/commits) [![Open issues](https://img.shields.io/github/issues/jamesonstone/kit)](https://github.com/jamesonstone/kit/issues) [![Pull requests](https://img.shields.io/github/issues-pr/jamesonstone/kit)](https://github.com/jamesonstone/kit/pulls) [![Release](https://img.shields.io/github/v/release/jamesonstone/kit)](https://github.com/jamesonstone/kit/releases)
<!-- END KIT-MANAGED README BADGES -->

## Major Update

Kit 2.0 intentionally removes low-use and duplicative commands. Existing
repositories keep their files and historical specifications, but should preview
`kit reconcile --include-files --dry-run --diff` before applying managed-file
updates. `kit reconcile` itself retains its existing behavior and remains the
canonical drift-maintenance surface.

See the [v2 migration guide](docs/migration-v2.md), [release
notes](docs/releases/v2.0.0.md), and [command guide](docs/commands.md).

## Primary Flow

```bash
kit init
kit capabilities context resolve --json
kit spec my-feature
kit context resolve --workflow implementation-delivery --feature my-feature --json
# coding agent plans, implements, validates, and curates repository memory
kit check my-feature
kit reconcile --all
```

Use `repository-bootstrap`, `implementation-delivery`,
`repository-maintenance`, `pr-feedback-repair`, or
`cross-repository-program-coordination` as the workflow slug. Resolution is
local-only, read-only, deterministic, and non-networked. A blocked result is a
hard evidence gap, not clean completion.

## Supported Command Surface

| Area | Commands |
| --- | --- |
| Bootstrap and memory | `kit init`, `kit spec`, `kit instructions` |
| Agent evidence | `kit capabilities`, `kit context resolve` |
| Execution prompts | `kit dispatch`, `kit pr fix` |
| Rules and maintenance | `kit rules add|list|view|link`, `kit registry status`, `kit reconcile`, `kit health` |
| Inspection and validation | `kit status`, `kit check`, `kit config check`, `kit aws verify` |
| Local usage | `kit usage [report|status|refresh|clear|enable|disable]` |
| Harness and utilities | `kit improve run`, `kit upgrade`, `kit version`, `kit completion` |

The separate `git-wt` binary remains available and unchanged. Kit rules use
native `git worktree` as the portable authority; `git wt` is an optional
manual convenience.

## Local Usage Data

Kit records minimal local command events by default so maintainers can identify
unused surfaces using evidence rather than intuition. It never records command arguments, output, repository names, paths, file content, environment values, or secrets; project identity is local and pseudonymous. Data remains on the
machine, is retained for at most 365 days, and is capped at 16 MiB total with
2 MiB shards.

```bash
kit usage status
kit usage report --since 90d
kit usage refresh
kit usage disable --global
kit usage clear --all --yes
```

A global disable is absolute. Project-level enable or disable is stored in that
repository's `.kit.yaml`. Usage commands do not record themselves.

## Install

```bash
go install github.com/jamesonstone/kit/cmd/kit@latest
GOBIN="$HOME/.local/bin" go install github.com/jamesonstone/kit/cmd/git-wt@latest
```

Or clone the repository and run `make build`. Enable repository-managed hooks
for the clone with `make install-git-hooks`.

## Documentation

- [Overview](docs/overview.md)
- [Commands](docs/commands.md)
- [Coding-agent workflow](docs/workflows.md)
- [Migration to v2](docs/migration-v2.md)
- [Agent routing](docs/agents/README.md)
- [Rules and references](docs/references/README.md)
- [Project Constitution](docs/CONSTITUTION.md)

## License

MIT

## Maintainers

<!-- BEGIN KIT-MANAGED README MAINTAINERS -->
Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
<!-- END KIT-MANAGED README MAINTAINERS -->
