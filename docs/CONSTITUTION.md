# CONSTITUTION

Kit v2 is a general-purpose harness for disciplined thought work.
Its strongest engine is a document-first, spec-driven workflow, but the
harness also supports ad hoc execution, catch-up, handoff, summarization,
review, and orchestration. The current command surface is packaged around
repository and software workflows, but the underlying concepts generalize to
research, strategy, operations, writing, policy, and other structured fields.
This constitution defines the invariant rules, patterns, and vision that guide
all decisions.

---

## PRINCIPLES

### 1. Harness First, Workflow Second

- Kit is a harness, not a single rigid workflow.
- Spec-driven planning is a first-class engine, not the only operating mode.
- Ad hoc work, recovery flows, and orchestration flows are part of the product surface, not side utilities.
- The harness should help teams choose the right level of structure for the work.
- Software is a current packaging default, not the conceptual boundary of the product.

### 2. Documents Are the Source of Truth

- Specifications drive code. Code serves specifications.
- All decisions must be traceable to a document.
- If reality diverges from documentation, update documentation first, then code.
- The repository should be understandable by reading docs alone.

### 3. Portable and Agent-Agnostic

- No vendor lock-in to any coding agent (Claude, Copilot, Codex, etc.).
- Documents use only markdown and YAML — universally readable.
- Repository instruction files (`AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`) stay aligned with canonical docs and summarize the active workflow contract for supported tools.
- `kit init` also creates shared local support files, including `.env`, `.envrc`, `.coderabbit.yaml`, `.github/pull_request_template.md`, and the optional `.github/workflows/auto-assign.yml` workflow, without overwriting custom local versions.
- Agents can be swapped with zero document changes.

### 4. Minimal Magic, Explicit State

- Prefer explicit over implicit behavior.
- No hidden databases, lock files, or external state.
- All state lives in the filesystem (markdown files + `.kit.yaml` + visible generated `.kit/` artifacts).
- Markdown remains authoritative; `.kit/runs/`, `.kit/loops/`, and `.kit/state.json` are local generated evidence/state surfaces that can be regenerated or deleted.
- Commands fail fast with actionable error messages.

### 5. Opinionated Defaults, Configurable Escapes

- Sensible defaults work out of the box.
- Configuration overrides via `.kit.yaml` for team customization.
- CLI flags always override configuration.
- Escape hatches exist but aren't encouraged.

### 6. Tooling Should Disappear

- Kit's job is done once documents are complete and correct.
- Application implementation remains outside Kit's product scope.
- Kit may run declared local verification commands or configured agent loops only when an explicit command contract says so.
- The CLI becomes unnecessary once understanding is achieved.
- Teams should reach clarity faster with fewer reworks.

### 7. Density Over Prose

- Documents prioritize brevity and precision.
- One sentence where possible.
- Bullet points over paragraphs.
- Code snippets only when unavoidable.

---

## CONSTRAINTS


### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Treat `docs/notes/<feature>` as optional source material, not canonical truth; promote durable decisions into `SPEC.md`, `docs/CONSTITUTION.md`, or durable references.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all `docs/**`, all `.kit/**`, or `.kit.yaml`.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->
### Non-Negotiable Rules

1. **V2 Single-SPEC Workflow**
   - Constitution → `kit spec <feature>` → `SPEC.md` phases: clarify → ready → implement → validate → reflect → deliver → complete
   - `SPEC.md` is the single durable feature artifact for v2 feature work
   - The v2 readiness gates happen inside `SPEC.md`: clarification complete, assumptions resolved, acceptance criteria binary-verifiable, task checklist mapped to criteria, validation mapped 1:1, delivery intent known
   - The readiness gate adversarially challenges `CONSTITUTION.md` and `SPEC.md` for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, validation gaps, delivery ambiguity, and scope creep before code begins
   - If a readiness gate fails, update `SPEC.md` first, then restart the relevant phase
   - Legacy staged artifacts (`BRAINSTORM.md`, `PLAN.md`, `TASKS.md`) are preserved as historical context or when a legacy staged command is explicitly used

