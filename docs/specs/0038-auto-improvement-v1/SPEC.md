---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: deliver
clarification:
  status: ready
  confidence: 99
  unresolved_questions: 0
feature:
  id: 0038
  slug: auto-improvement-v1
  dir: 0038-auto-improvement-v1
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0038-auto-improvement-v1
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
  - id: auto-improvement-strategy-note
    name: Auto-improvement strategy note
    type: note
    target: docs/notes/0038-auto-improvement-v1/references/AUTO_IMPROVEMENT.md
    relation: informs
    read_policy: must
    used_for: source strategy material migrated from root AUTO_IMPROVEMENT.md
    status: active
  - id: agents-readme
    name: Agent routing entrypoint
    type: doc
    target: docs/agents/README.md
    relation: guides
    read_policy: must
    used_for: repo-local workflow routing
    status: active
  - id: workflows
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    relation: constrains
    read_policy: must
    used_for: v2 SPEC workflow and readiness gates
    status: active
  - id: guardrails
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion bar, safety, and delivery hard gate
    status: active
  - id: tooling
    name: Tooling guidance
    type: doc
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: capabilities and project-directory workflow
    status: active
  - id: command-capabilities-rule
    name: Command capabilities rule
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: command-surface metadata requirements
    status: active
  - id: agent-team-orchestration-rule
    name: Agent team orchestration rule
    type: ruleset
    target: docs/references/rules/agent-team-orchestration.md
    relation: guides
    read_policy: conditional
    used_for: Agent Team Plan and verification topology
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## THESIS

# Kit Auto-Improvement Strategy

## Purpose

Define an exact implementation strategy for `kit improve`: a benchmark-backed,
daily self-improvement loop that can discover recurring Kit harness weaknesses,
propose small harness changes, validate them against regression criteria, and
open a pull request only when the change meaningfully improves Kit.

This design is grounded in two operating theories:

- Self-Harness: run the current harness on tasks, mine recurring failure
  patterns from execution traces, propose minimal harness edits, and accept
  candidates only after regression validation.
- autoresearch: keep the experiment loop small, measurable, repetitive, and
  auditable; use one or more objective scores to decide whether to keep or
  discard a proposed change.

## Strategy Comprehensiveness Rubric

Use this rubric to score whether the `kit improve` goal and implementation
strategy are complete enough to hand to a coding agent. Score each category
independently, then sum to 100.

| Category | Points | Complete Strategy Means |
|---|---:|---|
| Goal clarity and success definition | 15 | The command purpose, user outcome, V1 completion bar, and non-goals are explicit and measurable. |
| Harness scope and editable-surface boundaries | 10 | Allowed and forbidden surfaces are precise enough to block unsafe or irrelevant changes. |
| Benchmark and task model completeness | 15 | Tasks define fixtures, prompts, oracles, assertions, mutation policy, and expected behavior reproducibly. |
| Trace, weakness mining, and proposal pipeline | 15 | Runs emit enough evidence to cluster true failure modes and create targeted candidates without rerunning everything. |
| Validation, scoring, and held-out rigor | 15 | Candidate scoring compares before/after behavior, protects held-out regressions, handles flakes, and rejects gaming. |
| Safety, delivery, and regression controls | 10 | The workflow preserves dirty work, secrets, branch rules, PR gates, and allowed-surface enforcement. |
| Implementation phases and tests | 10 | Phases are ordered, testable, and include enough unit, CLI, fixture, and dry-run coverage. |
| Automation, reporting, and lifecycle | 10 | Scheduled runs, artifacts, retention, benchmark maintenance, PR evidence, and human review are defined. |

### Improvement Loop Ledger

This strategy was scored and refined using the rubric above. Each loop changed
one fact or strategy point, then re-scored the whole document.

| Loop | Score | Focused Improvement |
|---:|---:|---|
| 0 | 86 | Initial document had strong architecture but no explicit meta-rubric, incomplete task oracles, and underspecified isolation/scoring lifecycle. |
| 1 | 89 | Added this 100-point comprehensiveness rubric so future review can score strategy quality consistently. |
| 2 | 91 | Made benchmark tasks include explicit expected behavior, oracle type, mutation policy, and held-out eligibility. |
| 3 | 93 | Added suite manifests with held-in/held-out selection rules and hidden-from-proposer boundaries. |
| 4 | 95 | Added runner isolation and reproducibility rules for disposable fixture workspaces, stable binaries, and bounded network use. |
| 5 | 96 | Tightened weakness mining with reproducibility, counterexample, confidence, and flake-rate requirements. |
| 6 | 97 | Expanded candidate metadata with patch, rationale, expected score delta, negative controls, and rollback information. |
| 7 | 98 | Added score normalization and no-gaming rules so accepted candidates cannot win by broad rewrites or benchmark leakage. |
| 8 | 99 | Added artifact retention and benchmark lifecycle maintenance so reports remain auditable without committing trace dumps. |
| 9 | 100 | Added explicit stop conditions and human review gates for uncertain validation, unsafe surfaces, and PR delivery failures. |
| 10 | 100 | Added schema contracts so the implementation can define stable required fields, optional fields, and compatibility rules. |
| 11 | 100 | Added command contracts so every subcommand has deterministic output, exit status, dry-run, and JSON behavior. |
| 12 | 100 | Added fixture lifecycle rules for checked-in fixtures, generated fixtures, golden files, and fixture drift review. |
| 13 | 100 | Added oracle and assertion contracts so validation semantics are implementation-ready instead of implied by examples. |
| 14 | 100 | Added candidate application mechanics that use disposable copies and patch application instead of worktrees. |
| 15 | 100 | Added privacy and trace bounds for log capture, environment redaction, artifact upload, and private repo context. |
| 16 | 100 | Added baseline drift and versioning policy so old scorecards are not compared against incompatible main-branch behavior. |
| 17 | 100 | Added review ownership rules for accepted candidates, override paths, and issue-only fallback when PR evidence is weak. |
| 18 | 100 | Added failure taxonomy and triage routing so non-harness failures do not become low-quality harness rewrites. |
| 19 | 100 | Added explicit V1 non-goals to keep the auto-improvement loop bounded, auditable, and reversible. |

## Definition Of Harness

For Kit, the "harness" is the operating layer that shapes how coding agents
work inside Kit-managed repositories. V1 should treat these surfaces as
editable harness surfaces:

- `docs/agents/*`
- `docs/references/rules/*`
- generated prompt templates
- `kit capabilities` catalog descriptions
- `kit spec`, `kit loop prompt`, `kit pr fix`, and `kit dispatch` prompt text
- validation/check guidance
- small CLI checks that enforce already-documented harness rules

V1 should not use `kit improve` to rewrite unrelated product code, secrets,
GitHub permissions, dependency graphs, or protected branch settings.

## Target Command Surface

Add a top-level command:

```bash
kit improve
```

Recommended subcommands:

```bash
kit improve run
kit improve mine
kit improve propose
kit improve validate
kit improve pr
kit improve report
```

Recommended common flags:

```bash
--suite <name>              # default: default
--from <artifact-dir>       # read prior traces/reports
--candidate <id|path>       # select candidate proposal
--max-candidates <n>        # default: 3
--repeat <n>                # default: 1 locally, higher in scheduled mode
--dry-run
--json
--create-pr
```

## Command Contracts

Every `kit improve` command must support stable human-readable output and
machine-readable `--json` output. JSON output is part of the command contract
and should be covered by schema or golden-output tests.

Exit status contract:

- `0`: command completed and produced a valid result, even if the result is
  "no actionable clusters" or "candidate rejected".
- `1`: command failed because validation, assertions, or acceptance gates failed.
- `2`: command usage error, invalid flags, malformed suite/task/candidate input,
  or missing required artifact paths.
- `3`: command stopped because repository state, dirty worktree, identity,
  GitHub auth, or delivery contract checks are unsafe or ambiguous.
- `4`: command result is inconclusive because required reproducibility data,
  baseline evidence, network access, or fixture state was unavailable.

