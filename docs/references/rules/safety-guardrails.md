---
kind: ruleset
slug: safety-guardrails
description: Always-on git and GitHub guardrails for recon, identity, branch protection, secrets, and failure handling.
status: active
applies_to:
  - git
  - github
  - safety
  - coding-agent
read_policy_default: must
---

# Ruleset: safety-guardrails

## Purpose

- Define always-on safety checks for repository, GitHub, and git operations.
- Run before work-lane decisions or PR delivery workflow.
- Prevent primary-checkout edits, unsafe writes, identity mistakes,
  protected-branch mutations, blind retries, and unauthorized deletion.

## Applies When

- Always active before any coding-agent repository, GitHub, or git mutation,
  regardless of PR consent.
- Runs first, before `work-lane-gating`.
- Applies before assessing lane, branch, import-graph, or PR workflow questions.

## Rules

### Execution Order

1. Run read-only `safety-guardrails` recon and identity checks.
2. Apply `work-lane-gating`: default to a new worklane without asking, or honor
   an explicit same-scope direction to continue an existing lane. Treat an
   exact existing-PR lifecycle target set as continuation, not new work.
3. Record the complete Pull-Request Landing Plan.
4. Establish or verify the non-primary writable worktree.
5. Run `github-pr-delivery` only inside that lane.

Do not evaluate lane or import-graph questions before recon completes. The current branch scope cannot be assessed without knowing the branch.

### Canonical Authority Model

Use this matrix instead of inventing an authority boundary in each workflow:

| Action | Required authority |
| --- | --- |
| Read-only discovery | Implied by the task |
| In-scope implementation and safe recovery | Current accepted task |
| Issue, branch, commit, push, and ready PR | Accepted repository-mutating task through the default new lane, or explicit same-scope continuation |
| Review-thread mutation | Explicitly assigned repair/resolution authority |
| PR merge | Direct merge request or accepted bounded merge plan |
| Multi-repository merge program | Approved plan plus reconciled program ledger |
| Deployment or infrastructure mutation | Applicable infrastructure/deployment approval |
| Protection bypass, admin override, identity substitution | Prohibited |

- Authority is bounded by repository, target, action, intended effect, actor,
  and applicable environment. Revalidation and compatible retries preserve an
  unchanged authority boundary; material scope expansion requires follow-up
  authorization.
- PR-delivery consent, automatic lane allocation, subagent assignment, check
  success, and program-ledger existence never imply merge authority.
- Before an authorized merge, resolve the `pull-request-merge` workflow and
  follow `github-pr-merge`.

### Prohibited Actions

GitHub access is never permission to:

- Merge without a direct request or accepted bounded merge plan.
- Force-push protected branches.
- Delete branches.
- Change repository settings.
- Change secrets.
- Bypass review.
- Alter protected branch rules.

### Working-Tree Recon

Run before anything and show output:

```bash
pwd
git status --short --branch
git remote -v
git rev-parse --abbrev-ref HEAD
git worktree list --porcelain
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
gh pr list --head "$CURRENT_BRANCH" --state all --json number,url,state,isDraft,headRefName,baseRefName,assignees
```

- Identify current branch, default/base branch, repository owner/name, and the
  exact primary checkout.
- Identify active pull requests for the current branch before editing, committing, pushing, or mutating any PR.
- Confirm the active directory, branch, remote, and PR head branch match the intended work lane.
- Repeat this recon after thread resumes, user redirects, branch changes, or any sign that another thread may have moved the work forward.
- If the active branch, remote, or PR state does not match the expected issue branch or repository, resolve the mismatch autonomously when the intended lane can be proven safely; otherwise report the ambiguity and request the smallest input needed before mutating.
- If `gh` is unavailable, use an approved GitHub connector for the active PR lookup; if neither is available, report that the active PR check could not run before mutating.
- Do not overwrite, revert, or mix unrelated user changes.
- If unrelated dirty files exist, leave them alone.
- If dirty files overlap the requested work, preserve them and resolve the overlap autonomously when ownership and intent are evident; otherwise complete unblocked work and request the smallest clarification needed without discarding changes.
- Work in the existing checkout only when it is the exact non-primary worktree
  that owns the user-selected writable lane.
