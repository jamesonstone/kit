---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: deliver
clarification:
  status: ready
  confidence: 96
  unresolved_questions: 0
feature:
  id: 0039
  slug: reflect-scoring
  dir: 0039-reflect-scoring
relationships:
  - type: related_to
    target: 0031-executable-verification-harness
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: user-thesis
    name: User thesis attachment
    type: attachment
    target: /Users/jamesonstone/.codex/attachments/eea16278-af47-46c0-bfe4-e62b89474e1b/pasted-text.txt
    selector_type: artifact
    selector: pasted-text.txt
    relation: constrains
    read_policy: must
    used_for: feature thesis, non-goals, acceptance criteria, and stop-after-plan gate
    status: active
  - id: agents-routing
    name: Agent routing docs
    type: doc
    target: docs/agents/README.md
    selector_type: artifact
    selector: README.md
    relation: guides
    read_policy: must
    used_for: spec-driven routing and current feature artifact selection
    status: active
  - id: workflow-rules
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: v2 single-SPEC source-of-truth rules and readiness gates
    status: active
  - id: guardrails
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    selector_type: artifact
    selector: GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: dirty-worktree gate, validation expectations, and GitHub delivery hard gate
    status: active
  - id: rlm
    name: RLM routing
    type: doc
    target: docs/agents/RLM.md
    selector_type: artifact
    selector: RLM.md
    relation: guides
    read_policy: must
    used_for: narrow prior-work and codebase-discovery pass
    status: active
  - id: tooling
    name: Tooling docs
    type: doc
    target: docs/agents/TOOLING.md
    selector_type: artifact
    selector: TOOLING.md
    relation: constrains
    read_policy: must
    used_for: command-capability and agent-team routing requirements
    status: active
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: document-first state, explicit execution boundaries, and v2 workflow invariants
    status: active
  - id: project-progress
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    selector_type: heading
    selector: FEATURE PROGRESS TABLE
    relation: informs
    read_policy: evidence
    used_for: current feature index and prior-feature shortlist
    status: active
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0039-reflect-scoring
    selector_type: artifact
    selector: README.md
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input; only README and placeholders currently exist
    status: optional
  - id: command-capabilities-rule
    name: Command capabilities rule
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    selector_type: artifact
    selector: command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: required capability metadata updates when reflect command behavior changes
    status: active
  - id: agent-team-rule
    name: Agent team orchestration rule
    type: ruleset
    target: docs/references/rules/agent-team-orchestration.md
    selector_type: artifact
    selector: agent-team-orchestration.md
    relation: constrains
    read_policy: must
    used_for: implementation and verification lane plan after approval
    status: active
  - id: reflect-command
    name: Current reflect command
    type: code
    target: pkg/cli/reflect.go
    selector_type: artifact
    selector: reflect.go
    relation: informs
    read_policy: must
    used_for: legacy reflect boundary; not the implementation target after user clarification
    status: active
  - id: legacy-command
    name: Legacy command root
    type: code
    target: pkg/cli/legacy.go
    selector_type: symbol
    selector: legacyCmd
    relation: constrains
    read_policy: must
    used_for: command-surface reality that reflect is currently under kit legacy
    status: active
  - id: feature-phase-schema
    name: Feature phase schema
    type: code
    target: internal/feature/feature.go
    selector_type: artifact
    selector: feature.go
    relation: constrains
    read_policy: must
    used_for: phase-state and reflection marker conventions
    status: active
  - id: front-matter-schema
    name: Front matter schema
    type: code
    target: internal/document/metadata.go
    selector_type: artifact
    selector: metadata.go
    relation: constrains
    read_policy: must
    used_for: BRAINSTORM/SPEC/PLAN/TASKS front matter conventions
    status: active
  - id: artifact-templates
    name: Artifact templates
    type: code
    target: internal/templates/templates.go
    selector_type: artifact
    selector: templates.go
    relation: constrains
    read_policy: must
    used_for: v2 SPEC and legacy TASKS artifact defaults
    status: active
  - id: verify-engine
    name: Verification engine
    type: code
    target: internal/verify/tasks.go
    selector_type: artifact
    selector: tasks.go
    relation: supports
    read_policy: must
    used_for: existing task scope, expected files, and verify command parsing
    status: active
  - id: verify-execute
    name: Verification execution
    type: code
    target: internal/verify/execute.go
    selector_type: artifact
    selector: execute.go
    relation: supports
    read_policy: must
    used_for: existing command exit-code, stdout/stderr, status, and timestamp handling
    status: active
  - id: runstore
    name: Runstore evidence
    type: code
    target: internal/runstore/store.go
    selector_type: artifact
    selector: store.go
    relation: supports
    read_policy: conditional
    used_for: latest legacy verification evidence lookup
    status: active
  - id: quality-gates
    name: Makefile quality commands
    type: file
    target: Makefile
    selector_type: artifact
    selector: Makefile
    relation: constrains
    read_policy: must
    used_for: existing build, test, lint, vet, and all command sources
    status: active
  - id: precommit-gate
    name: Pre-commit gate
    type: file
    target: .githooks/pre-commit
    selector_type: artifact
    selector: pre-commit
    relation: informs
    read_policy: evidence
    used_for: actual pre-commit command source
    status: active
  - id: release-quality-gates
    name: Release quality gates
    type: file
    target: .github/workflows/release-publish.yml
    selector_type: artifact
    selector: release-publish.yml
    relation: informs
    read_policy: evidence
    used_for: CI release gate command source
    status: active
  - id: capabilities-catalog
    name: Capabilities catalog
    type: code
    target: pkg/cli/capabilities_catalog.go
    selector_type: symbol
    selector: capabilityCatalog
    relation: constrains
    read_policy: must
    used_for: loop workflow capability metadata that must change when reflection verdict writes are added
    status: active
  - id: loop-runner
    name: Loop workflow runner
    type: code
    target: pkg/cli/loop_runner.go
    selector_type: artifact
    selector: loop_runner.go
    relation: constrains
    read_policy: must
    used_for: mechanical v2 stage execution and the reflect-stage hook point
    status: active
  - id: loop-validator
    name: Loop stage validator
    type: code
    target: pkg/cli/loop_validate.go
    selector_type: artifact
    selector: loop_validate.go
    relation: constrains
    read_policy: must
    used_for: strict phase advancement checks and the missing reflect-verdict gate
    status: active
  - id: loop-prompt
    name: Loop prompt builder
    type: code
    target: pkg/cli/loop_prompt.go
    selector_type: artifact
    selector: loop_prompt.go
    relation: constrains
    read_policy: must
    used_for: v2 stage prompt contract for reflection
    status: active
  - id: spec-v2-prompt
    name: V2 supervisor prompt
    type: code
    target: pkg/cli/spec_v2_prompt.go
    selector_type: artifact
    selector: spec_v2_prompt.go
    relation: constrains
    read_policy: must
    used_for: reflection-phase instructions in the kit spec workflow
    status: active
  - id: executable-verification-harness
    name: Executable verification harness spec
    type: feature
    target: docs/specs/0031-executable-verification-harness/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: prior task bundle, run evidence, and reflect evidence-gating contract
    status: active
  - id: kit-capabilities
    name: Kit capabilities spec
    type: feature
    target: docs/specs/0033-kit-capabilities/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: capability drift expectations for changed command behavior
    status: active
  - id: self-improvement-loops-skill
    name: Self-improvement loops source material
    type: external
    target: https://github.com/muratcankoylan/Agent-Skills-for-Context-Engineering/blob/main/skills/self-improvement-loops/SKILL.md
    selector_type: url
    selector: https://github.com/muratcankoylan/Agent-Skills-for-Context-Engineering/blob/main/skills/self-improvement-loops/SKILL.md
    relation: informs
    read_policy: must
    used_for: self-improvement-aligned scoring principles, especially runtime-owned metrics, raw evidence binding, and lowest-rung workflow scope
    status: active
