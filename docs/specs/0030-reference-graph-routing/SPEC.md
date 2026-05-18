---
kit_metadata_version: 1
artifact: spec
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
  - id: front-matter-integration-spec
    name: Front matter integration
    type: feature
    target: docs/specs/0026-front-matter-integration
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: prior metadata model and migration constraints
    status: active
  - id: project-map-builder
    name: Project map
    type: code
    target: internal/feature/map.go
    selector_type: symbol
    selector: BuildProjectMap
    relation: implements
    read_policy: must
    used_for: reference graph extraction and map rendering
    status: active
  - id: reconcile-prompt-builder
    name: Reconcile prompt
    type: code
    target: pkg/cli/reconcile_prompt.go
    selector_type: symbol
    selector: buildReconcilePrompt
    relation: implements
    read_policy: must
    used_for: migration prompt generation
    status: active
  - id: metadata-reference-schema
    name: Metadata parser
    type: code
    target: internal/document/metadata.go
    selector_type: symbol
    selector: MetadataReference
    relation: constrains
    read_policy: must
    used_for: canonical references schema and validation
    status: active
---
# SPEC

## SUMMARY

Replace front matter `dependencies` with canonical graph-like `references` that describe the relevant source, target, relationship, selector, read policy, use case, and status. Extend `kit map` to emit a focused read plan and extend `kit reconcile` with a migration prompt mode for converting old dependency metadata into the new reference graph.

## PROBLEM

The current `dependencies` metadata can point to exact file lines, which is brittle and can push agents toward stale or overly broad context. It also collapses different meanings into one list: constraints, implementation surfaces, evidence, prior work, design inputs, and stale context all look like generic dependencies. Kit needs a more precise routing structure that keeps context small without sacrificing correctness.

## GOALS

- Destructively migrate feature front matter from `dependencies` to `references`.
- Stop supporting front matter `dependencies` as a canonical or fallback metadata format.
- Preserve body `## DEPENDENCIES` sections as readable required document sections, but do not treat them as canonical metadata.
- Add a reference schema with graph-oriented fields: `id`, `name`, `type`, `target`, `selector_type`, `selector`, `relation`, `read_policy`, `used_for`, and `status`.
- Add validation for required reference fields, enum values, policy consistency, selector type consistency, and discouraged unpinned line references.
- Extend `kit map` to display reference edges with resolver status.
- Add a focused `kit map <feature> --context` read plan that lists only relevant references, de-duplicates repeated target selectors, and does not inline document contents.
- Add JSON map/context output for agent-readable graph plans.
- Extend `kit reconcile` with a migration flag that produces a prompt for old-format to new-format conversion.
- Update prompts, templates, docs, and repo artifacts to instruct agents to maintain `references`.

## NON-GOALS

- Adding a graph database.
- Automatically reading referenced files into prompts.
- Removing required markdown section headings from feature artifacts.
- Changing task dependency semantics in `TASKS.md`.
- Guaranteeing automatic migration of arbitrary third-party markdown outside Kit-managed docs.

## USERS

- Users who want agents to load only the documentation needed for the immediate decision.
- Maintainers who need deterministic validation of feature metadata.
- Agents that need a compact, explicit read plan instead of broad context dumps.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: 0026-front-matter-integration
- builds on: 0016-document-map-relationships
- related to: 0017-reconcile-command
- related to: 0029-scaffold-workflows-prepare

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

- [SPEC-01] `references` must be the canonical front matter field for context-routing metadata.
- [SPEC-02] Front matter `dependencies` must be treated as deprecated and invalid when present.
- [SPEC-03] The migration must be destructive in this pass: touched/generated Kit docs should use `references`, not duplicate `dependencies`.
- [SPEC-04] A reference must include `name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status`.
- [SPEC-05] `selector_type` and `selector` are optional, but must be supported for stable targeted references.
- [SPEC-06] Supported `relation` values must include `constrains`, `supports`, `implements`, `verifies`, `guides`, `informs`, `supersedes`, `conflicts_with`, and `uses`.
- [SPEC-07] Supported `read_policy` values must be `must`, `conditional`, `evidence`, and `skip`.
- [SPEC-08] Supported `status` values must remain `active`, `optional`, and `stale`.
- [SPEC-09] Validation must warn on unpinned line-number references and prefer stable selectors such as headings, symbols, artifact IDs, node IDs, URLs, or commit-pinned permalinks.
- [SPEC-10] `kit map` must render reference edges with target, selector, relation, read policy, and status.
- [SPEC-11] `kit map <feature> --context` must emit a compact read plan grouped by read policy and must not inline referenced document contents.
- [SPEC-12] `kit reconcile --migrate-references` must produce a prompt that consistently converts old `dependencies` metadata into new `references` metadata.
- [SPEC-13] Migration guidance must map old `location` to new `target`, add a relation, add a read policy, and replace exact line ranges with stable selectors when possible.
- [SPEC-14] Prompt-producing commands must instruct agents to keep `references` current and avoid teaching legacy dependency fallback.
- [SPEC-15] Existing repo docs touched by this work must be migrated to the new front matter format.
- [SPEC-16] `kit map` must resolve local file, feature artifact, heading, symbol, command, URL, and node-id references where possible.
- [SPEC-17] Context plans must merge duplicate target selectors and keep the strongest read policy.
- [SPEC-18] `kit map <feature> --json`, `kit map <feature> --context --json`, and `kit map --all --json` must emit deterministic machine-readable map or context-plan data.
- [SPEC-19] Reference relation wording must define the relation as the referenced target's role relative to the source artifact.

## ACCEPTANCE

- `go test ./internal/document ./internal/feature ./pkg/cli` passes.
- `kit check --project` reports old front matter `dependencies` as invalid.
- `kit reconcile --migrate-references` emits migration-specific instructions even when the normal audit has no other findings.
- `kit map <feature>` shows reference links instead of dependency links.
- `kit map <feature> --context` shows a read plan and does not print referenced file contents.
- `kit map <feature> --context --json` emits grouped context-plan JSON.
- Resolvable local references show resolved status in map output.
- New generated feature artifacts contain front matter `references` when metadata exists.
- Repo feature docs no longer contain front matter `dependencies`.

## EDGE-CASES

- A legacy document has both `dependencies` and `references`.
- A reference target points to a URL, a local path, a Figma node, a command flag, or a feature directory.
- A stale reference should remain visible in metadata but default to `read_policy: skip`.
- A line-number location can only be made stable by adding a descriptive selector.
- A feature has no references.
- `kit map --context` is run non-interactively without a feature argument.

## OPEN-QUESTIONS

- none
