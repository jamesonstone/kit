```text
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║
██╔═██╗ ██║   ██║
██║  ██╗██║   ██║
╚═╝  ╚═╝╚═╝   ╚═╝
```

**General-Purpose Thought-Work Harness**

🎒 Portable harness without vendor lock-in.

Kit is a general-purpose harness for disciplined thought work.
Its deepest engine is a document-first, spec-driven workflow, but the harness
also supports ad hoc execution, catch-up, handoff, summarization, review, and
orchestration.

Today, the shipped command surface is packaged around repository and software
work. The underlying concepts are broader: they generalize to research,
strategy, operations, policy, writing, analysis, and other fields where you
need explicit constraints, structured exploration, planning, execution, and
reflection.

Harness principles:

- 🧰 harness-first, workflow-second
- 📄 documents are the source of truth
- 🧠 spec-driven planning is the strongest engine for ambiguous or high-risk work
- ⚡ ad hoc work stays lightweight but still verified
- 🤝 portable and agent-agnostic by default
- 🔍 explicit gates beat hidden automation
- 🔄 reflection closes the loop after code changes

### 🌍 Cross-Domain Concepts

The artifact model is broader than software:

| Kit Concept | In Software | In Research | In Strategy / Ops | In Writing / Policy |
| ----------- | ----------- | ----------- | ----------------- | ------------------- |
| `CONSTITUTION.md` | engineering constraints | study constraints | operating principles | editorial or policy constraints |
| `BRAINSTORM.md` | codebase research | literature scan | landscape scan | source gathering and framing |
| `SPEC.md` | feature requirements | research question or hypothesis | decision brief | argument or policy brief |
| `PLAN.md` | implementation plan | study design | rollout plan | outline and revision plan |
| `TASKS.md` | execution checklist | experiment tasks | workback schedule | drafting and review checklist |
| `implement` | coding and integration | running the study | executing the change | drafting and editing |
| `reflect` | verification and regression review | results review | retro and validation | revision review and critique |
| `catchup` / `summarize` / `handoff` | resume or transfer context | resume the investigation | transfer project state | transfer editorial context |

The names may be software-flavored today, but the structure is general:
constraints, research, specification, planning, execution, verification, and
transfer are common to most serious thought work.

## ⚙️ Installation

```bash
go install github.com/jamesonstone/kit/cmd/kit@latest
```

Or build from source:

```bash
git clone https://github.com/jamesonstone/kit.git
cd kit
make build
```

To enable the repository-managed Git hooks for this clone:

```bash
make install-git-hooks
```

This configures `core.hooksPath` to use `.githooks/`, including a `pre-commit`
hook that runs `make build` before every `git commit`. If the build fails, the
commit is blocked.

## 🚀 Quick Start

```bash
# initialize the project
kit init

# optionally capture research first
kit brainstorm my-feature

# write the spec
kit spec my-feature

# define the plan
kit plan my-feature

# break work into tasks
kit tasks my-feature

# start implementation
# runs the readiness gate, then outputs coding-agent context
kit implement my-feature

# reorient before resuming work
kit catchup my-feature

# inspect progress at any time
kit status

# mark all eligible active features complete
kit complete --all
```

## 🧰 Commands

### 🏁 Project Initialization

| Command                  | Description                                    |
| ------------------------ | ---------------------------------------------- |
| `kit init`               | Initialize a new Kit project                   |
| `kit scaffold [feature]` | Create a feature directory with pipeline files |

### 🔁 Core Development Loop

| Command                    | Description                                                                               |
| -------------------------- | ----------------------------------------------------------------------------------------- |
| `kit brainstorm [feature]` | Interactively create `BRAINSTORM.md` and track phase dependencies in a `/plan` prompt     |
| `kit spec <feature>`       | Create or open a feature specification, perform skills discovery, and track dependencies  |
| `kit plan <feature>`       | Create or open an implementation plan and track planning dependencies                     |
| `kit tasks <feature>`      | Create or open a task list                                                                |
| `kit implement [feature]`  | Run the implementation readiness gate and output implementation context for coding agents |
| `kit reflect [feature]`    | Output reflection/verification instructions                                               |
| `kit complete [feature]`   | Mark a feature complete; supports `--all` for all eligible active features                |
| `kit status`               | Show current feature status; supports `--json` and includes the running Kit version       |

### ✅ Verification & State

