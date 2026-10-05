```text
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║
██╔═██╗ ██║   ██║
██║  ██╗██║   ██║
╚═╝  ╚═╝╚═╝   ╚═╝

              repository-local contract for coding agents
```

Kit is a provider-neutral, repository-local contract for coding agents. It
installs one universal agent contract and contextual rules, keeps project memory
(Constitution, living specs, testing reference), validates it, and migrates any
existing Kit project to the current structure without touching the primary
checkout. Kit does not infer project truth, choose a model, or launch or
supervise agents.

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/kit)](https://github.com/jamesonstone/kit/commits) [![Open issues](https://img.shields.io/github/issues/jamesonstone/kit)](https://github.com/jamesonstone/kit/issues) [![Pull requests](https://img.shields.io/github/issues-pr/jamesonstone/kit)](https://github.com/jamesonstone/kit/pulls) [![Release](https://img.shields.io/github/v/release/jamesonstone/kit)](https://github.com/jamesonstone/kit/releases)
<!-- END KIT-MANAGED README BADGES -->

## How Kit Works

Kit gives every coding agent one compact universal contract (rendered into
`AGENTS.md`, `CLAUDE.md`, and Copilot instructions), contextual rules that are
read only when their trigger applies, and project memory in
`docs/CONSTITUTION.md`, `docs/specs/`, and `docs/references/testing.md`. The
binary ships the rules it was tested with and adds deterministic commands for
bootstrap, validation, drift maintenance, spec allocation, and AWS identity
checks. Agents plan, explore, and delegate with their own native tools.

New project:

```bash
kit init
kit spec my-feature
# coding agent plans, implements, validates, and curates repository memory
kit check --project
```

Existing Kit project, created by any Kit release:

```bash
kit reconcile --dry-run --diff   # preview
kit reconcile                    # migrate to the current structure
```

Mint deployment project with an existing team-first schema 2 `.mint.yaml` and
reviewed build, promotion and observation adapters:

```bash
kit init --mint                         # new Kit project
kit reconcile --mint --dry-run --diff    # existing Kit project: preview
kit reconcile --mint                    # create a missing controller
```

Kit uses the embedded Mint v0.5.0 renderer to generate repository Actions for
promotion, hotfix, rollback, resume, observation and reconciliation. It uses the
policy's environment names and adapter workflow names. Existing controllers,
policies and adapters stay project owned, even with `--force`. No registry,
provider or production environment is assumed. Package and artifact projects
retain their existing publication workflows. Scaffolding keeps the
`MINT_RELEASE_ENABLED` gate; it does not activate deployments. Review source-build
isolation, adapter manifests, runtime configuration identities and verified
baselines before enabling that gate. New-project `kit init` without `--mint`
continues to install lifecycle guidance without adopting Mint automatically.

Diagnose without changing anything:

```bash
kit health
```

Reconcile keeps everything the project wrote: guidance outside the managed
contract block, edited Kit sections, edited rules, and project rules. It removes
retired Kit files (`docs/agents/`, `docs/references/workflows/`, the progress
summary, retired rules) only when they are exactly as a Kit release generated
them and committed to Git, and reports every file it keeps. From the primary
checkout it writes to a `kit-reconcile` linked worktree for review.

## Command Surface

| Area | Commands |
| --- | --- |
| Bootstrap and memory | `kit init`, `kit spec` |
| Rules | `kit rules add\|list\|view` |
| Validation and maintenance | `kit check`, `kit reconcile`, `kit health`, `kit registry status`, `kit status`, `kit config check`, `kit aws verify` |
| Local usage | `kit usage [report\|status\|refresh\|clear\|enable\|disable]` |
| Utilities | `kit upgrade`, `kit version`, `kit completion` |

See the [command guide](docs/commands.md), the [v3 migration
guide](docs/migration-v3.md), and the [release notes](docs/releases/v3.0.0.md).

## Local Usage Data

Kit records minimal local command events by default so maintainers can identify
unused surfaces using evidence rather than intuition. It never records command arguments, output, repository names, paths, file content, environment values, or secrets; project identity is local and pseudonymous. Data remains on the
machine, is retained for at most 365 days, and is capped at 16 MiB total with
2 MiB shards. Unversioned development builds, test binaries, and processes with
`KIT_USAGE_DISABLED=1` never record.

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
go install github.com/jamesonstone/kit/v3/cmd/kit@latest
```

Or clone the repository and run `make build`. Enable repository-managed hooks
for the clone with `make install-git-hooks`.

## Documentation

- [Overview](docs/overview.md)
- [Commands](docs/commands.md)
- [Migration to v3](docs/migration-v3.md)
- [Historical v2 migration](docs/migration-v2.md)
- [Rules and references](docs/references/README.md)
- [Project Constitution](docs/CONSTITUTION.md)

## License

MIT

## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
