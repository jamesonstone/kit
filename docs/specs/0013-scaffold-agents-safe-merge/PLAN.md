# PLAN

## SUMMARY

- Add a safe overwrite confirmation gate and a deterministic append-only merge mode to `kit scaffold-agents`.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Record the scaffold safety contract in a dedicated feature spec before changing code.
- [PLAN-02][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14] Extend scaffold write handling to support explicit write modes, overwrite confirmation, and append-only preflight planning.
- [PLAN-03][SPEC-15][SPEC-16] Add a deterministic instruction-file section merge helper for append-only mode.
- [PLAN-04][SPEC-17][SPEC-18] Update command wiring, help text, and post-run guidance for the new flags and safer suggestions.
- [PLAN-05][SPEC-19][SPEC-20] Add tests for overwrite confirmation, append-only success/failure behavior, targeted selection, and flag validation.
- [PLAN-06][SPEC-21] Update shipped docs for the new scaffold-agents semantics and rerun verification.

## COMPONENTS

- `pkg/cli/scaffold_agents.go`
  - new flags
  - overwrite confirmation flow
  - append-only mode selection
  - safer completion guidance
- `pkg/cli/instruction_files.go`
  - write-mode modeling
  - append-only preflight planning
  - write result reporting
- `pkg/cli/instruction_file_merge.go`
  - instruction markdown section parsing
  - deterministic append-only merge planning
- `pkg/cli/instruction_files_test.go`
  - confirmation tests
  - append-only merge tests
  - flag-validation tests
- `README.md`
  - command behavior and flag docs
- `docs/specs/0000_INIT_PROJECT.md`
  - shipped scaffold-agents behavior summary

## DATA

- Input data comes from:
  - selected scaffold target paths
  - existing file contents when present
  - Kit instruction templates from `internal/templates`
- No new persistent state or artifact type is introduced.
- Append-only mode uses only top-level `##` markdown section anchors from the scaffolded instruction templates.

## INTERFACES

- Existing command:
  - `kit scaffold-agents`
  - alias: `kit scaffold-agent`
- New flags:
  - `--yes` / `-y`
  - `--append-only`
- Existing flags remain:
  - `--force`
  - `--agentsmd`
  - `--claude`
  - `--copilot`
- Mode rules:
  - default: create missing, skip existing, suggest append-only/force when skipped
  - `--force`: overwrite existing after confirmation when needed
  - `--force --yes`: overwrite existing without confirmation
  - `--append-only`: merge missing Kit-managed sections without overwriting matched existing sections

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical scaffold behavior and write-mode rules | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | shipped scaffold-agents behavior summary | active |
| scaffold agents command | code | `pkg/cli/scaffold_agents.go` | overwrite confirmation and append-only flow | active |
| instruction-file merge helpers | code | `pkg/cli/instruction_files.go`, `pkg/cli/instruction_file_merge.go` | deterministic append-only behavior | active |
| README | doc | `README.md` | user-facing command and flag docs | active |

## RISKS

- Append-only merge rules can become ambiguous if heading matching is too loose.
- Overwrite confirmation can break automation if no explicit bypass exists.
- Partial writes would be dangerous if append-only validation happens after mutation instead of before.
- Merge behavior can surprise users if extra sections are reordered incorrectly.
- Flag combinations can become confusing without clear validation errors.

## TESTING

- Add unit tests for:
  - overwrite confirmation accept/cancel behavior
  - `--force --yes` bypass behavior
  - append-only merge with missing template sections
  - append-only preservation of existing matched sections
  - append-only preservation of extra user sections
  - append-only failure when no recognizable anchors exist
  - append-only failure when duplicate recognized headings exist
  - flag validation for `--append-only --force` and `--yes` without `--force`
  - targeted selection behavior in append-only mode
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
