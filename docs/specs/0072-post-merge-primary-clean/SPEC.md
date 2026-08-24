---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0072"
  slug: post-merge-primary-clean
  dir: 0072-post-merge-primary-clean
relationships:
  - type: builds_on
    target: 0017-reconcile-command
    note: Adds post-merge primary leftover cleanup after the existing worktree-and-PR reconcile delivery path.
  - type: related_to
    target: 0063-explicit-work-lane-choice
    note: Keeps primary-checkout implementation read-only and tripwire preservation until the matching worktree PR has merged.
references:
  - id: github-pr-delivery
    name: GitHub pull-request delivery ruleset
    type: rule
    target: docs/references/rules/github-pr-delivery.md
    relation: implements
    read_policy: must
    used_for: durable post-merge primary leftover-cleanup contract
    status: active
  - id: work-lane-gating
    name: Work-lane gating ruleset
    type: rule
    target: docs/references/rules/work-lane-gating.md
    relation: implements
    read_policy: must
    used_for: tripwire exception after the matching worktree PR has merged
    status: active
  - id: managed-file-delivery
    name: Generated managed-file delivery instructions
    type: code
    target: pkg/cli/prompt_rules.go
    relation: implements
    read_policy: must
    used_for: reconcile, init, and health coding-agent delivery prompts
    status: active
  - id: deletion-safety
    name: Deletion safety ruleset
    type: rule
    target: docs/references/rules/deletion-safety.md
    relation: informs
    read_policy: evidence
    used_for: leftover untracked files after merge are ephemeral duplicates, not authoritative product state
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Stop leftover `kit reconcile` files on the primary default branch from
blocking `git pull` after the worktree pull request has merged. Keep the
current design that creates a worktree, adds the same files there, and
lands the remaining work through that PR. After remaining PR feedback is
addressed and that PR is merged, instruct the coding agent to run
`git clean -fd` on primary `main` and then pull.

## CONTEXT

- Issue #176 reports that `kit reconcile` used to add and update rules
  leaves stale unstaged or untracked copies on the primary default branch
  `main`. Generated instructions then create a worktree, add the same
  files there, and finish the remaining work. That worktree-and-PR path
  remains by design.
- After that PR merges, the leftover files on `main` block pulling the
  merged result. Untracked copies typically fail pull with working-tree
  files that would be overwritten.
- GH-160 deferred write-capable primary refreshes and asked the agent to
  create the worktree first, then rerun reconcile there. The user reports
  that worktree-first / pre-create approach failed in practice and must
  not be retried.
- Current generated delivery instructions also forbid `git clean`, so an
  agent that follows them cannot remove the leftovers after merge.
- `git clean -fd` removes untracked files. Leftover unstaged edits to
  already-tracked command-owned paths would still block pull and need a
  scoped restore of those exact paths. Unrelated dirty or untracked state
  must not be wiped.
- The leftovers are ephemeral local duplicates of content that became
  authoritative in the merged commit. Discarding them is leftover
  disposal, not deletion of persistent product state. The tripwire still
  forbids cleaning ungated primary changes before that merge.

## REQUIREMENTS

- Keep the current reconcile delivery design: command-owned rule and
  managed-file updates may appear on primary `main`; the coding agent
  creates or reuses the canonical worktree, adds the same files there,
  and completes remaining work through that pull request.
- Generated managed-file delivery instructions, with and without a
  command-owned snapshot, must tell the coding agent to run
  `git clean -fd` on the primary default-branch checkout only after
  remaining pull-request feedback is addressed and the worktree pull
  request has been merged.
- Those instructions must then tell the agent to pull the merged default
  branch onto the primary checkout.
- If leftover unstaged tracked modifications of the same command-owned
  paths would still block pull, restore those exact paths to `HEAD`
  before pulling.
- Do not run `git clean` before merge, inside the writable worktree, to
  create or clear a lane, or when unrelated untracked files are present.
  If unrelated dirty or untracked state exists, stop and report it.
- Durable `github-pr-delivery` and `work-lane-gating` rules must agree
  with the generated prompt. The tripwire continues to preserve ungated
  primary changes until the matching worktree PR has merged.
- Do not revert primary-checkout refresh deferral, add a public reconcile
  flag, or authorize merge by default. Cleanup runs only after an
  authorized merge has already happened.
- Do not retry pre-creating the canonical worktree so writes never land
  on `main`.

## ACCEPTED PLAN

1. Record this accepted plan in `0072` before implementation.
2. Extract shared managed-file closing instructions that keep secret and
   bulk-mutation prohibitions, then add the post-merge primary
   `git clean -fd`, scoped restore, and pull sequence.
