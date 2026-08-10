---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
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
- Pull-request feedback can arrive asynchronously after local implementation
  and after a provider status first becomes successful. A successful provider
  context is not sufficient evidence of completed review: the description may
  instead report a terminal skipped review with its exact reason.
- `kit init`, `kit reconcile`, `kit rules add|list|view`,
  `kit registry status`, and `kit pr fix` are the exhaustive protected legacy
  command paths.
- The separate `git-wt` binary remains unchanged.
- The first major-release implementation narrowed `kit init` too far: it
  materialized registry artifacts and routing, but omitted the complete
  repository bootstrap that existing users depend on.
- The accepted correction makes `kit init` the canonical bootstrap command.
  Its compatibility boundary includes the historical scaffold footprint and
  semantic coding-agent bootstrap prompt, even though the rest of the legacy
  CLI remains intentionally removed.
- The initial coding-agent-first implementation left feature specifications
  conditional on whether an agent judged rationale to be material and removed
  the former detailed template enforcement. An empty `docs/specs/` bootstrap
  plus advisory wording is insufficient to preserve feature history.

## SOURCE EVIDENCE AND HISTORICAL RELATIONSHIPS

- `docs/specs/0021-project-validation-and-instruction-registry/SPEC.md`
  establishes the registry provenance and reconciliation model extended here.
- Historical specifications under `docs/specs/` demonstrate the repository's
  long-lived feature rationale and remain evidence even when the CLI or runtime
  they describe is removed.
- The current `implementation-delivery` workflow and generated agent guidance
  make `SPEC.md` conditional, while the resolver has no feature-specification
  state. Those concrete artifacts are the implementation gap corrected here.
- SPEC 0058 is the current living V3 exemplar and must itself retain every
  required lifecycle section while this major release evolves.

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
- Add a registry-backed `pr-feedback-repair` workflow with an explicit
  `ruleset/agent-team-orchestration` dependency and every applicable delivery,
  lane, safety, testing, and source-size dependency.
- Restore `kit pr fix` as a narrow prompt-producing fallback over the
  declarative `pr-feedback-repair` workflow. The explicit adapter may call
  GitHub, select or prepare the exact writable same-repository PR-head lane,
  collect current active feedback, and explicitly resolve named verified
  threads; it must not launch agents, edit source, stage, commit, push, post
  comments, merge, or add network behavior to contract resolution.
- Define bounded await and event-triggered or one-shot collect intake modes
  over one repair workflow. Native host wakeups are preferred; a single
  token-free `gh` helper process is the bounded fallback.
- Make provider completion, skip, failure, head-change, unavailability,
  timeout, and rate-limit outcomes explicit. Pending and timeout are never
  clean completion.
- Specify rate-conscious status observation, quiet-window confirmation,
  paginated active-feedback collection, durable fingerprints, watcher
  deduplication, and bounded head epochs and repair passes as a structured,
  testable workflow contract.
- Make fresh init create or boundedly merge `.gitignore`, `.env`, `.envrc`,
  `Makefile`, `.coderabbit.yaml`, GitHub templates and auto-assignment,
  `README.md`, `docs/CONSTITUTION.md`,
  `docs/PROJECT_PROGRESS_SUMMARY.md`, `docs/specs/`, provider routers, the
  complete `docs/agents/` RLM tree, and project reference starters.
- Preserve existing project-owned content. Empty environment files and safe
  starters are create-if-missing; ignore patterns are append-only; README,
  Constitution, and routing changes are bounded managed sections.
- Restore safe user defaults at `~/.config/kit/.kit.yaml`, preserve unrelated
  existing keys, and consume those defaults for bootstrap and registry source
  selection without moving project provenance out of project `.kit.yaml`.
- Add the registry-backed `repository-bootstrap` workflow. The init prompt
  directs a coding agent to resolve it locally, progressively inspect actual
  repository evidence, and populate only demonstrated project context and
  verified command entrypoints.
