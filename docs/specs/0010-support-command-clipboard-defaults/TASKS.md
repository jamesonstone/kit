# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                                                  | STATUS | OWNER | DEPENDENCIES |
| ---- | --------------------------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Record support-command clipboard-default docs                         | done   | agent |              |
| T002 | Switch `handoff`, `summarize`, and `code-review` output flow          | done   | agent | T001         |
| T003 | Update README/help wording                                            | done   | agent | T002         |
| T004 | Run verification                                                      | done   | agent | T002, T003   |
| T005 | Narrow handoff scope now that prompt content has its own feature spec | done   | agent | T004         |

## TASK LIST

- [x] T001: Record support-command clipboard-default docs [PLAN-01]
- [x] T002: Switch `handoff`, `summarize`, and `code-review` output flow [PLAN-02]
- [x] T003: Update README/help wording [PLAN-03]
- [x] T004: Run verification [PLAN-04]
- [x] T005: Narrow handoff scope now that prompt content has its own feature spec [PLAN-01]

## TASK DETAILS

### T001

- **GOAL**: Record the approved behavior before code changes
- **SCOPE**:
  - add `docs/specs/0010-support-command-clipboard-defaults/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - the feature docs exist with the required sections
  - the behavior change is explicit and testable

### T002

- **GOAL**: Align the three support commands with the clipboard-first output contract
- **SCOPE**:
  - update `pkg/cli/handoff.go`
  - update `pkg/cli/summarize.go`
  - update `pkg/cli/code_review.go`
- **ACCEPTANCE**:
  - default command output acknowledges clipboard copy and does not print the output body
  - `--output-only` prints the raw output to stdout
  - `--output-only --copy` both prints and copies

### T003

- **GOAL**: Keep product messaging aligned with shipped behavior
- **SCOPE**:
  - update `README.md`
  - update command flag text as needed
- **ACCEPTANCE**:
  - README documents the full clipboard-first command set
  - help text reflects the new output contract

### T004

- **GOAL**: Verify the clipboard-default rollout without regressions
- **SCOPE**:
  - run tests, vet, and build
- **ACCEPTANCE**:
  - `go test ./...` passes
  - `make vet` passes
  - `make build` passes

### T005

- **GOAL**: Keep the clipboard-default feature docs accurate after handoff prompt content evolves
- **SCOPE**:
  - update `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - the feature docs no longer claim that handoff prompt content is frozen

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded contract.
- T003 depends on T002 because docs must reflect the shipped behavior.
- T004 depends on T002 and T003 because verification must validate the final surface.

## NOTES

- No additional notes.
