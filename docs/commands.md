# Kit Command Guide

The major-release CLI intentionally contains only the following command tree:

```text
kit
├── init
├── reconcile
├── contract resolve
├── registry add|list|status|view
└── rules add|list|view
```

## `kit init`

Materializes every downstream-visible registry artifact, provider routing
sections, and schema-v2 provenance. The complete catalog and write plan are
validated before mutation.

- `--dry-run`: validate and render a plan without writing.
- `--json`: emit the versioned plan as JSON.
- `--registry-repo`, `--registry-branch`, `--catalog`: configure a GitHub source.
- `--registry-path`: use a local registry root.

## `kit contract resolve`

Emits resolved-contract schema v1 as JSON. It is deterministic, local-only,
read-only, and performs no Git operation or model inference.

- `--feature <id>`: add an explicit feature applicability hint.
- `--path <path>`: add a path hint; repeatable.
- `--applies-to <tag>`: add an applicability tag; repeatable.
- `--workflow <slug>`: request a workflow; repeatable.

For pull-request review repair, request `--workflow pr-feedback-repair`. The
result selects its orchestration and GitHub-delivery dependencies locally; it
does not wait for or collect network feedback.

Blocked required artifacts still produce diagnostic JSON and exit with status
2. Input or project errors exit with status 1.

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
