---
kit_metadata_version: 1
artifact: tasks
feature:
  id: "0031"
  slug: executable-verification-harness
  dir: 0031-executable-verification-harness
summary: Task plan for implementing executable verification, run evidence, trace/replay, generated state, task bundles, eval fixtures, and docs updates.
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
  - id: executable-harness-plan
    name: Executable verification harness plan
    type: feature
    target: docs/specs/0031-executable-verification-harness/PLAN.md
    selector_type: artifact
    selector: PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation task sequencing
    status: active
  - id: executable-harness-spec
    name: Executable verification harness spec
    type: feature
    target: docs/specs/0031-executable-verification-harness/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: requirements and acceptance criteria
    status: active
---
# TASKS

## PROGRESS TABLE

| ID   | TASK                                                   | STATUS | OWNER | DEPENDENCIES |
| ---- | ------------------------------------------------------ | ------ | ----- | ------------ |
| T001 | Record executable verification feature docs            | done   | agent |              |
| T002 | Add verification declaration parsing and task bundles  | done   | agent | T001         |
| T003 | Implement `kit verify` execution and JSON output       | done   | agent | T002         |
| T004 | Persist local run artifacts under `.kit/runs/`         | done   | agent | T003         |
| T005 | Make `kit reflect` evidence-aware                     | done   | agent | T004         |
| T006 | Add `kit trace` and `kit replay`                       | done   | agent | T004         |
| T007 | Add generated `.kit/state.json`                        | done   | agent | T002, T004   |
| T008 | Add `kit eval` harness fixtures                        | done   | agent | T003, T005   |
| T009 | Update templates, prompts, docs, and validation        | done   | agent | T002-T008    |
| T010 | Run full verification and reconcile project summary    | done   | agent | T009         |
| T011 | Add executable verification reconcile advisory         | done   | agent | T009         |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record executable verification feature docs [PLAN-01] [PLAN-09]
- [x] T002: Add verification declaration parsing and task bundles [PLAN-01] [PLAN-07]
- [x] T003: Implement `kit verify` execution and JSON output [PLAN-02]
- [x] T004: Persist local run artifacts under `.kit/runs/` [PLAN-03]
- [x] T005: Make `kit reflect` evidence-aware [PLAN-04]
- [x] T006: Add `kit trace` and `kit replay` [PLAN-05]
- [x] T007: Add generated `.kit/state.json` [PLAN-06]
- [x] T008: Add `kit eval` harness fixtures [PLAN-08]
- [x] T009: Update templates, prompts, docs, and validation [PLAN-09]
- [x] T010: Run full verification and reconcile project summary [PLAN-09]
- [x] T011: Add executable verification reconcile advisory [PLAN-10]

## TASK DETAILS

For each task, provide:

### T001

- **GOAL**: Capture the approved executable harness refactor before implementation.
- **SCOPE**:
  - create `docs/specs/0031-executable-verification-harness/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
  - encode the implementation order from verification to evals
- **ACCEPTANCE**:
  - feature docs define `kit verify`, run artifacts, evidence-aware reflect, trace/replay, generated state, task bundles, evals, and RLM constraints
  - docs contain no placeholder-only required sections
  - `kit check executable-verification-harness` passes
- **VERIFY**:
  - `go run ./cmd/kit check executable-verification-harness`
- **EXPECTED FILES**:
  - `docs/specs/0031-executable-verification-harness/SPEC.md`
  - `docs/specs/0031-executable-verification-harness/PLAN.md`
  - `docs/specs/0031-executable-verification-harness/TASKS.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Low; docs-only, but incorrect scope here can over-expand the implementation.
- **ROLLBACK**: Remove `docs/specs/0031-executable-verification-harness/` and regenerate project summary if the feature direction is rejected.
- **NOTES**: This task is complete once these docs and summary updates are in place.

### T002

