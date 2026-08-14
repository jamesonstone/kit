---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0064"
  slug: "aws-agent-toolkit-guidance"
  dir: "0064-aws-agent-toolkit-guidance"
relationships:
  - type: builds_on
    target: 0041-aws-config-schema
  - type: related_to
    target: 0057-infrastructure-change-approval
references:
  - id: aws-agent-toolkit-setup
    name: AWS Agent Toolkit setup instructions
    type: documentation
    target: https://raw.githubusercontent.com/aws/agent-toolkit-for-aws/refs/heads/main/setup-instructions/setup.md
    relation: informs
    read_policy: must
    used_for: current installation, authentication, verification, and project-rule integration contract
    status: active
  - id: aws-agent-experience-rules
    name: AWS Agent Toolkit experience rules
    type: documentation
    target: https://raw.githubusercontent.com/aws/agent-toolkit-for-aws/refs/heads/main/rules/aws-agent-rules.md
    relation: constrains
    read_policy: must
    used_for: AWS skill, documentation, MCP, CLI, infrastructure-as-code, and secret-safety guidance
    status: active
  - id: aws-context-feature
    name: AWS project context feature
    type: feature
    target: docs/specs/0041-aws-config-schema/SPEC.md
    relation: constrains
    read_policy: must
    used_for: project-bound AWS identity verification and profile behavior
    status: active
  - id: infrastructure-approval
    name: Infrastructure change approval rule
    type: rule
    target: docs/references/rules/infrastructure-change-approval.md
    relation: constrains
    read_policy: must
    used_for: public-cloud and infrastructure-as-code mutation approval
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: guides
    read_policy: must
    used_for: complete local and hosted validation
    status: active
  - id: source-size-rule
    name: Source file size rule
    type: rule
    target: docs/references/rules/source-file-size.md
    relation: constrains
    read_policy: must
    used_for: changed source and test organization
    status: active
skills: []
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Integrate AWS Agent Toolkit guidance into Kit-managed projects so coding agents
use current AWS skills and official documentation, prefer the AWS MCP Server
when available, and fall back to the AWS CLI without weakening Kit's existing
identity, infrastructure-approval, delivery, or secret-safety boundaries.

## CONTEXT

- AWS's current setup instructions install AWS CLI v2, authenticate through
  `aws login`, verify the caller with STS, configure Agent Toolkit skills and
  the AWS MCP Server, and then direct each coding tool to place the AWS
  experience rules in its project rules file.
- The upstream AWS experience rules require checking for a relevant AWS skill,
  preferring that skill over general knowledge, verifying uncertain service
  details against current documentation, preferring the AWS MCP Server over a
  direct CLI interaction, using infrastructure as code, following AWS
  Well-Architected guidance, and applying a dedicated Secrets Manager safety
  flow.
- Copying the complete vendor rules into `AGENTS.md`, `CLAUDE.md`, and Copilot
  instructions would conflict with Kit's thin routing-table architecture and
  would become stale as AWS changes the upstream files.
- Kit already owns project-bound AWS identity through `.kit.yaml` and
  `kit aws verify`. That gate is stricter and more specific than the vendor's
  generic STS verification and must remain authoritative when configured.
- Kit already owns public-cloud and infrastructure-as-code mutation approval
  through `infrastructure-change-approval`; Agent Toolkit guidance must not
  create separate or implied mutation consent.
- Repository rules are downstream registry artifacts. A new rule placed under
  `docs/references/rules/` becomes available through Kit's existing GitHub
  registry and is installed during downstream refresh without a new command or
  policy engine.
- GitHub issue #149, branch `GH-149`, and canonical worktree
  `/Users/jamesonstone/worktrees/jamesonstone/kit/GH-149` own this change. The
  planned end state is one ready pull request targeting `main`.

## REQUIREMENTS

- Add one active downstream ruleset for AWS Agent Toolkit guidance with a
  must-read default and explicit AWS, AWS CLI, AWS MCP, Agent Toolkit,
  infrastructure, documentation, and secrets applicability.
- Keep the ruleset pointer-loaded. Generated project instructions must contain
  only a concise trigger that loads it before AWS-dependent work.
- Treat the upstream setup instructions and AWS experience rules on the AWS
  repository's `main` branch as live vendor sources. Setup or refresh tasks
  must retrieve the current files instead of relying on copied command details.
- Before AWS-dependent work, inspect the current host skill catalog, load the
  most relevant AWS Agent Toolkit skill through the host-supported mechanism,
  and prefer that skill's current guidance over model memory. Do not hard-code
  a `retrieve_skill` tool when the host exposes skills differently.
