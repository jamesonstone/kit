# Kit — Core Specification

## 1. Purpose

Kit is a general-purpose harness for disciplined thought work.

Its goal is to give teams the right amount of structure for the work in front
of them, using open standards and universally portable documents. Its deepest
engine is a document-first, spec-driven workflow, but the harness also supports
ad hoc work, catch-up, handoff, summarization, review, and orchestration.

The current command surface is packaged around repository and software work,
but the underlying concepts generalize to research, strategy, operations,
writing, policy, and other fields that benefit from explicit constraints,
exploration, specification, planning, execution, and reflection.

Kit intentionally avoids agent-specific tooling. Instead, it centralizes canonical documents and scaffolds lightweight pointer files for different coding agents and environments.

All canonical markdown files use **FULL CAPITALIZATION** (e.g., `CONSTITUTION.md`, `SPEC.md`) and snake_case for file naming (i.e. `PROJECT_PROGRESS_SUMMARY.md`). Use kebab-case for directories (e.g., `0001-feat-name`). Feature directories use numeric prefix + slug naming and may optionally begin with a `BRAINSTORM.md` research artifact.

---

## 2. Design Principles

- harness-first, workflow-second
- documents are the source of truth
- spec-driven workflow is a first-class engine, not the whole product
- ad hoc work should remain lightweight but verified
- software-oriented defaults should not narrow the conceptual scope of the harness
- markdown + yaml only
- no agent lock-in
- minimal magic, explicit state
- opinionated defaults, configurable escapes
- tooling should disappear once documents are correct

---

## 3. Artifact Model (Non-Negotiable)

Kit exposes the following structured artifact pipeline as one of its core
engines. The names are software-friendly, but the pattern is general across
domains:

**Project Initialization** (run once, update as needed):

1. **Constitution** — strategy, patterns, long-term vision (kept updated with priors)

**Optional Research Step**:

1. **Brainstorm** — codebase-aware research, affected files, options, and recommended strategy

**Core Development Loop**:

1. **Specification** — what is being built and why
2. **Plan** — how it will be built
3. **Tasks** — executable work units
4. **Implementation** — execution begins only after the implementation readiness gate passes
5. **Reflection** — verify correctness, refine understanding (loops back to Specification)

Kit's responsibility is to provide disciplined harnessing around planning,
state, verification, and transfer. In the structured path, its responsibility
ends once tasks are clear and validated.

---

## 4. Canonical Document Locations

### 4.1 Global

- `docs/CONSTITUTION.md`
  - single per repository
  - defines invariant rules and constraints

- `docs/PROJECT_PROGRESS_SUMMARY.md`
  - lives alongside `CONSTITUTION.md` to indicate top-level abstraction
  - generated and maintained by Kit
  - provides a high-level, fork-ready overview of the entire project
  - summarizes all features, their intent, implementation strategy, and current phase

### 4.2 Feature-Scoped

Each feature lives in its own directory:

```bash
docs/specs/<feature>/
  SPEC.md
  PLAN.md
  TASKS.md
  ANALYSIS.md (optional)
```

Defaults:

- `<feature>` uses the format `0001-feat-name`
- directory is created on first reference
- feature number is auto-assigned by scanning `docs/specs/` and incrementing

Slug validation:

- numeric prefix (auto-assigned)
- lowercase only
- kebab-case
- max 5 words

---

## 5. Repository Instruction Files

Kit scaffolds, but does not own, repository instruction files.

Examples:

- `AGENTS.md`
- `CLAUDE.md`
- `.github/copilot-instructions.md`

These files:

- contain links/paths to canonical documents
- summarize the active workflow contract and repository standards
- stay aligned with the canonical docs
- support multiple tools without changing the artifact pipeline

If canonical paths change, Kit can refresh these files.

### Example: `AGENTS.md`

```markdown
# AGENTS.md

## Kit is the source of truth

- Constitution: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `SPEC.md` (requirements)
  - `PLAN.md` (implementation plan)
  - `TASKS.md` (executable task list)
  - `ANALYSIS.md` (optional, analysis scratchpad)

## Workflow contract

- Specs drive code. Code serves specs.
- Optional research starts with `BRAINSTORM.md` when present.
- For any change:
  1. classify the work as spec-driven or ad hoc
  2. locate the relevant feature directory in `docs/specs/<feature>/`
  3. read `BRAINSTORM.md` when present, then `SPEC.md` → `PLAN.md` → `TASKS.md`
  4. implement tasks in order
  5. verify (tests / validation steps from plan)
  6. if reality diverges, update `SPEC.md` / `PLAN.md` / `TASKS.md` first, then code

## Multi-feature rule

- Never mix features in one `docs/specs/<feature>/` directory.
- If work spans features, update each feature's docs separately.
```

