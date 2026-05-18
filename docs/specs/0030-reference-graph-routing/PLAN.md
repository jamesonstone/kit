---
kit_metadata_version: 1
artifact: plan
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
  - id: reference-graph-spec
    name: Reference graph spec
    type: feature
    target: docs/specs/0030-reference-graph-routing/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: implementation scope and acceptance criteria
    status: active
  - id: metadata-reference-schema
    name: Metadata parser
    type: code
    target: internal/document/metadata.go
    selector_type: symbol
    selector: MetadataReference
    relation: implements
    read_policy: must
    used_for: canonical schema and validation
    status: active
  - id: feature-map-references
    name: Feature map
    type: code
    target: internal/feature/map.go
    selector_type: symbol
    selector: FeatureMap
    relation: implements
    read_policy: must
    used_for: project graph and reference links
    status: active
  - id: reconcile-migration-flag
    name: Reconcile command
    type: code
    target: pkg/cli/reconcile.go
    selector_type: command
    selector: kit reconcile --migrate-references
    relation: implements
    read_policy: must
    used_for: migration prompt behavior
    status: active
---
# PLAN

## SUMMARY

Implement a destructive migration from front matter `dependencies` to graph-like `references`, then expose the new metadata through validation, map output, context read plans, reconcile migration prompts, templates, prompts, tests, and repo docs.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-19] Update the document metadata model so `references` is canonical, selector/policy validation is explicit, and legacy front matter `dependencies` is invalid.
- [PLAN-02][SPEC-10][SPEC-11][SPEC-16][SPEC-17][SPEC-18] Extend the project map to collect and resolve reference links, add `kit map <feature> --context` as a de-duplicated read-plan output, and support JSON output.
- [PLAN-03][SPEC-12][SPEC-13] Add `kit reconcile --migrate-references` so the command can produce a focused old-to-new migration prompt.
- [PLAN-04][SPEC-14][SPEC-15] Update prompts, templates, agent docs, README, and existing feature docs to use references.
- [PLAN-05] Update tests and run focused verification.

## COMPONENTS

- `internal/document/metadata.go`
  - metadata structs
  - enum validation
  - upsert behavior
  - diagnostics for deprecated `dependencies`
- `internal/document/metadata_accessors.go`
  - canonical reference accessor
  - removal of dependency fallback from canonical reads
- `internal/feature/map.go`
  - reference link collection and ordering
- `pkg/cli/map.go`
  - reference rendering
  - `--context` read-plan output
- `pkg/cli/reconcile.go`
  - migration flag
- `pkg/cli/reconcile_prompt.go`
  - migration prompt rules and verification guidance
- Prompt/template files under `pkg/cli` and `internal/templates`
  - references wording
  - no legacy dependency fallback
- Feature docs under `docs/specs`
  - destructive front matter migration

## DATA

- New front matter field:
  - `references: []`
- Reference fields:
  - `id`: optional stable identifier for updating the same conceptual edge
  - `name`: human label
  - `type`: reference category such as `doc`, `code`, `feature`, `notes`, `design`, `url`, `profile`, or `command`
  - `target`: stable path, URL, feature directory, command flag, or external node reference
  - `selector_type`: optional targeting strategy
  - `selector`: optional stable selector
  - `relation`: graph edge semantics
  - `read_policy`: context-loading policy
  - `used_for`: reason this reference matters
  - `status`: lifecycle status
- Validation rules:
  - `selector_type` is constrained to `artifact`, `heading`, `symbol`, `command`, `url`, or `node_id`.
  - `status: stale` should normally pair with `read_policy: skip`.
  - `relation: constrains` should normally pair with `read_policy: must`.
  - `relation: verifies` should normally pair with `read_policy: evidence`.

## INTERFACES

- `kit map <feature>`
  - Shows reference links and resolver status in the project map.
- `kit map <feature> --context`
  - Shows a de-duplicated read plan grouped by `must`, `conditional`, `evidence`, and `skip`.
- `kit map <feature> --context --json`
  - Emits the context plan as grouped JSON.
- `kit map --all --json`
  - Emits the full project map as deterministic JSON for automation.
- `kit reconcile --migrate-references`
  - Emits a prompt that converts old front matter `dependencies` to canonical `references`.

## DEPENDENCIES

References are tracked in front matter.

## RISKS

- A destructive migration can break old tests and old docs in one pass; mitigate by updating tests and running focused checks.
- Exact line ranges can look precise while becoming stale quickly; mitigate by warning and preferring selectors.
- Adding more enum fields can make metadata more verbose; mitigate with clear defaults in command-generated references.
- A read plan that inlines file contents would defeat the context-budget goal; keep map output pointer-only.

## TESTING

- Run:
  - `go test ./internal/document ./internal/feature ./pkg/cli`
  - `go run ./cmd/kit map reference-graph-routing`
  - `go run ./cmd/kit map reference-graph-routing --context`
  - `go run ./cmd/kit reconcile --migrate-references --output-only`
  - `go run ./cmd/kit check --project`
  - `go run ./cmd/kit check --all`
