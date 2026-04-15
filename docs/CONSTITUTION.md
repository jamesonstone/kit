# CONSTITUTION

Kit is a general-purpose harness for disciplined thought work.
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
- Agents can be swapped with zero document changes.

### 4. Minimal Magic, Explicit State

- Prefer explicit over implicit behavior.
- No hidden databases, lock files, or external state.
- All state lives in the filesystem (markdown files + `.kit.yaml`).
- Commands fail fast with actionable error messages.

### 5. Opinionated Defaults, Configurable Escapes

- Sensible defaults work out of the box.
- Configuration overrides via `.kit.yaml` for team customization.
- CLI flags always override configuration.
- Escape hatches exist but aren't encouraged.

### 6. Tooling Should Disappear

- Kit's job is done once documents are complete and correct.
- Implementation happens outside Kit's scope.
- The CLI becomes unnecessary once understanding is achieved.
- Teams should reach clarity faster with fewer reworks.

### 7. Density Over Prose

- Documents prioritize brevity and precision.
- One sentence where possible.
- Bullet points over paragraphs.
- Code snippets only when unavoidable.

---

## CONSTRAINTS

### Non-Negotiable Rules

1. **Artifact Pipeline Order**
   - Constitution → Brainstorm (optional) → Specification → Plan → Tasks → Implementation → Reflection
   - Each artifact gates the next (unless `--force` or `allow_out_of_order: true`)
   - Promotion from Tasks to Implementation requires an implementation readiness gate
   - The readiness gate adversarially challenges `CONSTITUTION.md`, optional `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, and scope creep before code begins
   - If the readiness gate fails, update the canonical docs first, then restart implementation
   - Breaking this order requires explicit intent

2. **Document Structure**
   - All canonical markdown files use FULL CAPITALIZATION (e.g., `CONSTITUTION.md`, `SPEC.md`)
   - Use snake_case for multi-word file names (e.g., `PROJECT_PROGRESS_SUMMARY.md`)
   - Use kebab-case for directories (e.g., `0001-feat-name`)
   - Feature directories use numeric prefix + slug format

3. **Section Requirements**
   - Each document type has required sections that must be present
   - Required sections in `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md` must also be populated
   - Do not leave HTML TODO comments as the only content in a required section
   - If a required section has no feature-specific detail, replace the placeholder comment with `not applicable`, `not required`, or `no additional information required`
   - `CONSTITUTION.md`: PRINCIPLES, CONSTRAINTS, NON-GOALS, DEFINITIONS
   - `BRAINSTORM.md`: SUMMARY, USER THESIS, RELATIONSHIPS, CODEBASE FINDINGS, AFFECTED FILES, DEPENDENCIES, QUESTIONS, OPTIONS, RECOMMENDED STRATEGY, NEXT STEP
   - `SPEC.md`: SUMMARY, PROBLEM, GOALS, NON-GOALS, USERS, SKILLS, RELATIONSHIPS, DEPENDENCIES, REQUIREMENTS, ACCEPTANCE, EDGE-CASES, OPEN-QUESTIONS
   - `PLAN.md`: SUMMARY, APPROACH, COMPONENTS, DATA, INTERFACES, RISKS, TESTING
   - `TASKS.md`: PROGRESS TABLE, TASK LIST, TASK DETAILS, DEPENDENCIES, NOTES
- `RELATIONSHIPS` in `BRAINSTORM.md` and `SPEC.md` must be either `none` or explicit bullets using `builds on: <feature>`, `depends on: <feature>`, or `related to: <feature>`
- inline-code-wrapped relationship targets are valid if they normalize back to a canonical feature directory identifier

4. **Single Feature Per Directory**
   - Never mix features in one `docs/specs/<feature>/` directory
   - If work spans features, update each feature's docs separately
   - Feature directories are immutable once created

5. **No Implementation Details in Specs**
   - `SPEC.md` defines WHAT, not HOW
   - No code in specifications
   - No technology choices unless absolutely required

6. **Traceability**
   - Tasks link to plan items using `[PLAN-XX]` syntax
   - Plan items link to spec items using `[SPEC-XX]` syntax
   - Every claim in `PROJECT_PROGRESS_SUMMARY.md` must map to a feature document

7. **External Review Tools**
   - Do NOT run `coderabbit --prompt-only` unless the user explicitly asks for it or explicitly approves it first

8. **Git Worktrees**
   - When isolated checkouts are needed, create worktrees under `~/worktrees/`
   - Use a flat repo-prefixed leaf path such as `~/worktrees/<repo>-<branch>`
   - Never create worktrees inside the repository tree or nested per-project directories

### Code Quality Constraints

1. **Go Best Practices**
   - Single binary with subcommands (Cobra CLI)
   - No global state beyond package-level constants
   - Error handling must be explicit (`%w` for wrapped errors)
   - Test coverage for critical paths

2. **Package Structure**
   - `cmd/kit/` — main entry point only
   - `pkg/cli/` — command implementations
   - `internal/` — private packages (config, document, feature, rollup, templates)
   - No circular dependencies

3. **Error Messages**
   - Must be actionable (suggest fixes)
   - Include context (file paths, feature names)
   - Fail fast — don't continue with partial state

---

## CHANGE CLASSIFICATION

All work falls into one of two tracks. Classify before acting.

### Spec-Driven (Formal)

Use when ANY of these apply:

- Initiated via `kit brainstorm` or `kit spec`
- New feature or capability
- Substantial architectural or behavioral change
- Work that has existing spec docs in `docs/specs/<feature>/`
- Change affects multiple components or public interfaces

**Workflow**: Optional research + artifact pipeline — BRAINSTORM.md → SPEC.md → PLAN.md → TASKS.md → implement → reflect

**Clarification protocol for formal planning phases**:

- Ask in numbered batches of up to 10 questions with a recommended default for each question
- `yes` / `y` approves all recommended defaults in the current batch
- `yes 3, 4, 5` / `y 3, 4, 5` approves only those numbered defaults
- `no 2: <answer>` / `n 2: <answer>` overrides a numbered default
- `no` / `n` rejects the full batch and requires explicit replacements before proceeding

### Ad Hoc (Lightweight)

Use when ALL of these apply:

- Not initiated via `kit brainstorm` or `kit spec`
- Bug fix, security review, refactor, dependency update, config change, or small refinement
- Scope is contained and well-understood without formal specification

**Workflow**: Understand → implement → verify

**Documentation**: Update only practical docs (READMEs, inline docs, API docs). Do NOT create SPEC.md / PLAN.md / TASKS.md for ad hoc work.

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

- ❌ Execute code or run tests
- ❌ Manage agents directly or maintain prompt registries
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

### Artifact Pipeline

The ordered sequence of documents that drive development:

1. **Constitution** — Project-wide constraints, principles, long-term vision. Kept updated with priors. Single per repository.

2. **Brainstorm (BRAINSTORM.md)** — Optional codebase-aware research. Captures findings, affected files, open questions, and recommended strategy.

3. **Specification (SPEC.md)** — What is being built and why. Requirements, acceptance criteria, edge cases. No implementation details.

4. **Plan (PLAN.md)** — How it will be built. Strategy, components, interfaces, risks. Explains approach, not code.

5. **Tasks (TASKS.md)** — Atomic executable work units. Maps to plan items. Reflects real progress.

6. **Implementation** — Code execution. Outside Kit's core scope.
   - `kit implement` begins with an implementation readiness gate before code execution starts.

7. **Reflection** — Verify correctness, refine understanding. Loops back to specification if needed.

### Feature

A self-contained unit of work with its own directory under `docs/specs/`. Identified by:

- **Numeric prefix**: Auto-assigned sequential number (e.g., `0001`)
- **Slug**: Lowercase kebab-case name, max 5 words (e.g., `init-project`)
- **Directory**: Combined format (e.g., `0001-init-project`)

### Phase

The current state of a feature in the artifact pipeline:

- `brainstorm` — Has BRAINSTORM.md without SPEC.md
- `spec` — Has SPEC.md only, optionally preceded by BRAINSTORM.md
- `plan` — Has SPEC.md and PLAN.md, optionally preceded by BRAINSTORM.md
- `tasks` — Has SPEC.md, PLAN.md, and TASKS.md, optionally preceded by BRAINSTORM.md
- `implement` — Beyond Kit's scope

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

### Map

The read-only structural view rendered by `kit map`:

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
├── cmd/kit/main.go          # Entry point (5 lines)
├── pkg/cli/                  # Public CLI commands
│   ├── root.go              # Root command, banner, colors
│   ├── init.go              # kit init
│   ├── brainstorm.go        # kit brainstorm [feature]
│   ├── spec.go              # kit spec <feature>
│   ├── plan.go              # kit plan <feature>
│   ├── tasks.go             # kit tasks <feature>
│   ├── check.go             # kit check [feature]
│   ├── rollup.go            # kit rollup
│   ├── reconcile.go         # kit reconcile [feature]
│   ├── handoff.go           # kit handoff [feature]
│   ├── summarize.go         # kit summarize [feature]
│   ├── reflect.go           # kit reflect [feature]
│   ├── map.go               # kit map [feature]
│   └── scaffold_agents.go   # kit scaffold-agents
├── internal/
│   ├── config/config.go     # .kit.yaml loading, project root discovery
│   ├── document/document.go # Markdown parsing, validation, section extraction
│   ├── feature/feature.go   # Feature numbering, slug validation, directory management
│   ├── rollup/rollup.go     # PROJECT_PROGRESS_SUMMARY.md generation
│   └── templates/templates.go # Embedded document templates
└── docs/
    ├── CONSTITUTION.md      # This file
    ├── PROJECT_PROGRESS_SUMMARY.md
    └── specs/               # Feature directories
```

