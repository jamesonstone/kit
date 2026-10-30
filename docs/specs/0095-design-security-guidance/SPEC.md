---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0095"
  slug: "design-security-guidance"
  dir: "0095-design-security-guidance"
---
# SPEC

## PURPOSE

Provide useful design and security defaults during initialization and safe upgrades, with contextual agent discovery.

## CONTEXT

Issue: https://github.com/jamesonstone/kit/issues/272. Instruction artifacts already flow through init, reconcile planning, file targeting, atomic writes, linked-target guards, and delivery snapshots. The registry records installed hashes and distinguishes managed from local-custom rules. The universal contract is the canonical contextual discovery surface.

## REQUIREMENTS

Create canonical DESIGN.md and SECURITY.md. Preserve existing/customized documents, including vulnerability-reporting policies. Update unchanged Kit defaults. Preview without writes. Use relevant contextual references, meaningful existence/structure checks, lifecycle regression tests, and concise documentation. No policy engine, new configuration mechanism, security-settings changes, or reconciliation of other projects. Delivery and local binary verification follow the authorized PR gates.

## ACCEPTED PLAN

Embed two concise Markdown templates. Extend instruction-artifact enumeration and reuse the existing registry artifact hash/state contract for whole-document ownership. Missing files are created; exact current or previously installed unmodified defaults converge; all project edits remain intact. Replacement of customized guidance requires explicit --force --file selection. Extend project audit, tests, and documentation, then publish PR #273 from GH-272. Complete the authorized merge only after current CI and automatic review pass, then rebuild the local binary and verify it in disposable projects.

## DECISIONS

Use the existing installed-hash registry instead of a new policy engine or marker protocol. Unlike generated agent adapters, these documents are project-customizable whole documents; preserve every local edit, particularly reporting channels. Unchanged Kit defaults update; edited documents are local-custom with an advisory. Agent contract triggers reference the canonical documents without copying them. Validate only existence, readability, and nonempty content. Project-owned Markdown needs no prescribed heading style; canonical templates supply structured defaults and policy meaning is reviewed semantically.

## DISCOVERIES

Fresh init runs the shared refresh planner after scaffolding; this lets canonical documents adopt installed hashes without a separate init ownership mechanism. File-target refresh must persist ownership metadata even when .kit.yaml was not explicitly selected. Existing linked-target guards and atomic writes remain authoritative. The self-hosted documents are also customizable: unlike generated agent blocks, project guidance is not pinned byte-for-byte by repository tests, which would prevent Kit from maintaining its own reporting policy independently of downstream defaults.

Automatic CodeRabbit review identified two additional preservation/status cases. Existing policies now resolve in `.github/SECURITY.md`, `SECURITY.md`, then `docs/SECURITY.md` order, without creating or adopting a root placeholder unless explicitly forced by file target. Project completeness and audits use the same resolver. Registry status reads current guidance bytes, so edits are immediately local-custom before reconciliation. Follow-up review exposed the complexity of prescribing headings for project-owned Markdown: ATX-only detection rejected Setext, while an added Setext detector misclassified separators. The accepted simpler approach removes the heading parser entirely. Readable nonempty project Markdown remains project-owned, including plain paragraphs, Setext, and separators. Canonical templates remain structured; custom guidance meaning is not mechanically certified. This keeps validation within the requested scope and avoids a Markdown policy engine.

## VALIDATION

Passed `gofmt -l .` (empty output), `git diff --check`, `make vet`, `make test` (complete Go suite), `make lint` (0 issues), `make build`, and `KIT_USAGE_DISABLED=1 bin/kit check --project`. The project check retains the pre-existing nonblocking advisory for the project-owned retired docs/references/README.md. Lifecycle tests cover fresh init, existing reporting-policy preservation, missing-file reconciliation, unmodified-default upgrades, customized documents, blanket versus targeted force, dry-run content/registry immutability, contextual references, registry status, project validation, linked-file safety. The complete validation set passed again after both CodeRabbit fixes, including regressions for all three policy locations, explicit root replacement, contextual precedence, project completeness, and status before reconciliation. Project-check regressions cover ATX and Setext guidance, single-character underlines, plain Markdown, separators, whitespace-only documents, and conventional security-policy locations. Exact-head hosted CI and automatic-review evidence is recorded on PR #273.

## OUTCOME

Implemented embedded guidance and contextual discovery through the existing instruction-artifact flow. Reconciliation uses existing installed hashes and managed/local-custom states; no new policy engine, dependency, or configuration mechanism was added. Both defaults are self-hosted in Kit. Source changes were implemented in the isolated GH-272 worktree and delivered through PR #273. Local binary verification uses disposable project fixtures. No security settings or other repositories are reconciled by this feature.

## REPOSITORY MEMORY

This living spec retains ownership and migration rationale. Updated README, commands, and the Constitution artifact inventory; retained existing universal delivery rules. The full source requirement remains the original user attachment; issue #272 records its engineering and delivery scope.
