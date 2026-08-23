---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0070"
  slug: "deadline-mode"
  dir: "0070-deadline-mode"
relationships:
  - type: builds_on
    target: 0069-multi-agent-orchestration-evaluation-gate
  - type: related_to
    target: 0066-capability-aware-subagent-workflows
  - type: related_to
    target: 0067-agent-completion-output
references:
  - id: testing-and-environment-validation
    name: Testing and environment validation contract
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: constrains
    read_policy: must
    used_for: existing "run complete suite before handoff" default this rule explicitly supersedes when active
    status: active
  - id: agent-team-orchestration
    name: Capability-aware supervisor/child-agent contract
    type: rule
    target: docs/references/rules/agent-team-orchestration.md
    relation: constrains
    read_policy: must
    used_for: reused lifecycle states, independent-verifier requirement, and the two bundled hardening additions
    status: active
  - id: agent-completion-output
    name: Terminal reporting evidence-state contract
    type: rule
    target: docs/references/rules/agent-completion-output.md
    relation: constrains
    read_policy: must
    used_for: literal PARTIAL/SKIPPED reporting discipline for deferred validation
    status: active
  - id: work-lane-gating
    name: Scope-expiry idiom precedent
    type: rule
    target: docs/references/rules/work-lane-gating.md
    relation: supports
    read_policy: must
    used_for: "reused idiom: one authorization covers the accepted unit of work; ask again for materially new scope"
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

The user shared `deadline-agent-architecture.md`, a document they fed to Codex
under a real deadline on a separate project, combining subagent delegation
with an explicit "deadline mode" that safely narrowed testing/validation
scope and implementation complexity without weakening hard invariants. They
reported this was dramatically more usage-efficient (roughly 35% of their
usage limit consumed in 5 hours before adopting it, vs. 15% in the following
14 hours after) and asked kit to capture the same effectiveness. Subagent
delegation is already covered by `0069-multi-agent-orchestration-evaluation-gate`;
this feature is scoped to the two genuinely new ingredients the user
described: reducing testing/validation scope to only what's necessary when
necessary, and pausing implementation-complexity growth to focus on the
declared goal.

## CONTEXT

- `docs/references/rules/testing-and-environment-validation.md` states,
  unconditionally: "During development, run focused tests for fast feedback.
  Before handoff, run the complete applicable code-level suite and record any
  genuine blocker," and its Anti-Patterns forbid "Claiming 100 percent
  correctness ... from partial or unobserved evidence." A repo-wide grep for
  "deadline", "urgency", "time-constrained", "reduced scope", "testing
  budget" found zero relevant matches anywhere in kit before this feature.
- `docs/references/rules/agent-team-orchestration.md`'s
  `### Host-Owned Capacity And File Overlap` section only said lanes editing
  the same file/interface should serialize -- it did not name migration
  registries, contracts-under-active-revision, deployment state, runtime
  authority, or mutable external resources as explicit serialization
  triggers. There was no "reconcile a handoff against live state, don't trust
  its narrative" rule anywhere for ordinary single-repo subagent handoffs --
  only `docs/references/rules/cross-repository-program-coordination.md`'s
  `### Resume, Handoff, And Reconciliation`, scoped to multi-repo programs.
- `pkg/cli/agent_team_orchestration_rules_test.go`'s
  `TestAgentTeamOrchestrationFallbackOrderAndNeutrality` forbids the literal
  substrings `codex`, `claude`, `copilot`, `warp`, `openai`, `anthropic`,
  `gpt-`, `opus`, `sonnet`, `haiku`, `luna`, `terra` (case-insensitive) inside
  `agent-team-orchestration.md`, and forbids fixed numeric concurrency-cap
  lines via regex. Both bundled additions to that file respect this.
- Kit's existing 6 capability profiles and its existing lifecycle
  (`SCOPED -> ... -> PLAN_READY -> ... -> PR_READY -> MERGE_AUTHORIZED ->
  MERGED -> RELEASE_VERIFIED -> ... -> COMPLETE`) already cover the deadline
  document's 5 roles and its separate task-state vocabulary; this feature
  reuses kit's own terms throughout instead of importing the document's
  bespoke vocabulary (`GATE-SOURCE-HANDOFF`, `PLANNED -> MATERIALIZED -> ...`,
  literal `Luna`/`Terra`/`Sol` names).
- `0068-human-authorship` (commit `4d8fdae`) is the exact structural
  precedent: a new, conditional, pointer-loaded rule needs no
  CLAUDE.md/AGENTS.md/copilot-instructions.md Hard Gate, no
  `docs/agents/GUARDRAILS.md` touch, and no immutable instruction version
  bump -- unlike `0069`'s Hard Gate, deadline mode is inherently opt-in.
