---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: deliver
delivery_intent: issue_branch_pr_in_progress
summary: Modernize every generated prompt for GPT-5.6, prove the migration with a truthful deterministic benchmark, and preserve Kit's code-enforced workflow gates.
clarification:
  status: ready
  confidence: 99
  unresolved_questions: 0
feature:
  id: 0040
  slug: prompt-system-migration
  dir: 0040-prompt-system-migration
relationships:
  - type: builds_on
    target: 0038-auto-improvement-v1
  - type: builds_on
    target: 0022-typed-prompt-ir
  - type: builds_on
    target: 0024-frontend-profile
  - type: related_to
    target: 0025-v0-prompt-library
  - type: related_to
    target: 0035-loop-review
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0040-prompt-system-migration
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
  - id: openai-latest-model-guide
    name: OpenAI latest model guide
    type: documentation
    target: https://developers.openai.com/api/docs/guides/latest-model
    relation: guides
    read_policy: must
    used_for: GPT-5.6 prompt and autonomy guidance
    status: active
  - id: openai-codex-models
    name: OpenAI Codex models
    type: documentation
    target: https://developers.openai.com/codex/models
    relation: verifies
    read_policy: evidence
    used_for: Codex CLI GPT-5.6 model support
    status: active
  - id: agents-readme
    name: Agent routing entrypoint
    type: doc
    target: docs/agents/README.md
    relation: constrains
    read_policy: must
    used_for: repository task routing
    status: active
  - id: workflows
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    relation: constrains
    read_policy: must
    used_for: canonical v2 SPEC workflow
    status: active
  - id: guardrails
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion, safety, validation, and delivery gates
    status: active
  - id: tooling
    name: Tooling guidance
    type: doc
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: prompt tooling and project-directory workflow
    status: active
  - id: rlm
    name: Progressive disclosure guidance
    type: doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: compact context and just-in-time loading
    status: active
  - id: command-capabilities-rule
    name: Command capabilities rule
    type: ruleset
    target: docs/references/rules/command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: capability metadata and tests
    status: active
  - id: agent-team-rule
    name: Agent Team orchestration rule
    type: ruleset
    target: docs/references/rules/agent-team-orchestration.md
    relation: constrains
    read_policy: must
    used_for: bounded audit and verification delegation
    status: active
  - id: github-delivery-rule
    name: GitHub PR delivery rule
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: issue-first ready-PR delivery
    status: active
  - id: improve-v1-spec
    name: Auto-improvement v1 specification
    type: spec
    target: docs/specs/0038-auto-improvement-v1/SPEC.md
    relation: informs
    read_policy: must
    used_for: intended benchmark and scoring contracts
    status: active
skills: []
---
# SPEC

## THESIS

Modernize Kit's complete generated prompt system for the GPT-5.6 model family in
one cohesive delivery lane. The migration must make prompts leaner, more
outcome-oriented, more stage-specific, and less prone to unnecessary
clarification or approval pauses while preserving Kit's code-enforced state,
validation, evidence, and delivery guarantees.

The work includes the generated loop model, shared decorators, v2 supervisor,
workflow loop, code review, dispatch, toolbox, frontend profile, verbosity, and
legacy prompt surfaces. It also includes the tests, deterministic benchmarks,
documentation, generated configuration, and capability metadata needed to prove
the migration is complete.

`kit improve` may support that proof only after its real behavior is audited.
The before/after comparison must use identical benchmark definitions and must
not treat a successful command, prompt length, call count, or self-reported
confidence as task-quality evidence.

## CONTEXT

### Source Map

