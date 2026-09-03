---
kind: ruleset
slug: infrastructure-change-approval
description: Fences destructive public-cloud, Kubernetes, and infrastructure-as-code effects behind exact confirmation while allowing additive and rollback-preserving mutations to proceed autonomously.
status: active
registry_scope: downstream
applies_to:
  - cloud
  - infrastructure
  - infrastructure-as-code
  - aws
  - gcp
  - azure
  - kubernetes
  - terraform
  - pulumi
  - cloudformation
  - deployment
  - coding-agent
read_policy_default: must
---

# Ruleset: Infrastructure Change Approval

## Purpose

- Classify public-cloud, Kubernetes, and infrastructure-as-code mutations
  before they run.
- Proceed autonomously when the graph contains only additive or rollback-preserving effects, including routine application operations,
  production activation, and additive IAM, network, or resource
  create-or-update.
- Require exact manual confirmation only for delete, remove, destroy, purge,
  destructive replacement, state removal, history rewrite, data erasure,
  permission revocation, or loss of a supported recovery path.
- Preserve one-pass execution after destructive confirmation while failing
  closed on unresolved destructive ambiguity.

## Applies When

Inspect every public-cloud, Kubernetes, or infrastructure-as-code plan or
diff and classify each effect as create, update, replace, delete, or remove.

Confirmation is required only for destructive effects:

- Deleting, destroying, or removing public-cloud, Kubernetes, or
  infrastructure-as-code-managed infrastructure.
- Destructive replacement, purge, state removal, history rewrite, data
  erasure, permission revocation, or loss of a supported recovery path.
- A merge known to trigger one of those destructive effects. The merge is
  stopped because of the destructive effect, not because merging itself needs
  consent.

Proceed autonomously, with classification recorded, for:

- Additive or rollback-preserving create, update, import, move, or apply of
  public-cloud, Kubernetes, or IaC-managed resources.
- Additive IAM, network topology, persistent data-store, or cluster
  configuration changes that do not revoke permissions or remove a recovery
  path.
- Routine application operations defined below.
- Editing IaC source when Git preserves recovery and the planned graph is
  additive or rollback-preserving.

This rule does not cover read-only discovery. Project-local rules may define
a broader destructive scope. Unresolved destructive-effect classification fails closed.

## Rules

### Routine Application Operations

A routine application operation targets already-provisioned application
compute or artifact hosting. It does not create, replace, or delete provider
resources, infrastructure-class Kubernetes objects, or IaC-managed resources,
and it does not change IAM, network topology, persistent data stores, cluster
control plane, or secrets/KMS.

Routine application operations include:

- shipping a new container image, digest, or application artifact onto an
  existing ECS service, Kubernetes Deployment, or equivalent already-provisioned
  workload or artifact host;
- force-new-deployment, rolling restart, or equivalent restart of an existing
  service;
- operational ECS or equivalent interactions against existing services,
  including describe, logs, health, an update of an existing task definition
  that only changes image or ordinary runtime settings, and desired-count
  adjustments;
- merging a pull request whose only known cloud effect is existing CD rolling
  out a new application image or artifact to already-provisioned targets.

These are not infrastructure-approval batches. Record the target, image or
artifact identity, and workflow when useful. Do not stop for a
confirmation outline. AWS identity verification remains additive for
AWS-dependent work.

### Effect Classification

- Always inspect plans, diffs, and provider previews before mutation.
- Proceed autonomously when every classified effect is additive or
  rollback-preserving.
- Isolate destructive effects from ordinary merge, deploy, and release waves.
- A provider replacement containing a destroy requires exactly one
  exact-target confirmation for that destructive subset.
- Ordinary tracked-source edits remain autonomous when Git preserves recovery,
  unless they remove a supported product, migration, compatibility, or
  recovery surface.
- Explicit user holds such as "keep production default-off" prevail.

### Read-Only Discovery

- Read-only discovery may run before mutation when needed to identify the
  actual target and classify effects.
- Discovery must not alter cloud resources, Kubernetes objects, remote state,
  locks with persistent effects, or repository-owned infrastructure source.
- Verify target identity using the strongest project-local mechanism. For an
  enabled Kit AWS context, the separate AWS context gate remains mandatory.
- If the target cannot be resolved safely, ask for the smallest missing
  identity or scope information before proposing a destructive mutation.

### Consolidated Change Outline

Before the first destructive mutation, create one consolidated outline
containing:

- target identity: provider, account, project or subscription, environment,
  region or zone, cluster, and relevant source paths;
- intended actions: affected resources and whether each will be created,
  updated, replaced, deleted, imported, moved, or applied;
- execution boundary: the ordered batch, tools, and explicit exclusions;
- material impact and risk: availability, data, security or IAM, cost,
  dependencies, and any destructive or irreversible behavior;
