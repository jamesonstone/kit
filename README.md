```text
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║
██╔═██╗ ██║   ██║
██║  ██╗██║   ██║
╚═╝  ╚═╝╚═╝   ╚═╝
```

**Kit v2 Thought-Work Harness**

🎒 Portable harness without vendor lock-in.

Kit v2 is a general-purpose harness for disciplined thought work.
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

The v2 artifact model is broader than software:

| Kit Concept                                      | In Software                            | In Research                           | In Strategy / Ops                   | In Writing / Policy                     |
| ------------------------------------------------ | -------------------------------------- | ------------------------------------- | ----------------------------------- | --------------------------------------- |
| `CONSTITUTION.md`                                | engineering constraints                | study constraints                     | operating principles                | editorial or policy constraints         |
| `SPEC.md`                                        | feature workflow artifact              | research question, study plan, proof  | decision brief, rollout, evidence   | argument, outline, revision evidence    |
| `SPEC.md` acceptance criteria                    | binary behavior checks                 | falsifiable success criteria          | decision or rollout gates           | editorial acceptance standards          |
| `SPEC.md` validation map and evidence            | tests, runtime checks, docs review     | result evidence and audit trail       | operational validation              | source/proof trail and revision notes   |
| legacy v1 `BRAINSTORM.md` / `PLAN.md` / `TASKS.md` | historical staged artifacts            | historical staged artifacts           | historical staged artifacts         | historical staged artifacts             |
| `reconcile` / `resume` / `summarize` / `handoff` | reconcile, resume, or transfer context | reconcile or resume the investigation | reconcile or transfer project state | reconcile or transfer editorial context |

Feature artifacts use typed YAML front matter for canonical metadata such as
artifact identity, feature identity, relationships, dependencies, skills, and
summary/intent. New v2 feature work uses `SPEC.md` as the single durable
feature artifact; deprecated legacy v1 staged artifact files remain readable during migration.

The names may be software-flavored today, but the structure is general:
constraints, clarification, planning, execution, verification, reflection,
delivery, evidence, and transfer are common to most serious thought work.

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

# start the v2 single-SPEC feature workflow
kit spec my-feature

# for frontend-heavy work, opt into frontend-specific prompt guidance
kit spec dashboard-redesign --profile=frontend

# capture out-of-scope follow-up work without changing the active lane
kit legacy brainstorm --backlog shared-refactor

# review deferred backlog items
kit backlog

# paste the generated v2 supervisor prompt into your coding agent
# clarification, planning, tasks, implementation, validation,
# reflection, documentation updates, and delivery gating stay in SPEC.md

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
| `kit backlog`              | List deferred brainstorm items or use `--pickup` as the backlog-specific resume shortcut  |
| `kit spec <feature>`       | Run the v2 single-SPEC feature workflow and output the supervisor prompt                  |
| `kit legacy`               | List deprecated legacy v1 staged workflow commands retained for migration                 |
| `kit loop [feature]`       | Convenience alias for `kit loop workflow [feature]`                                      |
| `kit loop workflow [feature]` | Run the remaining workflow through a configured confidence-gated local agent loop      |
| `kit loop review [feature]` | Run a coding-agent correctness review loop over changed code                            |
| `kit resume [feature]`     | Resume backlog or in-flight work through the canonical prompt flow                        |
| `kit pause [feature]`      | Pause an in-flight feature without changing its underlying phase                          |
| `kit complete [feature]`   | Mark a feature complete; supports `--all` for all eligible active features                |
| `kit rm [feature]`         | Remove feature docs, retain notes by default, and show removed state in history/status; `kit remove` also works |

Run `kit legacy --help` to list the deprecated legacy v1 staged commands retained for
finishing existing `BRAINSTORM.md` / `PLAN.md` / `TASKS.md` work.

### 🔎 Inspect & Repair

