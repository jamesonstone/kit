# GitHub Copilot Repository Instructions

## Fast rules for chat and code review

- classify every request first
  - **spec-driven**: `kit brainstorm` / `kit spec` work, new capability, substantial behavioral or architectural change, existing spec-covered feature work, or cross-component/public-interface changes
  - **ad hoc**: contained bug fix, review, refactor, dependency update, config change, or small refinement
- for spec-driven work:
  - read `BRAINSTORM.md` when present, then `SPEC.md` → `PLAN.md` → `TASKS.md`
  - ask numbered clarification questions until you reach ≥95% confidence
  - include a recommended default, proposed solution, or assumption for every question
  - accept approvals via `yes` / `y`, partial approvals via `yes 3, 4`, and overrides via `no 2: <answer>`
  - implement tasks in order and update docs first if implementation changes behavior, requirements, or approach
- for ad hoc work:
  - follow understand → implement → verify
  - update existing spec docs when the change alters behavior, requirements, or approach
- always:
  - never mix multiple features in one `docs/specs/<feature>/` directory
  - keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with canonical docs
  - when creating a `git worktree`, use the flat `~/worktrees/` root with repo-prefixed leaf directories such as `~/worktrees/<repo>-<branch>`
  - prefer readable, maintainable code with explicit error handling and focused functions
  - fix all lint and test failures before completion and wait for the user's output before triaging findings they report
  - do NOT run `coderabbit --prompt-only`, `git add`, or `git commit` without explicit approval

---

## Source of truth

- Primary authority for repository workflow, constraints, and change policy: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `BRAINSTORM.md` (optional research)
  - `SPEC.md` (requirements)
  - `PLAN.md` (implementation plan)
  - `TASKS.md` (executable task list)
  - `ANALYSIS.md` (optional, analysis scratchpad)
- Keep repository instruction files aligned with the canonical docs: `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`

---

## Change Classification (Required First Step)

Classify each request before implementation.

### 1) Spec-Driven (Formal Track)

Use when any apply:

- request initiated through `kit brainstorm` or `kit spec`
- new feature or capability
- substantial architectural or behavioral change
- work touches code with existing feature specs under `docs/specs/<feature>/`
- changes cross component boundaries or public interfaces

Required flow:

- follow optional research + artifact pipeline: `BRAINSTORM.md` → `SPEC.md` → `PLAN.md` → `TASKS.md` → implementation → reflection

### 2) Ad Hoc (Lightweight Track)

Use when all apply:

- not initiated through `kit brainstorm` or `kit spec`
- bug fix, security review, refactor, dependency update, config change, or small refinement
- scope is contained and can be verified directly

Required flow:

- understand → implement → verify
- update only relevant practical docs (README/API docs/inline docs) when needed
- do not create spec artifacts for ad hoc work by default

### 3) Ad Hoc + Existing Feature Specs

If ad hoc work touches a feature with existing specs:

- default to updating `SPEC.md` / `PLAN.md` / `TASKS.md` when behavior, requirements, or approach changes
- skip spec updates only for mechanical edits (formatting, typo, dependency bump)

## Multi-feature rule

- Never mix features in one `docs/specs/<feature>/` directory.
- If work spans features, update each feature's docs separately.

---

## Communication Style

- Answer first; no preamble
- Short sentences, zero fluff, no exclamations
- Use numbered lists for questions or clarification
- Include only highest-signal facts; omit obvious context
- Use tight bullets when listing
- Code: production-ready, minimal comments, no extra text
- Numbers > simple claims: quantify, compare, give thresholds
- End with a concise TL;DR when appropriate

---

## Workflow: Plan → Act → Reflect (Spec-Driven Track)

### Phase 1: PLAN

- Locate the relevant feature directory in `docs/specs/<feature>/`
- Read `BRAINSTORM.md` when present, then `SPEC.md` → `PLAN.md` → `TASKS.md`

