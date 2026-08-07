---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0057
  slug: infrastructure-change-approval
  dir: 0057-infrastructure-change-approval
relationships:
  - type: builds_on
    target: 0041-aws-config-schema
  - type: builds_on
    target: 0046-autonomous-mutation-recovery
references:
  - id: instruction-template
    name: Managed instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: provider-neutral hard-gate routing
    status: active
  - id: guardrails-template
    name: Managed guardrails template
    type: code
    target: internal/templates/instruction_support_templates.go
    relation: implements
    read_policy: must
    used_for: downstream approval and autonomy boundary
    status: active
  - id: references-index
    name: Downstream ruleset index
    type: documentation
    target: docs/references/README.md
    relation: guides
    read_policy: must
    used_for: managed ruleset discovery and scope
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Require Kit-managed coding agents to outline public-cloud, Kubernetes, and
infrastructure-as-code mutations and obtain user confirmation before changing
them, while allowing each approved bounded batch to run to completion in one
pass without routine approval interruptions and always confirming deletion or
removal after its summary.

## CONTEXT

- Kit currently has an AWS identity gate, but it does not require an upfront
  mutation outline and does not cover GCP, Azure, Kubernetes, or
  provider-neutral infrastructure-as-code.
- The user selected cloud and infrastructure-as-code scope rather than the
  broader control-plane scope that would automatically include adjacent SaaS
  and general CI/CD settings.
- The user requires approval before any covered mutation, including edits to
  infrastructure-as-code sources as well as live provider changes.
- The approval interaction must be consolidated at the beginning so the agent
  can work autonomously through the exact approved batch.
- Infrastructure changes should be outlined as part of the task plan when one
  is used, and approval of that plan should count as the batch confirmation.
- A sufficiently detailed initial request may itself satisfy a non-deletion
  batch when it includes the required outline and clearly authorizes the
  bounded mutations.
- Deleting or removing infrastructure is the exception: the agent must always
  summarize the complete deletion batch and receive confirmation afterward,
  but that one confirmation authorizes the full deletion batch.

## REQUIREMENTS

- Add an active, mandatory, downstream-managed ruleset for public-cloud and
  infrastructure-as-code changes.
- Cover direct cloud commands, APIs, SDKs, or consoles for AWS, GCP, and Azure;
  Kubernetes mutations; Terraform, Pulumi, CloudFormation, CDK, Bicep, and
  comparable infrastructure-as-code source or apply operations.
- Exclude adjacent infrastructure SaaS and general CI/CD configuration unless
  the operation directly invokes a covered cloud or infrastructure-as-code
  mutation.
- Permit read-only discovery before approval only when it does not alter cloud
  resources, remote state, or repository-owned infrastructure source.
- Before the first covered mutation, require one consolidated outline that
  states the exact target context, intended resource actions, execution
  boundary, material availability/data/security/cost impact, rollback or
  recovery path, and validation evidence.
- Put the infrastructure outline into the task plan when planning is used and
  treat approval of that complete plan as the single batch confirmation.
- Require one explicit user confirmation after the outline unless a
  non-deletion initial request already contains and authorizes the complete
  batch.
- Always require explicit confirmation after a deletion or removal outline;
  the initial-request exception does not apply to destructive batches.
- Scope confirmation to the exact bounded batch. Within that scope, continue
  implementation, application, validation, routine failure recovery, and the
  rest of the task in one pass without command-by-command approval.
- When additional covered changes become necessary, collect all then-known
  changes into one follow-up outline, obtain one confirmation, and execute the
  complete follow-up batch in one pass.
- Treat provider identity, environment, region or cluster, resource set,
  action type, material impact, rollback path, or intended-effect changes
  outside the approved outline as one follow-up batch, not repeated prompts.
- Keep the existing AWS identity verification gate additive: an approved AWS
  batch still requires the configured account and credentials to verify before
  AWS-dependent work and immediately before AWS mutation.
- Route supported generated instruction versions and checked-in V3 provider
  instructions through the rule, and make stale downstream guidance visible
  to reconciliation.
- Prove exact ruleset validity, mandatory registry adoption, generated
  instruction routing, checked-in alignment, and reconciliation expectations.
