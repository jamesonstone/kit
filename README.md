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

| Kit Concept                                      | In Software                            | In Research                           | In Strategy / Ops                   | In Writing / Policy                     |
| ------------------------------------------------ | -------------------------------------- | ------------------------------------- | ----------------------------------- | --------------------------------------- |
| `CONSTITUTION.md`                                | engineering constraints                | study constraints                     | operating principles                | editorial or policy constraints         |
| `BRAINSTORM.md`                                  | codebase research                      | literature scan                       | landscape scan                      | source gathering and framing            |
| `SPEC.md`                                        | feature requirements                   | research question or hypothesis       | decision brief                      | argument or policy brief                |
| `PLAN.md`                                        | implementation plan                    | study design                          | rollout plan                        | outline and revision plan               |
| `TASKS.md`                                       | execution checklist                    | experiment tasks                      | workback schedule                   | drafting and review checklist           |
| `implement`                                      | coding and integration                 | running the study                     | executing the change                | drafting and editing                    |
| `reflect`                                        | verification and regression review     | results review                        | retro and validation                | revision review and critique            |
| `reconcile` / `resume` / `summarize` / `handoff` | reconcile, resume, or transfer context | reconcile or resume the investigation | reconcile or transfer project state | reconcile or transfer editorial context |

Feature artifacts use typed YAML front matter for canonical metadata such as
artifact identity, feature identity, relationships, dependencies, skills, and
summary/intent. Legacy markdown body sections remain readable during migration,
but newly generated docs write canonical metadata in front matter.

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
# initialize the project and user config
# creates/updates .kit.yaml and ~/.config/kit/.kit.yaml
# creates .coderabbit.yaml if missing
# copies a CONSTITUTION.md drafting prompt to your clipboard
kit init

# paste the copied prompt into your coding agent to build docs/CONSTITUTION.md

# later, after the repo has real contents, refresh durable project-level docs
kit prompt project refresh

# refresh Kit-managed scaffold docs/files to the current Kit defaults
kit init --refresh

# optionally capture research first
kit brainstorm my-feature

# for frontend-heavy work, opt into frontend-specific prompt guidance
kit brainstorm dashboard-redesign --profile=frontend

# capture out-of-scope follow-up work without changing the active lane
kit brainstorm --backlog shared-refactor

# review deferred backlog items
kit backlog

# write the spec
kit spec my-feature

# define the plan
kit plan my-feature

# break work into tasks
kit tasks my-feature

# start implementation
# runs the readiness gate, then outputs coding-agent context
kit implement my-feature

# inspect progress at any time
kit status

# inspect the full project overview
kit status --all

# reorient before resuming work
kit resume my-feature

# inspect a selected feature's document map and feature lineage
kit map

# inspect the full project document map
kit map --all

# inspect Kit command behavior before choosing a command
kit capabilities --search verify

# pause a feature without losing its phase
kit pause my-feature

# mark all eligible active features complete
kit complete --all

# remove a feature and all docs under its feature directory
# keeps notes by default and shows the feature as removed in history/status
kit rm my-feature --yes

