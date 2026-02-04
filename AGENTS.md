# AGENTS

## Kit is the source of truth

- Constitution: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `SPEC.md` (requirements)
  - `PLAN.md` (implementation plan)
  - `TASKS.md` (executable task list)
  - `ANALYSIS.md` (optional, analysis scratchpad)

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

## Workflow: Plan → Act → Reflect

Specs drive code. Code serves specs.

### Phase 1: PLAN

- Locate the relevant feature directory in `docs/specs/<feature>/`
- Read `SPEC.md` → `PLAN.md` → `TASKS.md`

- Ask clarifying questions until requirements, constraints, and success criteria are explicit
  - Clarification questions must be clearly labeled as one of:
    - **Fact-finding** (inputs, outputs, constraints, invariants)
    - **Decision-required** (tradeoffs the user must choose)
  - When appropriate, explicitly include the agent's **preferred solution** as **one option**, clearly labeled:
    - state assumptions
    - explain why it is preferred (performance, simplicity, safety, cost)
    - present alternatives if they are viable
  - Do not assume acceptance of the preferred solution without confirmation

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
- Review changes using `git diff` against the approved plan
- If CodeRabbit is available, run `coderabbit --prompt-only` and fix issues
- If implementation diverges from specs:
  - update `SPEC.md` / `PLAN.md` / `TASKS.md` first
  - then update code

---

## Definition of Done (DoD)

A feature or task is done only when all apply:

- Requirements satisfied per `SPEC.md`
- Code implemented per `PLAN.md` and `TASKS.md`
- Tests added or updated for changed behavior
- Observability updated when applicable:
  - logs, metrics, traces
- Security checklist reviewed if inputs, auth, or data storage changed
- Migrations and rollback plan documented if data model changed
- Relevant documentation updated

---

## Code Style Standards

- Use lowercase letters at the beginning of comments
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
