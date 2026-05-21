---
kit_metadata_version: 1
artifact: plan
feature:
  id: "0031"
  slug: executable-verification-harness
  dir: 0031-executable-verification-harness
summary: "Implement executable verification in layers: declaration schema, kit verify, run artifacts, evidence-aware reflect, trace/replay, generated state, task bundles, eval fixtures, and documentation updates."
relationships:
  - type: builds_on
    target: 0030-reference-graph-routing
  - type: builds_on
    target: 0027-implement-readiness-gate
  - type: related_to
    target: 0021-project-validation-and-instruction-registry
  - type: related_to
    target: 0028-project-refresh-advisory
references:
  - id: executable-harness-spec
    name: Executable verification harness spec
    type: feature
    target: docs/specs/0031-executable-verification-harness/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: implementation scope and acceptance criteria
    status: active
  - id: task-template
    name: Task template
    type: code
    target: internal/templates/templates.go
    selector_type: symbol
    selector: Tasks
    relation: constrains
    read_policy: must
    used_for: TASKS.md task-detail field migration
    status: active
  - id: reflect-command
    name: Reflect command
    type: code
    target: pkg/cli/reflect.go
    selector_type: symbol
    selector: buildReflectPrompt
    relation: implements
    read_policy: must
    used_for: evidence-aware reflection prompt changes
    status: active
  - id: check-command
    name: Check command
    type: code
    target: pkg/cli/check.go
    selector_type: symbol
    selector: runCheck
    relation: constrains
    read_policy: must
    used_for: structural validation and strict generated-state checks
    status: active
  - id: rlm-guidance
    name: RLM guidance
    type: doc
    target: docs/agents/RLM.md
    selector_type: artifact
    selector: RLM.md
    relation: constrains
    read_policy: must
    used_for: just-in-time context loading rules for command behavior
    status: active
---
# PLAN

## SUMMARY

Implement the executable harness as a narrow sequence of composable layers. Start by adding declarations and `kit verify`, then persist evidence, make reflection consume that evidence, add trace/replay views, generate non-authoritative JSON state and task bundles, add harness eval fixtures, and update docs/templates after behavior is proven.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-19][SPEC-20] Define task verification declarations and task bundle data structures without making JSON authoritative.
- [PLAN-02][SPEC-05][SPEC-06][SPEC-07][SPEC-10][SPEC-11] Add `kit verify` with dry-run, task scoping, conservative command parsing, execution, and JSON output.
- [PLAN-03][SPEC-08][SPEC-09][SPEC-11] Add local run artifact persistence under `.kit/runs/<run-id>/` with redacted bounded output and compact summaries.
- [PLAN-04][SPEC-12][SPEC-13] Make `kit reflect` consume latest verification evidence and surface missing or failing evidence.
- [PLAN-05][SPEC-14][SPEC-15][SPEC-16] Add `kit trace` and `kit replay` views over stored run artifacts.
- [PLAN-06][SPEC-17][SPEC-18][SPEC-20] Add generated `.kit/state.json` and state refresh/show behavior for agents and tools.
- [PLAN-07][SPEC-19][SPEC-20] Add machine-readable task bundles to generated state and command output surfaces.
- [PLAN-08][SPEC-21] Add `kit eval` fixtures for harness behavior and regression checks.
- [PLAN-09][SPEC-22][SPEC-23] Update templates, README, constitution, agent docs, checks, and tests after the command behavior is stable.
- [PLAN-10][SPEC-24][SPEC-25] Add advisory reconciliation for executable verification fields and an opt-in verification migration prompt.

## COMPONENTS

- `internal/verify/`
  - Parse task-level `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK`.
  - Normalize verification declarations into typed commands.
  - Execute commands without shell evaluation by default.
  - Produce deterministic result structs for CLI output and run artifacts.
- `internal/runstore/`
  - Create `.kit/runs/<run-id>/`.
  - Write `run.json`, bounded/redacted `stdout.txt` and `stderr.txt` files, and `summary.md`.
  - Maintain a small local index for feature/run lookup.
- `internal/state/`
  - Generate `.kit/state.json` from docs, feature map/status, task bundles, references, and latest run evidence.
  - Track source fingerprints so stale state can be detected.
- `pkg/cli/verify.go`
  - Register `kit verify`.
  - Support feature selection, `--task`, `--json`, `--dry-run`, `--no-write`, timeout, and explicit shell opt-in flags.
- `pkg/cli/trace.go`
  - Register `kit trace`.
  - Support feature-level listing and run-detail views.
- `pkg/cli/replay.go`
  - Register `kit replay`.
  - Rerun commands from a prior run and link the new artifact to the parent run.
