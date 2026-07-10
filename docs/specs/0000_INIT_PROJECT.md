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

## 3. V2 Artifact Model (Non-Negotiable)

Kit exposes the v2 single-`SPEC.md` workflow as its primary structured engine.
The names are software-friendly, but the pattern is general across domains:

**Project Initialization** (run once, update as needed):

1. **Constitution** — strategy, patterns, long-term vision (kept updated with priors)
2. **Project refresh prompt** — `kit project refresh` asks an agent to re-analyze the maturing repository and update durable project-level docs without rerunning init

**Optional Research Material**:

1. **Feature notes/reference material** — supporting artifacts, screenshots, research, constraints, prior context, draft responses, and local-only conversation history under tracked/ignored sections

**V2 Feature Workflow**:

1. **`kit spec <feature>`** — emits the v2 supervisor prompt and seeds the clarification gate
2. **`SPEC.md`** — single durable feature artifact carrying phase and clarification front matter plus thesis, context, clarifications, requirements, assumptions, acceptance criteria, implementation plan, task checklist, validation map, reflection notes, documentation updates, delivery decision, and evidence
3. **Implementation** — begins only after objective readiness gates pass
4. **Validation and reflection** — prove every acceptance criterion, update docs, record evidence, and route gaps back to implementation before delivery

Kit's responsibility is to provide disciplined harnessing around planning,
state, verification, and transfer. In the structured path, v2 keeps durable
workflow state in `SPEC.md`; legacy staged artifacts remain readable historical
context when present.

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
  SPEC.md                 # v2 durable feature workflow artifact
  BRAINSTORM.md (optional legacy staged research)
  PLAN.md       (optional legacy staged plan)
  TASKS.md      (optional legacy staged task list)
  ANALYSIS.md (optional)
```

Defaults:

- `<feature>` uses the format `0001-feat-name`
- directory is created on first reference
- feature number is reserved by Kit's allocator, using repo-shared Git common-dir state when available and falling back to local `docs/specs/` inspection

Slug validation:

- numeric prefix (allocator-reserved)
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
- support multiple tools without changing the v2 `SPEC.md` workflow contract

If canonical paths change, Kit can refresh these files.

### Example: `AGENTS.md`

```markdown
# AGENTS.md

## Kit is the source of truth

- Constitution: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `SPEC.md` (v2 workflow artifact)
  - `BRAINSTORM.md`, `PLAN.md`, `TASKS.md` (optional legacy staged artifacts)
  - `ANALYSIS.md` (optional, analysis scratchpad)

## Workflow contract

- Specs drive code. Code serves specs.
- V2 feature work starts with `kit spec <feature>`.
- `SPEC.md` carries requirements, plan, tasks, validation, reflection, delivery, and evidence.
- Legacy staged artifacts are historical context unless the user explicitly uses a legacy staged command.
- For any change:
  1. classify the work as spec-driven or ad hoc
  2. locate the relevant feature directory in `docs/specs/<feature>/`
  3. read `SPEC.md` first for v2 feature work
  4. use legacy staged docs only when they materially affect the current decision
  5. implement from the `SPEC.md` task checklist after readiness gates pass
  6. verify against the `SPEC.md` validation map and record evidence there
  7. if reality diverges, update `SPEC.md` first, then continue

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

### 6.2 `BRAINSTORM.md` (Legacy Staged Optional)

Required sections for legacy staged brainstorm docs and brainstorm docs touched
by legacy staged workflow commands:

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
- keep supporting inputs in front matter `references` with `name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status`
- `status` should distinguish `active`, `optional`, and `stale` inputs
- durable rulesets live under `docs/references/rules/<slug>.md` and must be linked through front matter `references` rather than copied into always-loaded instruction files
- when a reference is a Figma or MCP-driven design source, record the exact URL or file/node reference in `target` and use stable selectors when needed

---

### 6.3 `SPEC.md`

Required fixed v2 section order for newly generated specs and specs touched by
the v2 workflow:

- THESIS
- CONTEXT
- CLARIFICATIONS
- REQUIREMENTS
- ASSUMPTIONS
- ACCEPTANCE CRITERIA
- IMPLEMENTATION PLAN
- TASK CHECKLIST
- VALIDATION MAP
- REFLECTION NOTES
- DOCUMENTATION UPDATES
- DELIVERY DECISION
- EVIDENCE

Required front matter:

- `workflow_version: 2`
- `phase: clarify | ready | implement | validate | reflect | deliver | complete | blocked`
- `clarification.status: open | ready | blocked`
- `clarification.confidence: 0..100`
- `clarification.unresolved_questions: 0 or greater`

Rules:

- `SPEC.md` is the single durable v2 feature artifact
- keep thesis, context, clarifications, requirements, plan, tasks, validation, reflection, documentation updates, delivery decision, and evidence in this file
- acceptance criteria must be binary-verifiable and mapped 1:1 to validation
- the task checklist must be concise, durable, and mapped to lanes, acceptance criteria, status, and evidence
- implementation must not begin until `clarification.status` is `ready`, confidence meets `goal_percentage`, unresolved questions are zero, assumptions are accepted or removed, validation is mapped, touched files are predicted, delivery lane is resolved, and rollback is known
- validation evidence should be summarized inline and may reference detailed logs under `.kit/runs/...`
- legacy staged `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files remain readable historical context but are not the default v2 workflow contract
- `## RELATIONSHIPS` must be either `none` or one bullet per explicit cross-feature relationship using `builds on: <feature>`, `depends on: <feature>`, or `related to: <feature>`
- inline-code-wrapped targets after `builds on:`, `depends on:`, or `related to:` are valid, but Kit must normalize them back to the canonical feature ID
- relationship targets must use canonical feature directory identifiers such as `0007-catchup-command`
- keep broader supporting inputs in front matter `references` when available

