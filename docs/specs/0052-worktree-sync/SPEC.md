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
    used_for: issues GH-93, GH-125, and GH-127, their exact branches, and ready pull-request delivery
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
  managed environment links, conservative manual removal, explicit pruning, and
  native Git authority.
- Feature 0051 added same-repository PR-head resolution, but no command
  reconciles merged lanes or the local default branch.
- `internal/worktree/remove.go` already proves exact registration, refuses the
  current lane, inspects tracked, untracked, and ignored state, recognizes only
  verified managed `.env` and `.envrc` symlinks, and restores removed links if
  native removal fails.
- Manual `git wt remove` requires upstream and ahead-state proof and preserves
  the branch. Sync needs a separate proof because `fetch --prune` may remove the
  merged branch's upstream before cleanup.
- Git ancestry or a missing remote branch does not prove that a pull request
  merged, targeted the origin default branch, or came from this repository.
- GitHub issue #93 and branch `GH-93` are the delivery lane.
- In `jamesonstone/merge-controller`, merged PR #2 records local `GH-1` OID
  `7dc7b8d` as its exact head but squash commit `3748110` does not contain that
  OID as an ancestor. Sync therefore removed the proven-safe worktree and then
  failed when ordinary `git branch -d` applied a different ancestry proof.
- GitHub issue #125 and branch `GH-125` own the squash-merge repair.
- GitHub issue #127 and branch `GH-127` own the ignored root `bin/` cleanup
  exception requested for repositories that use the conventional Go build
  output directory.

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
  - no tracked, staged, untracked, or ignored material except verified managed
    `.env` and `.envrc` symlinks.
- REQ-009: Open, closed-unmerged, wrong-base, fork, missing, ambiguous, or
  head-OID-mismatched PR evidence preserves the lane and reports why.
- REQ-010: Reuse/refactor the existing manual-removal preflight and execution.
  Manual `git wt remove` retains upstream/ahead proof and branch preservation.
  Sync substitutes merged-PR plus exact-head proof.
- REQ-011: Ordinary sync removes every proven-safe worktree immediately using
  ordinary non-force `git worktree remove`, restores all removed managed
  environment links if that removal fails, and then deletes the exact local
  branch with ordinary non-force `git branch -d`. Before worktree removal, sync
  must create an exact PR-head proof ref through a create-only compare-and-swap
  and verify the branch's configured merge ref. Branch deletion must resolve
  that proof through a command-local temporary remote, so squash merges do not
  depend on default-branch ancestry while Git still refuses moved or reattached
  branches. Sync must remove its proof ref with the expected OID afterward.
- REQ-012: A local branch is never deleted before its worktree removal
  succeeds. Proof preparation failure preserves the entire lane. Missing or
  changed exact-ref evidence, checked-out branch state, branch-deletion failure,
  and proof cleanup failure are reported, make the overall command nonzero,
  and do not trigger force deletion or destructive recovery.
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
- REQ-018: Sync may discard one ignored repository-root `bin/` directory from
  an otherwise proven-safe merged lane immediately before native worktree
  removal. The exception applies only when Git reports the exact porcelain
  record `!! bin/` and `bin/` is an actual directory. A nested record such as
  `!! bin/generated/`, including one caused by a clean tracked sibling beneath
  `bin/`, preserves the lane. Tracked changes, ordinary untracked material,
  symlinks, nested `*/bin/` paths, and every other ignored path also preserve
  the lane. Manual `git wt remove` retains its existing all-ignored-material
  refusal.

### Non-goals

- Synchronizing remotes other than `origin`.
- Removing a worktree using branch ancestry, missing upstream, merged commit
  reachability, branch naming, age, or remote-branch deletion alone.
- Cleaning, stashing, resetting, rebasing, merging, force-removing, force
  deleting, deleting a branch ref that moved from the proven PR head OID,
  force-pushing, or discarding any lane state.
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
- AC-013: A real-Git squash-merged fixture removes both the clean canonical
  worktree and its exact local branch without requiring ancestry to the
  synthesized default-branch commit.
- AC-014: A branch ref moved or reattached after worktree removal is preserved,
  the temporary proof ref is cleaned, the failure is reported, and sync returns
  nonzero. Proof preparation failure preserves the worktree as well.
- AC-015: Real-Git tests prove sync removes an eligible merged lane containing
  ignored root `bin/` output, dry-run leaves that output untouched, and manual
  removal plus tracked, untracked, symlinked, nested, or other ignored material
  continue to fail closed. A clean tracked file beneath `bin/` beside nested
  ignored output also preserves the lane and its tracked content.

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

### GH-125 squash-merge repair

