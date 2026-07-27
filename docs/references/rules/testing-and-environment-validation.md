---
kind: ruleset
slug: testing-and-environment-validation
description: Defines code-level PR checks, environment suites, immutable evidence, status reporting, and safe production end-to-end validation.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - implementation
  - testing
  - validation
  - ci
  - github-actions
  - deployment
  - local
  - production
  - end-to-end
  - live-integration
read_policy_default: must
---

# Ruleset: testing-and-environment-validation

## Purpose

- Make correctness validation a layered, repeatable part of development and
  deployment in every Kit-managed project.
- Preserve fast, precise language-native tests while adding environment-level
  evidence for complete workflows and real integrations.
- Produce immutable, redacted evidence that subsequent tests, coding agents,
  and humans can inspect without overstating what a suite proved.
- Keep production validation isolated, bounded, attributable, and safe.

## Applies When

- A Kit-managed project implements, changes, validates, deploys, or releases
  code, build behavior, generated contracts, CI, or runtime integrations.
- A project creates or changes unit, component, contract, repository,
  integration, end-to-end, live-integration, smoke, or production tests.
- A pull-request or deployment workflow decides which correctness checks to
  run or which evidence to retain.

Load this rule before implementation and validation. Apply each test layer
when the repository has that responsibility or boundary. Do not invent a
deployment environment or external integration for a project that has none.

## Rules

### Confidence, Not Certainty

- Treat “near 100% correctness” as a risk-based confidence objective backed by
  comprehensive evidence, not as a mathematical or absolute guarantee.
- Map changed behavior and acceptance criteria to the narrowest tests that can
  prove them, then add broader integration and end-to-end evidence for
  boundary-spanning risks.
- Report skipped, partial, blocked, flaky, and unavailable validation
  literally. Do not convert unobserved behavior into a passing claim.
- Use coverage reports to find untested behavior, not as a substitute for
  meaningful assertions or as a universal percentage target.

### Preserve Code-Level Tests

- Keep unit and other code-level tests in the language and framework's native
  package, module, or test layout. Do not move them into the high-level suite
  directories defined below.
- End-to-end and live-integration tests supplement code-level tests and never
  replace them.
- Add or update the narrowest applicable tests for every behavior change,
  including success, failure, boundary, authorization, concurrency,
  idempotency, retry, and compatibility cases that materially affect risk.
- Unit tests isolate deterministic logic from real networks, clocks,
  randomness, processes, and persistent services unless the framework
  provides a controlled substitute.
- Component, controller, handler, client, contract, schema, and generated-code
  tests verify their owned interface and translation behavior.
- Repository and integration tests exercise real storage, queues, filesystems,
  protocols, or faithful ephemeral fixtures when those boundaries determine
  correctness.
- Keep tests deterministic and independently repeatable. Diagnose and fix
  flaky tests; do not hide them with unconditional retries, sleeps, relaxed
  assertions, or permanent skips.
- During development, run focused tests for fast feedback. Before handoff, run
  the complete applicable code-level suite and record any genuine blocker.

### Pull-Request CI

- GitHub-hosted projects with code must run all applicable code-level checks
  on `pull_request`. This includes formatting, linting, static or type
  analysis, unit, component, contract, repository, integration, schema,
  generated-code, build, and security checks owned by the project.
- Configure ephemeral services or faithful fixtures for integration checks.
  Do not silently downgrade an integration test to a mock-only unit test
  because CI lacks its boundary.
- Make correctness jobs required when repository policy supports required
  checks. Treat pending, skipped, cancelled, unavailable, and unobserved jobs
  as distinct from passing jobs.
- Run hermetic, stable, time-bounded local end-to-end suites in pull-request CI
  when their dependencies and credentials can be isolated safely.
- When a local end-to-end suite cannot run reliably in PR CI, document the
  reason and require fresh local milestone evidence before implementation
  handoff. The manual evidence does not replace mandatory code-level CI.
- Upload CI-generated high-level run directories as workflow artifacts.
  Never commit raw run evidence or let CI automatically edit the tracked
  status map.

### High-Level Suite Layout

- Put high-level executable suites under the repository-root `tests`
  directory, organized first by test type and then by environment:

```text
tests/
  end-to-end/
    local/
    production/
  live-integration/
    local/
    production/
  RUN_STATUS.md
```