| Source | Finding | Consequence |
|---|---|---|
| SRC-001 | The official GPT-5.6 guide recommends lean prompts, stating an instruction once, explicit autonomy boundaries, and task-specific tool guidance. | Remove repeated lifecycle prose and express goals, constraints, success, and output contracts directly. |
| SRC-002 | Official Codex model documentation supports `gpt-5.6` for Codex CLI execution. | Generated loop configuration and associated help/docs/tests can move from `gpt-5.5` to `gpt-5.6`. |
| SRC-003 | The default `kit improve` suite runs eight `kit capabilities ... --json` tasks and never renders or executes the prompt surfaces in scope. A task command may fail yet pass an assertion, and a failed manifest currently leaves the CLI at exit status 0. | Treat the default suite only as a capability smoke test; repair correctness reporting and add a prompt-system benchmark. |
| SRC-004 | Three pre-change prompt renders are byte-identical across twelve representative surfaces. The corpus is 1,760 lines, 19,299 words, 141,340 characters, and about 35,338 tokens. | The current prompts provide a stable deterministic baseline and a meaningful size comparison. |
| SRC-005 | Loop state transitions, readiness diagnostics, repository validation, evidence requirements, and delivery gates are implemented in Go rather than relying only on prompt prose. | Prompt text may be reduced, but these code-enforced gates and their tests must remain intact. |
| SRC-006 | GitHub issue #54 and branch `GH-54` are the active issue-first delivery lane. | All implementation, validation, reflection, documentation, and delivery evidence remains in one lane and one SPEC. |

### Prompt Surfaces

- Generated loop configuration and model help: `.kit.yaml`,
  `pkg/cli/init_loop_config.go`, `pkg/cli/loop.go`, docs, catalog, and tests.
- Shared decorators: `pkg/cli/skills_prompt.go`, `pkg/cli/subagents.go`, and
  `pkg/cli/prompt_profile.go`.
- V2 supervisor and loop stages: `pkg/cli/spec_v2_prompt.go`,
  `pkg/cli/loop_prompt.go`, and `pkg/cli/loop_prompt_command.go`.
- Other current prompt surfaces: code review, dispatch, built-in toolbox,
  frontend profile, final-response/verbosity guidance, and supporting
  instruction templates.
- Legacy prompt surfaces: brainstorm, plan, tasks, implement, and reflect.
- Benchmark implementation: `internal/improve`, `pkg/cli/improve.go`, and
  `docs/evals/kit-improve`.

### Pre-change Improve Audit

The default suite is reproducible for its narrow purpose: three runs passed all
eight capability assertions and recorded aggregate trace durations of 64 ms,
67 ms, and 64 ms. That stability does not establish prompt quality because the
suite makes no model calls and does not inspect prompt output.

A controlled missing-stdout assertion produced a 7/8 failed manifest and a
useful assertion message, but the CLI still exited successfully. The recorded
failure signature was the first static known failure mode rather than the
actual assertion failure. Candidate validation assigns a score of 1 when a
candidate is merely `proposed`; it does not compare behavioral task results.
`repeat`, held-out eligibility, persona, input prompt, expected behavior, and
seed data do not currently create the scoring rigor described by feature 0038.

Trustworthy current evidence is limited to deterministic command output,
assertion results, changed-file tracking, redacted stdout/stderr, task trace
duration, prompt corpus measurements, and repeated output hashes. Model
latency, token billing, cost, clarification turns, approval turns, tool calls,
and subagent routing are not observable from the default suite.

## CLARIFICATIONS

- The user explicitly placed every identified prompt improvement and benchmark
  repair in one cohesive goal; none is an optional follow-up.
- Repository evidence identifies all affected surfaces and preserves the
  existing code-enforced gates, so no material product decision remains open.
- "Lean" means removing redundant prompt prose while retaining each unique
  behavioral contract and relying on code for gates already enforced in code.
- "Stage-specific" means each loop iteration receives the durable goal/context
  plus only the instructions, success criteria, and output contract needed for
  its current phase.
- Outside an explicit clarification workflow, agents should ask only when a
  material choice cannot be discovered safely from repository or task context.
- Confidence is 99%. The remaining uncertainty concerns measured results, not
  an unresolved requirement, and will be resolved by implementation evidence.

## REQUIREMENTS

- REQ-001: Change every generated Codex loop default and public description
  from `gpt-5.5` to `gpt-5.6`, including migration recognition for previously
  generated defaults.
- REQ-002: Reduce shared Skills, subagent, and profile decorators to concise,
  non-duplicative routing and boundary guidance.
