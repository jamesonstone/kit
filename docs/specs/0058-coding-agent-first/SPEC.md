---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: validate
feature:
  id: 0058
  slug: coding-agent-first
  dir: 0058-coding-agent-first
relationships:
  - type: builds_on
    target: 0038-auto-improvement-v1
  - type: builds_on
    target: 0055-codex-thread-initialization
references:
  - id: accepted-plan
    name: Coding-agent-first major-release plan
    type: user-plan
    target: issue-133
    relation: governs
    read_policy: must
    used_for: product boundary and acceptance
    status: active
  - id: registry-predecessor
    name: Project validation and instruction registry
    type: spec
    target: docs/specs/0021-project-validation-and-instruction-registry/SPEC.md
    relation: builds_on
    read_policy: conditional
    used_for: provenance and local-custom merge semantics
    status: historical
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Ship Kit's next semantic-version major release as a provider-neutral,
repository-native contract layer for coding agents. Registry-backed rulesets
and declarative workflows become the product core; repository-local Markdown
remains authoritative and Kit resolves a deterministic machine-readable
contract without launching or supervising agents.

## CONTEXT

- Kit already distributes rulesets, records source provenance and hashes, and
  performs section-aware reconciliation, but those behaviors are embedded in
  a broad human-oriented CLI package.
- The existing product positioning centers repository memory and exposes many
  feature, prompt, loop, review, CI, and utility commands that are outside the
  accepted coding-agent-first boundary.
- The user explicitly selected an aggressive major-version reset rather than
  broad compatibility or dual-mode operation.
- `kit init`, `kit reconcile`, `kit rules add|list|view`, and
  `kit registry status` are the exhaustive protected legacy command paths.
- The separate `git-wt` binary remains unchanged.

## REQUIREMENTS

- Add a typed, versioned registry catalog for downstream-visible `ruleset` and
  declarative `workflow` artifacts with configurable source identity,
  provenance, applicability, dependencies, versions, and digests.
- Upgrade project configuration to schema v2 and record installed artifact
  state, materialized paths, source revisions, and content or section hashes.
- Add `kit contract resolve` as a local-only, read-only, deterministic JSON
  projection of the effective agent contract.
- Accept explicit feature, path, applicability-tag, and workflow hints without
  task-text or model inference.
- Block resolution when a required artifact is missing, invalid, conflicted,
  or has an unresolved dependency while still emitting diagnostic JSON.
- Make `kit init` materialize every downstream-visible ruleset and workflow,
  provider routing files, and provenance as one validated, idempotent plan.
- Make `kit reconcile` the canonical preview and apply surface for registry,
  configuration, and managed routing drift.
- Preserve local custom content, merge disjoint section edits, and refuse
  same-section conflicts unless the user explicitly accepts registry content
  for that exact artifact.
- Add generic `kit registry list|view|add|status`; implement protected
  `kit rules add|list|view` through the same core with a ruleset filter.
- Remove every other existing `kit` command and its dead implementation,
  configuration, templates, tests, and documentation.
- Keep `cmd/git-wt` and the worktree implementation behavior unchanged.
- Update README and canonical documentation concisely for the major release,
  including the primary `init -> resolve -> implement -> reconcile` flow and
  the schema-v1 migration path.
- Observable acceptance: a fresh fixture initializes all visible artifacts,
  resolves offline, preserves a local customization across a disjoint remote
  change, blocks a same-section conflict, and exposes only the protected plus
  new command surface.

## ACCEPTED PLAN

1. Extract a registry/catalog core and add typed workflow artifacts while
   retaining the proven provenance and section-merge semantics.
2. Add a separate contract resolver that reads only materialized local state
   and emits a stable v1 JSON contract.
3. Rebuild init, reconcile, rules, and registry commands as thin adapters over
   shared planning and application services.
4. Register only the protected commands, new registry administration, and
   contract resolution in the major-release root CLI; remove dead surfaces.
5. Upgrade `.kit.yaml`, routing documents, README, command guides, and release
   documentation without rewriting historical feature specifications.
6. Validate schemas, deterministic output, migration, merge safety,
   transactional initialization, command removal, git-wt preservation, full
   Go correctness, file-size limits, and hosted PR checks.

## DECISIONS

- This is a semantic-version major release; compatibility is intentionally
  exhaustive and limited to the six protected command paths and their purpose.
- Repository-local Markdown is canonical. Resolved JSON is a reproducible
  projection and never a second source of truth.
- Fresh initialization installs every registry artifact visible to downstream
  projects rather than prompting for a subset.
- Registry-backed workflows are instructions and gates, not executable agent
  runtimes.
- Contract resolution is deterministic and explicit; it does not inspect task
  prose or call a model.
- The generic registry lifecycle is canonical; `kit rules` is a filtered
  ruleset interface over it.
