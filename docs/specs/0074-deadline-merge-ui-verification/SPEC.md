---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0074"
  slug: deadline-merge-ui-verification
  dir: 0074-deadline-merge-ui-verification
relationships:
  - type: builds_on
    target: 0070-deadline-mode
    note: Adds merge/deployment-wave UI deferral to the existing deadline-mode testing budget without weakening required post-deployment production-suite checks.
  - type: related_to
    target: 0053-testing-and-environment-validation
    note: Pointer-loads the same explicit supersession so UI/browser walkthroughs wait until every authorized result is delivered.
  - type: related_to
    target: 0061-authorized-coding-agent-merge-autonomy
    note: Merge waves keep recording merge, workflow, and deployment evidence after each transition while deferring UI walkthroughs.
references:
  - id: deadline-mode
    name: Deadline mode
    type: rule
    target: docs/references/rules/deadline-mode.md
    relation: implements
    read_policy: must
    used_for: canonical urgency contract for skipping interleaved UI verification
    status: active
  - id: testing-rule
    name: Testing and environment validation
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: constrains
    read_policy: must
    used_for: existing browser and post-deployment requirements this feature defers without cancelling
    status: active
  - id: merge-rule
    name: GitHub PR merge
    type: rule
    target: docs/references/rules/github-pr-merge.md
    relation: constrains
    read_policy: must
    used_for: merge-wave continuation while UI walkthroughs wait for every delivered result
    status: active
  - id: issue
    name: Defer UI verification until merge/deploy results are delivered
    type: external
    target: https://github.com/jamesonstone/kit/issues/188
    relation: supports
    read_policy: must
    used_for: original ask and acceptance criteria
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Under an explicit deadline, keep authorized merge and deployment work moving.
Skip UI and browser walkthroughs until every result in that wave is delivered,
then run one final UI verification. Do not cancel that final check or weaken
required post-deployment production-suite evidence.

## CONTEXT

- `deadline-mode` already defers premature browser automation until the
  workflow exists, and it already runs full browser/operator acceptance at
  final activation readiness. That is not the same as a merge/deploy wave:
  agents can still stop after each merge or deployment to walk the UI.
- The user asked to update the urgency rules for merge/deployment with this
  sentiment: continue, skip UI verifications until all results are delivered,
  then do one final UI verification.
- Required post-deployment production-suite checks in
  `testing-and-environment-validation.md`'s `### Local And Production Execution`
  remain mandatory. This feature defers UI/browser walkthroughs, not health,
  hosted-workflow, or production-suite evidence.
- Topology: single-lane, because the change is tightly coupled policy across
  three rules and string-locked tests and needs continuous design judgment on
  the defer-versus-never-weaken boundary.

## REQUIREMENTS

- Under active deadline mode, continue authorized merge and deployment work
  without interleaved UI, browser, or operator walkthrough verification after
  each result.
- After every result in the authorized set is delivered, run one final UI
  verification of the delivered system.
- Report deferred UI checks as `PARTIAL` or `SKIPPED`, never `PASS`. If the
  wave has no UI surface, report the final UI check as `NOT_APPLICABLE`.
- Keep required post-deployment production-suite checks, merge
  authorization, infrastructure approval, and the final UI verification
  itself unweakened.
- Add non-contradictory pointers from `testing-and-environment-validation` and
  `github-pr-merge` so merge-orchestration and validation agents find the
  same contract.
- No Hard Gate, CONSTITUTION change, or immutable instruction version bump.
- Observable acceptance: focused tests lock the deferral and the one-final
  check; project validation passes.

## ACCEPTED PLAN

1. Record this accepted plan in `0074` before implementation.
2. Extend `deadline-mode` priority order, stop-doing, continue-doing, testing
   budget, anti-patterns, verification, and examples with the merge/deploy UI
   deferral.
3. Point `testing-and-environment-validation` and `github-pr-merge` at that
   contract without silently contradicting required production-suite checks.
