---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0053
  slug: testing-and-environment-validation
  dir: 0053-testing-and-environment-validation
relationships:
  - type: builds_on
    target: 0021-project-validation-and-instruction-registry
  - type: builds_on
    target: 0049-application-architecture-rules
  - type: related_to
    target: 0031-executable-verification-harness
references:
  - id: rules-registry
    name: Rules registry
    type: code
    target: pkg/cli/rules_registry.go
    relation: implements
    read_policy: must
    used_for: downstream ruleset visibility and refresh adoption
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: implements
    read_policy: must
    used_for: code-level CI, environment suites, evidence, and production safety
    status: active
  - id: testing-reference
    name: Testing reference
    type: documentation
    target: docs/references/testing.md
    relation: guides
    read_policy: must
    used_for: project-specific commands, environments, automation, and known gaps
    status: active
  - id: rlm-routing
    name: RLM routing
    type: documentation
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: just-in-time loading of the mandatory testing rule
    status: active
  - id: instruction-templates
    name: Repository instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: downstream testing and validation routing
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: create and verify issue 98 and its pull request
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the validated issue branch as a ready pull request
    required: true
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Provide one mandatory downstream testing rule that keeps language-native
code-level tests and pull-request CI authoritative while adding consistent
local, live-integration, and production end-to-end validation, evidence, and
production-safety contracts across Kit-managed projects.

## CONTEXT

- Kit distributes active downstream rules from `docs/references/rules`, but it
  currently has no universal rule for code-level test ownership, PR CI,
  environment suites, high-level evidence, or post-deployment validation.
- The existing application architecture rules mention tests at layer
  boundaries but do not define a complete cross-project correctness model.
- The managed `docs/references/testing.md` starter is intentionally empty and
  does not tell projects which commands, environments, automation, or gaps to
  record.
- End-to-end coverage must supplement unit, component, contract, repository,
  and integration tests rather than moving or replacing them.
- Local and production suites need durable, agent-readable evidence without
  turning tracked Markdown into an append-only execution log.
- Production end-to-end validation may require bounded synthetic writes, but
  must never use customer data, broad cleanup, infrastructure mutation, or
  unverified targets.
- Browser automation across Kit-managed repositories can leak user-installed
  Chrome processes and macOS code-sign clones when browser selection, session
  ownership, retries, and cleanup are left implicit.
- GitHub issue #98 and branch `GH-98` own this Kit-only change. Migrating the
  motivating LabCore scripts and workflows is a separate downstream task.
- GitHub issue #121 and branch `GH-121` own the browser-lifecycle extension to
  the same canonical testing rule. The extension remains Kit-only and does not
  mutate downstream repositories or global browser tooling.

## REQUIREMENTS

- Add an active downstream `testing-and-environment-validation` ruleset with
  `read_policy_default: must`.
- Require applicable unit, component, contract, repository, integration,
  generated-code, static-analysis, and other code-level checks at the
  narrowest useful layer.
- Require applicable code-level checks in pull-request CI. End-to-end tests
  supplement and never replace those checks.
- Run hermetic, stable, time-bounded local end-to-end suites in PR CI when
  feasible; otherwise require literal milestone evidence before handoff.
- Organize high-level executable tests by type and then environment without
  relocating language-native tests:
  `tests/end-to-end/{local,production}` and
  `tests/live-integration/{local,production}`.
- Standardize immutable evidence at
  `tmp/<UTC-date>/<stable-test-id>/<positive-run-number>/`, allocate without
  overwriting, and require redacted `output.txt`, machine-readable
  `result.json`, and clearly named structured artifacts.
- Define `result.json` fields for suite, environment, run identity, timing,
  result, exit code, source/deployment identity, assertions, cleanup, and
  artifact paths.
- Define the result vocabulary as `PASS`, `FAIL`, `PARTIAL`, `BLOCKED`,
  `SKIPPED`, and `NOT_APPLICABLE`.
- Require a tracked, curated `tests/RUN_STATUS.md` with one current row per
  suite and environment. Preserve material failures without appending every
  repeated successful execution.
- Require exact target and deployed-version verification before production
  validation and prevent a production-validated claim after a failed suite.
