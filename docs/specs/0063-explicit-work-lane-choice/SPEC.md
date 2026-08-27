---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0063"
  slug: "explicit-work-lane-choice"
  dir: "0063-explicit-work-lane-choice"
relationships:
  - type: builds_on
    target: 0050-safe-worktree-workflow
  - type: builds_on
    target: 0059-conservative-coding-agent-first
  - type: related_to
    target: 0061-authorized-coding-agent-merge-autonomy
references:
  - id: work-lane-gating
    name: Work lane gating rule
    type: rule
    target: docs/references/rules/work-lane-gating.md
    relation: constrains
    read_policy: must
    used_for: default lane routing and mutation tripwire
    status: active
  - id: safety-guardrails
    name: Safety guardrails rule
    type: rule
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: primary-checkout protection and dirty-state preservation
    status: active
  - id: github-pr-delivery
    name: GitHub pull request delivery rule
    type: rule
    target: docs/references/rules/github-pr-delivery.md
    relation: guides
    read_policy: must
    used_for: issue, branch, worktree, commit, and pull-request delivery
    status: active
  - id: github-pr-merge
    name: GitHub pull request merge rule
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: guides
    read_policy: must
    used_for: existing-PR lifecycle precedence and bounded in-place remediation
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: guides
    read_policy: must
    used_for: complete local and hosted validation
    status: active
  - id: source-size-rule
    name: Source file size rule
    type: rule
    target: docs/references/rules/source-file-size.md
    relation: constrains
    read_policy: must
    used_for: changed source and test organization
    status: active
skills: []
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Default every coding-agent repository mutation to a new worklane without
asking for a lane choice. Every change must happen in a proven writable
worktree with a concrete pull-request landing plan, while the clone's primary
checkout remains read-only and preserves user-owned state. Existing-lane
continuation remains available only when the user explicitly directs it.

## CONTEXT

- At the start of GH-186, the active rule, generated guidance, and immutable
  v10 instructions required an explicit new-versus-existing user choice before
  every repository or delivery mutation.
- The user now requires the opposite interaction contract: allocate a complete
  new worklane by default, do not ask the lane-choice question, and reserve
  questions for material implementation-intent or named-target ambiguity.
- The new default is broader than the earlier v2 clean-preflight path. It
  applies on clean, dirty, protected, and feature branches; current work is not
  continued unless the user explicitly directs continuation for the same scope.
- Existing-PR lifecycle work is already owned. Review repair, CI repair, base
  refresh, and dependency-ordered merge coordination must retain every targeted
  pull request's branch and identity instead of creating coordinator or
  recursively corrective pull requests.
- Feature 0050 established canonical linked worktrees at
  `~/worktrees/<owner>/<repository>/<lane>`, exact `GH-<issue-number>` durable
  lanes, and preservation of dirty worktree state.
- Published `kit instructions` versions v1 through v10 are immutable historical
  contracts. Additive v11 must become current rather than rewriting them.
- The historical v4-v10 gate required a fixed question and explicit answer;
  v5 and v6 added shorthand and full-form response mappings before GH-186
  superseded the prompt in active guidance.
- Historical user feedback on PR #161 showed that project coding agents treated
  `new worklane`, `new work lane`, and `new lane` as ambiguous prose instead of
  the complete new issue, exact branch, canonical worktree, and pull-request
  choice.
- GitHub issue #143, branch `GH-143`, and canonical worktree
  `~/worktrees/jamesonstone/kit/GH-143` own this change. The planned end state
  is one ready pull request targeting `main`.
- GitHub issue #186, branch `GH-186`, and canonical worktree
  `~/worktrees/jamesonstone/kit/GH-186` own the default-new-worklane reversal.
  Its planned end state is one ready pull request targeting `main`.

## REQUIREMENTS

- Before a coding agent creates, edits, deletes, moves, or generates any
  repository file, it must default to creating or reusing a new issue, exact
  issue branch, canonical worktree, and ready pull-request plan without asking.
- Read-only repository discovery, safety recon, context resolution, planning,
  and implementation-intent questions may precede lane establishment. Commands
  that write specs, generated files, configuration, source, tests,
  documentation, or Git state may not.
- Apply the default on clean, dirty, protected, and feature branches, when an
  issue is referenced, and when the request asks for a new pull request. The
  default applies only when no existing issue/branch/pull-request owner matches
  the accepted unit of work.
- Continue an existing lane only when the user explicitly directs that outcome
  for the same unit of work and recon proves its exact branch, non-primary
  owning worktree, issue scope, protected base, and pull-request target.