---

### 6.4 `PLAN.md` (Legacy Staged Optional)

Required sections for legacy staged plans and plans touched by legacy staged
workflow commands:

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
- keep implementation-strategy inputs in front matter `references` with exact targets, stable selectors, relations, read policies, and status values
- link applicable durable rulesets with `type: ruleset` and `target: docs/references/rules/<slug>.md`

---

### 6.5 `TASKS.md` (Legacy Staged Optional)

Top-level table (required):

| ID  | TASK | STATUS | OWNER | DEPENDENCIES |
| --- | ---- | ------ | ----- | ------------ |

Required sections:

- TASKS
- DEPENDENCIES
- NOTES

Rules:

- tasks are atomic
- tasks map to legacy plan items
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

- one row per live feature plus one row per removed feature tombstone
- sorted by ID ascending
- `PHASE` ∈ `clarify | ready | implement | validate | reflect | deliver | complete | blocked | removed`, with legacy staged states (`brainstorm | spec | plan | tasks`) allowed for historical projects
- `PAUSED` ∈ `yes | no`
- `SUMMARY` should prefer the concise `SPEC.md` `SUMMARY` section and only fall back to `PROBLEM` or brainstorm summary when needed
- removed feature rows use the retained tombstone metadata because their
  feature docs no longer exist
- removed feature summaries should point to retained `docs/notes/<feature>`
  content when it still exists
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

Project-scoped Kit commands traverse upward to find `.kit.yaml` and use its
location as the project root. This allows commands to run from any
subdirectory. Read-only command-discovery surfaces such as `kit capabilities`
are explicit exceptions: they do not require a Kit project root or load project
configuration.

### 7.1 Core Fields

```yaml
goal_percentage: 95
specs_dir: docs/specs
skills_dir: .agents/skills
constitution_path: docs/CONSTITUTION.md
allow_out_of_order: false # if true, kit legacy plan/tasks create missing prerequisites
loop:
  min_confidence: 95
  max_iterations: 20
  agent:
    command: codex
    args:
      - --ask-for-approval
      - never
      - exec
      - --model
      - gpt-5.6
      - --sandbox
      - workspace-write
      - --ignore-user-config
      - --color
      - never
      - "-"
project_refresh:
  constitution:
    feature_interval: 5
    max_age_days: 30
instruction_scaffold_version: 2
feature_state:
  0001-feat-name:
    paused: false
removed_features:
  - number: 1
    slug: feat-name
    dir_name: 0001-feat-name
    created_at: "2026-04-05T00:00:00Z"
    removed_at: "2026-05-06T12:00:00Z"
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

```yaml
feature_naming:
  numeric_width: 4
  separator: "-"
```

Feature names always use numeric prefix + slug format (e.g., `0001-feat-name`).

### 7.4 Prompt Library

Project-local prompt entries live in project-root `.kit.yaml`. Global prompt
entries use the same shape in `~/.config/kit/.kit.yaml`. `kit init` creates or
updates the global config with missing default fields without replacing existing
prompt entries.

```yaml
prompts:
  custom:
    review:
      content: |
        Review the current changes for correctness, edge cases, and tests.
      description: Custom review prompt
```

Rules:

- `content` is required.
- `description` is optional.
- noun and verb keys normalize to lowercase kebab-case.
- local prompts override global prompts with the same noun and verb.
- global prompts override built-in prompts when no local prompt exists.
- built-in prompts are the lowest-precedence fallback.

CLI flags always override `.kit.yaml`.

---

## 8. CLI Commands

### 8.1 Initialization

#### `kit init`

- create `.kit.yaml` if missing
- create or populate `~/.config/kit/.kit.yaml` with missing default fields
- create blank `.env` and default `.envrc` if missing
- include `.env` and `.envrc` in `.gitignore`
- create `.coderabbit.yaml` if missing
- create `.github/pull_request_template.md` if missing
- create `.github/workflows/auto-assign.yml` if missing, using project-local
  `github.default_assignees` with global config fallback and a non-blocking
  no-op when no assignees are configured
- create or refresh the Kit-managed `README.md` badge block when a GitHub
  repository is configured or discoverable from `origin`; default public
  repository badges cover last commit, open issues, pull requests, releases,
  and conventional CI workflows; private repositories skip public Shields GitHub
  metadata badges and keep only native GitHub Actions workflow badges when a
  conventional workflow exists; no default License badge is added
- create or refresh `## Maintainers` as the last README H2 with the managed
  Jameson / `jamesonstone` attribution
