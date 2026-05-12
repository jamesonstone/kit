---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: "0026"
  slug: front-matter-integration
  dir: 0026-front-matter-integration
relationships:
  - type: builds_on
    target: 0016-document-map-relationships
  - type: related_to
    target: 0004-brainstorm-first-workflow
  - type: related_to
    target: 0017-reconcile-command
  - type: related_to
    target: 0021-project-validation-and-instruction-registry
  - type: related_to
    target: 0025-v0-prompt-library
---
# BRAINSTORM

## SUMMARY

Kit currently stores feature relationships, phase dependencies, skills, and other file-linking metadata in markdown body sections and tables, which makes parsing, validation, map rendering, and prompt routing depend on brittle body structure. The likely direction is to introduce typed YAML front matter for canonical machine-readable metadata while preserving readable markdown bodies and legacy fallback during migration.

## USER THESIS

refactor the kit project to migrate all relational metadata and file-linking logic from the markdown body into structured yaml front-matter. Look at all `kit` commands and determine how front-matter can be leveraged.

## RELATIONSHIPS

- builds on: `0016-document-map-relationships`
- related to: `0004-brainstorm-first-workflow`
- related to: `0017-reconcile-command`
- related to: `0021-project-validation-and-instruction-registry`
- related to: `0025-v0-prompt-library`

## CODEBASE FINDINGS

1. `internal/document/document.go` is the central markdown parser and validator. It currently treats the whole file as markdown body, finds `##` sections with `sectionPattern`, validates required body sections through `RequiredSections`, and has no front-matter parser or metadata model.
2. `internal/document/relationships.go` makes `## RELATIONSHIPS` in `BRAINSTORM.md` and `SPEC.md` the machine-readable source for feature graph edges. It accepts `none` or bullets like `- builds on: 0001-example-feature`, normalizes inline-code-wrapped targets, and exposes strict plus relaxed parsing modes.
3. `internal/feature/map.go` consumes relationship metadata from `BRAINSTORM.md` and `SPEC.md` body sections, then consumes dependency metadata from `## DEPENDENCIES` markdown tables in `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`. This is the most direct current consumer of relational metadata and file-linking logic.
4. `internal/feature/map.go` and `pkg/cli/profile_dependencies.go` each contain markdown dependency-table parsing helpers. The duplication is a strong sign that dependency rows should move behind one typed metadata API instead of being parsed ad hoc from body tables.
5. `pkg/cli/check.go` validates feature documents by calling `document.ParseFile(...).Validate()`, so any front-matter schema must flow through `internal/document` rather than being bolted onto individual commands.
6. `pkg/cli/reconcile_audit_helpers.go` and `pkg/cli/reconcile_audit_tables.go` enforce current body-section and table contracts. `kit reconcile` will need to detect legacy body metadata, malformed front matter, missing canonical metadata, stale relationship targets, and migration drift.
7. `internal/templates/templates.go` currently scaffolds metadata-bearing sections in body markdown: `## RELATIONSHIPS` in `BRAINSTORM.md` and `SPEC.md`, `## DEPENDENCIES` in `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`, and `## SKILLS` in `SPEC.md`. Templates are the right place to seed the new front-matter shape.
8. `pkg/cli/brainstorm_notes.go`, `pkg/cli/profile_dependencies.go`, and `pkg/cli/brainstorm_backlog.go` mutate body metadata after scaffold time by replacing markdown sections or appending dependency rows. Those mutation paths should be consolidated behind front-matter read/update helpers.
9. `internal/rollup/rollup.go` extracts human summaries from body sections (`SUMMARY`, `PROBLEM`, `OPEN-QUESTIONS`, `APPROACH`) and generates `PROJECT_PROGRESS_SUMMARY.md`. Summary extraction is not purely relational metadata, but front matter could optionally provide concise summary/intent fields for deterministic rollup without replacing rich body content.
10. `internal/feature/feature.go` derives phase from file existence and task checkboxes, not from metadata. Phase should probably remain derived state to avoid drift, while front matter can describe artifact type, feature identity, relationships, dependencies, and external links.
11. `pkg/cli/skill_prompt.go` already teaches agents to create skill `SKILL.md` files with YAML front matter (`name`, `description`). Kit therefore has a conceptual precedent for markdown plus YAML metadata, but no reusable parser for Kit feature documents yet.
12. `internal/config/prompt_files.go` uses `gopkg.in/yaml.v3` and `yaml.Node` helpers to update `.kit.yaml` while preserving unrelated fields. That code is not a front-matter parser, but it is a useful local pattern for additive YAML writes and unknown-field preservation.
13. `kit map`, `kit check`, `kit reconcile`, `kit rollup`, `kit status`, `kit resume`, `kit handoff`, `kit summarize`, `kit skill mine`, and the workflow prompt commands all depend on feature-document metadata either directly or through shared helpers. A command-by-command migration without a shared document metadata layer would create inconsistent behavior.
14. `docs/CONSTITUTION.md` requires markdown and YAML, filesystem-backed explicit state, agent portability, no hidden databases, and documents readable without Kit. YAML front matter fits those constraints if the body stays human-readable and the metadata schema stays portable.
15. The feature notes directory contains only `.gitkeep`; there are no usable pre-brainstorm notes to read or promote into this artifact.

