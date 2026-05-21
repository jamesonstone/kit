---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0031"
  slug: executable-verification-harness
  dir: 0031-executable-verification-harness
summary: Add executable verification, local run evidence, generated agent state, trace/replay, and small harness evals while preserving Markdown as Kit's human source of truth.
relationships:
  - type: builds_on
    target: 0030-reference-graph-routing
  - type: builds_on
    target: 0027-implement-readiness-gate
  - type: related_to
    target: 0021-project-validation-and-instruction-registry
  - type: related_to
    target: 0028-project-refresh-advisory
references:
  - id: code-as-agent-harness-paper
    name: Code as Agent Harness paper
    type: paper
    target: https://arxiv.org/abs/2605.18747
    selector_type: url
    selector: https://arxiv.org/abs/2605.18747
    relation: informs
    read_policy: evidence
    used_for: harness-engineering rationale around executable verification, traces, shared state, and multi-agent coordination
    status: active
  - id: kit-rlm-routing
    name: Kit RLM routing guidance
    type: doc
    target: docs/agents/RLM.md
    selector_type: artifact
    selector: RLM.md
    relation: constrains
    read_policy: must
    used_for: just-in-time context loading and pointer-only reference behavior
    status: active
  - id: kit-state-constitution
    name: Kit state constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: "4. Minimal Magic, Explicit State"
    relation: constrains
    read_policy: must
    used_for: boundary between Markdown authority, generated JSON, and local evidence artifacts
    status: active
  - id: next-gen-runtime-notes
    name: V1 next-gen runtime notes
    type: doc
    target: docs/future/V1_NEXT_GEN.md
    selector_type: artifact
    selector: V1_NEXT_GEN.md
    relation: informs
    read_policy: conditional
    used_for: prior non-binding trace, replay, runtime state, and evaluation ideas
    status: active
  - id: current-check-command
    name: Current check command
    type: code
    target: pkg/cli/check.go
    selector_type: symbol
    selector: checkFeature
    relation: constrains
    read_policy: must
    used_for: existing structural validation boundary
    status: active
  - id: current-reflect-command
    name: Current reflect command
    type: code
    target: pkg/cli/reflect.go
    selector_type: symbol
    selector: buildReflectPrompt
    relation: constrains
    read_policy: must
    used_for: existing evidence-oriented reflection prompt behavior
    status: active
  - id: current-task-template
    name: Current task template
    type: code
    target: internal/templates/templates.go
    selector_type: symbol
    selector: Tasks
    relation: constrains
    read_policy: must
    used_for: TASKS.md schema and template migration
    status: active
---
# SPEC

## SUMMARY

Add a verification-first execution layer to Kit: declared checks in feature docs, `kit verify` execution, local run artifacts, evidence-aware reflection, trace/replay commands, generated agent-readable state, machine-readable task bundles, and small harness evals. Markdown remains the authoritative human source of truth; JSON and run artifacts are generated evidence surfaces.

## PROBLEM

Kit currently enforces strong document structure through `kit check`, readiness gates, and prompt instructions, but most correctness evidence still lives outside the harness in terminal output or agent summaries. That makes reflection weaker than it should be: agents can say tests passed without Kit having a structured record, future sessions cannot reliably inspect what was run, and multi-agent work still depends on humans or agents re-parsing Markdown tables instead of consuming a typed task/evidence surface.

## GOALS

- Make executable verification first-class without replacing Kit's document-first workflow.
- Add declared task and feature verification commands that Kit can run and report as structured JSON.
- Persist local, inspectable run artifacts under a visible `.kit/runs/` directory.
- Make `kit reflect` consume verification evidence instead of relying only on prompt instructions.
- Extend `TASKS.md` with executable but still human-readable fields: `VERIFY`, `EXPECTED FILES`, `RISK`, and optional `ROLLBACK`.
- Add trace and replay commands for deterministic-ish reconstruction of what commands ran and what artifacts were produced.
- Add generated `.kit/state.json` as an agent/tool cache derived from canonical docs and latest run evidence.
- Add machine-readable task bundles for safe multi-agent handoff.
- Add `kit eval` fixture checks for Kit's own harness behavior.
- Preserve RLM: commands should read current task/context first, follow explicit references, and avoid inlining whole documents when pointers are sufficient.

