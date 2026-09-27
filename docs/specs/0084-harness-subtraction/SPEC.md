---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0084"
  slug: "harness-subtraction"
  dir: "0084-harness-subtraction"
relationships:
  - type: builds_on
    target: 0082-universal-agent-contract
    note: Keeps the universal contract and removes the machinery around it.
---
# SPEC

## PURPOSE

Remove Kit machinery that no longer earns its complexity for current frontier coding agents, and define the canonical project structure that the final migration phase will converge existing projects onto (GH-219, modernization phase 3).

## CONTEXT

- After phase 2, Kit still shipped prompt wrappers (`dispatch`, `pr fix`, `pr orchestrate`), a context router (`context resolve` plus seven workflow manifests), a hand-maintained capability catalog, 17 frozen instruction versions, a self-improvement harness, host adapters, a generated progress summary, and a rule corpus of about 36k words fetched at runtime from mutable `main`.
- Usage data (Aug–Sep 2026, released binaries): 98.8% of calls were non-interactive agents; `context resolve` failed 30% of the time; `pr fix` had 1 call and `dispatch` 6 in the final two weeks; `pr orchestrate` 3 calls ever; `improve run` 8; `instructions` 16. High counts for `context resolve` and `capabilities` reflected instructions that required them, not demonstrated value.
- Every push to `main` already cuts a release, so fetching rules from `main` bought no speed and let a binary install rules it was never tested with.
- The 300-line source cap (spec 0054) was a Kit maintainability choice encoded as a universal rule for every project and language.

## REQUIREMENTS

- Every remaining command does deterministic work agents cannot reproduce as safely or cheaply.
- A released binary installs exactly the rules it shipped with; no runtime network dependency for rules.
- Mandatory pre-implementation context does not regress from phase 2 (about 2.5k tokens for a fresh project).
- Critical invariants from phase 2 remain intact; protected developer-experience files do not change.
- Legacy state stays recognizable for the final migration without keeping legacy runtime behavior.

## ACCEPTED PLAN

1. Delete prompt-producing and orchestration commands, context resolution and workflow manifests, the capability catalog and its rules, instruction archives, the improve harness, host adapters, the progress summary, and `rules link`.
2. Embed rules in the binary and replace section merging and base fetching with hash-based sync against the embedded version.
3. Consolidate and condense the rule corpus; make style rules optional.
4. Replace the universal 300-line rule with an opt-in `.kit.yaml` `source_file_line_limit` enforced by `kit check --project`.
5. Reduce v3 generated project memory to the Constitution and `docs/references/testing.md`.
6. Isolate usage telemetry from tests, development builds, and help lookups.
7. Resolve the duplicate `0077` spec number and make Kit's own project check pass in CI.

Single-lane execution: removals, rule changes, and test updates are tightly coupled.

## DECISIONS

- Kept: `init`, `spec`, `rules add|list|view`, `check`, `reconcile`, `health`, `registry status`, `status`, `config check`, `aws verify`, `usage *`, `upgrade`, `version`, `completion`. Each performs deterministic bootstrap, validation, convergence, allocation, identity verification, or evidence work.
- `registry status` stays because it is the cheap read-only drift check that scheduled maintenance calls and `health` builds on.
- `kit init`'s initialization prompt stays because the protected Makefile starter points users to it.
- Removed `kit pr fix` despite interactive use: its unique parts were PR-head worktree preparation and thread lookup, which agents perform directly with `git` and `gh` under the `delivery` rule; usage had fallen to near zero. `internal/worktreeprep` keeps only checkout inspection, used to keep `reconcile` writes out of the primary checkout.
- Merged `work-lane-gating`, `github-pr-delivery`, `safety-guardrails`, and `human-authorship` into one `delivery` rule; deleted `agent-completion-output` (its invariants live in the contract) and `source-file-size`. Retired slugs are listed in `pkg/cli/rules_retired.go` so historical spec references stay valid and the final migration can recognize legacy files.
- Rule sync no longer needs a remote base: identical files are managed, files unchanged since Kit wrote them take the embedded version, and edited files are preserved as local-custom. The section-merge "conflict" state is gone.
- `readme-header-tagline` and `llms-txt` use a new `optional` registry scope and install only through `kit rules add`. The `kit-maintainer` scope was removed with `command-capabilities`.
- Rule documents require only Purpose, Applies When, and Rules sections, removing the pressure to restate rules as anti-pattern and verification lists.
- The line-limit audit runs only when configured. Kit configures 300 for itself and enforces it in CI through `kit check --project`.
- Renumbered `0077-auto-assign-initiator` (created a day after `0077-unstructured-completion-output`) to `0083`. Historical validation logs that mention the duplicate were left literal.
- Removed legacy config fields (`goal_percentage`, `skills_dir`, `loop`, `prompts`, `feature_state`, `removed_features`, `project_refresh`) and registry provenance fields; decoding ignores them in existing files.
- Superseded: live registry fetching and three-way section merge (0021 registry), `context resolve` (0059), capability catalog (0033), instruction versions (0044), improve harness (0038), prompt wrappers (0008, 0034, 0060), the progress summary rollup, and the universal 300-line invariant (0054).

