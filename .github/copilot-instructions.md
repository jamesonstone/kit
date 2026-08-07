# GitHub Copilot Repository Instructions

## Native Planning

Use native planning for research and design. Before implementation, inspect code and repository documentation, then decide whether material rationale requires a living `SPEC.md`. Capture the accepted plan before code when it does. After validation, load `docs/references/rules/constitution-curation.md` and curate durable decisions into the correct repository document; code-and-test-sufficient work may report that no documentation update was required.

Start with `docs/agents/README.md`. Before implementing API or backend routes, handlers, services, repositories, persistence adapters, or gateways, load `docs/references/rules/backend-service-architecture.md`. Before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components, load `docs/references/rules/frontend-application-architecture.md`. Treat both rules as responsibility boundaries rather than mandatory directory names, and preserve stronger repo-local architecture.

Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md` and the project's `docs/references/testing.md`. Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them.

Before editing implementation/source or test files, load `docs/references/rules/source-file-size.md`. Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less, audit the complete affected scope before delivery, and audit the entire repository during whole-project reconcile and scheduled maintenance.

Before Git, GitHub, or AWS mutations, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*`. Repo-local Kit rules outrank generic defaults.

## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load `docs/references/rules/infrastructure-change-approval.md`.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Put one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence into the task plan when planning is used; otherwise present it once before the first covered mutation. Obtain one explicit user confirmation for the complete bounded batch.
- Approval of a task plan containing the complete outline counts as confirmation. A sufficiently detailed initial request may also count only when it clearly authorizes the exact bounded batch and the batch does not delete or remove infrastructure.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch.
- After confirmation, execute the exact approved batch and continue the rest of the task to completion in one pass without routine command-by-command approval.
- If additional covered infrastructure changes become necessary, collect all then-known changes into one follow-up outline, obtain one confirmation, and execute that follow-up batch in one pass. Do not re-confirm actions already included in an approved batch.
- Treat a material change to target identity, environment, region or cluster, resource set, action type, impact, or recovery as a follow-up batch; compatible tools, commands, and retries inside the approved boundary do not require another prompt.

## Final Response

Every implementation final response must include:

- Repository Memory
- Decision: created | updated | refactored | deleted | not required
- Rationale: why this persistence decision is correct
- Artifacts: paths or none
