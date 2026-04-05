# SPEC

## SUMMARY

- Add an implementation-readiness gate inside `kit implement` that requires an adversarial preflight over the feature docs before coding begins.
- Keep the workflow as a gate, not a new lifecycle phase, command, or artifact type.

## PROBLEM

- `kit implement` currently moves directly from document reading to task execution.
- That leaves no explicit semantic challenge step to catch contradictions, ambiguity, weak task coverage, or missing failure cases before code is written.
- The existing `kit check` command validates document structure, but it does not serve as a pre-implementation adversarial review.

## GOALS

- Make `kit implement` begin with an explicit implementation-readiness gate.
- Require the implementation prompt to run an adversarial preflight across `CONSTITUTION.md`, optional `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` before writing code.
- Require the prompt to surface contradictions, ambiguous requirements, missing edge cases, missing task coverage, hidden assumptions, and scope creep before execution.
- Require the prompted agent to stop coding when the readiness gate fails, update the canonical docs first, and rerun the readiness check before implementation begins.
- Keep the change as workflow semantics only: no new lifecycle phase, no new feature artifact, and no new top-level command.
- Keep `kit check` structural for v1 rather than turning it into a semantic/adversarial validator.
- Update workflow docs and user-facing wording so `kit implement` clearly communicates the readiness gate.
- Keep scaffolded repository instruction templates aligned with the readiness gate so `kit scaffold-agent` and `kit scaffold-agents` generate current workflow guidance.

## NON-GOALS

- Adding a new canonical phase between `tasks` and `implement`.
- Adding a new public command such as `kit preflight` in v1.
- Adding a user-configurable Git-hook-style extension system.
- Creating a new markdown artifact such as `READINESS.md` or `PREFLIGHT.md`.
- Changing `kit reflect` semantics.

## USERS

- Users who want to challenge specs, plans, and tasks before code is written.
- Coding agents that need a stricter go/no-go contract before implementation.
- Maintainers who want more resilience and correctness without bloating the visible lifecycle.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER                       | REQUIRED |
| ----- | ------ | ---- | ----------------------------- | -------- |
| none  | n/a    | n/a  | no additional skills required | no       |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical workflow rules and implementation gate contract | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | shipped workflow summary and gate wording | active |
| implement command | code | `pkg/cli/implement.go` | readiness-gate prompt behavior | active |
| status command | code | `pkg/cli/status.go` | next-step wording for completed work | active |
| repository instruction templates | code | `internal/templates/templates.go` | scaffolded workflow guidance | active |
| README | doc | `README.md` | user-facing workflow wording | active |

## REQUIREMENTS

- `kit implement` must describe an explicit implementation-readiness gate before any coding instructions.
- The readiness gate must be described in user-facing workflow language as an `implementation readiness gate`.
- The prompt may describe the internal challenge pass as an `adversarial preflight`.
- The readiness gate must instruct the agent to read `CONSTITUTION.md` first, then optional `BRAINSTORM.md`, then `SPEC.md`, `PLAN.md`, and `TASKS.md` before coding.
- The readiness gate must require the agent to challenge the document set for:
  - contradictions across documents
  - ambiguous or underspecified requirements
  - hidden assumptions
  - missing edge cases or failure modes
  - tasks that are too broad, missing, or not clearly traceable to the plan/spec
  - scope creep or requirements invented outside the binding docs
- The readiness gate must require a go/no-go decision before implementation starts.
- If the readiness gate fails, the prompt must instruct the agent to:
  - stop coding
  - update `SPEC.md`, `PLAN.md`, and/or `TASKS.md` first
  - refresh `PROJECT_PROGRESS_SUMMARY.md` when feature state or summary changes
  - rerun the readiness gate before implementation starts
- If the readiness gate passes, the prompt must instruct the agent to begin with the first incomplete task in `TASKS.md`.
- `kit implement` must remain a prompt-producing command only; Kit itself must not edit docs or code automatically.
- `kit check` must remain focused on structural/document validation in v1.
- The visible workflow model must remain `brainstorm → spec → plan → tasks → implement → reflect`.
- `kit status` may update its next-step wording to mention the readiness gate, but it must not introduce a new phase or readiness substate.
- `README.md`, `docs/CONSTITUTION.md`, and `docs/specs/0000_INIT_PROJECT.md` must describe the readiness gate consistently with the shipped behavior.
- The repository instruction templates used by `kit scaffold-agent` / `kit scaffold-agents` must include the readiness-gate workflow guidance so regenerated files stay aligned with the checked-in instruction files.

## ACCEPTANCE

- Running `kit implement <feature>` outputs a prompt that clearly begins with an implementation-readiness gate before code execution.
- The prompt tells the agent to run an adversarial preflight against the feature docs and the constitution.
- The prompt tells the agent to stop coding and update canonical docs first when the gate fails.
- The prompt tells the agent to rerun the readiness gate after updating docs and only then continue to implementation.
- The prompt tells the agent to start with the first incomplete task only after the gate passes.
- `kit status` guidance for task-complete features can mention the readiness gate without creating a new lifecycle state.
- `kit check` behavior remains structural in this feature.
- Workflow docs describe the readiness gate without adding a new phase.
- `kit scaffold-agent` and `kit scaffold-agents` generate instruction files that include the readiness gate.
- Automated tests cover the new `kit implement` prompt contract and any status wording changes.

## EDGE-CASES

- The feature has `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md`, but they disagree on scope.
- `TASKS.md` marks work complete even though acceptance criteria are missing from upstream docs.
- The adversarial preflight finds that the plan omits a failure mode implied by the spec.
- The prompt is regenerated with `--prompt-only` and must still include the readiness gate.
- `TASKS.md` exists but contains no incomplete checkboxes.

## OPEN-QUESTIONS

- none