- Schema-v1 migration is previewed and applied through reconcile. No legacy
  runtime mode survives the migration.
- `git-wt` remains independently supported and unchanged.

## DISCOVERIES

- Current registry state already records source repo, branch, commit, path,
  installed hashes, and `managed`, `local-custom`, and `conflict` states.
- Current refresh code already performs section-aware three-way merges but is
  coupled to `pkg/cli` and a hard-coded GitHub source.
- `kit reconcile --include-files` currently delegates to the init refresh
  planner, providing a proven seam for one shared materialization service.
- The current configuration schema is version 1 and tracks only `ruleset`
  registry artifacts; typed workflows and resolver state require schema 2.
- Correct preservation across multiple reconciliation runs requires comparing
  local content to the installed registry digest while retaining registry
  section hashes as the three-way merge base. The last observed local digest
  is provenance, not the merge base.
- Catalog target moves need separate treatment for the previous path: unchanged
  managed files can be retired, while customized old paths must remain
  untouched as the new target is installed.
- Relative local registry paths must resolve from the discovered project root,
  not from the caller's current subdirectory.
- The network-backed registry provider made the repository's Go 1.25.5 pin an
  active security boundary; Go 1.25.12 is the smallest compatible patch with
  no reachable standard-library vulnerabilities in the final scan.

## VALIDATION

- `go test ./...` passed for `cmd/kit`, `cmd/git-wt`, `internal/contract`,
  `internal/registry`, `internal/worktree`, and `pkg/agentcli`.
- `go test -race ./internal/registry ./internal/contract ./pkg/agentcli`
  passed.
- `go vet ./...` and `golangci-lint run ./...` passed with zero issues.
- `goreleaser check` validated the release configuration and updated Kit
  version injection path.
- `goreleaser build --snapshot --clean` built both binaries for Linux, macOS,
  and Windows on amd64 and arm64.
- `govulncheck ./...` under Go 1.25.12 reported zero reachable
  vulnerabilities.
- Both `cmd/kit` and the unchanged `cmd/git-wt` built successfully.
- The release-tag computation selected `v2.0.0` from current tag `v1.0.122`
  and `MAJOR_VERSION=2`.
- Automated tests cover duplicate and invalid catalogs, dependency graphs,
  source and target confinement, GitHub revision pinning, digest mismatch,
  deterministic resolved-contract golden JSON, explicit hints, workflow
  dependencies, conditional and required blocking, initialization rollback and
  idempotency, schema-v1 migration, routing preservation, remote-only,
  local-only, missing, moved, retired, disjoint-section, conflicting-section,
  and exact-accept reconciliation states.
- The repository migrated itself through a local-catalog reconcile preview and
  explicit apply; follow-up `kit registry status --json` reported `current`
  with 16 artifacts and zero changes.
- `kit contract resolve --workflow implementation-delivery --path README.md`
  returned ready schema-v1 JSON, the explicit workflow, mandatory rules, the
  path-selected README rule, provenance, states, reasons, and next action.
- The CLI allowlist test proves only `init`, `reconcile`, `contract`,
  `registry`, and `rules` roots exist; a direct removed-command smoke test
  rejects `kit spec`.
- `git diff -- cmd/git-wt internal/worktree` is empty, proving the separate
  binary implementation is unchanged.
- The complete version-control-eligible handwritten source and test audit found
  no file above 300 physical lines.

## OUTCOME

- Kit now exposes a coding-agent-first contract architecture with a typed
  catalog, schema-v2 project provenance, stable resolved-contract v1 JSON,
  declarative workflows, bounded routing, and a reusable registry core.
- Fresh initialization installs all 14 downstream rulesets and two workflows;
  repeat initialization is idempotent and delegates drift maintenance to
  reconcile.
- Reconciliation is preview-first, section-aware, transactional on failure,
  conflict-preserving, path-move-aware, and exact-accept only for overwrites.
- The broad Kit 1.x runtime, commands, prompt templates, evaluations,
  configuration, and command-specific automation were removed. Historical
  specifications remain intact.
- `MAJOR_VERSION` now selects v2.0.0 from the current v1 tag line and subsequent
  mainline releases increment within major version 2.
- README, overview, Constitution, workflows, commands, migration, release, and
  generated agent-routing documentation now describe the new product model.
- Implementation and local acceptance are complete on `GH-133`; ready
  pull-request delivery and hosted-check verification are the remaining
  delivery actions.

## REPOSITORY MEMORY

This specification is the canonical rationale, implementation history,
validation record, and major-release migration decision. Durable contract,
registry, reconciliation, and product-boundary invariants were promoted to the
refactored `docs/CONSTITUTION.md`; reusable execution contracts live under
`docs/references/workflows/`, and operator migration details live under
`docs/migrations/` and `docs/releases/`.
