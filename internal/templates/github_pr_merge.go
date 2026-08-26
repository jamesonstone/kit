package templates

const githubPRMergeGate = `## GitHub Merge Authorization Hard Gate

- Merge is a distinct mutation boundary. PR-delivery consent, automatic lane allocation, approval, check success, subagent assignment, and a program ledger never imply merge consent.
- Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set.
- Before any merge or merge-queue mutation, resolve ` + "`pull-request-merge`" + ` and load ` + "`docs/references/rules/github-pr-merge.md`" + `.
- Reconcile the authorization source, authenticated actor, expected head/base, repository merge policy, current reviews/checks, dependencies, and infrastructure or deployment effects before every wave.
- Only exact current ` + "`MERGE_READY`" + ` nodes may merge. Pending, missing, stale-head, or policy-ineligible skipped checks are not passing.
- Use one complete preflight snapshot per consequential mutation or wave; do not rerun unchanged checks or poll repeatedly unless material state changes or the evidence freshness window expires.
- Revalidating an unchanged authorized head does not require another prompt. A changed head invalidates readiness and prior merge authority; merging it requires fresh current-head evidence and explicit exact-head authorization. Adding a target or materially changing actor, method, environment, infrastructure effect, or recovery requires follow-up authorization.
- Never bypass protection, reviews, required checks, a merge queue, repository policy, or identity safeguards.
- Report merge, hosted workflow, deployment/runtime, and production evidence as separate claims.

`