# remove the feature notes too
kit rm my-feature --yes --notes
```

## 🧰 Commands

### 🏁 Setup

| Command        | Description                                                                       |
| -------------- | --------------------------------------------------------------------------------- |
| `kit init`     | Initialize project, user config, local env files, `.gitignore`, review config, and GitHub PR template |
| `kit scaffold` | Create empty workflow document structures, support directories, and agent files   |

### 🔁 Workflow

| Command                    | Description                                                                               |
| -------------------------- | ----------------------------------------------------------------------------------------- |
| `kit brainstorm [feature]` | Interactively create `BRAINSTORM.md` and output a research/documentation prompt          |
| `kit backlog`              | List deferred brainstorm items or use `--pickup` as the backlog-specific resume shortcut  |
| `kit spec <feature>`       | Create or open a feature specification, perform skills discovery, and track dependencies  |
| `kit plan <feature>`       | Create or open an implementation plan and track planning dependencies                     |
| `kit tasks <feature>`      | Create or open a task list                                                                |
| `kit loop [feature]`       | Run the remaining workflow through a configured confidence-gated local agent loop         |
| `kit resume [feature]`     | Resume backlog or in-flight work through the canonical prompt flow                        |
| `kit implement [feature]`  | Run the implementation readiness gate and output implementation context for coding agents |
| `kit reflect [feature]`    | Output reflection and verification instructions                                           |
| `kit pause [feature]`      | Pause an in-flight feature without changing its underlying phase                          |
| `kit complete [feature]`   | Mark a feature complete; supports `--all` for all eligible active features                |
| `kit rm [feature]`         | Remove feature docs, retain notes by default, and show removed state in history/status; `kit remove` also works |

### 🔎 Inspect & Repair

| Command                   | Description                                                                                      |
| ------------------------- | ------------------------------------------------------------------------------------------------ |
| `kit status`              | Show the active feature status, including paused state; supports `--json`                        |
| `kit status --all`        | Show the project-wide overview as a lifecycle matrix with state and task progress; supports JSON |
| `kit map [feature]`       | Select or show a feature map; supports `--all` for the full project document map                 |
| `kit capabilities`        | List command capabilities, mutation behavior, network use, and important flags; supports `--json`, `--full`, and `--search` |
| `kit check <feature>`     | Validate feature documents and populated required sections                                       |
| `kit check --project`     | Validate the repo-level document, init scaffold, and instruction contract                        |
| `kit verify [feature]`    | Run declared verification checks from `TASKS.md` and write local run evidence                    |
| `kit trace <target>`      | List feature verification runs or inspect one run ID                                             |
| `kit replay <run-id>`     | Rerun commands from a prior verification run and compare outcomes                                |
| `kit state [refresh]`     | Show or refresh generated pointer-only `.kit/state.json` for agents and tools                    |
| `kit eval`                | Run small local harness regression checks                                                        |
| `kit rules` / `kit rule`  | Import, preview, create, list, and link durable repo-local rulesets under `docs/references/rules/` |
| `kit reconcile [feature]` | Audit Kit-managed docs and init scaffold drift; supports `--migrate-verification` for advisory executable-check migration |

### 🧾 Prompt Utilities

| Command                        | Description                                                                                  |
| ------------------------------ | -------------------------------------------------------------------------------------------- |
| `kit prompt [noun] [verb]`     | Resolve and copy a reusable prompt from local, global, or built-in prompt libraries          |
| `kit prompt list`              | List effective merged prompts with origin and override metadata                              |
| `kit prompt project refresh`   | Prompt an agent to refresh durable project-level docs after the repo matures                 |
| `kit set prompt [noun] [verb]` | Create or update a local or global prompt through the editor                                 |
| `kit handoff [feature]`        | Prompt the current agent session to sync docs, reference inventories, and prepare a handoff |
| `kit summarize [feature]`      | Output context summarization instructions                                                    |
| `kit review-loop`              | Prepare a dispatch prompt from current unresolved PR review feedback                         |
| `kit dispatch`                 | Output a discovery-first prompt for clustering tasks and queueing subagents                  |
| `kit code-review`              | Output instructions for branch code review                                                   |
| `kit skill mine [feature]`     | Output skill extraction prompt for the active coding agent                                   |

### 🔧 Utilities

| Command          | Description                                 |
| ---------------- | ------------------------------------------- |
| `kit upgrade`    | Download and install the latest Kit release |
| `kit version`    | Print the installed Kit version             |
| `kit completion` | Generate shell autocompletion script        |

Hidden compatibility commands remain callable for migration, but they are no longer shown in
default help or primary docs: `kit update`, `kit skills`, `kit catchup`, `kit scaffold-agents`,
`kit rollup`, and `kit brainstorm --pickup`.

Prompt-producing commands default to subagent orchestration guidance. Pass
`--single-agent` when you explicitly want to keep the work in one lane.

Prompt-producing commands also support `--profile=frontend` for frontend-heavy
work. The profile keeps Kit's normal RLM flow, but adds frontend-specific
guidance for design-system fit, domain-appropriate UI, visual assets,
responsive behavior, browser or screenshot validation, interaction states, and
common generated-UI pitfalls. Feature-scoped commands can carry the profile
forward through the feature's front matter references once it has been recorded.

Use `kit dispatch` when you need the full overlap-clustering and queue-planning
workflow for a raw task set. Use the default prompt path when the agent should
use subagents opportunistically, and use `kit dispatch` when you want a formal
discovery report, overlap clustering, and explicit approval before launch.
Use `kit dispatch --pr <url|number>` to prefill the dispatch editor from
unresolved, non-outdated GitHub PR review threads. Add `--coderabbit` to keep
only CodeRabbit-authored review comments. `--pr` accepts a full GitHub PR URL, a
Markdown PR link, `owner/repo#123`, or a PR number resolved from the current
project's `origin` remote.
After the resulting fixes or no-op decisions are complete, use
`kit dispatch --pr <target> --resolve --yes` to resolve the currently matching
unresolved review threads on GitHub. Add `--coderabbit` to resolve only
CodeRabbit-authored review threads. Resolution is an explicit GitHub mutation
and is never part of the default prompt-generation path.