2. **Document Structure**
   - All canonical markdown files use FULL CAPITALIZATION (e.g., `CONSTITUTION.md`, `SPEC.md`)
   - Use snake_case for multi-word file names (e.g., `PROJECT_PROGRESS_SUMMARY.md`)
   - Use kebab-case for directories (e.g., `0001-feat-name`)
   - Feature directories use numeric prefix + slug format

3. **Section Requirements**
   - Each document type has required sections that must be present
   - Required sections in v2 `SPEC.md` must be populated
   - Required sections in legacy staged artifacts must be populated when those artifacts are actively used
   - Do not leave HTML TODO comments as the only content in a required section
   - If a required section has no feature-specific detail, replace the placeholder comment with `not applicable`, `not required`, or `no additional information required`
   - `CONSTITUTION.md`: PRINCIPLES, CONSTRAINTS, NON-GOALS, DEFINITIONS
   - v2 `SPEC.md`: THESIS, CONTEXT, CLARIFICATIONS, REQUIREMENTS, ASSUMPTIONS, ACCEPTANCE CRITERIA, IMPLEMENTATION PLAN, TASK CHECKLIST, VALIDATION MAP, REFLECTION NOTES, DOCUMENTATION UPDATES, DELIVERY DECISION, EVIDENCE
   - legacy `BRAINSTORM.md`: SUMMARY, USER THESIS, RELATIONSHIPS, CODEBASE FINDINGS, AFFECTED FILES, DEPENDENCIES, QUESTIONS, OPTIONS, RECOMMENDED STRATEGY, NEXT STEP
   - legacy `SPEC.md`: SUMMARY, PROBLEM, GOALS, NON-GOALS, USERS, SKILLS, RELATIONSHIPS, DEPENDENCIES, REQUIREMENTS, ACCEPTANCE, EDGE-CASES, OPEN-QUESTIONS
   - legacy `PLAN.md`: SUMMARY, APPROACH, COMPONENTS, DATA, INTERFACES, RISKS, TESTING
   - legacy `TASKS.md`: PROGRESS TABLE, TASK LIST, TASK DETAILS, DEPENDENCIES, NOTES
   - `RELATIONSHIPS` in `BRAINSTORM.md` and `SPEC.md` must be either `none` or explicit bullets using `builds on: <feature>`, `depends on: <feature>`, or `related to: <feature>`
   - inline-code-wrapped relationship targets are valid if they normalize back to a canonical feature directory identifier

4. **Single Feature Per Directory**
   - Never mix features in one `docs/specs/<feature>/` directory
   - If work spans features, update each feature's docs separately
   - Feature directories are immutable once created

5. **No Premature Implementation Details in Specs**
   - v2 `SPEC.md` captures implementation plan and task checklist only after clarification and readiness gates pass
   - Before the ready phase, `SPEC.md` defines WHAT and WHY before HOW
   - No code in specifications
   - No technology choices unless accepted requirements or repo-grounded constraints require them

6. **Traceability**
   - v2 task checklist items map to acceptance criteria and validation evidence inside `SPEC.md`
   - Legacy staged tasks link to plan items using `[PLAN-XX]` syntax, and plan items link to spec items using `[SPEC-XX]` syntax
   - Every claim in `PROJECT_PROGRESS_SUMMARY.md` must map to a feature document
   - Validation evidence belongs in `SPEC.md` Evidence and Validation Map sections; legacy executable verification may still use task-level `VERIFY` fields where available
   - Generated JSON state and run artifacts must point back to source documents instead of replacing them

7. **Execution Boundaries**
   - Prompt-only and inspection commands must not mutate files, git, GitHub, or external services
   - Local execution surfaces must be explicit in command help and `kit capabilities`
   - Verification, replay, eval, and loop commands may run local commands or configured agent commands only within their documented command contracts
   - Local run evidence under `.kit/runs/` and loop evidence under `.kit/loops/` is generated, inspectable, and non-authoritative
   - Kit must not become a general-purpose task runner, CI replacement, daemon, or hidden supervisor

8. **External Review Tools**
   - Do NOT run `coderabbit --prompt-only` unless the user explicitly asks for it or explicitly approves it first

9. **Project Directory Git Workflow**
   - Work in the existing project directory by default
   - Do not create or use git worktrees for agent work
   - If the current branch or dirty state is unsuitable, stop and ask the user how to proceed instead of creating an alternate checkout

### Code Quality Constraints

