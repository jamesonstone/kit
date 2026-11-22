---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0104"
  slug: "substantive-review-followup"
  dir: "0104-substantive-review-followup"
---
# SPEC

## PURPOSE

Keep independent review coverage when a pull request changes after its initial fresh-context review.

## CONTEXT

Kit PRs #299 and #301 named reviewed commits and described substantive edits after those reviews. The records show when each review ran and leave later coverage unclear. The owner approved a scoped follow-up rule for GH-305.

## REQUIREMENTS

- Keep the first independent cold review mandatory before opening a pull request.
- Record the reviewed commit or identifiable snapshot, scope, and outcome.
- Require independent review of substantive later changes. A lead diff check is sufficient for minor changes.
- Ship the prompts in both provider templates and refresh exact prior copies while preserving custom templates. Leave merge authority and required hosted reviews as their own rules.

## ACCEPTED PLAN

Update the delivery rule and the PR Evidence fields. Retain the previous exact template as a migration input. Verify managed artifact hashes, provider variants, and the project.

## DECISIONS

One pull request Evidence section holds the review history. An independent reviewer may inspect only the affected changes and dependencies in a follow-up, including the original independent reviewer. Classification depends on effect, including base changes, so a new commit alone does not trigger another review.

## DISCOVERIES

The embedded GitHub template is also the basis of the Linear template. One historical body preserves ownership detection for both variants.

## VALIDATION

- Focused template and embedded-rule tests passed, including reconciliation of the previous GitHub and Linear templates and preservation of a customized previous template.
- Build, vet, lint, gofmt, diff check, project check, and feature check passed. The project check reported the existing references compatibility advisory.
- Full local tests hit `TestWorkLedgerRoundTrip` and `TestRunRegistryStatusReportsCurrentAfterRefreshConverges`. Both failures also reproduced on the unchanged base. GitHub CI will run against the published head.
- Cold review at `97665d1` checked the complete rule, template, migration, and tests against GH-305 and found no actionable issues.

## OUTCOME

The embedded delivery rule requires a reviewed snapshot and follow-up only for substantive deltas. Kit's shipped and checked-in PR templates collect that evidence. Reconcile recognizes the previous evidence template by exact bytes for both providers.

## REPOSITORY MEMORY

The delivery rule owns reviewer scope and timing. The PR template prompts for the record.
