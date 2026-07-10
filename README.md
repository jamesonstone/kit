```text
╭──────────────────────────────────────────────────────────────────────╮
│                                                                      │
│   K                                                                  │
│      keep one durable SPEC.md as the source of truth                 │
│                                                                      │
│   I                                                                  │
│      interrogate ambiguity before implementation                     │
│                                                                      │
│   T                                                                  │
│      turn agent work into validation evidence                        │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│                    Kit v2 Thought-Work Harness                       │
│             idea → clarity → agent teams → evidence                  │
╰──────────────────────────────────────────────────────────────────────╯
```

🎒 **Kit v2 Thought-Work Harness** is a portable, agent-agnostic harness for
disciplined thought work.

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/kit)](https://github.com/jamesonstone/kit/commits) [![Open issues](https://img.shields.io/github/issues/jamesonstone/kit)](https://github.com/jamesonstone/kit/issues) [![Pull requests](https://img.shields.io/github/issues-pr/jamesonstone/kit)](https://github.com/jamesonstone/kit/pulls) [![Release](https://img.shields.io/github/v/release/jamesonstone/kit)](https://github.com/jamesonstone/kit/releases)
<!-- END KIT-MANAGED README BADGES -->

Its strongest engine is a document-first, spec-driven workflow for software
projects, but the same structure works for research, operations, writing,
policy, strategy, and other work where constraints, evidence, reflection, and
handoff matter.

## Start Here

| Need | Go To |
| --- | --- |
| 🧭 Understand what Kit is | [docs/overview.md](docs/overview.md) |
| ⚙️ Install and use commands | [docs/commands.md](docs/commands.md) |
| 🔁 Understand the v2 workflow | [docs/workflows.md](docs/workflows.md) |
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

# start the v2 single-SPEC workflow
kit spec my-feature

# inspect progress
kit status --all

# reorient before continuing
kit resume my-feature

# inspect command behavior before choosing a command
kit capabilities --search spec
```

The default feature workflow is:

```text
idea → kit spec <feature> → SPEC.md evidence and readiness gates → same-thread implementation → validation → reflection → delivery
```

In v2, `SPEC.md` is the single durable feature artifact. Legacy
`BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files remain readable historical
artifacts and are available only through `kit legacy` workflows.

## Common Commands

| Command | Purpose |
| --- | --- |
| `kit init` | Initialize or refresh Kit-managed project scaffolding |
| `kit spec <feature>` | Start or resume the evidence-first v2 `SPEC.md` workflow |
| `kit loop workflow <feature>` | Execute workflow phases through a configured local agent loop |
| `kit loop review` | Review changed code until local correctness converges |
| `kit pr fix` | Select or target a PR and prepare a review-feedback dispatch prompt |
| `kit status --all` | Show project-wide feature state |
| `kit map --all` | Show the project document map |
| `kit capabilities --search <term>` | Inspect command behavior and mutation boundaries |
| `kit improve run --suite prompt-system` | Run deterministic prompt regression and size checks |
| `kit prompt list` | List reusable prompt-library entries |
| `kit legacy --help` | List v1 staged workflow commands retained for migration |

See [docs/commands.md](docs/commands.md) for the full command guide.

## Documentation Map

- 🧭 [Overview](docs/overview.md) - product model, principles, positioning, and cross-domain concepts.
- ⚙️ [Commands](docs/commands.md) - installation, command groups, prompt behavior, scaffold refresh, and prompt libraries.
- 🔁 [Workflows](docs/workflows.md) - v2 single-`SPEC.md` workflow, v1 foundations, autonomous loops, usage examples, and project structure.
- 🧱 [Constitution](docs/CONSTITUTION.md) - project contract, invariants, and repository rules.
- 🤖 [Agent Docs](docs/agents/README.md) - repo-local agent routing and RLM guidance.
- 📌 [References](docs/references/README.md) - durable project references and rulesets.
- 🧪 [Testing Reference](docs/references/testing.md) - testing guidance.
- 🛠️ [Tooling Reference](docs/references/tooling.md) - durable tooling notes.

## License

MIT

## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
