# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record the shared instruction-registry and project-validation contract | done | agent | |
| T002 | Add the shared instruction registry and migrate current consumers | done | agent | T001 |
| T003 | Extend `kit check` with project-scoped validation | done | agent | T002 |
| T004 | Tighten shipped subagent guidance and update supporting docs | done | agent | T002 |
| T005 | Add tests and run verification | done | agent | T002, T003, T004 |

## TASK LIST

- [x] T001: Record the shared instruction-registry and project-validation contract [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T002: Add the shared instruction registry and migrate current consumers [PLAN-01] [PLAN-02]
- [x] T003: Extend `kit check` with project-scoped validation [PLAN-03]
- [x] T004: Tighten shipped subagent guidance and update supporting docs [PLAN-04]
- [x] T005: Add tests and run verification [PLAN-05]

## TASK DETAILS

### T001
- **GOAL**: capture the contract before code changes
- **SCOPE**:
  - create `SPEC.md`
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - `docs/specs/0021-project-validation-and-instruction-registry/` exists
  - the docs define the shared registry, project validation, and subagent guidance goals

### T002
- **GOAL**: remove duplicated instruction-model metadata from current consumers
- **SCOPE**:
  - add `internal/instructions`
  - migrate detection and repo-doc inventory helpers
  - update template support-file metadata to use the shared registry
- **ACCEPTANCE**:
  - current consumers no longer maintain their own `v2` docs-tree path sets
  - map, prompt helpers, and version detection agree on one source of truth

### T003
- **GOAL**: add a direct project validator for repo-level contract drift
- **SCOPE**:
  - add `kit check --project`
  - reuse repo-audit findings
  - keep feature-scoped check behavior stable
- **ACCEPTANCE**:
  - project validation reports concise findings and exits non-zero on any repo-level drift
  - feature-scoped validation still works

### T004
- **GOAL**: make the shipped subagent contract more precise
- **SCOPE**:
  - update shared subagent suffix wording
  - update repo-local docs that explain subagent use
- **ACCEPTANCE**:
  - RLM is described as discovery
  - dispatch/subagents are described as execution planning
  - the main agent remains responsible for synthesis and integration

### T005
- **GOAL**: prevent regression and verify the new contract
- **SCOPE**:
  - add registry tests
  - add `check --project` tests
  - run targeted and full verification
- **ACCEPTANCE**:
  - tests cover registry-backed detection and project validation
  - verification commands pass

## DEPENDENCIES

- T002 depends on T001 because the registry contract must be explicit before code changes
- T003 depends on T002 because project validation should use the shared registry-backed audit inputs
- T004 depends on T002 because repo-local subagent docs should use the centralized instruction model
- T005 depends on T002, T003, and T004 because tests must validate the final behavior

## NOTES

- `kit reconcile` remains the prompt-oriented migration surface
- this feature adds a direct validator, not a new migration engine
