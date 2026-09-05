---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0076"
  slug: "standing-merge-deploy-authority"
  dir: "0076-standing-merge-deploy-authority"
relationships:
  - type: builds_on
    target: 0075-autonomous-non-destructive-execution
    note: Narrows accepted-task autonomy to explicit bounded standing authority and restores hard stops for infrastructure-class and nonstandard deployment effects.
  - type: builds_on
    target: 0061-authorized-coding-agent-merge-autonomy
    note: Preserves exact current-head readiness while removing repeated authorization after in-scope head refreshes.
  - type: related_to
    target: 0073-infrastructure-approval-scope
    note: Retains the routine-application-operation distinction but does not let standing deployment authority cover infrastructure creation or replacement.
  - type: related_to
    target: 0065-deletion-safety
    note: Destructive and data-loss effects remain outside standing authority.
references:
  - id: merge-rule
    name: GitHub PR merge
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: implements
    read_policy: must
    used_for: standing authority, current-head readiness, and pause boundaries
    status: active
  - id: infrastructure-rule
    name: Infrastructure change approval
    type: rule
    target: docs/references/rules/infrastructure-change-approval.md
    relation: constrains
    read_policy: must
    used_for: standard deployment and infrastructure-class exclusions
    status: active
  - id: program-rule
    name: Cross-repository program coordination
    type: rule
    target: docs/references/rules/cross-repository-program-coordination.md
    relation: constrains
    read_policy: must
    used_for: dynamic in-scope PR binding and downstream LabCore propagation
    status: active
  - id: program-ledger
    name: Standing merge and deployment authority program
    type: reference
    target: docs/programs/standing-merge-deploy-authority/PROGRAM.md
    relation: informs
    read_policy: must
    used_for: Kit and LabCore delivery sequencing
    status: active
  - id: issue
    name: Allow standing merge and deploy authority
    type: external
    target: https://github.com/jamesonstone/kit/issues/198
    relation: supports
    read_policy: must
    used_for: accepted scope and delivery identity
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Let a human grant bounded standing authority for an accepted task, goal, or
program so agents can create and repair its governed pull requests, merge each
eligible current head, run only the repositories' existing standard deployment
workflows, verify runtime, and resume the goal without repetitive permission
prompts. Keep every safety and evidence gate that protects scope, identity,
policy, infrastructure, data, and production correctness.

## CONTEXT

- Feature 0075 treated every accepted task or active `/goal` as standing
  authority for non-destructive merge, deployment, production activation, and
  additive IAM, network, resource, and IaC mutation. That is broader than this
  request: standing authority must be explicitly granted for a bounded goal,
  and it must cover only merges and standard deployments.
- Active merge guidance already separates `MERGE_READY`, `BLOCKED`, and
  `UNKNOWN`, but several canonical and generated surfaces still require the
  "exact in-scope PR set" to be known before mutation. This makes later blocker
  PRs and current-head refreshes appear to require renewed consent.
- Exact PR numbers and head OIDs are evidence bound at the mutation boundary;
  they do not need to be known when the human grants semantic task-level or
  program-level authority.
- Kit owns the canonical downstream rule registry, instruction templates,
  workflow templates, reconcile audits, and refresh behavior. LabCore must be
  refreshed only after those outputs are stable.
- Coordinator: `jamesonstone/kit`. Canonical program ledger:
  `docs/programs/standing-merge-deploy-authority/PROGRAM.md`.
- Topology: single supervisor lane. The policy, templates, workflows, audits,
  golden output, and tests overlap tightly; LabCore propagation is serialized
  behind the exact Kit source commit. The active host policy does not authorize
  child-agent execution for this task.

## REQUIREMENTS

- Standing authority exists only when a human explicitly authorizes a bounded
  task, goal, or program to merge and/or deploy its resulting work. Accepting a
  generic implementation task, opening a lane, approving a PR, passing checks,
  or recording a ledger does not silently add merge or deployment authority.
- Record the authority source, goal and non-goals, repositories, protected
  bases, deployment environments, permitted actions, standard deployment
  workflows, actor, expiry/completion boundary, and explicit exclusions.
