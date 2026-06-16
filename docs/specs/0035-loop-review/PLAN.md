---
kit_metadata_version: 1
artifact: plan
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
  - id: spec
    name: Loop review spec
    type: feature
    target: docs/specs/0035-loop-review/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: implementation contract
    status: active
---
# PLAN

## SUMMARY

Refactor `kit loop` into a command group, keep the existing workflow runner as `kit loop workflow`, and add a new review runner that executes configured coding-agent passes against changed code until the agreed `done` contract is met.

## APPROACH

- Keep existing workflow loop functions and artifacts mostly unchanged.
- Add a separate review runner instead of forcing review semantics into the stage-based workflow loop.
- Reuse `.kit/loops/<run-id>/` artifact helpers.
- Reuse existing PR target parsing, CodeRabbit review-task extraction, and watch timing helpers.
- Keep default PR mode opportunistic; only `--watch` / `--wait-for-coderabbit` blocks for CodeRabbit completion.

## COMPONENTS

- [PLAN-COMPONENTS-01] `pkg/cli/loop.go`: command group wiring, workflow subcommand, default iteration change.
- [PLAN-COMPONENTS-02] `pkg/cli/loop_review.go`: review options, diff discovery, prompt builder, agent loop, result parser, artifact writer, and PR feedback polling.
- [PLAN-COMPONENTS-03] `pkg/cli/capabilities_catalog.go`: loop and loop review metadata.
- [PLAN-COMPONENTS-04] Docs: README, agent tooling, and init spec command reference.
- [PLAN-COMPONENTS-05] Tests: routing, parsing, base fallback, PR feedback handling, and capabilities.

## DATA

- Review run artifacts include run id, status, base ref, PR ref, correctness, stop reason, and per-iteration prompt/stdout/stderr paths.
- PR feedback is tracked by fingerprint so already-ingested feedback does not repeatedly force new passes.

## DEPENDENCIES

- Builds on the existing `0034-review-loop` PR review-thread intake and CodeRabbit filtering helpers.
- Relies on the configured loop agent in `.kit.yaml`.
- Uses Git for local changed-code discovery and `gh` only when PR mode is requested.

## INTERFACES

- `kit loop workflow [feature]`
- `kit loop review [feature]`
- `kit loop review --base <ref>`
- `kit loop review --pr <target>`
- `kit loop review --pr <target> --watch`
- `kit loop review --pr <target> --wait-for-coderabbit`

## RISKS

- Agent-reported correctness can be overstated; mitigate with prompt requirements and validation instructions.
- CodeRabbit status signals vary; default PR mode avoids hanging and strict mode uses existing timeout behavior.
- Existing users may call `kit loop [feature]`; preserve compatibility.

## TESTING

- Unit tests for command routing and parser behavior.
- Unit tests for base fallback with fake command runner.
- Unit tests for opportunistic PR feedback ingestion using existing fake review-loop seams.
- Capabilities tests for new nested command metadata.
- Full `go test ./...` verification.
