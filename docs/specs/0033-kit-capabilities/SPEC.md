---
kit_metadata_version: 1
artifact: spec
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
    used_for: artifact pipeline order, source-of-truth rules, read-only command expectations, populated-section rules, and public command constraints
    status: active
  - id: agent-routing
    name: Agent routing docs
    type: doc
    target: docs/agents/README.md
    selector_type: artifact
    selector: README.md
    relation: guides
    read_policy: must
    used_for: repo-local entrypoint and progressive disclosure routing for this spec phase
    status: active
  - id: workflow-rules
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: spec-driven classification, source-of-truth order, and clarification expectations
    status: active
  - id: rlm-rules
    name: RLM rules
    type: skill
    target: docs/agents/RLM.md
    selector_type: heading
    selector: Rules
    relation: guides
    read_policy: must
    used_for: just-in-time prior-work and codebase inspection bounds; selected as the execution-time skill for this feature
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
    used_for: skills discovery, project-directory workflow, and secondary input boundaries
    status: active
  - id: brainstorm
    name: Kit capabilities brainstorm
    type: feature
    target: docs/specs/0033-kit-capabilities/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, resolved defaults, codebase findings, options, and recommended strategy
    status: active
  - id: feature-map-command
    name: Feature map command
    type: command
    target: "kit map 0033-kit-capabilities"
    selector_type: command
    selector: "kit map 0033-kit-capabilities"
    relation: informs
    read_policy: evidence
    used_for: confirmed current spec phase, outgoing relationships, reference resolution, and feature artifact state
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
    used_for: grouped root help contract, canonical visible commands, hidden deprecated command behavior, and command documentation expectations
    status: active
  - id: reference-graph-routing
    name: Reference graph routing
    type: feature
    target: docs/specs/0030-reference-graph-routing/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: canonical front matter references schema and compact read-plan precedent
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
    used_for: thin table-of-contents instruction model and progressive disclosure guidance
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
    used_for: Cobra root command and public command registration evidence
    status: active
  - id: root-help
    name: Root help
    type: code
    target: pkg/cli/root_help.go
    selector_type: symbol
    selector: rootCommandSections
    relation: constrains
    read_policy: must
    used_for: visible command categories, command ordering, and root help discoverability contract
    status: active
  - id: root-help-tests
    name: Root help tests
    type: code
    target: pkg/cli/root_help_test.go
    selector_type: symbol
    selector: TestRootHelpGroupsCanonicalCommands
    relation: verifies
    read_policy: evidence
    used_for: expected grouped help behavior and hidden deprecated command assertions
    status: active
  - id: map-command
    name: Map command
    type: code
    target: pkg/cli/map.go
    selector_type: symbol
    selector: mapCmd
    relation: informs
    read_policy: conditional
    used_for: read-only command precedent and JSON output pattern
    status: active
  - id: verify-command
    name: Verify command
    type: code
    target: pkg/cli/verify.go
    selector_type: symbol
    selector: verifyCmd
    relation: informs
    read_policy: conditional
    used_for: command metadata examples for execution, .kit/runs writes, dry-run, no-write, shell, timeout, and JSON behavior
    status: active
  - id: dispatch-command
    name: Dispatch command
    type: code
    target: pkg/cli/dispatch.go
    selector_type: symbol
    selector: dispatchCmd
    relation: informs
    read_policy: conditional
    used_for: command metadata examples for prompt-only orchestration, PR/network input, and subagent planning
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
    used_for: local ci subprocess calls, GitHub reads, gh authentication requirement, and .kit.yaml default-branch cache write
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
    used_for: durable user-facing command contract and documentation update target
    status: active
---
# SPEC

## SUMMARY

Add `kit capabilities` as a read-only command-discovery surface that exposes Kit's command catalog as compact, versioned, agent-readable JSON with targeted detail on demand. The feature must help agents choose existing commands deliberately while keeping always-loaded instructions small and avoiding product-code mutation during discovery.

## PROBLEM

Coding agents working in Kit-managed repositories currently infer command choice from static instruction files, README tables, or broad repository context. That makes command selection vulnerable to stale guidance, hallucinated workflows, unnecessary full-context reads, and unsafe assumptions about whether a command writes files, calls network services, executes subprocesses, or mutates git state.

Kit already has document-context routing through `kit map` and repo-local RLM docs, but it does not have an equivalent runtime surface for command-selection metadata. Agents need a compact way to answer "which Kit command should I use now?" before loading detailed command contracts or running higher-risk commands.

## GOALS

