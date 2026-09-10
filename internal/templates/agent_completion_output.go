package templates

const agentCompletionOutputGate = `## Agent Completion Output Contract

- Before a substantial terminal completion or handoff response, load ` + "`docs/references/rules/agent-completion-output.md`" + ` when present.
- Write each response, including terminal completions and handoffs, in the shape its content calls for. Match length to consequence rather than to effort spent.
- A terminal response conveys what the user now has, what remains unfinished and why, anything blocking completion and what would clear it, and anything the reader must do next with the exact command or prompt when there is one.
- Say plainly whether the work is finished, partly finished, blocked, or failed.
- Keep blockers and unfinished scope as prominent as the successes.
- Report each check as observed: a failing, pending, unavailable, skipped, or never-run check is reported as exactly that, and literal states such as PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE are preserved verbatim.
- Report a check as passing only when it ran and passed, and a file or system as inspected only when it was inspected. When something could not be validated, say so and say why.
- Distinguish a verified fact from an inference and from a hypothesis.
- Make every repository, delivery, external-system, and infrastructure change recoverable from the response, with the identifiers a reader needs to find or undo it.
- Include an identifier when the reader needs it to act or would reasonably doubt the claim without it. A response is an account of where things stand, not an index of everything checked.
- For merge or release orchestration, report state changes and the smallest evidence set that proves each terminal node.
- Delivery, validation, repository-memory, orchestration, program, and environment contracts name facts that must reach the reader. Satisfy them on content; a heading alone satisfies none of them.

`
