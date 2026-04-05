# TASKS

## PROGRESS TABLE

| ID   | TASK                                                                     | STATUS | OWNER | DEPENDENCIES     |
| ---- | ------------------------------------------------------------------------ | ------ | ----- | ---------------- |
| T001 | Update canonical spec contracts for `SPEC.md` skills discovery           | done   | agent |                  |
| T002 | Update `kit spec` prompt generation to require skills discovery          | done   | agent | T001             |
| T003 | Add shared execution-time skills guidance to prompt-output commands      | done   | agent | T002             |
| T004 | Update repository instruction templates and checked-in instruction files | done   | agent | T001             |
| T005 | Add tests for templates and spec prompt output                           | done   | agent | T002, T003, T004 |
| T006 | Split broader spec dependencies from execution-time skills              | done   | agent | T002, T005       |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Update canonical spec contracts for `SPEC.md` skills discovery
- [x] T002: Update `kit spec` prompt generation to require skills discovery
- [x] T003: Add shared execution-time skills guidance to prompt-output commands
- [x] T004: Update repository instruction templates and checked-in instruction files
- [x] T005: Add tests for templates and spec prompt output
- [x] T006: Split broader spec dependencies from execution-time skills

## TASK DETAILS

### T001

- **GOAL**: make the `SPEC.md` skills section part of the canonical document contract
- **SCOPE**: update `docs/CONSTITUTION.md`, `docs/specs/0000_INIT_PROJECT.md`, `internal/document/document.go`, and `internal/templates/templates.go`
- **ACCEPTANCE**: spec validation and templates both require `## SKILLS`
- **NOTES**: keep the table shape fixed and include the default no-skill row

### T002

- **GOAL**: make `kit spec` instruct agents to perform skills discovery before sign-off
- **SCOPE**: update interactive and template spec prompt output
- **ACCEPTANCE**: prompt names repo-local skills, documented global inputs, and the requirement to populate `SPEC.md` `## SKILLS`

### T003

- **GOAL**: ensure prompt-output commands tell agents to use documented skills during execution
- **SCOPE**: add one shared prompt suffix applied through the prompt output pipeline
- **ACCEPTANCE**: prompt-output commands include a `## Skills` section

### T004

- **GOAL**: keep checked-in instruction files aligned with the new workflow
- **SCOPE**: update `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md`
- **ACCEPTANCE**: each file mentions the `kit spec` skills discovery phase and `SPEC.md` `## SKILLS`

### T005

- **GOAL**: verify templates and prompt output changed as intended
- **SCOPE**: add or update unit tests for spec prompts, shared prompt suffix, and instruction templates
- **ACCEPTANCE**: tests fail if required discovery inputs or `## SKILLS` instructions are removed

### T006

- **GOAL**: keep the spec workflow explicit about broader supporting dependencies without overloading `## SKILLS`
- **SCOPE**: update `SPEC.md` templates and prompt guidance to add a separate dependency inventory
- **ACCEPTANCE**: `kit spec` prompts keep `## SKILLS` and `## DEPENDENCIES` separate and require exact design locations

## DEPENDENCIES

- T001 must land before prompt or template tests can assert the new contract.
- T002 and T003 both depend on the updated contract from T001.
- T004 depends on T001 because the checked-in instruction files mirror the canonical docs.
- T005 depends on the completed behavior from T002, T003, and T004.

## NOTES

- This feature stays prompt-driven and document-driven.
- No runtime skill execution or automatic skill resolution is added.
