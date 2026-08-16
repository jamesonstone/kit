---
kind: ruleset
slug: infrastructure-change-approval
description: Requires one plan-level confirmation and one-pass execution per infrastructure batch, with name-aware material AWS targets and explicit deletion confirmation.
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
- Give the user one meaningful approval boundary per bounded change batch.
- Preserve one-pass execution and autonomous recovery after approval while
  preventing unreviewed scope, deletion, or impact expansion.

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
- Merging a pull request known to trigger deployment, Kubernetes,
  public-cloud, or infrastructure-as-code mutation. The merge is part of the
  covered mutation boundary even though the provider mutation occurs
  indirectly in a workflow.

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

Before the first covered mutation, create one consolidated outline containing:

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

For a merge-triggered mutation, the outline must additionally identify:

- the exact PR and triggering workflow;
- target account, environment, region, cluster, project, or subscription;
- expected infrastructure actions and material impact;
- rollback, recovery, or corrective-PR ownership; and
- the post-merge deployment, runtime, and provider evidence required.

Unknown triggering effects block the merge until inspected. A single accepted
plan may contain both the exact merge authorization and this infrastructure
approval. Do not ask twice when that one complete plan satisfies both
contracts.

- When the task uses a plan, include the complete infrastructure outline in
  that plan instead of creating a separate approval ceremony.
- Use read-only discovery to make the outline complete before asking. Do not
  split known changes into several summaries or approval prompts.

### Name-Aware Material AWS Targets

- Treat an AWS infrastructure batch as large or materially risky when it
  affects production or shared infrastructure, spans accounts, Regions, or a
  substantial resource set, or can materially change IAM or security, network
  routing, persistent data, availability, cost, or recovery.
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
- Fold this evidence into the existing consolidated outline and its one
  confirmation. Do not create a separate identity prompt or approval ceremony.

### One Confirmation And One-Pass Execution

- Obtain one explicit user confirmation of the complete outline before editing
  covered infrastructure source or performing a live mutation.
- User approval of a task plan that contains the complete infrastructure
  outline counts as confirmation; do not ask again before individual commands.
- For a batch with no deletion or removal, a sufficiently detailed initial
  request also counts as confirmation only when it contains the complete
  required outline and clearly authorizes the exact bounded mutations. A broad
  request such as "deploy it" or "fix the infra" is not confirmation.
- Confirmation authorizes the exact outlined batch, not unrelated follow-on
  changes or an open-ended task-wide infrastructure grant.
- After confirmation, execute the approved implementation, application,
  validation, routine failure recovery, and remaining task work to completion
  in one pass without asking for command-by-command approval.
- Compatible tools and diagnosed retries do not require renewed confirmation
  when the target, intended effect, material impact, and recovery boundary are
  unchanged.

### Deletion And Removal Exception

- Follow `deletion-safety` first: default the resource lifecycle to a
  recoverable soft-delete, disablement, quarantine, retained snapshot, or
  provider recovery control. If hard deletion remains necessary, combine the
  deletion-safety and infrastructure fields into one outline and one exact
  post-outline manual confirmation.
- Deleting, destroying, or removing infrastructure always requires explicit
  user confirmation after the consolidated outline, even when the initial
  request already asked for or authorized the deletion.
- This includes provider delete or destroy operations, Kubernetes object
  deletion, and infrastructure-as-code edits or plans that remove or replace a
  managed resource.
- One confirmation covers every deletion or removal named in the batch. After
  that confirmation, execute the whole deletion batch and its validation in
  one pass; do not ask again for each resource or command.
- A task-plan approval counts as the required deletion confirmation only when
  the plan contains the complete deletion outline and the user approves it
  after seeing that outline. An earlier broad or detailed request alone never
  satisfies this exception.

### Follow-Up Batches And Material Deviations

When additional covered infrastructure changes become necessary, use read-only
discovery to collect all then-known changes into one follow-up outline. Obtain
one confirmation for that follow-up batch, execute it to completion in one
pass, and continue the rest of the task. Do not create a separate prompt for
each newly discovered command or resource.

Treat the change as a follow-up batch when any of these fall outside the
approved outline or change materially:

- provider identity, account, project, subscription, environment, region,
  zone, or cluster;
- resource set, source scope, or action type, especially a new delete,
  replacement, import, or state move;
- expected availability, data, security, IAM, cost, dependency, destructive,
  or irreversible impact;
- rollback, recovery, validation, or intended outcome;
- an observed plan or provider response that differs materially from the
  approved outline.

Stop before the first mutation in the follow-up batch, not before every
subsequent command. Do not re-confirm actions already included in an approved
batch. A newly discovered deletion or removal always uses the deletion
confirmation boundary above.

## Anti-Patterns

- Treating the original goal as approval when it does not contain the required
  target, action, impact, recovery, and validation outline.
- Treating an initial request as deletion confirmation before the user sees the
  consolidated deletion outline.
- Editing Terraform, Pulumi, CloudFormation, CDK, Bicep, or Kubernetes sources
  before the covered batch is confirmed.
- Applying a plan whose deletes, replacements, target, or material impact were
  not in the approved outline.
- Asking for approval before every command or routine retry inside an unchanged
  approved batch.
- Interrupting execution with repeated prompts for infrastructure actions that
  were already included in the approved plan.
- Prompting separately for several follow-up changes that read-only discovery
  could consolidate into one new batch.
- Hiding uncertainty with generic language such as "minor cloud updates."
- Treating a successful command exit as proof that the intended infrastructure
  state is correct.
- Treating merge authorization as infrastructure approval, or starting a
  merge with unknown covered deployment effects.

## Verification

- Confirm the outline identifies target, actions, execution boundary, impact,
  rollback or recovery, and validation before the first mutation.
- For a large or materially risky AWS batch, confirm the outline includes the
  resolved account and Region names where available, always includes the
  stable account ID and Region code, and reports unavailable display labels
  explicitly without broadening access.
- Confirm the user approved the plan or outline once for the complete batch, or
  supplied a qualifying initial request for a non-deletion batch.
- Confirm every deletion or removal received explicit confirmation after its
  complete consolidated outline; confirm the batch was not re-prompted after
  that approval.
- Compare the final provider or infrastructure-as-code plan with the approved
  batch and fail closed on material deviations.
- Verify the target identity again at any project-required mutation boundary.
- Run the outlined post-change checks and report actual evidence, skipped
  validation, partial results, and rollback status literally.
- Confirm no covered mutation occurred outside the approved batch.
- Confirm additional required changes were consolidated into one follow-up
  batch and received one confirmation before their first mutation.
- For a merge-triggered mutation, confirm the workflow, exact target, expected
  actions and impact, recovery, and post-merge evidence were included before
  the merge and that one complete accepted plan was not redundantly confirmed.

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

Detailed non-deletion initial request that can count as confirmation:

```text
In AWS account payments-production (123456789012), Region US East
(N. Virginia) (us-east-1), update only the existing staging ECS service desired
count from 2 to 3. This adds one task and its normal cost, does not change data
or IAM, and can be rolled back to 2. Verify the account, service deployment,
ready task count, and health check. Proceed with exactly that change.
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

Follow-up batch required by a material deviation:

```text
The approved update planned an in-place change, but the provider plan now
replaces the database. Consolidate the replacement with any other newly
required changes, present one follow-up outline covering data, downtime,
recovery, and validation, and obtain one confirmation before that batch.
```
