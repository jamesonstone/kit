---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: 0065
  slug: deletion-safety
  dir: 0065-deletion-safety
relationships:
  - type: builds_on
    target: 0046-autonomous-mutation-recovery
  - type: builds_on
    target: 0057-infrastructure-change-approval
references:
  - id: safety-guardrails
    name: Existing deletion and recovery boundary
    type: documentation
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: universal deletion permission and recovery semantics
    status: active
  - id: instruction-templates
    name: Managed instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: provider-neutral hard-gate routing
    status: active
  - id: context-workflows
    name: Managed context workflows
    type: code
    target: internal/templates/context_workflows/implementation-delivery.md
    relation: implements
    read_policy: must
    used_for: mandatory workflow evidence selection
    status: active
  - id: registry-adoption
    name: Downstream ruleset adoption coverage
    type: code
    target: pkg/cli/init_refresh_ruleset_adoption_test.go
    relation: supports
    read_policy: must
    used_for: Kit refresh distribution
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make recoverability the default deletion behavior in every Kit-managed
project. An unqualified delete must preserve a supported restore path, and an
irreversible hard delete must remain prohibited until a human manually
confirms the exact current targets after seeing the complete consequences.

## CONTEXT

- Existing `safety-guardrails` prefers recoverable deletion where practical
  but requires permission only for large-scale or sensitive-file deletion.
  It does not universally default project behavior to soft delete.
- Existing `infrastructure-change-approval` requires post-outline confirmation
  for infrastructure deletion, but it does not establish soft delete as the
  default for application data, files, identities, artifacts, or other
  persistent project-managed state.
- The user requires one rule for every project that adopts the Kit harness,
  not a Kit-repository-only convention.
- Kit distributes active downstream rulesets through `kit init --refresh` and
  routes them through generated instructions, support references, context
  workflows, and reconciliation expectations.
- Kit preserves immutable `kit instructions` versions. The current v4 bytes
  cannot be changed, so the new universal contract requires an additive v5.
- Existing stricter controls remain in force. The new rule must compose with
  infrastructure approval, legal erasure, security revocation, branch
  prohibitions, and exact-run production-test cleanup without creating
  duplicate confirmation prompts.

## REQUIREMENTS

### Scope And Definitions

- Add an active downstream ruleset named `deletion-safety` with
  `read_policy_default: must`.
- Apply it to implementation, data, persistence, filesystems, identities,
  cloud and infrastructure, APIs, UIs, automation, cleanup, retention,
  migrations, and operational workflows in Kit-managed projects.
- Treat persistent project, user, business, or external-system state as
  covered. Task-owned ephemeral scratch that never became authoritative state
  is not a retained resource and remains governed by the normal exact-scope
  cleanup rules.
- Define soft delete as a reversible lifecycle transition with a supported,
  authorized restore path. Tombstones, archive states, quarantine, disablement,
  retained version history, and provider-native recovery controls may satisfy
  the contract when restoration is real and documented.
- Define hard delete as any action that removes the supported restore path,
  including purge, destroy, force deletion, empty-trash operations, destructive
  replacement, history rewrite, retention expiry, backup or snapshot deletion,
  and irreversible cascades.

### Default Behavior

- Interpret an unqualified request to delete or remove covered state as soft
  delete. Do not infer hard-delete authority from verbs such as delete, remove,
  clean up, decommission, reset, replace, or start over.
- When implementing deletion behavior, make the ordinary API, UI, CLI, service,
  and automation path soft-delete by default and provide an authorized restore
  path with explicit retention and visibility semantics.
- Keep hard delete as a separate privileged and auditable action. Do not hide
  it behind a normal-delete `force`, `hard`, or `permanent` flag that can bypass
  a server-enforced human confirmation boundary.
- Preserve restore authorization, tenant or owner isolation, uniqueness and
  referential-integrity behavior, idempotency, concurrency safety, audit
  history, and lifecycle-state observability.

### Specific Manual Hard-Delete Confirmation

- Before any hard delete, resolve and present the exact target identity or
  bounded selector, current target count, environment or account, dependent
  cascades, why soft delete is insufficient, the fact that supported restore
  will be impossible, backup or recovery state, retention or legal impact, and
  the planned verification evidence.