1. **Go Best Practices**
   - Single binary with subcommands (Cobra CLI)
   - No global state beyond package-level constants
   - Error handling must be explicit (`%w` for wrapped errors)
   - Test coverage for critical paths

2. **Package Structure**
   - `cmd/kit/` — main entry point only
   - `pkg/cli/` — command implementations
   - `internal/` — private packages for config, documents, features, instruction contracts, prompts, rollups, and templates
   - No circular dependencies

3. **Error Messages**
   - Must be actionable (suggest fixes)
   - Include context (file paths, feature names)
   - Fail fast — don't continue with partial state

4. **Code File Size**
   - Applies only to implementation/source code files
   - Prefer code files around 300 lines or less when splitting improves clarity and ownership
   - Treat the limit as guidance, not a mandate to fragment cohesive code; justify larger code files when the reason is not obvious
   - Excluded: documentation files, all `docs/**`, all `.kit/**`, and `.kit.yaml`
   - Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines

---

## CHANGE CLASSIFICATION

All work falls into one of two tracks. Classify before acting.

### Spec-Driven (Formal)

Use when ANY of these apply:

- Initiated via `kit spec` or explicit legacy staged commands under `kit legacy`
- New feature or capability
- Substantial architectural or behavioral change
- Work that has existing spec docs in `docs/specs/<feature>/`
- Change affects multiple components or public interfaces

**Workflow**: v2 single-SPEC workflow — `kit spec <feature>` creates or adopts `SPEC.md`, then the supervisor prompt drives clarify → ready → implement → validate → reflect → deliver → complete inside that artifact

**Clarification protocol for formal planning phases**:

- Ask in numbered batches of up to 10 questions with a recommended default for each question
- `yes` / `y` approves all recommended defaults in the current batch
- `yes 3, 4, 5` / `y 3, 4, 5` approves only those numbered defaults
- `no 2: <answer>` / `n 2: <answer>` overrides a numbered default
- `no` / `n` rejects the full batch and requires explicit replacements before proceeding

### Ad Hoc (Lightweight)

Use when ALL of these apply:

- Not initiated via `kit spec` or explicit legacy staged commands under `kit legacy`
- Bug fix, security review, refactor, dependency update, config change, or small refinement
- Scope is contained and well-understood without formal specification

**Workflow**: Understand → implement → verify

**Documentation**: Update only practical docs (READMEs, inline docs, API docs). Do NOT create feature `SPEC.md` or legacy staged artifacts for ad hoc work.

### Ad Hoc with Existing Specs

If an ad hoc change touches code covered by existing spec docs in `docs/specs/<feature>/`:

- **Default to updating** the spec docs if the change alters behavior, requirements, or approach
- **Skip spec updates** only if the change is purely mechanical (formatting, typo fix, dependency bump)

### Classification Examples

| Change                              | Track                 | Why                           |
| ----------------------------------- | --------------------- | ----------------------------- |
| New CLI command                     | Spec-driven           | New capability                |
| Fix nil pointer in existing handler | Ad hoc                | Bug fix, contained scope      |
| Security review of auth flow        | Ad hoc                | Review, no new feature        |
| Refactor package structure          | Ad hoc or Spec-driven | Depends on scope              |
| Add streaming support to export     | Spec-driven           | Substantial behavioral change |
| Update dependency version           | Ad hoc                | Mechanical change             |
| Fix typo in error message           | Ad hoc                | Trivial, mechanical           |

---

## NON-GOALS

Kit explicitly does NOT:

### Process & Execution

- ❌ Run undeclared arbitrary commands or act as a general-purpose task runner
- ❌ Manage agents directly or maintain hidden prompt registries outside YAML files
- ❌ Merge branches or manage PRs
- ❌ Replace CI/CD systems

### Data & State

- ❌ Maintain databases or external state
- ❌ Use lock files or semaphores
- ❌ Store credentials or secrets
- ❌ Track metrics or analytics

### Content & Format

- ❌ Invent new document formats
- ❌ Generate prose or content (only templates)
- ❌ Define understanding rubrics or scoring models
- ❌ Duplicate specifications in agent files

### Scope

- ❌ Manage multi-repository projects
- ❌ Handle deployment or infrastructure
- ❌ Provide a web interface or GUI
- ❌ Support non-markdown documentation

