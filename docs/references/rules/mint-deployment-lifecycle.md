---
kind: ruleset
slug: mint-deployment-lifecycle
description: Guides Mint releases and environment recovery through repository workflows, immutable artifacts, and verified state.
status: active
registry_scope: downstream
applies_to:
  - mint
  - deployment
  - release
  - hotfix
  - rollback
read_policy_default: conditional
---

# Ruleset: mint-deployment-lifecycle

## Purpose

Build a version once, then promote its immutable artifact through the project's
configured environments. Mint manages selection, review, history and release
publication. Project workflows build, retain, deploy and verify the artifact.

Use repository Actions to request a release, hotfix, rollback or observation.
A reviewed latest-following release has one PR that will track eligible builds
until approval freezes its selection. Completion records the outcome on that
same PR. Prefer a roll-forward release of corrected source when recovery is needed.

## Applies When

- The project uses Mint and the task concerns release adoption, publication,
  deployment, environment state, hotfix, roll-forward, rollback or recovery.
- A `.mint.yaml` file or a pinned Mint workflow is evidence of adoption;
  the file's presence alone does not establish runtime activation.

## Rules

### Establish the project contract

- Inspect `.mint.yaml`, the installed or pinned Mint version, repository Actions,
  activation gates and the project's deployment runbook before choosing an action.
  Do not install Mint or change a non-Mint project's release system by inference.
- Use the repository's published immutable Mint pin. Inspect that version's help
  and adapter contracts when needed; do not assume CLI or workflow inputs from a
  different version. Schema 1 adapters remain supported by Mint v0.4.0; moving
  their policy to schema 2 requires an adapter migration, not just a YAML rewrite.
- Schema 2 supports project-owned environment names and counts. Use the explicit
  requested environment. An omitted environment may use `default_environment`
  or the sole environment; never guess among multiple environments.
- Honor `mode`: `deployment` uses runtime adapters; `package` and `artifact`
  retain project publication workflows and reject runtime operations. Local scope
  does not provision a developer runtime or establish shared deployment evidence.
- Team-owned policies use `authorization: repository-write`: Mint verifies current
  effective GitHub repository write access, including team grants. Optional
  `assignees` only route work and do not grant authority. Legacy `human_login`
  policies retain their restrictions; changing authorization is a separate
  reviewed policy migration, not a target-only deployment request.
- Keep `.mint.yaml` as desired policy. Verified history and timestamped live
  observations are separate evidence; do not write apparent live versions into
  policy or treat an old observation as current runtime truth.

### Kit workflow scaffolding

- `kit init --mint` (new Kit project) or `kit reconcile --mint --dry-run --diff`
  followed by `kit reconcile --mint` (existing project) can create a missing
  controller from an existing schema 2 deployment `.mint.yaml` with
  `authorization: repository-write` and reviewed build/deploy/observe adapters.
  Kit embeds the published Mint v0.5.0 parser and renderer; generation works
  offline and does not require installing or executing Mint in the project.
- Existing controllers, adapters and policies remain project owned, even under
  `--force`. Package/artifact policies retain their publication workflows.
  Generation keeps the activation gate; complete adapter integration and verified
  baselines before enabling it. Scaffolding does not prove activation or deployment.
- Candidate producers must isolate selected source execution in credential-free
  containers or an equivalent project-reviewed sandbox, without host mounts or
  shared trusted caches. Transfer completed artifacts to a fresh trusted publisher;
  do not rebuild or execute selected source after acquiring write/cloud credentials.

### Repository-native operator paths

