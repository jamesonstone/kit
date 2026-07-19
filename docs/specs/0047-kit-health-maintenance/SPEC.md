---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0047
  slug: kit-health-maintenance
  dir: 0047-kit-health-maintenance
relationships:
  - type: builds_on
    target: 0033-kit-capabilities
  - type: builds_on
    target: 0045-constitution-curation
  - type: related_to
    target: 0046-autonomous-mutation-recovery
skills: []
references:
  - id: status-managed
    name: Kit-managed status
    type: code
    target: pkg/cli/status_kit_managed.go
    relation: uses
    read_policy: must
    used_for: bounded registry freshness planning and stable state vocabulary
    status: active
  - id: reconcile
    name: Reconcile command
    type: code
    target: pkg/cli/reconcile.go
    relation: uses
    read_policy: must
    used_for: safe managed-file and ruleset application
    status: active
  - id: project-check
    name: Project contract check
    type: code
    target: pkg/cli/check.go
    relation: uses
    read_policy: must
    used_for: post-refresh project health validation
    status: active
  - id: config
    name: Kit project configuration
    type: code
    target: internal/config/config.go
    relation: implements
    read_policy: must
    used_for: nullable automated-health management policy
    status: active
  - id: github-delivery
    name: GitHub PR delivery rule
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: guides
    read_policy: must
    used_for: same-PR additional-issue traceability
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Keep Kit-managed rules, instructions, configuration, scaffold files, and project health current through a focused maintenance workflow without making ordinary implementation agents track registry churn throughout product work.

## CONTEXT

- `kit status` already performs a bounded remote registry check and reports Kit-managed drift, but its broader feature output is unnecessarily large for a scheduled maintenance gate.
- `kit reconcile --include-files` already plans and safely applies generated files and downstream registry rulesets, preserving mergeable local changes and surfacing conflicts instead of overwriting them by default.
- `kit check --project` validates the resulting repository contract but does not fetch or apply registry updates.
- The user rejected continuous mid-implementation auto-refresh because rule changes can consume coding-agent context, expand unrelated diffs, and change the governing contract after planning.
- The accepted replacement is a dedicated maintenance command plus a weekly Codex automation that isolates each affected repository in a worktree and leaves every update in a ready, unmerged pull request for human review.
- The automation must scan Git repositories beneath `/Users/jamesonstone/go/src/github.com/**` that contain `.kit.yaml`, default to managing them, and honor only an explicit project-level `false` as an opt-out.
- This feature continues existing PR `#63` on branch `GH-62`, but issue `#66` independently owns the new scope and every new commit uses `GH-66` as its Conventional Commit scope.
- Concurrent issue `#64` already uses feature id `0046`, so this feature reserves `0047` to avoid a future spec collision.

## REQUIREMENTS

- Add visible root command group `kit registry` with read-only `kit registry status`.
- `kit registry status` must perform a bounded registry freshness check, print compact human output, support stable `--json`, and expose enough state for automation without requiring the full feature status payload.
- Reuse the existing state vocabulary where applicable: `current`, `refresh_available`, `attention_needed`, and `unknown`.
- Add visible root command `kit health` as the reviewed orchestration surface for automated Kit maintenance.
- Default `kit health` must apply every safe Kit-managed file and registry update identified by the existing refresh planner, without `--force`, then rerun registry status and `kit check --project`.
- `kit health` must not mutate Git, GitHub, arbitrary application code, or unresolved local-custom/conflicted content. It must report remaining attention so the scheduled coding agent can curate the maintenance diff in the worktree and rerun validation.
- Provide `kit health --dry-run --diff` for a read-only complete preview and `kit health --json` for stable automation-facing outcome data.
- Add nullable project configuration `health.managed`. Omitted `health`, null `health`, an empty mapping, or null `health.managed` means managed. Only explicit `health.managed: false` opts the repository out.
- An opted-out `kit health` or `kit registry status` invocation must return a clear successful skipped/disabled result without network or file mutation.
- Do not materialize the opt-out field into generated project configuration when the project uses the default managed behavior.
- Update root help, capability metadata, command docs, README command discovery, config examples, and focused tests for every new command and flag.
- Update `github-pr-delivery` so unrelated or tangential scope explicitly continued on an existing PR receives a new GitHub issue, uses the new issue number in commit scopes, remains on the existing PR head branch, appends `Closes #<new>` to the PR Ticket section, and refreshes the PR description and validation list for the combined diff.
- Preserve the normal one-issue/one-branch/one-PR rule unless the user explicitly requests the same-PR exception.
- Create an active Wednesday 1:00 PM `America/New_York` Codex automation after delivery.
- The automation must use the Kit project as its execution project, scan eligible `.kit.yaml` repositories, use one isolated worktree per repository, start each maintenance branch from current `origin/main`, and create or update one human-assigned issue, exact `GH-<issue>` branch, and ready unmerged PR per changed project.
- The automation must default to applying all safe health updates, semantically curate or refactor Kit-managed files when the health result requires attention, validate the complete diff, and let the human reject individual Kit file changes during PR review.
- The automation must skip repositories only for explicit `health.managed: false`, missing `main`, unavailable GitHub identity/access, or a genuine repository-specific blocker; one blocked repository must not stop other independent repositories.
- No-op repositories must produce no file, Git, or GitHub mutation.
- Observable acceptance: compact registry status works in current/update/attention/unknown/opt-out cases; health preview and apply are deterministic and safe; opt-out semantics round-trip through YAML; capabilities and help discover both commands; the same-PR issue rule is canonical; focused/full validation passes; the recurring automation is active on the requested schedule.
- Non-goals: automatically merging maintenance PRs, changing arbitrary product code, forcing registry content over unresolved customization, silently modifying dirty primary worktrees, continuous registry polling during implementation, or making `kit check --project` networked or mutating.