- A semantic selector may bind later-created pull requests when each PR is
  directly required by the authorized outcome, belongs to a governed in-scope
  lane, targets an authorized repository/base, and introduces no materially
  different effect. Materialize the exact current PRs and head OIDs during
  pre-merge reconciliation; do not ask again merely because those identifiers
  were unknown when authority was granted.
- An in-scope repair, base refresh, regenerated artifact, or ordinary current-
  head update invalidates prior readiness evidence but not standing authority.
  Re-run all current-head reviews, checks, mergeability, policy, dependency,
  identity, and effect classification; merge only after the refreshed node is
  `MERGE_READY`.
- Standing deployment authority covers only an existing repository-approved
  standard deployment workflow for an authorized environment and the exact
  artifact produced by an authorized merge. It includes normal image or
  application-artifact rollout onto already-provisioned targets and required
  runtime verification. It does not authorize novel provider commands,
  workflow changes, new targets, or materially different effects.
- Pause before scope, repository, base, environment, actor, identity, merge
  method, deployment workflow, or material-effect expansion. Also pause for
  admin/force/protection/review/check bypass; pending, missing, skipped without
  verified eligibility, stale, failed, or unattributed required checks; IAM,
  network, KMS, secrets, database schema or data-loss changes; infrastructure
  creation, replacement, or deletion; destructive deletion; nonstandard
  deployment effects; or any unresolved risk classification.
- Direct current human instructions take precedence. An explicit pause, hold,
  or revocation immediately prevents the affected mutation set and its
  dependents. Resume requires an explicit human instruction; a passing check,
  new head, retry, new session, or unchanged ledger cannot revive authority.
- The narrowest explicit instruction wins. A scoped pause may stop one
  repository, environment, PR class, or action while independent authorized
  work continues. Full revocation ends the authority. Goal completion or the
  recorded expiry boundary ends it automatically.
- Preserve exact pre-merge reconciliation and post-deployment verification as
  mandatory evidence, not authorization ceremonies.
- Update canonical rules, generated entrypoints, support docs, release prompt,
  refresh/reconcile behavior, and tests so every Kit-managed project receives
  the semantics through normal refresh/reconcile.
- Refresh LabCore's actual managed entrypoints and required managed rule copies
  from the exact updated Kit commit, preserving LabCore-local production and
  AWS constraints.
- Regression scenarios must prove: (a) later in-scope blocker PRs and refreshed
  heads merge/deploy under standing authority after current checks; (b) scope,
  actor/policy, failed/missing checks, destructive, infrastructure-class, data,
  and nonstandard deployment changes pause; and (c) explicit pause/revocation
  remains effective until explicit resume.
- Non-goals: implement a merge or deployment executor in Kit; merge or deploy
  these governance PRs; weaken current-head checks, reviews, branch protection,
  identity, infrastructure approval, deletion safety, or production evidence;
  authorize arbitrary future repositories, environments, or effects.

## ACCEPTED PLAN

1. Reconcile feature 0075 and the current remote rule/template/audit chain,
   then define the narrower standing-authority selector, precedence, pause,
   revocation, readiness, and standard-deployment contract here.
2. Update the canonical Kit rules, generated hard-gate and tooling templates,
   workflow and release-prompt sources, Constitution, and versioned/current
   instruction surfaces without editing historical instruction versions.
3. Update reconcile semantic audits and init/refresh materialization so stale
   exact-set/per-head-consent language is detected and managed downstream
   copies adopt the new contract safely.
4. Add focused scenario, template, golden, rule-selection, init-refresh, and
   reconcile tests for dynamic in-scope PR binding, current-head refresh,
   safety pauses, and pause/revocation precedence.
5. Run focused and full Kit validation, audit the affected source/test scope,
   curate repository memory, and deliver one ready PR from `GH-198`.
6. Build the exact Kit source, preview LabCore's managed refresh, apply only the
   required managed instruction/rule/workflow updates in `GH-507`, preserve
   local customizations, run LabCore checks and reconcile convergence, and
   deliver one ready PR.
7. Checkpoint exact issues, branches, heads, PRs, validation, and remaining
   landing order in the coordinator ledger. Stop before merge or deployment.

## DECISIONS

