# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID | FEATURE | PATH | PHASE | CREATED | SUMMARY |
| -- | ------- | ---- | ----- | ------- | ------- |
| 0001 | refactor-plan-command | `docs/specs/0001-refactor-plan-command` | tasks | 2026-01-19 | (no description) |
| 0002 | cicd-goreleaser-releases | `docs/specs/0002-cicd-goreleaser-releases` | reflect | 2026-03-04 | - Kit lacks an automated release pipeline for cross-platf... |
| 0003 | inplace-upgrade-update | `docs/specs/0003-inplace-upgrade-update` | spec | 2026-03-05 | - Kit users currently need manual update flows (for examp... |
| 0004 | brainstorm-first-workflow | `docs/specs/0004-brainstorm-first-workflow` | reflect | 2026-03-06 | Kit currently treats brainstorming as an external or stan... |
| 0005 | version-command | `docs/specs/0005-version-command` | reflect | 2026-03-06 | - Kit currently exposes version information only through ... |
| 0006 | skill-mine-command | `docs/specs/0006-skill-mine-command` | reflect | 2026-03-15 | - Kit has no built-in command for turning completed featu... |

## PROJECT INTENT

<!-- TODO: describe the overall project purpose -->

## GLOBAL CONSTRAINTS

See `docs/CONSTITUTION.md` for project-wide constraints and principles.

## FEATURE SUMMARIES

### refactor-plan-command

- **STATUS**: tasks
- **INTENT**: (see SPEC.md)
- **APPROACH**: (see PLAN.md)
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0001-refactor-plan-command/SPEC.md`, `docs/specs/0001-refactor-plan-command/PLAN.md`, `docs/specs/0001-refactor-plan-command/TASKS.md`

### cicd-goreleaser-releases

- **STATUS**: reflect
- **INTENT**: - Kit lacks an automated release pipeline for cross-platform binary distribution. - Releases are not consistently ver...
- **APPROACH**: - Add a `main` branch workflow that computes next `vMAJOR.MINOR.PATCH` tag from existing semantic tags and pushes it....
- **OPEN ITEMS**: - None.
- **POINTERS**: `docs/specs/0002-cicd-goreleaser-releases/SPEC.md`, `docs/specs/0002-cicd-goreleaser-releases/PLAN.md`, `docs/specs/0002-cicd-goreleaser-releases/TASKS.md`

### inplace-upgrade-update

- **STATUS**: spec
- **INTENT**: - Kit users currently need manual update flows (for example, reinstalling with Go tooling), which is slower and incon...
- **APPROACH**: (see PLAN.md)
- **OPEN ITEMS**: - Should prereleases be ignored by default, with no opt-in in this phase? - For `dev` builds, should the command refu...
- **POINTERS**: `docs/specs/0003-inplace-upgrade-update/SPEC.md`, `docs/specs/0003-inplace-upgrade-update/PLAN.md`, `docs/specs/0003-inplace-upgrade-update/TASKS.md`

### brainstorm-first-workflow

- **STATUS**: reflect
- **INTENT**: Kit currently treats brainstorming as an external or standalone activity, while the formal workflow starts at `SPEC.m...
- **APPROACH**: 1. formalize the workflow contract in repo docs and generated templates 2. add `BRAINSTORM.md` support and a dedicate...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0004-brainstorm-first-workflow/SPEC.md`, `docs/specs/0004-brainstorm-first-workflow/PLAN.md`, `docs/specs/0004-brainstorm-first-workflow/TASKS.md`

### version-command

- **STATUS**: reflect
- **INTENT**: - Kit currently exposes version information only through the root `--version` flag. - Users and scripts do not have a...
- **APPROACH**: - Implement a no-arg Cobra command with `RunE` that writes the resolved version to `cmd.OutOrStdout()`. - Keep format...
- **OPEN ITEMS**: - None.
- **POINTERS**: `docs/specs/0005-version-command/SPEC.md`, `docs/specs/0005-version-command/PLAN.md`, `docs/specs/0005-version-command/TASKS.md`

### skill-mine-command

- **STATUS**: reflect
- **INTENT**: - Kit has no built-in command for turning completed feature work into reusable agent skills. - Reusable implementatio...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-32][SPEC-36] Extend config with a configurable canonical skills directory and keep...
- **OPEN ITEMS**: - None.
- **POINTERS**: `docs/specs/0006-skill-mine-command/SPEC.md`, `docs/specs/0006-skill-mine-command/PLAN.md`, `docs/specs/0006-skill-mine-command/TASKS.md`

## LAST UPDATED

2026-03-15 17:10:51 EDT
