# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                                                       | STATUS | OWNER | DEPENDENCIES           |
| ---- | -------------------------------------------------------------------------- | ------ | ----- | ---------------------- |
| T001 | Add brainstorm workflow spec docs and canonical repo references            | done   | agent |                        |
| T002 | Add brainstorm artifact template and brainstorm phase detection            | done   | agent | T001                   |
| T003 | Rebuild `kit brainstorm` as interactive planning-only command              | done   | agent | T002                   |
| T004 | Thread `BRAINSTORM.md` through downstream prompt commands                  | done   | agent | T002                   |
| T005 | Remove `kit oneshot`, git branch automation, and stale config/code         | done   | agent | T001                   |
| T006 | Update root/help/status/handoff/rollup/docs for brainstorm-first workflow  | done   | agent | T002, T005             |
| T007 | Add tests and run verification                                             | done   | agent | T003, T004, T005, T006 |
| T008 | Add pre-editor instructions and keypress gating for editor-backed input    | done   | agent | T003, T007             |
| T009 | Switch `brainstorm`/`spec`/`plan`/`tasks` to clipboard-first prompt output | done   | agent | T003, T004, T007       |
| T010 | Extend clipboard-first prompt output to `implement` and `reflect`          | done   | agent | T004, T009             |
| T011 | Add phase dependency inventories to brainstorm and plan workflow prompts   | done   | agent | T003, T004             |
| T012 | Add side-effect-free `--prompt-only` regeneration to core workflow commands  | done   | agent | T003, T004, T009, T010 |
| T013 | Make supported multiline free-text flows vim-default with `--inline` opt-out | done   | agent | T003, T004, T008       |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Add brainstorm workflow spec docs and canonical repo references
- [x] T002: Add brainstorm artifact template and brainstorm phase detection
- [x] T003: Rebuild `kit brainstorm` as interactive planning-only command
- [x] T004: Thread `BRAINSTORM.md` through downstream prompt commands
- [x] T005: Remove `kit oneshot`, git branch automation, and stale config/code
- [x] T006: Update root/help/status/handoff/rollup/docs for brainstorm-first workflow
- [x] T007: Add tests and run verification
- [x] T008: Add pre-editor instructions and keypress gating for editor-backed input
- [x] T009: Switch `brainstorm`, `spec`, `plan`, and `tasks` to clipboard-first prompt output
- [x] T010: Extend clipboard-first prompt output to `implement` and `reflect`
- [x] T011: Add phase dependency inventories to brainstorm and plan workflow prompts
- [x] T012: Add side-effect-free `--prompt-only` regeneration to core workflow commands
- [x] T013: Make supported multiline free-text flows vim-default with `--inline` opt-out

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
  - support `Shift+Enter` and `Ctrl+J` for multiline thesis entry
  - support `--vim` and `--editor=vim` for editor-backed thesis entry
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
  - keep free-text prompt behavior consistent between `brainstorm` and `spec --interactive`
  - support the same editor-backed free-text mode in both commands
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
  - cover `Shift+Enter` multiline entry behavior and newline preservation helpers
  - cover editor flag resolution and editor-backed input helpers
  - run `go test ./...`
  - search for stale `oneshot`, branch, and outdated clarification-loop references
- **ACCEPTANCE**:
  - tests pass
  - targeted regressions are covered by tests
  - stale references are removed or intentionally documented
- **NOTES**: [PLAN-TESTING]

### T008

- **GOAL**: Make editor-backed free-text entry clearer before the editor opens
- **SCOPE**:
  - update the shared editor input helper
  - show a short step-specific instruction screen before editor launch
  - wait for any key before opening the editor
- **ACCEPTANCE**:
  - `brainstorm --vim` and `spec --interactive --vim` display step instructions before opening the editor
  - the editor opens only after an explicit key press
  - tests cover the new pre-editor interaction
- **NOTES**: [PLAN-COMPONENTS], [PLAN-TESTING]

### T009

- **GOAL**: Make the planning-stage prompt commands copy to the clipboard by default
- **SCOPE**:
  - update `brainstorm`, `spec`, `plan`, and `tasks`
  - keep raw stdout prompt output behind `--output-only`
  - preserve `--copy` as an explicit override for `--output-only`
  - keep `brainstorm --output <path>` writing files while also copying by default
