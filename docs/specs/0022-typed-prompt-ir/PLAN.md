# PLAN

## SUMMARY

- Add a typed prompt IR package, migrate all prompt-producing commands to it, and lock the highest-risk migrated outputs down with golden coverage.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Add an internal prompt IR package with a minimal typed block model and a markdown/plain-text renderer.
- [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Migrate prompt-producing commands to build through the IR while preserving the current output-wrapper and decorator flow.
- [PLAN-03][SPEC-07] Add reusable IR helpers for repeated prompt structures where that improves consistency without hiding command-local wording.
- [PLAN-04][SPEC-08][SPEC-09] Add exact-output golden tests for representative migrated prompt builders, with normalization for unstable paths and environment-specific values.
- [PLAN-05][SPEC-10] Keep cached project context out of scope and avoid behavior changes outside prompt construction.

## COMPONENTS

- `internal/promptdoc/*`
  - typed prompt block model
  - renderer
  - small builder helpers
- `pkg/cli/*`
  - migrate prompt-producing commands to the IR
- `pkg/cli/*test.go`
  - golden coverage for representative rendered prompts
- `pkg/cli/testdata/*.golden`
  - expected rendered prompt outputs

## DATA

- No new persisted project state.
- Golden fixtures live under repository testdata.

## INTERFACES

- Internal IR API consumed by CLI prompt builders.
- Existing user-facing command interfaces remain unchanged.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| prompt output wrappers | code | `pkg/cli/prompt_output.go` | preserve rendering boundary and existing clipboard behavior | active |
| prompt builders | code | `pkg/cli/*.go` prompt surfaces | migration targets | active |
| current prompt tests | code | `pkg/cli/*test.go` | retain current semantic coverage during migration | active |

## RISKS

- A broad migration can accidentally change prompt semantics if too much command-specific text is abstracted away.
- Golden tests can become brittle if unstable absolute paths are not normalized.
- Overusing raw escape hatches would reduce the value of the IR even if the migration technically lands.

## TESTING

- Add unit tests for IR rendering behavior.
- Add golden tests for representative migrated prompt builders.
- Keep existing prompt-substring tests unless they become redundant.
- Run:
  - `go test ./internal/promptdoc ./pkg/cli`
  - `go test ./...`
  - `make vet`
  - `make build`
