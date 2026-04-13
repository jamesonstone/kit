# Agents Docs

## Purpose

- This directory is the repo-local knowledge entrypoint for coding agents
- Top-level instruction files should route here instead of carrying the full operating manual
- Start with this index, then read only the linked docs needed for the current task

## Read Order

1. `docs/CONSTITUTION.md`
2. `docs/agents/WORKFLOWS.md`
3. `docs/agents/GUARDRAILS.md`
4. `docs/agents/TOOLING.md`
5. `docs/agents/RLM.md` when the task is repository-scale
6. relevant `docs/specs/<feature>/` docs when work is feature-scoped
7. `docs/references/*` only when a durable repo-local reference is relevant

## Directory Map

- `WORKFLOWS.md` — when to use the spec-driven or ad hoc path
- `RLM.md` — Kit's repository-scale context-routing pattern for broad discovery with progressive disclosure
- `TOOLING.md` — skills, dispatch, worktrees, and secondary global inputs
- `GUARDRAILS.md` — hard constraints, completion bar, and safety rules

## System Of Record

- Feature requirements, plans, and tasks live under `docs/specs/<feature>/`
- Broader repo references live under `docs/references/`
- Keep durable guidance here instead of expanding the injected top-level instruction files
