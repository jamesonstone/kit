---
kit_metadata_version: 1
artifact: "plan"
feature:
  id: "0027"
  slug: "implement-readiness-gate"
  dir: "0027-implement-readiness-gate"
dependencies:
  - name: "constitution contract"
    type: "doc"
    location: "docs/CONSTITUTION.md"
    used_for: "canonical workflow rules and implementation gate contract"
    status: "active"
  - name: "init project spec"
    type: "doc"
    location: "docs/specs/0000_INIT_PROJECT.md"
    used_for: "shipped workflow summary and gate wording"
    status: "active"
  - name: "implement command"
    type: "code"
    location: "pkg/cli/implement.go"
    used_for: "readiness-gate prompt behavior"
    status: "active"
  - name: "status command"
    type: "code"
    location: "pkg/cli/status.go"
    used_for: "next-step wording for completed work"
    status: "active"
  - name: "repository instruction templates"
    type: "code"
    location: "internal/templates/templates.go"
    used_for: "scaffolded workflow guidance"
    status: "active"
  - name: "README"
    type: "doc"
    location: "README.md"
    used_for: "user-facing workflow wording"
    status: "active"
---
# PLAN

## SUMMARY

- Add a pre-implementation readiness gate to `kit implement` and document it consistently across the core workflow docs.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Record the approved readiness-gate contract in feature docs before changing code.
- [PLAN-02][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11] Rewrite the `kit implement` prompt so it begins with an implementation-readiness gate and adversarial preflight instructions.
- [PLAN-03][SPEC-12][SPEC-13] Update adjacent workflow wording in `kit status`, `README.md`, `docs/CONSTITUTION.md`, `docs/specs/0000_INIT_PROJECT.md`, and scaffolded instruction templates without introducing a new phase.
- [PLAN-04][SPEC-14] Leave `kit check` unchanged for v1.
- [PLAN-05][SPEC-15][SPEC-16] Add focused tests for the implement prompt and status wording, then run verification.

## COMPONENTS

- `pkg/cli/implement.go`
  - implementation-readiness gate instructions
  - adversarial preflight guidance
  - go/no-go behavior before task execution
- `pkg/cli/status.go`
  - next-step wording for task-complete features
- `pkg/cli/implement_test.go`
  - prompt contract assertions
- `pkg/cli/status_test.go`
  - readiness-gate next-step assertions
- `README.md`
  - command description and workflow wording
- `docs/CONSTITUTION.md`
  - canonical workflow contract wording
- `docs/specs/0000_INIT_PROJECT.md`
  - shipped product spec summary
- `internal/templates/templates.go`
  - scaffolded repository instruction templates
- `internal/templates/templates_test.go`
  - readiness-gate template assertions

## DATA

- Input data remains the existing feature document set:
  - `docs/CONSTITUTION.md`
  - optional `BRAINSTORM.md`
  - `SPEC.md`
  - `PLAN.md`
  - `TASKS.md`
  - `PROJECT_PROGRESS_SUMMARY.md`
- No new persistent state or artifact type is introduced.

## INTERFACES

- Command surface remains:
  - `kit implement [feature]`
- Output shape remains clipboard-first prompt transport.
- Prompt semantics change so implementation begins with an implementation-readiness gate before execution.
- `kit status` remains phase-based and may only change its guidance text.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical workflow rules and implementation gate contract | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | shipped workflow summary and gate wording | active |
| implement command | code | `pkg/cli/implement.go` | readiness-gate prompt behavior | active |
| status command | code | `pkg/cli/status.go` | next-step wording for completed work | active |
| repository instruction templates | code | `internal/templates/templates.go` | scaffolded workflow guidance | active |
| README | doc | `README.md` | user-facing workflow wording | active |

## RISKS

- The readiness gate prompt can become verbose or repetitive if it duplicates too much of `reflect`.
- The adversarial preflight can blur into a new phase if wording implies durable workflow state.
- Status wording can overpromise a new readiness model if it implies a tracked substate.
- The implement prompt can accidentally encourage speculative doc rewrites unless the prompt keeps the binding-document rules explicit.

## TESTING

- Add unit tests for the `kit implement` prompt.
- Add unit tests for scaffolded instruction templates.
- Assert the prompt includes:
  - implementation readiness gate wording
  - adversarial preflight wording
  - contradictions/ambiguity/edge-case challenge instructions
  - stop-and-update-docs behavior on failure
  - rerun-gate behavior before coding
  - first-incomplete-task execution only after a pass
- Assert scaffolded instruction templates include the readiness-gate workflow guidance.
- Update or add unit tests for `kit status` next-step wording when tasks are complete.
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
