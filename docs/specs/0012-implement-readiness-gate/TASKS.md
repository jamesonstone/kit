# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                               | STATUS | OWNER | DEPENDENCIES |
| ---- | -------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Record implementation-readiness gate feature docs  | done   | agent |              |
| T002 | Update canonical workflow docs                     | done   | agent | T001         |
| T003 | Rewrite `kit implement` with readiness gate prompt | done   | agent | T001, T002   |
| T004 | Update status guidance and add tests               | done   | agent | T003         |
| T005 | Run verification                                   | done   | agent | T003, T004   |
| T006 | Align scaffolded instruction templates             | done   | agent | T002         |

## TASK LIST

- [x] T001: Record implementation-readiness gate feature docs [PLAN-01]
- [x] T002: Update canonical workflow docs [PLAN-03]
- [x] T003: Rewrite `kit implement` with readiness gate prompt [PLAN-02] [PLAN-04]
- [x] T004: Update status guidance and add tests [PLAN-05]
- [x] T005: Run verification [PLAN-05]
- [x] T006: Align scaffolded instruction templates [PLAN-03] [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: Capture the approved readiness-gate contract before changing workflow behavior
- **SCOPE**:
  - add `docs/specs/0012-implement-readiness-gate/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - the new feature docs exist with complete sections
  - the gate contract is explicit and testable

### T002

- **GOAL**: Keep the canonical workflow docs aligned with the shipped readiness gate
- **SCOPE**:
  - update `README.md`
  - update `docs/CONSTITUTION.md`
  - update `docs/specs/0000_INIT_PROJECT.md`
- **ACCEPTANCE**:
  - docs describe the readiness gate without adding a new phase
  - wording is consistent across the product surface

### T003

- **GOAL**: Make `kit implement` challenge docs before coding begins
- **SCOPE**:
  - update `pkg/cli/implement.go`
  - preserve the command surface and clipboard behavior
  - keep `kit check` unchanged
- **ACCEPTANCE**:
  - the prompt starts with the implementation-readiness gate
  - failure requires updating canonical docs before coding
  - passing the gate leads to the first incomplete task

### T004

- **GOAL**: Prove the readiness gate is visible and consistent where users need it
- **SCOPE**:
  - update `pkg/cli/status.go` wording if needed
  - add `pkg/cli/implement_test.go`
  - update `pkg/cli/status_test.go`
- **ACCEPTANCE**:
  - tests cover the readiness-gate prompt contract
  - status wording stays phase-based while mentioning the gate

### T005

- **GOAL**: Verify the new workflow semantics without regression
- **SCOPE**:
  - run tests, vet, and build
- **ACCEPTANCE**:
  - `go test ./...` passes
  - `make vet` passes
  - `make build` passes

### T006

- **GOAL**: Keep scaffolded repository instruction files aligned with the shipped readiness-gate workflow
- **SCOPE**:
  - update `internal/templates/templates.go`
  - update `internal/templates/templates_test.go`
- **ACCEPTANCE**:
  - scaffolded `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` include readiness-gate guidance
  - template tests catch future drift

## DEPENDENCIES

- T002 depends on T001 because canonical docs should follow the approved feature contract.
- T003 depends on T001 and T002 because the prompt should implement the documented semantics.
- T004 depends on T003 because tests and status wording must reflect the shipped prompt behavior.
- T005 depends on T003 and T004 because verification must validate the final surface.
- T006 depends on T002 because scaffolded instruction files must stay aligned with the documented workflow contract.

## NOTES

- not required
