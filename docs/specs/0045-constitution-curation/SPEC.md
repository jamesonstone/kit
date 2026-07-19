---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0045
  slug: constitution-curation
  dir: 0045-constitution-curation
relationships:
  - type: builds_on
    target: 0042-native-plan-repository-memory
  - type: builds_on
    target: 0043-init-makefile-scaffold
  - type: related_to
    target: 0028-project-refresh-advisory
skills:
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the completed change to the existing pull request
    required: true
references:
  - id: init-command
    name: Init command
    type: code
    target: pkg/cli/init.go
    relation: implements
    read_policy: must
    used_for: fresh-project prompt behavior and visible next steps
    status: active
  - id: setup-gate
    name: Spec setup gate
    type: code
    target: pkg/cli/spec_setup_gate.go
    relation: implements
    read_policy: must
    used_for: bootstrap Constitution readiness
    status: active
  - id: rule-registry
    name: Rules registry
    type: code
    target: pkg/cli/rules_registry.go
    relation: implements
    read_policy: must
    used_for: downstream ruleset visibility and adoption
    status: active
  - id: v3-instructions
    name: V3 repository instructions
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: always-loaded routing to just-in-time Constitution curation
    status: active
  - id: project-refresh
    name: Project refresh prompt
    type: code
    target: pkg/cli/project_refresh_prompt.go
    relation: guides
    read_policy: must
    used_for: periodic audit after continuous per-change curation
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Keep `docs/CONSTITUTION.md` aligned with demonstrated project-wide truth as implementation evolves, without forcing a new project to predict its complete identity or turning `kit init` into a long interview.

## CONTEXT

- Kit now uses native agent planning first, captures material accepted plans in V3 `SPEC.md` before implementation, and reconciles those specs with what actually shipped after validation.
- A fresh `kit init` currently writes the generated Constitution starter and immediately asks an agent to populate it by analyzing the codebase. Empty and sparse repositories do not contain enough product evidence for that request, so the agent is pushed toward invented intent or generic scaffold restatement.
- Initial product ideas belong in the accepted native plan and feature spec. The Constitution is the current durable project contract, while feature specs and Git history preserve how that contract evolved.
- Generated V3 repository instructions already require post-validation memory curation, but the detailed promotion criteria should live in one downstream registry rule instead of being duplicated across always-loaded provider files.
- Registry-visible downstream rulesets are adopted by `kit init` and refresh, but rules remain just-in-time context. A short pointer in the always-loaded V3 instructions is required to make the rule reliable during normal coding-agent finalization.
- `kit project refresh` remains a periodic semantic audit for missed, stale, or cross-feature patterns; it is a safety net rather than the primary mechanism for keeping the Constitution current.
- The user explicitly requested that this work extend existing issue `#62`, branch `GH-62`, and pull request `#63`.

## REQUIREMENTS

