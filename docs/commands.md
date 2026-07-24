# Kit Commands

This guide covers installation, command groups, prompt behavior, prompt
libraries, scaffold refresh, and common command paths.

## Installation

```bash
go install github.com/jamesonstone/kit/cmd/kit@latest
GOBIN="$HOME/.local/bin" go install github.com/jamesonstone/kit/cmd/git-wt@latest
```

Or build from source:

```bash
git clone https://github.com/jamesonstone/kit.git
cd kit
make build
```

`make build` builds `bin/kit` and `bin/git-wt`, then installs or updates
`~/.local/bin/git-wt`.

To enable repository-managed Git hooks:

```bash
make install-git-hooks
```

This configures `core.hooksPath` to use `.githooks/`, including a `pre-commit`
hook that runs `make build` before every commit.

The separately installed `git-wt` executable is an optional manual convenience
discovered by Git as `git wt`. Kit-managed rules and reconciled guidance use
native `git worktree` commands and do not depend on this wrapper. For the
portable workflow and optional command cheat sheet, see
[references/worktrees.md](references/worktrees.md). Use `git wt help` for
command discovery; Git reserves
`git <command> --help` for installed manual pages. Writable `issue`, `add`, and
`repair` lanes link the clone's primary checkout `.env` by default; append
`--no-link-env` when isolation is required. Detached `pr` lanes and migration do
not create environment links. `git wt cd GH-123` opens a child shell in an
exact registered lane for manual testing. To change the current shell's
directory, use `cd "$(git wt path GH-123)"`.

## Quick Start

```bash
kit init
kit project refresh
kit reconcile
kit spec my-feature
kit notes my-feature --add --source slack
kit spec dashboard-redesign --profile=frontend
kit status --all
kit registry status
kit health --dry-run --diff
kit resume my-feature
kit map --all
kit instructions
kit instructions --version=v3
kit instructions --version=v2
kit instructions --version=v1
kit capabilities --search verify
kit config check
kit aws verify
kit pause my-feature
kit complete --all
kit rm my-feature --yes
kit rm my-feature --yes --notes
```

## Setup

| Command | Description |
| --- | --- |
| `kit init` | Initialize project, user config, local env files, `.gitignore`, a project-owned Makefile starter, review config, GitHub PR template, managed README badges and Maintainers section, and optional auto-assignment workflow. |
| `kit scaffold` | Create empty workflow document structures, support directories, and agent files. |

## Workflow

| Command | Description |
| --- | --- |
| `kit backlog` | List deferred brainstorm items or use `--pickup` as the backlog-specific resume shortcut. |
| `kit spec <feature>` | Non-interactively scaffold, adopt, or orient a living specification for native agent planning, then remind the agent to check Kit-managed updates with `kit status`. |
| `kit notes [feature]` | Select, create, or add source-material notes under `docs/notes/<feature>`, including gitignored private conversation notes. |
| `kit legacy` | List deprecated legacy v1 staged workflow commands retained for migration. |
| `kit loop [feature]` | Deprecated bare V2 workflow-loop compatibility path. |
| `kit loop workflow [feature]` | Deprecated V2 workflow loop; rejects V3 specs with native-planning guidance. |
| `kit loop review [feature]` | Run a coding-agent correctness review loop over changed code. |
| `kit resume [feature]` | Resume backlog or in-flight work through the canonical prompt flow. |
| `kit pause [feature]` | Pause an in-flight feature without changing its underlying phase. |
| `kit complete [feature]` | Mark a feature complete and remind the agent to check Kit-managed updates with `kit status` before final delivery; supports `--all`. |
| `kit project refresh` | Generate or record a semantic refresh of durable project-level docs and Constitution cadence state. |
| `kit rm [feature]` | Remove feature docs, retain notes by default, and show removed state in history/status. `kit remove` also works. |

Run `kit legacy --help` to list v1 staged commands retained for finishing
existing `BRAINSTORM.md`, `PLAN.md`, or `TASKS.md` work.

## Inspect And Repair

