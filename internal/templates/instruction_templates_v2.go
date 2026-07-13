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

## GitHub Delivery Hard Gate

- In Kit-managed projects, issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before any GitHub delivery mutation, load ` + "`docs/agents/GUARDRAILS.md`" + ` and the relevant ` + "`docs/references/rules/*`" + ` delivery rules
- Repo-local Kit rules outrank global GitHub/plugin defaults; do not use generic branches, commits, PR bodies, or draft defaults when Kit defines the contract

## AWS Context Hard Gate

- If .kit.yaml defines an enabled aws context, run kit aws verify before the first AWS-dependent command in a task and again immediately before any AWS mutation
- Use the verified configured profile explicitly for every AWS-dependent command, including AWS CLI, SDK, Terraform, CDK, deployment, and project scripts, where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Treat the verified account and ARN as authoritative; on missing credentials, incomplete config, or mismatch, stop and follow docs/agents/GUARDRAILS.md instead of falling back to another profile or default

## Conditional Context

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

## GitHub Delivery Hard Gate

- In Kit-managed projects, issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before any GitHub delivery mutation, load ` + "`docs/agents/GUARDRAILS.md`" + ` and the relevant ` + "`docs/references/rules/*`" + ` delivery rules
- Repo-local Kit rules outrank global GitHub/plugin defaults; do not use generic branches, commits, PR bodies, or draft defaults when Kit defines the contract

## AWS Context Hard Gate

- If .kit.yaml defines an enabled aws context, run kit aws verify before the first AWS-dependent command in a task and again immediately before any AWS mutation
- Use the verified configured profile explicitly for every AWS-dependent command, including AWS CLI, SDK, Terraform, CDK, deployment, and project scripts, where supported
- After verification, never use default, another discovered profile, or ambient credentials
- Treat the verified account and ARN as authoritative; on missing credentials, incomplete config, or mismatch, stop and follow docs/agents/GUARDRAILS.md instead of falling back to another profile or default

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
- explicit reference link over broad search
- repo-local docs before global model/vendor instructions

## Rules

- Keep map work file-scoped or narrowly bounded so synthesis stays deterministic
- Prefer repo-local docs before secondary global inputs
- For v2 feature-scoped work, keep must-read inputs small: the current ` + "`SPEC.md`" + ` section or decision, plus directly linked references, relationships, rules, evidence, or historical staged artifacts only when they affect that decision
- Treat generated ` + "`.kit/state.json`" + ` and task bundles as pointer/index data; recurse back to canonical Markdown before changing behavior
- Treat rulesets under ` + "`docs/references/rules/`" + ` as just-in-time context; load only the linked ruleset sections whose ` + "`read_policy`" + ` and ` + "`applies_to`" + ` match the current decision
- Treat ` + "`docs/notes/<feature>`" + ` as optional source material, not canonical truth; load ` + "`docs/references/rules/feature-notes.md`" + ` when notes may materially affect the task
- For feature notes, read ` + "`docs/notes/<feature>/README.md`" + ` when the notes contract is unclear, then inspect only relevant files under ` + "`inbox/`" + `, ` + "`references/`" + `, or ` + "`responses/`" + `
- Do not load every note by default, ignore ` + "`.gitkeep`" + ` placeholders, and do not read ` + "`private/`" + ` unless the user explicitly points to local private context
- Promote durable conclusions from notes into ` + "`SPEC.md`" + `, ` + "`docs/CONSTITUTION.md`" + `, or durable references, and record materially used note files in front matter references
- Load ` + "`docs/references/rules/agent-team-orchestration.md`" + ` only when the immediate decision includes execution topology, subagent lanes, or read-only verification; do not load it for trivial single-lane tasks
- Use indices first: start with ` + "`kit map <feature>`" + ` and ` + "`docs/PROJECT_PROGRESS_SUMMARY.md`" + ` to shortlist candidate prior features under ` + "`docs/specs/`" + `
- Treat prior feature docs, repo references, and secondary global inputs as conditional reads only
- Do not load every ruleset by default; feature front matter references determine when a ruleset is must-read, conditional, evidence, or skipped
- Open a prior feature doc only when it affects a shared interface or contract, overlapping files or modules, migrations or data shape, acceptance criteria, or an explicit relationship or reference link
- Inspect at most 5 prior feature directories before narrowing further or asking a clarifying question
- Extract only the concrete facts that change the current feature; do not paraphrase entire prior docs into chat or copy irrelevant history into the active artifact
- Treat RLM as discovery and context selection first; do not jump straight into parallel execution while the candidate set is still broad
- Always update affected documentation and ensure touched documents stay current and properly formatted before finishing the work
- Record the docs, skills, and references that materially shaped the feature in canonical front matter references
- Use ` + "`kit dispatch`" + ` only when the work moves from broad discovery into multi-lane execution planning
`

const agentsTooling = `# Tooling

## Skills

- Repo-local canonical skills live under ` + "`.agents/skills/*/SKILL.md`" + `
- For feature-scoped work, start with the current feature's canonical front matter ` + "`skills`" + `, falling back to the legacy ` + "`SPEC.md`" + ` ` + "`## SKILLS`" + ` table only when front matter is absent
- Keep the selected skill set minimal and actionable

## Command Capability Discovery

