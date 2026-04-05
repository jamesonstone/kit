# SPEC

## SUMMARY

- Refactor `kit plan [feature]` into a dedicated implementation-plan step that
  scaffolds `PLAN.md`, enforces the spec-first workflow, and keeps project
  state in sync.
- Keep the command focused on plan creation and planning guidance, not task
  generation or implementation execution.

## PROBLEM

- Implementation strategy needs its own explicit artifact between requirements
  and tasks.
- Without a dedicated `plan` command, strategy details drift into `SPEC.md` or
  straight into code, which weakens traceability and makes task generation less
  deterministic.
- The workflow also needs one place to enforce `SPEC.md` as the prerequisite
  for planning and to keep `PROJECT_PROGRESS_SUMMARY.md` aligned with the
  highest completed artifact.

## GOALS

- Add a dedicated `kit plan [feature]` command for creating or opening
  `PLAN.md`.
- Keep the workflow spec-first by requiring `SPEC.md` before planning unless
  the user opts into out-of-order scaffolding.
- Scaffold a canonical `PLAN.md` structure that is explicit enough to drive
  `TASKS.md`.
- Keep planning inputs and dependencies visible in the plan artifact.
- Update `PROJECT_PROGRESS_SUMMARY.md` after plan creation so lifecycle state
  stays current.

## NON-GOALS

- Generate `TASKS.md` as part of the planning step.
- Start implementation or produce code changes.
- Replace `SPEC.md` as the binding definition of what is being built.
- Add non-markdown planning artifacts or hidden planning state.

## USERS

- Maintainers turning an approved spec into an execution-ready implementation
  strategy.
- Coding agents that need a bounded planning contract before task generation.
- Contributors resuming work on features that already have `SPEC.md`.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | workflow constraints and plan-stage rules | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical `kit plan` behavior and artifact model | active |
| plan command | code | `pkg/cli/plan.go` | shipped command behavior, prerequisite handling, and prompt contract | active |
| plan template | code | `internal/templates/templates.go` | required `PLAN.md` sections | active |
| rollup generator | code | `internal/rollup/rollup.go` | project summary updates after plan creation | active |

## REQUIREMENTS

- [SPEC-01] `kit plan [feature]` must create or open `PLAN.md` as the
  implementation-plan artifact for a feature.
- [SPEC-02] `SPEC.md` must be the default prerequisite for planning.
- [SPEC-03] When `--force` or out-of-order scaffolding is enabled, the command
  may create the missing prerequisite artifacts with canonical headers.
- [SPEC-04] When no feature argument is provided, the command must resolve a
  feature through interactive selection of features that are ready for
  planning.
- [SPEC-05] `PLAN.md` must include `SUMMARY`, `APPROACH`, `COMPONENTS`,
  `DATA`, `INTERFACES`, `DEPENDENCIES`, `RISKS`, and `TESTING`.
- [SPEC-06] The plan workflow must keep implementation-shaping dependencies
  explicit in `PLAN.md`.
- [SPEC-07] The command must refresh `PROJECT_PROGRESS_SUMMARY.md` after the
  plan artifact is created or updated.
- [SPEC-08] Planning guidance must treat `SPEC.md` as the fixed contract and
  keep `PLAN.md` focused on how execution will work.

## ACCEPTANCE

- Running `kit plan <feature>` with an existing spec creates `PLAN.md` when it
  is missing and leaves the file in place when it already exists.
- Running `kit plan <feature>` without `SPEC.md` fails with an actionable
  prerequisite error unless `--force` or out-of-order scaffolding is enabled.
- Running `kit plan` with no feature argument offers only features that are
  ready for planning.
- Generated `PLAN.md` files include the required sections, including
  `## DEPENDENCIES`.
- Successful plan creation updates `PROJECT_PROGRESS_SUMMARY.md`.
- The generated planning prompt keeps the agent in plan mode and makes the
  subsequent `TASKS.md` step obvious.

## EDGE-CASES

- `PLAN.md` already exists for the selected feature.
- The selected feature cannot be resolved from the provided reference.
- `SPEC.md` is missing and the user did not opt into out-of-order scaffolding.
- Interactive selection receives invalid input.
- The rollup update fails after `PLAN.md` is created.

## OPEN-QUESTIONS

- none
