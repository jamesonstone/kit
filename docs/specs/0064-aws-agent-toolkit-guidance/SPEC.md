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
  - id: aws-cli-profile-region
    name: AWS CLI profile configuration
    type: documentation
    target: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html
    relation: informs
    read_policy: must
    used_for: project default-Region semantics and profile Region preference
    status: active
  - id: aws-enabled-regions
    name: AWS CLI describe-regions reference
    type: documentation
    target: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-regions.html
    relation: informs
    read_policy: must
    used_for: live account-enabled Region discovery
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
identity, infrastructure-approval, delivery, or secret-safety boundaries. Bind
each enabled project AWS context to one explicit default Region selected through
the existing interactive initialization and reconciliation experience.

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
- The existing AWS schema binds a project to a verified profile and account but
  does not persist a Region. A valid profile/account pair therefore takes the
  local-only fast path even though AWS CLI calls still lack the project default
  Region required by the upstream setup and current AWS CLI guidance.
- `kit init` invokes interactive configuration remediation directly, while an
  interactive `kit reconcile` reaches the same remediation through Kit's root
  preflight. Extending that shared path keeps both experiences aligned without
  adding a second Region prompt implementation.
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
  before AWS mutation, then use only the verified configured profile and
  Region.
- Advance the project-config schema to version 2 and require a valid
  `aws.region` whenever AWS context is enabled. Preserve `aws.enabled: false` as
  a complete opt-out with no profile, account, or Region prompt.
- When an enabled profile/account binding lacks a Region, use the verified
  profile and AWS CLI `ec2 describe-regions` response to present a numbered
  selector of Regions currently enabled for that account. Do not guess from
  repository ownership, a profile name, ambient variables, or a copied list.
- Use the same shared remediation from interactive `kit init`, `kit reconcile`,
  and `kit config check`. Noninteractive, JSON, output-only, and dry-run paths
  must remain prompt-free and must not write configuration.
- Preserve the complete local-only fast path after profile, account, and Region
  are valid. Region discovery and STS verification run only during accepted
  interactive remediation or explicit `kit aws verify`.
- Make `kit aws verify` pass the configured Region explicitly to STS and report
  Region alongside the verified profile, account, and ARN.
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
  AWS resources, writing the selected Region into the user's AWS CLI profile,
  supporting multiple default Regions per project, vendoring the complete AWS
  rule text, hard-coding a static Region catalog or current Agent Toolkit
  regions or flags, adding a new Kit command, or replacing infrastructure
  approval.

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
5. Extend project-config schema version 2 with required `aws.region`, add live
   enabled-Region discovery and numbered selection to shared interactive
   remediation, and make verification use and report the configured Region.
6. Add schema, persistence, selector, init/reconcile-preflight, fast-path,
   verification, capability, and compatibility regression coverage.
7. Reconcile the checked-in generated artifacts, minimize unrelated generated
   summary churn, then run formatting, focused tests, full tests, race tests,
   vet, lint, builds, Kit checks, source-size, secret, and whitespace audits.
8. Curate the actual outcome into this spec and demonstrated project invariants
   only where appropriate, explicitly stage GH-149 paths, commit and push as
   Jameson Stone, and update ready pull request #150 assigned to Jameson Stone.

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
- Accepted: version the required Region as project-config schema 2. Existing
  schema-1 projects remain readable and receive the existing accepted migration
  prompt before Region selection instead of being silently rewritten.
- Accepted: discover enabled Regions live through the selected AWS profile.
  A static catalog would drift, while all-Region discovery could offer Regions
  the account has not enabled.
- Accepted: prompt only when enabled AWS context is incomplete. Requiring a
  Region choice on every reconcile would turn a one-time configuration repair
  into recurring noise and would destroy the current local-only fast path.
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
- Official AWS CLI documentation defines the profile Region as the default for
  requests and documents `ec2 describe-regions` as the account-enabled Region
  inventory. The project Region is therefore distinct from the Agent Toolkit
  service Region and must not be inferred from it.
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
- Verified the current official AWS CLI profile-configuration and EC2
  `describe-regions` references. They establish the profile Region as a
  request default and `describe-regions` as the account-enabled Region
  inventory used by the selector.
- `go test ./internal/config ./internal/templates ./pkg/cli -count=1` and the
  complete `go test ./... -count=1` suite passed. Focused coverage includes
  schema validation and persistence, existing profile/account migration,
  account-enabled Region choices including `us-east-1`, `us-east-2`, and
  `us-west-1`, explicit choice persistence, EOF no-write behavior, reconcile
  preflight routing, complete-context no-subprocess behavior, and Region-bound
  STS verification.
