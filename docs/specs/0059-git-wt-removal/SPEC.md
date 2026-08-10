---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0059"
  slug: "git-wt-removal"
  dir: "0059-git-wt-removal"
relationships:
  - type: builds_on
    target: 0051-resolve-pr-repair-worktrees
  - type: related_to
    target: 0052-worktree-sync
references:
  - id: "feature-notes"
    name: "Feature notes"
    type: "notes"
    target: "docs/notes/0059-git-wt-removal"
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

## VALIDATION

- Focused real-Git preparation and repair-context tests passed for canonical
  local and remote branches, exact reuse, same-repository PR heads, fork and
  detached-lane refusal, environment linking, collision rollback, and regular
  `.envrc` preservation.
- `go test ./... -count=1`, `go test -race ./... -count=1`, and `go vet ./...`
  passed.
- `make build`, `make build-windows`, and `goreleaser check` passed.
  `goreleaser release --snapshot` built six archives; their contents were only
  `README.md` plus `kit` or `kit.exe` and contained no `git-wt` artifact.
- `golangci-lint run --new-from-rev=origin/main ./...` passed with zero issues.
  Whole-repository lint still reports 43 pre-existing findings outside this
  change.
- Built-binary capability inspection returned no command beginning with
  `git wt`, and a `git wt` search returned an empty result.
- Improve run `20260810T211539.932313000Z-4a2ec7` passed the default suite with
  8/8 task runs and 16/16 assertions. Run
  `20260810T211544.542200000Z-24f50d` passed the prompt-system suite with 45/45
  task runs, 345/345 assertions, and determinism rate 1.0.
- Whole-project reconcile reported a complete source-file-size audit of 475
  eligible handwritten files with zero above 300 physical lines. Its three
  managed-refresh candidates are the two canonical rules intentionally changed
  by this branch plus pre-existing `.kit.yaml` registry refresh state; no
  refresh was applied.
- `kit check git-wt-removal`, `kit check --project`, canonical/template byte
  equality, `git diff --check`, and Gitleaks directory scanning passed.
- `govulncheck ./...` reports the existing Go 1.25.5 standard-library baseline:
  13 reachable advisories, all fixed by later Go 1.25 patch releases. This PR
  does not change Kit's Go toolchain and introduces no new dependency.

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
  `GH-139` own the Kit companion delivery. Ready Kit PR #140 is open and
  mergeable and contains validated implementation commit
  `3d1d882f6b3c0a404f9667c3dc5217636dc75cee`; hosted validation was still
  running when this checkpoint was recorded.

## REPOSITORY MEMORY

Decision: updated

Rationale: The source-ownership split, preservation of Kit's internal repair
preparer, historical decommission, and cross-repository delivery evidence are
material rationale that code and tests cannot preserve alone.

Artifacts:

- `docs/specs/0059-git-wt-removal/SPEC.md`
- `docs/programs/git-wt-extraction/PROGRAM.md`
- `docs/specs/0051-resolve-pr-repair-worktrees/SPEC.md`
- `docs/specs/0052-worktree-sync/SPEC.md`
- `docs/CONSTITUTION.md`
- `docs/references/worktrees.md`
