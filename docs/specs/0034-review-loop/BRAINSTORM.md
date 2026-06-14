---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0034
  slug: review-loop
  dir: 0034-review-loop
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0034-review-loop
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
  - id: dispatch-command
    name: Dispatch command
    type: code
    target: pkg/cli/dispatch.go
    selector_type: symbol
    selector: dispatchCmd
    relation: informs
    read_policy: must
    used_for: current prompt-only review intake and editor-backed dispatch behavior
    status: active
  - id: dispatch-pr-intake
    name: Dispatch PR review intake
    type: code
    target: pkg/cli/dispatch_pr.go
    selector_type: symbol
    selector: loadDispatchPRInput
    relation: informs
    read_policy: must
    used_for: current GitHub PR parsing, review-thread fetch, and editor population behavior
    status: active
  - id: dispatch-pr-extraction
    name: Dispatch PR review extraction
    type: code
    target: pkg/cli/dispatch_pr_extract.go
    selector_type: symbol
    selector: extractDispatchReviewTasks
    relation: informs
    read_policy: must
    used_for: current unresolved/non-outdated filtering, CodeRabbit prompt extraction, and dedupe behavior
    status: active
  - id: capabilities-catalog
    name: Capabilities catalog
    type: code
    target: pkg/cli/capabilities_catalog.go
    selector_type: symbol
    selector: capabilityCatalog
    relation: constrains
    read_policy: must
    used_for: ensuring command capability metadata describes review-loop intake behavior
    status: active
  - id: capabilities-tests
    name: Capabilities tests
    type: code
    target: pkg/cli/capabilities_test.go
    selector_type: symbol
    selector: TestCapabilitiesTargetedJSON
    relation: verifies
    read_policy: evidence
    used_for: regression coverage for dispatch review-loop capability metadata
    status: active
---
# BRAINSTORM

## SUMMARY

Kit already has the first piece of a CodeRabbit review workflow through `kit dispatch --pr --coderabbit`, but it does not yet orchestrate an end-to-end review loop. The likely direction is a dedicated `kit review-loop` command that reuses dispatch's PR-review intake, waits for current CodeRabbit feedback, semantically triages findings, and keeps all git/GitHub mutations human-gated.

## USER THESIS

Kit should make the CodeRabbit follow-up loop deterministic instead of relying on manual copy/paste from GitHub review comments. The command should collect only current unresolved review feedback, extract CodeRabbit's agent-ready prompt blocks when present, classify findings against the PR goal and current code, and route still-valid in-scope work through the existing dispatch/editor path before any staging, commit, push, PR comment, or conversation-resolution mutation occurs.

## Context Synthesis

The objective is to design and implement a Kit-native CodeRabbit review loop for pull requests that waits for CodeRabbit review completion, fetches only current unresolved and non-outdated feedback, semantically triages findings against the PR goal and current codebase, dispatches valid in-scope fixes to an agent, and preserves human-gated staging, commit, and push behavior [S1][S2][S3][S4]. Affected users are Kit operators using GitHub PR workflows with CodeRabbit review feedback [S1][S7]. Constraints: use `gh` and GitHub APIs, keep diagnosis read-only until explicit mutation approval, ignore stale or resolved comments, and reject blind implementation of irrelevant feedback [S2][S3][S4][S5][S6]. Done means CLI behavior exists for `--watch` plus semantic triage, with tests covering review-completion detection, review-thread filtering, classification, editor population, validation reporting, capability discovery, and mutation gating [S2][S3][S4][S9][S10][S11]. The selected direction is to add a dedicated `kit review-loop` command unless implementation inspection proves a narrower `kit dispatch --loop` flag fits better; either way, the implementation must reuse dispatch's PR intake instead of duplicating review-thread parsing [S1][S3][S9][S10][S11].

## Source Map

- [S1] discussion — User workflow: wait for `coderabbitai` feedback, fix, stage, commit, and repeat until each item is fixed or resolved by assessment. — conversation:side-coderabbit-loop
- [S2] discussion — Decision: add `--watch` mode to wait for CodeRabbit completion before dispatching a focused subagent pass. — conversation:side-coderabbit-watch
- [S3] discussion — Decision: semantically triage CodeRabbit findings against the stated work goal and broader codebase before fixing. — conversation:side-coderabbit-semantic-triage
- [S4] doc — Kit guardrail: Git and GitHub mutations require repo-local workflow rules and explicit approval. — `/Users/jamesonstone/go/src/github.com/jamesonstone/kit/AGENTS.md`
- [S5] link — GitHub Checks API exposes check-run status for commit refs. — https://docs.github.com/en/rest/checks/runs
- [S6] link — GitHub GraphQL exposes PR review threads for filtering review feedback state. — https://docs.github.com/en/graphql/reference/pulls
- [S7] link — CodeRabbit commands support manual review triggering and review-flow commands. — https://docs.coderabbit.ai/reference/review-commands
- [S8] link — GitHub webhooks support event-driven future handling for check runs and review events. — https://docs.github.com/en/webhooks/webhook-events-and-payloads
- [S9] code — `pkg/cli/dispatch.go::dispatchCmd` already provides prompt-only editor-backed dispatch and exposes `--pr` plus `--coderabbit`.
- [S10] code — `pkg/cli/dispatch_pr.go::loadDispatchPRInput` resolves PR URLs, Markdown links, owner/repo numbers, and current-repo PR numbers, then fetches GitHub review threads through `gh api graphql`.
- [S11] code — `pkg/cli/dispatch_pr_extract.go::extractDispatchReviewTasks` skips resolved and outdated threads, filters CodeRabbit authors, extracts `Prompt for AI Agents` details, strips repeated boilerplate, and dedupes tasks.
- [S12] code — `pkg/cli/capabilities_catalog.go::capabilityCatalog` now describes `dispatch` as PR review-thread and CodeRabbit review-loop intake.
- [S13] test — `pkg/cli/capabilities_test.go::TestCapabilitiesTargetedJSON` verifies dispatch capability metadata documents review-loop intake, review-thread filtering, and CodeRabbit prompt extraction.
- [S14] doc — `README.md` documents `kit dispatch --pr <url|number>` and `--coderabbit` for unresolved, non-outdated PR review threads.

