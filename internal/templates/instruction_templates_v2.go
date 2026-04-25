package templates

func tocRepositoryInstructions(title string) string {
	return `## Purpose

- This file is a routing table, not the full manual
- Start at ` + "`docs/agents/README.md`" + `, then load only the docs needed for the current decision
- Repo-local markdown under ` + "`docs/`" + ` is the system of record

## Runtime Routing

- ` + "`docs/agents/README.md`" + ` — classify the task and choose the next document
- ` + "`docs/agents/WORKFLOWS.md`" + ` — spec-driven versus ad hoc flow
- ` + "`docs/agents/GUARDRAILS.md`" + ` — completion, safety, and hard rules
- ` + "`docs/agents/RLM.md`" + ` — just-in-time context loading when broad context would be noisy
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, worktrees, and secondary inputs

## Conditional Context

- ` + "`docs/specs/<feature>/`" + ` — active feature artifacts only
- ` + "`docs/references/README.md`" + ` — durable repo references only when relevant
- ` + "`docs/CONSTITUTION.md`" + ` — project invariants when a decision depends on them

## Repo Knowledge Map

- ` + "`docs/agents/README.md`" + ` — runtime routing index
- ` + "`docs/agents/WORKFLOWS.md`" + ` — work classification and source-of-truth semantics
- ` + "`docs/agents/RLM.md`" + ` — progressive disclosure and context budget rules
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, worktrees, and secondary global inputs
- ` + "`docs/agents/GUARDRAILS.md`" + ` — completion bar, safety rules, and validation expectations
- ` + "`docs/references/README.md`" + ` — durable repo-local references that are broader than one feature
- ` + "`docs/specs/<feature>/`" + ` — feature-level source of truth for requirements, plans, and tasks

## Constraints

- Keep ` + title + ` short and stable so it fits easily into injected context
- Put durable workflow guidance in ` + "`docs/agents/*`" + ` rather than expanding this file
- Do not add an always-loaded monolithic instruction file
`
}

const tocCopilotInstructions = `# GitHub Copilot Repository Instructions

## Quick Rules

- Use this file as a map, not the full manual
- Start with ` + "`docs/agents/README.md`" + ` and then open only the linked docs needed for the current decision
- Treat ` + "`docs/specs/<feature>/`" + ` as the feature system of record
- Use ` + "`docs/agents/RLM.md`" + ` when full-context loading would be noisy or wasteful
- Keep context minimal and source-attributed

## Runtime Routing

- ` + "`docs/agents/README.md`" + ` — classify the task and choose the next document
- ` + "`docs/agents/WORKFLOWS.md`" + ` — workflow and source-of-truth rules
- ` + "`docs/agents/GUARDRAILS.md`" + ` — hard completion and safety rules
- ` + "`docs/agents/RLM.md`" + ` — just-in-time context routing
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, worktrees, and secondary inputs

## Non-Negotiable Rules

- Repo-local docs under ` + "`docs/`" + ` are the source of truth
- Always update affected documentation and keep touched docs properly formatted
- Keep context minimal and load only the docs and files relevant to the task
- Remove dead code and unnecessary exports or public surface when they are not strictly needed
- Do not treat ` + "`.claude/skills`" + ` as canonical discovery input
- Do not add an always-loaded monolithic instruction file

## Repo Knowledge Map

- ` + "`docs/agents/README.md`" + ` — repo-local entrypoint
- ` + "`docs/agents/WORKFLOWS.md`" + ` — work classification and execution flow
- ` + "`docs/agents/RLM.md`" + ` — progressive-disclosure pattern for broad discovery
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, worktrees, and secondary globals
- ` + "`docs/agents/GUARDRAILS.md`" + ` — hard rules and completion bar
- ` + "`docs/references/README.md`" + ` — durable repo-local references
- ` + "`docs/specs/<feature>/`" + ` — feature source of truth
`

const agentsREADME = `# Agents Docs

## Purpose

- Start here for repo-local agent guidance
- Classify the task, then load only the linked doc needed for the current decision
- Avoid reading all agent docs by default

## Runtime Routing

- ` + "`WORKFLOWS.md`" + ` → classify spec-driven vs ad hoc work and resolve source-of-truth order
- ` + "`GUARDRAILS.md`" + ` → completion, safety, validation, and hard rules
- ` + "`RLM.md`" + ` → context routing and progressive disclosure
- ` + "`TOOLING.md`" + ` → skills, dispatch, worktrees, and secondary inputs
- ` + "`docs/references/*`" + ` → durable reference material only when relevant
- ` + "`docs/specs/<feature>/*`" + ` → active feature artifacts only

## Loading Rule

- Identify the immediate decision before opening another file
- Prefer a specific section over a full file
- Stop loading once the decision is supported
- Treat repo-local docs as primary and global model/vendor instructions as secondary

## System Of Record

- Feature requirements, plans, and tasks live under ` + "`docs/specs/<feature>/`" + `
- Broader repo references live under ` + "`docs/references/`" + `
- Keep durable guidance here instead of expanding the injected top-level instruction files
`

