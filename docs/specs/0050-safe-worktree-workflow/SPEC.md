---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0050
  slug: safe-worktree-workflow
  dir: 0050-safe-worktree-workflow
references:
  - id: safety-guardrails
    name: Safety guardrails
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: dirty-state and destructive-action boundaries
    status: active
  - id: github-pr-delivery
    name: GitHub PR delivery
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: issue and pull-request lane semantics
    status: active
  - id: tooling
    name: Agent tooling
    type: documentation
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: project-directory and worktree guidance
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: verify the existing issue and pull-request lanes in Kit and LabCore
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the completed follow-up to the existing ready pull requests
    required: true
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Provide one safe, memorable `git wt` workflow for isolated Git issue and pull-request work while preserving in-flight primary checkouts and making the repository hierarchy predictable.

## CONTEXT

- The existing user-level `git wa` alias combines creation, pruning, forced removal, substring selection, and automatic `.env` symlinking in one opaque shell expression.
- Existing linked worktrees accumulate directly beneath `~/worktrees`, which hides repository ownership and makes similarly named lanes harder to scan.
- Kit-managed policy currently prohibits ordinary worktrees even when an unrelated dirty checkout makes a separate lane the safest way to preserve user work.
- A worktree has its own checkout, index, and `HEAD`, but shares refs, remotes, objects, configuration, and stash state with the other worktrees of the clone. Safe automation must respect both sides of that boundary.
- The accepted convention is `~/worktrees/<owner>/<repository>/<lane>`, with durable issue lanes named `GH-<number>` and temporary detached pull-request views named uppercase `PR-<number>`.
- The root checkout remains on `main`; each coding agent develops and tests in its assigned durable issue lane without checking that branch out in the root checkout.
- GitWT remains a thin wrapper around native Git worktree operations. Runtime services, databases, ports, Temporal state, and sibling repositories remain outside its scope.

## REQUIREMENTS

- Add a Kit-owned `git-wt` executable so standard Git external-command discovery makes it available as `git wt`.
- Default the hierarchy root to `~/worktrees`; allow a testable explicit override without requiring machine-specific paths in repository policy.
- Derive lowercase owner and repository segments from `origin`, preserve safe branch hierarchy below them, and reject absolute, empty, dot, or parent-traversal lane components.
- Provide a durable issue-lane command that creates or reuses exact `GH-<number>` from the freshly fetched remote default branch.
- Provide an existing-branch command that reuses a registered worktree, attaches a local branch, or creates a tracking branch from `origin`.
- Provide `PR-<number>` as a detached inspection lane fetched from the pull request head.
- Provide a repair command that resolves a same-repository pull request head and opens its durable branch worktree instead of editing the detached `PR-<number>` view.
- Provide read-only listing, exact safe removal, explicit pruning, root discovery, and dry-run-first migration of legacy flat linked worktrees.
- Removal must never use force, must refuse the current checkout, dirty state, ignored material, and local branch commits that are not present on the configured upstream.
- Migration must preflight every candidate and destination before applying, use `git worktree move`, preserve dirty contents, skip already hierarchical directories, and stop rather than overwrite or force through a conflict.
- Link the invoking checkout's repository-root `.env` into writable `issue`, `add`, and `repair` lanes by default, using a symlink only and allowing explicit `--no-link-env` isolation.
- Reusing an existing writable lane must ensure the expected `.env` symlink exists when linking is enabled.
- Missing source `.env` files must not prevent lane creation; report that no environment file was linked.
- Never overwrite a destination `.env`; refuse regular files and symlinks whose target is not the invoking checkout's source `.env`.
- Never automatically link `.envrc`, because it is executable shell configuration.
- Detached `pr` inspection lanes must not link `.env`, and `migrate` must preserve existing files and links without creating new ones.
- Safe removal may ignore only a verified GitWT-managed `.env` symlink, must unlink only that symlink before ordinary non-force worktree removal, and must restore it if removal fails.
- Do not add substring-based targeting, implicit pruning during listing, forced `nuke`, stash, reset, clean, or branch deletion.
- Project validation must not require ignored local-only `.env` or `.envrc` scaffold files in a linked checkout.
- Update canonical Kit rules, generated instruction sources, active checked-in guidance, prompts, and tests so managed projects may use worktrees only beneath `~/worktrees` with one active branch per worktree and without nesting them inside repositories.
- Keep subagents from independently creating, switching, moving, or removing worktrees; a supervisor may assign an already prepared worktree explicitly.
- Document the mental model, command map, naming rules, lifecycle, shared-state caveats, and PR-review workflow.
- Observable acceptance: focused integration tests exercise issue, branch, PR, repair, remove, prune, and dirty migration behavior; full Kit validation passes; the installed command replaces `git wa`; every legacy worktree is relocated with branch and dirty-state parity.
- Non-goals: reconciling every managed project immediately, automatically sharing `.envrc`, starting or stopping applications, multi-repository runtime orchestration, database reset or snapshot behavior, port allocation, Temporal namespace management, process supervision, automatic root-checkout branch switching, supporting fork pull-request repair automatically, deleting branches, force-removing worktrees, moving standalone clones, or merging either delivery pull request.

