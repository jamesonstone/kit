# Cursor Host Adapter

Cursor reads the shared repository AGENTS.md. Conversation naming uses
[the shared policy](../thread-naming.md), with manual history-title or CLI
`/rename` fallback when no safe native setter is exposed. Do not apply the
Codex initialization/pinning gate or invent a title field in Cursor hooks.
