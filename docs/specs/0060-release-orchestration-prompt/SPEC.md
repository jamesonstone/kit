---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0060"
  slug: "release-orchestration-prompt"
  dir: "0060-release-orchestration-prompt"
relationships:
  - type: builds_on
    target: 0059-conservative-coding-agent-first
  - type: related_to
    target: 0058-cross-repository-program-coordination
  - type: related_to
    target: 0061-authorized-coding-agent-merge-autonomy
references:
  - id: context-usage
    name: Coding agent context usage
    type: rule
    target: docs/references/rules/coding-agent-context-usage.md
    relation: guides
    read_policy: must
    used_for: downstream context resolution contract
    status: active
  - id: testing-rule
    name: Testing and environment validation
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: guides
    read_policy: must
    used_for: deterministic and integration validation
    status: active
  - id: source-size-rule
    name: Source file size
    type: rule
    target: docs/references/rules/source-file-size.md
    relation: constrains
    read_policy: must
    used_for: implementation and test organization
    status: active
  - id: infrastructure-rule
    name: Infrastructure change approval
    type: rule
    target: docs/references/rules/infrastructure-change-approval.md
    relation: guides
    read_policy: must
    used_for: generated release mutation boundary
    status: active
  - id: team-rule
    name: Agent team orchestration
    type: rule
    target: docs/references/rules/agent-team-orchestration.md
    relation: guides
    read_policy: conditional
    used_for: safe release preparation topology
    status: active
  - id: delivery-rule
    name: GitHub PR delivery
    type: rule
    target: docs/references/rules/github-pr-delivery.md
    relation: guides
    read_policy: must
    used_for: issue branch commit push and pull request delivery
    status: active
  - id: merge-rule
    name: GitHub PR merge
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: guides
    read_policy: must
    used_for: exact merge authority readiness identity and wave evidence
    status: active
  - id: in-place-remediation-issue
    name: Prefer in-place PR remediation in merge policy
    type: external
    target: https://github.com/jamesonstone/kit/issues/172
    relation: supports
    read_policy: must
    used_for: release-graph remediation behavior and delivery traceability
    status: active
skills: []
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Add `kit pr orchestrate`, a first-class prompt-producing command that converts
minimal user input, bounded repository discovery, and conservative defaults
into a deterministic coding-agent prompt for dependency-aware multi-repository
release delivery. The command prepares the release control contract; it never
executes the release itself.

## CONTEXT

- Kit v2 admits public commands through an exact protected-path allowlist and
  requires matching capability and bounded usage-telemetry registration.
- Generic prompt-library commands were intentionally removed in feature 0059.
  This feature is a bounded adapter under the retained `pr` command, not a
  restoration of the retired prompt-library architecture.
- Command handlers are expected to stay thin. Reusable policy, discovery,
  normalization, validation, provenance, and rendering belong in an internal
  package.
- Repository-local workflows are deterministic evidence contracts materialized
  by `kit init`. A release-specific workflow must be checked in and embedded,
  while downstream repositories that have not refreshed need an explicit safe
  fallback.
- Usage telemetry records only normalized command paths and bounded execution
  metadata. It never records arguments, repository paths, or prompt contents.
- Issue #141 and branch `GH-141` own this implementation. The accepted product
  behavior was clarified before implementation with 99 percent confidence and
  no unresolved decisions.
- A read-only reconcile preview found pre-existing managed-file drift in
  `.kit.yaml` and the README maintainer markers. That drift is not adopted as
  part of this feature; overlapping README edits will preserve current owned
  content.
- Downstream use exposed a replacement-first failure loop: the renderer turned
  every source correction into a new PR and graph node even when the existing
  PR could safely carry a routine, scope-preserving repair.

## REQUIREMENTS

- Register `kit pr orchestrate` in the v2 protected command tree, local usage
  telemetry, capabilities catalog, help, documentation, and exact-tree tests.
- Support repeatable `--repos` and mutually exclusive `--root`; bounded root
  discovery considers only the root and immediate child Git repositories.
- Support `--project`, `--organization`, `--context`, `--scope`, `--infra`,
  `--infra-provider`, `--infra-cli`, one `--environment`, `--verify`,
  `--integration-suite`, `--dry-run`, `--output-only`, and `--copy`.
- Ask questions only when stdin and stderr are terminals. Interactive runs
  without scope ask for it; noninteractive runs require explicit scope.
- Resolve explicit values before discovered values and discovered values before
  defaults, recording provenance for dry-run inspection.
- Accept tagged verification values `auto`, `command:`, `script:`, `endpoint:`,
  and `instruction:`. Integration additionally accepts `none`.
