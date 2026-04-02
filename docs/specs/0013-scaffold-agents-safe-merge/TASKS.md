# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                     | STATUS | OWNER | DEPENDENCIES |
| ---- | ---------------------------------------- | ------ | ----- | ------------ |
| T001 | Record scaffold safety feature docs      | done   | agent |              |
| T002 | Add scaffold overwrite confirmation gate | done   | agent | T001         |
| T003 | Add append-only instruction merge mode   | done   | agent | T001         |
| T004 | Update docs and command help             | done   | agent | T002, T003   |
| T005 | Add tests for scaffold safety modes      | done   | agent | T002, T003   |
| T006 | Run verification                         | done   | agent | T004, T005   |

## TASK LIST

- [x] T001: Record scaffold safety feature docs [PLAN-01]
- [x] T002: Add scaffold overwrite confirmation gate [PLAN-02] [PLAN-04]
- [x] T003: Add append-only instruction merge mode [PLAN-02] [PLAN-03]
- [x] T004: Update docs and command help [PLAN-04] [PLAN-06]
- [x] T005: Add tests for scaffold safety modes [PLAN-05]
- [x] T006: Run verification [PLAN-06]

## TASK DETAILS

### T001

- **GOAL**: Capture the approved non-destructive scaffold behavior before code changes
- **SCOPE**:
  - add `docs/specs/0013-scaffold-agents-safe-merge/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - feature docs exist with complete sections
  - overwrite confirmation and append-only semantics are explicit

### T002

- **GOAL**: Make destructive scaffold overwrites explicit and confirmable
- **SCOPE**:
  - add overwrite confirmation for `--force`
  - add `--yes` bypass
  - validate mode combinations
- **ACCEPTANCE**:
  - `--force` prompts before overwriting existing files
  - cancellation leaves files unchanged
  - `--force --yes` overwrites without prompting

### T003

- **GOAL**: Add deterministic append-only merging for instruction files
- **SCOPE**:
  - add append-only merge planning
  - preserve matched sections and extra sections
  - fail safely on ambiguous or anchorless files
- **ACCEPTANCE**:
  - append-only inserts missing Kit-managed sections in template order
  - matched existing sections are unchanged
  - ambiguous merges fail before any writes occur

### T004

- **GOAL**: Keep user-facing docs aligned with the shipped scaffold behavior
- **SCOPE**:
  - update `README.md`
  - update `docs/specs/0000_INIT_PROJECT.md`
  - update command flag/help text where needed
- **ACCEPTANCE**:
  - docs describe `--force`, `--yes`, and `--append-only`
  - docs mention the safer append-only suggestion when files already exist

### T005

- **GOAL**: Prove the scaffold safety modes work and fail safely
- **SCOPE**:
  - update `pkg/cli/instruction_files_test.go`
  - add merge helper tests if needed
- **ACCEPTANCE**:
  - tests cover confirmation, append-only success, append-only failure, and flag validation

### T006

- **GOAL**: Verify the final scaffold safety feature without regression
- **SCOPE**:
  - run tests, vet, and build
- **ACCEPTANCE**:
  - `go test ./...` passes
  - `make vet` passes
  - `make build` passes

## DEPENDENCIES

- T002 depends on T001 because the overwrite confirmation semantics must be documented first.
- T003 depends on T001 because append-only merge rules must be specified before implementation.
- T004 depends on T002 and T003 because docs must match the shipped behavior.
- T005 depends on T002 and T003 because the tests must reflect the final safety modes.
- T006 depends on T004 and T005 because verification must validate the final surface.

## NOTES

- not required