## AFFECTED FILES

1. `internal/document/document.go` — central parse/validate path; needs front-matter parsing, body extraction, metadata accessors, and validation hooks.
2. `internal/document/relationships.go` — current body relationship parser; likely becomes legacy parser or mapper between front-matter relationship entries and old labels.
3. `internal/document/section_content.go` — current visible-body content helpers; must ignore front matter when checking body sections.
4. `internal/feature/map.go` — map graph and dependency-link consumer; should read canonical front matter first, with legacy body fallback during migration.
5. `internal/feature/status.go` — summary/status extraction; may optionally use front-matter summary fields while preserving body fallback.
6. `internal/rollup/rollup.go` — project progress summary generation; may use front-matter metadata for summary, intent, links, and notes/design references where appropriate.
7. `internal/templates/templates.go` — scaffolds `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md`; must seed the new front-matter schema.
8. `pkg/cli/check.go` — feature/project validation entrypoint; must surface front-matter schema errors and migration drift.
9. `pkg/cli/reconcile_audit_helpers.go` — structured document audit path; must detect missing or stale canonical metadata and legacy-only body metadata.
10. `pkg/cli/reconcile_audit_tables.go` — current table contract checks; should be revised if dependencies/skills move out of required body tables.
11. `pkg/cli/brainstorm.go` and `pkg/cli/brainstorm_prompt.go` — create brainstorm docs, notes directories, and prompt agents to maintain dependencies/relationships.
12. `pkg/cli/brainstorm_notes.go` — current feature-notes dependency row mutation; should write front-matter dependencies instead of markdown tables.
13. `pkg/cli/profile_dependencies.go` — current frontend-profile/design-material dependency row mutation; should write front-matter dependencies/profile metadata.
14. `pkg/cli/brainstorm_backlog.go` — current backlog relationship mutation in `## RELATIONSHIPS`; should write front-matter relationships.
15. `pkg/cli/spec.go`, `pkg/cli/spec_output.go`, `pkg/cli/spec_context.go`, and `pkg/cli/spec_template.go` — create/spec prompts and instruct relationship/dependency/skills maintenance.
16. `pkg/cli/plan.go` and `pkg/cli/tasks.go` — create downstream artifacts and prompt agents to maintain dependency/task metadata.
17. `pkg/cli/implement.go`, `pkg/cli/reflect.go`, `pkg/cli/handoff_prompt.go`, `pkg/cli/summarize.go`, `pkg/cli/resume.go`, and `pkg/cli/skill_prompt.go` — prompt builders that should reference canonical front-matter metadata just in time instead of body metadata tables when available.
18. `pkg/cli/map.go` — terminal renderer for relationships/dependencies; output should remain read-only while its data source changes.
19. `internal/config/prompt_files.go` — local YAML-node update pattern that can inform front-matter writer design.
20. `docs/CONSTITUTION.md`, `README.md`, `docs/agents/*`, and `docs/references/*` — docs that may need concise updates if the canonical metadata contract changes.
21. Existing feature docs under `docs/specs/*/{BRAINSTORM.md,SPEC.md,PLAN.md,TASKS.md}` — migration/backcompat surface for legacy section/table metadata.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Feature notes | notes | docs/notes/0026-front-matter-integration | optional pre-brainstorm research input | optional |
| Constitution | doc | docs/CONSTITUTION.md | markdown/YAML portability, filesystem-backed state, document-source-of-truth constraints | active |
| Agents routing docs | doc | docs/agents/README.md, docs/agents/WORKFLOWS.md, docs/agents/RLM.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md | RLM routing, formal workflow semantics, completion rules | active |
| References index | doc | docs/references/README.md | confirmed durable references are conditional inputs only | active |
| kit map output | command output | `go run ./cmd/kit map 0026-front-matter-integration` | current phase, dependency baseline, relationship state | active |
| Project progress summary | doc | docs/PROJECT_PROGRESS_SUMMARY.md | prior-feature shortlist and current feature row | active |
| Brainstorm-first workflow | prior feature doc | docs/specs/0004-brainstorm-first-workflow/SPEC.md, docs/specs/0004-brainstorm-first-workflow/PLAN.md | brainstorm artifact, dependency table, prompt-only, rollup/phase precedents | active |
| Document map relationships | prior feature doc | docs/specs/0016-document-map-relationships/SPEC.md, docs/specs/0016-document-map-relationships/PLAN.md | current relationship body contract and map graph behavior | active |
| Reconcile command | prior feature doc | docs/specs/0017-reconcile-command/SPEC.md, docs/specs/0017-reconcile-command/PLAN.md | current drift audit and table/relationship validation expectations | active |
| Project validation and instruction registry | prior feature doc | docs/specs/0021-project-validation-and-instruction-registry/SPEC.md, docs/specs/0021-project-validation-and-instruction-registry/PLAN.md | `kit check --project` and shared registry precedent | active |
| Prompt library | prior feature doc | docs/specs/0025-v0-prompt-library/SPEC.md, docs/specs/0025-v0-prompt-library/PLAN.md | YAML config storage, prompt surfaces, dynamic prompt provider inventory | active |
| Document parser | code | internal/document/document.go, internal/document/section_content.go | markdown section parsing, required-section validation, placeholder handling | active |
| Relationship parser | code | internal/document/relationships.go | current relationship syntax and validation source | active |
| Map builder | code | internal/feature/map.go, pkg/cli/map.go | current relationship/dependency graph consumer and renderer | active |
| Reconcile/check validators | code | pkg/cli/check.go, pkg/cli/reconcile_audit_helpers.go, pkg/cli/reconcile_audit_tables.go | document drift detection and validation behavior | active |
| Templates | code | internal/templates/templates.go | generated feature artifact shape | active |
| Rollup/status extraction | code | internal/rollup/rollup.go, internal/feature/status.go | summary and phase extraction from feature docs | active |
| Metadata mutation helpers | code | pkg/cli/brainstorm_notes.go, pkg/cli/profile_dependencies.go, pkg/cli/brainstorm_backlog.go | current body-section/table updates that should move behind metadata helpers | active |
| YAML-node config helpers | code | internal/config/prompt_files.go | additive YAML write pattern and unknown-field preservation precedent | active |
| Skill front-matter precedent | code | pkg/cli/skill_prompt.go | existing Kit prompt guidance for YAML front matter in markdown skills | active |

