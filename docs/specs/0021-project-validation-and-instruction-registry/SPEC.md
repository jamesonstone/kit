# SPEC

## SUMMARY

- Centralize Kit's instruction-model metadata in one internal registry and add a project-scoped validation mode so repo-level contract drift is checked mechanically instead of being spread across prompt builders.

## PROBLEM

- The current `v1` versus `v2` instruction contract is duplicated across templates, prompt helpers, map output, and version-detection helpers.
- That duplication increases correctness risk because one command can learn a new repo doc or routing rule while another still uses stale assumptions.
- `kit check` only validates feature-scoped docs today, so repo-level instruction drift and thin-ToC contract breakage are mostly surfaced through `kit reconcile` prompts instead of a direct validator.
- Subagent guidance exists, but the shipped contract should distinguish RLM discovery from dispatch-style execution more explicitly.

## GOALS

- Introduce a single internal instruction-contract registry for version detection and repo-doc metadata.
- Make current consumers read instruction metadata from that shared registry instead of duplicating path lists and labels.
- Extend `kit check` with a project-scoped validation mode that checks repo-level instruction coherence in addition to feature docs.
- Tighten shipped subagent guidance so RLM discovery and dispatch/parallel execution are clearly separated.

## NON-GOALS

- Replacing `kit reconcile` as the prompt-oriented migration surface.
- Introducing a new long-lived project database or hidden cache.
- Rewriting every prompt builder into a typed prompt IR in this change.
- Launching subagents directly from the Kit binary.

## USERS

- Maintainers who need repo-level validation to fail fast when the instruction contract drifts.
- Coding agents that rely on a stable thin-ToC docs tree and consistent subagent guidance.
- Contributors updating prompt surfaces who need one source of truth for instruction-model metadata.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: `0012-default-subagent-orchestration`
- builds on: `0017-reconcile-command`
- builds on: `0020-versioned-instruction-model`
- related to: `0006-skill-mine-command`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| versioned instruction model | doc | `docs/specs/0020-versioned-instruction-model/SPEC.md` | current `v1` and `v2` contract that must move into a shared registry | active |
| subagent orchestration contract | doc | `docs/specs/0012-default-subagent-orchestration/SPEC.md` | preserve subagent-first defaults while clarifying RLM versus dispatch responsibilities | active |
| reconcile audits | code | `pkg/cli/reconcile_audit.go` | existing repo-level drift detection that project validation can reuse | active |
| check command | code | `pkg/cli/check.go` | existing validation surface to extend with project-scoped validation | active |
| instruction templates | code | `internal/templates/instruction_templates.go`, `internal/templates/instruction_templates_v2.go` | current instruction artifact content and support-file wiring | active |

## REQUIREMENTS

- [SPEC-01] Kit must expose a shared internal instruction-contract registry package that owns:
  - supported instruction-model versions
  - repo instruction artifact paths
  - repo-local knowledge-doc paths
  - labels and usage text for those docs
  - instruction-model detection rules
- [SPEC-02] Prompt builders, repo-doc helpers, and map output must consume the shared instruction-contract registry instead of duplicating hardcoded `v1` or `v2` path sets.
- [SPEC-03] `kit check` must accept `--project` to validate the repo-level contract.
- [SPEC-04] `kit check --project` must reject a feature argument and must not require `--all`.
- [SPEC-05] `kit check --project` must validate:
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
  - every feature under `docs/specs/`
  - Kit-managed instruction files
  - version-specific repo-local docs-tree artifacts
- [SPEC-06] `kit check --project` must reuse the current repo-audit engine or an equivalent shared audit layer instead of introducing a second drift implementation.
- [SPEC-07] `kit check --project` must return a non-zero error when any repo-level validation finding exists, even if the finding is surfaced as a warning in reconciliation output.
- [SPEC-08] `kit check --project` terminal output must stay concise and must show each finding with severity, file path, and issue text.
- [SPEC-09] The shipped subagent guidance must explicitly distinguish:
  - RLM as repository-scale discovery and context selection
  - dispatch/subagent orchestration as post-discovery execution planning
- [SPEC-10] The shipped subagent guidance must keep the main agent responsible for synthesis, integration, validation, and communication.

## ACCEPTANCE

- One internal package defines the instruction-model metadata used by detection, repo-doc routing, and map output.
- `kit check --project` fails when repo-level docs or instruction artifacts drift from the current Kit contract.
- `kit check --project` catches missing `v2` support docs in a thin-ToC repo and catches leftover `v2` support docs in a `v1` repo.
- Shared subagent guidance and repo-local docs describe RLM as discovery and dispatch as execution planning rather than conflating them.
- Existing feature-scoped `kit check <feature>` and `kit check --all` behavior remains intact.

## EDGE-CASES

- `.kit.yaml` is missing or contains no persisted instruction version, so detection must fall back safely.
- A repo uses `v1` and still has leftover `v2` docs-tree files.
- A repo uses `v2` but is missing one or more support files under `docs/agents/` or `docs/references/`.
- The project has feature-doc warnings that should still fail `--project` validation because the repo contract is not fully coherent.

## OPEN-QUESTIONS

- none
