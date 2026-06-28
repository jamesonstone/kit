# Kit Commands

This guide covers installation, command groups, prompt behavior, prompt
libraries, scaffold refresh, and common command paths.

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

To enable repository-managed Git hooks:

```bash
make install-git-hooks
```

This configures `core.hooksPath` to use `.githooks/`, including a `pre-commit`
hook that runs `make build` before every commit.

## Quick Start

```bash
kit init
kit project refresh
kit init --refresh
kit spec my-feature
kit spec dashboard-redesign --profile=frontend
kit status --all
kit resume my-feature
kit map --all
kit capabilities --search verify
kit pause my-feature
kit complete --all
kit rm my-feature --yes
kit rm my-feature --yes --notes
```

## Setup

| Command | Description |
| --- | --- |
| `kit init` | Initialize project, user config, local env files, `.gitignore`, review config, GitHub PR template, and optional auto-assignment workflow. |
| `kit scaffold` | Create empty workflow document structures, support directories, and agent files. |

## Workflow

| Command | Description |
| --- | --- |
| `kit backlog` | List deferred brainstorm items or use `--pickup` as the backlog-specific resume shortcut. |
| `kit spec <feature>` | Start or resume the v2 `SPEC.md` workflow and output the supervisor prompt. |
| `kit legacy` | List deprecated legacy v1 staged workflow commands retained for migration. |
| `kit loop [feature]` | Convenience alias for `kit loop workflow [feature]`. |
| `kit loop workflow [feature]` | Run the remaining workflow through a configured confidence-gated local agent loop. |
| `kit loop review [feature]` | Run a coding-agent correctness review loop over changed code. |
| `kit resume [feature]` | Resume backlog or in-flight work through the canonical prompt flow. |
| `kit pause [feature]` | Pause an in-flight feature without changing its underlying phase. |
| `kit complete [feature]` | Mark a feature complete; supports `--all`. |
| `kit project refresh` | Generate or record a semantic refresh of durable project-level docs and Constitution cadence state. |
| `kit rm [feature]` | Remove feature docs, retain notes by default, and show removed state in history/status. `kit remove` also works. |

Run `kit legacy --help` to list v1 staged commands retained for finishing
existing `BRAINSTORM.md`, `PLAN.md`, or `TASKS.md` work.

## Inspect And Repair

| Command | Description |
| --- | --- |
| `kit status` | Show active feature status, including paused state; supports `--json`. |
| `kit status --all` | Show the project-wide lifecycle matrix with state and task progress. |
| `kit map [feature]` | Select or show a feature map; supports `--all` for the full project document map. |
| `kit capabilities` | List command capabilities, mutation behavior, network use, and important flags. |
| `kit check <feature>` | Validate feature documents and required populated sections. |
| `kit check --project` | Validate repo-level docs, init scaffold, and instruction contract. |
| `kit pr fix` | Select or target an open PR and run the review feedback repair loop. |
| `kit trace <target>` | List feature verification runs or inspect one run ID. |
| `kit replay <run-id>` | Rerun commands from a prior verification run and compare outcomes. |
| `kit state [refresh]` | Show or refresh generated pointer-only `.kit/state.json`. |
| `kit eval` | Run small local harness regression checks. |
| `kit rules` / `kit rule` | Import, preview, create, list, and link repo-local rulesets. |
| `kit reconcile [feature]` | Audit Kit-managed docs and init scaffold drift. |

Inside the Kit source repository, every new command, subcommand, flag, alias,
or command behavior extension must update `kit capabilities` in the same
change. Downstream Kit-managed projects should use `kit capabilities` for
discovery, not maintain Kit's internal command catalog.

## Prompt Utilities

| Command | Description |
| --- | --- |
| `kit prompt [noun] [verb]` | Resolve and copy a reusable prompt from local, global, or built-in prompt libraries. |
| `kit prompt list` | List effective merged prompts with origin and override metadata. |
| `kit prompt project refresh` | Render the reusable prompt-library version of the project refresh prompt. |
| `kit set prompt [noun] [verb]` | Create or update a local or global prompt through the editor. |
| `kit handoff [feature]` | Prompt the current agent session to sync docs and prepare a handoff. |
| `kit summarize [feature]` | Output context summarization instructions. |
| `kit dispatch` | Output a discovery-first prompt for clustering tasks and queueing subagents. |
| `kit code-review` | Output instructions for branch code review. |
| `kit skill mine [feature]` | Output skill extraction prompt for the active coding agent. |

## Utilities

| Command | Description |
| --- | --- |
| `kit upgrade` | Download and install the latest Kit release. |
| `kit version` | Print the installed Kit version. |
| `kit completion` | Generate shell autocompletion scripts. |