delivery_intent: issue_branch_pr_later
---
# SPEC

## THESIS

Add a structured reflect verdict for Kit so the v2 `kit spec` reflection phase emits machine-checkable JSON next to the feature's `SPEC.md`.

The requested verdict schema is:

```go
type ReflectVerdict struct {
    TestsPass     bool   `json:"tests_pass"`
    LintDelta     int    `json:"lint_delta"`
    ScopeDrift    string `json:"scope_drift"`
    CycleTimeMin  int    `json:"cycle_time_min"`
    ReworkCount   int    `json:"rework_count"`
    PromptVersion string `json:"prompt_version"`
    Timestamp     string `json:"timestamp"`
}
```

This is step 1 of a larger harness-improvement plan. It must not touch dispatch, prompt curation, prompt stats, metrics aggregation, or self-editing agent logic.

The initial BRAINSTORM -> SPEC -> PLAN gate is complete. The user approved the recommended defaults on 2026-07-08, so this SPEC now authorizes the v2 task checklist and implementation work, but not issue, branch, commit, push, or PR mutation.

## CONTEXT

### Pre-Instruction Report

- Current SPEC path: `docs/specs/0039-reflect-scoring/SPEC.md`.
- Workflow version: `2`.
- Current phase: `deliver`.
- Clarification state: `status=ready`, `confidence=96`, `unresolved_questions=0`.
- Loaded instruction docs: `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md`, `docs/agents/TOOLING.md`.
- Loaded rule references: `docs/references/rules/command-capabilities.md`, `docs/references/rules/agent-team-orchestration.md`.
- Loaded external source material: self-improvement-loops `SKILL.md` from `muratcankoylan/Agent-Skills-for-Context-Engineering`.
- Loaded durable facts: `docs/CONSTITUTION.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`, `docs/notes/0039-reflect-scoring/README.md`, `kit map 0039-reflect-scoring`.
- Loaded prior feature context: `docs/specs/0031-executable-verification-harness/SPEC.md`, `docs/specs/0033-kit-capabilities/SPEC.md`.
- Loaded code and tooling context: `pkg/cli/loop_runner.go`, `pkg/cli/loop_validate.go`, `pkg/cli/loop_prompt.go`, `pkg/cli/spec_v2_prompt.go`, `pkg/cli/reflect.go`, `pkg/cli/legacy.go`, `pkg/cli/capabilities_catalog.go`, `internal/feature/feature.go`, `internal/feature/status.go`, `internal/document/metadata.go`, `internal/templates/templates.go`, `internal/verify`, `Makefile`, `.githooks/pre-commit`, release workflows.
- Current dirty-worktree classification: in-scope generated setup exists before implementation (`docs/specs/0039-reflect-scoring/SPEC.md`, `docs/notes/0039-reflect-scoring/**`, and `docs/PROJECT_PROGRESS_SUMMARY.md`). These are treated as user/Kit-owned setup changes; only this SPEC was edited in this planning pass.
- Readiness gate: passed after the user approved the recommended defaults.
- Agent team plan: use a single supervisor implementation lane for this feature because the change is tightly coupled around one runtime hook, one scorer, prompt/capability metadata, and focused tests. No separate specialist or verification subagents are spawned; validation remains the supervisor's responsibility.