- When the current checkout contains unrelated work or another active lane, preserve it and use a separate canonical worktree rather than switching, stashing, resetting, cleaning, or mixing branches.

### Worktree Lane Selection

- Canonical linked worktrees live only at `~/worktrees/<owner>/<repository>/<lane>`, outside every repository checkout.
- Before creating a worktree, inspect `git worktree list --porcelain` and reuse the exact registered branch worktree when one exists.
- Use exact uppercase `GH-<number>` for a durable issue branch lane.
- Use exact uppercase `PR-<number>` only for a temporary detached pull-request view. Never implement or repair in that detached view; reuse the pull request's writable head branch worktree.
- Keep one active branch in one worktree. Do not bypass Git's branch ownership check.
- Treat working files, the index, and `HEAD` as worktree-local. Treat objects, refs, remotes, most configuration, and stash entries as shared clone state.
- Never create linked worktrees inside a project directory, including `.worktrees/`.
- Treat the exact primary/root checkout as read-only for coding-agent work,
  even when it is clean, on a feature branch, or has a pull-request plan. Do
  not edit files, generate artifacts, stage, commit, or switch branches there.
- Perform every repository mutation directly in the user-selected non-primary
  durable worktree. Never edit the primary checkout with a plan to move the
  diff later.
- Use native `git worktree` commands and ordinary filesystem operations as the portable authority. Rules and reconciled guidance must not require `git-wt`, an alias, plugin, or other wrapper.
- For a writable lane, link the clone's primary checkout repository-root `.env` and `.envrc` by default when each source exists. Create only exact symlinks after proving the destinations do not exist, or accept already-matching symlinks during reuse; omit both links when isolation is required.
- Never copy environment contents or overwrite destination environment material. Preserve a repository- or user-supplied `.envrc`; because `.envrc` is executable shell configuration, review the primary source and retain path-specific direnv approval.
- Keep detached `PR-<number>` views environment-isolated, and preserve existing files and links during migration without creating new ones.
- Never use stash, reset, clean, force removal, branch deletion, or substring-based target selection to make a worktree operation succeed.
- List worktrees without pruning. Prune only through an explicit prune action after reviewing stale metadata.
- Remove only an exact registered path after proving it is not the current checkout, contains no tracked, untracked, or ignored material other than verified expected `.env` and `.envrc` symlinks, and has no unpushed commits. Verify that each link targets the matching primary-checkout source, unlink only those symlinks before ordinary non-force `git worktree remove`, and restore them if removal fails.
- Keep runtime services, databases, ports, Temporal state, process supervision, and sibling-repository orchestration outside the worktree workflow.
- Subagents may use only a worktree explicitly prepared and assigned by the supervisor. They may not independently create, switch, move, or remove worktrees.
- Load `docs/references/worktrees.md` for command usage and the complete mental model.

### Protected-Branch Detection

Treat these branch names as protected by assumption, case-insensitively, pending verification:

```text
main, master, develop, dev, production, prod, prd, staging, stage, stg
```

Verify actual protection before any write that could touch one:

```bash
# requires authenticated gh; CURRENT_BRANCH and BASE_BRANCH from recon
for b in "$CURRENT_BRANCH" "$BASE_BRANCH"; do
  if gh api "repos/{owner}/{repo}/branches/$b/protection" >/dev/null 2>&1; then
    echo "PROTECTED (verified): $b"
  else
    case " main master develop dev production prod prd staging stage stg " in
      *" $(echo "$b" | tr '[:upper:]' '[:lower:]') "*) echo "PROTECTED (assumed; could not verify): $b" ;;
      *) echo "unverified, not in assumed set: $b" ;;
    esac
  fi
done
```

- If a target branch is protected, verified or assumed, never commit to it, force-push it, or delete it. Branch off it instead.
- If `gh` cannot verify and the branch is in the assumed set, fail closed and treat it as protected.
- Treat the discovered base branch as protected by default.

### Secret Scanning