- REQ-003: Rewrite the v2 supervisor around the user goal, grounded context,
  constraints, approval boundaries, success criteria, and output contract.
- REQ-004: Make v2 workflow-loop prompts phase-specific and avoid reinjecting
  the complete lifecycle contract on every fresh agent subprocess.
- REQ-005: Limit clarification requests outside explicit clarification flows to
  material, non-discoverable decisions; permit safe repository discovery and
  in-scope action without unnecessary approval pauses.
- REQ-006: Modernize code-review, dispatch, built-in toolbox, frontend,
  verbosity/final-response, and legacy prompt surfaces using the same concise
  outcome-oriented conventions.
- REQ-007: Preserve state-transition diagnostics, validation, evidence,
  delivery gates, dirty-work protection, and other code-enforced safety rules.
- REQ-008: Do not add Programmatic Tool Calling, persisted reasoning, Pro mode,
  or API `text.verbosity` as instructions to Codex CLI prompts.
- REQ-009: Make `kit improve run` treat command failures and failed manifests as
  failures, emit cause-specific traces/signatures, and report benchmark facts
  rather than implying unsupported model-quality scores.
- REQ-010: Add a deterministic prompt-system benchmark and representative loop
  end-to-end coverage that exercise the actual generated prompts.
- REQ-011: Record task/assertion completeness, prompt size, timing where
  observable, benchmark provenance, and the explicit observability status of
  clarification, approval, validation/evidence, and routing metrics.
- REQ-012: Run the identical benchmark definition against the frozen pre-change
  binary and the post-change binary and report the comparison.
- REQ-013: Update tests, goldens, documentation, generated config examples,
  workflows, and capability metadata for all changed behavior.
- REQ-014: Complete implementation, validation, reflection, and ready-PR
  delivery through this canonical SPEC and GitHub lane #54.

### Non-goals

- No API-only reasoning or tool-calling feature is represented as a Codex CLI
  prompt instruction.
- No removal or weakening of code-enforced workflow, repository, safety,
  validation, evidence, or delivery gates.
- No attempt to measure live model quality, latency, or billing without an
  observable model call and reproducible data source.
- No general-purpose evaluation framework, statistical model judge, or broad
  rewrite of the feature-0038 improvement pipeline beyond what trustworthy
  prompt benchmarking requires.
- No separate issues, branches, PRs, worktrees, or partial-delivery stopping
  points for individual prompt surfaces.

## ASSUMPTIONS

- The official documentation and live repository are authoritative for model
  support and current prompt behavior.
- `gpt-5.6` is available to the generated `codex exec` invocation; this change
  does not add an API transport or model-specific API parameters.
- Deterministic prompt assertions and fake-agent loop scenarios are the correct
  local evidence for behavior that cannot be measured without model calls.
- The existing default improve suite remains useful as a capability smoke test
  when its limited interpretation is documented.
- The frozen `/tmp/kit-gh54-baseline` binary is the pre-change executable for
  the final identical-definition comparison; its SHA-256 will be recorded.
- If a size threshold conflicts with a required contract, correctness wins and
  the threshold is adjusted transparently rather than deleting the contract.

## ACCEPTANCE CRITERIA

- AC-001: Generated init config, loop help, docs, capability metadata, and tests
  use `gpt-5.6`; explicit old generated `gpt-5.5` defaults migrate safely while
  user-customized model values remain unchanged.
- AC-002: Prepared prompts contain concise Skills, subagent, and profile
  decorators with no repeated lifecycle or generic coding guidance.
- AC-003: The v2 supervisor clearly contains goal, context, constraints,
  approval boundaries, success criteria, and output contract, with each unique
  instruction stated once.
- AC-004: Captured loop-agent stdin for clarify, implement, validate, reflect,
  and deliver contains phase-specific work and omits unrelated phase contracts.
- AC-005: Non-clarification prompts ask only about material non-discoverable
  ambiguity and explicitly allow safe in-scope discovery/action without routine
  approval; explicit clarification prompts retain their question workflow.
