---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0033
  slug: kit-capabilities
  dir: 0033-kit-capabilities
relationships:
  - type: builds_on
    target: 0019-command-surface-simplification
  - type: builds_on
    target: 0030-reference-graph-routing
  - type: related_to
    target: 0016-document-map-relationships
  - type: related_to
    target: 0020-versioned-instruction-model
  - type: related_to
    target: 0021-project-validation-and-instruction-registry
references:
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: artifact pipeline order, read-only command expectations, populated-section rules, and command/package constraints
    status: active
  - id: agent-routing
    name: Agent routing docs
    type: doc
    target: docs/agents/README.md
    selector_type: artifact
    selector: README.md
    relation: guides
    read_policy: must
    used_for: repo-local entrypoint and progressive disclosure routing for this research phase
    status: active
  - id: workflow-rules
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: brainstorm-only phase classification, source-of-truth order, and clarification expectations
    status: active
  - id: rlm-rules
    name: RLM rules
    type: doc
    target: docs/agents/RLM.md
    selector_type: heading
    selector: Rules
    relation: constrains
    read_policy: must
    used_for: just-in-time prior-work and codebase inspection bounds
    status: active
  - id: guardrails
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    selector_type: heading
    selector: Completion Bar
    relation: constrains
    read_policy: must
    used_for: documentation completion bar, placeholder removal, and no-git/no-production-edit constraints
    status: active
  - id: tooling-doc
    name: Tooling docs
    type: doc
    target: docs/agents/TOOLING.md
    selector_type: heading
    selector: Skills
    relation: guides
    read_policy: conditional
    used_for: no feature-local skills were declared and repo-local docs remain primary
    status: active
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0033-kit-capabilities
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input; only .gitkeep exists in this pass
    status: optional
  - id: feature-map-command
    name: Feature map command
    type: command
    target: "kit map 0033-kit-capabilities"
    selector_type: command
    selector: "kit map 0033-kit-capabilities"
    relation: informs
    read_policy: evidence
    used_for: confirmed current brainstorm phase, missing SPEC/PLAN/TASKS, outgoing feature relationships, reference resolution, and feature-notes reference
    status: active
  - id: progress-summary
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    selector_type: heading
    selector: FEATURE PROGRESS TABLE
    relation: informs
    read_policy: evidence
    used_for: shortlisted prior features and confirmed 0033 is brainstorm-phase with no description yet
    status: active
  - id: wiring-search
    name: CLI wiring search
    type: command
    target: "rg -n \"rootCmd|AddCommand|configureRootHelp|commandOrder|rootCommandSections\" cmd pkg internal docs"
    selector_type: command
    selector: "rg -n \"rootCmd|AddCommand|configureRootHelp|commandOrder|rootCommandSections\" cmd pkg internal docs"
    relation: verifies
    read_policy: evidence
    used_for: concrete root command registration, help ordering, section rendering, and existing tests
    status: active
  - id: json-search
    name: JSON output search
    type: command
    target: "rg -n -- \"--json|json\" pkg/cli internal cmd docs/specs"
    selector_type: command
    selector: "rg -n -- \"--json|json\" pkg/cli internal cmd docs/specs"
    relation: verifies
    read_policy: evidence
    used_for: existing JSON flag and stable schema patterns across map, status, verify, trace, replay, state, eval, loop, and local ci worktree files
    status: active
  - id: git-status-evidence
    name: Worktree status evidence
    type: command
    target: "git status --short"
    selector_type: command
    selector: "git status --short"
    relation: informs
    read_policy: evidence
    used_for: identified pre-existing dirty worktree state, including local ci command files and untracked 0033 brainstorm artifacts
    status: active
  - id: command-surface-simplification
    name: Command surface simplification
    type: feature
    target: docs/specs/0019-command-surface-simplification
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: grouped root help contract, visible canonical commands, hidden deprecated command handling, and README command table expectations
    status: active
  - id: reference-graph-routing
    name: Reference graph routing
    type: feature
    target: docs/specs/0030-reference-graph-routing
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: canonical front matter references schema and distinction between document context plans and command capability discovery
    status: active
  - id: document-map-relationships
    name: Document map relationships
    type: feature
    target: docs/specs/0016-document-map-relationships
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: read-only map precedent, relationship semantics, and low-token orientation surface pattern
    status: active
  - id: instruction-model
    name: Versioned instruction model
    type: feature
    target: docs/specs/0020-versioned-instruction-model
    selector_type: artifact
    selector: SPEC.md
    relation: guides
    read_policy: conditional
    used_for: thin ToC/RLM instruction model and guidance to keep always-loaded instructions small
    status: active
  - id: instruction-registry
    name: Project validation and instruction registry
    type: feature
    target: docs/specs/0021-project-validation-and-instruction-registry
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: shared registry precedent and separation between RLM discovery and dispatch execution planning
    status: active
  - id: root-command
    name: Root command
    type: code
    target: pkg/cli/root.go
    selector_type: symbol
    selector: rootCmd
    relation: implements
    read_policy: must
    used_for: root Cobra command, version template, long description, and Execute error handling
    status: active
  - id: root-help
    name: Root help
    type: code
    target: pkg/cli/root_help.go
    selector_type: symbol
    selector: rootCommandSections
    relation: constrains
    read_policy: must
    used_for: command ordering, visible command sections, root-only grouped help, and help discoverability tests
    status: active
  - id: root-help-tests
    name: Root help tests
    type: code
    target: pkg/cli/root_help_test.go
    selector_type: symbol
    selector: TestRootHelpGroupsCanonicalCommands
    relation: verifies
    read_policy: evidence
    used_for: expected help sections, hidden deprecated command behavior, and command registration assertions
    status: active
  - id: map-command
    name: Map command
    type: code
    target: pkg/cli/map.go
    selector_type: symbol
    selector: mapCmd
    relation: implements
    read_policy: conditional
    used_for: read-only command style, --json output helper, --context distinction, and selected-feature behavior
    status: active
  - id: project-map-model
    name: Project map model
    type: code
    target: internal/feature/map.go
    selector_type: symbol
    selector: BuildProjectMap
    relation: informs
    read_policy: conditional
    used_for: existing document/reference map data model and instruction-doc registry consumption
    status: active
  - id: metadata-schema
    name: Metadata schema
    type: code
    target: internal/document/metadata.go
    selector_type: symbol
    selector: MetadataReference
    relation: constrains
    read_policy: must
    used_for: valid reference fields, relationship types, selector_type enum, read_policy enum, and status enum
    status: active
  - id: instruction-registry-code
    name: Instruction registry
    type: code
    target: internal/instructions/registry.go
    selector_type: symbol
    selector: SupportDocs
    relation: informs
    read_policy: conditional
    used_for: existing registry shape for docs metadata and possible pattern for command metadata ownership
    status: active
  - id: status-command
    name: Status command
    type: code
    target: pkg/cli/status.go
    selector_type: symbol
    selector: statusCmd
    relation: informs
    read_policy: conditional
    used_for: command-local --json pattern and backward-compatible payload caution
    status: active
  - id: verify-command
    name: Verify command
    type: code
    target: pkg/cli/verify.go
    selector_type: symbol
    selector: verifyCmd
    relation: informs
    read_policy: conditional
    used_for: command capability metadata examples for command execution, .kit/runs writes, and dry-run/no-write flags
    status: active
  - id: dispatch-command
    name: Dispatch command
    type: code
    target: pkg/cli/dispatch.go
    selector_type: symbol
    selector: dispatchCmd
    relation: informs
    read_policy: conditional
    used_for: command capability metadata examples for prompt-only orchestration, PR/network input, and subagent planning
    status: active
  - id: ci-command
    name: CI command
    type: code
    target: pkg/cli/ci.go
    selector_type: symbol
    selector: ciCmd
    relation: informs
    read_policy: conditional
    used_for: local worktree command registration, JSON flag, GitHub Actions diagnostic behavior, dispatch flag, and capability metadata for network/subprocess/file-write nuance
    status: active
  - id: ci-github-support
    name: CI GitHub support
    type: code
    target: pkg/cli/ci_github.go
    selector_type: symbol
    selector: cacheCIDefaultBranch
    relation: informs
    read_policy: conditional
    used_for: local ci command subprocess calls to git/gh, GitHub API reads, gh authentication requirement, and .kit.yaml default-branch cache write
    status: active
  - id: ci-tests
    name: CI command tests
    type: code
    target: pkg/cli/ci_test.go
    selector_type: symbol
    selector: TestRunCIDiagnosesDefaultBranchFailureAndCachesGitHubConfig
    relation: verifies
    read_policy: evidence
    used_for: verified local ci tests expect default-branch cache writes and JSON output containing agent prompt data
    status: active
  - id: readme-commands
    name: README command table
    type: doc
    target: README.md
    selector_type: heading
    selector: 🧰 Commands
    relation: guides
    read_policy: conditional
    used_for: user-facing command grouping and documentation update target
    status: active
  - id: init-project-command-contract
    name: Init project command contract
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    selector_type: heading
    selector: "8. CLI Commands"
    relation: guides
    read_policy: conditional
    used_for: canonical user-facing command contract and likely durable documentation target
    status: active
