---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: implement
feature:
  id: "0077"
  slug: auto-assign-initiator
  dir: 0077-auto-assign-initiator
relationships:
  - type: builds_on
    target: 0068-human-authorship
  - type: related_to
    target: 0000_INIT_PROJECT
references:
  - id: github-workflows-generator
    name: Auto-assign workflow generator
    type: code
    target: internal/templates/github_workflows.go
    relation: informs
    read_policy: must
    used_for: current static-assignee generation contract
    status: active
  - id: human-authorship
    name: Human authorship
    type: rule
    target: docs/references/rules/human-authorship.md
    relation: constrains
    read_policy: conditional
    used_for: human-only assignment constraint
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Assign the human who opened an issue or pull request in addition to
configured maintainers, instead of always assigning only the static
`github.default_assignees` list (today always `jamesonstone` in this
repository).

## CONTEXT

- `internal/templates/github_workflows.go` (`BuildAutoAssignWorkflow`)
  generates `.github/workflows/auto-assign.yml` from
  `github.default_assignees` (project `.kit.yaml` first, global config
  fallback).
- The generated script sets `const assignees = [...]` and calls
  `github.rest.issues.addAssignees` with only that static list. It never
  reads the triggering author, and it no-ops when the list is empty.
- Checked-in `.github/workflows/auto-assign.yml` currently contains only
  `"jamesonstone"`, so issues/PRs opened by any other human are still
  assigned to `jamesonstone` and never to their initiator.
- Agent delivery rules (`github-pr-delivery`, `human-authorship`,
  generated `AGENTS.md`/`CLAUDE.md` guardrails) already assign
  agent-created issues/PRs to the human user via `gh ... --assignee @me`.
  That path is initiator-aware and needs no change; the gap is only the CI
  auto-assign workflow.
- Issue #205, branch `GH-205`, and the canonical non-primary worktree are
  the authorized delivery lane. Merge is not authorized.

## REQUIREMENTS

### Initiator Plus Maintainers

- Resolve the initiator from the triggering payload for both events:
  `issues` (`context.payload.issue.user.login`) and `pull_request_target`
  (`context.payload.pull_request.user.login`).
- Final assignee set is configured maintainers plus the initiator,
  deduplicated case-insensitively, empty entries removed, `[bot]` authors
  skipped.
- When no configured maintainers exist, still assign the initiator instead
  of no-op. Only skip when neither a maintainer nor a human initiator
  resolves.
- Keep `pull_request_target` without `actions/checkout`, keep
  `continue-on-error: true`, and keep `issues: write` /
  `pull-requests: read` permissions unchanged.

### Observable Acceptance

- Issues opened by a non-maintainer are assigned to that opener plus
  configured maintainers.
- PRs opened by a non-maintainer are assigned the same way.
- Reopened events behave like opened events.
- Bot-opened events do not add the bot as assignee.
- Empty configured list still assigns the human opener.
- `go test ./internal/templates/... ./pkg/cli/...` passes; regenerated
  `.github/workflows/auto-assign.yml` contains initiator resolution and the
  configured maintainer.

### Non-Goals

- No change to agent `gh issue create --assignee @me` / `gh pr create`
  behavior; no change to `human-authorship` or `github-pr-delivery` text.
- No change to historical `docs/specs/0000_INIT_PROJECT.md` wording.
- No permission, trigger, or checkout-model change beyond assignment logic.

## ACCEPTED PLAN

1. Update `BuildAutoAssignWorkflow` to emit initiator resolution plus
   configured maintainers with dedupe and bot exclusion.
2. Regenerate the checked-in `.github/workflows/auto-assign.yml` via the
   generator path (`kit init --refresh` semantics) in the `GH-205`
   worktree.
3. Update focused tests: generator author/dedupe/bot/empty-config coverage
   and init refresh expectations.
4. Run focused Go tests, vet, fmt check, and diff review; deliver one ready
   PR from `GH-205` to `main` assigned to the human user.

## DECISIONS

- Additive (initiator + maintainers) rather than replacement so existing
  maintainer triage is preserved; the initiator is added, not substituted.
- Case-insensitive dedupe because GitHub logins are case-insensitive and
  the initiator may already be a configured maintainer.
- Skip `[bot]` initiator logins so automation authors do not become
  assignees; configured maintainers are left as configured.

## DISCOVERIES

- Global `~/.config/kit/.kit.yaml` supplies `github.default_assignees:
  [jamesonstone]`; repo-local `.kit.yaml` has no `github:` section, so the
  checked-in workflow bakes in the global fallback.
- `TestBuildAutoAssignWorkflowNoOpsWithoutAssignees` encodes the old
  empty-list no-op; it must be updated to the initiator-still-assigns
  contract.

## VALIDATION

- `go test ./internal/templates/ -count=1` passed.
- `go test ./pkg/cli/ ./internal/templates/ -count=1` passed (full
  affected packages).
- `gofmt -l`, `go vet ./internal/templates/ ./pkg/cli/`, `go build ./...`,
  and `git diff --check` passed.
- Node simulation of the emitted assignment loop verified: initiator added,
  same-login dedupe (including case-insensitive), `[bot]` initiator
  skipped, empty configured list still assigns the human initiator, null
  author resolves to maintainers only.
- Regenerated `.github/workflows/auto-assign.yml` via
  `kit init --refresh --file .github/workflows/auto-assign.yml` (global
  fallback `jamesonstone` preserved plus initiator logic); YAML parses.
- Affected source/test files audited ≤300 lines: `github_workflows.go`
  (83), `templates_test.go` (281), `init_configuration_test.go` (285),
  `init_auto_assign.go` (83).

## OUTCOME

- `BuildAutoAssignWorkflow` now emits `configured` plus initiator
  resolution with case-insensitive dedupe and `[bot]` exclusion.
- Checked-in `.github/workflows/auto-assign.yml` assigns the initiator and
  configured maintainers without checkout or permission changes.
- Focused tests updated to the initiator contract; empty configured list is
  initiator-only rather than no-op.
- Agent `gh ... --assignee @me` rules unchanged (already initiator-aware);
  no `human-authorship` or `github-pr-delivery` text change required.
- Delivery at issue #205, branch `GH-205`, ready pull request. Merge is not
  authorized.

## REPOSITORY MEMORY

- **Decision:** created
- **Rationale:** Initiator-plus-maintainer auto-assignment is a reusable
  delivery policy future agents need.
- **Artifacts:** `docs/specs/0077-auto-assign-initiator/SPEC.md`
