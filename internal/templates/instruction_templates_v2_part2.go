package templates

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
