---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: deliver
summary: Version Kit project configuration and bind each AWS-enabled project to one STS-verified CLI profile and account with fast automatic checks and interactive repair.
clarification:
  status: ready
  confidence: 97
  unresolved_questions: 0
feature:
  id: 0041
  slug: aws-config-schema
  dir: 0041-aws-config-schema
relationships:
  - type: builds_on
    target: 0021-project-validation-and-instruction-registry
  - type: related_to
    target: 0020-versioned-instruction-model
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0041-aws-config-schema
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
  - id: config-model
    name: Kit project config model
    type: code
    target: internal/config/config.go
    selector_type: symbol
    selector: Config
    relation: implements
    read_policy: must
    used_for: schema version, AWS context, loading, saving, and project-root discovery
    status: active
  - id: root-command
    name: Root command lifecycle
    type: code
    target: pkg/cli/root.go
    selector_type: symbol
    selector: rootCmd
    relation: implements
    read_policy: must
    used_for: fast automatic config preflight on project-aware commands
    status: active
  - id: init-command
    name: Init command
    type: code
    target: pkg/cli/init.go
    selector_type: symbol
    selector: runInit
    relation: implements
    read_policy: must
    used_for: new-project schema persistence and interactive AWS discovery
    status: active
  - id: instruction-templates
    name: Version 2 instruction templates
    type: code
    target: internal/templates/instruction_templates_v2.go
    selector_type: symbol
    selector: tocRepositoryInstructions
    relation: implements
    read_policy: must
    used_for: always-loaded AWS context hard gate and detailed guardrail routing
    status: active
  - id: command-capabilities-rule
    name: Command capabilities maintainer rule
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    selector_type: artifact
    selector: command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: command metadata updates for config and AWS command surfaces
    status: active
  - id: versioned-instruction-model
    name: Versioned instruction model
    type: feature
    target: docs/specs/0020-versioned-instruction-model/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: conditional
    used_for: persisted-version and safe-migration precedent
    status: active
  - id: project-validation
    name: Project validation and instruction registry
    type: feature
    target: docs/specs/0021-project-validation-and-instruction-registry/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: informs
    read_policy: must
    used_for: mechanical project-contract validation precedent
    status: active
  - id: capabilities-feature
    name: Kit capabilities
    type: feature
    target: docs/specs/0033-kit-capabilities/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: visible command metadata, JSON stability, and project-independent capability behavior
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## THESIS

Version .kit.yaml, add fast programmatic validation on project-aware Kit invocations, and provide an interactive config check/remediation path that can discover AWS CLI profiles, propose a single profile with a default-yes Y/n confirmation, verify the effective AWS account, and add the profile/account binding to the current change set without disrupting noninteractive or machine-readable commands.

## CONTEXT

- `internal/config.Config` is the typed `.kit.yaml` model, but the project config currently has no top-level schema version and no AWS context.
- `config.Load` overlays YAML on defaults, which is useful for compatibility but hides whether a field was physically absent; schema inspection therefore needs to retain raw version presence separately from the effective config.
- `kit init` creates and refreshes `.kit.yaml`, while `kit init --refresh --file=.kit.yaml --force` currently replaces the config with defaults. User-owned AWS context must not be lost during managed refresh.
- The Cobra root currently has no persistent pre-run hook. A shared hook can perform the automatic check, but project-independent commands and machine-readable output must preserve their existing contracts.
- `kit capabilities` is explicitly project-independent and read-only. It must not begin loading or repairing `.kit.yaml` merely because it runs inside a Kit repository.
- The fast path must be local-only: locate the project, read and parse `.kit.yaml`, compare the raw schema version, and return without subprocess or network work when the config is current and complete.
- AWS profile enumeration is local but subprocess-backed. STS caller-identity verification is network-backed and must occur only during explicit or accepted remediation, or through `kit aws verify`.
- A profile name is a local alias. The stable safety assertion is the resolved 12-digit AWS account ID; both must be stored when AWS is enabled.
- Always-loaded repository instructions are intentionally thin. The AWS gate should be a small conditional trigger there, with detailed behavior in `docs/agents/GUARDRAILS.md`.

## CLARIFICATIONS

