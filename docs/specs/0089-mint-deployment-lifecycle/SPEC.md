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
