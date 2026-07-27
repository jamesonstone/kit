---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: 0052
  slug: worktree-sync
  dir: 0052-worktree-sync
relationships:
  - type: builds_on
    target: 0050-safe-worktree-workflow
  - type: builds_on
    target: 0051-resolve-pr-repair-worktrees
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: safety-guardrails
    name: Safety guardrails
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: fail-closed removal, dirty-state preservation, and delivery safety
    status: active
  - id: github-pr-delivery
    name: GitHub PR delivery
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: issue GH-93, branch GH-93, and ready pull-request delivery
    status: active
  - id: command-capabilities
    name: Command capabilities
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: network, filesystem, Git, and GitHub behavior metadata
    status: active
  - id: worktree-guide
    name: Git worktrees
    type: documentation
    target: docs/references/worktrees.md
    relation: guides
    read_policy: must
    used_for: canonical lane, environment-link, pruning, and removal contracts
    status: active
  - id: safe-worktree-spec
    name: Safe worktree workflow
    type: spec
    target: docs/specs/0050-safe-worktree-workflow/SPEC.md
    relation: informs
    read_policy: must
    used_for: existing manual removal and native worktree invariants
    status: active
  - id: repair-worktree-spec
    name: Resolve PR repair worktrees
    type: spec
    target: docs/specs/0051-resolve-pr-repair-worktrees/SPEC.md
    relation: informs
    read_policy: conditional
    used_for: richer same-repository PR metadata and exact-head evidence
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
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Add one explicit `git wt sync` maintenance command that safely reconciles the
origin default branch, removes only worktrees proven to belong to merged
same-repository pull requests, deletes their now-unused local branches, prunes
stale worktree metadata, and reports every decision deterministically.

`git wt list` remains read-only and offline-capable: its bounded, fail-soft
open-pull-request annotation may consult `gh`, but failures never prevent
listing. Synchronization is never implicit in listing or navigation.

## CONTEXT

- Feature 0050 established the canonical external worktree hierarchy, exact
  managed `.env` link, conservative manual removal, explicit pruning, and
  native Git authority.
- Feature 0051 added same-repository PR-head resolution, but no command
  reconciles merged lanes or the local default branch.
- `internal/worktree/remove.go` already proves exact registration, refuses the
  current lane, inspects tracked, untracked, and ignored state, recognizes only
  the verified managed `.env` symlink, and restores that link if native removal
  fails.
- Manual `git wt remove` requires upstream and ahead-state proof and preserves
  the branch. Sync needs a separate proof because `fetch --prune` may remove the
  merged branch's upstream before cleanup.
- Git ancestry or a missing remote branch does not prove that a pull request
  merged, targeted the origin default branch, or came from this repository.
- GitHub issue #93 and branch `GH-93` are the delivery lane.

## REQUIREMENTS

- REQ-001: Add exactly one visible `git wt sync` command. Do not add implicit
  synchronization to `git wt list`, selectors, navigation, or other commands.
- REQ-002: Ordinary sync discovers the current repository and origin default
  branch, then runs one fetch that prunes only `origin`.
- REQ-003: `--dry-run` is strictly mutation-free: it performs no fetch, ref
  update, worktree removal, branch deletion, metadata pruning, symlink removal,
  or filesystem write. It may use live remote and GitHub reads.
- REQ-004: Support deterministic human output and `--json`, both rendered from
  one typed report containing options, repository/default state, lane
  decisions, actions, failures, and the refreshed worktree summary.
- REQ-005: Reconcile the local default branch as follows:
  - equal to `origin/<default>`: report current;
  - clean and strictly behind: use `git merge --ff-only origin/<default>` when
    the default branch is checked out, otherwise atomically update the local
    branch ref from its observed old OID to the remote OID;
  - dirty, ahead, or diverged: preserve and report without merge, reset,
    checkout, rebase, or rewrite.
- REQ-006: Never remove the primary worktree or the worktree from which sync is
  running.
- REQ-007: Consider only exact registered worktree paths beneath the canonical
  `~/worktrees/<owner>/<repository>` root. Preserve legacy, detached, foreign,
  or otherwise non-canonical registrations with an explicit reason.
- REQ-008: Removing a lane requires all of:
  - exact canonical registration;
  - a pull request whose state is merged into the discovered origin default
    branch;
  - a same-repository head branch;
  - unambiguous PR evidence for that branch;
  - local worktree `HEAD` exactly equal to the merged PR head OID;
  - no tracked, staged, untracked, or ignored material except the verified
    managed `.env` symlink.
