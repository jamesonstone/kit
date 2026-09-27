# Tooling

## Skills

- Repo-local canonical skills live under `.agents/skills/*/SKILL.md`
- For feature-scoped work, start with the current feature's canonical front matter `skills`, falling back to the legacy `SPEC.md` `## SKILLS` table only when front matter is absent
- Keep the selected skill set minimal and actionable

## Dispatch

- Use `kit dispatch` when broad work must be turned into safe multi-lane execution
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep broad or noisy discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Keep the main agent responsible for synthesis, integration, validation, and communication

## Project Directory

- Work in the existing project directory by default
- Do not create or use git worktrees for agent work
- If the current branch or dirty state is unsuitable, stop and ask the user how to proceed instead of creating an alternate checkout

## Secondary Global Inputs

- `~/.claude/CLAUDE.md`
- `${CODEX_HOME}/AGENTS.md`
- `${CODEX_HOME}/instructions.md`
- `${CODEX_HOME}/skills/*/SKILL.md`

- Treat these as secondary context after repo-local docs
- Do not use `.claude/skills` as canonical discovery input