---
# BRAINSTORM

## SUMMARY

Add a read-only `kit capabilities` command that exposes Kit's command surface as compact, versioned, agent-readable JSON so agents can choose the narrowest correct command without loading persistent full-context instructions. The implementation should add a small explicit command-capability catalog in the CLI layer, wire it into root help as an inspection surface, update agent-facing docs to discourage repeated full reads, and test drift against registered Cobra commands.

## USER THESIS

Coding agents need a compact, trustworthy way to discover Kit's command surface at runtime so they choose existing commands deliberately instead of guessing workflows from stale instructions or broad README context. The feature should expose command-selection metadata through a read-only `kit capabilities` surface while keeping detailed command contracts targeted and optional.

## Context Synthesis

Add a read-only Kit discovery surface so coding agents can inspect Kit commands, choose the narrowest command, and avoid hallucinated workflows while keeping always-loaded instructions small [S1][S2]. The affected users are Kit users delegating work to coding agents and agents deciding between `kit map`, `kit dispatch`, `kit verify`, the local worktree's `kit ci`, and related commands [S1]. The selected direction is `kit capabilities` with compact JSON by default, targeted command detail, optional full detail, and repo instruction updates that tell agents to treat the output as a discovery index, not persistent context [S2][S3]. Done means the CLI exposes versioned JSON, docs explain when and how to call it, repeated full reads are discouraged, and tests prove compact, targeted, full, and search output behavior [S2][S3].

