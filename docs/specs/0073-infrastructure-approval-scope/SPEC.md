---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0073"
  slug: infrastructure-approval-scope
  dir: 0073-infrastructure-approval-scope
relationships:
  - type: builds_on
    target: 0057-infrastructure-change-approval
    note: Narrows covered mutations so routine application operations are not infrastructure-approval batches, while preserving deletion confirmation.
  - type: builds_on
    target: 0061-authorized-coding-agent-merge-autonomy
    note: Stops treating every deployment-triggering merge as a covered infrastructure mutation.
  - type: related_to
    target: 0065-deletion-safety
    note: Deleting infrastructure still requires explicit post-outline confirmation and remains isolated from merge and release waves.
references:
  - id: infrastructure-rule
    name: Infrastructure change approval
    type: rule
    target: docs/references/rules/infrastructure-change-approval.md
    relation: implements
    read_policy: must
    used_for: covered-mutation and routine-application-operation boundary
    status: active
  - id: merge-rule
    name: GitHub PR merge
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: implements
    read_policy: must
    used_for: merge-plan recording of deploy effects without extra infra confirmation for image-only CD
    status: active
  - id: safety-rule
    name: Safety guardrails
    type: rule
    target: docs/references/rules/safety-guardrails.md
    relation: implements
    read_policy: must
    used_for: permission boundary for covered mutations versus routine operations
    status: active
  - id: issue
    name: Carve routine deploys and ECS ops out of infrastructure approval
    type: external
    target: https://github.com/jamesonstone/kit/issues/184
    relation: supports
    read_policy: must
    used_for: original ask and acceptance criteria
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Stop Kit-managed coding agents from treating every deployment image update
and ECS interaction as an infrastructure-approval batch, while keeping
create/replace/apply of infrastructure gated and requiring explicit manual
confirmation after an outline whenever infrastructure is deleted.

## CONTEXT

- Feature 0057 required one outline and one confirmation before any public-cloud,
  Kubernetes, or infrastructure-as-code mutation, including updates. Agents
  therefore stopped for ECS `update-service`, image rolls, desired-count
  changes, and merges whose only cloud effect was existing CD.
- Feature 0061 made a merge known to trigger deployment, Kubernetes,
  public-cloud, or IaC mutation part of that covered mutation boundary. That
  coupling is correct for Terraform applies and resource create/delete, and
  too rigid for shipping a new image onto already-provisioned compute.
- Issue #184 asks that repository rules not require approval for every
  deployment image and ECS interaction, and that deleting infrastructure still
  require explicit manual approval.
- Topology: single-lane, because the change is tightly coupled policy across
  rules, instruction templates, and string-locked tests and needs continuous
  design judgment on the carve-out boundary.

## REQUIREMENTS

- Define routine application operations on already-provisioned workloads as
  outside infrastructure-change-approval confirmation: image/digest or artifact
  rolls, force-new-deployment or rolling restart, and ECS or equivalent
  operational calls that do not create, replace, or delete infrastructure.
- Keep create, replace, import, move, and apply of infrastructure, IaC
  source/state, cluster/control-plane changes, IAM, network topology, and
  persistent data stores as covered mutations with one outline and one
  confirmation per batch.
- Treat a merge whose only known cloud effect is existing CD rolling a new
  application image or artifact as a recorded deployment effect, not a covered
  infrastructure batch. Merge authorization remains a separate gate.
- Require infrastructure approval before merge only when the merge is known to
  trigger a covered infrastructure mutation. Unknown create, replace, or delete
  effects still block until inspected.
- Always require a consolidated outline and explicit post-outline manual
  confirmation for deleting, destroying, or removing infrastructure. Merge
  consent, image deploy, ECS ops, and a broad "deploy it" never authorize
  deletion. Keep destructive work out of ordinary merge and release waves.
- Fail closed when it is uncertain whether an action creates, replaces, or
  deletes infrastructure.
- Align generated hard gates, checked-in instructions, references index,
  safety-guardrails, and locked tests. No new Kit command.
- Observable acceptance: image/ECS routine ops proceed without infra
  confirmation; covered create/replace/apply still confirm once; deletion
  always confirms after the outline; image-only merges do not invent an infra
  batch.

## ACCEPTED PLAN

