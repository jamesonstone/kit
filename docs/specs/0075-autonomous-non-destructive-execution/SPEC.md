---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0075"
  slug: autonomous-non-destructive-execution
  dir: 0075-autonomous-non-destructive-execution
relationships:
  - type: builds_on
    target: 0061-authorized-coding-agent-merge-autonomy
    note: Supersedes the extra merge-consent prompt. Keeps MERGE_READY, identity, reviews, CI, and protection as hard gates.
  - type: builds_on
    target: 0073-infrastructure-approval-scope
    note: Supersedes additive-infrastructure confirmation. Keeps exact confirmation only for delete, remove, or otherwise destructive effects.
  - type: related_to
    target: 0065-deletion-safety
    note: Destructive fencing still routes through deletion-safety for persistent-state hard delete.
  - type: related_to
    target: 0070-deadline-mode
    note: Deadline mode must not reintroduce non-destructive consent pauses.
references:
  - id: safety-rule
    name: Safety guardrails
    type: rule
    target: docs/references/rules/safety-guardrails.md
    relation: implements
    read_policy: must
    used_for: canonical authority matrix
    status: active
  - id: merge-rule
    name: GitHub PR merge
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: implements
    read_policy: must
    used_for: merge readiness without a second consent prompt
    status: active
  - id: infrastructure-rule
    name: Infrastructure change approval
    type: rule
    target: docs/references/rules/infrastructure-change-approval.md
    relation: implements
    read_policy: must
    used_for: destructive-effect fencing
    status: active
  - id: issue
    name: Authorize non-destructive merge, deploy, and additive infrastructure from accepted task scope
    type: external
    target: https://github.com/jamesonstone/kit/issues/190
    relation: supports
    read_policy: must
    used_for: original ask and acceptance criteria
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Stop Kit coding agents from pausing accepted task or `/goal` work to ask for
merge, deploy, or additive-infrastructure consent. Keep explicit manual
authorization only for deletes, removals, or otherwise destructive effects.

## CONTEXT

- Feature 0061 made merge a distinct consent boundary. Agents therefore stopped
  even after the user accepted the task or an active `/goal`, including for
  ordinary in-scope PR merges and in-scope head repairs.
- Feature 0073 carved routine image/ECS operations out of infrastructure
  approval, but still required confirmation for additive create, IAM, network,
  and other non-destructive infrastructure. Goal-oriented deploy and LabCore
  default-off additive batches still paused.
- `deadline-mode` preserved both gates, so urgency could not override them.
- There is no first-class `/goal` authority contract. Generic "continue until
  complete" language is subordinate to those gates.
- Downstream `local-custom` copies, including LabCore's production-rollout
  exception, can silently restore the old pauses. Kit must detect that
  conflict; mutating LabCore is out of scope for this issue.
- Topology: single-lane, because canonical rules, generated templates,
  workflows, and string-locked tests overlap heavily and need continuous
  design judgment.

## REQUIREMENTS

- An accepted task or active `/goal` is standing authority, within accepted
  product scope, for issue/branch/worktree/implementation/commit/push/ready PR,
  review/CI/CVE/Inspector/build/compatibility repair, ordinary and remediation
  merges, changed-head revalidation and subsequent merge, releases, production
  deployment/activation/rollback rehearsal/acceptance, and additive or
  rollback-preserving infrastructure and IaC mutations.
- Explicit user holds such as "do not merge" or "keep production default-off"
  prevail.
- Convert merge authorization into merge readiness. Keep `MERGE_READY`,
  reviews, CI, branch protection, actor identity, dependencies, and effect
  classification as hard gates. A changed head loses readiness, not
  accepted-task authority.
- Convert infrastructure approval into destructive-effect fencing. Proceed
  autonomously for additive or rollback-preserving effects. Isolate and obtain
  exact manual confirmation for delete, remove, destroy, purge, destructive
  replacement, state removal, history rewrite, data erasure, permission
  revocation, or loss of a supported recovery path.
- A merge that triggers deletion stops because of the destructive effect, not
  because merging itself needs consent.
- Unknown effects receive further inspection; unresolved destructive ambiguity
  fails closed.
- Failed CI, Inspector, or branch protection is repaired or blocked, never
  bypassed.
- Update generated instructions, Constitution, tooling, workflows, release
  prompts, and locked tests atomically.
