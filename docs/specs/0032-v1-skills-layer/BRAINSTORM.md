---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0032
  slug: v1-skills-layer
  dir: 0032-v1-skills-layer
relationships:
  - type: builds_on
    target: 0006-skill-mine-command
  - type: related_to
    target: 0025-v0-prompt-library
  - type: builds_on
    target: 0026-front-matter-integration
  - type: depends_on
    target: 0030-reference-graph-routing
  - type: depends_on
    target: 0031-executable-verification-harness
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0032-v1-skills-layer
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input; no usable files beyond placeholders were found during this pass
    status: optional
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: document-first source-of-truth rules, filesystem-only state, visible .kit artifact constraints, and no hidden database/runtime constraints
    status: active
  - id: agents-entrypoint
    name: Agents docs entrypoint
    type: doc
    target: docs/agents/README.md
    selector_type: heading
    selector: Runtime Routing
    relation: guides
    read_policy: must
    used_for: repo-local context routing and minimal document loading
    status: active
  - id: workflows-doc
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: formal planning workflow, source-of-truth order, and clarification protocol
    status: active
  - id: rlm-doc
    name: RLM rules
    type: doc
    target: docs/agents/RLM.md
    selector_type: heading
    selector: Rules
    relation: guides
    read_policy: must
    used_for: prior-work shortlist and just-in-time codebase research bounds
    status: active
  - id: guardrails-doc
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    selector_type: heading
    selector: Completion Bar
    relation: constrains
    read_policy: must
    used_for: populated brainstorm requirements, validation honesty, and no guessed file behavior
    status: active
  - id: tooling-doc
    name: Tooling and skills routing
    type: doc
    target: docs/agents/TOOLING.md
    selector_type: heading
    selector: Skills
    relation: constrains
    read_policy: must
    used_for: current canonical skills root and feature-scope skills metadata behavior
    status: active
  - id: references-index
    name: References index
    type: doc
    target: docs/references/README.md
    selector_type: heading
    selector: Purpose
    relation: guides
    read_policy: conditional
    used_for: placement precedent for durable repo-wide reference documents such as a master skills document
    status: active
  - id: spec-skills-discovery
    name: Spec skills discovery
    type: prior feature doc
    target: docs/specs/0009-spec-skills-discovery/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: existing feature-level skills discovery contract and .agents/skills discovery precedent
    status: active
  - id: kit-map-0032
    name: Current feature map
    type: command
    target: kit map 0032-v1-skills-layer
    selector_type: command
    selector: kit map 0032-v1-skills-layer
    relation: verifies
    read_policy: evidence
    used_for: confirmed brainstorm phase, no existing relationships, and optional notes reference resolution
    status: active
  - id: progress-summary
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    selector_type: heading
    selector: FEATURE PROGRESS TABLE
    relation: informs
    read_policy: conditional
    used_for: prior-feature shortlist and current feature phase state
    status: active
  - id: skillopt-paper
    name: SkillOpt paper
    type: url
    target: https://arxiv.org/abs/2605.23904
    relation: guides
    read_policy: must
    used_for: bounded text-space skill optimization, held-out validation gating, rejected-edit buffer, slow/meta update, and compact deployable skill artifact concepts
    status: active
  - id: skillopt-project-page
    name: SkillOpt project page
    type: url
    target: https://microsoft.github.io/SkillOpt/
    relation: informs
    read_policy: evidence
    used_for: operational loop summary for rollout, reflection, bounded edits, gate, memory, ablations, and transfer behavior
    status: active
  - id: skill-mine-spec
    name: Skill mine command spec
    type: prior feature doc
    target: docs/specs/0006-skill-mine-command/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: existing canonical skill mining command, .agents/skills root, Claude mirror, prompt-only non-mutating contract, and skill audit semantics
    status: active
  - id: prompt-library-plan
    name: Prompt library plan
    type: prior feature doc
    target: docs/specs/0025-v0-prompt-library/PLAN.md
    selector_type: heading
    selector: COMPONENTS
    relation: informs
    read_policy: conditional
    used_for: layered local/global/built-in precedence, YAML config extension pattern, deterministic merge behavior, and shadow metadata precedent
    status: active
  - id: front-matter-spec
    name: Front matter integration spec
    type: prior feature doc
    target: docs/specs/0026-front-matter-integration/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: canonical feature metadata, relationships, references, and skills front matter behavior
    status: active
  - id: reference-graph-spec
    name: Reference graph routing spec
    type: prior feature doc
    target: docs/specs/0030-reference-graph-routing/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: canonical references schema, read policies, selectors, map rendering, and context-plan behavior
    status: active
  - id: verification-harness-spec
    name: Executable verification harness spec
    type: prior feature doc
    target: docs/specs/0031-executable-verification-harness/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: verify, trace, replay, run artifact, eval, and pointer-only state contracts
    status: active
  - id: config-skills-path
    name: Config skills path
    type: code
    target: internal/config/config.go
    selector_type: symbol
    selector: Config.SkillsPath
    relation: constrains
    read_policy: must
    used_for: current configurable canonical skills directory defaulting to .agents/skills
    status: active
  - id: skill-command
    name: Skill command
    type: code
    target: pkg/cli/skill.go
    selector_type: symbol
    selector: runSkillMine
    relation: constrains
    read_policy: must
    used_for: existing skill root, mine subcommand behavior, feature resolution, and prompt-only output semantics
    status: active
  - id: skill-prompt-builder
    name: Skill prompt builder
    type: code
    target: pkg/cli/skill_prompt.go
    selector_type: symbol
    selector: buildSkillMinePrompt
    relation: constrains
    read_policy: must
    used_for: current skill bundle format, audit guidance, source-of-truth root, and Claude mirror behavior
    status: active
  - id: verify-task-bundles
    name: Verification task bundles
    type: code
    target: internal/verify/tasks.go
    selector_type: symbol
    selector: LoadTaskBundles
    relation: implements
    read_policy: must
    used_for: parsing TASKS.md VERIFY fields into typed commands and task bundles
    status: active
  - id: verify-run-model
    name: Verification run model
    type: code
    target: internal/verify/execute.go
    selector_type: symbol
    selector: Run
    relation: implements
    read_policy: must
    used_for: existing run status, command result, dry-run, no-declared-checks, and execution result data model
    status: active
  - id: runstore
    name: Run artifact store
    type: code
    target: internal/runstore/store.go
    selector_type: symbol
    selector: Write
    relation: implements
    read_policy: must
    used_for: visible .kit/runs artifact persistence, index updates, bounded output, and latest-run lookup
    status: active
  - id: verify-cli
    name: Verify CLI
    type: code
    target: pkg/cli/verify.go
    selector_type: symbol
    selector: runVerify
    relation: implements
    read_policy: must
    used_for: command surface and output behavior for replay/eval validation gates
    status: active
  - id: trace-cli
    name: Trace CLI
    type: code
    target: pkg/cli/trace.go
    selector_type: symbol
    selector: runTrace
    relation: implements
    read_policy: conditional
    used_for: trace listing and run-detail output for skill history and score provenance
    status: active
  - id: replay-cli
    name: Replay CLI
    type: code
    target: pkg/cli/replay.go
    selector_type: symbol
    selector: runReplay
    relation: implements
    read_policy: must
    used_for: current single-run replay semantics and parent-run comparison model
    status: active
  - id: state-generator
    name: Generated state
    type: code
    target: internal/state/state.go
    selector_type: symbol
    selector: Generate
    relation: implements
    read_policy: conditional
    used_for: pointer-only .kit/state.json generation and latest verification exposure
    status: active
  - id: eval-runner
    name: Harness eval runner
    type: code
    target: internal/eval/eval.go
    selector_type: symbol
    selector: Run
    relation: implements
    read_policy: conditional
    used_for: existing local harness eval pattern and negative-path regression cases
    status: active
  - id: reflect-evidence
    name: Reflection evidence lookup
    type: code
    target: pkg/cli/reflect.go
    selector_type: symbol
    selector: latestVerificationEvidenceStep
    relation: uses
    read_policy: conditional
    used_for: existing latest-run evidence consumption in reflection prompts
    status: active
  - id: metadata-model
    name: Document metadata model
    type: code
    target: internal/document/metadata.go
    selector_type: symbol
    selector: MetadataReference
    relation: constrains
    read_policy: must
    used_for: front matter reference fields, enums, validation, and skill metadata precedent
    status: active
  - id: feature-map
    name: Feature map references
    type: code
    target: internal/feature/map.go
    selector_type: symbol
    selector: BuildProjectMap
    relation: uses
    read_policy: conditional
    used_for: map relationship/reference rendering and resolver warnings
    status: active
  - id: promptlib-merge
    name: Prompt library merge
    type: code
    target: internal/promptlib/merge.go
    selector_type: symbol
    selector: Merge
    relation: informs
    read_policy: conditional
    used_for: deterministic layered precedence pattern for active skill resolution
    status: active
