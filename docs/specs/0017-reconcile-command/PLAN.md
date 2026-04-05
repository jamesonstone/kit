# PLAN

## SUMMARY

- Add a prompt-only `reconcile` command that audits Kit-managed project documents for contract drift and emits an agent prompt to repair those docs.
- Keep v1 read-only, whole-project by default, and feature-scoped when explicitly requested.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Add a new `pkg/cli/reconcile.go` command with project-wide default behavior, optional feature scoping, `--all`, and the shared prompt-output flags.
- [PLAN-02][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16][SPEC-17][SPEC-18][SPEC-19] Implement reconciliation audit helpers that inspect Kit-managed docs for missing sections, placeholder-only required content, malformed required tables, safe structural truncation, and bounded semantic drift.
- [PLAN-03][SPEC-20][SPEC-21][SPEC-22][SPEC-23] Add cross-document consistency checks for task alignment, relationship targets, rollup presence, and instruction-file drift.
- [PLAN-04][SPEC-24][SPEC-25][SPEC-26][SPEC-27][SPEC-28] Build a reconciliation prompt that groups findings, cites canonical contract sources, prescribes update actions, and emits concise deduplicated search guidance plus a compact response contract.
- [PLAN-05][SPEC-29][SPEC-30][SPEC-31] Integrate the command into root help and README with wording that distinguishes it from validation, catch-up, handoff, and instruction scaffolding.
- [PLAN-06][SPEC-06][SPEC-12][SPEC-17][SPEC-18][SPEC-20][SPEC-21][SPEC-23][SPEC-27][SPEC-28] Add focused tests for audit findings, clean-project behavior, prompt generation, and flag handling, then run normal verification commands.
- [PLAN-07] Add a human-readable terminal summary for non-`--output-only` reconcile runs without changing the raw prompt payload.
- [PLAN-08] Keep the compact prompt aligned with default orchestration by explicitly telling the coding agent to use subagents and queue work according to overlapping file changes, while omitting that line under `--single-agent`.

## COMPONENTS

- `pkg/cli/reconcile.go`
  - command registration
  - scope resolution
  - clean-result handling
- `pkg/cli/reconcile_audit.go`
  - document audit helpers
  - table and task-structure checks
  - rollup and instruction-file drift checks
- `pkg/cli/reconcile_prompt.go`
  - compact prompt builder
  - grouped file summaries
  - deduplicated search-guidance rendering
  - default subagent-overlap instruction
- `pkg/cli/reconcile_summary.go`
  - human-readable terminal summary rendering
- `pkg/cli/reconcile_test.go`
  - command and prompt tests
- `pkg/cli/root.go`
  - help ordering
- `README.md`
  - command documentation

## DATA

- Inputs come from:
  - `docs/CONSTITUTION.md`
  - `PROJECT_PROGRESS_SUMMARY.md`
  - feature docs under `docs/specs/`
  - embedded templates in `internal/templates/templates.go`
  - instruction-file write planning in `pkg/cli/instruction_files.go`
- No new persisted state is introduced.
- No network or external APIs are involved.

## INTERFACES

- New command:
  - `kit reconcile [feature]`
- Flags:
  - `--all`
  - `--copy`
  - `--output-only`
  - `--prompt-only`
- Output modes:
  - concise clean-result text when no findings exist
  - clipboard-first prompt output when findings exist

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| command wiring patterns | code | `pkg/cli/handoff.go` | project-vs-feature prompt command shape | active |
| check validation | code | `pkg/cli/check.go` | base document validation rules | active |
| instruction append-only planning | code | `pkg/cli/instruction_files.go` | instruction-file drift detection | active |
| template contract | code | `internal/templates/templates.go` | current required sections and table shapes | active |
| rollup generation | code | `internal/rollup/rollup.go` | progress-summary expectations | active |

## RISKS

- The command can duplicate `check` if it reports raw validation failures without migration guidance.
- The command can duplicate `handoff` if the prompt becomes a general session-transfer workflow instead of targeted reconciliation.
- Contract detection can become brittle if it relies on exact template prose instead of structural requirements and bounded semantics.
- Rollup drift checks can become noisy if they attempt full generated-content diffing instead of focused missing-entry validation.
- Instruction-file drift guidance can mislead users if append-only planning failures are not surfaced clearly with manual fallback instructions.
- The prompt can become too long for whole-project audits if finding rendering stays per-finding instead of grouped by file and category.

## TESTING

- Add unit tests for:
  - `--all` and feature-argument validation
  - clean-project/no-findings behavior
  - compact prompt content for project-wide and feature-scoped findings
  - malformed table detection
  - task ID alignment detection
  - relationship-target detection
  - instruction-file drift detection via append-only planning
  - human-readable reconcile summary rendering
- Run:
  - `make vet`
  - `make test`
  - `make build`
  - `./bin/kit reconcile --help`
  - `./bin/kit --help`