## QUESTIONS

Resolved batch 1 decisions:

1. Front matter should become canonical for newly generated and touched feature docs, while legacy body sections remain readable fallback during migration.
2. `## RELATIONSHIPS`, `## DEPENDENCIES`, and `## SKILLS` may be removed from required body-section contracts eventually, but transitional compatibility must remain until existing docs can migrate cleanly.
3. `SPEC.md` skills should move into front matter because they are structured execution metadata.
4. Task progress should remain in human-editable `TASKS.md` checkboxes/body tables, not front matter.
5. Feature phase should remain derived from file existence and `TASKS.md` completion state, not stored in front matter.
6. Relationship types should use machine enum values such as `builds_on`, `depends_on`, and `related_to`, with render-time mapping back to human labels.
7. Canonical artifact self-links should be derived from the feature directory, not stored in front matter; front matter should store external and cross-feature links only.
8. The first migration pass should not add a new public migration command or flag; use `kit check --project`, `kit reconcile`, and normal workflow commands.
9. `PROJECT_PROGRESS_SUMMARY.md` should remain generated markdown without canonical front matter.
10. Legacy-project validation should warn and fall back first; malformed front matter should become an error once a front-matter block exists.

Resolved batch 2 decisions:

1. Every canonical artifact front matter should include `kit_metadata_version: 1`.
2. Every canonical artifact front matter should include `artifact: brainstorm|spec|plan|tasks`.
3. Every canonical artifact front matter should include `feature.id`, `feature.slug`, and `feature.dir`; the feature directory remains authoritative if these values conflict.
4. Relationships should use typed list entries:

   ```yaml
   relationships:
     - type: depends_on
       target: 0016-document-map-relationships
   ```

5. Dependencies should use typed list entries:

   ```yaml
   dependencies:
     - name: Feature notes
       type: notes
       location: docs/notes/0026-front-matter-integration
       used_for: optional pre-brainstorm research input
       status: optional
   ```

6. `SPEC.md` skills should use typed list entries:

   ```yaml
   skills:
     - name: rlm
       source: repo
       path: docs/agents/RLM.md
       trigger: broad or noisy context work
       required: false
   ```

