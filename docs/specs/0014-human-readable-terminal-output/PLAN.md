# PLAN

## SUMMARY

- Add a shared human-readable formatting layer for terminal-facing UI surfaces and apply it to help output, clipboard acknowledgements, selectors, workflow guidance, and status rendering.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Record the approved scope and exclusions in a dedicated feature spec before code changes.
- [PLAN-02][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-10] Add shared terminal-formatting helpers for human-readable acknowledgements, headings, selector prompts, and TTY detection.
- [PLAN-03][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Apply the shared formatting to grouped root help, workflow guidance, editor-launch instructions, selector screens, and related command follow-up output.
- [PLAN-04][SPEC-01][SPEC-05] Improve human-readable status presentation
  without changing status content, and keep fleet views in fixed-width
  terminal columns instead of Markdown-style tables, with ANSI color gated on
  TTY output only.
- [PLAN-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09] Add or update tests to prove raw-output stability and new human-readable formatting behavior.
- [PLAN-06][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Update practical docs and rollup state to match the shipped UX.

## COMPONENTS

- `pkg/cli/root.go`
  - grouped root help and usage updates
  - shared clipboard acknowledgement integration
  - workflow guidance integration
- `pkg/cli/human_output.go`
  - TTY detection
  - shared human-readable formatting helpers
- `pkg/cli/status_output.go`
  - status section spacing and shared heading usage
- `pkg/cli/editor_input.go`
  - editor-launch guidance formatting
- selector and interactive output surfaces
  - `pkg/cli/brainstorm.go`
  - `pkg/cli/spec.go`
  - `pkg/cli/plan.go`
  - `pkg/cli/tasks.go`
  - `pkg/cli/implement.go`
  - `pkg/cli/reflect.go`
  - `pkg/cli/complete.go`
  - `pkg/cli/handoff.go`
  - `pkg/cli/catchup.go`
  - `pkg/cli/skill.go`
  - `pkg/cli/code_review.go`
  - `pkg/cli/init.go`
- `pkg/cli/output_test.go`
  - acknowledgement formatting assertions
- `pkg/cli/editor_input_test.go`
  - editor-launch guidance assertions
- `README.md`
  - user-facing output behavior notes

## DATA

- Inputs remain the same:
  - command metadata from Cobra
  - existing command output strings
  - terminal capability inferred from the destination writer when available
- No new persistent state is introduced.
- Raw output payloads for prompts and JSON remain untouched.

## INTERFACES

- Existing command names, flags, and arguments remain unchanged.
- Human-readable surfaces receive presentation-only changes.
- `--output-only` continues to bypass the human-readable acknowledgement path.
- `--json` continues to bypass human-readable rendering.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical command and output rules | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | shipped terminal-output behavior summary | active |
| root command wiring | code | `pkg/cli/root.go` | help and usage template updates | active |
| human output helpers | code | `pkg/cli/human_output.go` | shared formatting and TTY detection | active |
| status output | code | `pkg/cli/status_output.go` | human-readable status rendering | active |
| editor input | code | `pkg/cli/editor_input.go` | editor-launch guidance formatting | active |
| README | doc | `README.md` | user-facing output behavior notes | active |

## RISKS

- Over-formatting can reduce readability if emoji density becomes noisy.
- Shared helpers can accidentally change raw output if the bypass conditions are wrong.
- Cobra help-template changes can affect every command at once.
- TTY detection can diverge between runtime and tests if not isolated cleanly.

## TESTING

- Add or update unit tests for:
  - clipboard acknowledgement formatting
  - raw `--output-only` stability
  - editor-launch guidance formatting
  - selector header formatting helpers
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
