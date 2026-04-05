# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                              | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Record catchup feature artifacts                  | done   | agent |              |
| T002 | Implement catchup command and prompt              | done   | agent | T001         |
| T003 | Update help/docs surfaces                         | done   | agent | T002         |
| T004 | Add tests and run verification                    | done   | agent | T002, T003   |
| T005 | Switch `catchup` to clipboard-first prompt output | done   | agent | T004         |
| T006 | Add `--prompt-only` consistency flag to `catchup` | done   | agent | T005         |

## TASK LIST

- [x] T001: Record catchup feature artifacts [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05] [PLAN-06]
- [x] T002: Implement catchup command and prompt [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04]
- [x] T003: Update help/docs surfaces [PLAN-05]
- [x] T004: Add tests and run verification [PLAN-06]
- [x] T005: Switch `catchup` to clipboard-first prompt output [PLAN-07]
- [x] T006: Add `--prompt-only` consistency flag to `catchup` [PLAN-08]

## TASK DETAILS

### T001

- **GOAL**: Capture the feature contract before implementation
- **SCOPE**:
  - create `SPEC.md`
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - the new feature directory exists under `docs/specs/0007-catchup-command/`
  - the docs define how `catchup` differs from `handoff`, `summarize`, and `implement`

### T002

- **GOAL**: Add the `kit catchup` command and prompt generation logic
- **SCOPE**:
  - create `pkg/cli/catchup.go`
  - create `pkg/cli/catchup_prompt.go`
  - add selector, feature resolution, stage/state derivation, and prompt output
- **ACCEPTANCE**:
  - `kit catchup` works with selector or direct feature argument
  - prompt keeps the agent in plan mode and asks for explicit permission before implementation
  - complete features are handled as review/reopen triage

### T003

- **GOAL**: Expose the command cleanly in product help and docs
- **SCOPE**:
  - update `pkg/cli/root.go`
  - update `README.md`
- **ACCEPTANCE**:
  - `catchup` remains callable for compatibility
  - canonical help and README teach `resume` as the general resume command
  - deprecated `catchup` guidance stays distinct from `handoff` and `implement`

### T004

- **GOAL**: Prevent regression and verify the new prompt contract
- **SCOPE**:
  - add `pkg/cli/catchup_test.go`
  - run vet, tests, build, and help checks
- **ACCEPTANCE**:
  - prompt tests cover stage/state, planning-mode gating, and complete-phase wording
  - verification commands pass cleanly

### T005

- **GOAL**: Align `catchup` output with the clipboard-first prompt contract
- **SCOPE**:
  - update `pkg/cli/catchup.go`
  - keep raw stdout prompt output behind `--output-only`
  - preserve `--copy` as an explicit override for `--output-only`
- **ACCEPTANCE**:
  - default command output acknowledges clipboard copy and does not print the prompt body
  - `--output-only` prints the raw prompt to stdout
  - `--output-only --copy` both prints and copies
  - verification commands pass cleanly

### T006

- **GOAL**: keep `catchup` aligned with the shared feature-prompt command surface
- **SCOPE**:
  - register `--prompt-only`
  - preserve the existing prompt-only runtime behavior
  - cover the flag in help/docs verification
- **ACCEPTANCE**:
  - `kit catchup --prompt-only` is accepted
  - the flag does not change the repo mutation model
  - docs/help reflect the shared command-surface contract while keeping
    `catchup` as a hidden deprecated compatibility path

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded contract.
- T003 depends on T002 because docs must describe the final command surface.
- T004 depends on T002 and T003 because verification must validate the shipped behavior.
- T005 depends on T004 because the clipboard-first amendment follows the shipped prompt contract.
- T006 depends on T005 because the shared prompt-only flag sits on top of the shipped prompt-output surface.

## NOTES

- No additional notes.