- After that outline, require a specific manual confirmation from the human
  authorizing irreversible deletion of those exact targets. The confirmation
  may cover one fully enumerated batch.
- Initial requests, general task approval, generic plan approval, automation
  signals, scheduled retention, prior soft-delete approval, earlier confirmation
  for different targets, and broad cleanup intent do not count.
- Bind confirmation to the resolved current targets, actor, action, environment,
  and material consequences. Target drift, expanded cascades, a new environment,
  or a changed recovery posture requires a new outline and confirmation.
- For a bounded selector, bind confirmation to materialized target IDs or an
  immutable snapshot or version token as well as the resolved count. Immediately
  before execution, compare the current target set or version with the
  confirmed snapshot; abort on any difference, including identity changes with
  the same count, and require a new outline and confirmation.
- Product hard-delete flows must enforce an equivalent separate human act and
  record actor, time, target scope, and confirmation evidence. Client-side
  prompts alone are insufficient when the server can still purge directly.

### Composition And Compatibility

- Apply stricter repository, legal, privacy, security, infrastructure, and
  provider controls in addition to this rule. When another gate already
  requires a post-outline deletion confirmation, one confirmation may satisfy
  both only when its outline includes every field required by both rules.
- Revoke or disable compromised credentials immediately through a reversible
  lifecycle action where supported; irreversible erasure still follows the
  hard-delete confirmation boundary unless an external mandatory safety control
  performs it outside project or agent authority.
- Do not retroactively migrate every downstream data model in this issue.
  New and refreshed projects receive the rule, and later feature work must
  apply it whenever deletion behavior is created or changed.
- Preserve all immutable prior instruction versions and make v5 the current
  version containing the new contract.

### Observable Acceptance

- The canonical rule parses as a valid mandatory downstream ruleset and is
  adopted by the normal Kit initialization and refresh path.
- Every managed context workflow selects `deletion-safety` as required evidence.
- V1, V2, and V3 generated provider instructions, generated support docs,
  checked-in V3 guidance, and current versioned instructions route through the
  same soft-delete default and manual hard-delete confirmation boundary.
- Reconciliation detects missing routing or a weakened hard-delete boundary.
- Focused tests verify rule semantics, registry adoption, generated and
  checked-in alignment, workflow selection, immutable v1-v4 preservation, and
  the additive v5 contract.

## ACCEPTED PLAN

1. Create the canonical `deletion-safety` downstream ruleset and document it in
   the references index, safety and testing interactions, project Constitution,
   and this living spec.
2. Add one concise shared deletion-safety hard gate and compose it into every
   supported generated provider instruction and generated guardrails surface.
3. Route the rule through RLM, all seven embedded and checked-in context
   workflows, managed reference templates, and reconciliation expectations.
4. Preserve v1-v4 instruction bytes, add additive instruction version v5, and
   make v5 current.
5. Add focused ruleset, registry-adoption, template, workflow, reconciliation,
   and versioned-instruction tests without exceeding the 300-line source-file
   limit.
6. Run focused tests, formatting, complete tests, race tests, vet, lint, builds,
   Kit project checks, whole-project reconciliation and source-size audits,
   diff and secret review, then curate repository memory and deliver issue #153
   through one ready pull request.

## DECISIONS

- Use a dedicated ruleset rather than expanding `safety-guardrails` into a
  catch-all. Deletion lifecycle design and exact hard-delete confirmation are
  independently reusable responsibilities.
- Treat recoverability as a behavior contract, not a prescribed schema. A
  `deleted_at` column, archive state, quarantine, retained Git history, or
  provider-native recovery control is acceptable only when a supported restore
  path actually exists.
- Require post-outline confirmation for hard delete. A generic initial request
  cannot be sufficiently specific because exact targets, cascades, and current
  recovery state must be resolved first.
- Permit one exact confirmation to satisfy multiple deletion gates when the
  combined outline is complete; duplicate prompts add friction without adding
  authority.
- Exclude only non-authoritative task scratch from retained-state semantics.
  Ambiguous files or resources are covered and therefore default to soft delete.
- Preserve prior instruction versions and publish v5 rather than mutating the
  immutable v4 contract.
- Bind bounded hard-delete selectors to materialized identities or an immutable
  snapshot/version so target-set drift cannot detach confirmation from the
  records that were actually outlined.

## DISCOVERIES