---
# BRAINSTORM

## SUMMARY

`v1-skills-layer` should extend Kit from one-shot, prompt-only skill mining into a document-backed skill optimization lifecycle that can produce, validate, score, promote, reject, version, resolve, and garbage-collect bounded behavioral skill modules. The canonical lifecycle ledger should be a master skills document, while generated `.kit/` artifacts remain evidence, cache, export, or staging surfaces rather than the source of truth.

## USER THESIS

The user thesis is preserved below as the research seed for this brainstorm phase. Approved clarifications now supersede the initial `.kit/skills`-as-ledger assumption: use a master skills document as the canonical skills-layer artifact, keep candidate authoring external or prompt-assisted, let `kit reflect` notice and propose recursive improvement, and keep `kit skill` commands responsible for scoring, promotion, rejection, history, and pruning.

### Context Synthesis

The objective is to evolve the `kit` harness from a document-centric workflow orchestrator into a deterministic behavioral optimization system that continuously extracts, validates, scores, promotes, prunes, and versions reusable agent skills at global, domain, and project scopes. [S1][S2][S3] Affected users are autonomous coding agents and engineers using Kit workflows for implementation, validation, replay, and reflection loops. [S2] The selected direction uses replay-driven skill optimization instead of unconstrained memory accumulation, with skills represented as minimal executable policy modules rather than prose memories. [S1][S3] Definition of done: Kit persists candidate skills under `.kit/skills/candidates/`, validates candidates against replay/eval runs, promotes only score-improving skills into active scope directories, tracks rejected edits and scores, supports hierarchical skill resolution order, and exposes deterministic CLI commands for optimize/promote/reject/history/gc workflows. [S2][S3] The chosen architecture rejects vector-memory prompt accumulation and instead implements policy layering with lifecycle management, replay gates, and regression prevention. [S3]

