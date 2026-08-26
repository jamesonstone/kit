---
kind: ruleset
slug: cross-repository-program-coordination
description: Preserves dependency, delivery, deployment, evidence, checkpoint, and handoff state for accepted plans spanning multiple repositories.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - workflow
  - cross-repository
  - program
  - deployment
  - handoff
  - resume
  - dispatch
read_policy_default: must
---

# Ruleset: cross-repository-program-coordination

## Purpose

- Preserve operational knowledge while an accepted plan moves through multiple
  repositories, delivery lanes, deployments, agents, sessions, and milestones.
- Make the next safe work frontier, current blockers, accountable owner, and
  global completion evidence recoverable without relying on chat history or one
  agent's context window.
- Coordinate repository-local sources of truth without copying or replacing
  their requirements, implementation rationale, runbooks, or evidence.

## Applies When

Load and follow this rule before implementing or resuming an accepted plan when
both conditions are true:

1. the plan spans more than one repository; and
2. it includes at least one of these program signals:
   - deliverables depend on changes or interfaces in another repository;
   - deployment, activation, migration, or validation must occur in stages; or
   - execution is expected to cross agent or session handoffs.

Do not create a program ledger merely for read-only multi-repository research,
one atomic change mirrored across repositories, or several independent changes
whose order and completion do not affect one another. Escalate into this rule
when those tasks acquire a shared dependency, rollout, acceptance, or handoff
boundary.

This rule coordinates a program. It does not replace each repository's local
workflow, specification, issue, pull request, testing, infrastructure approval,
deployment, or production-safety rules.

## Rules

### One Coordinator And One Ledger

- Designate one coordinator repository before program implementation begins.
  Choose the repository that owns the integrated outcome or, when no product
  repository naturally owns it, the repository that durably owns program
  operations.
- Create or adopt exactly one canonical program ledger at
  `docs/programs/<program>/PROGRAM.md` in that coordinator repository.
- When qualifying work is already in progress, reconstruct the ledger from
  live repository, GitHub, runtime, and validation state and checkpoint that
  reconciled baseline before further mutation or dispatch. Do not backfill it
  from agent memory or chat alone.
- Treat the program ledger as hand-authored canonical Markdown until Kit
  explicitly supports a compatible program artifact. Generated state may index
  the ledger but must not replace it.
- Record the coordinator repository, program owner, scope, non-goals, global
  acceptance gates, and ledger location in the ledger itself.
- Participant repositories may point to the ledger. They must not create
  competing ledgers for the same program.

### Local Truth And Program Pointers

- Keep detailed requirements, architecture rationale, accepted implementation
  plans, decisions, discoveries, and local acceptance in each participant
  repository's canonical `SPEC.md` or other repo-local artifact.
- Keep issues, branches, commits, pull requests, runbooks, deployment records,
  and validation evidence in the system that owns them.
- The program ledger stores stable identities, current coordination state, and
  exact pointers to those sources. It must not copy full local specs, task
  lists, runbooks, logs, or test evidence.
- Identify repositories by durable owner/name or canonical remote. Never use an
  absolute workstation path as a repository identity.
- When local and program records disagree, reconcile against the authoritative
  local or live source and update the program ledger. Do not make the copied or
  older claim authoritative.

### Stable Program Model

Assign stable, human-readable IDs that remain unchanged across reordering:

- one program ID;
- milestone IDs for externally meaningful integrated outcomes;
- workstream IDs for repository-owned deliverables; and
- gate IDs for cross-workstream acceptance, deployment, or activation checks.

The ledger must contain:

- a participant repository table with owner, local spec, issue, branch, pull
  request, and operational-reference pointers when applicable;
- the authorization source and exact approved PR set, authenticated GitHub
  actor for each repository, expected PR head/base, merge method and repository
  policy, bounded in-place-remediation authority, replacement-PR criteria, and
  corrective or rollback owner;
- a dependency graph that names each workstream's prerequisites and consumers;
- the current ready frontier: only unblocked workstreams whose prerequisites
  and approvals are currently satisfied;
- pre-merge evidence, post-merge gates, and known infrastructure or deployment
  effects for every merge node;
- milestone and gate state with the evidence needed to advance them;
- interface, compatibility, migration, rollout, and rollback dependencies when
  they affect ordering or safety; and
- blockers with accountable owner, affected workstreams, impact, recorded time,
  and an exact unblock condition.

Use explicit `none`, `not applicable`, `unknown`, or `unobserved` values instead
of silently omitting a field that affects readiness or completion.

### Separate State Dimensions

Never overload one `done` field. Track these dimensions independently for each
workstream when applicable:

1. implementation: planned, in progress, blocked, or complete at an exact
   commit;
2. GitHub delivery: issue, branch, pull request, review/check, and merge state;
3. deployment/runtime: target environment, artifact or image identity,
   deployment state, activation state, and rollback readiness; and
4. validation: suites or gates run, result, observed version, evidence pointer,
   and observation time.

A merged pull request is not proof of deployment. A successful deployment
command is not proof of runtime readiness. A passing local suite is not proof
of production behavior. Report each claim only in the dimension its evidence
supports.

### Evidence Contract

For every claim that advances readiness or a global gate, record the applicable
exact evidence:

- repository and commit SHA;
- issue, branch, pull request, review, and hosted-check identity;
- target environment and deployment, release, artifact, or image identity;
- observed state and literal result;
- observation time in UTC;
- immutable or canonical evidence pointer; and
- actor or system that produced the observation when attribution matters.

Keep raw run evidence and the curated `tests/RUN_STATUS.md` entry in the
participant repository according to
`testing-and-environment-validation`. Store only their exact pointers and
program-level interpretation in the ledger. Redact secrets and protected data.

### Supervisor And Participant Authority

- Keep one accountable program supervisor responsible for program scope,
  dependency ordering, ready-frontier decisions, integration, checkpoint
  quality, merge-wave decisions, global validation, and completion reporting.
- Assign each workstream an accountable repository owner or agent. Participant
  agents report state and evidence for their assigned workstream; they may edit
  the coordinator ledger only when the supervisor explicitly assigns the exact
  coordinator checkout and ledger update.
- Participant agents must not expand program scope, change another
  workstream's ownership, bypass a dependency or gate, or mark the whole
  program complete unless the program supervisor explicitly assigns that
  authority.
- A participant may merge only specifically assigned PR nodes from the exact
  approved `MERGE_READY` frontier. Participant or subagent assignment alone
  does not create merge authority, and read-only verifiers never merge.
- The accepted bounded plan or direct user request creates merge authority.
  The ledger records and reconciles that authority; it never creates it.
- Before every merge wave, the supervisor reconciles the authorization source,
  approved PR set, actor, expected head/base, repository policy, current
  evidence, approvals, and ready frontier against live sources.
- Use `agent-team-orchestration` for the execution topology inside each ready
  wave. Its overlap, concurrency, verification, and delivery boundaries remain
  in force.
- Dispatch only the current ready frontier. Do not dispatch blocked downstream
  work merely to maximize concurrency.

### Checkpoints

Update the program ledger immediately after any material state transition:

- milestone or workstream start or completion;
- pull request, merge, deployment, activation, rollback, or validation change;
- new, changed, or cleared blocker;
- approval, exception, or accepted residual risk;
- interface, compatibility, sequencing, scope, or ownership decision;
- user redirect or material correction;
- planned context compaction; or
- agent or session handoff.

Each checkpoint must record the UTC time, supervisor, current milestone, state
changes, current ready frontier, blockers, next safe actions, evidence pointers,
and any claims that still require live observation. Prefer one curated current
checkpoint plus material historical decisions over an append-only activity log.

### Resume, Handoff, And Reconciliation

Before resume, handoff, dispatch, milestone advancement, or completion:

1. read the program ledger and the local artifacts for the affected frontier;
2. reconcile recorded repository, GitHub, deployment/runtime, and validation
   claims against the strongest currently available live sources;
3. record drift and update the ready frontier;
4. change stale, unavailable, or unverified claims to `unknown` or `unobserved`
   instead of carrying them forward as fact; and
5. checkpoint the reconciled state before assigning or performing more work.

Before each merge wave, also resolve `pull-request-merge` and follow
`github-pr-merge`. Revalidation of unchanged authorized heads does not require
another prompt. A changed head loses prior readiness and merge authority and
requires fresh exact-head authorization; adding a target or materially changing
actor, method, environment, infrastructure effect, or recovery also requires
follow-up authorization.

A handoff must identify the coordinator and ledger, active milestone, current
ready frontier, blockers and owners, material decisions, exact evidence,
unobserved claims, next safe actions, and prohibited or deferred scope. A chat
summary may point to this checkpoint but is not the handoff authority.

### Completion

The program supervisor may mark the program complete only when:

- every required workstream and global gate is satisfied;
- dependency, interface, compatibility, rollout, and rollback obligations are
  resolved;
- implementation, delivery, deployment/runtime, and validation claims have
  current exact evidence or an explicitly accepted exception;
- the final integrated outcome is observed in every required environment;
- blockers are closed or residual risk is explicitly accepted by the user;
- participant repository memory is current; and
- a final reconciled checkpoint records the actual outcome and remaining
  operational obligations.