- AC-006: Code-review, dispatch, toolbox, frontend, verbosity/final-response,
  brainstorm, plan, tasks, implement, and reflect prompts follow the concise
  outcome/success/output pattern and pass focused goldens/contracts.
- AC-007: Existing code-enforced state, readiness, validation, evidence,
  repository-safety, and delivery diagnostics continue to pass unchanged or
  stronger tests.
- AC-008: Generated prompts do not introduce instructions for Programmatic Tool
  Calling, persisted reasoning, Pro mode, or API `text.verbosity`.
- AC-009: An improve task with a nonzero command result fails; a failed run
  returns nonzero; failure evidence identifies the actual command/assertion;
  default-suite documentation labels it as a capability smoke test.
- AC-010: A committed prompt-system suite renders the changed prompt surfaces,
  runs repeated deterministic checks, records benchmark/binary provenance,
  output assertion completeness, prompt lines/words/bytes/estimated tokens,
  and command duration.
- AC-011: The final report compares the frozen baseline and candidate using the
  identical committed suite definition and records all requested metric fields,
  marking unavailable live-model metrics as unobservable rather than inferred.
- AC-012: The representative prompt corpus is materially smaller in total
  estimated tokens and words, with a target reduction of at least 30%, while no
  required-output or representative scenario assertion regresses.
- AC-013: Deterministic scenarios verify expected clarification, approval/
  autonomy, validation/evidence, and subagent/tool-routing contracts for both
  positive and negative cases.
- AC-014: `.kit.yaml`, workflow/docs references, capability catalog, goldens,
  unit tests, and end-to-end tests agree with the final behavior.
- AC-015: Formatting, vetting, unit/integration tests, build, lint where
  available, diff checks, verification-agent review, and literal PR checks are
  recorded; the ready PR closes issue #54.

## IMPLEMENTATION PLAN

### 1. Make the benchmark truthful before using it

- Repair improve runner failure propagation and cause-specific diagnostics.
- Add benchmark provenance and deterministic output metrics.
- Commit a prompt-system suite/fixture that invokes actual Kit prompt commands.
- Add negative controls proving command and assertion failures are
  distinguishable.
- Run the new suite against `/tmp/kit-gh54-baseline` before changing prompt
  implementations.

### 2. Migrate model configuration and shared preparation

- Change generated defaults and public metadata to `gpt-5.6`.
- Preserve refresh compatibility for the prior generated `gpt-5.5` block and
  avoid overwriting custom loop agents.
- Thin Skills, profile, and subagent decorators at their shared composition
  points so all downstream prompts benefit consistently.

### 3. Redesign v2 supervision and loop phases

- Factor a compact durable supervisor contract around goal, grounded context,
  constraints, approval boundaries, success criteria, and output.
- Generate a phase contract for only the current loop phase.
- Retain SPEC ownership and code-enforced phase transition validation.
- Add fake-agent capture tests for actual per-stage subprocess input.

### 4. Migrate remaining prompt surfaces

- Rewrite code-review and dispatch around actionable findings/outcomes.
- Replace toolbox prompts with concise short, planning, and instruction modes.
- Reduce frontend and verbosity/final-response guidance to unique constraints.
- Modernize legacy prompts without changing their explicit legacy command
  status or artifact contracts.

### 5. Lock behavior and documentation

- Update raw and final-prepared prompt goldens, focused contracts, migration
  tests, capability tests, E2E scenarios, and workflow path triggers.
- Update model/config/docs references and benchmark limitations.
- Run the full repository validation matrix and a read-only verification-agent
  audit, then repair any valid findings.

### 6. Compare, reflect, and deliver

- Build the candidate binary and rerun the exact committed prompt-system suite.
- Produce a before/after table for correctness/completeness, clarification,
  approval/autonomy, validation/evidence, size, timing/cost observability, and
  routing.
- Record reflection, remaining risks, and confidence in this SPEC.
- Stage explicit paths, commit, push `GH-54`, open a ready PR assigned to the
  repository owner, and report checks literally.

### Risks and Rollback

- Risk: removing prose can hide a unique contract. Mitigation: inventory
  contracts before rewriting and enforce final prepared prompt goldens plus
  representative E2E scenarios.
