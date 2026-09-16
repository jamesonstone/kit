# Codex Host Adapter

Applies only when the active coding host is Codex. All other hosts skip this file.

- Thread initialization: ordered rename then pin before the first commentary, planning, inspection, or command; verify from returned host state; on non-success begin with `Thread initialization: rename <status>; pin <status>.` See `docs/references/rules/codex-thread-initialization.md` for the normative contract.
- Browser: use the host built-in browser surface; do not control the user's active profile or launch external Chromium via automation unless the user explicitly requests it; terminate and verify task-owned browser processes before finishing.
- Subagents: inspect the live roster before delegating; root may use host-exposed launch, per-child configuration, follow-up, and wait controls; children must not spawn descendants; resolve semantic profiles from the live roster, never from static model IDs.
- Capability descriptors: strongest justified configuration for `architect` and `precision`; balanced read-heavy configuration for `orchestrator` and `mapper`; fast configuration for bounded `specialist` work; fresh strong configuration for `verifier`.

Conversation naming uses [the shared policy](../thread-naming.md), including
semantic evolution, exact child identity and native metadata-only setters.
