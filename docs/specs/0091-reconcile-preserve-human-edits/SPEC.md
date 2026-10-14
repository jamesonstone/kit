---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0091"
  slug: "reconcile-preserve-human-edits"
  dir: "0091-reconcile-preserve-human-edits"
---
# SPEC

## PURPOSE

`kit reconcile` must not undo deliberate downstream edits to Kit-managed
files. Running `kit reconcile --all` (v3.0.43) in `jamesonstone/rungrid`
downgraded a Dependabot bump of `actions/github-script` from v9 to v7 and
replaced a human-written README Maintainers line with "Maintained by the
[jamesonstone](https://github.com/jamesonstone) team." GitHub issue #259.

## CONTEXT

- `pkg/cli/init_auto_assign.go` rewrites a Kit-managed
  `.github/workflows/auto-assign.yml` whenever its content differs from
  `templates.BuildAutoAssignWorkflow`. Any action pin bump made by Dependabot
  counts as a difference, so reconcile reverted it.
- `pkg/cli/init_readme_maintainers.go` removed every Maintainers section and
  appended the generated default. This conflicts with the existing
  `readme-header-tagline` rule, which already says to preserve a deliberately
  authored maintainers list.
- GH-194 replaced the hardcoded Kit author credit with an owner-derived
  "team" credit. That wording is wrong for a personal GitHub account.

## REQUIREMENTS

- Reconcile never downgrades a semantic-major action pin (`owner/action@vN`)
  that is newer than the Kit template's pin for the same action.
- The auto-assign template pins the current `actions/github-script` major
  (v9).
- Generated Maintainers wording says "team" only for a confirmed
  organization owner.
- Reconcile keeps human-written Maintainers content, normalizes the heading
  to `## Maintainers`, and keeps the section as the last H2.
- Non-goals: other workflow templates, README badge logic.

## ACCEPTED PLAN

- Bump the template to `actions/github-script@v9` and Kit's own workflow.
- Add `preserveNewerWorkflowActionPins` and apply it to the generated
  auto-assign content before comparing it with the existing file.
- Resolve the owner type with `gh api users/<owner> --jq .type` and choose
  the wording from it.
- Replace a Maintainers section only when its body matches a Kit-generated
  default form; otherwise keep its body.

## DECISIONS

- Do both pin fixes (template bump and newer-major preservation). The bump
  fixes the current case; preservation stops the next Dependabot bump from
  regressing the same way.
- Only `@vN[.x.y]` refs are compared. SHA pins and branch refs have no
  comparable major, so the template pin applies unchanged.
- Unknown owner type (no `gh`, no network, API error) uses the neutral
  "Maintained by [<owner>](...)." wording. Rejected: default to "team",
  because it is wrong for personal accounts, which are the common case for
  Kit users.
- A section is Kit-generated when its body is exactly the fallback line or
  `Maintained by [the ][<x>](https://github.com/<x>)[ team].`. Everything
  else is human-authored. Consequence: the pre-GH-194 hardcoded author line is
  now treated as human-authored and kept. Rejected: special-casing that line,
  because repos such as rungrid kept it on purpose.

## DISCOVERIES

- The `readme-header-tagline` rule already required preserving authored
  maintainers content; the code had drifted from the rule.

## VALIDATION

- `gofmt -l .`, `make vet`, `make lint`, `make test`, `make build`,
  `kit check --project`. Results are in the pull request.
- New tests: `TestRunInitRefresh_KeepsNewerPinnedAutoAssignActionMajor`,
  `TestPreserveNewerWorkflowActionPins`,
  `TestRunInitRefresh_UserOwnedRepoMaintainersDoNotSayTeam`,
  `TestRunInitRefresh_UnknownOwnerTypeMaintainersDoNotSayTeam`,
  `TestRunInitRefresh_ReplacesGeneratedTeamDefaultForUserOwner`,
  `TestRunInitRefresh_PreservesHumanWrittenMaintainers`,
  `TestRunInitRefresh_PreservesCustomMaintainerSectionAndKeepsMaintainersLast`,
  `TestDefaultReadmeMaintainersSectionWording`.

## OUTCOME

- Implemented as planned. Remaining risk: a downstream repo that still has
  the pre-GH-194 hardcoded Kit author line keeps it; edit it by hand if it is
  wrong for that repo.

## REPOSITORY MEMORY

- Updated `docs/references/rules/readme-header-tagline.md` with the
  owner-type wording and the preservation behavior.
- No Constitution change: the behavior is a feature-level contract.
