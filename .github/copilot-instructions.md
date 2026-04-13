# GitHub Copilot Repository Instructions

## Quick Rules

- Use this file as a map, not the full manual
- Start with `docs/agents/README.md` and then open only the relevant linked docs
- Treat `docs/specs/<feature>/` as the feature system of record
- Use `docs/agents/RLM.md` only when the task is repository-scale; RLM is Kit's repository-scale context-routing pattern
- Keep context minimal and source-attributed

## Fallback Read Order

- If linked-doc traversal is weak or unavailable, read `docs/CONSTITUTION.md` first
- Then read the relevant `docs/specs/<feature>/` docs for the task in scope
- Use `docs/agents/RLM.md` only when the task is broad enough that loading the whole repo context would be noisy or wasteful

## Non-Negotiable Rules

- Repo-local docs under `docs/` are the source of truth
- Always update affected documentation and keep touched docs properly formatted
- Keep context minimal and load only the docs and files relevant to the task
- Remove dead code and unnecessary exports or public surface when they are not strictly needed
- Do not treat `.claude/skills` as canonical discovery input

## Repo Knowledge Map

- `docs/agents/README.md` — repo-local entrypoint
- `docs/agents/WORKFLOWS.md` — work classification and execution flow
- `docs/agents/RLM.md` — progressive-disclosure pattern for broad discovery
- `docs/agents/TOOLING.md` — skills, dispatch, worktrees, and secondary globals
- `docs/agents/GUARDRAILS.md` — hard rules and completion bar
- `docs/references/README.md` — durable repo-local references
- `docs/specs/<feature>/` — feature source of truth