- The user explicitly requested a versioned `.kit.yaml`, a programmatic config-check command, and an automatic attribute check on Kit use when performance remains fast.
- The user explicitly requested inline remediation and a default-yes `Y/n` prompt when exactly one AWS CLI profile is available.
- Accepted default: expose `kit config check` for explicit validation/remediation and `kit aws verify` for deterministic AWS identity verification.
- Accepted default: automatic checks prompt only in an interactive terminal. Noninteractive, JSON, raw-output, help, completion, version, upgrade, and project-independent capability paths never prompt or corrupt stdout.
- Accepted default: a missing or older supported schema is compatible and repairable; a config declaring a schema newer than the installed Kit binary is a hard error with upgrade guidance.
- Accepted default: `aws.profile` plus `aws.account_id` enables AWS context. `aws.enabled: false` is the explicit project-level opt-out that prevents repeated discovery prompts.
- Accepted default: when several profiles exist, present a numbered selector with a skip/disable choice; do not guess.
- Accepted default: no GitHub delivery is in scope. Work remains in the current dirty change set without staging or commits.
- Confidence is 97%. Remaining implementation details are repository-discoverable and do not require user clarification.

## REQUIREMENTS

- [REQ-01] Add top-level integer `schema_version` to `.kit.yaml` and define one current supported project-config schema version in `internal/config`.
- [REQ-02] New configs and saved migrated configs must persist the current schema version.
- [REQ-03] Config loading must preserve backward-compatible defaults while config inspection still distinguishes missing, older, current, and newer raw schema versions.
- [REQ-04] A newer-than-supported schema must fail before project-aware command execution and tell the user to upgrade Kit.
- [REQ-05] Add visible `kit config check` that validates `.kit.yaml`, reports schema state and semantic findings, and offers interactive repairs when safe.
- [REQ-06] `kit config check --json` must be noninteractive and emit a stable machine-readable report without repairing files.
- [REQ-07] Project-aware Kit commands must run the shared local config inspection automatically before their command body.
- [REQ-08] The current and complete automatic fast path must not execute AWS, Git, GitHub, or network subprocesses and must not write files.
- [REQ-09] Project-independent or output-contract-sensitive commands must skip automatic remediation prompts, including `capabilities`, help, completion, version, upgrade, JSON, and raw `--output-only` paths.
- [REQ-10] Add optional `aws` config with `profile`, quoted-string `account_id`, and an explicit disabled state.
- [REQ-11] Enabled AWS config requires a non-empty profile and an exactly 12-digit account ID.
- [REQ-12] When AWS config is absent and AWS CLI is unavailable or has no profiles, automatic remediation must quietly continue without error.
- [REQ-13] When exactly one AWS profile is discovered, interactive remediation must ask whether to add it using a default-yes `Y/n` prompt.
- [REQ-14] When several AWS profiles are discovered, interactive remediation must present an explicit numbered selection and must not infer one.
- [REQ-15] Before persisting an enabled AWS binding, Kit must call `aws sts get-caller-identity --profile <profile>`, validate the returned account ID, and store the verified profile/account pair together.
- [REQ-16] Partial AWS config must be reported and repairable; failed authentication or account mismatch must never persist a guessed or partial binding.
- [REQ-17] Add `kit aws verify` to resolve the configured profile, reject a conflicting `AWS_PROFILE`, call STS, display the verified profile/account/ARN, and fail nonzero on unavailable credentials or account mismatch.
- [REQ-18] Generated top-level agent instructions must contain a concise conditional AWS hard gate, with detailed verification and stop behavior in generated guardrails.
- [REQ-19] `kit init` must participate in interactive AWS setup for new projects while preserving noninteractive and output-only behavior.
- [REQ-20] Init refresh and force paths must preserve project-specific AWS context unless the user explicitly removes or replaces it.
- [REQ-21] Root help, `kit capabilities`, practical docs, tests, and the project rollup must describe the new config and AWS surfaces truthfully.

Non-goals:

- Enforcing AWS identity outside Kit or replacing IAM, SCP, permission-set, or least-privilege controls.
- Supporting multiple named AWS contexts per project in this first schema.
- Automatically running `aws sso login`, opening a browser, or choosing a profile when multiple candidates exist.
- Adding a generic multi-cloud abstraction before another provider creates a concrete requirement.
- Running AWS discovery or STS on the complete automatic fast path.

## ASSUMPTIONS

- AWS CLI v2 exposes `aws configure list-profiles` and `aws sts get-caller-identity` with the output shapes used by this feature.
- AWS account IDs are stored as strings so leading zeroes cannot be lost by YAML numeric parsing.
- A project with `aws.enabled: false` intentionally opts out of automatic AWS discovery until the config is changed.
- A project requiring several AWS accounts must continue using explicit profiles outside the single-context Kit contract until multi-context support is separately specified.
- No assumption blocks implementation.

## ACCEPTANCE CRITERIA