### Source Map

- [S1] link: SkillOpt paper defining executive-driven self-evolving agent skill optimization using replay/evaluation loops and externalized behavioral skills; https://arxiv.org/abs/2605.23904
- [S2] link: Kit repository containing workflow orchestration, replay, verify, trace, and skill-mining primitives; https://github.com/jamesonstone/kit
- [S3] discussion: Hierarchical behavioral policy overlays were selected over flat memory systems, with global/domain/project/session scopes and lifecycle pruning; conversation identifier: `kit-skill-architecture-thread`
- [S4] note: Skills must remain structured, composable, scoreable, and bounded in size to prevent prompt-landfill degradation; conversation identifier: `kit-skill-constraints`
- [S5] discussion: Candidate skills require replay/eval promotion gates and rejected-edit persistence for anti-regression behavior; conversation identifier: `kit-skill-promotion-loop`

### Coding Agent Instructions

Implement hierarchical skill optimization infrastructure inside Kit using replay/eval validated behavioral policy modules instead of conversational memory accumulation. [S1][S3] Use the SkillOpt paper as the primary architectural reference for candidate generation, executive optimization loops, replay-based evaluation, and promotion/rejection semantics. [S1] Preserve existing workflow orchestration semantics while extending the current skill mining pipeline into a deterministic optimize/promote/reject lifecycle. [S2][S5] Tradeoffs: replay-driven validation increases execution cost but prevents prompt drift and behavioral regressions; structured policy modules reduce flexibility but preserve composability and deterministic resolution. [S3][S4]

1. Inspect repository structure and identify existing workflow, replay, verify, trace, reflection, and skill functionality by exact file path and exported symbol. [S2] Inspect `README.md`, `docs/`, `cmd/`, `internal/`, `pkg/`, and locate symbols implementing `skill mine`, replay execution, validation gates, workflow state transitions, and trace persistence. [S2]

2. Read and extract operational concepts from the SkillOpt paper before implementation. [S1] Capture exact mechanisms for candidate skill generation, optimization loops, evaluation gating, persistent skill memory, and rejection semantics. Map those concepts onto Kit runtime behavior.

3. Reconcile brainstorm architecture with actual repository behavior. [S2][S3] Document conflicts between proposed hierarchical skill overlays and current runtime execution order. Mark unresolved mismatches as `CONFLICT` and choose deterministic precedence: `session -> project -> domain -> personal -> global`. [S3]

4. Produce a complete implementation strategy grounded in repository reality. [S2] Define exact directories, symbols, interfaces, storage formats, score calculation logic, replay validation pipeline, promotion semantics, and garbage-collection behavior. [S3][S5]

5. Enumerate concrete file edits and additions. [S2] Include CLI command implementations for `kit skill optimize`, `kit skill score`, `kit skill promote`, `kit skill reject`, `kit skill history`, and `kit skill gc`. [S5] Include configuration files, persistence formats, migration steps, dependency changes, replay integration points, and exported symbols by exact path and identifier. [S2]

6. Define deterministic data model changes. [S3][S5] Add candidate-skill persistence under `.kit/skills/candidates/`, rejected-edit storage under `.kit/skills/rejected/`, active scoped skills under `.kit/skills/{global,domain,project}/`, and score tracking under `.kit/skills/scores.json`. Include lifecycle metadata: `version`, `score`, `uses`, `last_used`, `last_failure`, `superseded_by`, `ttl_days`, and `decay_rate`. [S3]

7. Define validation commands and expected outputs. [S2][S5] Include replay/eval execution commands, promotion verification commands, score diff assertions, regression checks, and garbage-collection validation commands with deterministic expected output strings or exit codes.

8. Define unit, integration, replay, and negative-path tests. [S2][S5] Include tests for failed promotions, replay regressions, stale-skill pruning, overlay precedence resolution, duplicate-skill merge behavior, malformed metadata rejection, and rejected-edit replay prevention.

9. State risks, assumptions, and open questions explicitly. [S3][S4] Include owner and mitigation for replay runtime growth, skill explosion, stale-skill accumulation, prompt bloat, score instability, and conflicting overlay resolution semantics.

