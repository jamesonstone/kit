# PLAN

## SUMMARY

Introduce a real brainstorm artifact and visible brainstorm phase, then rewire CLI prompts and product docs around that model. Remove parallel workflow concepts (`oneshot`, branch automation) so Kit is consistently document-centered and planning-first. Keep the core workflow prompt commands clipboard-first by default so stdout prompt output is reserved for explicit `--output-only` usage, add a side-effect-free `--prompt-only` regeneration path for existing features, and make supported multiline free-text flows open a vim-compatible editor by default.

## APPROACH

1. formalize the workflow contract in repo docs and generated templates
2. add `BRAINSTORM.md` support and a dedicated brainstorm phase in feature/status/rollup logic
3. refactor `kit brainstorm` into the interactive, planning-only feature entrypoint
4. thread `BRAINSTORM.md` through downstream prompts as optional upstream context and phase dependency source
5. keep prompt output behavior command-scoped by adding a clipboard-first helper for the core workflow commands without changing support utilities
6. add a shared `--prompt-only` flag for feature-scoped prompt commands and branch artifact-writing commands into side-effect-free regeneration mode
7. make supported multiline free-text prompts editor-default and add an explicit `--inline` opt-out where inline entry already exists
8. remove `kit oneshot` and git branch automation from code, config, help, and docs
9. add tests for prompt generation, clipboard-first output, prompt-only regeneration, editor-default free-text flows, and phase detection, then run full verification

## COMPONENTS

- `pkg/cli/brainstorm.go`
  - interactive input flow
  - feature resolution/creation
  - brainstorm prompt generation
  - output/copy/file behaviors
- `pkg/cli/root.go`
  - shared prompt formatting
  - command-scoped clipboard-first output helper for brainstorm/spec/plan/tasks/implement/reflect
- `pkg/cli/prompt_only.go`
  - shared `--prompt-only` flag registration and lookup
- `pkg/cli/multiline_input.go`
  - shared free-text prompt setup
  - kitty keyboard protocol activation/restoration
  - `Shift+Enter` escape translation and newline insertion
- `pkg/cli/editor_input.go`
  - shared editor-backed free-text input
  - `--vim` and `--editor=vim` flag handling
  - `--inline` opt-out flag handling for inline-capable flows
  - pre-editor instruction screen
  - any-key confirmation before editor launch
  - editor launch, cancel, and submit semantics
- `internal/templates/templates.go`
  - brainstorm artifact template
  - generated agent pointer/template updates
  - dependency inventory tables for brainstorm and plan docs
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
  - default the multiline thesis to a vim-compatible editor
  - allow `--inline` to switch the thesis back to terminal multiline entry with `Shift+Enter` and `Ctrl+J`
  - allow `--vim` and `--editor=vim` to capture the thesis in a vim-compatible editor explicitly
  - create or reuse `docs/specs/<feature>/BRAINSTORM.md`
  - output a `/plan` prompt for a coding agent
  - require `BRAINSTORM.md` `## DEPENDENCIES` to track the supporting inputs used during the brainstorm phase
  - default to copying the prepared prompt to the clipboard and only print the prompt body for `--output-only`
  - allow `--prompt-only` to regenerate the prompt from an existing `BRAINSTORM.md` without asking for a new thesis, without writing files, and without updating rollups
- `kit spec`, `kit plan`, `kit tasks`, `kit implement`, `kit reflect`
  - include `BRAINSTORM.md` in file references and instructions when present
  - preserve shared clarification-loop approval semantics when the `>=95%` understanding workflow is active
  - accept `--prompt-only` to regenerate prompts for an existing feature without mutating repo docs
- `kit plan`
  - refresh `PLAN.md` `## DEPENDENCIES` with the resources that shape the implementation strategy
- `kit spec`, `kit plan`, `kit tasks`, `kit implement`, `kit reflect`
  - match `kit brainstorm` clipboard-first default output semantics
  - keep `--copy` available as an explicit override for `--output-only`
- `kit spec --interactive`
  - use the same editor-default multiline free-text behavior as `kit brainstorm`
  - support `--inline` to switch per-question responses back to terminal multiline entry
  - support editor-backed per-question responses via `--vim` and `--editor=vim`
- `kit status`
  - display brainstorm-only features correctly
  - include the running Kit version as minor informational metadata
- help/README/constitution/agent templates
  - show optional brainstorm before spec
  - remove `oneshot` and branching language

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | workflow model and prompt semantics | active |
| document templates | code | `internal/templates/templates.go` | brainstorm artifact and prompt scaffolding | active |
| rollup generator | code | `internal/rollup/rollup.go` | brainstorm-aware project summary output | active |
| instruction templates | doc | `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md` | repository instruction alignment | active |

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
- unit tests for brainstorm and plan dependency-inventory guidance
- unit tests for clipboard-first prompt output semantics, including default-copy acknowledgement and `--output-only` raw stdout behavior
- unit tests for prompt-only regeneration, including existing-feature selectors and missing-artifact failures
- unit tests for multiline input translation, including `Shift+Enter` escape handling and blank-line preservation hooks
- unit tests for editor resolution and editor-backed submit/cancel semantics helpers
- unit tests for default editor routing and `--inline` opt-out behavior on brainstorm/spec interactive flows
- unit tests for pre-editor instruction rendering and any-key launch gating
- unit tests for brainstorm-aware next-step/status behavior
- repository-wide search verification for removed `oneshot` and branching references
- full `go test ./...`