---

## DEFINITIONS

### V2 Single-SPEC Workflow

The default feature workflow that keeps durable state in one feature artifact:

1. **Constitution** — Project-wide constraints, principles, long-term vision. Kept updated with priors. Single per repository.

2. **Specification (SPEC.md)** — Single durable v2 feature artifact. Captures thesis, context, clarifications, requirements, assumptions, acceptance criteria, implementation plan, task checklist, validation map, reflection notes, documentation updates, delivery decision, and evidence.

3. **Feature Notes** — Optional source material under `docs/notes/<feature>`. Notes may hold raw context, screenshots, research, Slack/customer excerpts, draft responses, and local-only private context, but they do not replace `SPEC.md`, `docs/CONSTITUTION.md`, or durable references.

4. **Implementation** — Code execution after the `SPEC.md` readiness gates pass. The supervisor keeps task status, lane decisions, validation still needed, and rollback notes current in `SPEC.md`.

5. **Validation and Reflection** — Verification against every acceptance criterion, reflection on correctness and drift, documentation sync, delivery decision, and evidence recorded in `SPEC.md`.

Legacy staged artifacts (`BRAINSTORM.md`, `PLAN.md`, `TASKS.md`) may exist in upgraded projects and remain readable historical context. They are binding only when a legacy staged command is explicitly used.

### Feature

A self-contained unit of work with its own directory under `docs/specs/`. Identified by:

- **Numeric prefix**: Reserved by Kit's feature allocator, using repo-shared Git common-dir state when available and falling back to local `docs/specs/` inspection (e.g., `0001`)
- **Slug**: Lowercase kebab-case name, max 5 words (e.g., `init-project`)
- **Directory**: Combined format (e.g., `0001-init-project`)

Feature notes may exist separately under `docs/notes/<feature>` before, during, or after feature work. The standard scaffold is:

- `README.md` — note directory contract and feature pointer.
- `inbox/` — raw captured inputs, conversation excerpts, and transient context.
- `references/` — source material, links, screenshots, research, and external references.
- `responses/` — draft or sent responses tied to the feature.
- `private/` — local-only ignored context; only `README.md` and `.gitignore` are tracked.

Agents should load note files only when they materially affect the current decision, ignore `.gitkeep` placeholders, and promote durable conclusions into canonical project artifacts.

### Phase

The current lifecycle state of a feature. V2 `SPEC.md` also records its own workflow phase in front matter.

- `brainstorm` — Legacy staged research exists without `SPEC.md`
- `spec` — `SPEC.md` exists and is the v2 workflow artifact, optionally with historical staged context
- `plan` — Legacy staged `PLAN.md` exists
- `tasks` — Legacy staged `TASKS.md` exists
- `implement` — Implementation work is in progress beyond Kit's prompt-generation scope
- `reflect` — Implementation work is ready for validation, reflection, and documentation sync
- `complete` — Marked complete after reflection and lifecycle completion
- `removed` — Removed from `docs/specs/` but retained as history through `.kit.yaml`

### Understanding Percentage

A single integer (0–100) reported by the coding agent indicating:

- Completeness of requirements understanding
- Clarity of implementation approach
- Readiness to proceed to next phase
- Surfaced after each batch of up to 10 numbered clarification questions so the user can see progress

Kit does not define a scoring rubric — the agent determines the value.

### Repository Instruction File

A markdown file (e.g., `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`) that:

- Links to canonical documents
- Defines the active workflow contract for supported tools
- Summarizes repository standards and execution rules
- Must stay aligned with canonical project documents

### Project Root

The directory containing `.kit.yaml`. All Kit commands traverse upward to find this file. Enables running commands from any subdirectory.

### Rollup

The process of scanning all features and generating `PROJECT_PROGRESS_SUMMARY.md`:

- High-level briefing document
- Sufficient to onboard or fork the project
- Primary context input for any coding agent

### Project Refresh

The semantic refresh flow for updating durable project-level truth after a repository matures:

- Invoked with `kit prompt project refresh`
- Updates `docs/CONSTITUTION.md` only when durable project-wide rules, vocabulary, constraints, or conventions changed
- Uses `kit reconcile --all` for structural contract drift instead of duplicating reconciliation
- Remains advisory and docs-only; it does not rerun `kit init` or block lifecycle commands

