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
- Revalidating an authorized target, retrying a compatible merge path, or using
  the repository-required merge queue does not require another prompt.
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
  effects, including their approval state;
- post-merge deployment, runtime, and validation gates; and
- corrective-PR, failure-containment, recovery, and rollback ownership.

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

- A merge known to trigger deployment, Kubernetes, public-cloud, or
  infrastructure-as-code mutation is part of that covered mutation boundary.
- Before authorizing that node as `MERGE_READY`, the accepted plan must identify
  the triggering workflow, target account/environment/region/cluster, expected
  actions and impact, recovery or rollback, and post-merge evidence.
- The same accepted plan may contain both merge authorization and applicable
  infrastructure approval. Do not ask twice when one complete plan satisfies
  both contracts.
- Unknown indirect effects or incomplete infrastructure approval make the node
  `UNKNOWN` or `BLOCKED`; inspect them before merge.
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
  same-base operations whose order affects conflict, queue, or release state
  remain serialized.
- Reconcile authorization, head/base, actor, policy, checks, approvals, and
  dependencies immediately before every wave.

### Partial Failure And Corrective Work

- A failure on one node stops that node and its dependents. Preserve exact
  completed, queued, blocked, unknown, and unobserved state.
- Continue only independent authorized nodes whose readiness remains valid and
  whose failure isolation is proven.
- Source remediation becomes a normal corrective pull request. It is not
  automatically added to the authorized set and cannot be merged without
  follow-up authorization or an accepted plan that already names the bounded
  corrective policy.
- Never force, bypass, change identity, or broaden scope to recover a wave.

### Post-Merge Evidence

After each merge or queue transition, record:

- repository, PR, expected and observed head/base, merge method, merge or queue
  result, merge commit when available, actor, and observation time;
- the exact completed frontier and recalculated downstream frontier;
- hosted workflow state for the merged identity;
- deployment/runtime/production state as separate observed claims; and
- blockers, unknowns, corrective ownership, and next safe action.

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
- Starting a deployment-triggering merge before its effects and approval are
  known.
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
  serialization boundaries.
- Confirm identity or node failure remained isolated and exact partial state
  was preserved.
- Confirm post-merge evidence separates merge, workflow, deployment/runtime,
  and production validation.
- Confirm no bypass, admin override, silent identity substitution, or
  unauthorized scope expansion occurred.

## Examples

Authorized single PR:

```text
Direct request: merge owner/service#84.
Head/base, actor, policy, reviews, required checks, and deployment effects are
current and acceptable. State: MERGE_READY. Merge only #84.
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
