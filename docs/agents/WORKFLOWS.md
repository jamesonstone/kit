# Workflows

## Spec-Driven Work

- Use this path for new features, substantial behavioral changes, cross-component changes, or work that already has feature docs
- Do not load every artifact up front
- Start from `TASKS.md` to identify the next action
- Recurse into the relevant `PLAN.md` section for approach
- Recurse into the relevant `SPEC.md` requirement for scope and acceptance
- Use `BRAINSTORM.md` only for unresolved rationale
- Use prior feature docs only through explicit dependency or relationship links
- Ask clarification questions until confidence is high and unresolved assumptions are zero
- Run the implementation readiness gate before writing code
- Update docs first when the implementation changes behavior, requirements, or approach

## Source Of Truth

Authority order:

1. safety and permission constraints
2. current user request
3. `docs/CONSTITUTION.md`
4. `SPEC.md`
5. `PLAN.md`
6. `TASKS.md`
7. `BRAINSTORM.md`
8. repo conventions

Execution order for feature work:

1. `TASKS.md`
2. relevant `PLAN.md` section
3. relevant `SPEC.md` requirement
4. `docs/CONSTITUTION.md` only when needed

- `TASKS.md` controls next action
- `PLAN.md` controls approach
- `SPEC.md` controls requirements
- `CONSTITUTION.md` controls project invariants
- `BRAINSTORM.md` is non-binding research context

## Ad Hoc Work

- Use this path for contained bug fixes, reviews, dependency updates, config changes, or small refinements
- Inspect relevant files before editing
- Use existing repo patterns
- Verify directly with the smallest relevant checks
- Do not create feature docs unless scope requires it
- Update only the practical docs that changed, unless existing feature docs must also change

## Readiness Gate

- Challenge the active docs for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, and scope creep
- If the gate fails, update the canonical docs first, then continue

## Feature Docs

- `docs/specs/<feature>/` remains the source of truth for feature-scoped work
- Keep dependencies, relationships, and skills tables current when those docs are touched