Dry-run contract:

- `--dry-run` may read files, build plans, and inspect GitHub state when the
  command already supports GitHub inspection.
- `--dry-run` must not mutate source files, fixture templates, `.kit.yaml`,
  `.git`, GitHub issues, branches, PRs, or uploaded artifacts.
- `--dry-run --json` must include the same decision fields that a mutating run
  would use, with `would_*` fields for planned writes.

Subcommand output contract:

- `run` emits a run manifest, task trace paths, assertion summary, and
  inconclusive/failure reasons.
- `mine` emits weakness clusters, confidence, flake rate, representative trace
  paths, actionability, and proposed future eval coverage.
- `propose` emits candidate metadata, target cluster, patch path or prompt path,
  editable surfaces, regression risks, and rejection reason when no candidate is
  produced.
- `validate` emits baseline reproduction status, candidate status, held-in and
  held-out scorecards, repository validation, allowed-surface status, and final
  accept/reject/inconclusive decision.
- `pr` emits the delivery contract, issue/branch/PR identifiers, commit hash
  when created, PR URL when created, and exact validation evidence.
- `report` emits a summarized Markdown report and the JSON source paths used to
  produce it.

Default local behavior:

```bash
kit improve
```

Should run an interactive flow:

1. Select benchmark suite.
2. Confirm editable harness surfaces.
3. Run or select existing traces.
4. Mine weaknesses.
5. Generate candidate proposals.
6. Validate selected candidates.
7. Offer to create a PR only if the acceptance criteria pass.

Default scheduled behavior:

```bash
kit improve run --suite default --json
kit improve mine --from .kit/improve/latest --json
kit improve propose --from .kit/improve/latest --max-candidates 3 --json
kit improve validate --from .kit/improve/latest --json
kit improve pr --candidate <accepted-candidate> --create-pr
```

## Repository Layout

Add a versioned evaluation and artifact structure:

```text
docs/evals/kit-improve/
  README.md
  suites/
    default.yaml
    github-delivery.yaml
    refresh-registry.yaml
    review-loop.yaml
  tasks/
    capabilities-discovery.yaml
    coderabbit-review-loop.yaml
    github-pr-delivery-contract.yaml
    init-refresh-idempotency.yaml
    readme-rules-private-repo.yaml
    registry-rule-preservation.yaml
  fixtures/
    repos/
      minimal-kit-project/
      dirty-worktree-project/
      registry-drift-project/
      existing-pr-project/
  schemas/
    suite.schema.json
    task.schema.json
    trace.schema.json
    candidate.schema.json
    scorecard.schema.json
```

Generated run artifacts should live under ignored local state:

```text
.kit/improve/
  latest -> runs/<timestamp>/
  runs/
    <timestamp>/
      run.json
      traces/
      weakness-report.json
      candidates/
      validation/
      scorecard.json
      report.md
```

The `.kit/improve/runs/*` directory should not be committed by default. PRs
should include a summarized report, not raw trace dumps, unless the user asks.

## Schema Contracts

The schemas under `docs/evals/kit-improve/schemas/` are implementation
contracts, not loose documentation examples.

General schema rules:

- Every schema includes `schema_version`, stable `id` when the object is named,
  and enough provenance to reproduce the object.
- Unknown fields are rejected by default in committed suite/task definitions and
  allowed only under an explicit `metadata` object.
- Additive optional fields may be introduced in a minor schema revision.
- Required field removal, semantic reinterpretation, or incompatible enum
  changes require a major schema revision and migration note.
- CLI `--json` output must identify the schema name and version that validates
  the payload.

`suite.schema.json` required fields:

- `schema_version`
- `id`
- `title`
- `held_in`
- `held_out`
- `repeat`
- `minimum_tasks`
- `selection_rules`

`task.schema.json` required fields:

- `schema_version`
- `id`
- `title`
- `category`
- `fixture`
- `input_prompt` or `commands`
- `expected_behavior`
- `oracle`
- `mutation_policy`
- `allowed_surfaces`
- `assertions`
- `regression_tags`
- `held_out_eligible`

`trace.schema.json` required fields:

- `schema_version`
- `task_id`
- `suite`
- `kit_version`
- `git_commit`
- `started_at`
- `duration_ms`
- `status`
- `workspace_path`
- `repeat_index`
- `commands`
- `assertions`
- `changed_files`
- `allowed_surface_violations`
- `oracle_results`

`candidate.schema.json` required fields:

- `schema_version`
- `id`
- `target_cluster`
- `editable_surfaces`
- `summary`
- `expected_effect`
- `rationale`
- `regression_risks`
- `rollback`
- `status`

`scorecard.schema.json` required fields:

- `schema_version`
- `candidate_id`
- `baseline_run_id`
- `candidate_run_id`
- `score`
- `before`
- `after`
- `hard_gate_statuses`
- `acceptance`
- `reasons`
- `validation_commands`

## Fixture Lifecycle

Fixtures are benchmark inputs and should be reviewed with the same seriousness
as tests.

- Prefer checked-in fixture repositories for deterministic historical failure
  cases.
- Allow generated fixtures only when the generator is committed, deterministic,
  and covered by tests.
- Store golden expected output beside the task definition when assertions depend
  on long prompt text, report text, or JSON shape.
- Fixture updates must state whether they are adapting to intentional Kit
  behavior, adding a new historical failure, or removing obsolete behavior.
- Fixture drift must not be silently accepted. If a fixture no longer reproduces
  the intended baseline behavior, mark the task `needs-refresh` and require
  human review before changing the expected outcome.
- Do not let `kit improve propose` edit fixtures, golden files, task
  definitions, or suite split policies unless a human explicitly requests
  benchmark maintenance.

## Oracle And Assertion Contracts

Oracles should be small, deterministic adapters that turn task output into
pass/fail/inconclusive evidence.

Required V1 oracle types:

- `deterministic-cli`: checks command exit status, stdout/stderr fragments, and
  generated files.
- `snapshot`: compares normalized text or JSON snapshots with clear update
  rules.
- `json-schema`: validates JSON output against a committed schema.
- `git-state`: checks branch, diff, status, staged files, commits, and ignored
  paths.
- `prompt-contract`: checks generated prompt content for required and forbidden
  instruction fragments.
- `github-thread-state`: checks PR review threads, unresolved comment filters,
  and resolve-instruction behavior using mocked or recorded GitHub responses.

Assertion results must include:

- assertion type
- target command or artifact
- normalized expected value
- normalized actual value when safe to print
- status: `passed`, `failed`, or `inconclusive`
- concise failure message

An oracle must return `inconclusive` rather than `passed` when required inputs
are missing, redacted, malformed, or produced by an unsupported tool version.

## Candidate Application Mechanics

Candidate validation must keep the source checkout and user worktree clean.

- Apply candidate patches only to disposable copied validation workspaces under
  `.kit/improve/runs/<id>/validation/`.
- Do not use git worktrees for candidate validation.
- Do not apply candidate patches to the source checkout until `kit improve pr`
  has passed the normal Kit delivery hard gate and the user has approved PR
  creation.
- Validate candidates serially by default so logs, branches, and failure causes
  remain easy to inspect.
- If a candidate patch does not apply cleanly to the validation workspace, mark
  it `rejected:patch-conflict`.
- If multiple candidates touch overlapping surfaces, validate them independently
  first. Combining candidates requires a separate combined-candidate scorecard.
- Cleanup may delete disposable validation workspaces only after the report has
  recorded the candidate id, patch hash, trace bundle hash, and rejection or
  acceptance reason.

## Privacy, Security, And Trace Bounds

Traces should be useful for debugging without becoming a private data dump.

- Capture at most 200 lines of stdout and 200 lines of stderr per command by
  default, with a flag to raise the limit for local debugging.
- Store full logs only in local ignored artifacts unless the user explicitly
  approves upload.
- Redact environment variables by default and record only an allowlist of
  relevant version/config keys.
