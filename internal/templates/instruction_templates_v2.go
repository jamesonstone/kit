package templates

func tocRepositoryInstructions(title string) string {
	return `## Purpose

- This file is a table of contents, not the full manual
- Repo-local markdown under ` + "`docs/`" + ` is the system of record
- Start here, then load only the docs relevant to the current task

## Read This Next

- ` + "`docs/agents/README.md`" + ` for the repo-local entrypoint
- ` + "`docs/CONSTITUTION.md`" + ` for project-wide constraints and invariants
- the relevant ` + "`docs/specs/<feature>/`" + ` files when the work is feature-scoped

## Work Routing

- Use ` + "`docs/agents/WORKFLOWS.md`" + ` to classify work as spec-driven or ad hoc
- Use ` + "`docs/agents/RLM.md`" + ` when the task is repository-scale or needs broad discovery; RLM is Kit's repository-scale context-routing pattern
- Use ` + "`docs/agents/GUARDRAILS.md`" + ` for hard constraints that must not be relaxed

## Repo Knowledge Map

- ` + "`docs/agents/README.md`" + ` — entrypoint and navigation guide
- ` + "`docs/agents/WORKFLOWS.md`" + ` — spec-driven, ad hoc, and readiness-gate flow
- ` + "`docs/agents/RLM.md`" + ` — repository-scale discovery and progressive disclosure
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, worktrees, and secondary global inputs
- ` + "`docs/agents/GUARDRAILS.md`" + ` — completion bar, safety rules, and non-negotiable invariants
- ` + "`docs/references/README.md`" + ` — durable repo-local references that are broader than one feature
- ` + "`docs/specs/<feature>/`" + ` — feature-level source of truth for requirements, plans, and tasks

## Runtime Context

- Start small and load only the docs, skills, and files relevant to the task
- Treat global Claude or Codex files as secondary inputs after repo-local docs
- If docs, skills, or references materially shape a feature, record them in the feature dependency tables

## Tool-Specific Notes

- ` + title + ` should stay short and stable so it fits easily into injected context
- Put durable workflow guidance in ` + "`docs/agents/*`" + ` rather than expanding this file
`
}

const tocCopilotInstructions = `# GitHub Copilot Repository Instructions

## Quick Rules

- Use this file as a map, not the full manual
- Start with ` + "`docs/agents/README.md`" + ` and then open only the relevant linked docs
- Treat ` + "`docs/specs/<feature>/`" + ` as the feature system of record
- Use ` + "`docs/agents/RLM.md`" + ` only when the task is repository-scale; RLM is Kit's repository-scale context-routing pattern
- Keep context minimal and source-attributed

## Fallback Read Order

- If linked-doc traversal is weak or unavailable, read ` + "`docs/CONSTITUTION.md`" + ` first
- Then read the relevant ` + "`docs/specs/<feature>/`" + ` docs for the task in scope
- Use ` + "`docs/agents/RLM.md`" + ` only when the task is broad enough that loading the whole repo context would be noisy or wasteful

## Non-Negotiable Rules

- Repo-local docs under ` + "`docs/`" + ` are the source of truth
- Always update affected documentation and keep touched docs properly formatted
- Keep context minimal and load only the docs and files relevant to the task
- Remove dead code and unnecessary exports or public surface when they are not strictly needed
- Do not treat ` + "`.claude/skills`" + ` as canonical discovery input

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

- This directory is the repo-local knowledge entrypoint for coding agents
- Top-level instruction files should route here instead of carrying the full operating manual
- Start with this index, then read only the linked docs needed for the current task

## Read Order

1. ` + "`docs/CONSTITUTION.md`" + `
2. ` + "`docs/agents/WORKFLOWS.md`" + `
3. ` + "`docs/agents/GUARDRAILS.md`" + `
4. ` + "`docs/agents/TOOLING.md`" + `
5. ` + "`docs/agents/RLM.md`" + ` when the task is repository-scale
6. relevant ` + "`docs/specs/<feature>/`" + ` docs when work is feature-scoped
7. ` + "`docs/references/*`" + ` only when a durable repo-local reference is relevant

## Directory Map

- ` + "`WORKFLOWS.md`" + ` — when to use the spec-driven or ad hoc path
- ` + "`RLM.md`" + ` — Kit's repository-scale context-routing pattern for broad discovery with progressive disclosure
- ` + "`TOOLING.md`" + ` — skills, dispatch, worktrees, and secondary global inputs
- ` + "`GUARDRAILS.md`" + ` — hard constraints, completion bar, and safety rules

## System Of Record

- Feature requirements, plans, and tasks live under ` + "`docs/specs/<feature>/`" + `
- Broader repo references live under ` + "`docs/references/`" + `
- Keep durable guidance here instead of expanding the injected top-level instruction files
`

const agentsWorkflows = `# Workflows

## Spec-Driven Work

- Use this path for new features, substantial behavioral changes, cross-component changes, or work that already has feature docs
- Read ` + "`BRAINSTORM.md`" + ` when present, then ` + "`SPEC.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + `
- Ask clarification questions until confidence is high and unresolved assumptions are zero
- Run the implementation readiness gate before writing code
- Update docs first when the implementation changes behavior, requirements, or approach

## Ad Hoc Work

- Use this path for contained bug fixes, reviews, dependency updates, config changes, or small refinements
- Follow understand -> implement -> verify
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

- RLM is Kit's repository-scale context-routing pattern: discover broadly, narrow to the smallest relevant context, then synthesize from sourced reads
- Use this pattern for repository-scale analysis, broad audits, or tasks that span many files or services
- Use RLM when the task is broad enough that loading the whole repo context would be noisy or wasteful
- The goal is progressive disclosure: load only the relevant subset of repo knowledge instead of flooding context

## Trigger Signals

- codebase-wide analysis
- scan repository
- audit all integrations
- many files or services
- high uncertainty about where the relevant logic lives

## Execution Pattern

1. index candidate docs, files, skills, and references
2. filter to the smallest set likely to matter
3. map bounded reads or file-scoped workers across the filtered set
4. reduce those results into a sourced synthesis

## Rules

- Keep map work file-scoped or narrowly bounded so synthesis stays deterministic
- Prefer repo-local docs before secondary global inputs
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
- Keep repository-scale discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
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