- `pkg/cli/state.go`
  - Register state refresh/show behavior if a visible command is chosen.
  - Keep generated state pointer-only and deterministic.
- `pkg/cli/eval.go`
  - Register `kit eval`.
  - Run local fixture tests for check, verify, trace, replay, reflect evidence gating, and stale state.
- Existing docs/prompt surfaces
  - `pkg/cli/reconcile.go`
  - `pkg/cli/reconcile_audit*.go`
  - `pkg/cli/reconcile_prompt.go`
  - `pkg/cli/reflect.go`
  - `pkg/cli/implement.go`
  - `internal/templates/templates.go`
  - `internal/templates/instruction_templates*.go`
  - `docs/CONSTITUTION.md`
  - `docs/agents/*.md`
  - `README.md`

## DATA

- Verification command:
  - `id`: stable local command ID.
  - `source`: feature, task ID, and source file path.
  - `argv`: command and arguments.
  - `cwd`: default project root, optionally narrower relative directory.
  - `timeout`: optional duration.
  - `shell`: default false.
- Verification result:
  - command metadata.
  - start/end timestamps.
  - duration.
  - exit code or timeout status.
  - stdout/stderr artifact paths.
  - redaction flag.
  - pass/fail/no-declared-checks status.
- Run artifact:
  - `run_id`.
  - `parent_run_id` for replay.
  - feature ID/slug/dir.
  - task IDs covered.
  - selected commands.
  - result list.
  - changed/touched file summary when available.
  - model/agent metadata only when available from explicit inputs or environment-safe sources.
  - summary status.
- Generated state:
  - schema version.
  - generated timestamp.
  - source fingerprints.
  - feature status and task progress.
  - pointer-only references and RLM read plan data.
  - task bundles.
  - latest verification evidence per feature/task.
  - stale/healthy indicator.

## INTERFACES

- `kit verify [feature]`
  - `--task T###`
  - `--json`
  - `--dry-run`
  - `--no-write`
  - `--timeout <duration>`
  - `--allow-shell`
- `kit trace <feature-or-run-id>`
  - feature argument lists runs.
  - run ID argument shows compact run detail.
  - `--json` emits stable machine output.
- `kit replay <run-id>`
  - reruns recorded verification commands.
  - writes a new run with `parent_run_id`.
  - `--json` emits comparison output.
- `kit reflect [feature]`
  - keeps the existing prompt-output behavior.
  - adds latest verification evidence and missing-evidence warnings.
- `kit eval`
  - runs bundled local harness fixtures.
  - supports `--json` for CI.
- `kit reconcile --migrate-verification`
  - emits an advisory migration prompt only.
  - never edits feature docs directly.
  - warns active work about missing executable task fields without invalidating legacy docs.
- `TASKS.md` task detail fields:
  - `GOAL`
  - `SCOPE`
  - `ACCEPTANCE`
  - `VERIFY`
  - `EXPECTED FILES`
  - `RISK`
  - `ROLLBACK`
  - `NOTES`

## DEPENDENCIES

References are tracked in front matter.

## RISKS

- Shell command safety: mitigate by defaulting to argv execution and requiring explicit shell opt-in.
- Secret leakage in output artifacts: mitigate with bounded output, common secret redaction, no environment capture, and `--no-write`.
- Stale generated state: mitigate with source fingerprints, automatic stale warnings, and deterministic regeneration.
- Over-scoped command surface: mitigate by sequencing implementation and keeping each command narrowly tied to verification evidence.
- Legacy doc breakage: mitigate by making new fields optional under normal `kit check` and strict only when opted in.
- Verification-field churn: mitigate by warning only for active in-progress work and keeping migration prompt-only.
- RLM regression: mitigate by keeping state and trace views pointer-only and by testing that large document bodies are not inlined.
- Platform-specific command behavior: mitigate by testing command parsing/execution on Go's standard process APIs and documenting shell opt-in limits.

## TESTING

- Unit tests for task field parsing, command normalization, no-shell rejection, timeout behavior, output redaction, and run ID generation.
- Unit tests for run artifact read/write and stale generated-state detection.
- CLI tests for `kit verify --dry-run`, `kit verify --json`, `kit trace`, `kit replay`, `kit reflect`, and `kit eval`.
- CLI tests for advisory `kit reconcile` warnings and `kit reconcile --migrate-verification` prompt rules.
- Fixture tests for messy docs, incomplete tasks, stale summary/state, verify parser failures, and replay comparisons.
- Golden tests for JSON output schemas where stable ordering matters.
- End-to-end verification:
  - `go test ./...`
  - `go run ./cmd/kit check executable-verification-harness`
  - `go run ./cmd/kit check --project`
  - `go run ./cmd/kit check --all`