- Add an active downstream registry ruleset named `constitution-curation`.
- Route V3 Codex, Claude, and Copilot repository instructions to the rule after implementation and validation, without inlining the full ruleset into always-loaded files.
- Define the generated Constitution starter as a valid bootstrap state when no project-specific truth has yet been demonstrated.
- Make the curation rule require evidence from implemented behavior, validated outcomes, current specs, repository documentation, and recurring conventions; initial aspiration alone is insufficient.
- Promote only durable project-wide principles, constraints, non-goals, definitions, vocabulary, or workflow boundaries into the Constitution.
- Keep feature-specific rationale and superseded decisions in the relevant `SPEC.md`; exclude transient planning chatter and code-recoverable implementation detail from the Constitution.
- Permit and require a `not required` result when the completed work reveals no constitutional change.
- Require stale constitutional rules to be corrected or removed when current implementation disproves them, with material historical rationale retained in the relevant spec.
- Treat project-refresh cadence as a trigger to review, never as permission for an automatic or blind rewrite.
- Change the shared initialization prompt so it inspects available repository evidence, preserves the starter Constitution when evidence is insufficient, does not request an exhaustive project interview, and does not invent project commands.
- Preserve the safe starter Makefile unchanged when no verified repository-native commands exist; populate only verified thin wrappers when evidence exists.
- Treat the exact generated Constitution template as ready bootstrap state in spec setup and project reconciliation.
- Continue reporting incomplete setup for missing Constitutions and for partially customized Constitutions whose required sections remain empty or placeholder-only.
- Keep the immutable `kit instructions` `v1` payload unchanged.
- Make native `kit spec` output remind coding agents to run `kit status` and follow any Kit-managed refresh action before implementation.
- Make successful `kit complete` output remind coding agents to run `kit status` and follow any Kit-managed refresh action before final delivery.
- Keep both reminders non-blocking and advisory-only: do not run status automatically, add a managed-guidance version field, introduce a separate upgrade framework, or mutate managed files.
- Update generated templates, checked-in repository instruction files, core init documentation, durable Constitution guidance, focused tests, project rollup, issue scope, and PR description.
- Observable acceptance: fresh init output teaches evidence-based bootstrap behavior without the former exhaustive-drafting request; exact starter Constitutions pass setup and project reconciliation; partially customized incomplete Constitutions still fail the relevant gate; every checked-in V3 provider entrypoint routes to the new valid ruleset; full repository validation passes.
- Non-goals: background Constitution edits, mandatory interactive initialization, capturing a complete initial product vision, making the Constitution a changelog, removing periodic project refresh, modifying immutable instruction payloads, or automatically migrating local-custom downstream rules.

## ACCEPTED PLAN

1. Add a canonical downstream `constitution-curation` ruleset covering bootstrap, evidence, promotion, correction, no-op, refresh, and verification behavior.
2. Add a concise pointer to that rule in generated V3 provider instructions and align the checked-in instruction files.
3. Rewrite the shared init prompt and next steps around evidence-based curation and verified Makefile commands, leaving empty-project starters intact.
4. Share an exact-template bootstrap predicate between spec setup and project reconciliation so the generated starter is valid while partial placeholder documents remain actionable.
5. Add focused prompt, setup, reconciliation, instruction-template, and ruleset tests; update canonical docs and the project rollup.
6. Run formatting, focused tests, full Go validation, build, Kit checks, prompt-system checks, and diff review before explicit staging, commit, push, and existing-PR update.
7. Add concise `kit status` advisories to the native spec and completion lifecycle outputs, document the behavior, and verify it without changing managed-guidance versioning or refresh semantics.

## DECISIONS

- Accepted continuous rule-driven curation as the primary mechanism and periodic project refresh as a safety net.
- Accepted an exact generated-template bootstrap exemption; rejected broad heuristics that could hide partially authored but incomplete Constitutions.
- Accepted a short always-loaded pointer because a registry rule cannot guarantee its own just-in-time loading.
- Rejected a long `kit init` interview and rejected treating initial intent as constitutional truth.
- Rejected automatic Constitution mutation because semantic promotion, correction, and removal require reviewed agent judgment.
- Accepted lifecycle advisories that route coding agents to the existing `kit status` freshness check; rejected a new version field or parallel update/upgrade framework as unnecessary.

## DISCOVERIES

- The rules registry makes downstream rulesets available to managed projects, but a ruleset cannot guarantee its own just-in-time loading. Generated V3 provider entrypoints therefore need one concise post-validation pointer to the canonical rule.
- Spec setup already had stronger Constitution population checks than project reconciliation. The exact-template bootstrap exemption exposed that reconciliation checked required section presence but not meaningful Constitution content, so the audit now requires populated customized sections while bypassing only the untouched generated starter.
- Keeping the generated Constitution template byte-identical avoids turning existing bootstrap files into false customized documents and preserves deterministic exact-template detection.
- The periodic project-refresh machinery already records cadence without rewriting documentation. Its prompt and reflection advisory needed only to distinguish that broader audit from continuous per-change curation.
- `kit status` already owns the bounded registry freshness check, reports Kit-managed drift, and recommends the applicable refresh action. Routing lifecycle output to that command is sufficient and avoids duplicating network or upgrade behavior in `kit spec` and `kit complete`.
- The status reminder belongs only in native `kit spec` orientation; changing legacy prompt output would expand the compatibility surface without improving the current coding-agent path.

