# TASKS

## PROGRESS TABLE

| ID   | TASK                                                                      | STATUS | OWNER | DEPENDENCIES           |
| ---- | ------------------------------------------------------------------------- | ------ | ----- | ---------------------- |
| T001 | Add brainstorm workflow spec docs and canonical repo references           | done   | agent |                        |
| T002 | Add brainstorm artifact template and brainstorm phase detection           | done   | agent | T001                   |
| T003 | Rebuild `kit brainstorm` as interactive planning-only command             | done   | agent | T002                   |
| T004 | Thread `BRAINSTORM.md` through downstream prompt commands                 | done   | agent | T002                   |
| T005 | Remove `kit oneshot`, git branch automation, and stale config/code        | done   | agent | T001                   |
| T006 | Update root/help/status/handoff/rollup/docs for brainstorm-first workflow | done   | agent | T002, T005             |
| T007 | Add tests and run verification                                            | done   | agent | T003, T004, T005, T006 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Add brainstorm workflow spec docs and canonical repo references
- [x] T002: Add brainstorm artifact template and brainstorm phase detection
- [x] T003: Rebuild `kit brainstorm` as interactive planning-only command
- [x] T004: Thread `BRAINSTORM.md` through downstream prompt commands
- [x] T005: Remove `kit oneshot`, git branch automation, and stale config/code
- [x] T006: Update root/help/status/handoff/rollup/docs for brainstorm-first workflow
- [x] T007: Add tests and run verification

## TASK DETAILS

### T001

- **GOAL**: Create the formal feature documents for the brainstorm-first workflow change
- **SCOPE**:
  - add `docs/specs/0004-brainstorm-first-workflow/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - all three files exist with required sections
  - workflow decisions are captured in repo docs
- **NOTES**: complete when docs are committed to the working tree

### T002

- **GOAL**: Model `BRAINSTORM.md` and brainstorm-only features explicitly in core feature logic
- **SCOPE**:
  - add brainstorm template
  - add brainstorm phase constant and detection
  - update feature/status/rollup helpers to understand `BRAINSTORM.md`
- **ACCEPTANCE**:
  - features with only `BRAINSTORM.md` are not reported as `spec`
  - rollup/status logic can reference brainstorm artifacts safely
- **NOTES**: [PLAN-DATA], [PLAN-COMPONENTS]

### T003

- **GOAL**: Turn `kit brainstorm` into the interactive planning-only entrypoint
- **SCOPE**:
  - ask for feature name and thesis
  - create or reuse the numbered feature directory
  - create `BRAINSTORM.md`
  - output a `/plan` prompt with planning-only instructions
- **ACCEPTANCE**:
  - prompt starts with `/plan`
  - prompt instructs the coding agent to avoid implementation and use numbered batched clarification with recommended defaults, `yes` / `y` whole-batch approval, `yes 3, 4, 5` / `y 3, 4, 5` numbered approval, `no` / `n` overrides, uncertainties, and percentage-understanding progress until the specification is precise enough for a production-quality solution
  - `BRAINSTORM.md` is the persistence target
- **NOTES**: [PLAN-INTERFACES], [PLAN-APPROACH]

### T004

- **GOAL**: Make downstream workflow prompts consume `BRAINSTORM.md` when present
- **SCOPE**:
  - update `spec`, `plan`, `tasks`, `implement`, and `reflect`
  - keep `BRAINSTORM.md` optional
- **ACCEPTANCE**:
  - prompts include `BRAINSTORM.md` file references and guidance when present
  - prompts that use the `>=95%` clarification loop preserve the shared approval semantics across `spec`, `plan`, and `tasks`
  - commands still function when `BRAINSTORM.md` is absent
- **NOTES**: [PLAN-COMPONENTS], [PLAN-INTERFACES]

### T005

- **GOAL**: Remove parallel workflow and branch automation concepts from the product surface
- **SCOPE**:
  - delete `oneshot` command
  - remove branch-related flags and config
  - remove obsolete git helper code if unused
- **ACCEPTANCE**:
  - CLI/help no longer exposes `oneshot` or branch creation flags
  - config/schema/docs no longer describe branch automation
- **NOTES**: [PLAN-APPROACH], [PLAN-RISKS]

### T006

- **GOAL**: Align product messaging and reporting with the brainstorm-first workflow
- **SCOPE**:
  - update help text, status, handoff, rollup, README, constitution, and generated templates
  - show brainstorming as optional before spec
  - keep `kit status` output informative by including the running Kit version as a minor metadata line
- **ACCEPTANCE**:
  - visible workflow messaging is internally consistent
  - brainstorm-only features produce correct next-step guidance
  - `kit status` includes the running Kit version without overwhelming the main feature output
- **NOTES**: [PLAN-INTERFACES], [PLAN-RISKS]

### T007

- **GOAL**: Prove the refactor works and prevent regression
- **SCOPE**:
  - add/update tests
  - run `go test ./...`
  - search for stale `oneshot`, branch, and outdated clarification-loop references
- **ACCEPTANCE**:
  - tests pass
  - targeted regressions are covered by tests
  - stale references are removed or intentionally documented
- **NOTES**: [PLAN-TESTING]

## DEPENDENCIES

- T002 precedes T003 and T004 because prompt/status behavior depends on brainstorm artifact modeling
- T005 should land before T006 so docs/help reflect the final product surface
- T007 is the final gate after all behavior and docs are updated

## NOTES

- `BRAINSTORM.md` is optional but canonical when present
- brainstorming remains planning-only and does not execute implementation work