| Command                   | Description                                                                                      |
| ------------------------- | ------------------------------------------------------------------------------------------------ |
| `kit status`              | Show the active feature status, including paused state; supports `--json`                        |
| `kit status --all`        | Show the project-wide overview as a lifecycle matrix with state and task progress; supports JSON |
| `kit map [feature]`       | Select or show a feature map; supports `--all` for the full project document map                 |
| `kit capabilities`        | List command capabilities, mutation behavior, network use, and important flags; supports `--json`, `--full`, and `--search` |
| `kit check <feature>`     | Validate feature documents and populated required sections                                       |
| `kit check --project`     | Validate the repo-level document, init scaffold, and instruction contract                        |
| `kit trace <target>`      | List feature verification runs or inspect one run ID                                             |
| `kit replay <run-id>`     | Rerun commands from a prior verification run and compare outcomes                                |
| `kit state [refresh]`     | Show or refresh generated pointer-only `.kit/state.json` for agents and tools                    |
| `kit eval`                | Run small local harness regression checks                                                        |
| `kit rules` / `kit rule`  | Import, preview, create, list, and link durable repo-local rulesets under `docs/references/rules/` |
| `kit reconcile [feature]` | Audit Kit-managed docs and init scaffold drift; supports `--migrate-verification` for advisory executable-check migration |

Inside the Kit source repository, every new command, subcommand, flag, alias, or
command behavior extension must update `kit capabilities` in the same change.
Downstream Kit-managed projects should use `kit capabilities` for command
discovery; they should not maintain Kit's internal command catalog.

### 🧾 Prompt Utilities

| Command                        | Description                                                                                  |
| ------------------------------ | -------------------------------------------------------------------------------------------- |
| `kit prompt [noun] [verb]`     | Resolve and copy a reusable prompt from local, global, or built-in prompt libraries          |
| `kit prompt list`              | List effective merged prompts with origin and override metadata                              |
| `kit prompt project refresh`   | Prompt an agent to refresh durable project-level docs after the repo matures                 |
| `kit set prompt [noun] [verb]` | Create or update a local or global prompt through the editor                                 |
| `kit handoff [feature]`        | Prompt the current agent session to sync docs, reference inventories, and prepare a handoff |
| `kit summarize [feature]`      | Output context summarization instructions                                                    |
| `kit dispatch`                 | Output a discovery-first prompt for clustering tasks and queueing subagents                  |
| `kit code-review`              | Output instructions for branch code review                                                   |
| `kit skill mine [feature]`     | Output skill extraction prompt for the active coding agent                                   |

### 🔧 Utilities

| Command          | Description                                 |
| ---------------- | ------------------------------------------- |
| `kit upgrade`    | Download and install the latest Kit release |
| `kit version`    | Print the installed Kit version             |
| `kit completion` | Generate shell autocompletion script        |

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

Use `kit loop review` when you want Kit to run a configured coding agent over
changes not in the remote mainline until the agent reports at least 95%
correctness and no high, medium, or correctness-impacting issues remain.
Without `--pr`, it reviews the current branch relative to `origin/main`
(falling back to `main`) plus staged and unstaged changes.

Use `kit loop review --pr <target>` to fold current CodeRabbit feedback into
that repair loop. PR mode starts local review immediately, checks CodeRabbit
opportunistically during the loop, and does one quick feedback check before
finalizing. If CodeRabbit is still pending after local review is done, Kit exits
with a provisional status and a rerun command instead of hanging. Add `--watch`
or `--wait-for-coderabbit` when you explicitly want to wait up to the timeout.
Review prompts use one agent by default. Pass `--subagents` to let the parent
agent pre-analyze the diff, choose useful review lanes, and use subagents only
when the split is clear. In an interactive terminal, Kit asks before rerunning
when prior loop-review evidence exists or the current run reaches max
iterations. Human-readable runs stream emoji-marked runner progress and
child-agent stdout/stderr to stderr; `--json` keeps stdout machine-readable and
suppresses progress output. Agent command setup failures stop immediately with
stderr context.

Use `kit dispatch --loop --pr <target>` only when you want a human-reviewed
dispatch prompt from current unresolved PR review feedback instead of an agent
repair loop.

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
selected feature's prompt without mutating repository docs. In the v2 feature
workflow, `kit spec --prompt-only` emits the supervisor prompt without adding
v2 metadata, missing sections, notes directories, or rollup updates. Legacy
staged prompt commands such as `kit legacy brainstorm`, `kit legacy plan`,
`kit legacy tasks`, `kit legacy implement`, and `kit legacy reflect` retain
the same inspection-safe behavior for existing v1 artifact-stage flows.

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
the v2 `kit spec` / `workflow spec` supervisor prompt, support prompts,
`skill mine`, and `project init`. Use `kit spec <feature>` when Kit should
create or adopt `SPEC.md`; use `kit prompt kit spec` to render the active
feature's reusable prompt-library entry.