1. Anchor the merged PR head in a create-only temporary proof ref before
   worktree removal, recheck the exact local branch, and run ordinary
   `git branch -d` against that proof through a command-local remote mapping.
   Remove only the task-owned proof ref with its expected OID afterward.
2. Add real-Git coverage for successful squash-merged cleanup and fail-closed
   preservation when the branch ref changes before deletion.
3. Update canonical and embedded worktree guidance, run the complete affected
   validation matrix, curate this spec, and deliver through issue #125 and
   `GH-125`.

### GH-127 ignored root bin cleanup

1. Extend sync-only removal inspection to recognize literal ignored porcelain
   entries beneath repository-root `bin/` after proving the root is a real
   directory; preserve manual removal and every other dirty-state refusal.
2. Reinspect the lane immediately before deletion, remove only the exact
   worktree-root `bin/` directory, and then retain ordinary non-force worktree
   and local-branch removal.
3. Add real-Git success, dry-run, refusal, and cleanup-failure coverage; align
   help, capabilities, canonical and embedded worktree guidance; validate and
   deliver through issue #127 and `GH-127`.

## DECISIONS

- Ordinary `git wt sync` immediately applies every mutation that passes the
  complete safety proof; preview-only behavior is available through
  `--dry-run`.
- GH-125 refines V1's ordinary `git branch -d` choice. Sync first creates a
  task-owned remote-tracking proof ref at the exact merged PR head using a
  create-only expected-zero OID, then asks ordinary `git branch -d` to evaluate
  the local branch against that proof through command-local remote mapping.
  This supports squash merges while retaining Git's checked-out and moved-head
  refusals. The proof is removed only with its expected OID; sync never
  force-deletes and never deletes the real remote branch.
- Dry-run does not fetch or update any local ref. It reports from current local
  state plus live remote/GitHub reads.
- Batch GitHub lookup is an optimization only. Exact targeted fallback and
  explicit ambiguity preserve correctness when the batch result is incomplete.
- A failed independent candidate does not prevent inspection of later
  candidates, but any operational failure makes the final command nonzero.
- Human and JSON output are views of one report, not separately accumulated
  command logs.
- Ignored root `bin/` output is a sync-only disposable-build exception. It is
  recognized only from Git's exact ignored porcelain record `!! bin/` plus an
  exact non-symlink directory check, removed by exact absolute path, and never
  restored if a later native worktree or branch operation fails. Prefix or
  nested status records are insufficient because they can coexist with clean
  tracked material beneath `bin/`. Manual removal remains conservative because
  it lacks sync's complete merged-PR and exact-head proof.

## OPEN QUESTIONS

None. The user explicitly selected immediate application, local branch
deletion, strictly mutation-free dry-run behavior, and disposal of ignored
root `bin/` build output during proven merged-lane cleanup.

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
- `git branch -d` answers a commit-ancestry question that is incompatible with
  GitHub squash merges. Merge-controller PR #2 proves the stronger identity
  needed by sync: merged state, same-repository/default-base ownership, and an
  exact PR-head/local-head OID match. A task-owned proof ref lets ordinary
  `git branch -d` consume that identity while preserving its worktree-ownership
  guard; create-only and expected-old-OID proof updates fail closed on
  collisions or cleanup races.

## VALIDATION

### GH-127 ignored root bin cleanup

- Focused real-Git tests passed for automatic ignored root `bin/` removal,
  strictly non-mutating dry-run, manual-removal refusal, ordinary untracked
  root `bin/`, ignored nested `*/bin/`, other ignored material, root symlinks,
  tracked changes beneath `bin/`, a clean tracked sibling beside nested ignored
  output, build-output deletion failure, managed-link restoration, and native
  worktree-removal failure after build-output disposal.
- `make fmt`, `go test ./... -count=1`, `make vet`, and the affected
  `go test -race ./internal/worktree -run
  '^(TestSync|TestManualRemoveStillRefusesIgnoredRootBuildOutput)' -count=1`
  suite passed.
- `make build` passed, rebuilt both binaries, and installed the validated
  `git-wt` artifact and manpage.
- `kit capabilities git wt sync --json` and the `ignored root bin` capability
  search exposed the new file-deletion and refusal boundaries.
- `kit check worktree-sync`, `kit check safe-worktree-workflow`, and
  `kit check --project` passed.
- Canonical and embedded worktree guidance matched byte-for-byte; manpage lint,
  `git diff --check`, and `gitleaks git --redact --no-banner` passed.
- Every affected handwritten Go source and test file remained at most 300
  physical lines; `internal/worktree/remove.go` was the largest at 299 lines.
