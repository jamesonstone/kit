---
kind: ruleset
slug: deadline-mode
description: Narrows testing and implementation scope under an explicit, user-signaled deadline without weakening required approvals, security, or compatibility invariants.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - workflow
  - testing
  - implementation
  - prioritization
read_policy_default: conditional
---

# Ruleset: deadline-mode

## Purpose

- Let a coding agent narrow validation and implementation scope to what a real, user-declared deadline requires, instead of wasting time on evidence and cleanup the deadline makes low-value.
- Change prioritization and scope only. Deadline mode never weakens correctness, authority, or any required approval, security, or compatibility boundary.
- Require every narrowing to be explicit and recorded, never silent. A recorded single-lane or reduced-scope decision is a fully valid, first-class outcome, not a shortcut to hide.

## Applies When

- Load this ruleset only when the user explicitly signals a real deadline or time constraint in-thread: a concrete deadline, an explicit statement of being time-constrained, or an explicit request to work leaner or faster under time pressure.
- Do not infer deadline mode from repository signals, task size, calendar proximity, PR staleness, or an agent's own judgment that a task feels urgent.
- Never proactively suggest or enter deadline mode unprompted.
- Enter only at a safe checkpoint: not mid-mutation, and not while an unresolved verification gap is open.
- Does not apply to routine implementation work with no user-declared time pressure; `docs/references/rules/testing-and-environment-validation.md`'s default governs unless this ruleset is explicitly and currently active for the unit of work.

## Rules

### Priority Ordering

While deadline mode is active, prioritize in this order:

1. the correctness-critical path for the declared goal;
2. contracts and interfaces that unblock other dependent work;
3. focused implementation and repair on that critical path;
4. independent, high-risk verification (see `agent-team-orchestration.md`'s fresh independent `verifier`);
5. `PR_READY`/`MERGE_AUTHORIZED` delivery of the completed work, per `agent-team-orchestration.md`'s lifecycle and `github-pr-delivery.md`;
6. one compatibility or integration pass once a wave of work is complete; then
7. deferred operational acceptance (browser, operator, hardware, or full-production acceptance) once the feature exists.

### Stop Doing

While deadline mode is active, stop or defer:

- optional audits and broad cleanup unrelated to the declared goal;
- unrelated refactors;
- repeated broad test-suite reruns when current focused evidence already covers the change;
- premature browser automation, operator walkthroughs, or hardware acceptance before the workflow they exercise actually exists;
- low-value parallel exploration that does not sit on the correctness-critical path;
- re-running evidence that is already current and attributable to the exact head being validated.

Do not substitute repeated broad suites for focused evidence.

### Continue Doing — Never Weaken

Deadline mode must never weaken, skip, or narrow:

- required merge authorization — `docs/references/rules/github-pr-merge.md` and `pull-request-merge` still gate every merge exactly as always;
- required infrastructure approval — `docs/references/rules/infrastructure-change-approval.md`'s consolidated outline and one-pass execution still apply in full;
- independent final review — `docs/references/rules/agent-team-orchestration.md`'s fresh, read-only `verifier` requirement still applies to nontrivial implementation;
- required post-deployment tests — `docs/references/rules/testing-and-environment-validation.md`'s `### Local And Production Execution` production-suite requirement still applies after an actual deployment;
- default-off and fail-closed behavior, including `docs/references/rules/deletion-safety.md`'s soft-delete-by-default pattern as the concrete recurring instance;
- security, authentication, privacy, and tenant-isolation boundaries;
- database migration safety, and any immutable contract, hash, history, route, or replay-fixture guarantee the project owns.

Kit has no dedicated ruleset for migration safety or replay/contract-hash immutability today; treat these as plain, non-negotiable domain invariants rather than inventing a pointer to a rule that does not exist.

### Deadline Testing Budget

Reduce validation depth, not validation honesty:

- **Per pull request**: run the narrowest focused/affected tests that prove the change, plus every check `testing-and-environment-validation.md`'s `### Pull-Request CI` already makes mandatory in this project's hosted checks (formatting, linting, static/type analysis, build). Defer that same rule's "before handoff, run the complete applicable code-level suite" default — see the explicit supersession this ruleset records below.
- **Per completed wave**: run one compatibility or integration pass across the affected surfaces before starting the next wave.
- **After an actual deployment**: run exactly the required post-deployment acceptance check named in `testing-and-environment-validation.md`'s `### Local And Production Execution`. Do not skip this even under deadline mode.
- **At final activation readiness**: run full browser/operator/hardware/production acceptance per `testing-and-environment-validation.md`'s `### Safe Production Test Data` and browser-automation sections, plus a final review of the complete change.

### Explicit Supersession Of The Complete-Suite Default

- Deadline mode is an explicit, recorded, invariant-preserving supersession of `testing-and-environment-validation.md`'s "before handoff, run the complete applicable code-level suite" default. It is never a silent contradiction of that default.
- Record, in-thread, that deadline mode is active and exactly which validation was narrowed or deferred and why.
- Report deferred or reduced-scope validation literally, per `docs/references/rules/agent-completion-output.md`'s evidence-state discipline: use `PARTIAL` or `SKIPPED`, never `PASS`, for anything that did not actually run. Preserve `PENDING`, `UNKNOWN`, and `NOT_APPLICABLE` exactly as that rule already requires.

### Scope Expiry

One deadline-mode authorization covers the current unit of work the user declared it for, plus directly required tests, documentation, validation fixes, review fixes, and delivery for that same work. Ask again, or fall back to the ordinary `testing-and-environment-validation.md` default, before materially new or tangential scope, or once the safe checkpoint that started deadline mode no longer holds.

## Anti-Patterns

- Entering or suggesting deadline mode without an explicit user signal.
- Treating deadline mode as license to weaken merge authorization, infrastructure approval, independent review, required post-deployment tests, default-off/fail-closed behavior, security, or migration safety.
- Reporting deferred or narrowed validation as `PASS` instead of `PARTIAL` or `SKIPPED`.
- Letting one deadline-mode authorization silently cover materially new or tangential scope without asking again.
- Treating "the user has a deadline somewhere" as a standing license that outlives the checkpoint or unit of work it was declared for.

## Verification

- Confirm the user's explicit deadline or time-constraint signal is quoted or clearly identifiable before deadline mode is treated as active.
- Confirm the priority order was followed and any deferred item was actually low-priority under that order, not merely convenient to skip.
- Confirm every "Continue Doing" invariant was verified untouched.
- Confirm the current testing-budget tier matches the actual lifecycle point (per-PR vs. per-wave vs. post-deployment vs. activation readiness).
- Confirm deferred or reduced-scope validation was reported as `PARTIAL`/`SKIPPED`, never `PASS`.
- Confirm the authorization was re-asked when work materially expanded beyond the accepted scope.

## Examples

Entering deadline mode:

```text
User: "I have a hard deadline in 3 hours — let's move fast and only test what's actually load-bearing for this change."
Agent: records deadline-mode active for this unit of work, follows the priority
order, runs the per-PR testing budget, defers the broad suite rerun, and
reports the deferred validation as PARTIAL rather than PASS.
```

Declining to enter deadline mode:

```text
User: "Can you add this feature?" (no time-constraint signal)
Agent: proceeds under the ordinary testing-and-environment-validation.md
default — focused tests during development, the complete applicable
code-level suite before handoff.
```