1. Record this accepted plan in `0073` before implementation.
2. Narrow `infrastructure-change-approval` Applies When, add a Routine
   Application Operations section, keep deletion post-outline confirmation,
   and replace examples that treated image/ECS updates as covered batches.
3. Narrow `github-pr-merge` so only covered infrastructure mutations join the
   covered mutation boundary; record routine deploys observationally.
4. Align the shared hard gate, checked-in AGENTS/CLAUDE/Copilot/GUARDRAILS
   mirrors, safety-guardrails, references index, and pull-request-merge
   workflow language.
5. Update focused ruleset, merge-scenario, template, and consistency tests so
   the carve-out and deletion confirmation are both locked.
6. Note the superseded 0057/0061 "every update/deploy is covered" decisions
   without mechanically rewriting those completed specs.
7. Validate, curate repository memory, and open one ready pull request for
   issue #184.

## DECISIONS

- Accepted: "updating public-cloud resources" is too broad. Covered mutations
  are create, replace, import, move, apply, and delete of infrastructure,
  plus IAM, network, datastore, cluster-control-plane, and IaC changes.
- Accepted: routine application operations include existing-service image
  rolls, force-new-deployment, desired-count changes, and read/operational
  ECS calls that do not create or delete infrastructure.
- Accepted: already-provisioned application artifact hosts are in the same
  class as existing compute when the operation only ships a new artifact.
- Accepted: merge authorization does not imply infrastructure approval, and
  infrastructure approval is not required solely because merge will deploy a
  new image through existing CD.
- Accepted: deletion of infrastructure always needs explicit confirmation
  after the outline; plan approval counts only when that outline was visible.
- Rejected: treating every ECS API call or deployment-triggering merge as a
  covered infrastructure batch.
- Topology: single-lane, because tightly coupled, high-overlap, and requiring
  continuous design judgment.

## DISCOVERIES

- Locked tests currently require "known to trigger deployment, Kubernetes,
  public-cloud" on `github-pr-merge` and treat that merge as part of the
  covered mutation boundary. Those assertions must change with the rule.
- `kit spec` rewrote unrelated `PROJECT_PROGRESS_SUMMARY.md` created dates;
  that churn was reverted and only the `0073` row and summary are added.
- The shared hard-gate and reconcile snippets must keep existing deletion and
  one-pass strings while adding the routine-ops carve-out.

## VALIDATION

- PASS: `go test ./internal/templates ./pkg/cli -count=1`
- PASS: `go fmt ./...`, `git diff --check`, `go vet ./...`
- PASS: `go test ./... -count=1`
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` with zero issues
- PASS: `make build`; `./bin/kit check 0073-infrastructure-approval-scope`;
  `./bin/kit check --project`
- NOT_APPLICABLE: browser, live cloud, Kubernetes, IaC apply, and production
  suites; this change is local policy, templates, and tests
- Affected handwritten source/test files remain at or below 300 physical lines

## OUTCOME

- Routine application operations on already-provisioned workloads, including
  deployment image updates and ECS interactions that do not create or delete
  infrastructure, are no longer infrastructure-approval batches.
- Create, replace, apply, IAM, network, datastore, cluster-control-plane, and
  IaC mutations remain covered with one outline and one confirmation.
- Deleting, destroying, or removing infrastructure still requires explicit
  confirmation after the consolidated outline. Merge, image deploy, ECS ops,
  and "deploy it" never authorize deletion.
- Image-only CD merges record deployment effects and do not invent a covered
  infrastructure batch. Merge success remains distinct from deployment proof.

## REPOSITORY MEMORY

Decision: created

Rationale: The carve-out between routine application operations and covered
infrastructure mutations, plus the unchanged deletion-confirmation exception,
is consequential cross-project policy that tests cannot fully preserve.

Artifacts:

- `docs/specs/0073-infrastructure-approval-scope/SPEC.md`
- `docs/references/rules/infrastructure-change-approval.md`
- `docs/references/rules/github-pr-merge.md`
- `docs/references/rules/safety-guardrails.md`
- `docs/CONSTITUTION.md`
- `docs/PROJECT_PROGRESS_SUMMARY.md`

Constitution: promoted the covered-versus-routine-ops boundary and deletion
confirmation into Evidence Before Mutation because downstream Kit-managed
projects inherit it as a project-wide agent workflow invariant.
