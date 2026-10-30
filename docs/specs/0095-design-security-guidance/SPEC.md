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

Create canonical DESIGN.md and SECURITY.md. Preserve existing/customized documents, including vulnerability-reporting policies. Update unchanged Kit defaults. Preview without writes. Use relevant contextual references, meaningful existence/structure checks, lifecycle regression tests, and concise documentation. No policy engine, new configuration mechanism, settings changes, merge, or installation into other projects.

## ACCEPTED PLAN

Embed two concise Markdown templates. Extend instruction-artifact enumeration and reuse the existing registry artifact hash/state contract for whole-document ownership. Missing files are created; exact current or previously installed unmodified defaults converge; all project edits remain intact. Replacement of customized guidance requires explicit --force --file selection. Extend project audit, tests, and documentation, then publish a ready PR from GH-272.

## DECISIONS

Use the existing installed-hash registry instead of a new policy engine or marker protocol. Unlike generated agent adapters, these documents are project-customizable whole documents; preserve every local edit, particularly reporting channels. Unchanged Kit defaults update; edited documents are local-custom with an advisory. Agent contract triggers reference the canonical documents without copying them. Validate only readable, nonempty Markdown with a heading; project policy is not mechanically enforced.

## DISCOVERIES

Fresh init runs the shared refresh planner after scaffolding; this lets canonical documents adopt installed hashes without a separate init ownership mechanism. File-target refresh must persist ownership metadata even when .kit.yaml was not explicitly selected. Existing linked-target guards and atomic writes remain authoritative. The self-hosted documents are also customizable: unlike generated agent blocks, project guidance is not pinned byte-for-byte by repository tests, which would prevent Kit from maintaining its own reporting policy independently of downstream defaults.

## VALIDATION

Passed `gofmt -l .` (empty output), `git diff --check`, `make vet`, `make test` (complete Go suite), `make lint` (0 issues), `make build`, and `KIT_USAGE_DISABLED=1 bin/kit check --project`. The project check retains the pre-existing nonblocking advisory for the project-owned retired docs/references/README.md. Lifecycle tests cover fresh init, existing reporting-policy preservation, missing-file reconciliation, unmodified-default upgrades, customized documents, blanket versus targeted force, dry-run content/registry immutability, contextual references, registry status, project validation, linked-file safety. Hosted exact-head CI and review remain to be observed on the PR.

## OUTCOME

Implemented embedded guidance and contextual discovery through the existing instruction-artifact flow. Reconciliation uses existing installed hashes and managed/local-custom states; no new policy engine, dependency, or configuration mechanism was added. Both defaults are self-hosted in Kit. Changes are confined to GH-272; the primary checkout is unchanged. Delivery is a ready-for-review PR; no merge, release, installation, or security settings change is authorized.

## REPOSITORY MEMORY

This living spec retains ownership and migration rationale. Updated README, commands, and the Constitution artifact inventory; retained existing universal delivery rules. The full source requirement remains the original user attachment; issue #272 records its engineering and delivery scope.