- Never offer or ask for a new-versus-existing lane preference. Ask only when
  implementation intent or a user-named target is materially ambiguous and
  cannot be resolved from repository evidence.
- Treat an exact existing pull-request set targeted for review repair, CI
  repair, base refresh, conflict resolution, generated-artifact refresh, or
  ordered merge coordination as explicit continuation of those existing lanes.
- Reuse every targeted head branch and pull request for scope-preserving work.
  Never create a coordinator or corrective pull request solely to update or
  make another pull request mergeable.
- For a multi-PR merge or program plan, record one continuation entry per
  target. If source repair is not authorized, request bounded in-place repair
  authority rather than allocating a new lane or replacement pull request.
- Preserve `github-pr-merge` replacement criteria: replacement requires a
  material scope or architecture change, an unsafe original head, repository
  policy, or explicit user direction.
- One recorded lane allocation covers the accepted unit of work and its required tests,
  documentation, validation fixes, and delivery. Materially new or tangential
  accepted scope defaults to another new lane; routine subtasks do not create
  another lane.
- The default new lane must establish or reuse the human-assigned issue, exact
  `GH-<issue-number>` branch, canonical linked worktree, protected base, and
  ready pull-request plan before the first repository file mutation.
- Explicit continue-existing direction is valid only when recon proves a non-protected
  writable branch, its owning worktree, its issue scope, and a plan to create
  or update that branch's pull request. Ambiguous or detached lanes fail closed.
- The clone's primary/root checkout is read-only for coding-agent work,
  regardless of branch cleanliness or file type. It may be inspected and may
  supply shared environment symlinks, but it may not receive implementation,
  documentation, generated, staging, commit, or branch-switch mutations.
- If an agent detects an ungated edit or a root-checkout edit, it must stop,
  preserve all state, report the violation, and avoid staging, committing,
  pushing, resetting, cleaning, stashing, or silently transferring the change.
  Recovery begins only after the default or explicitly continued lane and exact
  ownership are proven.
- Align the normative rules, active generated instruction templates,
  checked-in generated guidance, Constitution, current versioned agent
  instructions, and regression tests without changing immutable v1-v10 payloads.
- Keep every changed handwritten Go source and test file at or below 300
  physical lines and pass complete repository validation.
- Observable acceptance: policy tests require automatic new-worklane routing,
  reject the routine lane-choice prompt and implicit existing-lane continuation,
  require the pull-request landing plan, protect the primary checkout for every
  file mutation, preserve tripwire state, and prove additive current v11 while
  v1-v10 hashes remain unchanged.
- Non-goals: implementing a runtime filesystem interceptor, automatically
  moving pre-existing dirty changes, creating a second pull request for an
  existing approved lane, changing native Git worktree semantics, or merging
  the delivery pull request.

## ACCEPTED PLAN

1. Historical GH-143: rewrite `work-lane-gating` around one unconditional pre-mutation user choice,
   a recorded pull-request landing plan, scoped choice reuse, and a fail-closed
   tripwire with no clean-default-branch exception.
2. Strengthen safety and delivery rules so the primary checkout is read-only,
   new work uses a canonical issue worktree, and continued work is allowed only
   in a proven owned writable lane with one create-or-update PR plan.
3. Put the hard gate in the shared and current instruction generators, refresh
   checked-in active guidance, and add immutable version v4 as the new current
   `kit instructions` contract while retaining the exact v1-v3 bytes.
4. Add focused consistency tests across rules, templates, checked-in guidance,
   versioned instructions, and command capability examples; split test files if
   needed to preserve the 300-line limit.
5. Run formatting, focused tests, complete tests and race tests, vet, lint,
   builds, Kit checks, reconcile, source-size, diff, and secret validation.
   Curate the demonstrated invariant into the Constitution and this spec.
6. Explicitly stage only GH-143 files, commit as the human user, push
   `GH-143`, open one ready PR assigned to the human user, and report hosted
   checks separately from local validation.
7. For issue #155 on the user-selected existing PR lane, add concise response
   semantics to the canonical rule, active templates and mirrors, current v5
   instructions, durable feature rationale, and focused regression coverage.
8. For the GH-160 / PR #161 follow-up, map common full-form new-lane phrases
   to the complete issue/branch/worktree/PR choice in the canonical rule,
   active templates and mirrors, reconcile drift expectations, and additive
   current v6 instructions while preserving immutable v1-v5 bytes.