10. Produce the final deliverable as a fully executable implementation document with exact file paths, symbols, commands, interfaces, acceptance checks, migration sequence, and repository-grounded assumptions only. [S2]

## RELATIONSHIPS

Canonical relationships are tracked in front matter.

- builds on: `0006-skill-mine-command`
- related to: `0025-v0-prompt-library`
- builds on: `0026-front-matter-integration`
- depends on: `0030-reference-graph-routing`
- depends on: `0031-executable-verification-harness`

## CODEBASE FINDINGS

1. Project constraints:
   - `docs/CONSTITUTION.md` defines Kit as a document-first harness. Markdown remains authoritative; `.kit/runs/` and `.kit/state.json` are generated local evidence/state surfaces.
   - `docs/CONSTITUTION.md` forbids hidden databases, hidden external state, secret storage, and direct management of agents. A skills layer must therefore use visible filesystem artifacts and should not depend on an external model API inside the Kit binary.
   - `docs/CONSTITUTION.md` also says Kit does not define understanding rubrics or scoring models today. `v1-skills-layer` would intentionally change that product boundary by adding skill scoring, so the spec must call out the constitutional delta explicitly.

2. Existing skills surface:
   - `internal/config/config.go` defines `Config.SkillsDir` with default `.agents/skills` and exposes `Config.SkillsPath(projectRoot string) string`.
   - `pkg/cli/skill.go` registers canonical `kit skill` and hidden deprecated `kit skills`; only `mine` exists today.
   - `pkg/cli/skill.go::runSkillMine` is prompt-only. It resolves a feature, requires `SPEC.md`, `PLAN.md`, and `TASKS.md`, builds a prompt, and writes no skill files itself.
   - `pkg/cli/skill_prompt.go::buildSkillMinePrompt` tells the agent to write canonical skills under `<skills_dir>/<feature-slug>/SKILL.md`, defaulting to `.agents/skills`, and duplicate the directory into `.claude/skills` as a Claude discovery mirror.
   - `pkg/cli/skill_prompt.go::skillAuditSteps` models stale-skill cleanup as prompt guidance only, with explicit human approval before deletion. There is no deterministic skill inventory, score, promotion gate, rejected-edit store, version history, or garbage collector.
   - `pkg/cli/skills_prompt.go::skillPromptSuffix` instructs agents to read feature front matter `skills` first and fallback to legacy `SPEC.md` `## SKILLS` only when front matter is absent.

3. CONFLICT: requested `.kit/skills` versus current `.agents/skills`.
   - User thesis requires candidate, rejected, score, and active scoped skills under `.kit/skills/...`.
   - Current repo docs and code treat `.agents/skills` as the repo-local canonical skills root and `.claude/skills` as a mirror.
   - Approved resolution: use a master skills document as the canonical lifecycle artifact instead of `.kit/skills/*` runtime state. `.kit/skills` may be used only for generated evidence, candidate staging, export cache, or temporary artifacts if the spec explicitly keeps it non-authoritative. `.agents/skills` remains the exported discovery surface for promoted project-scope skills, and `.claude/skills` remains optional compatibility mirroring only.
   - Candidate master document locations are `docs/SKILLS.md`, `docs/references/SKILLS.md`, or `docs/skills/README.md`. Current `docs/references/README.md` says `docs/references/*` is for durable repo-wide references broader than one feature, which makes `docs/references/SKILLS.md` the best repo-consistent default pending user approval.

4. Existing verification and replay surfaces:
   - `internal/verify/tasks.go::LoadTaskBundles` parses `TASKS.md` task detail fields, including `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK`, into `verify.TaskBundle`.
   - `internal/verify/tasks.go::ParseCommand` rejects shell syntax by default and requires explicit shell opt-in.
   - `internal/verify/execute.go::Run` captures `RunID`, `ParentRunID`, feature ref, task IDs, expected files, commands, results, status, start/end timestamps, and artifact directory.
   - `internal/verify/execute.go::ExecuteRun` supports pass/fail/dry-run/no-declared-checks states but does not compute semantic scores or aggregate a validation suite.
   - `pkg/cli/verify.go::runVerify` runs selected task commands and writes `.kit/runs/<run-id>/` artifacts through `internal/runstore` unless `--dry-run` or `--no-write` is used.
   - `pkg/cli/replay.go::runReplay` reruns the command set from one prior run and links the new run by `parent_run_id`; it does not replay a feature-level benchmark suite or compare candidate skill variants.
   - `pkg/cli/trace.go::runTrace` lists runs for a feature or displays compact run detail for a run ID. It is the closest existing history/provenance surface.
   - `internal/runstore/store.go::Write` writes `run.json`, bounded/redacted stdout/stderr artifacts, `summary.md`, and `.kit/runs/index.json`.

