# Workflows

## Spec-Driven Work

- Use this path for new features, substantial behavioral changes, cross-component changes, or work that already has feature docs
- Do not load every artifact up front
- In v2 feature work, start from `SPEC.md`; it is the single durable feature artifact
- Use `SPEC.md` sections for thesis, context, clarifications, requirements, assumptions, acceptance criteria, implementation plan, task checklist, validation map, reflection notes, documentation updates, delivery decision, and evidence
- Treat legacy `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` as historical context unless the user explicitly chooses a legacy staged command
- Use `BRAINSTORM.md` only for unresolved historical rationale
- Use `PLAN.md` and `TASKS.md` only for legacy staged flows or historical comparison
- Use prior feature docs only through explicit dependency or relationship links
- Ask clarification questions until confidence is high and unresolved assumptions are zero
- Run the v2 readiness gates before writing code: clarification complete, acceptance criteria binary-verifiable, task checklist mapped to criteria, validation mapped 1:1, delivery intent known
- Update docs first when the implementation changes behavior, requirements, or approach

## Source Of Truth

Authority order:

1. safety and permission constraints
2. current user request
3. `docs/CONSTITUTION.md`
4. `SPEC.md`
5. legacy `PLAN.md` / `TASKS.md` when the user explicitly chooses a staged flow
6. legacy `BRAINSTORM.md`
7. repo conventions

Execution order for feature work:

1. `SPEC.md`
2. relevant `SPEC.md` task checklist item, acceptance criterion, and validation map entry
3. legacy staged artifacts only when explicitly operating in a legacy staged flow
4. `docs/CONSTITUTION.md` only when needed

- `SPEC.md` controls requirements, plan, tasks, validation, reflection, delivery, and evidence
- `CONSTITUTION.md` controls project invariants
- `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` are non-binding historical context in v2 unless the user chooses a legacy staged flow

## Ad Hoc Work

- Use this path for contained bug fixes, reviews, dependency updates, config changes, or small refinements
- Inspect relevant files before editing
- Use existing repo patterns
- Verify directly with the smallest relevant checks
- Prefer declared `kit legacy verify` checks when legacy staged verification exists; otherwise run the smallest direct command that proves the change
- Do not create feature docs unless scope requires it
- Update only the practical docs that changed, unless existing feature docs must also change

## Readiness Gate

- Challenge `SPEC.md` for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, validation gaps, delivery ambiguity, and scope creep
- If the gate fails, update the canonical docs first, then continue

## Feature Docs

- `docs/specs/<feature>/` remains the source of truth for feature-scoped work
- v2 feature work keeps durable workflow state in `SPEC.md`
- `SPEC.md` front matter should include `workflow_version: 2` and a current `phase`
- Keep references, relationships, and skills metadata current when those docs are touched