9. For GH-186, replace the active lane-choice prompt with unconditional default
   new-worklane routing, retain only explicitly directed existing-lane
   continuation, update managed-file and reconcile drift semantics, publish
   additive current v11, and preserve immutable v1-v10 bytes.
10. Repair PR #187 in place by making existing-PR lifecycle work an explicit
    precedence rule across the canonical lane, safety, delivery, merge,
    generated-guidance, reconciliation, and v11 surfaces. Add regressions that
    reject coordinator and recursively corrective pull requests for
    scope-preserving review, CI, base-refresh, and ordered-merge work.

## DECISIONS

- Accepted: apply the gate to every coding-agent repository file mutation,
  including documentation and spec creation. A file-type exception would
  still permit root-checkout changes with no delivery lane.
- Superseded by GH-186: the earlier explicit choice was scoped to one accepted
  unit of work. The replacement preserves that scope boundary as one lane
  allocation; routine completion stays in it, while materially new accepted
  scope defaults to another new lane without prompting.
- Accepted: publish v4 as the new current versioned instruction contract.
  Rewriting v1-v3 would invalidate their documented immutability and hashes.
- Superseded by GH-186: the earlier rejection of automatic allocation reflected
  the then-current user requirement for an explicit choice. The user now
  requires automatic new-worklane allocation across every repository state;
  existing work is continued only through an explicit same-scope direction.
- Rejected: allow edits in a clean primary checkout with a later transfer plan.
  That makes user-owned root state depend on recovery and permits an ungated
  diff to exist before its pull-request lane exists.
- Accepted for the follow-up: use the first standalone `c`, `n`, or `y` token
  as the primary lane choice while preserving any trailing instructions. This
  keeps the common response concise without discarding user constraints.
- Retained for the follow-up: fail closed when trailing text contradicts the
  shorthand or no binary choice can be resolved. Shorthand reduces input
  ceremony but does not weaken the mutation gate.
- Accepted for the GH-160 follow-up: normalize `new lane`, `new work lane`,
  `new worklane`, and `new worktree` as explicit new-lane answers. Requiring
  users to repeat the longer generated question after already naming the lane
  adds friction and does not improve target authority.
- Accepted for the GH-160 follow-up: publish additive v6 instructions instead
  of changing immutable v5. Current generated guidance may advance, while
  every published version remains byte-stable.
- Accepted for GH-186: remove the lane-choice question and standalone response
  shorthand from active guidance. Those semantics remain historical in
  immutable versions, while additive v11 carries the new default.
- Accepted for GH-186: an accepted repository-mutating request authorizes the
  default issue-to-ready-PR lane. This removes a redundant permission prompt
  without weakening identity, primary-checkout, staging, validation, delivery,
  or separate merge-authorization boundaries.
- Accepted for the PR #187 repair: default-new routing applies only when the
  accepted unit has no existing issue/branch/pull-request owner. Exact
  existing-PR lifecycle targets take precedence and are explicit continuation.
- Accepted for the PR #187 repair: merge authority and in-place repair
  authority remain separate. Missing repair authority blocks head mutation but
  can never be satisfied by creating a coordinator or corrective pull request.
- Rejected during PR review: add unconditional runtime rejection to `kit init`,
  `kit health`, and init refresh. The lane gate governs coding-agent actions,
  while these commands also support direct human use and non-Git bootstrap. Kit
  has no persisted caller identity or Pull-Request Landing Plan to validate, so
  executable enforcement requires a separately designed attestation contract
  rather than silently changing these command interfaces.
- Retained during follow-up review: feature 0055's fail-visible thread
  initialization fallback. Title and pin operations are blocking, ordered
  pre-response actions, but unsupported, unavailable, or failed host operations
  must report exact status and then allow substantive work to continue. Making
  only unsupported operations non-blocking would contradict that durable
  provider-specific contract.

## DISCOVERIES

- The repository-shared feature allocator initially reserved ID 0062. Before
  delivery, refreshed `origin/main` introduced merged feature
  `0062-git-wt-removal`, so this lane advanced its shared reservation and spec
  identity to 0063 before integrating upstream.
- `kit spec` currently mutates repository memory immediately. Agents must
  establish the default or explicitly continued writable lane before invoking
  it; capability inspection may still run read-only beforehand.
- PR review found that immutable v4 scoped the lane gate inside new-session
  initialization and that the shared full template described only file writes.
  Resumed sessions and delivery mutations therefore needed explicit coverage in
  the generated instruction contract even though the canonical rule already
  required both.
