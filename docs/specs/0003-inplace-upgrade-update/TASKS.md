# TASKS

## PROGRESS TABLE

| ID   | TASK                               | STATUS | OWNER | DEPENDENCIES |
| ---- | ---------------------------------- | ------ | ----- | ------------ |
| T001 | Create plan and task artifacts     | done   | agent |              |
| T002 | Implement upgrade command behavior | done   | agent | T001         |
| T003 | Update help ordering and README    | done   | agent | T002         |
| T004 | Add tests and run verification     | done   | agent | T002, T003   |

## TASK LIST

- [x] T001: Create plan and task artifacts [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05] [PLAN-06]
- [x] T002: Implement upgrade command behavior [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T003: Update help ordering and README [PLAN-06]
- [x] T004: Add tests and run verification [PLAN-06]

## TASK DETAILS

### T001

- **GOAL**: Record the implementation approach before code changes
- **SCOPE**:
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - both files exist and trace the requested feature work

### T002

- **GOAL**: Add safe in-place self-update behavior
- **SCOPE**:
  - create `pkg/cli/upgrade.go`
  - register `upgrade` and `update`
  - fetch latest stable release
  - compare versions
  - select asset and verify checksum
  - extract and replace executable safely
- **ACCEPTANCE**:
  - both command names route to the same behavior
  - `dev` builds can upgrade
  - no failure path leaves a broken executable

### T003

- **GOAL**: Expose the new command in help and docs
- **SCOPE**:
  - update `pkg/cli/root.go`
  - update `README.md`
- **ACCEPTANCE**:
  - `kit --help` lists both `upgrade` and `update`
  - README utility table documents both entries

### T004

- **GOAL**: Prevent regression and verify the feature end to end
- **SCOPE**:
  - add `pkg/cli/upgrade_test.go`
  - run vet, tests, build, and help checks
- **ACCEPTANCE**:
  - helper behavior is covered with unit tests
  - required verification commands pass cleanly

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded plan.
- T003 depends on T002 because docs must describe the shipped command surface.
- T004 depends on T002 and T003 because verification must validate the final implementation and docs.