- Observable acceptance: new and refreshed projects receive the rule;
  applicable agents encounter the hard gate before mutation; one approval
  authorizes the complete unchanged batch; follow-up changes are consolidated;
  and deletion or removal always receives post-outline confirmation.

## ACCEPTED PLAN

1. Add the canonical `infrastructure-change-approval` downstream ruleset and
   list it in the managed references index.
2. Add one concise provider-neutral approval gate to V1, V2, and V3 managed
   instruction templates plus generated guardrails, preserving the narrower
   AWS identity gate as a second required check.
3. Align checked-in V3 `AGENTS.md`, `CLAUDE.md`, Copilot instructions,
   guardrails, and RLM routing with the template source.
4. Extend reconciliation expectations and focused template, ruleset, and
   downstream-adoption tests so weakened or missing approval guidance is
   detected.
5. Run focused and complete Go validation, generated-output and reconciliation
   checks, source-file-size audit, diff and secret review, then curate durable
   repository memory and deliver issue #131 as a ready pull request.

## DECISIONS

- "Before any mutation" includes both infrastructure-as-code source edits and
  live cloud or cluster mutations.
- Approval is one exact bounded batch, not one prompt per command and not an
  unbounded whole-task grant.
- Approval of a task plan containing the complete infrastructure outline is
  the preferred confirmation UX and must not be followed by routine prompts.
- Detailed initial instructions can count as approval for non-deletion work
  only when they contain the same target, action, impact, rollback, and
  validation information that a separate agent outline would require.
- Deletion or removal always requires confirmation after its consolidated
  outline, including when the initial request already asked for it; one
  confirmation covers every named destructive action in the batch.
- Additional infrastructure work is a follow-up batch: summarize all then-known
  changes once, confirm once, and execute once.
- Routine retries and compatible tool changes inside the approved target and
  intended effect remain autonomous under the existing mutation-recovery
  contract.
- Read-only discovery may precede approval so the initial outline can be
  evidence-based.

## DISCOVERIES

- `docs/references/rules/*.md` on Kit's `main` branch is the managed registry;
  downstream `kit init --refresh` adopts valid downstream-scoped rules that it
  does not already track.
- Generated provider instructions currently route AWS mutations through an
  identity-only gate, which is complementary to rather than a substitute for
  the new approval contract.
- V3 reconciliation uses exact guidance expectations for checked-in provider
  files and support docs, so the approval route can be made observable without
  adding a command or a new mutation mechanism.
- Review of the completed implementation exposed that the detailed ruleset
  covered Kubernetes while the concise hard gate and discovery routes named
  only public-cloud and infrastructure-as-code mutations. Direct Kubernetes
  mutations now remain explicit at every routing boundary.
- User review clarified that the approval boundary is a planning UX, not a
  command-interruption mechanism: approve each complete batch once, execute it
  and the rest of the task in one pass, and use the same one-summary UX for a
  later batch. Deletion or removal remains the explicit exception that always
  requires post-summary confirmation.
- The existing autonomous-recovery permission boundary said deletion was the
  only reason to request permission. It now defers to explicit repo-local
  approval gates so infrastructure confirmation and routine recovery do not
  contradict each other.
- No Kit command surface changes are needed; command-capability metadata is
  therefore outside this feature.

## VALIDATION

- PASS: focused ruleset, safety-boundary, V1/V2/V3 template,
  registry-adoption, checked-in alignment, reconciliation-expectation, and
  stale-guidance tests.
- PASS: `make fmt` and `git diff --check`.
- PASS: `go test ./... -count=1`; `internal/worktree` completed in 72.033s
  and `pkg/cli` completed in 32.042s on the captured final run.