- **GOAL**: Parse executable task declarations and expose machine-readable task bundles.
- **SCOPE**:
  - add parsing for `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK`
  - normalize verification command bullets into typed command structs
  - build task bundle structs from task details and dependencies
  - keep legacy task docs valid when fields are absent
- **ACCEPTANCE**:
  - parser extracts task declarations deterministically
  - invalid command declarations report actionable errors
  - task bundles include task ID, feature, dependencies, expected paths, verify commands, risk, rollback, and handoff fields
  - existing task parsing behavior remains compatible
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit check executable-verification-harness`
- **EXPECTED FILES**:
  - `internal/verify/`
  - `pkg/cli/verify.go`
- **RISK**: Medium; task parsing touches a shared document contract.
- **ROLLBACK**: Revert parser and bundle structs while leaving docs intact for a narrower follow-up.
- **NOTES**: Keep parsing explicit and avoid a general-purpose Markdown interpreter.

### T003

- **GOAL**: Add `kit verify` with dry-run, scoped execution, and structured JSON.
- **SCOPE**:
  - register `kit verify [feature]`
  - support `--task`, `--json`, `--dry-run`, `--timeout`, `--no-write`, and `--allow-shell`
  - execute commands without shell evaluation by default
  - return stable command result JSON
- **ACCEPTANCE**:
  - `kit verify <feature> --dry-run` lists selected commands without executing them
  - `kit verify <feature> --task T### --json` emits schema-valid JSON
  - commands with shell syntax fail unless shell opt-in is explicit
  - no declared checks are reported clearly and do not masquerade as passing tests
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit verify executable-verification-harness --task T001 --dry-run`
  - `go run ./cmd/kit verify executable-verification-harness --task T001 --json`
- **EXPECTED FILES**:
  - `pkg/cli/verify.go`
  - `internal/verify/`
  - `pkg/cli/root_help.go`
- **RISK**: Medium-high; command execution must be safe and predictable.
- **ROLLBACK**: Remove the command registration and keep declaration parsing for a later implementation.
- **NOTES**: Prefer Go process execution over shell invocation.

### T004

- **GOAL**: Persist verification evidence as local run artifacts.
- **SCOPE**:
  - create `.kit/runs/<run-id>/`
  - write `run.json`
  - write bounded and redacted stdout/stderr artifacts
  - write compact `summary.md`
  - maintain a local run index for feature/run lookup
- **ACCEPTANCE**:
  - each executed verification run produces inspectable local artifacts unless `--no-write` is set
  - artifact JSON includes feature, task IDs, commands, outcomes, durations, output paths, and redaction metadata
  - output capture is bounded and redacted for common secret patterns
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit verify executable-verification-harness --task T001 --json`
- **EXPECTED FILES**:
  - `internal/runstore/`
  - `pkg/cli/verify.go`
  - `.gitignore`
- **RISK**: High; storing execution output creates privacy and stale-artifact risks.
- **ROLLBACK**: Keep `kit verify --json` and disable persistence behind `--write` until storage is corrected.
- **NOTES**: Run artifacts are local evidence, not canonical requirements.

### T005

- **GOAL**: Make reflection evidence-backed.
- **SCOPE**:
  - load latest relevant run evidence for a feature
  - include pass/fail/missing evidence in `kit reflect` output
  - require final reflection instructions to cite verification evidence before marking complete
