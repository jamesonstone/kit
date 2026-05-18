---
kit_metadata_version: 1
artifact: plan
feature:
  id: "0026"
  slug: front-matter-integration
  dir: 0026-front-matter-integration
---
# PLAN

## SUMMARY

Implement front matter as a shared `internal/document` metadata layer, then migrate command consumers to read canonical metadata through that layer before falling back to legacy markdown sections. Keep required markdown headings for human readability and Constitution compatibility, but stop treating metadata tables as the primary machine contract when front matter exists.

## APPROACH

Sequence the implementation in four passes: establish the shared document metadata layer, migrate readers to front-matter-first accessors, migrate writers/templates/prompts to seed canonical metadata, then harden validation and tests around the dual-read migration.

- [PLAN-01][SPEC-01][SPEC-02][SPEC-19][SPEC-30] Add front-matter parsing to the document parser before section detection. Store the original file content, visible markdown body, parsed metadata, and metadata diagnostics so existing section validation continues to operate on body text only.
- [PLAN-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-10][SPEC-11][SPEC-12][SPEC-13] Define one typed metadata contract in `internal/document` for version, artifact, feature identity, relationships, dependencies, skills, and optional summary/intent.
- [PLAN-03][SPEC-02][SPEC-20][SPEC-29] Preserve legacy body parsers as fallback adapters. Front matter wins on conflict; legacy-only metadata stays readable; conflict diagnostics are exposed for project validation and reconcile.
- [PLAN-04][SPEC-21][SPEC-27][SPEC-28] Update template and mutation paths to seed or refresh front matter non-destructively. Keep section headings, but replace required metadata tables in new templates with concise body text when canonical front matter covers the same data.
- [PLAN-05][SPEC-22][SPEC-25][SPEC-26] Move map, project validation, and reconcile to shared metadata accessors instead of command-local table parsers. Keep `kit map` read-only and keep `kit reconcile` prompt-only.
- [PLAN-06][SPEC-16][SPEC-17][SPEC-18] Leave phase, task progress, and `PROJECT_PROGRESS_SUMMARY.md` body generation derived from files and checkboxes. Use optional summary/intent metadata only as an extraction aid with body fallback.
- [PLAN-07][SPEC-31][SPEC-32][SPEC-33][SPEC-34] Update prompt builders and agent guidance to tell agents to maintain canonical front matter while preserving RLM-style just-in-time context loading. Do not add or remove public commands.
- [PLAN-08][SPEC-23][SPEC-24][SPEC-25][SPEC-30][SPEC-35] Build validation in two tiers: feature-local checks fail malformed present front matter but do not warn solely for legacy docs, while project validation and reconcile report migration drift.
- [PLAN-09][SPEC-35] Add focused unit and CLI tests around parser behavior, metadata precedence, legacy fallback, templates, map output, project validation, reconcile findings, and prompt text.

Tradeoffs:

- Use a dual-read migration instead of a big-bang rewrite to avoid breaking existing projects.
- Keep metadata in artifact front matter instead of `.kit.yaml` so each document remains portable and self-describing.
- Preserve existing section headings because current validation, prompts, and Constitution language still rely on them; only the structured rows move to YAML.
- Do not promise YAML comment preservation in v1; preserving unknown fields and known values is the practical compatibility bar.
- Keep metadata helpers in `internal/document`, not `pkg/cli`, so CLI commands do not grow competing parsers.

## COMPONENTS

- [PLAN-01] `internal/document` front-matter parser
  - Splits a leading YAML front-matter block from visible markdown body.
  - Keeps line offsets accurate enough for actionable validation errors.
  - Ensures `GetSection`, `HasUnresolvedPlaceholders`, relationship validation, and first-paragraph extraction operate on body content.

- [PLAN-02] `internal/document` metadata model
  - Owns artifact enums, relationship enums, dependency status enums, feature identity, dependencies, skills, optional summary/intent, and diagnostics.
  - Maps legacy relationship labels to machine enum values and back to human labels for renderers.
  - Rejects malformed front matter and invalid enum values with file-specific errors.

