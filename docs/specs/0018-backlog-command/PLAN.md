# PLAN

## SUMMARY

Model backlog as a filtered view over paused brainstorm-phase features. Keep
`kit backlog` as the backlog-specific list and pickup surface, route the new
canonical `kit resume` command through the same pickup helper for backlog
targets, and keep deferred items out of active-feature status selection until
resumed.

## APPROACH

- keep persistence unchanged by reusing the existing paused lifecycle flag
- define backlog eligibility structurally: paused + `brainstorm` phase
- add shared backlog helpers for filtering, description extraction, selection,
  and pickup validation
- make `kit brainstorm --backlog` a capture-only flow
- make `kit backlog --pickup`, `kit resume`, and the deprecated
  `kit brainstorm --pickup` path share the same resume helper and prompt output
- update active-feature selection logic so deferred items do not become active

## COMPONENTS

- `pkg/cli/backlog.go`
  - register the command
  - render list output
  - handle pickup mode flags
- `pkg/cli/brainstorm.go` and supporting backlog helper file
  - add `--backlog` and `--pickup`
  - split capture-only backlog creation from normal brainstorm prompting
- `internal/feature`
  - add backlog classification and filtered listing helpers
  - add active-feature selection that excludes backlog items when lifecycle
    state is available
- `pkg/cli/status.go` and `pkg/cli/status_output.go`
  - route through the new active-feature selection behavior
  - improve no-active guidance when only backlog items exist
- docs
  - update README and `docs/specs/0000_INIT_PROJECT.md`

## DATA

- no new persisted data structures
- backlog membership is derived from:
  - feature paused state in `.kit.yaml`
  - current phase derived from feature artifacts
- backlog descriptions come from `BRAINSTORM.md`:
  - first choice: `## SUMMARY`
  - fallback: `## USER THESIS`

## INTERFACES

- `kit backlog`
  - list backlog items in a concise table
- `kit backlog --pickup [feature]`
  - resume a backlog item and output the brainstorm prompt
- `kit resume [feature]`
  - route backlog targets through the same pickup behavior
- `kit brainstorm --backlog [feature]`
  - capture and defer a brainstorm item without outputting a prompt
- `kit brainstorm --pickup [feature]`
  - deprecated compatibility path for the same pickup behavior

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| feature paused-state persistence | code | `internal/feature/lifecycle.go` | defer and pickup mutations | active |
| brainstorm prompt generation | code | `pkg/cli/brainstorm.go` | resumed backlog prompt output | active |
| brainstorm summary extraction | code | `internal/feature/status.go` | backlog description rendering | active |
| rollup generation | code | `internal/rollup/rollup.go` | post-mutation project summary refresh | active |

## RISKS

- status behavior can regress if active-feature filtering excludes too much
  work; mitigate with focused selection tests
- backlog capture can become confusing if it still behaves like a normal
  brainstorm prompt flow; mitigate by making `--backlog` capture-only
- pickup flows can drift apart between `backlog` and `brainstorm`; mitigate by
  sharing one resume helper

## TESTING

- unit tests for backlog eligibility and filtered listing
- command tests for:
  - `kit backlog` list output
  - `kit backlog --pickup`
  - `kit brainstorm --backlog`
  - `kit brainstorm --pickup`
- status tests for active-feature selection when backlog items exist
- targeted rollup assertions for backlog capture and pickup state changes
