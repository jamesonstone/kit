# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                            | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------- | ------ | ----- | ------------ |
| T001 | Add version command spec docs   | done   | agent |              |
| T002 | Implement `kit version` command | done   | agent | T001         |
| T003 | Update help and README surfaces | done   | agent | T002         |
| T004 | Add tests and run verification  | done   | agent | T002, T003   |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Add version command spec docs
- [x] T002: Implement `kit version` command
- [x] T003: Update help and README surfaces
- [x] T004: Add tests and run verification

## TASK DETAILS

### T001

- **GOAL**: Record the feature requirements and implementation plan
- **SCOPE**:
  - add `docs/specs/0005-version-command/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - all three files exist with required sections
  - requirements are explicit enough to implement directly
- **NOTES**: complete when files exist in the working tree

### T002

- **GOAL**: Add a root `version` command
- **SCOPE**:
  - create `pkg/cli/version.go`
  - print the current `Version` value to stdout
- **ACCEPTANCE**:
  - `kit version` exits successfully and prints the installed version
  - no args are accepted
- **NOTES**: use the existing linker-injected version source of truth

### T003

- **GOAL**: Expose the command in user-facing help and docs
- **SCOPE**:
  - update root command ordering
  - update local build metadata defaults if needed
  - update README command documentation
- **ACCEPTANCE**:
  - root help includes `version`
  - local Makefile builds do not report a stale hard-coded version
  - README lists `kit version`
- **NOTES**: keep wording short and script-friendly

### T004

- **GOAL**: Prevent regression and validate behavior
- **SCOPE**:
  - add tests for command output
  - run `go test ./...`
- **ACCEPTANCE**:
  - tests pass
  - command behavior is covered by automated tests
- **NOTES**: restore global version state in tests

## DEPENDENCIES

- T002 depends on T001 for the formal feature record.
- T003 and T004 depend on the command implementation.

## NOTES

- `kit version` should remain a thin wrapper around existing build metadata.