7. New docs should not generate duplicate body metadata tables from front matter in v1 unless a specific command needs a human display.
8. Legacy `## RELATIONSHIPS`, `## DEPENDENCIES`, and `## SKILLS` sections should remain accepted as fallback for now, but `kit check --project` and `kit reconcile` should mark them as legacy once front matter is supported.
9. Normal workflow commands should auto-seed missing front matter in existing docs they already mutate, using append-only and non-destructive behavior.
10. `kit reconcile` should remain prompt-only for this migration phase; it should detect and explain migration needs without writing metadata.

Resolved batch 3 decisions:

1. If front matter and legacy body metadata both exist but conflict, front matter wins; body metadata is fallback only.
2. Conflicts between front matter and legacy body metadata should surface as warnings or reconcile findings, not hard errors, unless the front matter itself is malformed.
3. Front-matter writes should preserve unknown fields and existing known values where practical; v1 should not promise comment preservation.
4. `kit check <feature>` should not warn about missing front matter during the initial migration. Keep missing-front-matter migration warnings in `kit check --project` and `kit reconcile`.
5. New templates should remove body `## RELATIONSHIPS`, `## DEPENDENCIES`, and `## SKILLS` once front matter exists, while validators still accept legacy sections during migration.

Open questions:

no unresolved questions after user approval of batches 1-3.

## OPTIONS

1. Big-bang canonical front matter only.
   - Pros: clean data model, fewer dual-source ambiguities after migration.
   - Cons: high breakage risk for existing feature docs, many command/test updates at once, conflicts with Kit's current ability to read older projects.
2. Dual-read migration with front matter canonical for new/touched docs.
   - Pros: preserves existing projects, lets `map`/`check`/`reconcile` adopt the new model without stranding legacy docs, fits append-only/document-first behavior.
   - Cons: requires explicit precedence rules and tests for front-matter-vs-body conflicts during transition.
3. Front matter for relationships only.
   - Pros: small first step focused on the graph surface that motivated `kit map`.
   - Cons: leaves dependency tables, skills tables, notes/design links, and prompt-routing inputs in body markdown, so most file-linking duplication remains.
4. Centralize metadata in `.kit.yaml` instead of artifact front matter.
   - Pros: easy YAML parsing with existing config patterns.
   - Cons: contradicts the user thesis and weakens feature-doc portability by separating metadata from the documents it describes.
5. Generate body tables from front matter on demand.
   - Pros: keeps human-readable tables while avoiding table parsing as source of truth.
   - Cons: requires a write policy and could create noisy diffs unless generation is tightly scoped.

## RECOMMENDED STRATEGY

Use a dual-read, front-matter-canonical migration. Add a shared `internal/document` metadata layer that parses an optional YAML front-matter block, strips it from markdown body parsing, validates typed metadata, and exposes relationships/dependencies/skills/file links through one API. New and touched feature artifacts should write canonical metadata to front matter; `kit map`, `kit check`, `kit reconcile`, prompt builders, rollup/status, and metadata mutation helpers should read front matter first and legacy body sections second until migration is complete.

Initial canonical front-matter scope should include relationships, dependencies, skills, feature/artifact identity, notes/design-material references, and optional summary/intent fields. Task checkboxes and artifact phase should remain body/filesystem-derived unless a later spec proves that storing them in front matter will not create drift.

Use machine enum values in YAML for relationship types (`builds_on`, `depends_on`, `related_to`) and map them back to body/prompt labels (`builds on`, `depends on`, `related to`) when rendering human-facing text. Do not store canonical artifact self-links because `docs/specs/<feature>/BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` are deterministic from the feature directory.

Seed every canonical artifact with `kit_metadata_version: 1`, an `artifact` enum, and feature identity fields. Treat the feature directory as authoritative over duplicated identity fields if they drift. In v1, do not generate duplicate body metadata tables from front matter by default; keep body prose human-readable and let commands render metadata when needed.

In conflict cases, front matter wins and legacy body metadata is fallback only. Surface front-matter-vs-body conflicts through `kit reconcile` and project-level validation warnings, not feature-local hard failures. Malformed front matter remains an error because it means the canonical metadata source cannot be trusted.

Do not add a new public migration command in the first pass. Use existing `kit check --project` and `kit reconcile` to surface legacy-only metadata and migration drift, and let normal workflow commands seed or refresh front matter when they create or touch artifacts. Treat missing front matter in legacy docs as warning/fallback at first, but treat malformed front matter as an error once present. Keep `kit reconcile` prompt-only for this migration phase. Front-matter writes should preserve unknown fields and existing known values where practical, but v1 should not promise comment preservation.

New generated templates should remove body `## RELATIONSHIPS`, `## DEPENDENCIES`, and `## SKILLS` once equivalent front matter exists. Validators and command readers should continue accepting those body sections as legacy fallback during the migration window.

## NEXT STEP

Run `kit spec front-matter-integration` to convert this research and approved migration decisions into a precise requirements contract.