- Permit bounded synthetic production writes only with dedicated credentials,
  limits, redaction, cleanup or documented retention, and names beginning
  `kit-e2e-<project>-<environment>-<run-id>-<resource>[-<ordinal>]`.
- Require cleanup to select both the test marker and exact run ID. Prohibit
  customer data, unrelated records, production reset, infrastructure or shared
  configuration mutation, and authentication weakening.
- Default automated browser work to Playwright-managed Chromium or Chrome for
  Testing. Treat the user's installed, auto-updating Google Chrome as an
  explicit, justified exception rather than a default.
- Require one uniquely named browser session per task or test run, reuse that
  session across repeated operations, and prohibit unbounded instance-creating
  retries.
- Distinguish task-owned browser sessions, processes, and automation daemons
  from user-owned browsers. Close task-owned resources and detach from
  user-owned resources without closing them.
- Put scoped browser cleanup in unconditional lifecycle handling that runs on
  success, failure, cancellation, timeout, and interrupted validation, then
  verify that every task-owned browser resource exited.
- Treat browser cleanup failure as validation failure. Prohibit broad process
  termination, shared temporary-directory deletion, code-sign protection
  weakening, and routine clone-directory deletion as cleanup substitutes.
- Treat read-only production smoke as `PARTIAL` when safe write isolation is
  unavailable.
- Require CI to upload run directories as artifacts without committing raw
  evidence or automatically changing `RUN_STATUS.md`.
- Keep non-deployable projects compliant through code-level CI and an explicit
  `NOT_APPLICABLE` production status rather than artificial environments.
- Expand the managed testing reference with project-specific commands, suite
  inventory, preflights, credential names, cleanup and retention, automation,
  manual fallbacks, and known gaps.
- Add concise generated V2/V3 and checked-in instruction routing without
  inlining the rule into always-loaded files.
- Verify ruleset validity, downstream refresh adoption, generated/checked-in
  instruction alignment, testing-reference alignment, and registry index
  visibility.

### Non-goals

- Claiming that any suite proves absolute or literal 100 percent correctness.
- Replacing language-native tests with shell scripts or high-level suites.
- Imposing one test framework, coverage percentage, deployment platform, or
  rollback mechanism on every project.
- Automatically creating test directories, workflows, credentials, or
  production data in downstream repositories.
- Migrating LabCore scripts, workflows, or evidence in this issue.
- Adding a browser launcher, wrapper, or daemon to Kit when repository
  inspection finds no Kit-owned browser automation implementation to enforce.
- Modifying global Codex skills, browser plugins, installed browsers, macOS
  code-signing settings, or downstream repositories.

## ACCEPTED PLAN

1. Add the canonical mandatory downstream rule with explicit code-level,
   high-level suite, evidence, status-map, automation, and production-safety
   contracts.
2. Expand the managed testing reference and ruleset index so projects can
   record their concrete commands, environment prerequisites, automation, and
   gaps.
3. Add concise testing gates to generated V2/V3 repository instructions,
   Copilot guidance, and RLM routing, then align checked-in artifacts.
4. Add focused ruleset, refresh-adoption, reference-template, and instruction
   routing tests.
5. Validate the focused and complete repository surfaces, curate repository
   memory, and deliver issue #98 through `GH-98` as a ready pull request.

Issue #121 extends that accepted plan without replacing it:

1. Add browser-automation applicability and one required browser lifecycle
   section to the existing canonical ruleset.
2. Route browser automation and browser testing explicitly through generated
   V2/V3 instructions, Copilot guidance, and RLM support guidance.
3. Strengthen ruleset and downstream refresh tests to assert lifecycle,
   ownership, safety, and exact propagated content.
4. Do not add enforcement code unless repository inspection finds a Kit-owned
   browser launcher, wrapper, example, or validation helper.
5. Run focused and complete validation, including output-only whole-project
   reconcile, then deliver issue #121 through `GH-121` as a ready pull request.

## DECISIONS

- Use one unified rule because code-level CI and environment suites are layers
  of one correctness model; separating them would make it easier for
  end-to-end guidance to displace lower-level tests.
