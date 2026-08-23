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
    name: Additive current instruction version
    type: code
    target: internal/instructions/versions/v8.md
    relation: implements
    read_policy: must
    used_for: immutable current kit instructions snapshot
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make terminal coding-agent completion output easy to scan and act on. Keep
the status-first envelope, but stop requiring a Markdown pipe table for
operator actions. Use a prioritized left-aligned list so the next step is
obvious without fighting renderer table layout.

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
- Instruction versions are immutable. `v7` is current on `main` and already
  contains the table contract as part of its frozen snapshot. The new
  current contract must be additive `v8`. Versions `v1` through `v7` remain
  byte-for-byte unchanged.
- Kit distributes the contract through the canonical ruleset, the shared
  instruction gate, checked-in AGENTS/CLAUDE/Copilot/GUARDRAILS mirrors,
  Constitution baseline text, the references index, health/reconcile
  expectations, and focused tests.
- Issue #174, branch `GH-174`, and worktree
  `~/worktrees/jamesonstone/kit/GH-174` are the authorized delivery lane.
  Merge is not authorized.

## REQUIREMENTS

### Operator Action List

- Keep the first human-readable line
  `# <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>`.
- Immediately after that heading, emit a prioritized bullet list titled
  `## Next actions`.
- Order items `Blocker`, `Incomplete`, `Next`, `Optional`, then `None`.
- Omit types that do not apply, except every PASS includes one `None` item.
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
- Add immutable instruction version `v8` as current. Do not mutate `v7` or
  earlier.
- Do not change CLI command behavior, flags, or JSON schemas.

### Non-Goals

- Do not merge this pull request.
- Do not rewrite historical `0067` outcome text into a claim that tables
  remain the current contract.
- Do not add a JSON schema or CLI flag for completion output.

## ACCEPTED PLAN

1. Record this accepted plan in `0071` before implementation.
2. Rewrite `docs/references/rules/agent-completion-output.md` so the
   operator action section is a prioritized list, and replace table
   examples with list examples.
3. Update the shared instruction gate and every generated/checked-in
   mirror, Constitution baseline, and references index sentence that still
   requires an action table.
4. Copy `v7.md` to `v8.md`, rewrite only the Agent completion output
   section, register `v8` as current, and keep `v1` through `v7` hashes
   unchanged.
5. Point focused ruleset, template, reconcile, safety-propagation, and
   instruction-version tests at the list-first contract. Keep `v6`/`v7`
   snapshot tests asserting their frozen table language.
6. Note in `0067` that the action-table decision is superseded for current
   guidance.
7. Validate, curate repository memory, and open one ready pull request for
   issue #174.

## DECISIONS

- Replace the operator action table rather than keeping it as a prefix and
  appending a superseding note. Frozen `v7` already contains the table
  language; repeating it in `v8` would tell agents to emit tables.
- Keep the four action fields (type, action, why, continue-with) as list
  structure instead of dropping why or continue-with. Readability comes
  from vertical scanning, not from dropping the copy-ready prompt.
- Keep PASS's explicit None item. The original table existed partly so
  completion could not be confused with omitted follow-up reporting.
- Publish additive `v8` instead of mutating `v7`. `v8` equals `v7` except
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

## VALIDATION

- PASS: `gofmt` on changed Go files; `go vet ./...`; `go test ./... -count=1`;
  focused `go test ./internal/templates ./internal/instructions ./pkg/cli -count=1`;
  `go test -race ./internal/instructions ./internal/templates ./pkg/cli -count=1`.
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` reported 0 issues.
- PASS: `git diff --check`.
- PASS: `gitleaks dir --no-banner --redact .` found no leaks.
- PASS: `kit check --project` using this branch's built `./bin/kit`.
- NOT_APPLICABLE: browser, end-to-end, live-integration, and production
  suites; this change is instruction-contract and unit-test scoped.
- Hosted GitHub checks remain PENDING until the pull request exists.

## OUTCOME

Terminal completion output now uses a status heading plus a prioritized
action list. Frozen `v6` and `v7` instruction snapshots still contain the
historical table contract. Current guidance is additive `v8`.

## REPOSITORY MEMORY

- Decision: updated
- Rationale: The operator-action table is harder to scan in chat renderers
  than a prioritized list. Keep status-first semantics, explicit PASS None,
  copy-ready Continue with prompts, and required evidence fields.
- Artifacts: `docs/specs/0071-list-first-completion-output/SPEC.md`,
  `docs/references/rules/agent-completion-output.md`,
  `docs/CONSTITUTION.md`, `internal/instructions/versions/v8.md`
