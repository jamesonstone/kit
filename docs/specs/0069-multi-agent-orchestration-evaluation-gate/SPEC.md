---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0069"
  slug: "multi-agent-orchestration-evaluation-gate"
  dir: "0069-multi-agent-orchestration-evaluation-gate"
relationships:
  - type: builds_on
    target: 0066-capability-aware-subagent-workflows
  - type: builds_on
    target: 0063-explicit-work-lane-choice
  - type: related_to
    target: 0068-human-authorship
references:
  - id: agent-team-orchestration
    name: Capability-aware supervisor/child-agent contract
    type: rule
    target: docs/references/rules/agent-team-orchestration.md
    relation: constrains
    read_policy: must
    used_for: existing lifecycle, capability profiles, and single-lane criteria this gate makes mandatory to evaluate
    status: active
  - id: work-lane-gating
    name: Work Lane Mutation Hard Gate precedent
    type: rule
    target: docs/references/rules/work-lane-gating.md
    relation: supports
    read_policy: must
    used_for: structural precedent for adding a new cross-file Hard Gate
    status: active
  - id: implementation-delivery-workflow
    name: Implementation delivery workflow contract
    type: documentation
    target: docs/references/workflows/implementation-delivery.md
    relation: constrains
    read_policy: must
    used_for: required-rule wiring for agent-team-orchestration
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make every coding agent (Claude Code, Codex, GitHub Copilot, Warp/Oz) always
evaluate multi-agent/parallel decomposition during native plan formation,
not just for work that already looks obviously large. The evaluation is
mandatory; the recorded answer is still allowed to be single-lane.

## CONTEXT

- Prompted by a user request to research multi-agent orchestration best
  practices (a Reddit-linked tool and Anthropic's own published
  orchestrator-worker research-system guidance both point at effort-scaled
  routing, not blanket parallelism) and implement equivalent behavior into
  Kit's own coding-agent workflow.
- Kit already ships a mature orchestrator/worker rule,
  `docs/references/rules/agent-team-orchestration.md` (supervisor/child-agent
  lifecycle, semantic capability profiles: `architect`, `orchestrator`,
  `mapper`, `specialist`, `precision`, `verifier`), introduced by
  `0066-capability-aware-subagent-workflows`. It was wired conditionally: the
  exact pre-existing RLM.md pointer read "only when the immediate decision
  includes execution topology, subagent lanes, or read-only verification; do
  not load it for trivial single-lane tasks" -- which means an agent handling
  a trivial task never opened the rule and therefore never learned the
  escape valve existed to record "single-lane, because trivial." That is a
  real self-contradiction, not just weak wording.
- `0063-explicit-work-lane-choice` is the closest precedent for adding a
  genuine new cross-file Hard Gate (its origin commit `ae9476a` touched
  CLAUDE.md/AGENTS.md/copilot-instructions.md stubs, GUARDRAILS.md
  procedural text, an immutable instruction version bump, and Go template
  wiring across the same call sites this feature needed).
- `docs/CONSTITUTION.md` carried a NON-GOAL, "Kit does not choose models,
  launch coding agents, supervise agent processes, or replace native agent
  planning," that read close to contradicting a mandatory-evaluation gate
  even though the gate never selects a concrete model or forces parallel
  execution.
- `agent-team-orchestration.md` is a `registry_scope: downstream` ruleset
  fetched live from `jamesonstone/kit`'s GitHub `main` branch on every
  untargeted `kit init`/`init refresh`/`kit reconcile` (see
  `pkg/cli/rules_registry_fetch.go`, `rules_registry.go`,
  `init_refresh_rulesets.go`). The Hard Gate skeleton (Go template text
  compiled into CLAUDE.md/AGENTS.md/copilot-instructions.md/GUARDRAILS.md) is
  compiled into the `kit` binary instead. These propagate to downstream
  projects on different schedules: the strengthened rule body reaches a
  project on its next untargeted refresh once this change is on `main`; the
  gate skeleton only reaches a project after it upgrades to a release
  containing this change and re-runs init/refresh.