5. Existing eval and reflection surfaces:
   - `internal/eval/eval.go::Run` is a local harness regression suite for Kit behavior. It is not a general scoring engine, but its `Report`/`CaseResult` shape is a useful precedent for deterministic pass/fail suites.
   - `pkg/cli/eval.go::runEval` exposes `kit eval` with text and JSON output and exits non-zero on failed cases.
   - `pkg/cli/reflect.go::latestVerificationEvidenceStep` already reads latest runstore evidence and warns when verification evidence is missing. Skill promotion should use the same evidence culture but with stricter score-diff gates.

6. Existing generated state:
   - `internal/state/state.go::Generate` produces pointer-only `.kit/state.json` from feature docs and latest run evidence. It is non-authoritative and includes source fingerprints.
   - The skills layer should not make `.kit/state.json` or `.kit/skills/*` authoritative. It may add pointer-only active skill summaries later, but the durable source should remain the master skills document plus referenced feature artifacts and run evidence.

7. Existing front matter and reference graph behavior:
   - `internal/document/metadata.go::MetadataReference` defines canonical `references` fields: `id`, `name`, `type`, `target`, `selector_type`, `selector`, `relation`, `read_policy`, `used_for`, and `status`.
   - `internal/document/metadata.go::MetadataSkill` defines current feature-level skill metadata: `name`, `source`, `path`, `trigger`, and `required`.
   - `internal/document/metadata.go` validates reference relations, read policies, statuses, and selector types. This brainstorm's front matter uses those enums.
   - `internal/feature/map.go::BuildProjectMap` and `internal/feature/reference_resolver.go::resolveReference` render and resolve references for `kit map`. New skill references should use existing metadata rather than inventing a second reference graph.

8. Existing layered registry precedent:
   - `internal/promptlib/merge.go::Merge` resolves built-in, global, and local prompt sources with deterministic precedence and shadow metadata.
   - `pkg/cli/prompt.go::loadPromptLibrary` loads built-in prompts, global config, then local config, and delegates merging to `promptlib`.
   - This is not a skill runtime, but it is a repo-proven pattern for layered resolution, stable sorting, duplicate rejection, and override visibility.

9. SkillOpt operational concepts mapped to Kit:
   - Rollout evidence maps to `.kit/runs/<run-id>/run.json`, `summary.md`, and selected task/eval runs.
   - Optimizer reflection maps to a Kit-generated prompt or external agent step unless the spec explicitly allows Kit to call a model, which current project constraints do not.
   - Bounded add/delete/replace edits map to candidate entries in the master skills document plus optional non-authoritative staging artifacts.
   - Held-out validation maps to replay/eval suites that compare active baseline score to candidate score.
   - Rejected-edit buffer maps to rejected candidate entries in the master skills document, including content hash, reason, score, and replay evidence pointers.
   - Slow/meta update maps to periodic history compaction or score decay, not deploy-time prompt bloat.
   - Exported best skill maps to active skill entries in the master skills document and an `.agents/skills` discovery export.

10. Approved lifecycle split:
   - `kit reflect` is where recursive self-improvement is noticed and proposed, not where it is silently applied.
   - `reflect = observe, extract candidates, recommend`.
   - `skill = score, promote, reject, prune`.
   - Reflection completion should not require promoting a skill. It should include a non-blocking skill discovery section by default and route actionable candidate work to `kit skill optimize <feature>` and `kit skill promote <candidate-id>`.

11. Approved deterministic model boundary:
   - Kit should not call an optimizer model itself in v1.
   - Candidate authoring stays external or prompt-assisted.
   - `kit skill optimize <feature> --candidate <path> --suite <name>` should deterministically assemble replay/eval evidence, accept a candidate skill path, run score gates, persist candidate/rejected/promoted artifacts through the canonical document workflow, and record history.
   - Direct model-backed optimization is deferred to a future feature that explicitly updates the Constitution, configuration, and security model.

12. Current working tree note:
   - `git status --short` already shows modified `docs/PROJECT_PROGRESS_SUMMARY.md` and untracked `docs/notes/0032-v1-skills-layer/` plus `docs/specs/0032-v1-skills-layer/` before this brainstorm update. Treat these as pre-existing feature-scaffolding changes; do not revert them.

## AFFECTED FILES

