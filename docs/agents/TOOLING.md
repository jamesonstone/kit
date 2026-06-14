# Tooling

## Skills

- Repo-local canonical skills live under `.agents/skills/*/SKILL.md`
- For feature-scoped work, start with the current feature's canonical front matter `skills`, falling back to the legacy `SPEC.md` `## SKILLS` table only when front matter is absent
- Keep the selected skill set minimal and actionable

## Command Capability Discovery

- Use `kit capabilities` when choosing among Kit commands and the mutation, network, write, or git behavior is not already obvious.
- Use `kit capabilities <command> --json` for one command path, including nested paths such as `rules add` or `skill mine`.
- Use `kit capabilities --search <term> --json` for compact filtered discovery, and `kit capabilities --full --json` only when hidden or deprecated compatibility commands matter.
- Treat `kit capabilities` itself as read-only: it does not require a Kit project root and does not load project config, write files, call the network, run subprocesses, or mutate git.
- When adding or extending any Kit command, subcommand, flag, alias, prompt surface, or command behavior, update `kit capabilities` in the same change.
- Load `docs/references/rules/command-capabilities.md` before command-surface work so command metadata stays complete for future coding agents.

## Dispatch

- Use `kit dispatch` when broad work must be turned into safe multi-lane execution
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep broad or noisy discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Keep the main agent responsible for synthesis, integration, validation, and communication

## Review Loop

- Use `kit review-loop --pr <target> --coderabbit` when current unresolved CodeRabbit PR review feedback should become a human-reviewed dispatch prompt.
- Use `kit review-loop --pr <target> --watch` to wait for CodeRabbit completion on the current PR head before collecting review feedback.
- `kit dispatch --loop --pr <target>` is an alias for the same review-loop workflow.
- Use `kit dispatch --pr <target> --coderabbit` only when you need raw unresolved CodeRabbit review-thread intake without review-loop watch, classification, or summary behavior.
- Treat review-loop as read-only by default: it may read GitHub through `gh`, open an editor, and copy output, but it must not stage, commit, push, post PR comments, or resolve review threads.
- After fixes or no-op decisions are complete, `kit dispatch --pr <target> --resolve --yes` may resolve currently matching unresolved review threads on GitHub; this is an explicit GitHub mutation and must not be run speculatively.

## Project Directory

- Work in the existing project directory by default.
- Do not create or use git worktrees for agent work.
- If the current branch or dirty state is unsuitable, stop and ask the user how to proceed instead of creating an alternate checkout.
- New feature numbers should stay numeric and human-readable; do not renumber them to match dependency order
- Dependency order should come from `builds on` and `depends on` relationships, not from rewriting directory prefixes

## Secondary Global Inputs

- `~/.claude/CLAUDE.md`
- `${CODEX_HOME}/AGENTS.md`
- `${CODEX_HOME}/instructions.md`
- `${CODEX_HOME}/skills/*/SKILL.md`

- Treat these as secondary context after repo-local docs
- Do not use `.claude/skills` as canonical discovery input
