---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0061"
  slug: "authorized-coding-agent-merge-autonomy"
  dir: "0061-authorized-coding-agent-merge-autonomy"
relationships:
  - type: builds_on
    target: 0046-autonomous-mutation-recovery
  - type: builds_on
    target: 0058-cross-repository-program-coordination
  - type: related_to
    target: 0059-conservative-coding-agent-first
  - type: related_to
    target: 0060-release-orchestration-prompt
references:
  - id: safety-rule
    name: Safety guardrails
    type: rule
    target: docs/references/rules/safety-guardrails.md
    relation: implements
    read_policy: must
    used_for: canonical authority model and prohibited actions
    status: active
  - id: merge-rule
    name: GitHub PR merge
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: implements
    read_policy: must
    used_for: authorized merge planning, execution, and verification
    status: active
  - id: lane-rule
    name: Work lane gating
    type: rule
    target: docs/references/rules/work-lane-gating.md
    relation: implements
    read_policy: must
    used_for: separation of PR delivery consent and merge authorization
    status: active
  - id: team-rule
    name: Agent team orchestration
    type: rule
    target: docs/references/rules/agent-team-orchestration.md
    relation: implements
    read_policy: must
    used_for: supervisor and participant merge authority
    status: active
  - id: program-rule
    name: Cross-repository program coordination
    type: rule
    target: docs/references/rules/cross-repository-program-coordination.md
    relation: implements
    read_policy: must
    used_for: authorized PR set, ready frontier, and wave evidence
    status: active
  - id: infrastructure-rule
    name: Infrastructure change approval
    type: rule
    target: docs/references/rules/infrastructure-change-approval.md
    relation: implements
    read_policy: must
    used_for: merges that trigger covered mutations
    status: active
  - id: testing-rule
    name: Testing and environment validation
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: implements
    read_policy: must
    used_for: merge-readiness evidence semantics
    status: active
  - id: in-place-remediation-issue
    name: Prefer in-place PR remediation in merge policy
    type: external
    target: https://github.com/jamesonstone/kit/issues/172
    relation: supports
    read_policy: must
    used_for: exact-head and source-remediation authority refinement
    status: active
  - id: concise-context-aware-orchestration-issue
    name: Strengthen merge orchestration rules
    type: external
    target: https://github.com/jamesonstone/kit/issues/180
    relation: supports
    read_policy: must
    used_for: concise preflight, concurrency, delegation, recovery, and reporting policy
    status: active
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Define one repository-native authority model that lets coding agents merge an
exact, policy-compliant pull-request set when a direct request or accepted
bounded plan authorizes it, without allowing PR-delivery consent, a program
ledger, subagent assignment, or successful checks to invent broader authority.

## CONTEXT

- Historical feature 0046 deliberately retained a categorical prohibition on
  autonomous merge while expanding routine mutation recovery. That historical
  decision remains evidence, but active guidance now needs a narrower
  capability: authorized merge autonomy with explicit scope and evidence.
- `safety-guardrails`, `work-lane-gating`, `github-pr-delivery`, agent-team,
  program-coordination, infrastructure, validation, generated instructions,
  and release-prompt surfaces currently express different pieces of consent
  and mutation policy. Updating only merge wording would leave contradictions.
- Pull-request delivery authority and merge authority have different blast
  radii. Creating an issue, branch, commit, push, or ready PR never authorizes
  integration into a protected base.
- A merge may indirectly trigger deployment, Kubernetes, public-cloud, or
  infrastructure-as-code mutation. Its authorization must cover those known
  effects before the merge, not after the workflow starts.
- Issue #141 and branch `GH-141` own this integration together with the
  `kit pr orchestrate` prompt consumer. Historical specs remain unchanged.
- Issue #172 and branch `GH-172` refine the repair boundary after downstream
  use showed that treating every source fix as a new corrective PR creates
  recursive delivery lanes for routine, scope-preserving remediation.
- Issue #180 and branch `GH-180` strengthen only the existing rule registry and
  generated repository gates. They add no Kit CLI invocation or executor.

## REQUIREMENTS

- Put one compact authority matrix in `safety-guardrails`:
  - read-only discovery: implied by the task;
  - in-scope implementation and safe recovery: the current accepted task;
  - issue, branch, commit, push, and ready PR: PR-delivery consent;
  - review-thread mutation: explicitly assigned repair/resolution authority;
  - PR merge: a direct merge request or accepted bounded merge plan;
  - multi-repository merge program: an approved plan plus reconciled program
    ledger;
  - deployment or infrastructure mutation: the applicable approval; and
  - protection bypass, admin override, or identity substitution: prohibited.