1. Existing files likely affected by the future implementation:
   - `internal/config/config.go`: add skill-layer configuration while preserving existing `skills_dir` behavior and `.kit.yaml` compatibility.
   - `pkg/cli/skill.go`: extend `kit skill` with `optimize`, `score`, `promote`, `reject`, `history`, and `gc`; keep `mine` and hidden `skills` compatibility behavior intact.
   - `pkg/cli/skill_prompt.go`: update mining prompt to route candidates into the master skills document lifecycle, or leave prompt-only mining as legacy input to candidate creation.
   - `pkg/cli/reflect.go`: add a non-blocking skill discovery section so reflection can notice reusable operating patterns, plan divergences, repeated validation failures, and skill helped/hurt evidence without silently applying skill changes.
   - `pkg/cli/root_help.go`: place new skill lifecycle commands without exposing deprecated `skills`.
   - `internal/verify/tasks.go`: likely reuse `TaskBundle`, `Command`, and `LoadTaskBundles`; avoid changing existing parsing unless candidate eval suites need an additional typed command source.
   - `internal/verify/execute.go`: likely reuse `Run`, `RunStatus`, `CommandResult`, and `ExecuteRun`; avoid adding skill-specific fields here unless generic run metadata is needed.
   - `internal/runstore/store.go`: reuse run artifact persistence and latest-run lookup; consider adding generic provenance links from skill score records to run IDs rather than expanding runstore into a skill database.
   - `pkg/cli/verify.go`, `pkg/cli/replay.go`, `pkg/cli/trace.go`, `pkg/cli/eval.go`: integrate only through exported/internal helpers where possible; do not duplicate execution logic.
   - `internal/state/state.go`: optionally add pointer-only skill inventory references after the master skills document schema is defined.
   - `internal/document/metadata.go`: avoid expanding feature front matter into a lifecycle ledger; feature front matter can reference skill provenance, evidence run IDs, source feature, and produced/used/evaluated/changed relations.
   - `internal/feature/map.go` and `internal/feature/reference_resolver.go`: only affected if `kit map` should render the master skills document as a reference or include skill provenance edges.
   - `internal/templates/templates.go`: update generated docs or task templates only if new verification/skill fields are required in future feature docs.
   - `README.md`, `docs/CONSTITUTION.md`, `docs/agents/TOOLING.md`, `docs/references/README.md`, and generated instruction templates: update only for changed product contracts, especially master skills document authority versus `.agents/skills` discovery/export behavior.

2. New files/packages likely needed:
   - `docs/references/SKILLS.md`: recommended master skills document location, pending approval; owns active skills, candidates, rejected edits, scores, versions, evidence pointers, supersession, TTL/decay policy, and pruning decisions.
   - `internal/skilldoc/` or `internal/skillstore/`: parser/mutator for the master skills document plus exported skill bundles; naming should reflect that markdown is canonical, not a hidden store.
   - `internal/skillstore/types.go`: data structures such as `Skill`, `Candidate`, `ScoreRecord`, `HistoryEntry`, `Rejection`, `Scope`, and `Resolution`.
   - `internal/skillstore/store.go`: path resolution, careful document updates, exported bundle writes, validation, and migration helpers.
   - `internal/skillstore/resolve.go`: deterministic precedence `session -> project -> domain -> personal -> global`.
   - `internal/skillstore/score.go`: score calculation, baseline/candidate comparison, decay, TTL, and strict improvement rules.
   - `internal/skillstore/gc.go`: stale candidate/rejected/active cleanup with dry-run support.
   - `pkg/cli/skill_optimize.go`: candidate generation/orchestration command.
   - `pkg/cli/skill_score.go`: explicit score and score-diff command.
   - `pkg/cli/skill_promote.go`: promotion gate and active-scope write/export command.
   - `pkg/cli/skill_reject.go`: rejected-edit persistence command.
   - `pkg/cli/skill_history.go`: score/history/provenance inspection command.
   - `pkg/cli/skill_gc.go`: garbage collection command.
   - Tests beside each new package and CLI file, following existing stdlib testing patterns.

3. Proposed durable document layout:
   - `docs/references/SKILLS.md` or the user-approved equivalent master document.
   - Required sections likely include `SUMMARY`, `ACTIVE SKILLS`, `CANDIDATES`, `REJECTED CANDIDATES`, `SCORES`, `HISTORY`, `SUPERSESSION`, `TTL AND DECAY`, `GC DECISIONS`, and `EVIDENCE`.
   - Feature front matter should record lightweight skill provenance only: skill ID, relation, evidence run ID, source feature, and whether the feature produced, used, evaluated, or changed the skill.
   - `.agents/skills/<skill-id>/SKILL.md` remains exported discovery output for promoted project-scope skills, not the lifecycle source of truth.
   - `.kit/skills/*`, if kept at all, must be generated or temporary and must point back to `docs/references/SKILLS.md`.

## DEPENDENCIES

References are tracked in front matter. The implementation strategy depends most heavily on these current repo surfaces:

1. `0006-skill-mine-command`: defines the existing source-of-truth skills root, transfer bundle format, and prompt-only mining semantics.
2. `0031-executable-verification-harness`: provides `kit verify`, `.kit/runs`, `kit trace`, `kit replay`, `.kit/state.json`, and local `kit eval` primitives.
3. `0030-reference-graph-routing` and `0026-front-matter-integration`: constrain front matter references, relationships, and skills metadata.
4. `0025-v0-prompt-library`: provides a deterministic layered precedence precedent but should not be reused as the skill store itself.
5. SkillOpt arXiv v2 and project page: supply the external optimization loop model, especially bounded edits, held-out gates, rejected-edit feedback, and compact deploy-time artifact goals.

## QUESTIONS