- PR #169 (issue #168) was still open, not merged, when this work started,
  and this feature edits the same file (`agent-team-orchestration.md`) #169
  already modified. This branch (`GH-170`) was created from `origin/GH-168`,
  not `origin/main`, to build on top of #169's not-yet-merged content rather
  than fork a stale pre-#169 copy. `docs/specs/` on `origin/main` topped out
  at `0068` at feature-start time, with `0069` claimed by the still-open
  PR #169, so this feature used `0070`.
- Issue #170, branch `GH-170`, and the canonical non-primary worktree are the
  authorized delivery lane. Merge is not authorized, and is additionally
  gated on PR #169 merging first since this branch depends on it.

## REQUIREMENTS

### New Conditional Rule: `deadline-mode`

- Load only on an explicit, in-thread user-signaled deadline or time
  constraint; never infer it from repository signals, task size, or calendar
  proximity, and never proactively suggest it.
- Enter only at a safe checkpoint, not mid-mutation or with an open
  verification gap.
- State a priority ordering, a stop-doing list, and a continue-doing
  (never-weaken) list that cross-references kit's actual rules by name
  wherever one exists (`github-pr-merge`, `infrastructure-change-approval`,
  `agent-team-orchestration`'s independent verifier,
  `testing-and-environment-validation`'s post-deployment tests,
  `deletion-safety`'s soft-delete-by-default).
- State a deadline testing budget (per-PR / per-wave / post-deployment /
  activation-readiness) reframed against `testing-and-environment-validation`'s
  own existing sections, keeping "Do not substitute repeated broad suites for
  focused evidence" verbatim.
- Require deadline mode's narrowing to be an explicit, recorded,
  invariant-preserving supersession of `testing-and-environment-validation`'s
  complete-suite-before-handoff default -- never silent -- with deferred
  validation reported as `PARTIAL`/`SKIPPED` per `agent-completion-output`,
  never as `PASS`.
- Scope-expire one authorization to its declared unit of work, reusing
  `work-lane-gating`'s established "ask again for materially new scope"
  idiom.

### Non-Goals

- No CLAUDE.md/AGENTS.md/.github/copilot-instructions.md Hard Gate.
- No `docs/agents/GUARDRAILS.md` touch.
- No `docs/CONSTITUTION.md` touch.
- No new immutable instruction version.
- No change to the ruleset's `read_policy_default` philosophy elsewhere in
  the repo (`deadline-mode` itself is `conditional`, matching every other
  opt-in rule).

### Bundled Orchestration Hardening (explicit user decision to bundle rather
than defer to a follow-up)

- Extend `agent-team-orchestration.md`'s `### Host-Owned Capacity And File
  Overlap` with explicit concurrency-serialization triggers (migration
  registry, contract under active revision, deployment state, runtime
  authority, mutable external resource), respecting the file's forbidden-term
  and no-fixed-numeric-cap test constraints.
- Add a new `### Handoff Reconciliation` rule requiring a coding agent to
  reconcile a subagent's handoff against live worktree/diff, git/GitHub
  state, the canonical plan, and artifact/deployment evidence rather than
  trusting its narrative alone, scoped to ordinary single-repo handoffs
  (complementing, not duplicating, `cross-repository-program-coordination`'s
  multi-repo-program version of the same principle).

### Observable Acceptance

- `docs/references/rules/deadline-mode.md` exists, parses as a valid
  conditional downstream ruleset, and contains every required section.
- `testing-and-environment-validation.md`'s default is explicitly and
  non-contradictorily superseded only by an active `deadline-mode`
  authorization.
- `docs/agents/RLM.md`, `docs/references/README.md`, and
  `docs/references/workflows/implementation-delivery.md` (both copies) all
  route to the new rule.
- `agent-team-orchestration.md` gains the concurrency-serialization
  vocabulary and the Handoff Reconciliation rule, both passing the existing
  forbidden-terms/no-fixed-numeric-cap tests.
- Zero references to `deadline-mode` exist in CLAUDE.md, AGENTS.md,
  .github/copilot-instructions.md, docs/agents/GUARDRAILS.md,
  docs/CONSTITUTION.md, or any `internal/instructions/versions/*.md`.
- New/extended regression coverage mirrors the `human-authorship` three-file
  test pattern.

## ACCEPTED PLAN

1. Write `docs/references/rules/deadline-mode.md` complete.
2. Add the one-sentence, non-contradictory supersession pointer to
   `testing-and-environment-validation.md`.
3. Add the concurrency-serialization-triggers bullet and the new
   `### Handoff Reconciliation` subsection to `agent-team-orchestration.md`.