### Init Refresh

The structural refresh flow for updating Kit-managed project files to the current Kit defaults:

- Invoked with `kit init --refresh`
- Creates missing Kit-managed scaffold files, instruction docs, support docs, and known registry rulesets
- Creates or refreshes the Kit-managed `.github/workflows/auto-assign.yml` workflow from project-local `github.default_assignees`, falling back to global `~/.config/kit/.kit.yaml`, and rendering a non-blocking no-op when no assignees are configured
- Merges or appends missing Kit-managed documentation sections by default instead of overwriting project-specific content
- Adopts existing registry rulesets into `.kit.yaml` registry state and syncs safe upstream ruleset updates from the Kit GitHub `main` branch
- Preserves local ruleset `status` while comparing registry content, because activation and silencing are project-local choices
- Skips and reports local-custom or conflicted rulesets instead of writing conflict markers or silently overwriting project guidance
- Migrates old verbose repository instruction files to the v2 thin ToC/RLM model when they still match known generated templates
- Uses `kit init --refresh --dry-run --diff` to preview managed-file changes without writing them
- Uses `kit init --refresh --force` for generated documentation overwrites and for accepting latest registry ruleset content while preserving local ruleset status
- Uses `kit init --refresh --file=<path> --force` for targeted overwrites such as `.envrc` or `docs/references/rules/<slug>.md`

### Map

The read-only structural view rendered by `kit map`:

- Opens an interactive feature selector by default and shows the full project view with `kit map --all`
- Shows global docs, feature docs, lifecycle state, and explicit feature-to-feature relationships
- Derives its state from canonical markdown docs and the filesystem
- Normalizes harmless inline-code formatting around relationship targets and warns on malformed lines instead of failing the whole map
- May color terminal output for scanability, while keeping non-TTY output plain
- Does not create another persisted graph document in the repository

---

## ARCHITECTURAL PATTERNS

### Package Organization

```bash
kit/
├── cmd/kit/main.go          # thin entry point
├── pkg/cli/                 # public CLI commands, prompt builders, and human output
│   ├── root*.go             # root command, banner, help, profiles
│   ├── brainstorm*.go       # brainstorm, backlog capture, notes, prompts
│   ├── spec*.go             # specification workflow and interactive inputs
│   ├── plan*.go             # plan workflow and prompt generation
│   ├── tasks*.go            # task workflow and prompt generation
│   ├── implement*.go        # readiness-gated implementation prompts
│   ├── reflect.go           # reflection prompt and refresh advisory
│   ├── status*.go           # active and project-wide lifecycle views
│   ├── map.go               # document map and relationship rendering
│   ├── reconcile*.go        # structural doc drift audit and prompt
│   ├── loop*.go             # workflow, review, and prompt loop surfaces
│   ├── pr*.go               # PR review-feedback command aliases
│   ├── prompt*.go           # prompt library, profiles, IR helpers, output wrappers
│   ├── scaffold.go          # empty workflow document structure scaffolding
│   ├── scaffold_agents*.go  # repository instruction scaffolding under kit scaffold agents
│   ├── complete.go          # completion lifecycle command
│   ├── pause.go             # pause lifecycle command
│   ├── remove.go            # remove lifecycle command
│   ├── resume.go            # canonical resume routing
│   ├── dispatch*.go         # task clustering and subagent dispatch prompt
│   ├── handoff*.go          # handoff prompt and doc-sync guidance
│   ├── summarize.go         # summarization prompt
│   ├── skill*.go            # skill mining prompts
│   ├── upgrade*.go          # self-upgrade support
│   └── version.go           # version command
├── internal/
│   ├── config/              # .kit.yaml loading, project root discovery, prompt config
│   ├── document/            # Markdown parsing, metadata, relationships, validation
│   ├── feature/             # feature identity, allocator, lifecycle, map, status
│   ├── instructions/        # versioned repository instruction registry
│   ├── promptdoc/           # typed prompt document rendering
│   ├── promptlib/           # prompt library merge, normalize, resolve, suggest
│   ├── rollup/              # PROJECT_PROGRESS_SUMMARY.md generation
│   └── templates/           # embedded project and instruction templates
└── docs/
    ├── CONSTITUTION.md      # this file
    ├── PROJECT_PROGRESS_SUMMARY.md
    ├── agents/              # repo-local agent routing docs
    ├── future/              # non-binding future architecture notes
    ├── notes/               # optional feature source material; private contents ignored
    ├── references/          # durable repo references and pointer-loaded rulesets
    └── specs/               # feature directories and core spec
```