## Coding Agent Instructions

Build the first-pass command as a deterministic, human-gated workflow: prefer a dedicated `kit review-loop` command unless repository inspection proves `kit dispatch --loop` fits the existing CLI structure with less surface area [S1][S3]. The tradeoff is deliberate: polling plus editor dispatch is simpler and safer than a daemon or webhook receiver, while preserving a later event-driven path [S2][S4][S8].

1. Inspect the repository and identify existing PR dispatch, GitHub API, CLI command, config, and editor functionality by exact file path and symbol; UNKNOWN exact entrypoints, so run `pwd`, `git status --short --branch`, and `rg -n "dispatch|coderabbit|reviewThreads|gh pr|cobra|editor" .` [S4].
2. Reconcile brainstorm decisions with actual code behavior, including current `kit dispatch --pr`, `--coderabbit`, editor population, GitHub GraphQL filtering, and mutation guardrails [S1][S3][S4][S6].
3. Produce a complete implementation strategy grounded in discovered files, symbols, command wiring, test style, and config conventions before editing code [S3][S4].
4. Enumerate concrete file edits, interfaces, data model changes, dependency updates, configuration changes, migration steps, validation commands, and tests in the implementation plan [S3][S4].
5. Implement `--watch` completion detection using GitHub check-run status, CodeRabbit review activity, a quiet window, and current PR head SHA tracking [S2][S5][S7].
6. Implement semantic triage classifications: `FIX`, `VALID_OUT_OF_SCOPE`, `FALSE_POSITIVE`, `STALE`, and `NEEDS_HUMAN`, with evidence, file path, line, PR-goal relationship, and recommended action [S3][S6].
7. Preserve read-only diagnosis until explicit approval; do not stage, commit, push, resolve conversations, or post GitHub comments without a repo-local delivery gate and user approval [S4].
8. Define acceptance checks with expected outputs: unit tests pass, CLI help lists the new command or flag, dry-run emits triage groups, stale/resolved comments are omitted, and mutation attempts stop before approval [S2][S3][S4].
9. State risks, open questions, and explicit assumptions with mitigation and owner, including CodeRabbit status signal gaps, API pagination, rate limits, private-repo auth, and command naming ownership [S5][S6][S7][S8].

## Resource Links

- [R1] GitHub REST Check Runs API — https://docs.github.com/en/rest/checks/runs — Lists check runs for a ref and exposes queued, in-progress, and completed states.
- [R2] GitHub GraphQL Pull Requests Reference — https://docs.github.com/en/graphql/reference/pulls — Documents PR review threads and pull request review data.
- [R3] GitHub Webhook Events and Payloads — https://docs.github.com/en/webhooks/webhook-events-and-payloads — Documents events for check runs, review comments, review threads, and workflow runs.
- [R4] GitHub Actions Events That Trigger Workflows — https://docs.github.com/actions/using-workflows/events-that-trigger-workflows — Official reference for workflow event triggers.
- [R5] GitHub REST Commit Statuses API — https://docs.github.com/en/rest/commits/statuses — Supports fallback status inspection for refs.
- [R6] CodeRabbit Review Commands — https://docs.coderabbit.ai/reference/review-commands — Documents `@coderabbitai review`, `full review`, `approve`, `resolve`, and autofix commands.
- [R7] CodeRabbit GitHub Checks — https://docs.coderabbit.ai/tools/github-checks — Explains how CodeRabbit reads GitHub Checks and posts remediation comments.
- [R8] CodeRabbit FAQ — https://docs.coderabbit.ai/faq — Describes CodeRabbit automatic and manual review triggering.
- [R9] CodeRabbit Glossary — https://docs.coderabbit.ai/reference/glossary — Defines Request Changes Workflow and approval behavior.
- [R10] CodeRabbit Manage Code Reviews — https://docs.coderabbit.ai/guides/commands — Documents pause, resume, manual review, and comment-resolution workflows.
- [R11] CodeRabbit Automatic Review Controls — https://docs.coderabbit.ai/configuration/auto-review — Documents configuration for automatic review scope and draft handling.
- [R12] CodeRabbit GitHub Platform Docs — https://docs.coderabbit.ai/platforms/github-com — Documents CodeRabbit GitHub integration setup and permissions.
- [R13] CodeRabbit Changelog — https://docs.coderabbit.ai/changelog — Notes review command completion-status changes.