- PASS: `go vet ./...`.
- PASS: `go test -race ./internal/templates ./pkg/cli -count=1`.
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` with zero
  issues.
- PASS: standalone `bin/kit` and `bin/git-wt` builds.
- PASS: built-binary feature check and all 54 feature checks.
- EXPECTED PRE-CURATION BLOCK: the first built-binary `check --project`
  reported only the missing 0057 progress-table row and feature-summary
  heading; this evidence triggered the required rollup curation.
- PASS: built-binary whole-project reconcile completed its source-file audit
  over 826 version-control-eligible candidates and 517 eligible handwritten
  source/test files with zero files above 300 physical lines; its only two
  findings were the same missing progress-summary entries.
- PASS: targeted generated-file dry-run showed no drift for AGENTS, Claude,
  Copilot, guardrails, or RLM. The local Kit-maintainer ruleset table remains a
  reviewed extension beyond the generated downstream references starter.
- PASS: after Constitution and progress-rollup curation, built-binary
  `check 0057-infrastructure-change-approval` and `check --project` both
  reported coherent contracts.
- PASS: final built-binary whole-project reconcile reported no reconciliation
  needed and repeated the complete 826-candidate, 517-eligible-file,
  zero-violation source-file audit.
- PASS: review repair added direct Kubernetes scope to the shared gate, RLM
  route, references index, checked-in mirrors, reconciliation expectations,
  and focused regression assertions without changing approval semantics.
- PASS: repair validation completed `make fmt`, `git diff --check`, focused
  template and CLI tests, `go test ./... -count=1`, `go vet ./...`, targeted
  race tests, new-diff golangci-lint, and standalone Kit and git-wt builds.
- PASS: the built repair binary validated feature 0057, all 54 features, and
  the project contract; whole-project reconciliation reported no semantic
  reconciliation need and repeated the 826-candidate, 517-eligible-file,
  zero-violation source-size audit.
- PASS: an independent read-only verifier found no blocking or non-blocking
  gaps across D001, generated and checked-in mirrors, approval semantics,
  regression coverage, and affected-file size limits.
- EXPECTED MANAGED-REFRESH DRIFT: a forced include-files dry run reported three
  refresh targets because the registry `main` source does not yet contain this
  branch's new approval rule and Kit-maintainer reference extensions. Applying
  that refresh would discard in-scope branch content, so no refresh was
  applied.
- PASS: one-confirmation UX refinement focused tests cover task-plan approval,
  the non-deletion initial-request exception, deletion/removal confirmation,
  one-pass execution, consolidated follow-up batches, no re-confirmation, and
  compatible retry autonomy.
- PASS: refinement validation completed `make fmt`, `git diff --check`,
  `go test ./... -count=1`, `go vet ./...`, targeted race tests, new-diff
  golangci-lint, and standalone Kit and git-wt builds.
- PASS: the built refinement binary validated feature 0057, all 54 features,
  the project contract, and whole-project reconciliation with 826 candidates,
  517 eligible source/test files, and zero files above 300 physical lines.
- PASS: an independent read-only verifier found no blocking or non-blocking
  gaps in planning integration, confirmation cardinality, destructive-action
  handling, follow-up batching, mirror alignment, regression coverage, or
  affected source/test file sizes.

## OUTCOME

- Kit now exposes a mandatory downstream
  `infrastructure-change-approval` ruleset covering public clouds,
  Kubernetes, and infrastructure-as-code source and apply mutations.
- V1, V2, and V3 managed instructions require one complete bounded outline and
  one plan-level confirmation before covered mutation, execute the batch and
  remaining task work in one pass, consolidate later required changes into one
  follow-up batch, and never re-prompt for already approved actions.
- A fully detailed and authorized initial request may satisfy only a
  non-deletion batch. Deletion or removal always requires explicit confirmation
  after the complete outline, with that one confirmation authorizing the whole
  destructive batch.
- The existing AWS identity gate remains additive, and autonomous failure
  recovery remains active after the exact infrastructure batch is approved.
- New projects and existing projects that run managed refresh adopt the rule
  from Kit's `main` registry without a new command surface.

## REPOSITORY MEMORY

Decision: created

Rationale: The plan-level approval boundary, scope exclusions, non-deletion
initial-request exception, one-pass batch lifetime, follow-up batching, and
deletion confirmation rule are consequential cross-project policy decisions
whose rationale is not recoverable from template assertions alone.

Artifacts:

- `docs/specs/0057-infrastructure-change-approval/SPEC.md`
- `docs/references/rules/infrastructure-change-approval.md`
- managed instruction, guardrail, registry-adoption, and reconciliation sources
- `docs/CONSTITUTION.md`
- `docs/PROJECT_PROGRESS_SUMMARY.md`
