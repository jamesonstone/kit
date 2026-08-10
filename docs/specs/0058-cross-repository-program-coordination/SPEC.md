---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0058"
  slug: "cross-repository-program-coordination"
  dir: "0058-cross-repository-program-coordination"
relationships:
  - type: builds_on
    target: 0042-native-plan-repository-memory
  - type: builds_on
    target: 0030-reference-graph-routing
  - type: related_to
    target: 0012-default-subagent-orchestration
  - type: related_to
    target: 0053-testing-and-environment-validation
references:
  - id: program-coordination-rule
    name: Cross-repository program coordination rule
    type: rule
    target: docs/references/rules/cross-repository-program-coordination.md
    relation: implements
    read_policy: must
    used_for: coordinator ownership, durable checkpoints, reconciliation, and handoff
    status: active
  - id: agent-team-rule
    name: Agent team orchestration rule
    type: rule
    target: docs/references/rules/agent-team-orchestration.md
    relation: guides
    read_policy: must
    used_for: per-wave execution topology and concurrency limits
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: guides
    read_policy: must
    used_for: repository-local validation and evidence ownership
    status: active
  - id: instruction-templates
    name: Repository instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: coding-agent activation of the program coordination rule
    status: active
  - id: rules-registry
    name: Rules registry
    type: code
    target: pkg/cli/rules_registry.go
    relation: implements
    read_policy: must
    used_for: downstream distribution of the program coordination rule
    status: active
skills:
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the validated issue branch as a ready pull request
    required: true
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Ensure that an accepted plan spanning multiple repositories, dependent
deliverables, staged deployments, or agent handoffs retains enough operational
state to resume safely and transfer accountability without relying on one chat,
one agent's context, or an overloaded task checklist.

## CONTEXT

- Kit already preserves feature-level rationale in each repository's
  `SPEC.md`, routes execution topology through `kit dispatch`, and provides
  local resume, handoff, map, state, and validation surfaces.
- Those surfaces are repository- and feature-scoped. None owns a program-wide
  dependency graph, ready frontier, deployment state, global acceptance gates,
  or a reconciled cross-repository handoff checkpoint.
- The existing `agent-team-orchestration` rule governs supervisor and subagent
  topology within an execution wave. It does not preserve program state across
  repositories, sessions, deployments, or agents.
- A new standalone command family would duplicate existing Kit verbs before
  the artifact and lifecycle contract are proven. This change therefore adds
  the durable behavioral rule and thin routing only.
- Repository-local specs, delivery records, runbooks, and validation evidence
  remain authoritative. Program coordination must point to that truth rather
  than copy it into a second requirements system.
- GitHub issue #135 and branch `GH-135` own this Kit-only change.

## REQUIREMENTS

- Add an active downstream `cross-repository-program-coordination` ruleset with
  `read_policy_default: must`.
- Trigger the rule before implementation or resumption only when an accepted
  plan spans multiple repositories and also has dependent deliverables, staged
  deployment or activation, or expected agent/session handoff.
- Require one coordinator repository and one canonical program ledger; retain
  detailed requirements, implementation state, delivery state, runbooks, and
  evidence in their authoritative participant repositories.
- Require stable program, milestone, workstream, and gate identities; an
  explicit dependency graph; and a current ready frontier.
- Track implementation, GitHub delivery, deployment/runtime, and validation as
  separate state dimensions instead of one overloaded completion status.
- Require exact pointers and evidence, including repository identity, local
  spec, issue/branch/PR, commit SHA, target environment, deployment identity,
  observation time, result, and evidence location when applicable.
- Require explicit interface, compatibility, rollout, rollback, blocker-owner,
  impact, and unblock-condition state where relevant.
- Keep one accountable program supervisor. Participant agents may update their
  assigned workstream but may not expand program scope or mark the program
  complete.
- Dispatch only the current ready frontier and continue to obey the existing
  agent-team concurrency and overlap rule within each wave.
- Checkpoint after material milestones, deployment changes, blockers,
  approvals, user redirects, context compaction, and handoff.
- Reconcile recorded state against live repositories, GitHub, and runtime
  evidence before resume, handoff, dispatch, or completion; stale or
  unobserved claims become unknown rather than assumed current.
- Add a compact activation gate to every generated coding-agent instruction
  version, plus just-in-time RLM, tooling, references-index, and reconcile
  coverage so agents reliably discover the rule.
- Keep all changed handwritten source and test files at or below 300 physical
  lines.
- Observable acceptance: the ruleset validates, generated V1/V2/V3 agent
  instructions contain the activation gate, support docs route the rule,
  stale guidance is detected by reconcile tests, downstream registry refresh
  tests include the rule, and the full repository validation suite passes.
- Non-goals: implement a `kit program` command, program artifact parser,
  state.json schema, cross-repository runtime reconciler, deployment driver,
  or a parallel family of program-only commands.

## ACCEPTED PLAN

1. Define the complete coordination contract in one downstream ruleset,
   including trigger boundaries, canonical ownership, state model, evidence,
   checkpoint/reconciliation behavior, handoff, anti-patterns, and verification.
2. Add one compact generated instruction gate that tells coding agents when to
   load and follow the rule without duplicating the ruleset body.
3. Route the rule through RLM, dispatch guidance, and the references index;
   preserve existing command ownership and per-repository sources of truth.
4. Extend template, ruleset, registry-refresh, and reconciliation tests with
   focused files or minimal additions that preserve the source-size contract.
