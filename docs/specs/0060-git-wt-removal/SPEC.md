---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0060"
  slug: "git-wt-removal"
  dir: "0060-git-wt-removal"
relationships:
  - type: builds_on
    target: 0051-resolve-pr-repair-worktrees
  - type: related_to
    target: 0052-worktree-sync
references:
  - id: "feature-notes"
    name: "Feature notes"
    type: "notes"
    target: "docs/notes/0060-git-wt-removal"
    relation: "informs"
    read_policy: "conditional"
    used_for: "optional pre-brainstorm research input"
    status: "optional"
  - id: "git-wt-extraction-program"
    name: "Git wt extraction program"
    type: "program"
    target: "docs/programs/git-wt-extraction/PROGRAM.md"
    relation: "guides"
    read_policy: "must"
    used_for: "cross-repository dependency and delivery state"
    status: "active"
---
# SPEC

## PURPOSE

Remove the generic `git wt` product from Kit after its implementation moved to
Kura. Kit should again ship only the repository-memory CLI while retaining the
narrow internal worktree preparation required by Kit's own pull-request repair
flows.

## CONTEXT

- Kit currently builds, installs, releases, documents, and advertises a second
  `git-wt` executable through `cmd/git-wt`, `internal/worktree`, the Makefile,
  GoReleaser, the command capability catalog, a manpage, and user docs.
- Kura PR `jamesonstone/kura#2` contains the extracted implementation at commit
  `195184cf25143a264cd549f7ab0880ca1cb0999c`; its hosted checks passed and the
  PR was open, ready, and mergeable when this work began.
- Kit's `pkg/cli/repair_context.go` imports `internal/worktree` to prepare exact
  same-repository PR-head and branch worktrees. That behavior remains a Kit
  responsibility even after the generic command leaves.
- Historical completed specs and the generated project progress summary are
  records of what Kit previously delivered. They remain intact rather than
  being rewritten as though `git wt` never existed.
- Kit coordinates the cross-repository transition through
  `docs/programs/git-wt-extraction/PROGRAM.md`. The Kit removal PR must not
  merge before Kura PR #2.
- Kit `main` subsequently merged the coding-agent-first v2 release at
  `0eb617e80839d3bc9ae326dbe3c63ddc5d0b0591`. That release retained `git-wt`
  while replacing substantial command, context, documentation, and test
  surfaces, so this removal must be reconciled on top of v2 rather than
  preserving the older pre-v2 layout.

## REQUIREMENTS

- Delete the `git-wt` command entry point, manpage, CLI-only implementation and
  tests, build/install targets, release artifacts, capability records, and
  current user-facing command documentation.
- Preserve Kit PR and CI repair behavior by moving only its required canonical
  writable-worktree preparation behind a narrowly named internal package.
- Keep native `git worktree` operations authoritative in Kit-managed guidance;
  do not make Kit depend on Kura or another wrapper.
- Keep the canonical worktree reference and its embedded template byte-identical.
- Remove current `git wt` references from runtime source, active docs, build and
  release configuration, while allowing explicit negative policy wording and
  immutable historical records to remain.
- Validate focused repair preparation, the complete Go suite and race suite,
  builds, lint, release configuration/snapshot, capabilities, improve suites,
  source size, vulnerabilities, secrets, feature memory, and project memory.
- Deliver through issue #139, exact branch `GH-139`, and a ready pull request.

Non-goals:

- Do not change Kura's implementation in this Kit PR.
- Do not merge either repository's pull request.
- Do not rewrite historical completed specifications or progress records.
- Do not remove Kit's target-aware PR repair and CI repair workflows.

## ACCEPTED PLAN

1. Reconstruct one Kit-owned program ledger from live Kit and Kura repository
   and GitHub state, with Kura delivery as the prerequisite for Kit removal.
2. Split the worktree ownership boundary: create a minimal internal preparer
   for Kit repair workflows, update its callers and focused tests, then remove
   the generic command implementation.
3. Remove `git-wt` from the Makefile, GoReleaser, command capabilities,
   manpage, README, command guide, canonical native worktree guide/template,
   Constitution architecture map, and related expectations.
4. Search current non-historical surfaces for residual ownership claims and
   reconcile only active documentation and code.
5. Run the complete local validation matrix, curate repository memory, update
   the program checkpoint, and deliver the ready Kit companion PR with an
   explicit dependency on Kura PR #2.
6. When Kit main advances, merge the current base without rebasing, preserve
   the coding-agent-first v2 contract, resolve any feature identity collision,
   and repeat the complete validation and hosted-check cycle on the integrated
   head.

## DECISIONS

- Kit remains responsible for preparing writable worktrees used by `kit pr fix`,
  PR-backed dispatch, review loops, and CI repair. That capability is retained
  as internal product infrastructure, not as an independently installable CLI.
- The new internal package will be named for preparation rather than the old
  product so its boundary is explicit and it cannot accidentally retain the
  generic command surface.
- Active Kit guidance will document the portable native Git workflow only.
  Kura ownership belongs in the local spec and program ledger, not in Kit's
  command catalog.
- The removal PR may be implemented and reviewed while Kura PR #2 is open, but
  its merge is blocked until Kura is merged. This keeps implementation work
  parallel without creating a distribution gap.
- Main's v2 feature already owns stable feature ID `0059`; this removal is
  renumbered from its pre-integration `0059` allocation to `0060` so both
  historical specifications remain addressable without an identity collision.

## DISCOVERIES

- `internal/worktree` is not wholly removable without replacement because
  `pkg/cli/repair_context.go` uses its public preparation methods. The generic
  command and the repair preparer therefore require an explicit source boundary
  rather than a directory-only deletion.