- PR #128 review repair added the clean tracked-sibling real-Git regression and
  reran focused and full tests, affected race tests, vet, build, change-scoped
  `golangci-lint`, feature/project checks, documentation parity and lint,
  diff checks, and secret scanning successfully.
- A built-binary live `merge-controller` dry run reported zero failures, all
  24 linked branch lanes as `would-remove / proven-safe-merged-lane`, and only
  the primary checkout as preserved. A SHA-256 digest over local refs,
  registered worktrees, and every lane's tracked, untracked, and ignored status
  was identical before and after the preview.

### GH-125 squash-merge repair

- Focused real-Git sync tests passed for squash-merged cleanup, exact-ref
  movement and reattachment refusal, proof preparation and collision handling,
  proof cleanup, ordinary merged cleanup, off-default invocation,
  branch-deletion failure reporting, and dry-run immutability.
- `go test ./internal/worktree -count=1` passed.
- `go test -race ./internal/worktree -run '^TestSync' -count=1` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `make build-git-wt` passed.
- `go test -race ./internal/worktree` reached the unchanged terminal-selector
  cancellation baseline failure in
  `TestSelectWorktreeTerminalCancellationRestoresPTY`; the complete affected
  sync race scope passed separately.
- `go run ./cmd/kit capabilities git wt sync --json` and the `merged worktree`
  capability search reported compare-and-swap exact-ref deletion and squash
  merge support.
- `go run ./cmd/kit check worktree-sync` and
  `go run ./cmd/kit check --project` passed.
- Every changed handwritten Go source and test file was at most 300 physical
  lines; the largest was 237 lines.
- A real built-binary `git wt sync --dry-run --json` smoke test reported no
  failures, and complete local/origin refs, registered worktrees, plus tracked,
  untracked, and ignored status snapshots were identical before and after.
- Canonical and embedded worktree guidance remained byte-identical, and
  `git diff --check` passed.

### GH-93 original delivery

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

### GH-127 ignored root bin cleanup

Implementation and local validation are complete on issue #127 and branch
`GH-127`. Sync now recognizes only the exact ignored porcelain record `!! bin/`
for an actual repository-root `bin/` directory, reinspects the lane immediately
before deletion, removes that exact disposable build-output directory, and
continues through ordinary non-force worktree removal and exact local-branch
deletion. Manual removal, nested or symlinked bin paths, clean or changed
tracked material beneath `bin/`, ordinary untracked files, and every other
ignored path remain fail-closed. If a later native operation fails, the
disposable build output stays deleted while the remaining worktree and branch
are preserved and the failure is reported.

### GH-125 squash-merge repair

Implementation and local validation are complete on issue #125 and branch
`GH-125`. Sync now anchors the already-proven merged PR head OID in a temporary,
create-only proof ref before removing the worktree, then uses ordinary
`git branch -d` through a command-local remote mapping and removes the proof by
expected OID. GitHub squash merges no longer fail the unrelated default-branch
ancestry check, while missing, moved, or reattached branches remain preserved
with an explicit operation failure. Dry-run behavior and all pre-existing lane
refusal proofs remain unchanged.

### GH-93 original delivery

Implementation and validation are complete on issue #93 and branch `GH-93`.
`git wt sync` is explicit, default-applying only for fully proven safe merged
lanes, branch-deleting only through ordinary `git branch -d`, deterministic in
human and JSON modes, and strictly local-mutation-free under `--dry-run`.
The ready pull request is
<https://github.com/jamesonstone/kit/pull/94>; review and merge remain.

## REPOSITORY MEMORY

Decision: updated

Rationale: GH-93 created the durable contract for automatic merged-lane removal
and local branch deletion. GH-125 updated that contract for squash merges, and
GH-127 adds one deliberate destructive-boundary exception for ignored root
`bin/` build output. The distinction between sync-only disposal and conservative
manual removal, exact full-record ignored-path recognition, immediate
reinspection, non-restoration after later failures, and all retained dirty-state
protections must survive beyond code and tests. Exact record matching is a
safety boundary: a nested ignored status can coexist with clean tracked content
under `bin/` and must not authorize deletion.

Constitution curation result: updated the project-wide worktree invariant with
the single validated sync-only ignored root `bin/` exception while retaining
the prohibition on stash, reset, clean, force removal, and disposal of every
other dirty or ignored path. Feature-specific implementation and failure
rationale remains in this spec and the canonical worktree reference.

Artifacts:

- `docs/specs/0052-worktree-sync/SPEC.md`
- `docs/references/worktrees.md`
- `internal/templates/worktrees_reference.md`
- `internal/worktree/remove_build_output.go`
- `internal/worktree/sync_build_output_test.go`
- `pkg/cli/capabilities_catalog_utilities.go`
- `docs/CONSTITUTION.md`
