// package templates provides embedded document templates for Kit.
package templates

// Constitution template per spec section 6.1
const Constitution = `# CONSTITUTION

## PRINCIPLES

<!-- TODO: define core principles that guide all decisions -->

## CONSTRAINTS

<!-- TODO: define invariant rules that must never be violated -->

## CHANGE CLASSIFICATION

<!-- all work falls into one of two tracks — classify before acting -->

### Spec-Driven (Formal)

<!-- use when: new features, kit spec/oneshot, substantial architectural or behavioral changes -->
<!-- workflow: full artifact pipeline — SPEC.md → PLAN.md → TASKS.md → implement → reflect -->

### Ad Hoc (Lightweight)

<!-- use when: bug fixes, security reviews, refactors, dependency updates, config changes, small refinements -->
<!-- workflow: understand → implement → verify -->
<!-- docs: update only practical docs (READMEs, inline docs, API docs) -->
<!-- do NOT create SPEC.md / PLAN.md / TASKS.md for ad hoc work -->

### Ad Hoc with Existing Specs

<!-- if change touches code with existing spec docs: default to updating them -->
<!-- skip spec updates only for purely mechanical changes (formatting, typo, dep bump) -->

## NON-GOALS

<!-- TODO: define what this project explicitly will not do -->

## DEFINITIONS

<!-- TODO: define key terms used throughout the project -->
`

// Spec template per spec section 6.2
const Spec = `# SPEC

## SUMMARY

<!-- TODO: 1-2 sentence business summary of this feature -->

## PROBLEM

<!-- TODO: describe the problem being solved -->

## GOALS

<!-- TODO: list what this feature must achieve -->

## NON-GOALS

<!-- TODO: list what this feature will not do -->

## USERS

<!-- TODO: identify who will use this feature -->

## REQUIREMENTS

<!-- TODO: list functional requirements -->

## ACCEPTANCE

<!-- TODO: define acceptance criteria -->

## EDGE-CASES

<!-- TODO: document edge cases and how they should be handled -->

## OPEN-QUESTIONS

<!-- TODO: list unresolved questions -->
`

// Plan template per spec section 6.3
const Plan = `# PLAN

## SUMMARY

<!-- TODO: brief overview of the implementation approach -->

## APPROACH

<!-- TODO: explain the strategy, not code -->

## COMPONENTS

<!-- TODO: list major components and their responsibilities -->

## DATA

<!-- TODO: describe data structures and storage -->

## INTERFACES

<!-- TODO: define APIs, contracts, and integration points -->

## RISKS

<!-- TODO: identify risks and mitigation strategies -->

## TESTING

<!-- TODO: describe testing strategy -->
`

// Tasks template per spec section 6.4
// IMPORTANT: tasks use markdown checkboxes for progress tracking:
//   - [ ] incomplete task
//   - [x] completed task
const Tasks = `# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | <!-- task description --> | todo | <!-- owner --> | <!-- deps --> |

## TASK LIST

Use markdown checkboxes to track completion:

- [ ] T001: <!-- task description -->

## TASK DETAILS

For each task, provide:

### T001
- **GOAL**: <!-- one sentence outcome -->
- **SCOPE**: <!-- tight bullets, no fluff -->
- **ACCEPTANCE**: <!-- concrete checks -->
- **NOTES**: <!-- only if necessary -->

## DEPENDENCIES

<!-- TODO: document task dependencies and ordering -->

## NOTES

<!-- TODO: additional context or implementation notes -->
`

// Analysis template per spec section 6.6
const Analysis = `# ANALYSIS

## UNDERSTANDING

**Current Understanding: 0%%**

<!-- understanding percentage tracked at top and bottom -->

## QUESTIONS

<!-- TODO: open questions for the user/team -->

## RESEARCH

<!-- technical investigation notes: library comparisons, performance benchmarks, compatibility findings -->

## CLARIFICATIONS

<!-- resolved questions with answers -->

## ASSUMPTIONS

<!-- documented assumptions made during analysis -->

## RISKS

<!-- identified risks or concerns -->

---

**Understanding: 0%%**
`

// ProjectProgressSummary template per spec section 6.5
const ProjectProgressSummary = `# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID | FEATURE | PATH | PHASE | CREATED | SUMMARY |
| -- | ------- | ---- | ----- | ------- | ------- |

## PROJECT INTENT

<!-- TODO: describe the overall project purpose -->

## GLOBAL CONSTRAINTS

<!-- TODO: summarize key constraints from CONSTITUTION.md -->

## FEATURE SUMMARIES

<!-- feature summaries will be generated here -->

## LAST UPDATED

<!-- timestamp updated by kit rollup -->
`

// AgentPointer returns a minimal template for agent pointer files (e.g., CLAUDE.md).
func AgentPointer(agentName string) string {
	return `# ` + agentName + `

## Source of truth

- Primary authority for repository workflow, constraints, and change policy: ` + "`docs/CONSTITUTION.md`" + `
- Feature specs live under: ` + "`docs/specs/<feature>/`" + `
  - ` + "`SPEC.md`" + ` (requirements)
  - ` + "`PLAN.md`" + ` (implementation plan)
  - ` + "`TASKS.md`" + ` (executable task list)
  - ` + "`ANALYSIS.md`" + ` (optional, analysis scratchpad)

## Workflow contract (classification-first)

- Classify every request before acting:
  - **Spec-driven**: use full pipeline for ` + "`kit spec`" + ` / ` + "`kit oneshot`" + `, new features, or substantial changes
  - **Ad hoc**: use lightweight flow for small fixes, reviews, refinements, and mechanical changes
- If ad hoc work touches an existing feature in ` + "`docs/specs/<feature>/`" + `, default to updating its spec docs when behavior, requirements, or approach changes
- For ad hoc changes, update only relevant practical docs (e.g., README/API docs) and avoid creating spec artifacts unless needed

## Multi-feature rule

- Never mix features in one ` + "`docs/specs/<feature>/`" + ` directory.
- If work spans features, update each feature's docs separately.
`
}

