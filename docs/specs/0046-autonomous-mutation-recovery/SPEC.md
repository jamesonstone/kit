---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0046
  slug: autonomous-mutation-recovery
  dir: 0046-autonomous-mutation-recovery
references:
  - id: safety-guardrails
    name: Safety guardrails
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: failure recovery and destructive-action boundaries
    status: active
  - id: github-pr-delivery
    name: GitHub PR delivery
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: issue branch commit push and PR recovery behavior
    status: active
  - id: work-lane-gating
    name: Work lane gating
    type: ruleset
    target: docs/references/rules/work-lane-gating.md
    relation: constrains
    read_policy: must
    used_for: safe recovery from an ungated or mismatched implementation lane
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Make Kit coding agents own routine failure recovery through completion instead of stopping after a failed in-scope mutation and asking the user to authorize a compatible retry such as authenticated `gh`.

## CONTEXT

- The current `safety-guardrails` ruleset requires agents to stop on every failure, forbids mutation retries, and waits for user instruction.
- Downstream agents consequently emit messages such as “Reply `retry with gh`,” even when the user already authorized the goal and the alternative tool would target the same repository, branch, issue, or pull request.
- The current rule conflates a changed authorization boundary with a changed implementation path. A compatible retry that preserves repository, target, scope, and intended effect should remain inside the original authorization.
- Increased autonomy must not permit blind repetition, scope expansion, destructive cleanup, protected-branch writes, force operations, review bypass, identity substitution, secret exposure, autonomous merge, or deletion without the required user decision.

## REQUIREMENTS

- Agents must own the requested outcome, diagnose routine failures, revise the approach, and continue until the goal is complete or a genuine external blocker remains.
- After a failed mutation, agents must capture the exact error, inspect current state with read-only checks, identify the cause, choose a safe compatible recovery path, retry within the original scope, and verify the resulting state.
- Switching from one supported mutation path to another, including an authenticated GitHub connector to authenticated `gh` or the reverse, must not require user permission when repository, target, scope, intended effect, and human identity are unchanged.
- Agents must not blindly repeat the same failed command. Each retry must follow diagnosis or a material change in evidence, state, parameters, or tool path.
- Existing prohibitions on force-push, protected-branch mutation, review bypass, autonomous merge, branch deletion, repository-setting changes, secret changes, destructive resets, broad staging, and agent attribution remain in force.
- Agents must ask permission before large-scale deletion or deleting sensitive files. Deletion scope and sensitivity must be resolved with read-only inspection before asking.
- Missing credentials, ambiguous identity or target, conflicting user-owned changes, unavailable external dependencies, or required external authorization are blockers that may require user input; they must not be framed as routine permission-to-retry requests.
- Generated `docs/agents/GUARDRAILS.md`, its template source, the downstream `safety-guardrails` registry ruleset, and GitHub delivery guidance must express one consistent autonomy boundary.
- Focused tests must prevent regression to blanket stop-and-ask behavior and must require the compatible-tool recovery rule plus the deletion permission boundary.
- Observable acceptance: generated and checked-in guardrails are aligned; ruleset validation and focused autonomy assertions pass; full formatting, vet, tests, build, prompt regression, feature validation, and project validation succeed.
- Non-goals: autonomous merge, permissionless large-scale or sensitive deletion, silent error suppression, unbounded blind retries, force-based recovery, bypassing repository delivery gates, changing user intent, or treating missing credentials and ambiguous targets as recoverable guesses.

## ACCEPTED PLAN

1. Replace blanket stop-on-failure language in `safety-guardrails` with a diagnose, reconcile, retry, and verify workflow bounded by the original authorization.
2. Align GitHub delivery and work-lane failure references so recoverable lane, fetch, validation, push, and PR-tool failures continue autonomously without allowing stale-base branching, force operations, or duplicate delivery state.
3. Update the generated and checked-in agent guardrails to remove routine approval prompts, state the deletion permission boundary, and distinguish genuine blockers from retry authorization.
4. Add focused ruleset and template assertions, run the complete validation surface, and curate the spec and project-wide contract to the validated outcome.

## DECISIONS