---

## 6. Document Structure (Canonical Templates)

All documents prioritize:

- density over prose
- clarity over style
- concepts over code

Code snippets are included **only when unavoidable**.

### 6.1 `CONSTITUTION.md`

Required sections:

- PRINCIPLES
- CONSTRAINTS
- NON-GOALS
- DEFINITIONS

Purpose:

- establish invariant rules
- prevent scope drift

---

### 6.2 `BRAINSTORM.md` (Optional)

Required sections for newly generated brainstorm docs and brainstorm docs touched by current workflow commands:

- SUMMARY
- USER THESIS
- RELATIONSHIPS
- CODEBASE FINDINGS
- AFFECTED FILES
- DEPENDENCIES
- QUESTIONS
- OPTIONS
- RECOMMENDED STRATEGY
- NEXT STEP

Rules:

- keep file paths concrete whenever possible
- `## RELATIONSHIPS` must be either `none` or one bullet per explicit cross-feature relationship using `builds on: <feature>`, `depends on: <feature>`, or `related to: <feature>`
- inline-code-wrapped targets after `builds on:`, `depends on:`, or `related to:` are valid, but Kit must normalize them back to the canonical feature ID
- relationship targets must use canonical feature directory identifiers such as `0007-catchup-command`
- keep supporting inputs in a dependency table with columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`
- `Status` should distinguish `active`, `optional`, and `stale` inputs
- when a dependency is a Figma or MCP-driven design source, record the exact URL or file/node reference in `Location`

---

### 6.3 `SPEC.md`

Required sections for newly generated specs and specs touched by current workflow commands:

- SUMMARY
- PROBLEM
- GOALS
- NON-GOALS
- USERS
- SKILLS
- RELATIONSHIPS
- DEPENDENCIES
- REQUIREMENTS
- ACCEPTANCE
- EDGE-CASES
- OPEN-QUESTIONS

Rules:

- one sentence where possible
- no implementation detail
- no speculative language
- keep `## SKILLS` focused on execution-time agent skills
- `## RELATIONSHIPS` must be either `none` or one bullet per explicit cross-feature relationship using `builds on: <feature>`, `depends on: <feature>`, or `related to: <feature>`
- inline-code-wrapped targets after `builds on:`, `depends on:`, or `related to:` are valid, but Kit must normalize them back to the canonical feature ID
- relationship targets must use canonical feature directory identifiers such as `0007-catchup-command`
- keep broader supporting inputs in `## DEPENDENCIES`

---

### 6.4 `PLAN.md`

Required sections for newly generated plans and plans touched by current workflow commands:

- SUMMARY
- APPROACH
- COMPONENTS
- DATA
- INTERFACES
- DEPENDENCIES
- RISKS
- TESTING

Rules:

