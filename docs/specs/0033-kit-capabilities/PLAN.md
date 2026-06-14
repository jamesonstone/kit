---
kit_metadata_version: 1
artifact: plan
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
skills:
  - name: rlm
    source: repo-local
    path: docs/agents/RLM.md
    trigger: analyze codebase; scan repository; large repository analysis; recursive language model context routing
    required: true
references:
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: artifact pipeline order, source-of-truth rules, read-only command expectations, populated-section rules, command/package constraints, and no hidden-state principle
    status: active
  - id: agent-routing
    name: Agent routing docs
    type: doc
    target: docs/agents/README.md
    selector_type: artifact
    selector: README.md
    relation: guides
    read_policy: must
    used_for: repo-local entrypoint and progressive disclosure routing for this plan phase
    status: active
  - id: workflow-rules
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: plan-phase source-of-truth order, clarification expectations, and docs-before-code sequencing
    status: active
  - id: rlm-rules
    name: RLM rules
    type: skill
    target: docs/agents/RLM.md
    selector_type: heading
    selector: Rules
    relation: guides
    read_policy: must
    used_for: just-in-time prior-work and codebase inspection bounds; selected as the planning-time skill for this feature
    status: active
  - id: guardrails
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    selector_type: heading
    selector: Completion Bar
    relation: constrains
    read_policy: must
    used_for: documentation completion bar, placeholder removal, no-git/no-production-edit constraints, and validation expectations
    status: active
  - id: tooling-doc
    name: Tooling docs
    type: doc
    target: docs/agents/TOOLING.md
    selector_type: heading
    selector: Skills
    relation: guides
    read_policy: conditional
    used_for: skills discovery, project-directory workflow, and RLM versus dispatch execution boundaries
    status: active
  - id: brainstorm
    name: Kit capabilities brainstorm
    type: feature
    target: docs/specs/0033-kit-capabilities/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, resolved defaults, codebase findings, options, and recommended implementation strategy
    status: active
  - id: spec
    name: Kit capabilities spec
    type: feature
    target: docs/specs/0033-kit-capabilities/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding requirements, acceptance criteria, edge cases, non-goals, and test evidence targets
    status: active
  - id: feature-map-command
    name: Feature map command
    type: command
    target: "kit map 0033-kit-capabilities"
    selector_type: command
    selector: "kit map 0033-kit-capabilities"
    relation: informs
    read_policy: evidence
    used_for: confirmed current plan phase, outgoing relationships, reference resolution, and feature artifact state
    status: active
  - id: progress-summary
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    selector_type: heading
    selector: FEATURE PROGRESS TABLE
    relation: informs
    read_policy: evidence
    used_for: prior-feature shortlist and feature phase/status tracking
    status: active
  - id: command-surface-simplification
    name: Command surface simplification
    type: feature
    target: docs/specs/0019-command-surface-simplification/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: root help category contract, canonical visible commands, hidden deprecated command behavior, and README command grouping
    status: active
  - id: reference-graph-routing
    name: Reference graph routing
    type: feature
    target: docs/specs/0030-reference-graph-routing/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: canonical front matter references schema, stable selectors, relation semantics, and compact read-plan precedent
    status: active
  - id: document-map-relationships
    name: Document map relationships
    type: feature
    target: docs/specs/0016-document-map-relationships/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: read-only map precedent and low-token orientation surface pattern
    status: active
  - id: instruction-model
    name: Versioned instruction model
    type: feature
    target: docs/specs/0020-versioned-instruction-model/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: guides
    read_policy: conditional
    used_for: thin table-of-contents instruction model and progressive disclosure guidance for agent-facing documentation
    status: active
  - id: instruction-registry
    name: Project validation and instruction registry
    type: feature
    target: docs/specs/0021-project-validation-and-instruction-registry/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: registry precedent and separation between RLM discovery and dispatch execution planning
    status: active
  - id: root-command
    name: Root command
    type: code
    target: pkg/cli/root.go
    selector_type: symbol
    selector: rootCmd
    relation: informs
    read_policy: conditional
    used_for: Cobra root command, public command registration evidence, version template, and Execute error handling
    status: active
  - id: root-help
    name: Root help
    type: code
    target: pkg/cli/root_help.go
    selector_type: symbol
    selector: rootCommandSections
    relation: constrains
    read_policy: must
    used_for: visible command categories, command ordering, root-only grouped help, and Inspect & Repair placement
    status: active
  - id: root-help-tests
    name: Root help tests
    type: code
    target: pkg/cli/root_help_test.go
    selector_type: symbol
    selector: TestRootHelpGroupsCanonicalCommands
    relation: verifies
    read_policy: evidence
    used_for: expected grouped help behavior, root command visibility assertions, and hidden deprecated command assertions
    status: active
  - id: map-command
    name: Map command
    type: code
    target: pkg/cli/map.go
    selector_type: symbol
    selector: mapCmd
    relation: informs
    read_policy: conditional
    used_for: read-only command precedent, JSON output pattern, feature map behavior, and distinction from command-surface discovery
    status: active
  - id: verify-command
    name: Verify command
    type: code
    target: pkg/cli/verify.go
    selector_type: symbol
    selector: verifyCmd
    relation: informs
    read_policy: conditional
    used_for: command metadata examples for execution, .kit/runs writes, dry-run, no-write, shell, timeout, and shared JSON helper behavior
    status: active
  - id: dispatch-command
    name: Dispatch command
    type: code
    target: pkg/cli/dispatch.go
    selector_type: symbol
    selector: dispatchCmd
    relation: informs
    read_policy: conditional
    used_for: command metadata examples for prompt-only orchestration, PR/network input, CodeRabbit filtering, and subagent planning
    status: active
  - id: status-command
    name: Status command
    type: code
    target: pkg/cli/status.go
    selector_type: symbol
    selector: statusCmd
    relation: informs
    read_policy: conditional
    used_for: command-local JSON pattern and backward-compatible payload caution
    status: active
  - id: ci-command
    name: CI command
    type: code
    target: pkg/cli/ci.go
    selector_type: symbol
    selector: ciCmd
    relation: informs
    read_policy: conditional
    used_for: local worktree command registration, GitHub Actions diagnostic surface, JSON flag, dispatch flag, and mutation metadata nuance
    status: active
  - id: ci-github-support
    name: CI GitHub support
    type: code
    target: pkg/cli/ci_github.go
    selector_type: symbol
    selector: cacheCIDefaultBranch
    relation: informs
    read_policy: conditional
    used_for: local ci subprocess calls, GitHub reads, gh authentication requirement, and .kit.yaml default-branch cache write metadata
    status: active
  - id: ci-tests
    name: CI command tests
    type: code
    target: pkg/cli/ci_test.go
    selector_type: symbol
    selector: TestRunCIDiagnosesDefaultBranchFailureAndCachesGitHubConfig
    relation: verifies
    read_policy: evidence
    used_for: existing local ci expectations for cache writes and JSON output when ci remains registered
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
  - id: readme-commands
    name: README command table
    type: doc
    target: README.md
    selector_type: heading
    selector: 🧰 Commands
    relation: guides
    read_policy: conditional
    used_for: user-facing command grouping and documentation update target after command behavior is stable
    status: active
  - id: init-project-command-contract
    name: Init project command contract
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    selector_type: heading
    selector: "8. CLI Commands"
    relation: guides
    read_policy: conditional
    used_for: durable user-facing command contract and documentation update target
    status: active