- Preserve clipboard-first init prompt behavior through `--output-only` and
  `--copy`, while keeping `--dry-run` and `--json` strictly write-free and
  secret-free.
- Keep `docs/specs/` as the canonical feature-history container. Fresh init
  creates only the directory and never invents a feature or restores
  `docs/specs/0000_INIT_PROJECT.md` downstream.
- Require every non-trivial feature implementation to create or adopt a living
  `docs/specs/<feature>/SPEC.md` before source implementation. This is a hard
  workflow phase gate, not an agent discretion or rationale-materiality test.
- Define and validate the living V3 SPEC contract: purpose, context, source
  evidence and historical relationships, requirements, non-goals, observable
  acceptance criteria, accepted plan, architecture and consequential decisions
  including superseded alternatives, discoveries, validation plan and literal
  results, actual outcome, delivery evidence, and repository-memory
  disposition.
- Make `workflow/implementation-delivery` unconditionally depend on a
  downstream `ruleset/feature-specification`, with explicit pre-source,
  pre-delivery, and completion gates.
- For an explicit `--feature` resolution, report the canonical spec path,
  structural state, required and missing sections, relevant historical specs,
  and phase permissions. Missing or incomplete specs permit spec authoring but
  block source implementation and delivery until the agent updates the spec
  and re-resolves.
- Remove `docs/notes` and `ruleset/feature-notes` from Kit's active product
  model and generated guidance without deleting downstream project-owned notes
  or rewriting historical specifications that mention them.

## NON-GOALS

- Do not restore `kit spec`, staged feature lifecycle commands, or broad legacy
  CLI compatibility to enforce the feature-history contract.
- Do not make Kit infer feature truth, author a feature spec, execute a plan,
  launch an agent, or mutate source through contract resolution.
- Do not invent a bootstrap feature, scaffold `0000_INIT_PROJECT.md`, delete
  existing downstream `docs/notes`, or mechanically rewrite historical specs.
- Do not turn historical specifications into current CLI documentation; agents
  use them as progressively disclosed rationale alongside current evidence.

## ACCEPTANCE CRITERIA

- Fresh init creates `docs/specs/` and no invented feature specification.
- Resolving `implementation-delivery` selects `feature-specification`
  transitively; resolving with `--feature` exposes deterministic feature-spec
  state and relevant historical specification paths.
- Missing, invalid, or structurally incomplete feature specs keep spec
  authoring allowed while source implementation and delivery are blocked;
  a structurally complete V3 spec unlocks those phases after re-resolution.
- The V3 structural contract requires every detailed lifecycle section and
  preserves material alternatives, discoveries, literal validation, outcome,
  delivery, and repository-memory disposition.
- Routing and workflow guidance require the spec before source edits and keep
  it live through integrated-diff validation and completion.
- Active catalogs and generated guidance do not install, select, scaffold, or
  recommend `docs/notes`; existing downstream content and historical specs are
  not automatically removed or rewritten.
- The reduced command tree remains exact and does not regain `kit spec` or any
  other removed feature lifecycle command.

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
7. Add the asynchronous PR-feedback workflow; validate its structured state
   classifier, request budget, dependencies, feedback sources, and resolver
   selection, then reconcile Kit's own catalog provenance and routing counts.
8. Add a bootstrap planner around the registry init plan so the complete
   repository scaffold, empty specs directory, project provenance, and
   separately reversible user-config merge are validated before mutation and
   applied atomically from the user's perspective.
9. Restore deterministic starter templates and the semantic bootstrap prompt,
   register `repository-bootstrap`, and cover the exact path allowlist,
   preservation, secrecy, rollback, idempotency, and prompt-output modes.
10. Restore only the `kit pr fix` compatibility path as a thin GitHub and
    worktree adapter. Reuse the resolved workflow for policy, retain
    clipboard/output/editor controls, add bounded await plus immediate collect,
    and require exact head and thread identifiers for explicit resolution.
11. Replace the active feature-notes artifact with a mandatory downstream
    feature-specification ruleset and make it an unconditional dependency of
    implementation delivery.