- Add an active downstream `github-pr-merge` ruleset and a
  `pull-request-merge` context workflow. A direct merge request or accepted
  bounded plan routes to that evidence before any merge mutation.
- State in `work-lane-gating` and `github-pr-delivery` that PR-delivery consent
  and automatic clean-preflight allocation never imply merge consent. Scope
  expansion needs follow-up authorization; revalidation of an already
  authorized target does not.
- Distinguish exact-head merge authority from source-remediation authority.
  Under a separately authorized bounded repair policy, routine fixes that stay
  inside the existing PR issue and declared scope must update that same head
  branch between waves, ordinarily by merging the current base into the head,
  applying or regenerating the correction, and committing and pushing without
  rebase, force-push, or retargeting. The changed node returns to `UNKNOWN` and
  needs fresh checks, review, revalidation, and exact-head authorization.
- Reserve replacement PRs for material issue-scope or architecture changes,
  original heads that cannot be updated safely, or explicit repository-policy
  or user direction. Never create recursively corrective PRs for routine
  conflicts, generated artifacts, dependency refreshes, or similar in-scope
  fixes.
- Make one accountable supervisor own merge-wave decisions. Participants may
  merge only specifically assigned authorized frontier nodes; read-only
  verification agents never merge; subagent assignment does not create merge
  authority or permit global-gate advancement.
- Permit independent ready nodes to merge concurrently while serializing
  dependency chains and same-base sensitive operations.
- Define independence across both source and deployment effects. Shared bases,
  services, environments, databases, migrations, queues, or acceptance gates
  remain serialized even when repositories differ.
- Use one complete preflight snapshot per consequential mutation or wave.
  Refresh only after material state change or freshness expiry, and monitor
  hosted or deployment state with event-driven waits or bounded backoff.
- Treat a merge known to trigger deployment, Kubernetes, public-cloud, or IaC
  mutation as part of that covered mutation boundary. The accepted plan must
  identify the workflow, target context, actions and impact, recovery, and
  post-merge evidence. One complete plan may satisfy merge and infrastructure
  approval without duplicate prompts; unknown effects block merge.
- During release orchestration, infrastructure deletion, destruction, purge,
  destructive replacement, and state removal are prohibited. Isolate them as a
  separate task with their own exact post-outline authorization.
- Allow explicitly requested lower-cost or lower-capability agents only for
  exact bounded ready-node merges and deployment monitoring when the host can
  prove sufficient capability. Keep graph, repair, recovery, wave, and
  acceptance decisions with the accountable supervisor.
- Diagnose failures before retrying, limit repair to authorized scope, rerun
  only affected evidence, and stop repeated unchanged failures.
- Keep terminal release reports status-first and concise; omit chronological
  command logs, repeated checks, unchanged polls, and routine tool detail.
- Define `MERGE_READY`, `BLOCKED`, and `UNKNOWN` validation states. Only
  `MERGE_READY` enters a merge frontier. Pending, missing, stale-head,
  policy-ineligible skipped, unattributed, or locally substituted required
  checks never pass. Merge success is not deployment or production evidence.
- Extend `PROGRAM.md` guidance with authorization source and approved PR set,
  authenticated actor, expected head/base, merge method and policy,
  dependencies/frontier, pre-merge evidence, post-merge gate,
  infrastructure/deployment effects, and corrective/rollback ownership.
- Clarify that an accepted plan or direct request creates authority; the ledger
  records and reconciles it but never creates it. Reconcile authority before
  every wave.
- Preserve identity safeguards at every repository boundary: verify the
  authenticated GitHub actor, accept only the expected human or explicitly
  authorized service identity, prohibit silent profile substitution, and
  verify the post-merge actor. One repository's identity failure blocks only
  its node and dependents.
- Align `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`,
  `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md`,
  `docs/agents/TOOLING.md`, instruction templates, workflow templates, and
  paired tests atomically. Every active delivery-mutation list must include
  merge distinctly and route it to `github-pr-merge`.
- Align the `kit pr orchestrate` renderer and golden with the same remediation
  default so release graphs preserve existing PR identity and model only
  exceptional replacement PRs as new graph nodes.
- Make `kit pr orchestrate` build the graph and bounded merge plan first,
  identify authorization, resolve `pull-request-merge`, merge only the
  reconciled ready frontier, avoid per-PR reconfirmation after plan acceptance,
  reauthorize only material scope expansion, and report partial waves
  literally. Kit generates the prompt; the coding agent performs mutations.
- Add an integrated active-policy consistency test excluding historical specs
  and notes. It must reject categorical current no-merge language,
  delivery-as-merge consent, ledger-as-authority, protection bypass,
  infrastructure-triggering merges without approval, verification-agent
  mutation authority, and generated mutation lists that omit merge.
