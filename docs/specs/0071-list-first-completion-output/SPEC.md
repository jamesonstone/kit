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
    note: Replaces the operator action table with a prioritized action list. Status heading, profiles, and evidence fields remain.
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
    used_for: status-first, action-first terminal response shape
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
    name: Additive current instruction version
    type: code
    target: internal/instructions/versions/v9.md
    relation: implements
    read_policy: must
    used_for: immutable current kit instructions snapshot
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Keep ordinary coding-agent conversation direct and readable while preserving
a status-first, action-first report for substantial completion and handoff.
When the structured contract applies, use a prioritized left-aligned list so
the next step is obvious without fighting renderer table layout.

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
- Instruction versions are immutable. `v8` is current on `main` and contains
  the always-structured list contract as part of its frozen snapshot. The new
  proportional contract must be additive `v9`. Versions `v1` through `v8`
  remain byte-for-byte unchanged.
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

- When the structured contract applies, keep the first human-readable line
  `# <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>`.
- Immediately after that heading, emit a prioritized bullet list titled
  `## Next actions`.
- Order items `Blocker`, `Incomplete`, `Next`, `Optional`, then `None`.
- Omit types that do not apply, except every structured PASS includes one
  `None` item.
- Each item uses a bold lead (`**Blocker — User: …**`) plus indented
  `Why:` and `Continue with:` lines.
- Required `Continue with` values remain copy-ready prompts or commands.
- Do not use a Markdown pipe table for operator actions in the normal
  terminal response. A higher-priority host schema may still require a
  table; preserve semantic order inside that wrapper.

### Unchanged Semantics

- Preserve PASS, PARTIAL, BLOCKED, and FAIL meaning.
- Preserve native evidence states such as PENDING, UNKNOWN, SKIPPED, and
  NOT_APPLICABLE.
- Preserve the eight named profiles plus fallback, left-aligned headings,
  and required delivery, validation, repository-memory, orchestration,
  program, and environment fields.

### Propagation

- Update the canonical ruleset, shared gate, Constitution baseline,
  references index, checked-in instruction mirrors, and reconcile
  expectations together.
- Add immutable instruction version `v9` as current. Do not mutate `v8` or
  earlier.
- Do not change CLI command behavior, flags, or JSON schemas.

### Non-Goals

- Do not merge this pull request.
- Do not rewrite historical `0067` outcome text into a claim that tables
  remain the current contract.
- Do not add a JSON schema or CLI flag for completion output.

## ACCEPTED PLAN

1. Update this living `0071` spec with the proportionality decision before
   implementation.
2. Add the semantic applicability gate, conversational exclusions, structured
   triggers, and paired examples to the canonical ruleset.
3. Propagate the same decision through the shared instruction gate, generated
   and checked-in mirrors, Constitution baseline, references index, RLM route,
   health/reconcile expectations, and focused tests.
4. Copy `v8.md` to `v9.md`, rewrite only the Agent completion output section,
   register `v9` as current, and keep `v1` through `v8` hashes unchanged.
5. Validate both sides of the boundary: ordinary conversational answers remain
   natural, while substantial implementation, delivery, blocked, incomplete,
   and evidence-bearing handoffs retain the structured contract.
6. Curate repository memory and open one ready pull request for issue #178.

## DECISIONS

- Replace the operator action table rather than keeping it as a prefix and
  appending a superseding note. Frozen `v7` already contains the table
  language; repeating it in `v8` would tell agents to emit tables.
- Keep the four action fields (type, action, why, continue-with) as list
  structure instead of dropping why or continue-with. Readability comes
  from vertical scanning, not from dropping the copy-ready prompt.
- Keep PASS's explicit None item. The original table existed partly so
  structured completion could not be confused with omitted follow-up
  reporting. Do not manufacture that item for ordinary conversation.
- Apply a semantic proportionality gate instead of a length threshold. A short
  blocked or delivery response can be operationally substantial, while a long
  conceptual explanation can still be ordinary conversation.
- Default to natural prose when classification is uncertain unless omitting
  structure could hide an operationally important fact.
- Publish additive `v9` instead of mutating `v8`. `v9` equals `v8` except
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

## VALIDATION

- PASS: focused template, instruction-version, rule, reconcile, and managed
  propagation coverage in `internal/instructions`, `internal/templates`, and
  `pkg/cli`.
- PASS: `go test ./... -count=1` and `go test -race ./... -count=1`.
- PASS: `make fmt`, `go vet ./...`,
  `golangci-lint run --new-from-rev=origin/main ./...` with zero issues, and
  `make build`.
- PASS: branch-built Kit listed and viewed `agent-completion-output`, rendered
  instruction `v9`, and resolved implementation context with no blockers.
- PASS: feature `0071`, all 69 features, and the project contract.
- PASS: whole-project reconcile reported no semantic reconciliation needed and
  audited 729 candidates / 378 eligible handwritten source/test files with
  zero above 300 physical lines.
- PASS: `kit health --dry-run --json` and
  `kit reconcile --include-files --dry-run` detected the expected pre-merge
  managed-registry drift for the locally changed canonical rule. Health state
  is literally `attention_needed`; it did not overwrite the branch rule.
- PASS: `git diff --check`; `gitleaks dir --no-banner --redact .` scanned
  5.15 MB and found no leaks; instruction `v1` through `v8` hashes are
  unchanged.
- NOT_APPLICABLE: browser, live-integration, deployment, runtime, and
  production suites; this is an instruction-contract change.
- Hosted GitHub checks remain PENDING until the pull request exists.

## OUTCOME

Ordinary conversational replies now use natural prose without a status heading,
Next actions section, synthetic None item, profile, or Repository Memory block.
The existing status/action/profile contract remains mandatory for substantial
completions and handoffs. Current guidance is additive `v9`; `v1` through `v8`
remain frozen.

## REPOSITORY MEMORY

- Decision: updated
- Rationale: List-first output fixed horizontal table readability but not
  over-application. Small questions must not carry a ceremonial PASS report;
  substantial handoffs still need status, action, and evidence discipline.
- Artifacts: `docs/specs/0071-list-first-completion-output/SPEC.md`,
  `docs/references/rules/agent-completion-output.md`,
  `docs/CONSTITUTION.md`, `internal/instructions/versions/v9.md`