- create `docs/CONSTITUTION.md` if missing
- scaffold configured agent instruction files and `.github/copilot-instructions.md`
- if files exist, preserve them; Kit-managed markdown documents may merge missing required sections
- output a prepared prompt for drafting `docs/CONSTITUTION.md`
- by default, copy that prompt to the clipboard instead of printing the prompt body
- make the first visible next step: paste the copied prompt into the agent to draft `docs/CONSTITUTION.md`
- support `--output-only` to print the raw prompt to stdout instead of copying it
- support `--copy` to also copy the prompt when `--output-only` is set
- support `--refresh` as the existing-project structural refresh mode for Kit-managed files
- support `--refresh` backfilling or upgrading known generated default `loop.agent.command` config needed by `kit loop review`
- support `--refresh --file=README.md` for targeted README badge-block and
  Maintainers section adoption or update
- support `--refresh --dry-run --diff` to print planned Kit-managed file changes without writing them
- support `--refresh --force` for generated documentation and ruleset overwrites
- after full `--refresh --force`, copy a documentation review prompt to the
  clipboard so an agent can update `docs/CONSTITUTION.md`, agent docs,
  references, command docs, and directly affected feature specs semantically
- support `--refresh --file=<path> --force` for targeted per-file scaffold overwrites

---

### 8.2 Feature Lifecycle

#### `kit legacy brainstorm [feature]`

- prompt interactively for a feature name and short issue or feature thesis
- create or reuse the feature directory (uses `0001-feat-name` format)
- when Git common-dir state is available, reserve the next feature number from the shared clone-local allocator before creating the directory
- create `BRAINSTORM.md` as the first artifact in the feature directory
- output a planning-only `/plan` prompt for a coding agent
- support `--backlog` to capture a deferred brainstorm item without outputting a
  planning prompt
- support `--prepare` to create the brainstorm document structure and notes
  directory before starting the prompt-driven brainstorm workflow
- keep `--pickup` as a compatibility path for resuming a deferred backlog item
  while teaching `kit resume <feature>` or `kit backlog --pickup <feature>` as
  the canonical resume flows
- require the agent to use indices first for prior work discovery: `kit map
  <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md`
- require prior feature docs to be conditional reads gated by explicit
  relevance: shared interfaces or contracts, overlapping files or modules,
  migrations or data shape, acceptance criteria, or explicit relationship or
  reference links
- require the agent to inspect at most 5 prior feature directories before
  narrowing further or asking a clarifying question
- require the agent to extract only concrete decision-shaping facts from prior
  work instead of replaying full historical docs into chat or active artifacts
- require the agent to keep `BRAINSTORM.md` `## DEPENDENCIES` current with the supporting inputs used during the research phase
- require the agent to populate every `BRAINSTORM.md` section and replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
- require the agent to resolve repository-discoverable ambiguity itself and ask concise numbered questions only for material non-discoverable choices, with a recommended default and why each answer changes the result
- require the agent to stop before finalizing while a material question remains, and otherwise continue when the configured understanding threshold is reached

---

#### `kit backlog [feature]`

- list deferred backlog items captured as paused legacy brainstorm-phase features
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

#### `kit legacy`

- list the legacy staged workflow commands retained for migration
- keep `kit spec <feature>` documented as the default v2 workflow entry point
- do not mutate repository files
- list `brainstorm`, `plan`, `tasks`, `implement`, and `reflect` with wording that marks them as legacy staged commands
- direct users to `kit <command> --help` for command-specific migration flags

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

- create or adopt `docs/specs/<feature>/SPEC.md` as the v2 durable feature artifact
- keep `kit spec` prompt-producing by default; direct execution belongs to explicit loop/run surfaces such as `kit loop workflow`
- add compact v2 front matter and missing v2 sections during normal execution, but avoid migration writes under `--prompt-only`
- add or backfill structured clarification metadata during normal execution:
  `clarification.status`, `clarification.confidence`, and
  `clarification.unresolved_questions`
- template uses the fixed v2 section order: Thesis, Context, Clarifications, Requirements, Assumptions, Acceptance Criteria, Implementation Plan, Task Checklist, Validation Map, Reflection Notes, Documentation Updates, Delivery Decision, Evidence
- prompt instructions define goals, context, constraints, approval boundaries, success criteria, stage-specific output contracts, a repo-grounded material-ambiguity gate, implementation rules, validation/verification, reflection, documentation sync, delivery, `SPEC.md` updates, and the final response contract
- prompt instructions require indices-first prior work discovery through
  `kit map <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md`
- prompt instructions require prior feature docs to stay conditional and
  relevance-gated before widening codebase discovery
- prompt instructions require extracted prior-work facts to stay concise and
  decision-shaping rather than historical replay
