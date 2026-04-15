# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record the `kit map` and relationship-section contract in canonical docs | done | codex | none |
| T002 | Add `RELATIONSHIPS` to brainstorm/spec validation, templates, and existing docs | done | codex | T001 |
| T003 | Implement explicit feature-relationship parsing and map state building | done | codex | T002 |
| T004 | Add the `kit map` CLI command and terminal rendering | done | codex | T003 |
| T005 | Update practical docs and prompt guidance for relationship metadata | done | codex | T002, T004 |
| T006 | Add tests and run verification for map output and relationship parsing | done | codex | T003, T004, T005 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record the `kit map` and relationship-section contract in canonical docs
- [x] T002: Add `RELATIONSHIPS` to brainstorm/spec validation, templates, and existing docs
- [x] T003: Implement explicit feature-relationship parsing and map state building
- [x] T004: Add the `kit map` CLI command and terminal rendering
- [x] T005: Update practical docs and prompt guidance for relationship metadata
- [x] T006: Add tests and run verification for map output and relationship parsing

## TASK DETAILS

### T001
- **GOAL**: lock the new command and document requirements into the canonical contract
- **SCOPE**: add a dedicated feature spec set and update core docs that define
  the canonical artifact model
- **ACCEPTANCE**: the command surface, relationship syntax, and no-new-markdown
  constraint are written down before code changes
- **NOTES**: keep the first pass ASCII and read-only

### T002
- **GOAL**: make relationship metadata part of the canonical brainstorm and spec shape
- **SCOPE**: update required sections, templates, and validation tests for
  `BRAINSTORM.md` and `SPEC.md`, and backfill existing feature docs so the new
  validation rule is immediately consistent
- **ACCEPTANCE**: new docs seed `RELATIONSHIPS` with `none`, and touched docs
  require populated relationship content
- **NOTES**: prefer explicit feature identifiers over prose-only references

### T003
- **GOAL**: extract a stable feature graph from explicit canonical relationships
- **SCOPE**: add parsing helpers and a map-friendly model for docs, lifecycle
  state, and resolved or unresolved edges
- **ACCEPTANCE**: relationship parsing is deterministic and does not require
  hidden state or inference
- **NOTES**: keep conflicting brainstorm/spec declarations visible rather than
  merging them away; normalize harmless inline-code wrappers around canonical
  feature IDs

### T004
- **GOAL**: give users an on-demand structural view of the project
- **SCOPE**: add `kit map [feature]`, feature resolution, project-wide and
  feature-scoped graphical terminal rendering, and help integration
- **ACCEPTANCE**: the command prints stable read-only output for both scopes
- **NOTES**: do not add JSON or Mermaid in this phase; malformed relationship
  lines should become visible warnings rather than hard failures

### T005
- **GOAL**: keep docs and prompt guidance aligned with the new relationship metadata
- **SCOPE**: update README, root help as needed, and workflow prompt guidance
  that shapes brainstorm or spec authoring
- **ACCEPTANCE**: users and agents are told where to record relationships and
  how `kit map` reads them
- **NOTES**: avoid adding relationship noise to unrelated status surfaces

### T006
- **GOAL**: verify the new command and metadata contract
- **SCOPE**: add focused parser, validator, rendering, and CLI tests; run
  targeted validation and a build
- **ACCEPTANCE**: relevant tests pass and cover the new graph and section logic
- **NOTES**: include unresolved-target and no-feature-project cases

## DEPENDENCIES

- T002 depends on T001 because validation and templates must follow the written
  contract
- T003 depends on T002 because the parser should target the approved section
  shape
- T004 depends on T003 because the command renders the derived map model
- T005 depends on T002 and T004 because docs should describe the final section
  and command surfaces
- T006 depends on T003-T005 because verification must validate the shipped
  implementation and docs

## NOTES

- `kit map` is read-only in this phase
- `RELATIONSHIPS` is canonical only when explicitly recorded in
  `BRAINSTORM.md` or `SPEC.md`
- `none` is the required empty-state value for relationship sections
- inline-code-wrapped relationship targets normalize back to canonical feature
  IDs during validation and rendering

<!-- REFLECTION_COMPLETE -->