### ⛏️ Skill Mining

| Command                    | Description                                                |
| -------------------------- | ---------------------------------------------------------- |
| `kit skill mine [feature]` | Output skill extraction prompt for the active coding agent |

`kit scaffold agents` creates or refreshes repository instruction files.
`kit scaffold spec <feature>` creates or additively adopts the v2 `SPEC.md`
scaffold plus notes/reference-material directories without emitting an agent
prompt. Legacy staged document scaffolds are available only through
`kit legacy` commands.

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
support docs, installs or upgrades known generated default `codex exec` loop
agent config when `.kit.yaml` is missing or still has the generated
`loop.agent` value, imports missing known registry rulesets, and records
ruleset registry state in `.kit.yaml`. Existing registry
rulesets are adopted
automatically: safe upstream updates from the Kit GitHub `main` branch are
applied, local activation status is preserved, and customized or conflicted
rules are skipped with a report. Existing local scaffold files such as `.envrc`,
`.coderabbit.yaml`, and the pull request template are skipped by default. Use
`kit init --refresh --force` to overwrite refreshable generated documentation
and accept latest registry ruleset content while preserving local ruleset
status, or `kit init --refresh --file=.envrc --force` to overwrite one existing
Kit-managed file explicitly. Use `kit init --refresh --dry-run --diff` to
preview the managed-file changes without writing them.

## 🗂️ Structured Engine: V2 Single-SPEC Workflow

The v2 single-`SPEC.md` workflow is Kit's most structured operating engine. It
is not the entire product, but it is the clearest path when a problem needs
deliberate clarification, planning, implementation, validation, reflection,
documentation sync, delivery gating, and evidence in any domain.

### 🏗️ Project Initialization

Run once, then refine as the project matures:

- use `kit prompt project refresh` when early feature work reveals durable project-level rules, vocabulary, or constraints that should update `CONSTITUTION.md`

```text
┌──────────────┐
│ Constitution │  ← global constraints, principles, priors
└──────────────┘
```

### 🧠 Optional Research Material

```text
┌──────────────┐
│ Notes/Inputs │  ← reference materials, screenshots, research, constraints
└──────────────┘
```

### 🔁 V2 Development Loop

```text
┌──────────────┐    ┌───────────────────────────────────────────────────────────────────────────────┐
│ Idea / Input │ ─▶ │ kit spec <feature>                                                           │
└──────────────┘    │ SPEC.md: clarify → ready → implement → validate → reflect → deliver/complete │
                    └───────────────────────────────────────────────────────────────────────────────┘
```

### 📝 Artifact Details

1. 📜 **Constitution** — strategy, patterns, long-term vision
2. 📐 **SPEC.md** — the v2 durable feature artifact: thesis, context, clarifications, requirements, assumptions, acceptance criteria, implementation plan, task checklist, validation map, reflection notes, documentation updates, delivery decision, and evidence
3. 🧠 **Legacy staged artifacts** — existing `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` remain historical context for upgraded projects and are used by legacy staged commands

When a core workflow command runs without a feature argument, its selector only
shows features that are eligible for that command's next stage. Completed stages
are omitted from earlier-stage selectors.
If `kit spec` has no eligible existing feature candidates to list, it prompts
for a new feature name and starts the interactive spec builder.

The v2 `kit spec` prompt treats `SPEC.md` as the single durable feature
artifact and requires every acceptance criterion to map to validation evidence.
Legacy staged prompts still populate their legacy artifacts when those commands
are used directly.

## 🧱 Foundations: V1 Staged Workflow

Kit v2 was built from the original staged workflow:

```text
brainstorm -> specification -> plan -> tasks -> implement -> reflection
```

