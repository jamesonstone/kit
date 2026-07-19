```text
K   K  IIIII  TTTTT
K  K     I      T
KKK      I      T
K  K     I      T
K   K  IIIII    T

          native plan → implementation → curated memory
```

🎒 **Kit** is a portable, agent-agnostic harness that turns consequential agent
work into durable repository memory.

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/kit)](https://github.com/jamesonstone/kit/commits) [![Open issues](https://img.shields.io/github/issues/jamesonstone/kit)](https://github.com/jamesonstone/kit/issues) [![Pull requests](https://img.shields.io/github/issues-pr/jamesonstone/kit)](https://github.com/jamesonstone/kit/pulls) [![Release](https://img.shields.io/github/v/release/jamesonstone/kit)](https://github.com/jamesonstone/kit/releases)
<!-- END KIT-MANAGED README BADGES -->

The host agent's native planning capability owns research, clarification,
design, and implementation planning. Kit ensures the accepted plan, material
decisions, discoveries, validation, and actual outcome survive in the
repository when code and tests alone cannot preserve the important why.

## Start Here

| Need | Go To |
| --- | --- |
| 🧭 Understand what Kit is | [docs/overview.md](docs/overview.md) |
| ⚙️ Install and use commands | [docs/commands.md](docs/commands.md) |
| 🔁 Understand the memory workflow | [docs/workflows.md](docs/workflows.md) |
| 📚 Browse all documentation | [docs/README.md](docs/README.md) |
| 🧱 Read project invariants | [docs/CONSTITUTION.md](docs/CONSTITUTION.md) |
| 🤖 Read agent routing docs | [docs/agents/README.md](docs/agents/README.md) |

## Install

```bash
go install github.com/jamesonstone/kit/cmd/kit@latest
```

Or build from source:

```bash
git clone https://github.com/jamesonstone/kit.git
cd kit
make build
```

Enable repository-managed hooks for this clone:

```bash
make install-git-hooks
```

## Quick Start

```bash
# initialize Kit in a repository
kit init

# scaffold durable feature memory after native planning
kit spec my-feature

# inspect progress
kit status --all

# check compact registry freshness, then apply safe Kit maintenance
kit registry status
kit health --dry-run --diff

# reorient before continuing
kit resume my-feature

# inspect command behavior before choosing a command
kit capabilities --search spec

# validate or repair project configuration
kit config check

# verify a configured AWS identity before AWS work
kit aws verify
```

The default feature workflow is:

```text
native agent plan → semantic memory decision → SPEC.md when required → implementation → validation → curated repository memory
```

`SPEC.md` is the durable home for material feature rationale. Project
invariants belong in `CONSTITUTION.md`, reusable practices in references or
rules, and domain knowledge in its existing canonical documentation. A
justified `not required` memory decision is valid when code and tests preserve
the complete durable truth.

## Common Commands

| Command | Purpose |
| --- | --- |
| `kit init` | Initialize or refresh project scaffolding, including a project-owned Makefile starter |
| `kit spec <feature>` | Non-interactively scaffold, adopt, or orient a living specification |
| `kit spec <feature> --legacy-supervisor` | Temporarily run the deprecated V2 lifecycle supervisor |
| `kit loop workflow <feature>` | Deprecated V2 compatibility loop; V3 specs use native planning |
| `kit loop review` | Review changed code until local correctness converges |
| `kit pr fix` | Select or target a PR and prepare a review-feedback dispatch prompt |
| `kit status --all` | Show project-wide feature state |
| `kit registry status` | Show compact registry and Kit-managed file freshness; supports `--json` |
| `kit health` | Apply safe Kit-managed updates and validate project health; preview with `--dry-run --diff` |
| `kit map --all` | Show the project document map |
| `kit capabilities --search <term>` | Inspect command behavior and mutation boundaries |
| `kit config check` | Validate schema-versioned `.kit.yaml` and offer safe interactive repairs |
| `kit aws verify` | Verify the configured AWS profile, account, and ARN against `.kit.yaml` |
| `kit improve run --suite prompt-system` | Run deterministic prompt regression and size checks |
| `kit instructions [--version=vN]` | Print current provider-neutral coding-agent instructions as raw Markdown; use `--version=vN` to retrieve an immutable historical version for reproducible use |
| `kit prompt list` | List reusable prompt-library entries |
| `kit legacy --help` | List v1 staged workflow commands retained for migration |

See [docs/commands.md](docs/commands.md) for the full command guide.

## Documentation Map

- 🧭 [Overview](docs/overview.md) - product model, principles, positioning, and cross-domain concepts.
- ⚙️ [Commands](docs/commands.md) - installation, command groups, prompt behavior, scaffold refresh, and prompt libraries.
- 🔁 [Workflows](docs/workflows.md) - native planning, semantic memory decisions, living specs, compatibility, and curation.
- 🧱 [Constitution](docs/CONSTITUTION.md) - project contract, invariants, and repository rules.
- 🤖 [Agent Docs](docs/agents/README.md) - repo-local agent routing and RLM guidance.
- 📌 [References](docs/references/README.md) - durable project references and rulesets.
- 🧪 [Testing Reference](docs/references/testing.md) - testing guidance.
- 🛠️ [Tooling Reference](docs/references/tooling.md) - durable tooling notes.

## License

MIT

## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
