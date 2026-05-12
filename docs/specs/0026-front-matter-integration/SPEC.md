---
kit_metadata_version: 1
artifact: spec
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
# SPEC

## SUMMARY

Adds typed YAML front matter as the canonical metadata layer for Kit feature artifacts while preserving legacy markdown-body fallback during migration. The feature covers relationships, dependencies, skills, artifact identity, file links, validation, mapping, prompts, and workflow readers without adding a new migration command or hidden state.

## PROBLEM

Kit stores machine-readable feature metadata in markdown body sections and tables. Relationships, dependencies, skills, notes links, design links, and prompt-routing inputs are parsed from human prose structures, which makes command behavior brittle, duplicates parsing logic, and forces validators, maps, prompts, and rollups to depend on body formatting instead of a typed document contract.

## GOALS

- Make YAML front matter the canonical source for structured feature-artifact metadata in newly generated and touched docs.
- Preserve existing projects by reading legacy body metadata as fallback during the migration window.
- Centralize metadata parsing, validation, and access through Kit's document layer.
- Keep feature documents readable as markdown without requiring Kit to understand the main body.
- Keep task progress and feature phase derived from human-editable files, not duplicated metadata.
- Let existing commands use structured metadata for just-in-time context loading, map rendering, validation, prompts, and reconciliation.
- Avoid adding public migration commands, hidden registries, hidden databases, or monolithic instruction files.

## NON-GOALS

- No new public command or public migration flag.
- No removal of legacy body fallback in the first implementation pass.
- No storage of task checkbox progress in front matter.
- No storage of feature phase in front matter.
- No canonical front matter requirement for `docs/PROJECT_PROGRESS_SUMMARY.md`.
- No hidden metadata database, lock file, generated registry, or external state.
- No promise that front-matter writes preserve YAML comments in v1.
- No automatic destructive rewrite of existing feature docs.
- No expansion of root instruction files or command surface.

## USERS

- Kit users who create and maintain feature docs through `kit brainstorm`, `kit spec`, `kit plan`, `kit tasks`, and `kit implement`.
- Coding agents that need deterministic, low-context routing through relationships, dependencies, skills, notes, and design inputs.
- Maintainers reviewing metadata drift through `kit check`, `kit check --project`, `kit reconcile`, `kit map`, `kit resume`, `kit rollup`, and prompt-producing commands.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| rlm | repo-local doc | docs/agents/RLM.md | analyze codebase, scan repository, large repository analysis, scan all files, recursive language model, or any front-matter migration task with broad command impact | yes |

## RELATIONSHIPS

