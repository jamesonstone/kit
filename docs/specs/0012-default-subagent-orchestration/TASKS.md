# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                                        | STATUS | OWNER | DEPENDENCIES |
| ---- | ----------------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Record default subagent orchestration feature docs          | done   | agent |              |
| T002 | Flip the shared orchestration helper to subagents-by-default | done   | agent | T001         |
| T003 | Update README and help-facing wording                       | done   | agent | T002         |
| T004 | Add regression tests and run verification                   | done   | agent | T002, T003   |

## TASK LIST

- [x] T001: Record default subagent orchestration feature docs [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04]
- [x] T002: Flip the shared orchestration helper to subagents-by-default [PLAN-01] [PLAN-02] [PLAN-03]
- [x] T003: Update README and help-facing wording [PLAN-04]
- [x] T004: Add regression tests and run verification [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: capture the approved cross-cutting orchestration change before code edits
- **SCOPE**:
  - add `docs/specs/0012-default-subagent-orchestration/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - the new feature docs exist with the required sections
  - the default orchestration contract is explicit and testable

### T002

- **GOAL**: make shared prompt outputs default to subagent orchestration
- **SCOPE**:
  - update `pkg/cli/subagents.go`
  - add `--single-agent`
  - keep `dispatch` on the no-shared-suffix path
- **ACCEPTANCE**:
  - shared prompt outputs include `## Subagent Orchestration` by default
  - `--single-agent` disables that suffix
  - dispatch output remains unchanged

### T003

- **GOAL**: align product messaging with the shipped flag and prompt behavior
- **SCOPE**:
  - update `README.md`
- **ACCEPTANCE**:
  - README describes subagents as the default
  - README documents `--single-agent` as the opt-out

### T004

- **GOAL**: prevent regression and validate the new default
- **SCOPE**:
  - update shared prompt tests
  - keep dispatch isolation covered
  - run tests, vet, and build
- **ACCEPTANCE**:
  - tests cover default suffix behavior, `--single-agent`, and dispatch isolation
  - `go test ./...`, `make vet`, and `make build` pass

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded contract.
- T003 depends on T002 because docs must describe the shipped behavior.
- T004 depends on T002 and T003 because verification must validate the final surface.

## NOTES

- `kit dispatch` remains the stricter approval-gated orchestration planner even though normal prompts now default to subagent guidance.
