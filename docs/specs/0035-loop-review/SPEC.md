---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0035
  slug: loop-review
  dir: 0035-loop-review
relationships:
  - type: builds_on
    target: 0034-review-loop
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: brainstorm
    name: Loop review brainstorm
    type: feature
    target: docs/specs/0035-loop-review/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: selected command shape and stop contract
    status: active
---
# SPEC

## SUMMARY

Expose `kit loop review` as an agent-driven correctness loop that reviews changes on the current branch relative to the remote mainline, optionally folds in CodeRabbit PR feedback, and repeats until the agent emits a compact `done` summary with at least 95% correctness.

## PROBLEM

Kit has a feature workflow loop and a PR review prompt-prep command, but it does not have a single command that repeatedly asks a coding agent to inspect changed code, fix correctness issues, validate, and stop only when no high, medium, or correctness-impacting issues remain.

## GOALS

- Reorganize loop functionality under a `kit loop` command group.
- Preserve legacy `kit loop [feature]` behavior while exposing `kit loop workflow [feature]`.
- Add `kit loop review` for changed-code correctness review and repair.
- Use remote mainline comparison by default, with a base-ref override.
- Stop only when the agent reports `Correctness: <n>%`, `n >= 95`, a dense issue/fix summary, and final line `done`.
- Default review iterations to 10.
- In PR mode, start local review immediately and poll CodeRabbit opportunistically.
- Allow strict PR mode through `--watch` or `--wait-for-coderabbit`.
- Keep git and GitHub mutations outside the review loop.

## NON-GOALS

- Do not stage, commit, push, post PR comments, resolve review threads, or merge pull requests.
- Do not replace existing `kit review-loop` prompt-prep behavior in this feature.
- Do not create a background daemon, hidden database, webhook server, or long-lived scheduler.
- Do not mathematically prove global correctness; the percentage is an agent-reported confidence gate backed by local validation.

## USERS

- Maintainers who want a repeatable local review and repair loop before publishing or updating a PR.
- Coding agents that need a deterministic stop contract for review passes.

## SKILLS

not required

## RELATIONSHIPS

- builds on: 0034-review-loop
- related to: 0033-kit-capabilities

## DEPENDENCIES

- `.kit.yaml` loop agent configuration.
- Git for branch diff discovery.
- `gh` only when `--pr` is supplied.

## REQUIREMENTS

- [SPEC-01] `kit loop` must expose `workflow` and `review` subcommands.
- [SPEC-02] Legacy `kit loop [feature]` must remain compatible with the existing feature workflow loop.
- [SPEC-03] `kit loop workflow [feature]` must run the existing feature workflow loop.
- [SPEC-04] `kit loop review` must compare changed code against `origin/main` by default, falling back to local `main`.
- [SPEC-05] `kit loop review --base <ref>` must override the comparison base.
- [SPEC-06] `kit loop review [feature]` must include feature docs context in the review prompt when the feature can be resolved.
- [SPEC-07] The review loop must use the configured loop agent and write artifacts under `.kit/loops/<run-id>/`.
- [SPEC-08] The default max iteration count must be 10.
- [SPEC-09] The agent stop output must include `Correctness: <n>%` near the top and end with a final non-empty line exactly equal to `done`.
- [SPEC-10] The command must continue looping when correctness is below the configured threshold or the final line is not `done`.
- [SPEC-11] The review prompt must instruct the agent to fix high, medium, and correctness-impacting issues and ignore non-blocking style churn unless it affects correctness.
- [SPEC-12] The review prompt must prohibit staging, commits, pushes, PR comments, and review-thread resolution.
- [SPEC-13] `kit loop review --pr <target>` must start local review immediately and poll CodeRabbit feedback opportunistically.
- [SPEC-14] PR feedback discovered during or after a pass must be included in the next pass.
- [SPEC-15] When local review reaches `done`, PR mode must perform one quick feedback check.
- [SPEC-16] If CodeRabbit is still pending at local completion, default PR mode must exit with a provisional status and rerun guidance instead of waiting.
- [SPEC-17] `--watch` and `--wait-for-coderabbit` must wait for CodeRabbit completion up to the existing timeout before finalizing.
- [SPEC-18] Existing `kit review-loop` prompt-prep behavior must remain available as a compatibility surface.
- [SPEC-19] Command capability metadata and user/agent docs must describe the new command shape and safety boundaries.

## ACCEPTANCE

- [ACCEPT-01] Focused tests cover loop command routing, default iteration count, review stop parsing, base fallback, PR opportunistic feedback ingestion, and provisional PR status.
- [ACCEPT-02] `kit loop --help` shows `workflow` and `review`.
- [ACCEPT-03] `kit capabilities loop review --json` documents command behavior.
- [ACCEPT-04] `go test ./...` passes.

## EDGE-CASES

- No branch diff and no working-tree diff should produce a clear no-change review target.
- CodeRabbit unavailable without feedback should not block default PR mode.
- Repeated already-ingested PR feedback must not cause an infinite loop after the agent reports local completion.
- A missing loop agent must fail with setup guidance.

## OPEN-QUESTIONS

none