## Source Map

- [S1] discussion: User requested functionality that makes Kit's command surface visible to coding agents through a detailed machine-readable command. Document identifier: side conversation message "should we include some new functionality in `kit`...".
- [S2] discussion: Chosen design is a read-only `kit capabilities` command with compact JSON, targeted command lookup, full mode, command metadata, examples, mutation level, and related commands. Document identifier: assistant response recommending `kit capabilities`.
- [S3] discussion: User asked whether instructions are needed to prevent repeated full `kit capabilities` reads; selected answer is to add explicit usage guidance and keep default output compact. Document identifier: side conversation message "do i need to include any additional instructions...".

## Coding Agent Instructions

Implement a read-only `kit capabilities` feature that exposes Kit's command surface through compact, versioned JSON and targeted detail while updating agent-facing docs to prevent repeated full-context loading [S1][S2][S3]. The key tradeoff is between discoverability and context cost: compact default output supports command selection, while targeted and full modes expose detailed contracts only after the agent narrows its decision [S2][S3].

1. Inspect the repository and identify existing CLI wiring by exact file path and symbol; run `rg -n "rootCmd|AddCommand|configureRootHelp|commandOrder|rootCommandSections" cmd pkg internal docs` and record the concrete files and symbols before editing [S1].
2. Reconcile brainstorm decisions with current code behavior: verify how commands are registered, how help text is rendered, and whether an existing command already exposes command metadata; mark missing facts as `UNKNOWN` with the exact inspection command used [S1][S2].
3. Produce a complete implementation strategy grounded in the current codebase: command name `capabilities`, default compact JSON, targeted `kit capabilities <command> --json`, optional `--full`, and optional `--search <term>` [S2][S3].
4. Enumerate file edits before coding: CLI command file, root command registration, help ordering, tests, and docs under agent tooling instructions; include interfaces, structs, JSON schema version, dependency changes, config changes, migrations, and validation commands [S2][S3].
5. Model capability records with fields for command name, category, summary, when-to-use, when-not-to-use, mutation level, network use, file writes, git mutation, important flags, examples, and related commands [S2].
6. Keep the command read-only: it must not write files, mutate git, call network services, or modify `.kit.yaml`; encode this in command metadata and tests [S2].
7. Update agent instructions to say `kit capabilities --json` is a discovery index, should be run only when command choice is uncertain, and should be followed by targeted lookups instead of repeated full output [S3].
8. Add tests for compact JSON, targeted command JSON, full JSON, search filtering, unknown command errors, schema version stability, and help listing [S2][S3].
9. Define acceptance checks with expected outputs: `go test ./...` exits `0`; `go run ./cmd/kit capabilities --json` returns valid compact JSON with `schema_version`; `go run ./cmd/kit capabilities dispatch --json` returns one command record; `go run ./cmd/kit capabilities --full --json` includes detailed arrays [S2][S3].
10. State risks, open questions, assumptions, mitigation, and owner: risk is duplicated command metadata drifting from implementation; mitigation is tests that compare registered command names to capability records; owner is the implementing coding agent [S2].

