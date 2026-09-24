# Host Adapters

Host-specific bindings live here so normative core files stay model-agnostic.

- Core uses only semantic profiles: `architect`, `orchestrator`, `mapper`, `specialist`, `precision`, `verifier`.
- Each host file translates those profiles into controls the active host confirms at runtime.
- Never copy model IDs, version names, or host operations into `AGENTS.md`, `CLAUDE.md`, `docs/agents/*`, or normative rules.
- Normative topology remains in `docs/references/rules/agent-team-orchestration.md`.

Files:

- `codex.md` — Codex browser and subagent bindings.
- `claude-code.md` — Claude Code capability descriptors without model-class pins.
- `copilot.md` — GitHub Copilot conservative profile usage.
- `warp.md` — Warp and Oz native orchestration notes.