## NON-GOALS

- Do not replace `SPEC.md`, `PLAN.md`, or `TASKS.md` with JSON.
- Do not introduce a daemon, live supervisor, remote model dependency, telemetry, or hosted service.
- Do not attempt to replay hidden model reasoning or private agent state.
- Do not make existing legacy feature docs fail only because they lack the new task fields.
- Do not run arbitrary shell strings with unrestricted shell evaluation by default.
- Do not store secrets, environment variables, or unbounded raw terminal output in run artifacts.
- Do not make multi-agent orchestration automatic; task bundles are contracts, not subagent launchers.

## USERS

- Maintainers who need evidence that a feature's declared acceptance criteria were actually verified.
- Coding agents that need compact, typed state and task bundles without loading every document into context.
- Reviewers who need to inspect the latest run, changed files, and verification failures before accepting a feature.
- Future Kit maintainers who need replayable harness evals to prevent regressions in check/implement/reflect behavior.

## SKILLS

Skills are tracked in front matter.

## RELATIONSHIPS

- builds on: 0030-reference-graph-routing
- builds on: 0027-implement-readiness-gate
- related to: 0021-project-validation-and-instruction-registry
- related to: 0028-project-refresh-advisory

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

- [SPEC-01] Markdown artifacts remain canonical. `.kit/state.json`, task bundles, and run JSON must be generated from `SPEC.md`, `PLAN.md`, `TASKS.md`, project metadata, and run artifacts.
- [SPEC-02] `TASKS.md` must support optional task-detail fields `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK`; new templates and prompts must include them.
- [SPEC-03] Existing feature docs without the new task fields must remain valid unless a user opts into stricter verification policy.
- [SPEC-04] Verification declarations must support simple command bullets such as ``- `go test ./...` `` and normalize them into a typed command model.
- [SPEC-05] `kit verify [feature]` must discover declared checks from `TASKS.md` first, then feature-level plan/testing context only when needed.
- [SPEC-06] `kit verify --task T###` must run only the selected task's declared checks and use the task's expected file list for scope reporting.
- [SPEC-07] `kit verify --json` must emit stable JSON containing feature, task scope, commands, exit codes, durations, output artifact paths, touched-file summary when available, and pass/fail status.
- [SPEC-08] `kit verify` must persist each run by default under `.kit/runs/<run-id>/` with `run.json`, bounded/redacted command output artifacts, and a compact `summary.md`.
- [SPEC-09] Run IDs must be deterministic enough for ordering and unique enough for concurrent work, using timestamp plus short entropy or similar local-only scheme.
- [SPEC-10] Verification execution must avoid shell evaluation by default; shell execution requires an explicit declaration or opt-in flag.
- [SPEC-11] Verification output capture must avoid storing environment variables and must redact common secret patterns before writing artifacts.
- [SPEC-12] `kit reflect [feature]` must read the latest relevant run evidence, summarize pass/fail state, and require explicit evidence when marking reflection complete.
- [SPEC-13] `kit reflect` must warn when no relevant verification evidence exists for a feature with declared checks.
- [SPEC-14] `kit trace <feature>` must list run artifacts for the feature with status, timestamps, task IDs, commands, and artifact paths.
- [SPEC-15] `kit trace <run-id>` must show a compact run detail view without inlining large stdout/stderr artifacts.
- [SPEC-16] `kit replay <run-id>` must rerun the recorded verification command set and write a new linked run artifact; it must not claim to reconstruct model reasoning.
- [SPEC-17] `.kit/state.json` must be generated, deterministic, and explicitly non-authoritative; it must include source hashes or mtimes sufficient to detect stale cache state.
- [SPEC-18] Agent-readable state must include feature status, task progress, task bundles, references, latest verification status, and next recommended command.
- [SPEC-19] Task bundles must include task ID, source feature, allowed/expected paths, dependencies, verification commands, risk, rollback guidance, and handoff requirements.
- [SPEC-20] RLM behavior must be preserved: generated state and command output should point to relevant sections/artifacts instead of copying whole documents into prompts or JSON.
- [SPEC-21] `kit eval` must run local fixtures that test Kit harness behavior, including messy docs, incomplete tasks, stale summaries, verify parser behavior, trace/replay shape, and reflect evidence gating.
- [SPEC-22] `kit check --project` must validate the repo-level contract for new generated state/run artifact guidance after docs and templates are updated.
- [SPEC-23] README, `docs/CONSTITUTION.md`, agent docs, and generated instruction templates must be updated only where the shipped contract changes.
- [SPEC-24] `kit reconcile` must treat missing executable task fields as advisory for active work and must not invalidate legacy or historical `TASKS.md` files solely because they lack `VERIFY`, `EXPECTED FILES`, `RISK`, or `ROLLBACK`.
- [SPEC-25] `kit reconcile --migrate-verification` must output a migration prompt only; it must not edit files directly or guess verification commands from prose.

