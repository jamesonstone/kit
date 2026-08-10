# Kit Command Guide

The major-release CLI intentionally contains only the following command tree:

```text
kit
├── init
├── reconcile
├── contract resolve
├── pr fix
├── registry add|list|status|view
└── rules add|list|view
```

## `kit init`

Canonically bootstraps the repository: safe local-environment and Makefile
starters, review and GitHub support, bounded README and Constitution sections,
the project progress summary, provider routers, the complete `docs/agents/`
RLM tree, project references, every downstream-visible registry artifact, and
schema-v2 provenance. The complete project and user-config plans are validated
before mutation.

- `--dry-run`: validate and render a plan without writing.
- `--json`: emit the versioned plan, file dispositions, prompt, diagnostics,
  and next actions without writing.
- `--output-only`: apply initialization and emit only the raw
  repository-bootstrap prompt.
- `--copy`: copy the prompt, including when combined with `--output-only`.
- `--registry-repo`, `--registry-branch`, `--catalog`: configure a GitHub source.
- `--registry-path`: use a local registry root.

Default prompt delivery is clipboard-first. The prompt directs the coding
agent to resolve `repository-bootstrap`; Kit itself never reads `.env`, grants
direnv trust, infers project truth, or launches an agent. Repeated init
backfills missing starters in schema-v2 projects while leaving registry drift
to `kit reconcile`; schema-v1 projects must migrate first.

## `kit contract resolve`

Emits resolved-contract schema v1 as JSON. It is deterministic, local-only,
read-only, and performs no Git operation or model inference.

- `--feature <directory>`: inspect the canonical living V3 feature spec and
  expose its structural state, historical links, and phase permissions.
- `--work-type feature|maintenance`: explicitly classify implementation work;
  Kit never infers maintenance from an omitted hint.
- `--path <path>`: add a path hint; repeatable.
- `--applies-to <tag>`: add an applicability tag; repeatable.
- `--workflow <slug>`: request a workflow; repeatable.

For pull-request review repair, request `--workflow pr-feedback-repair`. The
result selects its orchestration and GitHub-delivery dependencies locally; it
does not wait for or collect network feedback.

Blocked required artifacts still produce diagnostic JSON and exit with status
2. Input or project errors exit with status 1.

For features, combine `--workflow implementation-delivery` with
`--work-type feature --feature <feature>`. Genuinely mechanical maintenance
uses `--work-type maintenance`; this recorded hint is the sole feature-spec
exemption. Missing, invalid, unknown, or contradictory classification blocks
the contract. A missing, invalid, or incomplete feature spec keeps
`feature_spec.phase_permissions.spec_authoring` true while source
implementation and delivery remain false. Create or repair the reported
`docs/specs/<feature>/SPEC.md`, then re-run the same local resolution before
source edits. Kit validates structure and phase state; the coding agent authors
the spec from the accepted plan and repository evidence.

## `kit pr fix`

This protected fallback collects current unresolved, non-outdated PR feedback,
prepares the exact writable same-repository PR-head lane, and emits a coding-
agent supervisor prompt. It never launches agents or edits, stages, commits,
pushes, comments, or merges. Without `--pr`, select from current-repository
open PRs. `--pr` accepts a GitHub URL, Markdown link, `owner/repo#number`, or
current-repository number.

- `--output-only`, `--copy`: control prompt delivery; clipboard-first remains
  the default.
- `--edit`, `--editor`, `--vim`: explicitly edit normalized findings before
  prompt generation.
- `--coderabbit`: exclude human feedback and keep CodeRabbit-authored input.
- `--max-subagents`: default 3, hard maximum 4.
- `--include-dirty` or `--exclude-dirty`: explicitly assign pre-existing lane
  changes; overlap with excluded findings fails closed.
- `--wait [--timeout]`: use bounded, rate-aware CodeRabbit observation from the
  local `pr-feedback-repair` contract before collecting.
- `--resolve --head <sha> --thread <id> --yes`: after exact pushed-head
  verification, resolve only explicitly named current active review threads.

Wait failures emit pure
[`pr-feedback-await-v1`](../schemas/pr-feedback-await-v1.schema.json) JSON to
stdout, diagnostics to stderr, and exit with status 2. One-shot invocation
remains the path for late or human feedback. `kit contract resolve` remains
local-only; all network and worktree behavior is confined to this explicit
adapter.

## `kit reconcile`

Builds a read-only drift or schema-migration plan by default.

- `--json`: emit stable plan JSON.
- `--diff`: render affected-file previews.
- `--apply`: apply conflict-free planned changes.
- `--accept-registry <kind>/<slug>`: replace local customization for one exact
  artifact; repeatable.

Attention-required plans exit with status 2. Applying does not overwrite a
same-section conflict unless its exact artifact was accepted.

## `kit registry`

- `list [--json]`: list all typed catalog artifacts and installed state.
- `view <kind>/<slug> [--source|--local|--diff]`: inspect content and provenance.
- `add <kind>/<slug>`: materialize one typed registry artifact.
- `status [--json]`: report freshness, planned changes, and conflicts.

## `kit rules`

`rules add|list|view` use the generic registry core with a `ruleset` filter.
`rules add <slug>` may create a project-local ruleset when the slug is absent
from the configured registry.

## Separate `git-wt` binary

`git-wt` and the `git wt` alias remain unchanged. See
[Git worktrees](references/worktrees.md). They are not `kit` subcommands.
