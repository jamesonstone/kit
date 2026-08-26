package templates

const agentCompletionOutputGate = `## Agent Completion Output Contract

- Before a substantial terminal completion or handoff response, load ` + "`docs/references/rules/agent-completion-output.md`" + ` when present.
- This structured contract does not apply to intermediate progress commentary.
- Answer ordinary conversational requests naturally. Direct questions, definitions, confirmations, rewrites, brief explanations, small read-only lookups, concise recommendations, and focused clarification questions must not receive status tokens, canonical section headings, synthetic None items, task profiles, or repository-memory reporting.
- Use the structured contract when omitting it could hide a blocker, incomplete required scope, required operator action, unresolved failure, repository or external-system mutation, delivery artifact, multiple validation layers, material evidence, owner/dependency handoff, or when the user explicitly requests the canonical report.
- Do not classify by word count, token count, elapsed time, or tool-call count. When uncertain, prefer natural prose unless structure is necessary to preserve operationally important information.
- When the structured contract applies, emit exactly ` + "`## What happened`" + `, ` + "`## Deviations`" + `, and ` + "`## Next steps`" + ` in that order, with no other output section.
- Put ` + "`**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**`" + ` in the first What happened bullet. Do not add a separate status heading.
- Fold task-specific results, validation, delivery, coordination, and repository-memory evidence into concise What happened bullets. Use at most one nested evidence layer and state each fact once.
- Put blockers, incomplete scope, failures, warnings, pending or unknown evidence, skipped validation, and degraded execution under Deviations. Use one ` + "`**None.**`" + ` bullet when there are no deviations.
- Put independently actionable items under Next steps, required before optional. Name the actor and make every required continuation copy-ready. Use one ` + "`**None.**`" + ` bullet when no action remains.
- Use PASS only for complete scope and required validation, PARTIAL for usable incomplete work, BLOCKED for a specific external dependency, and FAIL for an unresolved known failure without an external stopping dependency.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE literally.
- Do not use Markdown pipe tables, additional profile headings, or separate Completed, Validation, Delivery, Feature State, Residual Notes, Coordination, or Repository Memory sections.
- Preserve every field required by active delivery, validation, repository-memory, orchestration, program, and environment contracts inside the three canonical sections without duplication.
- For merge or release orchestration, report only state changes, terminal evidence, and actionable next steps; omit chronological command logs, repeated checks, unchanged polling, and routine tool details.

`