- Adopt compatible existing directories instead of duplicating them. If an
  existing path has incompatible semantics, use the collision-free sibling
  `tests/end-to-end-kit` or `tests/live-integration-kit` with the same
  environment children and document the choice in
  `docs/references/testing.md`.
- Keep shared scenarios and assertions environment-neutral when practical.
  Environment wrappers own target URLs, credentials, preflight checks, and
  environment-specific setup or cleanup.
- End-to-end suites exercise a complete externally observable workflow across
  the components that own it. Live-integration suites exercise a narrower
  real external or shared boundary when a complete workflow is not the test's
  purpose.
- Do not relabel a narrow health probe or mocked component test as
  end-to-end.

### Local And Production Execution

- Run the applicable local suite after a coherent set of changes is developed
  and before calling implementation validation complete.
- Local suites use a production-like topology while keeping data and
  dependencies isolated from production.
- Run the applicable production suite only after deployment lands and after
  verifying the exact target environment and deployed version or source
  commit.
- Automate production validation as a post-deployment job when the project can
  do so safely and deterministically. Otherwise document an ordered operator
  command and require its evidence before claiming production validation.
- A failed production suite prevents a “production validated” claim.
  Remediation and rollback follow the project's documented release policy;
  this rule does not invent a universal rollback action.
- A non-deployable project records production suites as `NOT_APPLICABLE`
  rather than creating an artificial environment.

### Immutable Run Evidence

- Store every high-level execution beneath:

```text
tmp/<UTC-date>/<stable-test-id>/<positive-run-number>/
```

- Use a UTC date in `YYYY-MM-DD` form. Use a stable test identifier, normally
  the executable basename including its extension. Use positive integer run
  numbers scoped to the date and test identifier.
- Reserve a run directory without overwriting. Concurrent or repeated runs
  must atomically claim the next available positive integer or fail safely;
  they must never reuse, truncate, or replace an existing run.
- Give each execution a globally useful run ID in the form
  `<YYYYMMDDTHHMMSSZ>-<short-unique-suffix>` and use that ID consistently in
  evidence and synthetic resources.
- Every run directory contains:
  - `output.txt` with combined, readable, redacted execution output;
  - `result.json` with the final machine-readable result;
  - clearly named structured request, response, assertion, and diagnostic
    files needed to explain the run.
- `result.json` records at least:
  - schema version, project, suite, stable test ID, and environment;
  - run ID, run number, start and finish timestamps, and duration;
  - result, exit code, source commit, and deployed version when applicable;
  - target identity, assertion summary, cleanup status, and artifact paths.
- Use only `PASS`, `FAIL`, `PARTIAL`, `BLOCKED`, `SKIPPED`, or
  `NOT_APPLICABLE` as the final result.
- Write a final result even when setup, assertions, or cleanup fail. Preserve
  the original failure and report cleanup failure separately.
- Redact credentials, authorization headers, cookies, private keys, signed URL
  query strings, customer data, and other secrets from every artifact. Prefer
  structured JSON, JSON Lines, Markdown, and plain text with stable field
  names over terminal-only formatting.
- Keep `tmp/` ignored by Git. Subsequent tests may read immutable structured
  artifacts through documented paths, but must validate schema and status
  before trusting them.

### Curated Status Map

- Track `tests/RUN_STATUS.md` as a curated current-state map, not as an
  append-only command ledger.
- Keep exactly one current row per suite and environment with these columns:

```text
| Suite | Environment | Current status | Latest attempt | Latest pass | Source/deployment | Run ID | Evidence | Active finding |
```

- Refresh the tracked map at meaningful validation milestones or when status,
  deployed identity, evidence, or an active finding materially changes.
- Preserve unresolved failures and diagnostically important recent failures in
  `Active finding`. Do not add a new row for every repeated successful run.
- Link or name immutable local evidence and CI artifacts precisely. Do not
  claim that an expired, inaccessible, or unobserved artifact was inspected.
- Local and CI tooling may generate a candidate complete status view under the
  run directory, but must not silently rewrite or commit `RUN_STATUS.md`.

### Safe Production Test Data

- Permit bounded synthetic writes only when a genuine production workflow
  cannot be validated read-only and the project provides safe write isolation.
- Name every synthetic resource:

```text
kit-e2e-<project>-<environment>-<run-id>-<resource>[-<ordinal>]
```