## Prompt Profiles And Subagents

Prompt-producing commands default to subagent orchestration guidance. Pass
`--single-agent` when you explicitly want to keep work in one lane.

Prompt-producing commands also support `--profile=frontend` for frontend-heavy
work. The profile keeps Kit's normal RLM flow while adding frontend-specific
guidance for design-system fit, domain-appropriate UI, visual assets,
responsive behavior, browser or screenshot validation, interaction states, and
common generated-UI pitfalls.

## Dispatch And Review Loops

Use `kit dispatch` when you need formal overlap clustering and queue planning
for a raw task set. Use the default prompt path when an agent should use
subagents opportunistically.

Use `kit pr fix` as the default PR review repair entrypoint. With no flags it
lists open pull requests in the current repository and asks which one to repair.
Use `kit pr fix --pr <url|owner/repo#number|number>` to target a specific PR.
The command wraps the `kit loop review --pr` repair path, so it may write
`.kit/loops` evidence and the configured agent may edit local files, but it
does not stage, commit, push, post PR comments, or perform GitHub delivery.
After fixes or no-op decisions are validated, `kit pr fix` asks the delegated
agent to resolve all matching current unresolved review threads, including
human reviewer and CodeRabbit feedback, through
`kit dispatch --pr <target> --resolve --yes`.

Use `kit dispatch --pr <url|number>` to prefill the dispatch editor from
unresolved, non-outdated GitHub PR review threads. Add `--coderabbit` to keep
only CodeRabbit-authored review comments.

After fixes or no-op decisions are complete, use
`kit dispatch --pr <target> --resolve --yes` to resolve matching unresolved
review threads on GitHub. Resolution is an explicit GitHub mutation and is
not part of raw dispatch prompt generation.

Use `kit loop review` when changed code should be reviewed until the local
agent reports at least 95% correctness and no high, medium, or
correctness-impacting issues remain. Without `--pr`, it reviews current-branch
changes relative to `origin/main`, falling back to `main`, plus staged and
unstaged changes.

Use `kit loop review --pr <target>` to fold current CodeRabbit feedback into
that repair loop. Add `--watch` or `--wait-for-coderabbit` only when you want
to wait up to the timeout.

## Output Behavior

Prompt-producing commands, including the constitution prompt emitted by
`kit init`, copy generated output to the clipboard by default.

Use:

- `--output-only` to print the raw prompt or output to stdout
- `--output-only --copy` to print and copy
- `--prompt-only` on feature-scoped prompt commands to regenerate prompts
  without mutating repository docs

Human-readable terminal output uses semantic emoji markers, spacing, and ANSI
color when appropriate. Raw `--output-only` payloads and `--json` output avoid
human-readable wrappers.

## Prompt Library

`kit prompt` resolves reusable prompts by explicit noun and verb:

```bash
kit prompt coding-agent short
kit prompt
kit prompt coding-agent
kit prompt list
kit set prompt custom review
kit set prompt custom review --global
kit set prompt custom review --local --global
```

Prompt precedence is:

1. project-local `.kit.yaml`
2. global `~/.config/kit/.kit.yaml`
3. built-in Kit prompts

Prompt entries use nested YAML object form:

```yaml
prompts:
  custom:
    review:
      content: |
        Review the current changes for correctness, edge cases, and tests.
      description: Custom review prompt
```

Nouns and verbs normalize to lowercase kebab-case. Built-ins include
`coding-agent short`, `coding-agent long`, `coding-agent instructions`, the v2
`kit spec` / `workflow spec` supervisor prompt, support prompts, `skill mine`,
and `project init`.

Use `kit spec <feature>` when Kit should create or adopt `SPEC.md`. Use
`kit prompt kit spec` to render the active feature's reusable prompt-library
entry.

## Scaffold And Refresh

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
- `--version 2` uses thin entrypoints plus `docs/agents/` and `docs/references/`
- new repos default to `v2`
- existing repos keep their current model unless `--version` explicitly switches them
- switching models is repo-wide and requires `--force`

`kit init --refresh` is the consolidated refresh command for existing Kit
projects. It creates missing Kit-managed files, migrates known generated v1
instruction files to the v2 thin docs model, refreshes generated support docs,
imports known registry rulesets, and records ruleset registry state in
`.kit.yaml`. It also creates or refreshes the Kit-managed
`.github/workflows/auto-assign.yml` workflow. That workflow assigns new issues
and pull requests to `github.default_assignees` from the project `.kit.yaml`,
falls back to the global `~/.config/kit/.kit.yaml`, and safely no-ops when no
assignees are configured.

Use `kit init --refresh --dry-run --diff` to preview managed-file changes
without writing them.