- `pkg/cli/agent_team_orchestration_rules_test.go`'s
  `TestAgentTeamOrchestrationFallbackOrderAndNeutrality` already forbids
  concrete provider/model names (including `luna` and `terra`) inside
  `agent-team-orchestration.md`, confirming the existing project convention
  of abstracted capability tiers rather than pinned model names.
- Issue #168, branch `GH-168`, and the canonical non-primary worktree are
  the authorized delivery lane. Merge is not authorized.

## REQUIREMENTS

### Mandatory First-Pass Topology Evaluation

- Before finalizing any native implementation plan for a new feature, a
  substantial architectural or behavioral change, or a multi-file refactor,
  a coding agent must load `agent-team-orchestration.md` and evaluate
  multi-agent/parallel decomposition using that rule's existing lifecycle
  and capability profiles.
- The evaluation itself is never skipped for that class of work. The
  recorded decision may still be single-lane, using the rule's existing
  single-lane criteria (trivial, tightly coupled, high-overlap, requires
  continuous design judgment, user requested single-agent execution, host
  does not confirm separate execution).
- A single mechanical edit, a direct question, or read-only research that
  never forms an implementation plan does not trigger the gate.

### New Cross-File Hard Gate

- Add a "Multi-Agent Orchestration Evaluation Hard Gate" to all three
  Kit-managed instruction files (`CLAUDE.md`, `AGENTS.md`,
  `.github/copilot-instructions.md`) and `docs/agents/GUARDRAILS.md`,
  positioned immediately before the Work Lane Mutation Hard Gate in each
  file so on-page order mirrors temporal order (evaluation happens during
  plan formation; mutation gate fires later, before the first repository
  write).
- Source the gate from Go template constants
  (`multiAgentOrchestrationRoutingGate` for the short instruction-file form,
  `multiAgentOrchestrationHardGate` for the full GUARDRAILS.md form) so it
  propagates to every downstream Kit-managed project through the normal
  `kit init`/`init refresh` template path, mirroring
  `workLaneMutationRoutingGate`/`workLaneMutationHardGate`.
- Broaden the RLM.md pointer for `agent-team-orchestration.md` from
  "only when execution topology is already in question" to a mandatory
  first-pass check. Leave the TOOLING.md dispatch-time pointer conditional;
  it governs a different, later moment (`kit dispatch` execution planning,
  not plan formation).
- Mark `agent-team-orchestration` `required: true` in
  `docs/references/workflows/implementation-delivery.md` and its template
  mirror only. Leave the other four workflows' entries unchanged.
- Narrow the CONSTITUTION.md NON-GOALS bullet so it no longer reads as
  contradicting the new gate, while preserving that Kit still never selects
  a concrete model or forces parallel execution.
- Publish a new immutable instruction version, `v7`, matching the
  `0063-explicit-work-lane-choice` precedent for this class of change.

### Non-Goals

- No forced parallelism. A recorded single-lane decision remains a fully
  valid, first-class outcome.
- No hardcoded model names (no `Fable 5`, no `Sol`/`Terra`/`Luna`, no
  `Sonnet`/`Haiku`) anywhere in the new gate text or the strengthened rule.
  Use Kit's existing abstracted capability-profile vocabulary instead.
- No new competing topology vocabulary (no "light/medium/heavy" tiers).
  Reuse `agent-team-orchestration.md`'s existing lifecycle and profile
  terms.
- No change to `pkg/cli/subagents.go` / `kit dispatch`'s existing prompt
  suffix -- it is a distinct, later, post-plan mechanism already governed by
  the same rule file.
- No change to the ruleset's `read_policy_default` (`conditional`) or
  `applies_to` array -- a pinned test depends on the current values, and
  mandatoriness is enforced through the gate plus the workflow's
  `required: true` entry instead.
- No Codex-specific "Sol subagents inherit full config" caveat -- considered
  and explicitly excluded by the user.

### Observable Acceptance