- **ACCEPTANCE**:
  - `kit reflect <feature>` surfaces latest verification status
  - missing declared verification evidence is visible and actionable
  - existing prompt-only behavior and clipboard/output flags remain intact
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit reflect executable-verification-harness --output-only`
- **EXPECTED FILES**:
  - `pkg/cli/reflect.go`
  - `pkg/cli/testdata/reflect_feature_prompt.golden`
- **RISK**: Medium; reflection should not become noisy or silently block legitimate docs-only work.
- **ROLLBACK**: Remove run-evidence injection from reflect while keeping runstore APIs.
- **NOTES**: Keep evidence concise and pointer-based.

### T006

- **GOAL**: Add trace and replay commands over persisted verification runs.
- **SCOPE**:
  - add `kit trace <feature-or-run-id>`
  - add `kit replay <run-id>`
  - support `--json` for both commands
  - link replay runs back to parent run IDs
- **ACCEPTANCE**:
  - trace lists feature runs with timestamps, status, task IDs, commands, and artifact paths
  - trace run-detail view avoids inlining large stdout/stderr
  - replay reruns recorded commands and writes a linked run
  - replay docs clearly avoid claiming deterministic model reconstruction
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit trace executable-verification-harness --json`
- **EXPECTED FILES**:
  - `pkg/cli/trace.go`
  - `pkg/cli/replay.go`
  - `internal/runstore/`
- **RISK**: Medium; replay can be misunderstood as agent-state reconstruction.
- **ROLLBACK**: Keep trace listing and defer replay to a follow-up feature.
- **NOTES**: Replay is command evidence replay only.

### T007

- **GOAL**: Generate agent-readable state without making JSON authoritative.
- **SCOPE**:
  - define `.kit/state.json` schema
  - generate state from docs, feature map/status, task bundles, references, and latest run evidence
  - record source fingerprints for stale detection
  - add visible refresh/show behavior
- **ACCEPTANCE**:
  - generated state is deterministic
  - stale state can be detected from source fingerprints
  - state includes task bundles and latest verification status
  - state does not inline full `SPEC.md`, `PLAN.md`, or `TASKS.md` bodies
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit check --project`
- **EXPECTED FILES**:
  - `internal/state/`
  - `pkg/cli/state.go`
  - `.gitignore`
- **RISK**: High; generated JSON can accidentally become a second source of truth.
- **ROLLBACK**: Remove state command/output and continue using runstore plus map/status JSON.
- **NOTES**: Every state consumer must be able to regenerate from Markdown and run artifacts.

### T008

- **GOAL**: Add local harness eval fixtures for Kit itself.
- **SCOPE**:
  - register `kit eval`
  - add fixtures for messy docs, incomplete tasks, stale summary/state, verify parsing, trace/replay shape, and reflect evidence gating
  - support JSON output for CI
- **ACCEPTANCE**:
  - `kit eval` passes when fixtures match expected decisions
  - `kit eval` fails clearly when fixture output drifts
  - fixture docs remain small and purpose-specific
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit eval --json`
- **EXPECTED FILES**:
  - `pkg/cli/eval.go`
  - `internal/eval/`
  - `internal/verify/tasks_test.go`
  - `internal/verify/execute_test.go`
  - `internal/runstore/store_test.go`
- **RISK**: Medium; evals can become expensive or too coupled to prompt wording.
- **ROLLBACK**: Keep unit tests and defer public `kit eval`.
- **NOTES**: Prefer schema/behavior checks over brittle prose snapshots.

### T009

- **GOAL**: Update docs, templates, prompts, and validation for the shipped executable harness contract.
- **SCOPE**:
  - update task templates with new fields
  - update implement/reflect/handoff guidance to use declared verification and evidence
  - update README command tables
  - update `docs/CONSTITUTION.md` for visible local run/state files
  - update agent docs to preserve RLM and avoid whole-document loading
  - update `kit check --project` where needed