- The existing permission boundary is compatible because it already defers to
  explicit repo-local approval gates; the new rule must be named there so the
  universal exception is obvious and testable.
- All active downstream registry rulesets on Kit `main` are adopted by refresh;
  the focused mandatory-rules test is the explicit regression inventory.
- Context workflows are stored twice: embedded templates under
  `internal/templates/context_workflows/` and exact checked-in mirrors under
  `docs/references/workflows/`.
- Checked-in V3 guidance is generated from shared template constants, while
  `kit instructions` uses a separate immutable version registry.
- Synthetic production cleanup currently authorizes exact-run cleanup but does
  not distinguish soft from hard deletion; it must defer to this new rule.
- CodeRabbit identified that resolved counts alone cannot detect a bounded
  selector changing to different targets. The contract now requires
  materialized identities or an immutable snapshot/version token and an
  immediate pre-execution comparison.
- The generated V2 root instructions had relied on the remaining space below a
  fixed 100-line threshold for small project-owned additions. The mandatory
  gate consumed that incidental space, so reconciliation now uses the current
  generated contract plus an explicit 20-line customization allowance while
  retaining the existing duplicate-manual markers.

## VALIDATION

- Focused semantic, template, workflow, registry-adoption, reconciliation, and
  versioned-instruction tests passed across `internal/templates`,
  `internal/instructions`, `internal/context`, and `pkg/cli`.
- `go test ./... -count=1` and `go test -race ./... -count=1` passed for every
  package.
- `go vet ./...`, `golangci-lint run --new-from-rev=origin/main ./...`,
  `make fmt`, and `make build` passed; lint reported zero issues.
- The branch-built `bin/kit` parsed and rendered `deletion-safety`, listed it
  as an active local ruleset, emitted v5 with the registered immutable hash,
  and resolved `implementation-delivery` with `deletion-safety` as required
  evidence and no blocked diagnostics.
- The branch-built `bin/kit check 0065-deletion-safety`, `bin/kit check --all`,
  and `bin/kit check --project` passed. All 62 feature contracts and the
  project contract were coherent.
- `bin/kit reconcile --all --output-only` reported no reconciliation needed
  and a complete source-file-size audit: 683 version-control-eligible
  candidates, 349 eligible handwritten source/test files, and zero files above
  300 physical lines.
- SHA-256 verification preserved v1-v4 byte-for-byte and registered the final
  snapshot-bound v5 as
  `67122fe42ce3bcb65a9b1f355271395ebe4c65d43fef5b1b9632580cacf5e3d6`.
- `gitleaks dir --no-banner --redact .` scanned 4.70 MB and found no leaks.
- Hosted pull-request checks remain separate delivery evidence and must be
  revalidated on the final pushed head; they do not substitute for local
  validation.

## OUTCOME

- Added the mandatory downstream `deletion-safety` ruleset. Covered deletion
  now defaults to a supported recoverable lifecycle, while hard delete remains
  a separate privileged action prohibited until a human specifically confirms
  the exact current targets after the complete consequence outline.
- Bounded-selector confirmation is snapshot-bound and execution must abort on
  any target-set or version drift, even when the resolved count is unchanged.
- Routed the contract through all generated provider instructions, checked-in
  V3 guidance, the Constitution baseline, support references, all seven context
  workflows, refresh adoption, reconciliation expectations, and testing and
  infrastructure composition rules.
- Preserved immutable instruction versions v1-v4 and made additive v5 current.
- No downstream data model was mass-migrated. New and refreshed Kit projects
  receive the rule, and later deletion work must apply it to existing systems.

## REPOSITORY MEMORY

Decision: created

Rationale: The distinction between soft and hard deletion, exact confirmation
scope, non-authoritative scratch boundary, cross-gate composition, and immutable
instruction-version strategy are consequential cross-project policy decisions
that code assertions alone cannot preserve.

Artifacts:

- `docs/specs/0065-deletion-safety/SPEC.md`
- `docs/references/rules/deletion-safety.md`
- `docs/CONSTITUTION.md`
- `docs/references/rules/safety-guardrails.md`
- `docs/references/rules/infrastructure-change-approval.md`
- `docs/references/rules/testing-and-environment-validation.md`
- managed instruction, workflow, registry, reconciliation, and
  versioned-instruction sources and tests