3. Qualify the snapshot tripwire so leftovers stay preserved until the
   worktree PR merges, then follow the post-merge cleanup.
4. Add the same sequence to `github-pr-delivery` as leftover disposal
   after an authorized merge, and add the matching tripwire exception to
   `work-lane-gating`.
5. Point `0017` at this leftover-disposal sequence so transfer-time
   source restoration still forbids clean, while post-merge primary
   leftover cleanup is owned here.
6. Pin the new instruction language with focused Go tests.
7. Validate, curate repository memory, and open one ready pull request
   for issue #176.

Execute as one supervisor lane because the change is tightly coupled
across generated prompt text, two durable rulesets, and matching tests.
Splitting would create high-overlap contract drift.

## DECISIONS

- Keep write-on-main plus worktree copy as the intended agent workflow.
  Do not retry GH-160-style worktree pre-creation as the leftover fix.
- Use the exact command `git clean -fd` after merge, before pull, on the
  primary default-branch checkout only.
- Pair that command with a scoped restore of leftover command-owned
  unstaged tracked paths when those leftovers would still block pull.
  `git clean -fd` alone cannot remove tracked modifications.
- Abort instead of running a blanket clean when unrelated untracked or
  dirty state is present.
- Treat this as delivery leftover disposal, not implementation on the
  primary checkout and not a new merge-authorization path.
- Do not add an instruction version `v9`. The generated reconcile prompt
  and pointer-loaded delivery rules are the operator-facing contract.
  Frozen `v8` still forbids clean to create or clear a worktree.

## DISCOVERIES

- `kit spec` rejects a leading numeric prefix in the slug; `post-merge-primary-clean` is the accepted form and Kit assigned `0072`.
- `kit spec` rewrote every `PROJECT_PROGRESS_SUMMARY.md` created date. That churn was reverted; only the `0072` row and feature summary were added.
- Frozen instruction versions `v1` through `v8` still forbid clean to create or clear a worktree. The operator-facing change lives in generated managed-file prompts and pointer-loaded delivery rules, so a `v9` snapshot is not required.
- `git clean -fd` removes untracked leftovers only. The generated instructions also restore leftover unstaged tracked command-owned paths when those would still block pull.

## VALIDATION

- PASS: focused `go test ./pkg/cli -count=1 -run 'TestManagedFileDelivery|TestGitHubPRDelivery|TestWorkLaneGating'`
- PASS: `go test ./pkg/cli ./internal/templates ./internal/instructions -count=1`
- PASS: `go test -race ./pkg/cli -count=1 -run 'TestManagedFileDelivery|TestGitHubPRDelivery|TestWorkLaneGating'`
- PASS: `go vet ./pkg/cli ./internal/templates ./internal/instructions`
- PASS: `golangci-lint run --new-from-rev=origin/main ./pkg/cli ./internal/templates ./internal/instructions`
- PASS: `git diff --check`
- PASS: source-file-size audit of `pkg/cli/prompt_rules.go` (96), `pkg/cli/managed_file_delivery_test.go` (269), and `pkg/cli/rules_github_delivery_test.go` (103)
- PASS: `make build`; `./bin/kit check 0072-post-merge-primary-clean`; `./bin/kit check --project`
- NOT_APPLICABLE: browser, end-to-end, live-integration, and production suites; this change is generated-instruction and ruleset scoped
- PENDING: hosted GitHub checks until the pull request exists

## OUTCOME

Generated managed-file delivery instructions now tell the coding agent to run
`git clean -fd` on the primary default-branch checkout after remaining PR
feedback is addressed and the worktree PR has merged, then pull. Durable
`github-pr-delivery` and `work-lane-gating` rules agree. Worktree pre-creation
was not retried.

## REPOSITORY MEMORY

- Decision: updated
- Rationale: Leftover untracked reconcile files on primary `main` block pull after the worktree PR merges. The durable contract is post-merge leftover disposal, not a second attempt to pre-create the worktree. Primary remains read-only for implementation.
- Artifacts: `docs/specs/0072-post-merge-primary-clean/SPEC.md`, `docs/references/rules/github-pr-delivery.md`, `docs/references/rules/work-lane-gating.md`, `pkg/cli/prompt_rules.go`, `docs/specs/0017-reconcile-command/SPEC.md`
- Constitution: unchanged. Primary-checkout implementation remains read-only; leftover disposal after an authorized merge is a delivery-rule exception, not a new project-wide product invariant.
