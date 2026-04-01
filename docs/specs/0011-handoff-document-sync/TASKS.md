# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                                   | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------------------------------ | ------ | ----- | ------------ |
| T001 | Record handoff documentation-sync feature docs         | done   | agent |              |
| T002 | Adjust prior clipboard-default docs to remove conflict | done   | agent | T001         |
| T003 | Rewrite `kit handoff` prompt generation                | done   | agent | T001, T002   |
| T004 | Add handoff tests and update docs                      | done   | agent | T003         |
| T005 | Run verification                                       | done   | agent | T003, T004   |
| T006 | Verify dependency inventories during handoff           | done   | agent | T003, T004   |
| T007 | Add `--prompt-only` consistency flag to `handoff`      | done   | agent | T006         |

## TASK LIST

- [x] T001: Record handoff documentation-sync feature docs [PLAN-01]
- [x] T002: Adjust prior clipboard-default docs to remove conflict [PLAN-01]
- [x] T003: Rewrite `kit handoff` prompt generation [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T004: Add handoff tests and update docs [PLAN-06]
- [x] T005: Run verification [PLAN-06]
- [x] T006: Verify dependency inventories during handoff [PLAN-05] [PLAN-06]
- [x] T007: Add `--prompt-only` consistency flag to `handoff` [PLAN-07]

## TASK DETAILS

### T001

- **GOAL**: Capture the approved handoff prompt contract before changing code
- **SCOPE**:
  - add `docs/specs/0011-handoff-document-sync/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - the new feature docs exist with complete sections
  - the prompt contract is explicit and testable

### T002

- **GOAL**: Keep earlier support-command specs aligned with the new handoff contract
- **SCOPE**:
  - update `docs/specs/0010-support-command-clipboard-defaults/`
- **ACCEPTANCE**:
  - no existing spec claims that handoff content must remain unchanged

### T003

- **GOAL**: Make `kit handoff` drive documentation reconciliation before transfer
- **SCOPE**:
  - update `pkg/cli/handoff.go`
  - add `pkg/cli/handoff_prompt.go`
  - add documentation inventory table helpers
  - rewrite feature and project-wide prompts
- **ACCEPTANCE**:
  - the prompt tells the current agent session to update docs first when needed
  - the prompt includes the documentation inventory table and final response contract

### T004

- **GOAL**: Prove the new handoff prompt and docs stay aligned
- **SCOPE**:
  - add `pkg/cli/handoff_test.go`
  - update `README.md`
  - update `docs/specs/0000_INIT_PROJECT.md`
- **ACCEPTANCE**:
  - tests cover feature and project-wide prompt structure
  - docs describe the shipped handoff behavior

### T005

- **GOAL**: Verify the new handoff workflow without regression
- **SCOPE**:
  - run tests, vet, and build
- **ACCEPTANCE**:
  - `go test ./...` passes
  - `make vet` passes
  - `make build` passes

### T006

- **GOAL**: make handoff reconciliation include the phase dependency inventories in touched docs
- **SCOPE**:
  - update `pkg/cli/handoff_prompt.go`
  - update `pkg/cli/handoff_test.go`
- **ACCEPTANCE**:
  - handoff prompts tell the agent to refresh dependency inventories in touched `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`
  - the final response contract confirms documentation and dependency inventory sync

### T007

- **GOAL**: keep `handoff` aligned with the shared feature-prompt command surface
- **SCOPE**:
  - register `--prompt-only`
  - preserve the existing prompt-only runtime behavior
  - update docs/help expectations
- **ACCEPTANCE**:
  - `kit handoff --prompt-only` is accepted
  - the flag does not change the repo mutation model
  - docs/help reflect the shared command-surface contract

## DEPENDENCIES

- T002 depends on T001 because the new contract must be recorded before reconciling older specs.
- T003 depends on T001 and T002 because implementation must follow the updated spec set.
- T004 depends on T003 because docs and tests must reflect the shipped behavior.
- T005 depends on T003 and T004 because verification must validate the final surface.
- T007 depends on T006 because the shared prompt-only flag is layered on top of the shipped doc-sync prompt contract.

## NOTES

- No additional notes.
