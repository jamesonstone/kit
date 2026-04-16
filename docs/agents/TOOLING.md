# Tooling

## Skills

- Repo-local canonical skills live under `.agents/skills/*/SKILL.md`
- For feature-scoped work, start with the current feature's `SPEC.md` `## SKILLS` table
- Keep the selected skill set minimal and actionable

## Dispatch

- Use `kit dispatch` when broad work must be turned into safe multi-lane execution
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep repository-scale discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Keep the main agent responsible for synthesis, integration, validation, and communication

## Worktrees

- When isolated checkouts are needed, keep worktrees flat under `~/worktrees/`
- Use `git worktree add ~/worktrees/<repo>-<branch> <branch>` or `git worktree add -b <branch> ~/worktrees/<repo>-<branch> <base-ref>`
- New feature numbers should stay numeric and human-readable; do not renumber them to match dependency order
- When Git common dir state is available, Kit reserves the next feature number from shared clone-local allocator state so sibling worktrees do not collide
- Dependency order should come from `builds on` and `depends on` relationships, not from rewriting directory prefixes

## Secondary Global Inputs

- `~/.claude/CLAUDE.md`
- `${CODEX_HOME}/AGENTS.md`
- `${CODEX_HOME}/instructions.md`
- `${CODEX_HOME}/skills/*/SKILL.md`

- Treat these as secondary context after repo-local docs
- Do not use `.claude/skills` as canonical discovery input
