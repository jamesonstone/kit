---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0002
  slug: legacy-prompt
  dir: 0002-legacy-prompt
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Implement the sample behavior. [PLAN-01] | done | agent | |

## TASK LIST

- [x] T001: Implement the sample behavior. [PLAN-01]

## TASK DETAILS

### T001

- **GOAL**: Produce the requested sample output.
- **SCOPE**: Keep the change local to the fixture.
- **ACCEPTANCE**: Output is complete and correct.
- **VERIFY**: Inspect the deterministic command result.
- **EXPECTED FILES**: fixture output only.
- **RISK**: Prompt drift.
- **ROLLBACK**: Revert the fixture change.
- **DEPENDENCIES**: None.
- **NOTES**: Preserve evidence.