- Use capability-based automation: code-level PR CI is mandatory, local
  end-to-end CI is required when hermetic and practical, and production
  automation is preferred with a documented operator fallback.
- Treat “near 100% correctness” as a risk-based confidence objective backed by
  explicit evidence and literal residual gaps, never as a guarantee.
- Keep `tests/RUN_STATUS.md` tracked and curated rather than append-only.
  Immutable per-run detail belongs under ignored `tmp/` and CI artifacts.
- Allow bounded synthetic production writes because read-only probes cannot
  prove every workflow, but fail closed to `PARTIAL` when safe isolation is
  unavailable.
- Do not prescribe automatic rollback. A failed production suite blocks the
  validation claim and defers remediation to the project release policy.
- Keep browser lifecycle in the unified testing rule because browser process
  ownership and cleanup are validation responsibilities, not a separate Kit
  subsystem.
- Route browser-specific work explicitly while retaining the existing
  mandatory pre-implementation and pre-validation load gate for all projects.
- Do not invent a Kit browser wrapper solely to encode policy. Enforcement code
  is required only at an implementation surface Kit actually owns.

## DISCOVERIES

- `kit init --refresh` installs every active downstream registry ruleset and
  records its managed state, so no separate registry catalog entry is needed.
- Mandatory registry installation alone does not guarantee just-in-time
  loading; generated provider instructions and RLM need concise routing.
- The V3 testing reference is generated from the shared
  `referencesTesting` template, so the checked-in reference and generated
  downstream file can be kept identical by focused tests.
- The Kit source checkout reports a newly added registry rule as `untracked`
  until the rule merges into the configured `main` registry source. Local
  parsing, refresh-adoption tests, and generated-routing tests provide the
  pre-merge distribution evidence.
- Repository-wide searches found no Kit-owned Playwright, Chromium, Chrome for
  Testing, Google Chrome, browser-session, or browser-daemon implementation;
  the owned enforcement surfaces are the downstream rule, generated routing,
  registry refresh, reconcile, and their tests.

## VALIDATION

- `go test ./internal/templates -run
  'Test(InstructionTemplatesRouteTestingAndEnvironmentValidation|MemoryRepositoryInstructionsRouteApplicationArchitecture|MemoryRepositoryInstructionsRouteConstitutionCuration)$'
  -count=1` passed.
- `go test ./pkg/cli -run
  'Test(TestingAndEnvironmentValidationRegistryRulesetIsValid|RunInitRefresh_InstallsMandatoryDownstreamRules)$'
  -count=1` passed.
- `make fmt` passed.
- `go test ./... -count=1` passed across every package.
- `go vet ./...` passed.
- `go test -race ./internal/templates ./pkg/cli -count=1` passed.
- Both `cmd/kit` and `cmd/git-wt` built successfully into ignored `bin/`
  artifacts without installing over a user binary.
- `golangci-lint run --new-from-rev=origin/main ./...` passed with
  `0 issues`.
- `./bin/kit rules view testing-and-environment-validation` parsed and rendered
  the local mandatory downstream rule.
- `./bin/kit rules list` showed the active local rule with all expected
  applicability tags and the expected pre-merge `untracked` registry state.
- `./bin/kit check --all` passed all 50 features, including feature 0053.
- `./bin/kit check --project` passed after the completed feature was added to
  the curated project progress summary.
- `./bin/kit status --json` and
  `./bin/kit reconcile --include-files --dry-run --diff` were reviewed. The
  only proposed managed-file changes were unrelated `.kit.yaml` source-commit
  metadata churn and creation of `.envrc`, so they were intentionally not
  applied in this feature.
- `git diff --check` passed.
- No local or production end-to-end suite was applicable to this Kit
  registry/template change. LabCore environment execution remains explicitly
  outside issue #98.

### Issue #121 browser lifecycle extension

- Focused `pkg/cli` ruleset, exact refresh-adoption, reconcile-guidance, and
  stale-guidance tests passed with `-count=1`.
- `go test ./internal/templates -run
  'TestInstructionTemplatesRouteTestingAndEnvironmentValidation$' -count=1`
  passed for V2/V3, Copilot, RLM, and checked-in V3 routing.