### Brainstorm

The useful core is narrower than the thesis first appears:

- V2 reflection behavior is prompt-oriented and `SPEC.md` phase-oriented.
- The old `kit legacy reflect` command is not the implementation target for this feature.
- `kit loop workflow` is the mechanical v2 execution surface that can enforce the verdict before the workflow advances out of `reflect`.
- Self-improvement source material reinforces that this scoring step should be runtime-owned and evidence-bound, not a prompt-only instruction or a self-reported agent summary.
- Existing verification behavior already has task bundles, command exit codes, run status, expected files, and bounded run artifacts.
- The verdict should be additive: `SPEC.md` reflection notes and existing prompt text stay human-readable while `REFLECT.json` supplies the binary signal.
- The implementation should reuse existing command execution and run-evidence patterns where possible instead of adding an unrelated runner.
- The score must be generated from raw command, git, and diff evidence controlled by Kit runtime surfaces. Agent-written reports may inform human reflection but must not be trusted as metric input.
- This feature should stay at the workflow/evidence rung: add a measurable reflect signal without adding prompt evolution, harness self-edits, candidate search, or promotion logic.
- The riskiest remaining decisions are lint delta semantics and how to identify the v2 approval/start boundary for cycle time.

Options considered:

1. Add verdict generation to the v2 loop reflect stage.
   - Pros: matches the clarified product surface, gives Kit a deterministic gate before leaving `reflect`, and keeps legacy staged commands out of scope.
   - Cons: direct prompt-only users still need prompt instructions unless they run `kit loop workflow`.
2. Add only prompt instructions and leave JSON writing to the agent.
   - Pros: minimal code.
   - Cons: weaker binary guarantee; the feature would still depend on agent compliance.
3. Create a new internal package for reflect scoring.
   - Pros: clean unit tests and reusable boundaries.
   - Cons: more abstraction than needed unless the verdict code grows beyond the loop workflow.

Accepted direction after user clarification:

- Target the v2 `kit spec` reflection phase, especially `kit loop workflow` when it runs a feature whose `SPEC.md` phase is `reflect`.
- Add a deterministic scorer/verdict helper used by the loop runner, not the legacy reflect command.
- Update the v2 supervisor/loop prompt so manual reflection also knows `REFLECT.json` is required.
- Update `pkg/cli/capabilities_catalog.go` because `kit loop workflow` would gain a reflect-stage file write.
- Keep scoring implementation outside the editable agent prompt. The prompt can require the artifact, but the verifier/scorer must be enforced by Kit code.

### Source Map

