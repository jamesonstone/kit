---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0071"
  slug: list-first-completion-output
  dir: 0071-list-first-completion-output
relationships:
  - type: related_to
    target: 0067-agent-completion-output
    note: Replaces the table and later multi-section envelope with proportional three-section completion output while preserving status and evidence semantics.
  - type: builds_on
    target: 0044-versioned-agent-instructions
  - type: related_to
    target: 0069-multi-agent-orchestration-evaluation-gate
references:
  - id: agent-completion-output-spec
    name: Prior table-based completion contract
    type: spec
    target: docs/specs/0067-agent-completion-output/SPEC.md
    relation: supersedes
    read_policy: evidence
    used_for: historical table-based operator-action decision that this feature replaces
    status: optional
  - id: agent-completion-output
    name: Canonical terminal completion contract
    type: rule
    target: docs/references/rules/agent-completion-output.md
    relation: implements
    read_policy: must
    used_for: proportional conversational and three-section terminal response shape
    status: active
  - id: instruction-gate
    name: Shared completion routing gate
    type: code
    target: internal/templates/agent_completion_output.go
    relation: implements
    read_policy: must
    used_for: generated and checked-in provider instruction mirrors
    status: active
  - id: instruction-v8
    name: Frozen list-first instruction version
    type: code
    target: internal/instructions/versions/v8.md
    relation: informs
    read_policy: evidence
    used_for: immutable prior kit instructions snapshot
    status: optional
  - id: instruction-v9
    name: Frozen proportional instruction version
    type: code
    target: internal/instructions/versions/v9.md
    relation: informs
    read_policy: evidence
    used_for: immutable prior kit instructions snapshot
    status: optional
  - id: instruction-v10
    name: Additive current instruction version
    type: code
    target: internal/instructions/versions/v10.md
    relation: implements
    read_policy: must
    used_for: immutable current kit instructions snapshot
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Keep ordinary coding-agent conversation direct and make substantial completion
reports concise enough to scan. A structured response contains only what
happened, deviations, and next steps; task-specific evidence is folded into
those sections without duplicative profile headings.

## CONTEXT

- Issue #174 asked to remove the table formatting applied to completion
  output because it is harder to read than originally expected, and to
  return to prioritized bulleted lists or another structure that maximizes
  readability while keeping actionable steps first.
- Feature `0067-agent-completion-output` introduced a four-state status
  heading, one immediate operator action table, and left-aligned evidence
  profiles. The table was chosen so PASS still showed an explicit None row
  and so blockers could not hide under completed detail. Renderer behavior
  later showed that even one table is harder to scan than a tight list.
- Issue #178 showed a second readability failure: direct questions, brief
  explanations, confirmations, and ordinary side-chat exchanges still receive
  a large PASS envelope, synthetic None action, profile, and repository-memory
  report. The problem is over-application of the completion contract, not the
  horizontal alignment of its remaining structured report.
- A later production merge/deployment report demonstrated a third failure:
  even correctly structured substantial output becomes hard to read when the
  same facts are repeated across status, operational result, validation,
  feature state, residual notes, coordination, and repository-memory sections.
- The user selected status in the first What happened bullet, explicit `None`
  bullets for empty Deviations and Next steps sections, and one nested evidence
  layer only when needed.
- Instruction versions are immutable. `v9` was current on PR #179 at the start
  of this refinement and contains the proportional but multi-profile contract
  as part of its frozen snapshot.
  The three-section contract must be additive `v10`. Versions `v1` through
  `v9` remain byte-for-byte unchanged.
- Kit distributes the contract through the canonical ruleset, the shared
  instruction gate, checked-in AGENTS/CLAUDE/Copilot/GUARDRAILS mirrors,
  Constitution baseline text, the references index, health/reconcile
  expectations, and focused tests.
- Issue #178, branch `GH-178`, and worktree
  `~/worktrees/jamesonstone/kit/GH-178` are the authorized delivery lane.
  Merge is not authorized.

## REQUIREMENTS

### Proportionality Gate