Use `kit review-loop --pr <url|number> --coderabbit` when you want Kit to turn
current unresolved CodeRabbit review feedback into a human-reviewed dispatch
prompt. Add `--watch` to wait for CodeRabbit completion on the current PR head
before collecting comments. `kit dispatch --loop --pr <target>` is an alias for
the same review-loop workflow, while `kit dispatch --pr <target> --coderabbit`
remains the lower-level untriaged review-thread intake.

### 📋 Output Behavior

Prompt-producing commands, including the constitution prompt emitted by
`kit init`, copy their generated output to the clipboard by default. Pass
`--output-only` to print the raw prompt or output to stdout instead, or combine
`--output-only --copy` to do both.

`kit prompt <noun> <verb>` follows the same raw-output flags, but its default
human-readable output also prints the selected prompt body in a delimited block
with command, origin, and override metadata. v0 does not support `--source`,
`--no-copy`, auto-paste, or clipboard restore.

In interactive terminals, Kit also uses clearer section spacing and semantic
emoji markers for help, status, selectors, and other human-readable guidance.
Status views may also use ANSI color in a real terminal to highlight lifecycle
markers, state labels, file presence, and progress without changing non-TTY
output.
Raw `--output-only` payloads and `--json` output avoid human-readable wrappers;
prompt-library `coding-agent` payloads intentionally begin with `---` so they
can be pasted directly as instruction blocks.

Lifecycle views surface paused and removed work explicitly. `kit status` keeps
the active feature in focus, `kit status --all` provides the project overview
as a fixed-width lifecycle matrix with retained notes visibility, and
deferred brainstorm items stay available through `kit backlog`,
`kit backlog --pickup`, or `kit resume`.

### ♻️ Prompt Regeneration

Feature-scoped prompt commands also accept `--prompt-only` to regenerate the
selected feature's prompt without mutating repository docs:
`kit brainstorm`, `kit spec`, `kit plan`, `kit tasks`, `kit implement`,
`kit reflect`, `kit reconcile`, `kit handoff`, and `kit skill mine`. For
`brainstorm`, `spec`, `plan`, and `tasks`, `--prompt-only` skips scaffolding
and rollup writes, requires the existing artifact set, and uses the normal
existing-feature selector when no feature argument is provided.

### 📚 Prompt Library

`kit prompt` resolves reusable prompts by explicit noun and verb:

```bash
# direct lookup
kit prompt coding-agent short

# interactive noun and verb selectors
kit prompt
kit prompt coding-agent

# discovery
kit prompt list

# create or update prompts through the editor
kit set prompt custom review
kit set prompt custom review --global
kit set prompt custom review --local --global
```

Prompt precedence is:

1. project-local `.kit.yaml`
2. global `~/.config/kit/.kit.yaml`
3. built-in Kit prompts

`kit init` creates or updates the global config with missing default fields
without replacing existing prompt entries.

Prompt entries use nested YAML object form:

```yaml
prompts:
  custom:
    review:
      content: |
        Review the current changes for correctness, edge cases, and tests.
      description: Custom review prompt
```

Nouns and verbs normalize to lowercase kebab-case. `kit set prompt` defaults to
local save inside a Kit project, asks before saving globally outside a project,
and confirms before overwriting each selected scope. Built-ins include
`coding-agent short`, `coding-agent long`, `coding-agent instructions`,
workflow prompts, support prompts, `skill mine`, and `project init`.

### ⛏️ Skill Mining

| Command                    | Description                                                |
| -------------------------- | ---------------------------------------------------------- |
| `kit skill mine [feature]` | Output skill extraction prompt for the active coding agent |

