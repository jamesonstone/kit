# PLAN

## SUMMARY

- Switch `kit handoff`, `kit summarize`, and `kit code-review` to the shared clipboard-first output contract without changing `summarize` or `code-review` generated content.

## APPROACH

- [PLAN-01] Update the formal docs for the three commands before changing code.
- [PLAN-02] Route all three commands through the shared clipboard-first helper.
- [PLAN-03] Update README and help strings to reflect the new default output contract.
- [PLAN-04] Reuse existing clipboard-first helper tests and rerun repository verification.

## COMPONENTS

- `pkg/cli/handoff.go`
  - flag text
  - output helper usage
- `pkg/cli/summarize.go`
  - flag text
  - output helper usage
- `pkg/cli/code_review.go`
  - flag text
  - output helper usage
- `README.md`
  - clipboard-first command note

## DATA

- No new persistent state.
- No changes to `summarize` or `code-review` generated prompt bodies.
- `kit handoff` prompt content is owned by `0011-handoff-document-sync`.

## INTERFACES

- Commands:
  - `kit handoff [feature]`
  - `kit summarize [feature]`
  - `kit code-review`
- Flags:
  - `--copy`
  - `--output-only`
- Output shape:
  - default copies to clipboard and prints an acknowledgement
  - `--output-only` prints raw stdout
  - `--output-only --copy` both prints and copies

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| clipboard-first helper | code | `pkg/cli/root.go` | shared prompt and output transport behavior | active |
| human output style | code | `pkg/cli/human_output.go` | clipboard acknowledgement rendering | active |
| handoff prompt flow | code | `pkg/cli/handoff.go` | output transport for handoff prompts | active |
| summarize command | code | `pkg/cli/summarize.go` | output transport for summarize prompts | active |
| code review command | code | `pkg/cli/code_review.go` | output transport for review prompts | active |
| README | doc | `README.md` | user-facing output contract | active |

## RISKS

- Users accustomed to reading the full default output in the terminal may need to learn the `--output-only` escape hatch.
- Command help can become misleading if flag text is not updated together with behavior.
- The shared helper must preserve current content formatting and only change output transport.

## TESTING

- Reuse the existing clipboard-first helper tests.
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
