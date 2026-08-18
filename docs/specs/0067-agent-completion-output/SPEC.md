---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: 0067
  slug: agent-completion-output
  dir: 0067-agent-completion-output
relationships:
  - type: builds_on
    target: 0020-versioned-instruction-model
  - type: builds_on
    target: 0042-native-plan-repository-memory
  - type: related_to
    target: 0053-testing-and-environment-validation
  - type: related_to
    target: 0063-explicit-work-lane-choice
  - type: related_to
    target: 0066-capability-aware-subagent-workflows
references:
  - id: instruction-templates
    name: Managed agent instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: shared completion routing across current provider instructions
    status: active
  - id: context-workflows
    name: Managed task context workflows
    type: code
    target: internal/templates/context_workflows/implementation-delivery.md
    relation: implements
    read_policy: must
    used_for: mandatory task-completion evidence selection
    status: active
  - id: registry-adoption
    name: Downstream ruleset adoption coverage
    type: code
    target: pkg/cli/init_refresh_ruleset_adoption_test.go
    relation: supports
    read_policy: must
    used_for: refresh, health, and reconcile propagation
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Give every Kit-managed coding-agent task a terminal response that an operator
can scan and act on immediately. Completion starts with one literal overall
status, exposes blockers and unfinished work before detail, and uses a
task-specific evidence block without sacrificing existing validation,
delivery, orchestration, or repository-memory requirements.

## CONTEXT

- Current guidance already says to lead with the outcome, report validation
  literally, include blockers and the smallest next action, preserve
  repository-memory disposition, and report GitHub delivery identifiers.
  Those requirements are scattered across root instructions and several
  rules, so agents can satisfy them in inconsistent orders and shapes.
- The accepted implementation format is the action-first option: a literal
  status heading and an immediate operator-action table, followed by
  left-aligned evidence, delivery, and memory blocks when applicable.
- The same status and action envelope must support research, diagnosis,
  planning, validation, review, operations, coordination, and unclassified
  tasks. The requested deliverable, not incidental activity, selects one
  primary profile.
- Kit distributes universal policy through downstream rulesets, generated and
  checked-in instructions, context workflows, refresh/reconcile behavior, and
  focused propagation tests.
- `kit instructions` versions are immutable. V5 remains fixed, so the new
  current contract must be additive V6.
- Higher-priority host wrappers and machine directives can constrain literal
  response boundaries. The status heading is therefore the first
  human-readable line inside any required wrapper.
- Issue #162, branch `GH-162`, and the canonical non-primary worktree are the
  authorized delivery lane. Merge is not authorized.

## REQUIREMENTS

### Universal Terminal Envelope

- Add an active downstream ruleset named `agent-completion-output` with
  `read_policy_default: must`.
- Apply it to terminal task completion and handoff responses, not progress
  commentary or focused clarification questions.
- Make the first human-readable line exactly
  `# <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>`.
- Put an action table immediately after that heading with columns `Type`,
  `Action required`, `Why`, and `Continue with`.
- Order action rows as Blocker, Incomplete, Next, then Optional. A completed
  task includes a None row that states no action is required.
- Required follow-up rows contain an exact copy-ready prompt or command. Cells
  are never blank and successful work must not hide blockers or incomplete
  scope.

### Status Semantics

- Use PASS only when required scope and required validation are complete or
  explicitly not applicable.
- Use PARTIAL when useful work exists but required scope or evidence remains
  incomplete without an external stopping dependency.
- Use BLOCKED when completion requires specific external input, authority, or
  state; name the exact unblock action and resume prompt.
- Use FAIL for an unresolved known failure when an external blocker is not the
  stopping reason.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and
  NOT_APPLICABLE rather than translating them into success.

### Required Primary Profiles

- Implementation and delivery: Item, Result, Evidence; add left-aligned
  Validation, Delivery, and Repository Memory blocks when applicable.
- Research and discovery: Question, Finding, Evidence and confidence,
  Implication.
- Diagnosis and troubleshooting: Symptom, Root cause or hypothesis, Evidence,
  Confidence and impact.