- **ACCEPTANCE**:
  - generated tasks include executable fields
  - README and help list the new commands
  - constitution explains generated local evidence/state without making it authoritative
  - project checks validate the new contract
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit check --project`
  - `go run ./cmd/kit check --all`
- **EXPECTED FILES**:
  - `internal/templates/`
  - `pkg/cli/implement.go`
  - `pkg/cli/reflect.go`
  - `pkg/cli/root_help.go`
  - `README.md`
  - `docs/CONSTITUTION.md`
  - `docs/agents/`
- **RISK**: Medium-high; broad docs/template changes can drift from implementation reality.
- **ROLLBACK**: Revert docs/templates for unreleased command surfaces and keep implementation internal until complete.
- **NOTES**: Update docs after behavior lands, not before.

### T010

- **GOAL**: Prove the feature and reconcile generated project state.
- **SCOPE**:
  - run focused and full verification
  - regenerate project progress summary
  - inspect final diffs for scope creep
  - ensure all completed tasks have evidence
- **ACCEPTANCE**:
  - all declared checks pass or failures are documented with exact blockers
  - `docs/PROJECT_PROGRESS_SUMMARY.md` reflects this feature's current phase and task progress
  - no unrelated files are changed
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit check executable-verification-harness`
  - `go run ./cmd/kit check --project`
  - `go run ./cmd/kit check --all`
- **EXPECTED FILES**:
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
  - files touched by earlier tasks
- **RISK**: Low; final verification may expose earlier task drift.
- **ROLLBACK**: Fix failing task-specific work before marking the feature ready for reflection.
- **NOTES**: Do not mark additional tasks done unless their `VERIFY` evidence exists.

### T011

- **GOAL**: Add advisory reconciliation for executable verification fields without creating legacy-doc churn.
- **SCOPE**:
  - warn active in-progress features when task details lack executable verification fields
  - leave completed, historical, and legacy feature docs compatible by default
  - add `kit reconcile --migrate-verification` as a prompt-only migration path
  - ensure migration guidance forbids guessing commands from prose and separates proposed checks from confirmed checks
- **ACCEPTANCE**:
  - normal `kit check` remains non-strict for legacy task docs
  - `kit reconcile` emits warning-level findings for active task details missing `VERIFY`, `EXPECTED FILES`, `RISK`, or `ROLLBACK`
  - `kit reconcile --migrate-verification` outputs advisory prompt rules and does not edit files directly
  - migration prompt asks agents to run `kit verify <feature> --dry-run`, refresh `.kit/state.json`, then rerun `kit check <feature>` and `kit check --project`
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit reconcile executable-verification-harness --migrate-verification --output-only`
  - `go run ./cmd/kit check executable-verification-harness`
  - `go run ./cmd/kit check --project`
- **EXPECTED FILES**:
  - `pkg/cli/reconcile.go`
  - `pkg/cli/reconcile_audit.go`
  - `pkg/cli/reconcile_audit_helpers.go`
  - `pkg/cli/reconcile_prompt.go`
  - `pkg/cli/reconcile_test.go`
  - `docs/specs/0031-executable-verification-harness/SPEC.md`
  - `docs/specs/0031-executable-verification-harness/PLAN.md`
  - `docs/specs/0031-executable-verification-harness/TASKS.md`
- **RISK**: Medium; reconcile must guide migration without making old docs fail or encouraging invented commands.
- **ROLLBACK**: Remove the advisory audit and migration flag while keeping executable verification core behavior intact.
- **NOTES**: Keep this warning-only until a future strict policy is explicitly requested.

## DEPENDENCIES

- T001 must land first because it establishes the implementation contract.
- T002 must land before commands that consume task declarations.
- T003 must land before run persistence, reflect evidence, trace/replay, generated state, and eval fixtures.
- T004 must land before evidence-aware reflect and trace/replay.
- T009 should wait until command behavior is stable enough to document accurately.
- T010 is the final verification and reconciliation pass.
- T011 follows the shipped verification behavior and keeps migration advisory by default.

## NOTES

- RLM rule for implementers: start from the selected task detail, then load only linked `PLAN.md` and `SPEC.md` sections plus front matter references that materially affect the immediate decision.
- JSON state and task bundles exist to reduce repetitive Markdown parsing by agents and tools; they must never outrank the Markdown artifacts.
- Run artifacts are evidence and should be local, inspectable, bounded, redacted, and removable.