- rollback or recovery: how the prior safe state will be restored, or an
  explicit statement that rollback is unavailable and how failure is handled;
- validation: the read-only plan, post-change checks, and evidence that will
  establish the intended result.

The outline may cover multiple providers or tools only when every destructive
target and mutation is included in the same bounded batch.

For a merge-triggered destructive mutation, the outline must additionally
identify:

- the exact PR and triggering workflow;
- target account, environment, region, cluster, project, or subscription;
- expected destructive actions and material impact;
- rollback, recovery, or corrective-PR ownership; and
- the post-merge deployment, runtime, and provider evidence required.

Unknown destructive effects block the merge until inspected. A routine
application operation is not an unknown destructive effect. Do not invent a
confirmation ceremony for additive or rollback-preserving work.

- When the task uses a plan and destructive effects exist, include the
  complete infrastructure outline in that plan instead of creating a separate
  approval ceremony.
- Use read-only discovery to make a destructive outline complete before
  asking. Do not split known destructive changes into several summaries or
  approval prompts.

### Merge And Release Orchestration

- Build the dependency graph and infrastructure outline during analysis.
  Classify effects. Proceed with additive and routine application operations
  without confirmation.
- A merge or release whose only known cloud effect is additive or a routine
  application operation does not require infrastructure-change-approval
  confirmation. Record the triggering workflow and environment.
- Infrastructure deletion, destruction, purge, destructive replacement, and
  state removal are outside an ordinary merge or release-orchestration batch.
  Do not execute them there; isolate them as a separate task governed by this
  rule and `deletion-safety` with its own exact post-outline authorization.
- A non-destructive release batch may continue after the destructive node is
  removed only when the remaining graph remains complete.

### Name-Aware Material AWS Targets

- Treat an AWS destructive infrastructure batch as large or materially risky
  when it affects production or shared infrastructure, spans accounts,
  Regions, or a substantial resource set, or can materially change IAM or
  security, network routing, persistent data, availability, cost, or recovery.
- For such a batch, follow `aws-agent-toolkit-guidance` during read-only
  discovery to resolve the current account display name and Region long name
  where the verified identity, partition, API availability, and permissions
  allow it.
- Show the target once in the consolidated outline as
  `account name (account ID)` and `Region long name (Region code)`. Keep the
  STS-verified account ID, ARN, and Region code authoritative; names are
  display-only operator aids.
- If a name cannot be resolved, state `display name unavailable` beside the
  stable ID or code. Do not guess, change credentials, or broaden IAM access to
  obtain a label.
- Fold this evidence into the existing destructive outline and its one
  confirmation. Do not create a separate identity prompt or approval ceremony.

### One Confirmation And One-Pass Execution

- Obtain one explicit user confirmation of the complete destructive outline
  before editing infrastructure source that removes a recovery path or
  performing a live destructive mutation.
- User approval of a task plan that contains the complete infrastructure
  outline counts as confirmation only when that outline is the destructive
  batch; do not ask again before individual commands.
- Additive or rollback-preserving batches proceed without confirmation. A
  broad request such as "deploy it" or "fix the infra" still never authorizes
  deletion or removal.
- Confirmation authorizes the exact outlined destructive batch, not unrelated
  follow-on changes or an open-ended task-wide infrastructure grant.
- After confirmation, execute the approved destructive implementation,
  application, validation, routine failure recovery, and remaining task work
  to completion in one pass without asking for command-by-command approval.
- Compatible tools and diagnosed retries do not require renewed confirmation
  when the target, intended destructive effect, material impact, and recovery
  boundary are unchanged.

### Deletion And Removal Exception

- Follow `deletion-safety` first: default the resource lifecycle to a
  recoverable soft-delete, disablement, quarantine, retained snapshot, or
  provider recovery control. If hard deletion remains necessary, combine the
  deletion-safety and infrastructure fields into one outline and one exact
  post-outline manual confirmation.
- Deleting, destroying, or removing infrastructure always requires explicit user confirmation after the consolidated outline, even when the initial
  request already asked for or authorized the deletion.
- Merge authorization, routine application operations, image deployment, ECS
  operational interactions, and a broad request such as "deploy it"
  never authorize deletion or removal.
- This includes provider delete or destroy operations, Kubernetes object
  deletion, and infrastructure-as-code edits or plans that remove or
  destructively replace a managed resource.
- One confirmation covers every deletion or removal named in the batch. After
  that confirmation, execute the whole deletion batch and its validation in
  one pass; do not ask again for each resource or command.
- A task-plan approval counts as the required deletion confirmation only when
  the plan contains the complete deletion outline and the user approves it
  after seeing that outline. An earlier broad or detailed request alone never
  satisfies this exception.

### Follow-Up Batches And Material Deviations

