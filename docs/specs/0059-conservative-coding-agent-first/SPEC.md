---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0059"
  slug: "conservative-coding-agent-first"
  dir: "0059-conservative-coding-agent-first"
relationships:
  - type: builds_on
    target: 0042-native-plan-repository-memory
  - type: builds_on
    target: 0044-versioned-agent-instructions
  - type: builds_on
    target: 0047-kit-health-maintenance
  - type: builds_on
    target: 0058-cross-repository-program-coordination
references:
  - id: command-capabilities
    name: Command capabilities rule
    type: rule
    target: docs/references/rules/command-capabilities.md
    relation: guides
    read_policy: must
    used_for: exact command behavior and safety metadata
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: guides
    read_policy: must
    used_for: complete validation and literal evidence
    status: active
  - id: source-size-rule
    name: Source file size rule
    type: rule
    target: docs/references/rules/source-file-size.md
    relation: constrains
    read_policy: must
    used_for: implementation and test file organization
    status: active
  - id: delivery-rule
    name: GitHub PR delivery rule
    type: rule
    target: docs/references/rules/github-pr-delivery.md
    relation: guides
    read_policy: must
    used_for: issue, branch, commit, push, and pull request delivery
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: issue and pull request coordination
    required: true
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Ship Kit v2.0.0 as a conservative coding-agent-first release: retain the
small, proven human and automation interfaces that own repository setup,
memory, inspection, repair, and delivery while removing unused or duplicative
workflow surfaces. Add deterministic local context resolution so coding agents
can load the exact repository workflows, rules, specifications, strategies,
and implementation evidence relevant to their work. Add bounded local usage
telemetry so future command-removal decisions have evidence instead of guesses.

## CONTEXT

- Kit currently presents 71 command paths across repository memory, staged
  feature lifecycle, prompt libraries, loops, review, verification, and
  maintenance. Many are generated or compatibility surfaces rather than
  demonstrated external consumers.
- The protected repository and automation interfaces are `init`, `spec`,
  `status`, `registry status`, `health`, `capabilities`, `config check`,
  `aws verify`, `check`, `pr fix`, `improve run`, `rules add|list|view|link`,
  `reconcile`, `dispatch`, `instructions`, `upgrade`, `version`, and
  `completion`. The separate `git-wt` binary remains unchanged.
- The active weekly Kit health task spans 29 repositories and depends on the
  current `capabilities`, `registry status`, `status`, `reconcile`, `health`,
  and `check --project` behavior. That maintenance contract must remain intact.
- `kit aws verify` has executable downstream consumers and `kit improve run`
  is used by Kit's hosted workflows. Both remain public.
- Closed PR #134 attempted a much broader coding-agent contract rewrite. This
  implementation intentionally does not revive its wholesale registry,
  scaffold, or command reset. It preserves `internal/templates`, the current
  rules registry, repository-local Markdown authority, and current reconcile
  behavior.
- No command-usage history exists today. The accepted removal allowlist is the
  v2 compatibility boundary; telemetry begins with v2 and informs later
  releases rather than delaying this cleanup.
- GitHub issue #137 and branch `GH-137` own this implementation.

## REQUIREMENTS

- Publish the change as semantic-version major release v2.0.0.
- Preserve the exact approved command paths and remove every other current
  root or nested command, including active `notes` and `project refresh`
  surfaces. Keep `dispatch` behavior and safe flags intact.
- Keep `kit reconcile` behavior unchanged. Do not fold health, semantic
  project refresh, or usage maintenance into it.
- Keep the weekly health task's schedule, repository allowlist, safety,
  maintenance commands, and no-merge behavior unchanged. After v2 is installed,
  add one overall read-only usage analysis.
- Keep `kit spec` focused on V3 living repository memory. Remove the legacy
  supervisor and active notes scaffolding while continuing to read historical
  V1 and V2 specifications.
- Remove `docs/notes` as an active Kit concept without deleting downstream
  project-owned notes or mechanically rewriting historical specs.
- Add `kit context resolve` with versioned JSON and concise text output. It
  must be deterministic, repository-local, read-only, network-free, and free
  of Git mutation, model inference, or agent execution.
- Resolve ordered evidence for explicit workflows and optional feature/path
  hints, including required and optional local workflows, rules, specs,
  strategies, implementation patterns, reasons, and content digests.
- Add lightweight repository-bootstrap, implementation-delivery,
  repository-maintenance, PR-feedback-repair, and cross-repository program
  coordination workflows plus a mandatory coding-agent context-usage rule.
- Preserve the existing rules-registry architecture; do not introduce a typed
  generic artifact catalog or move network access into context resolution.
- Add local usage telemetry with global and project opt-out, machine-local
  salted project identities, no argument/path/secret capture, no network
  transmission, and best-effort failure isolation.