- Keep the leading `kit-e2e-` marker visible. When identifier limits require
  truncation, preserve that marker and store the complete identity in
  supported metadata:
  - `kit_e2e=true`
  - `kit_e2e_project=<project>`
  - `kit_e2e_environment=<environment>`
  - `kit_e2e_run_id=<run-id>`
  - `kit_e2e_created_at=<RFC3339 timestamp>`
  - `kit_e2e_expires_at=<RFC3339 timestamp>` when retention is temporary
- Before a mutating production test, verify the target environment, deployed
  identity, dedicated least-privilege credentials, write authorization, rate
  limit, cost limit, time limit, and cleanup or retention policy.
- Operate only on synthetic resources carrying the execution's exact run ID.
  Cleanup must select both the `kit-e2e-` marker and exact run ID; a broad
  prefix alone is never deletion authority.
- Record every created resource and the cleanup outcome in immutable evidence.
- Never use customer data, reset production state, mutate infrastructure or
  shared configuration, weaken authentication, change unrelated records, or
  broaden credentials to make a test pass.
- If safe write isolation is unavailable, restrict the suite to read-only
  probes and report `PARTIAL`, not complete end-to-end validation.

### Project Testing Reference

- Keep `docs/references/testing.md` current with:
  - language-native code-level commands and their PR workflow/check names;
  - high-level suite inventory and canonical invocation order;
  - local and production environment preflights;
  - credential and secret names without values;
  - synthetic-data, cleanup, retention, rate, cost, and timeout policy;
  - automated jobs, uploaded artifact locations, and manual fallbacks;
  - known gaps, partial coverage, blocked checks, and non-applicable
    environments.
- Keep feature-specific acceptance and observed validation in the active
  `SPEC.md`. The testing reference records reusable project-wide operation,
  not a duplicate history of every feature run.

## Anti-Patterns

- Replacing unit or integration tests with end-to-end scripts.
- Moving language-native tests into `tests/end-to-end` or
  `tests/live-integration`.
- Claiming 100 percent correctness, production validation, or hosted CI success
  from partial or unobserved evidence.
- Running only happy paths or using line coverage as the sole quality signal.
- Hiding flaky tests with retries, long sleeps, weak assertions, or permanent
  skips.
- Overwriting a prior run directory or storing all output in one mutable log.
- Committing raw run artifacts or appending every successful command to
  `RUN_STATUS.md`.
- Capturing secrets, signed URLs, customer data, or unredacted production
  payloads in evidence.
- Mutating production before exact target, deployment, identity, limits, and
  cleanup preflights pass.
- Cleaning by broad prefix, touching unrelated records, changing
  infrastructure, or weakening authentication.
- Calling a read-only smoke probe complete end-to-end coverage.
- Creating fake production suites for libraries or other non-deployable
  projects.

## Verification

- Confirm changed behavior is covered at the narrowest applicable code-level
  layer and broader boundary risks have integration or end-to-end evidence.
- Confirm all applicable code-level checks run in pull-request CI and their
  observed states are reported literally.
- Confirm high-level executables are organized by test type and environment
  without moving language-native tests.
- Confirm repeated and concurrent runs cannot overwrite evidence.
- Inspect `output.txt`, `result.json`, and structured artifacts for readability,
  schema completeness, redaction, source/deployment identity, assertions, and
  cleanup outcome.
- Confirm `RUN_STATUS.md` has one current row per suite and environment,
  preserves active findings, and is not automatically rewritten by CI.
- Confirm CI uploads high-level evidence without committing it.
- For production writes, confirm the exact `kit-e2e-` identity, metadata,
  target, credentials, limits, created-resource inventory, and exact-run
  cleanup proof.
- Confirm unavailable safe production writes produce read-only `PARTIAL`
  evidence and non-deployable projects use `NOT_APPLICABLE`.
- Run the project commands documented in `docs/references/testing.md` and
  record any skipped or blocked validation.

## Examples

Immutable LabCore-style run evidence:

```text
tmp/2026-07-27/labcore-biotouch-kit-order-prod-test.sh/1/
  output.txt
  result.json
  create-order-request.json
  create-order-response.json
  assertions.jsonl
```

Synthetic production order:

```text
kit-e2e-labcore-production-20260727T153012Z-a1b2c3-order-001
```

Curated status row:

```text
| order lifecycle | production | PASS | 2026-07-27T15:36:44Z | 2026-07-27T15:36:44Z | v1.4.2 / abc123 | 20260727T153012Z-a1b2c3 | Actions artifact order-production-1 | none |
```