- Scan traces for common tokens, private keys, `.env` values, GitHub tokens,
  service credentials, and local absolute paths before upload.
- Replace private repository paths with stable fixture-relative paths in
  committed reports whenever possible.
- Do not upload raw traces from private repositories unless a human explicitly
  approves the artifact upload after reviewing the redaction report.
- PR summaries should reference trace bundle hashes and curated excerpts, not
  raw logs.

## Baseline Drift And Versioning

Candidate validation is meaningful only when baseline and candidate runs are
comparable.

- Record the Kit source commit, eval suite hash, task hash, fixture hash, and
  binary hash for every baseline and candidate run.
- If `main` changes between baseline and candidate validation, rerun the
  baseline or mark the candidate `inconclusive:baseline-stale`.
- Scorecards from different suite, task, fixture, or binary hashes must not be
  compared as direct before/after evidence.
- Accepted candidates should include the baseline run id and candidate run id
  used for the PR decision.
- If a benchmark task is intentionally updated, close out old scorecards as
  historical evidence and start a new baseline series.
- Scheduled runs should keep the latest passing baseline for each suite and
  explicitly report when no fresh baseline exists.

## Benchmark Task Model

Each benchmark task should be declarative and reproducible.

Example:

```yaml
schema_version: 1
id: init-refresh-idempotency
title: kit init refresh is idempotent
category: registry-refresh
fixture: fixtures/repos/registry-drift-project
persona: coding-agent
timeout_seconds: 120
input_prompt: Run Kit refresh twice and verify the second run is a no-op.
expected_behavior: The second forced refresh reports no file changes and leaves the fixture clean.
oracle: deterministic-cli
mutation_policy: fixture-only
allowed_surfaces:
  - .kit.yaml
  - docs/**
  - .github/workflows/auto-assign.yml
commands:
  - kit init --refresh --force
  - kit init --refresh --force
assertions:
  - type: stdout_contains
    command_index: 1
    value: "Created: 0"
  - type: stdout_contains
    command_index: 1
    value: "Updated: 0"
  - type: git_diff_empty
regression_tags:
  - init
  - registry
  - idempotency
held_out_eligible: true
known_failure_modes:
  - refresh-not-idempotent:managed-rules
```

Required V1 task categories:

- Init and refresh idempotency.
- Registry-managed rule adoption and local status preservation.
- Local-custom rule preservation.
- `kit capabilities` command discoverability.
- GitHub delivery contract prompt correctness.
- Ready-for-review PR rule handling.
- CodeRabbit review extraction and resolve prompt behavior.
- README rules for public and private repositories.
- Dirty worktree preservation.

Each task must define:

- fixture source and fixture copy policy
- user or agent input prompt when a prompt-producing command is being tested
- expected behavior in plain language
- oracle type, such as `deterministic-cli`, `snapshot`, `json-schema`,
  `git-state`, `prompt-contract`, or `github-thread-state`
- assertions that map to the expected behavior
- allowed mutation surfaces inside the disposable fixture
- regression tags and held-out eligibility
- known failure modes when the task is based on a historical failure

## Suite Model

Each suite manifest should be explicit about task selection and split policy.

Example:

```yaml
schema_version: 1
id: default
title: Default Kit harness regression suite
held_in:
  include_tags:
    - init
    - github-delivery
    - review-loop
held_out:
  include_tags:
    - capabilities
    - dirty-worktree
    - readme
  hidden_from_proposer: true
repeat: 1
minimum_tasks: 8
selection_rules:
  - include at least one task from each required V1 category when available
  - do not put two fixture variants of the same failure in both held-in and held-out
```

Held-out tasks must not be shown to `kit improve propose`; proposal generation
receives only held-in failures, weakness clusters, editable surfaces, and broad
regression tags. Validation may run held-out tasks after candidates are created.

## Execution Trace Model

Each task run should produce a structured trace:

```json
{
  "schema_version": 1,
  "task_id": "init-refresh-idempotency",
  "suite": "default",
  "kit_version": "dev",
  "git_commit": "abc1234",
  "started_at": "2026-07-05T00:00:00Z",
  "duration_ms": 1000,
  "status": "failed",
  "workspace_path": ".kit/improve/runs/2026-07-05T000000Z/workspaces/init-refresh-idempotency",
  "baseline_trace_id": "baseline-abc1234",
  "repeat_index": 1,
  "seed": "default",
  "commands": [
    {
      "argv": ["kit", "init", "--refresh", "--force"],
      "exit_code": 0,
      "stdout_path": "traces/init-refresh/stdout.txt",
      "stderr_path": "traces/init-refresh/stderr.txt"
    }
  ],
  "assertions": [
    {
      "type": "git_diff_empty",
      "status": "failed",
      "message": "second refresh still changed managed rules"
    }
  ],
  "changed_files": [
    "docs/references/rules/github-pr-delivery.md"
  ],
  "allowed_surface_violations": [],
  "oracle_results": [
    {
      "oracle": "git-state",
      "status": "failed",
      "message": "fixture has tracked changes after the second refresh"
    }
  ],
  "failure_signature": "refresh-not-idempotent:managed-rules"
}
```

The trace must be detailed enough to support weakness mining without requiring
the proposer to rerun every command.

## Runner Isolation And Reproducibility

`kit improve run` must never run benchmark tasks directly against the source
checkout unless a task explicitly declares read-only source inspection.

Runner rules:

- Build or resolve one Kit binary for the run and record its path and version.
- Copy each fixture to a disposable workspace under `.kit/improve/runs/<id>/`.
- Run destructive setup, `git clean`, or fixture reset commands only inside the
  disposable workspace.
- Redact environment variables and known secret patterns before writing traces.
- Disable network by default unless the task declares `network: required`.
- Record OS, architecture, Go version, git version, `gh` version when used, and
  relevant environment toggles.
- Prefer deterministic time, random seed, and fixture data where possible.
- If reproducibility data is missing, mark the run `inconclusive` rather than
  using it as acceptance evidence.

## Weakness Mining

`kit improve mine` should read traces and produce clustered failure patterns.

The mining step should group failures by verifier-grounded signatures, not by
surface symptoms alone. Example clusters:

- `refresh-not-idempotent:managed-rules`
- `github-delivery:generic-pr-default-used`
- `review-loop:resolved-comments-not-filtered`
- `capabilities:missing-command-guidance`
- `dirty-worktree:unrelated-file-touched`

Each cluster should include:

- affected tasks
- representative trace paths
- observed failure mode
- likely harness surface
- whether the failure is plausibly addressable by a harness change
- confidence
- reproducibility count
- flake rate
- counterexamples where the same task or signature passed
- proposed evaluation tasks that should guard the cluster in future runs

Clusters that are not plausibly addressable by an editable harness surface
should be marked `not-actionable` and excluded from proposal generation.

An actionable cluster should usually require either:

- the same failure signature reproduced at least twice, or
- one high-severity failure that directly violates a documented Kit rule and has
  deterministic trace evidence.

Clusters with high flake rate or conflicting counterexamples should be marked
`needs-more-runs` unless the candidate is specifically a flake-reduction change.

## Harness Proposal

`kit improve propose` should generate candidate changes from actionable
weakness clusters.

Each candidate must be:

- tied to one primary failure cluster
- mapped to a concrete editable harness surface
- small enough to review
- materially distinct from other candidates
- explicit about the expected behavioral effect
- explicit about regression risk

Candidate metadata:

