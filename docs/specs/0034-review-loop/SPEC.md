---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0034
  slug: review-loop
  dir: 0034-review-loop
skills: []
relationships:
  - type: builds_on
    target: 0008-dispatch-command
  - type: related_to
    target: 0033-kit-capabilities
references:
  - id: constitution
    name: Constitution
    type: doc
    target: docs/CONSTITUTION.md
    selector_type: heading
    selector: CONSTRAINTS
    relation: constrains
    read_policy: must
    used_for: document-first workflow, no implementation details in specs, prompt-only/mutation boundaries, and external review tool guardrails
    status: active
  - id: agent-workflows
    name: Agent workflows
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: spec-phase clarification protocol, source-of-truth order, and readiness gate expectations
    status: active
  - id: agent-guardrails
    name: Agent guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    selector_type: heading
    selector: Completion Bar
    relation: constrains
    read_policy: must
    used_for: required section population, no git mutation without approval, and no CodeRabbit prompt execution without approval
    status: active
  - id: agent-rlm
    name: RLM guidance
    type: doc
    target: docs/agents/RLM.md
    selector_type: heading
    selector: Rules
    relation: guides
    read_policy: must
    used_for: prior-feature discovery bounds and just-in-time context loading
    status: active
  - id: agent-tooling
    name: Tooling guidance
    type: doc
    target: docs/agents/TOOLING.md
    selector_type: heading
    selector: Skills
    relation: guides
    read_policy: must
    used_for: repo-local skills discovery and command capability discovery expectations
    status: active
  - id: review-loop-brainstorm
    name: Review loop brainstorm
    type: doc
    target: docs/specs/0034-review-loop/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, selected direction, codebase findings, and open product questions
    status: active
  - id: dispatch-spec
    name: Dispatch command spec
    type: feature
    target: docs/specs/0008-dispatch-command/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: existing PR review-thread intake contract, CodeRabbit filtering, editor prefill, dedupe, and prompt-only dispatch behavior
    status: active
  - id: capabilities-spec
    name: Kit capabilities spec
    type: feature
    target: docs/specs/0033-kit-capabilities/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: command capability metadata requirements and discoverability expectations for dispatch and future review-loop command surfaces
    status: active
  - id: dispatch-command
    name: Dispatch command
    type: code
    target: pkg/cli/dispatch.go
    selector_type: symbol
    selector: dispatchCmd
    relation: informs
    read_policy: must
    used_for: current prompt-only editor-backed review intake surface
    status: active
  - id: dispatch-pr-intake
    name: Dispatch PR review intake
    type: code
    target: pkg/cli/dispatch_pr.go
    selector_type: symbol
    selector: loadDispatchPRInput
    relation: informs
    read_policy: must
    used_for: current PR target resolution, GitHub GraphQL review-thread fetch, and editor population behavior
    status: active
  - id: dispatch-pr-extraction
    name: Dispatch PR review extraction
    type: code
    target: pkg/cli/dispatch_pr_extract.go
    selector_type: symbol
    selector: extractDispatchReviewTasks
    relation: informs
    read_policy: must
    used_for: current unresolved/outdated filtering, CodeRabbit author filtering, prompt extraction, boilerplate removal, and dedupe behavior
    status: active
  - id: capabilities-catalog
    name: Capabilities catalog
    type: code
    target: pkg/cli/capabilities_catalog.go
    selector_type: symbol
    selector: capabilityCatalog
    relation: constrains
    read_policy: must
    used_for: current dispatch review-loop intake metadata and future review-loop command metadata expectations
    status: active
  - id: github-check-runs-api
    name: GitHub Check Runs API
    type: url
    target: https://docs.github.com/en/rest/checks/runs
    relation: informs
    read_policy: evidence
    used_for: current check-run status and conclusion fields for review-completion detection
    status: active
  - id: github-pr-graphql
    name: GitHub GraphQL pull request reference
    type: url
    target: https://docs.github.com/en/graphql/reference/pulls
    relation: informs
    read_policy: evidence
    used_for: current pull request review-thread fields, including resolved state, path, line, and comments
    status: active
  - id: coderabbit-review-commands
    name: CodeRabbit review commands
    type: url
    target: https://docs.coderabbit.ai/reference/review-commands
    relation: informs
    read_policy: evidence
    used_for: manual review, full review, pause, resume, and ignore command expectations
    status: active
  - id: coderabbit-github-checks
    name: CodeRabbit GitHub Checks
    type: url
    target: https://docs.coderabbit.ai/tools/github-checks
    relation: informs
    read_policy: evidence
    used_for: CodeRabbit's current GitHub Checks timing and failure-comment behavior
    status: active
---
# SPEC

## SUMMARY

Add a Kit-native review-loop prompt-prep workflow, exposed through `kit dispatch --loop`, that turns current unresolved CodeRabbit PR feedback into a human-reviewed dispatch prompt with optional review-completion waiting and no git or GitHub mutation without explicit approval.

## PROBLEM

