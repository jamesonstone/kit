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
- `kit improve validate` can accept or reject candidates using held-in and
  held-out scoring.
- `kit improve pr` can prepare a PR body with scorecard and evidence.
- Daily evaluation artifacts can be produced by GitHub Actions.
- Candidate PR creation remains gated by existing Kit GitHub delivery rules.