- explain strategy, not code
- name decisions explicitly
- defer code unless essential
- keep implementation-strategy dependencies in a dependency table with columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`

---

### 6.5 `TASKS.md`

Top-level table (required):

| ID  | TASK | STATUS | OWNER | DEPENDENCIES |
| --- | ---- | ------ | ----- | ------------ |

Required sections:

- TASKS
- DEPENDENCIES
- NOTES

Rules:

- tasks are atomic
- tasks map to plan items
- status reflects real progress

---

### 6.6 `PROJECT_PROGRESS_SUMMARY.md`

Purpose:

- single high-level briefing for the entire project
- sufficient context to onboard or fork
- primary context input for any coding agent

#### Required Top Section (Always First)

**FEATURE PROGRESS TABLE**

| ID  | FEATURE | PATH | PHASE | PAUSED | CREATED | SUMMARY |
| --- | ------- | ---- | ----- | ------ | ------- | ------- |

Rules:

- one row per feature
- sorted by ID ascending
- `PHASE` ∈ `brainstorm | spec | plan | tasks | implement | reflect | complete`
- `PAUSED` ∈ `yes | no`
- `SUMMARY` should prefer the concise `SPEC.md` `SUMMARY` section and only fall back to `PROBLEM` or brainstorm summary when needed
- `SUMMARY` must preserve the full intended meaning without truncation
- `SUMMARY` should normalize whitespace so it stays readable in a single markdown table row
- table is the authoritative project state

#### Required Sections (In Order)

- PROJECT INTENT
- GLOBAL CONSTRAINTS
- FEATURE SUMMARIES
- CROSS-FEATURE NOTES (optional)
- LAST UPDATED

#### Feature Summary Template

For each feature (order must match table):

- STATUS
- INTENT
- APPROACH
- OPEN ITEMS
- POINTERS

Rules:

- no code blocks unless unavoidable
- no duplicated specifications
- every claim must map to a feature document

---

### 6.7 `ANALYSIS.md` (Optional)

Purpose:

- optional scratchpad for the coding agent during analysis
- tracks understanding progression
- surfaces ambiguities, questions, and clarifications

Suggested sections:

- UNDERSTANDING (percentage at top)
- QUESTIONS (open questions for the user/team)
- RESEARCH (technical investigation notes)
- CLARIFICATIONS (resolved questions with answers)
- ASSUMPTIONS (documented assumptions made)
- RISKS (identified risks or concerns)

Rules:

- manually created when needed
- not managed by Kit CLI
- serves as working memory for iterative analysis

---

## 7. Configuration:

`.kit.yaml` lives at the project root and defines defaults.

All Kit commands traverse upward to find `.kit.yaml` and use its location as the project root. This allows commands to run from any subdirectory.

### 7.1 Core Fields

```yaml
goal_percentage: 95
specs_dir: docs/specs
skills_dir: .agents/skills
constitution_path: docs/CONSTITUTION.md
allow_out_of_order: false # if true, kit plan/tasks create missing prerequisites
feature_state:
  0001-feat-name:
    paused: false
```

### 7.2 Agents

```yaml
agents:
  - AGENTS.md
  - CLAUDE.md
  - .github/copilot-instructions.md
```

`agents` lists the repository instruction files Kit keeps aligned by default.

### 7.3 Feature Naming

```plaintext
feature_naming:
  numeric_width: 4
  separator: "-"
```

Feature names always use numeric prefix + slug format (e.g., `0001-feat-name`).

CLI flags always override `.kit.yaml`.

---

## 8. CLI Commands

### 8.1 Initialization

#### `kit init`

- create `.kit.yaml` if missing
- create `docs/CONSTITUTION.md` if missing
- scaffold configured agent instruction files and `.github/copilot-instructions.md`
- if files exist, attempt to merge (preserve existing content, add missing sections)
- output a prepared prompt for drafting `docs/CONSTITUTION.md`
- by default, copy that prompt to the clipboard instead of printing the prompt body
- make the first visible next step: paste the copied prompt into the agent to draft `docs/CONSTITUTION.md`
- support `--output-only` to print the raw prompt to stdout instead of copying it
- support `--copy` to also copy the prompt when `--output-only` is set

---

### 8.2 Feature Lifecycle

#### `kit brainstorm [feature]`

- prompt interactively for a feature name and short issue or feature thesis
- create or reuse the feature directory (uses `0001-feat-name` format)
- when Git common-dir state is available, reserve the next feature number from the shared clone-local allocator before creating the directory
- create `BRAINSTORM.md` as the first artifact in the feature directory
- output a planning-only `/plan` prompt for a coding agent
- support `--backlog` to capture a deferred brainstorm item without outputting a
  planning prompt
- keep `--pickup` as a compatibility path for resuming a deferred backlog item
  while teaching `kit resume <feature>` or `kit backlog --pickup <feature>` as
  the canonical resume flows
- require the agent to use indices first for prior work discovery: `kit map
  <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md`
- require prior feature docs to be conditional reads gated by explicit
  relevance: shared interfaces or contracts, overlapping files or modules,
  migrations or data shape, acceptance criteria, or explicit relationship or
  dependency links
- require the agent to inspect at most 5 prior feature directories before
  narrowing further or asking a clarifying question
- require the agent to extract only concrete decision-shaping facts from prior
  work instead of replaying full historical docs into chat or active artifacts
- require the agent to keep `BRAINSTORM.md` `## DEPENDENCIES` current with the supporting inputs used during the research phase
- require the agent to populate every `BRAINSTORM.md` section and replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
- require the agent to use numbered lists, ask clarifying questions in batches of up to 10, include a recommended default/proposed solution/assumption for every question, accept `yes` / `y` as full-batch approval and `yes 3, 4, 5` / `y 3, 4, 5` as numbered approval, support `no` / `n` overrides, state uncertainties, and output percentage-understanding progress after each batch
- require the agent to continue until the configured understanding threshold is reached and the specification is precise enough for a correct, production-quality solution before writing implementation artifacts

