package templates

const agentCompletionOutputGate = `## Agent Completion Output Contract

- Before a substantial terminal completion or handoff response, load ` + "`docs/references/rules/agent-completion-output.md`" + ` when present.
- This structured contract does not apply to intermediate progress commentary.
- Answer ordinary conversational requests naturally. Direct questions, definitions, confirmations, rewrites, brief explanations, small read-only lookups, concise recommendations, and focused clarification questions must not receive status tokens, canonical section headings, synthetic None items, task profiles, or repository-memory reporting.
- Use the structured contract when omitting it could hide a blocker, incomplete required scope, required operator action, unresolved failure, repository or external-system mutation, delivery artifact, multiple validation layers, material evidence, owner/dependency handoff, or when the user explicitly requests the canonical report.
- Do not classify by word count, token count, elapsed time, or tool-call count. When uncertain, prefer natural prose unless structure is necessary to preserve operationally important information.
- When the structured contract applies, use only ` + "`## What happened`" + `, ` + "`## Deviations`" + `, and ` + "`## Next steps`" + `, in that order. Always emit What happened; omit Deviations and Next steps entirely when empty instead of writing a None bullet.
- Open What happened with ` + "`**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**`" + ` on its own line, then one to three plain sentences of prose. Do not add a separate status heading.
- Write a briefing, not a transcript: target twelve rendered lines or fewer, at most five What happened bullets, one sentence plus at most one evidence clause per bullet, and at most one nested line.
- Use bullets only for genuinely separate outcomes, and reserve bold for the status line, blockers, and required actions.
- Include an identifier only when the reader needs it to act or would doubt the claim without it. Omit a commit SHA cited as proof that something was checked rather than as a delivery identifier a composing contract requires, test and suite name lists, file-and-line references for correct code, and counts of things inspected; name the validation that ran and its result in a few words.
- Put blockers, incomplete scope, failures, warnings, pending or unknown evidence, skipped validation, and degraded execution under Deviations, one sentence each. A PARTIAL, BLOCKED, or FAIL status always carries Deviations.
- Put independently actionable items under Next steps, required before optional. Name the actor, make every required continuation copy-ready, and never manufacture or speculatively offer an action.
- Use PASS only for complete scope and required validation, PARTIAL for usable incomplete work, BLOCKED for a specific external dependency, and FAIL for an unresolved known failure without an external stopping dependency.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE literally.
- Do not use Markdown pipe tables, additional profile headings, or separate Completed, Validation, Delivery, Feature State, Residual Notes, Coordination, or Repository Memory sections.
- Preserve every field required by active delivery, validation, repository-memory, orchestration, program, and environment contracts inside the canonical sections without duplication, using the shortest phrasing that keeps each fact recoverable.
- For merge or release orchestration, report only state changes, terminal evidence, and actionable next steps; omit chronological command logs, repeated checks, unchanged polling, and routine tool details.

`