4. Wire `docs/agents/RLM.md` (+ `agentsRLM` Go constant),
   `docs/references/README.md` (+ `referencesREADME` Go constant, hand-edited
   directly rather than regenerated -- see DISCOVERIES), and
   `docs/references/workflows/implementation-delivery.md` (+ template
   mirror) as `required: false`.
5. Scaffold `docs/specs/0070-deadline-mode/SPEC.md` via `kit spec
   deadline-mode` and populate it.
6. Add `pkg/cli/deadline_mode_rules_test.go`,
   `pkg/cli/deadline_mode_reconcile_test.go`,
   `internal/templates/deadline_mode_instructions_test.go`; extend
   `pkg/cli/agent_team_orchestration_rules_test.go`,
   `pkg/cli/reconcile_guidance_expectations.go`, and
   `pkg/cli/init_refresh_ruleset_adoption_test.go`.
7. Manually correct the recurring `docs/PROJECT_PROGRESS_SUMMARY.md`
   mtime-corruption regression (see DISCOVERIES) and hand-add this feature's
   own row and detail subsection.
8. Validate, deliver issue #170 through one ready pull request based on
   `GH-168`.

## DECISIONS

- Structural template: `0068-human-authorship`, not `0069`'s Hard Gate
  pattern. Deadline mode is inherently opt-in (explicit user signal only),
  so it gets no Hard Gate, no GUARDRAILS.md touch, no CONSTITUTION.md route,
  and no immutable instruction version bump -- matching every other
  conditional, pointer-loaded rule (`deletion-safety`,
  `infrastructure-change-approval`, `cross-repository-program-coordination`),
  not the two near-universal rules that do get a CONSTITUTION.md baseline
  bullet (`human-authorship`, `agent-completion-output`).
- Bundle the concurrency-serialization-triggers and Handoff Reconciliation
  additions into this same change rather than deferring them to a follow-up,
  per explicit user choice after being offered both options with a
  recommended default of deferring (the recommendation cited PR #169 being
  unmerged and touching the same file as a conflict-risk reason to wait).
  The user chose to bundle anyway.
- Reuse kit's existing lifecycle/profile vocabulary throughout instead of
  importing the deadline document's own task-state and role vocabulary --
  same "no second competing vocabulary" decision already made and
  test-enforced for `0069`.
- Hand-edit `docs/references/README.md` directly instead of regenerating it
  via `kit init --refresh --force`, after a `--dry-run --diff` preview showed
  that command would silently delete hand-maintained content not present in
  the Go template: the entire "Ruleset Index" table, a
  `rules/agent-team-orchestration.md` usage bullet, a `kit rules list`
  bullet, and an "AWS targets" phrase elsewhere in the same file. Only
  `docs/agents/RLM.md` was safe to regenerate (its dry-run diff contained
  only the intended new line), so that file was applied via the scoped
  `kit init --refresh --force --file docs/agents/RLM.md` command; `README.md`
  was edited by hand for both the usage bullet and the Ruleset Index row.
- Base branch `GH-170` on `origin/GH-168`, not `origin/main`, since this
  feature edits content PR #169 introduced that hadn't merged yet. Delivery
  of this PR is therefore gated on #169 merging first.

## DISCOVERIES

- `docs/references/README.md` is only partially Go-template-generated. Its
  `referencesREADME` constant covers the "Purpose" bullet list, but the
  "Ruleset Index" table and at least two of its bullets are hand-maintained
  content that predates or diverges from the current template and would be
  silently destroyed by an unscoped `kit init --refresh --force` on that
  file. This is a latent risk for any future change to this file, not
  something this feature caused or should fix.
- `kit spec`'s `docs/PROJECT_PROGRESS_SUMMARY.md` regeneration sources its
  `CREATED` column from each spec file's filesystem mtime
  (`internal/feature/feature.go:139`), which a fresh `git worktree add`
  checkout resets for every historical file. This recurred a second time in
  this same session (previously found and fixed during `0069`) --
  confirming it is a real, repeatable latent bug in kit itself, not a
  one-off. Reverted and hand-fixed again rather than shipping the corruption;
  still worth a dedicated follow-up issue against kit itself.
- `pkg/cli/init_refresh_ruleset_adoption_test.go`'s mandatory-downstream
  sample list is illustrative, not exhaustive (`agent-team-orchestration` was
  never added to it despite being a real conditional rule) -- confirmed
  again this session. Followed the more specific, more recent precedent
  (`human-authorship`, which was added) over the older, coarser one.

## VALIDATION