```json
{
  "schema_version": 1,
  "id": "candidate-001",
  "target_cluster": "github-delivery:generic-pr-default-used",
  "editable_surfaces": [
    "docs/references/rules/github-pr-delivery.md",
    "docs/agents/GUARDRAILS.md"
  ],
  "patch_path": ".kit/improve/runs/2026-07-05T000000Z/candidates/candidate-001/change.patch",
  "summary": "Make Kit delivery contract reload mandatory before PR mutation.",
  "expected_effect": "Agents stop before GitHub mutation and resolve repo-local delivery fields.",
  "expected_score_delta": {
    "held_in": "+2 tasks",
    "held_out": "0 regressions"
  },
  "rationale": "The failure cluster shows agents using global GitHub defaults after implementation, so a delivery-boundary rule is narrower than a broad prompt rewrite.",
  "negative_controls": [
    "docs-only task should not trigger a new issue/branch/PR gate"
  ],
  "regression_risks": [
    "May over-gate simple docs-only work if wording is too broad."
  ],
  "rollback": "Remove the added delivery-boundary paragraph and rerun github-delivery held-in tasks.",
  "status": "proposed"
}
```

The proposal engine must prefer precise harness edits over generic instruction
expansion. A large prompt rewrite should lose to a narrow rule or validator when
both address the same cluster.

Candidates must also explain why the fix belongs in the harness rather than in
unrelated product code. If the target cluster is best solved by a product bug
fix outside editable harness surfaces, mark the cluster `not-actionable` for
`kit improve propose`.

## Proposal Validation

`kit improve validate` should evaluate each candidate against the current
baseline.

Validation splits:

- Held-in: tasks whose failures motivated the proposal.
- Held-out: tasks not shown to the proposer and used as regression checks.
- Repository validation: normal project tests and checks.

Acceptance rule:

- held-in score improves, and
- held-out score does not regress, and
- required repository validation passes, and
- candidate changes only allowed harness surfaces, and
- candidate diff is reviewable and minimal.

The baseline and candidate runs should use the same suite, fixture versions,
repeat count, seed policy, and Kit binary build procedure. If the motivating
failure cannot be reproduced in the baseline, the candidate should be marked
`inconclusive` instead of accepted.

If evaluation is stochastic, repeat validation and score aggregate results.

Rejected candidates should be recorded with a reason. They should not be opened
as PRs.

## Scoring Model

Use a conservative scorecard. Example:

```text
score =
  + 10 * newly_passing_held_in_tasks
  + 15 * newly_passing_held_out_tasks
  - 25 * newly_failing_regression_tasks
  - 10 * flaky_or_inconclusive_tasks
  - 5  * broad_surface_penalty
  - 3  * prompt_bloat_penalty
  + 5  * simplification_bonus
```

Normalize candidate scores to a bounded range, such as `-100..100`, and keep the
accept/reject decision separate from the numeric score. A high positive score
must not override a hard failure in held-out regression, repository validation,
allowed-surface enforcement, or delivery rules.

Suggested PR threshold:

```text
score >= 10
newly_failing_regression_tasks == 0
required_validation_status == pass
allowed_surface_status == pass
```

No-gaming rules:

- Do not reward tasks that already passed before the candidate.
- Do not let candidates modify held-out task definitions.
- Do not let candidates weaken assertions, remove required tasks, or broaden
  allowed mutation surfaces without explicit human review.
- Apply prompt-bloat and broad-surface penalties before acceptance.
- Reject candidates that improve held-in results by making instructions vague,
  less enforceable, or harder to validate.

Scorecard output:

```json
{
  "schema_version": 1,
  "candidate_id": "candidate-001",
  "score": 20,
  "before": {
    "held_in_passed": 7,
    "held_in_total": 10,
    "held_out_passed": 8,
    "held_out_total": 10
  },
  "after": {
    "held_in_passed": 9,
    "held_in_total": 10,
    "held_out_passed": 8,
    "held_out_total": 10
  },
  "acceptance": "accepted",
  "reasons": [
    "held-in improved by 2",
    "held-out did not regress",
    "go test ./... passed"
  ]
}
```

## Pull Request Creation

`kit improve pr` should use the existing Kit GitHub delivery rules. It must not
invent a separate PR workflow.

The PR body should include:

```markdown
## Improvement Summary

- Target weakness:
- Harness surface changed:
- Why this should help:

## Scorecard

| Metric | Before | After | Delta |
|---|---:|---:|---:|
| Held-in pass count | 7/10 | 9/10 | +2 |
| Held-out pass count | 8/10 | 8/10 | 0 |
| Regression failures | 0 | 0 | 0 |

## Evidence

- Failure cluster:
- Representative traces:
- Validation commands:
- Rejected alternatives:

## Risk

- Possible overfitting:
- Prompt growth:
- Follow-up evals needed:
```

The PR should only be created when the candidate is accepted by the scorecard.
The PR must include the exact validation commands and observed results.

Before `kit improve pr` opens or updates a PR, it must re-run the normal Kit
delivery hard gate: issue assignment, branch identity, author/committer
identity, explicit staging, staged diff review, ready-for-review PR state, and
post-PR checks.

## GitHub Actions Strategy

Use GitHub Actions for deterministic execution and artifact persistence.

Recommended workflows:

```text
.github/workflows/kit-improve-eval.yml
.github/workflows/kit-improve-validate.yml
```

`kit-improve-eval.yml`:

- runs daily on `main`
- runs `kit improve run --suite default --json`
- uploads `.kit/improve/latest` as an artifact
- does not mutate source files

`kit-improve-validate.yml`:

- runs on candidate branches and PRs
- runs `kit improve validate --candidate <id> --json`
- uploads scorecards and validation reports
- fails if a candidate regresses held-out tasks or normal repo validation

Creative proposal generation should run in a scheduled Codex task or equivalent
agent environment, not directly inside a plain shell-only GitHub Action, unless
the runner has an approved agent execution environment.

## Scheduled Agent Strategy

Create a scheduled agent task that:

1. Fetches the latest `kit-improve-eval` artifact from `main`.
2. Runs `kit improve mine`.
3. Runs `kit improve propose --max-candidates 3`.
4. Applies each candidate in a disposable validation workspace.
5. Runs `kit improve validate`.
6. Creates a PR only for accepted candidates.
7. Leaves rejected candidates in the improvement report.

The scheduled agent should obey the same GitHub delivery hard gate used by all
Kit PR creation workflows.

## Artifact Retention And Benchmark Lifecycle

`kit improve` should keep enough evidence to audit decisions without committing
large or sensitive trace dumps.

- Commit suite definitions, task definitions, schemas, small fixtures, and
  curated scorecard summaries.
- Keep raw traces under `.kit/improve/runs/*` or GitHub Actions artifacts.
- Record content hashes for trace bundles referenced by PR summaries.
- Redact secrets before artifact upload.
- Prune local run artifacts by age or count, preserving the latest accepted and
  rejected candidate reports.
- Treat benchmark maintenance as part of command-surface maintenance: when a
  Kit command, prompt, rule, or validation behavior changes, add or update an
  eval task if the behavior is agent-facing.
- Require human review before deleting, weakening, or reclassifying benchmark
  tasks that protect a historical failure.

## Stop Conditions And Human Review Gates

The improvement loop should stop without creating a PR when:

- no actionable weakness clusters are found
- baseline failures cannot be reproduced
- every candidate is rejected or inconclusive
- a candidate touches a forbidden surface
- held-out tasks regress
- repository validation fails
- dirty worktree, branch, identity, auth, or PR state violates Kit delivery
  rules
- the proposed change needs product design judgment outside the editable
  harness surface

Human review is required before:

- changing the benchmark split policy
- weakening assertions or task oracles
- accepting a candidate with flaky evidence
- creating a PR that changes delivery, branch, commit, or review-thread rules
- uploading raw traces that may include project-local context

## Review Ownership And Approval Paths

`kit improve` may recommend and package improvements, but it must preserve a
human review boundary.

- The Kit maintainer owns acceptance of new harness behavior, benchmark split
  policy, and any change that increases agent autonomy.
- A passing scorecard is evidence for review, not permission to merge.
- If a candidate improves the suite but changes user-visible workflow semantics,
  `kit improve pr` should create or reuse an issue and describe the workflow
  tradeoff in the PR body.
- If evidence is promising but incomplete, create an issue or report instead of
  a PR.
- If the candidate needs product judgment, mark the weakness
  `needs-design-decision` and stop before generating a patch.
