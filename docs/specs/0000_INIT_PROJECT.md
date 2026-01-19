# Kit — Core Specification

## 1. Purpose

Kit is a document-centered CLI for spec-driven development.

Its goal is to help teams reach a high-confidence understanding of a problem and its solution _before_ implementation, using open standards and universally portable documents.

Kit intentionally avoids agent-specific tooling. Instead, it centralizes canonical documents and scaffolds lightweight pointer files for different coding agents and environments.

All canonical markdown files use **FULL CAPITALIZATION** (e.g., `CONSTITUTION.md`, `SPEC.md`) and snake_case for file naming (i.e. `PROJECT_PROGRESS_SUMMARY.md`). Use kebab-case for directories (e.g., `0001-feat-name`). Use lowercase for git branches (e.g., `0001-feat-name`) and allow them to be customizable via a configuration file and/or CLI flags.

---

## 2. Design Principles

- documents are the source of truth
- markdown + yaml only
- no agent lock-in
- minimal magic, explicit state
- opinionated defaults, configurable escapes
- tooling should disappear once documents are correct

---

## 3. Artifact Model (Non-Negotiable)

Kit enforces the following artifact pipeline:

**Project Initialization** (run once, update as needed):

1. **Constitution** — strategy, patterns, long-term vision (kept updated with priors)

**Core Development Loop**:

2. **Specification** — what is being built and why
3. **Plan** — how it will be built
4. **Tasks** — executable work units
5. **Implementation** — execution outside Kit's core scope
6. **Reflection** — verify correctness, refine understanding (loops back to Specification)

Kit's responsibility ends once tasks are clear and validated.

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

```
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

## 5. Agent Portability via Pointer Files

Kit scaffolds, but does not own, agent-specific files.

Examples:

- `AGENTS.md`
- `CLAUDE.md`
- `WARP.md`

These files:

- contain links/paths to canonical documents
- include minimal agent-specific constraints
- never duplicate specifications
- define the workflow contract for that agent

If canonical paths change, Kit can update pointers.

#### Example: `WARP.md`

```markdown
# WARP.md

## Kit is the source of truth

- Constitution: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `SPEC.md` (requirements)
  - `PLAN.md` (implementation plan)
  - `TASKS.md` (executable task list)
  - `ANALYSIS.md` (optional, analysis scratchpad)

## Workflow contract

- Specs drive code. Code serves specs.
- For any change:
  1. locate the relevant feature directory in `docs/specs/<feature>/`
  2. read `SPEC.md` → `PLAN.md` → `TASKS.md`
  3. implement tasks in order
  4. verify (tests / validation steps from plan)
  5. if reality diverges, update `SPEC.md` / `PLAN.md` / `TASKS.md` first, then code

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

### 6.2 `SPEC.md`

Required sections:

- PROBLEM
- GOALS
- NON-GOALS
- USERS
- REQUIREMENTS
- ACCEPTANCE
- EDGE-CASES
- OPEN-QUESTIONS

Rules:

- one sentence where possible
- no implementation detail
- no speculative language

---

### 6.3 `PLAN.md`

Required sections:

- SUMMARY
- APPROACH
- COMPONENTS
- DATA
- INTERFACES
- RISKS
- TESTING

Rules:

- explain strategy, not code
- name decisions explicitly
- defer code unless essential

---

### 6.4 `TASKS.md`

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

### 6.5 `PROJECT_PROGRESS_SUMMARY.md`

Purpose:

- single high-level briefing for the entire project
- sufficient context to onboard or fork
- primary context input for any coding agent

#### Required Top Section (Always First)

**FEATURE PROGRESS TABLE**

| ID  | FEATURE | PATH | PHASE | CREATED | SUMMARY |
| --- | ------- | ---- | ----- | ------- | ------- |

Rules:

- one row per feature
- sorted by ID ascending
- `PHASE` ∈ `spec | plan | tasks | implement`
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

### 6.6 `ANALYSIS.md` (Optional)

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
constitution_path: docs/CONSTITUTION.md
allow_out_of_order: false # if true, kit plan/tasks create missing prerequisites
```

### 7.2 Agents

```
agents:
  - AGENTS.md
  - CLAUDE.md
  - WARP.md
```

### 7.3 Branching

```plaintext
branching:
  enabled: true
  base_branch: main
  name_template: "{numeric}-{slug}"
```

### 7.4 Feature Naming

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
- scaffold configured agent pointer files
- if files exist, attempt to merge (preserve existing content, add missing sections)

---

### 8.2 Feature Lifecycle

#### `kit spec <feature>`

- scaffold `SPEC.md` template for manual editing
- template includes section headers with placeholder comments (e.g., `<!-- TODO: describe the problem -->`)
- create feature directory if missing (uses `0001-feat-name` format)
- create a new git branch matching the directory name
- update `docs/PROJECT_PROGRESS_SUMMARY.md`

Flags:

- `--no-branch` — create directory but skip git branch creation

---

#### `kit plan <feature>`

- scaffold `PLAN.md` template for manual editing
- plan items link to spec items using `[SPEC-01]` syntax
- update `PROJECT_PROGRESS_SUMMARY.md`

Prerequisites:

- `SPEC.md` must exist (error otherwise)

Flags:

- `--force` — create empty `SPEC.md` with headers if missing

---

#### `kit tasks <feature>`

- scaffold `TASKS.md` template for manual editing
- tasks link to plan items using `[PLAN-01]` syntax
- update `PROJECT_PROGRESS_SUMMARY.md`

Prerequisites:

- `PLAN.md` must exist (error otherwise)

Flags:

- `--force` — create empty `SPEC.md` and `PLAN.md` with headers if missing

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
  - current artifact phase (spec | plan | tasks | implement)
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
- required sections present
- traceability between spec → plan → tasks
- no unresolved placeholders

Flags:

- `--all` — validate all features in `docs/specs/`

Fails fast with explicit errors. Errors suggest fixes (e.g., "SPEC.md missing. Run `kit spec <feature>` first or use `--force`").

---

### 8.5 Agent Scaffolding

#### `kit scaffold-agents`

- create missing agent pointer files
- update document links if paths changed

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
- verify changes using git and coderabbit

Behavior:

- without feature argument: outputs generic verification instructions
- with feature argument: outputs instructions scoped to that feature's context

Reflection Process:

1. analyze git state (staged and unstaged changes)
2. understand the delta and intent of each change
3. cross-reference with repository context and codebase
4. verify correctness checklist (compiles, no errors, edge cases handled)
5. run `coderabbit --prompt-only` and fix any issues

---

### 8.8 Agent Handoff

#### `kit handoff [feature]`

Purpose:

- output context for starting a fresh coding agent session
- minimize information loss when switching agents due to token limits or rate limiting
- enable seamless continuation across Warp, Claude, Copilot, Codex, etc.

Behavior:

- without feature argument: outputs project-level context including:
  - Kit workflow explanation
  - project structure
  - list of features with their phases
  - immediate next steps
- with feature argument: outputs feature-specific context including:
  - feature location and phase
  - required reading (SPEC.md, PLAN.md, TASKS.md)
  - phase-appropriate next actions

Flags:

- `--copy` / `-c` — copy output to clipboard (pbcopy)

Use case: when you run out of tokens or hit rate limits, run `kit handoff` and paste the output into a new agent session to continue with minimal context loss.

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

## 10. Git Workflow

- feature creation defaults to a new branch
- branch names use the format `0001-feat-name`
- feature directory names match the branch name exactly
- multiple features can be in progress simultaneously (different branches, different directories)

Kit does not:

- merge branches
- manage PRs
- enforce git policies beyond branch creation
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
