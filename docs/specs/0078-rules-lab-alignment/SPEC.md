---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0078"
  slug: "rules-lab-alignment"
  dir: "0078-rules-lab-alignment"
references:
  - id: orchestration-rule
    name: Agent team orchestration
    type: rule
    target: docs/references/rules/agent-team-orchestration.md
    relation: constrains
    read_policy: must
    used_for: single-lane topology decision for tightly coupled rule edits
    status: active
  - id: rlm-rule
    name: RLM progressive disclosure
    type: rule
    target: docs/agents/RLM.md
    relation: implements
    read_policy: must
    used_for: context-budget operationalization
    status: active
  - id: issue
    name: Align deeply-rooted rules with cross-lab best practices
    type: external
    target: https://github.com/jamesonstone/kit/issues/206
    relation: supports
    read_policy: must
    used_for: accepted scope and delivery identity
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Align Kit's deeply-rooted agent rules with abstracted cross-lab fundamentals that hold for OpenAI, Claude, Google, and Meta models, preserving Kit's safety invariants while removing model-specific leakage and duplicated hierarchy.

## CONTEXT

- Inventory covers `AGENTS.md`, `CLAUDE.md`, `docs/agents/README.md`, `WORKFLOWS.md`, `GUARDRAILS.md`, `RLM.md`, `TOOLING.md`, `docs/CONSTITUTION.md`, 22 rulesets under `docs/references/rules/`, and the global operating contract.
- Most gates are already model-agnostic. Host-specific content lives in `AGENTS.md` Codex thread-init plus browser policy, `codex-thread-initialization.md`, and `TOOLING.md` illustrative mappings naming Opus, Sonnet, Haiku, Codex, Copilot, Warp.
- Cross-lab synthesis: Anthropic smallest-high-signal context plus progressive disclosure plus fresh verifier; OpenAI persistence plus plan-act-reflect plus contradiction-free hierarchy plus typed tools; Google isolated envs plus explicit loops plus human-in-loop plus ask-last; Meta typed contracts plus trajectory evals; HumanLayer stateless reducer plus own-control-flow.
- Topology: single-lane, because rule edits are tightly coupled, high-overlap docs requiring continuous design judgment; parallel lanes would diverge on wording and precedence.
- Landing plan: default new lane; repository `jamesonstone/kit`; issue #206; branch `GH-206`; worktree `~/worktrees/jamesonstone/kit/GH-206`; protected base `main`; create one ready PR.

## REQUIREMENTS

- De-model core: no hardcoded model or host IDs in normative core; host bindings live in `docs/references/host-adapters/` with semantic profiles only in core.
- Single precedence chain declared: `docs/CONSTITUTION.md` over `GUARDRAILS.md` over `rules/` over `SPEC.md` over chat; Goldilocks altitude with heuristics plus structure.
- Persistence with explicit exit conditions, tool budgets, bounded retries, and stall detection.
- Per-step plan-act-reflect micro-loop with ground-every-claim-in-tools.
- Typed tool-contract checklist: names, descriptions, examples, token-efficient outputs, native schemas, split-on-overlap.
- Fresh read-only verifier required for high-risk lanes; builder never marks own work done; trajectory evidence preserved.
- Context-budget ops: durable instructions top and bottom, task plus negative constraints last, delimit untrusted content as data, canonical examples.
- Resumable work via SPEC plus program ledger progress log plus re-orient checklist; lightweight rule-change evals before Constitution promotion.
- Non-goals: no vendor tricks, no model pins, no numeric concurrency caps, no auto-merge, no weakening of deletion, infra, or merge gates.
- Observable acceptance: `kit check --project` passes; `reconcile --all --output-only` clean or only expected managed-file drift; focused plus full Go tests pass; source-size audit clean; ready PR on `GH-206`.

## ACCEPTED PLAN

1. Populate this SPEC and resolve `implementation-delivery` context; load required evidence in order.
2. Implement demodel: add `docs/references/host-adapters/` reference docs, generalize `TOOLING.md` mappings to capability descriptors, point `AGENTS.md` Codex sections to adapters without expanding core.
3. Implement hierarchy: add precedence plus altitude to `docs/agents/README.md`, replace duplicated worklane plus merge text in `CLAUDE.md`, `AGENTS.md`, `TOOLING.md` with pointers to `GUARDRAILS.md`, `work-lane-gating.md`, `github-pr-merge.md`.
4. Implement fundamentals: persistence plus exit in `agent-team-orchestration.md`; reflect loop in `WORKFLOWS.md`; tool contracts in `TOOLING.md`; verifier hardening in orchestration plus testing rule; context ops in `RLM.md`; resumability in `WORKFLOWS.md`; eval discipline in testing rule.
5. Validate: formatting, `go test`, `go vet`, `kit check`, reconcile audit, source-size audit, self-review diff.
6. Curate memory per `constitution-curation.md`; deliver one ready PR from `GH-206`; stop before merge.