- Prefer the AWS MCP Server for AWS interactions when available because the
  upstream contract provides sandboxed execution, observability, and audit
  logging. Use the AWS CLI as the explicit fallback when the MCP server is not
  available or cannot perform the required operation, and use it when the user
  explicitly requires CLI execution.
- Resolve uncertain AWS API parameters, permissions, quotas, limits, error
  codes, regions, service availability, and setup commands from current
  official AWS documentation. State the evidence gap and stop rather than
  guessing when authoritative documentation cannot be reached.
- Preserve Kit's AWS context hard gate. When `.kit.yaml` enables AWS, run
  `kit aws verify` before the first AWS-dependent command and again immediately
  before AWS mutation, then use only the verified configured profile.
- When no enabled Kit AWS context exists but an AWS interaction is required,
  verify caller identity with STS and reconcile the account, ARN, region, and
  intended environment explicitly. Never silently accept ambient credentials
  as proof of the target.
- Preserve `infrastructure-change-approval` for AWS resource and
  infrastructure-as-code mutations. Agent Toolkit, skill, documentation, MCP,
  or CLI availability never implies mutation authority.
- Prefer AWS CDK or CloudFormation for infrastructure creation and apply the
  AWS Well-Architected Framework, while preserving stronger repository-local
  architecture and infrastructure rules.
- For secret, credential, API key, token, or password work, load the current
  AWS Secrets Manager skill required by the upstream rule. If it is unavailable,
  fail closed; never retrieve secret values into agent context or call the
  prohibited secret-value APIs as a workaround.
- Keep generated verbose, table-of-contents, and repository-memory instruction
  scaffolds aligned for Codex, Claude, Copilot, support docs, checked-in Kit
  guidance, the references index, and reconcile expectations.
- Add focused ruleset, template, downstream-adoption, and reconcile regression
  tests. Keep every changed handwritten Go source and test file at or below 300
  physical lines and pass complete repository validation.
- Observable acceptance: generated project instructions route AWS-dependent
  work to the new rule; the rule validates as a downstream registry artifact;
  downstream init refresh treats it as mandatory; and regression tests prove
  current AWS-source, skill, MCP/CLI, identity, IaC, documentation, approval,
  and secret-safety requirements.
- Non-goals: installing or authenticating Agent Toolkit in this task, mutating
  AWS resources, changing `.kit.yaml` schema or `kit aws verify`, vendoring the
  complete AWS rule text, hard-coding current Agent Toolkit regions or flags,
  adding a new Kit command, or replacing infrastructure approval.

## ACCEPTED PLAN

1. Add `docs/references/rules/aws-agent-toolkit-guidance.md` as the durable
   downstream rule, translating AWS's vendor-specific rule-file placement into
   Kit's pointer-loaded project architecture while linking the current official
   sources.
2. Add one compact shared AWS guidance routing gate and compose it into every
   active generated project-instruction variant and generated guardrails,
   leaving the existing AWS identity gate intact.
3. Route the new rule through generated and checked-in RLM/reference indices,
   downstream mandatory-rule adoption, and reconcile expectations.
4. Add focused tests for ruleset metadata and semantics, generated instruction
   parity, support-doc routing, refresh adoption, and managed guidance checks.
5. Reconcile the checked-in generated artifacts, minimize unrelated generated
   summary churn, then run formatting, focused tests, full tests, race tests,
   vet, lint, builds, Kit checks, source-size, secret, and whitespace audits.
6. Curate the actual outcome into this spec and demonstrated project invariants
   only where appropriate, explicitly stage GH-149 paths, commit and push as
   Jameson Stone, and open one ready pull request assigned to Jameson Stone.

## DECISIONS

- Accepted: add a dedicated ruleset. The AWS guidance spans tool selection,
  documentation freshness, execution paths, IaC, architecture, and secret
  safety; those responsibilities do not belong inside the identity-only AWS
  context gate or the mutation-only infrastructure approval rule.
- Accepted: reference AWS's moving `main` sources at execution time for setup
  and refresh work. Copying their complete contents into Kit would immediately
  create a second, stale source of truth.
- Accepted: translate `retrieve_skill` into host-supported skill loading. AWS
  defines the required outcome, while Codex and other agents expose different
  skill interfaces.
- Accepted: preserve Kit's stricter local gates. Vendor tooling guidance can
  narrow behavior but cannot weaken project identity or user approval.
- Rejected: paste the complete AWS experience rule into every root instruction
  file. That duplicates policy, expands always-loaded context, and bypasses the
  registry and reconciliation model.