## ACCEPTANCE

- Running `kit verify executable-verification-harness --task T002 --json` returns schema-valid JSON and writes a local run artifact when T002 has declared checks.
- Running `kit verify executable-verification-harness --dry-run` lists the commands that would run without executing them.
- Running `kit trace executable-verification-harness` lists relevant runs without requiring agents to parse `.kit/runs/` manually.
- Running `kit replay <run-id>` creates a new run linked to the prior run and compares command outcomes.
- Running `kit reflect executable-verification-harness` includes latest verification evidence or an explicit missing-evidence warning.
- Generated `.kit/state.json` is reproducible from docs and run artifacts, and `kit check --project` detects stale generated state when strict mode is enabled.
- New `TASKS.md` templates include `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK`.
- Existing legacy feature docs still pass normal `kit check --all` unless they contain invalid existing metadata.
- `kit eval` fails when fixture expectations are not met and passes when check/verify/trace/reflect behavior matches fixtures.
- `kit reconcile` warns, rather than fails, when the active feature has task details missing executable verification fields.
- `kit reconcile --migrate-verification` emits prompt instructions to inspect active `TASKS.md`, add fields only where checks are known, run `kit verify <feature> --dry-run`, refresh `.kit/state.json`, and rerun `kit check`.
- `go test ./...`, `go run ./cmd/kit check executable-verification-harness`, `go run ./cmd/kit check --project`, and `go run ./cmd/kit check --all` pass after implementation.

## EDGE-CASES

- A task has no `VERIFY` field: `kit verify` reports `no_declared_checks` for that task and does not treat it as a command failure unless strict mode is requested.
- A command exits non-zero: the run is persisted with failure status, and `kit reflect` surfaces the failure without hiding partial evidence.
- A command times out: the run records timeout status and bounded output captured before termination.
- A command requires shell syntax: Kit rejects it by default and points to the explicit shell opt-in path.
- A run output contains a likely secret: stored artifacts contain redacted text and the run records that redaction occurred.
- `.kit/state.json` is stale: commands that consume it regenerate or warn with the stale source details.
- `.kit/runs/` is missing or deleted: trace commands report no local run history without breaking document workflows.
- Multiple tasks share the same verify command: Kit may de-duplicate in dry-run/state views, but run artifacts must still map evidence back to each task.
- A feature has large stdout/stderr artifacts: human views show paths and summaries, not full output.
- Work happens in a git worktree: run artifacts remain local to that checkout unless a future feature explicitly defines shared storage.
- Legacy tasks omit executable fields: normal validation and reconciliation keep them compatible unless the feature is actively being implemented, verified, or reflected.
- A migration prompt sees prose-only acceptance criteria: it asks for proposed runnable checks separately from confirmed commands instead of guessing.

## OPEN-QUESTIONS

- Should `kit verify` persist run artifacts by default for all users, or should first-run output clearly advertise `--no-write` for ephemeral checks?
- Should strict verification policy be a `.kit.yaml` option in this feature or a follow-up after default behavior settles?
- Should `.kit/state.json` be refreshed implicitly by read commands, or should a visible `kit state refresh` command be the primary interface?