- Expose a top-level `kit capabilities` command for command-surface discovery.
- Make `kit capabilities --json` the primary machine-readable contract for coding agents.
- Provide compact default JSON suitable for command selection without full command-contract loading.
- Provide targeted command detail for one command or nested command path after the agent narrows its choice.
- Provide optional full and search modes without encouraging repeated full-catalog reads.
- Describe command behavior in safety-relevant terms: file writes, network use, subprocess execution, git mutation, destructive behavior, hidden/deprecated status, important flags, examples, and related commands.
- Keep the command itself read-only: no file writes, no `.kit` artifacts, no `.kit.yaml` changes, no network calls, no git commands, and no delegated execution of other Kit commands.
- Update agent-facing and user-facing docs so agents use `kit capabilities` as a discovery index only when command choice is uncertain.
- Add validation that keeps capability records aligned with registered commands and root help visibility.

## NON-GOALS

- Do not implement, stabilize, or redesign `kit ci`; only describe it in capability metadata when it is registered in the implementation worktree.
- Do not replace `kit map`, `kit map --context`, `kit status`, `kit check`, or repo-local RLM docs.
- Do not make `kit capabilities` call other Kit commands internally.
- Do not introduce a hidden database, lock file, generated markdown catalog, persistent command registry document, or external service dependency.
- Do not infer detailed safety metadata solely from Cobra help text.
- Do not remove or make agents prefer hidden deprecated compatibility commands.
- Do not change existing JSON payloads for `status`, `map`, `verify`, `trace`, `replay`, `state`, `eval`, `loop`, or local `ci` output.
- Do not add GitHub delivery, branch, commit, push, issue, or PR behavior.

## USERS

- Coding agents choosing among Kit commands during repository work.
- Humans delegating work to coding agents and wanting safer, more deliberate command selection.
- Kit maintainers adding or changing commands who need command-discovery drift to fail in tests.
- Automation that needs a stable command capability index before deciding which Kit command to invoke.

## SKILLS

Skills are tracked in front matter.

## RELATIONSHIPS

Relationships are tracked in front matter.

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

- [SPEC-01] Kit must expose a visible top-level command named `capabilities`.
- [SPEC-02] `kit capabilities --json` must emit valid JSON with integer `schema_version: 1`.
- [SPEC-03] The default JSON shape must be compact and optimized for command selection, not full command documentation.
- [SPEC-04] Compact command records must include enough metadata for a coding agent to decide whether to inspect or run a command:
  - command path
  - category
  - summary
  - mutation level
  - network use summary
  - file-write summary
  - git mutation summary
  - hidden/deprecated state
  - important flags
  - related commands
- [SPEC-05] Targeted lookup must support top-level commands and nested command paths.
- [SPEC-06] Targeted JSON output must return exactly one command record when the command path is known.
- [SPEC-07] Targeted records must include detailed command-selection fields:
  - when to use
  - when not to use
  - examples
  - caveats
  - detailed flag behavior
  - related command notes
- [SPEC-08] `--full --json` must emit detailed records for every included command.
- [SPEC-09] `--search <term> --json` must return compact records filtered by command path, category, summary, flags, when-to-use text, and related commands without implying full output.
- [SPEC-10] Unknown targeted command paths must fail non-zero with an actionable error and suggestions from known command paths when possible.
- [SPEC-11] The command must include capability metadata for visible canonical root commands.
- [SPEC-12] Hidden or deprecated compatibility commands must either appear only in full/targeted output with explicit hidden/deprecated metadata or be excluded by a tested explicit rationale.
- [SPEC-13] The compact default must avoid steering agents toward hidden deprecated compatibility commands.
- [SPEC-14] Command metadata must distinguish the behavior of the capability command itself from the behavior of the commands it describes.
- [SPEC-15] The `capabilities` command itself must be read-only and must not:
  - require a project root for basic catalog output
  - write files
  - write `.kit.yaml`
  - write `.kit/state.json`
  - write `.kit/runs`
  - write `.kit/loops`
  - create or edit feature docs or notes
  - run git commands
  - call network services
  - execute other Kit commands
