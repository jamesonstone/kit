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
- CODEBASE FINDINGS
- AFFECTED FILES
- DEPENDENCIES
- QUESTIONS
- OPTIONS
- RECOMMENDED STRATEGY
- NEXT STEP

Rules:

- keep file paths concrete whenever possible
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
- `SUMMARY` ≤ 120 characters
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
```

`agents` controls agent-specific files only. Repo-wide Copilot instructions always scaffold to `.github/copilot-instructions.md`.

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

---

### 8.2 Feature Lifecycle

#### `kit brainstorm [feature]`

- prompt interactively for a feature name and short issue or feature thesis
- create or reuse the feature directory (uses `0001-feat-name` format)
- create `BRAINSTORM.md` as the first artifact in the feature directory
- output a planning-only `/plan` prompt for a coding agent
- require the agent to keep `BRAINSTORM.md` `## DEPENDENCIES` current with the supporting inputs used during the research phase
- require the agent to populate every `BRAINSTORM.md` section and replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
- require the agent to use numbered lists, ask clarifying questions in batches of up to 10, include a recommended default/proposed solution/assumption for every question, accept `yes` / `y` as full-batch approval and `yes 3, 4, 5` / `y 3, 4, 5` as numbered approval, support `no` / `n` overrides, state uncertainties, and output percentage-understanding progress after each batch
- require the agent to continue until the configured understanding threshold is reached and the specification is precise enough for a correct, production-quality solution before writing implementation artifacts

---

#### `kit spec <feature>`

- scaffold `SPEC.md` template for manual editing
- template includes section headers with placeholder comments (e.g., `<!-- TODO: describe the problem -->`)
- template includes `## SKILLS` and `## DEPENDENCIES` tables for newly generated specs
- prompt instructions require every `SPEC.md` section to be populated; if a section has no additional detail, replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`
- create feature directory if missing (uses `0001-feat-name` format)
- update `docs/PROJECT_PROGRESS_SUMMARY.md`

---

#### `kit plan <feature>`

- scaffold `PLAN.md` template for manual editing
- scaffold includes a `## DEPENDENCIES` table for implementation-strategy inputs
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

### 8.6 Context Summarization

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

### 8.7 Reflection

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

### 8.8 Agent Handoff

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
