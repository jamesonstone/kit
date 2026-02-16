# CLAUDE

## Source of truth

- Primary authority for repository workflow, constraints, and change policy: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `SPEC.md` (requirements)
  - `PLAN.md` (implementation plan)
  - `TASKS.md` (executable task list)
  - `ANALYSIS.md` (optional, analysis scratchpad)

## Workflow contract (classification-first)

- Classify every request before acting:
  - **Spec-driven**: use full pipeline for `kit spec` / `kit oneshot`, new features, or substantial changes
  - **Ad hoc**: use lightweight flow for small fixes, reviews, refinements, and mechanical changes
- If ad hoc work touches an existing feature in `docs/specs/<feature>/`, default to updating its spec docs when behavior, requirements, or approach changes
- For ad hoc changes, update only relevant practical docs (e.g., README/API docs) and avoid creating spec artifacts unless needed

## Multi-feature rule

- Never mix features in one `docs/specs/<feature>/` directory.
- If work spans features, update each feature's docs separately.
