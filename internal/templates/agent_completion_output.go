package templates

const agentCompletionOutputGate = `## Agent Completion Output Contract

- Before a terminal task response, load ` + "`docs/references/rules/agent-completion-output.md`" + ` when present. This contract does not apply to progress commentary or focused clarification questions.
- Make the first human-readable line ` + "`# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>`" + `. A required host wrapper may surround the response, but no human-readable preamble may precede the status.
- Immediately follow with a prioritized action list ordered Blocker, Incomplete, Next, Optional, then None; every PASS includes a None item.
- Make required follow-ups copy-ready. Never leave Why or Continue with blank, and never hide blockers or incomplete work below completed detail.
- Use PASS only for complete scope and required validation, PARTIAL for usable incomplete work, BLOCKED for a specific external dependency, and FAIL for an unresolved known failure without an external stopping dependency.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE literally.
- After the action list, use left-aligned headings and CommonMark list or key/value blocks. Do not use a Markdown pipe table unless a higher-priority schema requires it.
- Select one primary profile from the requested deliverable: implementation, research, diagnosis, planning, validation, review, operations, coordination, or fallback. Start each detail item with a short bold lead label and put long rationale on indented continuation lines.
- Preserve every field required by active delivery, validation, repository-memory, orchestration, program, and environment contracts inside the canonical profile blocks.

`