### Command Pattern

Most stateful workflow commands follow the same structure:

1. Find project root via `config.FindProjectRoot()` when project context is required
2. Load configuration via `config.Load(projectRoot)`
3. Resolve feature or project scope via feature helpers when needed
4. Perform the action: scaffold, validate, render, prompt, or mutate lifecycle docs
5. Update rollup only when feature or project summary state can change
6. Output next steps, validation results, or agent prompts

Prompt-only and inspection commands may skip feature resolution, document writes, or rollup updates by design.

### Error Handling Pattern

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create %s: %w", path, err)
}

// Suggest fixes in error messages
return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first or use --force", slug)

// Use warnings for non-fatal issues
fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
```

### Document Validation Pattern

```go
doc, err := document.ParseFile(path, document.TypeSpec)
if err != nil {
    return err
}
for _, e := range doc.Validate() {
    errors = append(errors, e.Error())
}
if doc.HasUnresolvedPlaceholders() {
    warnings = append(warnings, "has unresolved TODO placeholders")
}
```

---

## CODE STYLE CONVENTIONS

### Naming

- **Packages**: lowercase, single word (`config`, `document`, `feature`)
- **Files**: lowercase, descriptive (`config.go`, `document.go`)
- **Types**: PascalCase (`Config`, `Feature`, `Document`)
- **Functions**: PascalCase for exported, camelCase for internal
- **Variables**: camelCase (`projectRoot`, `specsDir`)
- **Constants**: PascalCase for exported, camelCase for internal

### Go Idioms

- Use `%w` for error wrapping
- Return errors, don't panic
- Prefer early returns over deep nesting
- Use `filepath.Join()` for paths
- Use `strings.Builder` for string concatenation

### CLI Conventions

- Use emoji sparingly for visual feedback (🎒 📁 ✓ ⚠ ❌ ✅ 📋 📝 🔎 📊 🔍)
- Provide copy-pasteable prompts for coding agents
- Show next steps after every command
- Support `--help` on all commands

---

## DEPENDENCIES

| Dependency                   | Purpose                             | Version |
| ---------------------------- | ----------------------------------- | ------- |
| `github.com/chzyer/readline` | Interactive terminal input support  | v1.5.1  |
| `github.com/spf13/cobra`     | CLI framework                       | v1.10.2 |
| `golang.org/x/term`          | Terminal capability and TTY helpers | v0.41.0 |
| `gopkg.in/yaml.v3`           | YAML parsing for `.kit.yaml`        | v3.0.1  |

### Why These Dependencies?

- **Cobra**: Industry standard for Go CLIs. Provides subcommands, flags, help generation.
- **YAML v3**: Required for `.kit.yaml` configuration. v3 has better error messages.
- **readline / x/term**: Support interactive terminal flows, editor gates, and terminal-aware output.

### Dependency Constraints

Kit intentionally keeps direct dependencies small:

- No database drivers
- No service-specific SDKs
- No testing frameworks beyond stdlib
- Transitive dependencies are accepted only through direct dependencies that serve current CLI needs

---

## CONFIGURATION REFERENCE

### `.kit.yaml` Full Schema

The same top-level schema is used for project-local `.kit.yaml` and global
`~/.config/kit/.kit.yaml`. `kit init` populates both locations, while command
state such as `feature_state` remains project-local in practice.

```yaml
# Understanding threshold percentage (0-100)
goal_percentage: 95

# Location of feature specs relative to project root
specs_dir: docs/specs
# Location of reusable agent skills relative to project root
skills_dir: .agents/skills

# Location of constitution file
constitution_path: docs/CONSTITUTION.md

# If true, kit legacy plan/tasks create missing prerequisites
allow_out_of_order: false
# Autonomous workflow loop policy and local agent command
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
      - gpt-5.5
      - --sandbox
      - workspace-write
      - --ignore-user-config
      - --color
      - never
      - "-"
