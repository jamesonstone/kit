---
kind: ruleset
slug: source-file-size
description: Defines the 300-line handwritten source and test file limit, audit scope, exclusions, semantic splitting, and verification.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - implementation
  - testing
  - validation
  - refactor
  - reconcile
  - maintenance
read_policy_default: must
---

# Ruleset: Source File Size

## Purpose

- Keep handwritten implementation and test files small enough for clear
  ownership, review, maintenance, and agent context loading.
- Turn the historical approximate guideline into one observable contract that
  new work and periodic reconcile maintenance can verify consistently.
- Preserve behavior and architecture while splitting by responsibility.

## Applies When

- An agent creates, edits, reviews, validates, or delivers handwritten
  implementation/source or test files.
- `kit reconcile --all` or scheduled Kit maintenance audits a whole project.
- A project adopts or refreshes Kit-managed instructions.

Load this rule before editing implementation/source or test files. For normal
implementation, audit the complete affected source/test scope before delivery.
For whole-project reconcile and scheduled maintenance, audit the complete
version-control-eligible repository scope.

## Rules

### Exact Limit

- Every in-scope file must contain at most 300 physical lines.
- Count the final physical line even when the file has no trailing newline.
- Treat 301 or more lines as an actionable reconcile finding.
- If a safe split cannot be completed, keep the finding visible and report the
  exact file and blocker. Do not silently accept or suppress it.

### In-Scope Files

- Include cached and untracked non-ignored handwritten implementation/source
  and test files that are eligible for version control.
- Include common programming languages, executable scripts, styles, UI
  templates, schemas, queries, and source-like test fixtures.
- Include tests under the same limit as production implementation.
- In a Git repository, use the equivalent of
  `git ls-files --cached --others --exclude-standard` as the candidate set.

### Exclusions

- Exclude documentation files, every file under `docs/**` and `.kit/**`, and
  `.kit.yaml`.
- Exclude ignored files and untracked files that are not version-control
  eligible.
- Exclude vendored dependency trees such as `vendor/**`, `third_party/**`, and
  `node_modules/**`.
- Exclude generated files only when a recognized generated marker or generated
  artifact name establishes that they are not handwritten.
- Do not infer an exclusion from file size, unfamiliar syntax, or an ambiguous
  directory name.

### Semantic Splitting

- Split by cohesive responsibility, domain concept, layer, scenario, or test
  concern.
- Preserve stable public entry points and existing behavior unless the task
  independently requires an interface or behavior change.
- Use responsibility-based filenames. Do not create arbitrary `part1`,
  `part2`, or other numbered shards as the final organization.
- Keep test-only helpers in language-native test files and preserve discovery
  conventions such as Go's `_test.go` suffix.
- Remove obsolete or duplicate code exposed by the split when doing so is
  behavior-preserving and within scope.
- Do not satisfy the line limit through minification, compressed formatting,
  removed useful whitespace, long generated expressions, or moving code into
  documentation or configuration.

### Reconcile And Scheduled Maintenance

- `kit reconcile` identifies exact violations and produces coding-agent
  guidance; it does not rewrite product source files itself.
- A reconcile repair may edit only listed oversized source/test files, directly
  required tests, and directly required canonical documentation.
- Scheduled Kit maintenance may apply a responsibility-preserving split only
  when the current whole-project reconcile audit identifies the file.
- A scheduled split must not change product behavior, dependencies, public
  interfaces, runtime configuration, deployment state, or production data.
- `kit health` remains limited to Kit-managed files and project validation; it
  does not mutate product source.

### Validation Requirements

- Rerun the whole-project source-file audit after maintenance. It must report
  no in-scope file above 300 lines.
- Run formatting, language-native tests, static analysis, builds, generated
  checks, and other project validation applicable to every moved symbol and
  affected boundary.
- Review the complete diff for accidental behavior, visibility, ordering,
  initialization, registration, fixture-discovery, and build-inclusion changes.
- Report skipped, blocked, unavailable, or failing checks literally.

## Anti-Patterns

- Treating “approximately 300” as permission to skip the audit.
- Auditing only the files remembered by the current agent during whole-project
  reconcile maintenance.
- Splitting implementation without moving its focused tests or preserving test
  discovery.
- Accepting arbitrary numbered files as permanent architecture.
- Changing behavior or interfaces to make a mechanical split easier.
- Claiming a clean source-file-size audit when candidate enumeration failed.

## Verification

- Confirm the candidate set includes cached and untracked non-ignored files.
- Confirm documentation, Kit state, vendored trees, ignored files, and proven
  generated files are excluded.
- Confirm every in-scope file is at most 300 physical lines.
- Confirm filenames express responsibility and public entry points remain
  stable.
- Confirm applicable formatting, tests, static analysis, and builds passed.
- Confirm the final reconcile audit is clean or reports an exact blocker.

## Examples

Semantic split:

```text
order_service.go (418 lines)
  -> order_service.go (public entry points and orchestration)
  -> order_validation.go (validation rules)
  -> order_persistence.go (persistence translation)
```

Excluded generated file:

```text
api.pb.go begins with "Code generated ... DO NOT EDIT."
```