- Answer direct questions, definitions, confirmations, rewrites, brief
  explanations, small read-only lookups, concise recommendations, and ordinary
  conversational exchanges naturally without a completion envelope.
- Conversational responses must not emit a PASS/PARTIAL/BLOCKED/FAIL heading,
  `## Next actions`, a synthetic None item, a task profile, or Repository
  Memory.
- Require the structured contract when omission could hide a blocker,
  incomplete required scope, required operator action, unresolved failure,
  repository or external-system mutation, delivery artifact, multiple
  validation layers, material evidence, owner/dependency handoff, or an
  explicitly requested canonical report.
- Classify by operational consequence, never by word count, token count,
  elapsed time, or number of tool calls. When uncertain, prefer natural prose
  unless structure preserves operationally important information.

### Structured Operator Action List

- When the structured contract applies, emit exactly these headings in order:
  `## What happened`, `## Deviations`, and `## Next steps`.
- Put `**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**` in the
  first What happened bullet. Do not add a separate status heading.
- Put concise outcomes and only operationally useful evidence under What
  happened. Use at most one nested evidence layer.
- Put every divergence from requested or expected scope under Deviations,
  including blockers, incomplete work, failures, warnings, pending or unknown
  evidence, skipped validation, and degraded execution. Planned exclusions are
  not deviations.
- Put one independently actionable item per Next steps bullet. Name the actor,
  distinguish required from optional actions, and include a copy-ready prompt
  or command for required follow-up.
- Keep Deviations and Next steps present even when empty, using exactly one
  `**None.**` bullet.
- Do not emit separate Validation, Delivery, Feature State, Residual Notes,
  Coordination, Repository Memory, or task-profile sections. Fold their
  non-duplicative facts into the three canonical sections.
- Do not use a Markdown pipe table in the normal response. A higher-priority
  host schema may require another shape, but it must preserve the same three
  semantic groups.

### Unchanged Semantics

- Preserve PASS, PARTIAL, BLOCKED, and FAIL meaning.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and
  NOT_APPLICABLE.
- Preserve task-specific content requirements and required delivery,
  validation, repository-memory, orchestration, program, and environment
  fields inside the three sections without repeating facts.

### Propagation

- Update the canonical ruleset, shared gate, Constitution baseline,
  references index, checked-in instruction mirrors, and reconcile
  expectations together.
- Add immutable instruction version `v10` as current. Do not mutate `v9` or
  earlier.
- Do not change CLI command behavior, flags, or JSON schemas.

### Non-Goals

- Do not merge this pull request.
- Do not rewrite historical `0067` outcome text into a claim that tables
  remain the current contract.
- Do not add a JSON schema or CLI flag for completion output.

## ACCEPTED PLAN

1. Record the user-selected three-section format in this living `0071` spec.
2. Replace the structured envelope and separate task profiles with What
   happened, Deviations, and Next steps, including a complex production example
   that proves evidence can stay concise without being lost.
3. Propagate the format through the shared gate, checked-in mirrors,
   Constitution baseline, related reporting rules, reconcile expectations, and
   focused tests.
4. Copy `v9.md` to `v10.md`, rewrite only the Agent completion output section,
   register `v10` as current, and keep `v1` through `v9` hashes unchanged.
5. Revalidate the complete branch and push a follow-up commit to PR #179.

## DECISIONS

- Replace the operator action table rather than keeping it as a prefix and
  appending a superseding note. Frozen `v7` already contains the table
  language; repeating it in `v8` would tell agents to emit tables.
- Replace the separate status heading and profile sections with exactly three
  semantic sections. Readability requires eliminating duplicated categories,
  not merely left-aligning them.
- Keep status semantics but place the status in the first What happened bullet.
- Keep Deviations and Next steps visible with an explicit None bullet when
  empty. This preserves immediate operator certainty without the old action
  envelope.
- Preserve required evidence by nesting it once under the outcome it supports,
  not by creating more top-level sections.
- Apply a semantic proportionality gate instead of a length threshold. A short
  blocked or delivery response can be operationally substantial, while a long
  conceptual explanation can still be ordinary conversation.
- Default to natural prose when classification is uncertain unless omitting
  structure could hide an operationally important fact.
