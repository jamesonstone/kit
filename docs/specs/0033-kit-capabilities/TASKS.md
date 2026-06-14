---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0033
  slug: kit-capabilities
  dir: 0033-kit-capabilities
summary: Task plan for implementing the read-only kit capabilities command, capability catalog, drift tests, documentation, and verification.
relationships:
  - type: builds_on
    target: 0019-command-surface-simplification
  - type: builds_on
    target: 0030-reference-graph-routing
  - type: related_to
    target: 0016-document-map-relationships
  - type: related_to
    target: 0020-versioned-instruction-model
  - type: related_to
    target: 0021-project-validation-and-instruction-registry
skills:
  - name: rlm
    source: repo-local
    path: docs/agents/RLM.md
    trigger: analyze codebase; scan repository; large repository analysis; recursive language model context routing
    required: true
references:
  - id: kit-capabilities-spec
    name: Kit capabilities spec
    type: feature
    target: docs/specs/0033-kit-capabilities/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding requirements, acceptance criteria, edge cases, and non-goals
    status: active
  - id: kit-capabilities-plan
    name: Kit capabilities plan
    type: feature
    target: docs/specs/0033-kit-capabilities/PLAN.md
    selector_type: artifact
    selector: PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation task sequencing, component boundaries, interfaces, risks, and verification evidence mapping
    status: active
  - id: kit-capabilities-brainstorm
    name: Kit capabilities brainstorm
    type: feature
    target: docs/specs/0033-kit-capabilities/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: conditional
    used_for: upstream research context and resolved defaults when task execution encounters ambiguity
    status: active
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: artifact pipeline order, read-only command expectations, no hidden state, code quality, and validation requirements
    status: active
  - id: feature-map-command
    name: Feature map command
    type: command
    target: "kit map 0033-kit-capabilities"
    selector_type: command
    selector: "kit map 0033-kit-capabilities"
    relation: informs
    read_policy: evidence
    used_for: feature phase, relationship, and reference resolution before implementation
    status: active
  - id: progress-summary
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    selector_type: heading
    selector: FEATURE PROGRESS TABLE
    relation: informs
    read_policy: evidence
    used_for: highest-artifact phase tracking and final summary reconciliation
    status: active
---
# TASKS

## PROGRESS TABLE

| ID   | TASK                                               | STATUS | OWNER | DEPENDENCIES |
| ---- | -------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Run readiness preflight and command inventory      | done   | agent |              |
| T002 | Add capability catalog and record projections      | done   | agent | T001         |
| T003 | Implement `kit capabilities` command modes         | done   | agent | T002         |
| T004 | Wire root help and visible command placement       | done   | agent | T003         |
| T005 | Add JSON, search, and error-path tests             | done   | agent | T003         |
| T006 | Add drift and read-only side-effect tests          | done   | agent | T004, T005   |
| T007 | Update user-facing and agent-facing documentation  | done   | agent | T006         |
| T008 | Run final verification and reconcile feature state | done   | agent | T007         |

## TASK LIST

- [x] T001: Run readiness preflight and command inventory [PLAN-APPROACH] [PLAN-RISKS]
- [x] T002: Add capability catalog and record projections [PLAN-COMPONENTS] [PLAN-DATA]
- [x] T003: Implement `kit capabilities` command modes [PLAN-APPROACH] [PLAN-INTERFACES]
- [x] T004: Wire root help and visible command placement [PLAN-COMPONENTS] [PLAN-INTERFACES]
- [x] T005: Add JSON, search, and error-path tests [PLAN-TESTING]
- [x] T006: Add drift and read-only side-effect tests [PLAN-RISKS] [PLAN-TESTING]
- [x] T007: Update user-facing and agent-facing documentation [PLAN-APPROACH] [PLAN-COMPONENTS]
- [x] T008: Run final verification and reconcile feature state [PLAN-TESTING]

## TASK DETAILS

### T001

- **GOAL**: Confirm the implementation can start without contradictions, hidden assumptions, or stale command-surface facts.
- **SCOPE**:
  - run the implementation readiness gate over `CONSTITUTION.md`, `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md`
  - inspect current root command registration, root help ordering, visible command sections, and hidden/deprecated command policy
  - decide from live Cobra registration whether `ci` is in scope as metadata only
  - record any blocking mismatch by updating feature docs before code
