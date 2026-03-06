```text
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║
██╔═██╗ ██║   ██║
██║  ██╗██║   ██║
╚═╝  ╚═╝╚═╝   ╚═╝
```

**Spec-Driven Development Toolkit**

🎒 Portable specification-driven development framework without vendor lock-in.

Kit is a document-centered CLI that helps teams reach high-confidence understanding of a problem and its solution _before_ implementation.

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

# optionally create brainstorm research first
kit brainstorm my-feature

# create a feature specification
kit spec my-feature

# create implementation plan
kit plan my-feature

# create task list
kit tasks my-feature

# start implementation (outputs context for coding agents)
kit implement my-feature

# check status anytime
kit status
```

## Commands

### Project Initialization

| Command    | Description                  |
| ---------- | ---------------------------- |
| `kit init` | Initialize a new Kit project |

### Core Development Loop

| Command                    | Description                                                             |
| -------------------------- | ----------------------------------------------------------------------- |
| `kit brainstorm [feature]` | Interactively create `BRAINSTORM.md` and a planning-only `/plan` prompt |
| `kit spec <feature>`       | Create or open a feature specification                                  |
| `kit plan <feature>`       | Create or open an implementation plan                                   |
| `kit tasks <feature>`      | Create or open a task list                                              |
| `kit implement [feature]`  | Output implementation context for coding agents                         |
| `kit status`               | Show current feature status for coding agents                           |

### Verification & State

| Command               | Description                                |
| --------------------- | ------------------------------------------ |
| `kit check <feature>` | Validate feature documents                 |
| `kit rollup`          | Generate PROJECT_PROGRESS_SUMMARY.md       |
| `kit code-review`     | Output instructions for branch code review |

### Context Management

| Command                   | Description                                 |
| ------------------------- | ------------------------------------------- |
| `kit handoff [feature]`   | Output context for fresh agent session      |
| `kit summarize [feature]` | Output context summarization instructions   |
| `kit reflect [feature]`   | Output reflection/verification instructions |

### Utility

| Command               | Description                                      |
| --------------------- | ------------------------------------------------ |
| `kit agentsmd`        | Create or overwrite AGENTS.md with full template |
| `kit scaffold-agents` | Create or update agent pointer files             |
| `kit completion`      | Generate shell autocompletion script             |

## Artifact Pipeline

**Project Initialization** (run once, update as needed):

```text
┌──────────────┐
│ Constitution │  ← global constraints, principles, priors
└──────────────┘
```

**Optional Research Step**:

```text
┌──────────────┐
│ Brainstorm   │  ← codebase research, options, affected files
└──────────────┘
```

**Core Development Loop**:

```text
┌──────────────┐    ┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐
│ Brainstorm   │ ─▶ │ Specification │ ─▶ │ Plan │ ─▶ │ Tasks │ ─▶ │ Implementation │ ─▶ │ Reflection │ ─┐
└──────────────┘    └───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │
       ▲                                                                            │
       └────────────────────────────────────────────────────────────────────────────┘
```

**Artifact Details**:

1. **Constitution** — strategy, patterns, long-term vision (kept updated)
2. **Brainstorm** — optional research artifact with codebase findings and strategy
3. **Specification** — what is being built and why
4. **Plan** — how it will be built
5. **Tasks** — executable work units
6. **Implementation** — execution outside Kit's core scope
7. **Reflection** — verify correctness, refine understanding

## Brainstorm — Interactive Research Entry Point

`kit brainstorm` is now the optional front door for new feature work. It asks for:

1. the feature name using the same normalization rules as `kit spec`
2. a short user thesis describing the issue or feature

Then Kit:

- creates or reuses `docs/specs/<feature>/`
- creates `BRAINSTORM.md` as the first artifact in that directory
- outputs a planning-only prompt that starts with `/plan`
- tells the coding agent to research the codebase, ask questions, and avoid implementation
- requires the agent to continue until understanding reaches at least `95%`

### Why this matters

`BRAINSTORM.md` becomes the durable bridge between early ideation and the formal artifact pipeline. When present, downstream commands use it as research context while still treating `SPEC.md`, `PLAN.md`, and `TASKS.md` as the binding execution contract.

### Typical flow

```text
You / team idea
  ↓
kit brainstorm my-feature
  ↓
BRAINSTORM.md + planning-only /plan prompt
  ↓
kit spec my-feature
  ↓
kit plan my-feature
  ↓
kit tasks my-feature
  ↓
kit implement my-feature
  ↓
kit reflect my-feature
```

### Usage

```bash
# interactive brainstorm for a new feature
kit brainstorm my-feature

# open or continue an existing brainstorm
kit brainstorm my-feature --copy

# write the generated /plan prompt to a file
kit brainstorm my-feature --output tmp/brainstorm-prompt.md
```

### What goes in `BRAINSTORM.md`

- summary of the issue or opportunity
- user thesis in the user's own words
- codebase findings and relevant architecture notes
- affected files with concrete paths
- unresolved questions and viable options
- recommended strategy and the next workflow step

## Project Structure

```text
.kit.yaml                    # configuration
docs/
  CONSTITUTION.md            # project-wide constraints
  PROJECT_PROGRESS_SUMMARY.md
  specs/
    0001-my-feature/
      BRAINSTORM.md         # optional research artifact
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