12. Extend local feature resolution with deterministic living-spec structural
    inspection, phase permissions, historical-spec discovery, diagnostics, and
    re-resolution actions while preserving schema-v1 compatibility and the
    resolver's read-only, local-only boundary.
13. Tighten routing, workflow, migration, release, and command documentation;
    remove active `docs/notes` guidance; and add focused/golden coverage for the
    feature-history harness without rewriting historical specifications.

## ARCHITECTURE AND DECISIONS

- This is a semantic-version major release; compatibility is intentionally
  exhaustive and limited to the seven protected command paths and their
  purpose.
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
- Asynchronous GitHub intake normally belongs to the coding agent or host. The
  protected `kit pr fix` fallback may perform explicit bounded GitHub intake,
  lane preparation, and named-thread resolution while Kit continues to own the
  declarative workflow, contract validation, and local resolver projection.
- `kit pr fix` is not a repair runtime. It prepares and records the safe lane,
  generates the supervisor prompt, and optionally waits or resolves explicitly
  named verified threads; the coding agent owns reasoning, edits, validation,
  delivery, reflection, and repair decisions.
- `SUCCESS` plus a `Review completed` description means provider completion;
  `SUCCESS` plus `Review skipped:` is a terminal non-clean skip whose suffix is
  preserved exactly. Unknown success descriptions fail closed as unavailable.
- Await mode uses a bounded, jittered schedule rather than constant polling;
  collect mode can run once long after provider completion and includes active
  human feedback by default.
- A repair supervisor owns one writable PR-head lane and at most three default
  independent repair lanes, with a hard maximum of four. Shared files are
  serialized, excess work is queued, and nontrivial repairs receive a
  read-only verification lane.
- Canonical bootstrap compatibility is deliberately broader than ordinary
  protected-command compatibility: `kit init` restores its complete scaffold
  duty, while `kit pr fix` is the sole restored prompt fallback. No dispatch,
  loop, broader PR, feature, review, or lifecycle command returns.
- Kit writes deterministic valid starters and resolves a declarative
  `repository-bootstrap` contract; the coding agent, not Kit, infers project
  truth from repository evidence.
- `.env` is never read by bootstrap planning, prompt rendering, provenance, or
  JSON output. Existing `.env` and `.envrc` content is never overwritten, and
  Kit never grants direnv trust.
- Existing schema-v2 projects may use repeated init to backfill missing
  bootstrap starters without reconciling registry or routing drift. Schema-v1
  projects still fail closed into a previewed reconcile migration.
- Legacy `BRAINSTORM.md`, `PLAN.md`, `TASKS.md`,
  `docs/specs/0000_INIT_PROJECT.md`, runtime state/evidence, transcripts,
  mined skills, Kit-internal fixtures, and downstream copies of Kit product
  guides remain excluded.
- A non-trivial feature is represented by an explicit feature hint and the
  implementation-delivery workflow. The resolver does not infer feature scope
  from prose; agents must resolve with `--feature <feature>` before feature
  source work.
- Feature-spec authoring and source implementation are distinct phases in one
  resolved contract. Missing or incomplete SPEC content blocks source and
  delivery, but never blocks the documentation work needed to satisfy the gate.
- Structural completeness is deterministic Markdown/front-matter validation,
  not semantic truth inference. The coding agent remains accountable for
  evidence, accepted-plan fidelity, and keeping the living document accurate.
- `feature-specification` replaces `feature-notes` in the typed catalog so the
  artifact count remains stable while the mandatory dependency set becomes
  stronger. Existing downstream note directories remain project-owned data.
- Historical specs are immutable rationale inputs discovered through the
  specs index, progress summary, explicit relationships, and repository
  evidence. Their mention of removed concepts does not make those concepts
  active again.

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
- CodeRabbit reports both `Review completed` and `Review skipped: <reason>`
  through successful status contexts, so state and exact description must be
  classified together. Review threads are a separate paginated source and
  must not be fetched during status polling.