4. Lock the wording in focused deadline-mode, merge, and testing-rule tests.
5. Hand-add only this feature's `PROJECT_PROGRESS_SUMMARY.md` row after
   reverting `kit spec` mtime corruption.
6. Validate, curate repository memory, and open one ready pull request for
   issue #188.

## DECISIONS

- Canonical home is `deadline-mode`. Merge/deployment urgency is already that
  opt-in ruleset; do not make UI deferral the default for every merge.
- `github-pr-merge` and `testing-and-environment-validation` get pointers, not
  a second competing contract.
- UI/browser/operator walkthroughs are the deferred class. Hosted workflow,
  deployment/runtime, and required production-suite checks still record after
  each transition.
- If the unit of work is the merge or deployment wave itself, the one final UI
  verification is the activation-readiness UI check; do not run a second
  interleaved UI pass.
- Topology: single-lane, because tightly coupled, high-overlap, and requiring
  continuous design judgment.

## DISCOVERIES

- `kit spec` rewrote unrelated `PROJECT_PROGRESS_SUMMARY.md` created dates
  from filesystem mtime again. That churn was reverted; only the `0074` row
  and summary subsection are added by hand.

## VALIDATION

- `go test ./pkg/cli -run 'TestDeadlineMode|TestGitHubPRMerge|TestTestingAndEnvironmentValidation'` and `go test ./internal/templates -run 'TestMergeOrchestrationRulesStayContextAwareAndConcise|TestInstructionTemplatesRouteDeadlineMode|TestImplementationDeliverySelectsDeadlineMode'` passed.
- `go test ./...` passed, including `TestDeadlineModeRegistryRulesetIsValid`, `TestDeadlineModeIsIntegratedWithRelatedRules`, `TestGitHubPRMergeRulesetIsValid`, `TestGitHubPRMergeRulesetCoversRequiredScenarios/deadline-mode_UI_deferral`, and `TestTestingAndEnvironmentValidationRegistryRulesetIsValid`.
- `make fmt`, `go vet ./...`, and `make build` passed.
- `golangci-lint run --new-from-rev=origin/main ./...` reported 0 issues.
- `git diff --check` reported no whitespace errors.
- `gitleaks detect --source . --no-git` found no leaks.
- A grep for `deadline-mode` across `CLAUDE.md`, `AGENTS.md`, `.github/copilot-instructions.md`, `docs/agents/GUARDRAILS.md`, `docs/CONSTITUTION.md`, and `internal/instructions/versions` found zero matches.
- `kit check 0074-deadline-merge-ui-verification` and `kit check --project` passed.
- Browser, live-integration, deployment, and production tests: `NOT_APPLICABLE` for this ruleset change.

## OUTCOME

- Extended `deadline-mode` so urgent merge and deployment work continues, UI/browser walkthroughs wait until every authorized result is delivered, and one final UI verification then runs. Required post-deployment production-suite checks remain mandatory.
- Pointed `testing-and-environment-validation` and `github-pr-merge` at the same contract without a second competing rule.
- Locked the wording in focused deadline-mode, merge, testing-rule, and merge-orchestration tests.
- Reverted `kit spec` mtime corruption in `PROJECT_PROGRESS_SUMMARY.md` and hand-added only the `0074` row and subsection.
- Delivery remains at the authorized issue #188 branch (`GH-188`) and ready-pull-request boundary. Merge is not authorized.

## REPOSITORY MEMORY

- **Decision:** created
- **Rationale:** This follow-up records a user-stated merge/deployment
  urgency sentiment that code cannot preserve: skip interleaved UI checks
  until every result lands, then run one final UI verification, without
  weakening required production-suite evidence.
- **Artifacts:** `docs/specs/0074-deadline-merge-ui-verification/SPEC.md`,
  `docs/references/rules/deadline-mode.md`,
  `docs/references/rules/testing-and-environment-validation.md`,
  `docs/references/rules/github-pr-merge.md`
