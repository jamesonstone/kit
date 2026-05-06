# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Define pause/remove lifecycle contract in canonical docs | done | codex | none |
| T002 | Add persisted pause state to config and feature models | done | codex | T001 |
| T003 | Implement `kit pause` and explicit-resume auto-unpause behavior | done | codex | T002 |
| T004 | Implement `kit remove` with confirmation, deletion, and state cleanup | done | codex | T002 |
| T005 | Update rollup, status, and active-only flows for paused state | done | codex | T002, T003, T004 |
| T006 | Add regression tests and verify command behavior | done | codex | T003, T004, T005 |
| T007 | Promote `kit rm` as the feature removal command while preserving `kit remove` | done | oz | T004 |
| T008 | Retain removed feature tombstones in project progress summary | done | oz | T004, T007 |
| T009 | Surface removed features in status/rm and add notes removal controls | done | oz | T008 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Define pause/remove lifecycle contract in canonical docs
- [x] T002: Add persisted pause state to config and feature models
- [x] T003: Implement `kit pause` and explicit-resume auto-unpause behavior
- [x] T004: Implement `kit remove` with confirmation, deletion, and state cleanup
- [x] T005: Update rollup, status, and active-only flows for paused state
- [x] T006: Add regression tests and verify command behavior
- [x] T007: Promote `kit rm` as the feature removal command while preserving `kit remove`
- [x] T008: Retain removed feature tombstones in project progress summary
- [x] T009: Surface removed features in status/rm and add notes removal controls

## TASK DETAILS

### T001
- **GOAL**: record the approved lifecycle semantics before code changes
- **SCOPE**: add dedicated feature spec docs and update core lifecycle docs if
  the generated progress-summary contract changes
- **ACCEPTANCE**: `SPEC.md`, `PLAN.md`, and `TASKS.md` are complete and
  consistent with approved defaults
- **NOTES**: keep paused as a separate flag, not a phase

### T002
- **GOAL**: persist and expose paused state in a single consistent model
- **SCOPE**: extend `.kit.yaml`, feature listing, feature status, and shared
  helpers
- **ACCEPTANCE**: pause state loads and saves cleanly and is available to CLI
  flows without ad hoc parsing
- **NOTES**: removing a feature must clear its persisted state entry

### T003
- **GOAL**: let users pause in-flight features and resume them explicitly
- **SCOPE**: add `kit pause`, selection flow, complete-phase guard, idempotence,
  and auto-unpause for explicit feature-scoped workflow commands
- **ACCEPTANCE**: paused features can be resumed by explicit work commands
  and by the canonical `kit resume` command
- **NOTES**: `kit status` still reports the highest-numbered feature even if it
  is paused

### T004
- **GOAL**: remove features safely and completely
- **SCOPE**: add `kit remove`, selector flow, confirmation, `--yes`, directory
  deletion, paused-state cleanup, and rollup regeneration
- **ACCEPTANCE**: deleted features disappear from disk and active lifecycle
  selectors while rollup regeneration records their removed state
- **NOTES**: do not rewrite arbitrary non-Kit markdown

### T005
- **GOAL**: make paused state visible in generated lifecycle views
- **SCOPE**: update rollup, status text/json, all-features tables, and
  active-only multi-feature flows
- **ACCEPTANCE**: paused state renders separately from phase and paused features
  are excluded from active-only flows other than the active-feature `status`
  view while `status --all` exposes the project overview
- **NOTES**: keep progress-summary ordering stable

### T006
- **GOAL**: verify lifecycle behavior and prevent regressions
- **SCOPE**: add focused tests for pause, remove, rollup, status, and
  auto-unpause flows; run targeted validation
- **ACCEPTANCE**: relevant tests pass and cover the new lifecycle controls
- **NOTES**: include remove confirmation, `--yes` cases, and resume-path coverage

### T007
- **GOAL**: make `kit rm` the concise removal command users can run directly
- **SCOPE**: expose `rm` as the primary command name, retain `remove` as a
  compatibility alias, and update user-facing docs
- **ACCEPTANCE**: `kit rm <feature> --yes` removes the feature directory and
  docs through the existing destructive removal flow, and `kit remove` remains
  callable
- **NOTES**: do not broaden removal beyond the target feature directory and
  Kit-managed lifecycle state

### T008
- **GOAL**: keep project history visible after feature docs are deleted
- **SCOPE**: persist removed-feature tombstones outside `docs/specs/`, render
  them as `removed` rows in `PROJECT_PROGRESS_SUMMARY.md`, and keep them out of
  active selectors/status views
- **ACCEPTANCE**: after `kit rm <feature> --yes`, the feature directory is gone,
  `.kit.yaml` contains removal metadata, and `PROJECT_PROGRESS_SUMMARY.md`
  retains the feature with `PHASE` set to `removed`
- **NOTES**: do not preserve the deleted feature docs themselves

### T009
- **GOAL**: keep removed features and retained notes visible after docs are
  deleted
- **SCOPE**: update `kit rm` output and interactive prompts, add `--notes`,
  retain `docs/notes/<feature>` by default, and include removed tombstones in
  `kit status --all`
- **ACCEPTANCE**: removed features show `REMOVED` in `kit status --all`, `kit rm`
  reports removed status and notes retention/removal, notes are kept by default,
  and notes are removed only through the interactive prompt or `--notes`
- **NOTES**: default retention preserves follow-up research context

## DEPENDENCIES

- T002 depends on the lifecycle contract captured in T001
- T003 and T004 depend on the shared state model from T002
- T005 depends on the final pause/remove behavior from T003 and T004
- T006 depends on the completed command and rendering changes from T003-T005
- T007 depends on the existing remove flow from T004
- T008 depends on the existing remove flow and primary `rm` surface
- T009 depends on removed tombstones and feature notes paths

## NOTES

- paused is a persisted flag, not a new workflow phase
- feature numbering continues to derive from existing directories
- `kit status` remains latest-feature-first even when the latest feature is
  paused
- removed feature notes are retained by default for follow-up work

<!-- REFLECTION_COMPLETE -->
