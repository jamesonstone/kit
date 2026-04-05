# PLAN

## SUMMARY

- Update the canonical spec contract, repo instruction templates, and prompt builders so `kit spec` performs skills discovery and `SPEC.md` stores the selected skills.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Update document validation and templates so `SPEC.md` requires a `## SKILLS` section with the fixed table shape and default row.
- [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Update `kit spec` prompt generation to treat skills discovery as a first-class phase and to name repo-local and documented global inputs explicitly.
- [PLAN-02A][SPEC-08][SPEC-09][SPEC-10][SPEC-11] Keep `SPEC.md` dependency inventories separate from `## SKILLS` and require exact locations for design dependencies.
- [PLAN-03][SPEC-12] Add a shared prompt suffix that tells coding agents to consult documented skills before execution, so every prompt-output command stays aligned.
- [PLAN-04][SPEC-13][SPEC-14] Update repository instruction templates and checked-in instruction files to describe the new workflow and to keep `.claude/skills` mirror-only.

## COMPONENTS

- `internal/document/document.go`
- `internal/templates/templates.go`
- `pkg/cli/spec.go`
- `pkg/cli/subagents.go`
- `pkg/cli/skills_prompt.go`
- `AGENTS.md`
- `CLAUDE.md`
- `.github/copilot-instructions.md`

## DATA

- Canonical repo-local skill root: `.agents/skills/*/SKILL.md`
- Documented global inputs:
  - `~/.claude/CLAUDE.md`
  - `${CODEX_HOME}/AGENTS.md`
  - `${CODEX_HOME}/instructions.md`
  - `${CODEX_HOME}/skills/*/SKILL.md`
- Feature-specific selected skills persist only in `SPEC.md`.

## INTERFACES

- `SPEC.md` gains a required `## SKILLS` section and a dependency inventory for newly generated or touched specs.
- `kit spec` prompt contract adds a mandatory skills discovery phase.
- Prompt-output commands inherit a shared `## Skills` instruction block.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical workflow and section requirements | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical skills and dependency inventory contract | active |
| spec template | code | `internal/templates/templates.go` | required `SPEC.md` and prompt section shapes | active |
| document validation | code | `internal/document/document.go` | required section parsing and validation | active |
| spec prompt flow | code | `pkg/cli/spec.go` | skills discovery prompt content | active |
| shared skills guidance | code | `pkg/cli/subagents.go` | prompt-output skills instruction block | active |
| repository instruction files | doc | `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md` | checked-in workflow contract alignment | active |

## RISKS

- Generic prompt augmentation could become noisy if the shared skills note is too verbose.
- The spec prompt could become harder to follow if skills discovery is not ordered clearly before sign-off.
- Repository instruction files can drift from templates if checked-in files are not updated together.

## TESTING

- Verify the `SPEC.md` template includes the new `## SKILLS` section and table.
- Verify `kit spec` prompt output names all required discovery inputs, the `## SKILLS` table requirement, and the separate dependency inventory rules.
- Verify the shared prompt suffix appears on prompt-output commands.
- Verify repository instruction templates mention the new workflow and do not describe `.claude/skills` as canonical.
