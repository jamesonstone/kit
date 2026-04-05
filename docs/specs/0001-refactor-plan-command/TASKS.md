# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record the base `kit plan` command contract | done | agent | |
| T002 | Implement prerequisite checks and plan scaffolding | done | agent | T001 |
| T003 | Add plan-specific prompt guidance and feature selection | done | agent | T002 |
| T004 | Update project rollup behavior for the plan stage | done | agent | T002 |
| T005 | Verify the refactored planning workflow | done | agent | T003, T004 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record the base `kit plan` command contract [PLAN-01] [PLAN-03]
- [x] T002: Implement prerequisite checks and plan scaffolding [PLAN-01]
- [x] T003: Add plan-specific prompt guidance and feature selection [PLAN-02]
  [PLAN-05]
- [x] T004: Update project rollup behavior for the plan stage [PLAN-04]
- [x] T005: Verify the refactored planning workflow [PLAN-01] [PLAN-03]
  [PLAN-04] [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: define the dedicated `kit plan` stage before changing behavior
- **SCOPE**:
  - record the planning artifact contract
  - capture prerequisite and rollup expectations
- **ACCEPTANCE**:
  - `SPEC.md`, `PLAN.md`, and `TASKS.md` describe the same base workflow

### T002

- **GOAL**: make planning a first-class command with explicit prerequisites
- **SCOPE**:
  - create or open `PLAN.md`
  - enforce `SPEC.md` by default
  - support explicit out-of-order scaffolding
- **ACCEPTANCE**:
  - `kit plan <feature>` scaffolds `PLAN.md`
  - missing prerequisites fail clearly unless forced

### T003

- **GOAL**: make the planning step useful for both humans and coding agents
- **SCOPE**:
  - add feature selection for eligible features
  - add prompt guidance that reads `CONSTITUTION.md` and `SPEC.md` first
- **ACCEPTANCE**:
  - interactive mode selects only features ready for planning
  - prompt output keeps the agent focused on `PLAN.md`

### T004

- **GOAL**: keep repository state aligned after planning
- **SCOPE**:
  - regenerate `PROJECT_PROGRESS_SUMMARY.md`
  - keep the highest completed artifact current
- **ACCEPTANCE**:
  - successful plan creation refreshes the rollup view

### T005

- **GOAL**: prove the refactored planning workflow is stable
- **SCOPE**:
  - validate prerequisite handling
  - validate template structure
  - validate prompt and rollup behavior
- **ACCEPTANCE**:
  - command checks cover success and prerequisite failure paths

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded contract.
- T003 and T004 depend on T002 because they refine the shipped planning flow.
- T005 depends on T003 and T004 because verification covers the final surface.

## NOTES

- This feature records the base `kit plan` refactor.
- Later feature specs may extend prompt behavior, dependency inventories, or
  lifecycle rules without changing this feature's core intent.
