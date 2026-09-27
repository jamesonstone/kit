---
kind: ruleset
slug: infrastructure-change-approval
description: Batches infrastructure mutations behind one outline and one human approval, with deletion always confirmed.
status: active
registry_scope: downstream
applies_to:
  - cloud
  - infrastructure-as-code
  - kubernetes
  - aws
  - gcp
  - azure
  - terraform
read_policy_default: conditional
---

# Ruleset: infrastructure-change-approval

## Purpose

- Keep irreversible or costly infrastructure changes under explicit human control without prompting for every command.

## Applies When

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, including IAM, network, KMS or secrets, and persistent data stores or schemas.

## Rules

### Routine Operations

- Shipping an already-merged artifact through an existing, repository-approved deployment workflow to already-provisioned resources (image rollout, restart, force-new-deployment) is routine when a standing human grant names that workflow and environment. It needs no infrastructure outline.
- Anything else, including new targets, workflow changes, IAM, network, KMS or secrets, schema or data-loss changes, and creating, replacing, or deleting infrastructure, needs its own approval below. Generic task acceptance never authorizes deployment.

### One Outline, One Approval

- Read-only discovery (plans, diffs, describe calls) may run first, but must not change resources, objects, remote state, or infrastructure source. Verify the target identity (account, project, cluster, region); for AWS follow `aws-agent-toolkit-guidance`.
- Before the first mutation, present one outline: target identity; resources and whether each is created, updated, replaced, or deleted; the ordered execution boundary and exclusions; material impact and risk (availability, data, security, cost); rollback or recovery; and validation.
- Get one explicit approval for that batch, then execute it in one pass without per-command prompts. A task plan that contains the complete outline counts as approval when it names no deletions.
- A material deviation (new target, action type, deletion, or an observed plan that differs from the outline) needs a new outline and approval.

### Deletion

- Deleting, destroying, or removing infrastructure always needs explicit confirmation after the outline, even when the original request asked for it. Follow `deletion-safety`. Merge authority, routine deployments, and standing grants never authorize deletion.
- During merge or release orchestration, isolate infrastructure deletion as a separate task with its own confirmation.
