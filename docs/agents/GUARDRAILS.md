# Guardrails

## Hard Rules

- `docs/CONSTITUTION.md` is the canonical project contract
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree
- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for
- Never mix multiple features in one `docs/specs/<feature>/` directory
- Update docs first when reality diverges from documented behavior

## GitHub Delivery Hard Gate

When the user asks to create or mutate an issue, branch, commit, push, or pull request in a Kit-managed project, stop before any GitHub or git mutation.

A Kit-managed project is any repository containing `.kit.yaml`, `docs/CONSTITUTION.md`, or `docs/agents/README.md`.

Before creating or mutating issues, branches, staging, commits, pushes, or PRs, agents must:

1. Load repo-local workflow entrypoints:
   - `.kit.yaml`
   - `docs/agents/README.md`
   - `docs/agents/GUARDRAILS.md`
   - `docs/agents/TOOLING.md`
   - any referenced `docs/references/rules/*` rulesets relevant to git, GitHub, branches, issues, commits, or PRs
   - `.github/pull_request_template.md` and issue templates when present
2. Run delivery recon and report the result:
   - `pwd`
   - `git status --short --branch`
   - `git remote -v`
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
   - staging rule
   - commit message format
   - PR draft/ready convention
   - PR template headings
   - required validation commands
4. Present a short Delivery Contract and wait for explicit user approval if any field is unknown, ambiguous, missing, or conflicts with generic agent defaults.
5. Never use global defaults such as `codex/<slug>` branches, ad hoc issue bodies, ad hoc PR bodies, draft PRs, `git add -A`, `git add .`, or generic commit messages when repo-local Kit rules define different behavior.
6. If repo-local delivery rules cannot be found or are incomplete, stop and ask. Do not invent a substitute workflow.

Before executing GitHub delivery, output:

```text
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
```

If any field is unknown, stop.

The `PR title format` field must resolve to Conventional Commits title format with the GitHub issue as scope:
`<type>(<issue_number>): <gitmoji> <short title message>`.

## No Generic GitHub Defaults In Kit Projects

In a Kit-managed project, global agent/plugin GitHub workflows are fallback tools only. They do not define process.

Do not create:

- `codex/*` branches
- ad hoc issue bodies
- ad hoc PR bodies
- draft PRs by default
- commits using generic messages
- PRs that omit the repo template

unless the repo-local Kit rules explicitly require them or the user explicitly overrides the Kit contract.

## Completion Bar

- For v2 feature work, populate all required `SPEC.md` sections and keep front matter `workflow_version`, `phase`, references, relationships, and skills current
- For legacy staged workflows, populate all required sections in the staged artifact being used
- Replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
- Always update affected documentation and ensure touched docs are current and properly formatted before calling work complete
- Never claim tests passed unless they ran
- Prefer citing `kit legacy verify` run IDs or exact command output when declared verification exists
- Never claim files were inspected unless they were inspected
- Never guess file contents, APIs, or behavior
- If validation cannot run, state why
- Fix relevant lint and test failures before calling work complete
- Keep canonical front matter references and relationships current when those docs are touched

## Code Hygiene

- Remove dead code, unused exports, and public surfaces that are not strictly necessary
- If a symbol is only used locally, reduce its visibility instead of keeping it exported
- Keep implementation/source code files around 300 lines or less when splitting improves clarity
- Do not apply the 300-line guideline to documentation files, `docs/**`, `.kit/**`, or `.kit.yaml`

## Safety

- Prefer explicit error handling over silent failure
- Keep changes minimal and reversible
- Do not run `git add` or `git commit` without explicit approval
- Do not run `coderabbit --prompt-only` unless explicitly requested or approved