5. Validate focused behavior, generated-file equality, the complete Go suite,
   race/build/static checks, rule and project checks, and changed-file size.
6. Curate the final spec and any demonstrated project-wide invariant, then
   deliver issue #135 as a ready pull request.

## DECISIONS

- Accepted: introduce a rule and activation gate now because durable agent
  behavior is the missing layer and can be shipped without inventing premature
  command semantics.
- Accepted: describe a single canonical program ledger as the coordination
  boundary while keeping participant repositories authoritative for local
  detail.
- Rejected: make `program` a new peer command family. Future implementation
  should extend existing scaffold, map, state, check, resume, handoff, and
  dispatch verbs only after the ledger contract is proven.
- Rejected: use agent transcripts, chat summaries, copied specs, or generated
  JSON as canonical program memory.

## DISCOVERIES

- Existing infrastructure-approval templates and tests establish the supported
  pattern for a rule-backed activation gate across V1, V2, and V3 instruction
  artifacts.
- Registry refresh already installs every visible downstream ruleset, so the
  new rule needs downstream metadata and coverage rather than a new installer.
- Several template and test files are near the 300-line ceiling; new focused
  files and shared constants are required instead of growing large files.
- Reconciliation guidance can test the V3 root gate, RLM route, and tooling
  frontier independently, making weakened activation or execution guidance
  observable without duplicating the full rule in generated instructions.
- Pull-request review exposed that V2 reconciliation covered the RLM route but
  not the generated tooling frontier. V2 and V3 tooling now have symmetric
  stale-guidance expectations and focused removal tests.
- The deterministic prompt-system suite remained unchanged and fully stable;
  the new rule is repository instruction context rather than a new prompt or
  command implementation.
- A single supervisor lane was required by the active runtime instructions;
  no specialist or verification agents were spawned.

## VALIDATION

- PASS: focused template, ruleset, mandatory registry-adoption, and stale
  reconciliation tests.
- PASS: `make fmt`, `git diff --check`, `go vet ./...`,
  `go test ./... -count=1`, and `go build ./...`.
- PASS: `go test -race ./internal/templates ./pkg/cli -count=1`.
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` with zero issues.
- PASS: `go run ./cmd/kit check --project` reported a coherent project
  contract, and `go run ./cmd/kit check --all` passed all 55 features.
- PASS: `go run ./cmd/kit reconcile --output-only` reported no reconciliation
  need and audited 839 version-control-eligible candidates, 521 eligible
  handwritten source/test files, and zero files above 300 physical lines.
- PASS: every affected Go file is at or below 300 lines; the largest affected
  file is `internal/templates/instruction_templates_v2.go` at 282 lines.
- PASS: prompt-system run `20260810T160834.958631000Z-da47ef` passed all 45
  task runs and 345 assertions with 15/15 repeated tasks stable and a 1.0
  determinism rate.
- PASS: `gitleaks git --redact --no-banner` scanned 299 commits and found no
  leaks.
- PASS: review repair added V2 tooling-frontier reconciliation coverage; its
  focused template and CLI tests passed before the complete validation rerun.
- EXPECTED UNRELATED MANAGED REFRESH: `kit status` and the reviewed
  `kit reconcile --include-files --dry-run --diff --output-only` preview only
  existing `.kit.yaml` registry metadata updates from Kit's current `main`.
  Applying that unrelated refresh is outside issue #135, so no managed-file
  mutation was performed.
- NOT APPLICABLE: browser, live-integration, deployment, and production
  validation; this change adds a rule, generated instruction routing, and
  deterministic reconciliation behavior without a runtime service.

## OUTCOME

- Added the active downstream
  `cross-repository-program-coordination` ruleset. It defines the trigger,
  single coordinator and ledger, local source-of-truth boundary, stable IDs,
  dependency and ready-frontier model, separate implementation/delivery/
  deployment/validation states, exact evidence, supervisor authority,
  checkpoint cadence, live reconciliation, handoff, and completion gates.
- Added one compact activation gate to generated V1, V2, and V3 coding-agent
  instructions and aligned checked-in V3 AGENTS, Claude, and Copilot files.
- Added RLM, dispatch-tooling, references-index, downstream-adoption, generated
  template, ruleset validation, and stale-reconciliation coverage.
- Preserved existing command ownership. This feature does not add a `program`
  command family, parser, state schema, deployment driver, or live
  cross-repository reconciler.
- Remaining delivery dependency: downstream registry clients cannot install
  this new rule from Kit's `main` branch until this pull request is merged.
- Future implementation may add a first-class program artifact by extending
  existing scaffold, map, state, check, resume, handoff, and dispatch verbs;
  that work requires its own accepted design and is not implied by this rule.

## REPOSITORY MEMORY

- Created this V3 spec because the coordination trigger, source-of-truth
  boundary, state model, and rejected command-family design contain material
  rationale that code and tests alone cannot preserve.
- Created
  `docs/references/rules/cross-repository-program-coordination.md` as the
  reusable downstream behavioral contract.
- Updated `docs/CONSTITUTION.md` with the demonstrated project-wide invariant:
  qualifying multi-repository programs use one coordinator ledger and advance
  only from reconciled evidence and a current ready frontier.
- Updated generated and checked-in coding-agent routing, RLM, tooling, and the
  references index so agents can discover the rule without copying it.
- Updated `docs/PROJECT_PROGRESS_SUMMARY.md` through Kit's feature lifecycle
  rollup so the feature remains discoverable from repository memory.
