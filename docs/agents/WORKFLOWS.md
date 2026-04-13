# Workflows

## Spec-Driven Work

- Use this path for new features, substantial behavioral changes, cross-component changes, or work that already has feature docs
- Read `BRAINSTORM.md` when present, then `SPEC.md`, `PLAN.md`, and `TASKS.md`
- Ask clarification questions until confidence is high and unresolved assumptions are zero
- Run the implementation readiness gate before writing code
- Update docs first when the implementation changes behavior, requirements, or approach

## Ad Hoc Work

- Use this path for contained bug fixes, reviews, dependency updates, config changes, or small refinements
- Follow understand -> implement -> verify
- Update only the practical docs that changed, unless existing feature docs must also change

## Readiness Gate

- Challenge the active docs for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, and scope creep
- If the gate fails, update the canonical docs first, then continue

## Feature Docs

- `docs/specs/<feature>/` remains the source of truth for feature-scoped work
- Keep dependencies, relationships, and skills tables current when those docs are touched