1. Resolved: use a master skills document as the canonical skills-layer artifact, not `.kit/skills/*` runtime state.
   - `.kit/skills/*` must not become a hidden or competing ledger.
   - Feature front matter records lightweight skill provenance only.

2. Resolved: include all five precedence scopes in v1.
   - Precedence is `session -> project -> domain -> personal -> global`.
   - Remaining detail: choose storage/representation for session and personal scopes in the master document.

3. Resolved: Kit must not call an optimizer model itself in v1.
   - Candidate authoring stays external or prompt-assisted.
   - `kit skill optimize` deterministically assembles replay/eval evidence, accepts a candidate skill path, runs score gates, persists candidate/rejected/promoted outcomes through the canonical document workflow, and records history.
   - Direct model-backed optimization is deferred to a future feature with explicit Constitution, configuration, and security-model changes.

4. Resolved: score candidates against a named replay/eval suite.
   - Default v1 score is weighted pass rate.
   - Promotion requires strict score improvement, no required regression, and persisted score diff.
   - Default check weight is `1.0` until a future spec adds per-check weights.

5. Resolved: keep lifecycle/scoring metadata out of `SKILL.md` front matter.
   - `SKILL.md` remains compact and trigger-facing.
   - The master skills document owns `version`, `score`, `uses`, `last_used`, `last_failure`, `superseded_by`, `ttl_days`, `decay_rate`, promotion history, rejected candidates, and GC decisions.

6. Resolved: promotion exports to `.agents/skills` by default and keeps `.claude/skills` optional/configured.
   - `.agents/skills` is discovery/export output, not lifecycle authority.
   - `.claude/skills` remains compatibility mirroring only.

7. Resolved: `kit skill optimize` v1 accepts candidate input.
   - Required shape: `kit skill optimize <feature> --candidate <path> --suite <name>`.
   - Prompt-output mode should help author candidates from reflect/replay/eval evidence.

8. Resolved: rejected edits block blind duplicate promotion.
   - Rejection records should include candidate content hash, reason, validation suite, score, and evidence pointers.
   - A future or flagged retry path can be allowed when new evidence makes a previously rejected candidate relevant again.

9. Resolved: GC is dry-run by default.
   - Destructive deletion/pruning requires `--yes`.
   - Active best-scoring skills must not be pruned.

10. Resolved: the next `SPEC.md` should include a small Constitution update.
    - The update should define a bounded skill-score exception while preserving the broader no-hidden-state and document-first constraints.

11. Open: choose the master skills document path.
    - Recommended default: `docs/references/SKILLS.md` because `docs/references/README.md` defines that folder as durable repo-wide context broader than one feature.

12. Open: choose whether lifecycle commands mutate the master document directly.
    - Recommended default: direct, careful mutation for `promote`, `reject`, and `gc --yes`; dry-run/report/prompt-only modes for review and candidate authoring.

13. Open: choose the master document record format.
    - Recommended default: human-readable markdown sections plus fenced YAML record blocks for machine-updated skill/candidate/score/history entries.

14. Open: choose where full candidate skill content lives before promotion.
    - Recommended default: candidate content is copied into `docs/references/skills/candidates/<candidate-id>/SKILL.md` or equivalent durable docs-adjacent path, while the master document owns metadata and pointers.

15. Open: choose the first-class validation suite format.
    - Recommended default: define suites in the master skills document first, with each suite listing feature/task verify selectors, run IDs, eval cases, required flags, and optional weights.

## OPTIONS

1. Option A: Minimal extension of existing `kit skill mine`.
   - Behavior: keep `.agents/skills` canonical; add more prompt text for replay/eval validation; no `.kit/skills` lifecycle store.
   - Pros: smallest implementation and least conflict with current docs.
   - Cons: fails the user thesis because candidates, rejected edits, scores, promotion gates, scoped active layers, and GC remain non-deterministic.
   - Fit: not recommended.

2. Option B: master skills document with `.agents/skills` export.
   - Behavior: a durable markdown document owns active skills, candidates, rejected edits, scores, history, scopes, supersession, TTL/decay policy, and GC decisions; promoted project skills export into `.agents/skills` for current agent discovery.
   - Pros: satisfies the requested deterministic lifecycle while preserving Kit's document-first source-of-truth model and avoiding hidden state.
   - Cons: requires careful machine edits to markdown and a compatibility story for existing `skills_dir` semantics.
   - Fit: recommended default.

3. Option C: Reuse `internal/promptlib` as the skill layer.
   - Behavior: model skills like prompts with local/global/built-in precedence.
   - Pros: reuses a proven layered merge pattern.
   - Cons: prompt entries are free-form text, not scoreable lifecycle artifacts; promptlib has no replay gates, candidate store, rejection buffer, active scopes, or GC.
   - Fit: use as design precedent only, not as the implementation package.