- prompt instructions require every `SPEC.md` section to be populated; if a section has no additional detail, replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`
- prompt instructions require binary-verifiable acceptance criteria mapped 1:1 to validation evidence
- create feature directory if missing (uses `0001-feat-name` format)
- when Git common-dir state is available, reserve the next feature number from the shared clone-local allocator before creating the directory
- without a feature argument, show eligible v2 features or prompt for a new feature name according to the interactive selector flow
- when no eligible existing feature candidates are available, prompt for a new feature name and
  start the interactive spec builder
- update `docs/PROJECT_PROGRESS_SUMMARY.md`

---

#### `kit notes [feature]`

- manage optional source-material directories under `docs/notes/<feature>`
- with a feature argument, create or refresh that feature's notes scaffold
- without a feature argument, show an interactive selector for existing features
  or creating a new feature notes directory
- create feature directories when needed so notes can be captured before
  `SPEC.md` exists
- scaffold `README.md`, `inbox/.gitkeep`, `references/.gitkeep`,
  `responses/.gitkeep`, `private/.gitignore`, and `private/README.md`
- preserve local files and create missing scaffold files only
- treat notes as source material, not canonical truth; durable conclusions must
  move into `SPEC.md`, `docs/CONSTITUTION.md`, or another canonical project doc
- use `docs/references/rules/feature-notes.md` when deciding how agents should
  load, reference, promote, or ignore feature notes
- keep `private/` contents ignored by git while tracking the private directory
  contract files
- support `--add` to create a timestamped note template with front matter:
  `kind`, `source`, `status`, `sensitivity`, `captured_at`, and `feature`
- support `--section`, `--source`, `--status`, `--sensitivity`, `--private`,
  `--copy-path`, and `--json`

---

#### `kit legacy plan <feature>` (Legacy Staged)

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
- without a feature argument, show only legacy spec-phase features that have `SPEC.md`
  without `PLAN.md`
- update `PROJECT_PROGRESS_SUMMARY.md`

Prerequisites:

- `SPEC.md` must exist (error otherwise)

Flags:

- `--force` — create empty `SPEC.md` with headers if missing

---

#### `kit legacy tasks <feature>` (Legacy Staged)

- scaffold `TASKS.md` template for manual editing
- prompt instructions require every `TASKS.md` section to be populated; if a section has no additional detail, replace placeholder-only content with `not applicable`, `not required`, or `no additional information required`
- tasks link to plan items using `[PLAN-01]` syntax
- without a feature argument, show only legacy plan-phase features that have `SPEC.md`
  and `PLAN.md` without `TASKS.md`
- update `PROJECT_PROGRESS_SUMMARY.md`

Prerequisites:

- `PLAN.md` must exist (error otherwise)

Flags:

- `--force` — create empty `SPEC.md` and `PLAN.md` with headers if missing

---

#### `kit loop workflow [feature]`

- run the v2 feature workflow through a configured local agent loop
- keep `kit spec` as the v2 prompt-producing entry point; `kit loop workflow` is the explicit execution surface for the same `SPEC.md` contract
- keep legacy staged commands under `kit legacy`, such as `kit legacy plan`, `kit legacy tasks`, `kit legacy implement`, and `kit legacy reflect`
  as compatibility prompt/artifact builders rather than primary v2 workflow commands
- resolve the next strict stage from canonical docs, not file existence alone
- reject advancement past clarify unless `clarification.status` is `ready`,
  confidence meets the loop threshold, `clarification.unresolved_questions` is
  `0`, acceptance criteria use stable IDs, and validation maps those IDs
- during clarify, research discoverable repository facts and update `SPEC.md`;
  if user decisions remain, stop with exact questions instead of guessing
- require each agent turn to end with `KIT_LOOP_RESULT` JSON containing stage,
  status, confidence, and blockers
- stop on nonzero agent exit, low confidence, blockers, malformed loop result,
  failed strict doc validation, failed verification evidence, or max iterations
- write local loop evidence under `.kit/loops/<run-id>/`
- without a feature argument, use the active feature; fail if no active feature exists
- legacy `kit loop [feature]` remains a compatibility alias for this workflow

Flags:

- `--dry-run` — show the next loop action without invoking the configured agent
- `--until <stage>` — run until `clarify`/`spec`, `ready`, `implement`,
  `validate`, `reflect`, `deliver`, or `complete` is complete
- `--min-confidence <0-100>` — override `loop.min_confidence`
- `--max-iterations <n>` — override `loop.max_iterations`
- `--json` — output the loop report as JSON

---

#### `kit loop review [feature]`

- run a configured coding-agent correctness loop over changed code
- without `--pr`, review current-branch changes relative to `origin/main`,
  falling back to local `main`, plus staged and unstaged changes
- with `[feature]`, include the feature docs as review context
- require final agent output to start with `Correctness: <n>%`, include dense
  issue/fix bullets, and end with a final line exactly equal to `done`
- continue looping until correctness is at least the configured threshold and
  no high, medium, or correctness-impacting issues remain
- default to `loop.max_iterations`, then 20
- write local loop evidence under `.kit/loops/<run-id>/`
- never stage, commit, push, post PR comments, or resolve review threads
- with `--pr <target>`, start local review immediately and opportunistically
  ingest current unresolved CodeRabbit feedback during later passes
- when local review reaches `done`, perform one quick PR feedback check
- if CodeRabbit is still pending in default PR mode, exit with a clear
  provisional result and rerun command instead of waiting
- with `--watch` or `--wait-for-coderabbit`, wait up to the existing timeout
  before finalizing PR mode

Flags:

- `--base <ref>` — override the comparison base
- `--pr <target>` — ingest CodeRabbit feedback from a PR
- `--watch` — wait for CodeRabbit completion before finalizing PR mode
- `--wait-for-coderabbit` — alias for `--watch`
- `--dry-run` — show the first review prompt without invoking the agent
- `--min-confidence <0-100>` — override `loop.min_confidence`
- `--max-iterations <n>` — override `loop.max_iterations`
- `--subagents` — allow parent review pre-analysis to choose subagents when the changed-code lanes are clearly independent
- `--json` — output the loop review report as JSON

---

#### `kit pr fix`

- provide the default human-facing PR review feedback prompt entrypoint
- without `--pr`, list open pull requests in the current repository and ask
  which one to repair
- with `--pr <target>`, accept the same PR target forms as dispatch PR intake:
  full GitHub PR URL, Markdown PR link, `owner/repo#123`, or current-repo PR
  number