- Follow-up review found that v4 had not yet inherited repository-start routing,
  thread-operation result verification, capability discovery for context
  resolution, or the separate merge-authority boundary demonstrated by
  features 0055 and 0061. These are instruction-contract gaps rather than
  changes to the work-lane policy itself.
- Final hosted review found that v4 also stopped after context selection without
  carrying forward the repository's native-planning and living-spec decision,
  testing-environment rule loads, or source-file-size gate. The published
  contract now preserves those pre-implementation boundaries explicitly.
- V2 root-instruction auditing deliberately keeps routing entrypoints thin.
  The implementation therefore uses a compact always-loaded hard gate and
  keeps the full operational contract in Guardrails and the normative ruleset.
- Existing managed-file delivery guidance normalized primary-checkout changes
  into a later transfer workflow. The corrected guidance treats such snapshots
  only as command-owned evidence, trips the gate, and refuses automatic
  transfer, staging, restoration, or delivery.
- Refreshed `origin/main` also introduced the complete authorized-merge model.
  The integrated lane contract keeps issue, branch, push, and ready-PR delivery
  distinct from merge authority and preserves the merge policy consistency
  checks without restoring automatic lane consent.
- The codified question is consumed through generated agent instructions; the
  repository has no separate executable response parser for this gate. The
  canonical template and ruleset therefore own shorthand interpretation, with
  regression tests proving every active generated surface carries it.
- Full-form lane interpretation has the same prompt-contract boundary as
  shorthand. It must be repeated in the downstream ruleset, compact and full
  templates, checked-in mirrors, current versioned instructions, and reconcile
  semantic-drift expectations or projects can remain structurally current but
  behaviorally ambiguous.

## VALIDATION

- Focused rule, template, versioned-instruction, managed-file, health, init,
  and merge-authority consistency tests passed after integrating current
  `origin/main`.
- `go test ./... -count=1` passed across every package, including
  `internal/releaseprompt`, `internal/worktreeprep`, and `pkg/cli`.
- `go test -race ./... -count=1` passed across every package.
- `make fmt`, `make vet`, `go build ./cmd/kit`, and
  `golangci-lint run --new-from-rev=origin/main ./...` passed; lint reported
  `0 issues`.
- `git diff --check origin/main...HEAD` passed.
- `kit check explicit-work-lane-choice`, `kit check --all`, and
  `kit check --project` passed; all 60 visible features and the project
  contract were coherent.
- `kit reconcile --all --output-only` reported no reconciliation needed and
  audited 668 version-control-eligible candidates and 338 eligible handwritten
  source/test files with none above 300 physical lines.
- The default `kit instructions` payload hashed to
  `96fc2b3bbd4f458ef55ae32910d737dd1ea35110d6443d6ee8e03d389d851986`,
  matching immutable v4.
- PR-review repair passed focused instruction, template, and CLI tests; complete
  tests and race tests; build, vet, and changed-code lint; feature, all-feature,
  and project checks; and whole-project reconcile. Reconcile audited 667
  version-control-eligible candidates and 338 eligible handwritten source/test
  files with none above 300 physical lines.
- The repaired working tree passed `gitleaks dir --redact --no-banner .`, which
  scanned 5.62 MB with no leaks.
- Follow-up v4 repair passed focused instruction tests, complete tests and race
  tests, build, vet, changed-code lint, all feature/project checks, and
  whole-project reconcile. Reconcile again found 0 of 338 eligible handwritten
  source/test files above 300 physical lines, and the 5.99 MB working-tree
  secret scan found no leaks.
- Final hosted-review repair passed focused and complete tests, the complete
  race suite, formatting, build, vet, changed-code lint, all 60 feature checks,
  the project-contract check, and whole-project reconcile. Reconcile found 0
  of 338 eligible handwritten source/test files above 300 physical lines, and
  the 6.53 MB working-tree secret scan found no leaks.
- Initial delivery `gitleaks dir --redact --no-banner .` scanned 5.25 MB with no leaks, and
  `gitleaks git --redact --no-banner --log-opts='origin/main..HEAD' .` scanned
  the GH-143 commit range with no leaks.
- Issue #155 shorthand follow-up passed focused template, instruction, and CLI
  tests; complete tests and race tests; formatting, build, vet, and changed-code
  lint; feature, all-feature, project, and reconcile checks. All 62 features
  passed, reconcile audited 683 version-control-eligible candidates and 349
  eligible handwritten source/test files with none above 300 physical lines,
  and the 5.43 MB working-tree secret scan found no leaks.