- `go test -race ./internal/config ./internal/templates ./pkg/cli -count=1`
  passed for the complete affected packages.
- Verified all CodeRabbit findings against current behavior and fixed the
  valid gaps. Every generated root instruction variant now makes the verified
  account, ARN, and Region authoritative, requires the configured profile and
  Region explicitly where supported, and forbids default, discovered-profile,
  or ambient-credential fallback. Region-selector coverage asserts every
  stubbed option, and a failed Region remediation is atomic even after the user
  accepts schema migration; unrelated AWS failures retain the established
  accepted-schema persistence behavior.
- `go fmt ./...`, `go vet ./...`, `go build ./...`, and
  `golangci-lint run --new-from-rev=origin/main ./...` passed with no issues.
- `go run ./cmd/kit config check --json` reported schema 2, current, valid,
  and the repository's explicit disabled AWS context. Capability inspection
  reports live enabled-Region discovery for interactive repair and configured
  profile-plus-Region use for `kit aws verify`.
- `kit check 0064-aws-agent-toolkit-guidance` and `kit check --project`
  passed. `kit rules list` parses the new active downstream rule and reports it
  as locally untracked until publication from GitHub `main` after merge.
- `kit reconcile --all --output-only` reported no reconciliation needed and a
  complete source-file-size audit: 675 version-control-eligible candidates,
  344 eligible handwritten source/test files, and zero above 300 physical
  lines.
- `gitleaks git --redact --no-banner` scanned 331 commits and approximately
  12.15 MB with no leaks for the guidance baseline; the final
  `gitleaks dir --redact --no-banner .` working-tree scan checked approximately
  4.44 MB with no leaks.
  `git diff --check` passed.
- Live AWS setup, authentication, Region discovery, deployment, runtime, and
  production acceptance were not performed: this repository's AWS context is
  disabled, the CLI interactions are covered with controlled subprocess
  stubs, and this delivery performs no AWS-dependent command or AWS mutation.

## OUTCOME

AWS Agent Toolkit routing and project default-Region configuration are
implemented on the GH-149 / PR #150 lane.

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
- Project config schema 2 requires `aws.region` for enabled AWS contexts. The
  checked-in `.kit.yaml` is migrated to schema 2 while retaining its explicit
  disabled AWS opt-out without profile, account, or Region bindings.
- Interactive `kit init`, `kit reconcile`, and `kit config check` share one
  remediation flow. Existing verified profile/account bindings skip profile
  reselection, fetch the account-enabled Region inventory live, and present a
  numbered default-Region selector; complete bindings remain local-only.
- `kit aws verify` rejects missing or invalid Regions, passes the configured
  profile and Region explicitly to STS, and reports Region with account and
  ARN evidence.
- Region discovery and selection fail without persisting a partial schema or
  AWS binding, including when schema migration was accepted earlier in the
  same interactive flow.
- Secret handling fails closed on the current `aws-secrets-manager` skill and
  prohibits secret-value retrieval into agent context while retaining the
  upstream dynamic-reference and `asm-exec` runtime-resolution boundary.
- Canonical templates, active checked-in Codex/Claude/Copilot instructions,
  generated Guardrails, RLM routing, the references index, reconcile
  expectations, and mandatory downstream refresh adoption are aligned.
- V3 root instructions retain the complete verified-identity contract rather
  than relying on a profile name alone: account, ARN, and Region are
  authoritative, explicit project bindings are required, and ambient or
  default credential fallback is prohibited.
- Focused regression tests enforce ruleset metadata and semantics, every
  supported instruction scaffold, V3 Copilot identity parity, support-document
  routing, registry adoption, and managed-guidance reconciliation.
- No AWS CLI, Agent Toolkit, MCP server, shell configuration, user AWS profile,
  credentials, cloud resource, or infrastructure state was installed,
  authenticated, or changed by this repository delivery.

## REPOSITORY MEMORY

Decision: created

Rationale: The vendor-freshness boundary, host-portable skill loading,
MCP-versus-CLI selection, live enabled-Region discovery, project-local Region
ownership, and precedence of Kit identity, approval, and secret safety are
material project rationale that code and tests cannot preserve alone. The
Constitution remains unchanged because this feature applies the existing
registry-backed-rule, evidence-before-mutation, and generated-scaffold
invariants rather than establishing a new project-wide constitutional rule.

Artifacts:

- `docs/specs/0064-aws-agent-toolkit-guidance/SPEC.md`
- `docs/references/rules/aws-agent-toolkit-guidance.md`
- `docs/references/README.md`
