# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record reconcile feature artifacts | done | agent | |
| T002 | Implement reconcile command and audit helpers | done | agent | T001 |
| T003 | Build reconciliation prompt and clean-result flow | done | agent | T002 |
| T004 | Update help and README surfaces | done | agent | T003 |
| T005 | Add tests and run verification | done | agent | T002, T003, T004 |
| T006 | Compress reconcile prompt and add terminal summary | done | agent | T005 |
| T007 | Add explicit subagent overlap instruction to the default prompt | done | agent | T006 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record reconcile feature artifacts [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05] [PLAN-06]
- [x] T002: Implement reconcile command and audit helpers [PLAN-01] [PLAN-02] [PLAN-03]
- [x] T003: Build reconciliation prompt and clean-result flow [PLAN-04]
- [x] T004: Update help and README surfaces [PLAN-05]
- [x] T005: Add tests and run verification [PLAN-06]
- [x] T006: Compress reconcile prompt and add terminal summary [PLAN-04] [PLAN-06] [PLAN-07]
- [x] T007: Add explicit subagent overlap instruction to the default prompt [PLAN-04] [PLAN-08]

## TASK DETAILS

### T001
- **GOAL**: record the approved contract before code changes
- **SCOPE**:
  - create `SPEC.md`
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - `docs/specs/0017-reconcile-command/` exists
  - the docs define how `reconcile` differs from `check`, `catchup`, `handoff`, and `scaffold-agents`

### T002
- **GOAL**: add repo-wide and feature-scoped reconciliation audits
- **SCOPE**:
  - create `pkg/cli/reconcile.go`
  - create audit helpers for required docs, sections, tables, task alignment, rollup drift, and instruction-file drift
  - keep the command read-only
- **ACCEPTANCE**:
  - `kit reconcile` audits the whole project
  - `kit reconcile <feature>` audits the selected feature
  - findings include severity and canonical source information

### T003
- **GOAL**: generate an actionable reconciliation prompt only when the audit finds issues
- **SCOPE**:
  - create prompt builder helpers
  - emit clean success when no findings exist
  - include exact search guidance and fixed response sections
- **ACCEPTANCE**:
  - prompts start with `/plan`
  - prompts forbid unrelated code changes
  - clean projects do not emit prompt bodies

### T004
- **GOAL**: expose the command clearly in shipped help and docs
- **SCOPE**:
  - update `pkg/cli/root.go`
  - update `README.md`
- **ACCEPTANCE**:
  - root help shows `reconcile`
  - README explains when to use `reconcile` instead of adjacent commands

### T005
- **GOAL**: prevent regression and verify the shipped command contract
- **SCOPE**:
  - add focused tests
  - run vet, test, build, and help checks
- **ACCEPTANCE**:
  - tests cover clean and dirty audit paths
  - verification commands pass cleanly

### T006
- **GOAL**: make reconcile output materially shorter and easier to scan
- **SCOPE**:
  - compress repo-wide prompt rendering by grouping findings by file
  - deduplicate search guidance
  - reduce the response contract
  - add a compact human-readable terminal summary for non-`--output-only` runs
- **ACCEPTANCE**:
  - repo-wide prompt no longer renders one full paragraph block per finding
  - terminal runs show a compact graphical summary without changing the raw prompt payload
  - tests cover the shorter prompt contract and terminal summary rendering

### T007
- **GOAL**: make the default reconcile prompt explicitly tell the coding agent how to orchestrate overlapping documentation work
- **SCOPE**:
  - add an explicit subagent overlap instruction to the default raw prompt
  - omit that line when `--single-agent` is set
  - cover the conditional behavior in tests
- **ACCEPTANCE**:
  - default prompt says to use subagents and queue work according to overlapping file changes
  - `--single-agent` omits that instruction from the raw prompt

## DEPENDENCIES

- T002 depends on T001 because implementation follows the recorded contract
- T003 depends on T002 because the prompt needs the final finding model
- T004 depends on T003 because docs should describe the shipped command shape
- T005 depends on T002, T003, and T004 because verification must validate the final surface
- T006 depends on T005 because prompt compression refines the shipped surface after the first implementation pass
- T007 depends on T006 because the explicit orchestration line refines the compact prompt contract

## NOTES

- v1 is intentionally prompt-only and documentation-scoped