### Command Pattern

Each command follows the same structure:

1. Find project root via `config.FindProjectRoot()`
2. Load configuration via `config.Load(projectRoot)`
3. Resolve feature via `feature.Resolve()` or `feature.EnsureExists()`
4. Perform action (create/validate documents)
5. Update rollup via `rollup.Update()`
6. Output next steps and agent prompts

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

| Dependency               | Purpose                    | Version |
| ------------------------ | -------------------------- | ------- |
| `github.com/spf13/cobra` | CLI framework              | v1.10.2 |
| `gopkg.in/yaml.v3`       | YAML parsing for .kit.yaml | v3.0.1  |

### Why These Dependencies?

- **Cobra**: Industry standard for Go CLIs. Provides subcommands, flags, help generation.
- **YAML v3**: Required for `.kit.yaml` configuration. v3 has better error messages.

### No Additional Dependencies

Kit intentionally keeps dependencies minimal:

- No database drivers
- No HTTP clients
- No external services
- No testing frameworks beyond stdlib

---

## CONFIGURATION REFERENCE

### `.kit.yaml` Full Schema

```yaml
# Understanding threshold percentage (0-100)
goal_percentage: 95

# Location of feature specs relative to project root
specs_dir: docs/specs

# Location of constitution file
constitution_path: docs/CONSTITUTION.md

# If true, kit plan/tasks create missing prerequisites
allow_out_of_order: false

# Agent pointer files to scaffold on kit init
agents:
  - AGENTS.md
  - CLAUDE.md
  - .github/copilot-instructions.md

# Feature directory naming
feature_naming:
  numeric_width: 4 # Pads to 0001, 0002, etc.
  separator: '-' # Between number and slug
```

---

## LONG-TERM VISION

### Kit 1.0 (Current)

- ✅ Document-centered workflow
- ✅ Feature lifecycle management
- ✅ Agent portability via pointer files
- ✅ Validation and rollup generation
- ✅ Context handoff between agents

### Kit 1.x (Near-term)

- [ ] Template customization via `.kit/templates/`
- [ ] Plugin system for custom validators
- [ ] Multi-language support for agent prompts
- [ ] Integration with common editors (VS Code, Neovim)

### Kit 2.0 (Future)

- [ ] Team collaboration features
- [ ] Specification diffing and versioning
- [ ] AI-assisted spec completion (opt-in)
- [ ] Metrics and insights (local only, no telemetry)

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

Last reviewed: 2026-01-19
