---
kit_metadata_version: 1
artifact: "plan"
feature:
  id: "0015"
  slug: "pause-remove-commands"
  dir: "0015-pause-remove-commands"
dependencies:
  - name: "config persistence"
    type: "code"
    location: "internal/config/config.go"
    used_for: "storing lifecycle state"
    status: "active"
  - name: "feature selectors"
    type: "code"
    location: "internal/feature/feature.go`, `pkg/cli/*"
    used_for: "consistent resolution and filtering"
    status: "active"
  - name: "rollup contract"
    type: "doc"
    location: "docs/specs/0000_INIT_PROJECT.md"
    used_for: "canonical progress-summary format"
    status: "active"
  - name: "existing lifecycle flows"
    type: "code"
    location: "pkg/cli/status.go`, `pkg/cli/complete.go`, `pkg/cli/handoff_prompt.go"
    used_for: "pause-aware behavior"
    status: "active"
---
# PLAN

## SUMMARY

Add persisted lifecycle metadata in `.kit.yaml`, wire new `pause` and `remove`
commands into the CLI, expose `rm` as the primary removal command, and update
lifecycle views so status, rollup, paused state, and removed feature history
stay consistent without changing the underlying active phase model.

## APPROACH

- extend config with a small per-feature state map keyed by feature directory
  name
- centralize pause lookups and pause mutations in `internal/feature`
- add a shared auto-unpause helper for explicit feature-scoped workflow
  commands
- implement `kit pause` as a non-destructive lifecycle toggle for non-complete
  features
- implement `kit rm` as a destructive lifecycle command with confirmation,
  directory deletion, state cleanup, and rollup regeneration
- keep `kit remove` as a compatibility alias for the same destructive flow
- record removed-feature tombstones in `.kit.yaml` so deleted feature docs are
  gone but `PROJECT_PROGRESS_SUMMARY.md` still shows the feature as removed
- keep `docs/notes/<feature>` by default so follow-up features can reuse notes,
  with interactive and `--notes` deletion paths for users who want full cleanup
- update rollup and status rendering to surface paused state separately from
  phase
- update `kit status --all` and `kit rm` output so removed tombstones are
  visible instead of hidden
- keep default `status` active-feature focused and move the fleet view into the
  explicit `status --all` mode
- exclude paused features from active-only multi-feature flows except `status`
  and `status --all`

## COMPONENTS

- `internal/config/config.go`
  - persist per-feature paused state and removed-feature tombstones in
    `.kit.yaml`
- `internal/feature/feature.go`
  - attach paused state to feature listings and helpers
- `internal/feature/status.go`
  - include paused state, removed metadata, and optional notes status in the
    feature status payload while keeping active feature selection number-based
- `internal/rollup/rollup.go`
  - render paused state and removed feature tombstones in the progress table and
    feature summaries, including retained notes pointers when available
- `pkg/cli/pause.go`
  - add pause command, selection, validation, persistence, and rollup update
- `pkg/cli/remove.go`
  - add rm/remove command surface, confirmation, deletion, state cleanup,
    removed tombstone persistence, notes retention/removal, removed history
    output, and rollup update
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
  - add `removed_features` tombstones keyed by feature directory metadata for
    history after docs are deleted
- `feature.Feature`
  - add `Paused bool`
- `feature.FeatureStatus`
  - add `Paused bool`, removed metadata, and optional notes status
- `rollup.FeatureSummary`
  - add `Paused bool`
  - add `Removed bool` and removed timestamp support

## INTERFACES

- new command: `kit pause [feature]`
  - no feature argument: interactive selector over non-complete features
  - explicit feature: pause that feature
- new command: `kit rm [feature] [--yes]`
  - no feature argument: interactive selector over all existing features
  - confirmation required unless `--yes` is set
  - notes under `docs/notes/<feature>` are retained by default
  - interactive runs ask whether to remove notes when notes exist
  - `--notes` removes notes too
- compatibility alias: `kit remove [feature] [--yes]`
  - invokes the same command path as `kit rm`
- status json payload
  - include `paused` under `active_feature`
- progress summary table
  - add `PAUSED` column next to `PHASE`
  - render removed tombstones with `PHASE` set to `removed`
- status all-features matrix
  - render removed tombstones with `State` set to `REMOVED`
  - include whether notes are retained

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
- deleting feature docs would otherwise remove all rollup evidence that the
  feature existed
  - mitigate by storing a minimal tombstone outside the deleted feature
    directory
- paused-state rendering could conflict with existing phase assumptions in docs
  - mitigate by updating the core project spec and keeping paused as a separate
    flag, not a replacement phase
- destructive remove flow could delete the wrong feature
  - mitigate with clear selector labeling and confirmation that names the target
- users may need removed-feature notes for a follow-up feature
  - mitigate by retaining notes by default and requiring an explicit interactive
    choice or `--notes` flag to delete them

## TESTING

- unit tests for config-backed pause persistence helpers
- unit tests for pause-aware feature listing and status payloads
- CLI tests for `kit pause` selection, rejection of complete features, and
  idempotence
- CLI tests for auto-unpause on explicit feature-scoped workflow commands
- CLI tests for `kit remove` confirmation, `--yes`, directory deletion, state
  cleanup, tombstone persistence, notes retention/removal, and removed rollup
  rendering
- status and rollup tests for the new paused column and paused summary output
- rollup tests for removed tombstones when the feature directory no longer
  exists
- status tests for removed tombstones and notes retention in `kit status --all`