- builds on: `0016-document-map-relationships`
- related to: `0004-brainstorm-first-workflow`
- related to: `0017-reconcile-command`
- related to: `0021-project-validation-and-instruction-registry`
- related to: `0025-v0-prompt-library`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Feature notes | notes | docs/notes/0026-front-matter-integration | optional pre-brainstorm research input; currently only `.gitkeep` exists | optional |
| Constitution | doc | docs/CONSTITUTION.md | markdown/YAML portability, document-source-of-truth rule, explicit filesystem state constraints | active |
| Agent routing docs | doc | docs/agents/README.md, docs/agents/WORKFLOWS.md, docs/agents/RLM.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md | spec workflow, RLM loading rules, source-of-truth semantics, validation expectations | active |
| References index | doc | docs/references/README.md | confirmed durable repo references are conditional inputs only | active |
| Brainstorm | feature doc | docs/specs/0026-front-matter-integration/BRAINSTORM.md | upstream findings, approved migration decisions, affected files | active |
| Kit map output | command output | `go run ./cmd/kit map 0026-front-matter-integration` | current feature phase, relationships, dependency baseline | active |
| Project progress summary | doc | docs/PROJECT_PROGRESS_SUMMARY.md | feature state and prior-feature shortlist | active |
| Brainstorm-first workflow | prior feature doc | docs/specs/0004-brainstorm-first-workflow/SPEC.md, docs/specs/0004-brainstorm-first-workflow/PLAN.md | workflow artifact and prompt-only precedents | active |
| Document map relationships | prior feature doc | docs/specs/0016-document-map-relationships/SPEC.md, docs/specs/0016-document-map-relationships/PLAN.md | current relationship graph contract and map behavior | active |
| Reconcile command | prior feature doc | docs/specs/0017-reconcile-command/SPEC.md, docs/specs/0017-reconcile-command/PLAN.md | prompt-only drift detection and audit behavior | active |
| Project validation and instruction registry | prior feature doc | docs/specs/0021-project-validation-and-instruction-registry/SPEC.md, docs/specs/0021-project-validation-and-instruction-registry/PLAN.md | `kit check --project` and registry validation precedent | active |
| Prompt library | prior feature doc | docs/specs/0025-v0-prompt-library/SPEC.md, docs/specs/0025-v0-prompt-library/PLAN.md | YAML config storage and prompt provider precedent | active |
| Document parser | code | internal/document/document.go, internal/document/section_content.go | markdown parsing, required-section validation, placeholder handling | active |
| Relationship parser | code | internal/document/relationships.go | existing relationship syntax, labels, and normalization behavior | active |
| Map builder | code | internal/feature/map.go, pkg/cli/map.go | relationship and dependency graph consumption and rendering | active |
| Reconcile/check validators | code | pkg/cli/check.go, pkg/cli/reconcile_audit_helpers.go, pkg/cli/reconcile_audit_tables.go | drift detection, project validation, section/table validation | active |
| Templates | code | internal/templates/templates.go | generated feature artifact shape | active |
| Rollup/status extraction | code | internal/rollup/rollup.go, internal/feature/status.go | summary and phase extraction behavior | active |
| Metadata mutation helpers | code | pkg/cli/brainstorm_notes.go, pkg/cli/profile_dependencies.go, pkg/cli/brainstorm_backlog.go | current body metadata mutation paths | active |
| YAML-node config helpers | code | internal/config/prompt_files.go | existing additive YAML write and unknown-field preservation pattern | active |
| Skill front-matter precedent | code | pkg/cli/skill_prompt.go | existing Kit guidance for YAML front matter in markdown skill files | active |
| Secondary global inputs | docs/skills | /Users/jamesonstone/.claude/CLAUDE.md, /Users/jamesonstone/.codex/AGENTS.md, /Users/jamesonstone/.codex/instructions.md, /Users/jamesonstone/.codex/skills/*/SKILL.md | inspected after repo-local inputs; no additional execution skill selected | optional |

## REQUIREMENTS

- [SPEC-01] `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` must support an optional YAML front-matter block at the start of the file.
- [SPEC-02] Front matter must be canonical when present; legacy markdown body metadata must be fallback only.
- [SPEC-03] New and touched canonical feature artifacts must include `kit_metadata_version: 1`.
- [SPEC-04] New and touched canonical feature artifacts must include an artifact type with one of these values: `brainstorm`, `spec`, `plan`, or `tasks`.
- [SPEC-05] New and touched canonical feature artifacts must include feature identity fields for numeric id, slug, and canonical feature directory.
- [SPEC-06] The feature directory must remain authoritative if front-matter feature identity conflicts with the filesystem path.
- [SPEC-07] Front matter must support relationship entries with a typed relationship kind and target feature directory.
- [SPEC-08] Relationship kinds must use stable machine values: `builds_on`, `depends_on`, and `related_to`.
- [SPEC-09] Human-facing renderers may display relationship labels as `builds on`, `depends on`, and `related to`, but machine storage must use the enum values in [SPEC-08].
- [SPEC-10] Front matter must support dependency entries with name, type, location, used-for text, and status.
- [SPEC-11] Dependency status must be one of `active`, `optional`, or `stale`.
- [SPEC-12] Front matter must support `SPEC.md` skill entries with name, source, path, trigger, and required fields.
- [SPEC-13] Front matter may support optional summary or intent fields for deterministic status and rollup display, but narrative context must remain in markdown body sections.
- [SPEC-14] Canonical self-links for `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` must be derived from the feature directory instead of stored redundantly.
- [SPEC-15] External links, notes links, design-material links, prior-feature links, and cross-artifact links that affect routing must be represented as structured relationships or dependencies.
- [SPEC-16] Task progress must remain derived from `TASKS.md` checkboxes and progress tables.
- [SPEC-17] Feature phase must remain derived from file existence and task completion state.
- [SPEC-18] `PROJECT_PROGRESS_SUMMARY.md` must remain generated markdown without requiring canonical front matter.
- [SPEC-19] Markdown section parsing must ignore front matter so required-section checks operate on the visible document body.
- [SPEC-20] Existing legacy `## RELATIONSHIPS`, `## DEPENDENCIES`, and `## SKILLS` sections must remain readable as fallback during the migration window.
- [SPEC-21] New templates must stop requiring duplicate body metadata tables when equivalent canonical front matter exists.
- [SPEC-22] `kit map` must read relationships and dependencies from front matter first, then legacy body metadata when front matter is absent.
- [SPEC-23] `kit check <feature>` must error on malformed front matter when a front-matter block exists.
- [SPEC-24] `kit check <feature>` must not warn solely because a legacy feature document lacks front matter during the initial migration.
- [SPEC-25] `kit check --project` must detect legacy-only metadata, missing canonical metadata in generated/touched docs, front-matter/body conflicts, malformed front matter, invalid enum values, and stale relationship targets.
- [SPEC-26] `kit reconcile` must remain prompt-only and must report front-matter migration needs, legacy-only metadata, stale links, and front-matter/body conflicts without writing changes.
- [SPEC-27] Workflow commands that create or mutate feature artifacts must seed or refresh front matter using non-destructive, append-only behavior where practical.
- [SPEC-28] Front-matter writes must preserve unknown fields and existing known values where practical.
- [SPEC-29] Front-matter/body conflicts must prefer front matter and surface as project-level warnings or reconcile findings unless the front matter itself is malformed.
- [SPEC-30] Malformed front matter must be treated as a validation error because the canonical metadata source cannot be trusted.
- [SPEC-31] Prompt-producing commands must instruct agents to maintain canonical front matter and use legacy body metadata only as fallback during migration.
- [SPEC-32] RLM guidance must remain intact: commands and prompts must route through the smallest relevant metadata and artifact links instead of inlining all feature docs by default.
- [SPEC-33] The implementation must not add `kit verify`, `kit validate`, `kit doctor`, `kit agent`, `kit instructions`, or any other new public command for this migration.
- [SPEC-34] The implementation must not remove existing public commands.
- [SPEC-35] Tests must cover parsing, validation, template generation, map rendering, reconcile/check drift detection, prompt behavior, and legacy fallback.

## ACCEPTANCE

- New `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` artifacts created by workflow commands contain valid front matter with metadata version, artifact type, and feature identity.
- Existing feature docs without front matter continue to parse and remain usable through legacy body metadata fallback.
- A document with front matter and a markdown body still passes required-section validation based on the body content only.
- Malformed front matter produces an actionable validation error with the affected file path.
- Invalid front-matter relationship kinds or dependency statuses produce actionable validation errors.
- When front matter and legacy body metadata conflict, map/prompt consumers use front matter and project-level validation or reconcile reports the drift.
- `kit map` renders the same relationship and dependency graph from front matter that it previously rendered from body metadata.
- `kit check <feature>` does not warn solely for missing front matter in legacy docs during the migration window.
- `kit check --project` identifies legacy-only metadata, missing canonical metadata in generated/touched docs, malformed front matter, stale relationship targets, and front-matter/body conflicts.
- `kit reconcile` reports front-matter migration findings without writing files.
- Workflow commands that already create or mutate metadata seed or refresh front matter without destructive body rewrites.
- New templates do not require duplicate body `RELATIONSHIPS`, `DEPENDENCIES`, or `SKILLS` tables when equivalent front matter exists.
- Prompt-producing commands preserve just-in-time context loading and reference canonical front matter instead of requiring full-context body-table reads.
- `go test ./...` passes after implementation.
- `kit check front-matter-integration` passes for this feature after implementation docs are complete.

## EDGE-CASES

- No front matter exists: read legacy body metadata, do not fail feature-local validation solely for migration status.
- Front matter delimiter appears later in the body: treat it as body text unless it starts the file as a valid block.
- Front matter exists but YAML is malformed: fail validation with file path and parse context.
- Front matter exists but required metadata fields are missing in a new or touched artifact: report as migration/schema drift in project validation or reconcile.
- Front matter feature identity differs from the directory: directory wins, and validation reports the mismatch.
- Relationship target is inline-code-wrapped in legacy body metadata: normalize it during fallback.
- Relationship target points to a missing feature directory: report stale relationship target.
- Dependency list is empty: treat it as no dependencies, not as malformed metadata.
- Dependency status is absent or unknown: report invalid metadata when front matter is present.
- Skills are present in both front matter and legacy body table: front matter wins and conflict is reported when values differ.
- Unknown front-matter fields exist: preserve them where practical and do not fail validation solely because they are unknown.
- `.gitkeep` appears in notes or design-material directories: ignore it as a placeholder dependency input.
- Existing body metadata sections contain `none`: treat as no legacy entries during fallback.
- `PROJECT_PROGRESS_SUMMARY.md` references missing downstream artifacts: preserve existing generated-summary behavior and do not require front matter.
- Project docs outside `docs/specs/<feature>/` have markdown front matter for other reasons: do not force feature-artifact schema onto non-feature docs.

## OPEN-QUESTIONS

none; brainstorm decision batches 1-3 resolved the migration scope, schema, precedence, validation, and command-surface assumptions.
