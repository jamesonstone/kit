---
kind: ruleset
slug: deployment-recovery
description: Recovers stalled authorized deployments through existing pipelines before changing workflows or declaring delivery blocked.
status: active
registry_scope: downstream
applies_to:
  - deployment
  - release
  - recovery
read_policy_default: conditional
---

# Ruleset: deployment-recovery

## Purpose

Complete authorized delivery with safe pipeline recovery and evidence of the
running artifact and product behavior.

## Applies When

- Following a release or deployment, especially when an expected pipeline run
  is missing, stalled, failed, or cancelled.

## Rules

- Follow the project's deployment runbook and existing authority. This rule
  grants no new permission; `infrastructure-change-approval` governs deployment
  effects and workflow changes, and provider rules govern identity and credentials.
- Establish the intended source, artifact, environment, and current pipeline
  stage. A missing publication run, a failed publication, and a failed rollout
  are different conditions. Keep unproven causes `UNKNOWN`.
- After confirming the target, authorization, and retry safety, try a bounded
  recovery through the existing pipeline before proposing workflow changes or
  declaring delivery blocked. Complete root-cause analysis is not required
  before a safe retry. Investigate deterministic or recurring failures; do not
  blindly repeat them or repeat steps with uncertain persistent side effects.

| Observed state | Next action |
| --- | --- |
| Queued or running | Wait with bounded backoff; inspect concurrency or stalled progress before starting another run. |
| Failed or cancelled | Retry the matching run when safe; use its failed stage and logs to guide repair if recovery fails. |
| Missing run | Check the live workflow for a supported start mechanism. Use it when authorized; do not assume a rerun or manual trigger exists. |
| Publication succeeded | Follow the downstream deployment and recover that stage without unnecessarily republishing. |
| Deployment succeeded | Verify the running artifact, health, and required product acceptance and regression checks. |

- Preserve pipeline dependencies and normal automatic deployment. Check the
  provider's rerun semantics; retrying an older source does not deliver newly
  merged work. Confirm any newly started run includes the intended change and stays
  within authorized scope, especially when its branch has advanced.
- If authentication is missing or expired, ask the human to reauthenticate
  through the configured mechanism. Do not switch identities or credentials
  or treat an authentication failure as evidence of a workflow defect.
- An approval boundary blocks that action and its dependents, not every
  recovery path or independent authorized task. If a workflow change is needed,
  explain why supported recovery cannot complete delivery and follow its
  approval boundary. Never bypass protections to manufacture a trigger.
- Before reporting `BLOCKED`, identify the affected stage, attempted or unsafe
  recovery paths, supporting error or missing capability, and smallest action
  needed to proceed. An unexplained historical trigger failure alone is not a
  blocker when a safe authorized recovery path remains.
- Report source/merge, publication, deployment, runtime health, and product
  acceptance separately. Tie deployment evidence to the intended source and
  running artifact; preserve the task's production invariants. Pipeline success
  alone does not establish that existing user workflows still work.