const agentsWorkflows = `# Workflows

## Spec-Driven Work

- Use this path for new features, substantial behavioral changes, cross-component changes, or work that already has feature docs
- Do not load every artifact up front
- Start from ` + "`TASKS.md`" + ` to identify the next action
- Recurse into the relevant ` + "`PLAN.md`" + ` section for approach
- Recurse into the relevant ` + "`SPEC.md`" + ` requirement for scope and acceptance
- Use ` + "`BRAINSTORM.md`" + ` only for unresolved rationale
- Use prior feature docs only through explicit dependency or relationship links
- Ask clarification questions until confidence is high and unresolved assumptions are zero
- Run the implementation readiness gate before writing code
- Update docs first when the implementation changes behavior, requirements, or approach

## Source Of Truth

Authority order:

1. safety and permission constraints
2. current user request
3. ` + "`docs/CONSTITUTION.md`" + `
4. ` + "`SPEC.md`" + `
5. ` + "`PLAN.md`" + `
6. ` + "`TASKS.md`" + `
7. ` + "`BRAINSTORM.md`" + `
8. repo conventions

Execution order for feature work:

1. ` + "`TASKS.md`" + `
2. relevant ` + "`PLAN.md`" + ` section
3. relevant ` + "`SPEC.md`" + ` requirement
4. ` + "`docs/CONSTITUTION.md`" + ` only when needed

- ` + "`TASKS.md`" + ` controls next action
- ` + "`PLAN.md`" + ` controls approach
- ` + "`SPEC.md`" + ` controls requirements
- ` + "`CONSTITUTION.md`" + ` controls project invariants
- ` + "`BRAINSTORM.md`" + ` is non-binding research context

## Ad Hoc Work

- Use this path for contained bug fixes, reviews, dependency updates, config changes, or small refinements
- Inspect relevant files before editing
- Use existing repo patterns
- Verify directly with the smallest relevant checks
- Do not create feature docs unless scope requires it
- Update only the practical docs that changed, unless existing feature docs must also change

## Readiness Gate

- Challenge the active docs for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, and scope creep
- If the gate fails, update the canonical docs first, then continue

## Feature Docs

- ` + "`docs/specs/<feature>/`" + ` remains the source of truth for feature-scoped work
- Keep dependencies, relationships, and skills tables current when those docs are touched
`

const agentsRLM = `# RLM

## Purpose

- RLM is Kit's just-in-time context-routing pattern
- Use it for any task where loading full context would be noisy or wasteful
- The goal is progressive disclosure: load only the smallest relevant subset of repo knowledge needed for the immediate decision

## Trigger Signals

- codebase-wide analysis
- scan repository
- audit all integrations
- many files or services
- high uncertainty about where the relevant logic lives
- feature work with many possible prior docs or references
- any request where broad upfront reading would slow correctness

## Runtime Loop

1. identify the immediate decision
2. load the smallest relevant artifact
3. extract only required facts
4. act if context is sufficient
5. recurse only when uncertainty remains
6. stop loading once the decision is supported

## Context Budget Rules

- specific section over full file
- current feature over all features
- explicit dependency link over broad search
- repo-local docs before global model/vendor instructions

## Rules

- Keep map work file-scoped or narrowly bounded so synthesis stays deterministic
- Prefer repo-local docs before secondary global inputs
- For feature-scoped work, keep must-read inputs small: the current ` + "`TASKS.md`" + ` entry plus the linked ` + "`PLAN.md`" + ` and ` + "`SPEC.md`" + ` sections
- Use indices first: start with ` + "`kit map <feature>`" + ` and ` + "`docs/PROJECT_PROGRESS_SUMMARY.md`" + ` to shortlist candidate prior features under ` + "`docs/specs/`" + `
- Treat prior feature docs, repo references, and secondary global inputs as conditional reads only
- Open a prior feature doc only when it affects a shared interface or contract, overlapping files or modules, migrations or data shape, acceptance criteria, or an explicit relationship or dependency link
- Inspect at most 5 prior feature directories before narrowing further or asking a clarifying question
- Extract only the concrete facts that change the current feature; do not paraphrase entire prior docs into chat or copy irrelevant history into the active artifact
- Treat RLM as discovery and context selection first; do not jump straight into parallel execution while the candidate set is still broad
- Always update affected documentation and ensure touched documents stay current and properly formatted before finishing the work
- Record the docs, skills, and references that materially shaped the feature in dependency tables
- Use ` + "`kit dispatch`" + ` only when the work moves from broad discovery into multi-lane execution planning
`

