---
kind: ruleset
slug: infrastructure-change-approval
description: Requires a consolidated user-approved outline before public-cloud, Kubernetes, or infrastructure-as-code mutations.
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

- Make public-cloud, Kubernetes, and infrastructure-as-code changes explicit
  and reviewable before mutation.
- Give the user one meaningful approval boundary at the beginning of a bounded
  change batch.
- Preserve autonomous execution and recovery after approval while preventing
  unreviewed scope or impact expansion.

## Applies When

- Creating, updating, replacing, deleting, importing, moving, or applying
  public-cloud resources through AWS, GCP, Azure, or comparable provider
  commands, APIs, SDKs, or consoles.
- Mutating Kubernetes resources or cluster configuration.
- Editing or applying infrastructure-as-code source, configuration, or state,
  including Terraform, Pulumi, CloudFormation, CDK, Bicep, and comparable
  tools.
- Running a deployment path that directly performs one of those covered
  mutations.

This rule does not automatically cover adjacent infrastructure SaaS or general
CI/CD configuration unless the operation directly invokes a covered
public-cloud, Kubernetes, or infrastructure-as-code mutation. Project-local
rules may define a broader scope.

## Rules

### Read-Only Discovery

- Read-only discovery may run before approval when needed to identify the
  actual target and produce an evidence-based outline.
- Discovery must not alter cloud resources, Kubernetes objects, remote state,
  locks with persistent effects, or repository-owned infrastructure source.
- Verify target identity using the strongest project-local mechanism. For an
  enabled Kit AWS context, the separate AWS context gate remains mandatory.
- If the target cannot be resolved safely, ask for the smallest missing
  identity or scope information before proposing mutation.

### Consolidated Change Outline

Before the first covered mutation, present one consolidated outline containing:

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

The outline may cover multiple providers or tools only when every target and
mutation is included in the same bounded batch.

### Confirmation And Execution

- Obtain explicit user confirmation of the complete outline before editing
  covered infrastructure source or performing a live mutation.
- A sufficiently detailed initial request counts as confirmation only when it
  contains the complete required outline and clearly authorizes the exact
  bounded mutations. A broad request such as "deploy it" or "fix the infra"
  is not confirmation.
- Confirmation authorizes the exact outlined batch, not unrelated follow-on
  changes or an open-ended task-wide infrastructure grant.
- After confirmation, execute the approved implementation, application,
  validation, and routine failure recovery to completion without asking for
  command-by-command approval.
- Compatible tools and diagnosed retries do not require renewed confirmation
  when the target, intended effect, material impact, and recovery boundary are
  unchanged.

### Material Deviations

Stop before the next covered mutation, revise the outline, and obtain renewed
confirmation when any of these change materially:

- provider identity, account, project, subscription, environment, region,
  zone, or cluster;
- resource set, source scope, or action type, especially a new delete,
  replacement, import, or state move;
- expected availability, data, security, IAM, cost, dependency, destructive,
  or irreversible impact;
- rollback, recovery, validation, or intended outcome;
- an observed plan or provider response that differs materially from the
  approved outline.

Do not split a known batch into repeated approval prompts. Renew approval only
for a material deviation or a newly proposed batch.

## Anti-Patterns

- Treating the original goal as approval when it does not contain the required
  target, action, impact, recovery, and validation outline.
- Editing Terraform, Pulumi, CloudFormation, CDK, Bicep, or Kubernetes sources
  before the covered batch is confirmed.
- Applying a plan whose deletes, replacements, target, or material impact were
  not in the approved outline.
- Asking for approval before every command or routine retry inside an unchanged
  approved batch.
- Hiding uncertainty with generic language such as "minor cloud updates."
- Treating a successful command exit as proof that the intended infrastructure
  state is correct.

## Verification

- Confirm the outline identifies target, actions, execution boundary, impact,
  rollback or recovery, and validation before the first mutation.
- Confirm the user explicitly approved the outline or supplied and authorized
  an initial request that already meets every outline requirement.
- Compare the final provider or infrastructure-as-code plan with the approved
  batch and fail closed on material deviations.
- Verify the target identity again at any project-required mutation boundary.
- Run the outlined post-change checks and report actual evidence, skipped
  validation, partial results, and rollback status literally.
- Confirm no covered mutation occurred outside the approved batch.

## Examples

Separate outline and confirmation:

```text
Target: GCP project analytics-prod, us-central1, GKE cluster primary.
Actions: update the payments Deployment image and set its minimum replicas to 4.
Impact: rolling restart; no planned downtime or data change; estimated compute cost increases.
Recovery: restore the prior image digest and replica count.
Validation: inspect the server-side diff, rollout status, ready replicas, and service health.

Proceed with this bounded batch?
```

Detailed initial request that can count as confirmation:

```text
In AWS account 123456789012, region us-east-1, update only the existing staging
ECS service desired count from 2 to 3. This adds one task and its normal cost,
does not change data or IAM, and can be rolled back to 2. Verify the account,
service deployment, ready task count, and health check. Proceed with exactly
that change.
```

Material deviation requiring renewed confirmation:

```text
The approved update planned an in-place change, but the provider plan now
replaces the database. Stop and present the replacement, data, downtime,
recovery, and validation implications before any apply.
```