- Use ` + "`kit capabilities`" + ` when choosing among Kit commands and the mutation, network, write, or git behavior is not already obvious.
- Use ` + "`kit capabilities <command> --json`" + ` for one command path, including nested paths such as ` + "`rules add`" + ` or ` + "`skill mine`" + `.
- Use ` + "`kit capabilities --search <term> --json`" + ` for compact filtered discovery, and ` + "`kit capabilities --full --json`" + ` only when hidden or deprecated compatibility commands matter.
- Treat ` + "`kit capabilities`" + ` itself as read-only: it does not require a Kit project root and does not load project config, write files, call the network, run subprocesses, or mutate git.
- In downstream Kit-managed projects, load ` + "`docs/references/rules/kit-capabilities-usage.md`" + ` when command discovery affects the task.
- Downstream projects should use ` + "`kit capabilities`" + ` for command discovery; do not maintain Kit's internal command catalog from a downstream project.

## Dispatch

- Use ` + "`kit dispatch`" + ` when broad work must be turned into a safe Agent Team Plan
- Load ` + "`docs/references/rules/agent-team-orchestration.md`" + ` when dispatch, direct subagent execution, or read-only verification topology affects the task
- Keep one accountable supervisor responsible for scope, integration, validation, evidence, delivery gating, and final reporting
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep single-lane work in one supervisor lane when the task is trivial, tightly coupled, high-overlap, high-ambiguity, cannot spawn subagents, or the user requested single-agent execution
- Default to at most 3 concurrent lanes; never exceed 4
- Keep broad or noisy discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Use read-only verification subagents by default after implementation unless a recorded exception applies

## PR Review Feedback

- Use ` + "`kit pr fix`" + ` as the default PR review feedback entrypoint when current PR review feedback should become an editable dispatch prompt.
- With no ` + "`--pr`" + `, ` + "`kit pr fix`" + ` lists open pull requests in the current repository and asks which one to repair.
- Use ` + "`kit pr fix --pr <target>`" + ` when the PR is known; accepted targets match dispatch PR intake: URL, Markdown link, ` + "`owner/repo#number`" + `, or current-repo number.
- ` + "`kit pr fix`" + ` uses the prompt-producing ` + "`kit dispatch --pr`" + ` path: it pre-populates the editor with unresolved review feedback, lets the user edit the task list, and copies the resulting dispatch prompt for a coding agent.
- The generated PR-fix prompt requires a post-push reflection cycle before review-thread resolution: the coding agent must review the pushed diff in context, confirm the PR head still matches the commit it pushed, and only then resolve verified addressed conversations.
- ` + "`kit pr fix`" + ` does not run the loop agent, edit files, write ` + "`.kit/loops`" + ` evidence, stage, commit, push, post PR comments, or resolve review threads.
- Use ` + "`kit loop review`" + ` when changed code should be locally reviewed and repaired by the configured loop agent until the final response reports at least 95% correctness and ends with ` + "`done`" + `.
- Without ` + "`--pr`" + `, ` + "`kit loop review`" + ` reviews current-branch changes relative to ` + "`origin/main`" + `, falling back to local ` + "`main`" + `, plus staged and unstaged changes.
- Use ` + "`kit loop review --pr <target>`" + ` when current unresolved CodeRabbit PR feedback should be opportunistically folded into the repair loop while local review starts immediately.
- Use ` + "`kit loop review --pr <target> --watch`" + ` or ` + "`--wait-for-coderabbit`" + ` only when finalization should block for CodeRabbit completion.
- Review prompts use one agent by default; pass ` + "`--subagents`" + ` to let the parent review agent pre-analyze the diff and choose subagents only when the lanes are clearly independent under ` + "`agent-team-orchestration`" + ` limits.
- Use ` + "`kit dispatch --loop --pr <target>`" + ` when current unresolved CodeRabbit PR review feedback should become a human-reviewed dispatch prompt instead of an agent repair loop.
- Use ` + "`kit dispatch --pr <target> --coderabbit`" + ` only when you need raw unresolved CodeRabbit review-thread intake without review-loop watch, classification, or summary behavior.
- Treat ` + "`kit loop review`" + ` as local repair only: it may edit files through the configured agent and write ` + "`.kit/loops`" + ` evidence, but it must not stage, commit, push, post PR comments, or resolve review threads.
- After fixes or no-op decisions are complete, validation has run, the repair is pushed, and reflection confirms no other code was pushed after the repair commit, resolve matching current unresolved review threads on the PR, including human reviewer and CodeRabbit feedback, with ` + "`kit dispatch --pr <target> --resolve --yes`" + `.
- Resolve only feedback verified as fixed or intentionally no-op; do not resolve unfixed, uncertain, stale, or unrelated feedback.
- ` + "`kit dispatch --pr <target> --resolve --yes`" + ` is an explicit GitHub mutation and must not be run speculatively.

## Project Directory

- Work in the existing project directory by default
- Do not create or use git worktrees for agent work
- If the current branch or dirty state is unsuitable, stop and ask the user how to proceed instead of creating an alternate checkout

## Secondary Global Inputs

- ` + "`~/.claude/CLAUDE.md`" + `
- ` + "`${CODEX_HOME}/AGENTS.md`" + `
- ` + "`${CODEX_HOME}/instructions.md`" + `
- ` + "`${CODEX_HOME}/skills/*/SKILL.md`" + `

- Treat these as secondary context after repo-local docs
- Do not use ` + "`.claude/skills`" + ` as canonical discovery input
`