4. Option D: Model-backed SkillOpt executor inside Kit.
   - Behavior: Kit directly invokes an optimizer model to generate bounded skill edits from rollouts.
   - Pros: closest to the paper's full executive optimization loop.
   - Cons: conflicts with current model-agnostic/no-external-state posture, introduces credentials/secrets questions, and creates a much larger security and dependency surface.
   - Fit: not recommended for v1 without an explicit constitutional/product change.

5. Option E: `.kit/skills` generated runtime ledger.
   - Behavior: keep scores/history/candidates in JSON under `.kit/skills`.
   - Pros: easier for commands to update atomically than markdown.
   - Cons: rejected by the approved clarification because it makes generated/local runtime state compete with the canonical document model.
   - Fit: not recommended.

## RECOMMENDED STRATEGY

1. Product boundary:
   - Treat v1 as deterministic skill lifecycle infrastructure, not full autonomous model training.
   - Keep candidate authoring external or prompt-assisted, but make validation, scoring, promotion, rejection, history, resolution, and GC deterministic inside Kit.
   - Add a non-blocking skill discovery section to `kit reflect`; reflection observes and proposes candidates but does not silently apply them.
   - Preserve `kit skill mine` as a prompt surface, and add lifecycle subcommands under the same `kit skill` root.

2. Storage:
   - Use a master skills document as the optimization source of truth.
   - Recommended default path is `docs/references/SKILLS.md`.
   - Keep deploy-time skill artifacts compact by keeping lifecycle metadata in the master document and exported `SKILL.md` content focused on trigger-facing procedural guidance.
   - Export promoted project-scope skills to `.agents/skills`; keep `.claude/skills` optional/configured.
   - Treat `.kit/skills/*`, if present, as generated cache, temporary staging, or evidence output that must point back to the master skills document.

3. Resolution:
   - Implement deterministic precedence exactly as requested: `session -> project -> domain -> personal -> global`.
   - Detect duplicate skill IDs across scopes and report which lower-precedence skills were shadowed.
   - Keep merge behavior explicit and inspectable, borrowing the shadow/override language from `internal/promptlib`.

4. Validation and scoring:
   - Reuse `internal/verify` command parsing/execution and `internal/runstore` artifact writing.
   - Define a skill validation suite in the master skills document as a deterministic list of existing run IDs, feature/task verify commands, and/or Kit eval cases.
   - Score baseline active skill and candidate skill against the same suite.
   - Promotion requires strictly higher candidate score, no required regression, valid metadata, non-duplicate candidate hash, and persisted score evidence.
   - Rejection persists the candidate hash, score, reason, and evidence pointers in the master skills document to prevent blind retry of harmful edits.

5. CLI:
   - `kit skill optimize <feature> --candidate <path> --suite <name>`: create/refresh candidate record in the master skills document, run validation, write score evidence, and print promotion/rejection recommendation.
   - `kit skill score <candidate-or-skill> --suite <name> --json`: compute or show score and baseline diff.
   - `kit skill promote <candidate-id> --scope <scope>`: require passing score gate, update active scoped skill in the master document, and export project-scope `SKILL.md` when applicable.
   - `kit skill reject <candidate-id> --reason <text>`: persist rejected edit metadata and reason in the master document.
   - `kit skill history [skill-id]`: render lifecycle events, scores, run IDs, and supersession chain.
   - `kit skill gc --dry-run`: list stale candidates, rejected edits, and superseded active skills from the master document; require `--yes` for pruning edits.

6. Validation commands to carry into SPEC/PLAN:
   - `go test ./internal/skillstore ./pkg/cli` or the final approved parser package path.
   - `go test ./internal/verify ./internal/runstore ./internal/state`
   - `go run ./cmd/kit skill optimize v1-skills-layer --candidate <fixture> --suite <fixture> --json`
   - Expected successful promotion output should include `Status: promoted`, `Score delta: +`, `Master document: docs/references/SKILLS.md`, and exported skill path when applicable.
   - Expected failed promotion output should exit non-zero and include `candidate did not improve held-out validation score`.
   - `go run ./cmd/kit skill gc --dry-run` should exit 0 and include `Dry run: no files deleted`.
   - `go run ./cmd/kit check v1-skills-layer`
   - `go run ./cmd/kit check --project`
   - `go test ./...`

7. Test categories:
   - Unit tests: master document parsing, record update preservation, metadata validation, score calculation, duplicate hash detection, precedence resolution, TTL/decay math, and malformed metadata rejection.
   - CLI tests: all new subcommands, JSON/text output, dry-run behavior, missing candidate, invalid scope, duplicate rejected edit, and promotion failure.
   - Integration tests: candidate -> score -> promote -> history -> gc lifecycle on temp project fixtures.
   - Replay tests: candidate fails when replay regresses, passes when score improves, and records parent run/evidence links.
   - Negative-path tests: stale rejected edit blocks blind retry, malformed `metadata.json` fails, duplicate active skill reports shadowing, and active best-scoring skill is not pruned.

## NEXT STEP

Resolve the open questions above, update this brainstorm with accepted defaults or overrides, then move to `kit spec v1-skills-layer`. Do not implement code during this phase.
