# PLAN

## SUMMARY

- Add a dedicated CLI path for planning that validates prerequisites, scaffolds
  a canonical `PLAN.md`, and emits guidance that turns an approved `SPEC.md`
  into an execution-ready implementation strategy.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Keep `pkg/cli/plan.go` as the single
  entry point for prerequisite enforcement and `PLAN.md` creation so the
  workflow stays explicit.
- [PLAN-02][SPEC-04] Reuse feature discovery and selection helpers to target
  features that are ready for planning.
- [PLAN-03][SPEC-05][SPEC-06] Use the embedded plan template to enforce the
  required plan sections and dependency inventory.
- [PLAN-04][SPEC-07] Regenerate `PROJECT_PROGRESS_SUMMARY.md` after plan
  creation so project state reflects the highest completed artifact.
- [PLAN-05][SPEC-08] Emit a planning prompt that reads the constitution and
  spec inputs first, then keeps the output focused on implementation strategy
  rather than tasks or code.

## COMPONENTS

- `pkg/cli/plan.go`
  - command registration
  - feature resolution
  - prerequisite checks
  - `PLAN.md` creation and prompt output
- `internal/templates/templates.go`
  - canonical `PLAN.md` structure
- `internal/feature/feature.go`
  - feature listing and selection support
- `internal/rollup/rollup.go`
  - progress-summary refresh after plan creation

## DATA

- `PLAN.md`
  - persistent implementation-strategy artifact under the feature directory
- feature references
  - direct feature arguments or interactive selections resolved from
    `docs/specs/`
- `PROJECT_PROGRESS_SUMMARY.md`
  - regenerated lifecycle summary after successful plan creation
- no new hidden state or non-markdown storage is introduced

## INTERFACES

- CLI entry point: `kit plan [feature]`
- prerequisite contract:
  - `SPEC.md` must exist unless `--force` or out-of-order scaffolding is
    enabled
- artifact writes:
  - create `PLAN.md` when missing
  - optionally scaffold `SPEC.md` when planning is forced out of order
  - refresh `PROJECT_PROGRESS_SUMMARY.md`
- user interaction:
  - explicit feature argument or interactive feature selection
- prompt contract:
  - read `CONSTITUTION.md` and `SPEC.md`
  - write `PLAN.md` as a how-focused implementation strategy

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| plan command | code | `pkg/cli/plan.go` | command flow, prompts, and prerequisite handling | active |
| feature resolution helpers | code | `internal/feature/feature.go` | feature lookup and selection | active |
| plan template | code | `internal/templates/templates.go` | canonical plan structure | active |
| rollup generator | code | `internal/rollup/rollup.go` | lifecycle summary refresh | active |
| constitution contract | doc | `docs/CONSTITUTION.md` | workflow constraints and traceability rules | active |

## RISKS

- Planning can drift into task-writing or coding if the prompt is not explicit.
  Mitigate by treating `SPEC.md` as fixed and keeping `PLAN.md` scoped to how.
- Out-of-order scaffolding can create thin artifacts. Mitigate by keeping it
  opt-in and using canonical templates.
- Rollup updates can fail after plan creation. Mitigate by surfacing that error
  clearly without hiding the created artifact.

## TESTING

- command tests for prerequisite errors and `--force` scaffolding
- command tests for interactive feature selection
- template checks for required `PLAN.md` sections, including `## DEPENDENCIES`
- prompt-output checks to confirm the agent stays in planning mode
- rollup update checks after successful plan creation