## ACCEPTED PLAN

1. Add pointer-based `health.managed` configuration semantics and tests without changing the current schema version or materializing the default.
2. Refactor the existing Kit-managed status builder just enough to expose a compact registry-status projection and implement the `registry status` command with human and JSON renderers.
3. Implement `health` by reusing the existing refresh planner/apply path and project-contract checker, with dry-run/diff, JSON, opt-out, bounded network handling, and explicit remaining-attention reporting.
4. Add focused command/config/refresh tests, capability records, root-help ordering, README/command documentation, and configuration guidance.
5. Add the explicit same-PR/new-issue exception to `github-pr-delivery` and curate any demonstrated project-wide invariant after validation.
6. Run formatting, vet, focused/full/race tests, build, changed-lines lint, command discovery, V3 spec/project checks, prompt-system validation where affected, end-to-end fixture checks, and full diff/secret review.
7. Commit on `GH-62` with `GH-66` scopes, push only after confirming the live PR head has not changed, append issue `#66` and combined validation to PR #63, and keep it ready and unmerged.
8. Create and verify the Wednesday 1:00 PM Eastern cross-project Codex automation, with per-repository worktrees and open-PR-only delivery.

## DECISIONS

- Accepted a boundary maintenance workflow instead of continuous implementation-agent tracking.
- Accepted `health.managed` as a nullable pointer policy: only explicit false opts out.
- Accepted safe non-force application as the default; unresolved customization becomes reviewed agent work in the maintenance PR rather than silent overwrite.
- Accepted `kit health` as a file/project orchestration command only; Git and GitHub delivery remain owned by the scheduled coding agent under each repository's rules.
- Accepted a narrowly explicit exception for new issue-scoped commits on an existing PR branch when the user requests same-PR continuation.
- Rejected making `kit check --project` networked or mutating.

## DISCOVERIES

- `kit status` already supplies bounded registry access, deterministic state names, and safe next actions, so the compact command can reuse rather than duplicate registry planning.
- `kit reconcile --include-files` already fast-forwards managed rules, section-merges non-conflicting local changes, records conflicts, and avoids destructive force by default.
- The existing PR head branch is not checked out in the primary worktree, allowing this task to use `GH-62` directly in an isolated worktree without disturbing concurrent `GH-64` work.
- Reusing the project checker for JSON health output required a writer-aware checker seam so human diagnostics can be buffered without corrupting machine-readable stdout.
- The first end-to-end apply fixture exposed a pre-existing Constitution refresh defect: the baseline replacement was bounded by the next heading instead of its managed end marker, so repeated refreshes could remain non-idempotent and consume custom constraint text. The refresh now replaces only the marked block, preserves surrounding content, and converges after one application.
- Live registry preview on this PR branch correctly reports refresh work relative to current `main`; the health command keeps that preview read-only and does not apply unrelated base-branch refreshes during feature implementation.

## VALIDATION

- `make fmt` and `git diff --check` completed without formatting or whitespace errors.
- `go vet ./...` passed.
- `go test ./... -count=1` passed across every package.
- `go test -race ./internal/config ./pkg/cli -count=1` passed.
- `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`.
- `make build` produced `bin/kit` successfully.
- Focused configuration, registry status, health, command-discovery, writer-error, and Constitution idempotency tests passed.
- `./bin/kit capabilities health --json` and `./bin/kit capabilities registry status --json` returned the documented mutation, network, file-write, flag, and opt-out contracts.
- `./bin/kit check 0047-kit-health-maintenance` passed.
- `./bin/kit registry status --json` completed a live bounded registry preview and reported the branch's current planned changes without writes.
- `./bin/kit health --dry-run --diff` completed successfully and produced a read-only full managed-file preview.
- `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit --json` passed all 45 task runs and all 345 assertions with deterministic output across all 15 repeated tasks.
- `./bin/kit check --project` still reports two pre-existing blocking V2 metadata findings in `docs/specs/0038-auto-improvement-v1/SPEC.md`, plus existing scaffold and compatibility advisories. This feature does not touch that legacy spec, and every changed feature/code validator passes.
- Codex automation `weekly-kit-health` was created, viewed, and confirmed active for Wednesdays at 1:00 PM Eastern with the Kit project target and the user-authorized per-repository worktree contract.

## OUTCOME

- Added compact human and JSON registry freshness reporting with current, refresh, attention, unknown, and disabled states.
- Added safe default health orchestration that applies conflict-free managed refreshes, preserves conflicted or customized content, reruns registry freshness, and validates the project contract.
- Added nullable default-on health management configuration, explicit opt-out behavior, command discovery, documentation, and regression coverage.
- Corrected Constitution baseline refresh idempotency and custom-content preservation because recurring health depends on convergence.
- Curated same-PR additional-issue traceability and the explicitly authorized scheduled-health worktree exception into the delivery rules and project Constitution.
- Activated recurring automation `weekly-kit-health` and prepared the validated issue-scoped implementation for delivery on existing PR #63; no merge is part of this feature.

## REPOSITORY MEMORY

Decision: created

Rationale: The separation between ordinary implementation and scheduled Kit maintenance, nullable opt-out semantics, health orchestration boundaries, and same-PR issue traceability are durable cross-command and cross-repository contracts.

Artifacts:

- `docs/specs/0047-kit-health-maintenance/SPEC.md`
- `docs/references/rules/github-pr-delivery.md`
