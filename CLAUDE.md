# CLAUDE

## Purpose

- This file is a table of contents, not the full manual
- Repo-local markdown under `docs/` is the system of record
- Start here, then load only the docs relevant to the current task

## Read This Next

- `docs/agents/README.md` for the repo-local entrypoint
- `docs/CONSTITUTION.md` for project-wide constraints and invariants
- the relevant `docs/specs/<feature>/` files when the work is feature-scoped

## Work Routing

- Use `docs/agents/WORKFLOWS.md` to classify work as spec-driven or ad hoc
- Use `docs/agents/RLM.md` when the task is repository-scale or needs broad discovery; RLM is Kit's repository-scale context-routing pattern
- Use `docs/agents/GUARDRAILS.md` for hard constraints that must not be relaxed

## Repo Knowledge Map

- `docs/agents/README.md` — entrypoint and navigation guide
- `docs/agents/WORKFLOWS.md` — spec-driven, ad hoc, and readiness-gate flow
- `docs/agents/RLM.md` — repository-scale discovery and progressive disclosure
- `docs/agents/TOOLING.md` — skills, dispatch, worktrees, and secondary global inputs
- `docs/agents/GUARDRAILS.md` — completion bar, safety rules, and non-negotiable invariants
- `docs/references/README.md` — durable repo-local references that are broader than one feature
- `docs/specs/<feature>/` — feature-level source of truth for requirements, plans, and tasks

## Runtime Context

- Start small and load only the docs, skills, and files relevant to the task
- Treat global Claude or Codex files as secondary inputs after repo-local docs
- If docs, skills, or references materially shape a feature, record them in the feature dependency tables

## Tool-Specific Notes

- CLAUDE should stay short and stable so it fits easily into injected context
- Put durable workflow guidance in `docs/agents/*` rather than expanding this file