- Workflow-specific structured front matter can make asynchronous intake
  policy testable without adding a network command or changing resolved
  contract schema v1. Catalog and front-matter dependencies must match so an
  orchestration dependency cannot silently disappear during materialization.
- One compact GraphQL observation can return PR state, head SHA, provider
  status and description, and rate budget at a verified current cost of one;
  the durable contract still consumes the returned cost rather than assuming
  that cost permanently.
- Historical init behavior was clipboard-first and append-only for
  `.gitignore`; it created an empty `.env`, a non-trusting `.envrc`, and a
  `.DEFAULT_GOAL := help` Makefile without guessed project commands. Those
  semantics remain appropriate for deterministic starters.
- The current resolver validates catalog artifacts only. It requires a
  separate feature-spec state projection so incomplete project-owned feature
  memory cannot be mistaken for registry drift or repaired by reconcile.
- Existing agent docs explicitly say feature specs are conditional, confirming
  that documentation-only tightening would still leave machine resolution
  unable to enforce the required pre-source gate.
- Resolved-contract v1 can preserve compatibility through one optional
  `feature_spec` projection. It reports deterministic structural state and
  separate spec-authoring, source-implementation, and delivery permissions
  without changing local-only resolver ownership or requiring a new command.
- A safe feature hint maps to exactly one project-confined
  `docs/specs/<feature>/SPEC.md`; relationship targets map only to existing
  confined historical SPEC paths, while the progress summary and specs
  directory remain lightweight RLM indices.
- Replacing `feature-notes` with `feature-specification` keeps the downstream
  catalog at 14 rulesets and four workflows while changing implementation
  delivery from five to six explicit dependencies. Reconcile retires only the
  old managed artifact record and does not scan or delete downstream notes.

## VALIDATION PLAN AND MAP

- Catalog and dependency tests prove `feature-specification` replaces
  `feature-notes`, is downstream-visible, and is selected transitively by
  implementation delivery.
- Resolver unit and JSON-schema tests cover missing, invalid, incomplete, and
  complete feature specs; spec-authoring/source/delivery permissions; stable
  required-section diagnostics; deterministic historical discovery; and
  strictly local read-only behavior.
- Bootstrap and command-tree tests prove fresh init creates only the empty
  specs container and no removed feature lifecycle surface returns.
- Documentation and routing assertions prove the pre-source, live-update,
  pre-delivery, and completion gates while excluding active `docs/notes`
  guidance.
- Full Go tests, focused race tests, formatting, vet, lint, security checks,
  binary and release builds, the affected-source 300-line audit, self-hosted
  resolution, exact pushed-head verification, and hosted PR status complete
  the integrated acceptance evidence.

## VALIDATION

- `go test ./...` passed for `cmd/kit`, `cmd/git-wt`, `internal/contract`,
  `internal/registry`, `internal/worktree`, and `pkg/agentcli`.
- `go test -race ./internal/bootstrap ./internal/registry
  ./internal/contract ./pkg/agentcli` passed.
- The all-package race attempt reached the unchanged `internal/worktree` PTY
  cancellation helper and failed twice because the selector returned before
  cancellation. The normal complete suite passes and
  `git diff -- cmd/git-wt internal/worktree` remains empty; no out-of-scope
  timing change was made.
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
- The repository reconciled itself through a local-catalog preview and explicit
  apply; follow-up `kit registry status --json` reported `current` with 18
  artifacts and zero changes.
- `kit contract resolve --workflow implementation-delivery --path README.md`
  returned ready schema-v1 JSON, the explicit workflow, mandatory rules, the
  path-selected README rule, provenance, states, reasons, and next action.
- The CLI allowlist test proves only `init`, `reconcile`, `contract`, `pr`,
  `registry`, and `rules` roots exist; `pr` exposes only `fix`, and a direct
  removed-command smoke test rejects `kit spec`.
- `git diff -- cmd/git-wt internal/worktree` is empty, proving the separate
  binary implementation is unchanged.
- The complete version-control-eligible handwritten source and test audit found
  no file above 300 physical lines.