| ID | Source | Selector | Claim / Fact | Used For | Maps To | Status |
| -- | ------ | -------- | ------------ | -------- | ------- | ------ |
| SRC-001 | User thesis attachment | `pasted-text.txt` | Requested additive `REFLECT.json` verdict schema, fail-closed behavior, no prompt curation/dispatch/self-editing scope, and stop-after-plan gate. | Scope, non-goals, acceptance criteria, phase gate | AC-001..AC-010 | confirmed |
| SRC-002 | `docs/agents/WORKFLOWS.md` + `docs/CONSTITUTION.md` | v2 single-SPEC guidance | V2 feature work keeps durable requirements, plan, task checklist, validation, reflection, delivery, and evidence inside `SPEC.md`; legacy staged docs are historical unless explicitly used. | Workflow state and no-v1-artifact decision | AC-010 | confirmed |
| SRC-003 | `go run ./cmd/kit map 0039-reflect-scoring` | command output | Feature `0039-reflect-scoring` currently has `SPEC.md` only; no BRAINSTORM, PLAN, TASKS, or ANALYSIS artifacts exist. | Stop before TASKS and preserve v2 single-artifact flow | AC-010 | confirmed |
| SRC-004 | `pkg/cli/reflect.go` | `reflectCmd`, `runReflect`, `buildReflectPrompt` | Legacy reflect is a deprecated v1 staged prompt command; after user clarification it is not the target for this feature. | Scope boundary | AC-006, AC-010 | confirmed |
| SRC-005 | `pkg/cli/legacy.go`; `go run ./cmd/kit reflect --help`; `go run ./cmd/kit legacy reflect --help` | `legacyCmd` and CLI output | There is no root `kit reflect` command in this checkout; the v2 reflection phase lives inside `kit spec`/`kit loop workflow`, not a root command. | Command-surface boundary | AC-001, AC-009, AC-010 | confirmed |
| SRC-006 | `internal/document/metadata.go` | `Metadata`, `MetadataClarification`, `UpsertMetadata` | Front matter schema includes artifact, workflow version, phase, delivery intent, clarification state, feature metadata, relationships, references, and skills. | Phase/front-matter conventions | AC-010 | confirmed |
| SRC-007 | `internal/templates/templates.go` | `BuildSpecArtifactForFeature`, `BuildTasksArtifactForFeature` | New v2 SPECs get `workflow_version: 2`, `phase: clarify`, and open clarification metadata; TASKS has canonical front matter when created through legacy scaffolding. | Artifact conventions | AC-010 | confirmed |
| SRC-008 | `internal/feature/feature.go` | `Phase`, `DeterminePhase`, `ReflectionCompleteMarker` | V2 phase front matter is authoritative; legacy TASKS checkbox state and reflection marker are fallback behavior. | Phase-state schema and reflect lifecycle | AC-001, AC-010 | confirmed |
| SRC-009 | `pkg/cli/loop_runner.go`, `pkg/cli/loop_validate.go`, `pkg/cli/loop_prompt.go` | v2 loop workflow | `kit loop workflow` resolves the current `SPEC.md` phase, builds a stage prompt, invokes the configured agent, updates rollup, checks verification failures, and validates phase advancement; it does not currently require `REFLECT.json` before leaving `reflect`. | V2 reflect-stage scorer hook and gate | AC-001, AC-005, AC-006, AC-009 | confirmed |
| SRC-010 | `internal/verify/execute.go`; `pkg/cli/verify.go` | `ExecuteRun`, `runVerify` | Existing verification execution records command argv, raw command, exit code, stdout/stderr, status, and timestamps; `kit legacy verify` writes run artifacts unless disabled. | TestsPass and fail-closed command evidence | AC-002, AC-005 | confirmed |
| SRC-011 | `Makefile`, `.githooks/pre-commit`, release workflows | `make build`, `make test`, `make lint`, `make vet` | Pre-commit runs `make build`; release workflows run `make vet` and `make test`; `make lint` exists but no CI/pre-commit hook currently runs lint. | Test/lint source and lint blocker | AC-002, AC-003, AC-008 | confirmed |
| SRC-012 | `docs/references/rules/command-capabilities.md`; `pkg/cli/capabilities_catalog.go`; `kit capabilities loop workflow --json` | `capabilityCatalog` | Command behavior changes inside Kit must update capabilities; current loop workflow metadata documents loop artifacts but not reflect-stage `REFLECT.json` writes. | Capability metadata work | AC-009 | confirmed |
| SRC-013 | `docs/specs/0031-executable-verification-harness/SPEC.md` | requirements `SPEC-02`, `SPEC-05`, `SPEC-07`, `SPEC-08`, `SPEC-12` | Prior verification harness established task fields, JSON run evidence, run artifacts, and reflect evidence gating. | Reuse existing harness instead of new runner | AC-001..AC-005 | confirmed |
| SRC-014 | `docs/specs/0033-kit-capabilities/SPEC.md` | `SPEC-29` | Kit command behavior extensions must update `kit capabilities` in the same change. | Capability acceptance criterion | AC-009 | confirmed |
| SRC-015 | `git status --short --branch` and `git diff -- docs/PROJECT_PROGRESS_SUMMARY.md` | worktree output | Worktree is on `main`, tracking `origin/main`, with in-scope generated 0039 docs and progress-summary updates already present before implementation. | Dirty-worktree gate | AC-010 | confirmed |
| SRC-016 | User clarification | current conversation | Scoring belongs to the reflection phase of the current `kit spec` workflow, not the deprecated legacy reflect command. | Accepted command/workflow target | AC-001, AC-006, AC-009, AC-010 | confirmed |
| SRC-017 | `pkg/cli/spec_v2_prompt.go` | `Reflection Phase` | The generated v2 supervisor prompt defines reflection as a post-validation review recorded in `SPEC.md`, but does not currently require `REFLECT.json`. | Prompt update target | AC-001, AC-006 | confirmed |
| SRC-018 | External self-improvement loops source | linked `SKILL.md` | The scoring signal should be outside agent-editable prompt content, tied to raw artifacts rather than agent reports, scoped to the lowest sufficient workflow change, and leave evaluator/surface expansion decisions human-gated. | Self-improvement-aligned constraints | AC-001, AC-005, AC-006, AC-011 | confirmed |

## CLARIFICATIONS

### Resolved By Repo Research

1. The current reflect command file is `pkg/cli/reflect.go`.
2. The current callable reflect surface is `kit legacy reflect`, not root `kit reflect`.
3. Phase/front-matter conventions live in `internal/document/metadata.go`, `internal/templates/templates.go`, and `internal/feature/feature.go`.
4. Existing task verification parsing and execution lives in `internal/verify` and `pkg/cli/verify.go`.
5. Existing local quality commands are Make targets: `make build`, `make test`, `make lint`, `make vet`.
6. Existing pre-commit only runs `make build`.
7. Existing release CI runs `make vet` and `make test`.
8. No `.golangci.yml` or other lint configuration file was found in the repo root scan.
9. User clarified that scoring belongs to the reflection phase of the current `kit spec` workflow.
10. V2 loop reflection is handled through `kit loop workflow`, `SPEC.md phase: reflect`, `pkg/cli/loop_runner.go`, `pkg/cli/loop_validate.go`, and `pkg/cli/spec_v2_prompt.go`.

### Open Question Batch 1

Confidence: 96%. Unresolved questions: 0. Gate: implementation is approved with the defaults below.

1. Resolved by user clarification: target the v2 `kit spec` reflection phase, not `kit legacy reflect` or a restored root `kit reflect` command.

