---
kind: ruleset
slug: github-pr-merge
description: Authorizes and gates exact pull-request merges without weakening repository policy, infrastructure approval, identity, or evidence requirements.
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

- Allow a coding agent to merge an exact pull-request set when a direct user
  request or accepted bounded merge plan authorizes it.
- Keep authorization, repository policy, current-head evidence, identity,
  deployment effects, and post-merge proof explicit and independently
  verifiable.
- Prevent PR-delivery consent, a program ledger, subagent assignment, or green
  checks from inventing merge authority.

## Applies When

Load this rule and resolve the `pull-request-merge` workflow before any merge or
merge-queue mutation.

The rule applies only when one of these creates merge authority:

- the user directly requests one or more exact pull-request merges; or
- the user accepts a bounded merge plan naming the complete authorized set.

Opening, updating, or approving a ready pull request does not trigger this rule
and does not authorize merge. Automatic clean-preflight delivery allocation,
review-thread resolution, check success, participant assignment, and program
ledger existence are also not authorization.

## Rules

### Authorization Boundary

- Record the authorization source and exact approved pull-request set before
  mutation.
- Treat that exact existing pull-request set as explicit continuation under
  `work-lane-gating`. Do not create a new coordination issue, branch, worktree,
  or pull request merely to prepare or execute the merge plan.
- Exact-head merge authority freezes the commit eligible for merge; it does not
  authorize source repair. An authorized in-place repair keeps the same pull
  request in the graph but invalidates its readiness and prior exact-head merge
  authority until the new head is explicitly authorized.
- Revalidating an unchanged authorized head, retrying a compatible merge path,
  or using the repository-required merge queue does not require another prompt.
- Use one complete current-state snapshot per consequential mutation or wave.
  Do not rerun unchanged checks, reread identical evidence, or poll repeatedly
  inside its declared freshness window.
- Adding a pull request, repository, target base, deployment environment,
  infrastructure effect, merge method, admin path, or identity is material
  scope expansion and requires follow-up authorization.
- Protection bypass, admin override, review bypass, required-check bypass,
  force-push, and silent identity substitution are prohibited even when merge
  authority exists.

### Bounded Merge Plan

Before the first merge, record:

- authorization source and approved PR set;
- repository identity and authenticated GitHub actor for every repository;
- expected PR head OID, base branch, and current PR state;
- repository-approved merge method or merge-queue policy;
- dependency edges and the current authorized ready frontier;
- required review and hosted-check policy plus current-head evidence;
- known deployment, Kubernetes, public-cloud, and infrastructure-as-code
  effects, classified as routine application operations or covered
  infrastructure mutations, including approval state when covered;
- post-merge deployment, runtime, and validation gates; and
- bounded in-place-remediation authority, replacement-PR criteria,
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

- `MERGE_READY`: all required current-head evidence is present, attributable,
  acceptable under repository policy, and every dependency and approval is
  satisfied;
- `BLOCKED`: a required gate failed or an explicit dependency or approval is
  unmet; or
- `UNKNOWN`: evidence is missing, stale, unavailable, ambiguous, or cannot be
  attributed to the expected head, target, policy, or actor.

Only `MERGE_READY` nodes may enter the ready frontier. These are never passing:

- pending or missing expected checks;
- skipped checks without verified policy eligibility;
- checks for an earlier head OID;
- local tests substituted for required hosted checks;
- review, mergeability, base, policy, actor, or deployment effects that are
  unknown; or
- a successful merge treated as deployment or production proof.

Head or base drift invalidates readiness. Recompute evidence and the frontier
before mutation.

### Repository Policy And Merge Method

- Inspect the repository's allowed merge methods, branch protection, rulesets,
  required reviews and checks, merge queue, and documentation-only policy.
- Use the required merge queue when policy requires it; queue admission still
  needs current `MERGE_READY` evidence and exact authorization.
- Use only a repository-permitted merge method. Do not choose an admin or
  bypass variant to make a blocked merge succeed.
- For documentation-only squash merges, preserve the repository's eligible
  skip directive and synthesized-message requirements only when the complete
  diff and required-check policy qualify. A skip is not a passing check.

### Infrastructure And Deployment Effects

- Record known deployment, Kubernetes, public-cloud, and infrastructure-as-code
  effects in the merge plan, including routine application operations such as
  existing CD rolling a new image onto already-provisioned compute.
- A merge known to trigger a covered infrastructure mutation is part of that
  covered mutation boundary. Before authorizing that node as `MERGE_READY`,
  the accepted plan must identify the triggering workflow, target
  account/environment/region/cluster, expected actions and impact, recovery
  or rollback, and post-merge evidence.
- A merge whose only known cloud effect is a routine application operation
  does not require infrastructure-change-approval confirmation. Record the
  workflow and environment; do not invent a covered infrastructure batch.
- The same accepted plan may contain both merge authorization and applicable
  infrastructure approval. Do not ask twice when one complete plan satisfies
  both contracts.
- Unknown create, replace, or delete effects, or incomplete approval for a
  covered infrastructure mutation, make the node `UNKNOWN` or `BLOCKED`;
  inspect them before merge. An image-only CD path is not an unknown covered
  mutation once classified as a routine application operation.
- Merge success never implies workflow success, deployed identity, runtime
  readiness, production validation, or rollback readiness.