- If the candidate modifies delivery, review-thread, branch, commit, or
  identity rules, require explicit maintainer approval in the PR body checklist.

Human override is allowed only when it is explicit and recorded:

- `override_acceptance_reason`
- skipped or failed checks
- risk accepted
- follow-up issue when needed
- reviewer or maintainer who accepted the override

## Failure Taxonomy And Triage Routing

Mining should classify failures before proposing changes. The class determines
whether `kit improve propose` is allowed to act.

| Failure Class | Meaning | Routing |
|---|---|---|
| `harness-rule-gap` | The documented or generated agent instructions are incomplete, contradictory, or too weak. | Candidate may edit allowed harness docs/templates. |
| `harness-enforcement-gap` | A rule exists but Kit does not check or surface it. | Candidate may add narrow CLI validation or prompt checks. |
| `command-bug` | Kit command behavior violates documented behavior. | Create issue or candidate only if the command is an allowed harness surface. |
| `benchmark-gap` | The suite missed a historical failure or has weak assertions. | Require human-reviewed benchmark maintenance. |
| `fixture-drift` | Fixture no longer represents the intended scenario. | Mark `needs-refresh`; do not propose harness patch. |
| `external-service` | Failure depends on GitHub, network, auth, or third-party behavior. | Mark inconclusive unless recorded/mock evidence is enough. |
| `product-code` | Fix belongs in an app or downstream repo, not Kit harness. | Mark not actionable for `kit improve propose`. |
| `ambiguous` | Evidence is insufficient or contradictory. | Run more repeats or ask for human review. |

A cluster should not become a candidate until it has one primary failure class.
Mixed-class clusters should be split or marked `ambiguous`.

## V1 Non-Goals

V1 should be deliberately narrow.

- Do not autonomously merge PRs.
- Do not automatically weaken or delete benchmarks.
- Do not train, fine-tune, or persist model weights.
- Do not rewrite downstream product code as part of the improvement loop.
- Do not use private project traces as public benchmark data without explicit
  approval.
- Do not create broad prompt rewrites when a small rule, template, validator, or
  test can address the failure.
- Do not treat a numeric score as sufficient to bypass hard delivery, security,
  privacy, or human-review gates.
- Do not make scheduled automation mutate source files unless it is opening a
  reviewed PR through the Kit delivery contract.

## Implementation Phases

### Phase 1: Evaluation Schema And Runner

- Add eval suite/task schemas.
- Add fixture repository support.
- Add `kit improve run`.
- Emit structured traces and command/assertion results.
- Add initial default suite with 8-12 tasks.

Validation:

```bash
go test ./pkg/cli -run TestImprove
go test ./...
go run ./cmd/kit improve run --suite default --json
```

### Phase 2: Weakness Mining

- Add failure signature generation.
- Add cluster report generation.
- Add `kit improve mine`.
- Mark clusters actionable or not actionable.

Validation:

```bash
go test ./pkg/cli -run TestImproveMine
go run ./cmd/kit improve mine --from .kit/improve/latest --json
```

### Phase 3: Candidate Proposal

- Add candidate proposal prompt/template.
- Restrict editable surfaces.
- Add candidate metadata and patch storage.
- Add `kit improve propose`.

Validation:

```bash
go test ./pkg/cli -run TestImprovePropose
go run ./cmd/kit improve propose --from .kit/improve/latest --max-candidates 3 --json
```

### Phase 4: Validation And Scoring

- Add held-in and held-out split handling.
- Add candidate application in isolated validation workspace.
- Add scoring and acceptance rules.
- Add `kit improve validate`.

Validation:

```bash
go test ./pkg/cli -run TestImproveValidate
go run ./cmd/kit improve validate --candidate candidate-001 --json
```

### Phase 5: PR Packaging

- Add `kit improve pr`.
- Reuse existing GitHub delivery rules.
- Generate scorecard-rich PR bodies.
- Assign issue and PR according to Kit delivery settings.

Validation:

```bash
go test ./pkg/cli -run TestImprovePR
go run ./cmd/kit improve pr --candidate candidate-001 --dry-run
```

### Phase 6: Daily Automation

- Add eval workflow.
- Add validation workflow.
- Add scheduled agent runbook.
- Document operational setup.

Validation:

```bash
gh workflow run kit-improve-eval.yml
gh run watch
gh run view --log
```

## Required Tests

- Suite/task schema loading.
- Schema validation rejects unknown committed fields outside `metadata`.
- `--json` payloads include schema name and version.
- Command exit status mapping for success, validation failure, usage error,
  unsafe repo state, and inconclusive evidence.
- Dry-run behavior does not mutate source, fixture, `.git`, or GitHub state.
- Fixture copy and cleanup.
- Fixture drift detection and `needs-refresh` routing.
- Trace emission.
- Trace redaction and output line limit behavior.
- Assertion pass/fail behavior.
- Oracle inconclusive behavior for missing, redacted, malformed, or unsupported
  inputs.
- Failure signature stability.
- Weakness cluster grouping.
- Failure taxonomy routing for harness, benchmark, fixture, external-service,
  product-code, and ambiguous failures.
- Candidate allowed-surface enforcement.
- Candidate rejected when no editable surface changes.
- Candidate patch conflict rejection.
- Candidate validation uses disposable validation workspaces and does not use
  worktrees.
- Baseline stale detection when suite, task, fixture, source, or binary hashes
  differ.
- Held-in improvement accepted only when held-out does not regress.
- Regression rejection.
- Scorecard math.
- Human override fields are required when accepting skipped or failed checks.
- PR body generation.

## Risks And Mitigations

- Risk: prompt bloat.
  - Mitigation: add prompt-bloat penalty and require failure-cluster evidence.
- Risk: overfitting to the benchmark.
  - Mitigation: keep held-out tasks hidden from proposal generation.
- Risk: flaky results.
  - Mitigation: support repeats and mark inconclusive runs.
- Risk: agent mutates unsafe surfaces.
  - Mitigation: enforce editable surface allowlist before validation and PR.
- Risk: benchmark suite becomes stale.
  - Mitigation: require new Kit command/rule changes to add or update eval tasks
    where behavior is agent-facing.
- Risk: generated PRs are noisy.
  - Mitigation: create PRs only above the acceptance threshold and include
    rejected alternatives in the report.

## V1 Success Criteria

V1 is complete when:

- `kit improve run` can run a default benchmark suite and emit traces.
- `kit improve mine` can cluster failures into actionable weakness reports.
- `kit improve propose` can produce bounded candidate harness changes.
- `kit improve validate` can accept prompt-candidates for human review or
  reject/inconclusively report malformed candidate metadata.
- `kit improve pr` can prepare a PR body with scorecard and evidence.
- Daily evaluation artifacts can be produced by GitHub Actions.
- Candidate PR creation remains gated by existing Kit GitHub delivery rules.

## CONTEXT

### Pre-Instruction Report

