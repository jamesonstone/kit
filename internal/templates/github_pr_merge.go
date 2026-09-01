package templates

const githubPRMergeGate = `## GitHub Merge Readiness Hard Gate

- Merge is a distinct mutation boundary. PR-delivery consent, automatic lane allocation, approval, check success, subagent assignment, and a program ledger never invent merge readiness.
- An accepted task or active ` + "`/goal`" + ` authorizes in-scope ordinary and remediation merges. Do not stop for a separate merge-consent prompt.
- Before any merge or merge-queue mutation, resolve ` + "`pull-request-merge`" + ` and load ` + "`docs/references/rules/github-pr-merge.md`" + `.
- Reconcile the accepted-scope source, authenticated actor, expected head/base, repository merge policy, current reviews/checks, dependencies, and destructive versus non-destructive effects before every wave.
- Only exact current ` + "`MERGE_READY`" + ` nodes may merge. Pending, missing, stale-head, or policy-ineligible skipped checks are not passing.
- Use one complete preflight snapshot per consequential mutation or wave; do not rerun unchanged checks or poll repeatedly unless material state changes or the evidence freshness window expires.
- Revalidating an unchanged ` + "`MERGE_READY`" + ` head does not require another prompt. A changed head invalidates readiness, not accepted-task authority; revalidate current-head evidence, then merge in-scope work without requesting fresh consent. Adding a target or materially expanding product scope requires clarification. An explicit user hold such as "do not merge" prevails.
- Never bypass protection, reviews, required checks, a merge queue, repository policy, or identity safeguards.
- Report merge, hosted workflow, deployment/runtime, and production evidence as separate claims.
- Stop for exact manual confirmation only when the merge is known to trigger a destructive effect.

`