## Resource Links

- NONE

## RELATIONSHIPS

Relationships are tracked in front matter. This brainstorm builds on `0019-command-surface-simplification` and `0030-reference-graph-routing`, and is related to `0016-document-map-relationships`, `0020-versioned-instruction-model`, and `0021-project-validation-and-instruction-registry`.

## CODEBASE FINDINGS

1. `docs/CONSTITUTION.md` classifies this as spec-driven work because it is a new CLI capability and a public command-surface change. The active phase is brainstorm only: research and documentation can change, but production code, tests, runtime config, generated artifacts, git state, and implementation files must not be modified in this phase.
2. `docs/specs/0033-kit-capabilities/BRAINSTORM.md` is the only feature artifact currently present for `0033-kit-capabilities`. `kit map 0033-kit-capabilities` reported `phase: brainstorm`, `docs: B present / S missing / P missing / T missing / A missing`, no incoming relationships, outgoing relationships to `0019-command-surface-simplification`, `0030-reference-graph-routing`, `0016-document-map-relationships`, `0020-versioned-instruction-model`, and `0021-project-validation-and-instruction-registry`, and one optional feature-notes reference.
3. `docs/notes/0033-kit-capabilities` contains only `.gitkeep`; there are no usable pre-brainstorm note files to copy into this artifact. The feature-notes front matter reference remains `status: optional`.
4. The required CLI wiring inspection command was run exactly as requested: `rg -n "rootCmd|AddCommand|configureRootHelp|commandOrder|rootCommandSections" cmd pkg internal docs`. It found `pkg/cli/root.go::rootCmd`, `pkg/cli/root.go::init`, `pkg/cli/root_help.go::commandOrder`, `pkg/cli/root_help.go::rootCommandSections`, `pkg/cli/root_help.go::configureRootHelp`, many `rootCmd.AddCommand(...)` registrations across `pkg/cli/*.go`, local worktree registration for `pkg/cli/ci.go::ciCmd`, and root-help tests in `pkg/cli/root_help_test.go`.
5. `pkg/cli/root.go` owns the global Cobra root command (`rootCmd`), root long description, version template, and top-level `Execute()` error handling. A new `capabilities` command should be registered from a new CLI file through `rootCmd.AddCommand(capabilitiesCmd)`, following the command-file pattern used by `pkg/cli/map.go`, `pkg/cli/status.go`, `pkg/cli/verify.go`, and `pkg/cli/dispatch.go`.
6. `pkg/cli/root_help.go` owns visible command ordering and grouped root help. In this dirty worktree, `commandOrder` currently places `status` at 20, `map` at 21, `rm`/`remove` at 22, `check` at 23, `ci` at 24, `verify` at 25, `trace` at 26, `replay` at 27, `state` at 28, `eval` at 29, and `rules`/`rollup` at 30. `rootCommandSections` currently groups visible commands into Setup, Workflow, Inspect & Repair, Prompt Utilities, and Utilities, with `ci` already listed under Inspect & Repair. `capabilities` should likely be added to Inspect & Repair near `map`, because it is read-only discovery/inspection rather than a workflow mutation or prompt utility.
7. `pkg/cli/root_help.go::renderRootHelp` renders grouped root help from `rootCommandSections`, while `findVisibleSubcommand` omits hidden/deprecated commands unless the command is `help`. Any implementation that expects root help discoverability must update both `commandOrder` and `rootCommandSections`.
8. `pkg/cli/root_help_test.go::TestRootHelpGroupsCanonicalCommands` asserts visible sections and selected commands, including the local worktree's `ci` command, while `TestDeprecatedCommandsRemainRegisteredAndHidden` confirms `update`, `skills`, `catchup`, and `rollup` remain registered but hidden/deprecated. A future capabilities test should verify `capabilities` appears in root help, and a separate drift test should decide whether hidden compatibility commands are included in capability records with `hidden: true` / `deprecated` metadata or deliberately excluded.
9. No existing machine-readable command metadata registry was found. Evidence: `rg -n "capabilities|capability|metadata registry|command metadata|mutation level|command surface" .` found only broad prose references, prior planning notes, and the current brainstorm, not an existing data model that can be reused for command capabilities.
10. Existing JSON output is command-local. `pkg/cli/map.go::outputMapJSON`, `pkg/cli/verify.go::outputJSON`, and the local worktree's `pkg/cli/ci_render.go::renderCIDiagnosisJSON` use `json.NewEncoder(...).SetIndent("", "  ")`; `pkg/cli/status_render.go` uses `json.MarshalIndent`. A new capabilities command can use the existing `outputJSON` helper from `pkg/cli/verify.go` if kept in package `cli`, or introduce a small command-local helper only if naming/coupling is clearer.
11. `pkg/cli/map.go` is the closest read-only command precedent. It has `--json`, `--all`, and `--context`, builds data through `internal/feature.BuildProjectMap`, and writes no repository files. It is document-context discovery, not command-surface discovery; `kit capabilities` should not overload `kit map --context`, because map references answer "what docs should I read?" while capabilities answers "which Kit command should I run?".
12. `internal/feature/map.go::BuildProjectMap` already consumes `internal/instructions/registry.go` for instruction docs and emits JSON-able document/reference structs. This is a useful registry precedent, but its domain is repository document artifacts. Command capability metadata should probably live close to Cobra command declarations in `pkg/cli` unless it becomes shared outside CLI.
13. `internal/document/metadata.go::MetadataReference` defines the canonical front matter reference schema used here: `name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status` are required; `selector_type` must be one of `artifact`, `heading`, `symbol`, `command`, `url`, or `node_id` when present.
14. `docs/specs/0019-command-surface-simplification/SPEC.md` constrains root help grouping and hidden compatibility behavior. It says root help grouping applies only to the root command surface, hidden deprecated commands remain callable, and README/canonical docs should teach canonical commands rather than old surfaces.
15. `docs/specs/0030-reference-graph-routing/SPEC.md` constrains this brainstorm's metadata and clarifies that read plans must not inline referenced document contents. For `kit capabilities`, the analogous default should be compact command-selection data, with targeted lookup for full contract details.
16. `docs/specs/0020-versioned-instruction-model/SPEC.md` and `docs/agents/RLM.md` support the core context-cost thesis: always-loaded instruction files should stay thin, agents should use repo-local docs and indices just in time, and material context should be recorded in feature front matter references.
17. `docs/specs/0021-project-validation-and-instruction-registry/SPEC.md` supports separating RLM discovery from dispatch/subagent execution. `kit capabilities` should be a discovery index, not a subagent dispatcher, not a prompt launcher, and not a hidden workflow policy store.
18. `README.md` has the human command table under `## Commands`, and `docs/specs/0000_INIT_PROJECT.md` has the durable command contract under command sections such as `kit map [feature]` and `kit scaffold agents`. Both are likely documentation update targets after the command shape is finalized.
19. `docs/agents/README.md`, `docs/agents/RLM.md`, and `docs/agents/TOOLING.md` are the agent-facing docs most likely to need concise usage guidance. The guidance should say: run `kit capabilities --json` only when command choice is uncertain; prefer `kit capabilities <command> --json` after narrowing; avoid repeated `--full` reads; treat output as a discovery index, not persistent project context.
20. `pkg/cli/verify.go::verifyCmd` demonstrates why capability metadata needs mutation flags. `kit verify` is useful to agents, supports `--dry-run`, `--no-write`, `--json`, `--allow-shell`, and `--timeout`, and can write `.kit/runs` artifacts unless `--dry-run` or `--no-write` is used. A simple "read-only yes/no" field is insufficient for all commands; the catalog should expose mutation level plus file writes, network use, and git mutation as separate fields.
21. `pkg/cli/dispatch.go::dispatchCmd` demonstrates why command metadata needs conditional network/input notes. `kit dispatch` is prompt-only by default but `--pr` fetches unresolved PR review threads and `--coderabbit` filters those comments. Capability records should distinguish default behavior from flag-dependent network use.
22. The local worktree contains pre-existing `pkg/cli/ci*.go` files and a visible `ci` root command. `pkg/cli/ci.go::ciCmd` is diagnostic and says it does not edit source, rerun CI, commit, push, or mutate GitHub state, but its behavior is not read-only in the strict filesystem sense: `pkg/cli/ci_github.go::requireGHAuth` shells out to `gh auth status`, fetch helpers run `git` and `gh` subprocesses, and `pkg/cli/ci_github.go::cacheCIDefaultBranch` can write `.kit.yaml` when default branch metadata is discovered. Its capability record needs network/subprocess/file-write detail, not a flat mutation label.
23. `pkg/cli/ci.go::ciCmd` also has `--json`, `--dispatch`, `--copilot`, `--no-copilot`, `--pr`, `--run`, `--job`, `--workflow`, `--repo-path`, and `--log-lines` flags. Capability metadata should separate diagnostic GitHub reads from optional dispatch-editor behavior and optional Copilot-assisted diagnosis attempts.
24. `pkg/cli/status.go::statusCmd` demonstrates the existing pattern of preserving stable JSON payloads. A `capabilities` payload should include `schema_version` from v1 and tests should pin it so future shape changes are intentional.
25. `docs/references/tooling.md` is currently generic and has no durable command metadata guidance. It is optional as a future documentation target; agent-facing `docs/agents/*` and README/INIT command docs are higher-value for v1.
26. `UNKNOWN`: whether the local untracked `pkg/cli/ci*.go` implementation will be committed before or with `kit capabilities` is outside this brainstorm. The capabilities spec should require the catalog to reflect commands registered in the implementation worktree and should not implement, stabilize, or redesign `kit ci` as part of this feature.