Render the terminal program result through the `agent-completion-output`
three-section contract. Group repository, owner, dependency, deployment,
runtime, and acceptance evidence by workstream under What happened. Put
unresolved dependencies under Deviations and exact handoffs under Next steps.

### Safety And Existing Gates

- Continue to obey every participant repository's instructions and ownership
  boundaries. The program ledger does not grant cross-repository mutation
  authority.
- The accepted plan or direct user request creates authority; the ledger only
  records it. Never infer merge permission from ledger existence, a ready
  frontier, participant assignment, or check success.
- Infrastructure changes still require the applicable consolidated approval
  boundary, target verification, rollback plan, and provider-specific gates.
- Never store secrets, credentials, customer data, raw logs, agent transcripts,
  or copied private context in the ledger.
- Do not infer deployment, activation, validation, or production success from a
  plan, commit, merge, command exit, or outdated checkpoint.

## Anti-Patterns

- Treating one long plan or task checklist as durable program state.
- Calling a set of unrelated repository edits a program solely to add process.
- Creating multiple ledgers or making every participant repository duplicate
  the full program state.
- Copying local specs, runbooks, test output, or deployment logs into the
  coordinator repository.
- Using chat transcripts, agent memory, generated JSON, absolute paths, or the
  current agent's context as the only program record.
- Using one `done` status for implementation, merge, deployment, activation,
  and validation.
- Dispatching downstream work before its dependencies and approvals are met.
- Marking a milestone complete from stale, indirect, or compatibility-only
  evidence.
- Treating `services stable`, a successful CLI exit, or a merged pull request as
  exact proof of deployed identity and runtime behavior.
- Allowing a participant agent to change global scope or completion state
  without supervisor authority.
- Treating the program ledger or participant assignment as merge authority.
- Maintaining an append-only diary instead of a concise current checkpoint and
  material decision history.

## Verification

- Confirm the trigger applies and one coordinator repository is explicit.
- Confirm exactly one canonical `PROGRAM.md` ledger exists for the program.
- Confirm every participant has durable identity and authoritative local
  pointers without duplicated requirements or evidence.
- Confirm stable IDs, dependency graph, current ready frontier, global gates,
  blockers, and next safe actions are present and internally consistent.
- Confirm implementation, GitHub delivery, deployment/runtime, and validation
  are represented as separate state dimensions.
- Confirm readiness and completion claims cite exact, current evidence and
  literal observation times.
- Confirm dispatch contains only the reconciled ready frontier and still obeys
  agent-team concurrency and overlap limits.
- Confirm every merge wave contains only the exact authorized `MERGE_READY`
  frontier, uses the expected actor/head/base/method, and resolves
  `pull-request-merge` before mutation.
- Confirm routine scoped remediation preserved its existing pull request and
  changed heads received fresh checks, review, revalidation, and exact-head
  authorization before reentering the frontier.
- Confirm a checkpoint follows each material transition and every handoff.
- Confirm resume, handoff, dispatch, and completion reconcile against live
  repository, GitHub, runtime, and validation sources.
- Confirm unknown or unavailable live state is reported as `unknown` or
  `unobserved`, not assumed from an old checkpoint.
- Confirm final completion satisfies all global gates and records remaining
  operational obligations or accepted residual risk.
- Confirm the ledger contains no secrets, protected data, transcripts, raw
  logs, copied local specs, or absolute workstation paths.

## Examples

Program trigger:

```text
The frontend depends on new contracts in two services, database migration must
precede service rollout, and production activation will happen in three waves.
Use one coordinator ledger and dispatch only workstreams whose prerequisites
and approvals are currently satisfied.
```

Not a program trigger:

```text
Inspect twelve repositories for the same deprecated setting and report which
ones contain it. This is read-only research with no dependent delivery,
deployment, shared acceptance gate, or expected handoff; do not create a
program ledger.
```

Separate state dimensions:

```text
WS-api
  implementation: complete at 9f1c2ab
  delivery: PR #84 merged at 2026-08-10T14:20:00Z
  deployment: staging release api@sha256:... observed ready; production unknown
  validation: staging contract suite PASS; production end-to-end unobserved
```

Reconciled handoff:

```text
Coordinator: owner/workflow-control
Ledger: docs/programs/dynamic-workflows/PROGRAM.md
Milestone: M2-staging-integration
Ready frontier: WS-frontend, GATE-staging-e2e
Blocked: WS-production-activation; owner platform; unblock when the exact image
digest is observed healthy in staging and GATE-staging-e2e passes.
Next: dispatch WS-frontend, run the staging gate, then checkpoint. Do not start
production activation from the prior chat summary.
```
