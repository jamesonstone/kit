---
kit_metadata_version: 1
artifact: tasks
feature:
  id: "0030"
  slug: reference-graph-routing
  dir: 0030-reference-graph-routing
relationships:
  - type: builds_on
    target: 0026-front-matter-integration
  - type: builds_on
    target: 0016-document-map-relationships
  - type: related_to
    target: 0017-reconcile-command
  - type: related_to
    target: 0029-scaffold-workflows-prepare
references:
  - id: reference-graph-plan
    name: Reference graph plan
    type: feature
    target: docs/specs/0030-reference-graph-routing/PLAN.md
    selector_type: artifact
    selector: PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation task sequencing
    status: active
---
# TASKS

## PROGRESS TABLE

| ID   | TASK                                                | STATUS | OWNER | DEPENDENCIES |
| ---- | --------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Record reference graph feature docs                 | done   | agent |              |
| T002 | Update metadata schema and validation               | done   | agent | T001         |
| T003 | Extend map output, resolver, and context read plan  | done   | agent | T002         |
| T004 | Add reconcile migration prompt behavior             | done   | agent | T002         |
| T005 | Update prompts, templates, and repo docs            | done   | agent | T002, T003   |
| T006 | Migrate existing feature front matter to references | done   | agent | T002, T005   |
| T007 | Update tests and run verification                   | done   | agent | T003, T004, T006 |

## TASK LIST

- [x] T001: Record reference graph feature docs [PLAN-01]
- [x] T002: Update metadata schema and validation [PLAN-01]
- [x] T003: Extend map output, resolver, and focused context read plan [PLAN-02]
- [x] T004: Add reconcile migration prompt behavior [PLAN-03]
- [x] T005: Update prompts, templates, and repo docs [PLAN-04]
- [x] T006: Migrate existing feature front matter to references [PLAN-04]
- [x] T007: Update tests and run verification [PLAN-05]

## TASK DETAILS

### T001

- **GOAL**: Capture the approved reference graph migration before implementation.
- **SCOPE**:
  - create `docs/specs/0030-reference-graph-routing/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
- **ACCEPTANCE**:
  - docs define destructive migration, schema, map read plans, and reconcile migration prompt behavior

### T002

- **GOAL**: Make `references` the canonical metadata model.
- **SCOPE**:
  - add reference structs and enums
  - validate required fields and enum values
  - validate selector type and read policy consistency
  - reject front matter `dependencies`
  - warn on unpinned line-number references
- **ACCEPTANCE**:
  - old front matter `dependencies` fails validation
  - generated references pass validation

### T003

- **GOAL**: Expose reference graph metadata through project maps.
- **SCOPE**:
  - collect reference links from canonical front matter
  - resolve local files, feature artifacts, headings, symbols, commands, URLs, and node IDs where possible
  - render reference links in normal map output
  - add de-duplicated `kit map <feature> --context`
  - add JSON output for map/context plans
- **ACCEPTANCE**:
  - map output lists target, selector, relation, read policy, status, and resolver status
  - context output is a pointer-only read plan
  - context JSON is deterministic and grouped by read policy

### T004

- **GOAL**: Give users a consistent migration prompt path.
- **SCOPE**:
  - add a reconcile migration flag
  - include migration rules and verification steps in generated prompts
- **ACCEPTANCE**:
  - `kit reconcile --migrate-references --output-only` emits a migration prompt

### T005

- **GOAL**: Stop teaching old dependency metadata.
- **SCOPE**:
  - update workflow prompts
  - update templates
  - update README and agent docs
- **ACCEPTANCE**:
  - prompt text says `references`, `target`, `relation`, and `read_policy`
  - prompt text does not instruct legacy dependency fallback for canonical references

### T006

- **GOAL**: Bring existing repo feature docs into the new format.
- **SCOPE**:
  - convert front matter `dependencies` to `references`
  - replace `location` with `target`
  - add relation and read policy values
  - avoid exact line selectors unless pinned or unavoidable
- **ACCEPTANCE**:
  - `rg '^dependencies:' docs/specs` finds no front matter dependency blocks

### T007

- **GOAL**: Prove the migration works.
- **SCOPE**:
  - update tests
  - run focused Go tests
  - run Kit document checks
- **ACCEPTANCE**:
  - verification commands pass or failures are documented with root cause

## DEPENDENCIES

- T002 depends on T001 because implementation follows the documented schema.
- T003 depends on T002 because map output reads canonical references.
- T004 depends on T002 because migration prompts target the new schema.
- T005 depends on T002 and T003 because prompts must describe the shipped metadata and map behavior.
- T006 depends on T002 and T005 because docs should migrate after the schema and wording are settled.
- T007 depends on T003, T004, and T006 because verification covers the final behavior.

## NOTES

- not required