- [AC-01] `config.Default()` and a fresh `kit init` persist the current top-level schema version.
- [AC-02] Tests distinguish missing, older/current, and newer raw schema versions while preserving existing default-overlay loading behavior.
- [AC-03] A newer schema prevents a project-aware command from running and returns actionable upgrade guidance.
- [AC-04] `kit config check` reports a coherent current config successfully; `--json` emits valid JSON and never prompts or writes.
- [AC-05] Automatic checking runs for a representative project-aware command but skips project-independent and machine-readable/raw-output paths.
- [AC-06] Tests prove the complete fast path performs no AWS subprocess and no config write.
- [AC-07] Missing AWS CLI and zero-profile discovery are clean no-ops.
- [AC-08] A single discovered profile renders a default-yes `Y/n` prompt; accepting verifies STS and atomically persists profile plus account ID.
- [AC-09] Multiple discovered profiles require an explicit selection; declining or selecting the disable option never guesses a profile.
- [AC-10] Partial AWS config, invalid account IDs, authentication failures, and account mismatches are reported without persisting an invalid binding.
- [AC-11] `kit aws verify` succeeds only when STS identity matches `.kit.yaml` and reports profile, account, and ARN.
- [AC-12] Generated instruction tests contain the conditional AWS context hard gate and detailed stop-on-mismatch guidance.
- [AC-13] Init, refresh, and force tests prove schema/AWS fields are created or preserved without breaking noninteractive behavior.
- [AC-14] Root help and capability metadata expose `config`, `config check`, `aws`, and `aws verify` with accurate mutation/network behavior.
- [AC-15] Focused config/CLI tests, `go test ./...`, `go vet ./...`, `git diff --check`, and `kit check 0041-aws-config-schema` pass; `kit check --project` introduces no finding attributable to this feature, with any pre-existing unrelated findings recorded exactly.

## IMPLEMENTATION PLAN

1. Extend `internal/config` with the schema constant, typed AWS config, raw-file inspection, semantic findings, migration helpers, and atomic persistence through the existing save path.
2. Add a small CLI config-check service shared by explicit `kit config check`, the root automatic hook, and interactive `kit init` integration.
3. Keep the automatic hook local-only on the complete fast path; gate prompts on terminal/output mode and isolate project-independent commands.
4. Add AWS CLI discovery and STS verification behind injectable subprocess helpers and bounded contexts so tests remain deterministic.
5. Add `kit aws verify` as the explicit runtime identity check and use the same verifier during remediation.
6. Preserve project-owned config fields during refresh/force instead of resetting the entire config object to defaults.
7. Add the conditional AWS gate to generated instructions and update root help, capabilities, commands docs, and project rollup.
8. Run focused tests while implementing, then full Go tests and Kit document/project checks. Review the final diff for prompt/output regressions and unintended config rewriting.

Rollback: remove the root preflight registration and new commands, retain backward-compatible parsing of `schema_version` and `aws` if configs have already adopted them, and avoid deleting user configuration automatically.

## TASK CHECKLIST

- [x] [TASK-01] Add schema/AWS config model, inspection, validation, migration, and unit tests. [AC-01][AC-02][AC-03][AC-10]
- [x] [TASK-02] Add shared config-check/remediation service and AWS discovery/verifier helpers. [AC-04][AC-07][AC-08][AC-09][AC-10][AC-11]
- [x] [TASK-03] Add `kit config check`, `kit aws verify`, root automatic checking, and init integration. [AC-03][AC-04][AC-05][AC-06][AC-11][AC-13]
- [x] [TASK-04] Preserve AWS/schema state during refresh and force paths. [AC-13]
- [x] [TASK-05] Update generated instructions, root help, capabilities, and practical docs. [AC-12][AC-14]
- [x] [TASK-06] Run full validation, review the diff, update reflection/evidence, and synchronize the rollup. [AC-15]

## VALIDATION MAP

| Acceptance | Validation |
| --- | --- |
| AC-01 | `go test ./internal/config ./pkg/cli -run 'Test(Default|RunInit).*Schema'` |
| AC-02 | `go test ./internal/config -run 'Test.*Schema'` |
| AC-03 | `go test ./pkg/cli -run 'TestAutomaticConfigCheck.*Newer'` |
| AC-04 | `go test ./pkg/cli -run 'TestRunConfigCheck'` plus `go run ./cmd/kit config check --json` |
| AC-05 | `go test ./pkg/cli -run 'TestAutomaticConfigCheck'` |
| AC-06 | focused fast-path test with injected AWS runner and config writer counters |
| AC-07 | `go test ./pkg/cli -run 'TestAWSConfigRemediation.*(MissingCLI|NoProfiles)'` |
| AC-08 | `go test ./pkg/cli -run 'TestAWSConfigRemediation.*Single'` |
| AC-09 | `go test ./pkg/cli -run 'TestAWSConfigRemediation.*Multiple'` |
| AC-10 | `go test ./internal/config ./pkg/cli -run 'Test.*AWS.*(Partial|Invalid|Mismatch|Failure)'` |
| AC-11 | `go test ./pkg/cli -run 'TestRunAWSVerify'` |
| AC-12 | `go test ./internal/templates -run 'Test.*AWS'` |
| AC-13 | `go test ./pkg/cli -run 'TestRunInit.*(Schema|AWS|Force|Refresh)'` |
| AC-14 | `go test ./pkg/cli -run 'Test(RootHelp|Capabilities).*'` and targeted capability JSON commands |
| AC-15 | `go test ./...`; `go vet ./...`; `go run ./cmd/kit check 0041-aws-config-schema`; `go run ./cmd/kit check --project`; `git diff --check`, with unrelated pre-existing findings recorded |