- [PLAN-03] `internal/document` metadata accessors and writers
  - Exposes relationships, dependencies, and skills through front-matter-first accessors.
  - Provides legacy fallback readers for existing `RELATIONSHIPS`, `DEPENDENCIES`, and `SKILLS` sections.
  - Provides non-destructive upsert helpers for commands that already create or mutate metadata.

- [PLAN-04] Templates and scaffolded artifacts
  - Seeds `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` with canonical front matter.
  - Retains required section headings with body text that points humans to canonical metadata where appropriate.
  - Avoids duplicate machine-readable metadata tables in newly generated docs unless a command explicitly renders one for display.

- [PLAN-05] Metadata consumers
  - `internal/feature/map.go` reads relationships and dependencies through shared accessors.
  - `pkg/cli/check.go` and reconcile audit helpers consume metadata diagnostics and conflict reports.
  - Rollup/status readers optionally use summary/intent metadata, then fall back to body sections.
  - Prompt builders replace table-specific instructions with front-matter-first maintenance language.

- [PLAN-06] Existing mutation paths
  - `pkg/cli/brainstorm_notes.go`, `pkg/cli/profile_dependencies.go`, and `pkg/cli/brainstorm_backlog.go` move from markdown-table/section writes to metadata upserts.
  - Existing body mutation behavior remains fallback-only for legacy docs when a safe front-matter write is not available.

- [PLAN-07] Test surface
  - `internal/document` tests own parser, schema, fallback, diagnostics, and writer behavior.
  - `internal/feature` and `pkg/cli` tests own consumer behavior and user-visible output.

## DATA

Canonical front matter:

```yaml
kit_metadata_version: 1
artifact: spec
feature:
  id: "0026"
  slug: front-matter-integration
  dir: 0026-front-matter-integration
summary: optional one-line summary
intent: optional one-line intent
relationships:
  - type: builds_on
    target: 0016-document-map-relationships
references:
  - name: Feature notes
    type: notes
    target: docs/notes/0026-front-matter-integration
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
skills:
  - name: rlm
    source: repo-local doc
    path: docs/agents/RLM.md
    trigger: broad/noisy context routing
    required: true
```

Enums:

- `artifact`: `brainstorm`, `spec`, `plan`, `tasks`.
- `relationship.type`: `builds_on`, `depends_on`, `related_to`.
- Human relationship labels: `builds on`, `depends on`, `related to`.
- `reference.status`: `active`, `optional`, `stale`.

Derived state:

- Feature phase stays derived from file existence and task completion.
- Task progress stays derived from `TASKS.md` checkbox and progress-table content.
- Canonical artifact self-links are derived from `docs/specs/<feature>/`.
- `PROJECT_PROGRESS_SUMMARY.md` remains generated markdown.

Diagnostics:

- Malformed front matter is an error.
- Invalid enum values are errors.
- Missing front matter in legacy docs is a project-level migration finding, not a feature-local warning.
- Front-matter/body conflicts are warnings or reconcile findings; front matter remains authoritative.

## INTERFACES

- `document.Parse` / `document.ParseFile`
  - Side effect: none.
  - Contract: parse optional leading front matter, parse body sections from visible markdown, and retain diagnostics for validation.

- `Document.Validate`
  - Side effect: none.
  - Contract: validate body sections plus present front matter; fail malformed present front matter; do not require front matter for feature-local legacy docs.

- Metadata accessors
  - Side effect: none.
  - Contract: return canonical relationships, dependencies, skills, summary, intent, and conflict diagnostics using front matter first and body fallback second.

- Metadata upsert helpers
  - Side effect: write the target markdown artifact.
  - Contract: seed or refresh front matter without destructive body rewrites; preserve unknown front-matter fields where practical.

- `kit map [feature]`
  - Output contract unchanged.
  - Reads relationships/references through shared metadata accessors.
  - Continues to render unresolved relationship targets and warnings deterministically.

