---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: 0051
  slug: resolve-pr-repair-worktrees
  dir: 0051-resolve-pr-repair-worktrees
relationships:
  - type: builds_on
    target: 0050-safe-worktree-workflow
  - type: builds_on
    target: 0034-review-loop
  - type: related_to
    target: 0035-loop-review
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: safety-guardrails
    name: Safety guardrails
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: dirty-state preservation and exact worktree selection
    status: active
  - id: github-pr-delivery
    name: GitHub PR delivery
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: writable PR-head lanes, push targets, and delivery gates
    status: active
  - id: command-capabilities
    name: Command capabilities
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: mutation and network metadata updates
    status: active
  - id: tooling
    name: Agent tooling
    type: documentation
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: PR repair, dispatch, loop-review, and worktree contracts
    status: active
  - id: safe-worktree-spec
    name: Safe worktree workflow
    type: spec
    target: docs/specs/0050-safe-worktree-workflow/SPEC.md
    relation: informs
    read_policy: must
    used_for: canonical lane hierarchy and reusable repair behavior
    status: active
  - id: review-loop-spec
    name: Review-loop prompt preparation
    type: spec
    target: docs/specs/0034-review-loop/SPEC.md
    relation: informs
    read_policy: must
    used_for: shared PR intake and prompt-only boundaries
    status: active
  - id: loop-review-spec
    name: Loop review
    type: spec
    target: docs/specs/0035-loop-review/SPEC.md
    relation: informs
    read_policy: conditional
    used_for: configured-agent checkout routing
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: create and verify the tracking issue and pull request
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the validated issue branch as a ready pull request
    required: true
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Make target-bearing Kit commands infer and prepare the exact writable worktree
that owns the requested pull request or branch, so the user does not have to
navigate manually and generated agents cannot confidently repair or push from
the wrong checkout.

## CONTEXT

- `kit pr fix`, `kit dispatch --pr`, and `kit dispatch --loop --pr` currently
  render `os.Getwd()` as the working directory even when the selected pull
  request belongs to another registered worktree.
- `kit ci --dispatch` similarly treats `--repo-path` or the invoking directory
  as the repair lane, although those values identify a repository rather than a
  target branch checkout.
- `kit loop review --pr` pins its agent process to the invoking Kit project,
  but it does not prove that the project owns the requested PR head.
- Git already records exact branch-to-worktree ownership through
  `git worktree list --porcelain`; selection by recency, substring, or an
  interactive terminal list is unnecessary and unsafe for automation.
- Feature 0050 established canonical external worktree paths, same-repository
  PR-head repair lanes, default exact `.env` symlinks, fork refusal, detached
  inspection boundaries, and native Git as the portable authority.
- The user confirmed that PR-fix remains a prompt-producing workflow, while Kit
  may create or attach a missing writable worktree when needed. Dirty target
  worktrees require an include-or-exclude question whose answer is carried into
  the generated prompt.
- GitHub issue #89 and branch `GH-89` are the delivery lane.

## REQUIREMENTS

- REQ-001: Resolve a canonical repository, PR number and URL, same-repository
  head branch, expected head SHA, exact worktree path, and push target before
  generating a PR repair prompt.
- REQ-002: Reuse an exact registered PR-head worktree when present; otherwise
  attach the local branch or create a tracking worktree from `origin` at the
  canonical `~/worktrees/<owner>/<repository>/<branch>` path.
- REQ-003: Reuse the safe worktree package behavior for `.env` linking, fork
  refusal, durable branch validation, and non-destructive lane creation rather
  than invoking or depending on the optional `git wt` executable.
- REQ-004: Never select by worktree recency, fuzzy branch matching, substring,
  or the interactive `git wt list` selector.
- REQ-005: When the resolved worktree has tracked, staged, or untracked changes,
  show their porcelain status and ask whether they belong in the target PR or
  branch repair.
- REQ-006: Record `include`, `exclude`, or `none` explicitly in the prompt.
  Inclusion makes the existing diff in-scope for review and validation;
  exclusion requires preserving and not staging or modifying those paths, and
  stopping on overlap rather than stashing, resetting, cleaning, or discarding.
- REQ-007: Apply shared PR repair resolution to `kit pr fix`,
  `kit dispatch --pr`, and `kit dispatch --loop --pr`.
