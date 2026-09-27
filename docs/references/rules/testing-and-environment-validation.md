---
kind: ruleset
slug: testing-and-environment-validation
description: Keeps code-level tests primary, validation honest, and browser and production testing safe.
status: active
registry_scope: downstream
applies_to:
  - testing
  - validation
  - ci
  - browser-automation
  - production
read_policy_default: conditional
---

# Ruleset: testing-and-environment-validation

## Purpose

- Base completion claims on tests that actually ran against the changed behavior.
- Keep browser and production validation from harming the user's machine, data, or customers.

## Applies When

- Changing behavior, adding or running tests, configuring CI, or running end-to-end, browser, or production validation.

## Rules

### Code-Level Tests

- Keep unit, integration, and contract tests in the language's native framework; they stay primary. End-to-end and live suites supplement them and never replace them.
- Add or update the narrowest tests that prove each behavior change. Keep tests deterministic; fix flakiness instead of retrying it away.
- Before handoff, run the project's validation commands from `docs/references/testing.md` and keep that file current when commands change.
- Projects with code on GitHub should run their code-level checks on every pull request. Do not silently downgrade an integration test to a mock because CI lacks a dependency.

### Honest Results

- Report skipped, partial, flaky, blocked, and unavailable validation as exactly that. Coverage is a guide, not proof.
- Keep generated run evidence out of Git (for example under an ignored `tmp/`), and redact secrets from any evidence.

### Browser Automation

- Never drive the user's active browser profile or personal sessions unless the user explicitly asks. Prefer the host's built-in or an isolated browser.
- Own and name each browser session you start, close task-owned sessions and automation processes in teardown, and never terminate processes you did not start.

### Production Validation

- Run production suites only after the deployment they validate has landed and only with authorization for that environment. Verify the target environment and deployed version first.
- Write only clearly named synthetic data scoped to the run, never touch customer data, and clean up under `deletion-safety`. When safe write isolation is unavailable, validate read-only.
- A failed or unrun production suite means production is not validated.
