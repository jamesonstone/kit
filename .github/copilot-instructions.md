# GitHub Copilot Repository Instructions

## Native Planning

Use native planning for research and design. Before implementation, inspect code and repository documentation, then decide whether material rationale requires a living `SPEC.md`. Capture the accepted plan before code when it does. After validation, load `docs/references/rules/constitution-curation.md` and curate durable decisions into the correct repository document; code-and-test-sufficient work may report that no documentation update was required.

Start with `docs/agents/README.md`. Before implementing API or backend routes, handlers, services, repositories, persistence adapters, or gateways, load `docs/references/rules/backend-service-architecture.md`. Before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components, load `docs/references/rules/frontend-application-architecture.md`. Treat both rules as responsibility boundaries rather than mandatory directory names, and preserve stronger repo-local architecture.

Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md` and the project's `docs/references/testing.md`. Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them.

Before editing implementation/source or test files, load `docs/references/rules/source-file-size.md`. Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less, audit the complete affected scope before delivery, and audit the entire repository during whole-project reconcile and scheduled maintenance.

## Multi-Agent Orchestration Evaluation Hard Gate

- Before finalizing any native implementation plan for a new feature, a substantial architectural or behavioral change, or a multi-file refactor, load `docs/references/rules/agent-team-orchestration.md` and evaluate whether the work benefits from multi-agent or parallel decomposition using that rule's lifecycle and semantic capability profiles.
- A single mechanical edit, a direct question, or read-only research that never forms an implementation plan does not trigger this gate.
- Record the decision before the plan is finalized: either a multi-lane Agent Team Plan, or `single-lane, because <reason>` using that rule's single-lane criteria. Never skip the evaluation silently, even when the recorded answer is single-lane.
- This gate fires during plan formation and precedes the Work Lane Mutation Hard Gate below, which fires later, before the first repository mutation.

## Work Lane Mutation Hard Gate

- Before any coding-agent repository file or delivery mutation, including issue, branch, staging, commit, push, worktree, and pull-request mutations, load `docs/agents/GUARDRAILS.md` and `work-lane-gating` first, complete read-only safety recon, then ask exactly: "Before I make any repository changes, should I create a new GitHub issue, GH-<issue-number> branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?"
- Interpret a leading standalone response token case-insensitively: `c` means continue existing; `n` or `y` means new lane. In a longer response, shorthand is the primary lane choice and the remaining text is supplemental lane instructions. Treat the case-insensitive full-form answers `new lane`, `new work lane`, `new worklane`, and `new worktree` as the new-lane choice: create or reuse the human-assigned GitHub issue, exact `GH-<issue-number>` branch, canonical non-primary worktree, and ready pull-request plan; ambiguous or contradictory responses fail closed.
- Wait for the explicit choice and record a Pull-Request Landing Plan covering the repository, issue, branch, canonical non-primary worktree, protected base, and create-or-update PR target. Verify that plan still matches before every mutation. Never infer the choice from clean state or a generic PR request.
- Treat the primary/root checkout as read-only. If an ungated or root change exists, preserve it: Do not stage, commit, push, stash, reset, clean, discard, or silently transfer it.


Before Git, GitHub, or AWS mutations, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*`. Repo-local Kit rules outrank generic defaults.

## GitHub Merge Authorization Hard Gate

- Merge is a distinct mutation boundary. PR-delivery consent, automatic lane allocation, approval, check success, subagent assignment, and a program ledger never imply merge consent.
- Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set.
- Before any merge or merge-queue mutation, resolve `pull-request-merge` and load `docs/references/rules/github-pr-merge.md`.
- Reconcile the authorization source, authenticated actor, expected head/base, repository merge policy, current reviews/checks, dependencies, and infrastructure or deployment effects before every wave.
- Only exact current `MERGE_READY` nodes may merge. Pending, missing, stale-head, or policy-ineligible skipped checks are not passing.
- Use one complete preflight snapshot per consequential mutation or wave; do not rerun unchanged checks or poll repeatedly unless material state changes or the evidence freshness window expires.
- Revalidating an unchanged authorized head does not require another prompt. A changed head invalidates readiness and prior merge authority; merging it requires fresh current-head evidence and explicit exact-head authorization. Adding a target or materially changing actor, method, environment, infrastructure effect, or recovery requires follow-up authorization.
- Never bypass protection, reviews, required checks, a merge queue, repository policy, or identity safeguards.
- Report merge, hosted workflow, deployment/runtime, and production evidence as separate claims.

## Cross-Repository Program Coordination Gate

- Before implementing or resuming an accepted plan that spans multiple repositories and includes dependent deliverables, staged deployment or activation, or expected agent or session handoff, load `docs/references/rules/cross-repository-program-coordination.md`.
- Designate one coordinator repository and create or adopt one canonical `docs/programs/<program>/PROGRAM.md` ledger before implementation; participant repositories remain authoritative for local specs, delivery state, runbooks, and evidence.
- Dispatch only the reconciled ready frontier, checkpoint every material transition and handoff, and reconcile recorded claims against live repositories, GitHub, runtime, and validation evidence before resume or completion.

## Deletion Safety Hard Gate