`kit scaffold agents` creates or refreshes repository instruction files.
`kit scaffold brainstorm`, `kit scaffold spec`, `kit scaffold plan`, and
`kit scaffold tasks` create empty workflow document scaffolds without emitting
agent prompts.

When instruction files already exist:

- default mode skips them and suggests safer next steps
- `--append-only` merges missing Kit-managed sections without overwriting matched existing content
- `--force` overwrites existing files after confirmation
- `--force --yes` overwrites existing files without prompting for automation use

Instruction scaffold versions:

- `--version 1` keeps the legacy verbose `AGENTS.md` / `CLAUDE.md` model
- `--version 2` uses thin entrypoints plus `docs/agents/` and `docs/references/` for repo-local ToC and RLM routing
- new repos default to `v2`
- existing repos keep their current model unless `--version` explicitly switches them
- switching models is a repo-wide change and requires `--force`

`kit init --refresh` is the consolidated refresh command for existing Kit
projects. It creates missing Kit-managed files, migrates known generated v1
instruction files to the v2 thin docs model, refreshes generated instruction
support docs, imports missing known registry rulesets, and records ruleset
registry state in `.kit.yaml`. Existing registry rulesets are adopted
automatically: safe upstream updates from the Kit GitHub `main` branch are
applied, local activation status is preserved, and customized or conflicted
rules are skipped with a report. Existing local scaffold files such as `.envrc`,
`.coderabbit.yaml`, and the pull request template are skipped by default. Use
`kit init --refresh --force` to overwrite refreshable generated documentation
and accept latest registry ruleset content while preserving local ruleset
status, or `kit init --refresh --file=.envrc --force` to overwrite one existing
Kit-managed file explicitly. Use `kit init --refresh --dry-run --diff` to
preview the managed-file changes without writing them.

## 🗂️ Structured Engine: Artifact Pipeline

The artifact pipeline is Kit's most structured operating engine. It is not the
entire product, but it is the clearest path when a problem needs deliberate
discovery, planning, and execution control in any domain.

### 🏗️ Project Initialization

Run once, then refine as the project matures:

- use `kit prompt project refresh` when early feature work reveals durable project-level rules, vocabulary, or constraints that should update `CONSTITUTION.md`

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
       ▲                                                                                                │
       └────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 📝 Artifact Details

1. 📜 **Constitution** — strategy, patterns, long-term vision
2. 🧠 **Brainstorm** — optional research, findings, relationships, dependencies, strategy
3. 📐 **Specification** — what is being built, why, and how it relates to prior features
4. 🗺️ **Plan** — how it will be built
5. ✅ **Tasks** — executable work units
6. 🛠️ **Implementation** — execution after the readiness gate passes
7. 🔍 **Reflection** — verification and learning after implementation

When a core workflow command runs without a feature argument, its selector only
shows features that are eligible for that command's next stage. Completed stages
are omitted from earlier-stage selectors.
If `kit spec` has no pre-spec features to list, it prompts for a new feature
name and starts the interactive spec builder.

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
- requires the coding agent to keep front matter `relationships` current with explicit prior-feature lineage, falling back to `## RELATIONSHIPS` only for legacy docs without front matter
- requires the coding agent to keep front matter `references` current with the inputs used during the brainstorm phase, including `target`, `relation`, and `read_policy`
- opens `$EDITOR` by default for the multiline thesis, falling back to a vim-compatible editor when `$EDITOR` is unset, with step instructions and a press-any-key launch gate
- supports `--inline` to use terminal multiline entry with `Shift+Enter` and `Ctrl+J`, including consecutive blank lines
- keeps `--vim` and `--editor=vim` as explicit controls when a vim-compatible editor is desired
- outputs a research/documentation prompt without native agent mode commands
- tells the coding agent to research the codebase, use numbered lists, ask questions in batches of up to 10, and avoid implementation
- requires the agent to include recommended defaults, accept `yes` / `y` for whole-batch approval and `yes 3, 4, 5` / `y 3, 4, 5` for numbered approval, state uncertainties, output percentage-understanding progress after each batch, and continue until the spec is precise enough for a production-quality solution
- supports `--backlog` to capture a deferred brainstorm item without outputting a research prompt
- keeps `--pickup` callable as a hidden compatibility path while teaching `kit resume <feature>` or `kit backlog --pickup <feature>` as the primary resume flows

### 💡 Why this matters