- Ask clarifying questions until you reach ≥95% confidence that you understand the problem and desired solution
- Use numbered lists
- Ask questions in batches of up to 10
- For every question, include your current best recommended default, proposed solution, or assumption
- State uncertainties
- Accept lean approvals for the current batch:
  - `yes` / `y` approves all recommended defaults in the batch
  - `yes 3, 4, 5` / `y 3, 4, 5` approves only those numbered defaults in the batch
  - `no 2: <answer>` / `n 2: <answer>` rejects a numbered default and provides the override
  - `no` / `n` rejects all recommended defaults in the batch and requires explicit replacements before proceeding
- Treat all unapproved questions in a batch as unresolved
- After each batch of up to 10 questions, output your current percentage understanding so the user can see progress
- After each batch, reassess and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution

- Identify ambiguities, missing context, edge cases, and failure modes
- Reference existing codebase structure and patterns
- Design solution approaches aligned with existing conventions
- Consider dependencies, impacts, backward compatibility, and integration points
- Include measurable constraints when relevant:
  - latency, throughput, memory, query count, cost, limits
- Present strategy for approval before proceeding

### Phase 2: ACT

- Implement tasks strictly in order from `TASKS.md`
- Follow all code style guidelines and architectural standards
- Ensure explicit error handling and input validation
- Add or update tests required by the plan
- Provide a final summary of all files changed

### Phase 3: REFLECT

- Verify using tests and validation steps defined in the plan
- Confirm correctness, edge cases, and failure handling
- Ensure code is formatted, linted, and tested
- Fix all lint and test failures before completion, including failures outside the immediate scope
- Review changes using `git diff` against the approved plan
- Do NOT run `coderabbit --prompt-only` unless the user explicitly asks for it or explicitly approves it first
- If implementation diverges from specs:
  - update `SPEC.md` / `PLAN.md` / `TASKS.md` first
  - then update code

---

## Quality gate policy

- Always leave the project in a working state.
- Fix all lint and test failures before completion, including failures outside the immediate scope.
- Wait for the user's output before triaging or fixing findings.

---

## Workflow: Understand → Implement → Verify (Ad Hoc Track)

### Phase 1: UNDERSTAND

- Confirm scope and constraints directly from request + code context
- Identify impacted files and risks
- If feature specs exist for impacted behavior, default to updating them

### Phase 2: IMPLEMENT

- Apply focused code changes
- Keep changes minimal and reversible
- Preserve existing architecture and style constraints

### Phase 3: VERIFY

- Run the smallest relevant validation (build/tests/lint as applicable)
- Confirm no unintended behavior changes
- Update only relevant practical docs when behavior or usage changes

---

## Definition of Done (DoD)

A change is done when all applicable conditions are met for its track.

### Spec-Driven DoD

- Requirements satisfied per `SPEC.md`
- Code implemented per `PLAN.md` and `TASKS.md`
- Tests added or updated for changed behavior
- Observability updated when applicable:
  - logs, metrics, traces
- Security checklist reviewed if inputs, auth, or data storage changed
- Migrations and rollback plan documented if data model changed
- Relevant documentation updated

### Ad Hoc DoD

- Requested change implemented and validated
- Existing specs updated when required by change classification
- Relevant practical docs updated only when behavior/usage changed
- No unnecessary artifact creation

---

## Code Style Standards

- ALWAYS use lowercase letters at the beginning of comments
- DO NOT add punctuation at the end of comments unless it improves readability
- Use comments sparingly and only where necessary
- Use docstrings/comments only for public or externally consumed APIs
- REST APIs: OpenAPI/Swagger is the primary documentation source
- Prefer self-explaining code over explanatory comments
- Use TODO/FIXME/NOTE prefixes only with context
- Default line length: 100 characters (max 120 when unavoidable)
- Use language idioms appropriately

---

## Output Requirements

- Output full functions and classes in complete fidelity
- Optimize for readability and maintainability over cleverness
- Use descriptive names for variables, functions, classes, parameters, methods
- Keep functions focused on a single responsibility
- Prefer explicit error handling over silent failures
- Soft file size limit: 200 lines
- Hard file size limit: 300 lines (exceptions require justification)
- Split files that exceed limits

