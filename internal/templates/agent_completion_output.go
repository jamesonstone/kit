package templates

const agentCompletionOutputGate = `## Agent Completion Output Contract

- Before a substantial terminal completion or handoff response, load ` + "`docs/references/rules/agent-completion-output.md`" + ` when present.
- There is no required response format. Write every response, including terminal completions and handoffs, in the shape the content calls for, and match length to consequence rather than to effort spent.
- Do not emit the retired envelope: no ` + "`## What happened`" + `, ` + "`## Deviations`" + `, or ` + "`## Next steps`" + ` headings, no ` + "`**Status: ...**`" + ` token, no ` + "`**None.**`" + ` item, and no announcement that there are no deviations, risks, or next steps.
- Do not replace it with a different fixed template applied to every task.
- Whatever shape a response takes, never leave the reader wrong about a blocker, incomplete scope, a required next action, a failing or unobserved check, a repository or external-system mutation, or how confident you are.
- State plainly whether work is finished, partly finished, blocked, or failed. No token, label, or taxonomy is required.
- Report a failing, pending, unavailable, skipped, or never-run check as exactly that, never folded into success, and preserve literal states such as PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE.
- Never report a check as passing unless it ran and passed, or a file as inspected unless it was inspected. When something could not be validated, say so and say why.
- Give a required action with enough context to act on it, and the exact command or prompt when there is one.
- Include an identifier only when the reader needs it to act or would doubt the claim without it. Omit a commit SHA cited as proof that something was checked, test and suite name lists, file-and-line references for correct code, and counts of things inspected.
- Delivery, validation, repository-memory, orchestration, program, and environment contracts still bind. They name facts that must reach the reader, never a layout, and a heading alone never satisfies one.

`