`BRAINSTORM.md` becomes the durable bridge between early ideation and the formal artifact pipeline. When present, downstream commands use it as research context while still treating `SPEC.md`, `PLAN.md`, and `TASKS.md` as the binding execution contract.

### 🪜 Typical flow

```text
You / team idea
  ↓
kit brainstorm my-feature
  ↓
BRAINSTORM.md + research/documentation prompt
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

### 🔂 Autonomous Loop

`kit loop [feature]` can run the remaining workflow after the user has started
with either `kit brainstorm <feature>` or `kit spec <feature>`. It resolves the
current strict stage, wraps that stage prompt with a required
`KIT_LOOP_RESULT` JSON contract, sends the prompt to the configured local agent
command over stdin, validates confidence and document state, then repeats until
the target stage is complete or a blocker appears.

```yaml
loop:
  min_confidence: 95
  max_iterations: 20
  agent:
    command: your-agent
    args: ["run", "--stdin"]
```

```bash
# see the next loop action without running an agent
kit loop my-feature --dry-run

# run until reflection is complete
kit loop my-feature

# stop after task generation is complete
kit loop my-feature --until tasks
```

Loop evidence is written under `.kit/loops/<run-id>/`. Existing workflow
commands remain manual prompt-producing commands; only `kit loop` invokes the
configured agent command.

`kit implement` begins with an implementation readiness gate that adversarially
challenges `CONSTITUTION.md`, optional `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`,
and `TASKS.md` before any code work starts. If the gate fails, update the
canonical docs first, then rerun the gate before implementing.

### ▶️ Usage

```bash
# interactive brainstorm for a new feature
kit brainstorm my-feature

# prepare notes and BRAINSTORM.md before starting the brainstorm prompt
kit brainstorm my-feature --prepare

# frontend-heavy brainstorm with design-material scaffolding
kit brainstorm dashboard-redesign --profile=frontend

# capture a deferred follow-up feature and leave it in backlog
kit brainstorm --backlog shared-refactor

# open or continue an existing brainstorm
kit brainstorm my-feature

# print the raw prompt to stdout instead of copying it
kit brainstorm my-feature --output-only

# regenerate the brainstorm prompt from an existing BRAINSTORM.md without touching repo docs
kit brainstorm my-feature --prompt-only

# opt out of default editor mode and use inline multiline entry
kit brainstorm my-feature --inline

# write the generated brainstorm prompt to a file
kit brainstorm my-feature --output tmp/brainstorm-prompt.md

# list deferred backlog items
kit backlog

# resume a deferred backlog item
kit backlog --pickup shared-refactor

# canonical general resume flow
kit resume shared-refactor
```

`kit spec <feature> --interactive` now opens `$EDITOR` for each free-text answer
by default, falling back to a vim-compatible editor when `$EDITOR` is unset. Use `kit spec <feature> --interactive --inline`
to opt back into terminal multiline entry.

### 📄 What goes in `BRAINSTORM.md`

- summary of the issue or opportunity
- user thesis in the user's own words
- explicit relationship to prior features or `none`
- codebase findings and relevant architecture notes
- affected files with concrete paths
- unresolved questions and viable options
- recommended strategy and the next workflow step

## 🏛️ Project Structure

```text
.kit.yaml                    # configuration and local prompt overrides
docs/
  CONSTITUTION.md            # project-wide constraints
  PROJECT_PROGRESS_SUMMARY.md
  notes/
    0001-my-feature/
      .gitkeep
      design/                # frontend materials when --profile=frontend is used
        .gitkeep
        screenshots/
          .gitkeep
        references/
          .gitkeep
  specs/
    0001-my-feature/
      BRAINSTORM.md         # optional research artifact
      SPEC.md
      PLAN.md
      TASKS.md
      ANALYSIS.md            # optional
  references/
    rules/
      frontend-ui.md          # optional durable pointer-loaded rulesets
```

New `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` files include
front matter. `kit map`, `kit status`, `kit rollup`, `kit check`, and
prompt-producing commands read front matter first and fall back to legacy body
metadata when front matter is absent.

## ✨ Positioning

Kit is broader than a spec generator. It is a harness for disciplined
development workflows:

- structured planning when scope is unclear
- lightweight ad hoc execution when the change is contained
- recovery tools such as `reconcile`, `resume`, `summarize`, and `handoff`
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