- **ACCEPTANCE**:
  - readiness gate finds no task gaps, scope creep, or unresolved assumptions
  - live command inventory identifies visible root commands, selected nested commands, hidden/deprecated commands, and registered `ci` status
  - no product code is changed before the readiness gate passes
- **VERIFY**:
  - `kit check 0033-kit-capabilities`
  - `kit map 0033-kit-capabilities`
  - `rg -n rootCmd cmd pkg internal docs`
  - `rg -n AddCommand cmd pkg internal docs`
  - `rg -n configureRootHelp cmd pkg internal docs`
  - `rg -n commandOrder cmd pkg internal docs`
  - `rg -n rootCommandSections cmd pkg internal docs`
- **EXPECTED FILES**:
  - no file changes expected unless the readiness gate finds documentation drift
- **RISK**: Low; preflight only, but skipping it can let implementation start from stale command facts.
- **ROLLBACK**: not required.
- **NOTES**: If readiness fails, update canonical feature docs first and do not continue to T002.

### T002

- **GOAL**: Add the static capability catalog and projection layer that all output modes use.
- **SCOPE**:
  - create catalog record types, schema version constants, compact/detail projections, deterministic sorting, lookup, search, and suggestions
  - include visible canonical root commands and required nested command paths
  - represent hidden/deprecated compatibility commands only for full or direct targeted lookup, with explicit metadata
  - include `ci` metadata if `ci` remains registered, without modifying `ci` behavior
  - keep ownership in `pkg/cli`; do not add an internal registry or third-party dependency
- **ACCEPTANCE**:
  - catalog records expose compact fields and detailed fields defined in PLAN `DATA`
  - `mutation_level` uses the planned enum values
  - safety notes distinguish default behavior from flag-dependent behavior for `verify`, `dispatch`, and registered `ci`
  - no persisted files, config fields, generated catalogs, or `.kit` artifacts are introduced
- **VERIFY**:
  - `go test ./pkg/cli`
- **EXPECTED FILES**:
  - `pkg/cli/capabilities_catalog.go`
  - `pkg/cli/capabilities_test.go`
- **RISK**: Medium; duplicated metadata can drift if record definitions are too loose.
- **ROLLBACK**: Remove the new catalog file and catalog tests.
- **NOTES**: Split catalog and command files only when it improves scanability.

### T003

- **GOAL**: Implement the visible `kit capabilities` command and its compact, targeted, full, search, and human output modes.
- **SCOPE**:
  - register a visible top-level `capabilities` command
  - add `--json`, `--full`, and `--search <term>`
  - support targeted lookup for top-level and nested command paths by joining positional arguments
  - emit compact, targeted, full, and search JSON payloads with integer `schema_version: 1`
  - render concise human text when `--json` is absent
  - reject targeted `--full` and targeted `--search` with actionable errors
  - return actionable unknown-command errors with suggestions when possible
  - avoid project-root lookup, config loading, file writes, network calls, subprocesses, git commands, and delegated Kit command execution
- **ACCEPTANCE**:
  - `kit capabilities --json` emits valid compact JSON with `capabilities` included
  - `kit capabilities dispatch --json` emits exactly one detailed record with dispatch safety notes
  - `kit capabilities --full --json` emits detailed records for all included commands
  - `kit capabilities --search verify --json` emits compact filtered records only
  - invalid combinations and unknown targeted commands exit non-zero with actionable text
- **VERIFY**:
  - `go test ./pkg/cli`
  - `go run ./cmd/kit capabilities --json`
  - `go run ./cmd/kit capabilities dispatch --json`
  - `go run ./cmd/kit capabilities --full --json`
  - `go run ./cmd/kit capabilities --search verify --json`
  - `go test ./pkg/cli -run TestCapabilitiesUnknownCommandIsActionable`
- **EXPECTED FILES**:
  - `pkg/cli/capabilities.go`
  - `pkg/cli/capabilities_catalog.go`
  - `pkg/cli/capabilities_test.go`
- **RISK**: Medium; mode handling can accidentally blur compact versus detailed payloads.
- **ROLLBACK**: Remove the command registration and command file while keeping catalog work isolated for revision.
- **NOTES**: The unknown-command verification is expected to exit non-zero; inspect the error text rather than treating the non-zero exit as a task failure.