- Publish additive `v10` instead of mutating `v9`. `v10` equals `v9` except
  the Agent completion output section.
- Update the Kit-managed Constitution baseline sentence because the
  required terminal envelope is a project-wide instruction contract, not
  feature-local rationale.
- Execute as one supervisor lane because the change is tightly coupled
  across one ruleset, shared instruction gate, frozen instruction
  snapshots, and matching tests; splitting would create high-overlap
  contract drift.

## DISCOVERIES

- `origin/main` already made `v7` current for the multi-agent
  orchestration gate. This work cannot treat `v7` as the new current
  version.
- `TestAgentInstructionsV7RequiresMultiAgentOrchestrationGate` previously
  asserted `CurrentAgentVersion == "v7"`. That assertion now lives on the
  `v8` tests so `v7` remains a frozen snapshot test.
- Feature-to-feature `relationships[].type` cannot be `supersedes`; that
  token is a `references[].relation` value. Record the 0067 replacement as
  `related_to` plus a `supersedes` reference.
- Required-snippet reconcile maps cannot encode a forbidden leftover
  operator-action table. Rejection belongs in a separate forbidden check
  on AGENTS.md, Copilot instructions, and GUARDRAILS.md, plus managed
  safety-guidance propagation tests.
- Renderer alignment was not the remaining root cause. A vertically aligned
  list still overwhelms a small answer when the completion contract applies to
  every terminal turn; the durable fix is to narrow applicability.
- Section proliferation is independently harmful even for substantial tasks.
  Profiles remain content checklists, not output headings; one fact must appear
  once in the section that best supports operator understanding or action.
- Hosted review identified that reconcile initially enforced the three-section
  shape without independently enforcing the structured-output trigger. The
  trigger is now required in AGENTS, Copilot, and Guardrails expectations with
  a removal regression fixture so conversational exemptions cannot suppress
  structured reporting for blockers, mutations, or handoffs.

## VALIDATION

- PASS: focused template, instruction-version, rule, related-contract,
  example-boundary, managed propagation, and reconcile drift coverage.
- PASS: `go test ./... -count=1` and `go test -race ./... -count=1`.
- PASS: `make fmt`, `go vet ./...`,
  `golangci-lint run --new-from-rev=origin/main ./...` with zero issues, and
  `make build`.
- PASS: branch-built Kit viewed the canonical rule, rendered instruction
  `v10`, and resolved implementation context with no blockers.
- PASS: feature `0071`, all 69 features, and the project contract.
- PASS: whole-project reconcile reported no semantic reconciliation needed and
  audited 731 candidates / 379 eligible handwritten source/test files with
  zero above 300 physical lines.
- PASS: `kit health --dry-run --json` and
  `kit reconcile --include-files --dry-run` detected expected pre-merge
  managed-registry drift without overwriting branch-owned rules. Health state
  remains literally `attention_needed`.
- PASS: `git diff --check`; `gitleaks dir --no-banner --redact .` scanned
  5.53 MB and found no leaks; instruction `v1` through `v9` hashes are
  unchanged.
- NOT_APPLICABLE: browser, live-integration, deployment, runtime, and
  production suites; this is an instruction-contract change.
- Hosted PR #179 checks require exact-head revalidation after the follow-up
  commit.

## OUTCOME

Ordinary conversation remains natural. Substantial completion now contains
exactly What happened, Deviations, and Next steps. Status is the first What
happened bullet, empty Deviations and Next steps use one None bullet, and
task-specific evidence is folded in once with at most one nested layer. Current
guidance is additive `v10`; `v1` through `v9` remain frozen.

## REPOSITORY MEMORY

- Decision: updated
- Rationale: Proportional applicability fixed small replies, but substantial
  reports remained repetitive. Three semantic sections retain operator-facing
  evidence and action while removing duplicated output categories.
- Artifacts: `docs/specs/0071-list-first-completion-output/SPEC.md`,
  `docs/references/rules/agent-completion-output.md`,
  `docs/CONSTITUTION.md`, `internal/instructions/versions/v10.md`
