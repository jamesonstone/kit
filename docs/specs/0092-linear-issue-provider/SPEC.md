---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0092"
  slug: "linear-issue-provider"
  dir: "0092-linear-issue-provider"
---
# SPEC

## PURPOSE

Support GitHub Issues (default) or Linear per repository. Keep Mint changelog
references useful across mixed history, and scaffold Mint Actions where absent.

## CONTEXT

Kit delegates ticket writes to authorized native agents; it has no ticket API
client. Canonical instructions are embedded and reconciled. Mint reads Git history.

## REQUIREMENTS

- Keep existing GitHub behavior and history compatible.
- Offer provider selection at init and persist non-secret workspace/team/project
  settings per repository. Support explicit configuration changes and reconcile.
- Populate Linear tickets with the same ask, scope, acceptance and verification
  content. Verify targets/assignment, duplicates, errors and uncertain writes.
- Use Linear identifiers in branches/commit scopes and full issue URL footers.
- Preserve links and attribution in mixed-history changelogs without reading private
  Linear content or creating live test tickets.
- Keep Mint integration narrow: scaffold files or add preparation to a selected job,
  preserve existing Mint uses and project workflow choices, and describe ambiguity.

## ACCEPTED PLAN

Implement repository provider selection and canonical ticket guidance; consume
Mint’s offline link support and add minimal, project-owned Actions scaffolding.

## DISCOVERIES

Kit creates tickets through native-agent instructions. Mint needs explicit URL
parsing from commit footers, without a new provider API or credential store.

## DECISIONS

Native agent capabilities own ticket API operations, auth and lifecycle actions.
Kit stores no credentials and adds no provider SDK or background issue service.
Tests of ticket writes are shipped-instruction contracts, not live API tests.

Issue configuration uses one planner separately on the invoking and canonical
write targets. Patching preserves unrelated YAML/comments and rejects concurrent
edits. Dry runs stay nonmutating and expose canonical-target guidance in text and
JSON target_notice. Configuration changes report explicit reconcile next actions.

Mint v0.9.0 contains the companion offline Linear footer parser. Kit consumes that
released library and pins scaffolded Actions to its verified commit. Consumer pins
and installed tools are separate surfaces. Mint never fetches private Linear issue
content; committed full URLs provide links, with commit references for attribution.

No workflows: provide a project-owned starter using canonical Mint release
resolution, changelog/artifact and publication steps. Existing Mint Action use:
preserve it without assessing completeness. Other existing workflows: require
explicit workflow/job/before-step placement and insert preparation only, preserving
existing bytes and settings. Ambiguous placement reports guidance and creates no
extra diagnostic workflow. Keep structural YAML/path and concurrent-write safety.

Kit does not certify or enforce Mint release triggers, publisher ownership,
credentials, permissions, concurrency, runners, checkout policy or runtime lifecycle.
Removed the trigger parser, remote-HEAD query/bounds framework, publisher detection,
release-owner flag, handoff heuristics and adoption assessments. No adapter framework
replaces them. Existing pre-PR controller scaffolding remains available.

Explicit kit mint check/update and ordinary installed-only reconcile upkeep remain
as separately requested. Missing Mint is skipped by reconcile without network or
installation; only explicit update can install it. Supported outdated Go-bin tools
update; current/newer tools are preserved. Unsupported ownership/dev/symlink cases
fail safely; dry runs do not install. No repository/CI pin changes follow from local
maintenance. The redundant proposed kit install command was removed before merge.

## VALIDATION

Keep representative provider defaults/config/switching, custom-template/config
preservation, identifier/footer instructions, dry-run/no-write/failure boundaries,
backend version decisions, file edit safety and real mixed-history changelog tests.
Removed redundant command matrices, copied input maps and now-obsolete trigger,
remote and publication-policy tests. One resolve step-ID assertion remains because
Actions output references depend on it; the integration fixture supplies its own ID.

Run full Go tests, vet, build, lint, formatting and project checks before delivery.
Independent correctness review and final-head hosted CI/review evidence belong on
the live PR. An existing edited retired README advisory is nonblocking.

## OUTCOME

Kit issue GH-264 / branch GH-264 / PR #265 remains open for review, with no merge
authorized. Mint #33 merged under separate authorization and released v0.9.0.
No global tool, live Linear workspace, consumer workflow or production system was
changed. Real ticket execution needs an explicitly authorized test destination.

## REPOSITORY MEMORY

This spec records the final feature boundary and the rejected enforcement layer.
README carries essential usage only; Maintainers remains its final section.