- `docs/references/worktrees.md` and
  `internal/templates/worktrees_reference.md` intentionally mirror each other;
  they must change together.
- Kit's current `main` has no general pull-request CI workflow. Hosted checks
  may therefore be limited to path-filtered Kit Improve and auto-assignment;
  complete local validation remains required and hosted state must be reported
  literally.
- Kura PR #2 merged its validated head unchanged at
  `2ec9cbe058bdf6da8a3a0e1f2b9f6dd717137239` on 2026-08-10. Its merge and
  post-merge CI are proven, while a Kura release and host installation remain
  unobserved.
- Hosted review found that preparation still needed to reject the protected
  primary checkout and revalidate a reused environment link's source type.
  Both checks now fail closed with focused real-Git regression coverage.
- Coding-agent-first v2 removes the legacy capability test file and changes the
  terminal cancellation test inside the generic worktree package. Both
  delete/modify conflicts resolve to deletion because the package and its old
  test surface remain outside Kit after extraction.
- Reconciliation against v2 preserves its local context resolver, reduced
  command tree, usage telemetry, workflow evidence, and release behavior while
  removing only the separately shipped `git-wt` artifact and ownership claims.

## VALIDATION

- Focused real-Git preparation, repair-context, v2 documentation, context, and
  feature-resolution tests passed after the base merge. The complete
  `go test ./... -count=1`, `go test -race ./... -count=1`, and `go vet ./...`
  suites also passed on Go 1.25.12.
- `make build`, `make build-windows`, `goreleaser check`, and
  `goreleaser release --snapshot --clean` passed. All six snapshot archives
  produced from clean merge commit `34bdc6b65e8ce5c84289089c9e4ddce1bf70723e`
  contain only `README.md` plus `kit` or `kit.exe`; no `git-wt` artifact exists.
- Changed-path and whole-repository GolangCI-Lint 2.11.2 both report zero
  issues. The full lint run exposed one orphaned v2 capability test helper after
  the last `git-wt` capability test was removed; deleting that unused helper
  restored a clean whole-project result.
- Built-binary `capabilities --search 'git wt' --json` returns an empty command
  list. Only `cmd/kit/main.go` remains under `cmd/`, and release output contains
  no file named for `git-wt`.
- Improve run `20260811T121642.831412000Z-73940e` passed the default suite with
  8/8 task runs and 18/18 assertions. Run
  `20260811T121642.839737000Z-17830e` passed the prompt-system suite with 24/24
  task runs, 114/114 assertions, and determinism rate 1.0. Both runs record
  merge commit `34bdc6b65e8ce5c84289089c9e4ddce1bf70723e` as provenance.
- Whole-project reconcile reported a complete source-file-size audit of 647
  version-control-eligible candidates and 314 eligible handwritten source/test
  files, with zero above 300 physical lines. Its semantic documentation audit
  was clean; four managed-file refresh candidates remain pending and were not
  applied.
- `kit context resolve` is unblocked for workflow `implementation-delivery`,
  feature `0060-git-wt-removal`, the program ledger, and worktree preparer.
  `kit check git-wt-removal`, `kit check --project`, canonical/template byte
  equality, `git diff --check`, and repeated Gitleaks directory scans passed.
- `govulncheck ./...` reports zero vulnerabilities affecting called code. It
  observes one advisory in a required module whose vulnerable symbols Kit does
  not call.
- Before the v2 base integration, hosted Kit Improve validation and CodeRabbit
  passed at repaired implementation head `594ba9c4674da7287c02d477d131cb7639be940a`,
  and all three valid review findings were resolved. Those pre-integration
  results remain historical evidence only.
- The first integrated PR head `eea9434fc08b2b4cbc6655237d48789b74fca437`
  reached `MERGEABLE/CLEAN` with hosted validation and CodeRabbit passing.
  CodeRabbit identified one new valid test-stability finding: real-Git fixture
  commands were not context bounded. Commit
  `3595b010bc313c11626415e4dbf705b49bf5b443` now uses `exec.CommandContext`
  with `t.Context()` and a ten-second timeout; focused and full Go tests,
  focused race tests, vet, both lint scopes, diff checks, and Gitleaks pass.

## OUTCOME

- Kit now builds, installs, releases, documents, and advertises only the `kit`
  executable. The generic `git-wt` entry point, implementation, manpage,
  capability metadata, dependencies, and command documentation are removed.
- Kit repair commands retain native canonical writable-lane preparation in
  `internal/worktreeprep`, without depending on Kura or exposing a generic
  worktree lifecycle command.
- Active worktree policy is native-only, and historical feature records retain
  explicit decommission notes instead of erasing the prior delivery.
- Kura PR #2 is merged and its post-merge CI passed. Issue #139 and branch
  `GH-139` own the Kit companion delivery. The removal is locally reconciled
  against coding-agent-first v2 at ordinary merge commit
  `34bdc6b65e8ce5c84289089c9e4ddce1bf70723e`, and its stable feature identity
  is now `0060`. The integrated head passed hosted validation and review, and
  the resulting bounded-fixture follow-up is locally validated pending push and
  fresh hosted observation.

## REPOSITORY MEMORY

Decision: updated

Rationale: The source-ownership split, preservation of Kit's internal repair
preparer, historical decommission, and cross-repository delivery evidence are
material rationale that code and tests cannot preserve alone.

Artifacts:

- `docs/specs/0060-git-wt-removal/SPEC.md`
- `docs/programs/git-wt-extraction/PROGRAM.md`
- `docs/specs/0051-resolve-pr-repair-worktrees/SPEC.md`
- `docs/specs/0052-worktree-sync/SPEC.md`
- `docs/CONSTITUTION.md`
- `docs/references/worktrees.md`