## AFFECTED FILES

1. `pkg/cli/capabilities.go` (new): likely home for `capabilitiesCmd`, flags (`--json`, `--full`, `--search`), argument handling, compact/full payload rendering, targeted lookup, and command-local errors.
2. `pkg/cli/capabilities_catalog.go` (new or folded into `capabilities.go`): likely home for the explicit capability record model and static catalog. Keep folded into one file if concise; split only if the catalog makes `capabilities.go` hard to scan.
3. `pkg/cli/capabilities_test.go` (new): focused tests for compact JSON, targeted command JSON, full JSON, search filtering, unknown command errors, schema version stability, read-only behavior, and drift against registered commands.
4. `pkg/cli/root_help.go`: add `capabilities` to `commandOrder` and the Inspect & Repair `rootCommandSections` entry, likely adjacent to `map`.
5. `pkg/cli/root_help_test.go`: assert `capabilities` appears in grouped root help and stays in the intended section.
6. `README.md`: add `kit capabilities` to the Inspect & Repair command table and explain compact/targeted/full usage without encouraging persistent full-context loading.
7. `docs/specs/0000_INIT_PROJECT.md`: add the durable command contract for `kit capabilities`, including read-only behavior and JSON schema expectations.
8. `pkg/cli/ci*.go` (existing local worktree files): no direct edits are recommended for this feature, but the capability catalog must include `ci` metadata if the command remains registered when implementation begins.
9. `docs/agents/README.md`: optionally mention `kit capabilities --json` as a command-choice discovery index when the agent is uncertain which Kit command applies.
10. `docs/agents/RLM.md`: likely add one concise rule distinguishing command-surface discovery (`kit capabilities`) from document/reference discovery (`kit map <feature>` and `kit map <feature> --context`).
11. `docs/agents/TOOLING.md`: likely add the operational guidance to use targeted `kit capabilities <command> --json` after narrowing and avoid repeated `--full` reads.
12. `docs/references/tooling.md`: optional durable command-reference note if the guidance needs more detail than belongs in always-loaded agent docs.
13. `internal/*`: no new internal package is recommended for v1 unless later spec work shows non-CLI consumers need the command catalog. Keeping the catalog in `pkg/cli` minimizes abstraction and mirrors current Cobra ownership.

