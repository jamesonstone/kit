# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID   | FEATURE                            | PATH                                                 | PHASE   | CREATED    | SUMMARY                                                      |
| ---- | ---------------------------------- | ---------------------------------------------------- | ------- | ---------- | ------------------------------------------------------------ |
| 0001 | refactor-plan-command              | `docs/specs/0001-refactor-plan-command`              | tasks   | 2026-01-19 | (no description)                                             |
| 0002 | cicd-goreleaser-releases           | `docs/specs/0002-cicd-goreleaser-releases`           | reflect | 2026-03-04 | - Kit lacks an automated release pipeline for cross-platf... |
| 0003 | inplace-upgrade-update             | `docs/specs/0003-inplace-upgrade-update`             | reflect | 2026-03-16 | - Kit users currently need manual update flows (for examp... |
| 0004 | brainstorm-first-workflow          | `docs/specs/0004-brainstorm-first-workflow`          | reflect | 2026-03-06 | Kit currently treats brainstorming as an external or stan... |
| 0005 | version-command                    | `docs/specs/0005-version-command`                    | reflect | 2026-03-06 | - Kit currently exposes version information only through ... |
| 0006 | skill-mine-command                 | `docs/specs/0006-skill-mine-command`                 | reflect | 2026-03-15 | - Kit has no built-in command for turning completed featu... |
| 0007 | catchup-command                    | `docs/specs/0007-catchup-command`                    | reflect | 2026-03-23 | - Kit has `status`, `handoff`, `summarize`, and `implemen... |
| 0008 | dispatch-command                   | `docs/specs/0008-dispatch-command`                   | reflect | 2026-03-27 | - Kit has prompt generators for planning, catch-up, imple... |
| 0009 | spec-skills-discovery              | `docs/specs/0009-spec-skills-discovery`              | reflect | 2026-03-29 | - Kit captures reusable skills after implementation, but ... |
| 0010 | support-command-clipboard-defaults | `docs/specs/0010-support-command-clipboard-defaults` | reflect | 2026-03-31 | - These three support commands still print their full pro... |
| 0011 | handoff-document-sync              | `docs/specs/0011-handoff-document-sync`              | reflect | 2026-03-31 | - `kit handoff` currently focuses on orienting a fresh ag... |
| 0012 | implement-readiness-gate           | `docs/specs/0012-implement-readiness-gate`           | reflect | 2026-04-02 | - `kit implement` currently moves directly from document ... |
| 0013 | scaffold-agents-safe-merge         | `docs/specs/0013-scaffold-agents-safe-merge`         | reflect | 2026-04-02 | - `kit scaffold-agents` currently has binary behavior: - ... |
| 0014 | human-readable-terminal-output     | `docs/specs/0014-human-readable-terminal-output`     | reflect | 2026-04-04 | - Kit mixes plain text, sparse emoji usage, dense sectio... |

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

- **STATUS**: reflect
- **INTENT**: - Kit users currently need manual update flows (for example, reinstalling with Go tooling), which is slower and incon...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Add a new `pkg/cli/upgrade.go` file that registers `upgrade` and `update` as r...
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

### catchup-command

- **STATUS**: reflect
- **INTENT**: - Kit has `status`, `handoff`, `summarize`, and `implement`, but no feature-scoped command dedicated to helping a cod...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04] Create a new `pkg/cli/catchup.go` command with optional feature argum...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0007-catchup-command/SPEC.md`, `docs/specs/0007-catchup-command/PLAN.md`, `docs/specs/0007-catchup-command/TASKS.md`

### dispatch-command

- **STATUS**: reflect
- **INTENT**: - Kit has prompt generators for planning, catch-up, implementation, reflection, and skill mining, but no prompt-only ...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-1...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0008-dispatch-command/SPEC.md`, `docs/specs/0008-dispatch-command/PLAN.md`, `docs/specs/0008-dispatch-command/TASKS.md`

### spec-skills-discovery

- **STATUS**: reflect
- **INTENT**: - Kit captures reusable skills after implementation, but the specification workflow does not currently tell coding ag...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Update document validation and templates so `SPEC.md` requires a `## SKILLS` s...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0009-spec-skills-discovery/SPEC.md`, `docs/specs/0009-spec-skills-discovery/PLAN.md`, `docs/specs/0009-spec-skills-discovery/TASKS.md`

### support-command-clipboard-defaults

- **STATUS**: reflect
- **INTENT**: - These three support commands still print their full prompt or output body by default while the rest of Kit's prompt...
- **APPROACH**: - [PLAN-01] Update the formal docs for the three commands before changing code. - [PLAN-02] Route all three commands ...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0010-support-command-clipboard-defaults/SPEC.md`, `docs/specs/0010-support-command-clipboard-defaults/PLAN.md`, `docs/specs/0010-support-command-clipboard-defaults/TASKS.md`

### handoff-document-sync

- **STATUS**: reflect
- **INTENT**: - `kit handoff` currently focuses on orienting a fresh agent session, but it does not explicitly require the current ...
- **APPROACH**: - [PLAN-01] Record the handoff prompt contract before changing code. - [PLAN-02] Add prompt-building helpers that gen...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0011-handoff-document-sync/SPEC.md`, `docs/specs/0011-handoff-document-sync/PLAN.md`, `docs/specs/0011-handoff-document-sync/TASKS.md`

### implement-readiness-gate

- **STATUS**: reflect
- **INTENT**: - `kit implement` currently moves directly from document reading to task execution. - That leaves no explicit semanti...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Record the approved readiness-gate contract in feat...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0012-implement-readiness-gate/SPEC.md`, `docs/specs/0012-implement-readiness-gate/PLAN.md`, `docs/specs/0012-implement-readiness-gate/TASKS.md`

### scaffold-agents-safe-merge

- **STATUS**: reflect
- **INTENT**: - `kit scaffold-agents` currently has binary behavior: - skip existing files by default - overwrite existing files bl...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Record the scaffold safety contract in a d...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0013-scaffold-agents-safe-merge/SPEC.md`, `docs/specs/0013-scaffold-agents-safe-merge/PLAN.md`, `docs/specs/0013-scaffold-agents-safe-merge/TASKS.md`

### human-readable-terminal-output

- **STATUS**: reflect
- **INTENT**: - Kit mixes plain text, sparse emoji usage, dense section layouts, and default Cobra help formatting across its huma...
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Record the approved scope and exclusions in a dedicated fe...
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0014-human-readable-terminal-output/SPEC.md`, `docs/specs/0014-human-readable-terminal-output/PLAN.md`, `docs/specs/0014-human-readable-terminal-output/TASKS.md`

## LAST UPDATED

2026-04-04 07:28:53 EDT
