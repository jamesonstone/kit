---
kind: ruleset
slug: aws-agent-toolkit-guidance
description: Verifies AWS identity, uses current AWS guidance and tools, and never exposes secret values.
status: active
registry_scope: downstream
applies_to:
  - aws
  - infrastructure
  - secrets
read_policy_default: conditional
---

# Ruleset: aws-agent-toolkit-guidance

## Purpose

- Keep AWS work on the intended account and Region, grounded in current AWS guidance, and free of secret exposure.

## Applies When

- Any AWS-dependent command, code, or infrastructure work.

## Rules

### Identity

- When `.kit.yaml` enables an AWS context, run `kit aws verify` before the first AWS-dependent command and again immediately before any AWS mutation. Treat the verified account, ARN, and Region as authoritative; a profile name alone proves nothing.
- Use the verified profile and Region explicitly on every command. Stop on missing credentials or a mismatch, and never fall back to default, ambient, or other profiles.
- Without a Kit AWS context, confirm identity with `aws sts get-caller-identity` before mutations and ask when the intended account is unclear.

### Guidance And Tools

- Load the host's relevant AWS skills when available; otherwise use current official AWS documentation. Verify uncertain APIs, permissions, quotas, and limits before relying on them.
- Prefer the AWS MCP Server when available; fall back to the AWS CLI. Do not install, authenticate, or reconfigure AWS tooling without explaining and getting approval for it.
- Prefer CDK or CloudFormation for new infrastructure, and follow `infrastructure-change-approval` for every infrastructure mutation.

### Secrets

- Never read secret values into the session: do not call `secretsmanager get-secret-value` or `ssm get-parameter --with-decryption`. Reference secrets by name or ARN and resolve them at runtime in the application or deployment.
