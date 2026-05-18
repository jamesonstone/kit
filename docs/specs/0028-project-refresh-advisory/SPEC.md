---
kit_metadata_version: 1
artifact: "spec"
feature:
  id: "0028"
  slug: "project-refresh-advisory"
  dir: "0028-project-refresh-advisory"
relationships:
  - type: "builds_on"
    target: "0017-reconcile-command"
  - type: "builds_on"
    target: "0025-v0-prompt-library"
  - type: "related_to"
    target: "0027-implement-readiness-gate"
references:
  - name: constitution contract
    type: doc
    target: docs/CONSTITUTION.md
    relation: informs
    read_policy: conditional
    used_for: project-level source-of-truth semantics and advisory-gate constraints
    status: active
  - name: init project spec
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: uses
    read_policy: conditional
    used_for: shipped project-initialization and prompt-library behavior
    status: active
  - name: reconcile command
    type: code
    target: pkg/cli/reconcile.go, pkg/cli/reconcile_prompt.go
    relation: implements
    read_policy: conditional
    used_for: structural drift audit precedent and project-wide documentation prompt behavior
    status: active
  - name: prompt library
    type: code
    target: pkg/cli/prompt_builtin_kit.go, pkg/cli/prompt_builtin_render.go
    relation: implements
    read_policy: conditional
    used_for: built-in prompt registration and runtime rendering
    status: active
  - name: reflect command
    type: code
    target: pkg/cli/reflect.go
    relation: implements
    read_policy: conditional
    used_for: late workflow advisory gate
    status: active
  - name: complete command
    type: code
    target: pkg/cli/complete.go
    relation: implements
    read_policy: conditional
    used_for: post-completion advisory output
    status: active
  - name: README
    type: doc
    target: README.md
    relation: guides
    read_policy: conditional
    used_for: user-facing command guidance
    status: active
---
# SPEC

## SUMMARY

- Add a built-in `kit prompt project refresh` prompt that asks a coding agent to re-analyze a maturing repository and update durable project-level docs.
- Add a soft project-refresh advisory gate to late workflow surfaces so users are reminded to refresh project truth without blocking normal feature progress.

## PROBLEM

- `kit init` can run before the repository has enough real code, commands, or workflow history for `docs/CONSTITUTION.md` to capture durable project truth.
- As the first feature matures, the constitution can remain structurally valid while becoming semantically stale.
- `kit reconcile --all` catches contract drift, but it is not designed to ask an agent to re-evaluate newly emerged project-wide rules, vocabulary, conventions, and long-term constraints.

## GOALS

- Expose `kit prompt project refresh` as the explicit manual command for refreshing project-level documentation.
- Keep the refresh prompt documentation-scoped and prompt-only.
- Ask the agent to inspect current repository state, feature history, command behavior, and existing Kit-managed docs before editing.
- Update `docs/CONSTITUTION.md` only with durable project-wide rules, conventions, vocabulary, and constraints.
- Tell the agent to use `kit reconcile --all` for structural contract drift and `kit rollup` when project summary content changes.
- Add advisory wording to `kit reflect` and `kit complete` that points to `kit prompt project refresh` when late feature work may have revealed durable project-level changes.
- Keep the advisory soft: no hard failure, no new persisted state, and no new lifecycle phase.

## NON-GOALS

- Re-running `kit init` as a destructive or automatic project migration.
- Adding a new top-level `kit project` command in this feature.
- Automatically editing `docs/CONSTITUTION.md`.
- Adding freshness timestamps, marker files, lock files, or hidden project state.
- Making `kit reflect` or `kit complete` fail when project docs might be stale.
- Replacing `kit reconcile --all`.

## USERS

- Users who initialize a repository before its real project shape has emerged.
- Maintainers who want a deliberate way to refresh `CONSTITUTION.md` after early feature work.
- Coding agents that need a bounded prompt for semantic project-doc refresh without drifting into product code.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: 0017-reconcile-command
- builds on: 0025-v0-prompt-library
- related to: 0027-implement-readiness-gate

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | project-level source-of-truth semantics and advisory-gate constraints | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | shipped project-initialization and prompt-library behavior | active |
| reconcile command | code | `pkg/cli/reconcile.go, pkg/cli/reconcile_prompt.go` | structural drift audit precedent and project-wide documentation prompt behavior | active |
| prompt library | code | `pkg/cli/prompt_builtin_kit.go, pkg/cli/prompt_builtin_render.go` | built-in prompt registration and runtime rendering | active |
| reflect command | code | `pkg/cli/reflect.go` | late workflow advisory gate | active |
| complete command | code | `pkg/cli/complete.go` | post-completion advisory output | active |
| README | doc | `README.md` | user-facing command guidance | active |

## REQUIREMENTS

- Add built-in prompt identity `project refresh` to the prompt library.
- `kit prompt project refresh` must require a Kit project context.
- The rendered prompt must begin with `/plan`.
- The prompt must be docs-only and must explicitly forbid product code, runtime, and test changes unless the user separately asks for them.
- The prompt must ask the agent to inspect:
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
  - repository instruction docs
  - current feature docs under `docs/specs/`
  - current code, command, and package structure
  - recent git status and diffs
- The prompt must distinguish semantic project refresh from structural reconciliation.
- The prompt must direct structural drift to `kit reconcile --all`.
- The prompt must direct generated rollup drift to `kit rollup`.
- The prompt must update `docs/CONSTITUTION.md` only when durable project-level truth changed.
- The prompt may update repository instruction docs only when canonical routing or workflow guidance changed.
- The prompt must preserve existing project wording when it remains accurate.
- The prompt must require verification with `kit check --project`, and with `kit check --all` when feature docs or instruction files are touched.
- The prompt must use a concise fixed response contract with `Findings`, `Updates`, and `Verification`.
- `kit reflect` must include a soft project-refresh advisory gate in the reflection prompt.
- The reflection advisory must ask the agent to decide whether current work revealed durable project-wide rules, constraints, vocabulary, or workflow changes.
- `kit complete` must print a non-blocking advisory after successful completion.
- Advisory text must point to `kit prompt project refresh`.
- The advisory must not change lifecycle state, block the command, or require additional flags.
- README and the core project spec must describe the manual refresh prompt and advisory behavior.

## ACCEPTANCE

- `kit prompt list` includes `project refresh`.
- `kit prompt project refresh --output-only` emits a docs-only project refresh prompt that starts with `/plan`.
- The prompt tells the agent to refresh `docs/CONSTITUTION.md` only for durable project-level changes.
- The prompt tells the agent to use `kit reconcile --all` for structural contract drift.
- The prompt tells the agent to verify with `kit check --project`.
- `kit reflect` prompt output includes a soft project-refresh advisory gate.
- `kit complete` prints a non-blocking advisory that points to `kit prompt project refresh`.
- Existing `kit init` behavior remains unchanged.
- Tests cover the new built-in prompt, reflection advisory, and completion advisory.

## EDGE-CASES

- The project has no feature docs yet; the refresh prompt should still ask the agent to inspect available repo structure and update only if durable facts exist.
- `docs/CONSTITUTION.md` is structurally valid but semantically stale.
- `kit reconcile --all` reports structural drift; the refresh prompt should route that work through reconciliation instead of duplicating the audit.
- A feature changed implementation details but no project-level rules; the advisory should allow "no project refresh needed."
- `kit complete --all` completes multiple features; the advisory should print once after the batch succeeds.

## OPEN-QUESTIONS

- none
