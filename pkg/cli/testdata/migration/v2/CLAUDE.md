# CLAUDE

## Purpose

- This file is a routing table, not the full manual
- Start at `docs/agents/README.md`, then load only the docs needed for the current decision
- Repo-local markdown under `docs/` is the system of record

## Pasted Text Attachments

- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for

## Runtime Routing

- `docs/agents/README.md` — classify the task and choose the next document
- `docs/agents/WORKFLOWS.md` — spec-driven versus ad hoc flow
- `docs/agents/GUARDRAILS.md` — completion, safety, and hard rules
- `docs/agents/RLM.md` — just-in-time context loading when broad context would be noisy
- `docs/agents/TOOLING.md` — skills, dispatch, project-directory workflow, and secondary inputs

## Conditional Context

- `docs/specs/<feature>/` — active feature artifacts only
- `docs/references/README.md` — durable repo references only when relevant
- `docs/CONSTITUTION.md` — project invariants when a decision depends on them

## Repo Knowledge Map

- `docs/agents/README.md` — runtime routing index
- `docs/agents/WORKFLOWS.md` — work classification and source-of-truth semantics
- `docs/agents/RLM.md` — progressive disclosure and context budget rules
- `docs/agents/TOOLING.md` — skills, dispatch, project-directory workflow, and secondary global inputs
- `docs/agents/GUARDRAILS.md` — completion bar, safety rules, and validation expectations
- `docs/references/README.md` — durable repo-local references that are broader than one feature
- `docs/specs/<feature>/` — feature-level source of truth for requirements, plans, and tasks

## Constraints

- Keep CLAUDE short and stable so it fits easily into injected context
- Put durable workflow guidance in `docs/agents/*` rather than expanding this file
- Do not add an always-loaded monolithic instruction file