const agentsTooling = `# Tooling

## Skills

- Repo-local canonical skills live under ` + "`.agents/skills/*/SKILL.md`" + `
- For feature-scoped work, start with the current feature's ` + "`SPEC.md`" + ` ` + "`## SKILLS`" + ` table
- Keep the selected skill set minimal and actionable

## Dispatch

- Use ` + "`kit dispatch`" + ` when broad work must be turned into safe multi-lane execution
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep broad or noisy discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Keep the main agent responsible for synthesis, integration, validation, and communication

## Worktrees

- When isolated checkouts are needed, keep worktrees flat under ` + "`~/worktrees/`" + `
- Use ` + "`git worktree add ~/worktrees/<repo>-<branch> <branch>`" + ` or ` + "`git worktree add -b <branch> ~/worktrees/<repo>-<branch> <base-ref>`" + `

## Secondary Global Inputs

- ` + "`~/.claude/CLAUDE.md`" + `
- ` + "`${CODEX_HOME}/AGENTS.md`" + `
- ` + "`${CODEX_HOME}/instructions.md`" + `
- ` + "`${CODEX_HOME}/skills/*/SKILL.md`" + `

- Treat these as secondary context after repo-local docs
- Do not use ` + "`.claude/skills`" + ` as canonical discovery input
`

const agentsGuardrails = `# Guardrails

## Hard Rules

- ` + "`docs/CONSTITUTION.md`" + ` is the canonical project contract
- Keep ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, and ` + "`.github/copilot-instructions.md`" + ` aligned with the repo-local docs tree
- Never mix multiple features in one ` + "`docs/specs/<feature>/`" + ` directory
- Update docs first when reality diverges from documented behavior

## Completion Bar

- Populate all required sections in ` + "`BRAINSTORM.md`" + `, ` + "`SPEC.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + `
- Replace placeholder-only sections with ` + "`not applicable`" + `, ` + "`not required`" + `, or ` + "`no additional information required`" + `
- Always update affected documentation and ensure touched docs are current and properly formatted before calling work complete
- Never claim tests passed unless they ran
- Never claim files were inspected unless they were inspected
- Never guess file contents, APIs, or behavior
- If validation cannot run, state why
- Fix relevant lint and test failures before calling work complete
- Keep dependency and relationship sections current when those docs are touched

## Code Hygiene

- Remove dead code, unused exports, and public surfaces that are not strictly necessary
- If a symbol is only used locally, reduce its visibility instead of keeping it exported

## Safety

- Prefer explicit error handling over silent failure
- Keep changes minimal and reversible
- Do not run ` + "`git add`" + ` or ` + "`git commit`" + ` without explicit approval
- Do not run ` + "`coderabbit --prompt-only`" + ` unless explicitly requested or approved
`

const referencesREADME = `# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature
- Keep long-lived background context here instead of in injected top-level instruction files
- Link these files from feature dependency tables when they materially shape work

## Starter Files

- ` + "`testing.md`" + ` — repo-wide testing norms and evidence expectations
- ` + "`tooling.md`" + ` — local tooling and command references that are broader than one feature
- ` + "`external-systems.md`" + ` — durable notes about external systems, APIs, or integrations
`

const referencesTesting = `# Testing Reference

## Purpose

- Record durable repo-wide testing guidance that is broader than one feature
- Keep feature-specific testing details in the current feature's ` + "`PLAN.md`" + ` or ` + "`TASKS.md`" + `

## Current State

- add project-specific testing guidance here when it becomes stable enough to reuse across features
`

const referencesTooling = `# Tooling Reference

## Purpose

- Record durable repo-wide tooling notes, command references, and local development expectations
- Keep short-lived implementation notes in feature docs instead of here

## Current State

- add project-specific tooling notes here when they become stable enough to reuse across features
`

const referencesExternalSystems = `# External Systems Reference

## Purpose

- Record durable notes about external systems, APIs, providers, or design sources that recur across features
- Keep feature-specific dependency details in feature docs and dependency tables

## Current State

- add durable external-system notes here when they become stable enough to reuse across features
`