- Before designing deletion behavior or deleting persistent project, user, business, or external-system state, load `docs/references/rules/deletion-safety.md`.
- An unqualified delete means soft delete: use a reversible lifecycle state with a supported, authorized, and tested restore path. Task-owned ephemeral scratch that never became authoritative state is outside this retained-state definition; ambiguity remains covered.
- Treat purge, destroy, force deletion, empty-trash operations, destructive replacement, history rewrite, retention expiry, backup or snapshot deletion, cryptographic erasure, and irreversible cascades as hard delete.
- Make the normal product and operational path soft-delete by default. Keep hard delete as a separate privileged, auditable, server-enforced action; a client prompt or `force` flag alone is insufficient.
- Before any hard delete, resolve and present the exact targets, or a bounded selector first resolved to the exact current target set with its current count and materialized target IDs or an immutable snapshot/version token, environment, cascades, why soft delete is insufficient, the loss of restore, backup state, retention or legal impact, and verification plan.
- After that outline, obtain a specific manual confirmation from the human for those exact current targets. Initial requests, general task or plan approval, automation, retention schedules, prior soft-delete approval, and broad cleanup language do not count.
- Bind confirmation to the actor, action, exact targets or immutable snapshot/version, environment, and consequences. Immediately before execution, compare the current target set or version with the confirmed snapshot; any difference requires a new outline and confirmation.
- Preserve stricter repository, legal, privacy, security, infrastructure, and provider controls. One post-outline confirmation may satisfy multiple deletion gates only when the combined outline contains every required field.

## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load `docs/references/rules/infrastructure-change-approval.md`.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Put one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence into the task plan when planning is used; otherwise present it once before the first covered mutation. Obtain one explicit user confirmation for the complete bounded batch.
- Approval of a task plan containing the complete outline counts as confirmation. A sufficiently detailed initial request may also count only when it clearly authorizes the exact bounded batch and the batch does not delete or remove infrastructure.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch.
- During merge or release orchestration, do not execute infrastructure deletion, destruction, purge, destructive replacement, or state removal; isolate it as a separate task with its own exact post-outline authorization.
- After confirmation, execute the exact approved batch and continue the rest of the task to completion in one pass without routine command-by-command approval.
- If additional covered infrastructure changes become necessary, collect all then-known changes into one follow-up outline, obtain one confirmation, and execute that follow-up batch in one pass. Do not re-confirm actions already included in an approved batch.
- Treat a material change to target identity, environment, region or cluster, resource set, action type, impact, or recovery as a follow-up batch; compatible tools, commands, and retries inside the approved boundary do not require another prompt.

## AWS Context Hard Gate

- Before AWS-dependent work, load `docs/references/rules/aws-agent-toolkit-guidance.md` and use its current AWS skill, official documentation, AWS MCP Server or CLI fallback, identity, infrastructure-approval, and secret-safety routing; repo-local Kit gates remain authoritative. If `.kit.yaml` defines an enabled AWS context, run `kit aws verify` before the first AWS-dependent command and again immediately before AWS mutation
- Treat the verified account, ARN, and Region as authoritative; a profile name alone is not proof of identity
- Use the verified configured profile and Region explicitly for every AWS-dependent command where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Stop on missing credentials, incomplete configuration, or identity mismatch

## Agent Completion Output Contract

- Before a substantial terminal completion or handoff response, load `docs/references/rules/agent-completion-output.md` when present.
- This structured contract does not apply to intermediate progress commentary.
- Answer ordinary conversational requests naturally. Direct questions, definitions, confirmations, rewrites, brief explanations, small read-only lookups, concise recommendations, and focused clarification questions must not receive status tokens, canonical section headings, synthetic None items, task profiles, or repository-memory reporting.
- Use the structured contract when omitting it could hide a blocker, incomplete required scope, required operator action, unresolved failure, repository or external-system mutation, delivery artifact, multiple validation layers, material evidence, owner/dependency handoff, or when the user explicitly requests the canonical report.
- Do not classify by word count, token count, elapsed time, or tool-call count. When uncertain, prefer natural prose unless structure is necessary to preserve operationally important information.
- When the structured contract applies, emit exactly `## What happened`, `## Deviations`, and `## Next steps` in that order, with no other output section.
- Put `**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**` in the first What happened bullet. Do not add a separate status heading.
- Fold task-specific results, validation, delivery, coordination, and repository-memory evidence into concise What happened bullets. Use at most one nested evidence layer and state each fact once.
- Put blockers, incomplete scope, failures, warnings, pending or unknown evidence, skipped validation, and degraded execution under Deviations. Use one `**None.**` bullet when there are no deviations.
- Put independently actionable items under Next steps, required before optional. Name the actor and make every required continuation copy-ready. Use one `**None.**` bullet when no action remains.
- Use PASS only for complete scope and required validation, PARTIAL for usable incomplete work, BLOCKED for a specific external dependency, and FAIL for an unresolved known failure without an external stopping dependency.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE literally.
- Do not use Markdown pipe tables, additional profile headings, or separate Completed, Validation, Delivery, Feature State, Residual Notes, Coordination, or Repository Memory sections.
- Preserve every field required by active delivery, validation, repository-memory, orchestration, program, and environment contracts inside the three canonical sections without duplication.
- For merge or release orchestration, report only state changes, terminal evidence, and actionable next steps; omit chronological command logs, repeated checks, unchanged polling, and routine tool details.
