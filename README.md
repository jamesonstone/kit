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

# start implementation (outputs context for coding agents)
kit implement my-feature

# check status anytime
kit status
```

## Commands

### Project Initialization

| Command              | Description                            |
| -------------------- | -------------------------------------- |
| `kit init`           | Initialize a new Kit project           |

### Core Development Loop

| Command                           | Description                                                       |
| --------------------------------- | ----------------------------------------------------------------- |
| `kit oneshot <feature>`           | **Flagship** — scaffold all artifacts + combined agent prompt     |
| `kit spec <feature>`              | Create or open a feature specification                            |
| `kit plan <feature>`              | Create or open an implementation plan                             |
| `kit tasks <feature>`             | Create or open a task list                                        |
| `kit implement [feature]`         | Output implementation context for coding agents                   |
| `kit status`                      | Show current feature status for coding agents                     |

### Verification & State

| Command              | Description                                |
| -------------------- | ------------------------------------------ |
| `kit check <feature>`| Validate feature documents                 |
| `kit rollup`         | Generate PROJECT_PROGRESS_SUMMARY.md       |
| `kit code-review`    | Output instructions for branch code review |

### Context Management

| Command                   | Description                                       |
| ------------------------- | ------------------------------------------------- |
| `kit handoff [feature]`   | Output context for fresh agent session            |
| `kit summarize [feature]` | Output context summarization instructions         |
| `kit reflect [feature]`   | Output reflection/verification instructions       |
| `kit brainstorm [topic]`  | Generate a brainstorming scaffold document        |

### Utility

| Command              | Description                                          |
| -------------------- | ---------------------------------------------------- |
| `kit agentsmd`       | Create or overwrite AGENTS.md with full template     |
| `kit scaffold-agents`| Create or update agent pointer files                 |
| `kit completion`     | Generate shell autocompletion script                 |

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

## Oneshot — The Flagship Command

The core loop above (spec → plan → tasks → implement → reflect) is the full workflow. **`kit oneshot`** collapses it into a single command. It is the fastest way to get value from Kit.

The idea: do your deep thinking *before* you enter code.

### The Two-Phase Research Model

```text
┌─────────────────────────────────────────────────────────────────────────┐
│  Phase A: Foundation Research (you + Foundation LLM + Notion)           │
│                                                                         │
│    You  ◄──────►  Foundation LLM  ◄──────►  Notion / Notes              │
│                                                                         │
│    Iterate many times. Brainstorm, challenge assumptions, explore       │
│    tradeoffs. Refine until you have a succinct, information-dense       │
│    specification — the "brainstorming spec".                            │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Phase B: Codebase-Aware Refinement (kit oneshot)                       │
│                                                                         │
│    kit oneshot my-feature --spec-file brainstorm.md                     │
│                                                                         │
│    Kit scaffolds SPEC.md, PLAN.md, TASKS.md and outputs a prompt        │
│    that drives a coding agent through a new line of questioning —       │
│    one that takes into account the codebase as it actually is:          │
│    existing patterns, architecture, constraints, and conventions.       │
│                                                                         │
│    The agent fills out every document, asks for clarification,          │
│    and reaches >= 95% understanding before entering the                 │
│    pre-implementation phase.                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

**Phase A** is where the hard intellectual work happens. Use a Foundation LLM (Claude, GPT, Gemini — whatever you prefer) connected to Notion or your note-taking tool of choice. Iterate *many times*. The goal is a short, dense specification that captures the problem, constraints, goals, and rough approach — without any codebase-specific detail.

**Phase B** is where Kit takes over. The brainstorming spec you built in Notion becomes the input to `kit oneshot`. Kit creates all the artifact files, then outputs a comprehensive prompt that drives a coding agent through codebase-aware refinement. The agent reads your codebase, applies its patterns, and fills out SPEC.md, PLAN.md, and TASKS.md — enhancing the research you already did with the reality of the code as it exists today.

### Usage

```bash
# interactive — Kit prompts you to paste your brainstorming spec
kit oneshot my-feature

# from a file — pipe your Notion export or brainstorm document directly
kit oneshot my-feature --spec-file docs/brainstorm-export.md

# inline — for short specs
kit oneshot my-feature --spec "Add CSV export with streaming for large datasets"

# copy the agent prompt to clipboard instead of printing
kit oneshot my-feature --spec-file brainstorm.md --copy
```

After running, paste the generated prompt into your coding agent. The agent drives the entire workflow autonomously — clarifying, drafting, and refining each document — until all artifacts are complete and ready for `kit implement`.

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