2. How should `LintDelta` be computed when no CI/pre-commit hook currently runs lint and there is no lint config file?
   - Recommended default: use the existing `make lint` target, count current `golangci-lint` findings as the delta when the command runs, and fail closed if the target/tool output is unavailable or unparseable.
   - Assumption: for step 1, "delta" can mean "new findings observed by the existing lint target for this reflect run" rather than a true base-vs-branch comparison.
   - Decision: accepted by user approval on 2026-07-08. Do not add branch-baseline lint machinery in this step.

3. How should the v2 approval/start boundary be identified for `cycle_time_min` and `rework_count`?
   - Recommended default: use the latest commit that changed the feature `SPEC.md` into `phase: ready` as the approval boundary; fail closed if git history cannot identify it.
   - Assumption: in the v2 workflow, "TASKS approval" maps to the point where `SPEC.md` readiness gates are complete and implementation is approved to start.
   - Decision: accepted by user approval on 2026-07-08. Do not add a new persisted approval-marker schema in this step.

## REQUIREMENTS

- REQ-001: The v2 reflection phase must write `REFLECT.json` in the same feature directory as `SPEC.md`.
- REQ-002: `REFLECT.json` must use the requested JSON keys: `tests_pass`, `lint_delta`, `scope_drift`, `cycle_time_min`, `rework_count`, `prompt_version`, and `timestamp`.
- REQ-003: Missing data sources must use zero values only when the field is explicitly optional or untracked; required evidence such as git log and test/lint command evidence must fail closed.
- REQ-004: If required inputs are missing, unavailable, or unparseable, the command must exit non-zero and must not write a partial or guessed `REFLECT.json`.
- REQ-005: Existing legacy reflect markdown/prompt content must remain byte-for-byte unchanged; v2 prompt changes must be limited to reflecting the new verdict requirement.
- REQ-006: `TestsPass` must be derived from existing test command exit codes, not a custom test parser.
- REQ-007: `LintDelta` must be derived from existing lint command output and exit status, subject to the unresolved lint semantics in Q2.
- REQ-008: `ScopeDrift` must compare branch-touched files against declared scope from v2 `SPEC.md` Task Checklist, Source Map, or explicit expected-file lines; legacy `TASKS.md` parsing may be reused only as an implementation helper where it fits.
- REQ-009: Scope drift tiers are `none`, `minor`, and `major`; `none` means touched files are a subset of declared files, `minor` means at most two unlisted files and no declared-file misses, and `major` means more than two unlisted files or any declared file untouched.
- REQ-010: `CycleTimeMin` and `ReworkCount` must derive from git history between the v2 approval/start boundary and reflect invocation, subject to the unresolved boundary semantics in Q3.
- REQ-011: `PromptVersion` must be empty unless prompt template version tracking already exists in a directly discoverable source.
- REQ-012: `Timestamp` must be RFC3339 in UTC or with explicit offset.
- REQ-013: No prompt curation, dispatch, prompt stats, metrics aggregation, or self-editing logic may be changed.
- REQ-014: No new external dependency may be added unless the plan explicitly justifies it and the user approves it.
- REQ-015: Because `kit loop workflow` behavior would gain a reflect-stage `REFLECT.json` write, `pkg/cli/capabilities_catalog.go` must be updated in the implementation.
- REQ-016: Verdict values must be computed by Kit runtime code from raw command/git/diff evidence, not from agent-written summaries, reflection prose, or the final `KIT_LOOP_RESULT` line.
- REQ-017: Prompt text may instruct agents to preserve or surface evidence, but the scoring implementation and pass/fail enforcement must live outside the mutable prompt content.
- REQ-018: Keep this feature at the lowest sufficient self-improvement rung: a workflow/evidence signal only. Do not add candidate search, prompt evolution, self-editing, promotion, or evaluator-tuning behavior.
- REQ-019: Any future change to evaluator semantics, editable surfaces, or automatic promotion remains human-gated and out of scope for this step.

## ASSUMPTIONS

- Accepted from user thesis: this feature is additive and must preserve existing reflect markdown content.
- Accepted from user thesis: binary-verifiable JSON is the goal; prose parsing is not acceptable.
- Accepted from repo research: v2 `SPEC.md` remains the only durable feature artifact for this workflow; the task checklist is embedded in this file rather than split into legacy `TASKS.md`.
- Accepted from user clarification: target the v2 `kit spec` reflection phase and `kit loop workflow`, not `kit legacy reflect`.
- Accepted from repo research: existing command execution and run-evidence patterns should be reused instead of duplicated where possible.
- Accepted from external source material: self-improvement-aligned metrics must be runtime-owned, evidence-bound, and protected from agent self-reporting or prompt-only enforcement.
- Accepted from user approval: compute `LintDelta` from existing `make lint` output rather than adding branch-baseline machinery.
- Accepted from user approval: identify the v2 approval/start boundary from the latest commit that changed the feature `SPEC.md` into `phase: ready`.
- Rejected: creating legacy BRAINSTORM.md, PLAN.md, or TASKS.md files for this v2 feature.
- Rejected: creating issues, branches, commits, pushes, PRs, or review-thread mutations during this implementation pass.

## ACCEPTANCE CRITERIA