- REQ-008: Apply target-aware worktree resolution to CI dispatch prompts,
  preferring an inferred open PR for the diagnosed branch, otherwise attaching
  a non-default target branch and refusing to treat the protected default
  branch as a writable repair lane.
- REQ-009: Route `kit loop review --pr` automatically into the resolved PR-head
  worktree before loading config, computing diffs, or launching the agent.
- REQ-010: Normalize generic dispatch prompts to the current Git top-level or
  Kit project root rather than an arbitrary nested invocation directory.
- REQ-011: Generated PR prompts must use a canonical PR URL for head checks and
  review-thread resolution so a bare number never changes meaning with cwd.
- REQ-012: Preserve the existing clipboard, editor, output-only, reflection,
  validation, explicit review-thread resolution, staging, commit, push, and
  GitHub delivery boundaries.
- REQ-013: Update command help, capability metadata, generated V3 guidance,
  checked-in documentation, tests, and repository memory for the new
  conditional worktree and `.env` mutation behavior.

### Non-goals

- Automatically repair fork pull requests.
- Edit detached `PR-<number>` inspection worktrees.
- Automatically stage, commit, push, resolve review threads, post comments, or
  merge a pull request.
- Infer a new issue or pull request for a failure on a protected default branch.
- Reconcile divergent or overlapping user-owned changes through stash, reset,
  clean, rebase, force operations, or branch deletion.
- Make the optional `git wt` executable an execution dependency.

## ACCEPTED PLAN

1. Refactor `internal/worktree` just enough to expose reusable, output-free
   branch and PR-repair preparation while retaining the existing manual command
   behavior and real-Git integration tests.
2. Add one CLI repair-context resolver that validates repository identity,
   prepares the exact lane, inspects branch/HEAD/status, prompts on dirty state,
   and produces canonical prompt metadata.
3. Extend dispatch prompt options with repair context and render exact
   repository, PR, branch, SHA, worktree, existing-change decision, and push
   target plus non-destructive execution instructions.
4. Use the shared resolver in PR-fix, dispatch PR intake, review-loop prompt
   preparation, CI dispatch, and PR-mode loop review; normalize generic
   dispatch separately to the current project/worktree root.
5. Cover reuse, creation, fork refusal, repository mismatch, dirty
   include/exclude, prompt rendering, CI branch inference, canonical PR
   targeting, and loop-review routing with focused tests.
6. Update capability metadata, command/docs guidance, the canonical init
   contract, and generated V3 sources, then run complete validation and curate
   the demonstrated invariant after review.
7. Deliver the complete change through issue #89, branch `GH-89`, one validated
   commit, and one ready pull request to `main`.

## DECISIONS

- Keep `kit pr fix` prompt-producing: it may prepare the required local lane but
  still does not edit application files, stage, commit, push, comment, resolve,
  or merge.
- Treat worktree preparation and the exact `.env` symlink as disclosed,
  conditional Git/filesystem mutation required to establish the repair lane.
- Ask about existing dirty changes after resolving the exact target lane, with
  exclusion as the safe default.
- Include exact dirty status in the prompt so the receiving agent can preserve
  the answer instead of rediscovering an ambiguous checkout.
- Refuse forks and protected-default-branch repair when no existing PR supplies
  a safe writable head.
- Keep native Git authoritative; internal reuse of Kit's worktree package is
  not dependency on the optional external wrapper.

## DISCOVERIES

- The initially invoked `kit spec resolve-pr-repair-worktrees --output-only`
  unexpectedly entered the deprecated V2 editor flow after allocating feature
  0051. Because no thesis was captured and no implementation began, the empty
  placeholder was semantically replaced with this accepted V3 plan.
- The existing `kit ci` capability metadata incorrectly claimed that
  `--dispatch` reran workflows. Current code only opens a repair-prompt editor,
  so the catalog was corrected while adding the conditional worktree behavior.
- Generated V3 worktree policy is backed by
  `internal/templates/worktrees_reference.md`; updating only the checked-in
  reference correctly failed template-alignment tests and both sources now
  match.
- A dirty worktree on a non-interactive EOF cannot be treated as a user
  exclusion. Repair prompt generation now stops until it receives an explicit
  yes, no, or accepted default-no newline response.
- Review of the generated `git -C` instructions found that `strconv.Quote`
  emitted double-quoted Go literals, which still permit shell expansion.
  Repair worktree arguments now use POSIX single-argument quoting, including
  safe embedded-apostrophe handling.
