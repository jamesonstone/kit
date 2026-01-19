# kit

🎒 Portable specification-driven development framework without vendor lock-in.

Kit is a document-centered CLI that helps teams reach high-confidence understanding of a problem and its solution *before* implementation.

## Installation

```bash
go install github.com/jamesonstone/kit/cmd/kit@latest
```

Or build from source:

```bash
git clone https://github.com/jamesonstone/kit.git
cd kit
make build
```

## Quick Start

```bash
# initialize a new project
kit init

# create a feature specification
kit spec my-feature

# create implementation plan
kit plan my-feature

# create task list
kit tasks my-feature

# validate documents
kit check my-feature
```

## Commands

| Command | Description |
| ------- | ----------- |
| `kit init` | Initialize a new Kit project |
| `kit spec <feature>` | Create a feature specification |
| `kit plan <feature>` | Create an implementation plan |
| `kit tasks <feature>` | Create a task list |
| `kit check <feature>` | Validate feature documents |
| `kit rollup` | Generate PROJECT_PROGRESS_SUMMARY.md |
| `kit scaffold-agents` | Create agent pointer files |
| `kit summarize [feature]` | Output context summarization instructions |
| `kit reflect [feature]` | Output reflection/verification instructions |
| `kit handoff [feature]` | Output context for fresh agent session |

## Artifact Pipeline

**Project Initialization** (run once, update as needed):

```text
┌──────────────┐
│ Constitution │  ← global constraints, principles, priors
└──────────────┘
```

**Core Development Loop**:

```text
┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐
│ Specification │ ─▶ │ Plan │ ─▶ │ Tasks │ ─▶ │ Implementation │ ─▶ │ Reflection │ ─┐
└───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │
       ▲                                                                            │
       └────────────────────────────────────────────────────────────────────────────┘
```

**Artifact Details**:

1. **Constitution** — strategy, patterns, long-term vision (kept updated)
2. **Specification** — what is being built and why
3. **Plan** — how it will be built
4. **Tasks** — executable work units
5. **Implementation** — execution outside Kit's core scope
6. **Reflection** — verify correctness, refine understanding

## Project Structure

```text
.kit.yaml                    # configuration
docs/
  CONSTITUTION.md            # project-wide constraints
  PROJECT_PROGRESS_SUMMARY.md
  specs/
    0001-my-feature/
      SPEC.md
      PLAN.md
      TASKS.md
      ANALYSIS.md            # optional
```

## Inspiration

Kit is inspired by GitHub's [spec-kit](https://github.com/github/spec-kit), which pioneered the concept of specification-driven development. However, spec-kit proved too verbose for my personal workflow. Kit distills the core ideas into a lighter, more portable tool.

## Documentation

See [docs/specs/0000_INIT_PROJECT.md](docs/specs/0000_INIT_PROJECT.md) for the full specification.

## License

MIT

## Maintainer

[@jamesonstone](https://github.com/jamesonstone)