- Add scenario coverage for authorized single-PR and multi-repository batches,
  missing authorization, dependency ordering, head drift, pending/missing
  checks, extra PR rejection, participant authority, infrastructure-triggering
  merges, partial-wave failure, actor mismatch, merge queues, and
  documentation-only squash policy.
- Keep historical specifications unchanged and state in current memory that
  this feature supersedes the former manual-only policy only for active
  guidance.
- Keep every changed handwritten source and test file at or below 300 lines.
- Non-goals: implement a GitHub merge executor in Kit, weaken reviews/checks or
  branch protection, grant cross-repository authority from a ledger, authorize
  deployments implicitly, bypass repository merge policy, or rewrite history.

## ACCEPTED PLAN

1. Define `github-pr-merge` and `pull-request-merge` as the canonical bounded
   merge contract, then align safety, delivery, lane, team, program,
   infrastructure, validation, Constitution, routing, and generated guidance.
2. Add deterministic active-policy consistency checks and focused scenario
   tests that make contradictory consent, authority, evidence, identity, and
   infrastructure semantics fail closed.
3. Complete `kit pr orchestrate` so its rendered prompt consumes the authority
   model and produces a reconciled, dependency-aware merge plan while Kit
   itself remains prompt-only.
4. Keep historical 0046, 0058, and 0059 specifications unchanged; link this
   V3 spec and release-prompt spec to record the superseding active policy.
5. Validate focused policy, templates, workflow materialization, command
   behavior, complete Go/static/build checks, source size, project context,
   and deterministic manual flows before explicit delivery on `GH-141`.
6. On `GH-172`, replace the replacement-first repair rule with authorized
   in-place remediation, invalidate old-head readiness and merge authority, and
   update the checked-in and embedded workflows and generated merge gates.
7. Align `kit pr orchestrate`, its golden, scenario tests, and the related
   release-orchestration feature memory, then rerun focused and complete
   validation before ready-PR delivery.
8. On `GH-180`, strengthen the existing merge, infrastructure, team, and
   completion rules plus generated hard gates and focused tests. Do not add or
   change a Kit command.

## DECISIONS

- Accepted: merge authority is created only by a direct request or accepted
  bounded plan and is preserved across compatible revalidation and retries.
- Accepted: PR-delivery consent, subagent assignment, check success, and a
  program ledger are evidence or workflow inputs, never merge authorization.
- Accepted: one complete accepted plan may jointly carry merge authorization
  and covered infrastructure approval when it names the full bounded batch.
- Accepted: fail closed as `UNKNOWN` when exact current-head evidence, identity,
  repository policy, or indirect deployment effects cannot be established.
- Accepted: current evidence is reused until a material state transition or
  declared freshness expiry; redundant unchanged-state rechecking is rejected.
- Accepted: release concurrency requires source and deployment independence,
  and bounded mechanical participants never own coordinator judgment.
- Accepted: destructive infrastructure work is outside release orchestration
  and must be isolated behind a separate approval boundary.
- Accepted: exact-head authority freezes only the commit eligible to merge; it
  does not require a new PR for separately authorized routine source repair.
  Any changed head loses prior readiness and merge authority.
- Accepted: preserve the original PR by default for routine scoped fixes.
  Replacement PRs are exceptional and become separately authorized graph
  nodes only when scope, architecture, head safety, policy, or user direction
  requires them.
- Rejected: categorical active no-merge language. It erases the distinction
  between unauthorized automation and user-authorized agent execution.
- Rejected: a Kit merge command. The current product boundary is a
  deterministic prompt generator and repository evidence harness.

## DISCOVERIES

- Active safety guidance currently lists merge among prohibited actions and
  must be narrowed without weakening force-push, protection, review, identity,
  secret, or repository-setting safeguards.
- Generated V1/V2/V3 files and checked-in V3 instruction artifacts repeat the
  delivery mutation boundary in several templates; consistency must be tested
  as an integrated active surface rather than only with substring unit tests.
- The existing program ledger already separates merge, deployment, and
  validation states, providing a compatible place to record bounded authority
  and frontier evidence without becoming an authorization source.
- The generated merge gate increased the V2 root template near its historical
  line threshold; the audit baseline now permits one small local policy section
  without weakening duplicate-manual pattern detection.
- The active downstream rule and `kit pr orchestrate` independently encoded a
  replacement-first corrective-PR loop. Updating only the registry rule would
  allow the release prompt to recreate the same infinite regress, so both
  surfaces and their checked-in and embedded derivatives must move together.
- Issue #180 confirmed that the necessary behavior already belongs to the
  rules registry. A new command would duplicate the existing prompt and
  orchestration surfaces instead of strengthening downstream policy.