# Repository instruction scaffold model
instruction_scaffold_version: 2

# Agent pointer files to scaffold on kit init
agents:
  - AGENTS.md
  - CLAUDE.md
  - .github/copilot-instructions.md

# GitHub integration defaults
github:
  repository: owner/repo
  default_branch: main
  # Project-local assignees take precedence over global config.
  # Omit or set null to fall back to ~/.config/kit/.kit.yaml.
  # Set [] to make the generated workflow no-op.
  default_assignees:
    - github-login
# Feature lifecycle state
feature_state:
  0001-feat-name:
    paused: false
removed_features:
  - number: 1
    slug: feat-name
    dir_name: 0001-feat-name
    created_at: "2026-01-01"
    removed_at: "2026-01-02"

# Feature directory naming
feature_naming:
  numeric_width: 4 # Pads to 0001, 0002, etc.
  separator: '-' # Between number and slug

# Local prompt library entries
prompts:
  coding-agent:
    short:
      content: |
        Clarify the task, inspect the codebase, then propose the next change.
      description: Short coding-agent planning prompt
```

Prompt library rules:

- Project-local `.kit.yaml` entries have highest precedence.
- Global prompt entries live in `~/.config/kit/.kit.yaml` and override built-ins when no local prompt exists.
- `kit init` creates or updates the global config with missing default fields without replacing existing prompt entries.
- Built-in prompts have lowest precedence.
- Prompt identities normalize to lowercase kebab-case.
- Prompt entries require `content` and may include `description`.
- Unknown prompt metadata must not break reads.
- v0 does not support `--source`, `--no-copy`, auto-paste, clipboard restore, stdin setters, or file setters.

---

## LONG-TERM VISION

### Kit 1.0 (Current)

- ✅ Document-centered workflow
- ✅ Feature lifecycle management
- ✅ Front-matter-first metadata with legacy body fallback
- ✅ Prompt library with built-in, global, and project-local prompt precedence
- ✅ Agent portability via versioned pointer files and repo-local routing docs
- ✅ Validation, reconciliation, mapping, and rollup generation
- ✅ Context resume, handoff, summarization, review, and dispatch prompts
- ✅ Soft project refresh advisory for mature repositories

### Kit 1.x (Near-term)

- [ ] Template customization via `.kit/templates/`
- [ ] Plugin system for custom validators
- [ ] Multi-language support for agent prompts
- [ ] Integration with common editors (VS Code, Neovim)
- [ ] Formalize selected future architecture notes from `docs/future/` into normal feature specs

### Kit 2.0 (Future)

- [ ] Team collaboration features
- [ ] Specification diffing and versioning
- [ ] AI-assisted spec completion (opt-in)
- [ ] Metrics and insights (local only, no telemetry)

### V1 Next-Generation Direction

`docs/future/V1_NEXT_GEN.md` is non-binding until converted into normal feature specs.
It records a possible evolution from static prompt generation toward a local intent and alignment runtime:

- compact intent contracts for current work sessions
- provider-neutral event streams for observed agent activity
- deterministic policy checks before model-backed supervision
- narrow interventions such as context injection, nudges, pauses, or blocks
- host adapters for coding-agent CLIs, desktop agents, MCP-capable clients, Warp, and PTY fallback
- visible local runtime files that remain inspectable and removable

Current Kit workflows remain document-first and prompt-centered until a formal feature spec changes the shipped product contract.

### Guiding Principle

Kit will always prioritize:

1. Documents over tools
2. Simplicity over features
3. Portability over convenience
4. Explicit over magic

The CLI should become unnecessary once understanding is achieved.

---

## SUCCESS CRITERIA

Kit is successful if:

1. **Documents remain readable without Kit** — Any markdown viewer works
2. **Agents can be swapped with zero document changes** — No lock-in
3. **Teams reach clarity faster with fewer reworks** — Measurable improvement
4. **The CLI becomes unnecessary once understanding is achieved** — Tool disappears

---

## MAINTENANCE

This constitution should be updated when:

- Core principles change (rare)
- New patterns emerge from usage
- Constraints prove too restrictive or too loose
- Definitions need clarification

Last reviewed: 2026-05-17
