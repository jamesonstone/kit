# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record the typed prompt IR feature contract | done | agent | |
| T002 | Add the internal prompt IR package and rendering helpers | done | agent | T001 |
| T003 | Migrate prompt-producing commands to the IR | done | agent | T002 |
| T004 | Add golden tests and verification | done | agent | T003 |

## TASK LIST

- [x] T001: Record the typed prompt IR feature contract [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T002: Add the internal prompt IR package and rendering helpers [PLAN-01]
- [x] T003: Migrate prompt-producing commands to the IR [PLAN-02] [PLAN-03]
- [x] T004: Add golden tests and verification [PLAN-04] [PLAN-05]

## TASK DETAILS

### T001
- **GOAL**: capture the approved prompt-IR migration contract before code changes
- **SCOPE**:
  - create `SPEC.md`
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - `docs/specs/0022-typed-prompt-ir/` exists
  - the docs define scope, renderer boundary, and golden-test expectations

### T002
- **GOAL**: introduce a small typed prompt document model
- **SCOPE**:
  - add `internal/promptdoc`
  - support the required block types
  - add rendering helpers
- **ACCEPTANCE**:
  - prompt builders can construct prompts without directly depending on `strings.Builder`
  - the renderer produces plain strings suitable for current output helpers
- **NOTES**:
  - added `internal/promptdoc/doc.go`
  - added `pkg/cli/prompt_ir_helpers.go`
  - covered rendering behavior in `internal/promptdoc/doc_test.go`

### T003
- **GOAL**: route all prompt-producing commands through the IR
- **SCOPE**:
  - migrate prompt builders across the scoped commands
  - keep decorators and output wrappers unchanged
  - add reusable IR helpers only where they improve consistency
- **ACCEPTANCE**:
  - all prompt-producing commands in scope build through the IR
  - prompt meaning and command intent remain preserved
- **NOTES**:
  - migrated `brainstorm`, `spec`, `plan`, `tasks`, `implement`, `reflect`, `catchup`, `handoff`, `reconcile`, `dispatch`, `skill mine`, `summarize`, `code-review`, and the init-time constitution prompt
  - preserved `prompt_output.go`, `skills_prompt.go`, and `subagents.go` as post-render decorators/output wrappers
  - removed the remaining builder-wrapper escape hatches from `skill mine` so the prompt body now builds through typed sections instead of `Raw(renderBuilderText(...))`

### T004
- **GOAL**: lock in correctness for the migrated prompt surface
- **SCOPE**:
  - add golden tests for representative migrated builders
  - normalize unstable paths or environment values
  - run targeted and full verification
- **ACCEPTANCE**:
  - golden tests cover representative migrated builders
  - verification commands pass
- **NOTES**:
  - added direct golden coverage for `code-review`, generic `summarize`, feature-scoped `summarize`, and `reflect`
  - retained command-local semantic tests for other migrated prompt builders to catch prompt-contract regressions without snapshotting every surface
  - normalized golden comparisons for CRLF and trailing newline drift in `pkg/cli/prompt_golden_test.go`
  - verification passed: `go test ./internal/promptdoc ./pkg/cli`, `go test ./...`, `make vet`, `make build`, `git diff --check`

## DEPENDENCIES

- T002 depends on T001 because the IR contract must be explicit before code changes
- T003 depends on T002 because commands need the IR package before they can migrate
- T004 depends on T003 because goldens and verification must reflect the migrated builders

## NOTES

- Cached project context is intentionally out of scope
