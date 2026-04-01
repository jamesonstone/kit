# PLAN

## SUMMARY

- Add a prompt-only `catchup` command that helps a coding agent resume work on a specific feature without moving directly into implementation.
- Reuse existing feature status, selector, and clipboard-first prompt-output patterns so the new surface is additive and does not duplicate `handoff`, `summarize`, or `implement`.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04] Create a new `pkg/cli/catchup.go` command with optional feature argument, selector fallback, and standard `--copy` / `--output-only` behavior.
- [PLAN-02][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09] Resolve the selected feature with existing feature/status helpers and derive stage plus state from `feature.GetFeatureStatus(...)` and current next-action guidance.
- [PLAN-03][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16] Build a feature-scoped `/plan` prompt that tells the coding agent how to catch up on the selected feature, ask questions first, stay in plan mode, and request explicit approval before implementation.
- [PLAN-04][SPEC-17] Add complete-phase-specific prompt wording so completed features are treated as review/reopen triage rather than resumed implementation.
- [PLAN-05][SPEC-18] Register the command in help ordering and README with wording that clearly distinguishes it from `handoff`, `summarize`, and `implement`.
- [PLAN-06][SPEC-19] Add focused tests for prompt generation and state rendering, then run the normal verification commands.
- [PLAN-07] Switch `catchup` to the shared clipboard-first helper while keeping `--output-only` and `--copy` behavior explicit.
- [PLAN-08] Register the shared `--prompt-only` flag on `catchup` so the command surface matches the rest of Kit's feature-scoped prompt commands.

## COMPONENTS

- `pkg/cli/catchup.go`
  - command registration
  - feature selector
  - command execution
- `pkg/cli/catchup_prompt.go`
  - prompt builder
  - stage/state formatting helpers
- `pkg/cli/catchup_test.go`
  - prompt-generation tests
- `pkg/cli/root.go`
  - command order entry
- `README.md`
  - context-management command docs

## DATA

- Input data comes from:
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
  - selected feature docs under `docs/specs/<feature>/`
  - `feature.GetFeatureStatus(...)`
- No new persistent state is introduced.
- No external HTTP or APIs are involved.

## INTERFACES

- New command:
  - `kit catchup [feature]`
- Flags:
  - `--copy`
  - `--output-only`
  - `--prompt-only`
- Output shape:
  - prompt-only, passed through the shared clipboard-first prompt helper
  - workflow footer via `printWorkflowInstructions(...)`

## RISKS

- `catchup` can drift into `handoff` duplication if it starts carrying project-wide session-transfer guidance.
- `catchup` can drift into `implement` duplication if the prompt tells the agent to begin execution instead of asking questions first.
- State wording can become inconsistent if it invents a parallel stage model instead of using `feature.GetFeatureStatus(...)`.
- Complete-phase handling can be misleading if the prompt assumes every resumed feature needs more implementation.
- Clipboard-first output can become confusing if acknowledgement text replaces the prompt body without a documented `--output-only` escape hatch.

## TESTING

- Add unit tests for the prompt builder.
- Assert prompt includes:
  - `/plan`
  - selected feature slug
  - current stage and state
  - `CONSTITUTION.md`
  - `PROJECT_PROGRESS_SUMMARY.md`
  - ordered feature-doc reading instructions
  - “stay in plan mode”
  - “ask questions first”
  - explicit permission before implementation
  - optional `kit summarize <feature>` reference
  - complete-phase reopen/review wording
- Add or reuse unit tests for clipboard-first prompt output semantics.
- Verify `kit catchup --help` exposes `--prompt-only`.
- Run:
  - `make vet`
  - `make test`
  - `make build`
  - `./bin/kit catchup --help`
  - `./bin/kit --help`
