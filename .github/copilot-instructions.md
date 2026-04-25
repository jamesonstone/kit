# GitHub Copilot Repository Instructions

## Quick Rules

- Use this file as a map, not the full manual
- Start with `docs/agents/README.md` and then open only the linked docs needed for the current decision
- Treat `docs/specs/<feature>/` as the feature system of record
- Use `docs/agents/RLM.md` when full-context loading would be noisy or wasteful
- Keep context minimal and source-attributed

## Runtime Routing

- `docs/agents/README.md` — classify the task and choose the next document
- `docs/agents/WORKFLOWS.md` — workflow and source-of-truth rules
- `docs/agents/GUARDRAILS.md` — hard completion and safety rules
- `docs/agents/RLM.md` — just-in-time context routing
- `docs/agents/TOOLING.md` — skills, dispatch, worktrees, and secondary inputs

## Non-Negotiable Rules

- Repo-local docs under `docs/` are the source of truth
- Always update affected documentation and keep touched docs properly formatted
- Keep context minimal and load only the docs and files relevant to the task
- Remove dead code and unnecessary exports or public surface when they are not strictly needed
- Do not treat `.claude/skills` as canonical discovery input
- Do not add an always-loaded monolithic instruction file

## Repo Knowledge Map

- `docs/agents/README.md` — repo-local entrypoint
- `docs/agents/WORKFLOWS.md` — work classification and execution flow
- `docs/agents/RLM.md` — progressive-disclosure pattern for broad discovery
- `docs/agents/TOOLING.md` — skills, dispatch, worktrees, and secondary globals
- `docs/agents/GUARDRAILS.md` — hard rules and completion bar
- `docs/references/README.md` — durable repo-local references
- `docs/specs/<feature>/` — feature source of truth