- Risk: benchmark thresholds reward size over correctness. Mitigation: required
  behavior assertions are hard gates; size is reported separately.
- Risk: old generated `gpt-5.5` blocks stop refreshing. Mitigation: explicit
  migration fixture tests for generated-old, generated-current, and custom
  configurations.
- Risk: a prompt shared by current and legacy commands changes unintended
  behavior. Mitigation: surface-specific goldens and command tests.
- Rollback: revert the single GH-54 commit/PR. No persistent data migration or
  external service mutation is involved.

## TASK CHECKLIST

| Task | Acceptance | Status | Evidence |
|---|---|---|---|
| T-001 Complete repository, workflow, prompt-surface, and delivery recon | AC-015 | complete | Issue #54; branch `GH-54`; source map |
| T-002 Audit improve execution, scoring, traces, stability, and failure discrimination | AC-009, AC-011 | complete | Three baseline runs and controlled failure summarized in Context |
| T-003 Capture repeated pre-change prompt corpus and size/hash baseline | AC-010, AC-011, AC-012 | complete | Baseline table in Evidence |
| T-004 Write canonical ready SPEC for the complete goal | AC-001 through AC-015 | complete | This document |
| T-005 Repair improve runner semantics and add prompt benchmark/provenance/metrics | AC-009, AC-010, AC-011 | complete | Runner/CLI tests; prompt-system baseline run `20260710T123420.394230000Z-2c0dcc` |
| T-006 Migrate loop model and generated-default compatibility | AC-001, AC-014 | complete | Config/capability/docs plus generated-old/current/custom migration tests |
| T-007 Thin shared Skills, subagent, and profile decorators | AC-002, AC-012 | complete | Compact-layer tests, prepared prompt goldens, and benchmark metrics |
| T-008 Rewrite v2 supervisor and phase-specific loop prompts | AC-003, AC-004, AC-005, AC-007 | complete | Supervisor golden and six-phase fake-agent stdin capture E2E |
| T-009 Modernize current and legacy prompt surfaces | AC-005, AC-006, AC-008 | complete | Focused contracts, goldens, and prompt-system suite |
| T-010 Update docs, workflows, capability metadata, examples, and generated config | AC-001, AC-014 | complete | Documentation/config/catalog/workflow diff review |
| T-011 Run identical post-change benchmark and full validation | AC-007, AC-009 through AC-015 | complete | Runs `20260710T132322.293199000Z-ce029f` and `20260710T132322.832544000Z-73ff54`; full validation below |
| T-012 Complete reflection and ready-PR delivery | AC-011, AC-012, AC-015 | complete | Implementation commit `cf207b0`; ready PR #55; initial checks recorded pending and terminal state reported in handoff |

## VALIDATION MAP

| Acceptance | Validation | Evidence target |
|---|---|---|
| AC-001 | Init/refresh migration tests; `rg 'gpt-5\\.5|gpt-5\\.6'`; capabilities tests | Test output and final diff |
| AC-002 | Final prepared-prompt goldens; decorator unit tests; prompt corpus metrics | Golden diffs and benchmark JSON |
| AC-003 | V2 supervisor golden and required/forbidden contract assertions | `pkg/cli` tests and golden |
| AC-004 | Fake `codex`/agent E2E capturing stdin for five phases | Stage capture test output |
| AC-005 | Positive and negative clarification/autonomy prompt assertions | Prompt suite and focused tests |
| AC-006 | Surface goldens/contracts for current, frontend, toolbox, and legacy prompts | `pkg/cli` tests and prompt suite |
| AC-007 | Existing loop transition/validation tests plus `go test ./...` | Full test log |
| AC-008 | Forbidden-string scan over rendered benchmark corpus | Prompt-suite assertions |
| AC-009 | Improve command-failure, assertion-failure, and CLI-exit negative tests | `internal/improve` and CLI tests |
| AC-010 | `kit improve run --suite prompt-system --json` against both binaries (suite-defined repeat 3) | Baseline/candidate run manifests |
| AC-011 | Same suite-definition hash and before/after report with observability fields | SPEC Evidence and PR body |
| AC-012 | Aggregate and per-surface words/tokens plus all correctness gates | Comparison table |
| AC-013 | Deterministic scenario assertions for clarification, approval/autonomy, evidence, and routing | Prompt suite and E2E output |
| AC-014 | Docs/config/catalog/golden review and relevant workflow tests | Final diff and test output |
| AC-015 | `gofmt`, `go vet ./...`, `go test ./...`, `go build ./...`, `make lint`, `git diff --check`, verifier audit, ready PR checks | Final validation and delivery evidence |

