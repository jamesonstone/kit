# TASKS

## PROGRESS TABLE

| ID   | TASK                                       | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------------------ | ------ | ----- | ------------ |
| T001 | Record dispatch feature artifacts          | done   | agent |              |
| T002 | Implement dispatch command and prompt      | done   | agent | T001         |
| T003 | Update help and README surfaces            | done   | agent | T002         |
| T004 | Add tests and run verification             | done   | agent | T002, T003   |
| T005 | Add pre-editor instructions and keypress gating | done | agent | T002, T004 |

## TASK LIST

- [x] T001: Record dispatch feature artifacts [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T002: Implement dispatch command and prompt [PLAN-01] [PLAN-02] [PLAN-03]
- [x] T003: Update help and README surfaces [PLAN-04]
- [x] T004: Add tests and run verification [PLAN-05]
- [x] T005: Add pre-editor instructions and keypress gating [PLAN-01] [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: Capture the dispatch command contract before code changes
- **SCOPE**:
  - create `SPEC.md`
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - the new feature directory exists under `docs/specs/0008-dispatch-command/`
  - the docs record the approved `kit dispatch` behavior

### T002

- **GOAL**: Add the `kit dispatch` command, task normalization, and prompt-generation logic
- **SCOPE**:
  - create `pkg/cli/dispatch.go`
  - create `pkg/cli/dispatch_input.go`
  - create `pkg/cli/dispatch_tasks.go`
  - create `pkg/cli/dispatch_prompt.go`
- **ACCEPTANCE**:
  - `kit dispatch` supports file, stdin, and default editor-backed input
  - task normalization preserves top-level task boundaries and nested detail correctly
  - the generated prompt enforces dry-run discovery, overlap clustering, and approval gating

### T003

- **GOAL**: Expose the new command in product help and docs
- **SCOPE**:
  - update `pkg/cli/root.go`
  - update `README.md`
- **ACCEPTANCE**:
  - help output includes `dispatch`
  - README documents the new command clearly

### T004

- **GOAL**: Prevent regression and verify the new prompt contract
- **SCOPE**:
  - add `pkg/cli/dispatch_test.go`
  - run vet, tests, build, and help checks
- **ACCEPTANCE**:
  - tests cover prompt requirements, normalization, precedence, and validation
  - verification commands pass cleanly

### T005

- **GOAL**: Make default editor-backed capture clearer before the editor opens
- **SCOPE**:
  - update the shared editor input helper used by `dispatch`
  - show a short instruction screen before the editor opens
  - wait for any key before launching the editor
- **ACCEPTANCE**:
  - naked `kit dispatch` shows instructions before the editor opens
  - the editor opens only after an explicit key press
  - tests cover the new pre-editor interaction

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded contract.
- T003 depends on T002 because docs must describe the final command surface.
- T004 depends on T002 and T003 because verification must validate the shipped behavior.