// AgentsMD is the comprehensive AGENTS.md template with full workflow and standards.
const AgentsMD = `# AGENTS

## Source of truth

- Primary authority for repository workflow, constraints, and change policy: ` + "`docs/CONSTITUTION.md`" + `
- Feature specs live under: ` + "`docs/specs/<feature>/`" + `
  - ` + "`SPEC.md`" + ` (requirements)
  - ` + "`PLAN.md`" + ` (implementation plan)
  - ` + "`TASKS.md`" + ` (executable task list)
  - ` + "`ANALYSIS.md`" + ` (optional, analysis scratchpad)
---

## Change Classification (Required First Step)

Classify each request before implementation.

### 1) Spec-Driven (Formal Track)

Use when any apply:

- request initiated through ` + "`kit spec`" + ` or ` + "`kit oneshot`" + `
- new feature or capability
- substantial architectural or behavioral change
- work touches code with existing feature specs under ` + "`docs/specs/<feature>/`" + `
- changes cross component boundaries or public interfaces

Required flow:

- follow full artifact pipeline: ` + "`SPEC.md`" + ` → ` + "`PLAN.md`" + ` → ` + "`TASKS.md`" + ` → implementation → reflection

### 2) Ad Hoc (Lightweight Track)

Use when all apply:

- not initiated through ` + "`kit spec`" + ` or ` + "`kit oneshot`" + `
- bug fix, security review, refactor, dependency update, config change, or small refinement
- scope is contained and can be verified directly

Required flow:

- understand → implement → verify
- update only relevant practical docs (README/API docs/inline docs) when needed
- do not create spec artifacts for ad hoc work by default

### 3) Ad Hoc + Existing Feature Specs

If ad hoc work touches a feature with existing specs:

- default to updating ` + "`SPEC.md`" + ` / ` + "`PLAN.md`" + ` / ` + "`TASKS.md`" + ` when behavior, requirements, or approach changes
- skip spec updates only for mechanical edits (formatting, typo, dependency bump)

## Multi-feature rule
## Multi-feature rule

- Never mix features in one ` + "`docs/specs/<feature>/`" + ` directory.
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

- Locate the relevant feature directory in ` + "`docs/specs/<feature>/`" + `
- Read ` + "`SPEC.md`" + ` → ` + "`PLAN.md`" + ` → ` + "`TASKS.md`" + `

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

- Implement tasks strictly in order from ` + "`TASKS.md`" + `
- Follow all code style guidelines and architectural standards
- Ensure explicit error handling and input validation
- Add or update tests required by the plan
- Provide a final summary of all files changed

### Phase 3: REFLECT

- Verify using tests and validation steps defined in the plan
- Confirm correctness, edge cases, and failure handling
- Ensure code is formatted, linted, and tested
- Review changes using ` + "`git diff`" + ` against the approved plan
- If CodeRabbit is available, run ` + "`coderabbit --prompt-only`" + ` and fix issues
- If implementation diverges from specs:
  - update ` + "`SPEC.md`" + ` / ` + "`PLAN.md`" + ` / ` + "`TASKS.md`" + ` first
  - then update code

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

- Requirements satisfied per ` + "`SPEC.md`" + `
- Code implemented per ` + "`PLAN.md`" + ` and ` + "`TASKS.md`" + `
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
- Place shared/common types in a dedicated ` + "`types.go`" + `
- Place service definitions, constructors, and shared options in main package files
- Co-locate related elements:
  - type definitions
  - helper methods
  - constants
  - service logic
- Use noun-based file names (` + "`conditions.go`" + `, not ` + "`list_conditions.go`" + `)

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
- Document architectural decisions in ` + "`/docs/adr`" + `
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

- **NEVER** run ` + "`git add`" + ` or ` + "`git commit`" + ` without user approval
- Use conventional commit messages with "gitmojis" in the title to improve commit message readability

---

## Document Management

- Do NOT generate periodic progress-tracking documents
- Update the active specification documents instead
- Do NOT create separate summary documents

---

## State Summarization

Trigger: ` + "`summarize-state`" + ` or ` + "`/compact state`" + `

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
`

// FeatureSummaryTemplate returns a template for a feature summary in PROJECT_PROGRESS_SUMMARY.md
const FeatureSummaryTemplate = `### {{.FeatureName}}

- **STATUS**: {{.Phase}}
- **INTENT**: {{.Intent}}
- **APPROACH**: {{.Approach}}
- **OPEN ITEMS**: {{.OpenItems}}
- **POINTERS**: ` + "`{{.Path}}/SPEC.md`" + `, ` + "`{{.Path}}/PLAN.md`" + `, ` + "`{{.Path}}/TASKS.md`" + `
`

// Brainstorm returns a scaffold template for brainstorming sessions.
func Brainstorm(topic string) string {
	return `# Brainstorm: ` + topic + `

## Context
- [What prompted this?]
- [Who is affected?]
- [What exists today?]

## Problem Statement
We are unable to __________ because __________.

## Goals
- [What must be achieved?]
- [What constraints exist?]

## Options Considered
1. **Option A** — [description]
2. **Option B** — [description]
3. **Do nothing** — [why this is/isn't viable]

## Open Questions
- [ ] [Question 1]
- [ ] [Question 2]

## Decision Criteria
- [What would make Option X the winner?]

## Next Steps
- [ ] [Action item]
`
}
