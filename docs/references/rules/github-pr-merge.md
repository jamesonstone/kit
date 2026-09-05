---
kind: ruleset
slug: github-pr-merge
description: Binds explicit bounded standing authority to exact current pull requests while preserving readiness, identity, policy, deployment, and risk gates.
status: active
registry_scope: downstream
applies_to:
  - git
  - github
  - pull-request
  - merge
  - merge-queue
  - cross-repository
  - coding-agent
read_policy_default: must
---

# Ruleset: GitHub PR Merge

## Purpose

- Allow a coding agent to merge pull requests covered by explicit bounded
  standing authority without asking again for each later PR number or head OID.
- Keep merge readiness, repository policy, current-head evidence, identity,
  destructive-effect classification, and post-merge proof explicit and
  independently verifiable.
- Prevent generic task acceptance, PR-delivery consent, a program ledger,
  subagent assignment, or green checks from inventing standing authority or
  `MERGE_READY`.

## Applies When

Load this rule and resolve the `pull-request-merge` workflow before any merge or
merge-queue mutation.

Standing merge authority exists only when a human explicitly authorizes a
bounded task, goal, or program to merge its resulting work. Record that grant
with its goal and non-goals, repositories, bases, environments, permitted
actions, actor, expiry or completion boundary, and exclusions.

The grant may use a semantic scope before pull-request numbers or final head
OIDs exist. Bind later pull requests only when each is directly required by the
authorized outcome, belongs to a governed in-scope lane, targets an authorized
repository and base, and introduces no materially different effect. Materialize
the exact PR and current head during pre-merge reconciliation. Do not ask again
merely because those identifiers were unknown when authority was granted.

Accepting a generic implementation task, opening or approving a pull request,
automatic lane allocation, review resolution, check success, participant
assignment, and program-ledger existence do not create standing authority or
`MERGE_READY`.

The most recent direct human instruction wins. An explicit pause, hold, or
revocation immediately supersedes standing authority for the affected actions
and dependents. Only an explicit human resume or new grant restores it; a new
head, passing check, retry, session, or unchanged ledger cannot.

## Rules

### Authorization Boundary

- Record the standing-authority source and selector, then resolve and record
  the exact current in-scope pull-request set before each mutation wave.
- Treat the resolved existing pull-request set as explicit continuation under
  `work-lane-gating`. Do not create a new coordination issue, branch, worktree,
  or pull request merely to prepare or execute the merge plan.
- Exact-head identity is a readiness constraint. An in-scope in-place repair
  keeps the same pull request in the graph but invalidates its readiness until
  the new head is revalidated. It does not require a new consent prompt.
- Revalidating an unchanged `MERGE_READY` head, retrying a compatible merge
  path, or using the repository-required merge queue does not require another
  prompt.
- A later PR or refreshed head inside the recorded selector retains authority.
  Adding or changing repository, target base, environment, actor, identity,
  merge method, deployment workflow, product scope, or material effect is
  expansion and requires explicit updated authority.
- Protection bypass, admin override, review bypass, required-check bypass,
  force-push, and silent identity substitution are prohibited even when merge
  authority exists.

### Bounded Merge Plan

Before the first merge, record:

- standing-authority source, selector, current pause/revocation state, and the
  resolved in-scope PR set;
- repository identity and authenticated GitHub actor for every repository;
- expected PR head OID, base branch, and current PR state;
- repository-approved merge method or merge-queue policy;
- dependency edges and the current ready frontier;
- required review and hosted-check policy plus current-head evidence;
- known deployment, Kubernetes, public-cloud, database, and infrastructure-as-
  code effects, including whether any standard deployment is named by the
  grant and whether any excluded risk class is present;
- post-merge deployment, runtime, and validation gates; and
- in-scope in-place-remediation policy, replacement-PR criteria,
  failure-containment, recovery, and rollback ownership.

Use explicit `none`, `not applicable`, `unknown`, or `unobserved`. Any field
that affects readiness and remains unknown blocks that node.

### Identity And Repository Boundary

- Before every repository boundary, resolve the repository owner/name and
  verify the authenticated GitHub actor.
- The actor must be the expected human user or an explicitly authorized service
  identity named by the accepted plan. Never silently switch accounts,
  profiles, tokens, or identities.
- Recheck identity immediately before mutation and verify the post-merge event
  identifies the expected actor.
- Identity failure blocks only that repository node and its dependents. It does
  not contaminate independent authorized nodes whose identity remains valid.
- Existing human Git author and committer rules remain unchanged; a GitHub
  merge actor is not permission to substitute commit identity.

### Merge Readiness

Classify every authorized node as exactly one of:

- `MERGE_READY`: standing authority currently covers the exact node, all
  required current-head evidence is present and attributable, repository
  policy accepts it, every dependency is satisfied, and no excluded or
  unresolved material-risk class is present;
- `BLOCKED`: a required gate failed, authority is paused or revoked, or the
  node requires scope expansion or an excluded risk class; or