| Command               | Description                                                |
| --------------------- | ---------------------------------------------------------- |
| `kit check <feature>` | Validate feature documents and populated required sections |
| `kit rollup`          | Generate PROJECT_PROGRESS_SUMMARY.md                       |
| `kit code-review`     | Output instructions for branch code review                 |

### 🔄 Context Management

| Command                   | Description                                                                                  |
| ------------------------- | -------------------------------------------------------------------------------------------- |
| `kit handoff [feature]`   | Prompt the current agent session to sync docs, dependency inventories, and prepare a handoff |
| `kit summarize [feature]` | Output context summarization instructions                                                    |
| `kit catchup [feature]`   | Output a feature catch-up prompt that stays in plan mode                                     |

### 🤝 Agent Orchestration

| Command        | Description                                                                 |
| -------------- | --------------------------------------------------------------------------- |
| `kit dispatch` | Output a discovery-first prompt for clustering tasks and queueing subagents |

Prompt-producing commands default to subagent orchestration guidance. Pass
`--single-agent` when you explicitly want to keep the work in one lane.

Use `kit dispatch` when you need the full overlap-clustering and queue-planning
workflow for a raw task set. Use the default prompt path when the agent should
use subagents opportunistically, and use `kit dispatch` when you want a formal
discovery report, overlap clustering, and explicit approval before launch.

### 📋 Output Behavior

Prompt-producing commands that expose `--output-only` copy their generated
output to the clipboard by default. Pass `--output-only` to print the raw
prompt or output to stdout instead, or combine `--output-only --copy` to do
both.

In interactive terminals, Kit also uses clearer section spacing and semantic
emoji markers for help, status, selectors, and other human-readable guidance.
Raw `--output-only` payloads and `--json` output stay unchanged.

### ♻️ Prompt Regeneration

Feature-scoped prompt commands also accept `--prompt-only` to regenerate the
selected feature's prompt without mutating repository docs:
`kit brainstorm`, `kit spec`, `kit plan`, `kit tasks`, `kit implement`,
`kit reflect`, `kit catchup`, `kit handoff`, and `kit skill mine`. For
`brainstorm`, `spec`, `plan`, and `tasks`, `--prompt-only` skips scaffolding
and rollup writes, requires the existing artifact set, and uses the normal
existing-feature selector when no feature argument is provided.

### ⛏️ Skill Mining

| Command                     | Description                                                |
| --------------------------- | ---------------------------------------------------------- |
| `kit skill mine [feature]`  | Output skill extraction prompt for the active coding agent |
| `kit skills mine [feature]` | Alias for `kit skill mine`                                 |

### 🔧 Utility

| Command               | Description                                                                                 |
| --------------------- | ------------------------------------------------------------------------------------------- |
| `kit upgrade`         | Download and install the latest Kit release                                                 |
| `kit update`          | Alias for `kit upgrade`                                                                     |
| `kit version`         | Print the installed Kit version                                                             |
| `kit scaffold-agents` | Create or refresh repository instruction files and Copilot rules with safer overwrite modes |
| `kit completion`      | Generate shell autocompletion script                                                        |

`kit scaffold-agents` also supports the singular alias `kit scaffold-agent`.

When instruction files already exist:

- default mode skips them and suggests safer next steps
- `--append-only` merges missing Kit-managed sections without overwriting matched existing content
- `--force` overwrites existing files after confirmation
- `--force --yes` overwrites existing files without prompting for automation use

## 🗂️ Structured Engine: Artifact Pipeline

The artifact pipeline is Kit's most structured operating engine. It is not the
entire product, but it is the clearest path when a problem needs deliberate
discovery, planning, and execution control in any domain.

### 🏗️ Project Initialization

Run once, then refine as the project matures:

```text
┌──────────────┐
│ Constitution │  ← global constraints, principles, priors
└──────────────┘
```

### 🧠 Optional Research Step

```text
┌──────────────┐
│ Brainstorm   │  ← codebase research, options, affected files
└──────────────┘
```

### 🔁 Core Development Loop

```text
┌──────────────┐    ┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐
│ Brainstorm   │ ─▶ │ Specification │ ─▶ │ Plan │ ─▶ │ Tasks │ ─▶ │ Implementation │ ─▶ │ Reflection │ ─┐
└──────────────┘    └───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │
       ▲                                                                            │
       └────────────────────────────────────────────────────────────────────────────┘
```

### 📝 Artifact Details

