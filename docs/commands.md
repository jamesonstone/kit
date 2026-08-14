# Kit Commands

Kit 2.0 exposes a deliberately reduced command tree. Command names below are
supported; flags and output formats remain command-specific.

## Agent Evidence

| Command | Purpose |
| --- | --- |
| `kit capabilities [command]` | Read-only command capability and side-effect discovery. |
| `kit context resolve` | Deterministically select ordered local workflow, rule, spec, reference, and source evidence. |

`kit context resolve` accepts `--workflow`, `--feature`, and repeatable
`--path` hints plus `--json`. It performs no network access, writes, Git
operations, model inference, or agent launch. Missing required evidence returns
a blocked contract and nonzero status.

## Bootstrap And Feature Memory

| Command | Purpose |
| --- | --- |
| `kit init` | Canonical repository bootstrap and managed-file refresh. |
| `kit spec [feature]` | Create, adopt, or orient a living V3 `SPEC.md`; existing V1/V2 specs remain readable. |
| `kit instructions` | Print versioned provider-neutral agent instructions. |

Fresh initialization preserves existing project-owned files and materializes
the current rules and workflow starters. Use `kit init --refresh --dry-run
--diff` to preview managed bootstrap changes.

## Rules And Maintenance

| Command | Purpose |
| --- | --- |
| `kit rules add` | Import or create a repository-local ruleset. |
| `kit rules list` | List installed and available rulesets. |
| `kit rules view` | Inspect a local or registry ruleset. |
| `kit rules link` | Link a ruleset from feature metadata. |
| `kit registry status` | Report registry and managed-file freshness. |
| `kit reconcile` | Preserve the existing project/file/rule/document reconciliation interface. |
| `kit health` | Apply safe managed updates and validate the project contract; supports `--dry-run --diff`. |

Preview managed reconciliation before applying it:

```bash
kit reconcile --include-files --dry-run --diff
kit reconcile --include-files
```

## Agent Execution Adapters

| Command | Purpose |
| --- | --- |
| `kit dispatch` | Produce an Agent Team Plan prompt after native planning; optional PR/watch modes remain bounded. |
| `kit pr fix` | Select or target a PR and produce a repair prompt from current unresolved review feedback. |
| `kit pr orchestrate` | Resolve bounded repository scope into a deterministic dependency-aware release prompt; Kit does not execute the release. |

These commands may perform their documented GitHub or exact-lane preparation,
but do not launch or supervise coding agents. Resolve
`pr-feedback-repair` context before agent repair work. Release agents resolve
`release-orchestration` and then `pull-request-merge` before any authorized
merge or merge-queue mutation.

## Inspection And Validation

| Command | Purpose |
| --- | --- |
| `kit status` | Show current feature and Kit-managed state. |
| `kit check` | Validate feature or project documents. |
| `kit config check` | Validate and safely repair `.kit.yaml`, including interactive AWS profile, account, and enabled-Region selection. |
| `kit aws verify` | Verify the configured AWS profile, account, and Region. |
| `kit improve run` | Run deterministic Kit harness benchmark suites. |

## Local Usage

| Command | Purpose |
| --- | --- |
| `kit usage` / `kit usage report` | Aggregate bounded local command usage; default window is 90 days. |
| `kit usage status` | Show effective collection state, storage bounds, coverage, and diagnostics. |
| `kit usage refresh` | Validate, rotate, and prune usage storage; supports `--dry-run`. |
| `kit usage clear` | Remove all or filtered usage events with confirmation. |
| `kit usage enable` / `disable` | Set exactly one `--global` or `--project` preference. |

Usage collection is local-only and records no arguments, output, raw project
identity, paths, content, environment values, or secrets. A global disable
overrides project settings. Usage commands are excluded from their own data.

## Utilities

- `kit upgrade`
- `kit version`
- `kit completion`
- normal `kit help`

## Removed In Version 2

The following former top-level groups are absent: backlog, brainstorm, catchup,
CI diagnosis, completion lifecycle, eval, handoff, implement, legacy staged
commands, loop runtime, map, notes, pause/resume/remove lifecycle, plan/tasks,
project refresh, prompt library, reflect, replay/state/trace, scaffold, set,
skill, summarize, and verify. `dispatch` is explicitly retained.

Existing repository files are not deleted. Use the [migration
guide](migration-v2.md) to replace command references safely.