The pre-change and post-change benchmark commands will use the same committed
suite and fixture. The only intended variable is `--kit-binary`. Binary and
suite SHA-256 values must appear in the run manifest. If the implementation
uses a different final flag spelling, this invariant remains mandatory.

## REFLECTION NOTES

### Prompt and workflow result

- Repeated lifecycle walkthroughs, generic coding advice, universal
  clarification batches, blanket brevity/TL;DR rules, and duplicated Skills,
  profile, and subagent guidance were removed. Unique authority, approval,
  repository safety, validation/evidence, delivery, and final-output contracts
  remain either in their narrow prompt surface or in code-enforced gates.
- The v2 supervisor now defines the durable goal and boundaries once. Captured
  loop-agent stdin proves all six phases receive their own outcome, inputs,
  actions, success checks, and output contract without the full phase table or
  unrelated phase instructions.
- No deterministic required-output behavior regressed. The final expanded
  suite passes 45/45 task runs and 345/345 assertions; the full Go test, vet,
  and build matrix also passes.
- The representative corpus exceeds the 30% target without using size as an
  acceptance score: words fell 64.8% and estimated tokens fell 62.7%, while
  output-contract assertions remained hard gates. Model capability output grew
  slightly because it now carries the correct GPT-5.6 metadata.

### Improve trust boundary

- Trustworthy: command success/failure, required substring and absence checks,
  changed-file checks, cause-specific trace evidence, suite/runner/evaluated-
  binary hashes, repeated-output determinism, lines/words/bytes/estimated-token
  counts, and local command duration.
- Useful proxies, not live behavior: deterministic prompt wording for
  clarification, approval/autonomy, validation/evidence, and routing. Focused
  tests and captured fake-agent stdin prove the generated contract, not how a
  hosted model will follow it.
- Unobservable because the suite makes zero model calls: actual clarification
  turns, approval pauses, model task quality, provider latency, billed tokens or
  cost, tool selection, and subagent routing. Local duration is not model
  latency. `kit improve validate` now validates candidate metadata only and
  returns score 0; it makes no behavioral acceptance claim.

### Independent verification

Three read-only audit lanes were used: two initial improve-validity/coverage
audits and one final integrated verifier. The final verifier found omitted Warp
and brainstorm benchmark coverage, stale clarification and brevity copy,
misleading improve candidate/PR claims, mislabeled benchmark coverage, and dead
prompt helpers. Each confirmed finding was repaired and retested. Its final
verdict reported no unresolved implementation defect; it explicitly retained
the live-model observability limits above.

### Remaining risk

The residual risk is model-following behavior that deterministic local tests
cannot observe. The mitigation is the lean contract structure, hard code-level
workflow gates, representative prompt assertions, and literal CI/PR review.
There is no data migration or persistent external runtime change; rollback is
the single GH-54 change set.

## DOCUMENTATION UPDATES

| Document | Required update | Status |
|---|---|---|
| `.kit.yaml` | Generated loop model example `gpt-5.6` | updated and verified |
| `docs/workflows.md` | Loop model, stage-specific behavior, and benchmark semantics | updated and verified |
| `docs/CONSTITUTION.md` | Generated model default and material-clarification policy | updated and verified |
| `docs/commands.md` and `README.md` | Prompt-system benchmark and concise prompt behavior | updated and verified |
| `docs/evals/kit-improve/README.md` | Smoke-suite limits, trusted metrics, and blind spots | updated and verified |
| Prompt-system suite/task/fixture/schema definitions | Reproducible benchmark contract | added and verified |
| `pkg/cli/capabilities_catalog.go` | Model, prompt, and improve behavior metadata | updated and tested |
| Golden files and test fixtures | Final generated prompt/config outputs | updated and tested |
| `.github/workflows/kit-improve-*.yml` | Prompt-system coverage and truthful failure behavior | updated; local commands verified |
| `docs/specs/0040-prompt-system-migration/SPEC.md` | Validation, reflection, comparison, delivery evidence | complete through ready-PR creation; live checks remain external state |