- `SPEC.md` path: `docs/specs/0038-auto-improvement-v1/SPEC.md`.
- Workflow version: `2`.
- Current phase: `deliver`.
- Loaded repo instruction docs: `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md`, and `docs/agents/TOOLING.md`.
- Loaded relevant rules: `docs/references/rules/command-capabilities.md` and `docs/references/rules/agent-team-orchestration.md`.
- Feature notes loaded: `docs/notes/0038-auto-improvement-v1/README.md` and `docs/notes/0038-auto-improvement-v1/private/README.md`; `.gitkeep` files ignored.
- User thesis known so far: implement V1 of `kit improve`, a benchmark-backed self-improvement loop for Kit harness changes.
- Delivery intent known so far: user requested issue, branch, and PR creation; implementation is ready to commit and push to existing PR `#47`.
- Confidence: `99%`.
- Unresolved questions: `0`.
- Readiness gate status: passed; implementation has started.
- Accepted assumptions: DA-001, DA-002, DA-003 with file move, DA-004, DA-006, and DA-007.
- Rejected assumptions: DA-005 as originally phrased; user requested issue, branch, and PR creation for this feature.
- Defaulted assumptions: use existing project directory, do not use worktrees, keep `SPEC.md` as the single durable feature artifact.
- Still-unverified assumptions: none for scoped V1.
- Acceptance criteria inventory: locked `AC-001` through `AC-010` below.
- Touched areas: `pkg/cli/*`, `pkg/cli/capabilities_catalog.go`, `pkg/cli/root_help.go`, `internal/improve/**`, `docs/evals/kit-improve/**`, `.github/workflows/**`, `.gitignore`, and this `SPEC.md`.
- Validation strategy: targeted `go test ./internal/improve ./pkg/cli`, full `go test ./...`, sequential `kit improve` CLI flow, capability output checks, and `git diff --check`.
- Rollback checkpoint: before implementation, this SPEC and generated feature notes/progress entries are the only feature-scoped durable changes; future implementation should remain separable by feature files and tests.
- Evidence locations: inline command summaries in this `SPEC.md`, generated run artifacts under `.kit/runs/...` when used, and future `docs/evals/kit-improve/**` fixtures/schemas.
- Dirty worktree summary: `docs/PROJECT_PROGRESS_SUMMARY.md` modified; `docs/specs/0038-auto-improvement-v1/` and `docs/notes/0038-auto-improvement-v1/` untracked.
- Dirty worktree classification: generated feature docs/progress are in-scope for this SPEC workflow; former root `AUTO_IMPROVEMENT.md` was moved to `docs/notes/0038-auto-improvement-v1/references/AUTO_IMPROVEMENT.md` per user instruction and is source material, not canonical truth.
- Source Map status: initialized below; several implementation-scope facts remain unverified until current code structure is inspected in the implementation phase.
- Agent Team Plan status: recorded below. Current runtime has no separate subagent tool exposed, so implementation will use a single supervisor lane with an explicit read-only self-verification pass.

### Source Map

| ID | Source | Selector | Claim / Fact | Used For | Maps To | Status |
|---|---|---|---|---|---|---|
| SRC-001 | User attached prompt | `Context Provided By User` | User supplied a detailed V1 strategy for `kit improve`, including command surface, schemas, runner, mining, proposal, validation, scoring, PR packaging, automation, non-goals, and tests. | Requirements and draft acceptance criteria | AC-001..AC-010 | confirmed |
| SRC-002 | `docs/specs/0038-auto-improvement-v1/SPEC.md` | front matter | Feature is `0038-auto-improvement-v1`, workflow version is `2`, phase is `clarify`, and delivery intent is `issue_branch_pr_later`. | Workflow state and delivery gate | AC-010 | confirmed |
| SRC-003 | `git status --short --branch` | command output | Current branch is `main...origin/main`; dirty state includes modified `docs/PROJECT_PROGRESS_SUMMARY.md` and untracked `docs/specs/0038-auto-improvement-v1/` and `docs/notes/0038-auto-improvement-v1/`. | Dirty-worktree gate | tasks, delivery decision | confirmed |
| SRC-004 | `docs/agents/WORKFLOWS.md` | `Spec-Driven Work`, `Readiness Gate` | V2 feature work must use `SPEC.md`, ask clarification until high confidence, and pass readiness gates before writing code. | Clarification and readiness gating | AC-010 | confirmed |
| SRC-005 | `docs/agents/GUARDRAILS.md` | `Completion Bar`, `GitHub Delivery Hard Gate` | Completion requires current docs, passing relevant validation or explained skips, self-review before staging, and repo-local delivery contract before any GitHub mutation. | Validation, reflection, delivery planning | AC-009, AC-010 | confirmed |
| SRC-006 | `docs/references/rules/command-capabilities.md` | `Rules`, `Verification` | Any Kit command or command extension must update `pkg/cli/capabilities_catalog.go` and verify targeted/search capability output. | Capability metadata requirement | AC-008 | confirmed |
| SRC-007 | `go run ./cmd/kit capabilities --search improve --json` | command output | Current capability search returns no commands for `improve`; `kit improve` is not currently discoverable. | Existing behavior baseline | AC-001, AC-008 | confirmed |
| SRC-008 | `pkg/cli` and `internal` file inventory | `find pkg/cli cmd internal` | Existing CLI commands are implemented in `pkg/cli`; reusable primitives live under `internal/*`; no current `pkg/cli/improve.go` exists in the inspected file list. | Predicted touched files | AC-001..AC-007 | confirmed |
| SRC-009 | `docs/references/rules/agent-team-orchestration.md` | `Required Agent Team Plan` | Non-trivial implementation should plan specialist lanes and read-only verification, but actual spawned agents must be distinguished from logical lanes. | Future implementation topology | AC-010 | confirmed |
| SRC-010 | `docs/notes/0038-auto-improvement-v1/README.md` | `Directories` | Feature notes are optional source material; durable decisions must be promoted into `SPEC.md`. | Notes handling | AC-010 | confirmed |
| SRC-011 | User clarification response | `y 3` | User instructed the former root `AUTO_IMPROVEMENT.md` to move into `docs/notes/0038-auto-improvement-v1` and be used as source material. | Notes handling and dirty-worktree ownership | AC-010 | confirmed |
| SRC-012 | User clarification response | `n 5` | User requested creating the issue, branch, and pull request for this feature rather than only preparing a PR package after validation. | Delivery decision | AC-010 | confirmed |
| SRC-013 | GitHub issue | `#46` | Issue `#46` was created for this feature and assigned to `jamesonstone`. | Delivery traceability | AC-010 | confirmed |
| SRC-014 | Git branch | `GH-46` | Branch `GH-46` was created from freshly fetched `origin/main`; branch HEAD matched `origin/main` immediately after checkout. | Delivery traceability | AC-010 | confirmed |
| SRC-015 | GitHub pull request | `#47` | Pull request `#47` was created ready for review from `GH-46` into `main` and assigned to `jamesonstone`. | Delivery traceability | AC-010 | confirmed |
| SRC-016 | User clarification response | `yes` | User approved small checked-in fixture repositories under `docs/evals/kit-improve/fixtures/repos/` with no large generated traces committed. | Fixture scope and storage model | AC-002, AC-003, AC-009 | confirmed |
| SRC-017 | `internal/improve/*` | package implementation | Added strict suite/task loading, disposable fixture execution, trace artifact writing, output redaction, allowed-surface checks, weakness mining, prompt-candidate generation, metadata validation, report rendering, and PR body rendering. | Core implementation | AC-003..AC-007, AC-009 | confirmed |
| SRC-018 | `pkg/cli/improve.go` | command registration | Added top-level `kit improve` plus `run`, `mine`, `propose`, `validate`, `report`, and `pr` subcommands with `--json` support and PR creation remaining gated. | CLI surface | AC-001, AC-007, AC-008 | confirmed |
| SRC-019 | `docs/evals/kit-improve/**` | committed eval assets | Added default suite, eight read-only benchmark task definitions, schemas, README, and small fixture repos; tasks invoke the active Kit binary through `{{kit}}` and avoid generated trace commits. | Eval assets | AC-002, AC-003, AC-009 | confirmed |
| SRC-020 | `.github/workflows/kit-improve-*.yml` | workflow files | Added scheduled/manual eval workflow and PR/manual validation workflow for the new command surface. | Automation | AC-007, AC-009 | confirmed |
| SRC-021 | `pkg/cli/capabilities_catalog.go`, `pkg/cli/root_help.go` | command metadata | Added root-help and `kit capabilities` records for `improve` and all nested improve subcommands. | Discoverability | AC-001, AC-008 | confirmed |
| SRC-022 | Validation commands | command output | `go test ./...` passed; sequential `kit improve` flow passed with `run: pass traces=8`, `mine clusters: 0`, `propose candidates: 0`, `dry-run: dry_run`, and capability search reporting `7 improve commands`. | Validation evidence | AC-001..AC-010 | confirmed |

## CLARIFICATIONS