- Accepted: authority is semantic at grant time and exact at each mutation
  boundary. Unknown future PR numbers or head OIDs do not force repeated
  permission prompts.
- Accepted: changed heads lose readiness evidence, not in-scope standing
  authority.
- Accepted: standard deployment is an allowlisted existing workflow and
  environment, not a synonym for every non-destructive cloud or IaC mutation.
- Accepted: current explicit pause/revocation outranks earlier standing
  authority and cannot be cleared by automation.
- Superseded from 0075: generic accepted-task scope is not enough by itself,
  and additive IAM/network/resource/IaC work is not covered by standing
  merge/deploy authority.
- Rejected: exact PR/head enumeration at grant time. It prevents a bounded goal
  from autonomously repairing blockers that do not yet have delivery IDs.
- Rejected: weakening exact current-state reconciliation. The friction to
  remove is authorization repetition, not safety validation.

## DISCOVERIES

- Feature 0075 had already removed repeated merge consent, but did so by
  granting every accepted task broad non-destructive merge, production
  activation, additive IAM/network/resource, and IaC authority. The required
  fix is a narrowing and precision pass, not simply removing old exact-head
  wording.
- Current instruction generation has four coupled surfaces: concise merge and
  infrastructure gate constants, V1/V2/V3 composition templates, checked-in
  entrypoints refreshed by `kit init --refresh`, and additive `kit instructions`
  versions. Reconcile adds a separate semantic audit for local-custom policy
  copies.
- LabCore uses instruction scaffold v2. Its `github-pr-merge` rule is managed,
  while `infrastructure-change-approval` is `local-custom` because live and
  production deployments require the `lsmc` AWS identity and stricter effects.
  Downstream propagation must update that rule semantically rather than replace
  it with the generic Kit copy.
- `kit spec` still regenerates unrelated historical progress-summary text from
  filesystem metadata. Restored `PROJECT_PROGRESS_SUMMARY.md` to `origin/main`
  plus only the 0076 row and summary before validation.
- Current `main` contained one 301-line handwritten source file,
  `internal/templates/instruction_templates_v2.go`, which blocked project
  validation. Moving the cohesive `agentsRLM` template to
  `instruction_templates_v2_rlm.go` preserved output and brought both files
  below the limit.
- PR #199 review confirmed five valid gaps. The resolved release policy needed
  all standard-deployment bounds; changed-head workflow wording needed to keep
  agent repair permission separate from standing merge authority; forbidden
  phrase checks needed symmetric whitespace normalization; the semantic audit
  needed the exact-current `MERGE_READY` gate; and unreadable policy documents
  needed visible warnings rather than silent skips.
- CodeRabbit's blocker-repair finding was too broad if applied literally to
  every changed head. Agent-performed source, commit, or push remediation
  requires explicit blocker-repair permission, but a selector-matching human or
  external head change retains standing merge authority after it returns from
  `UNKNOWN` to exact-current `MERGE_READY` with fresh evidence.

## VALIDATION

- PASS: focused policy, template, versioned-instruction, release-prompt,
  `kit init --refresh`, `kit health`, and `kit reconcile --include-files`
  tests via `go test ./internal/instructions ./internal/templates
  ./internal/releaseprompt ./pkg/cli -count=1`.
- PASS: `go test ./... -count=1`.
- PASS: `go test -race ./... -count=1`.
- PASS: `go vet ./...`, `golangci-lint run ./...` with zero issues,
  `go build ./...`, and `make build`.
- PASS: `./bin/kit check 0076-standing-merge-deploy-authority` and
  `./bin/kit check --project`.
- PASS: targeted `./bin/kit init --refresh --force --dry-run --diff` reports
  zero planned changes for AGENTS, CLAUDE, Copilot, GUARDRAILS, and TOOLING.
- PASS: `./bin/kit reconcile --all --output-only` reports no reconciliation
  needed and a complete audit of 746 candidates, 385 eligible handwritten
  source/test files, and zero files above 300 lines.
- PASS: LabCore full `go test ./... -count=1`, `go vet ./...`,
  `golangci-lint run --new-from-rev=origin/main ./...`, `make docs-check`, Kit
  config check, repository-maintenance context resolution, managed-rule and
  workflow byte comparisons, three-entrypoint section comparisons,
  `git diff --check`, and diff-scoped gitleaks.