- Rejected: hard-code `us-east-1`, installer flags, or authentication lifetime
  in Kit's durable rule. Those are setup details that the live upstream source
  owns and may change.

## DISCOVERIES

- The Agent Toolkit setup instructions explicitly distinguish the user's
  default AWS Region from the toolkit service Region. This reinforces the
  decision to retrieve the current setup document rather than persist today's
  region behavior in Kit.
- Agent Toolkit skills are visible through the current host skill catalog, but
  the upstream `retrieve_skill` operation is not a portable Kit interface.
- Kit's registry enumerates downstream rules from
  `docs/references/rules/` on GitHub `main`; adding the rule file is sufficient
  to publish it after merge, while refresh installs downstream rules and records
  provenance in each consuming project's `.kit.yaml`.

## VALIDATION

- Retrieved both current AWS-owned `main` sources during implementation. The
  212-line setup document still requires AWS CLI v2, browser-based `aws login`,
  STS identity verification, Agent Toolkit setup, skill-catalog verification,
  and project-rule installation; the 24-line experience rule still requires
  skill-first guidance, documentation verification, MCP preference with CLI
  fallback, infrastructure as code, Well-Architected practice, and secret-safe
  runtime resolution.
- `go test ./internal/templates ./pkg/cli -count=1` and the complete
  `go test ./... -count=1` suite passed.
- `go test -race ./internal/templates ./pkg/cli -count=1` passed for the
  complete affected packages.
- `go fmt ./...`, `go vet ./...`, `go build ./...`, and
  `golangci-lint run --new-from-rev=origin/main ./...` passed with no issues.
- `kit check 0064-aws-agent-toolkit-guidance` and `kit check --project`
  passed. `kit rules list` parses the new active downstream rule and reports it
  as locally untracked until publication from GitHub `main` after merge.
- `kit reconcile --all --output-only` reported no reconciliation needed and a
  complete source-file-size audit: 672 version-control-eligible candidates,
  341 eligible handwritten source/test files, and zero above 300 physical
  lines.
- `gitleaks git --redact --no-banner` scanned 331 commits and approximately
  12.15 MB with no leaks; the final `gitleaks dir --redact --no-banner .`
  working-tree scan checked approximately 4.06 MB with no leaks.
  `git diff --check` passed.
- AWS setup, authentication, AWS API interaction, deployment, runtime, and
  production acceptance were not applicable: this delivery changes project
  guidance and performs no AWS-dependent command or AWS mutation.

## OUTCOME

- Kit now owns one active downstream `aws-agent-toolkit-guidance` rule that
  routes AWS work through current Agent Toolkit skills and official
  documentation, chooses the AWS MCP Server when supported, and preserves AWS
  CLI execution when MCP cannot satisfy the operation or the user explicitly
  requires the CLI.
- Setup and refresh paths retrieve the current official AWS setup and
  experience-rule files instead of preserving drift-prone installer, Region,
  authentication-lifetime, flag, or recovery details in Kit.
- The new rule preserves `kit aws verify` as the authoritative identity gate
  for enabled project contexts and preserves the existing infrastructure
  approval rule as the authority for AWS and infrastructure-as-code mutation.
- Secret handling fails closed on the current `aws-secrets-manager` skill and
  prohibits secret-value retrieval into agent context while retaining the
  upstream dynamic-reference and `asm-exec` runtime-resolution boundary.
- Canonical templates, active checked-in Codex/Claude/Copilot instructions,
  generated Guardrails, RLM routing, the references index, reconcile
  expectations, and mandatory downstream refresh adoption are aligned.
- Focused regression tests enforce ruleset metadata and semantics, every
  supported instruction scaffold, V3 Copilot identity parity, support-document
  routing, registry adoption, and managed-guidance reconciliation.
- No AWS CLI, Agent Toolkit, MCP server, shell configuration, user credentials,
  `.kit.yaml` context, cloud resource, or infrastructure state was installed,
  authenticated, or changed by this repository delivery.

## REPOSITORY MEMORY

Decision: created

Rationale: The vendor-freshness boundary, host-portable skill loading,
MCP-versus-CLI selection, and precedence of Kit identity, approval, and secret
safety are material project rationale that code and tests cannot preserve
alone. The Constitution remains unchanged because this feature applies the
existing registry-backed-rule, evidence-before-mutation, and generated-scaffold
invariants rather than establishing a new project-wide constitutional rule.

Artifacts:

- `docs/specs/0064-aws-agent-toolkit-guidance/SPEC.md`
- `docs/references/rules/aws-agent-toolkit-guidance.md`
- `docs/references/README.md`