## DECISIONS

- Accepted: single-lane execution for this rule edit set due to tight coupling and continuous design judgment.
- Accepted: host-adapters as `docs/references/` reference docs, not registry rulesets, to avoid registry churn for illustrative mappings.
- Accepted: keep `AGENTS.md` Codex gate self-contained pointer plus detail in adapter, preserving pre-inspection invariant while removing model IDs from normative core.
- Accepted: tool contracts belong in `TOOLING.md`, not backend architecture, because agent tools are harness contracts, not product services.
- Accepted: hierarchy via additive precedence plus source-of-truth pointers rather than deletion, preserving `v3GuidanceExpectations` reconcile convergence.
- Accepted: parallel template updates in `internal/templates/` to preserve exact generator equality for `AGENTS.md`, `CLAUDE.md`, `TOOLING.md`, and `RLM.md`.

## DISCOVERIES

- `docs/agents/*` in Kit repo are canonical for Kit itself; `internal/templates/*` scaffolds downstream projects and is intentionally divergent, so this change touches docs only unless validation proves template drift.
- `kit context resolve --workflow implement` is blocked; correct slug is `implementation-delivery`.
- Baseline `kit check --project` passes and reconcile reports no action needed before edits.
- Checked-in `AGENTS.md`, `CLAUDE.md`, `docs/agents/TOOLING.md`, and `docs/agents/RLM.md` must equal the V3 generator output; docs-only edits break `TestCheckedInCapabilityAdapterMatchesGeneratedArtifacts` and related tests, so template sources in `internal/templates/` were updated in parallel to preserve exact equality.
- `v3GuidanceExpectations` in `pkg/cli/reconcile_guidance_expectations.go` pins worklane and merge wording; hierarchy was implemented additively via precedence plus source-of-truth pointers rather than deletion to preserve reconcile convergence.

## VALIDATION

- PASS: `go test ./internal/templates -run TestCheckedInCapabilityAdapterMatchesGeneratedArtifacts|TestMemoryInstructionsPreserveProjectOrientedWorktrees|TestMemoryRepositoryInstructionsRouteConstitutionCuration|TestMemoryRepositoryInstructionsRouteApplicationArchitecture|TestCapabilityAwareHostAdapterIsSharedAndProviderNeutral -count=1 -v`.
- PASS: `go test ./... -count=1` across all packages.
- PASS: `go vet ./...` clean; `gofmt -l internal/templates/` clean; `git diff --check` clean.
- PASS: `/tmp/kit check --project` coherent; `/tmp/kit reconcile --all --output-only` reports no reconciliation needed with source-size audit 758 candidates, 389 eligible files, 0 above 300 lines.
- PASS: self-review of complete diff against plan acceptance, repo-local rules, and human-authorship; no secrets staged.

## OUTCOME

- Rule updates integrated in `GH-206` worktree across docs plus parallel template sources; ready PR pending creation; merge not performed and not authorized.

## REPOSITORY MEMORY

- Artifacts: `docs/specs/0078-rules-lab-alignment/SPEC.md`; updated `docs/agents/README.md`, `RLM.md`, `TOOLING.md`, `WORKFLOWS.md`, `AGENTS.md`, `CLAUDE.md`, `docs/references/rules/agent-team-orchestration.md`, `docs/references/rules/testing-and-environment-validation.md`; new `docs/references/host-adapters/README.md`, `codex.md`, `claude-code.md`, `copilot.md`, `warp.md`; updated `internal/templates/instruction_templates_v3.go`, `instruction_templates_v2_tooling.go`, `instruction_templates_v2_rlm.go`, `instruction_templates_v3_context.go`, `capability_aware_host_adapter.go`, `codex_browser_policy.go`, `capability_aware_host_adapter_test.go`.
- Decision: Constitution update not required. Rationale: changes refine existing gates toward model-agnostic fundamentals without establishing new project-wide invariants beyond current Constitution contract; durable rationale lives in this SPEC and the updated rules.
- Curation: no `docs/CONSTITUTION.md` change; reusable practices live in updated rules and host-adapters as implemented.
