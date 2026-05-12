---
kit_metadata_version: 1
artifact: tasks
feature:
  id: "0026"
  slug: front-matter-integration
  dir: 0026-front-matter-integration
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add front-matter body parsing in `internal/document` [PLAN-APPROACH] [PLAN-COMPONENTS] | done | agent | |
| T002 | Add typed metadata schema, validation, and diagnostics [PLAN-DATA] [PLAN-COMPONENTS] | done | agent | T001 |
| T003 | Add front-matter-first accessors with legacy fallback [PLAN-APPROACH] [PLAN-INTERFACES] | done | agent | T002 |
| T004 | Add non-destructive metadata upsert helpers [PLAN-COMPONENTS] [PLAN-INTERFACES] | done | agent | T003 |
| T005 | Update feature artifact templates and creation paths [PLAN-COMPONENTS] [PLAN-INTERFACES] | done | agent | T004 |
| T006 | Migrate map, status, and rollup readers to metadata accessors [PLAN-COMPONENTS] [PLAN-INTERFACES] | done | agent | T003 |
| T007 | Migrate check and reconcile validation to metadata diagnostics [PLAN-APPROACH] [PLAN-INTERFACES] | done | agent | T003 |
| T008 | Migrate existing metadata mutation paths to front matter [PLAN-COMPONENTS] [PLAN-RISKS] | done | agent | T004, T005, T007 |
| T009 | Update prompt builders and agent-facing instructions [PLAN-APPROACH] [PLAN-INTERFACES] | done | agent | T005, T006, T007, T008 |
| T010 | Add parser, metadata, writer, and command tests [PLAN-TESTING] | done | agent | T001, T002, T003, T004, T005, T006, T007, T008, T009 |
| T011 | Update affected documentation and generated project summary [PLAN-INTERFACES] [PLAN-TESTING] | done | agent | T009, T010 |
| T012 | Run final validation and fix relevant failures [PLAN-TESTING] | done | agent | T010, T011 |

## TASK LIST

- [x] T001: Add front-matter body parsing in `internal/document` [PLAN-APPROACH] [PLAN-COMPONENTS]
- [x] T002: Add typed metadata schema, validation, and diagnostics [PLAN-DATA] [PLAN-COMPONENTS]
- [x] T003: Add front-matter-first accessors with legacy fallback [PLAN-APPROACH] [PLAN-INTERFACES]
- [x] T004: Add non-destructive metadata upsert helpers [PLAN-COMPONENTS] [PLAN-INTERFACES]
- [x] T005: Update feature artifact templates and creation paths [PLAN-COMPONENTS] [PLAN-INTERFACES]
- [x] T006: Migrate map, status, and rollup readers to metadata accessors [PLAN-COMPONENTS] [PLAN-INTERFACES]
- [x] T007: Migrate check and reconcile validation to metadata diagnostics [PLAN-APPROACH] [PLAN-INTERFACES]
- [x] T008: Migrate existing metadata mutation paths to front matter [PLAN-COMPONENTS] [PLAN-RISKS]
- [x] T009: Update prompt builders and agent-facing instructions [PLAN-APPROACH] [PLAN-INTERFACES]
- [x] T010: Add parser, metadata, writer, and command tests [PLAN-TESTING]
- [x] T011: Update affected documentation and generated project summary [PLAN-INTERFACES] [PLAN-TESTING]
- [x] T012: Run final validation and fix relevant failures [PLAN-TESTING]

## TASK DETAILS

### T001

- **GOAL**: Parse optional YAML front matter without breaking existing markdown body section behavior.
- **SCOPE**:
  - Update `internal/document/document.go` so `Parse` and `ParseFile` split a valid leading front-matter block before section detection.
  - Preserve original content and expose visible body content for section parsing, placeholder checks, and paragraph extraction.
  - Treat `---` outside the first file block as body text.
  - Keep existing public section APIs usable by current callers.
- **ACCEPTANCE**:
  - `internal/document` tests prove documents with and without front matter expose the same expected body sections.
  - Required-section validation ignores the front-matter block.
  - Placeholder detection does not treat YAML comments or front-matter content as markdown body placeholders.
  - Evidence artifact: updated `internal/document` tests passing.
- **NOTES**: This task must not change command output behavior yet.

### T002

- **GOAL**: Define the typed front-matter metadata contract and validation diagnostics.
- **SCOPE**:
  - Add metadata structs/enums for `kit_metadata_version`, `artifact`, `feature`, `relationships`, `dependencies`, `skills`, optional `summary`, and optional `intent`.
  - Validate artifact values, relationship enum values, dependency status values, required feature identity fields, and malformed YAML.
  - Normalize relationship machine values and human labels.
  - Keep feature directory identity authoritative over front-matter identity drift.
- **ACCEPTANCE**:
  - Malformed present front matter produces a file-specific validation error.
  - Invalid artifact, relationship type, dependency status, and missing required metadata are covered by tests.
  - Relationship enum values map to human labels used by renderers.
  - Evidence artifact: metadata validation tests in `internal/document`.