- route the selected PR through the prompt-producing `kit dispatch --pr` path:
  prepopulate the editor from unresolved review feedback, let the user edit the
  task list, and copy the resulting dispatch prompt for a coding agent
- preserve the delivery boundary: do not run the loop agent, edit files, write
  `.kit/loops` evidence, stage, commit, push, post PR comments, resolve review
  threads, or perform GitHub delivery
- after fixes or no-op decisions are validated, resolve matching current
  unresolved review threads on the PR, including human reviewer and CodeRabbit
  feedback, through `kit dispatch --pr <target> --resolve --yes`
- resolve only feedback verified as fixed or intentionally no-op; do not
  resolve unfixed, uncertain, stale, or unrelated feedback
- keep `kit dispatch --pr <target> --coderabbit` as the raw unresolved
  review-thread prompt intake path
- keep `kit dispatch --pr <target> --resolve --yes` as the explicit mutation
  path after fixes or no-op decisions are complete

Flags:

- `--pr <target>` — target a PR without using the selector
- `--coderabbit` — include only CodeRabbit-authored review comments
- `--editor <cmd>` — open review tasks in a specific editor command
- `--vim` — open review tasks in a vim-compatible editor
- `--copy` — copy generated prompt output even with `--output-only`
- `--output-only` — print prompt text instead of copying it
- `--max-subagents <n>` — maximum concurrent subagents allowed in the
  generated prompt; default 3, hard ceiling 4

---

#### `kit legacy implement [feature]` (Legacy Staged)

- output implementation context for a coding agent in the legacy staged workflow
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
- without a feature argument, show only legacy implement-phase features with incomplete
  task checkboxes; omit task-template, reflection-ready, and complete features
- if the target feature is paused, clear the paused flag before continuing

---

#### `kit pause [feature]`

- resolve the target feature using existing feature reference rules
- without a feature argument, show an interactive selector of non-complete features
- persist paused state in `.kit.yaml` without changing the feature's underlying phase
- reject complete features with an actionable error
- update `PROJECT_PROGRESS_SUMMARY.md`

---

#### `kit rm [feature]`

- resolve the target feature using existing feature reference rules
- without a feature argument, show an interactive selector of existing features
- require explicit confirmation before deletion unless `--yes` is set
- delete the feature directory and all files under it
- retain `docs/notes/<feature>` by default so follow-up features can reuse
  research notes
- when running interactively and notes exist, ask whether to remove the notes
  as part of the same removal flow
- support `--notes` to remove `docs/notes/<feature>` along with the feature
  docs
- remove any persisted paused lifecycle state for the deleted feature from
  `.kit.yaml`
- record a removed-feature tombstone in `.kit.yaml`
- update `PROJECT_PROGRESS_SUMMARY.md`
- retain the feature in `PROJECT_PROGRESS_SUMMARY.md` with `PHASE` set to
  `removed`
- print removal output that makes the final `removed` state and notes
  retention or deletion visible
- when no feature argument is provided, show already-removed tombstones as
  removed history alongside the live removal selector

#### `kit remove [feature]`

- compatibility alias for `kit rm [feature]`

---

#### `kit map [feature]`

- render a read-only terminal graph of the canonical document hierarchy and current project state
- without a feature argument, open the interactive feature selector and show the selected feature map
- with `--all`, show global docs plus every feature directory, current phase, paused state, canonical docs, and explicit relationship edges
- order project-wide feature rendering by `builds on` and `depends on` relationships when those edges provide a usable dependency order
- with a feature argument, show a feature-scoped view plus incoming or outgoing relationships that touch that feature
- derive relationship edges from explicit `## RELATIONSHIPS` sections in `BRAINSTORM.md` and `SPEC.md`
- normalize harmless inline-code formatting around relationship targets before validating or rendering them
- if a relationship line is malformed, keep the valid map output and surface the skipped line as a warning instead of failing the read-only command
- when writing to a terminal, map output may color labels and state markers for scanability without changing non-TTY output
- do not create another persisted markdown graph document

---

#### `kit capabilities [command]`

- render a read-only catalog of Kit command behavior for humans and agents
- require no Kit project root and do not load project configuration
- perform no file writes, network calls, subprocess execution, delegated Kit
  command execution, or git mutation
- default output lists visible canonical commands with compact metadata
- support `--json` with `schema_version: 1` and `kind: capabilities_index`
- support targeted command paths such as `kit capabilities legacy verify --json`,
  `kit capabilities pr fix --json`, `kit capabilities rules add --json`, and
  `kit capabilities skill mine --json`
  with `kind: capability_detail`
- support `--full --json` with detailed records, including hidden and
  deprecated compatibility commands labeled explicitly
- support `--search <term> --json` with compact filtered records for visible
  commands only
- reject `--search <term>` plus positional command paths and `--full` plus
  positional command paths with actionable non-zero errors
- return actionable unknown-command errors with suggestions when possible
- include mutation level, network use, file-write behavior, git mutation,
  hidden/deprecated state, important flags, and related commands in compact
  records