- Planning and design: Decision, Chosen approach, Rationale, Acceptance signal.
- Validation and testing: Check, Scope, Status, Evidence or gap.
- Review and audit: Severity, Finding, Location or evidence, Required action.
- Operations, deployment, and monitoring: Target, Action or observation,
  Status, Evidence or recovery.
- Coordination and handoff: Workstream, Owner, State, Dependency or next
  handoff.
- Fallback: Item, Result, Evidence or limitation.
- Select exactly one primary profile from the requested deliverable. Add
  supplemental structures only when another active contract requires their
  data.

### Readability And Composition

- Keep only the operator action section as a Markdown table. Render every
  task-specific profile with left-aligned CommonMark list or key/value blocks,
  short bold lead labels, and indented supporting evidence.
- Preserve all exact GitHub delivery, repository-memory, validation,
  orchestration-conformance, program-coordination, and environment-evidence
  fields required by adjacent rules, mapped into the canonical profile blocks.
- Higher-priority system, developer, client, and host schemas retain
  precedence. Preserve semantic order inside their required wrapper or use the
  closest structurally equivalent fields when CommonMark lists are prohibited.
- Do not change command behavior, add a CLI flag, or add a JSON schema.

### Observable Acceptance

- The rule parses as a valid mandatory downstream ruleset and is available to
  normal rules list/view and initialization paths.
- V1, V2, and V3 generated provider instructions plus checked-in current
  guidance route terminal reporting through the same rule.
- Every managed context workflow selects the rule as required evidence.
- Adjacent reporting rules retain their existing required information without
  defining a competing response order.
- The action table is the only normal Markdown table; every selected profile
  keeps a consistent left edge through headings and list or key/value blocks.
- Health and `reconcile --include-files` install the rule and managed guidance;
  semantic drift in the terminal envelope is detected.
- Instruction versions V1 through V5 remain byte-for-byte unchanged and V6 is
  registered as current.

## ACCEPTED PLAN

1. Create the canonical ruleset and record its cross-project rationale in this
   living spec, the references index, and the project Constitution.
2. Add one concise shared completion routing contract and compose it through
   all generated provider instructions, support guidance, and checked-in
   mirrors.
3. Require the rule in all embedded and checked-in context workflows and
   reconcile overlapping final-report clauses in delivery, validation,
   orchestration, program coordination, and repository-memory guidance.
4. Preserve V1 through V5 instructions, add additive V6, and make V6 current.
5. Add focused rule, template, workflow, registry, health, reconcile, drift,
   and version tests while preserving the source-file-size contract.
6. Run focused and complete validation, curate repository memory, explicitly
   stage only issue #162 changes, and deliver one ready pull request.

## DECISIONS

- Use one mandatory ruleset with profiles, not separate rules per task type.
  The universal envelope is the stable operator contract while profiles keep
  task evidence proportional.
- Make the implementation/delivery profile use the accepted action-first
  option. The action table remains present on PASS so the operator can
  distinguish completion from omitted follow-up reporting.
- Keep tables only where they materially improve actionability. Renderer
  evidence showed independently sized detail tables are centered separately,
  so left-aligned profile blocks provide a more readable shared edge.
- Select profiles by the requested outcome. An implementation task remains an
  implementation profile even when research, diagnosis, review, and testing
  occurred during execution.
- Treat the first line as the first human-readable line so host-mandated tags,
  directives, or structured wrappers can compose without weakening the
  operator-visible contract.
- Preserve literal external states instead of inventing one universal row
  status vocabulary. The four-state vocabulary governs only the overall task.
- Add V6 rather than mutating V5 or earlier immutable instructions.

## DISCOVERIES

- The conversation renderer centers each Markdown table according to its
  intrinsic width. Multiple detail tables therefore produce different left
  edges even when every cell is left-aligned; Markdown alignment markers do
  not control the table container. The stable solution is one action table
  followed by ordinary left-aligned content.
- V3 guardrails are derived from the shared guardrails template and then
  extended with repository-memory completion guidance. Composing the new gate
  in both layers created a duplicate recognized section; the final design
  composes it once in the shared guardrails source.