When additional destructive infrastructure changes become necessary, use
read-only discovery to collect all then-known changes into one follow-up
outline. Obtain one confirmation for that follow-up batch, execute it to
completion in one pass, and continue the rest of the task. Do not create a
separate prompt for each newly discovered command or resource.

Treat the change as a follow-up destructive batch when any of these fall
outside the confirmed outline or change materially:

- provider identity, account, project, subscription, environment, region,
  zone, or cluster;
- resource set, source scope, or action type, especially a new delete,
  destructive replacement, import, or state move;
- expected availability, data, security, IAM, cost, dependency, destructive,
  or irreversible impact;
- rollback, recovery, validation, or intended outcome;
- an observed plan or provider response that differs materially from the
  confirmed destructive outline.

Stop before the first mutation in the follow-up destructive batch, not before
every subsequent command. Do not re-confirm actions already included in an
approved batch. A newly discovered deletion or removal always uses the
deletion confirmation boundary above. A newly discovered additive or routine
application operation is not a follow-up infrastructure batch.

## Anti-Patterns

- Stopping for infrastructure confirmation on additive create, update, IAM,
  network, or production activation after the user already accepted the task
  or `/goal`.
- Treating an initial request as deletion confirmation before the user sees the
  consolidated deletion outline.
- Asking for infrastructure confirmation before every deployment image update
  or ECS interaction against an existing service.
- Applying a plan whose deletes, destructive replacements, target, or material
  destructive impact were not confirmed.
- Asking for approval before every command or routine retry inside an unchanged
  confirmed destructive batch.
- Hiding uncertainty with generic language such as "minor cloud updates."
- Treating a successful command exit as proof that the intended infrastructure
  state is correct.
- Treating merge itself as needing consent, or starting a merge with unknown
  destructive effects.
- Treating an image-only CD merge as a destructive infrastructure batch.

## Verification

- Confirm additive and rollback-preserving effects proceeded without a
  confirmation prompt.
- Confirm routine application operations, including deployment image updates
  and ECS interactions that do not create or delete infrastructure, did not
  receive an infrastructure-approval prompt.
- For a large or materially risky AWS destructive batch, confirm the outline
  includes the resolved account and Region names where available, always
  includes the stable account ID and Region code, and reports unavailable
  display labels explicitly without broadening access.
- Confirm every deletion or removal received explicit confirmation after its
  complete consolidated outline; confirm the batch was not re-prompted after
  that approval.
- Compare the final provider or infrastructure-as-code plan with the classified
  graph and fail closed on unresolved destructive deviations.
- Verify the target identity again at any project-required mutation boundary.
- Run the outlined post-change checks and report actual evidence, skipped
  validation, partial results, and rollback status literally.
- Confirm additional required destructive changes were consolidated into one
  follow-up batch and received one confirmation before their first mutation.
- For an image-only, additive, or routine-ops merge, confirm deployment effects
  were recorded and that no destructive infrastructure batch was invented.

## Examples

Additive create proceeds autonomously:

```text
Target: GCP project analytics-prod, us-central1, GKE cluster primary.
Actions: create a new payments-worker Deployment, Service, and backend
config in the existing cluster.
Impact: additional compute cost; no planned downtime or data change.
Recovery: Git and the prior manifest remain; delete the new objects after
draining if needed.
Classification: additive. Proceed without a confirmation prompt. Verify
rollout status, ready replicas, and service health.
```

Routine application operation, no infrastructure confirmation:

```text
Existing ECS service api-staging already runs in AWS account 123456789012,
us-east-1. Register a new task-definition revision that only changes the
container image digest and call update-service --force-new-deployment.
This is a routine application operation. Do not request
infrastructure-change-approval confirmation. Verify the deployment and
healthy task count.
```

Image-only merge, no destructive infrastructure batch:

```text
In-scope merge of owner/service#84. Hosted workflow deploy-staging will
roll the new image onto the existing ECS service. No delete or destructive
replace of infrastructure. Record the workflow and environment. Merge
success is not deployment proof.
```

Planned deletion that always requires confirmation after the summary:

```text
Target: AWS account 123456789012, us-east-1, staging VPC.
Actions: delete the unused staging NAT gateway and release its elastic IP.
Impact: staging private subnets lose outbound access until the replacement
path is enabled; no production or data impact; hourly cost decreases.
Recovery: recreate the gateway and re-associate a new elastic IP.
Validation: inspect the final plan, route tables, gateway absence, and billing
inventory.

Proceed with this complete deletion batch?
```

Provider replacement containing a destroy:

```text
The additive update planned an in-place change, but the provider plan now
destroys and recreates the database. Isolate that destructive subset, present
one exact-target outline covering data, downtime, recovery, and validation,
and obtain one confirmation before that destroy. Remaining additive nodes
may continue.
```
