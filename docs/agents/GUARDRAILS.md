# Guardrails

## Hard Rules

- `docs/CONSTITUTION.md` is the canonical project contract
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree
- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for
- Never mix multiple features in one `docs/specs/<feature>/` directory
- Update docs first when reality diverges from documented behavior

## Work Lane Mutation Hard Gate

Before a coding agent performs any repository file or delivery mutation, it
must:

1. Load `docs/agents/GUARDRAILS.md` and
   `docs/references/rules/work-lane-gating.md`.
2. Complete read-only safety recon, including the current branch, dirty state,
   remote, active pull requests, registered worktrees, and exact primary
   checkout.
3. Ask exactly:

   > Before I make any repository changes, should I create a new GitHub issue, `GH-<issue-number>` branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?

   Interpret the response's first standalone token after trimming surrounding
   whitespace, case-insensitively: `c` means continue existing, while
   `n` or `y` means new lane. When shorthand leads a longer response,
   shorthand is the primary lane choice and the remaining text is supplemental
   lane instructions. Treat the case-insensitive full-form answers `new lane`, `new work lane`, `new worklane`, and `new worktree` as the new-lane choice:
   create or reuse the human-assigned GitHub issue, exact `GH-<issue-number>` branch, canonical non-primary worktree, and ready pull-request plan;
   ambiguous or contradictory responses fail closed.
4. Wait for the user's explicit choice unless that exact choice is already
   recorded for the same unit of work.
5. Record a Pull-Request Landing Plan with the repository, issue, branch,
   non-primary worktree, protected base, and create-or-update PR target, then
   verify that plan still matches before every mutation.

- This gate covers source, tests, documentation, specs, plans, generated files,
  configuration, and every other repository file. It also covers every delivery
  mutation, including issue, branch, staging, commit, push, worktree, and
  pull-request mutations, as well as merges. Read-only discovery and planning
  may precede it; write-capable commands such as `kit spec`,
  `kit init`, and `kit reconcile` may not.
- Never infer the choice from a clean default branch, an issue reference, or a
  generic request to produce a pull request.
- For a new lane, create or reuse one human-assigned issue, exact
  `GH-<issue-number>` branch, canonical linked worktree at
  `~/worktrees/<owner>/<repository>/GH-<issue-number>`, and one ready-PR plan
  before editing files.
- Continue existing work only after proving the non-protected branch, its exact
  owning linked worktree, issue scope, protected base, and create-or-update PR
  target. Reuse an existing pull request; do not create a second delivery lane.
- Treat the clone's primary/root checkout as read-only for coding-agent work,
  regardless of branch or cleanliness. Never edit there with a plan to move the
  diff later.
- One choice covers directly required tests, documentation, validation fixes,
  and delivery. Ask again for materially new or tangential scope.
- If an ungated or primary-checkout change is detected, stop and preserve it.
  Do not stage, commit, push, stash, reset, clean, discard, or silently transfer
  it; follow `work-lane-gating` recovery.

---

## GitHub Delivery Hard Gate

When the user asks to create or mutate an issue, branch, commit, push, pull request, or merge in a Kit-managed project, stop before any GitHub or git mutation.

A Kit-managed project is any repository containing `.kit.yaml`, `docs/CONSTITUTION.md`, or `docs/agents/README.md`.

Before creating or mutating issues, branches, staging, commits, pushes, PRs, or merges, agents must:

1. Load repo-local workflow entrypoints:
   - `.kit.yaml`
   - `docs/agents/README.md`
   - `docs/agents/GUARDRAILS.md`
   - `docs/agents/TOOLING.md`
   - any referenced `docs/references/rules/*` rulesets relevant to git, GitHub, branches, issues, commits, PRs, or merges
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
   - self-review and no-known-errors gate before staging or commit
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
- Worktree path:
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

## GitHub Merge Authorization Hard Gate

- Merge is a distinct mutation boundary. PR-delivery consent, automatic lane allocation, approval, check success, subagent assignment, and a program ledger never imply merge consent.
- Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set.
- Before any merge or merge-queue mutation, resolve `pull-request-merge` and load `docs/references/rules/github-pr-merge.md`.
- Reconcile the authorization source, authenticated actor, expected head/base, repository merge policy, current reviews/checks, dependencies, and infrastructure or deployment effects before every wave.
- Only exact current `MERGE_READY` nodes may merge. Pending, missing, stale-head, or policy-ineligible skipped checks are not passing.
- Revalidating an authorized target does not require another prompt. Adding a target or materially changing actor, method, environment, infrastructure effect, or recovery requires follow-up authorization.
- Never bypass protection, reviews, required checks, a merge queue, repository policy, or identity safeguards.
- Report merge, hosted workflow, deployment/runtime, and production evidence as separate claims.

## Deletion Safety Hard Gate

- Before designing deletion behavior or deleting persistent project, user, business, or external-system state, load `docs/references/rules/deletion-safety.md`.
- An unqualified delete means soft delete: use a reversible lifecycle state with a supported, authorized, and tested restore path. Task-owned ephemeral scratch that never became authoritative state is outside this retained-state definition; ambiguity remains covered.
- Treat purge, destroy, force deletion, empty-trash operations, destructive replacement, history rewrite, retention expiry, backup or snapshot deletion, cryptographic erasure, and irreversible cascades as hard delete.
- Make the normal product and operational path soft-delete by default. Keep hard delete as a separate privileged, auditable, server-enforced action; a client prompt or `force` flag alone is insufficient.
- Before any hard delete, resolve and present the exact targets, or a bounded selector first resolved to the exact current target set with its current count and materialized target IDs or an immutable snapshot/version token, environment, cascades, why soft delete is insufficient, the loss of restore, backup state, retention or legal impact, and verification plan.
- After that outline, obtain a specific manual confirmation from the human for those exact current targets. Initial requests, general task or plan approval, automation, retention schedules, prior soft-delete approval, and broad cleanup language do not count.
- Bind confirmation to the actor, action, exact targets or immutable snapshot/version, environment, and consequences. Immediately before execution, compare the current target set or version with the confirmed snapshot; any difference requires a new outline and confirmation.
- Preserve stricter repository, legal, privacy, security, infrastructure, and provider controls. One post-outline confirmation may satisfy multiple deletion gates only when the combined outline contains every required field.

## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load `docs/references/rules/infrastructure-change-approval.md`.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Put one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence into the task plan when planning is used; otherwise present it once before the first covered mutation. Obtain one explicit user confirmation for the complete bounded batch.
- Approval of a task plan containing the complete outline counts as confirmation. A sufficiently detailed initial request may also count only when it clearly authorizes the exact bounded batch and the batch does not delete or remove infrastructure.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch.
- After confirmation, execute the exact approved batch and continue the rest of the task to completion in one pass without routine command-by-command approval.
- If additional covered infrastructure changes become necessary, collect all then-known changes into one follow-up outline, obtain one confirmation, and execute that follow-up batch in one pass. Do not re-confirm actions already included in an approved batch.
- Treat a material change to target identity, environment, region or cluster, resource set, action type, impact, or recovery as a follow-up batch; compatible tools, commands, and retries inside the approved boundary do not require another prompt.

## AWS Context Hard Gate

Before AWS-dependent work, load `docs/references/rules/aws-agent-toolkit-guidance.md` and use its current AWS skill, official documentation, AWS MCP Server or CLI fallback, identity, infrastructure-approval, and secret-safety routing; repo-local Kit gates remain authoritative. When .kit.yaml defines an enabled aws context, agents must:

1. Run kit aws verify before the first AWS-dependent command in the task.
2. Run kit aws verify again immediately before any command that can mutate AWS resources or deploy through AWS-backed tooling.
3. Treat the returned account ID, ARN, and Region as authoritative. A profile name alone is not proof of identity because environment credentials can change resolution.
4. Use the verified configured profile and Region explicitly for AWS CLI, SDK, Terraform, CDK, deployment, and project scripts where supported.
5. Stop on missing AWS CLI, expired or unavailable credentials, incomplete .kit.yaml AWS fields, or an account mismatch. Read .kit.yaml and ask the user when the intended context remains ambiguous.
6. Never fall back to default, another discovered profile, or ambient credentials after verification fails.

## Completion Bar

- For V3 feature work, satisfy the phase-aware living-spec gates and keep front matter `workflow_version`, `phase`, references, relationships, and skills current; preserve version-specific requirements for legacy specs
- For legacy staged workflows, populate all required sections in the staged artifact being used
- Replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
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
- Load `docs/references/rules/source-file-size.md` before editing implementation/source or test files
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less
- Before delivery, audit the complete affected source/test scope; whole-project reconcile and scheduled maintenance audit the entire repository
- Exclude documentation files, `docs/**`, `.kit/**`, `.kit.yaml`, ignored files, vendored dependencies, and proven generated files
- Split oversized files by semantic responsibility while preserving stable public entry points and behavior; do not use minification or arbitrary numbered chunks
- If a safe split cannot be completed, report the exact file and blocker instead of silently accepting the violation

## Safety

- Prefer explicit error handling over silent failure
- Keep changes minimal and reversible
- Preserve the checkout that owns each lane; put separate lanes only beneath `~/worktrees/<owner>/<repository>/<lane>` and never inside a repository
- Use native `git worktree` commands and ordinary filesystem operations as the portable authority; do not require a wrapper, alias, or plugin
- Never stash, reset, clean, force-remove, or delete a branch to create or clear a worktree
- Link the primary checkout's `.env` and `.envrc` into writable lanes by default when each exists, using only exact verified symlinks; omit both links when isolation is required
- Never copy environment contents or overwrite destination environment material; preserve a repository- or user-supplied `.envrc`, and remember that direnv approval remains path-specific; keep runtime services, databases, ports, Temporal state, process supervision, and sibling repositories outside the worktree workflow
- Resolve all in-scope issues autonomously and continue until the goal is fully complete or a genuine blocker remains; diagnose before retrying, preserve target and scope, and verify the recovered state
- Do not ask for routine approval to switch supported tools, including authenticated `gh`, when the authorized mutation is unchanged
- Follow `docs/references/rules/deletion-safety.md` before designing deletion behavior or deleting persistent project, user, business, or external-system state; default to soft delete and obtain post-outline specific manual confirmation before every hard delete
- Outside explicit repo-local approval gates, ask permission only before large-scale deletion or deleting sensitive files
- Treat missing credentials, ambiguous identity or target, conflicting user-owned changes, and required external authorization as blockers requiring the smallest missing input, not as routine retry-permission requests
- Do not run `coderabbit --prompt-only` unless explicitly requested or approved


## Repository Memory Completion Gate

- Inspect existing repository memory before implementation.
- Create or adopt a spec before code when material rationale exists.
- After implementation and validation, curate durable rationale into the correct canonical documents.
- A justified `not required` decision is valid when code and tests preserve the complete durable truth.
- Every implementation final response must include `Repository Memory`, a valid decision, rationale, and artifact paths or `none`.
