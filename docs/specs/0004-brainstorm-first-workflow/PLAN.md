# PLAN

## SUMMARY

Introduce a real brainstorm artifact and visible brainstorm phase, then rewire CLI prompts and product docs around that model. Remove parallel workflow concepts (`oneshot`, branch automation) so Kit is consistently document-centered and planning-first. Keep the core workflow prompt commands clipboard-first by default so stdout prompt output is reserved for explicit `--output-only` usage.

## APPROACH

1. formalize the workflow contract in repo docs and generated templates
2. add `BRAINSTORM.md` support and a dedicated brainstorm phase in feature/status/rollup logic
3. refactor `kit brainstorm` into the interactive, planning-only feature entrypoint
4. thread `BRAINSTORM.md` through downstream prompts as optional upstream context
5. keep prompt output behavior command-scoped by adding a clipboard-first helper for the core workflow commands without changing support utilities
6. remove `kit oneshot` and git branch automation from code, config, help, and docs
7. add tests for prompt generation, clipboard-first output, and phase detection, then run full verification

## COMPONENTS

- `pkg/cli/brainstorm.go`
  - interactive input flow
  - feature resolution/creation
  - brainstorm prompt generation
  - output/copy/file behaviors
- `pkg/cli/root.go`
  - shared prompt formatting
  - command-scoped clipboard-first output helper for brainstorm/spec/plan/tasks/implement/reflect
- `pkg/cli/multiline_input.go`
  - shared free-text prompt setup
  - kitty keyboard protocol activation/restoration
  - `Shift+Enter` escape translation and newline insertion
- `pkg/cli/editor_input.go`
  - shared editor-backed free-text input
  - `--vim` and `--editor=vim` flag handling
  - pre-editor instruction screen
  - any-key confirmation before editor launch
  - editor launch, cancel, and submit semantics
- `internal/templates/templates.go`
  - brainstorm artifact template
  - generated agent pointer/template updates
- `internal/feature/feature.go`
  - brainstorm phase constant
  - phase detection based on `BRAINSTORM.md`
- `internal/feature/status.go`
  - brainstorm file/status reporting
- `internal/rollup/rollup.go`
  - brainstorm-aware summary extraction and pointers
- downstream CLI commands
  - optional `BRAINSTORM.md` context in prompts and handoff/status flows
- product surface cleanup
  - remove `oneshot`
  - remove branch automation and related config

## DATA

- feature directory contents become:
  - optional `BRAINSTORM.md`
  - `SPEC.md`
  - `PLAN.md`
  - `TASKS.md`
  - optional `ANALYSIS.md`
- phase ordering becomes:
  - `brainstorm`
  - `spec`
  - `plan`
  - `tasks`
  - `implement`
  - `reflect`
  - `complete`
- `.kit.yaml` removes the `branching` block

## INTERFACES

- `kit brainstorm`
  - default interactive mode
  - prompt for feature name
  - prompt for multiline thesis
  - allow `Shift+Enter` and `Ctrl+J` to insert newlines without submit
  - allow `--vim` and `--editor=vim` to capture the thesis in a vim-compatible editor
  - create or reuse `docs/specs/<feature>/BRAINSTORM.md`
  - output a `/plan` prompt for a coding agent
  - default to copying the prepared prompt to the clipboard and only print the prompt body for `--output-only`
- `kit spec`, `kit plan`, `kit tasks`, `kit implement`, `kit reflect`
  - include `BRAINSTORM.md` in file references and instructions when present
  - preserve shared clarification-loop approval semantics when the `>=95%` understanding workflow is active
- `kit spec`, `kit plan`, `kit tasks`, `kit implement`, `kit reflect`
  - match `kit brainstorm` clipboard-first default output semantics
  - keep `--copy` available as an explicit override for `--output-only`
- `kit spec --interactive`
  - use the same multiline free-text input behavior as `kit brainstorm`
  - support editor-backed per-question responses via `--vim` and `--editor=vim`
- `kit status`
  - display brainstorm-only features correctly
  - include the running Kit version as minor informational metadata
- help/README/constitution/agent templates
  - show optional brainstorm before spec
  - remove `oneshot` and branching language

## RISKS

- phase reordering could break status, handoff, or rollup assumptions
  - mitigate with explicit phase ordering updates and tests
- changing the shared prompt output helper could silently alter unrelated commands
  - mitigate with a dedicated helper used only by the core workflow commands
- removing `oneshot` and branching could leave stale references in docs or template generators
  - mitigate with repo-wide search verification
- brainstorm prompt generation could diverge from required planning-only behavior
  - mitigate with focused string-based tests
- downstream commands could assume `SPEC.md` is always the first feature artifact
  - mitigate with brainstorm-aware selection and fallback logic

## TESTING

- unit tests for brainstorm phase detection and ordering
- unit tests for brainstorm prompt generation, including `/plan` prefix plus numbered-list, approval-syntax, and percentage-progress clarification requirements
- unit tests for clipboard-first prompt output semantics, including default-copy acknowledgement and `--output-only` raw stdout behavior
- unit tests for multiline input translation, including `Shift+Enter` escape handling and blank-line preservation hooks
- unit tests for editor resolution and editor-backed submit/cancel semantics helpers
- unit tests for pre-editor instruction rendering and any-key launch gating
- unit tests for brainstorm-aware next-step/status behavior
- repository-wide search verification for removed `oneshot` and branching references
- full `go test ./...`