### T004

- **GOAL**: Make `capabilities` discoverable in root help under Inspect & Repair.
- **SCOPE**:
  - add `capabilities` to `commandOrder` near `map`
  - add `capabilities` to the Inspect & Repair `rootCommandSections` entry
  - update root help assertions for visible command grouping
  - preserve hidden/deprecated command behavior and existing root help categories
- **ACCEPTANCE**:
  - root help lists `capabilities` under Inspect & Repair
  - hidden/deprecated commands remain omitted from default root help
  - no existing command ordering or help grouping regresses
- **VERIFY**:
  - `go test ./pkg/cli -run TestRootHelp`
  - `go test ./pkg/cli -run TestDeprecatedCommands`
  - `go run ./cmd/kit --help`
- **EXPECTED FILES**:
  - `pkg/cli/root_help.go`
  - `pkg/cli/root_help_test.go`
- **RISK**: Low; root help changes are localized but affect public discoverability.
- **ROLLBACK**: Revert root help ordering and section edits.
- **NOTES**: Keep root help grouping root-only.

### T005

- **GOAL**: Prove JSON contract, filtering, targeted lookup, and actionable error behavior.
- **SCOPE**:
  - test compact JSON schema and compact-only fields
  - test targeted detail for `dispatch`
  - test targeted detail for `ci` when `ci` is registered
  - test full JSON includes detailed fields
  - test search filtering, zero matches, and compact search shape
  - test unknown targeted command suggestions
  - test invalid flag combinations
- **ACCEPTANCE**:
  - tests pin `schema_version: 1`
  - tests prove full-only detail fields do not leak into compact search output
  - tests cover `dispatch` PR/network notes and registered `ci` GitHub/subprocess/cache/Copilot notes
  - tests prove unknown and invalid-command paths are actionable
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestCapabilities'`
  - `go run ./cmd/kit capabilities --json`
  - `go run ./cmd/kit capabilities dispatch --json`
  - `go run ./cmd/kit capabilities --search verify --json`
- **EXPECTED FILES**:
  - `pkg/cli/capabilities_test.go`
- **RISK**: Medium; weak tests would let downstream agents consume an unstable JSON contract.
- **ROLLBACK**: Revert the failing tests and repair command behavior before proceeding.
- **NOTES**: Use JSON decoding assertions instead of brittle full-output string comparisons where possible.

### T006

- **GOAL**: Guard the catalog against command drift and prove the new command is read-only.
- **SCOPE**:
  - compare visible registered root commands to catalog records or explicit tested exclusions
  - compare selected nested command paths to real Cobra commands
  - assert hidden/deprecated compatibility commands are excluded from compact default and labeled in full/targeted output when represented
  - assert important flags for safety-sensitive commands stay reflected in metadata
  - run compact, targeted, full, search, and error paths in a temporary project and assert no forbidden files are created or modified
- **ACCEPTANCE**:
  - visible registered root command drift fails tests
  - catalog records referencing absent Cobra paths fail tests
  - `.kit.yaml`, `.kit/state.json`, `.kit/runs`, `.kit/loops`, feature docs, and notes remain untouched by `kit capabilities`
  - `capabilities` itself has read-only metadata distinct from described command behavior
- **VERIFY**:
  - `go test ./pkg/cli -run TestCapabilityCatalog`
  - `go test ./pkg/cli -run TestCapabilitiesDoesNotRequireProjectRootOrWriteFiles`
  - `go test ./pkg/cli`
- **EXPECTED FILES**:
  - `pkg/cli/capabilities_test.go`
  - `pkg/cli/root_help_test.go`
- **RISK**: High; drift and side-effect regressions are the primary long-term failure modes.
- **ROLLBACK**: Revert drift/read-only tests only after repairing catalog or command behavior; do not weaken safety checks to pass.
- **NOTES**: If current test names differ, use an equivalent targeted `go test ./pkg/cli -run ...` pattern and keep full `go test ./pkg/cli`.

### T007

