# PLAN

## SUMMARY

Add a read-only `kit map` command backed by existing feature and document state,
extend brainstorm and spec contracts with explicit `RELATIONSHIPS` sections,
and keep relationship parsing lightweight and deterministic.

## APPROACH

- record the command and document-contract changes in canonical docs first
- extend brainstorm and spec templates plus document validation with a required
  `RELATIONSHIPS` section seeded with `none`
- backfill existing brainstorm and spec docs with explicit `RELATIONSHIPS`
  sections so the new validation rule does not strand the repo in an invalid
  intermediate state
- add a small relationship parser that reads explicit bullets from
  `BRAINSTORM.md` and `SPEC.md`
- normalize harmless inline-code formatting around relationship targets while
  still rejecting real prose in validation paths
- implement `kit map` as a terminal-first renderer over canonical docs and
  feature state
- keep read-only map rendering resilient by warning on malformed relationship
  lines instead of failing the entire view
- keep the map read-only and avoid introducing persisted derived graph files
- update practical docs and command help after the command shape is stable

## COMPONENTS

- `docs/CONSTITUTION.md`
  - document the new command and the explicit relationship-field rule
- `docs/specs/0000_INIT_PROJECT.md`
  - update canonical document requirements and the `kit map` contract
- `internal/document/document.go`
  - require `RELATIONSHIPS` for brainstorm and spec docs
- `internal/templates/templates.go`
  - add `RELATIONSHIPS` sections seeded with `none`
- existing feature docs under `docs/specs/*`
  - add `RELATIONSHIPS` sections with `none` where no concrete lineage is
    declared in this pass
- `internal/feature/relationships.go`
  - parse explicit feature relationships from canonical docs
- `internal/feature/map.go` or `internal/mapview/*`
  - build a map-friendly state model from features, docs, and relationships
- `pkg/cli/map.go`
  - add the new command and render project or feature views
- `pkg/cli/root.go`
  - register and order the command in help output

## DATA

- `BRAINSTORM.md`
  - new required `RELATIONSHIPS` section with `none` or explicit relationship
    bullets
- `SPEC.md`
  - new required `RELATIONSHIPS` section with `none` or explicit relationship
    bullets
- existing feature docs
  - backfilled `RELATIONSHIPS: none` when no specific lineage is recorded yet
- relationship edge
  - source feature identifier
  - source document type
  - relationship type
  - target feature identifier
  - resolved or unresolved target state
- map view model
  - global docs
  - feature docs with existence and optionality
  - current phase and paused state
  - outgoing and relevant incoming edges

## INTERFACES

- new command: `kit map [feature]`
  - no argument: project-wide map
  - feature argument: feature-scoped map
- relationship syntax in docs:
  - `none`
  - `- builds on: 0007-catchup-command`
  - `- builds on: <feature>` with the target optionally wrapped in inline code
  - `- depends on: 0009-spec-skills-discovery`
  - `- related to: 0011-handoff-document-sync`
- first-pass output:
  - human-readable terminal text graph with box drawing by default
  - ASCII-safe fallback when the environment cannot render Unicode reliably
  - no `--json`
  - no persisted markdown output

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| feature enumeration | code | `internal/feature/feature.go` | current repo feature list and phase state | active |
| lifecycle state | code | `internal/config/config.go`, `internal/feature/lifecycle.go` | paused-state annotation in map output | active |
| document parser | code | `internal/document/document.go` | section extraction and validation | active |
| templates | code | `internal/templates/templates.go` | seeded relationship sections | active |
| root help surface | code | `pkg/cli/root.go` | command ordering and user discovery | active |

## RISKS

- map output can become a second status view instead of a structural view
  - mitigate by focusing on hierarchy, lineage, and doc ownership first
- relationship syntax may be too loose to parse reliably
  - mitigate by constraining supported labels and feature identifier format
- brainstorm and spec relationships can diverge
  - mitigate by showing both source documents and avoiding silent merging rules
- a project-wide map can become noisy with many features
  - mitigate by keeping feature rows compact and offering a feature-scoped view

## TESTING

- unit tests for relationship parsing from brainstorm and spec docs
- unit tests for unresolved-target handling
- unit tests for project-wide and feature-scoped map rendering
- validator tests for new `RELATIONSHIPS` requirements
- CLI tests for `kit map` command resolution and output
- targeted verification that rollup, status, and existing prompts still behave
  after template and validator changes
