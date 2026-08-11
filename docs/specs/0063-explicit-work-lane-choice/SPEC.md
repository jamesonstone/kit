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

## DISCOVERIES

- The repository-shared feature allocator initially reserved ID 0062. Before
  delivery, refreshed `origin/main` introduced merged feature
  `0062-git-wt-removal`, so this lane advanced its shared reservation and spec
  identity to 0063 before integrating upstream.
- `kit spec` currently mutates repository memory immediately. Under the new
  contract, agents must select and establish the writable lane before invoking
  it; capability inspection may still run read-only beforehand.
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
  `607762ed53f64dd2c795efa51915cbc2a7a8187cde1d4639980e9c4f477277f2`,
  matching immutable v4.
- `gitleaks dir --redact --no-banner .` scanned 5.25 MB with no leaks, and
  `gitleaks git --redact --no-banner --log-opts='origin/main..HEAD' .` scanned
  the GH-143 commit range with no leaks.
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
  Tooling guidance carries the new gate. `kit instructions` now defaults to
  immutable v4; the byte hashes of published v1-v3 remain unchanged.
- Regression tests enforce the ruleset wording, root protection, generated and
  checked-in alignment, managed-command tripwire, new current version, and
  immutable historical versions.
- The integrated result preserves current merge-autonomy, release-orchestration,
  and git-wt-removal policy. Pull-request delivery still never implies merge
  consent, and feature IDs 0060 through 0063 remain unique.
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
