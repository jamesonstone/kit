# SPEC

## SUMMARY

Add a first-class `kit backlog` command and matching `kit brainstorm` backlog
flags so users can capture out-of-scope follow-up features, list them later,
and pick them back up without introducing a second document format. The
command remains the backlog-specific surface after command-surface
simplification, while `kit resume` becomes the canonical general resume entry
point.

## PROBLEM

Users often discover legitimate follow-up work while actively defining or
implementing another feature. That follow-up work is real enough to capture,
but it is intentionally out of scope for the current implementation. Today the
closest durable artifact is a normal feature directory, but Kit has no focused
workflow for recording that future work as deferred, listing those deferred
items later, or resuming one without it taking over the active feature lane.

## GOALS

- add `kit backlog` as a read-only backlog view over deferred feature work
- keep backlog items in the existing feature directory model instead of adding
  a new markdown artifact type
- let users create deferred backlog items from `kit brainstorm`
- let users pick up a backlog item from `kit backlog`, `kit resume`, or the
  deprecated compatibility path `kit brainstorm --pickup`
- make backlog list output concise and human-readable
- keep backlog items out of active-feature status selection until picked up

## NON-GOALS

- add a separate `FEATURE_BACKLOG.md` or any other new persistent document type
- add a new persisted lifecycle state beyond the existing paused flag
- support backlog items that start in `spec`, `plan`, `tasks`, or later phases
- add batch pickup, priority ordering, due dates, or owners in this phase
- change the generated `PROJECT_PROGRESS_SUMMARY.md` into an editable scratchpad

## USERS

- maintainers who need to capture follow-up work without derailing the current
  in-scope feature
- coding agents that need an explicit, durable place to park future work
- maintainers who want a short list of deferred items they can resume later

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: 0004-brainstorm-first-workflow
- builds on: 0015-pause-remove-commands
- related to: 0007-catchup-command

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| brainstorm command | code | `pkg/cli/brainstorm.go` | capture and resume deferred brainstorm items | active |
| feature lifecycle state | code | `internal/config/config.go`, `internal/feature/lifecycle.go` | paused-state persistence and resume behavior | active |
| feature model | code | `internal/feature/feature.go`, `internal/feature/status.go` | phase derivation and active-feature selection | active |
| rollup generator | code | `internal/rollup/rollup.go` | generated project summary after backlog mutations | active |
| status command | code | `pkg/cli/status.go`, `pkg/cli/status_output.go` | active-feature presentation after backlog capture | active |
| project command contract | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical CLI behavior updates | active |

## REQUIREMENTS

- `kit brainstorm --backlog` must create or reuse a feature directory, ensure
  `BRAINSTORM.md` exists, mark the feature paused, and refresh
  `PROJECT_PROGRESS_SUMMARY.md`
- `kit brainstorm --backlog` must keep vim-compatible editor input as the
  default experience when it needs to collect a new brainstorm thesis
- `kit brainstorm --backlog` must not output a planning prompt during deferred
  capture; it is a capture-only flow
- `kit resume [feature]` must resolve an existing backlog item, clear its
  paused state, refresh `PROJECT_PROGRESS_SUMMARY.md`, and output the same
  brainstorm planning prompt that a normal active brainstorm resume uses
- `kit backlog` with no arguments must render a concise table with columns
  `feature` and `description`
- `kit backlog` descriptions must prefer `BRAINSTORM.md` `SUMMARY` and fall
  back to `USER THESIS`
- `kit backlog --pickup [feature]` must clear the paused state for the selected
  backlog item, refresh `PROJECT_PROGRESS_SUMMARY.md`, and output the brainstorm
  planning prompt
- `kit backlog` help and user-facing docs must identify `kit resume` as the
  canonical general resume flow while preserving `kit backlog --pickup` as the
  backlog-specific shortcut
- `kit backlog --pickup` with no feature argument must show an interactive
  selector over eligible backlog items
- a backlog item must be defined as a paused feature whose current phase is
  `brainstorm`
- backlog items must remain normal numbered feature directories under
  `docs/specs/`
- `kit status` must exclude backlog items from active-feature selection so a
  newly captured deferred item does not replace the current in-scope feature
- if every existing feature is a backlog item, `kit status` must report no
  active feature in progress rather than treating a deferred item as active
- `kit backlog` must reject pickup attempts for features that are not backlog
  items with an actionable error
- `kit brainstorm --pickup` must remain callable for compatibility but must be
  hidden from default help and documented as deprecated in favor of
  `kit resume <feature>` or `kit backlog --pickup <feature>`
- `kit backlog --pickup` and `kit resume` must support existing feature
  reference rules and interactive selection when no feature argument is
  provided where appropriate
- `kit backlog` must not introduce or require a new `.kit.yaml` schema field
- shipped docs and help text must explain backlog capture, backlog listing, and
  backlog pickup clearly

## ACCEPTANCE

- running `kit brainstorm --backlog` creates a paused brainstorm-phase feature
  and updates `PROJECT_PROGRESS_SUMMARY.md`
- running `kit backlog` lists only paused brainstorm-phase features in a two
  column human-readable table
- running `kit backlog --pickup <feature>` clears paused state and outputs a
  brainstorm planning prompt for that feature
- running `kit resume <feature>` performs the same pickup flow for backlog
  targets
- running `kit brainstorm --pickup <feature>` still performs the same pickup
  flow while emitting deprecation guidance
- after capturing a backlog item, `kit status` still focuses on the active
  non-backlog feature when one exists
- when all features are backlog items, `kit status` reports that there is no
  active feature in progress
- automated tests cover backlog filtering, backlog list rendering, pickup
  resume behavior, and active-feature selection with deferred items present

## EDGE-CASES

- capturing a backlog item for an existing feature that already has
  `BRAINSTORM.md`
- capturing a backlog item when no other features exist yet
- picking up a backlog item by slug or full directory name
- running `kit backlog` when no backlog items exist
- running `kit backlog --pickup` when no backlog items exist
- attempting to pick up a paused feature in a later phase
- attempting to pick up an unpaused brainstorm feature
- creating multiple backlog items while one active feature remains in progress
- running `kit status` when only backlog items exist

## OPEN-QUESTIONS

- none