- `UNKNOWN`: evidence is missing, stale, unavailable, ambiguous, or cannot be
  attributed to the expected head, target, policy, or actor.

Only exact current `MERGE_READY` nodes may merge or enter the ready frontier.
These are never passing:

- pending or missing expected checks;
- skipped checks without verified policy eligibility;
- checks for an earlier head OID;
- local tests substituted for required hosted checks;
- review, mergeability, base, policy, actor, deployment workflow, environment,
  or material effects that are unknown; or
- a successful merge treated as deployment or production proof.

Head or base drift invalidates readiness. Recompute evidence and the frontier
before mutation. A changed in-scope head invalidates readiness, not standing
authority. After fresh current-head checks, review, mergeability, policy,
identity, dependencies, and effect classification pass, it may re-enter the
frontier without a new authorization prompt.

### Repository Policy And Merge Method

- Inspect the repository's allowed merge methods, branch protection, rulesets,
  required reviews and checks, merge queue, and documentation-only policy.
- Use the required merge queue when policy requires it; queue admission still
  needs current `MERGE_READY` evidence.
- Use only a repository-permitted merge method. Do not choose an admin or
  bypass variant to make a blocked merge succeed.
- For documentation-only squash merges, preserve the repository's eligible
  skip directive and synthesized-message requirements only when the complete
  diff and required-check policy qualify. A skip is not a passing check.

### Infrastructure And Deployment Effects

- Record known deployment, Kubernetes, public-cloud, and infrastructure-as-code
  effects in the merge plan, including routine application operations such as
  existing CD rolling a new image onto already-provisioned compute.
- Standing deployment authority covers only a repository-approved existing
  standard workflow named by the grant for an authorized environment, limited
  to the merged artifact on already-provisioned targets plus required runtime
  verification.
- IAM, network, KMS, secrets, database schema or data-loss changes,
  infrastructure creation, replacement, or deletion, destructive deletion,
  and nonstandard deployment effects are outside standing merge/deploy
  authority. Route them to their own applicable approval boundary.
- Unknown create, replace, delete, data, security, or deployment effects
  require inspection. Unresolved classification makes the node `UNKNOWN` or
  `BLOCKED`, never `MERGE_READY`.
- Merge success never implies workflow success, deployed identity, runtime
  readiness, production validation, or rollback readiness.

### Supervisor, Participants, And Waves

- One accountable supervisor owns authorization reconciliation, graph state,
  ready-frontier decisions, merge waves, failure containment, and final
  reporting.
- A participant may merge only specifically assigned PR nodes that are both in
  the in-scope set and current ready frontier.
- Subagent assignment alone does not create merge authority. A participant may
  not expand the in-scope set, bypass dependencies, or advance a global gate.
- Read-only verification agents never merge, queue, resolve review state, or
  perform another delivery mutation.
- Independent `MERGE_READY` nodes may merge concurrently. Dependency chains and
  nodes coupled through a base, service, environment, database, migration,
  queue, deployment, or acceptance gate remain serialized.
- Use one complete current-state snapshot per consequential mutation or wave.
  Reconcile the standing-authority selector and pause state, head/base, actor,
  policy, checks, effects, and dependencies once immediately before every
  wave. Refresh only
  when a material fact changes or the evidence freshness window expires.

### Partial Failure And Corrective Work

- A failure on one node stops that node and its dependents. Preserve exact
  completed, queued, blocked, unknown, and unobserved state.
- Continue only independent in-scope nodes whose readiness remains valid and
  whose failure isolation is proven.
- Treat routine remediation as an update to the existing pull request when it
  stays within that pull request's issue and declared scope, its head branch can
  be updated safely with ordinary commits, and repository policy permits the
  update. Do not create recursive corrective pull requests for minor conflicts,
  generated artifacts, dependency refreshes, or other scope-preserving fixes.
- Standing authority authorizes in-scope in-place repair when its recorded
  permitted actions include blocker repair. A specific repair hold remains a
  blocker.
- Perform in-scope in-place remediation between merge waves. Ordinarily merge
  the current base into the existing head branch, apply or regenerate the
  repair, commit, and push to the same branch without rebasing, force-pushing,
  or retargeting. Then mark the node `UNKNOWN`, remove it from the ready
  frontier, rerun every required current-head readiness check, and restore
  `MERGE_READY` without a new authorization prompt when the selector still
  matches.
- Use a replacement pull request only when remediation materially changes the
  issue scope or architecture, the original head cannot be updated safely, or
  repository policy or the user explicitly requires replacement. A replacement
  is a new node. A replacement that still satisfies the standing selector is
  revalidated and merged without a new authorization prompt; expansion is not
  automatically covered.
- Never force, bypass, change identity, or broaden product scope to recover a
  wave.
- Diagnose a failure before retrying. Within in-scope repair, rerun only
  affected evidence and refresh the failed node plus its dependents; repeated
  unchanged failure becomes `BLOCKED`.

### Post-Merge Evidence

After each merge or queue transition, record:

- repository, PR, expected and observed head/base, merge method, merge or queue
  result, merge commit when available, actor, and observation time;
