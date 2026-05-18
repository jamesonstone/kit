---
kit_metadata_version: 1
artifact: "spec"
feature:
  id: "0015"
  slug: "pause-remove-commands"
  dir: "0015-pause-remove-commands"
references:
  - name: feature lifecycle model
    type: code
    target: internal/feature/feature.go
    relation: implements
    read_policy: conditional
    used_for: phase derivation and feature resolution
    status: active
  - name: feature status model
    type: code
    target: internal/feature/status.go
    relation: implements
    read_policy: conditional
    used_for: active-feature selection and status payloads
    status: active
  - name: project rollup generator
    type: code
    target: internal/rollup/rollup.go
    relation: implements
    read_policy: conditional
    used_for: '`PROJECT_PROGRESS_SUMMARY.md` regeneration'
    status: active
  - name: status output
    type: code
    target: pkg/cli/status.go`, `pkg/cli/status_render.go
    relation: implements
    read_policy: conditional
    used_for: user-facing lifecycle views
    status: active
  - name: complete command
    type: code
    target: pkg/cli/complete.go
    relation: implements
    read_policy: conditional
    used_for: active-only lifecycle flow behavior
    status: active
  - name: project spec
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: informs
    read_policy: conditional
    used_for: canonical progress-summary contract
    status: active
---
# SPEC

## SUMMARY

Add `kit pause`, `kit rm`, and the `kit remove` compatibility alias so users
can explicitly pause in-flight work or remove a feature's docs while keeping
Kit's generated progress views, removed-history rows, retained notes, and
selectors consistent.

## PROBLEM

Kit currently has no lifecycle controls for work that should stop without being
completed or for feature directories that should be removed entirely. That
forces users to manage state by hand across `docs/specs/`, `.kit.yaml`,
`PROJECT_PROGRESS_SUMMARY.md`, and `kit status`, which is error-prone and can
leave stale active-feature views behind.

## GOALS

- add `kit pause [feature]` to persist a paused flag for non-complete features
- add `kit rm [feature]` to delete a feature directory after confirmation
- keep `kit remove [feature]` available as a compatibility alias for `kit rm`
- keep paused state separate from the underlying workflow phase
- show paused state in `kit status` and `PROJECT_PROGRESS_SUMMARY.md`
- exclude paused features from active-only multi-feature flows other than
  `kit status` and the new explicit `kit status --all` overview mode
- automatically clear paused state when a user resumes a feature through an
  explicit feature-scoped workflow command
- remove deleted features from active selectors and status views
- retain removed feature tombstones so `PROJECT_PROGRESS_SUMMARY.md` shows the
  feature as `removed` after its docs are deleted
- show removed tombstones in `kit status --all` and `kit rm` history output
- retain feature notes by default when removing docs, with an explicit
  interactive choice and `--notes` flag to delete notes too

## NON-GOALS

- redesign the lifecycle model beyond the paused flag and explicit resume
  commands
- preserve removed feature numbers for future feature allocation
- preserve removed feature documents after `kit rm`
- delete feature notes by default when feature docs are removed
- rewrite arbitrary historical markdown outside Kit-managed lifecycle views
- add bulk remove or bulk pause flows

## USERS

- maintainers who need to suspend a feature without losing its actual phase
- maintainers who need to remove abandoned or mistaken feature directories
- coding agents that rely on accurate active-feature and rollup state

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| feature lifecycle model | code | `internal/feature/feature.go` | phase derivation and feature resolution | active |
| feature status model | code | `internal/feature/status.go` | active-feature selection and status payloads | active |
| project rollup generator | code | `internal/rollup/rollup.go` | `PROJECT_PROGRESS_SUMMARY.md` regeneration | active |
| status output | code | `pkg/cli/status.go`, `pkg/cli/status_render.go` | user-facing lifecycle views | active |
| complete command | code | `pkg/cli/complete.go` | active-only lifecycle flow behavior | active |
| project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical progress-summary contract | active |

## REQUIREMENTS

- `kit pause [feature]` must resolve the feature by existing feature reference
  rules and must support interactive selection when no feature argument is
  provided
- `kit pause` must reject complete features with an actionable error
- `kit pause` must persist paused state in `.kit.yaml` under a per-feature
  state map keyed by feature directory name
- `kit pause` must be idempotent and must not mutate feature documents beyond
  regenerating Kit-managed lifecycle views
- paused state must not replace or rewrite the feature's underlying workflow
  phase
- explicit feature-scoped workflow commands that continue work on a paused
  feature must clear the paused flag before proceeding
- `kit status` must continue to treat the highest-numbered non-backlog feature
  as the active feature even if that feature is paused