- Prefer local Git metadata. Permit at most one cached targeted `gh repo view`
  fallback per repository when local identity or default-branch evidence is
  insufficient. Do not enumerate PRs, poll, run cloud CLIs, inspect script
  bodies, or read secret payloads.
- Keep destructive infrastructure changes behind explicit manual approval and
  validate infrastructure mode/provider/CLI combinations conservatively.
- Merge the supplied objective and LabCore template semantics into one complete
  prompt covering graph construction, readiness, compatibility, review,
  conflict remediation, merge/deployment waves, infrastructure, verification,
  in-place PR remediation, exceptional replacement PRs, and literal final
  reporting.
- Make the prompt consume the active merge-authority contract: build and
  authorize the bounded graph first, resolve `pull-request-merge`, merge only
  the reconciled ready frontier, reauthorize only material scope expansion,
  and report partial waves literally.
- Keep routine, scope-preserving source remediation on the existing PR head
  under bounded repair authority. Invalidate old-head readiness and merge
  authority, rerun current-head checks and review, and require fresh exact-head
  authorization. Create a replacement graph node only for material scope or
  architecture change, an unsafe original head, or explicit policy/user
  direction.
- Render no unresolved template placeholders or empty optional sections.
- Add a `release-orchestration` context workflow based on
  `implementation-delivery`; require infrastructure and topology evidence only
  at their applicability boundaries, and apply cross-repository program
  coordination conditionally under its existing trigger.
- Default TTY output copies the prompt and acknowledges it. Non-TTY and
  `--output-only` emit raw prompt output. `--copy` explicitly adds clipboard
  output. `--dry-run` always emits a deterministic Markdown bundle containing
  resolved YAML/provenance and the prompt without clipboard access.
- Keep every affected handwritten source and test file at or below 300 lines.
- Observable acceptance: command/capability/telemetry catalogs agree; config,
  discovery, rendering, terminal, dry-run, workflow materialization, and
  fallback tests pass; complete repository validation passes.
- Non-goals: construct the actual release graph, enumerate or merge PRs, modify
  repositories, launch coding agents, deploy software, mutate infrastructure,
  persist release configuration in `.kit.yaml`, support multiple promotion
  environments in v1, or restore generic prompt-library behavior.

## ACCEPTED PLAN

1. Add the exact v2 command admission, telemetry, capability, parent-help, and
   public command-tree changes for `pr orchestrate`.
2. Create `internal/releaseprompt` with typed input/config/provenance models,
   bounded Git/GitHub discovery, deterministic precedence and validation, and
   responsibility-split prompt rendering.
3. Add a thin Cobra adapter for flags, terminal-only conditional questions,
   exact stdout/stderr routing, clipboard behavior, and dry-run output.
4. Add and embed the `release-orchestration` downstream workflow, including
   initialization refresh coverage and explicit fallback instructions for
   repositories that have not materialized it.
5. Consume the authorized merge-autonomy contract in the rendered prompt and
   keep capability metadata explicit that Kit generates instructions while the
   coding agent performs any separately authorized GitHub mutation.
6. Update the Constitution, README, command/overview documentation, feature
   memory, and prompt golden as one integration set.
7. Run focused tests, complete Go tests, vet, lint, builds, project checks,
   manual CLI flows, source-size and diff audits, then explicitly stage,
   commit, push, and open a ready pull request.
8. On `GH-172`, replace the corrective-PR-first failure loop in the renderer,
   workflow, golden, and tests with in-place remediation by default and
   exceptional replacement nodes, while preserving exact-head authorization.

## DECISIONS

- Accepted: keep run configuration invocation-local rather than extending
  `.kit.yaml`; release intent is transient and may contain repository-specific
  operational detail.
- Accepted: use one canonical merged prompt rather than expose multiple
  templates or modes.
- Accepted: place domain behavior in `internal/releaseprompt`; `pkg/cli` owns
  only Cobra, terminal interaction, and presentation.
- Accepted: use a dedicated exact renderer without ambient profile, skill, or
  subagent suffixes because the release prompt is itself a complete execution
  contract.
- Accepted: one environment target in v1. Ordered multi-environment promotion
  can be added later without weakening current safety gates.
- Accepted: ordinary v2 usage telemetry may record the command path during
  dry-run because it is machine-local, best-effort, argument-free, and
  independent of release-side mutations.
- Rejected: make every multi-repository release create a program ledger.
  Existing cross-repository coordination applicability remains authoritative.
- Rejected: invoke broad GitHub discovery from the generator. The downstream
  agent owns release-set and graph construction.
- Accepted: preserve the existing PR node for routine scoped correction and
  represent its new head as an `UNKNOWN` revision until fresh evidence and
  exact-head authorization exist.
- Rejected: turn every source correction into a new corrective PR. It creates
  recursive graph growth and discards useful issue, discussion, review, and
  scope continuity without improving exact-head safety.