- `kit check <feature>`
  - Errors on malformed present front matter and invalid front-matter schema.
  - Does not warn only because a legacy feature lacks front matter.

- `kit check --project`
  - Reports legacy-only metadata, missing canonical metadata in generated/touched docs, stale relationship targets, malformed front matter, invalid enum values, and front-matter/body conflicts.

- `kit reconcile [feature]`
  - Remains prompt-only.
  - Reports front-matter migration findings and exact files/search hints without writing changes.

- Workflow commands
  - `kit brainstorm`, `kit spec`, `kit plan`, and `kit tasks` generate front matter in new artifacts.
  - Existing metadata mutation paths upsert front matter in docs they already touch.
  - Prompt-only modes describe the canonical front-matter contract and legacy fallback behavior.

- Rollup/status/resume/handoff/summarize prompts
  - Use metadata where relevant but preserve body fallback and RLM context routing.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Constitution | doc | docs/CONSTITUTION.md | document-source-of-truth, markdown/YAML portability, filesystem state, command-surface constraints | active |
| Brainstorm | feature doc | docs/specs/0026-front-matter-integration/BRAINSTORM.md | upstream code findings, migration decisions, affected files | active |
| Specification | feature doc | docs/specs/0026-front-matter-integration/SPEC.md | binding requirements and acceptance criteria | active |
| Agent routing docs | doc | docs/agents/README.md, docs/agents/WORKFLOWS.md, docs/agents/RLM.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md | RLM discovery, source-of-truth order, validation expectations | active |
| Project progress summary | doc | docs/PROJECT_PROGRESS_SUMMARY.md | current feature phase and prior-feature shortlist | active |
| Kit map output | command output | `go run ./cmd/kit map 0026-front-matter-integration` | relationship/dependency baseline and scope confirmation | active |
| Brainstorm-first workflow | prior feature doc | docs/specs/0004-brainstorm-first-workflow/SPEC.md, docs/specs/0004-brainstorm-first-workflow/PLAN.md | reference inventory and workflow artifact precedents | active |
| Document map relationships | prior feature doc | docs/specs/0016-document-map-relationships/SPEC.md, docs/specs/0016-document-map-relationships/PLAN.md | existing relationship contract and map rendering behavior | active |
| Reconcile command | prior feature doc | docs/specs/0017-reconcile-command/SPEC.md, docs/specs/0017-reconcile-command/PLAN.md | prompt-only audit behavior and relationship-target drift checks | active |
| Project validation and instruction registry | prior feature doc | docs/specs/0021-project-validation-and-instruction-registry/SPEC.md, docs/specs/0021-project-validation-and-instruction-registry/PLAN.md | project-level validation and shared registry precedent | active |
| Prompt library | prior feature doc | docs/specs/0025-v0-prompt-library/SPEC.md, docs/specs/0025-v0-prompt-library/PLAN.md | YAML-backed prompt storage and prompt-output compatibility | active |
| Document parser | code | internal/document/document.go, internal/document/section_content.go | parser seam, section validation, placeholder handling | active |
| Relationship parser | code | internal/document/relationships.go | legacy relationship fallback and label normalization | active |
| Map builder | code | internal/feature/map.go, pkg/cli/map.go | graph and dependency-link consumer | active |
| Check/reconcile validators | code | pkg/cli/check.go, pkg/cli/reconcile_audit_helpers.go, pkg/cli/reconcile_audit_tables.go | validation, migration drift, table-contract replacement surface | active |
| Templates | code | internal/templates/templates.go | generated artifact shape | active |
| Metadata mutation helpers | code | pkg/cli/brainstorm_notes.go, pkg/cli/profile_dependencies.go, pkg/cli/brainstorm_backlog.go | existing dependency/relationship mutation paths | active |
| Rollup/status extraction | code | internal/rollup/rollup.go, internal/feature/status.go | summary, intent, phase, and progress extraction behavior | active |
| Prompt builders | code | pkg/cli/spec_context.go, pkg/cli/plan.go, pkg/cli/implement.go, pkg/cli/handoff_prompt.go, pkg/cli/summarize.go, pkg/cli/resume.go | agent-facing instruction updates | active |
| YAML-node config helpers | code | internal/config/prompt_files.go | additive YAML write pattern and unknown-field preservation precedent | active |
| Feature notes | notes | docs/notes/0026-front-matter-integration | optional pre-brainstorm inputs; only `.gitkeep` exists | optional |
| Secondary global inputs | docs/skills | /Users/jamesonstone/.claude/CLAUDE.md, /Users/jamesonstone/.codex/AGENTS.md, /Users/jamesonstone/.codex/instructions.md, /Users/jamesonstone/.codex/skills/*/SKILL.md | inspected after repo-local docs; no plan-shaping skill selected | optional |

## RISKS

- Risk: dual-read behavior creates two sources of truth for longer than intended.
  - Mitigation: centralize precedence in `internal/document`; expose conflicts to `kit check --project` and `kit reconcile`; keep front matter authoritative.

- Risk: changing section parsing to ignore front matter can break existing validation or summary extraction.
  - Mitigation: preserve original content, visible body content, and existing section APIs; add parser regression tests with and without front matter.

- Risk: template changes conflict with Constitution-required section headings.
  - Mitigation: keep headings required and populated; move only structured metadata rows to front matter.

- Risk: command-local dependency parsers keep drifting after the metadata layer lands.
  - Mitigation: route `map`, frontend-profile dependency detection, notes dependencies, and reconcile checks through shared accessors.

- Risk: front-matter writes accidentally discard unknown fields or body content.
  - Mitigation: use `yaml.Node`-style upserts, preserve body bytes outside the front-matter block, and test unknown-field preservation.

- Risk: project-level migration warnings become too noisy for legacy repositories.
  - Mitigation: keep feature-local checks quiet for missing front matter, classify migration-only findings as warnings, and reserve errors for malformed present metadata.

- Risk: prompt builders keep telling agents to update legacy tables.
  - Mitigation: update shared prompt text and focused command prompts together; test representative prompt output.

## TESTING

- Unit tests for `internal/document` parsing:
  - no front matter
  - valid front matter plus required body sections
  - malformed YAML
  - delimiter inside body
  - invalid artifact, relationship type, and dependency status
  - unknown-field preservation on write

- Unit tests for metadata accessors:
  - front-matter relationships/references/skills
  - legacy fallback
  - front-matter/body conflict precedence
  - inline-code legacy relationship target normalization
  - empty dependency arrays and legacy `none`

- Template tests:
  - generated artifacts contain required front matter
  - required markdown section headings remain present and populated
  - new generated docs avoid duplicate metadata tables where canonical front matter covers the data

- Map tests:
  - `kit map` reads front-matter relationships and dependencies
  - unresolved relationship targets remain visible
  - output ordering stays stable
  - legacy docs render through fallback

- Check/reconcile tests:
  - `kit check <feature>` fails malformed present front matter
  - `kit check <feature>` does not warn for legacy-only missing front matter
  - `kit check --project` reports legacy-only metadata, conflicts, invalid enums, missing canonical metadata in generated/touched docs, and stale targets
  - `kit reconcile` reports migration findings without writing

- Mutation-path tests:
  - notes dependency and frontend profile/design dependency upsert front matter
  - backlog relationship upserts front matter
  - unknown fields and body content remain intact

- Prompt tests:
  - spec/plan/tasks/implement/handoff/resume/summarize prompts refer to canonical front matter and legacy fallback
  - RLM language continues to load only relevant linked context

- End-to-end validation:
  - `go test ./...`
  - `go run ./cmd/kit check front-matter-integration`
  - targeted `go run ./cmd/kit map front-matter-integration`
  - `go run ./cmd/kit check --project` after expected migration findings are either implemented as warnings or the repository docs are updated.