## REFLECTION NOTES

- A version integer plus raw-presence inspection was simpler and safer than a general migration framework. Typed loading keeps backward-compatible defaults, while inspection still detects whether the file itself is missing, older than, current with, or newer than the binary's schema.
- Automatic checking is intentionally bounded to one local file read and local raw/typed YAML decoding on a complete config. AWS profile discovery and STS are deferred until an interactive repair is actually needed or the user explicitly runs `kit aws verify`.
- AWS is modeled as one verified project context rather than a generic provider abstraction. A profile alias is never stored without the account ID returned by STS, and a conflicting ambient `AWS_PROFILE` stops verification instead of silently changing project context.
- Node-level YAML updates preserve unknown and unrelated project fields. Explicit init refresh and force behavior preserve project-owned AWS context rather than reconstructing it from defaults.
- An explicit `aws.enabled: false` records a deliberate opt-out and avoids prompting on every future command. Missing AWS CLI or an empty profile list remains a quiet no-op.
- The project-wide validation command still reports existing metadata debt in features 0032 and 0038. This feature adds no project-check finding and does not broaden scope by rewriting unrelated historical specs.

## DOCUMENTATION UPDATES

- `docs/specs/0041-aws-config-schema/SPEC.md` — canonical feature contract and implementation evidence.
- Generated `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, and `docs/agents/GUARDRAILS.md` templates — updated with the AWS context hard gate.
- `README.md`, `docs/commands.md`, root help, and capability metadata — updated for config inspection, remediation, and AWS verification.
- `docs/PROJECT_PROGRESS_SUMMARY.md` — synchronized with the current delivery phase.

## DELIVERY DECISION

The user requested issue-first pull-request delivery after a clean review. GitHub issue #56 is assigned to `jamesonstone`, and branch `GH-56` was created from refreshed `origin/main`. Delivery targets a ready-for-review PR using the repository template and the required Conventional Commit title shape.

## EVIDENCE

Pre-implementation evidence:

- Live repo inspection confirmed the typed config, root command, init/refresh, instruction-template, capability, and project-validation integration points.
- `go run ./cmd/kit status --json` identified existing managed `.kit.yaml` refresh drift before feature edits; no refresh was applied implicitly.

Implementation and validation evidence:

- `go test ./... -count=1` passed across all Go packages after the final implementation and documentation alignment.
- `go vet ./...` passed.
- `go test -race ./internal/config ./pkg/cli` passed.
- `go build ./cmd/kit` passed; the local build artifact was removed after validation.
- Focused schema, AWS, config-check, capability, and generated-instruction tests passed.
- Tests cover missing/older/current/newer schemas; newer-schema detection before incompatible typed decoding; single-profile default-yes acceptance and decline; multi-profile explicit selection; missing CLI and zero-profile no-ops; STS authentication failure and valid account mismatch without writes; malformed account repair; quoted-account enforcement and repair; automatic fast-path subprocess/write avoidance; init/refresh/force preservation and newer-schema rejection; AWS verification success; and conflicting `AWS_PROFILE` rejection.
- `go run ./cmd/kit config check --json` reported schema version 1 as current and valid. This repository intentionally remains AWS-unconfigured because no Kit-repository AWS context was specified.
- `go run ./cmd/kit capabilities config check --json` and `go run ./cmd/kit capabilities aws verify --json` emitted the documented command contracts, and root help exposed both command groups.
- `go run ./cmd/kit check 0041-aws-config-schema` passed.
- `git diff --check` passed.
- `go run ./cmd/kit check --project` reported 15 unrelated pre-existing findings: two missing-clarification warnings in features 0032 and 0038, plus 13 invalid `governs` relation and `loaded` status errors in feature 0038. It reported no finding for feature 0041, `.kit.yaml`, or the updated instructions.
- Final review found and repaired malformed-account remediation, raw quoted-string validation, quote-format remediation/verification integration, and newer-schema inspection ordering before GitHub delivery began.
- GitHub issue #56 (`https://github.com/jamesonstone/kit/issues/56`) is open and assigned to `jamesonstone`; branch `GH-56` was created exactly from refreshed `origin/main`.
