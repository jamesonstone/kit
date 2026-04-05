# SPEC

## SUMMARY

- Improve human-readable terminal output with clearer spacing, semantic emoji markers, and more readable help sections.
- Keep generated coding-agent prompts, scaffolded agent instruction files, `--output-only` raw stdout, and `--json` output unchanged.

## PROBLEM

- Kit mixes plain text, sparse emoji usage, dense section layouts, and default Cobra help formatting across its human-facing CLI surfaces.
- The inconsistent presentation makes interactive flows slower to scan, especially in selectors, workflow guidance, status output, and help text.
- Terminal applications cannot reliably change font size, so readability improvements must come from spacing, grouping, and consistent visual cues.

## GOALS

- Improve scanability of human-facing terminal output.
- Add semantic emoji markers to human-readable help, status, selection, and guidance surfaces.
- Increase whitespace and section separation where it improves readability.
- Preserve current command behavior and clipboard-first semantics aside from
  explicit command-surface simplification changes.
- Restrict the new presentation layer to human-readable terminal flows.

## NON-GOALS

- Changing generated coding-agent prompt bodies.
- Changing scaffolded repository instruction file contents.
- Changing `--output-only` raw output payloads.
- Changing `--json` output payloads.
- Introducing third-party UI dependencies.
- Attempting terminal font-size control.

## USERS

- Engineers using Kit interactively in a terminal.
- Users relying on help output to discover commands and flags.
- Users switching frequently between planning, implementation, and support commands.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Cobra help templates | library | `github.com/spf13/cobra` | help and usage presentation updates | active |
| terminal detection | library | `golang.org/x/term` | TTY-only human formatting decisions | active |
| existing CLI output helpers | code | `pkg/cli/root.go`, `pkg/cli/status_output.go`, `pkg/cli/editor_input.go` | shared formatting integration points | active |

## REQUIREMENTS

- Human-readable help output must use clearer section headings and spacing.
- Root help must support grouped command sections without requiring the same
  grouped layout for subcommand help.
- Human-readable clipboard acknowledgements must be easier to scan than the current single-line plain text.
- Human-readable workflow guidance must use consistent section styling and spacing.
- Human-readable selection prompts must use consistent headers and input prompts across commands.
- Human-readable editor-launch instructions must remain accurate while becoming easier to scan.
- Human-readable status output must preserve all current information while improving section separation.
- Human-readable fleet or progress tables must render as fixed-width terminal
  columns instead of Markdown-style pipe tables.
- Human-readable terminal status output may use ANSI color when writing to a
  TTY, but buffered or non-TTY output must remain readable without color.
- Existing raw payloads for `--output-only` must remain byte-for-byte unchanged.
- Existing `--json` output payloads must remain unchanged.
- Generated coding-agent prompt bodies must remain unchanged.
- Scaffolded repository instruction file templates must remain unchanged.
- The implementation must avoid adding non-standard runtime dependencies.

## ACCEPTANCE

- `kit --help` shows grouped section headings and spacing for the visible
  canonical commands, and `kit <command> --help` continues to show clearer
  section headings and spacing than the shipped default template.
- Human-readable clipboard acknowledgements include semantic emoji and clearer spacing.
- Human-readable workflow guidance, selector prompts, and editor-launch guidance use consistent styled headings.
- `kit status` preserves its current information and becomes easier to scan.
- `kit status --all` renders its feature matrix as fixed-width terminal columns.
- terminal `kit status` and `kit status --all` output may use color without
  changing non-TTY text output expectations.
- `--output-only` still prints the exact raw prompt or output text.
- `--json` still returns the same JSON structure.
- Existing generated coding-agent prompt content remains unchanged.
- Tests cover the new shared formatting helpers and preserve raw-output behavior.

## EDGE-CASES

- Help output rendered in non-TTY contexts.
- Human-readable output captured in tests through buffers instead of real terminals.
- Commands that mix shared clipboard acknowledgements with command-specific follow-up guidance.
- Commands that write to custom writers instead of `os.Stdout`.
- Existing tests that assert exact acknowledgement strings.

## OPEN-QUESTIONS

- none