Kit can already prefill `kit dispatch` from unresolved PR review threads, but the user still manually waits for CodeRabbit, copies comments from GitHub, removes repeated boilerplate, judges whether findings are still valid, and decides what should be fixed. That manual loop is slow, error-prone, and easy to run against stale, resolved, outdated, or out-of-scope comments.

The feature must make review follow-up deterministic while preserving Kit's core safety boundary: diagnosis and prompt preparation may be automated, but staging, committing, pushing, resolving conversations, and posting PR comments remain human-gated.

## GOALS

- Provide a Kit review-loop prompt-prep workflow for pull requests with CodeRabbit feedback through `kit dispatch --loop`.
- Fetch only current, unresolved, non-outdated PR review-thread feedback for the selected PR.
- Preserve the existing `kit dispatch --pr --coderabbit` review-task extraction semantics for CodeRabbit prompt blocks, fallback comment cleanup, shared boilerplate removal, and dedupe.
- Optionally wait until the relevant CodeRabbit review signal has completed for the current PR head before preparing work.
- Classify each current finding against the PR goal and current checkout so still-valid in-scope work is separated from stale, false-positive, out-of-scope, and human-decision items.
- Populate the editor with a concise, deduped, reviewable dispatch input before producing the final dispatch prompt.
- Keep the workflow read-only until the user explicitly chooses to mutate git or GitHub state through existing repo-local delivery rules.
- Expose the command behavior through `kit capabilities` under the runnable dispatch surface.

## NON-GOALS

- Do not automatically stage, commit, push, merge, resolve review threads, post PR comments, or trigger GitHub mutations.
- Do not resolve review threads as part of default `kit dispatch --loop` or `kit dispatch --pr` prompt generation; resolution is allowed only through an explicit follow-up mutation command after fixes or no-op decisions are complete.
- Do not replace CodeRabbit, reinterpret CodeRabbit configuration, or implement a webhook daemon.
- Do not blindly apply every CodeRabbit finding without validating it against the current code and PR goal.
- Do not include top-level PR comments unless a later approved requirement explicitly expands scope beyond review threads.
- Do not change existing `kit dispatch --pr --coderabbit` behavior except where explicitly required to share reusable review-loop semantics.
- Do not require a new hidden database or long-lived background process.

## USERS

- Kit operators working on GitHub pull requests that receive CodeRabbit review feedback.
- Coding agents that need a deterministic, repo-local command for collecting and triaging actionable review work.
- Maintainers who want review follow-up to preserve traceability, PR scope, and human approval boundaries.

## SKILLS

No additional execution-time skills are required for this specification. Repo-local canonical skills were not present under `.agents/skills`, and the documented global skills reviewed did not add a narrower, required execution skill for defining this feature.

## RELATIONSHIPS

Relationships are tracked in front matter.

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

- [SPEC-01] Kit must expose the review-loop prompt-prep workflow through `kit dispatch --loop`.
- [SPEC-02] The workflow must accept PR targets in the same user-facing forms supported by dispatch PR intake: GitHub PR URL, Markdown PR link, `owner/repo#number`, or current-repo PR number.
- [SPEC-03] The workflow must support a CodeRabbit-focused mode that includes CodeRabbit-authored review-thread comments and excludes non-CodeRabbit authors.
- [SPEC-04] The workflow must exclude resolved review threads.
- [SPEC-05] The workflow must exclude outdated review threads or comments.
- [SPEC-06] The workflow must exclude top-level PR comments in v1.
- [SPEC-07] The workflow must extract CodeRabbit `Prompt for AI Agents` content when present.
- [SPEC-08] When a CodeRabbit prompt block is absent, the workflow must fall back to a cleaned review-comment body.
- [SPEC-09] The workflow must include repeated CodeRabbit shared review instructions once at most.
- [SPEC-10] The workflow must dedupe equivalent review tasks by normalized task text plus source path and line.
- [SPEC-11] The workflow must preserve source evidence for each task, including at minimum file path, line when available, author, and review-thread or comment URL.
- [SPEC-12] The workflow must support a watch mode that waits for relevant CodeRabbit review activity to finish before collecting review tasks.
- [SPEC-13] Watch mode must be tied to the current PR head SHA so findings from a previous commit are not treated as current work.
- [SPEC-14] Watch mode must stop with an actionable timeout message instead of falling back to stale review state.
- [SPEC-15] Watch mode must wait 90 seconds before the first polling request, then poll every 15 seconds until completion or timeout.
- [SPEC-16] Watch mode must use a 15 minute timeout.
- [SPEC-17] Watch mode must require a 60 second quiet window after CodeRabbit appears complete before collecting review tasks.
- [SPEC-18] The workflow must classify collected findings into `FIX`, `VALID_OUT_OF_SCOPE`, `FALSE_POSITIVE`, `STALE`, and `NEEDS_HUMAN`.
- [SPEC-19] Classification must include a short reason and enough evidence for a human to decide whether to accept, edit, or remove each item in the editor.
- [SPEC-20] Only `FIX` items should be included in the default dispatch task block; non-fix classifications must remain visible in a review summary.
- [SPEC-21] The editor prefill must be concise, deduped, and safe for a human to edit before prompt generation.
- [SPEC-22] If no actionable current review feedback remains, the command must report that state clearly and must not open an empty editor.
- [SPEC-23] The command must remain read-only by default: no file writes except editor/clipboard/output behavior, no git staging or commits, no pushes, no PR comment posts, and no review-thread resolution.
- [SPEC-24] Any future mutation mode must require explicit user approval and must honor repo-local GitHub delivery rules before mutation.
- [SPEC-24A] `kit dispatch --pr <target> --resolve --yes` must resolve currently matching unresolved, non-outdated GitHub review threads after fixes or no-op decisions are complete, and must require `--yes` so the mutation is never accidental.
- [SPEC-25] The command capability metadata must describe default read-only behavior, GitHub network reads, CodeRabbit filtering, watch behavior, editor behavior, alias behavior, and mutation boundaries.
- [SPEC-26] README and agent-facing docs must describe when to use `dispatch --loop` and `dispatch --pr --coderabbit`.
- [SPEC-27] The workflow must fail with actionable errors when `gh` is unavailable, unauthenticated, missing repository access, or unable to query review threads.
- [SPEC-28] The workflow must handle paginated review-thread results.
- [SPEC-29] The workflow must preserve existing dispatch PR intake acceptance behavior unless this spec explicitly changes it.