Current confidence: `99%`.

Unresolved questions: `0`.

Implementation readiness gate is passed for V1 scope.

Clarification batch 1 answers:

- Q1 accepted default: V1 uses deterministic CLI/report/prompt mechanics and does not embed a model runtime.
- Q2 accepted default: V1 may include GitHub Actions workflow files, while external scheduled agent setup remains documented/environment-specific.
- Q3 accepted with override: move `AUTO_IMPROVEMENT.md` into `docs/notes/0038-auto-improvement-v1` and use it as source material.
- Q4 accepted: fixtures are small checked-in repositories under `docs/evals/kit-improve/fixtures/repos/`; large generated traces are not committed.
- Q5 rejected as phrased: user requested issue, branch, and PR creation for this feature.
- Q6 accepted default: committed eval definitions live under `docs/evals/kit-improve/**`; ignored local run state lives under `.kit/improve/**`.
- Q7 accepted default: clarification stays single-lane; implementation topology is planned after readiness.
- Follow-up answer accepted Q4 default: fixtures are small checked-in repositories under `docs/evals/kit-improve/fixtures/repos/`; large generated traces are not committed.

No open clarification questions remain.

## REQUIREMENTS

Requirements:

- REQ-001: Add a top-level `kit improve` workflow for benchmark-backed Kit harness self-improvement.
- REQ-002: Support subcommands `run`, `mine`, `propose`, `validate`, `pr`, and `report`.
- REQ-003: Store committed eval definitions, schemas, suites, tasks, and small fixtures under `docs/evals/kit-improve/**`.
- REQ-004: Store generated run artifacts under ignored `.kit/improve/**`.
- REQ-005: Implement deterministic task execution, trace emission, weakness clustering, candidate metadata, validation scoring, and report generation.
- REQ-006: Enforce editable-surface, privacy, redaction, baseline-drift, and no-worktree rules.
- REQ-007: Update command help and `kit capabilities` metadata for the new command surface.
- REQ-008: Add targeted tests and documentation for the new workflow.
- REQ-009: Preserve delivery hard-gate behavior; no autonomous merge or generic GitHub delivery defaults.
- REQ-010: Keep `SPEC.md` current as the single durable feature workflow artifact.

Draft non-goals:

- NG-001: Do not autonomously merge PRs.
- NG-002: Do not train or fine-tune models.
- NG-003: Do not rewrite downstream product code as part of `kit improve`.
- NG-004: Do not upload raw private traces without explicit approval.
- NG-005: Do not use git worktrees for candidate validation.

## ASSUMPTIONS

Accepted assumptions:

- A-001: Work remains in the existing project directory; no worktrees.
- A-002: `SPEC.md` is the canonical durable feature artifact.
- A-003: V1 uses deterministic CLI/report/prompt mechanics and does not embed a model runtime.
- A-004: GitHub Actions workflow files may be added; external scheduled agent setup remains documented/environment-specific.
- A-005: The former root `AUTO_IMPROVEMENT.md` is source material under `docs/notes/0038-auto-improvement-v1/references/AUTO_IMPROVEMENT.md`.
- A-006: Committed eval definitions live under `docs/evals/kit-improve/**`; ignored local run state lives under `.kit/improve/**`.
- A-007: Clarification is single-lane; implementation topology will be planned after readiness.
- A-008: Fixtures are small checked-in repositories under `docs/evals/kit-improve/fixtures/repos/`; raw traces are ignored artifacts.

Default assumptions:

- none.

Removed assumptions:

- DA-005 removed as originally phrased. User requested issue, branch, and PR creation for this feature.

Blocking assumptions:

- none.

## ACCEPTANCE CRITERIA

Locked V1 acceptance criteria:

- AC-001: `kit improve` is registered as a visible top-level command with subcommands `run`, `mine`, `propose`, `validate`, `pr`, and `report`.
- AC-002: `docs/evals/kit-improve/**` contains committed suite/task/schema/fixture scaffolding matching the V1 strategy.
- AC-003: `kit improve run` can execute a deterministic suite against disposable fixture copies and emit structured traces without mutating the source checkout.
- AC-004: `kit improve mine` can read traces and produce actionable/not-actionable weakness clusters with confidence, reproducibility, and flake metadata.
- AC-005: `kit improve propose` can produce bounded candidate metadata and prompt artifacts tied to actionable clusters and allowed surfaces.
- AC-006: `kit improve validate` can validate candidate metadata and produce scorecard evidence while preserving the human review boundary; autonomous patch application is out of V1 scope.
- AC-007: `kit improve report` and `kit improve pr` can produce reviewable Markdown/JSON evidence and a PR package/body without bypassing the Kit delivery hard gate.
- AC-008: `kit capabilities` and root help discover and accurately describe `kit improve`, its subcommands, mutation behavior, artifacts, caveats, and examples.
- AC-009: Tests cover strict suite/task loading, command contracts, fixture lifecycle, trace redaction, allowed-surface checks, assertion/oracle behavior, weakness mining, candidate metadata validation, reporting, and capability metadata.
- AC-010: `SPEC.md` records Source Map, clarification decisions, implementation plan, task checklist, validation map, evidence, reflection, documentation updates, and delivery state.

## IMPLEMENTATION PLAN

Implementation plan:

1. Lock clarified V1 scope and acceptance criteria in this `SPEC.md`.
2. Inspect existing command patterns in `pkg/cli`, command capability metadata, runstore/verify helpers, and eval-related internals.
3. Define eval schemas and structs for suites, tasks, traces, candidates, scorecards, reports, and command JSON payloads.
4. Add CLI command registration and subcommand handlers for `kit improve`.
5. Implement fixture copying, deterministic command execution, assertion/oracle evaluation, trace writing, redaction, and dry-run behavior.
6. Implement weakness mining, proposal metadata generation, candidate validation/scoring, report rendering, and PR-body packaging.
7. Add committed eval definitions, small fixtures, schemas, docs, help, and capability metadata.
8. Add tests mapped to every acceptance criterion.
9. Validate, reflect, update this SPEC, and only then run delivery hard gate if requested.

Rollback strategy:

- Keep new command files, eval definitions, workflow files, and tests separable by feature.
- Do not mutate existing Git/GitHub state during implementation.
- If validation fails broadly, revert or remove the new command surface and committed eval files while preserving this SPEC evidence.

### Agent Team Plan

- Supervisor responsibilities: maintain this `SPEC.md`, integrate changes, enforce scope, update evidence, run validation, perform self-review, and manage delivery-state updates.
- Implementation lanes actually spawned as subagents: none.
- Read-only verification lanes actually spawned as subagents: none.
- Single-lane exception: the active runtime does not expose a separate subagent execution tool; use one supervisor lane with serialized implementation and an explicit read-only self-verification pass before reflection.
- Logical-only lanes:
  - CLI command lane: `pkg/cli/improve*.go`, root help, capability metadata, command tests.
  - Eval model lane: internal structs/parsers/runners/scoring and tests.
  - Docs/eval artifact lane: `docs/evals/kit-improve/**`, workflow docs, SPEC updates.
- Predicted touched files:
  - `pkg/cli/root_help.go`
  - `pkg/cli/capabilities_catalog.go`
  - new `pkg/cli/improve*.go` files and tests
  - possible new `internal/improve/**` or similarly scoped internal package
  - `docs/evals/kit-improve/**`
  - `.gitignore` or existing ignore config for `.kit/improve/**`
  - `.github/workflows/kit-improve-*.yml`
  - `docs/specs/0038-auto-improvement-v1/SPEC.md`
- Overlap risks: command handlers and internal model tests may evolve together; serialize CLI and model changes until the package boundaries are clear.
- Max concurrency: 1 actual lane in this runtime.
- Validation/review lanes: focused tests during implementation; final read-only self-verification comparing diff against ACs and Source Map.

## TASK CHECKLIST

