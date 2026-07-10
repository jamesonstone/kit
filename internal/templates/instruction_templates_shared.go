package templates

const sharedRepositoryInstructionsCore = `## Source of truth

- Primary authority for repository workflow, constraints, and change policy: ` + "`docs/CONSTITUTION.md`" + `
- Feature specs live under: ` + "`docs/specs/<feature>/`" + `
  - ` + "`SPEC.md`" + ` (v2 workflow artifact: requirements, plan, task checklist, validation, reflection, delivery, evidence)
  - ` + "`BRAINSTORM.md`" + ` (optional legacy staged research)
  - ` + "`PLAN.md`" + ` (optional legacy staged implementation plan)
  - ` + "`TASKS.md`" + ` (optional legacy staged executable task list)
  - ` + "`ANALYSIS.md`" + ` (optional, analysis scratchpad)
- Keep repository instruction files aligned with the canonical docs: ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, ` + "`.github/copilot-instructions.md`" + `

## Pasted Text Attachments

- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for

---

## GitHub Delivery Hard Gate

When the user asks to create or mutate an issue, branch, staging, commit, push, or pull request in a Kit-managed project, stop before any GitHub or git mutation. Issue, branch, staging, commit, push, and PR actions are mutation boundaries.

- A Kit-managed project is any repository containing ` + "`.kit.yaml`" + `, ` + "`docs/CONSTITUTION.md`" + `, or ` + "`docs/agents/README.md`" + `
- Repo-local Kit rules outrank global GitHub/plugin defaults; delivery rules outrank global GitHub/plugin workflows
- Load ` + "`.kit.yaml`" + `, ` + "`docs/agents/README.md`" + `, ` + "`docs/agents/GUARDRAILS.md`" + `, ` + "`docs/agents/TOOLING.md`" + `, relevant ` + "`docs/references/rules/*`" + ` rulesets, and GitHub templates before delivery mutation
- Run and report delivery recon: ` + "`pwd`" + `, ` + "`git status --short --branch`" + `, ` + "`git remote -v`" + `, current branch, default/base branch, active PRs for the current branch, matching issues, and git author/committer identity
- Resolve and output a Delivery Contract covering repository, base branch, issue source, issue number/link, branch name/base, branch-status-staleness check, staging method, commit format, PR title format, PR template, draft/ready state, required checks, cross-repo dependencies, and unknowns/blockers
- Resolve PR title format as Conventional Commits title format with the GitHub issue as scope: ` + "`<type>(<issue_number>): <gitmoji> <short title message>`" + `
- If any Delivery Contract field is unknown, ambiguous, missing, or conflicts with generic defaults, stop and ask before mutating
- Do not use ` + "`codex/*`" + ` branches, ad hoc issue/PR bodies, draft PRs by default, bulk staging, generic commit messages, or PRs that omit the repo template unless the Kit contract or user explicitly overrides it

---

## Change Classification (Required First Step)

Classify each request before implementation.

### 1) Spec-Driven (Formal Track)

Use when any apply:

- request initiated through ` + "`kit spec`" + ` or explicit legacy staged ` + "`kit legacy`" + ` commands
- new feature or capability
- substantial architectural or behavioral change
- work touches code with existing feature specs under ` + "`docs/specs/<feature>/`" + `
- changes cross component boundaries or public interfaces

Required flow:

- use the v2 single-` + "`SPEC.md`" + ` workflow by default: clarify → ready → implement → validate → reflect → deliver → complete inside ` + "`SPEC.md`" + `
- treat legacy ` + "`BRAINSTORM.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` as historical context unless the user explicitly chooses a legacy staged command

### 2) Ad Hoc (Lightweight Track)

Use when all apply:

- not initiated through ` + "`kit spec`" + ` or explicit legacy staged ` + "`kit legacy`" + ` commands
- bug fix, security review, refactor, dependency update, config change, or small refinement
- scope is contained and can be verified directly

Required flow:

- understand → implement → verify
- update only relevant practical docs (README/API docs/inline docs) when needed
- do not create spec artifacts for ad hoc work by default

### 3) Ad Hoc + Existing Feature Specs

If ad hoc work touches a feature with existing specs:

- default to updating ` + "`SPEC.md`" + ` when behavior, requirements, approach, validation, reflection, documentation, or delivery state changes
- update legacy staged artifacts only when the user explicitly chooses a legacy staged flow or the historical artifact would otherwise mislead future work
- skip spec updates only for mechanical edits (formatting, typo, dependency bump)

## Multi-feature rule

- Never mix features in one ` + "`docs/specs/<feature>/`" + ` directory.
- If work spans features, update each feature's docs separately.

## Document Completeness

- For v2 feature work, every required ` + "`SPEC.md`" + ` section must be populated
- For legacy staged workflows, every required section in the active staged artifact must be populated
- Do not leave HTML TODO comments as the only content in a section
- If a section has no additional detail, replace the placeholder comment with ` + "`not applicable`" + `, ` + "`not required`" + `, or ` + "`no additional information required`" + `

---

## Communication Style

- Lead with the outcome and make detail proportional to the task
- Include the evidence, caveats, risks, and next action needed to evaluate the result
- Use numbered lists for actual questions or clarification
- Use compact sections or bullets only when they improve scanning
- Code: production-ready, minimal comments, no extra text
- Numbers > simple claims: quantify, compare, give thresholds
- Avoid repetition, but never trade correctness or required context for brevity

---

## Workflow: Clarify → Act → Validate → Reflect (Spec-Driven Track)

### Phase 1: PLAN

- Locate the relevant feature directory in ` + "`docs/specs/<feature>/`" + `
- Read ` + "`SPEC.md`" + ` first for v2 feature work
- Use ` + "`BRAINSTORM.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` only as legacy staged context when they materially affect the current decision

- Resolve repository-discoverable facts before asking the user
- Ask concise numbered questions only for material choices that change scope, behavior, risk, validation, or delivery and cannot be inferred safely
- Include a recommended default and impact for each question; stop after the questions while input is required
- When no material question remains, proceed with the current user request and canonical docs without requesting routine approval
- Record constraints, edge cases, dependencies, compatibility, measurable limits, and durable decisions in ` + "`SPEC.md`" + `

### Phase 2: ACT

- Before writing code for spec-driven work, run the v2 readiness gates: adversarially challenge ` + "`CONSTITUTION.md`" + ` and ` + "`SPEC.md`" + ` for contradictions, ambiguity, hidden assumptions, missing failure modes, task gaps, validation gaps, delivery ambiguity, and scope creep. If a gate fails, update ` + "`SPEC.md`" + ` first, then continue.
- Implement tasks from the ` + "`SPEC.md`" + ` task checklist and keep task status current there
- Map validation 1:1 to ` + "`SPEC.md`" + ` acceptance criteria and record evidence in ` + "`SPEC.md`" + `
- Legacy staged flows may still use ` + "`TASKS.md`" + ` and ` + "`kit legacy verify <feature> --task <task-id>`" + ` evidence when those artifacts are active
- Follow all code style guidelines and architectural standards
- Ensure explicit error handling and input validation
- Add or update tests required by the plan
- Provide a final summary of all files changed
`