## DELIVERY DECISION

- Delivery intent: create/update issue, branch, commit, push, and ready PR.
- GitHub issue: #54, `Modernize generated prompts for GPT-5.6`.
- Branch: `GH-54`, created from the fetched `origin/main` head.
- Planned commit/PR title:
  `feat(GH-54): :sparkles: modernize generated prompts for GPT-5.6`.
- Planned PR ticket line: `Closes #54`.
- Assignee: `jamesonstone`.
- Implementation commit: `cf207b0f6017d93d98d69018872a5545a2855e94`.
- Ready PR: #55,
  `https://github.com/jamesonstone/kit/pull/55`; open, assigned to
  `jamesonstone`, head `GH-54`, base `main`, and `isDraft: false`.
- Initial literal checks after PR creation: `Assign configured maintainers` —
  pending; `validate` — pending. Checks are mutable external state, so their
  terminal state is reported literally in the final handoff after this SPEC
  update is pushed.

## EVIDENCE

### Improve Validity Baseline

| Run | Result | Passed tasks | Failed tasks | Aggregate trace duration |
|---|---|---:|---:|---:|
| `20260710T121201.977941000Z-0802c4` | passed | 8 | 0 | 64 ms |
| `20260710T121202.066270000Z-f53099` | passed | 8 | 0 | 67 ms |
| `20260710T121202.159713000Z-77edf3` | passed | 8 | 0 | 64 ms |
| Controlled missing-output assertion | failed manifest | 7 | 1 | not used as quality score |

All three repeated prompt-render directories produced identical SHA-256 values
for all twelve files. The current improve suite made zero model calls, so model
latency, cost, conversational turns, and live tool/subagent routing are
unobservable in this baseline.

### Pre-change Prompt Corpus

Estimated tokens are `ceil(characters / 4)` per prompt and are a transparent
size proxy, not provider billing tokens.

| Surface | Lines | Words | Characters | Estimated tokens |
|---|---:|---:|---:|---:|
| Code review | 133 | 1,077 | 7,428 | 1,857 |
| Dispatch | 77 | 716 | 5,035 | 1,259 |
| Frontend code review | 148 | 1,469 | 10,226 | 2,557 |
| Legacy implement | 200 | 2,545 | 18,656 | 4,664 |
| Legacy plan | 187 | 2,121 | 15,872 | 3,968 |
| Legacy reflect | 215 | 2,314 | 16,728 | 4,182 |
| Legacy tasks | 172 | 1,876 | 13,752 | 3,438 |
| Loop prompt | 114 | 1,572 | 11,545 | 2,887 |
| V2 SPEC supervisor | 433 | 5,185 | 39,130 | 9,783 |
| Toolbox instructions | 66 | 232 | 1,622 | 406 |
| Toolbox long | 14 | 158 | 1,104 | 276 |
| Toolbox short | 1 | 34 | 242 | 61 |
| **Total** | **1,760** | **19,299** | **141,340** | **35,338** |

Detailed local baseline manifests and prompt snapshots are under the ignored
`.kit/improve/gh54-baseline/` directory. The final durable comparison and binary
and suite hashes will be recorded here and in the PR before delivery.

### Final Identical-Definition Comparison

Both final runs used suite definition SHA-256
`714bacb7f28dfa181d761cf44394de353bed0cc64d98d54f9ba3af765ce618bd`
and runner SHA-256
`15197ff11287a7e01cd765b94b5ac859ee892f162c24bc69021a7a4bea470e58`.
The only evaluated-program variable was the Kit binary:

