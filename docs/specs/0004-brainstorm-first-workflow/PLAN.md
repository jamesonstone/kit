# PLAN

## SUMMARY

Introduce a real brainstorm artifact and visible brainstorm phase, then rewire CLI prompts and product docs around that model. Remove parallel workflow concepts (`oneshot`, branch automation) so Kit is consistently document-centered and planning-first.

## APPROACH

1. formalize the workflow contract in repo docs and generated templates
2. add `BRAINSTORM.md` support and a dedicated brainstorm phase in feature/status/rollup logic
3. refactor `kit brainstorm` into the interactive, planning-only feature entrypoint
4. thread `BRAINSTORM.md` through downstream prompts as optional upstream context
5. remove `kit oneshot` and git branch automation from code, config, help, and docs
6. add tests for prompt generation and phase detection, then run full verification

## COMPONENTS

- `pkg/cli/brainstorm.go`
  - interactive input flow
  - feature resolution/creation
  - brainstorm prompt generation
  - output/copy/file behaviors
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
  - create or reuse `docs/specs/<feature>/BRAINSTORM.md`
  - output a `/plan` prompt for a coding agent
- `kit spec`, `kit plan`, `kit tasks`, `kit implement`, `kit reflect`
  - include `BRAINSTORM.md` in file references and instructions when present
- `kit status`
  - display brainstorm-only features correctly
- help/README/constitution/agent templates
  - show optional brainstorm before spec
  - remove `oneshot` and branching language

## RISKS

- phase reordering could break status, handoff, or rollup assumptions
  - mitigate with explicit phase ordering updates and tests
- removing `oneshot` and branching could leave stale references in docs or template generators
  - mitigate with repo-wide search verification
- brainstorm prompt generation could diverge from required planning-only behavior
  - mitigate with focused string-based tests
- downstream commands could assume `SPEC.md` is always the first feature artifact
  - mitigate with brainstorm-aware selection and fallback logic

## TESTING

- unit tests for brainstorm phase detection and ordering
- unit tests for brainstorm prompt generation, including `/plan` prefix plus numbered-list and percentage-progress clarification requirements
- unit tests for brainstorm-aware next-step/status behavior
- repository-wide search verification for removed `oneshot` and branching references
- full `go test ./...`