- The new Hard Gate section appears in `CLAUDE.md`, `AGENTS.md`,
  `.github/copilot-instructions.md`, and `docs/agents/GUARDRAILS.md`,
  generated from Go template source, immediately before Work Lane Mutation
  Hard Gate in each.
- `docs/agents/RLM.md`'s `agent-team-orchestration.md` pointer reads as a
  mandatory first-pass evaluation; `docs/agents/TOOLING.md`'s dispatch-time
  pointer is unchanged.
- `agent-team-orchestration.md` carries a new "First-Pass Topology Decision"
  rule and a broadened "Applies When" section, with `read_policy_default`
  and `applies_to` unchanged.
- `implementation-delivery.md` (both copies) mark
  `agent-team-orchestration` `required: true`.
- `docs/CONSTITUTION.md`'s NON-GOALS bullet is narrowed and internally
  consistent with the new gate.
- Immutable instruction version `v7` exists with a pinned SHA-256 and is
  `CurrentAgentVersion`.
- Focused tests cover the new gate text, its ordering relative to Work Lane
  Mutation Hard Gate, the broadened RLM pointer, and the workflow's
  `required: true` entry.

## ACCEPTED PLAN

1. Add `internal/templates/instruction_templates_agent_team_gate.go` with
   the short and full gate constants.
2. Wire the constants into the six existing
   `workLaneMutationRoutingGate`/`workLaneMutationHardGate` call sites across
   `instruction_templates_v2.go`, `instruction_templates_v3.go`,
   `instruction_support_templates.go`, and `instruction_templates_shared.go`.
3. Broaden the `agentsRLM` pointer sentence in `instruction_templates_v2.go`
   (shared by both the TOC and V3 scaffolds via `memoryRLM()`).
4. Strengthen `docs/references/rules/agent-team-orchestration.md`: broaden
   "Applies When," add a "First-Pass Topology Decision" rule.
5. Flip `agent-team-orchestration` to `required: true` in
   `docs/references/workflows/implementation-delivery.md` and
   `internal/templates/context_workflows/implementation-delivery.md`.
6. Narrow the CONSTITUTION.md NON-GOALS bullet.
7. Publish `internal/instructions/versions/v7.md` (ported from `v6.md` with
   the new gate's procedural text inserted before the repository
   context/mutation gate section), bump `CurrentAgentVersion`, add the v7
   entry and SHA-256 pin.
8. Regenerate `CLAUDE.md`, `AGENTS.md`, `.github/copilot-instructions.md`,
   `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md` via
   `kit init --refresh --force --file <path>` (scoped to avoid touching the
   network-fetched ruleset registry) rather than hand-editing, to guarantee
   byte-identity with the template output.
9. Extend `pkg/cli/agent_team_orchestration_rules_test.go`,
   `internal/templates/templates_test.go`, and
   `internal/instructions/agent_versions_test.go`; add a new focused test
   file for the gate's cross-file presence and ordering.
10. Record this spec and the progress summary, then deliver issue #168
    through one ready pull request.

## DECISIONS

- Position the new gate immediately before Work Lane Mutation Hard Gate
  rather than after it, so on-page order matches the temporal order
  established in `docs/agents/WORKFLOWS.md`'s Agent-First Contract
  (Coding Agent Context Gate -> this gate -> Work Lane Mutation Hard Gate).
- Make the gate an agent-owned recorded evaluation, not a user-facing
  verbatim question like Work Lane Mutation's lane choice. The topology
  decision is the coding agent's own judgment call against
  `agent-team-orchestration.md`'s criteria, not something only the human can
  authorize.
- Leave `docs/agents/TOOLING.md`'s dispatch-time pointer and
  `pkg/cli/subagents.go`'s existing prompt suffix untouched. Both govern a
  distinct, later moment (post-plan execution topology), and broadening them
  to match RLM.md's plan-time wording would blur two genuinely different
  decision points.
