# Agents Docs

## Purpose

- Start here for repo-local agent guidance
- Classify the task, then load only the linked doc needed for the current decision
- Avoid reading all agent docs by default

## Runtime Routing

- `WORKFLOWS.md` → classify spec-driven vs ad hoc work and resolve source-of-truth order
- `GUARDRAILS.md` → completion, safety, validation, and hard rules
- `RLM.md` → context routing and progressive disclosure
- `TOOLING.md` → skills, dispatch, worktrees, and secondary inputs
- `docs/references/*` → durable reference material only when relevant
- `docs/specs/<feature>/*` → active feature artifacts only

## Loading Rule

- Identify the immediate decision before opening another file
- Prefer a specific section over a full file
- Stop loading once the decision is supported
- Treat repo-local docs as primary and global model/vendor instructions as secondary

## System Of Record

- Feature requirements, plans, and tasks live under `docs/specs/<feature>/`
- Broader repo references live under `docs/references/`
- Keep durable guidance here instead of expanding the injected top-level instruction files