## DISCOVERIES

- `kit check --project` counted warnings as blocking, and the stale-reference and constraining-reference policies contradicted each other for a retired constraint. Stale references are now exempt from the constraining check.
- Tests for the usage report recorded real events; test binaries now never record, so tests seed the store directly.
- Kit's own `.kit.yaml` still listed retired rules after refresh because refresh does not retire rules; they were pruned by hand here, and retirement is final-migration work.
- Independent pre-merge review (the PR exceeded CodeRabbit's file limit) found that earlier releases recorded section-merged rules as `managed` with the hash of content that still held local edits, and marked them with per-section `sections` state. The hash-based sync would have overwritten those edits, so `sections` stays decodable and such entries, like legacy `conflict` entries, are treated as local-custom.
- The same review found `kit rules add <slug>` wrote a blank stub for shipped optional rules; it now installs the embedded rule as managed and names the replacement for a retired slug. Usage recording now uses the resolved build version, so `go install` binaries still record; only unversioned builds are isolated, and `make build` stamps a tag, so local builds do record unless `KIT_USAGE_DISABLED=1`.
- Legacy generated support documents and their templates are gone from the source tree; the final migration recovers their known content from release tags rather than keeping templates alive.

## VALIDATION

- PASS: `gofmt -l .` (no output), `go vet ./...`, `go test ./... -count=1`, `golangci-lint run ./...` (0 issues), `make build`, `git diff --check`, `kit check --project` on Kit itself, and `actionlint .github/workflows/ci.yml`.
- PRE-EXISTING: `actionlint` rejects the `concurrency.queue` key in the two release workflows, which this change did not touch.
- Measurements (words; tokens ≈ words × 4/3; fresh project in an empty repository):

| Metric | Before | After |
| --- | ---: | ---: |
| CLI command tree entries | 36 | 25 |
| Prompt-producing commands | 7 | 3 |
| Kit-owned generated files on fresh init | 40 | 19 |
| Kit-owned generated documentation words | 43,408 | 10,013 |
| Always-loaded contract words | 1,107 | 1,079 |
| Mandatory implementation-task tokens | ~2,517 (4 files) | ~2,177 (3 files) |
| Default contextual rules / words | 22 / 36,104 | 13 / 6,219 |
| Legacy instruction archive files | 17 | 0 |
| Remote rule dependencies | 1 (GitHub API) | 0 |

## OUTCOME

Delivered as planned. Remaining legacy state for the final migration: pre-contract entry files and scaffold versions 1 and 2, retired support documents and rule files in existing projects, orphaned registry entries, leftover `docs/references/workflows/`, `docs/PROJECT_PROGRESS_SUMMARY.md`, and legacy staged spec artifacts.

## REPOSITORY MEMORY

- Created this spec; renumbered spec `0083-auto-assign-initiator`.
- Updated `docs/CONSTITUTION.md` (command surface, rules distribution, line limit, telemetry), `docs/commands.md`, `docs/overview.md`, `docs/README.md`, `docs/references/README.md`, `docs/references/testing.md`, `docs/migration-v3.md`, and `README.md`.
- Deleted `docs/PROJECT_PROGRESS_SUMMARY.md`, `docs/workflows.md`, `docs/agents/README.md`, `docs/references/workflows/`, `docs/references/host-adapters/`, `docs/references/worktrees.md`, `docs/references/tooling.md`, `docs/references/external-systems.md`, and `docs/evals/`.