## ACCEPTANCE

- [ACCEPT-01] `go test ./...` exits 0.
- [ACCEPT-02] CLI help exposes `kit dispatch --loop` and its important flags.
- [ACCEPT-03] Running the workflow with a GitHub PR URL, Markdown PR link, `owner/repo#number`, and current-repo PR number resolves the same PR target as dispatch PR intake.
- [ACCEPT-04] Tests prove resolved review threads are excluded.
- [ACCEPT-05] Tests prove outdated review threads or comments are excluded.
- [ACCEPT-06] Tests prove non-CodeRabbit authors are excluded in CodeRabbit-focused mode.
- [ACCEPT-07] Tests prove CodeRabbit `Prompt for AI Agents` blocks are extracted when present.
- [ACCEPT-08] Tests prove cleaned review-comment bodies are used when no CodeRabbit prompt block exists.
- [ACCEPT-09] Tests prove repeated CodeRabbit shared review instructions appear once at most.
- [ACCEPT-10] Tests prove duplicate review tasks collapse by normalized body plus source path and line.
- [ACCEPT-11] Tests prove watch mode uses the current PR head SHA and does not accept review state from an older head.
- [ACCEPT-12] Tests prove watch mode waits 90 seconds before polling, then polls every 15 seconds, uses a 15 minute timeout, and waits for a 60 second quiet window after CodeRabbit appears complete.
- [ACCEPT-13] Tests prove timeout and unavailable-review states are reported clearly without producing stale dispatch work.
- [ACCEPT-14] Tests prove triage output separates `FIX`, `VALID_OUT_OF_SCOPE`, `FALSE_POSITIVE`, `STALE`, and `NEEDS_HUMAN` findings.
- [ACCEPT-15] Tests prove only `FIX` items are included in the default dispatch task block while skipped classifications remain visible in summary output.
- [ACCEPT-16] Tests prove no-actionable-feedback output does not open the editor.
- [ACCEPT-17] Tests prove the default workflow does not stage files, commit, push, post PR comments, resolve conversations, or write project files.
- [ACCEPT-17A] Tests prove review-thread resolution requires `--resolve --yes`, filters to current unresolved non-outdated review threads, and can be limited to CodeRabbit-authored threads.
- [ACCEPT-18] Tests prove `kit dispatch --loop` routes to the review-loop prompt-prep behavior.
- [ACCEPT-19] `go run ./cmd/kit capabilities review-loop --json` fails because the compatibility root command has been removed.
- [ACCEPT-20] `go run ./cmd/kit capabilities --search review-loop --json` identifies the relevant review-loop surface.
- [ACCEPT-21] README and agent-facing docs distinguish `kit dispatch --loop` and `kit dispatch --pr --coderabbit`.
- [ACCEPT-22] `kit map 0034-review-loop` resolves the completed spec relationships and material references.
- [ACCEPT-23] `kit check 0034-review-loop` passes after the spec is finalized.

## EDGE-CASES

- CodeRabbit review is still running for the current PR head.
- CodeRabbit review completed for a previous PR head, but the branch has newer commits.
- GitHub returns mixed CodeRabbit and human review-thread comments.
- GitHub returns resolved, unresolved, outdated, and current review threads in the same page.
- Review threads are paginated.
- A review thread contains multiple comments.
- A CodeRabbit comment has no `Prompt for AI Agents` block.
- A CodeRabbit comment repeats the shared boilerplate instruction.
- A finding points at a file or line that no longer exists locally.
- A finding is valid but outside the current PR goal.
- A finding is a false positive after inspecting current code.
- The PR has no actionable review threads.
- `gh` is not installed, unauthenticated, unauthorized for the repository, offline, or returns malformed data.
- The current directory has no GitHub `origin` remote and the user provided only a PR number.
- The user cancels or clears the editor content.
- Clipboard support is unavailable.

## OPEN-QUESTIONS

none