- **ACCEPTANCE**:
  - default command output acknowledges clipboard copy and does not print the prompt body
  - `--output-only` prints the raw prompt to stdout
  - `--output-only --copy` both prints and copies
  - tests cover the new output behavior
- **NOTES**: [PLAN-COMPONENTS], [PLAN-INTERFACES], [PLAN-TESTING]

### T010

- **GOAL**: Make the execution-stage core workflow commands clipboard-first by default
- **SCOPE**:
  - update `implement` and `reflect`
  - keep raw stdout prompt output behind `--output-only`
  - preserve `--copy` as an explicit override for `--output-only`
- **ACCEPTANCE**:
  - default command output acknowledges clipboard copy and does not print the prompt body
  - `--output-only` prints the raw prompt to stdout
  - `--output-only --copy` both prints and copies
  - tests and verification still pass
- **NOTES**: [PLAN-COMPONENTS], [PLAN-INTERFACES], [PLAN-TESTING]

### T011

- **GOAL**: keep brainstorm and plan artifacts explicit about the resources that shape each phase
- **SCOPE**:
  - update brainstorm prompt requirements
  - update plan prompt requirements
  - add dependency inventory tables to canonical templates
- **ACCEPTANCE**:
  - newly generated or touched `BRAINSTORM.md` docs track phase dependencies
  - `kit plan` prompts require `PLAN.md` to track implementation-strategy dependencies
  - tests cover the new prompt guidance
- **NOTES**: [PLAN-APPROACH], [PLAN-INTERFACES], [PLAN-TESTING]

### T012

- **GOAL**: let users regenerate existing workflow prompts without mutating repo docs
- **SCOPE**:
  - add a shared `--prompt-only` flag to `brainstorm`, `spec`, `plan`, `tasks`, `implement`, and `reflect`
  - make `brainstorm`, `spec`, `plan`, and `tasks` skip scaffolding and rollup writes when `--prompt-only` is used
  - reuse existing-feature selectors for prompt regeneration when no feature argument is provided
  - add regression tests for selector filtering and missing-artifact failures
- **ACCEPTANCE**:
  - `kit brainstorm --prompt-only`, `kit spec --prompt-only`, `kit plan --prompt-only`, and `kit tasks --prompt-only` regenerate prompts for existing eligible features without mutating feature docs or rollups
  - `kit implement --prompt-only` and `kit reflect --prompt-only` are accepted as consistency flags and preserve the current prompt-only behavior
  - tests cover prompt-only regeneration and missing-artifact failures
- **NOTES**: [PLAN-COMPONENTS], [PLAN-INTERFACES], [PLAN-TESTING]

### T013

- **GOAL**: make supported multiline free-text flows open a vim-compatible editor by default
- **SCOPE**:
  - update shared free-text input config
  - make `kit brainstorm` thesis entry editor-default
  - make `kit spec --interactive` answers editor-default
  - add `--inline` as the explicit opt-out for inline-capable flows
  - keep `kit dispatch` editor-default without adding a new inline mode
- **ACCEPTANCE**:
  - `kit brainstorm` thesis entry opens the editor by default and `kit brainstorm --inline` restores terminal multiline entry
  - `kit spec --interactive` opens the editor by default and `kit spec --interactive --inline` restores terminal multiline entry
  - contradictory flag combinations such as `--inline --vim` fail fast
  - tests cover default editor routing and the inline opt-out
- **NOTES**: [PLAN-COMPONENTS], [PLAN-INTERFACES], [PLAN-TESTING]

## DEPENDENCIES

- T002 precedes T003 and T004 because prompt/status behavior depends on brainstorm artifact modeling
- T005 should land before T006 so docs/help reflect the final product surface
- T007 is the final gate after all behavior and docs are updated
- T012 depends on the shipped core workflow prompt surfaces because regeneration reuses those prompt builders instead of introducing a parallel code path
- T013 depends on the existing shared editor helper and brainstorm/spec free-text flows because it changes the default routing instead of introducing a new input system

## NOTES

- `BRAINSTORM.md` is optional but canonical when present
- brainstorming remains planning-only and does not execute implementation work
