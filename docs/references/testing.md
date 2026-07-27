# Testing Reference

## Purpose

- Record the project's durable commands, suites, environments, automation, and evidence expectations
- Follow `rules/testing-and-environment-validation.md` for the mandatory cross-project testing and production-safety contract
- Keep feature-specific testing details in the current feature's `SPEC.md` VALIDATION and OUTCOME sections; legacy staged flows may still use `PLAN.md` or `TASKS.md`

## Code-Level Validation

| Layer | Command | PR workflow or check | Required | Notes |
| --- | --- | --- | --- | --- |
| Project-specific | Document the canonical command | Document the GitHub Actions job | yes or no | Record fixtures, services, and scope |

## High-Level Suites

| Suite | Type | Environment | Command | Automation | Evidence |
| --- | --- | --- | --- | --- | --- |
| Project-specific | end-to-end or live-integration | local or production | Document the ordered invocation | PR, post-deploy, or manual fallback | `tmp/<UTC-date>/<test>/<run-number>/` |

## Environment Preflights

- Document exact local topology, target identity, deployed-version checks, dependencies, and timeouts
- Document which environments are not applicable instead of creating artificial suites

## Credentials And Test Data

- List credential and secret names without values
- Document synthetic-data naming, rate and cost limits, cleanup, and retention

## Evidence And Retention

- Keep `tmp/` ignored and record CI artifact locations and retention
- Keep `tests/RUN_STATUS.md` curated at meaningful validation milestones

## Automation And Fallbacks

- Map code-level checks to pull-request jobs and high-level suites to PR or post-deployment jobs when feasible
- Document ordered operator commands when safe automation is unavailable

## Known Gaps

- Record partial, blocked, skipped, and unavailable validation literally
