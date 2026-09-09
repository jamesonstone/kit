---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0077"
  slug: "completion-output-density"
  dir: "0077-completion-output-density"
relationships:
  - type: builds_on
    target: 0071-list-first-completion-output
    note: Keeps the three-section contract and status line, and bounds how much a structured report may carry inside them.
  - type: related_to
    target: 0067-agent-completion-output
    note: Continues the retreat from the original table-and-profile envelope that made completion output dense.
references:
  - id: completion-rule
    name: Canonical terminal completion contract
    type: rule
    target: docs/references/rules/agent-completion-output.md
    relation: implements
    read_policy: must
    used_for: proportionality gate, three-section shape, density budget, and evidence inclusion test
    status: active
  - id: instruction-gate
    name: Shared completion routing gate
    type: code
    target: internal/templates/agent_completion_output.go
    relation: implements
    read_policy: must
    used_for: always-loaded gate mirrored into generated and checked-in provider instructions
    status: active
  - id: instruction-v13
    name: Frozen prior instruction version
    type: code
    target: internal/instructions/versions/v13.md
    relation: informs
    read_policy: evidence
    used_for: immutable prior kit instructions snapshot
    status: optional
  - id: instruction-v14
    name: Current instruction version
    type: code
    target: internal/instructions/versions/v14.md
    relation: implements
    read_policy: must
    used_for: immutable current kit instructions snapshot carrying the density contract
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make a structured completion report readable in one pass. The three-section
contract from `0071` produced correct reports that were exhausting to read: a
wall of bold-lead bullets, each carrying several sentences of packed
identifiers, inline links, and file-and-line references. This feature keeps the
sections and the status line and constrains what may go inside them.

## CONTEXT

- Issue #203 reported that real completion output is "way too information
  dense" and hard to read, with a representative report as evidence.
- The density was traceable to specific clauses in the rule rather than to
  agent drift. Agents were following the contract as written.
- The rule's own worked examples modeled the dense style, and examples steer
  behavior more strongly than prose does.

## REQUIREMENTS

- Preserve the proportionality gate, so ordinary conversation stays free of
  completion scaffolding.
- Preserve the three canonical headings, their order, the four status values,
  and literal native evidence states.
- Bound report length, bullet count, sentence count per bullet, and nesting
  depth with concrete numbers rather than adjectives.
- Replace the maximalist evidence preference with an inclusion test that names
  what to include and what to leave out.
- Make What happened prose-first, with bullets reserved for genuinely separate
  outcomes.
- Omit empty Deviations and Next steps sections instead of emitting a `None`
  bullet, while still requiring Deviations for every PARTIAL, BLOCKED, or FAIL.
- Restrict bold to the status line, blockers, and required actions.
- Rewrite the worked examples so they model the intended brevity.
- Propagate the contract through the always-loaded gate, the checked-in
  provider instruction files, the Constitution, and a new frozen instruction
  version.

## ACCEPTED PLAN

Single-lane, because the work is tightly coupled and high-overlap: one
contract's exact wording must propagate identically into the ruleset, the
always-loaded gate, four generated instruction files, the Constitution, a
frozen instruction snapshot, and string-exact tests. Parallel lanes would
collide on the same text.

1. Diagnose which clauses in the ruleset caused the observed density.
2. Rewrite the ruleset around a density budget, an evidence inclusion test,
   prose-first What happened, and omitted empty sections.
3. Rewrite the worked examples to demonstrate the budget.
4. Compress the always-loaded gate and regenerate the checked-in provider
   instruction files from the generator.
5. Publish frozen `v14`, equal to `v13` outside the completion section.
6. Update string-exact expectations, add reconcile detection for the
   superseded `None`-bullet and nested-evidence mandates, and add a `v14`
   contract test.
7. Validate, curate repository memory, and open one ready pull request for
   issue #203.

## DECISIONS

- **Keep the status line as `**Status: <VALUE> — <sentence>.**`** It is one
  short line, it is the most scannable part of a report, and five other
  contracts compose against it. The density lived in the bullets, not here.
- **Move the status line out of a bullet onto its own line.** A one-item
  bullet list followed by a paragraph reads badly, and prose is now the
  default shape for the section that follows.
- **Omit empty sections rather than emitting `- **None.**`.** The `None`
  bullet was meant to stop an agent silently dropping the section that would
  have held a blocker, but an agent willing to hide a blocker will write
  "None." just as readily. The guard was weak and the noise was constant. The
  requirement that PARTIAL, BLOCKED, and FAIL always carry Deviations replaces
  it with a check that actually binds.
- **Express limits as numbers.** "Concise" had already failed. Twelve rendered
  lines, five bullets, one sentence plus one evidence clause, and one nesting
  level are arguable, but they are checkable and they bind.
- **Invert the evidence default.** The prior rule preferred exact identifiers
  "over vague claims," which reads as an instruction to include everything
  verified. The replacement asks whether the reader needs the identifier to
  act, or would doubt the claim without it.
- **Do not update `.kit.yaml` registry hashes here.** Precedent in this
  repository is that rule-content changes ship without touching the registry,
  and a separate reconcile-maintenance chore refreshes `source_commit` and
  `installed_hash`. Issue #202 already tracks that work.
- **Leave `docs/references/README.md` alone.** It is a curated superset of the
  generator output, not a generated file; regenerating it silently drops the
  hand-maintained ruleset index.

## DISCOVERIES

- `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, and
  `docs/agents/GUARDRAILS.md` are byte-compared against the generator by
  `TestCheckedInCapabilityAdapterMatchesGeneratedArtifacts` and
  `TestMemoryCopilotInstructionsPreserveMutationRouting`. Hand-editing them is
  viable only when the edit reproduces generator output exactly; regenerating
  is the safer path.
- `docs/references/README.md` looks generated and is not. It is written by the
  same `InstructionSupportFiles` call but the checked-in copy carries an extra
  hand-maintained ruleset index, so regenerating it loses content.
- `internal/instructions/agent_versions_test.go` uses the next unreleased
  version string as its "unavailable version" sentinel, so every version bump
  must move that sentinel forward as well.

## VALIDATION

- `go build ./...` succeeds.
- `go vet ./...` reports nothing.
- `go test ./...` passes.
- `go test -race ./...` passes.
- The rewritten ruleset, the compressed gate, and the frozen `v14` snapshot are
  each locked by their own assertions, including forbidden-string checks for
  the superseded `None`-bullet and nested-evidence mandates.
- `v14` is byte-identical to `v13` outside the agent completion output section,
  asserted by `TestAgentInstructionsV14PreservesV13OutsideCompletionSection`.

## OUTCOME

Structured reports keep their three sections, their status line, and every
composing contract's required facts, and now have a stated ceiling on length,
bullet count, sentence count, nesting, and evidence. Reports with nothing to
flag and nothing to hand off end after What happened.

## REPOSITORY MEMORY

- Feature rationale for the density limits and the `None`-bullet removal lives
  in this spec.
- The Constitution gained the omission-and-brevity clause on its existing
  completion-output line; no new invariant was introduced.
- No reusable practice or domain documentation changed.