- AC-001: A clean v2 reflection run writes valid `REFLECT.json` next to `SPEC.md`, matching the requested schema and RFC3339 timestamp format.
- AC-002: Unit tests cover a `tests_pass=false` path derived from an existing command result exit code.
- AC-003: Unit tests cover a `lint_delta>0` path derived from existing lint command output.
- AC-004: Unit tests cover every `scope_drift` tier: `none`, `minor`, and `major`.
- AC-005: Unit tests cover a missing or unparseable git-log input path that exits non-zero and does not write `REFLECT.json`.
- AC-006: Legacy reflect markdown/prompt output remains byte-for-byte unchanged, while the v2 supervisor/loop prompt explicitly requires `REFLECT.json`.
- AC-007: No new external dependency is added unless this SPEC is updated with justification before implementation.
- AC-008: `go vet ./...`, relevant Go tests, and the existing lint path are clean or fail closed with documented blocker evidence.
- AC-009: `kit capabilities loop workflow --json` accurately reports reflect-stage `REFLECT.json` file-write behavior after implementation.
- AC-010: No dispatch, prompt curation, prompt stats, metrics aggregation, self-editing agent logic, or GitHub delivery mutation is changed by this feature.
- AC-011: Tests prove `REFLECT.json` values are derived from raw command/git/diff evidence and cannot be satisfied by agent summary text alone.

## IMPLEMENTATION PLAN

Approved implementation:

1. Resolve the two remaining open questions and update this SPEC before writing implementation code.
2. Generate a concise v2 Task Checklist inside this SPEC; do not create legacy `TASKS.md` for this feature unless explicitly requested.
3. Keep the command-surface patch minimal:
   - Modify the v2 `kit loop workflow` reflect stage.
   - Do not restore root `kit reflect`.
   - Do not add scoring behavior to `kit legacy reflect`.
4. Add a small verdict helper file, likely `pkg/cli/reflect_verdict.go` or `pkg/cli/loop_reflect_verdict.go`, rather than growing `pkg/cli/loop_runner.go` substantially.
5. Define the verdict data type and JSON write path:
   - Use the requested JSON keys.
   - Write atomically to a temp file in the feature directory, then rename.
   - Do not write if any required input fails.
   - Do not let the agent supply, overwrite, or certify the metric values through prompt output.
6. Add the v2 reflect-stage hook:
   - In `executeLoop`, when `before.Stage == loopStageReflect` and the agent reports done, generate and validate `REFLECT.json` before accepting advancement out of `reflect`.
   - Add strict validation in or near `loop_validate.go` so a missing/invalid verdict blocks the workflow.
   - Keep `REFLECT.json` generated evidence, not canonical source of truth.
7. Reuse existing evidence and command patterns:
   - Use existing Make targets for test/lint execution unless Q2 changes that.
   - Reuse `internal/verify.CommandResult` style fields or a small testable command-runner seam for exit codes and output.
   - Parse declared scope from v2 `SPEC.md` Task Checklist/Source Map/expected-file lines before falling back to legacy TASKS helpers.
   - Bind every field to raw evidence already available in command output, git output, diff file lists, or loop/run artifacts; if the evidence is missing, fail closed rather than accepting an agent report.
8. Derive touched files from git:
   - Use merge-base against the configured/default branch when available, otherwise fail closed with an actionable error.
   - Normalize paths repo-relative and slash-separated before comparison.
9. Derive cycle time and rework count from git:
   - Locate the v2 approval/start boundary per Q3.
   - Compute elapsed minutes from that commit timestamp to invocation.
   - Count commits after the boundary that touch the same declared/touched file set.
10. Preserve legacy reflect output and update v2 prompt guidance:
   - Keep `buildReflectPrompt` output unchanged.
   - Update `pkg/cli/spec_v2_prompt.go` and loop prompt tests so the v2 reflection phase requires `REFLECT.json`.
11. Update command capability metadata:
   - Change `loop workflow` metadata to mention reflect-stage `REFLECT.json` writes.
   - Leave `legacy reflect` metadata unchanged unless implementation touches it.
12. Add focused tests:
   - Pure helper tests for schema, scope tiers, lint parsing, git-log failure, and rework counting.
   - Loop workflow test with temp repo/feature fixture proving no partial file on failure and no stage advancement without valid verdict evidence.
   - Test proving agent stdout/final loop JSON cannot spoof `tests_pass`, `lint_delta`, or `scope_drift`.
   - Golden test confirming legacy reflect prompt content is unchanged and v2 prompt content includes the verdict requirement.
   - Capability test for updated file-write behavior.
13. Run validation:
   - `go test ./pkg/cli`
   - `go test ./...`
   - `go vet ./...`
   - `make lint` if available; otherwise document the exact tool availability failure as a blocker rather than fabricating lint evidence.
   - `go run ./cmd/kit capabilities loop workflow --json`

Tradeoffs:

- Keeping verdict helpers in `pkg/cli` avoids a new abstraction, but tests must keep the helpers small and explicit.
- A true branch-baseline `LintDelta` is more accurate but likely needs either tool-specific flags or a temporary checkout strategy; the recommended default avoids that scope for step 1.
- Hooking `kit loop workflow` gives mechanical enforcement; direct prompt-only `kit spec` users will rely on the updated prompt unless a later feature adds a standalone v2 reflect-scoring command.
- Evidence-bound scoring adds more plumbing than prompt-only scoring, but it is the self-improvement-safe option because future optimizer loops will target any gap between the metric and the intended behavior.

