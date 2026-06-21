# Agents Docs

## Purpose

- Start here for repo-local agent guidance
- Classify the task, then load only the linked doc needed for the current decision
- Avoid reading all agent docs by default

## Runtime Routing

- `WORKFLOWS.md` → classify spec-driven vs ad hoc work and resolve source-of-truth order
- `GUARDRAILS.md` → completion, safety, validation, and hard rules
- `RLM.md` → context routing and progressive disclosure
- `TOOLING.md` → skills, command capability discovery, dispatch, project-directory workflow, and secondary inputs
- `docs/references/*` → durable reference material only when relevant
- `docs/references/rules/*` → durable rulesets only when linked from feature references or directly relevant
- `docs/specs/<feature>/*` → active feature artifacts only

## Loading Rule

- Identify the immediate decision before opening another file
- Prefer a specific section over a full file
- Stop loading once the decision is supported
- Treat repo-local docs as primary and global model/vendor instructions as secondary

## System Of Record

- V2 feature requirements, implementation plan, task checklist, validation map, reflection notes, delivery decision, and evidence live in `docs/specs/<feature>/SPEC.md`
- Legacy staged `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files may exist as historical context or when a legacy staged command is explicitly used
- Broader repo references live under `docs/references/`
- Durable repo-local rulesets live under `docs/references/rules/` and should be pointer-loaded through feature references
- Keep durable guidance here instead of expanding the injected top-level instruction files
