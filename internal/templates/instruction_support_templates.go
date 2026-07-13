package templates

const agentsGuardrails = `# Guardrails

## Hard Rules

- ` + "`docs/CONSTITUTION.md`" + ` is the canonical project contract
- Keep ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, and ` + "`.github/copilot-instructions.md`" + ` aligned with the repo-local docs tree
- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for
- Never mix multiple features in one ` + "`docs/specs/<feature>/`" + ` directory
- Update docs first when reality diverges from documented behavior

## GitHub Delivery Hard Gate

When the user asks to create or mutate an issue, branch, commit, push, or pull request in a Kit-managed project, stop before any GitHub or git mutation.

A Kit-managed project is any repository containing ` + "`.kit.yaml`" + `, ` + "`docs/CONSTITUTION.md`" + `, or ` + "`docs/agents/README.md`" + `.

Before creating or mutating issues, branches, staging, commits, pushes, or PRs, agents must:

1. Load repo-local workflow entrypoints:
   - ` + "`.kit.yaml`" + `
   - ` + "`docs/agents/README.md`" + `
   - ` + "`docs/agents/GUARDRAILS.md`" + `
   - ` + "`docs/agents/TOOLING.md`" + `
   - any referenced ` + "`docs/references/rules/*`" + ` rulesets relevant to git, GitHub, branches, issues, commits, or PRs
   - ` + "`.github/pull_request_template.md`" + ` and issue templates when present
2. Run delivery recon and report the result:
   - ` + "`pwd`" + `
   - ` + "`git status --short --branch`" + `
   - ` + "`git remote -v`" + `
   - current branch
   - default/base branch
   - active PRs for the current branch
   - existing matching issues
   - current git author and committer identity
3. Resolve the repo-local delivery contract before mutation:
   - issue system and required ticket format
   - issue reuse/create rules
   - branch naming convention
   - base branch refresh and staleness rules
   - self-review and no-known-errors gate before staging or commit
   - staging rule
   - commit message format
   - PR draft/ready convention
   - PR template headings
   - required validation commands
4. Present a short Delivery Contract and wait for explicit user approval if any field is unknown, ambiguous, missing, or conflicts with generic agent defaults.
5. Never use global defaults such as ` + "`codex/<slug>`" + ` branches, ad hoc issue bodies, ad hoc PR bodies, draft PRs, ` + "`git add -A`" + `, ` + "`git add .`" + `, or generic commit messages when repo-local Kit rules define different behavior.
6. If repo-local delivery rules cannot be found or are incomplete, stop and ask. Do not invent a substitute workflow.

Before executing GitHub delivery, output:

` + "```text" + `
Delivery Contract:
- Repository:
- Base branch:
- Issue source:
- Issue number/link:
- Branch name:
- Branch base:
- Branch/status/staleness check:
- Staging method:
- Commit format:
- PR title format:
- PR template:
- Draft or ready:
- Required checks:
- Cross-repo dependencies:
- Unknowns/blockers:
` + "```" + `

If any field is unknown, stop.

The ` + "`PR title format`" + ` field must resolve to Conventional Commits title format with the GitHub issue as scope:
` + "`<type>(<issue_number>): <gitmoji> <short title message>`" + `.

## No Generic GitHub Defaults In Kit Projects

In a Kit-managed project, global agent/plugin GitHub workflows are fallback tools only. They do not define process.

Do not create:

- ` + "`codex/*`" + ` branches
- ad hoc issue bodies
- ad hoc PR bodies
- draft PRs by default
- commits using generic messages
- PRs that omit the repo template

unless the repo-local Kit rules explicitly require them or the user explicitly overrides the Kit contract.

## AWS Context Hard Gate

When .kit.yaml defines an enabled aws context, agents must:

1. Run kit aws verify before the first AWS-dependent command in the task.
2. Run kit aws verify again immediately before any command that can mutate AWS resources or deploy through AWS-backed tooling.
3. Treat the returned account ID and ARN as authoritative. A profile name alone is not proof of identity because environment credentials can change resolution.
4. Use the verified configured profile explicitly for AWS CLI, SDK, Terraform, CDK, deployment, and project scripts where supported.
5. Stop on missing AWS CLI, expired or unavailable credentials, incomplete .kit.yaml AWS fields, or an account mismatch. Read .kit.yaml and ask the user when the intended context remains ambiguous.
6. Never fall back to default, another discovered profile, or ambient credentials after verification fails.

## Completion Bar

- For v2 feature work, populate all required ` + "`SPEC.md`" + ` sections and keep front matter ` + "`workflow_version`" + `, ` + "`phase`" + `, references, relationships, and skills current
- For legacy staged workflows, populate all required sections in the staged artifact being used
- Replace placeholder-only sections with ` + "`not applicable`" + `, ` + "`not required`" + `, or ` + "`no additional information required`" + `
- Always update affected documentation and ensure touched docs are current and properly formatted before calling work complete
- Never claim tests passed unless they ran
- Never claim files were inspected unless they were inspected
- Never guess file contents, APIs, or behavior
- If validation cannot run, state why
- Fix relevant lint and test failures before calling work complete
- Before staging or committing, self-review the diff against the ask, acceptance criteria, and repo-local rules; fix known relevant errors first
- Keep canonical front matter references and relationships current when those docs are touched

## Code Hygiene

- Remove dead code, unused exports, and public surfaces that are not strictly necessary
- If a symbol is only used locally, reduce its visibility instead of keeping it exported
- Keep implementation/source code files around 300 lines or less when splitting improves clarity
- Do not apply the 300-line guideline to documentation files, ` + "`docs/**`" + `, ` + "`.kit/**`" + `, or ` + "`.kit.yaml`" + `

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
- Link these files from feature front matter references when they materially shape work
- Store durable rulesets under ` + "`rules/<slug>.md`" + ` and link them with ` + "`kit rules link`" + ` instead of copying rules into agent instruction files
- Use ` + "`rules/kit-capabilities-usage.md`" + ` in downstream projects for Kit command discovery guidance
- Use ` + "`rules/feature-notes.md`" + ` when deciding how to load, reference, promote, or ignore source material under ` + "`docs/notes/<feature>`" + `
- Use ` + "`kit rules add`" + ` to import or activate available registry rulesets from the Kit GitHub ` + "`main`" + ` branch
- Use ` + "`kit rules view <slug>`" + ` to preview a local or registry ruleset before importing it
- Use ` + "`kit init --refresh`" + ` to adopt existing registry rules into ` + "`.kit.yaml`" + ` registry state and pick up safe upstream ruleset updates
- Use ` + "`kit rules add --custom`" + ` for the interactive ` + "`$EDITOR`" + ` ruleset builder
- ` + "`kit rule`" + ` is the singular alias for ` + "`kit rules`" + `

## Starter Files

- ` + "`testing.md`" + ` — repo-wide testing norms and evidence expectations
- ` + "`tooling.md`" + ` — local tooling and command references that are broader than one feature
- ` + "`external-systems.md`" + ` — durable notes about external systems, APIs, or integrations
- ` + "`rules/`" + ` — pointer-loaded durable rulesets such as frontend UI rules, testing rules, API conventions, security constraints, or domain rules
- ` + "`../notes/<feature>/`" + ` — optional feature source material; not canonical truth and private contents remain ignored
`

const referencesTesting = `# Testing Reference

## Purpose

- Record durable repo-wide testing guidance that is broader than one feature
- Keep feature-specific testing details in the current feature's ` + "`SPEC.md`" + ` Validation Map and Evidence sections; legacy staged flows may still use ` + "`PLAN.md`" + ` or ` + "`TASKS.md`" + `

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
- Keep feature-specific reference details in feature docs as canonical front matter references

## Current State

- add durable external-system notes here when they become stable enough to reuse across features
`
