# AGENTS

## Codex Thread Initialization Hard Gate

- For every newly created Codex task, this is a blocking pre-response gate. Complete it before the first commentary message and before planning, repository inspection, shell or network commands, or any other task action; only the minimum capability lookup needed to locate the two thread operations may precede it.
- First, call the available thread-title operation (`set_thread_title` when available) with `[<project>] <description>`.
- Second, call the available thread-pin operation (`set_thread_pinned` when available).
- Both actions are required and ordered. Never defer either supported operation to a later interaction.
- Derive `<project>` from the host-provided repository or working-directory context and `<description>` from the user request without inspecting the repository first. Keep the description lowercase and at most four words.
- Verify each operation from its returned state when the host exposes one.
- If an operation is unsupported, unavailable, or fails, do not silently skip it or retry indefinitely. After resolving both actions in order, begin the first commentary with `Thread initialization: rename <status>; pin <status>.`, include a concise reason for every non-success status, then continue the requested work.
- For a continued Codex task, preserve its current title and pin state unless either is missing or the user explicitly requests a change.

## Browser policy

- For interactive browser work, use Codex's built-in browser through `@Browser`.
- Do not use `@Chrome`, control my active Chrome profile, or launch external
  Chrome or Chromium through Playwright, Selenium, Cypress, or browser MCP tools
  unless I explicitly request it.
- If `@Browser` is unavailable, report the limitation instead of silently
  falling back.
- When I explicitly authorize an external browser, terminate and verify all
  task-owned browser and automation processes before finishing.

## Purpose

- This file is a routing table, not the full manual
- Run `kit contract resolve --json` before implementation and when scope changes
- Start at `docs/agents/README.md` and load only the guidance needed for the current decision
- Use native agent planning for research, clarification, design, and implementation planning
- Treat repo-local markdown under `docs/` as persistent repository memory

## Repository Memory Gate

- Before implementation, inspect relevant code and existing repository memory
- Decide semantically whether the work contains material rationale that code and tests cannot preserve
- When material rationale exists, create or adopt `docs/specs/<feature>/SPEC.md` before editing implementation files and capture the accepted native plan
- When code and tests are sufficient, do not create documentation solely to satisfy a process; record `not required` in the final Repository Memory report
- During implementation, keep material decisions and discoveries current in the spec
- After implementation and validation, load `docs/references/rules/constitution-curation.md`; curate feature rationale into `SPEC.md`, demonstrated project invariants into `docs/CONSTITUTION.md`, reusable practices into `docs/references/` or `docs/references/rules/`, and domain knowledge into its existing canonical documentation
- Remove transient planning chatter and code-recoverable detail during curation; retain material superseded decisions with rationale

## Final Response Contract

- Every implementation final response must include:
  - `Repository Memory`
  - `Decision: created | updated | refactored | deleted | not required`
  - `Rationale: <why this is the correct persistence decision>`
  - `Artifacts: <paths or none>`

## Runtime Routing

- `docs/agents/README.md` — resolve and interpret the local agent contract
- `docs/agents/WORKFLOWS.md` — implementation delivery and registry maintenance
- `docs/agents/GUARDRAILS.md` — completion, safety, and hard rules
- `docs/agents/RLM.md` — just-in-time context loading
- `docs/agents/TOOLING.md` — Kit and native-agent ownership boundaries

## Testing And Validation Gate

- Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md` and the project's `docs/references/testing.md`
- Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them

## Source File Size Gate

- Before editing implementation/source or test files, load `docs/references/rules/source-file-size.md`
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less
- Audit the complete affected source/test scope before delivery; requested whole-project and scheduled maintenance audits cover the entire repository

## Application Architecture Gate

- Before implementing API or backend routes, controllers or handlers, services, repositories, persistence adapters, or gateways, load `docs/references/rules/backend-service-architecture.md`
- Before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components, load `docs/references/rules/frontend-application-architecture.md`
- Treat both rules as responsibility boundaries rather than mandatory directory names, and preserve stronger repo-local architecture

## GitHub Delivery Hard Gate

- Issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before a delivery mutation, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*` delivery rules
- Repo-local Kit rules outrank generic GitHub or plugin defaults

## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load `docs/references/rules/infrastructure-change-approval.md`.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Put one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence into the task plan when planning is used; otherwise present it once before the first covered mutation. Obtain one explicit user confirmation for the complete bounded batch.
- Approval of a task plan containing the complete outline counts as confirmation. A sufficiently detailed initial request may also count only when it clearly authorizes the exact bounded batch and the batch does not delete or remove infrastructure.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch.
- After confirmation, execute the exact approved batch and continue the rest of the task to completion in one pass without routine command-by-command approval.
- If additional covered infrastructure changes become necessary, collect all then-known changes into one follow-up outline, obtain one confirmation, and execute that follow-up batch in one pass. Do not re-confirm actions already included in an approved batch.
- Treat a material change to target identity, environment, region or cluster, resource set, action type, impact, or recovery as a follow-up batch; compatible tools, commands, and retries inside the approved boundary do not require another prompt.

## Knowledge Map

- `docs/specs/<feature>/SPEC.md` — material feature rationale and living implementation history
- `docs/CONSTITUTION.md` — project invariants
- `docs/references/` — reusable repo-wide knowledge and practices
- domain documentation — canonical domain behavior and interfaces
- `docs/notes/<feature>/` — optional source material, never canonical truth by itself

## Constraints

- Keep AGENTS short and stable
- Put durable workflow guidance in `docs/agents/*` instead of expanding always-loaded files
- Do not ingest or depend on agent transcripts as repository memory

<!-- BEGIN KIT AGENT CONTRACT -->
## Kit Agent Contract

- Run `kit contract resolve --json` before implementation and whenever task scope materially changes.
- Treat repository-local rulesets and workflows returned by the resolver as the authoritative contract.
- The resolver is local-only and read-only; use `kit registry status` for remote freshness and `kit reconcile` to preview drift.
- Installed contract: 14 ruleset(s), 2 workflow(s).
<!-- END KIT AGENT CONTRACT -->