## RELATIONSHIPS

Relationships are tracked in front matter.

## CODEBASE FINDINGS

- `kit dispatch` is the current review-feedback ingestion point. It remains prompt-only, opens the editor with collected tasks, normalizes the edited task list, and then emits the existing overlap-discovery dispatch prompt [S9].
- `--pr` already accepts full GitHub PR URLs, Markdown PR links, `owner/repo#123`, and plain PR numbers resolved from the current repository's `origin` remote [S10][S14].
- GitHub review-thread collection already uses GraphQL pagination through `gh api graphql`, so the review-loop implementation should build on that API boundary rather than introduce a second PR-comment collector [S10].
- Existing extraction already skips resolved or outdated threads, supports CodeRabbit-only filtering, extracts `Prompt for AI Agents` details, removes repeated CodeRabbit boilerplate, falls back to cleaned comment bodies, and dedupes by path, line, and normalized body [S11].
- `kit capabilities` must stay accurate as the command surface evolves. In this pass, it documents the current review-loop intake behavior under `dispatch` instead of advertising an unimplemented `review-loop` command [S12][S13].
- A future dedicated `review-loop` command should add its own capability record only when the command is registered and runnable. Until then, `kit capabilities --search review-loop` should lead agents to the existing `dispatch --pr --coderabbit` intake surface [S12][S13].

## AFFECTED FILES

- `pkg/cli/dispatch.go` — current prompt-only dispatch command and flag surface.
- `pkg/cli/dispatch_pr.go` — PR target resolution, GitHub GraphQL review-thread fetch, and editor prefill entrypoint.
- `pkg/cli/dispatch_pr_extract.go` — review-thread filtering, CodeRabbit prompt extraction, comment cleanup, and dedupe logic.
- `pkg/cli/dispatch_test.go` — current coverage for PR parsing, CodeRabbit filtering, dedupe, and no-action behavior.
- `pkg/cli/capabilities_catalog.go` — command capability metadata that must describe current and future review-loop surfaces.
- `pkg/cli/capabilities_test.go` — regression coverage for capabilities metadata and search behavior.
- `README.md` — human-facing docs for dispatch PR review intake.
- `docs/specs/0034-review-loop/BRAINSTORM.md` — source artifact for the next specification pass.

## DEPENDENCIES

References are tracked in front matter.

## QUESTIONS

- Should the first runnable command be `kit review-loop` or `kit dispatch --loop` after implementation inspection? Current recommendation is `kit review-loop`, with dispatch reused internally for editor/prompt intake.
- What exact signal should define "CodeRabbit is done": a check-run conclusion, no in-progress CodeRabbit checks, latest CodeRabbit review timestamp plus quiet window, or a combined heuristic?
- Should the loop ever post skip/fix summaries back to GitHub, or should v1 remain local/editor-only with the human deciding what to stage, commit, push, and resolve?
- What PR-goal source should semantic triage use first: PR title/body, linked issue, feature docs in the checkout, or user-provided task text?
- What timeout and polling interval should `--watch` use, and should those defaults be configurable in `.kit.yaml`?

## OPTIONS

1. Dedicated `kit review-loop` command:
   - Pros: clear workflow ownership, discoverable capability metadata, room for watch/triage/report flags without overloading dispatch.
   - Cons: adds a new root command and must carefully reuse dispatch internals to avoid duplicate PR review parsing.
2. `kit dispatch --loop`:
   - Pros: smallest visible command-surface change and directly extends the current review-comment intake path.
   - Cons: dispatch is intentionally prompt-only and queue-planning focused; adding polling, triage, and loop state risks blurring that contract.
3. External script or prompt-only recipe:
   - Pros: fastest to prototype and least invasive.
   - Cons: keeps the current manual workflow brittle, harder to test, and harder for agents to discover through `kit capabilities`.

## RECOMMENDED STRATEGY

Implement a dedicated `kit review-loop` command in the next phase, but reuse the existing dispatch PR intake/extraction/editor functions as the foundation. Keep v1 read-only through diagnosis and triage: fetch current review state, wait for CodeRabbit completion when requested, classify actionable items, and then populate the existing dispatch/editor flow. Do not stage, commit, push, resolve conversations, or post GitHub comments without an explicit repo-local delivery gate.

`kit capabilities` should describe both layers as they become real: `dispatch` remains the current PR review and CodeRabbit prompt intake command, while `review-loop` should receive its own record only after the command is registered.

## NEXT STEP

Run `kit spec 0034-review-loop` after resolving the command-shape and CodeRabbit-completion signal questions.
