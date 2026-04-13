# PLAN

## SUMMARY

- Centralize instruction-model metadata in one internal registry, route current consumers through it, and extend `kit check` with project-scoped validation that reuses repo-audit findings.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02] Add a shared internal instruction-contract registry package for version detection and repo-doc metadata.
- [PLAN-02][SPEC-02] Refactor current consumers to use the shared registry instead of duplicating hardcoded path sets.
- [PLAN-03][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08] Extend `kit check` with `--project` and reuse the repo-audit engine for project-scoped validation output.
- [PLAN-04][SPEC-09][SPEC-10] Tighten the shipped subagent guidance in shared prompt suffixes and repo-local docs so RLM and dispatch stay distinct.
- [PLAN-05] Add focused tests and verification for registry use, project validation, and subagent guidance wording.

## COMPONENTS

- `internal/instructions/*`
  - shared instruction-model metadata
  - version detection
  - repo-doc inventory helpers
- `internal/templates/instruction_templates*.go`
  - consume shared instruction metadata for support-file scaffolding
- `pkg/cli/check.go`
  - new `--project` validation mode
- `pkg/cli/reconcile_audit.go`
  - reusable repo-audit findings consumed by validation
- `pkg/cli/repo_docs.go`
  - repo-doc helpers backed by shared metadata
- `pkg/cli/spec_context.go`
  - shared metadata for repo-doc routing
- `internal/feature/map.go`
  - shared metadata for global instruction-doc inventory
- `pkg/cli/subagents.go`
  - tightened shared subagent guidance

## DATA

- No new persisted state.
- Project validation derives current-state findings from repo files and the existing config.

## INTERFACES

- `kit check --project`
- internal instruction-contract registry API consumed by CLI and feature packages

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| reconcile audit engine | code | `pkg/cli/reconcile_audit.go` | repo-level validation findings | active |
| versioned scaffold templates | code | `internal/templates/instruction_templates*.go` | source content for scaffolded instruction artifacts | active |
| current prompt helpers | code | `pkg/cli/repo_docs.go`, `pkg/cli/spec_context.go` | current metadata consumers to migrate | active |

## RISKS

- Registry extraction can accidentally change path ordering or labels that current prompt tests rely on.
- Project validation can become noisy if repo-audit warnings are not presented clearly.
- Subagent wording changes can unintentionally conflict with `kit dispatch` if RLM and execution planning are not separated precisely.

## TESTING

- Add unit tests for shared instruction detection and metadata lookup.
- Add CLI tests for `kit check --project` success and failure cases.
- Update map and repo-doc tests if registry-backed ordering changes.
- Run:
  - `go test ./internal/instructions ./internal/feature ./pkg/cli`
  - `go test ./...`
  - `make build`
