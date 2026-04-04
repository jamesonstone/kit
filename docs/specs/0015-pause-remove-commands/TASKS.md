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

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Define pause/remove lifecycle contract in canonical docs
- [x] T002: Add persisted pause state to config and feature models
- [x] T003: Implement `kit pause` and explicit-resume auto-unpause behavior
- [x] T004: Implement `kit remove` with confirmation, deletion, and state cleanup
- [x] T005: Update rollup, status, and active-only flows for paused state
- [x] T006: Add regression tests and verify command behavior

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
  without a dedicated resume command
- **NOTES**: `kit status` still reports the highest-numbered feature even if it
  is paused

### T004
- **GOAL**: remove features safely and completely
- **SCOPE**: add `kit remove`, selector flow, confirmation, `--yes`, directory
  deletion, paused-state cleanup, and rollup regeneration
- **ACCEPTANCE**: deleted features disappear from disk and from Kit-managed
  lifecycle views
- **NOTES**: do not rewrite arbitrary non-Kit markdown

### T005
- **GOAL**: make paused state visible in generated lifecycle views
- **SCOPE**: update rollup, status text/json, all-features tables, and
  active-only multi-feature flows
- **ACCEPTANCE**: paused state renders separately from phase and paused features
  are excluded from active-only flows other than `status`
- **NOTES**: keep progress-summary ordering stable

### T006
- **GOAL**: verify lifecycle behavior and prevent regressions
- **SCOPE**: add focused tests for pause, remove, rollup, status, and
  auto-unpause flows; run targeted validation
- **ACCEPTANCE**: relevant tests pass and cover the new lifecycle controls
- **NOTES**: include remove confirmation and `--yes` cases

## DEPENDENCIES

- T002 depends on the lifecycle contract captured in T001
- T003 and T004 depend on the shared state model from T002
- T005 depends on the final pause/remove behavior from T003 and T004
- T006 depends on the completed command and rendering changes from T003-T005

## NOTES

- paused is a persisted flag, not a new workflow phase
- feature numbering continues to derive from existing directories
- `kit status` remains latest-feature-first even when the latest feature is
  paused