---

#### `kit backlog [feature]`

- list deferred backlog items captured as paused brainstorm-phase features
- render a concise markdown table with `feature` and `description`
- use `BRAINSTORM.md` `SUMMARY` for the description when available and fall
  back to `USER THESIS`
- support `--pickup` to clear paused state for a backlog item and output the
  brainstorm planning prompt
- keep `kit backlog --pickup` as the backlog-specific shortcut while
  `kit resume` becomes the canonical general resume command
- without a feature argument under `--pickup`, show an interactive selector of
  backlog items
- do not create a new backlog-specific markdown artifact

---

#### `kit resume [feature]`

- act as the canonical command for resuming work
- when the target is a backlog item, reuse backlog pickup behavior and output
  the brainstorm planning prompt
- when the target is not a backlog item, reuse the catch-up prompt behavior
- without a feature argument, show a mixed selector ordered as:
  - paused non-backlog features
  - active in-flight feature
  - backlog items labeled as backlog
- support the shared clipboard-first prompt output contract

---

#### `kit spec <feature>`

- scaffold `SPEC.md` template for manual editing
- template includes section headers with placeholder comments (e.g., `<!-- TODO: describe the problem -->`)
- template includes `## SKILLS` and `## DEPENDENCIES` tables for newly generated specs
- prompt instructions require indices-first prior work discovery through
  `kit map <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md`
- prompt instructions require prior feature docs to stay conditional and
  relevance-gated before widening codebase discovery
- prompt instructions require extracted prior-work facts to stay concise and
  decision-shaping rather than historical replay
- prompt instructions require every `SPEC.md` section to be populated; if a section has no additional detail, replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`
- create feature directory if missing (uses `0001-feat-name` format)
- when Git common-dir state is available, reserve the next feature number from the shared clone-local allocator before creating the directory
- update `docs/PROJECT_PROGRESS_SUMMARY.md`

---

#### `kit plan <feature>`

- scaffold `PLAN.md` template for manual editing
- scaffold includes a `## DEPENDENCIES` table for implementation-strategy inputs
- prompt instructions require indices-first prior work discovery through
  `kit map <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md`
- prompt instructions require prior feature docs to stay conditional and
  relevance-gated so only the docs that materially shape the plan are loaded
- prompt instructions require extracted prior-work facts to stay concise and
  decision-shaping rather than historical replay
- prompt instructions require every `PLAN.md` section to be populated; if a section has no additional detail, replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`
- plan items link to spec items using `[SPEC-01]` syntax
- update `PROJECT_PROGRESS_SUMMARY.md`

Prerequisites:

- `SPEC.md` must exist (error otherwise)

Flags:

- `--force` — create empty `SPEC.md` with headers if missing

---

#### `kit tasks <feature>`

- scaffold `TASKS.md` template for manual editing
- prompt instructions require every `TASKS.md` section to be populated; if a section has no additional detail, replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`
- tasks link to plan items using `[PLAN-01]` syntax
- update `PROJECT_PROGRESS_SUMMARY.md`

Prerequisites:

- `PLAN.md` must exist (error otherwise)

Flags:

- `--force` — create empty `SPEC.md` and `PLAN.md` with headers if missing

---

#### `kit implement [feature]`

- output implementation context for a coding agent
- begin with an implementation readiness gate before any code execution instructions
- require an adversarial preflight across `CONSTITUTION.md`, optional `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md`
- require the coding agent to use indices-first prior work discovery through
  `kit map <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md`
- require prior feature docs to stay conditional and relevance-gated so only
  the docs that materially shape the implementation or refactor surface are
  loaded
- require extracted prior-work facts to stay concise and decision-shaping
  rather than historical replay
- require the coding agent to challenge contradictions, ambiguity, hidden assumptions, missing edge cases, missing task coverage, and scope creep before coding
- if the readiness gate fails, require the agent to update `SPEC.md`, `PLAN.md`, and/or `TASKS.md` first, refresh `PROJECT_PROGRESS_SUMMARY.md` when needed, then rerun the gate
- only after the readiness gate passes should the agent begin with the first incomplete task in `TASKS.md`
- if the target feature is paused, clear the paused flag before continuing

---

#### `kit pause [feature]`

- resolve the target feature using existing feature reference rules
- without a feature argument, show an interactive selector of non-complete features
- persist paused state in `.kit.yaml` without changing the feature's underlying phase
- reject complete features with an actionable error
- update `PROJECT_PROGRESS_SUMMARY.md`