- Regenerate the on-disk `CLAUDE.md`/`AGENTS.md`/copilot-instructions.md/
  GUARDRAILS.md/RLM.md via `kit init --refresh --force --file <path>`
  (file-scoped) instead of `--refresh` unscoped or hand-editing. Unscoped
  refresh also touches the network-fetched ruleset registry, which would
  have re-fetched the pre-edit `agent-team-orchestration.md` from
  `jamesonstone/kit`'s GitHub `main` and could have overwritten the local
  strengthening made in this same change before it ever reached `main`.
  File-scoping avoided that entirely; a `--dry-run --diff` pass confirmed
  the planned diff matched intent exactly (5 files updated, TOOLING.md
  correctly skipped) before applying it for real.
- Publish immutable instruction version `v7` rather than leaving `v6`
  current. `0063-explicit-work-lane-choice`'s origin commit did the same for
  a comparable cross-file Hard Gate addition; a lighter, purely
  pointer-loaded rule addition like `0068-human-authorship` did not need a
  version bump, but this feature changes always-loaded gate text, matching
  the heavier precedent instead.
- Keep `read_policy_default: conditional` and `applies_to` unchanged on the
  ruleset itself. A pinned test
  (`TestAgentTeamOrchestrationRegistryRulesetIsCapabilityAware`) depends on
  the current values, and mandatoriness for this feature is enforced through
  the new gate plus the workflow's `required: true` entry, not through the
  ruleset's own default read policy.

## DISCOVERIES

- `docs/references/rules/*.md` ruleset bodies and the Go-template-compiled
  instruction/GUARDRAILS text propagate to downstream projects through two
  independent mechanisms on two independent schedules (see CONTEXT). This
  was not obvious from a static read of `internal/templates/` alone; it
  required tracing `pkg/cli/rules_registry_fetch.go`,
  `rules_registry.go`, and `init_refresh_rulesets.go`.
- `memoryRLM()` and `memoryTooling()` (in
  `internal/templates/instruction_templates_v3_context.go`) derive the V3
  scaffold's RLM.md/TOOLING.md output from the same `agentsRLM`/
  `agentsTooling` constants used by the TOC scaffold, via targeted
  `strings.Replace`/`strings.ReplaceAll` substitutions on unrelated anchor
  text. One edit to `agentsRLM`'s pointer sentence therefore updates both
  scaffold versions' RLM.md output without a separate V3-specific edit.
  Similarly, `memoryGuardrails()` derives V3's GUARDRAILS.md from the shared
  `agentsGuardrails` constant, so one edit to `agentsGuardrails` updates both
  scaffold versions' GUARDRAILS.md output.
- `docs/references/workflows/implementation-delivery.md` and
  `internal/templates/context_workflows/implementation-delivery.md` were
  already byte-identical before this change, confirming Kit dogfoods its own
  template output for workflow contracts exactly as it does for instruction
  files.
- `TestAgentTeamOrchestrationFallbackOrderAndNeutrality` already forbids
  `luna` and `terra` (among other provider/model terms) inside
  `agent-team-orchestration.md`. This independently corroborated the
  decision to keep the strengthened rule text fully abstracted before any
  code was written.

## VALIDATION

- `kit init --refresh --dry-run --diff --force --file CLAUDE.md --file AGENTS.md --file .github/copilot-instructions.md --file docs/agents/GUARDRAILS.md --file docs/agents/RLM.md --file docs/agents/TOOLING.md` confirmed the exact planned diff (5 updated, TOOLING.md skipped) before applying it for real with the same command minus `--dry-run`.
- `go test ./...` passed across every package, including the new/extended coverage in `internal/instructions` (`agent_versions_v7_test.go`, updated `agent_versions_v6_test.go`/`agent_versions_test.go`), `internal/templates`, and `pkg/cli` (new `multi_agent_orchestration_gate_test.go`, extended `agent_team_orchestration_rules_test.go`, fixed `instructions_test.go` v7/v8 pins).
- `make fmt`, `go vet ./...`, and `make build` passed with no changes/output beyond this feature's own files.
- `golangci-lint run --new-from-rev=origin/main ./...` reported 0 issues.
- `git diff --check` reported no whitespace errors.
- `kit check 0069-multi-agent-orchestration-evaluation-gate` and `kit check --project` passed after this OUTCOME section was populated (both initially failed on the placeholder OUTCOME, as expected).
- `kit reconcile --all --dry-run --diff --include-files` was run to preview drift. It correctly flagged `agent-team-orchestration`'s registry state as `local-custom` (expected: this feature intentionally strengthens the local rule body ahead of the still-unmodified upstream registry copy on `main`) and separately surfaced pre-existing, unrelated upstream drift for `github-pr-delivery`, `safety-guardrails`, and a missing `human-authorship` registry entry. That unrelated drift predates this change and was left untouched rather than folded into this feature's diff. Nothing was written since the pass was `--dry-run`.
- `gitleaks detect --source . --no-git` found no leaks.

