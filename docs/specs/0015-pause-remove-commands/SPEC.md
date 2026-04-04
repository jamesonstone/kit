# SPEC

## SUMMARY

Add `kit pause` and `kit remove` so users can explicitly pause in-flight work
or remove a feature cleanly while keeping Kit's generated progress views and
selectors consistent.

## PROBLEM

Kit currently has no lifecycle controls for work that should stop without being
completed or for feature directories that should be removed entirely. That
forces users to manage state by hand across `docs/specs/`, `.kit.yaml`,
`PROJECT_PROGRESS_SUMMARY.md`, and `kit status`, which is error-prone and can
leave stale active-feature views behind.

## GOALS

- add `kit pause [feature]` to persist a paused flag for non-complete features
- add `kit remove [feature]` to delete a feature directory after confirmation
- keep paused state separate from the underlying workflow phase
- show paused state in `kit status` and `PROJECT_PROGRESS_SUMMARY.md`
- exclude paused features from active-only multi-feature flows other than
  `kit status`
- automatically clear paused state when a user resumes a feature through an
  explicit feature-scoped workflow command
- remove deleted features from generated lifecycle views and persisted state

## NON-GOALS

- add a dedicated `kit resume` command in this phase
- preserve removed feature numbers for future feature allocation
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

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| feature lifecycle model | code | `internal/feature/feature.go` | phase derivation and feature resolution | active |
| feature status model | code | `internal/feature/status.go` | active-feature selection and status payloads | active |
| project rollup generator | code | `internal/rollup/rollup.go` | `PROJECT_PROGRESS_SUMMARY.md` regeneration | active |
| status output | code | `pkg/cli/status.go`, `pkg/cli/status_output.go` | user-facing lifecycle views | active |
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
- `kit status` must continue to treat the highest-numbered feature as the
  active feature even if that feature is paused
- `kit status` text output must show whether the active feature is paused
- `kit status --json` must include paused state for the active feature
- `kit status` all-features output must add a paused column
- `PROJECT_PROGRESS_SUMMARY.md` feature progress table must add a paused column
- `PROJECT_PROGRESS_SUMMARY.md` feature summaries must reflect paused state in a
  stable, generated way
- active-only multi-feature flows other than `kit status` must exclude paused
  features by default
- `kit remove [feature]` must resolve the feature by existing feature reference
  rules and must support interactive selection when no feature argument is
  provided
- `kit remove` must require explicit confirmation before deletion and must
  support `--yes` to skip the confirmation prompt
- `kit remove` must delete the target feature directory and all files under it
- `kit remove` must remove any persisted paused state for the deleted feature
- `kit remove` must regenerate `PROJECT_PROGRESS_SUMMARY.md` after deletion
- deleted features must disappear from `kit status`, selectors, and generated
  lifecycle views after removal
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
- `kit status` shows the latest feature even when paused and labels it paused
- `kit status --json` includes paused state in the active-feature payload
- the `kit status` all-features table includes a paused column
- `PROJECT_PROGRESS_SUMMARY.md` includes a paused column in the feature progress
  table and a paused field in each generated feature summary
- paused features are excluded from active-only flows such as handoff's active
  feature set and `kit complete --all`
- running `kit remove <feature>` asks for confirmation, deletes the feature
  directory on approval, clears any persisted state for that feature, and
  regenerates `PROJECT_PROGRESS_SUMMARY.md`
- running `kit remove <feature> --yes` performs the same deletion without
  prompting
- removed features no longer appear in status views or generated lifecycle docs
- automated tests cover pause persistence, auto-unpause on explicit resume,
  remove confirmation and deletion, and paused-state rendering in status/rollup

## EDGE-CASES

- pausing a feature that is already paused
- pausing the current highest-numbered feature
- attempting to pause a complete feature
- removing a paused feature
- removing the current highest-numbered feature
- removing a feature when `PROJECT_PROGRESS_SUMMARY.md` does not yet exist
- removing a feature that has only a subset of workflow artifacts
- explicit workflow commands targeting a paused feature by partial or numeric
  reference
- interactive pause or remove selection with invalid input
- deleting the final remaining feature in the project

## OPEN-QUESTIONS

- none
