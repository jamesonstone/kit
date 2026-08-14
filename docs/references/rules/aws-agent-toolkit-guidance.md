---
kind: ruleset
slug: aws-agent-toolkit-guidance
description: Routes AWS work through current Agent Toolkit skills, official documentation, MCP or CLI execution, verified identity, infrastructure approval, and secret-safe handling.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - aws
  - aws-cli
  - aws-mcp
  - agent-toolkit
  - cloud
  - infrastructure
  - documentation
  - secrets
read_policy_default: must
---

# Ruleset: AWS Agent Toolkit Guidance

## Purpose

- Keep AWS-dependent work aligned with current AWS Agent Toolkit skills and
  official AWS documentation instead of model memory or copied setup details.
- Define one deterministic choice among an AWS skill, the AWS MCP Server, and
  the AWS CLI without weakening Kit's project identity, mutation approval, or
  secret-safety rules.
- Keep vendor guidance current by linking AWS-owned sources rather than
  vendoring their complete contents into Kit-managed project instructions.

## Applies When

- A coding agent plans or performs work involving AWS services, APIs, SDKs,
  CLI commands, infrastructure, deployments, permissions, quotas, regions,
  observability, security, credentials, or secrets.
- A task installs, repairs, updates, or verifies AWS CLI v2 or the AWS Agent
  Toolkit.
- A task must choose between an Agent Toolkit skill, the AWS MCP Server, and
  direct AWS CLI execution.

## Rules

### Current Vendor Sources

- For Agent Toolkit setup, repair, or refresh, retrieve and read the complete
  current AWS setup instructions before acting:
  `https://raw.githubusercontent.com/aws/agent-toolkit-for-aws/refs/heads/main/setup-instructions/setup.md`.
- Retrieve and read the complete current AWS experience rule before setup or
  when this rule may have drifted:
  `https://raw.githubusercontent.com/aws/agent-toolkit-for-aws/refs/heads/main/rules/aws-agent-rules.md`.
- Treat those AWS-owned `main` files as the authority for installer URLs,
  supported operating systems, authentication flow, setup flags, toolkit
  service Region, skill-catalog verification, credential lifetime, and error
  recovery. Do not preserve those changing details as Kit invariants.
- If either required source is unavailable, report the evidence gap and stop
  the affected setup or refresh path. Do not substitute remembered commands.
- Explain each installation or authentication mutation and its purpose. Never
  request an access key, secret key, session token, password, or other AWS
  credential from the user; follow the current AWS browser-based login flow.

### Skill And Documentation Routing

- Before AWS-dependent work, inspect the host's current skill catalog and load
  the most relevant AWS Agent Toolkit skill through the host-supported skill
  mechanism. Prefer the loaded skill's current instructions over general model
  knowledge.
- Do not assume the host exposes a tool literally named `retrieve_skill`.
  Use it when provided by Agent Toolkit; otherwise use the host's documented
  skill discovery and loading surface.
- When no relevant skill is available, use current official AWS documentation
  and state that the skill layer was unavailable. A missing skill is not
  permission to guess service behavior.
- Verify uncertain API parameters, IAM permissions, quotas, limits, service or
  Region availability, error codes, setup commands, and compatibility against
  current official AWS documentation. Prefer an AWS MCP documentation
  capability when available, then official AWS web documentation.
- State unresolved uncertainty explicitly and stop before a mutation whose
  target, permissions, effect, or recovery cannot be confirmed.

### MCP And CLI Execution

- Prefer the AWS MCP Server for supported AWS interactions when it is
  available, authenticated, and compatible with the task. Its sandboxing,
  observability, and audit trail are part of the selection decision.
- Use the AWS CLI directly when the AWS MCP Server is unavailable, does not
  support the required operation, the user explicitly requires the CLI, or MCP
  would obscure a required project-local command contract. Report the
  selection and preserve the same identity, approval, and validation
  boundaries.
- Do not install, update, authenticate, or reconfigure the AWS CLI, Agent
  Toolkit, MCP server, skills, shell startup files, or user configuration as an
  incidental prerequisite. Perform setup only when the user requested it or
  accepted it as part of the bounded task.
- Use explicit Regions for Region-scoped operations and use ASCII hyphens, not
  em dashes, in AWS resource names and descriptions.