| Command | Description |
| --- | --- |
| `kit status` | Show active feature status, local Kit-managed refresh state, and project refresh status; supports `--json`. |
| `kit status --all` | Show the project-wide lifecycle matrix plus local Kit-managed refresh state. |
| `kit registry status` | Show compact registry and Kit-managed file freshness for scheduled maintenance; supports `--json` and does not write files. |
| `kit health` | Apply all conflict-free Kit-managed rules, instructions, configuration, README, workflow, and scaffold updates, then run the project contract check. Use `--dry-run --diff` for a read-only preview or `--json` for automation. |
| `kit map [feature]` | Select or show a feature map; supports `--all` for the full project document map. |
| `kit capabilities` | List command capabilities, mutation behavior, network use, and important flags. |
| `kit config check` | Validate schema-versioned `.kit.yaml`; interactive terminals can add safe missing fields, while `--json` is read-only. |
| `kit aws verify` | Call STS and verify that the configured AWS profile resolves to the account configured in `.kit.yaml`. |
| `kit check <feature>` | Validate feature documents and required populated sections. |
| `kit check --project` | Validate repo-level docs, init scaffold, and instruction contract. |
| `kit pr fix` | Select or target an open PR and copy a dispatch prompt from review feedback; editing is opt-in. |
| `kit trace <target>` | List feature verification runs or inspect one run ID. |
| `kit replay <run-id>` | Rerun commands from a prior verification run and compare outcomes. |
| `kit state [refresh]` | Show or refresh generated pointer-only `.kit/state.json`. |
| `kit eval` | Run small local harness regression checks. |
| `kit improve run` | Run deterministic fixture suites. `default` is capability smoke coverage; `prompt-system` renders representative prompts three times and supports `--kit-binary` for identical-definition comparisons. |
| `kit rules` / `kit rule` | Import, preview, create, list, and link repo-local rulesets. |
| `kit reconcile [feature]` | Audit Kit-managed docs and init scaffold drift. Without a feature argument, the interactive menu asks whether to include files, force changes, and output the coding-agent prompt. Use `--include-files --dry-run --diff` to preview managed-file updates. |

Inside the Kit source repository, every new command, subcommand, flag, alias,
or command behavior extension must update `kit capabilities` in the same
change. Downstream Kit-managed projects should use `kit capabilities` for
discovery, not maintain Kit's internal command catalog.

### Project Configuration And AWS Context

Project `.kit.yaml` files carry a top-level integer `schema_version`. Kit performs a local schema and semantic inspection before project-aware commands. The current, complete fast path reads only `.kit.yaml`; it does not run AWS, Git, GitHub, or network subprocesses and does not write files.

Scheduled Kit health maintenance is enabled by default. Omitted, null, or empty
health configuration remains managed; only an explicit `false` opts a project
out:

```yaml
health:
  managed: false
```

`kit registry status` and `kit health` return a successful `disabled` result
without registry access or file writes for an opted-out project. Otherwise,
`kit health` applies safe non-force refreshes and reports local customizations or
conflicts for semantic curation and pull-request review. The command never
stages, commits, pushes, opens a pull request, or changes arbitrary product code.

When an interactive command finds a compatible missing or older schema, it offers to update the file inline. `kit config check` exposes the same validation explicitly, and `kit config check --json` reports state without prompting or writing.

AWS context is optional:

```yaml
schema_version: 1
aws:
  profile: acme-development
  account_id: "123456789012"
```

If AWS context is absent and AWS CLI profiles are available, interactive remediation offers to associate one with the project. A single profile uses a default-yes `Y/n` prompt; multiple profiles require an explicit numbered selection. Kit verifies the selected profile with STS before writing the profile and account ID together. Choosing not to use the only profile records `aws.enabled: false`, preventing repeated discovery prompts for projects that do not use AWS.

Run `kit aws verify` before the first AWS-dependent command in a task and immediately before AWS mutations. It uses the configured project profile, rejects a conflicting `AWS_PROFILE`, and fails when the resolved account differs from `aws.account_id`. Kit never runs `aws sso login` or chooses among multiple profiles automatically.

## Prompt Utilities

| Command | Description |
| --- | --- |
| `kit instructions` | Print the current provider-neutral coding-agent instructions as raw Markdown; use `--version=vN` to select an earlier immutable version. |
| `kit prompt [noun] [verb]` | Resolve and copy a reusable prompt from local, global, or built-in prompt libraries. |
| `kit prompt list` | List effective merged prompts with origin and override metadata. |
| `kit prompt project refresh` | Render the reusable prompt-library version of the project refresh prompt. |
| `kit set prompt [noun] [verb]` | Create or update a local or global prompt through the editor. |
| `kit plan challenge` | Read a copied Codex for Mac `/plan`, supplement it with a material adversarial-review contract, and copy the paste-ready prompt back without launching or calling a model. Use `--output-only` to inspect the raw prompt without replacing the clipboard. |
| `kit handoff [feature]` | Prompt the current agent session to sync docs and prepare a handoff. |
| `kit summarize [feature]` | Output context summarization instructions. |
| `kit dispatch` | Turn an accepted plan into a post-plan execution topology for independent lanes. |
| `kit code-review` | Output instructions for branch code review. |
| `kit skill mine [feature]` | Output skill extraction prompt for the active coding agent. |