## DEPENDENCIES

References are tracked in front matter.

1. Runtime dependencies: no new third-party dependencies are expected. Existing `github.com/spf13/cobra` and the Go standard library `encoding/json`, `sort`, and `strings` are sufficient.
2. Config changes: none expected. The command should not read or mutate `.kit.yaml` for v1 unless needed to resolve project-local instruction docs, and current thesis does not require that.
3. Data migrations: none expected. The command emits derived static/runtime metadata and should not create persisted artifacts.
4. Network dependencies: none. The command must not call GitHub, package managers, remote registries, or any MCP/service endpoint.
5. File writes: none. Tests should prove no project files are created or modified by `kit capabilities`, including `.kit.yaml`, `.kit/state.json`, `.kit/runs`, feature docs, or notes.
6. Git mutation: none. Capability metadata should expose `git_mutation: false` for `capabilities` itself.
7. Capability-record dependencies: if `ci` remains registered, its record should expose that `kit ci` itself can call `git`/`gh`, read GitHub data, and conditionally write `.kit.yaml`; these are metadata facts about an adjacent command and must not become behavior inside `kit capabilities`.
8. Validation dependencies for implementation phase: `go test ./...`, `go run ./cmd/kit capabilities --json`, `go run ./cmd/kit capabilities dispatch --json`, `go run ./cmd/kit capabilities ci --json` when `ci` is registered, `go run ./cmd/kit capabilities --full --json`, and root help checks.