- `kit status` text output must show whether the active feature is paused
- `kit status --json` must include paused state for the active feature
- `kit status --all` must render the all-features output as a fixed-width
  lifecycle matrix and must add a paused or backlog indicator
- `PROJECT_PROGRESS_SUMMARY.md` feature progress table must add a paused column
- `PROJECT_PROGRESS_SUMMARY.md` feature summaries must reflect paused state in a
  stable, generated way
- active-only multi-feature flows other than `kit status` must exclude paused
  features by default
- `kit rm [feature]` must resolve the feature by existing feature reference
  rules and must support interactive selection when no feature argument is
  provided
- `kit remove [feature]` must invoke the same removal flow as `kit rm [feature]`
- `kit rm` must require explicit confirmation before deletion and must
  support `--yes` to skip the confirmation prompt
- `kit rm` must delete the target feature directory and all files under it
- `kit rm` must remove any persisted paused state for the deleted feature
- `kit rm` must retain `docs/notes/<feature>` by default
- `kit rm` must offer an interactive notes-removal prompt when notes exist and
  deletion is not running under `--yes`
- `kit rm --notes` must remove `docs/notes/<feature>` along with the feature
  docs without changing the meaning of `--yes`
- `kit rm` output must make the final `removed` state and notes
  retention/deletion visible
- `kit rm` must persist a removed-feature tombstone outside the deleted feature
  directory with enough metadata to render project history
- `kit rm` must regenerate `PROJECT_PROGRESS_SUMMARY.md` after deletion
- `PROJECT_PROGRESS_SUMMARY.md` must retain a row for removed features with
  `PHASE` set to `removed`
- deleted features must disappear from default active `kit status` and active
  selectors after removal while remaining visible in `kit status --all`, `kit rm`
  removed history, and project progress history
- `kit status --all` must include removed tombstones with state `REMOVED` and a
  notes-retention marker
- adding pause support must not change the meaning of existing workflow phases
  for unpaused features

## ACCEPTANCE

- running `kit pause <feature>` marks that feature paused in `.kit.yaml`,
  regenerates `PROJECT_PROGRESS_SUMMARY.md`, and leaves the feature's phase
  unchanged
- running `kit pause` with no arguments shows an interactive selector of
  non-complete features
- running `kit pause <feature>` twice does not fail and does not duplicate state
- running `kit plan <feature>`, `kit tasks <feature>`, `kit implement <feature>`,
  or another explicit feature-scoped workflow command on a paused feature clears
  the paused flag before continuing
- `kit status` shows the latest non-backlog feature even when paused and labels
  it paused
- `kit status --json` includes paused state in the active-feature payload
- the `kit status --all` matrix includes paused or backlog state
- `PROJECT_PROGRESS_SUMMARY.md` includes a paused column in the feature progress
  table and a paused field in each generated feature summary
- paused features are excluded from active-only flows such as handoff's active
  feature set and `kit complete --all`
- running `kit rm <feature>` asks for confirmation, deletes the feature
  directory on approval, clears any persisted state for that feature, and
  regenerates `PROJECT_PROGRESS_SUMMARY.md`
- running `kit rm <feature> --yes` performs the same deletion without
  prompting
- running `kit remove <feature> --yes` performs the same deletion through the
  compatibility alias
- removed features no longer appear in default active status or active selectors
- removed features remain in `PROJECT_PROGRESS_SUMMARY.md` with `PHASE` set to
  `removed`
- removed features appear in `kit status --all` as `REMOVED`
- removed features appear in `kit rm` removed-history output
- feature notes are retained by default after `kit rm <feature> --yes`
- feature notes are removed when `kit rm <feature> --yes --notes` is used
- interactive removal asks whether to remove notes when notes exist
- automated tests cover pause persistence, auto-unpause on explicit resume,
  remove confirmation and deletion, removed tombstone rendering, and
  paused-state rendering in status/rollup

## EDGE-CASES

- pausing a feature that is already paused
- pausing the current highest-numbered feature
- pausing a brainstorm-phase feature that is later used as a backlog item
- attempting to pause a complete feature
- removing a paused feature
- removing the current highest-numbered feature
- removing a feature when `PROJECT_PROGRESS_SUMMARY.md` does not yet exist
- removing a feature that has only a subset of workflow artifacts
- removing a feature after it was paused
- rendering removed feature history after the feature docs no longer exist
- rendering removed feature history when feature notes are retained
- removing feature notes during interactive and flag-driven removal
- explicit workflow commands targeting a paused feature by partial or numeric
  reference
- interactive pause or remove selection with invalid input
- deleting the final remaining feature in the project

## OPEN-QUESTIONS

- none