- Bound usage storage to 365 days, 16 MiB total, and 2 MiB shards. Add report,
  status, refresh, clear, enable, and disable commands; bare `usage` aliases
  `usage report`; usage commands do not record themselves.
- Continue parsing `loop`, `prompts`, `feature_state`, `removed_features`, and
  `project_refresh` configuration as inert compatibility data for one major
  release without acting on or deleting it.
- Update root help, README, command/migration/release documentation, generated
  agent routing, capabilities, release tagging, tests, and goldens.
- Keep every touched handwritten source and test file at or below 300 physical
  lines and preserve overall command performance.
- Observable acceptance: exact command-tree golden passes; removed commands
  are absent; usage privacy/bounds/failure tests pass; context resolution is
  deterministic and local-only; init backfills workflows without overwrite;
  protected command compatibility passes; next tag calculation selects
  v2.0.0; complete repository validation passes.
- Non-goals: launch or supervise coding agents, send telemetry, infer project
  truth, replace repository-local Markdown, rewrite the registry wholesale,
  modify `git-wt`, change reconcile semantics, delete downstream notes, or
  reactivate removed legacy commands behind hidden aliases.

## ACCEPTED PLAN

1. Characterize and freeze the exact v2 command tree and protected weekly,
   reconcile, health, dispatch, PR-repair, AWS, improvement, and `git-wt`
   contracts.
2. Add a small internal usage package and thin CLI adapters with bounded JSONL
   storage, privacy-safe configuration precedence, reporting, maintenance,
   opt-out, and failure-isolated root-command instrumentation.
3. Add a small internal context package and thin `kit context resolve` adapter
   that reads only declared repository-local workflow and reference evidence.
4. Add five lightweight managed workflow documents and the downstream
   coding-agent context-usage ruleset through existing template/init/registry
   mechanisms, preserving `internal/templates` and reconcile behavior.
5. Remove unprotected command registrations, implementations, tests,
   capabilities, templates, and active documentation. Retain internal helpers
   still required by protected commands such as `dispatch --loop` and
   `improve run`.
6. Remove active notes and project-refresh behavior, leave retired config
   fields parseable, preserve historical specs and downstream project content,
   and document canonical replacements.
7. Update release automation so the first post-merge tag is v2.0.0 and later
   tags increment within v2.
8. Run focused and complete validation, self-review the integrated diff,
   explicitly stage, commit, push, and open a ready pull request. After merge
   and installed-release verification, update the weekly task with the
   accepted one-time-per-run usage analysis.

## DECISIONS

- Accepted: use an explicit v2 allowlist and delete other commands now rather
  than retain them for an observational telemetry period.
- Accepted: telemetry event history lives outside repositories in bounded
  global JSONL shards. `.kit.yaml` stores only live configuration because
  tracked event logs would dirty repositories and make reads and cleanup
  progressively more expensive.
- Accepted: a global disable is absolute; a project disable suppresses only
  that project; project enable cannot override global disable.
- Accepted: telemetry recording is best-effort and cannot change another
  command's stdout, stderr, JSON, latency class, or exit status.
- Accepted: context resolution is a reproducible read plan over materialized
  local evidence, not a second source of repository truth and not a networked
  registry client.
- Accepted: replace semantic `project refresh` with the
  `repository-maintenance` workflow rather than changing `kit reconcile`.
- Accepted: preserve all current safe `dispatch` behavior, including prompt
  helpers that share review-loop language, while removing the executable loop
  command family.
- Accepted: keep inactive legacy config fields parseable for one major release
  but provide no hidden dual runtime.
- Rejected: store command events in project `.kit.yaml`; it creates contention,
  unbounded rewrites, tracked noise, and cleanup problems.
- Rejected: restore PR #134's typed registry and broad scaffold rewrite; it is
  disproportionate to this conservative pivot and conflicts with the accepted
  compatibility boundary.

## DISCOVERIES

- The current rules registry is implemented in `pkg/cli` against
  `docs/references/rules` on GitHub; there is no standalone
  `internal/registry` package on current `main`.
- `internal/templates` is active and contains V3 instruction, scaffold,
  ruleset, and workflow/spec templates. It should be evolved, not removed.
- `project refresh` only generates semantic Constitution-curation guidance and
  optionally records cadence state. The new repository-maintenance workflow
  can preserve that semantic duty without mutating reconcile.
- `map --context` is directly duplicative of the new deterministic context
  resolver and has no separate protected consumer.
- `dispatch --loop` is a prompt-routing flag, not the executable loop runtime;
  shared prompt helpers may remain internal after the loop command disappears.
- The release workflow currently increments the latest tag's patch version on
  every main push, so it would emit another v1 tag unless the v2 transition is
  made explicit.
