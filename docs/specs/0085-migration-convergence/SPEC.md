---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0085"
  slug: "migration-convergence"
  dir: "0085-migration-convergence"
relationships:
  - type: builds_on
    target: 0084-harness-subtraction
    note: Migrates existing projects onto the structure phase 3 defined.
---
# SPEC

## PURPOSE

Make `kit init` and `kit reconcile` converge every supported Kit project onto one canonical structure, so a project no longer depends on which Kit release created it (GH-221, modernization phase 4).

## CONTEXT

- After phase 3, refresh left pre-contract entry files unchanged, never removed retired files, kept retired rule registrations, and still carried v1/v2 generators and v2/v3 support-document audits.
- Real projects (39 local checkouts with a committed `.kit.yaml`) spanned scaffold v1 through the phase 2 contract release. Their retired support documents came from 157 releases whose templates changed often, and many had been hand- or agent-edited.
- Earlier releases fetched rules from the default branch at install time, so installed rules match Kit's rule history rather than the release that installed them.

## REQUIREMENTS

- Fresh init produces the canonical structure; reconcile migrates v1, v2, pre-contract v3, phase 2, and current projects to the same Kit-owned structure. A second reconcile changes nothing.
- Project-owned and edited content is never silently destroyed; ambiguous files are kept and reported.
- Protected developer-experience files keep their behavior. The primary checkout stays read-only for reconcile.
- Legacy support is recognition only; no legacy generator remains.

## ACCEPTED PLAN

1. Fingerprint every release's generated output and all rule history into migration-only data (`internal/legacy`).
2. Migrate entry files onto one managed block, keeping sections Kit never wrote; stop on edited Kit sections unless forced; never guess at broken markers.
3. Remove unedited retired Kit files only when Git can restore them; prune retired registrations and obsolete `.kit.yaml` keys; record scaffold version 4 once entry files converge.
4. Make reconcile apply migration by default, writing to a linked `kit-reconcile` worktree from the primary checkout.
5. Delete v1/v2 generators, the section-merge path, and legacy audits; derive the audit from the same migration plan.
6. Prove convergence and preservation with release-derived fixtures and scratch runs on real projects.

Single-lane execution: the planner, entry-file migration, config, and reconcile changes share one refresh plan.

## DECISIONS

- Ownership is decided by section-level fingerprints of what released Kits generated, not by filenames or templates. `scripts/harvest-legacy-fingerprints.sh` built all 157 release tags, ran `kit init` under the default and each explicit scaffold version, and `internal/legacy/fingerprintgen` recorded fingerprints (36 KB) plus every rule blob in Kit's history. A section matches when its normalized text equals any generated version, which also recognizes files that older refreshes assembled by appending sections from different releases.
- Sections are classified as Kit (matches a release), edited Kit (heading Kit wrote, text no release wrote), or project (heading Kit never wrote). A retired document is removed only when every section is Kit. An entry file with edited Kit sections is left unchanged with a named diagnostic; `--force` replaces Kit sections and still keeps project sections. `--force` never overwrites a whole entry file or the testing reference.
- The progress summary is recognized by the rollup's fixed structure: every release regenerated it wholesale from specs, so a file with exactly that structure is derived Kit data; any extra section makes it project-maintained.
- Removal is soft deletion under the deletion-safety rule: only files tracked and unmodified in Git are removed, so `git checkout HEAD -- <path>` restores them; outside Git, or for uncommitted copies, nothing is removed and the file is reported.
- Retired rules lose their registration either way; an edited copy becomes an ordinary project rule. No retired rule is auto-mapped to a replacement: its replacement is either a default rule reconcile already installs (`delivery`) or no longer exists. `source-file-size` does not set `source_file_line_limit`, because phase 3 made the limit opt-in.
- A rule identical to any version Kit ever shipped is Kit-owned and updates to the embedded version, which fixes projects whose older registries recorded differently computed hashes.
- `instruction_scaffold_version` 4 marks the canonical structure and is recorded only when every entry file carries the block, because pre-contract v3 projects also record 3. A `.kit.yaml` without the key loads as legacy (0), never as current. The Constitution baseline and unedited `CONSTRAINTS` section move to the contract pointer only in the same pass.
- `.kit.yaml` is rewritten only when its decoded content changes (obsolete keys, retired registrations, version); comments or formatting alone never trigger a rewrite.
- From the primary checkout, reconcile creates or reuses branch `kit-reconcile` in `~/worktrees/<owner>/<repo>/kit-reconcile` (from the locally known `origin/HEAD`, no fetch), links `.env`/`.envrc`, applies there, and prints review and PR steps. `kit init --refresh` and `kit health` keep writing in place as explicit commands.
- `kit reconcile` has no interactive menu; `--include-files` is accepted and hidden because files are always included.

## DISCOVERIES

- A released v2 scaffold normally wrote `instruction_scaffold_version: 3`; v1.0.60 is the representative scaffold-2 release and v1.0.0 predates the key.
- Real-project scratch migration found two convergence bugs, both fixed and covered: the Constitution merge appended sections without a trailing newline (second pass not a no-op), and `config.Load` defaulted a missing scaffold version to current.
- `auditStandingAuthorityPolicy` was reachable only through the deleted legacy audit and is now called from the plan-based audit.

## VALIDATION

- PASS: `gofmt -l .` (none), `go vet ./...`, `go test ./... -count=1`, `golangci-lint run ./...` (0 issues), `make build`, `git diff --check`, `kit check --project` on Kit (one non-blocking note for Kit's own `docs/references/README.md`), `kit check 0085-migration-convergence`.
- Fixtures (`pkg/cli/testdata/migration`, exact release output): v1 (v1.0.0), v2 (v1.0.60), pre-contract v3 (v3.0.24), phase 2 (v3.0.25). Each reconciles to the fresh-init structure (scaffold version, config keys, entry blocks, baseline, default rule state and hashes, no retired paths) and the second reconcile plans no changes; a fresh init reconciles as a no-op.
- Preservation and safety tests: project text around the block, edited Kit sections without and with `--force`, incomplete markers and duplicate blocks, edited retired rule, project rule, missing generated files, uncommitted and locally modified retired files, non-Git projects, progress-summary ownership, protected developer files, clean checkouts without `.env`/`.envrc`, primary-checkout worktree creation and reuse.
- Scratch migration of 39 real Kit projects (clones only; originals untouched): 27 converged to version 4, 12 kept entry files whose Kit sections were edited (reported by name), 0 primary-checkout changes, 39/39 second passes planned no changes.

## OUTCOME

Delivered as planned.

## REPOSITORY MEMORY

- Created this spec; updated `docs/CONSTITUTION.md` (convergence, ownership, removal, and reconcile invariants), `README.md`, `docs/commands.md`, `docs/migration-v3.md`, and `docs/README.md`.
