---
kit_metadata_version: 1
artifact: "tasks"
feature:
  id: "0029"
  slug: "scaffold-workflows-prepare"
  dir: "0029-scaffold-workflows-prepare"
relationships:
  - type: "builds_on"
    target: "0013-scaffold-agents-safe-merge"
  - type: "builds_on"
    target: "0019-command-surface-simplification"
  - type: "related_to"
    target: "0004-brainstorm-first-workflow"
---
# TASKS

## PROGRESS TABLE

| ID   | TASK                                      | STATUS | OWNER | DEPENDENCIES |
| ---- | ----------------------------------------- | ------ | ----- | ------------ |
| T001 | Record scaffold workflow feature docs     | done   | agent |              |
| T002 | Add brainstorm prepare mode               | done   | agent | T001         |
| T003 | Convert scaffold into workflow namespace  | done   | agent | T001, T002   |
| T004 | Move agent scaffolding under scaffold     | done   | agent | T003         |
| T005 | Update docs and internal command guidance | done   | agent | T002, T004   |
| T006 | Add tests                                 | done   | agent | T002, T004   |
| T007 | Run rollup and verification               | done   | agent | T005, T006   |

## TASK LIST

- [x] T001: Record scaffold workflow feature docs [PLAN-01]
- [x] T002: Add brainstorm prepare mode [PLAN-01]
- [x] T003: Convert scaffold into workflow namespace [PLAN-02]
- [x] T004: Move agent scaffolding under scaffold [PLAN-03]
- [x] T005: Update docs and internal command guidance [PLAN-04]
- [x] T006: Add tests [PLAN-05]
- [x] T007: Run rollup and verification [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: Capture the requested scaffold command semantics before implementation
- **SCOPE**:
  - add `docs/specs/0029-scaffold-workflows-prepare/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - feature docs define prepare mode, scaffold namespace, and agent command move

### T002

- **GOAL**: Let users create brainstorm folders and files before the brainstorm prompt starts
- **SCOPE**:
  - add `--prepare` to `kit brainstorm`
  - create feature directory, notes directory, optional design directories, and `BRAINSTORM.md`
  - skip thesis capture and prompt output
- **ACCEPTANCE**:
  - prepare mode is idempotent
  - prepare mode rejects prompt-only and prompt-output flags

### T003

- **GOAL**: Make `kit scaffold` mean one-workflow filesystem preparation
- **SCOPE**:
  - replace the old hidden full-pipeline scaffold behavior
  - add workflow subcommands for brainstorm, spec, plan, and tasks
  - use existing templates and rollup behavior
- **ACCEPTANCE**:
  - scaffold subcommands create only the intended phase scaffold
  - output uses the requested recycle wording

### T004

- **GOAL**: Make `kit scaffold agents` the canonical instruction-file scaffold command
- **SCOPE**:
  - register the existing scaffold-agents command under `scaffold agents`
  - stop registering root `scaffold-agents`
  - preserve existing flags and safe write behavior
- **ACCEPTANCE**:
  - `kit scaffold agents --help` shows the existing version table and flags
  - root `kit scaffold-agents` is no longer the visible/canonical command

### T005

- **GOAL**: Keep user-facing and internal guidance aligned with the new command surface
- **SCOPE**:
  - update README
  - update `docs/specs/0000_INIT_PROJECT.md`
  - update `docs/CONSTITUTION.md`
  - update internal reconciliation and refresh-prompt guidance
- **ACCEPTANCE**:
  - docs use `kit scaffold agents`
  - docs describe `kit brainstorm --prepare`

### T006

- **GOAL**: Prove the new scaffold command surface works and old assumptions are removed
- **SCOPE**:
  - add or update CLI tests
  - update root help tests
  - update instruction scaffold tests
- **ACCEPTANCE**:
  - tests fail if scaffold subcommands disappear
  - tests fail if old root help still teaches `scaffold-agents`

### T007

- **GOAL**: Verify implementation and refresh project summary
- **SCOPE**:
  - run rollup
  - run Go tests, vet, build, and Kit document checks
- **ACCEPTANCE**:
  - verification commands pass or failures are reported with root cause

## DEPENDENCIES

- T002 depends on T001 because prepare behavior should follow the documented contract.
- T003 depends on T001 and T002 because `scaffold brainstorm` reuses prepare mode.
- T004 depends on T003 because agent scaffolding moves under the scaffold namespace.
- T005 depends on T002 and T004 because docs describe the shipped command surface.
- T006 depends on T002 and T004 because tests cover the final command routing.
- T007 depends on T005 and T006 because verification runs after docs and tests are current.

## NOTES

- not required