## DISCOVERIES

- `pkg/cli/root.go` prunes every path not admitted by
  `internal/commandset`, so command registration without allowlist admission
  would silently hide the feature.
- Capabilities and public-tree tests derive their expected surfaces from the
  same protected catalog; telemetry paths intentionally cover protected
  user-facing commands except usage-maintenance commands.
- Existing shared prompt output helpers append ambient orchestration content
  and write through package-global output. This feature needs an exact,
  command-writer-based path for deterministic piping and tests.
- Context workflow artifacts have both checked-in and embedded copies, and
  `kit init --refresh` preserves project customization. Both copies and their
  agreement tests must change together.
- The expanded generated instruction gate brought the V2 table-of-contents
  template close to the historical 90-line audit threshold. A small
  project-owned section then produced a false monolithic-manual finding, so
  the narrow baseline was raised to 100 lines while duplicate-manual snippet
  detection remains authoritative.
- The baseline `go test ./...` passes on commit
  `0eb617e80839d3bc9ae326dbe3c63ddc5d0b0591`.
- The release workflow, hard-rule list, graph vocabulary, failure loop,
  completion gate, reporting language, expected lifecycle, and golden all
  repeated corrective-PR-first semantics; partial wording changes would leave
  a contradictory generated prompt.

## VALIDATION

- Focused `internal/releaseprompt`, instruction-template, context-workflow,
  policy-consistency, and `pkg/cli` tests pass.
- Complete Go tests pass, including the exact `pkg/cli` suite; complete race
  tests pass, including the exact `pkg/cli` race suite.
- `go vet ./...`, `golangci-lint run ./...`, and `go build ./...` pass with no
  issues.
- `kit context resolve` succeeds unblocked for `release-orchestration` and
  `pull-request-merge` with the current feature and source hints.
- `go run ./cmd/kit check --project` reports a coherent contract.
- `go run ./cmd/kit reconcile --all --output-only` reports no reconciliation
  needed and audits 382 eligible handwritten source/test files with none over
  300 physical lines.
- Manual capability, raw noninteractive prompt, dry-run provenance, bounded
  root discovery, workflow-presence, and authority-frontier checks pass.
- `git diff --check` passes. Hosted pull-request checks remain a delivery
  observation, not local validation evidence.
- `GH-172` focused renderer and golden tests pass together with merge-rule,
  active-policy, generated-gate, workflow-materialization, and release-workflow
  assertions. The golden contains in-place same-branch remediation and no
  replacement-first source-correction instruction.
- Complete uncached Go and race tests, formatting, vet, lint, both builds,
  project validation, all 65 feature checks, and diff hygiene pass.
- The built binary resolves `release-orchestration` and `pull-request-merge`
  unblocked. A strict `pr orchestrate --dry-run` observes every new remediation
  boundary and rejects the retired corrective-PR-first phrases.
- Whole-project reconcile reports no drift and audits 712 candidates and 369
  eligible handwritten source/test files with zero above 300 lines. Gitleaks
  scans 370 commits and 13.15 MB with no leaks.
- Browser, deployment, infrastructure, live-integration, and production
  validation are `NOT_APPLICABLE` because this change renders instructions but
  performs no release, deployment, or provider mutation.

## OUTCOME

- `kit pr orchestrate` is a protected, telemetered, capability-described v2
  command with explicit repository scope, deterministic resolution provenance,
  bounded local/GitHub metadata discovery, tagged validation, exact output and
  clipboard behavior, and a deterministic dry-run bundle.
- The complete renderer constructs an authority-aware Global Release Graph,
  consumes `pull-request-merge`, restricts merges to the current authorized
  `MERGE_READY` frontier, separates merge/deployment/production evidence, and
  preserves partial waves literally.
- The renderer now keeps routine scoped remediation on the existing PR and
  creates replacement graph nodes only for material or unsafe changes. Every
  changed head loses old-head readiness and merge authority before it can
  return to the merge frontier.
- The embedded and checked-in `release-orchestration` workflow extends
  implementation delivery while keeping infrastructure, agent-team, and
  cross-repository program evidence conditional at their applicability
  boundaries.
- Kit remains a prompt generator: it does not enumerate PRs, execute merges,
  deploy software, mutate infrastructure, or launch an agent.

## REPOSITORY MEMORY

Created this V3 specification because the public v2 command boundary,
invocation-local configuration contract, discovery limits, rendering boundary,
downstream context fallback, infrastructure approval semantics, and rejected
alternatives contain material rationale that code and tests alone cannot
preserve.

Updated it for issue #172 because defaulting to in-place PR repair while
preserving exact-head authorization is a durable release-graph decision that
the renderer and golden alone cannot explain.