---
# PLAN

## SUMMARY

Implement `kit capabilities` as a CLI-local, read-only command catalog backed by explicit editorial metadata and guarded by Cobra drift tests. The implementation should add the smallest command surface needed for compact JSON discovery, targeted detail, full detail, search filtering, root-help discoverability, and concise documentation updates while preserving existing command behavior and JSON payloads.

## APPROACH

1. [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-05][SPEC-08][SPEC-09][SPEC-15][SPEC-26][ACCEPT-02][ACCEPT-04][ACCEPT-06][ACCEPT-07][ACCEPT-12] Build the command as a read-only `pkg/cli` surface and keep planning metadata in RLM mode. The command should not require a Kit project root for basic catalog output and must not load config, write files, call network services, execute subprocesses, run git, or delegate to other Kit commands.
2. [PLAN-02][SPEC-04][SPEC-07][SPEC-11][SPEC-12][SPEC-16][SPEC-17][SPEC-18][SPEC-19][SPEC-20][ACCEPT-03][ACCEPT-04][ACCEPT-05][ACCEPT-10][ACCEPT-11] Use an explicit static catalog for safety-relevant metadata, with Cobra used for registration, visibility, deprecation, and flag drift checks only. Cobra help is not a sufficient source for when-to-use guidance, mutation notes, examples, caveats, or related-command guidance.
3. [PLAN-03][SPEC-03][SPEC-06][SPEC-08][SPEC-09][SPEC-10][SPEC-21][ACCEPT-02][ACCEPT-04][ACCEPT-06][ACCEPT-07][ACCEPT-08] Implement output as views over the same records: compact index, targeted detail, full detail, and compact search results. Keep `schema_version: 1` stable and reject ambiguous flag combinations with actionable errors.
4. [PLAN-04][SPEC-22][SPEC-23][SPEC-24][SPEC-25][ACCEPT-09][ACCEPT-13] Add root-help and documentation integration after the command contract is stable. Place `capabilities` under Inspect & Repair near `map`, and teach agents to use compact discovery only when command choice is uncertain, followed by targeted lookup instead of repeated full reads.
5. [PLAN-05][SPEC-19][SPEC-27][SPEC-28][ACCEPT-01][ACCEPT-05] Treat local `ci` files as an adjacent command surface, not part of this feature's implementation scope. Include `ci` metadata if the command remains registered, but do not implement, stabilize, or redesign `ci`.

