---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0106"
  slug: "follow-latest-mint"
  dir: "0106-follow-latest-mint"
---
# SPEC

## PURPOSE

Follow Mint's latest official stable GitHub Release without a version in Kit code or configuration. Let release-only integrations cross Mint's v2 refactor without promoting the legacy deployment controller onto an incompatible runtime.

## CONTEXT

Kit GH-309. The current Mint release is v1.1.2. Kit embeds a v0.9.0 controller renderer and uses fixed v0.5.0 action commits to release Kit. Mint GH-46 preserves release primitives and removes the controller in v2.

## REQUIREMENTS

Resolve the stable release at runtime in new release workflows and in Kit's own release jobs. Preserve existing project-owned workflow structure and reference formats. Refresh supported static release refs during init and reconcile. Keep existing action refs intact when the newer release removes their commands or inputs. Fail explicitly on missing release or renderer support before writing files.

## ACCEPTED PLAN

Use a shared release selection script and the published Mint composite action checkout. Remove the embedded Mint Go dependency and invoke the published CLI only for explicit v1 controller generation. Validate local, workflow and live read-only paths and deliver alongside the research note.

## DECISIONS

Use GitHub's stable latest release, not an unreleased branch or a hardcoded major version. Execute upstream code selected for each workflow run and carry the resolved commit across Kit's split tag and publish jobs. Existing workflow pins remain reviewable and retain their project-owned shape. The v1 controller renderer uses its own immutable action ref. Mint v2 has no controller scaffold.

## DISCOVERIES

GitHub Actions requires a literal step uses value. Mint does not publish a floating latest tag. A checkout inside the repository's Git metadata directory keeps the working tree clean and lets GitHub Actions load Mint by a literal local path. Mint v1.1.2 needs a newer Go toolchain than Kit. The renderer builds in a temporary standalone module so Go can select that toolchain even when the target project has no Go module. Workflow invocations turn off Go workspace discovery so a project go.work cannot absorb Mint's separate checkout. Both release publication paths build Kit artifacts before invoking Mint's toolchain setup. The v2 candidate removes version-config from the Action and keeps release-version-main. Reconcile compares inputs to the published Action to avoid dropping the component scope. A job that keeps an incompatible Mint step also keeps its other Mint refs because the Action invocations share an installed binary on the runner.

## VALIDATION

Local validation covers the Go suite, vet, lint, build, project checks and workflow syntax. Release selection rejects nonstable and malformed tags and missing releases. Live read-only lookup selected v1.1.2, then v1.1.3 after Mint published it during review, without changing Kit. A scratch init produced and linted the release starter. An explicit v1 init rendered a two-environment controller with the verified Mint commit, and a second reconcile dry run proposed zero changes. Release resolution and changelog generation passed against the published CLI on a synthetic Git history. A v2 release is not yet published, so compatibility with its controller removal is exercised with fixtures. Release publication and deployment were not run.

## OUTCOME

Kit and newly scaffolded release steps follow the official latest stable Mint release at workflow runtime. Release selection uses no Kit configuration. The opt-in controller scaffold uses the resolved Mint release instead of a compiled dependency. Existing controllers and installation steps with ambiguous downstream adapter usage retain their references after Mint removes controller support, pending a project-owned migration. Reconciliation with an existing controller can still plan other work after the latest release removes controller generation. A runner without GitHub CLI receives a setup diagnostic.

## REPOSITORY MEMORY

This spec records the release selection and v1 compatibility boundaries. Update the current workflow instructions and README with the resulting interface.