| Task | Status | Lane | Maps To | Expected Evidence |
|---|---|---|---|---|
| T-001: Resolve clarification batch and lock readiness gates | complete | supervisor | AC-010 | Updated Clarifications, Requirements, Assumptions, ACs, Validation Map |
| T-002: Inspect command, eval, verify, runstore, and capability patterns | complete | supervisor | AC-001..AC-009 | Source Map entries and implementation notes |
| T-003: Implement eval schemas and data model | complete | supervisor | AC-002, AC-009 | `docs/evals/kit-improve/schemas/*.json`; strict YAML load tests |
| T-004: Implement `kit improve run` and trace emission | complete | supervisor | AC-003 | `kit improve run --suite default --json` passed with 8 traces |
| T-005: Implement `mine`, `propose`, `validate`, `report`, and `pr` mechanics | complete | supervisor | AC-004..AC-007 | Sequential CLI flow passed; metadata candidates remain human-review gated |
| T-006: Update capabilities, help, docs, and eval definitions | complete | supervisor | AC-002, AC-008, AC-010 | `kit capabilities improve --json`; `kit capabilities --search improve --json` returned 7 commands |
| T-007: Run validation, read-only verification, reflection, and delivery gate | complete | supervisor | AC-001..AC-010 | `go test ./...`, CLI flow, `git diff --check`, and diff review evidence recorded |

## VALIDATION MAP

| AC | Validation |
|---|---|
| AC-001 | `go test ./internal/improve ./pkg/cli` passed; `root_help` and command registration tests cover the visible command/subcommand surface. |
| AC-002 | `go test ./internal/improve` passed strict suite/task loading tests; `docs/evals/kit-improve/**` contains suite, task, schema, fixture, and README assets. |
| AC-003 | `go run ./cmd/kit improve run --suite default --json` passed with `status=pass` and `traces=8`; traces ran in disposable `.kit/improve/runs/*/workspaces/*` copies. |
| AC-004 | `go run ./cmd/kit improve mine --from .kit/improve/latest --json` passed with `clusters=0` for the passing suite; unit tests cover synthetic cluster proposal. |
| AC-005 | `go run ./cmd/kit improve propose --from .kit/improve/latest --max-candidates 3 --json` passed with `candidates=0` for the passing suite; unit tests cover generated candidate prompt metadata from a synthetic weakness report. |
| AC-006 | `go test ./internal/improve` passed candidate metadata validation coverage; V1 scorecards produce `accepted-for-review` for well-formed prompt candidates and do not apply patches autonomously. |
| AC-007 | `go run ./cmd/kit improve report --from .kit/improve/latest` and `go run ./cmd/kit improve pr --from .kit/improve/latest --issue 46` passed; PR creation remains intentionally gated. |
| AC-008 | `go run ./cmd/kit capabilities improve --json` passed and described `improve`; `go run ./cmd/kit capabilities --search improve --json` passed and returned `7 improve commands`. |
| AC-009 | `go test ./...` passed; `git diff --check` passed; line-count check shows all new/changed Go files under roughly 300 lines. |
| AC-010 | This `SPEC.md` records clarified decisions, implementation plan, Source Map, task statuses, validation evidence, reflection, documentation updates, and delivery state. |

## REFLECTION NOTES

- Implementation matched the accepted V1 scope: deterministic local evals, trace artifacts, weakness mining, prompt-candidate packaging, metadata validation, report/PR body generation, capability metadata, and GitHub Actions eval/validate workflows.
- A correctness pass found and fixed three issues before staging:
  - default tasks initially used `echo` placeholders; they now invoke the active Kit binary through `{{kit}}` and check real local command metadata
  - run IDs initially used second precision; they now use the existing high-precision `verify.NewRunID` helper to prevent local artifact collisions
  - `propose` initially required `weakness-report.json`; it now derives that report from traces when missing
- Safety/readability pass added strict YAML decoding, allowed-surface violation reporting, trace output redaction, and a helper-file split to keep new Go files below the rough 300-line readability target.
- V1 intentionally does not embed a model runtime and does not apply candidate patches autonomously. `kit improve validate` validates metadata and produces human-review scorecards; future work can add patch-backed candidate validation in disposable workspaces.
- Generated trace artifacts remain ignored under `.kit/improve/**`; only suite/task/schema/fixture definitions and command code are committed.

## DOCUMENTATION UPDATES

- `docs/specs/0038-auto-improvement-v1/SPEC.md`: updated as the canonical feature artifact and delivery evidence record.
- `docs/evals/kit-improve/README.md`: added overview of committed eval assets and ignored local artifacts.
- `docs/evals/kit-improve/suites/default.yaml`: added default held-in suite selecting eight V1 benchmark tasks.
- `docs/evals/kit-improve/tasks/*.yaml`: added eight read-only benchmark task contracts.
- `docs/evals/kit-improve/schemas/*.json`: added V1 schema contracts for suites, tasks, traces, candidates, and scorecards.
- `.github/workflows/kit-improve-eval.yml` and `.github/workflows/kit-improve-validate.yml`: added eval/validation workflow entrypoints.
- `pkg/cli/capabilities_catalog.go` and root help: updated so coding agents can discover the new command and subcommands through `kit capabilities`.

## DELIVERY DECISION

User requested creating a new issue, branch, and pull request for this feature using Kit-managed repository rules.

- Issue: `#46` (`https://github.com/jamesonstone/kit/issues/46`), assigned to `jamesonstone`.
- Branch: `GH-46`, created from freshly fetched `origin/main`.
- PR: `#47` (`https://github.com/jamesonstone/kit/pull/47`), ready for review and assigned to `jamesonstone`.

Implementation, validation, and reflection are complete for the scoped V1. The remaining delivery action is to commit and push these implementation changes to the existing ready-for-review PR `#47`, then report observed PR/CI state.

## EVIDENCE

Initial evidence:

- `git status --short --branch`: `## main...origin/main`; dirty files classified in Source Map `SRC-003`.
- `go run ./cmd/kit capabilities --search improve --json`: returned an empty command list, confirming no current `kit improve` capability record.
- `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md`, `docs/agents/TOOLING.md`, `docs/references/rules/command-capabilities.md`, and `docs/references/rules/agent-team-orchestration.md` inspected for workflow and command-surface requirements.
- `AUTO_IMPROVEMENT.md` moved to `docs/notes/0038-auto-improvement-v1/references/AUTO_IMPROVEMENT.md` per user instruction.
- GitHub issue `#46` created and verified assigned to `jamesonstone`.
- Branch `GH-46` created from fetched `origin/main`; checkout assertion and branch-base assertion passed.
- Pull request `#47` created ready for review from `GH-46` into `main` and assigned to `jamesonstone`.
- User approved Q4 default: small checked-in fixtures under `docs/evals/kit-improve/fixtures/repos/`, no large generated traces committed.
- Implementation phase started after readiness; existing command, eval, verify, root help, capability, and ignore patterns inspected.
- `go test ./internal/improve ./pkg/cli`: passed.
- `go test ./...`: passed.
- Sequential CLI flow passed:
  - `go run ./cmd/kit improve run --suite default --json`: `status=pass`, `traces=8`.
  - `go run ./cmd/kit improve mine --from .kit/improve/latest --json`: `clusters=0`.
  - `go run ./cmd/kit improve propose --from .kit/improve/latest --max-candidates 3 --json`: `candidates=0`.
  - `go run ./cmd/kit improve report --from .kit/improve/latest`: rendered `# Kit Improve Report`.
  - `go run ./cmd/kit improve pr --from .kit/improve/latest --issue 46`: rendered PR body beginning with `## Description`.
  - `go run ./cmd/kit improve run --suite default --dry-run --json`: `status=dry_run`.
- `go run ./cmd/kit capabilities improve --json`: returned command detail for `improve`.
- `go run ./cmd/kit capabilities --search improve --json`: returned `7 improve commands`.
- `git diff --check`: passed.
- `wc -l internal/improve/*.go pkg/cli/improve.go`: largest new/changed Go file is `pkg/cli/improve.go` at 250 lines.