- the exact completed frontier and recalculated downstream frontier;
- hosted workflow state for the merged identity;
- deployment/runtime/production state as separate observed claims; and
- blockers, unknowns, corrective ownership, and next safe action.

Prefer event-driven waits or bounded backoff for hosted and deployment state.
Do not emit unchanged polling as progress or reproduce a chronological command
log in the terminal result.

Report partial waves literally. Do not call a queued, merged, deployed, or
healthy state by another name.

Under an active `docs/references/rules/deadline-mode.md` authorization,
continue authorized merge and deployment work without interleaving UI or
browser walkthrough verification after each result. After every result in
the authorized set is delivered, run one final UI verification. Record merge,
hosted-workflow, and deployment/runtime evidence after each transition as
usual. Do not treat the deferred UI check as license to skip required
post-deployment production-suite evidence.

## Anti-Patterns

- Requiring renewed authorization solely because an in-scope PR number or final
  head OID was unknown when standing authority was granted.
- Treating generic task acceptance as standing merge authority.
- Treating PR-delivery consent, automatic lane allocation, review resolution,
  check success, or a program ledger as `MERGE_READY`.
- Merging an extra PR because it appears related or ready.
- Merging from stale head evidence or substituting local checks for required
  hosted checks.
- Assigning merge authority implicitly to every subagent or verifier.
- Using admin merge, protection bypass, identity substitution, or an
  unsupported merge method.
- Starting a merge with IAM, network, KMS, secrets, database schema or data-
  loss changes, infrastructure create/replace/delete, destructive deletion, or
  nonstandard deployment effects under standing merge/deploy authority.
- Inventing an infrastructure-approval batch solely because merge will deploy
  a new application image through existing CD.
- Creating replacement or recursively corrective pull requests for routine,
  scope-preserving remediation that can safely stay on the existing PR head.
- Updating a PR head and merging it under readiness or evidence bound to
  the prior head.
- Rechecking unchanged evidence or polling without a state transition merely
  to recreate an already-current snapshot.
- Treating merge success as deployment, runtime, production, or integration
  evidence.
- Interleaving UI or browser walkthroughs after each merge or deployment
  result during an active deadline-mode wave instead of one final UI
  verification after all results are delivered.

## Verification

- Confirm an explicit standing-authority grant covers the resolved exact
  current PR set and that pause, revocation, expiry, and completion state were
  reconciled.
- Confirm `pull-request-merge` context resolved and every required artifact was
  loaded.
- Confirm the authenticated actor, repository, expected head/base, merge
  method, policy, dependencies, checks, reviews, and destructive-effect
  classification were revalidated immediately before mutation.
- Confirm only `MERGE_READY` in-scope nodes entered each wave.
- Confirm independent concurrency did not cross dependency or same-base
  source, deployment, or acceptance serialization boundaries.
- Confirm each wave used one current preflight snapshot and refreshed evidence
  only after material change or expiry.
- Confirm identity or node failure remained isolated and exact partial state
  was preserved.
- Confirm routine in-scope remediation preserved the existing pull request,
  invalidated prior head evidence, and obtained fresh current-head checks and
  review without a new authorization prompt.
- Confirm replacement PRs were limited to material scope or architecture
  change, an unsafe or inaccessible original head, or explicit policy or user
  direction.
- Confirm post-merge evidence separates merge, workflow, deployment/runtime,
  and production validation.
- Confirm that under active deadline mode, UI or browser walkthroughs waited
  until every authorized merge and deployment result was delivered, then one
  final UI verification ran.
- Confirm no bypass, admin override, silent identity substitution, or
  out-of-scope expansion occurred.

## Examples

Later in-scope blocker PR:

```text
The human granted standing authority for the bounded checkout-recovery goal in
owner/service, targeting main and the standard staging deploy workflow. Blocker
PR #84 was created later in a governed lane. Its current head/base, actor,
policy, reviews, required checks, dependencies, and effects are current and
acceptable. State: MERGE_READY. Merge #84, run only the named standard deploy,
verify runtime, and resume the goal without another authorization prompt.
```

Unauthorized extra PR:

```text
Standing authority covers service PRs required by the checkout-recovery goal.
#91 changes the production network and is not covered. State: BLOCKED pending
explicit expanded authority and the infrastructure approval contract.
```

Pause and revocation:

```text
The human paused production deployment while keeping staging authorized. Stop
production and its dependents immediately. Passing checks and a new head do not
resume it; continue only independent staging work until the human explicitly
resumes or grants replacement authority.
```

Partial wave:

```text
Wave 2 merged service-a#84. service-b#87 became UNKNOWN after head drift, so it
and its dependent UI#90 stopped. Independent docs#12 remains MERGE_READY.
```

Deadline-mode merge/deploy wave:

```text
Deadline mode is active. Wave 1 merges service#84 and ui#90; both configured
deployments complete. UI verification stays SKIPPED until both results are
delivered, then one final UI verification runs against the delivered system.
```