- `go test ./...` passed across every package on the first run after each
  new/extended test file, including
  `TestDeadlineModeRegistryRulesetIsValid`,
  `TestDeadlineModeIsIntegratedWithRelatedRules`,
  `TestDeadlineModeDoesNotAddHardGateOrConstitutionRoute`,
  `TestReconcileFindsStaleDeadlineModeGuidance` (all 4 subtests),
  `TestInstructionTemplatesRouteDeadlineModeRule`,
  `TestImplementationDeliverySelectsDeadlineModeOptionally`,
  `TestAgentTeamOrchestrationRegistryRulesetIsCapabilityAware`,
  `TestAgentTeamOrchestrationFallbackOrderAndNeutrality`, and
  `TestRunInitRefresh_InstallsMandatoryDownstreamRules`.
- `make fmt`, `go vet ./...`, and `make build` passed with no unexpected
  changes.
- `golangci-lint run --new-from-rev=origin/GH-168 ./...` reported 0 issues.
- `git diff --check` reported no whitespace errors.
- `gitleaks detect --source . --no-git` found no leaks.
- A grep for `deadline-mode` (case-insensitive) across `CLAUDE.md`,
  `AGENTS.md`, `.github/copilot-instructions.md`,
  `docs/agents/GUARDRAILS.md`, `docs/CONSTITUTION.md`, and every
  `internal/instructions/versions/*.md` found zero matches, confirming this
  stayed a conditional, pointer-loaded rule with no Hard Gate and no version
  bump.
- `kit reconcile --all --dry-run --diff --include-files` confirmed the only
  expected drift was `agent-team-orchestration` and
  `testing-and-environment-validation` both correctly flipping to
  `local-custom` state (this feature intentionally strengthens both ahead of
  the still-unmerged upstream registry copies); pre-existing, unrelated
  upstream drift on `github-pr-delivery`, `safety-guardrails`, and a missing
  `human-authorship` registry entry was surfaced and deliberately left
  untouched, matching the same pattern observed during `0069`.
- `kit check 0070-deadline-mode` and `kit check --project` both passed after
  this OUTCOME section was populated.

## OUTCOME

- Added `docs/references/rules/deadline-mode.md`: a new, conditional,
  pointer-loaded ruleset with a priority ordering, a stop-doing/continue-doing
  invariant-preserving split cross-referencing kit's real rules by name, a
  deadline testing budget reframed against
  `testing-and-environment-validation`'s own sections, an explicit
  non-contradictory supersession clause, and a scope-expiry clause reusing
  `work-lane-gating`'s established idiom.
- Added the one-sentence supersession pointer to
  `testing-and-environment-validation.md`'s complete-suite-before-handoff
  default.
- Wired the new rule through `docs/agents/RLM.md` (+ `agentsRLM` constant),
  `docs/references/README.md` (+ `referencesREADME` constant and a hand-added
  Ruleset Index row), and `docs/references/workflows/implementation-delivery.md`
  (both copies) as `required: false`.
- Bundled both requested orchestration-hardening additions into
  `agent-team-orchestration.md`: explicit concurrency-serialization triggers
  in `### Host-Owned Capacity And File Overlap`, and a new
  `### Handoff Reconciliation` rule, both passing the file's existing
  forbidden-terms and no-fixed-numeric-cap tests unchanged.
- Confirmed, by direct grep, that no Hard Gate, GUARDRAILS.md touch,
  CONSTITUTION.md touch, or immutable instruction version bump was
  introduced -- matching the accepted plan's non-goals exactly.
- Added new focused test coverage mirroring the `human-authorship`
  three-file pattern, plus a fourth negative test
  (`TestDeadlineModeDoesNotAddHardGateOrConstitutionRoute`) not present in
  that precedent, added because this feature's core design constraint
  (staying conditional/opt-in) is itself worth regression-testing directly.
- Manually corrected a second occurrence of the `PROJECT_PROGRESS_SUMMARY.md`
  mtime-corruption regression rather than shipping it.
- Delivery remains at the authorized issue #170 branch (`GH-170`, based on
  `GH-168`) and ready-pull-request boundary. Merge is not authorized, and is
  additionally sequenced behind PR #169 merging first.

## REPOSITORY MEMORY

- **Decision:** created
- **Rationale:** This feature adds a new project-wide capability (an
  explicit, invariant-preserving way to narrow validation scope under a real
  deadline), depends on non-obvious discoveries (the partially-templated
  references/README.md, the recurring PROJECT_PROGRESS_SUMMARY.md mtime bug)
  future maintainers need without re-deriving them from source, and records
  an explicit user decision (bundle vs. defer the orchestration-hardening
  additions) that shaped the final scope.
- **Artifacts:** `docs/specs/0070-deadline-mode/SPEC.md`,
  `docs/references/rules/deadline-mode.md`,
  `docs/references/rules/agent-team-orchestration.md`