- The managed-propagation follow-up passed focused health, reconcile, registry,
  and existing-section drift tests; complete tests and race tests; formatting,
  build, vet, and changed-code lint; all 62 feature checks; project validation;
  and whole-project reconcile. Reconcile audited 684 version-control-eligible
  candidates and 350 eligible handwritten source/test files with none above 300
  physical lines, and the 6.16 MB working-tree secret scan found no leaks.
- Current v5 instructions hash to
  `cf68ece8fe95d51733fa835460e0788b89392d22fb4c46522c543f91f3ba6dc7`;
  immutable v1-v4 hashes remain unchanged.
- Final staged-diff and hosted pull-request checks remain delivery steps and
  will be recorded separately after local completion.
- GH-160 full-form mapping passed focused instruction, template, ruleset,
  reconcile-drift, health, and managed-refresh propagation tests; complete Go
  tests; race tests for instructions, templates, and CLI; formatting, vet,
  changed-code lint, and build.
- Built Kit checks passed for features 0063 and 0017, all 63 features, and the
  project contract. Current `kit instructions` renders the four equivalent
  full-form choices and capabilities advertise v6.
- Immutable v1-v5 instruction hashes remained unchanged. Additive v6 hashes to
  `6e46f43483957a434c6e3e7e9982f45807e499f486f21061307d74e2538f6e91`
  after integration with the additive completion-output contract.
- The whole-project managed dry-run reported a complete source-size audit of
  701 version-control-eligible candidates, 362 eligible handwritten
  source/test files, and zero violations. Its sole planned `.kit.yaml` change
  marks the branch-ahead ruleset `local-custom`; that expected pre-merge
  self-registry metadata change was not applied.
- A fresh read-only verifier found no actionable issues and confirmed canonical
  rule, compact/full template, checked-in mirror, reconcile drift, managed
  propagation, additive v6, immutable v1-v5, focused-test, diff, formatting,
  and source-size evidence.
- CodeRabbit found that the first drift assertions could pass on generic issue
  or PR wording without proving the exact branch/worktree/ready-plan clause.
  GH-160 now requires one contiguous complete-lane clause in every managed
  surface, removes that clause in per-file stale-guidance regressions, and
  passes the complete Go, race, format, vet, lint, and build validation again.
- Review of PR #187 found one precedence ambiguity: the general default-new gate
  could be read before the stronger merge rule and allocate a coordination
  worklane for an existing PR set. The repair makes exact existing-PR lifecycle
  targets explicit continuation at the always-loaded boundary.
- GH-186 focused instruction, template, ruleset, managed-file, reconciliation,
  health, and CLI tests pass. `make all`, `go test -race ./... -count=1`,
  `go build ./...`, changed-code `golangci-lint`, the Windows build, and
  `goreleaser check` pass.
- The built Kit binary passes feature 0063, all 70 feature checks, and the
  project contract. Context resolution is unblocked; whole-project reconcile
  reports no drift and audits 735 version-control-eligible candidates and 381
  eligible handwritten source/test files with zero above 300 physical lines.
- Current v11 hashes to
  `ddb2a92de00dcef09288f33532eb95164efa450ce00ef7687015b10c06c95f08`;
  immutable v10 remains
  `9c4d87348f0481b552b2dd44024ad0e3fbd82ec4a568fbb31555d9bb8de94162`.
- `gitleaks dir --redact --no-banner .` scanned 5.61 MB with no leaks, and
  `git diff --check` passes.
- Browser, live-integration, deployment, infrastructure, and production
  validation are `NOT_APPLICABLE`; this change updates local policy, templates,
  prompt generation, and embedded instructions without executing a runtime or
  provider mutation.
- `PENDING`: explicit staging, commit, push, ready pull-request creation, and
  hosted checks remain delivery steps.

## OUTCOME

- The active work-lane rule now defaults every accepted coding-agent repository
  mutation with no existing owner to a complete new
  issue/branch/worktree/ready-PR lane without asking. Explicit same-scope
  continuation and exact existing-PR lifecycle targets retain their lanes.
- The lane-choice question and response shorthand were removed from active
  guidance. One allocation covers required completion work, while materially
  new accepted scope receives another default new lane.
- Review repair, CI repair, base refresh, and ordered merge coordination reuse
  every targeted existing pull-request head. Scope-preserving work creates no
  coordinator or recursively corrective pull request.