### Supervisor, Participants, And Waves

- One accountable supervisor owns authorization reconciliation, graph state,
  ready-frontier decisions, merge waves, failure containment, and final
  reporting.
- A participant may merge only specifically assigned PR nodes that are both in
  the approved set and current ready frontier.
- Subagent assignment alone does not create merge authority. A participant may
  not expand the approved set, bypass dependencies, or advance a global gate.
- Read-only verification agents never merge, queue, resolve review state, or
  perform another delivery mutation.
- Independent `MERGE_READY` nodes may merge concurrently. Dependency chains and
  nodes coupled through a base, service, environment, database, migration,
  queue, deployment, or acceptance gate remain serialized.
- Reconcile authorization, head/base, actor, policy, checks, approvals, and
  dependencies once immediately before every wave. Refresh only when a
  material fact changes or the evidence freshness window expires.

### Partial Failure And Corrective Work

- A failure on one node stops that node and its dependents. Preserve exact
  completed, queued, blocked, unknown, and unobserved state.
- Continue only independent authorized nodes whose readiness remains valid and
  whose failure isolation is proven.
- Treat routine remediation as an update to the existing pull request when it
  stays within that pull request's issue and declared scope, its head branch can
  be updated safely with ordinary commits, and repository policy permits the
  update. Do not create recursive corrective pull requests for minor conflicts,
  generated artifacts, dependency refreshes, or other scope-preserving fixes.
- Exact-head merge authority alone does not authorize source remediation. The
  user or an accepted plan must separately authorize a bounded in-place repair
  policy before the head changes.
- Perform authorized in-place remediation between merge waves. Ordinarily merge
  the current base into the existing head branch, apply or regenerate the
  repair, commit, and push to the same branch without rebasing, force-pushing,
  or retargeting. Then mark the node `UNKNOWN`, remove it from the ready
  frontier, rerun required current-head checks and review, and obtain fresh
  exact-head merge authorization before it can become `MERGE_READY` again.
- Use a replacement pull request only when remediation materially changes the
  issue scope or architecture, the original head cannot be updated safely, or
  repository policy or the user explicitly requires replacement. A replacement
  is a new node and is not automatically added to the authorized set.
- Never force, bypass, change identity, or broaden scope to recover a wave.
- Diagnose a failure before retrying. Within separately authorized repair
  scope, rerun only affected evidence and refresh the failed node plus its
  dependents; repeated unchanged failure becomes `BLOCKED`.

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

## Anti-Patterns

- Treating PR-delivery consent, automatic lane allocation, review resolution,
  check success, or a program ledger as merge authorization.
- Merging an extra PR because it appears related or ready.
- Merging from stale head evidence or substituting local checks for required
  hosted checks.
- Assigning merge authority implicitly to every subagent or verifier.
- Using admin merge, protection bypass, identity substitution, or an
  unsupported merge method.
- Starting a merge that creates, replaces, or deletes infrastructure before
  those covered effects and approval are known.
- Inventing an infrastructure-approval batch solely because merge will deploy
  a new application image through existing CD.
- Creating replacement or recursively corrective pull requests for routine,
  scope-preserving remediation that can safely stay on the existing PR head.
- Updating a PR head and merging it under authorization or evidence bound to
  the prior head.
- Rechecking unchanged evidence or polling without a state transition merely
  to recreate an already-current snapshot.
- Treating merge success as deployment, runtime, production, or integration
  evidence.

## Verification

- Confirm a direct request or accepted bounded plan created authority for the
  exact PR set.
- Confirm `pull-request-merge` context resolved and every required artifact was
  loaded.
- Confirm the authenticated actor, repository, expected head/base, merge
  method, policy, dependencies, checks, reviews, and approvals were revalidated
  immediately before mutation.
- Confirm only `MERGE_READY` authorized nodes entered each wave.
- Confirm independent concurrency did not cross dependency or same-base
  source, deployment, or acceptance serialization boundaries.
- Confirm each wave used one current preflight snapshot and refreshed evidence
  only after material change or expiry.
- Confirm identity or node failure remained isolated and exact partial state
  was preserved.
- Confirm routine authorized remediation preserved the existing pull request,
  invalidated prior head evidence and merge authority, and obtained fresh
  current-head checks, review, revalidation, and exact-head authorization.
- Confirm replacement PRs were limited to material scope or architecture
  change, an unsafe or inaccessible original head, or explicit policy or user
  direction.
- Confirm post-merge evidence separates merge, workflow, deployment/runtime,
  and production validation.
- Confirm no bypass, admin override, silent identity substitution, or
  unauthorized scope expansion occurred.

## Examples

Authorized single PR:

```text
Direct request: merge owner/service#84.
Head/base, actor, policy, reviews, required checks, and deployment effects are
current and acceptable. The hosted workflow rolls a new image onto the existing
ECS service. That is a routine application operation, not a covered
infrastructure batch. State: MERGE_READY. Merge only #84.
```

Unauthorized extra PR:

```text
The accepted plan authorizes #84 and #87. #91 is green and related but is not
in the approved set. State: BLOCKED pending follow-up authorization.
```

Partial wave:

```text
Wave 2 merged service-a#84. service-b#87 became UNKNOWN after head drift, so it
and its dependent UI#90 stopped. Independent docs#12 remains MERGE_READY.
```
