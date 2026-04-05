# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add canonical feature docs for backlog capture and pickup | done | agent | |
| T002 | Add backlog classification and active-feature filtering helpers | done | agent | T001 |
| T003 | Implement `kit backlog` list and pickup flows | done | agent | T002 |
| T004 | Extend `kit brainstorm` with `--backlog` and `--pickup` | done | agent | T002 |
| T005 | Update status/help/docs for backlog semantics | done | agent | T002, T003, T004 |
| T006 | Add tests and run verification | done | agent | T002, T003, T004, T005 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Add canonical feature docs for backlog capture and pickup
- [x] T002: Add backlog classification and active-feature filtering helpers
- [x] T003: Implement `kit backlog` list and pickup flows
- [x] T004: Extend `kit brainstorm` with `--backlog` and `--pickup`
- [x] T005: Update status/help/docs for backlog semantics
- [x] T006: Add tests and run verification

## TASK DETAILS

### T001
- **GOAL**: record the approved backlog model before code changes
- **SCOPE**: create `SPEC.md`, `PLAN.md`, and `TASKS.md` for this feature
- **ACCEPTANCE**: the new feature directory documents requirements, approach,
  dependencies, and execution order
- **NOTES**: no additional information required

### T002
- **GOAL**: derive backlog membership without adding new persisted state
- **SCOPE**: add backlog-item helpers and active-feature filtering that skips
  deferred brainstorm items
- **ACCEPTANCE**: paused brainstorm items are identifiable and status can avoid
  treating them as active work
- **NOTES**: keep the logic small and explicit

### T003
- **GOAL**: ship a dedicated backlog list and pickup command
- **SCOPE**: add `kit backlog`, render the two-column table, support selector
  fallback, and reuse brainstorm prompt output on pickup
- **ACCEPTANCE**: backlog list and pickup work from the CLI with actionable
  errors for invalid pickup targets
- **NOTES**: list mode stays read-only

### T004
- **GOAL**: let users capture or resume deferred items directly from brainstorm
- **SCOPE**: add `--backlog` capture-only mode and `--pickup` resume mode
- **ACCEPTANCE**: deferred capture pauses brainstorm items and pickup clears
  paused state before outputting the planning prompt
- **NOTES**: `kit brainstorm --pickup` remains a compatibility path, while
  `kit backlog --pickup` and `kit resume` are the taught resume flows
- **NOTES**: keep existing vim-default free-text behavior for new captures

### T005
- **GOAL**: align product messaging with backlog semantics
- **SCOPE**: update status guidance, help ordering, README, and
  `docs/specs/0000_INIT_PROJECT.md`
- **ACCEPTANCE**: shipped docs describe backlog capture, backlog listing, and
  backlog pickup accurately, with `kit resume` as the canonical general resume
  command
- **NOTES**: keep `PROJECT_PROGRESS_SUMMARY.md` generated, not user-edited

### T006
- **GOAL**: verify behavior and guard against regression
- **SCOPE**: add focused tests and run `gofmt` plus `go test ./...`
- **ACCEPTANCE**: all relevant tests pass and new cases cover the changed flows
- **NOTES**: include status behavior when only backlog items exist

## DEPENDENCIES

- T002 depends on T001 because active-feature filtering must follow the approved
  backlog contract
- T003 and T004 depend on T002 because both command surfaces rely on the shared
  backlog helpers
- T005 depends on T003 and T004 because help and docs should reflect final
  command behavior
- T006 depends on all implementation tasks because verification targets the
  shipped behavior

## NOTES

- backlog items are deferred brainstorm-phase features, not a new artifact type
