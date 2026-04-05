# TASKS

## PROGRESS TABLE

| ID   | TASK                                             | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------------------------ | ------ | ----- | ------------ |
| T001 | Record human-readable terminal output feature docs | done | agent |              |
| T002 | Add shared human-readable formatting helpers     | done   | agent | T001         |
| T003 | Apply formatting to help and terminal guidance   | done   | agent | T002         |
| T004 | Improve status and selector readability          | done   | agent | T002         |
| T005 | Update tests for human-readable output changes   | done   | agent | T003, T004   |
| T006 | Update practical docs and verify                 | done   | agent | T005         |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record human-readable terminal output feature docs [PLAN-01]
- [x] T002: Add shared human-readable formatting helpers [PLAN-02]
- [x] T003: Apply formatting to help and terminal guidance [PLAN-03]
- [x] T004: Improve status and selector readability [PLAN-04]
- [x] T005: Update tests for human-readable output changes [PLAN-05]
- [x] T006: Update practical docs and verify [PLAN-06]

## TASK DETAILS

### T001

- **GOAL**: Capture the approved UI scope and exclusions before code changes
- **SCOPE**:
  - add `docs/specs/0014-human-readable-terminal-output/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - feature docs exist with complete required sections
  - exclusions for prompts, instruction files, raw output, and JSON are explicit

### T002

- **GOAL**: Centralize human-readable terminal presentation logic
- **SCOPE**:
  - add shared heading and acknowledgement helpers
  - add terminal-detection helper(s)
  - preserve raw-output bypass behavior
- **ACCEPTANCE**:
  - shared helpers can format headings, selector prompts, and clipboard acknowledgements
  - raw prompt and JSON outputs are unaffected

### T003

- **GOAL**: Improve the readability of help and guidance flows
- **SCOPE**:
  - update Cobra help and usage presentation
  - update workflow guidance and editor-launch instructions
  - update command follow-up guidance where needed
- **ACCEPTANCE**:
  - root help output uses grouped canonical sections with clearer headings and spacing
  - workflow and guidance text uses consistent semantic emoji markers

### T004

- **GOAL**: Make status and selection screens easier to scan
- **SCOPE**:
  - update status rendering
  - update feature selectors and input prompts
- **ACCEPTANCE**:
  - default `status` keeps the active-feature information with improved section separation
  - `status --all` presents the explicit project overview cleanly
  - terminal status views may use TTY-only color without changing buffered or JSON output
  - selector screens share consistent headers and prompts

### T005

- **GOAL**: Prove the new presentation layer works without changing raw outputs
- **SCOPE**:
  - update output-related tests
  - add helper tests where needed
- **ACCEPTANCE**:
  - tests cover shared human-readable formatting
  - tests confirm raw prompt output remains unchanged

### T006

- **GOAL**: Leave docs and verification aligned with the shipped UX
- **SCOPE**:
  - update `README.md`
  - update `docs/PROJECT_PROGRESS_SUMMARY.md`
  - run tests, vet, and build
- **ACCEPTANCE**:
  - docs describe the terminal UX behavior accurately
  - `go test ./...`, `make vet`, and `make build` pass

## DEPENDENCIES

- T002 depends on T001 because the approved scope and exclusions must be documented first.
- T003 depends on T002 because the shared formatter should drive help and guidance changes.
- T004 depends on T002 because status and selector changes should reuse the shared formatter.
- T005 depends on T003 and T004 because tests must reflect the final shipped output.
- T006 depends on T005 because docs and verification must match the final behavior.

## NOTES

- not required
