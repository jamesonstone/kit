---
kind: workflow
slug: pr-feedback-repair
description: Collects asynchronous pull-request feedback and drives bounded coding-agent repair without making Kit an agent runtime.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - github
  - pull-request
  - review
  - feedback
  - repair
  - coderabbit
read_policy_default: conditional
dependencies:
  - ruleset/agent-team-orchestration
  - ruleset/github-pr-delivery
  - ruleset/safety-guardrails
  - ruleset/work-lane-gating
  - ruleset/testing-and-environment-validation
  - ruleset/source-file-size
pr_feedback:
  schema_version: 1
  modes: [await, collect]
  watcher_key_fields: [repository, pull_request, head_sha]
  wake_events: [status, review, review-comment]
  status_query_fields: [pull_request_state, head_ref_oid, provider_state, provider_description, rate_limit]
  provider_context: CodeRabbit
  completed_description: Review completed
  skipped_description_prefix: "Review skipped:"
  status_schedule_seconds: [0, 90, 180, 360, 600, 900, 1200, 1500]
  quiet_window_seconds: 60
  default_timeout_seconds: 1500
  max_timeout_seconds: 3600
  jitter_percent: 10
  max_status_requests_per_head: 9
  request_budget_per_head: 32
  rate_reserve_points: 500
  rate_reserve_percent: 10
  max_head_epochs: 2
  max_repair_passes: 2
  collection:
    page_size: 50
    max_pages: 20
    sources:
      - review-threads
      - requested-change-reviews
      - trusted-top-level-comments
    exclude_resolved: true
    exclude_outdated: true
    include_human_threads: true
    prompt_marker: Prompt for AI Agents
    trusted_comment_marker: "<!-- kit:pr-feedback -->"
    fingerprint_fields: [normalized_task, path, line]
---

# Workflow: PR Feedback Repair

## Purpose

Give one accountable coding-agent supervisor a provider-neutral contract for
collecting asynchronous pull-request feedback, verifying it against the
current head, repairing valid findings, and closing only proven addressed
threads.

`kit contract resolve` reads this repository-local workflow and its rules
without GitHub, network, write, or Git activity. The protected `kit pr fix`
fallback may explicitly collect GitHub feedback, wait within this contract,
prepare the exact writable PR-head lane, emit the supervisor prompt, and
resolve only confirmed named threads. It does not launch or supervise agents
or perform repair, staging, commit, push, comment, merge, or implicit thread
resolution.

## Intake Modes

Both modes enter the same repair phases after collecting feedback:

- `await` is a bounded wait when CodeRabbit feedback is expected for one exact
  pull-request head. Prefer a host webhook, status, review, or review-comment
  wakeup. When the host cannot wake the task, run one shell/helper process that
  uses `gh` and sleeps without model turns.
- `collect` is a one-shot scan after a host event, on explicit request, or long
  after feedback arrived. It fetches active provider and human feedback
  immediately and never requires a preceding provider wait.

Deduplicate active watchers by `repository + pull request number + expected
head SHA`. Persist status and feedback fingerprints in the host state
directory, not in tracked project files. Use an atomic per-key lock so a
second watcher reports the existing observation instead of polling again.

## Deterministic Await

Record repository identity, pull-request number, expected remote head, local
head, writable push target, timeout, request budget, and state-file location
before waiting. A compact status query must retrieve pull-request state,
`headRefOid`, the CodeRabbit `StatusContext` state and description, and the
rate budget in one request:

```graphql
query PRFeedbackStatus($owner: String!, $name: String!, $number: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      state
      headRefOid
      commits(last: 1) {
        nodes {
          commit {
            statusCheckRollup {
              contexts(first: 100) {
                nodes {
                  __typename
                  ... on StatusContext {
                    context
                    state
                    description
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  rateLimit { cost limit remaining resetAt }
}
```

Filter the returned contexts to the exact `CodeRabbit` name. Do not fetch
review threads during status observation. The verified query cost is one, but
the helper must consume the returned cost rather than assume it forever.

Use at most the eight scheduled observations at `t=0`, 90 seconds, 3, 6, 10,
15, 20, and 25 minutes for each head, with bounded 10 percent jitter after the
initial request. Constant 15-second polling is not the default. When
`Review completed` is first observed, wait for one 60-second quiet window,
repeat the compact query once, require the same head and completion
description, and only then collect feedback. The default timeout is 25 minutes
and may be configured only up to the one-hour ceiling. A timeout leaves a
non-clean `timed-out` result.

Before another request, stop if the per-head request budget would be exceeded
or the returned remaining points fall below the greater of 500 points or 10
percent of the limit. On HTTP 403 or 429, preserve `Retry-After` when present,
otherwise preserve `resetAt`, emit `rate-limited`, and exit without a rapid
retry. A future host wakeup or explicit collect may resume after that time.

## Terminal Classification

Status state and description form one decision; `SUCCESS` alone is unsafe.