## QUESTIONS

1. Resolved default: include hidden/deprecated compatibility commands in the full catalog with explicit `hidden` and `deprecated` metadata, but keep compact default focused on canonical visible commands unless search/targeted lookup asks for a hidden command. This prevents agents from blindly using old surfaces while keeping drift checks possible.
2. Resolved default: support nested command paths in capability records (`scaffold agents`, `skill mine`, `prompt list`, `set prompt`, and `rules add/list/view/link`) because agents need the narrowest actual command, not only top-level namespaces.
3. Resolved default: make JSON the primary contract and require `--json` for machine-readable output, while allowing a concise human text view for direct terminal use. The user thesis specifically asks for compact JSON by default; the exact CLI should interpret this as `kit capabilities --json` being the agent path, not forcing raw JSON on humans who forget `--json`.
4. Resolved default: compact records include only the fields needed for command selection: command path, category, summary, mutation level, network use summary, file write summary, git mutation, hidden/deprecated state, important flags, and related commands. Full/targeted records include when-to-use, when-not-to-use, examples, caveats, and detailed flag behavior.
5. Resolved default: do not auto-generate all metadata from Cobra help. Cobra can supply names, short help, flags, hidden/deprecated status, and registration drift checks, but when-to-use, when-not-to-use, mutation level, examples, and related commands need intentional editorial metadata.
6. Resolved default: if `--search <term>` returns multiple matches, emit the same compact record shape filtered by command path, category, summary, flags, when-to-use, and related commands. Search should not imply `--full`.
7. Resolved default: unknown targeted commands should return a non-zero error with an actionable message and suggestions from known command paths when possible.
8. Resolved default: schema version should start at `1` as integer `schema_version`, not a string, matching `.kit/state.json`, verification runs, loop reports, and runstore indices.
9. Resolved default: `capabilities` belongs in root help under Inspect & Repair near `map`; it is an inspection/discovery command, not a workflow step.
10. Resolved default: include the local worktree's `ci` command in capability records if it is still registered during implementation, with explicit metadata for GitHub network reads, `git`/`gh` subprocesses, `.kit.yaml` default-branch cache writes, optional dispatch prompt behavior, and optional Copilot use. Do not implement or redesign `ci` in this feature.
11. Current understanding: 97%. No user-blocking questions remain for moving from brainstorm to `kit spec kit-capabilities`; the implementation spec can lock these defaults unless the user overrides them.

## OPTIONS

1. Option A - static CLI-local catalog plus Cobra drift checks.
   - Pros: simplest viable solution, explicit metadata for agent decision-making, no new package boundary, no config or migration surface, easy to test against `rootCmd.Commands()`.
   - Cons: metadata can drift from command behavior unless tests compare registered commands, flags, hidden/deprecated state, and root help sections to capability records.
2. Option B - generate metadata primarily from Cobra commands.
   - Pros: low duplication for names, flags, short help, hidden/deprecated state, and aliases.
   - Cons: cannot reliably infer when-to-use, when-not-to-use, mutation level, network use, examples, or related commands; risks shallow output that does not solve hallucinated workflows.
3. Option C - create an internal command registry used by Cobra and capabilities.
   - Pros: strongest long-term single source of truth if many commands need shared metadata.
   - Cons: larger refactor, touches many command files, higher blast radius, and premature for a read-only discovery command.