- Authorization is scoped by intended effect and target, not by the specific supported tool first attempted. A connector-to-`gh` retry is therefore routine recovery when its mutation semantics are unchanged.
- Autonomous recovery requires new evidence or a revised path; blind repetition remains prohibited.
- Permission is reserved for large-scale deletion and deletion of sensitive files. Requests for missing credentials, identity confirmation, or target clarification are information/blocker handling rather than permission for routine work.
- Existing GitHub and destructive-action prohibitions remain hard boundaries and are not weakened by this feature.

## DISCOVERIES

- `docs/references/rules/safety-guardrails.md` is the direct source of the blanket stop-and-wait behavior.
- `docs/references/rules/github-pr-delivery.md` repeats that behavior for implementation failures, base refresh failures, and push rejection handling.
- `docs/references/rules/work-lane-gating.md` retained a dependent instruction to leave changes in place and await user instruction after a gate violation; aligning it prevents the obsolete stop-and-wait behavior from re-entering through a continuously coupled ruleset.
- The generated `docs/agents/GUARDRAILS.md` source also requires explicit approval before staging or committing, which conflicts with goal-owned execution after the delivery contract is resolved.
- A focused generated-file assertion is sufficient to keep the V3 guardrail template and checked-in `docs/agents/GUARDRAILS.md` byte-aligned while separate ruleset tests protect the downstream recovery contract.
- The pre-implementation spec check rejected the legacy `governs` reference relation. Replacing it with supported V3 `constrains` metadata restored validation without weakening the relationship.
- Feature number `0046` avoids the `0044` and `0045` directories already present on the open `GH-62` lane while allowing this independent branch to remain based on `origin/main`.

## VALIDATION

- `go test ./pkg/cli -run '^(TestSafetyGuardrailsRegistryRulesetRequiresAutonomousRecovery|TestGitHubPRDeliveryRulesetUsesAutonomousRecovery|TestWorkLaneGatingRulesetUsesAutonomousRecovery)$' -count=1` — passed.
- `go test ./internal/templates -run '^TestMemoryGuardrailsPreserveAutonomousRecovery$' -count=1` — passed and confirmed the generated and checked-in V3 guardrails are byte-aligned.
- `make fmt` and `make vet` — passed.
- `go test ./... -count=1` — passed across all packages.
- `go test -race ./internal/templates ./pkg/cli -count=1` — passed.
- `make build` — passed and rebuilt `bin/kit` at `v1.0.91`.
- `golangci-lint run --new-from-rev=origin/main ./...` — passed with `0 issues`.
- `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit` run `20260719T122309.583174000Z-041f36` — passed all 45 traces after final ruleset alignment.
- `./bin/kit check autonomous-mutation-recovery` — passed.
- Initial `./bin/kit check --project` correctly blocked because the new feature was not yet represented in `docs/PROJECT_PROGRESS_SUMMARY.md`; `./bin/kit complete autonomous-mutation-recovery` marked the feature complete and refreshed the summary.
- Final `./bin/kit check --project` passed with 15 pre-existing compatibility warnings and a non-blocking project-refresh advisory.
- `git diff --check` — passed.

## OUTCOME

- Kit-managed agents now diagnose and recover routine in-scope failures autonomously, may switch to authenticated `gh` without requesting retry permission when the authorized mutation is unchanged, and continue until completion or a genuine external blocker.
- Blind repetition, force-based recovery, protected operations, autonomous merge, destructive cleanup, identity substitution, and duplicate delivery-state creation remain prohibited.
- Permission is reserved for large-scale deletion and deletion of sensitive files; missing credentials or ambiguous identity, target, ownership, or external authorization are reported as smallest-input blockers rather than retry approval prompts.
- Registry rules, GitHub delivery guidance, generated V3 guardrails, checked-in agent guidance, focused tests, the Constitution, and durable feature rationale now express the same boundary.
- Delivery uses issue `#64` and branch `GH-64`.

## REPOSITORY MEMORY

Decision: created

Rationale: Failure-recovery authorization and destructive-action boundaries are consequential cross-repository policy decisions. Future agents need the reasoning and rejected unsafe interpretations, not only the final rule text.

Artifacts:

- `docs/specs/0046-autonomous-mutation-recovery/SPEC.md`
- `docs/references/rules/safety-guardrails.md`
- `docs/references/rules/github-pr-delivery.md`
- `docs/references/rules/work-lane-gating.md`
- `internal/templates/instruction_support_templates.go`
- `docs/agents/GUARDRAILS.md`
- `docs/CONSTITUTION.md`
- focused ruleset and generated-guidance tests