Tradeoff decisions:

1. Choose a `pkg/cli` catalog instead of a new internal command registry for v1. This minimizes blast radius and keeps command-selection metadata close to Cobra ownership.
2. Keep the catalog static and intentional, then test it against registered commands. This duplicates some command facts but preserves the editorial safety metadata that agents need.
3. Do not extend `kit map --context`. Map remains document/reference routing; capabilities is command-selection routing.
4. Make `--json` the machine-readable contract while allowing concise human text for terminal use.
5. Reject `--search` with a targeted command. Search returns a filtered compact index and should not be mixed with a single-command lookup.
6. Reject `--full` with a targeted command. Targeted lookup already returns detailed data, while `--full` means all included commands.
7. Include hidden/deprecated compatibility commands only in full output or direct targeted lookup, with explicit metadata. Compact default should steer agents toward canonical visible commands.

## COMPONENTS

1. `pkg/cli/capabilities.go`
   - Own the Cobra command, flags, argument validation, mode selection, human rendering, JSON rendering, and actionable errors.
   - Accept no more than one targeted command path, while allowing nested command paths by joining positional arguments.
   - Keep command execution side-effect free.
2. `pkg/cli/capabilities_catalog.go`
   - Own capability record types, schema version constants, compact/detail projections, deterministic sorting, lookup, search, and suggestions.
   - Stay in `pkg/cli` unless a later feature creates a real non-CLI consumer.
   - Split from `capabilities.go` only if it keeps each source file easier to scan; do not introduce an abstraction layer for its own sake.
3. `pkg/cli/capabilities_test.go`
   - Cover compact JSON, targeted JSON, full JSON, search filtering, invalid flag combinations, unknown-command suggestions, schema stability, and read-only behavior.
   - Include drift tests comparing catalog records to registered Cobra commands and nested command paths.
4. `pkg/cli/root_help.go` and `pkg/cli/root_help_test.go`
   - Add `capabilities` to command ordering and the Inspect & Repair section.
   - Assert help placement and default compact exclusion of hidden/deprecated commands.
5. User-facing documentation
   - Update `README.md` and `docs/specs/0000_INIT_PROJECT.md` after behavior is implemented.
   - Keep wording short and command-focused.
6. Agent-facing documentation
   - Update `docs/agents/README.md`, `docs/agents/RLM.md`, and/or `docs/agents/TOOLING.md` with one concise discovery rule.
   - Distinguish command discovery from document/reference routing.
7. Progress tracking
   - Keep `docs/PROJECT_PROGRESS_SUMMARY.md` aligned with the highest completed artifact after planning, tasks, implementation, and reflection phases.

## DATA

No persisted data, migration, config field, generated markdown catalog, hidden database, lock file, or `.kit` artifact is required.

Payloads:

1. Compact index: `{ "schema_version": 1, "kind": "capabilities_index", "generated_by": "kit capabilities", "commands": [...] }`.
2. Targeted detail: `{ "schema_version": 1, "kind": "capability_detail", "generated_by": "kit capabilities", "command": {...} }`.
3. Full detail: `{ "schema_version": 1, "kind": "capabilities_full", "generated_by": "kit capabilities", "commands": [...] }`.
4. Search result: `{ "schema_version": 1, "kind": "capabilities_search", "generated_by": "kit capabilities", "query": "...", "commands": [...] }`.

Record shapes:

1. Compact fields: `command`, `category`, `summary`, `mutation_level`, `network_use`, `file_writes`, `git_mutation`, `hidden`, `deprecated`, `important_flags`, and `related_commands`.
2. Detailed fields add `when_to_use`, `when_not_to_use`, `examples`, `caveats`, and detailed flag behavior.
3. `mutation_level` values should be a small ordered enum: `none`, `writes_files`, `executes_commands`, `network`, `git`, `destructive`.
4. Safety fields should distinguish default behavior from flag-dependent behavior. Use structured notes rather than a flat boolean when the nuance affects command choice.
5. Search should match command path, category, summary, flags, when-to-use text, and related commands, then return compact records only.

## INTERFACES

Commands:

1. `kit capabilities`
   - Human text compact index.
   - No project-root requirement.
   - No writes, subprocesses, network calls, git commands, or delegated Kit command execution.