- Detect superseded non-destructive consent language in instruction files and
  in `github-pr-merge` / `infrastructure-change-approval` rulesets so
  `local-custom` copies cannot silently restore the old pauses.
- Non-goals: mutating downstream repositories; weakening protection, identity,
  or deletion-safety; auto-merging out-of-scope PRs.

## ACCEPTED PLAN

1. Record this accepted plan in `0075` before implementation.
2. Rewrite the authority matrix, `github-pr-merge`, `infrastructure-change-approval`,
   work-lane, delivery, deadline, team, program, Constitution, and generated
   gates so accepted-task scope authorizes non-destructive merge/deploy/infra.
3. Keep `MERGE_READY` and destructive confirmation as the only remaining
   human-stop gates besides explicit holds.
4. Lock the new behavior in ruleset, template, reconcile, and scenario tests.
5. Add reconcile detection for superseded consent phrases in instruction files
   and the two policy rulesets.
6. Note superseded 0061/0073 decisions without mechanically rewriting those
   completed specs.
7. Validate, curate repository memory, and open one ready pull request for
   issue #190.

## DECISIONS

- Topology: single-lane, because tightly coupled, high-overlap, and requiring
  continuous design judgment.
- Accepted: accepted task and active `/goal` are the same standing authority
  class for in-scope non-destructive work.
- Accepted: merge readiness is evidence, not a second consent ceremony.
- Accepted: additive IAM, network, resource create-or-update, production
  image deploy, and activation are autonomous.
- Rejected: keeping a separate merge-consent prompt "just in case."
- Rejected: treating every Terraform apply or IAM create as a confirmation
  batch.
- Out of scope: editing LabCore or other downstream repositories in this PR.

## DISCOVERIES

- Downstream `local-custom` copies of `github-pr-merge` and
  `infrastructure-change-approval` can restore the old pauses even after Kit
  templates change. Reconcile now warns when those two rulesets miss the new
  required phrases or still contain superseded consent language.
- Several locked tests match raw file text, not whitespace-normalized text.
  Required consent-removal phrases must appear on one physical line in those
  documents.
- Mutating LabCore or other downstream repositories remains out of scope for
  issue #190; Kit-side detection is the downstream hook.

## VALIDATION

- PASS: `go test ./pkg/cli ./internal/templates ./internal/releaseprompt -count=1`
- PASS: `go test ./... -count=1`
- PASS: `go fmt ./...`, `git diff --check`, `go vet ./...`
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` with zero issues
- PASS: `make build`; `./bin/kit check 0075-autonomous-non-destructive-execution`;
  `./bin/kit check --project`
- NOT_APPLICABLE: browser, live cloud, Kubernetes, IaC apply, and production
  suites; this change is local policy, templates, and tests
- Affected handwritten source/test files remain at or below 300 physical lines

## OUTCOME

- An accepted task or active `/goal` is standing authority for in-scope
  non-destructive merge, deploy, and additive infrastructure. Agents do not
  stop for a second merge or additive-infra consent prompt.
- `MERGE_READY`, identity, reviews, CI, branch protection, and explicit user
  holds remain hard gates. A changed head loses readiness, not accepted-task
  authority.
- Infrastructure confirmation is now destructive-effect fencing. Additive and
  rollback-preserving work, including IAM/network create-or-update and
  production activation, proceeds autonomously. Delete, remove, destroy, purge,
  destructive replacement, state removal, history rewrite, data erasure,
  permission revocation, and loss of recovery still require exact confirmation.
- Generated instruction files, Constitution, tooling, merge and release
  workflows, and the release prompt all use the same authority model.
- Historical specs `0061` and `0073` remain historical and record that 0075
  superseded their extra consent decisions.

## REPOSITORY MEMORY

Decision: created

Rationale: The shift from extra merge/infra consent pauses to accepted-task
authority plus destructive-only confirmation is consequential cross-project
policy that tests cannot fully preserve.

Artifacts:

- `docs/specs/0075-autonomous-non-destructive-execution/SPEC.md`
- `docs/references/rules/github-pr-merge.md`
- `docs/references/rules/infrastructure-change-approval.md`
- `docs/references/rules/safety-guardrails.md`
- `docs/CONSTITUTION.md`
- `docs/PROJECT_PROGRESS_SUMMARY.md`

Constitution: promoted accepted-task merge readiness and destructive-only
infrastructure confirmation into Evidence Before Mutation because downstream
Kit-managed projects inherit it as a project-wide agent workflow invariant.