- The generated Copilot document ends directly with the shared completion
  contract. Its checked-in mirror requires exactly one trailing newline, so
  the Copilot generator trims only the gate's extra separator newline while
  other consumers retain the blank line before their next section.
- Fresh repository bootstrap resolves every required workflow rule immediately
  after initialization. The repository-bootstrap fixture therefore had to
  install `agent-completion-output` alongside its other required rules.
- A branch-local downstream rule is intentionally `untracked` by the live
  rules inventory until the source lands on the registry branch. Live health
  and include-files previews consequently compare edited managed rules against
  current `main`; focused registry stubs prove the post-landing propagation
  behavior without fabricating live registry state.
- CodeRabbit correctly identified two missing regression boundaries: literal
  `NOT_APPLICABLE` coverage in both canonical and V6 assertions, and semantic
  reconcile coverage for removing the completion rule from generated
  references indices. Both gaps were valid and are now covered for V2 and V3.

## VALIDATION

- `go test ./internal/templates ./internal/instructions ./internal/context ./pkg/cli -count=1` passed.
- `go test ./... -count=1` and `go test -race ./... -count=1` passed after
  the final source change.
- `make fmt`, `go vet ./...`,
  `golangci-lint run --new-from-rev=origin/main ./...`, `make build`, and
  `git diff --check` passed; lint reported zero issues.
- The branch-built `bin/kit` parsed and rendered
  `agent-completion-output`, listed every task profile, rendered V6, and
  resolved `implementation-delivery` with the new rule as required evidence
  and no blocked diagnostics.
- `bin/kit check 0067-agent-completion-output`, `bin/kit check --all`, and
  `bin/kit check --project` passed; all 64 feature contracts and the project
  contract were coherent.
- V1 through V5 SHA-256 hashes remained unchanged. V6 is current, preserves V5
  as an exact prefix, and has SHA-256
  `77ea2d59321f411a160cb8ef18b4b821e0e08ccd2170ec7f047825cdf5ba93a0`.
- Focused propagation coverage passed for both `kit health` and
  `kit reconcile --include-files`. The live read-only previews correctly
  reported branch-versus-main registry drift; reconcile's semantic audit was
  clean and its source-size audit checked 703 version-control-eligible
  candidates and 362 handwritten source/test files with zero above 300
  physical lines.
- Focused hybrid-format tests prove every managed provider contract requires
  left-aligned detail blocks, canonical rule tests reject the former profile
  table headers, and stale references-index routing is detected for V2 and V3.
- `gitleaks dir --no-banner --redact .` scanned 5.24 MB and found no leaks.

## OUTCOME

- Added the mandatory downstream `agent-completion-output` ruleset with a
  four-state terminal heading, one immediate operator action table, eight
  requested left-aligned goal-specific profiles, and one fallback profile.
- Routed the contract through V1, V2, and V3 provider instructions, current
  checked-in guidance, guardrails, RLM, the references index, the Constitution,
  all seven context workflows, registry adoption, health/reconcile propagation,
  and semantic drift detection.
- Consolidated GitHub delivery, validation, repository-memory, orchestration,
  and cross-repository program reporting into the canonical profiles without
  removing their required evidence.
- Retained the operator action queue as the single normal Markdown table and
  converted every task-specific detail profile to left-aligned CommonMark
  lists or key/value blocks so independently sized tables cannot drift across
  the response.
- Preserved immutable instruction versions V1 through V5 and made additive V6
  current. No command, flag, or JSON schema was added.
- Delivery remains at the authorized issue #162 branch and ready-pull-request
  boundary. Merge is not authorized.

## REPOSITORY MEMORY

Decision: created

Rationale: The cross-project status vocabulary, task-profile selection,
action-first ordering, host-wrapper composition, and instruction-version
strategy are consequential policy decisions that code assertions alone cannot
preserve.

Artifacts:

- `docs/specs/0067-agent-completion-output/SPEC.md`
- `docs/references/rules/agent-completion-output.md`
- `docs/CONSTITUTION.md`
- managed instruction, workflow, registry, reconciliation, and versioned
  instruction sources and tests
