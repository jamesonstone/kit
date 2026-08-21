---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0068"
  slug: human-authorship
  dir: 0068-human-authorship
relationships:
  - type: builds_on
    target: 0017-reconcile-command
  - type: related_to
    target: 0063-explicit-work-lane-choice
  - type: related_to
    target: 0067-agent-completion-output
references:
  - id: safety-guardrails
    name: Git author identity guardrail
    type: rule
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: existing human git author and committer identity checks
    status: active
  - id: github-pr-delivery
    name: GitHub pull-request delivery
    type: rule
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: existing commit trailer and PR-body attribution prohibitions
    status: active
  - id: registry-adoption
    name: Downstream ruleset adoption coverage
    type: code
    target: pkg/cli/init_refresh_ruleset_adoption_test.go
    relation: supports
    read_policy: must
    used_for: refresh, health, and reconcile propagation
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make displayed authorship human-only in every Kit-managed project. Coding
agents may implement the work, but commits, pull requests, issues, comments,
trailers, and other attribution surfaces must show only the human user.

## CONTEXT

- `safety-guardrails` and `github-pr-delivery` already require a human git
  author and committer and forbid some agent trailers. Those clauses are
  delivery-specific and do not cover every attribution surface.
- Plugin defaults still insert `Co-authored-by`, "Generated with …", and
  similar credits. The project contract must forbid those credits wherever an
  agent would write them.
- Kit distributes downstream policy through registry rulesets on GitHub
  `main`. `kit init --refresh`, `kit health`, and `kit reconcile --include-files`
  install active downstream rulesets that are missing locally.
- The full authorship contract must stay pointer-loaded. Always-loaded
  instruction files should route to the ruleset instead of inlining it.
- Issue #166, branch `GH-166`, and the canonical non-primary worktree are the
  authorized delivery lane. Merge is not authorized.

## REQUIREMENTS

### Human-Only Displayed Authorship

- Add an active downstream ruleset named `human-authorship` with
  `read_policy_default: conditional`.
- Apply it to git author/committer identity, GitHub authors and assignees,
  commit and PR/issue text, review comments, trailers, badges, file headers,
  and equivalent generated-by credits.
- Do not load it for ordinary implementation or documentation edits that do
  not set authorship or insert credit.
- Never set a coding agent, assistant, bot, or tool as author, committer,
  co-author, or credited generator.
- Human DCO `Signed-off-by` trailers remain allowed; agent or tool
  `Signed-off-by` trailers are not.

### Distribution And Context Budget

- Mark the ruleset `registry_scope: downstream` so it is visible to projects
  and installed by refresh, health, and `kit reconcile --include-files`.
- Point `github-pr-delivery` and `safety-guardrails` at the ruleset instead of
  duplicating a third complete copy.
- Route load timing through RLM and the references index.
- Do not bump immutable `kit instructions` versions solely to inline the full
  rule into always-loaded provider files.
- Do not add a command, flag, or JSON schema.

### Observable Acceptance

- The rule parses as a valid conditional downstream ruleset.
- Refresh/health/reconcile adoption coverage installs it as managed.
- Generated and checked-in RLM and references indices route to it.
- Adjacent delivery and safety rules name it as the authorship contract.
- Focused tests cover ruleset semantics, pointers, and registry adoption.

## ACCEPTED PLAN

1. Add `docs/references/rules/human-authorship.md` as the canonical downstream
   ruleset.
2. Install it through the mandatory downstream adoption inventory and the
   health/reconcile managed-safety stub.
3. Add load pointers in `github-pr-delivery`, `safety-guardrails`, RLM,
   references indices, Constitution baseline, and implementation-delivery as
   optional workflow evidence.
4. Add focused ruleset, adoption, template, and reconcile-drift tests without
   exceeding the 300-line source-file limit.
5. Record the living spec and progress summary, then deliver issue #166
   through one ready pull request.

## DECISIONS

- Keep a dedicated ruleset rather than expanding `safety-guardrails` into a
  catch-all. Displayed authorship is independently reusable and should load
  only at attribution boundaries.
- Use `conditional` rather than `must` so ordinary coding turns do not pay the
  full rule. RLM, delivery, and safety pointers make the load timing explicit.
- Publish the rule from `jamesonstone/kit` rather than a downstream project.
  Downstream custom copies would not enter the GitHub registry that reconcile
  fetches.
- Preserve immutable instruction versions. A one-line RLM/constitution pointer
  plus the ruleset is enough without a new instruction version.

## DISCOVERIES

- All active downstream registry rulesets on Kit `main` are adopted by
  refresh; the focused mandatory-rules test is the explicit regression
  inventory.
- Context workflows are stored twice: embedded templates and checked-in
  mirrors under `docs/references/workflows/`.
- A branch-local downstream rule is `untracked` in live `kit rules list`
  until it lands on the registry branch. Focused registry stubs prove
  post-landing propagation without fabricating live GitHub `main` state.
- `limina-dev/sortr` PR #42 added a project-local copy by mistake. The
  canonical rule belongs in this repository.

## VALIDATION

- `go test ./pkg/cli ./internal/templates ./internal/instructions ./internal/context -count=1` passed.
- `make fmt`, `go vet ./pkg/cli ./internal/templates`, `make build`, and `git diff --check` passed.
- `./bin/kit check 0068-human-authorship` and `./bin/kit check --project` passed.
- Branch-local `kit rules list` is expected to show the new file as untracked until GitHub `main` publishes the registry listing.

## OUTCOME

- Added the conditional downstream `human-authorship` ruleset covering git, GitHub, commit, pull-request, issue, and other attribution surfaces.
- Refresh/health/reconcile adoption coverage installs it as a managed downstream rule.
- RLM, the references index, Constitution baseline, `github-pr-delivery`, `safety-guardrails`, and optional `implementation-delivery` evidence now route to it.
- Immutable instruction versions were not bumped. The full contract remains pointer-loaded.
- Delivery remains at the authorized issue #166 branch and ready-pull-request boundary. Merge is not authorized.

## REPOSITORY MEMORY

- **Decision:** created
- **Rationale:** Human-only displayed authorship is a reusable downstream
  contract and a project-wide invariant.
- **Artifacts:** `docs/specs/0068-human-authorship/SPEC.md`,
  `docs/references/rules/human-authorship.md`
