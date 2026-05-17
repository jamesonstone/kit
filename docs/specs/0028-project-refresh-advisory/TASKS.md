---
kit_metadata_version: 1
artifact: "tasks"
feature:
  id: "0028"
  slug: "project-refresh-advisory"
  dir: "0028-project-refresh-advisory"
relationships:
  - type: "builds_on"
    target: "0017-reconcile-command"
  - type: "builds_on"
    target: "0025-v0-prompt-library"
  - type: "related_to"
    target: "0027-implement-readiness-gate"
---
# TASKS

## PROGRESS TABLE

| ID   | TASK                                             | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------------------------ | ------ | ----- | ------------ |
| T001 | Record project refresh feature docs             | done   | agent |              |
| T002 | Add `project refresh` built-in prompt           | done   | agent | T001         |
| T003 | Add soft project-refresh advisory gate          | done   | agent | T001, T002   |
| T004 | Update user-facing docs                         | done   | agent | T001, T002   |
| T005 | Add focused tests                               | done   | agent | T002, T003   |
| T006 | Run rollup and verification                     | done   | agent | T004, T005   |

## TASK LIST

- [x] T001: Record project refresh feature docs [PLAN-01]
- [x] T002: Add `project refresh` built-in prompt [PLAN-01]
- [x] T003: Add soft project-refresh advisory gate [PLAN-02] [PLAN-03]
- [x] T004: Update user-facing docs [PLAN-04]
- [x] T005: Add focused tests [PLAN-05]
- [x] T006: Run rollup and verification [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: Capture the approved manual-plus-soft-gate behavior before changing code
- **SCOPE**:
  - add `docs/specs/0028-project-refresh-advisory/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - docs define the manual prompt and advisory behavior
  - docs explicitly keep the advisory soft and prompt-only

### T002

- **GOAL**: Expose a manual project refresh prompt through the existing prompt library
- **SCOPE**:
  - register `project refresh`
  - add a prompt builder that requires Kit project context
  - keep `kit init` unchanged
- **ACCEPTANCE**:
  - `kit prompt project refresh --output-only` emits a docs-only prompt
  - prompt starts with `/plan`
  - prompt distinguishes refresh from `kit reconcile --all`

### T003

- **GOAL**: Make late workflow steps ask whether project-level truth changed without blocking progress
- **SCOPE**:
  - add a reflection prompt step
  - print completion advisory after successful feature completion
- **ACCEPTANCE**:
  - advisory points to `kit prompt project refresh`
  - advisory does not alter lifecycle state
  - advisory does not make commands fail

### T004

- **GOAL**: Teach the new refresh surface where users discover setup and prompt utilities
- **SCOPE**:
  - update `README.md`
  - update `docs/specs/0000_INIT_PROJECT.md`
- **ACCEPTANCE**:
  - docs describe `kit prompt project refresh`
  - docs explain that late workflow advisory is soft

### T005

- **GOAL**: Cover the new prompt and advisory contract with focused tests
- **SCOPE**:
  - update built-in prompt tests
  - update reflection golden prompt
  - update completion tests
- **ACCEPTANCE**:
  - tests fail if `project refresh` disappears
  - tests fail if reflection loses the advisory
  - tests fail if completion stops printing the advisory

### T006

- **GOAL**: Refresh generated project summary and verify the implementation
- **SCOPE**:
  - run `kit rollup`
  - run relevant Go tests
  - run normal verification commands when feasible
- **ACCEPTANCE**:
  - `docs/PROJECT_PROGRESS_SUMMARY.md` includes feature 0028
  - validation commands pass or failures are reported with root cause

## DEPENDENCIES

- T002 depends on T001 because code should follow the approved feature contract.
- T003 depends on T001 and T002 because advisory text points to the new manual prompt.
- T004 depends on T001 and T002 because docs should describe the shipped prompt identity.
- T005 depends on T002 and T003 because tests cover the prompt and advisory behavior.
- T006 depends on T004 and T005 because verification should run after docs and tests are current.

## NOTES

- The advisory is intentionally soft because semantic project freshness cannot be detected reliably without noisy false positives.