- Before staging, verify no secrets, credentials, `.env` files, local tokens, private keys, or machine-local config were added accidentally.

### Git Author Identity

- Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, comment, or other attribution text.
- Inspect before committing and show output:

```bash
git config user.name
git config user.email
git var GIT_AUTHOR_IDENT
git var GIT_COMMITTER_IDENT
```

- If identity is missing, ambiguous, or not the human user's, stop and ask.
- Never fall back to a default, tool, bot, or agent identity.
- Do not change global git config unless the user explicitly asks.
- Prefer per-commit or repo-local config when a config change is needed.

### Autonomous Failure Recovery

Agents own the requested outcome. On a lint, test, template, tool, authentication, state, push, or other workflow failure:

1. Capture the exact error and current state.
2. Diagnose the cause with read-only inspection before another mutation.
3. Preserve the authorized repository, target, scope, intended effect, and human identity.
4. Choose a safe compatible recovery path, retry autonomously, and verify the resulting state.
5. Continue until the goal is complete or a genuine external blocker remains.

- A failed tool or connector does not revoke authorization for the same intended mutation. Use another supported authenticated path, including `gh`, without asking the user to reply with retry permission when repository, target, scope, intended effect, and human identity are unchanged.
- Do not blindly repeat the same failed command. Retry only after diagnosis or a material change in evidence, state, parameters, or tool path.
- Do not use `--force`, `--force-with-lease`, `git rebase`, `git reset --hard`, `git add -A`, `git add .`, an amend to an already-pushed commit, or branch/issue/PR recreation as a recovery shortcut.
- Missing credentials, ambiguous identity or target, conflicting user-owned changes, unavailable external dependencies, or required external authorization are genuine blockers. Complete unblocked work, report the evidence, and request only the smallest missing input; do not frame this as permission for a routine retry.
- Autonomous recovery never invents an existing-lane continuation direction,
  Pull-Request Landing Plan, or exact ownership proof. When continuation was not
  explicitly directed, recover through the default new worklane without asking.
  Preserve ungated and primary-checkout changes under the work-lane tripwire
  instead of staging, committing, pushing, discarding, or silently transferring
  them.

### Permission Boundary

- Resolve all in-scope implementation, validation, and delivery issues autonomously and continue until the requested goal is fully complete or a genuine external blocker remains.
- The default new-worklane route is a mandatory permission boundary. An
  accepted repository-mutating task authorizes that route without another
  prompt. Continue an existing lane only after an explicit same-scope user
  direction; do not infer continuation from current repository state.
- Review repair, CI repair, base refresh, and ordered merge coordination for an
  exact existing pull-request set preserve those pull-request lanes. Do not
  create coordination or corrective pull requests for scope-preserving work.
- Follow `deletion-safety` before designing deletion behavior or deleting
  persistent project, user, business, or external-system state. Default to
  soft delete and obtain post-outline specific manual confirmation for the
  exact current targets before every hard delete.
- Honor explicit repo-local approval gates. For covered public-cloud, Kubernetes, or infrastructure-as-code mutations, follow `infrastructure-change-approval` before mutation; use one plan-level confirmation per complete batch, then execute that batch and compatible retries without re-prompting; consolidate additional required changes into one follow-up batch, and always obtain post-outline confirmation for deletion or removal. Routine application operations, including deployment image updates and ECS interactions that do not create or delete infrastructure, are not covered mutations and do not require that confirmation.
- Outside explicit repo-local approval gates, ask permission only before large-scale deletion or deleting sensitive files.
- Before requesting hard-delete confirmation, resolve the exact targets, scope,
  sensitivity, cascades, and recoverability with read-only inspection. An
  unqualified deletion request never authorizes irreversible purge.
- This permission boundary does not authorize actions prohibited above. Never ask for permission to bypass protected branches, review, identity, secret, force-push, merge-authorization, or repository-setting safeguards.

## Anti-Patterns