Rollback strategy:

- Remove the new verdict helper/test files.
- Revert the loop-stage hook and validation changes.
- Revert the v2 prompt and capability metadata changes.
- Delete any generated local `REFLECT.json` test artifacts from fixtures.
- Preserve this SPEC history and record the rollback reason in Evidence.

## TASK CHECKLIST

- [x] TASK-001 - Accept default clarification decisions and move the v2 SPEC into implementation mode. Expected files: `docs/specs/0039-reflect-scoring/SPEC.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`.
- [x] TASK-002 - Add the runtime-owned reflect verdict scorer and atomic `REFLECT.json` writer. Expected files: `pkg/cli/loop_reflect_verdict.go`, `pkg/cli/loop_reflect_verdict_test.go`.
- [x] TASK-003 - Hook verdict generation into the v2 loop workflow reflect stage without changing legacy reflect behavior. Expected files: `pkg/cli/loop_runner.go`, `pkg/cli/loop_reflect_verdict.go`, `pkg/cli/loop_reflect_verdict_test.go`.
- [x] TASK-004 - Update v2 prompt guidance and command capability metadata for the reflect-stage `REFLECT.json` artifact. Expected files: `pkg/cli/spec_v2_prompt.go`, `pkg/cli/capabilities_catalog.go`, `pkg/cli/*_test.go`.
- [x] TASK-005 - Add focused validation coverage for test exit evidence, lint parser evidence, scope-drift tiers, git-boundary fail-closed behavior, spoof-resistant scoring, v2 prompt wording, capability metadata, and legacy reflect stability. Expected files: `pkg/cli/loop_reflect_verdict_test.go`, `pkg/cli/*_test.go`.
- [x] TASK-006 - Run validation and record evidence in this SPEC. Expected files: `docs/specs/0039-reflect-scoring/SPEC.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`.

## VALIDATION MAP

| AC | Validation Method | Evidence Target | Current Status |
| -- | ----------------- | --------------- | -------------- |
| AC-001 | Loop workflow reflect-stage test with temp feature repo plus JSON schema decode | `TestExecuteLoopRunsConfiguredAgentUntilComplete` writes and decodes `REFLECT.json` | pass |
| AC-002 | Unit test command result with non-zero test exit | `TestBuildLoopReflectVerdictUsesRawCommandEvidence` | pass |
| AC-003 | Unit test lint output parser with count > 0 | `TestParseLintIssueCount` | pass |
| AC-004 | Unit tests for scope drift tiers | `TestClassifyReflectScopeDriftTiers` | pass |
| AC-005 | Unit/command test for missing git log and no file write | `TestWriteLoopReflectVerdictFailsClosedWithoutReadyBoundary` | pass |
| AC-006 | Existing legacy reflect golden plus updated v2 supervisor prompt golden | `TestBuildReflectPrompt_Golden`, `TestBuildSpecV2SupervisorPrompt_Golden`, v2 prompt contract checks | pass |
| AC-007 | `go.mod` / `go.sum` diff review | no dependency diff | pass |
| AC-008 | `go test ./...`, `go vet ./...`, lint path | tests/vet pass; `make lint` reports pre-existing repo-wide findings; diff-only lint reports `0 issues` | pass with documented residual lint blocker |
| AC-009 | `go run ./cmd/kit capabilities loop workflow --json` and capability tests | capability JSON includes reflect-stage `REFLECT.json` write and raw-evidence caveat | pass |
| AC-010 | Scope review against diff and `rg` for forbidden surfaces | no diff in legacy reflect, dispatch, improve, prompt stats, metrics, self-editing surfaces, or dependency files | pass |
| AC-011 | Unit and loop tests that ignore agent-provided metric claims and compute values only from raw evidence | failing `make test` exit code beats spoofed `tests_pass=true` text in `TestBuildLoopReflectVerdictUsesRawCommandEvidence` | pass |

## REFLECTION NOTES

Reflection completed after validation.

- The implementation stayed on the v2 `kit loop workflow` reflect-stage boundary and left `kit legacy reflect` unchanged.
- The scorer is runtime-owned: tests, lint, touched files, ready-boundary git history, and rework count are read through command/git evidence; agent output cannot satisfy the verdict by itself.
- `REFLECT.json` is written atomically only after required evidence is available and parseable. The missing-ready-boundary test proves the fail-closed no-file-write path.
- The accepted lint default is implemented, but the repository currently has broad lint debt. `make lint` fails on pre-existing findings outside this feature; diff-only lint reports `0 issues`.
- This remains the lowest sufficient self-improvement rung: no prompt evolution, dispatch, prompt stats, metrics aggregation, self-editing, candidate promotion, or evaluator tuning was added.

## DOCUMENTATION UPDATES

Current planning updates:

- `docs/specs/0039-reflect-scoring/SPEC.md`: populated from scaffold with repo-grounded context, Source Map, clarified requirements, acceptance criteria, candidate plan, validation map, and open questions.
- `docs/specs/0039-reflect-scoring/SPEC.md`: updated after user clarification to target the v2 `kit spec` reflection phase rather than legacy reflect.
- `docs/specs/0039-reflect-scoring/SPEC.md`: updated after reviewing external self-improvement source material to make the verdict runtime-owned, raw-evidence-bound, and scoped below self-editing/promotion behavior.

