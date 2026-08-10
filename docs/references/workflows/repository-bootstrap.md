---
kind: workflow
slug: repository-bootstrap
description: Populates deterministic Kit starters from verified repository evidence using progressive context loading.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - bootstrap
  - repository-memory
  - rlm
  - tooling
  - testing
read_policy_default: conditional
dependencies:
  - ruleset/safety-guardrails
  - ruleset/testing-and-environment-validation
  - ruleset/source-file-size
  - ruleset/constitution-curation
  - ruleset/readme-header-tagline
---

# Workflow: Repository Bootstrap

## Purpose

Turn Kit's deterministic repository starters into evidence-backed project
memory and verified command entrypoints without making Kit infer project truth
or supervise a coding agent.

## Context Loading

1. Start with the resolved routing entrypoints and `docs/agents/RLM.md`.
2. Read this workflow, every dependency, and the project reference indices.
3. Inspect only relevant manifests, package metadata, build scripts, CI, tests,
   documentation, code boundaries, current specifications, history, and
   external-system evidence.
4. Expand context only when an unresolved bootstrap decision requires it.

Do not ingest the entire documentation tree, all registry artifacts, secret
files, generated caches, transcripts, or unrelated historical specifications.

## Workflow

1. Record which bootstrap files Kit created, boundedly merged, or preserved.
2. Establish the repository's demonstrated identity, implementation boundaries,
   toolchain, native commands, validation layers, and recurring integrations.
3. Curate only durable project-wide truth into `docs/CONSTITUTION.md`.
4. Populate `docs/PROJECT_PROGRESS_SUMMARY.md` from current specifications and
   repository history. Retain its valid empty state when no current feature
   evidence exists; never invent features.
5. Replace placeholder rows in `docs/references/testing.md` with actual safe,
   verified commands, environments, checks, and known gaps.
6. Populate `docs/references/tooling.md` from package, toolchain, build, and
   configuration evidence. Populate `docs/references/external-systems.md` only
   with verified integrations, never guessed secrets or production state.
7. Add Makefile targets only as thin wrappers around proven repository-native
   commands. Leave the help-only starter unchanged when none are proven.
8. Review README identity and only the bounded Kit badge and Maintainers
   sections without replacing surrounding project prose.
9. Validate every safe populated command and review the complete bootstrap
   diff under the selected safety, testing, source-size, README, and
   Constitution-curation rules.

## Gates

- `.env` values are never read, copied, summarized, or exposed.
- Existing `.envrc` is preserved and no trust or allow command is executed.
- Credentials, endpoints, owners, environments, and deployment state require
  direct repository or external-system evidence; absence remains an evidence
  gap.
- Starter placeholders are not project truth and cannot justify Constitution,
  tooling, testing, integration, or progress claims.

## Completion

Report files changed and preserved, validation performed, unresolved evidence
gaps, and the Repository Memory decision with rationale and artifacts. Every
project-specific claim points to repository evidence, and safe starters remain
valid wherever evidence was insufficient.