## VALIDATION

- Focused init, setup-gate, reconciliation, ruleset, V3 instruction-template, and project-refresh tests passed in `pkg/cli` and `internal/templates`.
- The first focused run revealed the reconciliation population-check asymmetry; after the narrow audit correction, the focused suite passed.
- `make fmt`, `git diff --check`, and `go vet ./...` passed.
- `go test ./... -count=1` passed across all packages after updating the intentionally changed built-in prompt and reflection golden expectations.
- `go test -race ./internal/templates ./pkg/cli -count=1` passed.
- `make build` passed and produced `bin/kit` at version `v1.0.91`.
- `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit` passed with 45 traces in run `20260717T195803.113251000Z-6ccd07`.
- Full-repository `golangci-lint run ./...` remains blocked by 45 pre-existing findings outside this feature's diff; changed-lines validation with `golangci-lint run --new-from-rev=origin/main ./...` passed with `0 issues`.
- `./bin/kit check 0045-constitution-curation` passed at the deliver gate.
- `./bin/kit complete 0045-constitution-curation` passed after the explicit deliver transition, set the phase to `complete`, and refreshed `docs/PROJECT_PROGRESS_SUMMARY.md`.
- Final `./bin/kit check 0045-constitution-curation` passed.
- Final `./bin/kit check --project` exited successfully with 15 historical V2 compatibility advisories and a due project-refresh cadence warning; no blocking finding was introduced by this feature.
- Follow-up focused tests for native spec orientation, completion advisories, and the capability catalog passed in `pkg/cli`.
- Follow-up `make fmt`, `go vet ./...`, `go test ./... -count=1`, focused race tests, `make build`, and `golangci-lint run --new-from-rev=origin/main ./...` passed.
- Follow-up `./bin/kit capabilities spec` and `./bin/kit capabilities complete` confirmed that both lifecycle commands remain network-free and describe the advisory-only status handoff.
- Follow-up `./bin/kit check 0045-constitution-curation`, `./bin/kit check --project`, and `git diff --check` passed; project validation retained only the same 15 historical V2 compatibility advisories and due project-refresh warning.

## OUTCOME

- Fresh `kit init` now treats the exact generated Constitution as a valid bootstrap, asks the agent to inspect repository evidence instead of defining the entire project, and leaves both Constitution and Makefile starters unchanged when evidence is insufficient.
- Spec setup and project reconciliation accept only the untouched generated starter as bootstrap. Missing Constitutions and partially customized documents with empty required sections remain actionable.
- The active downstream `constitution-curation` ruleset defines evidence order, promotion boundaries, stale-rule correction, no-op reporting, and cadence semantics.
- Generated and checked-in V3 Codex, Claude, and Copilot entrypoints load the rule after validation, making continuous curation part of normal coding-agent usage.
- Project refresh remains a reviewed periodic audit for missed, stale, or cross-feature truth; it does not edit the Constitution automatically or replace per-change curation.
- Native `kit spec` now reminds coding agents to run `kit status` before implementation, and successful `kit complete` repeats that reminder before final delivery. Neither command runs the network check or refreshes managed files itself.
- The immutable `kit instructions` `v1` payload remains unchanged.

## REPOSITORY MEMORY

Decision: created

Rationale: Evidence-based Constitution evolution, bootstrap semantics, and the division between continuous curation and periodic refresh are durable product decisions that must survive beyond the implementation diff.

Artifacts:

- `docs/specs/0045-constitution-curation/SPEC.md`
- `docs/references/rules/constitution-curation.md`
- `docs/CONSTITUTION.md`
- `docs/specs/0000_INIT_PROJECT.md`