2. `kit capabilities --json`
   - Compact JSON index for agents.
   - Includes visible canonical root commands and compact nested commands needed for direct command choice.
3. `kit capabilities <command> --json`
   - Detailed JSON for exactly one top-level or nested command path.
   - Supports direct lookup of hidden/deprecated commands while labeling them explicitly.
4. `kit capabilities --full --json`
   - Detailed JSON for all included records, including hidden/deprecated compatibility records when represented.
   - Intended for uncommon audits, not repeated command-choice routing.
5. `kit capabilities --search <term> --json`
   - Compact filtered JSON.
   - Zero matches returns a successful empty result with the original query.

Invalid combinations:

1. `kit capabilities --search <term> <command>` exits non-zero with guidance to choose either search or targeted lookup.
2. `kit capabilities --full <command>` exits non-zero with guidance that targeted lookup already returns detailed command data.
3. Unknown targeted command paths exit non-zero with suggestions when a close command path exists.

Files and artifacts touched during implementation:

1. Product code: `pkg/cli/capabilities.go`, optional `pkg/cli/capabilities_catalog.go`, `pkg/cli/root_help.go`.
2. Tests: `pkg/cli/capabilities_test.go`, `pkg/cli/root_help_test.go`.
3. Docs: `README.md`, `docs/specs/0000_INIT_PROJECT.md`, selected `docs/agents/*` files.
4. Feature docs and rollup: `docs/specs/0033-kit-capabilities/*`, `docs/PROJECT_PROGRESS_SUMMARY.md`.
5. Not touched: `.kit.yaml`, `.kit/state.json`, `.kit/runs`, `.kit/loops`, generated artifacts, Git state, and existing command JSON contracts outside `capabilities`.

## DEPENDENCIES

References are tracked in front matter.

Runtime dependencies:

1. Use existing Cobra and Go standard library packages only.
2. No new third-party module, external service, MCP tool, dataset, asset, or generated registry is required.
3. No config, project-root, migration, or persisted state dependency is required.

## RISKS

1. Metadata drift from Cobra registration or real command behavior.
   - Mitigation: add drift tests for visible root commands, selected nested commands, hidden/deprecated policy, aliases, and important flags.
2. Safety metadata becomes too coarse for commands with flag-dependent behavior.
   - Mitigation: require default and flag-dependent notes for commands such as `verify`, `dispatch`, and registered `ci`.
3. Compact output grows until it stops being useful as a discovery index.
   - Mitigation: keep detailed fields out of compact records and reserve full output for explicit `--full --json`.
4. Hidden/deprecated commands accidentally appear as recommended paths.
   - Mitigation: exclude them from compact default and test direct/full handling separately.
5. Local `ci` worktree state changes before implementation.
   - Mitigation: make the catalog reflect commands registered in the implementation worktree and scope `ci` work to metadata only.
6. A shared registry refactor expands the feature beyond the SPEC.
   - Mitigation: keep v1 ownership in `pkg/cli`; revisit only through a later spec if another consumer needs command metadata.
7. Read-only guarantees regress through future project-root or config loading.
   - Mitigation: add no-write tests around `.kit.yaml`, `.kit/state.json`, `.kit/runs`, `.kit/loops`, feature docs, and notes.

## TESTING

Evidence mapped to acceptance:

1. Unit tests for mode selection and JSON rendering cover compact index, targeted detail, full detail, search result, zero search matches, invalid flag combinations, and unknown-command suggestions. Evidence for [ACCEPT-02], [ACCEPT-04], [ACCEPT-06], [ACCEPT-07], and [ACCEPT-08].
2. Drift tests compare the catalog to Cobra registration for visible root commands, represented hidden/deprecated commands, selected nested commands, aliases, and important flags. Evidence for [ACCEPT-03], [ACCEPT-10], and [ACCEPT-11].
3. Root help tests assert `capabilities` appears under Inspect & Repair and hidden/deprecated commands stay out of default root help. Evidence for [ACCEPT-09] and [ACCEPT-11].
4. Side-effect tests run compact, targeted, full, search, and error paths from a temporary project and assert no `.kit.yaml`, `.kit/state.json`, `.kit/runs`, `.kit/loops`, feature docs, or notes are created or modified. Evidence for [ACCEPT-12].
5. Documentation checks verify README, `docs/specs/0000_INIT_PROJECT.md`, and agent-facing docs describe compact, targeted, full, and search usage without repeated full-context loading. Evidence for [ACCEPT-13].
6. Feature-doc validation runs `kit map 0033-kit-capabilities` and `kit check 0033-kit-capabilities`. Evidence for [ACCEPT-14] and [ACCEPT-15].
7. Repository verification runs `go test ./...` after implementation. Evidence for [ACCEPT-01].