- [SPEC-16] Capability records must represent flag-dependent behavior where safety properties change.
- [SPEC-17] Capability records for `kit verify` must expose that default verification can execute declared checks and write `.kit/runs` evidence, while `--dry-run` and `--no-write` reduce that behavior.
- [SPEC-18] Capability records for `kit dispatch` must expose that default behavior is prompt output, while `--pr` can fetch PR review data and `--coderabbit` filters those comments.
- [SPEC-19] If `kit ci` is registered in the implementation worktree, its capability record must expose GitHub/network reads, `git`/`gh` subprocess use, optional `.kit.yaml` default-branch cache writes, optional dispatch prompt behavior, and optional Copilot-assisted diagnosis attempts.
- [SPEC-20] Capability records must include nested commands that agents naturally need to choose directly, including `scaffold agents`, `skill mine`, `prompt list`, `set prompt`, and `rules add/list/view/link`.
- [SPEC-21] The JSON contract must be stable enough for tests and downstream agents; future incompatible changes must require an intentional schema-version change.
- [SPEC-22] Root help must list `capabilities` under the Inspect & Repair category.
- [SPEC-23] User-facing docs must document `kit capabilities` in the Inspect & Repair command group.
- [SPEC-24] Agent-facing docs must instruct agents to run `kit capabilities --json` only when command choice is uncertain, then prefer targeted lookup over repeated `--full` reads.
- [SPEC-25] Agent-facing docs must clarify that `kit capabilities` answers command-selection questions, while `kit map` and `kit map --context` answer document/reference-routing questions.
- [SPEC-26] Downstream planning for this feature must preserve RLM routing and record `parallelization_mode: "rlm"` in planning notes or execution metadata.
- [SPEC-27] The implementation must not add third-party dependencies.
- [SPEC-28] The implementation must preserve existing command behavior and existing command JSON payloads outside the new `capabilities` surface.
- [SPEC-29] Every new or changed Kit command, subcommand, flag, alias, prompt surface, or command behavior extension must update `kit capabilities` in the same change.
- [SPEC-30] Human detail output must include agent-readable guidance for safe command choice, including when to use the command, when not to use it, examples, caveats when present, important flag safety notes, and related commands.

## ACCEPTANCE

- [ACCEPT-01] `go test ./...` exits 0.
- [ACCEPT-02] `go run ./cmd/kit capabilities --json` emits valid JSON with `schema_version: 1` and a compact `commands` list.
- [ACCEPT-03] The compact `commands` list includes `capabilities` itself and all visible canonical root commands, except any command covered by an explicit tested exclusion.
- [ACCEPT-04] `go run ./cmd/kit capabilities dispatch --json` emits a targeted payload with exactly one command record and includes prompt-only plus flag-dependent PR/network notes.
- [ACCEPT-05] If `ci` is registered, `go run ./cmd/kit capabilities ci --json` emits a targeted payload that identifies GitHub/network reads, `git`/`gh` subprocess use, optional `.kit.yaml` cache writes, optional dispatch behavior, and optional Copilot behavior.
- [ACCEPT-06] `go run ./cmd/kit capabilities --full --json` emits detailed command records with when-to-use, when-not-to-use, examples, caveats, important flags, and related commands.
- [ACCEPT-07] `go run ./cmd/kit capabilities --search verify --json` emits filtered compact results and does not include full-only detail fields.
- [ACCEPT-08] `go run ./cmd/kit capabilities does-not-exist --json` exits non-zero with an actionable unknown-command message and suggestions when possible.
- [ACCEPT-09] Tests prove `capabilities` is listed under Inspect & Repair in root help.
- [ACCEPT-10] Tests prove visible registered root commands have capability records or explicit tested exclusions.
- [ACCEPT-11] Tests prove hidden/deprecated compatibility commands are not shown in the compact default unless directly targeted or included in full output with hidden/deprecated metadata.
- [ACCEPT-12] Tests prove `kit capabilities --json`, targeted lookup, `--full --json`, `--search --json`, and unknown-command paths do not create or modify `.kit.yaml`, `.kit/state.json`, `.kit/runs`, `.kit/loops`, feature docs, or notes.
- [ACCEPT-13] README, `docs/specs/0000_INIT_PROJECT.md`, and agent-facing docs describe compact, targeted, full, and search usage without encouraging repeated full-context loading.
- [ACCEPT-14] `kit map 0033-kit-capabilities` resolves the spec relationships and material references after this spec is complete.
- [ACCEPT-15] `kit check 0033-kit-capabilities` passes after the spec exists and is populated.
- [ACCEPT-16] Tests prove visible command records include agent guidance fields and targeted human detail output renders them.
- [ACCEPT-17] Repository rules document that command-surface changes must update `kit capabilities`.

## EDGE-CASES

- A registered visible command lacks a capability record.
- A capability record references a command path that Cobra does not recognize.
- A nested command has a visible parent but the nested command is hidden, deprecated, or absent.
- A hidden deprecated command is directly targeted by name.
- Search returns zero matches.
- Search returns both canonical and deprecated matches.
- A command has default read-only behavior but a flag that introduces network calls, file writes, subprocess execution, or prompt/editor behavior.
- A command has aliases or compatibility names that should not be promoted in compact output.
- `kit capabilities` runs outside a Kit-managed project.
- `kit capabilities --json` runs in a repository with dirty worktree state.
- `--full` and targeted lookup are requested together.
- `--search` and targeted lookup are requested together.
- Human text output is requested without `--json`.
- The implementation worktree contains local commands not present in the current release, such as `ci`.
- Future commands are added without updating capability metadata.

## OPEN-QUESTIONS

none