- **NOTES**: Missing front matter in existing docs is not a feature-local error.

### T003

- **GOAL**: Expose one metadata read API that prefers front matter and falls back to legacy markdown sections.
- **SCOPE**:
  - Add accessors for relationships, dependencies, skills, summary, intent, and metadata conflicts.
  - Reuse legacy parsers for `## RELATIONSHIPS`, `## DEPENDENCIES`, and `## SKILLS` fallback.
  - Ensure front matter wins when front matter and body metadata disagree.
  - Surface conflict diagnostics without blocking normal reads.
- **ACCEPTANCE**:
  - Tests cover front-matter-only, legacy-only, matching dual-source, and conflicting dual-source documents.
  - Legacy inline-code-wrapped relationship targets still normalize correctly.
  - Empty dependency lists and legacy `none` values return no entries.
  - Evidence artifact: accessor tests with clear fixture names.
- **NOTES**: Command packages should not introduce new table parsers after this task.

### T004

- **GOAL**: Provide safe metadata upsert helpers for commands that already write feature docs.
- **SCOPE**:
  - Add writer helpers that create or update front matter while preserving body text.
  - Preserve unknown front-matter fields and existing known values where practical.
  - Support upserting relationships, dependencies, skills, feature identity, artifact type, summary, and intent.
  - Avoid promising YAML comment preservation.
- **ACCEPTANCE**:
  - Tests prove unknown front-matter fields survive an upsert.
  - Tests prove body content before and after metadata upserts remains intact.
  - Upserts create front matter for legacy docs without destructive body rewrites.
  - Evidence artifact: writer/upsert tests in `internal/document`.
- **NOTES**: Use existing `yaml.Node` patterns where they fit local conventions.

### T005

- **GOAL**: Generate canonical front matter in new feature artifacts.
- **SCOPE**:
  - Update `internal/templates/templates.go` so new `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` include required front matter.
  - Keep required markdown headings present and populated.
  - Remove duplicate machine-readable body tables from new docs where equivalent canonical front matter exists.
  - Update `kit brainstorm`, `kit spec`, `kit plan`, and `kit tasks` creation paths if they need feature-specific metadata values at write time.
- **ACCEPTANCE**:
  - Template tests assert required front matter exists in generated artifacts.
  - Template tests assert required headings still exist and are not placeholder-only after generation rules are applied.
  - New generated docs do not require body `RELATIONSHIPS`, `DEPENDENCIES`, or `SKILLS` tables as the canonical source.
  - Evidence artifact: updated template and workflow creation tests.
- **NOTES**: Keep task progress body structures unchanged.

### T006

- **GOAL**: Make read-only map/status/rollup consumers use canonical metadata without changing derived state rules.
- **SCOPE**:
  - Update `internal/feature/map.go` and `pkg/cli/map.go` to read relationships and dependencies through metadata accessors.
  - Update `internal/feature/status.go` and `internal/rollup/rollup.go` to prefer optional summary/intent metadata with body fallback.
  - Keep phase derived from files and task completion.
  - Keep task progress derived from `TASKS.md` checkboxes and progress tables.
- **ACCEPTANCE**:
  - `kit map` output remains stable for legacy fixtures.
  - New tests prove `kit map` renders front-matter relationships and dependencies.
  - Rollup/status tests prove summary/intent metadata is preferred when present and body fallback still works.
  - Evidence artifact: updated `internal/feature`, `pkg/cli/map`, and `internal/rollup` tests.
- **NOTES**: `kit map` remains read-only.

### T007

- **GOAL**: Make feature and project validation enforce the migration contract.
- **SCOPE**:
  - Update `Document.Validate` and `pkg/cli/check.go` to error on malformed present front matter.
  - Keep `kit check <feature>` quiet for missing front matter in legacy docs.
  - Update `kit check --project` audit paths to report legacy-only metadata, generated/touched docs missing canonical metadata, stale targets, invalid metadata, and front-matter/body conflicts.
  - Update reconcile audit helpers to report the same migration findings without writing files.
- **ACCEPTANCE**:
  - `kit check <feature>` tests cover malformed front matter and legacy missing-front-matter tolerance.
  - `kit check --project` tests cover migration warnings/errors and stale targets.
  - `kit reconcile` tests prove findings are prompt-only and file-specific.
  - Evidence artifact: updated `pkg/cli/check` and `pkg/cli/reconcile` tests.
- **NOTES**: Preserve current severity behavior unless SPEC explicitly changes it.

### T008

- **GOAL**: Move existing relationship/dependency mutation paths behind front-matter upserts.
- **SCOPE**:
  - Update `pkg/cli/brainstorm_notes.go` to upsert feature notes dependencies in front matter.
  - Update `pkg/cli/profile_dependencies.go` to upsert frontend profile and design material dependencies in front matter.
  - Update `pkg/cli/brainstorm_backlog.go` to upsert backlog relationships in front matter.
  - Keep legacy body mutation only as fallback when needed for pre-front-matter documents.
