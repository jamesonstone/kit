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
    used_for: explicit lane choice and mutation tripwire
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

Make the coding-agent work lane an explicit user decision before any repository
file mutation. Every change must happen in a proven writable worktree with a
concrete pull-request landing plan, while the clone's primary checkout remains
read-only and preserves user-owned state.

## CONTEXT

- The current work-lane rule gates only source code and production-affecting
  configuration. It explicitly excludes standalone documentation, specs,
  planning artifacts, and other repository writes.
- The current rule also treats a clean current default branch as automatic
  consent to create a lane without asking the user. That behavior conflicts
  with the requested unconditional choice.
- Current generated agent instructions repeat the automatic clean-default-branch
  path and permit work in an existing checkout when it owns a lane. They do not
  make the primary checkout categorically read-only for coding agents.
- Feature 0050 established canonical linked worktrees at
  `~/worktrees/<owner>/<repository>/<lane>`, exact `GH-<issue-number>` durable
  lanes, and preservation of dirty worktree state.
- Published `kit instructions` versions v1 through v3 are immutable historical
  contracts. A new current version is required rather than rewriting those
  payloads.
- The gate's fixed question currently requires a free-form explicit answer and
  does not codify concise `c`, `n`, or `y` responses or mixed responses that
  append lane-specific instructions.
- GitHub issue #143, branch `GH-143`, and canonical worktree
  `~/worktrees/jamesonstone/kit/GH-143` own this change. The planned end state
  is one ready pull request targeting `main`.

## REQUIREMENTS

- Before a coding agent creates, edits, deletes, moves, or generates any
  repository file, it must ask whether to create a new issue, exact issue
  branch, canonical worktree, and pull request or continue in the existing
  branch/worktree and land through that lane's pull request.
- Read-only repository discovery, safety recon, context resolution, planning,
  and questions may precede the choice. Commands that write specs, generated
  files, configuration, source, tests, documentation, or Git state may not.
- Remove automatic clean-default-branch consent and implicit consent from a
  generic request to produce a pull request. Only the user's recorded lane
  choice satisfies the gate.
- After trimming surrounding whitespace, accept a case-insensitive leading
  standalone `c` as continue existing and `n` or `y` as new lane. When the
  response contains more text, the shorthand is the primary choice and the
  remaining text is retained as supplemental instructions within that lane.
- Preserve explicit full-form answers. Contradictory or unresolved responses
  remain ambiguous and must fail closed before mutation.
- One recorded choice covers the accepted unit of work and its required tests,
  documentation, validation fixes, and delivery. Materially new or tangential
  scope requires a new choice; routine subtasks do not repeatedly re-prompt.
- A new-lane choice must establish or reuse the human-assigned issue, exact
  `GH-<issue-number>` branch, canonical linked worktree, protected base, and
  ready pull-request plan before the first repository file mutation.
- A continue-existing choice is valid only when recon proves a non-protected
  writable branch, its owning worktree, its issue scope, and a plan to create
  or update that branch's pull request. Ambiguous or detached lanes fail closed.
- The clone's primary/root checkout is read-only for coding-agent work,
  regardless of branch cleanliness or file type. It may be inspected and may
  supply shared environment symlinks, but it may not receive implementation,
  documentation, generated, staging, commit, or branch-switch mutations.
- If an agent detects an ungated edit or a root-checkout edit, it must stop,
  preserve all state, report the violation, and avoid staging, committing,
  pushing, resetting, cleaning, stashing, or silently transferring the change.
  Recovery begins only after the lane choice and exact ownership are proven.
- Align the normative rules, active generated instruction templates,
  checked-in generated guidance, Constitution, current versioned agent
  instructions, and regression tests without changing immutable v1-v3 payloads.
- Keep every changed handwritten Go source and test file at or below 300
  physical lines and pass complete repository validation.
- Observable acceptance: policy tests reject automatic consent, require the
  exact binary lane choice and pull-request landing plan, protect the primary
  checkout for every file mutation, preserve tripwire state, and prove the
  new current instructions version while v1-v3 hashes remain unchanged.
- Non-goals: implementing a runtime filesystem interceptor, automatically
  moving pre-existing dirty changes, creating a second pull request for an
  existing approved lane, changing native Git worktree semantics, or merging
  the delivery pull request.

## ACCEPTED PLAN

1. Rewrite `work-lane-gating` around one unconditional pre-mutation user choice,
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

