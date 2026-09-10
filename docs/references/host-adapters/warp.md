# Warp and Oz Host Adapter

Applies only when the active coding host is Warp, Oz, or another host that reads `AGENTS.md` with native orchestration.

- Use native parent-child orchestration, continuation, and per-child configuration only when the host exposes them.
- Parallelism and admission remain host-owned; never invent a numeric cap or probe capacity with disposable work.
- Skip Codex-only thread operations and browser surfaces described in `codex.md`.
