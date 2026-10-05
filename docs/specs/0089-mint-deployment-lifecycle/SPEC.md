---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0089"
  slug: "mint-deployment-lifecycle"
  dir: "0089-mint-deployment-lifecycle"
---
# SPEC

## PURPOSE

Teach all Kit-managed agents the Mint release and deployment lifecycle without requiring users to explain its CLI structure.

## CONTEXT

Mint v0.4.0 is published at a435d0dd5bddeaec34f7523e344d093184fcabbf. Kit embeds default rules and routes them from one universal contract. Deployment recovery already supplies generic retry guidance. The separate Kit pin upgrade PR #246 remains unchanged.

## REQUIREMENTS

Install a default conditional Mint rule through init and reconcile across supported generations. Preserve locally customized guidance and non-Mint behavior. Explain repository Actions, YAML requests, arbitrary environments, package/artifact modes, immutable bundles, state evidence, rollback pauses and activation boundaries. Do not activate deployments or change cloud settings.

## ACCEPTED PLAN

Add one embedded downstream rule, route it in the canonical contract and all checked-in agent entries, index it, and test installation, migration, convergence and local customization preservation. Validate the complete Go suite and project gates; deliver a ready PR.

## DECISIONS

Use conditional default distribution instead of an optional rule: every project receives it without inferring adoption or activation. Keep detailed guidance outside the universal contract. Pin the reference to the reviewed feature-bearing Mint source; actual consumer pins and runbooks remain authoritative. Do not add Kit deployment commands or duplicate Mint implementation.

## DISCOVERIES

Existing registry metadata provides default installation; no new registry engine or configuration field is needed. Current schema 1 adopters need their real workflow paths rather than an invented schema 2 activation.

## VALIDATION

PASS: `go test ./pkg/cli -run TestMintLifecycle -count=1`, `go test ./...`, `make vet`, `make lint` (zero issues), `make build`, `bin/kit check --project`, gofmt and diff checks. Distribution tests exercise fresh init, all migration fixtures, absent-rule repair, second-reconcile convergence and local rule preservation. Project validation reports one nonblocking compatibility advisory for the existing customized references index. Guidance reviewed against the published Mint v0.4.0 lifecycle reference and CLI help; this validates distribution and documented contracts, not agent performance or live deployments.

## OUTCOME

Implemented a default embedded rule and conditional routing in every agent entry file. Kit initialization and reconciliation distribute the guidance without requiring an optional-rule selection. Projects receive it after installing the release containing this change and reconciling; no downstream rollout, deployment activation, consumer merge or cloud mutation occurred. Ready PR delivery is the remaining publication boundary.

## REPOSITORY MEMORY

Created this spec and mint-deployment-lifecycle rule; updated the reference index and universal contract routing. The Constitution remains unchanged because the distribution architecture is unchanged.

Team ownership update: the rule now documents repository-write authorization and separate optional assignment. Its adapter reference points to Mint team-authorization source b97969136d5a43d0982c46c6f185868db16d14bf; legacy consumers retain their existing restrictions until a reviewed migration.

## Workflow scaffolding (GH-249)

Accepted extension: `kit init --mint` and `kit reconcile --mint` create a missing environment controller from the project's existing schema 2 `.mint.yaml`. Reuse Mint v0.5.0's public policy parser and workflow renderer as a compiled Go dependency, with the corresponding immutable action commit. No network lookup, Mint executable installation or project code execution is needed for generation. Only deployment-mode, repository-write policies qualify; package/artifact consumers keep their publication flows. Configured adapter files must exist, and their workflow names supply exact callback triggers. Existing policy, controller and adapter files remain project owned, including under force. Project environments remain arbitrary; no provider, registry, language, environment count or production requirement is introduced.

Scaffolding is explicitly requested, retains `MINT_RELEASE_ENABLED` and creates no activation evidence. Reconcile includes generation in its normal dry-run, atomic apply and protected worktree delivery paths. Kit's own legacy publication actions move to the published v0.5.0 pin without changing their CLI contract. The existing personal policy switches to repository-write authorization.

Validation for GH-249: PASS full `make test`, `make vet`, `make lint` (zero issues), `make build`, project check (one existing nonblocking references-index advisory), formatting and workflow lint. New native tests cover environment choices/callbacks, token/cache boundaries, unsupported modes and personal policies, missing adapters, link rejection, preflight before init mutation, force preservation, dry-run and repeated generation, and primary-checkout delivery to a linked worktree. A fresh CLI invocation generated the controller, preserved policy bytes, passed actionlint and reconciled in dry-run. Generation was reviewed against the published renderer; provider adapters, runtime activation and deployment were not executed. Existing controllers stay project owned and require their own reviewed upgrades; no provider-specific producer/deployer scaffold is invented.

Adversarial review: reject duplicate adapter workflow names and the controller's own name before rendering. GitHub workflow_run matches names, so distinct file paths alone cannot prevent ambiguous or self-triggering callbacks. Expression-bearing callback names are rejected by the upstream renderer. Native regression tests cover each case.