- PASS: LabCore ready PR #508 at `57ad7a446ab5c19bb14a51a82e41d523654b78d4`
  is human-authored and assigned, mergeable, and preserves the Docs-Site
  `not-required` evidence.
- PENDING: LabCore CodeQL checks and Kit CodeRabbit. Kit hosted `validate` and
  auto-assignment passed on policy commit `7d08f7e` before this checkpoint
  update.
- KNOWN PRE-EXISTING: full LabCore lint reports 66 findings; Kit project
  validation reports 44 blocking historical spec/progress/source-size findings;
  the whole-tree gitleaks scan reports four existing documentation examples.
  The diff-scoped lint, docs, tests, vet, and gitleaks checks pass and none of
  those findings is in this change.
- PASS: PR-review repair focused suites for `internal/releaseprompt`,
  `internal/templates`, and `pkg/cli`; full `go test ./... -count=1`; full
  `go test -race ./... -count=1`; `go vet ./...`; `golangci-lint run ./...`;
  `go build ./...`; `make build`; workflow mirror comparisons; `gofmt` and
  `git diff --check`.
- PASS: fresh independent read-only verification of D001-D005 and the complete
  16-file repair diff. The verifier confirmed every finding, test boundary,
  mirror, source-size limit, and repair-versus-merge authority distinction.
- PENDING: final commit/push, current-head hosted validation, reflection, and
  verified review-thread resolution.

## OUTCOME

- Kit now distinguishes generic task acceptance from explicit bounded standing
  merge/deploy authority. The standing selector may bind later governed PRs and
  refreshed heads, but each exact current node still needs full `MERGE_READY`
  evidence.
- Standard deployment is limited to a named existing workflow, authorized
  environment and actor, exact merged artifact, already-provisioned resources,
  and required runtime/rollback verification.
- Scope/repository/base/environment/actor/identity/method/workflow/effect
  expansion, unsafe policy or checks, IAM/network/KMS/secrets, database schema
  or data loss, infrastructure creation/replacement/deletion, destructive, and
  nonstandard deployment effects remain separate stops.
- Direct human pause, hold, or revocation has precedence and persists until
  explicit resume or replacement authority; completion or expiry ends the
  grant.
- Additive instruction version v12 preserves immutable v1-v11, and generated
  entrypoints, workflows, release prompts, audits, and regression tests share
  the same contract.
- LabCore's three entrypoints now agree on the standing merge/deploy contract.
  Managed merge/team/program/deadline rules and merge/release workflows are
  byte-identical to Kit policy commit `7d08f7e`; local-custom delivery, safety,
  work-lane, and infrastructure rules preserve stricter LabCore policy.
- `.github/workflows/deploy.yaml` is the only current LabCore standard
  deployment path eligible for a named standing grant; live/production still
  requires immediate `kit aws verify` for the configured `lsmc` identity.
- Kit PR #199 and LabCore PR #508 are ready for human review. Merge and
  deployment remain explicitly unperformed and unauthorized.
- PR #199's repair now closes D001-D005 with focused regressions. Because D002
  and D004 update managed merge workflows/rules after policy source `7d08f7e`,
  LabCore PR #508 must be refreshed in its own existing lane before it can
  claim parity with the final Kit head.

## REPOSITORY MEMORY

Decision: created.

Rationale: standing-authority selectors, current-head evidence binding,
standard-deployment limits, and pause/revocation precedence are consequential
cross-project governance that code and substring tests cannot fully preserve.

Artifacts:

- `docs/specs/0076-standing-merge-deploy-authority/SPEC.md`
- `docs/programs/standing-merge-deploy-authority/PROGRAM.md`
- `docs/CONSTITUTION.md`
- `docs/references/rules/github-pr-merge.md`
- `docs/references/rules/infrastructure-change-approval.md`
- `docs/references/rules/safety-guardrails.md`
- generated entrypoints, workflows, release prompt, semantic audit, and v12
  instruction sources listed by the delivered diff
- LabCore repository memory: not required; the downstream change is a managed
  instruction/rule refresh whose durable rationale and sequencing live in this
  Kit spec and the coordinator program ledger.