## Utilities

| Command | Description |
| --- | --- |
| `kit upgrade` | Download and install the latest Kit release. |
| `kit version` | Print the installed Kit version. |
| `kit completion` | Generate shell autocompletion scripts. |
| `git wt` | Optional manual wrapper for durable issue lanes, exact path lookup, detached PR views, repair lanes, default writable-lane `.env` links, safe removal, pruning, and legacy migration beneath `~/worktrees`; reconciled rules use native Git. |

## Prompt Profiles And Subagents

## Feature Notes

Use `kit notes` to prepare or add optional source material for a feature before
or during the `kit spec` workflow. Notes live under `docs/notes/<feature>/`:

- `inbox/` stores raw captured notes and conversation excerpts.
- `references/` stores source material, links, examples, and assets.
- `responses/` stores draft or sent responses.
- `private/` stores local-only conversation history and is ignored by git.

`kit notes <feature>` ensures the scaffold. `kit notes <feature> --add` creates
a timestamped note template with front matter for `kind`, `source`, `status`,
`sensitivity`, `captured_at`, and `feature`. Add `--private` for sensitive
conversation context that should not enter the repository. Notes are source
material; promote durable decisions into `SPEC.md`, `docs/CONSTITUTION.md`, or
another canonical document before relying on them for implementation.
Use `docs/references/rules/feature-notes.md` for the agent-facing rules on
loading, referencing, promoting, and ignoring notes.

Prompt-producing commands default to accountable-supervisor orchestration
guidance. The shared decorator stays compact: it uses subagents only when
low-overlap lanes improve correctness or throughput, defaults to at most 3
concurrent lanes, and never exceeds 4. Pass
`--single-agent` when you explicitly want to keep work in one lane.

Prompt-producing commands also support `--profile=frontend` for frontend-heavy
work. The profile keeps Kit's normal RLM flow while adding frontend-specific
guidance for design-system fit, domain-appropriate UI, visual assets,
responsive behavior, browser or screenshot validation, interaction states, and
common generated-UI pitfalls.

Generated prompts resolve repository-discoverable facts before asking the user.
Outside explicit clarification workflows, they ask only for material choices
that cannot be inferred safely and do not require routine approval for safe
in-scope discovery or reversible edits. External, irreversible, production,
Git, and GitHub mutations remain behind their explicit authorization and
repo-local gates.

## Dispatch And Review Loops

Use `kit dispatch` after native planning when an accepted plan needs formal
overlap clustering and Agent Team Plan queueing. Dispatch supports execution
topology; it does not own feature research or design. Dispatch prompts use
`--max-subagents` to cap concurrent spawned agents; the default is 3 and the
hard ceiling is 4.

Use `kit pr fix` as the default PR review feedback entrypoint. With no flags it
lists open pull requests in the current repository and asks which one to repair.
Use `kit pr fix --pr <url|owner/repo#number|number>` to target a specific PR.
The command uses the same prompt-producing path as `kit dispatch --pr`: it
copies a dispatch prompt built from unresolved, non-outdated review threads
directly to the clipboard for a coding agent. Pass `--edit` to review and change
the task list in the default editor first; `--vim` and `--editor <cmd>` also opt
into editing. It does not run the loop agent, edit files, write `.kit/loops`
evidence, stage, commit, push, post PR comments, resolve review threads, or
perform GitHub delivery. The generated prompt tells the coding agent to run a
post-push reflection cycle, confirm the PR head still matches its pushed
commit, and only then resolve verified addressed conversations.

Use `kit dispatch --pr <url|number>` to prefill the dispatch editor from
unresolved, non-outdated GitHub PR review threads. Add `--coderabbit` to keep
only CodeRabbit-authored review comments.

After fixes or no-op decisions are complete, use
`kit dispatch --pr <target> --resolve --yes` to resolve matching unresolved
review threads on GitHub. Resolution is an explicit GitHub mutation and is
not part of raw dispatch prompt generation. Use broad resolution only after the
post-push reflection proves every active conversation in scope was addressed and
no other code was pushed after the repair commit.

Use `kit loop review` when changed code should be reviewed until the local
agent reports at least 95% correctness and no high, medium, or
correctness-impacting issues remain. Without `--pr`, it reviews current-branch
changes relative to `origin/main`, falling back to `main`, plus staged and
unstaged changes.

