package templates

func tocRepositoryInstructions(title string) string {
	return `## Purpose

- This file is a routing table, not the full manual
- Start at ` + "`docs/agents/README.md`" + `, then load only the docs needed for the current decision
- Repo-local markdown under ` + "`docs/`" + ` is the system of record

## Pasted Text Attachments

- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for

## Runtime Routing

- ` + "`docs/agents/README.md`" + ` — classify the task and choose the next document
- ` + "`docs/agents/WORKFLOWS.md`" + ` — spec-driven versus ad hoc flow
- ` + "`docs/agents/GUARDRAILS.md`" + ` — completion, safety, and hard rules
- ` + "`docs/agents/RLM.md`" + ` — just-in-time context loading when broad context would be noisy
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, project-directory workflow, and secondary inputs

` + multiAgentOrchestrationRoutingGate + workLaneMutationRoutingGate + `
## Testing And Validation Gate

- Before implementation or validation, including browser automation and browser testing, load ` + "`docs/references/rules/testing-and-environment-validation.md`" + ` and the project's ` + "`docs/references/testing.md`" + `
- Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them

## GitHub Delivery Hard Gate

- In Kit-managed projects, issue, branch, staging, commit, push, PR, and merge actions are distinct mutation boundaries
- Before any GitHub delivery mutation, load ` + "`docs/agents/GUARDRAILS.md`" + ` and the relevant ` + "`docs/references/rules/*`" + ` delivery rules
- Repo-local Kit rules outrank global GitHub/plugin defaults; do not use generic branches, commits, PR bodies, or draft defaults when Kit defines the contract
- Assign every created or reused GitHub issue and pull request to the human user, such as with ` + "`gh issue create --assignee @me`" + ` or ` + "`gh pr create --assignee @me`" + ` after confirming the authenticated ` + "`gh`" + ` login is the human user; never assign a coding agent, assistant, bot, or automated identity

` + githubPRMergeGate + crossRepositoryProgramCoordinationGate + deletionSafetyGate + infrastructureChangeApprovalGate + `## AWS Context Hard Gate

- ` + awsAgentToolkitGuidanceRoute + ` If .kit.yaml defines an enabled aws context, run kit aws verify before the first AWS-dependent command in a task and again immediately before any AWS mutation
- Use the verified configured profile and Region explicitly for every AWS-dependent command, including AWS CLI, SDK, Terraform, CDK, deployment, and project scripts, where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Treat the verified account, ARN, and Region as authoritative; on missing credentials, incomplete config, or mismatch, stop and follow docs/agents/GUARDRAILS.md instead of falling back to another profile or default

` + agentCompletionOutputGate + `## Conditional Context

- ` + "`docs/specs/<feature>/`" + ` — active feature artifacts only
- ` + "`docs/references/README.md`" + ` — durable repo references only when relevant
- ` + "`docs/CONSTITUTION.md`" + ` — project invariants when a decision depends on them

## Repo Knowledge Map

- ` + "`docs/agents/README.md`" + ` — runtime routing index
- ` + "`docs/agents/WORKFLOWS.md`" + ` — work classification and source-of-truth semantics
- ` + "`docs/agents/RLM.md`" + ` — progressive disclosure and context budget rules
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, project-directory workflow, and secondary global inputs
- ` + "`docs/agents/GUARDRAILS.md`" + ` — completion bar, safety rules, and validation expectations
- ` + "`docs/references/README.md`" + ` — durable repo-local references that are broader than one feature
- ` + "`docs/specs/<feature>/SPEC.md`" + ` — v2 feature source of truth for requirements, plan, tasks, validation, reflection, delivery, and evidence

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

## Pasted Text Attachments

- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for

## Runtime Routing

- ` + "`docs/agents/README.md`" + ` — classify the task and choose the next document
- ` + "`docs/agents/WORKFLOWS.md`" + ` — workflow and source-of-truth rules
- ` + "`docs/agents/GUARDRAILS.md`" + ` — hard completion and safety rules
- ` + "`docs/agents/RLM.md`" + ` — just-in-time context routing
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, project-directory workflow, and secondary inputs

` + multiAgentOrchestrationRoutingGate + workLaneMutationRoutingGate + `
## Testing And Validation Gate

- Before implementation or validation, including browser automation and browser testing, load ` + "`docs/references/rules/testing-and-environment-validation.md`" + ` and the project's ` + "`docs/references/testing.md`" + `
- Preserve language-native code-level tests and pull-request checks; end-to-end and live-integration suites supplement rather than replace them

## GitHub Delivery Hard Gate

- In Kit-managed projects, issue, branch, staging, commit, push, PR, and merge actions are distinct mutation boundaries
- Before any GitHub delivery mutation, load ` + "`docs/agents/GUARDRAILS.md`" + ` and the relevant ` + "`docs/references/rules/*`" + ` delivery rules
- Repo-local Kit rules outrank global GitHub/plugin defaults; do not use generic branches, commits, PR bodies, or draft defaults when Kit defines the contract
- Assign every created or reused GitHub issue and pull request to the human user, such as with ` + "`gh issue create --assignee @me`" + ` or ` + "`gh pr create --assignee @me`" + ` after confirming the authenticated ` + "`gh`" + ` login is the human user; never assign a coding agent, assistant, bot, or automated identity

` + githubPRMergeGate + crossRepositoryProgramCoordinationGate + deletionSafetyGate + infrastructureChangeApprovalGate + `## AWS Context Hard Gate

- ` + awsAgentToolkitGuidanceRoute + ` If .kit.yaml defines an enabled aws context, run kit aws verify before the first AWS-dependent command in a task and again immediately before any AWS mutation
- Use the verified configured profile and Region explicitly for every AWS-dependent command, including AWS CLI, SDK, Terraform, CDK, deployment, and project scripts, where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Treat the verified account, ARN, and Region as authoritative; on missing credentials, incomplete config, or mismatch, stop and follow docs/agents/GUARDRAILS.md instead of falling back to another profile or default

` + agentCompletionOutputGate + `## Non-Negotiable Rules

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
- ` + "`docs/agents/TOOLING.md`" + ` — skills, dispatch, project-directory workflow, and secondary globals
- ` + "`docs/agents/GUARDRAILS.md`" + ` — hard rules and completion bar
- ` + "`docs/references/README.md`" + ` — durable repo-local references
- ` + "`docs/specs/<feature>/SPEC.md`" + ` — v2 feature source of truth
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
- ` + "`TOOLING.md`" + ` → skills, dispatch, project-directory workflow, and secondary inputs
- ` + "`docs/references/*`" + ` → durable reference material only when relevant
- ` + "`docs/references/rules/*`" + ` → durable rulesets only when linked from feature references or directly relevant
- ` + "`docs/specs/<feature>/*`" + ` → active feature artifacts only

## Loading Rule

- Identify the immediate decision before opening another file
- Prefer a specific section over a full file
- Stop loading once the decision is supported
- Treat repo-local docs as primary and global model/vendor instructions as secondary

## System Of Record

- V2 feature requirements, implementation plan, task checklist, validation map, reflection notes, delivery decision, and evidence live in ` + "`docs/specs/<feature>/SPEC.md`" + `
- Legacy staged ` + "`BRAINSTORM.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` files may exist as historical context or when a legacy staged command is explicitly used
- Broader repo references live under ` + "`docs/references/`" + `
- Durable repo-local rulesets live under ` + "`docs/references/rules/`" + ` and should be pointer-loaded through feature references
- Keep durable guidance here instead of expanding the injected top-level instruction files
`

const agentsWorkflows = `# Workflows

## Work Lane Precondition

- Read-only discovery and planning may run before a lane exists
- Before any repository write, default to a new worklane without asking and
  record the complete pull-request landing plan
- Continue an existing lane only when the user explicitly directs it for the
  same unit of work and exact ownership can be proven
- Treat exact existing-PR review, CI, base-refresh, and ordered-merge work as
  continuation of every targeted head; never create a coordinator or
  corrective pull request for scope-preserving work
- Create or update feature artifacts only inside the selected non-primary
  writable worktree

## Spec-Driven Work

- Use this path for new features, substantial behavioral changes, cross-component changes, or work that already has feature docs
- Do not load every artifact up front
- In v2 feature work, start from ` + "`SPEC.md`" + `; it is the single durable feature artifact
- Use ` + "`SPEC.md`" + ` sections for thesis, context, clarifications, requirements, assumptions, acceptance criteria, implementation plan, task checklist, validation map, reflection notes, documentation updates, delivery decision, and evidence
- Treat legacy ` + "`BRAINSTORM.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` as historical context unless the user explicitly chooses a legacy staged command
- Use ` + "`BRAINSTORM.md`" + ` only for unresolved historical rationale
- Use ` + "`PLAN.md`" + ` and ` + "`TASKS.md`" + ` only for legacy staged flows or historical comparison
- Use prior feature docs only through explicit reference or relationship links
- Resolve repository-discoverable facts first; ask only about material non-discoverable choices, and begin implementation only when unresolved assumptions are zero
- Run the v2 readiness gates before writing code: clarification complete, acceptance criteria binary-verifiable, task checklist mapped to criteria, validation mapped 1:1, delivery intent known
- Update docs first when the implementation changes behavior, requirements, or approach

## Source Of Truth

Authority order:

1. safety and permission constraints
2. current user request
3. ` + "`docs/CONSTITUTION.md`" + `
4. ` + "`SPEC.md`" + `
5. legacy ` + "`PLAN.md`" + ` / ` + "`TASKS.md`" + ` when the user explicitly chooses a staged flow
6. legacy ` + "`BRAINSTORM.md`" + `
7. repo conventions

Execution order for feature work:

1. ` + "`SPEC.md`" + `
2. relevant ` + "`SPEC.md`" + ` task checklist item, acceptance criterion, and validation map entry
3. legacy staged artifacts only when explicitly operating in a legacy staged flow
4. ` + "`docs/CONSTITUTION.md`" + ` only when needed

- ` + "`SPEC.md`" + ` controls requirements, plan, tasks, validation, reflection, delivery, and evidence
- ` + "`CONSTITUTION.md`" + ` controls project invariants
- ` + "`BRAINSTORM.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` are non-binding historical context in v2 unless the user chooses a legacy staged flow

## Ad Hoc Work

- Use this path for contained bug fixes, reviews, dependency updates, config changes, or small refinements
- Inspect relevant files before editing
- Use existing repo patterns
- Verify directly with the smallest relevant checks
- Do not create feature docs unless scope requires it
- Update only the practical docs that changed, unless existing feature docs must also change

## Readiness Gate

- Challenge ` + "`SPEC.md`" + ` for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, validation gaps, delivery ambiguity, and scope creep
- If the gate fails, update the canonical docs first, then continue

## Feature Docs

- ` + "`docs/specs/<feature>/`" + ` remains the source of truth for feature-scoped work
- v2 feature work keeps durable workflow state in ` + "`SPEC.md`" + `
- ` + "`SPEC.md`" + ` front matter should include ` + "`workflow_version: 2`" + ` and a current ` + "`phase`" + `
- Keep references, relationships, and skills metadata current when those docs are touched
`