```text
┌─────────────┐   ┌───────────────┐   ┌─────────────┐   ┌─────────────┐   ┌──────────────┐
│ BRAINSTORM  │ → │ SPECIFICATION │ → │    PLAN     │ → │    TASKS    │ → │  REFLECTION  │
└─────────────┘   └───────────────┘   └─────────────┘   └─────────────┘   └──────────────┘
      idea             clarified           approach          checklist          review
```

That foundation still matters: v1 made ambiguity explicit, forced planning
before execution, kept task progress durable, and closed the loop with review.
Kit v2 keeps those semantics but removes the user-facing command sequence.
`kit spec <feature>` now creates one durable `SPEC.md` whose phases carry the
same work: clarify, ready, implement, validate, reflect, deliver, complete.

Historical `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files remain readable and
non-disruptive in upgraded projects. Their commands live under `kit legacy` for
finishing old staged work; they are no longer the default feature workflow.

### 🪜 Typical flow

```text
You / team idea
  ↓
kit spec my-feature
  ↓
SPEC.md + v2 supervisor prompt
  ↓
clarify → ready → implement → validate → reflect → deliver/complete
```

### 🔂 Autonomous Loop

`kit loop workflow [feature]` is the execution wrapper for prompt-level
workflow automation. In v2, the user-facing entry point remains
`kit spec <feature>` and the durable state remains `SPEC.md`; direct execution
stays behind explicit loop/run behavior. The `kit loop [feature]` root form is
a convenience alias for the same v2 workflow loop.

```yaml
loop:
  min_confidence: 95
  max_iterations: 10
  agent:
    command: codex
    args: ["--ask-for-approval", "never", "exec", "--model", "gpt-5.5", "--sandbox", "workspace-write", "--ignore-user-config", "--color", "never", "-"]
```

```bash
# see the next loop action without running an agent
kit loop workflow my-feature --dry-run

# run until the configured workflow target is complete
kit loop workflow my-feature

# stop after a v2 phase target when needed
kit loop workflow my-feature --until validate

# review changed code until local correctness converges
kit loop review

# review changed code and opportunistically ingest CodeRabbit feedback
kit loop review --pr 14
```

Loop evidence is written under `.kit/loops/<run-id>/`. `kit spec` remains
prompt-producing by default; only explicit loop/run behavior invokes the
configured agent command.

The v2 supervisor prompt performs the readiness gate inside `SPEC.md` before
implementation begins. It requires clarified assumptions, binary-verifiable
acceptance criteria, a task checklist, a validation map, documentation sync,
reflection notes, and evidence before delivery.

### ▶️ Usage

```bash
# start the v2 single-SPEC workflow
kit spec my-feature

# capture initial context interactively before generating the v2 supervisor prompt
kit spec my-feature --interactive

# regenerate the v2 supervisor prompt without v2 adoption writes
kit spec my-feature --prompt-only

# legacy staged workflow: prepare notes and BRAINSTORM.md
kit legacy brainstorm my-feature --prepare

# frontend-heavy v2 workflow with design-material scaffolding
kit spec dashboard-redesign --profile=frontend

# capture a deferred follow-up feature and leave it in backlog
kit legacy brainstorm --backlog shared-refactor

# print the raw prompt to stdout instead of copying it
kit spec my-feature --output-only

# list legacy staged commands retained for existing v1 work
kit legacy --help

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

### 🧭 Legacy Staged Commands

Use `kit legacy <command>` only when finishing existing v1 staged work or
capturing backlog research that intentionally lives outside the active v2 lane.
Those commands preserve `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` behavior for
migration, but new feature work should start with `kit spec <feature>`.

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
      SPEC.md               # v2 durable feature workflow artifact
      BRAINSTORM.md         # optional legacy staged research artifact
      PLAN.md               # optional legacy staged plan artifact
      TASKS.md              # optional legacy staged task artifact
      ANALYSIS.md            # optional
  references/
    rules/
      frontend-ui.md          # optional durable pointer-loaded rulesets
```

New v2 `SPEC.md` files include front matter with `workflow_version: 2` and a
workflow `phase`. Legacy staged artifacts still include front matter when
created, and `kit map`, `kit status`, `kit check`, and prompt-producing
commands read front matter first and fall back to legacy body metadata when
front matter is absent.

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