- include when-to-use, when-not-to-use, examples, caveats, aliases, and detailed
  flag behavior in detail and full records
- human detail output must include agent-readable guidance for safe command
  selection, including when to use the command, when not to use it, examples,
  caveats when present, important flag safety notes, and related commands
- in the Kit source repository, every new or changed command, subcommand, flag,
  alias, prompt surface, or behavior extension must update `kit capabilities`
  in the same change
- Kit maintainer command-surface work must follow
  `docs/references/rules/command-capabilities.md`
- downstream Kit-managed projects must receive `kit-capabilities-usage` guidance
  through registry refresh and must not be told to edit Kit's internal
  `pkg/cli/capabilities_catalog.go`

---

#### `kit status`

- default text output remains focused on the active feature only
- show current summary, phase, paused state, file presence, progress, and next
  recommended action for the active feature
- include a compact Kit-managed files section that reports whether generated
  project files have local refresh changes available and whether locally tracked registry rules
  are conflict, local-custom, or missing
- keep the existing default `--json` active-feature fields while adding a
  top-level `kit_managed` object for local managed-file and registry refresh state
- support `--all` as the explicit fleet overview mode
- `--all` text output shows every feature in a terminal-friendly fixed-width
  lifecycle matrix with paused or backlog state and available task progress
- `--all` includes removed feature tombstones with `State` set to `REMOVED`
  and a notes-retention marker
- `kit status` may fetch the Kit ruleset registry so its managed-file refresh
  state uses the same planning engine as refresh; use `kit reconcile
  --include-files --dry-run --diff` to preview registry and managed-file
  updates, and `kit reconcile` to interactively apply reviewed updates or
  audit Kit-managed documentation drift
- when writing to a terminal, status views may color lifecycle markers and
  state labels for scanability without changing non-TTY output
- `--all --json` uses a dedicated all-features payload and does not replace the
  default `--json` contract

---

### 8.3 Prompt Library

#### `kit dispatch --loop --pr <target>`

- prepare a dispatch prompt from current unresolved PR review-thread feedback
- accept the same PR target forms as dispatch PR intake: full GitHub PR URL,
  Markdown PR link, `owner/repo#123`, or current-repo PR number
- support `--coderabbit` to include only CodeRabbit-authored review comments
  and extract `Prompt for AI Agents` blocks when present
- support `--watch` to wait for CodeRabbit completion on the current PR head
  before collecting feedback
- classify current findings as `FIX`, `VALID_OUT_OF_SCOPE`,
  `FALSE_POSITIVE`, `STALE`, or `NEEDS_HUMAN`
- open the editor only when actionable `FIX` findings remain, and include
  non-fix classifications in the summary output
- remain read-only by default: no project-file writes, no git mutation, no PR
  comments, and no review-thread resolution
- `kit dispatch --loop --pr <target>` is the prompt-prep workflow for current
  unresolved PR review feedback
- leave `kit dispatch --pr <target> --coderabbit` as the lower-level
  untriaged review-thread intake
- after fixes or no-op decisions are complete, support
  `kit dispatch --pr <target> --resolve --yes` to resolve currently matching
  unresolved review threads on GitHub; this is an explicit mutation and must
  not be part of default review-loop or dispatch prompt generation

#### `kit prompt [noun] [verb]`

- resolve reusable prompts by explicit noun and verb
- with no args, show a noun selector followed by a verb selector
- with one arg, show the verb selector for that noun
- with two args, resolve directly or fail with nearest-match guidance
- copy the selected prompt body to the clipboard by default
- print command name, origin, override metadata, and a delimited prompt body in human-readable default mode
- support `--output-only` to print only the raw prompt body and skip copying
- support `--output-only --copy` to print and copy the same raw prompt body

#### `kit prompt list`

- render the effective merged prompt library as a deterministic table
- include command name, description, origin, and override metadata
- show only effective prompts after precedence is applied

Built-in project prompts:

- `project init` drafts the initial `docs/CONSTITUTION.md` prompt
- `project refresh` refreshes durable project-level docs after the repository has real contents
- `project refresh` is docs-only, starts with `/plan`, uses `kit reconcile --all` for structural drift, and verifies with `kit check --project`

#### `kit set`

- delegate to `kit set prompt` in v0 because `prompt` is the only configurable resource

#### `kit set prompt [noun] [verb]`

- create or update prompt entries through the existing editor flow, defaulting to `$EDITOR` and falling back to a vim-compatible editor when `$EDITOR` is unset
- default to project-local `.kit.yaml` when run inside a Kit project
- ask whether to save globally when run outside a Kit project with no scope flag
- support `--local`, `--global`, and `--local --global`
- write global prompts to `~/.config/kit/.kit.yaml`, creating the file when needed
- confirm before overwriting each selected scope
- reject stdin, `--file`, auto-paste, clipboard restore, `--source`, and `--no-copy` in v0

Prompt precedence:

1. project-local `.kit.yaml`
2. global `~/.config/kit/.kit.yaml`
3. built-in Kit prompts

Prompt entries use nested YAML object form:

```yaml
prompts:
  custom:
    review:
      content: |
        Review the current changes for correctness, edge cases, and tests.
      description: Custom review prompt
```

Nouns and verbs normalize to lowercase kebab-case.

#### `kit rules`

