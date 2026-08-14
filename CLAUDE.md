# CLAUDE

## Purpose

- This file is a routing table, not the full manual
- Start at `docs/agents/README.md` and load only the guidance needed for the current decision
- Use native agent planning for research, clarification, design, and implementation planning
- Treat repo-local markdown under `docs/` as persistent repository memory

## Work Lane Mutation Hard Gate

- Before any coding-agent repository file or delivery mutation, including issue, branch, staging, commit, push, worktree, and pull-request mutations, load `docs/agents/GUARDRAILS.md` and `work-lane-gating` first, complete read-only safety recon, then ask exactly: "Before I make any repository changes, should I create a new GitHub issue, GH-<issue-number> branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?"
- Wait for the explicit choice and record a Pull-Request Landing Plan covering the repository, issue, branch, canonical non-primary worktree, protected base, and create-or-update PR target. Verify that plan still matches before every mutation. Never infer the choice from clean state or a generic PR request.
- Treat the primary/root checkout as read-only. If an ungated or root change exists, preserve it: Do not stage, commit, push, stash, reset, clean, discard, or silently transfer it.


## Coding Agent Context Gate

- When Kit command behavior is not already established, run `kit capabilities <command> --json` before choosing the command
- Before implementation, maintenance, PR repair, or repository bootstrap, run `kit context resolve --workflow <slug> --json` with relevant feature and path hints
- Load every required selected artifact before acting; load optional evidence only when its applicability boundary is reached
- Treat a blocked contract as a hard evidence gap and rerun resolution after material scope changes
- `kit context resolve` is local-only and read-only; it never fetches, writes, mutates Git, infers truth, or launches an agent

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

- `docs/agents/README.md` — classify the work and choose the next document
- `docs/agents/WORKFLOWS.md` — native planning, implementation, and repository-memory lifecycle
- `docs/agents/GUARDRAILS.md` — completion, safety, and hard rules
- `docs/agents/RLM.md` — just-in-time context loading
- `docs/agents/TOOLING.md` — skills, post-plan dispatch, and secondary inputs

## Testing And Validation Gate

- Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md` and the project's `docs/references/testing.md`
- Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them

## Source File Size Gate

- Before editing implementation/source or test files, load `docs/references/rules/source-file-size.md`
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less
- Audit the complete affected source/test scope before delivery; whole-project reconcile and scheduled maintenance audit the entire repository

## Application Architecture Gate

- Before implementing API or backend routes, controllers or handlers, services, repositories, persistence adapters, or gateways, load `docs/references/rules/backend-service-architecture.md`
- Before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components, load `docs/references/rules/frontend-application-architecture.md`
- Treat both rules as responsibility boundaries rather than mandatory directory names, and preserve stronger repo-local architecture

## GitHub Delivery Hard Gate

- Issue, branch, staging, commit, push, PR, and merge actions are distinct mutation boundaries
- Before a delivery mutation, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*` delivery rules
- Repo-local Kit rules outrank generic GitHub or plugin defaults

## GitHub Merge Authorization Hard Gate

- Merge is a distinct mutation boundary. PR-delivery consent, automatic lane allocation, approval, check success, subagent assignment, and a program ledger never imply merge consent.
- Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set.
- Before any merge or merge-queue mutation, resolve `pull-request-merge` and load `docs/references/rules/github-pr-merge.md`.
- Reconcile the authorization source, authenticated actor, expected head/base, repository merge policy, current reviews/checks, dependencies, and infrastructure or deployment effects before every wave.
- Only exact current `MERGE_READY` nodes may merge. Pending, missing, stale-head, or policy-ineligible skipped checks are not passing.
- Revalidating an authorized target does not require another prompt. Adding a target or materially changing actor, method, environment, infrastructure effect, or recovery requires follow-up authorization.
- Never bypass protection, reviews, required checks, a merge queue, repository policy, or identity safeguards.
- Report merge, hosted workflow, deployment/runtime, and production evidence as separate claims.

## Cross-Repository Program Coordination Gate

- Before implementing or resuming an accepted plan that spans multiple repositories and includes dependent deliverables, staged deployment or activation, or expected agent or session handoff, load `docs/references/rules/cross-repository-program-coordination.md`.
- Designate one coordinator repository and create or adopt one canonical `docs/programs/<program>/PROGRAM.md` ledger before implementation; participant repositories remain authoritative for local specs, delivery state, runbooks, and evidence.
- Dispatch only the reconciled ready frontier, checkpoint every material transition and handoff, and reconcile recorded claims against live repositories, GitHub, runtime, and validation evidence before resume or completion.

## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load `docs/references/rules/infrastructure-change-approval.md`.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Put one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence into the task plan when planning is used; otherwise present it once before the first covered mutation. Obtain one explicit user confirmation for the complete bounded batch.
- Approval of a task plan containing the complete outline counts as confirmation. A sufficiently detailed initial request may also count only when it clearly authorizes the exact bounded batch and the batch does not delete or remove infrastructure.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch.
- After confirmation, execute the exact approved batch and continue the rest of the task to completion in one pass without routine command-by-command approval.
- If additional covered infrastructure changes become necessary, collect all then-known changes into one follow-up outline, obtain one confirmation, and execute that follow-up batch in one pass. Do not re-confirm actions already included in an approved batch.
- Treat a material change to target identity, environment, region or cluster, resource set, action type, impact, or recovery as a follow-up batch; compatible tools, commands, and retries inside the approved boundary do not require another prompt.

## AWS Context Hard Gate

- Before AWS-dependent work, load `docs/references/rules/aws-agent-toolkit-guidance.md` and use its current AWS skill, official documentation, AWS MCP Server or CLI fallback, identity, infrastructure-approval, and secret-safety routing; repo-local Kit gates remain authoritative. If `.kit.yaml` defines an enabled AWS context, run `kit aws verify` before the first AWS-dependent command and again immediately before AWS mutation
- Use only the verified configured profile and Region; stop on missing credentials, incomplete configuration, or identity mismatch

## Knowledge Map

- `docs/specs/<feature>/SPEC.md` — material feature rationale and living implementation history
- `docs/CONSTITUTION.md` — project invariants
- `docs/references/` — reusable repo-wide knowledge and practices
- domain documentation — canonical domain behavior and interfaces

## Constraints

- Keep CLAUDE short and stable
- Put durable workflow guidance in `docs/agents/*` instead of expanding always-loaded files
- Do not ingest or depend on agent transcripts as repository memory
