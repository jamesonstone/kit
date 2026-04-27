# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record simplification docs and amend affected specs | done | agent | |
| T002 | Implement root help grouping and deprecated compatibility surfaces | done | agent | T001 |
| T003 | Implement `kit resume` routing and backlog guidance updates | done | agent | T001 |
| T004 | Implement `kit status --all` text and JSON behavior | done | agent | T001 |
| T005 | Update README, canonical workflow docs, and command help text | done | agent | T002, T003, T004 |
| T006 | Add regression tests and run verification | done | agent | T002, T003, T004, T005 |

## TASK LIST

- [x] T001: Record simplification docs and amend affected specs [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T002: Implement root help grouping and deprecated compatibility surfaces [PLAN-03] [PLAN-04]
- [x] T003: Implement `kit resume` routing and backlog guidance updates [PLAN-01] [PLAN-05]
- [x] T004: Implement `kit status --all` text and JSON behavior [PLAN-02]
- [x] T005: Update README, canonical workflow docs, and command help text [PLAN-05]
- [x] T006: Add regression tests and run verification [PLAN-06]

## TASK DETAILS

### T001

- **GOAL**: record the approved simplification contract before code changes
- **SCOPE**:
  - add `docs/specs/0019-command-surface-simplification/`
  - amend affected existing specs whose shipped contracts now change
- **ACCEPTANCE**:
  - new spec docs exist with complete required sections
  - affected existing specs describe canonical and deprecated surfaces accurately
- **NOTES**: update docs before touching code

### T002

- **GOAL**: simplify the visible root command surface without breaking old entry points
- **SCOPE**:
  - group root help output
  - hide deprecated compatibility commands from default help
  - keep deprecated commands callable
- **ACCEPTANCE**:
  - root help shows grouped canonical commands only
  - deprecated commands remain invokable
- **NOTES**: keep subcommand help low risk

### T003

- **GOAL**: ship one canonical resume command
- **SCOPE**:
  - add `pkg/cli/resume.go`
  - route backlog items through the shared backlog pickup helper
  - route non-backlog features through catch-up prompt behavior
  - update backlog and brainstorm guidance text
- **ACCEPTANCE**:
  - `kit resume` works with explicit feature references and selector fallback
  - backlog and catch-up behavior remain consistent
- **NOTES**: reuse existing helpers instead of duplicating state logic

### T004

- **GOAL**: make project overview an explicit `status` mode
- **SCOPE**:
  - add `status --all`
  - move all-features output out of default `status`
  - add the dedicated all-features JSON payload
- **ACCEPTANCE**:
  - default `status` stays focused on the active feature
  - `status --all` exposes the fleet view in both text and JSON
  - human-readable `status --all` uses a fixed-width lifecycle matrix instead
    of a Markdown-style table
- **NOTES**: do not break the existing default `status --json` shape

### T005

- **GOAL**: teach the shipped surface accurately
- **SCOPE**:
  - update `README.md`
  - update `docs/specs/0000_INIT_PROJECT.md`
  - update touched command help text
- **ACCEPTANCE**:
  - README and canonical workflow docs teach `resume`, `status --all`,
    `upgrade`, and `skill` as canonical commands
  - deprecated commands appear only as migration guidance
- **NOTES**: keep `kit backlog` visible

### T006

- **GOAL**: prove compatibility and prevent regression
- **SCOPE**:
  - add or update focused CLI tests
  - run the normal verification suite
- **ACCEPTANCE**:
  - tests cover resume routing, status all-features output, root help grouping,
    and deprecated visibility
  - verification commands pass
- **NOTES**: include deprecated `brainstorm --pickup`

## DEPENDENCIES

no additional information required

## NOTES

not required