- The original repair resolver passed caller context into worktree preparation
  but not into its follow-up Git and GitHub subprocesses. The same caller
  context now reaches repository validation, branch/status inspection, PR
  inference, and default-branch discovery so cancellation and deadlines stop
  stalled commands.
- Dispatch review-loop preparation originally validated repository ownership
  only after PR metadata, optional CodeRabbit waiting, and task loading. It now
  resolves the repair lane first and preserves that context through the
  existing no-actionable-feedback path.

## VALIDATION

- `go test ./internal/worktree ./pkg/cli`: passed.
- `go test -race ./internal/worktree ./pkg/cli -count=1`: passed.
- `go vet ./...`: passed.
- `go test ./...`: passed after aligning the generated V3 worktree reference.
- `golangci-lint run --new-from-rev=origin/main ./...`: passed with zero issues
  after the only findings were repaired.
- Native macOS Kit and `git-wt` builds, Windows amd64 Kit and `git-wt` builds,
  and `goreleaser check`: passed.
- `kit improve run --suite prompt-system --kit-binary
  /tmp/kit-gh89-final-build.Nr094Z/kit --json`: passed final run
  `20260726T151543.366542000Z-79c462` with 45/45 task runs, 345/345 assertions,
  and determinism rate 1.0.
- `go run ./cmd/kit check resolve-pr-repair-worktrees`: passed.
- `go run ./cmd/kit check --project`: passed after adding feature 0051 to the
  project progress summary.
- Targeted `kit capabilities ... --json` inspection for `pr fix`, `dispatch`,
  `loop review`, and `ci` reports the new conditional worktree behavior;
  `kit capabilities --search worktree --json` discovers every affected repair
  surface.
- `git diff --check`: passed.
- Gitleaks scan of the complete intended diff, including new untracked source
  and spec files: passed with no leaks.
- PR review repair validation passed `go test ./...`,
  `go test -race ./pkg/cli -count=1`, `go vet ./...`,
  `golangci-lint run --new-from-rev=origin/main ./...`, formatting, and
  whitespace checks. Focused tests cover shell metacharacters and apostrophes,
  canceled and deadline-terminated subprocesses, fail-fast review-loop order,
  and the preserved no-actionable-feedback result.

## OUTCOME

- Kit now exposes shared output-free worktree preparation for exact branches
  and same-repository PR heads while preserving the existing `git wt` output
  contract.
- PR fix, raw PR dispatch, dispatch review-loop, PR-mode configured loop review,
  and CI dispatch all resolve the exact writable target lane. Generic dispatch
  normalizes to the current Git top-level.
- Missing lanes are attached or created at the canonical external path with
  existing safe `.env` behavior. Fork PRs, repository mismatches, ambiguous PR
  inference, and unowned protected-default-branch repair are refused.
- Dirty target lanes require an explicit include-or-exclude decision. Every
  generated repair prompt records the canonical target, branch, remote and
  local heads, worktree, original status, decision, and push target with
  non-destructive preservation instructions.
- Command help, capability metadata, generated and checked-in V3 guidance,
  canonical command docs, worktree policy, project Constitution, feature
  summary, and focused integration tests now describe the same behavior.
- Review hardening makes generated repair commands shell-safe, keeps local
  subprocesses cancellable, and fails on repository ownership before remote
  review waits or task intake.
- Delivery uses issue #89 and branch `GH-89`; the ready pull request targets
  `main` and remains unmerged for review.

### GH-139 ownership split

- Feature 0059 removes Kit's independently installable `git-wt` product while
  preserving this feature's target-aware preparation behavior in the narrower
  `internal/worktreeprep` package.
- The retained preparer continues to reuse or create the exact writable
  same-repository PR head, link the primary checkout's environment files when
  requested, and refuse fork, closed, or detached inspection lanes.
- Kit repair remains implemented with native Git and GitHub CLI operations; it
  does not depend on Kura or another external worktree wrapper.

## REPOSITORY MEMORY

Decision: created

Rationale: Exact target-lane inference, conditional worktree creation, and
dirty-state ownership are cross-command product and safety decisions that code
and tests cannot fully explain without the accepted boundaries.

Artifacts:

- `docs/specs/0051-resolve-pr-repair-worktrees/SPEC.md`