- Frozen baseline: `/tmp/kit-gh54-baseline`, SHA-256
  `167dea4d29fee308cd416da66fb595a774f9827a8a09178f937ff542151328d6`,
  run `20260710T132322.293199000Z-ce029f`.
- Final candidate: `/tmp/kit-gh54-final`, SHA-256
  `15197ff11287a7e01cd765b94b5ac859ee892f162c24bc69021a7a4bea470e58`,
  run `20260710T132322.832544000Z-73ff54`.

| Metric | Frozen baseline | Final candidate | Change |
|---|---:|---:|---:|
| Task success | 36/45 (80.0%) | 45/45 (100%) | +9 passing runs |
| Required-output assertions | 327/345 (94.78%) | 345/345 (100%) | +18 passing assertions |
| Repeated-output determinism | 15/15 (100%) | 15/15 (100%) | unchanged |
| Prompt/output lines | 6,195 | 2,640 | -57.4% |
| Prompt/output words | 63,411 | 22,323 | -64.8% |
| Prompt/output bytes | 498,867 | 186,144 | -62.7% |
| Estimated tokens | 124,737 | 46,548 | -62.7% |
| Aggregate local command duration | 300 ms | 294 ms | -2.0%; timing noise only |
| Provider latency and cost | unobservable | unobservable | zero model calls |
| Live clarification/approval turns | unobservable | unobservable | deterministic contract proxy only |
| Live tool/subagent routing | unobservable | unobservable | deterministic contract proxy only |

The baseline's nine failed task runs are three repeats each of the new
brainstorm material-ambiguity contract, the compact Warp plan contract, and the
GPT-5.6 capability assertion. This is expected and demonstrates that the
strengthened suite distinguishes the targeted pre-change failures. The final
candidate passes every task and assertion.

### Per-Surface Estimated Tokens

Values are the first deterministic repeat; the other two repeats are hash-
identical for every surface.

| Surface | Baseline | Candidate | Reduction |
|---|---:|---:|---:|
| Code review | 1,846 | 660 | 64.2% |
| Dispatch | 1,282 | 634 | 50.5% |
| Frontend code review | 2,550 | 909 | 64.4% |
| Legacy brainstorm | 3,837 | 1,297 | 66.2% |
| Legacy implement | 4,012 | 1,594 | 60.3% |
| Legacy plan | 3,563 | 1,311 | 63.2% |
| Legacy Warp plan | 3,575 | 1,355 | 62.1% |
| Legacy reflect | 3,618 | 1,682 | 53.5% |
| Legacy tasks | 2,975 | 1,316 | 55.8% |
| Feature loop prompt | 2,914 | 1,154 | 60.4% |
| Model capability | 524 | 562 | -7.3% (correct metadata added) |
| Toolbox instructions | 406 | 276 | 32.0% |
| Toolbox long | 276 | 196 | 29.0% |
| Toolbox short | 61 | 54 | 11.5% |
| V2 supervisor | 10,140 | 2,516 | 75.2% |

### Final Validation

- `go test ./...` — passed for all packages.
- `go vet ./...` — passed.
- `go build ./...` — passed.
- `git diff --check` — passed.
- JSON parsing for every `docs/evals/kit-improve/schemas/*.json` — passed.
- `kit check prompt-system-migration` — passed with no findings.
- Final default capability smoke run
  `20260710T132356.880856000Z-6b513c` — passed 8/8 tasks and 16/16
  assertions.
- `make lint` — failed on 44 repository-wide findings: 23 `errcheck`, 6
  `staticcheck`, and 15 `unused`. The unchanged `origin/main` snapshot fails the
  same command with 59 findings, so this branch introduces no net lint debt and
  removes 15 findings; the remaining unrelated debt is outside GH-54.
- `kit check --project` — failed on the same 13 invalid relation/status errors
  in historical feature 0038 and two warnings present when run with the frozen
  baseline binary. The new feature 0040 passes its dedicated validator.
- Read-only verifier — no unresolved defect; focused tests, vet, build, diff
  check, rendered forbidden-instruction review, and AC audit passed.