- Do not commit directly to a protected or assumed-protected branch.
- Do not clean up a failure with destructive git commands.
- Do not hide unrelated dirty work inside the requested change.
- Do not put worktrees inside repositories, improvise flat paths outside the project hierarchy, edit detached `PR-<number>` views, or force worktree cleanup.
- Do not stage secrets, `.env` files, tokens, private keys, or machine-local config.
- Do not proceed to lane gating before branch and repository recon is complete.
- Do not mutate the primary checkout or use it as a temporary edit location.
- Do not ask for a lane choice. Default to a new worklane, and do not infer
  existing-lane continuation from clean state, current branch, issue references,
  or a generic request for a pull request.
- Do not commit when author or committer identity is missing, ambiguous, or not the human user's.
- Do not ask the user to authorize a compatible `gh` or connector retry that preserves the already-authorized mutation.
- Do not treat autonomous failure recovery as permission to bypass an explicit infrastructure-change approval gate.
- Do not blindly repeat a failed mutation without new evidence or a revised recovery path.
- Do not perform large-scale deletion or delete sensitive files without explicit permission.
- Do not hard-delete covered state without the exact post-outline manual
  confirmation required by `deletion-safety`.
- Do not treat delivery consent, agent assignment, successful checks, or a
  program ledger as authority to merge.

## Verification

- Confirm `pwd`, `git status --short --branch`, `git remote -v`, and `git rev-parse --abbrev-ref HEAD` were run and shown.
- Confirm current branch, default/base branch, and repository owner/name were identified.
- Confirm `git worktree list --porcelain` identified the exact primary checkout
  and intended writable worktree.
- Confirm active PRs for the current branch were checked, or that an unavailable PR lookup was explicitly reported before mutation.
- Confirm the active directory, branch, remote, and PR state matched the intended work lane before editing, committing, pushing, or mutating a PR.
- Confirm the default new lane or explicitly continued existing lane and the
  complete Pull-Request Landing Plan were recorded before any repository file
  mutation.
- Confirm the primary checkout received no coding-agent file, index, commit, or
  branch mutation.
- Confirm protected or assumed-protected branches were not written to.
- Confirm overlapping dirty files were preserved and either resolved safely from evidence or reported as a genuine blocker without destructive cleanup.
- Confirm every separate lane reused or created the exact `~/worktrees/<owner>/<repository>/<lane>` path, and that no nested worktree or duplicate branch checkout was created.
- Confirm detached `PR-<number>` views remained inspection-only and writable repairs used the pull request head branch.
- Confirm writable lanes linked only the expected `.env` and `.envrc` sources by default or intentionally omitted both links for isolation, preserved existing destination material, and did not copy environment contents, stash, reset, clean, force-remove, delete a branch, or discard tracked, untracked, ignored, or unpushed state.
- Confirm author and committer identity were inspected before any commit.
- Confirm secret scanning happened before staging.
- Confirm routine failures were diagnosed, recovered autonomously, and verified, or that a genuine blocker was reported with the smallest required user input.
- Confirm compatible authenticated tool-path changes did not trigger routine permission requests.
- Confirm explicit repo-local approval gates were satisfied before their covered mutations.
- Confirm large-scale deletion and sensitive-file deletion did not occur without explicit permission.
- Confirm covered deletion defaulted to a recoverable lifecycle and every hard
  delete had exact, current, post-outline manual confirmation.
- Before any merge, confirm a direct request or accepted bounded plan created
  authority for the exact PR set and `pull-request-merge` resolved without a
  required evidence gap.

## Examples

Working-tree recon:

```bash
pwd
git status --short --branch
git remote -v
git rev-parse --abbrev-ref HEAD
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
gh pr list --head "$CURRENT_BRANCH" --state all --json number,url,state,isDraft,headRefName,baseRefName,assignees
```

Git identity inspection:

```bash
git config user.name
git config user.email
git var GIT_AUTHOR_IDENT
git var GIT_COMMITTER_IDENT
```

Compatible GitHub recovery:

```text
The connector PR mutation failed, but repository, branch, PR intent, and human identity are unchanged.
Re-run read-only PR and authentication checks, use authenticated gh for the same mutation, and verify the resulting PR state without asking the user to authorize the retry.
```