### Identity And Mutation Authority

- When `.kit.yaml` defines an enabled AWS context, run `kit aws verify` before
  the first AWS-dependent command in the task and again immediately before an
  AWS or AWS-backed deployment mutation.
- Treat the verified account ID and ARN as authoritative. Use only the
  configured verified profile for AWS CLI, SDK, CDK, CloudFormation,
  Terraform, deployment, and project-script operations where supported.
- After Kit verification fails, never fall back to default, ambient, or another
  discovered profile.
- When no enabled Kit AWS context exists but AWS access is required, run an STS
  caller-identity check through the selected AWS execution path and reconcile
  the account, ARN, Region, and intended environment. Ambient credentials alone
  are not proof that the target is correct.
- Before mutating AWS resources or infrastructure-as-code source,
  configuration, or state, load and follow
  `docs/references/rules/infrastructure-change-approval.md`. Skill, MCP, CLI,
  login, or identity success never supplies mutation approval.

### Infrastructure And Secret Safety

- Prefer AWS CDK or CloudFormation for infrastructure creation over direct
  imperative resource creation. Preserve stronger repository-local
  infrastructure ownership and architecture rules.
- Apply the AWS Well-Architected Framework to material architecture and
  operational decisions, using current official guidance for the affected
  workload and service.
- Before work involving a secret, credential, API key, token, or password,
  load the current Agent Toolkit `aws-secrets-manager` skill. If that required
  skill is unavailable, stop the secret-handling path.
- Never call `secretsmanager get-secret-value` or
  `secretsmanager batch-get-secret-value`, and never call a Secrets Manager
  agent daemon directly to place secret material in agent context.
- Use the current skill's runtime-resolution flow, including an AWS Secrets
  Manager dynamic reference and `asm-exec`, so the consuming process resolves
  the value without exposing it to the agent, logs, prompts, or evidence.

## Anti-Patterns

- Copying the complete AWS setup or experience rule into `AGENTS.md`,
  `CLAUDE.md`, Copilot instructions, prompts, or another always-loaded file.
- Treating an installed AWS CLI, a profile name, successful login, or ambient
  credentials as proof of the intended account or environment.
- Guessing from model memory when a relevant current AWS skill or official
  documentation is available.
- Using direct CLI calls merely because MCP discovery was skipped, or using
  MCP merely because it exists when it cannot preserve the required command or
  evidence contract.
- Running installer pipelines, editing shell startup files, opening an auth
  flow, or reconfiguring user tooling as an unannounced prerequisite.
- Creating infrastructure imperatively when the repository owns it through CDK
  or CloudFormation.
- Retrieving secret values into agent context, command output, logs, temporary
  evidence, or repository files.
- Treating vendor guidance as permission to bypass Kit's AWS identity,
  infrastructure approval, work-lane, delivery, or testing rules.

## Verification

- Confirm the current upstream setup and AWS experience rule were retrieved
  when setup, repair, refresh, or drift analysis was in scope.
- Record the relevant AWS skill loaded, or state why no relevant skill was
  available and which official documentation replaced it.
- Record whether the AWS MCP Server or AWS CLI was selected and why.
- Confirm the exact AWS account, ARN, Region, environment, and profile source
  before any AWS-dependent mutation.
- Confirm `kit aws verify` ran at both required boundaries when the project has
  enabled AWS context.
- Confirm infrastructure approval was recorded before AWS or
  infrastructure-as-code mutation and that no additional target or effect was
  silently added.
- Confirm secret workflows loaded `aws-secrets-manager`, used runtime
  resolution, and emitted no secret material.
- Report local validation, AWS interaction evidence, deployment/runtime
  evidence, and production acceptance as separate claims.

## Examples

Current setup-source check:

```text
Retrieve setup-instructions/setup.md and rules/aws-agent-rules.md from the
official aws/agent-toolkit-for-aws main branch before changing local setup.
```

Execution selection:

```text
Relevant AWS skill loaded -> current official docs checked -> AWS MCP Server
used when supported; otherwise AWS CLI with the same verified profile.
```

Identity boundary:

```text
.kit.yaml AWS enabled -> kit aws verify -> read-only AWS work -> kit aws verify
again -> approved AWS mutation.
```