---

#### `kit remove [feature]`

- resolve the target feature using existing feature reference rules
- without a feature argument, show an interactive selector of existing features
- require explicit confirmation before deletion unless `--yes` is set
- delete the feature directory and all files under it
- remove any persisted lifecycle state for the deleted feature from `.kit.yaml`
- update `PROJECT_PROGRESS_SUMMARY.md`

---

#### `kit map [feature]`

- render a read-only terminal graph of the canonical document hierarchy and current project state
- without a feature argument, show global docs plus every feature directory, current phase, paused state, canonical docs, and explicit relationship edges
- order project-wide feature rendering by `builds on` and `depends on` relationships when those edges provide a usable dependency order
- with a feature argument, show a feature-scoped view plus incoming or outgoing relationships that touch that feature
- derive relationship edges from explicit `## RELATIONSHIPS` sections in `BRAINSTORM.md` and `SPEC.md`
- normalize harmless inline-code formatting around relationship targets before validating or rendering them
- if a relationship line is malformed, keep the valid map output and surface the skipped line as a warning instead of failing the read-only command
- when writing to a terminal, map output may color labels and state markers for scanability without changing non-TTY output
- do not create another persisted markdown graph document

---

#### `kit status`

- default text output remains focused on the active feature only
- show current summary, phase, paused state, file presence, progress, and next
  recommended action for the active feature
- keep the existing default `--json` payload shape for the active-feature view
- support `--all` as the explicit fleet overview mode
- `--all` text output shows every feature in a terminal-friendly fixed-width
  lifecycle matrix with paused or backlog state and available task progress
- when writing to a terminal, status views may color lifecycle markers and
  state labels for scanability without changing non-TTY output
- `--all --json` uses a dedicated all-features payload and does not replace the
  default `--json` contract

---

### 8.3 Roll-Up

#### `kit rollup`

Purpose:

- analyze all feature specifications under `docs/specs/`
- generate or update `PROJECT_PROGRESS_SUMMARY.md`

Behavior:

- summarize each feature’s intent, approach, and implementation state
- include a table of all features with:
  - feature name
  - directory path
  - short summary
  - current artifact phase (`brainstorm | spec | plan | tasks | implement | reflect | complete`)
  - paused state
  - spec creation date
- remain callable as a maintenance command, not a primary workflow step taught
  to new users

`PROJECT_PROGRESS_SUMMARY.md` is intended to be:

- a high-level briefing document
- sufficient to onboard or fork the project
- safe to hand to any coding agent as primary context

This command is also executed automatically as the final stage of feature creation and refinement.

---

### 8.4 Verification

#### `kit check <feature>`

Validates:

- required documents exist
- required sections present and populated
- traceability between spec → plan → tasks
- no unresolved placeholders

Flags:

- `--all` — validate all features in `docs/specs/`

Fails fast with explicit errors. Errors suggest fixes (e.g., "SPEC.md missing. Run `kit spec <feature>` first or use `--force`").

---

### 8.5 Agent Scaffolding

#### `kit scaffold-agents`

- create missing repository instruction files
- overwrite existing repository instruction files only when `--force` is set
- prompt for confirmation before `--force` overwrites existing instruction files
- support `--yes` / `-y` to skip the overwrite confirmation prompt when `--force` is used
- support `--append-only` to merge missing Kit-managed sections without overwriting matched existing content
- scaffold `.github/copilot-instructions.md` alongside configured agent files
- support the alias `kit scaffold-agent`
- without targeted flags, scaffold configured agent files plus `.github/copilot-instructions.md`
- `--agentsmd` scaffolds only `AGENTS.md`
- `--claude` scaffolds only `CLAUDE.md`
- `--copilot` scaffolds only `.github/copilot-instructions.md`
- allow combining targeted flags to scaffold multiple specific built-in files in one run
- in default mode, suggest `--append-only` and `--force` when existing instruction files are skipped

---

### 8.6 Documentation Reconciliation

#### `kit reconcile [feature]`

Purpose:

- audit Kit-managed docs against the current Kit document contract
- output a prompt for a coding agent to reconcile stale or missing documentation
- keep v1 scoped to documentation only

Behavior:

- without feature argument: audits the whole project by default
- with feature argument: audits the selected feature plus related rollup drift
- emits a short clean result when no reconciliation is needed
- emits a clipboard-first prompt when reconciliation findings exist

Findings:

- missing required docs or sections
- placeholder-only required sections
- malformed `SKILLS`, `DEPENDENCIES`, or `PROGRESS TABLE` tables
- task-ID drift across `PROGRESS TABLE`, `TASK LIST`, and `TASK DETAILS`
- stale `RELATIONSHIPS` targets
- stale `PROJECT_PROGRESS_SUMMARY.md` coverage
- repository instruction-file drift detectable through append-only planning

Verification:

- run `kit check --all` for project-wide reconciliation or `kit check <feature>` for feature-scoped reconciliation
- run the maintenance command `kit rollup` when reconciled changes affect
  `PROJECT_PROGRESS_SUMMARY.md`

---

### 8.7 Context Summarization

#### `kit summarize [feature]`

Purpose:

- output instructions for context window summarization
- focus on retaining facts necessary for strategy, implementation, and process
- use with coding agents: `/compact` (Warp), `/summarize` (Claude), etc.

Behavior:

- without feature argument: outputs generic best-practice instructions
- with feature argument: outputs instructions scoped to that feature's context

Fact Retention Principles:

- **KEEP**: decisions, file paths, APIs, configs, errors, dependencies, test criteria
- **DISCARD**: pleasantries, redundant explanations, speculative discussions, verbose errors

---

### 8.8 Reflection

#### `kit reflect [feature]`

Purpose:

- output instructions for reflecting on recent changes
- ensure 100% implementation correctness
- verify changes using git, lint, and tests

Behavior:

- without feature argument: outputs generic verification instructions
- with feature argument: outputs instructions scoped to that feature's context

Reflection Process:

1. analyze git state (staged and unstaged changes)
2. understand the delta and intent of each change
3. cross-reference with repository context and codebase
4. verify correctness checklist (compiles, no errors, edge cases handled)
5. run lint and tests, then fix ALL failures (including out-of-scope failures) before completion
6. do not run `coderabbit --prompt-only` unless the user explicitly asks for it or explicitly approves it first

---

### 8.9 Agent Handoff

#### `kit handoff [feature]`

Purpose:

- prepare the current coding agent session to reconcile docs before transfer
- minimize information loss when switching agents due to token limits or rate limiting
- enable seamless continuation across Warp, Claude, Copilot, Codex, etc.

Behavior:

- without feature argument: shows an interactive numbered selector of all features under `docs/specs/` plus `0` for no specific feature
- selecting `0`: outputs project-level context including:
  - documentation inventory and current development status across active non-paused features
  - instructions to reconcile stale docs before handoff
  - dependency-inventory verification for touched feature docs
- with feature argument: outputs feature-specific context including:
  - feature location and phase
  - required reading (`BRAINSTORM.md` when present, then `SPEC.md`, `PLAN.md`, `TASKS.md`)
  - instructions to refresh dependency tables in `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` when those docs exist
  - a final response contract for concise documentation sync and recent-context summary

Flags:

- `--copy` / `-c` — copy output to clipboard (pbcopy)

Use case: when you run out of tokens or hit rate limits, run `kit handoff`, let the current agent reconcile docs and dependency inventories, then transfer the final handoff summary.

---

## 9. Understanding Percentage Model

Understanding is determined entirely by the active coding agent.

Kit does not define a scoring rubric or weighting model.

Responsibilities:

- the agent evaluates completeness, clarity, and readiness
- the agent reports a single integer percentage (0–100)
- Kit only compares the reported value against the configured goal

Defaults:

- goal percentage comes from `.kit.yaml`
- CLI flags override configuration

---

## 10. Workflow Scope

- feature directories use the format `0001-feat-name`
- multiple features can be in progress simultaneously
- `BRAINSTORM.md` is optional but first-class when present
- `SPEC.md`, `PLAN.md`, and `TASKS.md` remain the binding execution artifacts

Kit does not:

- manage branches or PRs
- enforce git policies
- maintain any state beyond files (no database, no lock files)

---

## 11. Non-Goals

Kit explicitly does not:

- execute code
- manage agents directly
- maintain prompt registries
- invent new document formats
- define understanding rubrics
- replace CI/CD systems

---

## 12. Implementation Notes

- Kit is implemented in **Go**
- follows best practices for modern Go CLIs
- uses a single binary with subcommands
- favors explicit flags over hidden state

---

## 13. Success Criteria

Kit is successful if:

- documents remain readable without Kit
- agents can be swapped with zero document changes
- teams reach clarity faster with fewer reworks
- the CLI becomes unnecessary once understanding is achieved