## VALIDATION

- The new ruleset validates structurally and its scenario tests cover exact
  single-PR and multi-repository authority, missing consent, dependency order,
  head drift, pending/missing checks, extra targets, participant scope,
  infrastructure-triggered merges, partial waves, identity mismatch, merge
  queues, and documentation-only squash policy.
- Active-policy consistency tests reject categorical no-merge language,
  delivery or ledger authority invention, verifier mutation, infrastructure
  approval gaps, protection bypass, and generated mutation lists that omit
  merge.
- Checked-in and embedded instructions and workflows agree. Both
  `pull-request-merge` and `release-orchestration` resolve unblocked.
- Complete Go and race tests, vet, lint, builds, project validation, source-size
  audit, manual prompt flows, and diff checks pass as recorded in related spec
  0060.
- `GH-172` focused merge-rule, active-policy, generated-gate, checked-in and
  embedded workflow, release-workflow, and release-renderer/golden tests pass.
- `make fmt`, `go test ./... -count=1`, `go test -race ./... -count=1`,
  `go vet ./...`, `golangci-lint run ./...`, `go build ./...`, and
  `make build` pass; lint reports zero issues.
- The built `kit` binary reports a coherent project contract and all 65
  features pass. Both `pull-request-merge` and `release-orchestration` resolve
  unblocked against the updated rule, workflows, specs, and source hints.
- `kit reconcile --all --output-only` reports no reconciliation need and a
  complete source-file-size audit of 712 version-control-eligible candidates,
  369 eligible handwritten source/test files, and zero files above 300 lines.
- A strict local `kit pr orchestrate --dry-run` contains the in-place repair,
  same-branch non-rewrite, old-head invalidation, fresh authorization, and
  exceptional replacement boundaries and omits the replacement-first phrases.
- The canonical Kit rule and the `kp` `GH-26` downstream rule are byte-identical
  at SHA-256
  `eb4f8af3d0abe5e1b6d3111cad328c9d73b092013c7da99220fd8dac9a6e432d`;
  their pull-request workflow phase and completion sections also match.
- `gitleaks git --redact --no-banner` scans 370 commits and 13.15 MB with no
  leaks; `git diff --check` passes.
- Browser, deployment, infrastructure, live-integration, and production
  validation are `NOT_APPLICABLE`; this change updates local policy, templates,
  and prompt generation without executing a release or provider mutation.
- `GH-180` focused template/rule tests, `make all`, the complete race suite,
  `golangci-lint`, all 69 feature checks, and `git diff --check` pass. The
  reconcile audit checks 378 eligible handwritten source/test files with zero
  above 300 physical lines. Its sole `.kit.yaml` refresh advisory is identical
  on clean `main` and remains outside this issue. Gitleaks scans 287 commits and
  10.82 MB with no leaks.

## OUTCOME

- Active guidance now uses one canonical authority matrix and downstream
  `github-pr-merge` rule. Only a direct request or accepted bounded plan creates
  authority for an exact PR set; delivery consent, checks, assignment, and a
  program ledger never do.
- `MERGE_READY`, `BLOCKED`, and `UNKNOWN` are exact current-head evidence
  states. Only the authorized ready frontier may merge, with independent nodes
  concurrent and dependencies or same-base sensitive operations serialized.
- Identity, repository policy, infrastructure effects, corrective ownership,
  and post-merge hosted/deployment/runtime/production proof remain explicit.
- Routine scope-preserving remediation now stays on the existing pull request
  under bounded repair authority. A changed head reenters as `UNKNOWN` and
  cannot merge without fresh checks, review, revalidation, and exact-head
  authorization; exceptional replacement PRs remain new graph nodes.
- Historical specifications remain unchanged. This active rule supersedes the
  former manual-only merge policy only for explicitly authorized agent work;
  Kit itself still performs no merge mutation.
- Merge waves now use one meaningful current snapshot, source-and-deployment
  independence, bounded recovery, event-driven monitoring, and concise
  terminal evidence. Destructive infrastructure remains outside release waves.

## REPOSITORY MEMORY

Created this V3 specification because authorization sources, authority
non-sources, evidence states, indirect infrastructure effects, identity
boundaries, superseded manual-only policy, and rejected executor behavior are
material rationale that code and tests alone cannot preserve.

Updated it for issue #172 because the distinction between freezing an exact
merge head and preserving the existing PR repair lane is consequential policy
that must survive downstream registry refresh and release-prompt generation.

Updated it for issue #180 because concise preflight, safe cross-deployment
concurrency, bounded mechanical delegation, destructive-infrastructure
isolation, and terse terminal reporting are durable registry policy rather
than a new CLI behavior.
