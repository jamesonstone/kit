---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0077"
  slug: "unstructured-completion-output"
  dir: "0077-unstructured-completion-output"
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
    name: Frozen density-budget instruction version
    type: code
    target: internal/instructions/versions/v14.md
    relation: informs
    read_policy: evidence
    used_for: immutable snapshot of the superseded density budget
    status: optional
  - id: instruction-v15
    name: Current instruction version
    type: code
    target: internal/instructions/versions/v15.md
    relation: implements
    read_policy: must
    used_for: immutable current kit instructions snapshot with no required response format
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Remove every required completion format and let the agent write terminal
responses in its own shape. Keep only the facts a response must not leave out,
so freeing the layout does not also free the agent from reporting blockers,
unfinished scope, unobserved checks, and mutations honestly.

## CONTEXT

- Issue #203 reported that completion output was "way too information dense"
  and hard to read, with a representative report as evidence.
- The first attempt kept the three-section envelope and bounded its density.
  It cut a sixteen-line report to five, but the shape was still imposed and
  every report still looked the same regardless of what it had to say.
- The user then asked to remove the structure altogether and rely on the
  model's own default reporting behavior. That supersedes the density budget
  before it shipped; both attempts are recorded here because the density work
  is why the format-free contract keeps the facts it keeps.
- Three successive contracts — tables in `0067`, sections in `0071`, a density
  budget in this feature's first pass — each tried to make reports scannable by
  prescribing shape, and each produced uniform, ceremonial output. The pattern
  is the evidence that shape was the wrong lever.

## REQUIREMENTS

- Impose no response format: no mandatory headings, status token, section
  order, bullet style, table ban, length budget, or empty-section declaration.
- Forbid reproducing the retired envelope and forbid substituting a new fixed
  template for it.
- Preserve, as content requirements rather than layout, the facts a reader must
  not be left wrong about: blockers, incomplete scope, required action, failing
  or unobserved checks, mutations, and confidence.
- Preserve literal provider evidence states and the prohibition on reporting
  unrun work as passing.
- Preserve the evidence inclusion test, and the merge and release reporting
  constraint that keeps a terminal result from becoming a command log.
- Rewrite every composing contract that named the retired headings so it names
  required facts instead.
- Propagate through the always-loaded gate, the generated provider instruction
  files, the Constitution, and a new frozen instruction version.

## ACCEPTED PLAN

Single-lane, because the work is tightly coupled and high-overlap: one
contract's removal propagates identically into the ruleset, the always-loaded
gate, four generated instruction files, five composing rules, the Constitution,
a frozen instruction snapshot, and string-exact tests. Parallel lanes would
collide on the same text.

1. Replace the ruleset body with a format-free content and honesty contract.
2. Compress the always-loaded gate to match and regenerate the checked-in
   provider instruction files.
3. Rewrite the passages in `github-pr-delivery`,
   `testing-and-environment-validation`, `agent-team-orchestration`,
   `cross-repository-program-coordination`, and `constitution-curation` that
   named the retired headings.
4. Publish frozen `v15`, equal to `v14` outside the completion section.
5. Teach reconcile to flag the three-section, status-token, and density-budget
   mandates as superseded guidance.
6. Validate, curate repository memory, and update the existing pull request for
   issue #203.

## DECISIONS

- **Remove format, keep content.** Dropping the honesty requirements alongside
  the layout would have been a real regression: without them nothing stops a
  report claiming success for checks that never ran. The contract now says what
  must reach the reader and never how to lay it out.
- **Forbid a replacement template explicitly.** The likeliest failure mode is
  not chaos but a new self-invented envelope applied to every task. That is
  named as an anti-pattern rather than left implicit.
- **Drop the PASS/PARTIAL/BLOCKED/FAIL taxonomy, keep the obligation.** The
  agent must still say plainly whether work is finished, partly finished,
  blocked, or failed; no token carries it.
- **Keep the merge and release reporting constraint.** It survives as content —
  report state changes and the smallest evidence proving each terminal node,
  not a polling history — because it prevents a specific observed failure and
  does not prescribe a shape.
- **Retire the proportionality gate as a named mechanism.** It existed to keep
  ceremony out of ordinary conversation. With no ceremony to gate, one line
  about matching shape to content replaces it.
- **Supersede the density budget rather than ship it.** It was correct about the
  problem and wrong about the lever. `v14` keeps it as a frozen snapshot.
- **Continue lane `GH-203` rather than open a second.** The requirement was
  revised before delivery, on the same artifact; a new lane would have produced
  a conflicting pull request and forced the first closed.
- **Do not update `.kit.yaml` registry hashes here.** Repository precedent is
  that rule-content changes ship without them and a separate reconcile chore,
  tracked by issue #202, refreshes `source_commit` and `installed_hash`.

## DISCOVERIES

- `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, and
  `docs/agents/GUARDRAILS.md` are byte-compared against the generator.
  Regenerating them is safer than hand-editing.
- `docs/references/README.md` looks generated and is not: it carries a
  hand-maintained ruleset index that regenerating silently drops.
- `internal/instructions/agent_versions_test.go` uses the next unreleased
  version string as its unavailable-version sentinel, so every bump moves it.
- The retired headings had leaked beyond the completion rule into the
  repository-memory gate in `instruction_templates_v3.go`, which required a
  concise What happened bullet for the memory decision.
- Splitting superseded-guidance detection out of
  `reconcile_guidance_expectations.go` kept both files under the 300-line limit
  on a real responsibility boundary rather than an arbitrary cut.

## VALIDATION

- `go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...`
  all pass.
- `kit check --all` passes for 74 features and `kit check --project` reports a
  coherent contract, including the source-file-size gate.
- The ruleset, the gate, and frozen `v15` are each locked by assertions,
  including forbidden-string checks for the three-section headings, the status
  token, the density budget, and the None bullet.
- `v15` is byte-identical to `v14` outside the agent completion output section.

## OUTCOME

Terminal responses have no required shape. The agent writes what the situation
calls for, and a short list of facts — blockers, unfinished scope, required
action, unobserved checks, mutations, confidence — must survive whatever shape
it chooses. Composing delivery, validation, orchestration, and program
contracts still bind on content and no longer dictate layout.

## REPOSITORY MEMORY

- Rationale for removing format, for keeping the content and honesty
  requirements, and for the superseded density budget lives in this spec.
- The Constitution's completion-output line now states that no format is
  required and names the facts that must not go missing.
- No reusable practice or domain documentation changed.