- REQ-009: Open, closed-unmerged, wrong-base, fork, missing, ambiguous, or
  head-OID-mismatched PR evidence preserves the lane and reports why.
- REQ-010: Reuse/refactor the existing manual-removal preflight and execution.
  Manual `git wt remove` retains upstream/ahead proof and branch preservation.
  Sync substitutes merged-PR plus exact-head proof.
- REQ-011: Ordinary sync removes every proven-safe worktree immediately using
  ordinary non-force `git worktree remove`, restores the managed `.env` link if
  that removal fails, and then deletes the exact local branch with ordinary
  `git branch -d`.
- REQ-012: A local branch is never deleted before its worktree removal
  succeeds. Branch-deletion failure is reported, makes the overall command
  nonzero, and does not trigger force deletion or destructive recovery.
- REQ-013: After lane processing, ordinary sync prunes stale worktree metadata
  with ordinary `git worktree prune --verbose`. Dry-run reports what would be
  pruned without invoking a mutating prune.
- REQ-014: Use an injectable rich PR resolver returning number, state,
  `mergedAt`, base/head names, head OID, cross-repository status, and URL.
  Resolve branches in one fast batch query with exact targeted fallback for
  unmatched registered branches.
- REQ-015: Fetch, GitHub, inspection, removal, branch-deletion, pruning, and
  output failures fail closed. Independent safe candidates continue when
  possible, while any operation failure makes the overall command return
  nonzero after rendering the report.
- REQ-016: Repeated sync is idempotent: removed lanes stay absent, preserved
  lanes retain their state, the default branch remains reconciled, and the
  second report contains no duplicate mutation.
- REQ-017: Update help, README summary if useful, command documentation,
  canonical and embedded worktree guidance, command capabilities, tests, and
  repository memory with the exact behavior.

### Non-goals

- Synchronizing remotes other than `origin`.
- Removing a worktree using branch ancestry, missing upstream, merged commit
  reachability, branch naming, age, or remote-branch deletion alone.
- Cleaning, stashing, resetting, rebasing, merging, force-removing, force
  deleting, force-pushing, or discarding any lane state.
- Removing detached PR inspection lanes automatically.
- Deleting remote branches.
- Changing application processes, databases, ports, Temporal state, runtime
  configuration, or sibling repositories.
- Making sync implicit in `git wt list`.

## ACCEPTANCE

- AC-001: Command parsing accepts `sync`, `sync --dry-run`, `sync --json`, and
  both flags together, and rejects unknown or duplicate-invalid arguments.
- AC-002: Tests prove equal, clean-behind checked-out, clean-behind
  not-checked-out, dirty, ahead, and diverged default-branch decisions.
- AC-003: Tests prove exact merged-PR lane removal plus local branch deletion.
- AC-004: Tests preserve open, closed-unmerged, wrong-base, fork, missing,
  ambiguous, and OID-mismatched PR lanes with deterministic reasons.
- AC-005: Tests preserve dirty, staged, untracked, ignored, current, primary,
  non-canonical, and detached worktrees.
- AC-006: Tests prove managed `.env` removal and restoration when worktree
  removal fails.
- AC-007: Tests prove merged cleanup still works after origin pruning deletes
  the branch's upstream.
- AC-008: Tests prove fetch, GitHub, worktree removal, branch deletion, pruning,
  and output failures fail closed; independent candidates continue and the
  aggregate result is nonzero.
- AC-009: Tests prove dry-run causes no repository, ref, worktree, branch,
  metadata, symlink, or filesystem mutation.
- AC-010: Tests prove human/JSON parity, deterministic ordering, repeat-run
  idempotence, and output-error propagation.
- AC-011: A real built-binary `git wt sync --dry-run` smoke test leaves pre/post
  worktree, branch, ref, status, and filesystem state identical.
- AC-012: Full formatting, focused/full tests, vet, `build-git-wt`, Kit
  feature/project checks, diff/secret review, and ready-PR delivery pass.

## ACCEPTED PLAN

1. Define typed sync options, report, default-branch state, lane decisions,
   actions, and aggregate error behavior in `internal/worktree/sync.go`.
2. Add a rich injectable GitHub resolver with one batch PR query and exact
   branch fallback, keeping same-repository and merged-base proof explicit.
3. Extract reusable removal inspection and execution from `remove.go`, keeping
   manual upstream/ahead proof and managed-link restoration unchanged.