Use `kit loop review --pr <target>` to fold current CodeRabbit feedback into
that repair loop. Add `--watch` or `--wait-for-coderabbit` only when you want
to wait up to the timeout.

## Output Behavior

Prompt-producing commands, including the Constitution and Makefile setup prompt
emitted by `kit init`, copy generated output to the clipboard by default. The
init prompt maps applicable targets such as `make dev`, `make test`, and
`make check` to verified repository-native commands; it does not leave guessed
or placeholder recipes in the project-owned Makefile. The exact generated
Constitution starter is a valid bootstrap state, so initialization leaves its
sections unchanged until implemented repository evidence supports durable
project-wide truth. Normal coding-agent work then applies the Constitution
curation rule after validation to keep that truth current as the project evolves.

Use:

- `--output-only` to print the raw prompt or output to stdout
- `--output-only --copy` to print and copy
- `--prompt-only` on feature-scoped prompt commands to regenerate prompts
  without mutating repository docs

Human-readable terminal output uses semantic emoji markers, spacing, and ANSI
color when appropriate. Raw `--output-only` payloads and `--json` output avoid
human-readable wrappers.

`kit instructions` always writes raw Markdown to stdout without a banner or
clipboard side effect. It defaults to the explicitly configured current embedded
version, currently `v3`; use `kit instructions --version=v3` for an exact current
selection or an earlier selector such as `kit instructions --version=v1` when a
reproducible historical payload is required.

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
`coding-agent short`, `coding-agent long`, `coding-agent instructions`, the
legacy V2 `kit spec` / `workflow spec` supervisor prompt, support prompts,
`skill mine`, and `project init`.

Use plain `kit spec <feature>` when Kit should create or adopt durable feature
memory around a native plan. Use `kit prompt kit spec` only to render the
legacy active-feature supervisor prompt during compatibility.

## Scaffold And Refresh

`kit scaffold agents` creates or refreshes repository instruction files.
`kit scaffold spec <feature>` creates the current `SPEC.md` scaffold plus
notes/reference-material directories without emitting an agent prompt. Legacy
staged document scaffolds are available only through `kit legacy` commands.

When instruction files already exist:

- default mode skips them and suggests safer next steps
- `--append-only` merges missing Kit-managed sections without overwriting matched existing content
- `--force` overwrites existing files after confirmation
- `--force --yes` overwrites existing files without prompting for automation use

Instruction scaffold versions:

- `--version 1` keeps the legacy verbose `AGENTS.md` / `CLAUDE.md` model
- `--version 2` keeps the legacy thin ToC/RLM instruction model
- `--version 3` uses native-plan repository-memory entrypoints plus `docs/agents/` and `docs/references/`
- new repos default to instruction scaffold version 3
- exact generated V2 instruction artifacts migrate atomically to V3 during a full refresh
- customized V2 instructions stay on V2 until reviewed or explicitly replaced with `--version 3 --force`
- switching models is repo-wide and requires `--force`

`kit reconcile` is the consolidated reviewed refresh command for existing Kit
projects. When files are included, it creates missing Kit-managed files,
atomically migrates exact generated V2 instruction files to V3, preserves
customized V2 instructions with an advisory, refreshes generated support docs,
imports known registry rulesets, and records ruleset registry state in
`.kit.yaml`. It also creates or refreshes the Kit-managed README badge block
when a GitHub repository is configured or discoverable from `origin`. Default
public-repository badges cover last commit, open issues, pull requests,
releases, and conventional CI workflows. Private repositories skip public
Shields GitHub metadata badges and keep only native GitHub Actions workflow
badges when a conventional workflow exists. License badges are not added by
default. It also creates or refreshes `## Maintainers` as the last README H2
with the managed Jameson / `jamesonstone` attribution. It also creates or
refreshes the Kit-managed `.github/workflows/auto-assign.yml` workflow. That
workflow assigns new issues and pull requests to `github.default_assignees` from
the project `.kit.yaml`, falls back to the global `~/.config/kit/.kit.yaml`, and
safely no-ops when no assignees are configured.

Run `kit reconcile` interactively to choose whether to include files, force
changes, and output the follow-up coding-agent prompt.

Use `kit reconcile --include-files --dry-run --diff` to preview managed-file
changes without writing them.

Use `kit reconcile --include-files --force` after reviewing local generated-file
changes when you want to accept refreshed generated guidance. When requested,
Kit outputs a documentation review prompt so an agent can update
`docs/CONSTITUTION.md`, agent docs, references, command docs, and directly
affected feature specs semantically.