- Bootstrap tests assert the exact 24-file fresh-init allowlist plus
  `docs/specs/`, create-if-missing preservation, append-only ignore updates,
  bounded README/Constitution/routing sections, local-custom preservation,
  `.env` secrecy, `.envrc` no-trust behavior, user-default consumption,
  schema-v1 fail-closed handling, rollback, idempotency, prompt golden output,
  and write-free JSON and dry-run modes.
- A self-hosted fresh fixture initialized from the local catalog, created an
  empty `.env`, resolved `repository-bootstrap` as ready, reported registry
  state `current`, and returned a write-free repeated-init plan with exactly 24
  bootstrap file dispositions.
- `gosec ./internal/bootstrap ./pkg/agentcli` reported zero issues and
  `gitleaks dir . --redact` found no leaks.
- PR-feedback tests cover completed versus skipped successful contexts, unknown
  success, provider failure, head change, timeout, HTTP 403/429 retry evidence,
  rate reserve, bounded schedule and quiet confirmation, watcher keys and host
  wakeups, collection pagination, human sources, fingerprints, prompt and
  trusted-comment markers, late collect mode, dependency drift, and hard
  timeout/request/page/head/pass ceilings.
- Protected-adapter tests cover all accepted target forms, interactive
  selection, human and CodeRabbit filtering, prompt extraction and
  deduplication, exact writable-head lane checks, NUL-safe dirty paths,
  clipboard/output/editor controls, subagent bounds, prompt-only behavior,
  structured wait failures, and explicit current-thread resolution.
- Self-hosted one-shot collection against PR #134 found no active thread,
  review-body, or marked-comment feedback and kept `--output-only` stdout
  empty. Bounded await classified the live successful CodeRabbit context as
  `skipped-with-reason`, preserved its exact `Review skipped: 541 files exceed
  the limit of 300` reason and rate evidence in schema-v1 JSON, and exited 2.
- `gosec ./internal/prfix ./pkg/agentcli` reported zero issues. The whole-tree
  scan retained six pre-existing findings only in unchanged `internal/registry`
  and `internal/worktree` paths; no finding points to the restored adapter.
- `kit contract resolve --workflow pr-feedback-repair` returned ready local-only
  JSON containing the workflow, `agent-team-orchestration`, GitHub delivery,
  safety, work-lane, testing, and source-size dependencies. The exact CLI
  allowlist exposes only the protected `pr fix` path under the `pr` root.
- Final validation passed `go test ./...`, focused race tests, `go vet ./...`,
  `golangci-lint run ./...` with zero issues, `govulncheck ./...` with zero
  reachable vulnerabilities, both direct binary builds, `goreleaser check`,
  and the full snapshot build matrix.
- Ready PR #134 remains the single delivery lane for the original major reset
  and its accepted bootstrap and `pr fix` corrections. GitHub reports the
  issue and PR assigned to `jamesonstone`; exact pushed-head and hosted-state
  evidence is refreshed after every update rather than treating a successful
  CodeRabbit context as completed review.
- Feature-spec resolver tests cover missing, invalid, incomplete, ready, and
  delivery-ready states; safe feature paths; spec-authoring/source/delivery
  permissions; exact required V3 sections; transitive rule selection;
  deterministic historical relationships; and the stable projection golden.
- The self-hosted SPEC 0058 resolution reports `ready`, workflow version 3,
  zero missing sections, source and delivery permission, both declared
  historical specs, and the progress-summary/specs-directory RLM indices.
- A fresh external fixture initialized all 14 rulesets and four workflows,
  created `docs/specs/` without `0000_INIT_PROJECT.md` or `docs/notes`, resolved
  repository bootstrap, and returned exit 2 plus spec-authoring-only permission
  for a missing feature spec.
- Repository reconciliation retired the managed `feature-notes` record, added
  `feature-specification`, updated the six-dependency implementation workflow
  and bounded routing, and then reported `current` with 18 artifacts.