## OUTCOME

- Added `internal/templates/instruction_templates_agent_team_gate.go` with the short (`multiAgentOrchestrationRoutingGate`) and full (`multiAgentOrchestrationHardGate`) gate constants, wired into all six existing `workLaneMutation*` call sites so the new Hard Gate renders immediately before Work Lane Mutation Hard Gate in `CLAUDE.md`, `AGENTS.md`, `.github/copilot-instructions.md`, and `docs/agents/GUARDRAILS.md` across both the TOC and V3 instruction scaffolds.
- Broadened the `agent-team-orchestration.md` pointer sentence in `docs/agents/RLM.md` to a mandatory first-pass evaluation; left `docs/agents/TOOLING.md`'s dispatch-time pointer unchanged as planned.
- Strengthened `docs/references/rules/agent-team-orchestration.md`: broadened "Applies When" and added a "First-Pass Topology Decision" rule, without touching `read_policy_default` or `applies_to`.
- Marked `agent-team-orchestration` `required: true` in `docs/references/workflows/implementation-delivery.md` and its template mirror only.
- Narrowed the `docs/CONSTITUTION.md` NON-GOALS bullet to reconcile it with the new mandatory-evaluation gate.
- Published immutable instruction version `v7` (`internal/instructions/versions/v7.md`, `agent_versions.go`, pinned SHA-256 in `agent_versions_test.go`), and added dedicated `agent_versions_v7_test.go` coverage (content + diff-preserving-vs-v6 tests) mirroring the existing per-version test convention; removed the now-stale `CurrentAgentVersion == "v6"` assertion from `agent_versions_v6_test.go`; fixed `pkg/cli/instructions_test.go`'s unavailable-version fixture from v7 to v8.
- Regenerated `CLAUDE.md`, `AGENTS.md`, `.github/copilot-instructions.md`, `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md` via file-scoped `kit init --refresh --force`, confirmed byte-identical to template output; `docs/agents/TOOLING.md` correctly unchanged.
- Manually reverted and re-applied a minimal, correct edit to `docs/PROJECT_PROGRESS_SUMMARY.md` after discovering `kit spec`'s table regeneration sources its `CREATED` column from each `SPEC.md`'s filesystem mtime (`internal/feature/feature.go:139`), which a fresh `git worktree add` checkout resets for every historical file to today's date. Reverted the accidental 65-row mass date corruption and hand-added only this feature's one summary row and one per-feature detail subsection, matching the existing format exactly.
- No changes to `pkg/cli/subagents.go`, `kit dispatch`'s prompt suffix, the ruleset's `read_policy_default`/`applies_to`, or any workflow other than `implementation-delivery`, matching the accepted plan's non-goals.
- Delivery remains at the authorized issue #168 branch and ready-pull-request boundary. Merge is not authorized.

## REPOSITORY MEMORY

- **Decision:** created
- **Rationale:** This feature changes a project-wide invariant (the
  CONSTITUTION.md NON-GOALS bullet), adds a new always-loaded Hard Gate
  propagated to every downstream Kit-managed project, and depends on
  non-obvious propagation-mechanism findings that future maintainers need
  without re-deriving them from source.
- **Artifacts:** `docs/specs/0069-multi-agent-orchestration-evaluation-gate/SPEC.md`,
  `docs/references/rules/agent-team-orchestration.md`,
  `internal/templates/instruction_templates_agent_team_gate.go`
