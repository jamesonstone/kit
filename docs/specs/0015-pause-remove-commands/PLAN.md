# PLAN

## SUMMARY

Add a persisted paused flag in `.kit.yaml`, wire new `pause` and `remove`
commands into the CLI, and update lifecycle views so status and rollup reflect
paused state without changing the underlying phase model.

## APPROACH

- extend config with a small per-feature state map keyed by feature directory
  name
- centralize pause lookups and pause mutations in `internal/feature`
- add a shared auto-unpause helper for explicit feature-scoped workflow
  commands
- implement `kit pause` as a non-destructive lifecycle toggle for non-complete
  features
- implement `kit remove` as a destructive lifecycle command with confirmation,
  directory deletion, state cleanup, and rollup regeneration
- update rollup and status rendering to surface paused state separately from
  phase
- keep default `status` active-feature focused and move the fleet view into the
  explicit `status --all` mode
- exclude paused features from active-only multi-feature flows except `status`
  and `status --all`

## COMPONENTS

- `internal/config/config.go`
  - persist per-feature paused state in `.kit.yaml`
- `internal/feature/feature.go`
  - attach paused state to feature listings and helpers
- `internal/feature/status.go`
  - include paused state in the feature status payload while keeping active
    feature selection number-based
- `internal/rollup/rollup.go`
  - render paused state in the progress table and feature summaries
- `pkg/cli/pause.go`
  - add pause command, selection, validation, persistence, and rollup update
- `pkg/cli/remove.go`
  - add remove command, confirmation, deletion, state cleanup, and rollup update
- existing explicit feature-scoped commands
  - clear pause before continuing work on an explicitly targeted feature
- `pkg/cli/status_output.go`
  - render paused state in both active-feature and all-features views
- `pkg/cli/complete.go`, `pkg/cli/handoff_prompt.go`
  - exclude paused features from active-only flows

## DATA

- `.kit.yaml`
  - add `feature_state` map keyed by feature directory name
  - each entry stores `paused: true|false`
- `feature.Feature`
  - add `Paused bool`
- `feature.FeatureStatus`
  - add `Paused bool`
- `rollup.FeatureSummary`
  - add `Paused bool`

## INTERFACES

- new command: `kit pause [feature]`
  - no feature argument: interactive selector over non-complete features
  - explicit feature: pause that feature
- new command: `kit remove [feature] [--yes]`
  - no feature argument: interactive selector over all existing features
  - confirmation required unless `--yes` is set
- status json payload
  - include `paused` under `active_feature`
- progress summary table
  - add `PAUSED` column next to `PHASE`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| config persistence | code | `internal/config/config.go` | storing lifecycle state | active |
| feature selectors | code | `internal/feature/feature.go`, `pkg/cli/*` | consistent resolution and filtering | active |
| rollup contract | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical progress-summary format | active |
| existing lifecycle flows | code | `pkg/cli/status.go`, `pkg/cli/complete.go`, `pkg/cli/handoff_prompt.go` | pause-aware behavior | active |

## RISKS

- duplicated pause checks across commands could drift
  - mitigate by adding shared helpers in `internal/feature`
- removing a feature without clearing persisted state could leave stale config
  - mitigate by making removal clean up `.kit.yaml` before regenerating rollup
- paused-state rendering could conflict with existing phase assumptions in docs
  - mitigate by updating the core project spec and keeping paused as a separate
    flag, not a replacement phase
- destructive remove flow could delete the wrong feature
  - mitigate with clear selector labeling and confirmation that names the target

## TESTING

- unit tests for config-backed pause persistence helpers
- unit tests for pause-aware feature listing and status payloads
- CLI tests for `kit pause` selection, rejection of complete features, and
  idempotence
- CLI tests for auto-unpause on explicit feature-scoped workflow commands
- CLI tests for `kit remove` confirmation, `--yes`, directory deletion, and
  state cleanup
- status and rollup tests for the new paused column and paused summary output