- The correction passed `go test ./...`, affected-package race tests,
  `go vet ./...`, `golangci-lint run ./...`, `govulncheck ./...`, focused gosec with
  zero issues, gitleaks with no leaks, both binary builds, `goreleaser check`,
  the complete snapshot build matrix, schema parsing, the exact command-tree
  tests, and the full version-control-eligible 300-line source/test audit.
- The full race suite reproduced only the unchanged `internal/worktree` PTY
  cancellation timing failure. The normal full suite and all affected-package
  race tests pass, the six whole-tree gosec findings remain in unchanged
  registry/worktree code, and `git diff -- cmd/git-wt internal/worktree` is
  empty.
- Diff assertions confirm no historical specification other than living SPEC
  0058 changed, active routing does not reference `feature-notes` or
  `docs/notes`, and the reduced command tree still rejects `kit spec`.

## OUTCOME

- Kit now exposes a coding-agent-first contract architecture with a typed
  catalog, schema-v2 project provenance, stable resolved-contract v1 JSON,
  declarative workflows, bounded routing, and a reusable registry core.
- Fresh initialization installs all 14 downstream rulesets and four workflows
  plus the complete deterministic environment, GitHub, repository-memory, RLM,
  and reference bootstrap. Repeat initialization backfills missing starters,
  preserves customization, and delegates registry drift to reconcile.
- Reconciliation is preview-first, section-aware, transactional on failure,
  conflict-preserving, path-move-aware, and exact-accept only for overwrites.
- The broad Kit 1.x runtime, commands, prompt templates, evaluations,
  configuration, and command-specific automation were removed. Historical
  specifications remain intact.
- `MAJOR_VERSION` now selects v2.0.0 from the current v1 tag line and subsequent
  mainline releases increment within major version 2.
- README, overview, Constitution, workflows, commands, migration, release, and
  generated agent-routing documentation now describe the new product model.
- Asynchronous PR feedback now has one registry-backed repair workflow with
  bounded await and late collect intake, fail-closed terminal states,
  rate-conscious GitHub observation, active human feedback, deterministic
  fingerprints, bounded repair passes, and the former `kit pr fix` supervisor
  semantics. The restored adapter performs only explicit bounded GitHub intake,
  lane preparation, prompt delivery, and confirmed named-thread resolution;
  Kit still performs no agent supervision or repair execution.
- `repository-bootstrap` restores the former init prompt's semantic duty through
  a local resolved workflow: the coding agent progressively inspects evidence
  and populates only verified Constitution, progress, testing, tooling,
  integration, Makefile, and README content. Kit itself does not infer that
  truth or read secrets.
- Every non-trivial feature now resolves a mandatory living V3 specification
  contract. Missing or incomplete specs expose authoring permission but block
  source and delivery; complete specs unlock source, and only deliver/complete
  phases unlock delivery. Relevant historical specs remain progressively
  discoverable without restoring `kit spec` or the active `docs/notes` model.
- Implementation, local acceptance, ready pull-request delivery, exact-head
  verification, and hosted-check observation are complete on `GH-133` and PR
  #134. Human review and merge remain external repository actions.

## DELIVERY EVIDENCE

- GitHub issue #133, branch `GH-133`, and ready PR #134 are the single accepted
  delivery lane for the complete coding-agent-first major release and every
  accepted correction.
- Before this correction's delivery commit, local `HEAD`, `origin/GH-133`, the
  GitHub branch ref, and PR head all matched
  `87d9de4d01990d1ae278c0697b2c352f0d647f69`; the issue and PR were open and
  assigned to `jamesonstone`, and the PR remained mergeable.
- The correction's exact pushed head and hosted-check state are recorded after
  push. Human review and merge remain external actions.

## REPOSITORY MEMORY

This specification is the canonical rationale, implementation history,
validation record, and major-release migration decision. Durable contract,
registry, reconciliation, and product-boundary invariants were promoted to the
refactored `docs/CONSTITUTION.md`; reusable execution contracts live under
`docs/references/workflows/`, and operator migration details live under
`docs/migrations/` and `docs/releases/`.
