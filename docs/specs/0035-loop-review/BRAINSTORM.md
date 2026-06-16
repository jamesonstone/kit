---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0035
  slug: loop-review
  dir: 0035-loop-review
relationships:
  - type: builds_on
    target: 0034-review-loop
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: command mutation boundaries, explicit state, and document-first traceability
    status: active
  - id: workflow-rules
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: public command change workflow and docs-before-code sequencing
    status: active
  - id: command-capabilities
    name: Command capabilities rule
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    selector_type: artifact
    selector: command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: loop command metadata updates
    status: active
  - id: existing-loop
    name: Existing loop command
    type: code
    target: pkg/cli/loop.go
    selector_type: symbol
    selector: loopCmd
    relation: informs
    read_policy: must
    used_for: current feature workflow loop and artifact model
    status: active
  - id: existing-review-loop
    name: Existing review-loop command
    type: feature
    target: docs/specs/0034-review-loop/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: must
    used_for: current PR review-thread intake and CodeRabbit filtering contract
    status: active
---
# BRAINSTORM

## SUMMARY

Add `kit loop review` as a coding-agent repair loop that reviews changes not in the remote mainline, repeats local fix passes until the agent reports at least 95% correctness, and optionally ingests CodeRabbit PR feedback without blocking local progress by default.

## USER THESIS

The review loop should behave like a practical correctness harness: start immediately, inspect changed code, fix high, medium, and correctness-impacting issues, validate, and stop only when the coding agent emits a dense summary ending in `done`.

## RELATIONSHIPS

- builds on: 0034-review-loop
- related to: 0033-kit-capabilities

## CODEBASE FINDINGS

- `pkg/cli/loop.go` already runs configured local agents, records `.kit/loops/<run-id>/` artifacts, and enforces a confidence threshold.
- `pkg/cli/review_loop.go` currently prepares dispatch prompts from PR review threads; it is read-only prompt preparation, not an agent repair loop.
- `pkg/cli/dispatch_pr.go` already resolves PR references, fetches review threads, filters current unresolved non-outdated CodeRabbit comments, extracts agent prompt blocks, and dedupes tasks.
- `pkg/cli/review_loop_watch.go` already has CodeRabbit check-status timing constants and a fakeable clock seam.
- `pkg/cli/capabilities_catalog.go` must change with the public `loop review` command surface.

## AFFECTED FILES

- `pkg/cli/loop.go`
- `pkg/cli/loop_review.go`
- `pkg/cli/loop_review_test.go`
- `pkg/cli/review_loop.go`
- `pkg/cli/capabilities_catalog.go`
- `pkg/cli/capabilities_test.go`
- `pkg/cli/root_help.go`
- `README.md`
- `docs/agents/TOOLING.md`
- `docs/specs/0000_INIT_PROJECT.md`

## DEPENDENCIES

- Existing configured loop agent from `.kit.yaml`.
- Existing Git and GitHub CLI commands for diff and PR feedback discovery.
- Existing CodeRabbit review-thread extraction from `0034-review-loop`.

## QUESTIONS

none

## OPTIONS

- Put review under `kit loop review`: chosen because the behavior is an actual agent loop.
- Extend root `kit review-loop`: rejected because that command already means PR prompt preparation.
- Add a separate daemon or scheduler: rejected because Kit favors explicit filesystem state and foreground commands.

## RECOMMENDED STRATEGY

Make `kit loop` a parent command. Keep legacy `kit loop [feature]` compatible, add `kit loop workflow [feature]` for the existing feature workflow loop, and add `kit loop review` for changed-code correctness repair. Reuse the existing PR review-thread collector for CodeRabbit feedback and the existing `.kit/loops` artifact pattern for auditability.

## NEXT STEP

Implement the command refactor, review runner, docs, capability metadata, and focused tests.