## DECISIONS

- Accepted: apply the gate to every coding-agent repository file mutation,
  including documentation and spec creation. A file-type exception would
  still permit root-checkout changes with no delivery lane.
- Accepted: make the choice scoped to one accepted unit of work. Asking again
  for its tests or required documentation would add ceremony without changing
  lane ownership; a material scope pivot must re-open the gate.
- Accepted: publish v4 as the new current versioned instruction contract.
  Rewriting v1-v3 would invalidate their documented immutability and hashes.
- Rejected: retain the clean-default-branch automatic allocation path. The
  user explicitly requires a choice, and automatic allocation cannot express
  whether existing work should be continued.
- Rejected: allow edits in a clean primary checkout with a later transfer plan.
  That makes user-owned root state depend on recovery and permits an ungated
  diff to exist before its pull-request lane exists.
- Accepted for the follow-up: use the first standalone `c`, `n`, or `y` token
  as the primary lane choice while preserving any trailing instructions. This
  keeps the common response concise without discarding user constraints.
- Retained for the follow-up: fail closed when trailing text contradicts the
  shorthand or no binary choice can be resolved. Shorthand reduces input
  ceremony but does not weaken the mutation gate.
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
- `kit spec` currently mutates repository memory immediately. Under the new
  contract, agents must select and establish the writable lane before invoking
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

## OUTCOME

- The active work-lane rule now requires the explicit binary user choice and a
  complete Pull-Request Landing Plan before every coding-agent repository or
  delivery mutation, including docs, specs, configuration, and generators.
- Clean-default-branch and generic-PR-request consent paths were removed. One
  decision covers required completion work, while material scope pivots re-open
  the gate.
- Safety and delivery rules now make the exact primary checkout read-only and
  accept continued work only in a proven non-primary owned worktree with one
  create-or-update pull-request route.
- Managed-file prompts preserve ungated or primary-checkout snapshots as
  evidence and no longer instruct agents to transfer and restore root changes
  automatically.
- Generated current AGENTS, Claude, Copilot, workflow, Guardrails, RLM, and
  Tooling guidance carries the gate. Current v5 instructions add deletion
  safety and shorthand lane responses while immutable v1-v4 remain unchanged.
- PR-review repair makes the gate unconditional in newly created and resumed
  v4 sessions, requires repository context resolution before write-capable
  workflows, and keeps generated compact/full guidance ordered across rule
  loading, recon, explicit choice, and plan verification for file and delivery
  mutations.
- Follow-up repair makes thread title and pin operations ordered, verified, and
  fail-visible; starts repository work at the routing entrypoint; discovers
  unknown context-resolve behavior through capabilities; and states explicitly
  that a landing lane never authorizes a merge outside the exact bounded
  `MERGE_READY` contract.
- Final review repair completes v4's execution preconditions by requiring native
  planning, the semantic living-spec decision and accepted plan when material,
  testing-environment rule loads, and the source-file-size gate before coding.
- Regression tests enforce the ruleset wording, `c`/`n`/`y` shorthand semantics,
  root protection, generated and checked-in alignment, managed-command
  tripwire, current-version behavior, and immutable historical versions.
- Managed-health and `kit reconcile --include-files` regression coverage proves
  missing root/provider guidance and Guardrails are restored with shorthand
  semantics. Reconcile also reports existing-section shorthand drift for
  reviewed semantic curation instead of silently treating structural freshness
  as current guidance.
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
  that every coding-agent mutation has an explicit human-selected PR lane and
  occurs outside the primary checkout.
- Updated `docs/references/rules/work-lane-gating.md`,
  `safety-guardrails.md`, and `github-pr-delivery.md` as the reusable normative
  contract, and kept `docs/references/README.md` routing and its project-owned
  ruleset index current.
- Refreshed generated active instruction artifacts from their canonical
  templates. Preserved published v1-v3 instruction files unchanged and added
  v4 rather than rewriting history.
- Issue #155 updated this spec, the work-lane ruleset, active generated guidance,
  and current v5 regression coverage because concise response interpretation is
  durable workflow behavior. The Constitution remains unchanged because the
  shorthand specializes the existing explicit-choice invariant rather than
  introducing a new project-wide invariant.
- The work-lane ruleset now declares downstream registry scope explicitly so
  scheduled health and reconcile refreshes install it in every managed project.