- **ACCEPTANCE**:
  - Tests prove notes, frontend profile/design, and backlog relationships are written to front matter.
  - Tests prove `.gitkeep` placeholder inputs do not become active dependencies.
  - Tests prove legacy docs remain usable and non-destructive after mutation.
  - Evidence artifact: updated mutation-path tests.
- **NOTES**: Do not add a migration command or flag.

### T009

- **GOAL**: Update agent prompts to maintain front matter and preserve RLM routing.
- **SCOPE**:
  - Update brainstorm/spec/plan/tasks/implement/reflect/handoff/resume/summarize prompt text that currently instructs agents to edit body metadata tables.
  - Update skill discovery guidance so `SPEC.md` skills are represented through canonical metadata when front matter exists.
  - Preserve just-in-time loading language and legacy fallback instructions.
  - Avoid expanding root instruction files or adding monolithic manuals.
- **ACCEPTANCE**:
  - Prompt tests assert front-matter-first wording appears where metadata maintenance is required.
  - Prompt tests assert RLM and legacy fallback language remains present.
  - No prompt instructs agents to use body metadata tables as the canonical source when front matter exists.
  - Evidence artifact: updated prompt test snapshots or string assertions.
- **NOTES**: Keep public command names and flags unchanged.

### T010

- **GOAL**: Add comprehensive regression coverage across parser, metadata, writers, and command consumers.
- **SCOPE**:
  - Add or update tests for `internal/document`, `internal/feature`, `internal/templates`, `internal/rollup`, and `pkg/cli`.
  - Cover happy paths, malformed front matter, invalid enums, legacy fallback, conflicts, unknown fields, map rendering, reconcile/check findings, mutation paths, and prompt text.
  - Ensure tests fail without the front-matter implementation rather than only asserting existing behavior.
- **ACCEPTANCE**:
  - Test names clearly describe the behavior under test.
  - Each SPEC acceptance criterion has at least one direct test or validation evidence path.
  - `go test ./...` reaches the new tests.
  - Evidence artifact: changed test files and passing test output.
- **NOTES**: This task may be worked incrementally with implementation tasks, but is not done until coverage spans all completed behavior.

### T011

- **GOAL**: Update user-facing docs and generated summaries affected by the metadata contract.
- **SCOPE**:
  - Update README or repo-local docs only where behavior or workflow guidance changed.
  - Keep root instruction files thin if touched at all.
  - Run the project summary refresh after TASKS and implementation state change.
  - Keep docs consistent with the no-new-command migration strategy.
- **ACCEPTANCE**:
  - Affected docs mention front matter as canonical metadata without duplicating a full manual.
  - `docs/PROJECT_PROGRESS_SUMMARY.md` reflects the highest completed artifact phase.
  - No root instruction file becomes a full workflow manual.
  - Evidence artifact: doc diff plus rollup output.
- **NOTES**: Skip README changes only if implementation creates no user-facing workflow behavior beyond generated artifacts.

### T012

- **GOAL**: Prove the implementation is complete and fix relevant failures before sign-off.
- **SCOPE**:
  - Run `go test ./...`.
  - Run `go run ./cmd/kit check front-matter-integration`.
  - Run `go run ./cmd/kit map front-matter-integration`.
  - Run `go run ./cmd/kit check --project` and handle expected migration findings according to the implemented contract.
  - Fix failures caused by this feature before marking tasks complete.
- **ACCEPTANCE**:
  - Validation commands run and their outcomes are recorded.
  - Build/test failures caused by this feature are fixed.
  - Any remaining validation findings are either expected migration warnings or explicitly documented with next action.
  - Evidence artifact: final validation command output and updated `TASKS.md` completion state.
- **NOTES**: Reflect review fixed two in-scope issues: metadata upsert callers now surface malformed front-matter errors instead of silently no-oping, and front-matter feature identity validation now checks `feature.id`, `feature.slug`, and `feature.dir` against the containing feature directory. `go test ./...`, `go vet ./...`, `go run ./cmd/kit check front-matter-integration`, and `go run ./cmd/kit map front-matter-integration` passed. The project-level reconciliation added canonical front matter to legacy feature artifacts, moved the readiness-gate feature to `0027-implement-readiness-gate`, and `go run ./cmd/kit check --project` plus `go run ./cmd/kit check --all` now pass.

## DEPENDENCIES

- Execute tasks in numeric order unless a later implementation readiness gate proves a different order is safer.
- T005 and T008 must not start before metadata write helpers exist.
- T009 should wait until the new metadata contract and command behavior are stable enough for prompt text to be precise.
- Project-level validation drift was resolved by moving the readiness-gate feature to `docs/specs/0027-implement-readiness-gate` and adding canonical front matter to legacy feature artifacts.
- Expected migration warnings remain for older feature docs without canonical YAML front matter.

## NOTES

The implementation readiness gate must run before production code changes. If the gate finds contradictions between SPEC, PLAN, and TASKS, update the canonical docs before implementing.