- Safety and delivery rules now make the exact primary checkout read-only and
  accept continued work only in a proven non-primary owned worktree with one
  create-or-update pull-request route.
- Managed-file prompts preserve ungated or primary-checkout snapshots as
  evidence and no longer instruct agents to transfer and restore root changes
  automatically.
- Generated current AGENTS, Claude, Copilot, workflow, Guardrails, and Tooling
  guidance carries the default. Additive v11 is current while immutable v1-v10
  retain their published historical contracts.
- Historical v4 repair made the prior gate unconditional in newly created and
  resumed sessions, required repository context resolution before write-capable
  workflows, and ordered rule loading, recon, lane routing, and plan
  verification. GH-186 preserves that order while changing the route default.
- Follow-up repair makes thread title and pin operations ordered, verified, and
  fail-visible; starts repository work at the routing entrypoint; discovers
  unknown context-resolve behavior through capabilities; and states explicitly
  that a landing lane never authorizes a merge outside the exact bounded
  `MERGE_READY` contract.
- Final review repair completes v4's execution preconditions by requiring native
  planning, the semantic living-spec decision and accepted plan when material,
  testing-environment rule loads, and the source-file-size gate before coding.
- Regression tests enforce default new-worklane wording, absence of the routine
  lane-choice prompt, explicit-only existing continuation, root protection,
  generated and checked-in alignment, managed-command tripwire,
  current-version behavior, and immutable historical versions.
- Managed-health and `kit reconcile --include-files` regression coverage proves
  missing root/provider guidance and Guardrails are restored with the default
  lane semantics. Reconcile also reports existing-section routing drift for
  reviewed semantic curation instead of silently treating structural freshness
  as current guidance.
- The GH-160 aliases remain historical in immutable v6. Active guidance no
  longer waits for those phrases because it allocates the same complete new
  lane by default.
- Compact and full generators, checked-in AGENTS/Claude/Copilot/Guardrails,
  downstream reconcile drift detection, and managed health/reconcile refreshes
  now carry the same default. Additive v11 is current while v1-v10 remain
  immutable.
- The integrated result preserves current merge-autonomy, release-orchestration,
  and git-wt-removal policy. Pull-request delivery still never implies merge
  consent, and feature IDs 0060 through 0065 remain unique.
- Remaining limitation: this is a repository and prompt contract, not an OS
  filesystem interceptor. Enforcement is deterministic in generated guidance,
  validation, and delivery prompts, but a non-compliant external agent could
  still ignore instructions.

## REPOSITORY MEMORY

- Created this living spec because the lane boundary, root-checkout invariant,
  scoped-consent semantics, and immutable instruction-version decision are
  material rationale that code and tests alone would not preserve.
- Updated `docs/CONSTITUTION.md` with the demonstrated project-wide invariant
  that every coding-agent mutation defaults to a new PR lane without prompting,
  explicitly directed continuation remains bounded, and all implementation
  occurs outside the primary checkout.
- Updated `docs/references/rules/work-lane-gating.md`,
  `safety-guardrails.md`, and `github-pr-delivery.md` as the reusable normative
  contract, and kept `docs/references/README.md` routing and its project-owned
  ruleset index current.
- Refreshed generated active instruction artifacts from their canonical
  templates. Preserved published v1-v3 instruction files unchanged and added
  v4 rather than rewriting history.
- Issue #155 updated this spec, the work-lane ruleset, generated guidance, and
  historical v5 regression coverage because concise response interpretation is
  durable workflow behavior. The Constitution remains unchanged because the
  shorthand specializes the existing explicit-choice invariant rather than
  introducing a new project-wide invariant.
- The work-lane ruleset now declares downstream registry scope explicitly so
  scheduled health and reconcile refreshes install it in every managed project.
- GH-160 updated this spec, the downstream ruleset, generated and checked-in
  guidance, reconcile expectations, and additive v6 because ordinary full-form
  lane vocabulary is durable workflow behavior. Constitution curation remains
  not required: the aliases specialize the existing explicit-choice invariant.
- GH-186 updated this spec, Constitution, downstream work-lane, safety, and
  delivery rules, generated and checked-in guidance, managed-file prompts,
  reconcile expectations, and additive v11. This reversal is material durable
  rationale: the default now creates a new worklane without asking, while
  explicit continuation remains bounded and merge authorization stays separate.
- PR #187 was repaired in place to add existing-PR lifecycle precedence across
  the same canonical artifacts and v11. This preserves the established
  no-corrective-PR invariant while keeping missing repair authority fail-closed.