- **GOAL**: Document `kit capabilities` for humans and agents without encouraging repeated full-catalog reads.
- **SCOPE**:
  - update README command documentation under Inspect & Repair
  - update `docs/specs/0000_INIT_PROJECT.md` with the durable command contract
  - update selected agent docs with concise usage guidance
  - clarify that `kit capabilities` answers command-selection questions while `kit map` and `kit map --context` answer document/reference-routing questions
  - keep docs short and avoid duplicating the full catalog
- **ACCEPTANCE**:
  - docs mention compact, targeted, full, and search usage
  - docs instruct agents to run `kit capabilities --json` only when command choice is uncertain
  - docs prefer targeted lookup over repeated `--full` reads
  - docs do not claim `kit capabilities` replaces map, status, check, or RLM docs
- **VERIFY**:
  - `rg -n capabilities README.md docs/specs/0000_INIT_PROJECT.md docs/agents`
  - `rg -n "kit map" README.md docs/specs/0000_INIT_PROJECT.md docs/agents`
  - `rg -n -- --full README.md docs/specs/0000_INIT_PROJECT.md docs/agents`
  - `rg -n -- --search README.md docs/specs/0000_INIT_PROJECT.md docs/agents`
  - `kit check 0033-kit-capabilities`
  - `kit map 0033-kit-capabilities`
- **EXPECTED FILES**:
  - `README.md`
  - `docs/specs/0000_INIT_PROJECT.md`
  - `docs/agents/README.md`
  - `docs/agents/RLM.md`
  - `docs/agents/TOOLING.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Medium; overlong docs can recreate the full-context loading problem this feature solves.
- **ROLLBACK**: Revert docs wording and keep code/tests if behavior is correct.
- **NOTES**: Update only the agent docs that need concise command-discovery guidance.

### T008

- **GOAL**: Prove the complete feature and reconcile feature state before handoff.
- **SCOPE**:
  - run formatter and full repository tests
  - run the SPEC acceptance commands
  - run feature document validation and map resolution
  - update `docs/PROJECT_PROGRESS_SUMMARY.md` to reflect the latest completed artifact and implementation status
  - record any failing command with root cause instead of marking it done
- **ACCEPTANCE**:
  - all required verification commands pass or failures have documented root cause and owner
  - feature docs contain no placeholder-only required sections
  - progress summary maps claims to feature docs
  - no existing command JSON payload outside `capabilities` is intentionally changed
- **VERIFY**:
  - `gofmt -w pkg/cli/capabilities.go pkg/cli/capabilities_catalog.go pkg/cli/capabilities_test.go pkg/cli/root_help.go pkg/cli/root_help_test.go`
  - `go test ./...`
  - `go run ./cmd/kit capabilities --json`
  - `go run ./cmd/kit capabilities dispatch --json`
  - `go run ./cmd/kit capabilities --full --json`
  - `go run ./cmd/kit capabilities --search verify --json`
  - `go test ./pkg/cli -run TestCapabilitiesUnknownCommandIsActionable`
  - `kit check 0033-kit-capabilities`
  - `kit map 0033-kit-capabilities`
- **EXPECTED FILES**:
  - all files touched by T002-T007
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Medium; broad verification may reveal unrelated dirty-worktree failures.
- **ROLLBACK**: Revert feature files if verification shows the approach is unsalvageable; otherwise fix the failing task before completion.
- **NOTES**: The unknown-command CLI path is covered by `TestCapabilitiesUnknownCommandIsActionable` so the verification harness can stay exit-code based.

## DEPENDENCIES

- T001 must finish before code edits because the constitution requires an implementation readiness gate.
- T002 must precede T003 because every command mode renders catalog projections.
- T003 must precede T004-T006 because help, output tests, drift tests, and read-only tests need the command to exist.
- T004 can run after T003 and before final drift tests so root help placement is part of the test surface.
- T005 and T006 must finish before T007 so documentation reflects tested behavior, not intended behavior.
- T008 depends on all implementation, tests, and docs tasks.

## NOTES

- Understanding is 97%; no missing decisions block task execution.
- Keep `parallelization_mode: "rlm"` during implementation planning; execute these tasks linearly unless the implementation agent proves non-overlapping workstreams after T003.
- Do not implement, stabilize, or redesign `kit ci`; describe it only if still registered.
- Do not add GitHub delivery, branch, commit, push, issue, or PR behavior.

<!-- REFLECTION_COMPLETE -->