| Result | Evidence and behavior |
| --- | --- |
| `pending` | Status is `EXPECTED` or `PENDING`, or the context has not appeared. Continue only within the bounded schedule. |
| `completed` | Status is `SUCCESS` and description is exactly `Review completed`, followed by the stable quiet-window confirmation. Collect feedback; this does not mean the review is clean. |
| `skipped-with-reason` | Status is `SUCCESS` and description starts with `Review skipped:`. Preserve the complete description and suffix exactly, stop as terminal non-clean, and do not infer zero findings. |
| `provider-failure` | Status is `ERROR` or `FAILURE`. Preserve provider state and description. |
| `head-changed` | `headRefOid` differs from the expected head. Stop this watcher before collecting or repairing. |
| `unavailable` | The PR is not open, the provider is unavailable, or a successful/terminal description is unknown. Fail closed with the exact reason. |
| `timed-out` | The bounded await expires while pending. Pending or timeout is never clean completion. |
| `rate-limited` | HTTP 403/429, request budget exhaustion, or the configured reserve is reached. Preserve retry/reset evidence. |

Emit a `pr-feedback-await-v1` structured result containing schema version,
mode, terminal state, repository, pull-request number, expected and observed
heads, provider state and description, exact reason, retry/reset evidence,
request count, and returned rate-budget evidence. Fingerprint status from head,
context state, and exact description so repeated wakeups remain observable
without persisting provider bodies.

## Feedback Collection

Collect only after confirmed `completed`, a human review or review-comment
event, or an explicit `collect` invocation. Use paginated GraphQL
`reviewThreads`; continue through every page up to the configured bound and
paginate a thread's comments when its connection has another page. Exclude
threads whose current `isResolved` or `isOutdated` value is true.

For each active item preserve its GraphQL node ID, source path, current or
original line, author login, full body, URL, review source, and head SHA.
Include human-authored review threads by default. From CodeRabbit bodies,
retain the full comment and separately extract any `<details>` block whose
summary contains `Prompt for AI Agents`, including its fenced prompt text.

Also collect non-empty review bodies whose review state is
`CHANGES_REQUESTED`. Top-level PR comments are not general feedback input:
include one only when its author is explicitly allowlisted by the caller, or
when an `OWNER`, `MEMBER`, or `COLLABORATOR` author includes the exact
`<!-- kit:pr-feedback -->` marker. Preserve path and line as null when that
source has no inline location.

Deduplicate by a SHA-256 fingerprint of normalized task, source path, and line.
Preserve node and thread IDs, author, URL, and body as evidence on the retained
item. Persist fingerprints per watcher key without persisting feedback bodies.
Hitting a page or request bound is `unavailable`, not a partial clean
collection.

## Supervisor And Lane Preflight

One supervisor owns scope, findings, lane assignment, integration, validation,
delivery, and thread resolution. Before spawning any repair agent, write an
Agent Team Plan with predicted files, overlap, serialized work, queued work,
validation, and verification ownership.

- Default to at most three independent low-overlap concurrent lanes; four is
  the hard ceiling. Queue excess work.
- Serialize shared or ambiguous files. Never let parallel repair lanes edit
  the same unclear boundary.
- After a nontrivial repair, assign a separate read-only verification lane.
- Subagents may not create, switch, move, or remove worktrees. They may not
  stage, commit, push, mutate the pull request, or resolve threads.

Repair only from the exact writable same-repository PR-head branch and its
registered worktree. Never repair from a detached `PR-<number>` inspection
worktree. Record expected remote head, observed local head, push remote and
branch, and explicit ownership of every pre-existing dirty path. Stop on a
head mismatch or ambiguous dirty-change ownership.

## Repair Phases

1. Resolve this workflow and read every dependency before mutation.
2. Establish the exact writable PR-head lane and record the preflight evidence.
3. Run one intake mode and normalize the deduplicated active findings.
4. Verify every finding against current `HEAD`, its current path and line, and
   the integrated code. Fix only findings that remain valid; record a concise
   evidence-based reason for every skip.
5. Partition valid findings into independent lanes under the concurrency and
   overlap rules. Queue work that cannot safely run in parallel.
6. Integrate all repairs in the supervisor lane and validate the full combined
   diff, not only each lane's fragment.
7. Run the required read-only verification lane after nontrivial repair and
   close every gap before delivery unless the user explicitly accepts it.
8. Explicitly stage intended files and push one coherent repair batch to the
   recorded remote PR-head branch.
9. Reflect after the push: verify the exact remote head, re-check the complete
   diff and hosted states, then fetch active threads again.
10. Resolve only review threads whose exact finding was verified, addressed by
    the pushed head, and confirmed absent or corrected. Never bulk-resolve,
    resolve outdated identities blindly, or treat a provider status as proof.

A push starts a new head epoch. Permit at most two head epochs and two repair
passes for one workflow run; stop with an explicit remaining-feedback result
instead of creating an infinite review-and-push loop.

## Completion

- Intake reports one explicit terminal state; pending, skipped, unavailable,
  timeout, and rate limit are never reported as clean completion.
- Every repair or skip is tied to refreshed current-head evidence.
- The integrated diff passes applicable validation and read-only verification.
- One coherent pushed batch is verified at the exact remote head.
- Only proven addressed review threads are explicitly resolved, and any late
  human or provider feedback remains visible for event-triggered or one-shot
  collection.

## Verification

- `kit contract resolve --workflow pr-feedback-repair` selects this workflow,
  `agent-team-orchestration`, and every declared delivery dependency without
  performing network access or writes.
- Tests distinguish completed and skipped `SUCCESS`, head change, timeout,
  provider failure, unavailable context, rate reserve, HTTP 403/429, bounded
  scheduling, quiet confirmation, pagination bounds, deduplication fields,
  human feedback sources, and late collect mode.
- Command-tree and prompt goldens prove `kit pr fix` is the sole protected PR
  fallback and that dispatch, loop, and the broader PR family remain absent.