4. Option D - extend `kit map --context` to include command capabilities.
   - Pros: reuses existing read-only map and reference output patterns.
   - Cons: conflates document/reference routing with command selection; increases context cost and makes `map` less focused.
5. Recommendation: choose Option A for v1, with a small amount of Cobra-derived verification. It is the narrowest production-ready change that fits the user thesis and the current codebase.

## RECOMMENDED STRATEGY

1. Add `kit capabilities [command]` as a read-only top-level command in `pkg/cli/capabilities.go`.
2. Add flags:
   - `--json`: emit machine-readable JSON for agents.
   - `--full`: include detailed fields for every returned command.
   - `--search <term>`: filter compact results without loading the full catalog into context.
3. Use argument behavior:
   - no argument: return compact capability index.
   - `<command>` or nested command path: return one targeted detailed record when `--json` is set; if a human text mode exists, print a concise readable detail view.
   - unknown command: fail non-zero with suggestions.
4. Define a v1 payload:
   - top-level compact payload: `schema_version`, `kind`, `generated_by`, `commands`.
   - targeted payload: `schema_version`, `kind`, `command`.
   - full payload: `schema_version`, `kind`, `commands`, with detailed arrays included.
5. Define capability record fields:
   - `command`: canonical command path.
   - `category`: one of the root help categories where possible.
   - `summary`.
   - `when_to_use`.
   - `when_not_to_use`.
   - `mutation_level`: recommended values `none`, `writes_files`, `executes_commands`, `network`, `git`, or `destructive`; choose one primary value and keep booleans below for precision.
   - `network_use`: structured text/boolean noting default and flag-dependent behavior.
   - `file_writes`: structured text/boolean noting default and flag-dependent behavior.
   - `git_mutation`: boolean plus note; expected false for all current inspected commands except future explicit GitHub delivery commands, if any.
   - `important_flags`: names, short descriptions, and safety notes.
   - `examples`: compact command examples, only in targeted/full detail.
   - `related_commands`: command paths and relationship notes.
   - `hidden`, `deprecated`, and `aliases` where relevant.
6. Keep the `capabilities` command itself strictly read-only: no project-root requirement unless needed for future project-aware filtering, no config writes, no `.kit` writes, no network calls, no subprocess execution except normal CLI process execution, and no git commands.
7. Add drift tests:
   - registered visible root commands have capability records or an explicit exclusion list.
   - capability records reference real Cobra commands for canonical paths and nested paths.
   - root help includes `capabilities` under Inspect & Repair.
   - hidden/deprecated commands are either represented with hidden/deprecated metadata or explicitly excluded from compact default with tested rationale.
   - currently registered local commands such as `ci` either have capability records or are covered by an explicit exclusion with rationale.
8. Add output tests:
   - compact JSON includes `schema_version: 1`, `commands`, compact fields only, and `capabilities` itself.
   - targeted `dispatch` JSON returns exactly one command record and includes prompt-only/network notes.
   - targeted `ci` JSON exists when `ci` is registered and includes GitHub/network, subprocess, optional `.kit.yaml` cache-write, and optional dispatch/Copilot notes.
   - full JSON includes detailed arrays such as `when_to_use`, `when_not_to_use`, `examples`, and `related_commands`.
   - search filters without switching to full detail.
   - unknown command errors are actionable.
9. Update docs after the command behavior is stable:
   - `README.md` command table.
   - `docs/specs/0000_INIT_PROJECT.md` durable command contract.
   - `docs/agents/README.md`, `docs/agents/RLM.md`, and/or `docs/agents/TOOLING.md` with concise guidance: use `kit capabilities --json` only when command choice is uncertain; then use targeted lookup; avoid repeated `--full` reads; do not treat output as persistent repo context.
10. Avoid scope creep:
   - do not implement, stabilize, or redesign `kit ci` in this feature; reflect it only as command metadata if it remains registered in the implementation worktree.
   - do not create a new persistent command metadata document.
   - do not refactor existing command files into a shared registry unless the SPEC phase finds a hard requirement.
   - do not make `capabilities` call `kit map`, `kit status`, `kit check`, or any other command internally.

## NEXT STEP

Run `kit spec kit-capabilities` to turn this brainstorm into a binding `SPEC.md`.
