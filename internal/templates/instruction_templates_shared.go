package templates

const sharedRepositoryInstructionsCore = `## Source of truth

- Primary authority for repository workflow, constraints, and change policy: ` + "`docs/CONSTITUTION.md`" + `
- Feature specs live under: ` + "`docs/specs/<feature>/`" + `
  - ` + "`BRAINSTORM.md`" + ` (optional research)
  - ` + "`SPEC.md`" + ` (requirements)
  - ` + "`PLAN.md`" + ` (implementation plan)
  - ` + "`TASKS.md`" + ` (executable task list)
  - ` + "`ANALYSIS.md`" + ` (optional, analysis scratchpad)
- Keep repository instruction files aligned with the canonical docs: ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, ` + "`.github/copilot-instructions.md`" + `

---

## Change Classification (Required First Step)

Classify each request before implementation.

### 1) Spec-Driven (Formal Track)

Use when any apply:

- request initiated through ` + "`kit brainstorm`" + ` or ` + "`kit spec`" + `
- new feature or capability
- substantial architectural or behavioral change
- work touches code with existing feature specs under ` + "`docs/specs/<feature>/`" + `
- changes cross component boundaries or public interfaces

Required flow:

- follow optional research + artifact pipeline: ` + "`BRAINSTORM.md`" + ` → ` + "`SPEC.md`" + ` → ` + "`PLAN.md`" + ` → ` + "`TASKS.md`" + ` → implementation → reflection

### 2) Ad Hoc (Lightweight Track)

Use when all apply:

- not initiated through ` + "`kit brainstorm`" + ` or ` + "`kit spec`" + `
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

- Never mix features in one ` + "`docs/specs/<feature>/`" + ` directory.
- If work spans features, update each feature's docs separately.

## Document Completeness

- For ` + "`BRAINSTORM.md`" + `, ` + "`SPEC.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + `, every required section must be populated
- Do not leave HTML TODO comments as the only content in a section
- If a section has no additional detail, replace the placeholder comment with ` + "`not applicable`" + `, ` + "`not required`" + `, or ` + "`no additional information required`" + `

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
- Read ` + "`BRAINSTORM.md`" + ` when present, then ` + "`SPEC.md`" + ` → ` + "`PLAN.md`" + ` → ` + "`TASKS.md`" + `

- Ask clarifying questions until you reach ≥95% confidence that you understand the problem and desired solution
- Use numbered lists
- Ask questions in batches of up to 10
- For every question, include your current best recommended default, proposed solution, or assumption
- State uncertainties
- Accept lean approvals for the current batch:
  - ` + "`yes`" + ` / ` + "`y`" + ` approves all recommended defaults in the batch
  - ` + "`yes 3, 4, 5`" + ` / ` + "`y 3, 4, 5`" + ` approves only those numbered defaults in the batch
  - ` + "`no 2: <answer>`" + ` / ` + "`n 2: <answer>`" + ` rejects a numbered default and provides the override
  - ` + "`no`" + ` / ` + "`n`" + ` rejects all recommended defaults in the batch and requires explicit replacements before proceeding
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

- Before writing code for spec-driven work, run an implementation readiness gate: adversarially challenge ` + "`CONSTITUTION.md`" + `, optional ` + "`BRAINSTORM.md`" + `, ` + "`SPEC.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, and scope creep. If the gate fails, update docs first, then code.
- Implement tasks strictly in order from ` + "`TASKS.md`" + `
- Follow all code style guidelines and architectural standards
- Ensure explicit error handling and input validation
- Add or update tests required by the plan
- Provide a final summary of all files changed
`