Planned implementation documentation updates:

- `pkg/cli/capabilities_catalog.go`: updated `loop workflow` metadata to report the reflect-stage `REFLECT.json` write and raw-evidence caveat.
- `pkg/cli/spec_v2_prompt.go` and `pkg/cli/testdata/spec_v2_supervisor_prompt.golden`: updated v2 reflection guidance so agents know the verdict is runtime-owned and must not be fabricated.
- README/root help were not updated because no visible command, flag, or root help contract changed.
- Dispatch, prompt curation, prompt stats, metrics aggregation, and self-editing docs were intentionally not updated because those surfaces stayed out of scope.

## DELIVERY DECISION

User requested GitHub delivery on 2026-07-08 using Kit-managed repository rules.

- Issue: `#50` (`https://github.com/jamesonstone/kit/issues/50`), assigned to `jamesonstone`.
- Branch: `GH-50`, created from refreshed `origin/main`.
- PR: pending at commit time; create ready PR after commit and push.

## EVIDENCE

Planning evidence gathered:

- `git status --short --branch`: on `main...origin/main`; existing in-scope changes include `docs/PROJECT_PROGRESS_SUMMARY.md`, `docs/notes/0039-reflect-scoring/**`, and `docs/specs/0039-reflect-scoring/**`.
- `go run ./cmd/kit map 0039-reflect-scoring`: feature phase `clarify`, `SPEC.md` present, BRAINSTORM/PLAN/TASKS/ANALYSIS missing, notes reference resolved.
- `go run ./cmd/kit reflect --help`: failed with unknown root command `reflect`.
- `go run ./cmd/kit legacy reflect --help`: confirmed current reflect surface and prompt-only v1 staged wording.
- `go run ./cmd/kit capabilities legacy reflect --json`: current metadata reports mutation/file writes as none.
- `go run ./cmd/kit capabilities legacy verify --json`: confirmed verify may execute commands and write run artifacts unless dry-run/no-write.
- Repo search confirmed no `.golangci.yml` or equivalent lint config file was present.
- `.githooks/pre-commit`: runs `make build`.
- Release workflows: run `make vet` and `make test`.
- `Makefile`: defines `build`, `test`, `lint`, `fmt`, `vet`, and `all`.
- User clarification: scoring belongs to the reflection phase of what now occurs as part of the `kit spec` workflow.
- `pkg/cli/loop_runner.go`: `executeLoop` runs stage prompts and validates phase advancement.
- `pkg/cli/loop_validate.go`: strict loop validation currently does not require `REFLECT.json`.
- `pkg/cli/spec_v2_prompt.go`: current Reflection Phase prompt text requires reflection notes in `SPEC.md` but not a structured verdict artifact.
- External source review: self-improvement loop guidance supports runtime-enforced scoring outside mutable prompt content, raw evidence binding, lowest-sufficient-rung changes, and human-gated evaluator/surface expansion.

Implementation and validation evidence:

- `pkg/cli/loop_reflect_verdict.go`: added `ReflectVerdict`, runtime evidence collection, lint finding parsing, expected-file scope drift classification, ready-boundary git history parsing, rework counting, and atomic `REFLECT.json` writes.
- `pkg/cli/loop_runner.go`: writes the reflect verdict when `before.Stage == loopStageReflect` before accepting stage advancement.
- `pkg/cli/spec_v2_prompt.go`: v2 Reflection Phase now states `kit loop workflow` writes runtime-owned `REFLECT.json` and agents must not fabricate verdict values.
- `pkg/cli/capabilities_catalog.go`: `loop workflow` capability now reports the reflect-stage `REFLECT.json` file write and raw-evidence caveat.
- `go test ./pkg/cli`: passed.
- `go test ./...`: passed.
- `go vet ./...`: passed with no output.
- `make lint`: failed with 58 pre-existing repo-wide issues (`errcheck`, `staticcheck`, `unused`); no reported finding referenced `pkg/cli/loop_reflect_verdict.go` or `pkg/cli/loop_reflect_verdict_test.go`.
- `golangci-lint run --new-from-rev=HEAD ./...`: passed with `0 issues`.
- `go run ./cmd/kit capabilities loop workflow --json`: passed and reported `writes loop prompts, stdout, stderr, and run summaries under .kit/loops; when the v2 reflect stage completes, writes runtime-owned REFLECT.json next to SPEC.md`.
- `git diff -- pkg/cli/reflect.go pkg/cli/dispatch.go pkg/cli/improve.go pkg/cli/prompt_stats.go`: no diff.
- `git diff -- go.mod go.sum`: no diff.
- GitHub delivery recon: repository `jamesonstone/kit`; base branch `main`; current branch was `main`; no active PR existed for `main`; no matching open issue was found for `reflect scoring REFLECT json`; git author and committer resolved to `Jameson Stone <jameson@stone.tc>`; GitHub login resolved to `jamesonstone`.
- `gh issue create --title "Add v2 reflect scoring" --assignee @me`: created `https://github.com/jamesonstone/kit/issues/50`.
- `git fetch origin main`: refreshed remote base.
- `git checkout -b GH-50 origin/main`: created `GH-50` from `origin/main`; branch HEAD matched `origin/main`; no existing PR was found for `GH-50`.