- manage durable repo-local rulesets under `docs/references/rules/<slug>.md`
- `kit rule` is a singular alias for `kit rules`
- keep Markdown rulesets as human source of truth with front matter:
  - `kind: ruleset`
  - `slug`
  - `status`
  - `applies_to`
  - `read_policy_default`
- `kit rules add` with no slug opens an interactive registry selector backed by the Kit GitHub `main` branch so users can import or activate available rulesets and toggle existing registry rules active or inactive
- registry rulesets may include a front matter `description` field that explains the rule's function
- registry selector status/state text is highly visible and colorized when terminal output supports color, and selector rows include the ruleset description when present
- terminal selector mode supports viewing the highlighted rule before applying changes
- inactive registry rules are preserved locally with `status: optional` instead of being deleted
- locally modified rulesets are preserved; activation and deactivation change only ruleset status
- `kit rules view <slug>` previews a local ruleset when installed, or the matching registry ruleset before importing it
- `kit rules add --custom` runs the interactive builder, opens `$EDITOR` for rule context by default, saves the ruleset, and copies an agent optimization prompt for semantic cleanup
- `kit rules add <slug>` creates a concise custom ruleset template non-interactively and refuses to overwrite unless `--force` is used
- `kit rules add --custom` and `kit rules add <slug>` support `--must`, `--conditional`, `--evidence`, and `--skip` to set `read_policy_default`
- `kit rules list` renders rulesets in stable slug order with slug, path, status, and `applies_to`
- `kit rules link <feature> <slug> --read-policy must|conditional` adds or refreshes one canonical feature `references` entry without duplicating existing references
- feature docs decide when a ruleset is loaded by setting `read_policy: must`, `conditional`, or `skip`
- agents must load rulesets just in time and only the sections relevant to the current implementation decision
- do not inline rulesets into `AGENTS.md`, `CLAUDE.md`, copilot instructions, or prompt bodies by default

---

### 8.4 Roll-Up

#### Project Progress Summary

Purpose:

- analyze all feature specifications under `docs/specs/`
- generate or update `PROJECT_PROGRESS_SUMMARY.md`

Behavior:

- summarize each feature’s intent, approach, and implementation state
- include a table of all features with:
  - feature name
  - directory path
  - short summary
  - current workflow phase (`clarify | ready | implement | validate | reflect | deliver | complete | blocked | removed`, with legacy staged states allowed for historical projects)
  - paused state
  - spec creation date
- include removed feature tombstones from `.kit.yaml` even after the feature
  docs are deleted
- include retained notes pointers for removed features when
  `docs/notes/<feature>` still exists
- refresh automatically from normal feature lifecycle commands

`PROJECT_PROGRESS_SUMMARY.md` is intended to be:

- a high-level briefing document
- sufficient to onboard or fork the project
- safe to hand to any coding agent as primary context

Project summary refresh is executed automatically as the final stage of feature
creation and refinement.

---

### 8.5 Verification

#### `kit legacy verify [feature]`

- run verification commands declared in legacy staged `TASKS.md`
- write evidence under `.kit/runs/...` unless `--dry-run` or `--no-write` is used
- remain available for migration and reusable local evidence
- do not serve as the normal standalone verification step in the v2 happy path
- let the v2 `kit spec` supervisor prompt reference useful run evidence from `SPEC.md`

---

#### `kit check <feature>`

Validates:

- required documents exist
- required sections present and populated
- traceability between spec → plan → tasks
- no unresolved placeholders
- ruleset references point to existing, valid files under `docs/references/rules/`

Flags:

- `--all` — validate all features in `docs/specs/`
- `--project` — validate repo-level docs, instruction docs, and ruleset documents

Fails fast with explicit errors. Errors suggest fixes (e.g., "SPEC.md missing. Run `kit spec <feature>` first or use `--force`").

---

### 8.6 Empty Workflow Scaffolding

#### `kit scaffold`

Purpose:

- create empty workflow document structures and supporting directories
- do not output workflow prompts
- do not start an agent phase

Subcommands:

- `kit scaffold brainstorm <feature>` — create or reuse the feature directory, create `BRAINSTORM.md`, and create the full `docs/notes/<feature>` scaffold
- `kit scaffold spec <feature>` — create or reuse the feature directory, create `SPEC.md`, and create the full `docs/notes/<feature>` scaffold
- `kit scaffold plan <feature>` — require `SPEC.md` and create `PLAN.md`
- `kit scaffold tasks <feature>` — require `PLAN.md` and create `TASKS.md`
- `kit scaffold agents` — create or refresh repository instruction files

Completion output:

- `♻️ <doc_type/workflow> directory and files empty scaffolding created. Please prepare your notes, documents, images, and examples for the <doc_type/workflow> phase`

#### `kit legacy brainstorm <feature> --prepare`

- aliases the brainstorm workflow scaffold behavior
- creates `BRAINSTORM.md` and the full `docs/notes/<feature>` scaffold
- current notes scaffolding includes `README.md`, `inbox/`, `references/`, `responses/`, and tracked private-directory guardrails
- when the frontend profile is active, also creates design-materials directories
- does not ask for a brainstorm thesis
- does not output or copy the brainstorm prompt
- rejects prompt-output flags and interactive thesis flags

---

### 8.7 Agent Scaffolding

#### `kit scaffold agents`