- Legacy `feature_state` and `removed_features` were still influencing status
  and rollup output after their command owners were removed. Treating the
  fields as compatibility input required making those consumers inert too,
  while continuing to parse and preserve explicitly configured values.
- Workflow documents must use the same path-confinement boundary as their
  referenced evidence. Reading workflow symlinks directly would have allowed a
  blocked contract to inspect and digest a file outside the project root.
- Fresh-init resolution cannot be validated against the pre-merge `main`
  registry because the new context-usage ruleset does not exist there yet.
  The init acceptance test therefore injects the exact registry documents and
  proves the post-merge repository-bootstrap contract end to end.
- The protected `kit improve run` adapter was still backed by default and
  `prompt-system` suites that asserted retired commands and incidental legacy
  capability wording. The hosted suite name remains stable, but its task set
  now measures the v2 agent contract, local context resolution, usage and
  protected command capabilities, the living-spec boundary, PR repair, and
  dispatch prompt determinism.
- Go 1.25.5 exposed reachable standard-library vulnerabilities during the
  release scan. Raising the declared patch toolchain to Go 1.25.12 removes the
  reachable findings without changing the language or module compatibility
  boundary.
- Race instrumentation can delay the PTY helper startup beyond the old 200 ms
  test sentinel. Delaying the sentinel until after the helper's cancellation
  timer makes the test assert terminal restoration instead of racing queued
  input; the `git-wt` implementation itself remains unchanged.
- A pre-merge registry-backed reconcile preview necessarily sees `main` rather
  than this branch, so it proposes the soon-to-be-retired feature-notes rule
  and other base-branch drift. The preview remained read-only and is not an
  implementation defect; post-merge registry behavior is covered by local
  tests using the branch's exact artifacts.

## VALIDATION

- PASS: focused config, feature, rollup, template, usage, context, and CLI
  tests, including exact command-tree, removed-command, workflow
  materialization, dry-run preservation, fresh-bootstrap resolution, privacy,
  storage bounds, path confinement, deterministic JSON, documentation, and
  release-transition coverage.
- PASS: `go test ./... -count=1` plus an explicit uncached
  `go test ./pkg/cli -count=1`.
- PASS: `go test -race ./... -count=1` with Go 1.25.12, including three
  consecutive focused PTY cancellation runs.
- PASS: `go vet ./...` and `golangci-lint run ./...` (`0 issues`).
- PASS: `govulncheck ./...` found zero reachable vulnerabilities; one required
  module contains an uncalled vulnerable symbol.
- PASS: current-platform and Windows amd64 builds for both `kit` and `git-wt`,
  plus a complete GoReleaser snapshot for Darwin, Linux, and Windows on amd64
  and arm64.
- PASS: `gitleaks detect --source . --no-banner --redact` scanned 304 commits
  and about 10.94 MB with no leaks.
- PASS: whole-project source-size audit checked 361 eligible handwritten
  source/test files and found none above 300 physical lines; the largest is
  exactly 300 lines.
- PASS: self-hosted capability, context, usage, and removed-command checks with
  isolated local configuration; `kit check --project` and all 56 historical
  and current features passed.
- PASS: the updated `default` improve suite passed 8/8 task runs and the
  compatibility-named `prompt-system` agent-contract suite passed 24/24 task
  runs with deterministic repeated output.
- PASS: `kit reconcile --all --output-only` remained prompt-producing and a
  `--include-files --dry-run --diff` registry preview performed no writes. Its
  expected pre-merge base-registry drift is recorded in DISCOVERIES.
- PENDING: final formatting/diff audit after this living-spec update, explicit
  staging, exact-head push, pull request creation, and hosted checks.
- NOT APPLICABLE: browser, live-integration, deployment, and production tests;
  Kit is a local Go CLI and this feature adds no deployed service.

## OUTCOME

- Implemented the conservative v2 command allowlist and removed unprotected
  CLI runtimes, dead packages, tests, prompt snapshots, and active docs while
  leaving historical specs intact.
- Added local `kit.context/v1` resolution, five repository-local workflows,
  context-usage rules, init materialization, bounded `kit.usage/v1` storage,
  usage controls/reporting, exact command and capability enforcement, v2
  migration/release docs, and the major tag transition.
- Preserved `internal/templates`, registry-backed rules, `kit reconcile`,
  dispatch, PR repair, health, the weekly automation boundary, and `git-wt`.
- Local implementation and release validation are complete; delivery remains
  in progress on `GH-137`.

## REPOSITORY MEMORY

- Created this V3 specification because the public compatibility boundary,
  telemetry privacy model, deterministic context contract, rejected registry
  rewrite, notes retirement, release transition, and weekly activation order
  contain consequential rationale that code and tests alone cannot preserve.
- Historical specifications remain preserved as evidence of the product's
  evolution, including decisions now superseded by the conservative v2 pivot.