| Intent | Supported path |
| --- | --- |
| Normal release | Merge source, let the trusted producer retain an eligible candidate, then use the configured automatic, manual or reviewed environment path. |
| Explicit promotion | Use **Mint environment control** Actions `promote` with the environment and optional exact eligible version. |
| Hotfix | Use Actions `hotfix` with a fix PR or issue and reason. A reviewed and merged `hotfix(GH-123): :firetruck: ...` PR title also triggers the trusted isolated hotfix path; multi-environment title routing needs an explicit default. |
| Roll-forward | Merge corrected or reverted source and release a new version through the normal promotion path. |
| Rollback | Use Actions `rollback` with the environment, reason and optional retained verified version. Review the recovery release; omission selects the previous distinct verified artifact. |
| Resume after rollback | Use Actions `resume` as a fresh request. It clears rollback's latest-following pause; it does not remove a fixed policy target. |
| Inspect runtime | Use Actions `observe`, then inspect desired, verified and observed state and observation time. |
| Uncertain completed deployment | Use Actions `reconcile` with the intent and a fresh successful trusted observer run, after inspecting the original completed adapter run. |

- These operations apply to an integrated schema 2 controller. For an existing
  schema 1 consumer, use its actual repository forms and workflows. Do not claim
  generated controls or adapters are installed because the upstream release has them.
- A target-only `.mint.yaml` PR is another request path: change one environment's
  exact `target` and recovery `operation`/`reason`, replacing `follow: latest` when
  selecting a fixed version. `target` and `follow` are mutually exclusive. Requests
  other than `resume` require an exact target. This PR is the release approval PR.
- Keep changes to checks, adapters, runtime configuration, prerequisites and
  publisher authority separate from deployment request approval. A request cannot
  weaken its own authorization. Require checks and independent human approval on
  the current head. Never add a separate status or changelog approval PR.
- Preserve isolated hotfix fixes in the default branch so later releases retain
  the correction. Source review, release approval and deployment are distinct stages.

### Artifact and activation boundaries

- Promote exact immutable digests without rebuilding. Environment tags such as
  `dev` are readable aliases, not artifact identity or deployment evidence. A
  multi-component project requires the complete bundle, not one passing component.
- Keep build configuration separate from approved runtime configuration. Rollback
  restores the retained artifact using current trusted runtime configuration; it
  does not reverse data, migrations or external side effects. Successful rollback
  pauses latest following until a fresh resume. Review configuration recovery
  explicitly if configuration itself caused the failure.
- Honor environment prerequisites and at most one shared canonical publisher.
  Do not require a production environment or assume an ECS/ECR, container, language
  or package-manager implementation. Project adapters own provider operations.
- Policy parsing, workflow generation and version upgrades are preparation.
  Runtime activation needs reviewed adapters, approved configuration identities,
  verified shared baselines and enabled project gates. Never fabricate these or
  enable gates as a side effect of updating Mint. Existing approval rules apply.
- Keep dependency builds separate from release-write tokens and deployment
  credentials. Do not share caches from dispatched hotfix source with trusted builds.

### Recovery and evidence

- Follow `deployment-recovery` within Mint's intent and deployment fences. Never
  clear a fence, invent success, or rerun a deployment with uncertain side effects
  just because a dispatch failed or an observation shows a working service.
- Reconciliation requires an authenticated observation taken after the original
  adapter completed. Its complete artifact and configuration must match the frozen
  intent or previous verified baseline; conflicting evidence retains the fence.
- Retry failed publication separately after verified deployment. Do not redeploy
  to repair publication or replace a canonical release for a configuration-only
  deployment of an already published version.
- Report source/merge, candidate, approval, dispatch, verified deployment,
  publication and live observation separately, with exact environment, version,
  digest, run/PR links and observation time. Dispatch acceptance is not deployment.

## Reference

The [Mint v0.5.0 lifecycle and adapter contract](https://github.com/jamesonstone/mint/blob/b97969136d5a43d0982c46c6f185868db16d14bf/docs/references/environment-lifecycle.md)
provides schema, workflow and manifest details. The project's pinned implementation
and reviewed runbook determine its supported actions. Kit embeds this rule locally;
reading it does not require a network fetch or grant deployment authority.