- create missing repository instruction files
- overwrite existing repository instruction files only when `--force` is set
- prompt for confirmation before `--force` overwrites existing instruction files
- support `--yes` / `-y` to skip the overwrite confirmation prompt when `--force` is used
- support `--append-only` to merge missing Kit-managed sections without overwriting matched existing content
- scaffold `.github/copilot-instructions.md` alongside configured agent files
- without targeted flags, scaffold configured agent files plus `.github/copilot-instructions.md`
- `--agentsmd` scaffolds only `AGENTS.md`
- `--claude` scaffolds only `CLAUDE.md`
- `--copilot` scaffolds only `.github/copilot-instructions.md`
- allow combining targeted flags to scaffold multiple specific built-in files in one run
- in default mode, suggest `--append-only` and `--force` when existing instruction files are skipped

---

### 8.8 Documentation Reconciliation

#### `kit reconcile [feature]`

Purpose:

- audit Kit-managed docs and init scaffold artifacts against the current Kit contract
- include current Kit-managed file and ruleset refreshes when requested
- output a prompt for a coding agent to reconcile stale or missing documentation and scaffold drift
- keep reconciliation scoped to Kit-managed docs and scaffold files, with no product-code edits

Behavior:

- without feature argument: audits the whole project by default
- when run interactively without flags, asks `include files?`, `force these changes?`, and `output coding-agent prompt too?`
- `--include-files` refreshes Kit-managed project files, registry rulesets, README badges, config backfills, init scaffold artifacts, and instruction docs before auditing documentation
- `--force` replaces supported generated files and registry rulesets only when the user intentionally accepts those changes
- `--dry-run --diff` previews included file refreshes without writing files
- with feature argument: audits the selected feature plus related rollup drift
- whole-project audits include repo-local artifacts created or updated by `kit init`
- emits a short clean result when no reconciliation is needed
- emits a clipboard-first prompt when reconciliation findings exist

Findings:

- missing `.gitignore` or missing current Kit-managed `.gitignore` entries
- missing local init scaffold artifacts such as `.env` or `.envrc`
- missing tracked init scaffold artifacts such as `.coderabbit.yaml`, `.github/pull_request_template.md`, or `.github/workflows/auto-assign.yml`
- missing required docs or sections
- placeholder-only required sections
- malformed `SKILLS`, `DEPENDENCIES`, or `PROGRESS TABLE` tables
- task-ID drift across `PROGRESS TABLE`, `TASK LIST`, and `TASK DETAILS`
- stale `RELATIONSHIPS` targets
- invalid or missing ruleset reference targets
- advisory active-feature warnings when a frontend feature appears to lack a linked frontend ruleset
- stale `PROJECT_PROGRESS_SUMMARY.md` coverage
- repository instruction-file drift detectable through append-only planning

Verification:

- run `kit check --all` for project-wide reconciliation or `kit check <feature>` for feature-scoped reconciliation
- run `kit reconcile --include-files --dry-run --diff` when reconcile reports init scaffold drift, then apply the intended refresh with `kit reconcile`
- refresh `PROJECT_PROGRESS_SUMMARY.md` when reconciled changes affect the
  project summary

---

### 8.9 Context Summarization

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

### 8.10 Reflection

#### `kit legacy reflect [feature]` (Legacy Staged)

Purpose:

- output instructions for reflecting on recent changes in the legacy staged workflow
- ensure 100% implementation correctness
- verify changes using git, lint, and tests

Behavior:

- without feature argument: show only legacy reflect-phase features whose task
  checkboxes are complete and whose `TASKS.md` does not yet contain the
  reflection-complete marker
- with feature argument: outputs instructions scoped to that feature's context

Reflection Process:

1. analyze git state (staged and unstaged changes)
2. understand the delta and intent of each change
3. cross-reference with repository context and codebase
4. verify correctness checklist (compiles, no errors, edge cases handled)
5. run lint and tests, then fix ALL failures (including out-of-scope failures) before completion
6. run the soft project-refresh advisory gate: if work revealed durable project-level rules, run `kit project refresh`; otherwise state no project refresh is needed
7. do not run `coderabbit --prompt-only` unless the user explicitly asks for it or explicitly approves it first

Completion advisory:

- after `kit complete` succeeds, Kit prints a non-blocking reminder to run `kit project refresh` if the completed work changed durable project-level truth

---

### 8.11 Agent Handoff

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
  - reference-inventory verification for touched feature docs
- with feature argument: outputs feature-specific context including:
  - feature location and phase
  - required reading (`SPEC.md` first for v2 work; legacy staged artifacts only when present and materially relevant)
  - instructions to refresh front matter references in touched `SPEC.md` files and in legacy staged docs when those docs are active
  - a final response contract for concise documentation sync and recent-context summary

Flags:

- `--copy` / `-c` — copy output to clipboard (pbcopy)

Use case: when you run out of tokens or hit rate limits, run `kit handoff`, let the current agent reconcile docs and reference inventories, then transfer the final handoff summary.

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
- `SPEC.md` is the binding v2 execution artifact
- legacy staged artifacts remain readable historical context and are binding only when a legacy staged command is explicitly used

Kit does not:

- manage branches or PRs
- enforce git policies
- maintain any state beyond files (no database, no lock files)

---

## 11. Non-Goals

Kit explicitly does not:

- execute code
- manage agents directly
- maintain hidden prompt registries outside YAML files
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
