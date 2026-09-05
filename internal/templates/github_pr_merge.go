package templates

const githubPRMergeGate = `## GitHub Standing Merge Authority Hard Gate

- Merge is a distinct mutation boundary. PR-delivery consent, automatic lane allocation, approval, check success, subagent assignment, and a program ledger never invent merge readiness.
- Standing merge authority exists only when a human explicitly authorizes a bounded task, goal, or program to merge its resulting work. Generic task acceptance does not create it. Record repositories, bases, environments, permitted actions, actor, expiry or completion, and exclusions.
- Standing authority may bind later-created in-scope PRs and refreshed heads. Resolve the exact current PR and head before mutation; do not ask again solely because its number or final OID was unknown when authority was granted.
- Before any merge or merge-queue mutation, resolve ` + "`pull-request-merge`" + ` and load ` + "`docs/references/rules/github-pr-merge.md`" + `.
- Reconcile the standing-authority selector and pause state, authenticated actor, expected head/base, repository policy, current reviews/checks, dependencies, deployment workflow, environment, and material effects before every wave.
- Only exact current ` + "`MERGE_READY`" + ` nodes may merge. Pending, missing, stale-head, or policy-ineligible skipped checks are not passing.
- Use one complete preflight snapshot per consequential mutation or wave; do not rerun unchanged checks or poll repeatedly unless material state changes or the evidence freshness window expires.
- A changed in-scope head invalidates readiness, not standing authority. Revalidate current-head evidence, then merge without renewed authorization. Scope, repository, base, environment, actor, identity, method, workflow, or material-effect expansion requires explicit updated authority.
- Never bypass protection, reviews, required checks, a merge queue, repository policy, or identity safeguards.
- Report merge, hosted workflow, deployment/runtime, and production evidence as separate claims.
- IAM, network, KMS, secrets, database-schema or data-loss changes, infrastructure creation/replacement/deletion, destructive deletion, nonstandard deployment effects, and unresolved risk classifications are not covered by standing merge/deploy authority.
- The most recent direct human instruction wins. Pause, hold, or revocation stops affected actions and dependents until explicit human resume or replacement authority.

`