- `make fmt`, `go test ./... -count=1`, `go vet ./...`, and
  `go test -race ./internal/templates ./pkg/cli -count=1` passed.
- `cmd/kit` and `cmd/git-wt` built independently into ignored `bin/`
  artifacts; no global binary or browser tooling was installed or changed.
- `golangci-lint run --new-from-rev=origin/main ./...` passed with `0 issues`.
- `./bin/kit rules view testing-and-environment-validation` parsed and rendered
  the browser lifecycle section, and `./bin/kit rules list` showed
  `browser-automation` and `browser-testing` applicability.
- `./bin/kit check --project` passed, and `./bin/kit check --all` passed all 52
  feature contracts including feature 0053.
- `./bin/kit reconcile --all --output-only` reported no reconciliation needed
  and a complete source-file audit: 812 version-control-eligible candidates,
  506 eligible handwritten source/test files, and 0 above 300 physical lines.
- `./bin/kit reconcile --all --include-files --dry-run --diff --output-only`
  proposed only `.kit.yaml` registry bookkeeping. It correctly treated the
  locally changed testing rule as `local-custom` until issue #121 merges into
  the configured `main` registry source; the dry-run diff was not applied.
- `git diff --check` passed. No browser end-to-end execution applied because
  Kit owns policy, routing, registry refresh, and reconcile rather than a
  browser runtime or launcher.

## OUTCOME

- Added the active mandatory downstream
  `testing-and-environment-validation` rule.
- Preserved language-native code-level tests and pull-request CI as the
  correctness foundation, with environment suites defined as additive
  evidence.
- Standardized type-first high-level suite layout, immutable UTC/run-number
  evidence, machine-readable results, curated status reporting, capability-
  based automation, and literal partial or non-applicable states.
- Added the universal `kit-e2e-` synthetic-resource identity, metadata,
  production preflights, exact-run cleanup proof, and prohibited mutation
  boundaries.
- Routed generated V2/V3 agents, Copilot, and RLM to the rule and project
  testing reference without inlining the full policy.
- Expanded the managed testing reference with project-specific commands,
  suite inventory, environment preflights, credentials, evidence, automation,
  fallbacks, and known gaps.
- Added focused validity, downstream refresh-adoption, generator-alignment,
  and registry-index tests.
- Kept the delivery Kit-only; no LabCore files, workflows, tests, or production
  systems changed.

### Issue #121 browser lifecycle extension

- Added one required `Browser Automation Lifecycle` section to the existing
  canonical testing rule, covering managed Chromium selection, explicit
  installed-Chrome exceptions, named session reuse, task versus user ownership,
  unconditional cleanup, scoped termination, exit verification, and macOS
  code-sign safety.
- Added `browser-automation` and `browser-testing` applicability and surfaced
  browser lifecycle ownership in the maintained rules index.
- Routed browser automation and browser testing explicitly through generated
  V2/V3 agents, Copilot instructions, RLM guidance, and checked-in V3 artifacts.
- Strengthened ruleset and refresh-adoption tests so the lifecycle contract and
  exact registry source content propagate together through Kit's reconcile
  refresh path.
- Added no browser wrapper because Kit owns no browser launcher, session
  manager, daemon, validation helper, or example. No global Codex files,
  installed browsers, macOS settings, or downstream repositories changed.

## REPOSITORY MEMORY

Decision: updated

Rationale: The testing taxonomy, environment evidence interface, curated
status model, capability-based automation, and production synthetic-data
safety defaults now include durable browser lifecycle ownership and cleanup
requirements that code and tests alone cannot preserve across downstream
projects. The existing Constitution already defines downstream registry
distribution and repository-memory behavior; no project-wide Constitution
invariant changed, so the Constitution remains unchanged.

Artifacts:

- `docs/specs/0053-testing-and-environment-validation/SPEC.md`
- `docs/references/rules/testing-and-environment-validation.md`
- `docs/references/testing.md`
- `docs/agents/RLM.md`
- `docs/references/README.md`
- `docs/PROJECT_PROGRESS_SUMMARY.md`