---

## Architecture & Structure

- Separate transport, domain, and infrastructure concerns
- Dependencies must point inward toward domain logic
- Prefer composition over inheritance
- Avoid deep nesting (max 3–4 levels)
- Extract complex logic into well-named helper functions
- Separate business logic from framework or I/O code
- Organize by feature or domain, not by technical type

---

## Package Structure Guidelines

- Keep files under 300 lines when possible
- One file per resource type, domain object, or logical grouping
- Place shared/common types in a dedicated `types.go`
- Place service definitions, constructors, and shared options in main package files
- Co-locate related elements:
  - type definitions
  - helper methods
  - constants
  - service logic
- Use noun-based file names (`conditions.go`, not `list_conditions.go`)

---

## Testing Standards

- Unit tests for domain and business logic
- Integration tests for databases and external dependencies
- Contract tests for APIs and public interfaces
- Property or fuzz tests for parsers and validators when applicable
- Focus on diff coverage:
  - ≥80%% of changed lines should be covered by tests

---

## Performance & SLOs

- Define performance expectations when relevant:
  - latency p50 / p95 / p99
  - throughput
  - memory and CPU bounds
- Enforce query and call budgets on hot paths
- Measure before and after when modifying critical paths

---

## Build and Development Standards

- Use appropriate build automation (Makefile, package.json scripts, etc.)
- Standard commands:
  - build, run, test, clean, lint, format
- Include database commands when applicable:
  - db-start, db-migrate, db-seed, db-reset
- Support containerization with Docker when appropriate

---

## Documentation Standards

- Use OpenAPI/Swagger specifications for all REST APIs
- Maintain a comprehensive README.md:
  - setup
  - usage
  - contribution guidelines
- Document architectural decisions in `/docs/adr`
- Use CommonMark for all markdown files
- Document required environment variables and configuration

---

## Logging and Monitoring

- Use structured logging with consistent fields:
  - event, level, component, trace_id
- Add Emojis as log-statement-prefixes to improve readability
- Use correlation IDs for request tracing
- Never log secrets, tokens, or PII

---

## Error Handling & Security

- Use specific error or exception types
- Standard error mapping when applicable:
  - validation → 400
  - auth/authz → 401/403
  - not found → 404
  - conflict → 409
  - rate limit → 429
  - downstream failure → 502/503
  - invariant violation → 500
- Validate and sanitize all external inputs
- Use secure defaults and fail securely
- No hardcoded secrets

---

## Git Rules

- **NEVER** run `git add` or `git commit` without user approval
- Use conventional commit messages with "gitmojis" in the title to improve commit message readability
- When creating a `git worktree`, use `git worktree add ~/worktrees/<repo>-<branch> <branch>` or `git worktree add -b <branch> ~/worktrees/<repo>-<branch> <base-ref>`
- Keep the `~/worktrees/` directory flat across all projects; do NOT create worktrees inside the repository tree or nested per-project directories

---

## Document Management

- Do NOT generate periodic progress-tracking documents
- Update the active specification documents instead
- Do NOT create separate summary documents

---

## State Summarization

Trigger: `summarize-state` or `/compact state`

Produce two outputs:

### A) CURRENT_STATE.md

- Human-readable
- Complete
- Source of truth

### B) AGENT_CONTEXT

- ≤400 tokens
- Facts only
- Optimized for prompt injection

Required sections (in order):

1. SYSTEM OVERVIEW
2. CURRENT ARCHITECTURE
3. DIRECTORY STRUCTURE (EXISTS TODAY)
4. IMPLEMENTED (PRODUCTION-READY)
5. STUBBED / NOT IMPLEMENTED
6. API CONTRACTS & EXTERNAL DEPENDENCIES
7. INVARIANTS (MUST NOT CHANGE)
8. SAFE TO CHANGE
9. NEXT PHASE GOALS
10. OUT OF SCOPE