4. Implement ordinary and strictly non-mutating dry-run default reconciliation,
   lane inspection/removal, exact branch deletion, pruning, and refreshed
   listing.
5. Render human and JSON output from the same report, with stable lane ordering,
   explicit preservation reasons, and aggregate failures.
6. Add focused real-Git and injected-failure tests for every acceptance group.
7. Update command/help/capability/docs/template surfaces, validate end to end,
   curate repository memory, and deliver through issue #93 and `GH-93`.

## DECISIONS

- Ordinary `git wt sync` immediately applies every mutation that passes the
  complete safety proof; preview-only behavior is available through
  `--dry-run`.
- V1 deletes the exact local branch after successful worktree removal using
  ordinary `git branch -d`; it never force-deletes and never deletes remotely.
- Dry-run does not fetch or update any local ref. It reports from current local
  state plus live remote/GitHub reads.
- Batch GitHub lookup is an optimization only. Exact targeted fallback and
  explicit ambiguity preserve correctness when the batch result is incomplete.
- A failed independent candidate does not prevent inspection of later
  candidates, but any operational failure makes the final command nonzero.
- Human and JSON output are views of one report, not separately accumulated
  command logs.

## OPEN QUESTIONS

None. The user explicitly selected immediate application, local branch
deletion, and strictly mutation-free dry-run behavior.

## DISCOVERIES

- The initial `kit spec worktree-sync --output-only` invocation allocated
  feature 0052 but unexpectedly entered the deprecated V2 editor path. No
  thesis was captured and no implementation began, so the empty placeholder
  was semantically replaced with this accepted V3 spec.
- On macOS, Git reports temporary worktree paths through the physical
  `/private/var` path while the configured test root may use `/var`. Canonical
  lane proof therefore resolves both the registered path and project root
  before comparing them; it still requires the exact physical destination.
- Ordinary `git branch -d` normally chooses its merge check from the invoking
  worktree's current branch when an upstream disappeared. Sync supplies
  command-local upstream configuration pointing at `origin/<default>` so
  deletion remains non-force and is independent of which worktree invoked
  sync.
- The real-repository dry-run found one stale weekly-health registration and
  previewed its native prune message while leaving registrations, refs,
  statuses, directories, and environment links byte-for-byte unchanged.
- The installed GitHub CLI supports all required rich fields in one `pr list`
  query: number, state, merge time, base/head names, head OID,
  cross-repository status, and URL.

## VALIDATION

- `gofmt` completed for every changed Go file.
- Focused real-Git sync tests passed for default-branch states, exact merged
  removal, local branch deletion, PR refusal cases, dirty/ignored/current/
  primary/detached/non-canonical preservation, managed environment links,
  deleted upstreams, dry-run immutability, independent failure continuation,
  output parity/errors, and idempotence.
- `go test ./...` passed.
- `go vet ./...` passed.
- `make build-git-wt` passed.
- `go run ./cmd/kit capabilities git wt sync --json` returned the complete
  network, write, Git mutation, flag, example, and caveat contract.
- `go run ./cmd/kit capabilities --search 'merged worktree' --json` discovered
  `git wt sync`.
- A real built-binary `git wt sync --dry-run --json` smoke test compared
  pre/post registered worktrees, local and origin refs, every worktree status,
  directory identity, and `.env` state with `cmp`; they were identical.
- `git diff --check` passed before repository-memory curation.

## OUTCOME

Implementation and validation are complete on issue #93 and branch `GH-93`.
`git wt sync` is explicit, default-applying only for fully proven safe merged
lanes, branch-deleting only through ordinary `git branch -d`, deterministic in
human and JSON modes, and strictly local-mutation-free under `--dry-run`.
The ready pull request is
<https://github.com/jamesonstone/kit/pull/94>; review and merge remain.

## REPOSITORY MEMORY

Decision: created

Rationale: Automatic merged-lane removal and local branch deletion introduce a
consequential destructive boundary. The exact GitHub, OID, canonical-path,
dirty-state, dry-run, failure-aggregation, and restoration rationale must
survive beyond code and tests.

Constitution curation result: no project-wide constitutional rule changed.
The new destructive boundary is feature-specific and is fully captured in this
spec plus the canonical worktree reference, so `docs/CONSTITUTION.md` remains
unchanged.

Artifacts:

- `docs/specs/0052-worktree-sync/SPEC.md`
- `docs/references/worktrees.md`
- `internal/templates/worktrees_reference.md`