1. 📜 **Constitution** — strategy, patterns, long-term vision
2. 🧠 **Brainstorm** — optional research, findings, dependencies, strategy
3. 📐 **Specification** — what is being built and why
4. 🗺️ **Plan** — how it will be built
5. ✅ **Tasks** — executable work units
6. 🛠️ **Implementation** — execution after the readiness gate passes
7. 🔍 **Reflection** — verification and learning after implementation

Spec-driven prompts must populate every section in `BRAINSTORM.md`, `SPEC.md`,
`PLAN.md`, and `TASKS.md`. If a section has no additional detail, replace the
placeholder comment with `not applicable`, `not required`, or
`no additional information required`.

## 🧠 Brainstorm — Interactive Research Entry Point

`kit brainstorm` is now the optional front door for new feature work. It asks for:

1. the feature name using the same normalization rules as `kit spec`
2. a short user thesis describing the issue or feature

Then Kit:

- creates or reuses `docs/specs/<feature>/`
- creates `BRAINSTORM.md` as the first artifact in that directory
- requires the coding agent to keep the `## DEPENDENCIES` table current with the inputs used during the brainstorm phase
- opens a vim-compatible editor by default for the multiline thesis, with step instructions and a press-any-key launch gate
- supports `--inline` to use terminal multiline entry with `Shift+Enter` and `Ctrl+J`, including consecutive blank lines
- keeps `--vim` and `--editor=vim` as explicit editor controls, though vim-mode is already the default for multiline free-text responses
- outputs a planning-only prompt that starts with `/plan`
- tells the coding agent to research the codebase, use numbered lists, ask questions in batches of up to 10, and avoid implementation
- requires the agent to include recommended defaults, accept `yes` / `y` for whole-batch approval and `yes 3, 4, 5` / `y 3, 4, 5` for numbered approval, state uncertainties, output percentage-understanding progress after each batch, and continue until the spec is precise enough for a production-quality solution

### 💡 Why this matters

`BRAINSTORM.md` becomes the durable bridge between early ideation and the formal artifact pipeline. When present, downstream commands use it as research context while still treating `SPEC.md`, `PLAN.md`, and `TASKS.md` as the binding execution contract.

### 🪜 Typical flow

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

`kit implement` begins with an implementation readiness gate that adversarially
challenges `CONSTITUTION.md`, optional `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`,
and `TASKS.md` before any code work starts. If the gate fails, update the
canonical docs first, then rerun the gate before implementing.

### ▶️ Usage

```bash
# interactive brainstorm for a new feature
kit brainstorm my-feature

# open or continue an existing brainstorm
kit brainstorm my-feature

# print the raw prompt to stdout instead of copying it
kit brainstorm my-feature --output-only

# regenerate the brainstorm prompt from an existing BRAINSTORM.md without touching repo docs
kit brainstorm my-feature --prompt-only

# opt out of default vim-mode and use inline multiline entry
kit brainstorm my-feature --inline

# write the generated /plan prompt to a file
kit brainstorm my-feature --output tmp/brainstorm-prompt.md
```

`kit spec <feature> --interactive` now opens a vim-compatible editor for each
free-text answer by default. Use `kit spec <feature> --interactive --inline`
to opt back into terminal multiline entry.

### 📄 What goes in `BRAINSTORM.md`

- summary of the issue or opportunity
- user thesis in the user's own words
- codebase findings and relevant architecture notes
- affected files with concrete paths
- unresolved questions and viable options
- recommended strategy and the next workflow step

## 🏛️ Project Structure

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

## ✨ Positioning

Kit is broader than a spec generator. It is a harness for disciplined
development workflows:

- structured planning when scope is unclear
- lightweight ad hoc execution when the change is contained
- recovery tools such as `catchup`, `summarize`, and `handoff`
- review and orchestration tools such as `code-review` and `dispatch`

Spec-driven development principles remain a core engine inside that harness,
not the only identity of the tool.

## ✨ Inspiration

Kit is inspired by GitHub's [spec-kit](https://github.com/github/spec-kit),
which pioneered the concept of specification-driven development. Kit keeps that
discipline where it helps most, then broadens it into a lighter, more portable,
general-purpose harness.

## 📚 Documentation

See [docs/specs/0000_INIT_PROJECT.md](docs/specs/0000_INIT_PROJECT.md) for the full specification.

## ⚖️ License

MIT

## 👤 Maintainer

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/jamesonstone">
        <img src="https://github.com/jamesonstone.png" width="100px;" alt="Jameson Stone"/>
        <br />
        <sub><b>Jameson Stone</b></sub>
      </a>
      <br />
      <sub>Lead Maintainer</sub>
    </td>
  </tr>
</table>