## ACCEPTED PLAN

1. Pass a default-enabled environment-link option through the shared writable-lane path used by `issue`, `add`, and `repair`, while keeping `pr`, `migrate`, and `GIT_WT_ROOT` behavior unchanged.
2. Add a narrow symlink manager that links only `.env`, reports a missing source, refuses destination collisions, and validates an existing lane's expected link.
3. Extend conservative removal to recognize only a verified expected `.env` symlink, preserve all other dirty, ignored, and unpublished-state protections, and restore the link if native removal fails.
4. Add focused real-Git integration coverage plus command-help assertions for creation, reuse, opt-out, collisions, safe removal, and no-copy semantics.
5. Align the canonical guide, README, command docs, Constitution, active registry rules, generated V3 support files, and current V3 instruction payload while preserving immutable V1 and V2 payloads.
6. Validate Kit fully, then update only LabCore's dedicated worktree-policy PR if its managed rules and V3 guidance require the new contract.

## DECISIONS

- Use an external `git-wt` binary instead of a Git alias so the workflow is testable, documented, and maintainable in Kit.
- Keep `PR-<number>` detached and inspection-only. Writable repair always targets the pull request's durable same-repository head branch.
- Make migration preview-only unless `--apply` is explicit.
- Make safe removal conservative. Manual intervention is preferable to losing ignored files, untracked work, or unpushed commits.
- Preserve arbitrary safe branch path components beneath owner and repository while reserving uppercase `GH-<number>` and `PR-<number>` identities for standard lanes.
- Make `.env` symlinking the one opinionated writable-lane convenience. It shares configuration without copying credentials and remains explicitly disableable with `--no-link-env`.
- Recognize a removable environment link by exact destination name, symlink type, and target match only; do not add a broad `.env` deletion or dirty-state exception.
- Keep application processes, databases, ports, Temporal state, and sibling repository coordination outside GitWT so the command remains a thin Git worktree wrapper.

## DISCOVERIES

- The existing flat worktree inventory contains both clean and dirty linked worktrees; most branches have no upstream, so migration must not depend on clean or published state.
- Kit already resolves feature allocation through the shared Git common directory, confirming that worktrees share clone-level Git state even though their checkout state is isolated.
- Active no-worktree language exists in registry rules, generated tooling guidance, legacy instruction versions, dispatch/improvement prompts, and the Constitution.
- Published `kit instructions` versions are immutable. The historical `v1` and `v2` payloads retain their hashes and former no-worktree contract; new current `v3` carries the project-oriented policy.
- Git reserves `git <command> --help` for manual-page discovery before invoking an external subcommand. Inline discovery therefore uses `git wt help` or direct `git-wt --help`.
- A real dry run found 29 legacy linked worktrees with collision-free destinations: 25 dirty and 4 clean.
- The existing writable-lane implementation already converged issue, add, repair, and reuse behavior through `addBranch`; passing one default-enabled option through that path keeps the environment policy consistent without touching detached PR or migration flows.
- Git reports an ignored or untracked managed link as an exact `.env` porcelain entry. Filtering only that entry after verifying the symlink target preserves every existing dirty, ignored, and unpublished removal protection.
- GitWT must restore the original symlink text rather than recreate a normalized target when native worktree removal fails, so relative and absolute matching links retain their original representation.
- LabCore's live `GH-78` / PR `#79` is unrelated order-to-hold product work. Its dedicated worktree-policy lane is issue `#80`, branch `GH-80`, and PR `#81`; only that lane may receive downstream policy updates.

## VALIDATION

