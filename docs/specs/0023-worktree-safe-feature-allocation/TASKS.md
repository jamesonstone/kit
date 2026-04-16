# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record the shared allocator and logical-ordering contract in feature docs | done | codex | none |
| T002 | Implement Git-common-dir feature-number reservation and wire it into feature creation | done | codex | T001 |
| T003 | Add duplicate-prefix detection to project validation and harden active-row comparison | done | codex | T001 |
| T004 | Order project-wide map output by dependency relationships without renumbering directories | done | codex | T001 |
| T005 | Add tests and run verification for allocator, duplicate detection, and logical ordering | done | codex | T002, T003, T004 |

## TASK LIST

- [x] T001: Record the shared allocator and logical-ordering contract in feature docs
- [x] T002: Implement Git-common-dir feature-number reservation and wire it into feature creation
- [x] T003: Add duplicate-prefix detection to project validation and harden active-row comparison
- [x] T004: Order project-wide map output by dependency relationships without renumbering directories
- [x] T005: Add tests and run verification for allocator, duplicate detection, and logical ordering

## TASK DETAILS

### T001
- **GOAL**: make the worktree-safe numbering model explicit before code changes
- **SCOPE**: record allocator behavior, fallback rules, duplicate detection, and logical ordering
- **ACCEPTANCE**: the feature contract explains why numbering stays numeric and how dependency order remains separate
- **NOTES**: reservation order and dependency order must remain distinct concepts

### T002
- **GOAL**: prevent duplicate numeric reservations across worktrees in the same clone
- **SCOPE**: add a shared allocator in the Git common dir and update feature creation to use it
- **ACCEPTANCE**: sequential reservations from separate worktrees in one clone cannot return the same number
- **NOTES**: gaps are acceptable; duplicate prefixes are not

### T003
- **GOAL**: make existing duplicate prefixes visible and avoid mislabeling active features
- **SCOPE**: add duplicate-prefix audits and compare active rows with a unique feature identity
- **ACCEPTANCE**: project validation fails on duplicate prefixes and `status --all` marks only the true active row
- **NOTES**: legacy duplicate repos should fail clearly rather than behave ambiguously

### T004
- **GOAL**: keep feature directories iterative while still presenting logical dependency order
- **SCOPE**: apply topological ordering to project-wide map output using `builds on` and `depends on`
- **ACCEPTANCE**: prerequisites appear before dependents in `kit map`, but directory names stay unchanged
- **NOTES**: `related to` stays informational and must not affect ordering

### T005
- **GOAL**: prove the allocator and ordering behavior is correct
- **SCOPE**: add focused tests and run targeted plus full verification
- **ACCEPTANCE**: allocator, duplicate detection, status hardening, and map ordering are covered by automated tests
- **NOTES**: include fallback behavior when no shared Git allocator is available

## DEPENDENCIES

- T002 depends on T001 because the shared allocator semantics must be recorded before implementation
- T003 depends on T001 because duplicate handling and active-row rules must follow the written contract
- T004 depends on T001 because logical ordering rules must be explicit before the renderer changes
- T005 depends on T002, T003, and T004 because verification must cover the shipped behavior

## NOTES

- directory numbering remains chronological or reservation-based
- dependency order is derived from `RELATIONSHIPS`, not the directory prefix