- `go test ./internal/worktree -count=1` passed integration coverage for issue lanes, existing branches, detached PR views, same-repository repair branches, safe removal, explicit pruning, and dirty migration.
- `make fmt` and `make vet` passed.
- `go test ./... -count=1` passed across every package.
- `go test -race ./internal/worktree ./internal/templates ./pkg/cli -count=1` passed.
- `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`.
- `goreleaser check` validated the two-binary release configuration.
- `make build` produced both `bin/kit` and `bin/git-wt`; `make build-windows` produced both Windows executables.
- `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit --json` final run `20260723T212544.763147000Z-d1a63f` passed all 45 task runs and all 345 assertions with deterministic output across 15 repeated tasks.
- `./bin/kit check safe-worktree-workflow` passed.
- Focused reconcile tests proved primary checkouts still report missing local environment scaffold while linked checkouts do not require ignored `.env` or `.envrc` files.
- Pre-completion `./bin/kit check --project` first exposed the linked-checkout environment-file mismatch; after the validator fix and progress-summary refresh, the final project check passed coherently.
- `make install-git-wt`, `git wt help`, and `git wt root` passed; the installed binary and build artifact had identical SHA-256 values.
- Post-PR review regression coverage passed for dot-segment project identity rejection, offline local-lane reuse, linked-worktree versus submodule detection, narrow local-only scaffold filtering, `.envrc` policy, and V3 repository-memory completion guidance.
- `git wt migrate` previewed 29 collision-free moves. `git wt migrate --apply` moved all 29 with `git worktree move`; post-move verification matched every branch, `HEAD`, status count, and complete status hash, preserving all 25 dirty and 4 clean worktrees. A second preview reported no legacy flat linked worktrees.
- LabCore downstream validation ran `make check` successfully and passed explicit worktree-policy assertions. Its `kit check --project` remains blocked by pre-existing invalid reference relations in feature `0012` and `0013` plus pre-existing scaffold/progress warnings; no unrelated feature artifacts were changed.
- `git diff --check` passed in both repositories.
- Follow-up focused integration tests passed for default linking through issue, add, and repair; opt-out; missing sources; destination collisions; existing-lane reuse; detached PR isolation; migration preservation; matching-link removal; regular and unexpected destination refusal; restoration on Git removal failure; and help output.
- Follow-up `make fmt`, `go vet ./...`, and `go test ./... -count=1` passed across all packages.
- Follow-up race tests passed for `internal/worktree`, `internal/templates`, `internal/instructions`, and `pkg/cli`.
- Follow-up `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`.
- Follow-up `make build`, `make build-windows`, and `goreleaser check` passed for both `kit` and `git-wt`.
- `./bin/git-wt help` showed `--no-link-env` for issue, add, and repair while leaving detached PR usage unchanged.
- Immutable V1 and V2 instruction payloads retained SHA-256 values `50cbfd80732e7b1912dc65f160cbf8555d2da95cb79079f33d7131cd51a86be5` and `811842c5c87a1b8c7f82831c7c76739071921583c44b0ab9c5dc62cbc08b27fc`.
- Follow-up `./bin/kit check safe-worktree-workflow`, `./bin/kit check --all`, and `./bin/kit check --project` passed; all 47 features and the project contract were coherent.
- Follow-up prompt-system run `20260724T143503.212928000Z-6e1a13` passed all 45 task runs and all 345 assertions with deterministic output across 15 repeated tasks.

## OUTCOME

- Added the Kit-owned `git-wt` executable with durable `GH-<number>` issue lanes, existing-branch reuse, detached `PR-<number>` views, writable PR-head repair, read-only listing, conservative exact removal, explicit pruning, canonical root discovery, and dry-run-first migration.
- Installed `git-wt` at `~/.local/bin/git-wt`, removed only the obsolete global `alias.wa`, and intentionally removed forced cleanup, substring targeting, and implicit list-time pruning from the workflow.
- Writable issue, add, repair, and existing-lane reuse now symlink the invoking checkout's `.env` by default with explicit `--no-link-env` isolation; missing sources remain successful, `.envrc` remains unshared, and detached PR or migration flows create no links.
- Safe removal recognizes only an exact expected `.env` symlink, preserves regular or unexpected destinations, retains every other dirty, ignored, and unpublished-state refusal, and restores the link if native non-force removal fails.
- GitWT remains limited to lane paths, branches, native worktree operations, and the `.env` convenience; it does not orchestrate applications, databases, ports, Temporal state, processes, or sibling repositories.
- Migrated the live worktree root to lowercase owner/repository hierarchy while preserving each branch and dirty checkout exactly.
- Added immutable current agent instructions `v3`, generated/legacy template alignment, prompt boundaries, active registry policy, release/build/install support, and a practical worktree reference guide.
- Updated project validation to recognize Git-file linked checkouts and avoid pressuring them to recreate or share ignored environment files.
- Hardened project identity containment, preserved offline reuse of existing local lanes, and distinguished linked-worktree metadata from submodule `.git` files after review.
- Updated LabCore's active rules and guidance on issue `#80` and branch `GH-80` without changing its existing `GH-78` lane or reconciling any other managed project.
- Kit delivery uses issue `#78` and branch `GH-78`; both repositories remain ready for their normal commit, push, and ready-pull-request gates.

## REPOSITORY MEMORY

Decision: created

Rationale: This changes cross-repository execution policy and establishes a durable local workflow whose safety model is not fully expressed by code alone.

Artifacts:

- `cmd/git-wt`
- `internal/worktree`
- `docs/references/worktrees.md`
- Kit-managed instruction and registry policy sources
- `internal/instructions/versions/v3.md`
- LabCore `docs/agents/*` and `docs/references/*` worktree guidance
