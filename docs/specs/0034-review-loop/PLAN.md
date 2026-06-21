---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0034
  slug: review-loop
  dir: 0034-review-loop
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
    used_for: no hidden-state principle, no implementation details in specs, project-directory workflow, and external review tool guardrails
    status: active
  - id: workflow-rules
    name: Workflow rules
    type: doc
    target: docs/agents/WORKFLOWS.md
    selector_type: heading
    selector: Spec-Driven Work
    relation: constrains
    read_policy: must
    used_for: plan-phase source-of-truth order, docs-before-code sequencing, and clarification expectations
    status: active
  - id: rlm-rules
    name: RLM rules
    type: doc
    target: docs/agents/RLM.md
    selector_type: heading
    selector: Rules
    relation: guides
    read_policy: must
    used_for: prior-work and codebase inspection bounds for this plan
    status: active
  - id: guardrails
    name: Guardrails
    type: doc
    target: docs/agents/GUARDRAILS.md
    selector_type: heading
    selector: Completion Bar
    relation: constrains
    read_policy: must
    used_for: populated-plan requirements, validation expectations, and no-git-mutation safety boundary
    status: active
  - id: tooling-doc
    name: Tooling docs
    type: doc
    target: docs/agents/TOOLING.md
    selector_type: heading
    selector: Command Capability Discovery
    relation: guides
    read_policy: conditional
    used_for: capabilities documentation expectations and dispatch versus RLM boundary
    status: active
  - id: brainstorm
    name: Review loop brainstorm
    type: feature
    target: docs/specs/0034-review-loop/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, current dispatch review-intake findings, and selected command direction
    status: active
  - id: spec
    name: Review loop spec
    type: feature
    target: docs/specs/0034-review-loop/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding command, watch, triage, safety, documentation, and acceptance contract
    status: active
  - id: feature-map-command
    name: Feature map command
    type: command
    target: "kit map 0034-review-loop"
    selector_type: command
    selector: "kit map 0034-review-loop"
    relation: informs
    read_policy: evidence
    used_for: confirmed plan phase, feature relationships, and resolved references
    status: active
  - id: progress-summary
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    selector_type: heading
    selector: FEATURE PROGRESS TABLE
    relation: informs
    read_policy: evidence
    used_for: feature phase/status tracking and prior-feature shortlist
    status: active
  - id: dispatch-spec
    name: Dispatch command spec
    type: feature
    target: docs/specs/0008-dispatch-command/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: existing PR review-thread intake, CodeRabbit filtering, editor prefill, dedupe, and prompt-only dispatch contract
    status: active
  - id: dispatch-plan
    name: Dispatch command plan
    type: feature
    target: docs/specs/0008-dispatch-command/PLAN.md
    selector_type: heading
    selector: COMPONENTS
    relation: informs
    read_policy: conditional
    used_for: current dispatch component boundaries and editor/prompt pipeline precedent
    status: active
  - id: capabilities-spec
    name: Kit capabilities spec
    type: feature
    target: docs/specs/0033-kit-capabilities/SPEC.md
    selector_type: heading
    selector: REQUIREMENTS
    relation: constrains
    read_policy: must
    used_for: command capability metadata, search, targeted lookup, and read-only discovery expectations
    status: active
  - id: capabilities-plan
    name: Kit capabilities plan
    type: feature
    target: docs/specs/0033-kit-capabilities/PLAN.md
    selector_type: heading
    selector: COMPONENTS
    relation: informs
    read_policy: conditional
    used_for: static capability catalog and command registration/update precedent
    status: active
  - id: dispatch-command
    name: Dispatch command
    type: code
    target: pkg/cli/dispatch.go
    selector_type: symbol
    selector: dispatchCmd
    relation: informs
    read_policy: must
    used_for: current dispatch flags, command registration, editor input path, and prompt output pipeline
    status: active
  - id: dispatch-pr-intake
    name: Dispatch PR review intake
    type: code
    target: pkg/cli/dispatch_pr.go
    selector_type: symbol
    selector: loadDispatchPRInput
    relation: informs
    read_policy: must
    used_for: current PR target resolution, GitHub GraphQL review-thread fetch, pagination, and editor population behavior
    status: active
  - id: dispatch-pr-extraction
    name: Dispatch PR review extraction
    type: code
    target: pkg/cli/dispatch_pr_extract.go
    selector_type: symbol
    selector: extractDispatchReviewTasks
    relation: informs
    read_policy: must
    used_for: current resolved/outdated filtering, CodeRabbit author filtering, prompt extraction, shared-instruction handling, and dedupe
    status: active
  - id: dispatch-tests
    name: Dispatch tests
    type: code
    target: pkg/cli/dispatch_test.go
    selector_type: artifact
    selector: dispatch_test.go
    relation: verifies
    read_policy: evidence
    used_for: existing test fixtures and regression patterns for PR review intake
    status: active
  - id: root-help
    name: Root help
    type: code
    target: pkg/cli/root_help.go
    selector_type: symbol
    selector: rootCommandSections
    relation: constrains
    read_policy: must
    used_for: visible command ordering and command category placement
    status: active
  - id: capabilities-catalog
    name: Capabilities catalog
    type: code
    target: pkg/cli/capabilities_catalog.go
    selector_type: symbol
    selector: capabilityCatalog
    relation: constrains
    read_policy: must
    used_for: review-loop and dispatch alias metadata updates
    status: active
  - id: ci-github
    name: CI GitHub helpers
    type: code
    target: pkg/cli/ci_github.go
    selector_type: symbol
    selector: ciRunner
    relation: informs
    read_policy: conditional
    used_for: command-runner seam and GitHub CLI JSON parsing precedent for testable gh integrations
    status: active
  - id: github-check-runs-api
    name: GitHub Check Runs API
    type: url
    target: https://docs.github.com/en/rest/checks/runs
    relation: informs
    read_policy: evidence
    used_for: check-run status, conclusion, and ref-based review-completion signal
    status: active
  - id: github-pr-graphql
    name: GitHub GraphQL pull request reference
    type: url
    target: https://docs.github.com/en/graphql/reference/pulls
    relation: informs
    read_policy: evidence
    used_for: pull request review-thread fields and PR metadata for review task evidence
    status: active
  - id: coderabbit-review-commands
    name: CodeRabbit review commands
    type: url
    target: https://docs.coderabbit.ai/reference/review-commands
    relation: informs
    read_policy: evidence
    used_for: CodeRabbit review lifecycle signals and manual command behavior
    status: active
  - id: coderabbit-github-checks
    name: CodeRabbit GitHub Checks
    type: url
    target: https://docs.coderabbit.ai/tools/github-checks
    relation: informs
    read_policy: evidence
    used_for: CodeRabbit GitHub Checks timing and completion behavior
    status: active
---
# PLAN

## SUMMARY

Implement review-loop prompt preparation as a thin orchestration layer over the existing dispatch PR review intake: use `kit dispatch --loop` to resolve a PR, optionally wait for current-head CodeRabbit completion, fetch and classify current review-thread findings, prefill the same dispatch editor pipeline with only `FIX` items, and keep default execution read-only. Add an explicit `kit dispatch --pr <target> --resolve --yes` follow-up path for resolving already-handled review threads after fixes or no-op decisions are complete.

## APPROACH

- Keep dispatch's current PR target parsing, review-thread GraphQL pagination, CodeRabbit prompt extraction, shared-instruction handling, and dedupe as shared lower-level helpers; do not fork a second PR-review collector.
- Add a review-loop runner layer that is responsible for PR metadata, watch timing, classification, summary rendering, and dispatch command wiring.
- Treat `kit dispatch --loop` as the runnable prompt-prep command path.
- Use a small injectable command-runner seam for GitHub CLI calls and clock/sleeper behavior so watch-mode tests can run without real waiting or network calls.
- Keep v1 timing constants fixed by default: 90 second initial wait before polling, 15 second polling, 15 minute timeout, and 60 second quiet window after CodeRabbit appears complete.
- Use PR title/body, linked issue when available, and discoverable Kit feature docs as triage context, but keep classification output evidence-based and conservative.
- Keep classification conservative: uncertain or ambiguous findings become `NEEDS_HUMAN`, current but out-of-scope findings become `VALID_OUT_OF_SCOPE`, obsolete line/file findings become `STALE`, disproven findings become `FALSE_POSITIVE`, and only still-valid in-scope findings become `FIX`.
- Render a review summary before editor launch so skipped classifications stay visible while only `FIX` items enter the dispatch task block.
- Preserve the command's default read-only boundary: network reads, editor/clipboard output, and stdout/stderr only.
- Update help, README, agent-facing docs, and `kit capabilities` after command behavior is stable.

## COMPONENTS

- `pkg/cli/review_loop.go`
  - owns shared review-loop prompt-prep execution
- `pkg/cli/dispatch.go`
  - adds `--loop` as the alias trigger
  - adds `--resolve --yes` as the explicit review-thread resolution path
  - rejects incompatible dispatch inputs when `--loop` is set
  - routes alias execution to the same shared review-loop runner
- `pkg/cli/dispatch_pr_resolve.go`
  - collects current unresolved, non-outdated review threads matching dispatch filters
  - supports CodeRabbit-only filtering for resolution
  - resolves GitHub review threads through the GraphQL `resolveReviewThread` mutation only when `--yes` is present
- `pkg/cli/review_loop_types.go`
  - defines review-loop options, PR context, finding, classification, watch state, and summary records
  - keeps transient data local to command execution
- `pkg/cli/review_loop_github.go`
  - fetches PR metadata needed for current head SHA, title, body, URL, and linked issue hints
  - fetches CodeRabbit-related checks or review activity for watch mode
  - reuses dispatch PR target resolution and review-thread fetch helpers
- `pkg/cli/review_loop_watch.go`
  - implements initial wait, polling, timeout, completion detection, and quiet-window behavior behind injectable clock/sleep interfaces
  - reports timeout and unavailable-review states without producing stale work
- `pkg/cli/review_loop_triage.go`
  - classifies current findings into `FIX`, `VALID_OUT_OF_SCOPE`, `FALSE_POSITIVE`, `STALE`, and `NEEDS_HUMAN`
  - attaches concise reasons and source evidence
- `pkg/cli/review_loop_render.go`
  - renders classification summaries
  - builds the editor prefill from `FIX` findings and existing dispatch review-task rendering
  - preserves common CodeRabbit instruction handling
- `pkg/cli/review_loop_test.go`
  - covers command routing, watch timing, classification, rendering, no-action behavior, and read-only guarantees
- `pkg/cli/capabilities_catalog.go`
  - updates `dispatch` metadata for the `--loop` workflow
- `README.md`, `docs/agents/TOOLING.md`, and `docs/specs/0000_INIT_PROJECT.md`
  - document when to use `dispatch --loop` and `dispatch --pr --coderabbit`

## DATA

- `reviewLoopOptions`
  - PR reference
  - CodeRabbit-only mode
  - watch mode
  - dispatch output flags
  - editor/input configuration
  - max-subagent setting reused from dispatch prompt generation
- `reviewLoopPRContext`
  - repository owner/name
  - PR number and URL
  - PR title/body
  - current head SHA
  - optional linked issue hints
  - optional local feature-doc context
- `reviewLoopFinding`
  - source path
  - line or start line
  - author
  - comment URL
  - extracted task body
  - raw cleaned comment body when needed for evidence
  - normalized dedupe key
- `reviewLoopClassification`
  - enum values: `FIX`, `VALID_OUT_OF_SCOPE`, `FALSE_POSITIVE`, `STALE`, `NEEDS_HUMAN`
  - reason text
  - source evidence
  - PR-goal relationship
- `reviewLoopWatchState`
  - PR head SHA
  - CodeRabbit check/review activity status
  - latest observed completion time
  - timeout deadline
  - quiet-window deadline
- Persistent data
  - no new persistent artifact or hidden state
  - no `.kit.yaml` writes
  - no `.kit/` run or loop artifacts

## INTERFACES

- `kit dispatch --loop --pr <target> [--coderabbit] [--watch]`
  - network-read PR review workflow
  - waits with `--watch` using the fixed timing contract before collecting current review tasks
- `kit dispatch --pr <target> --resolve --yes [--coderabbit]`
  - explicit post-fix/no-op mutation path for resolving currently matching unresolved review threads
  - requires `--yes` and does not run as part of prompt generation
- Output behavior
  - prints a concise review-loop summary
  - opens the editor only when actionable `FIX` items exist
  - emits the normal dispatch prompt after the human reviews the editor content
  - preserves clipboard/output-only behavior from dispatch prompt output helpers
- Failure behavior
  - actionable errors for missing `--pr`, invalid PR target, missing GitHub remote, missing or unauthenticated `gh`, GitHub GraphQL/API errors, malformed responses, timeout, and unavailable current review state
- Side effects
  - allowed: GitHub network reads, editor temp file behavior, clipboard/stdout output
  - forbidden: project-file writes, `.kit.yaml` writes, git staging, commits, pushes, PR comments, review-thread resolution, branch mutation, or CodeRabbit command execution
  - exception: `kit dispatch --pr <target> --resolve --yes` may resolve matching GitHub review threads after the user has explicitly confirmed fixes or no-op decisions are complete

## DEPENDENCIES

References are tracked in front matter.

## RISKS

- CodeRabbit completion signals may be incomplete or inconsistent across repositories.
  - Mitigation: combine current PR head SHA, check/review activity, and quiet-window logic; report uncertainty as timeout or unavailable state instead of collecting stale work.
- Semantic triage can overstate confidence.
  - Mitigation: keep classification conservative, require reasons/evidence, and route uncertain findings to `NEEDS_HUMAN`.
- Alias behavior can drift from canonical command behavior.
  - Mitigation: make both command paths call the same runner and add equivalence tests.
- Dispatch PR intake can become coupled to editor behavior too early for review-loop classification.
  - Mitigation: split reusable fetch/extract/build helpers from editor launch before adding review-loop orchestration.
- Watch-mode tests can become slow or flaky.
  - Mitigation: inject clock/sleeper behavior and command-runner responses; never sleep in unit tests.
- The command could accidentally become mutating through future convenience flags.
  - Mitigation: document mutation boundaries in capability metadata and add tests that assert no git/GitHub mutation commands are invoked by default.
- Capability/help/docs can lag behind the command surface.
  - Mitigation: update capability catalog, root help, README, and agent-facing docs in the same implementation pass and cover visible command metadata with tests.

## TESTING

- Unit tests for PR target reuse:
  - URL, Markdown link, `owner/repo#number`, and current-repo number resolve through shared dispatch behavior.
- Unit tests for review-thread filtering and extraction:
  - resolved threads excluded
  - outdated threads excluded
  - CodeRabbit-only filtering
  - prompt block extraction
  - cleaned-comment fallback
  - shared boilerplate included once
  - duplicate findings collapsed
- Unit tests for watch mode:
  - initial 90 second wait
  - 15 second polling
  - 15 minute timeout
  - 60 second quiet window
  - current-head SHA enforcement
  - timeout and unavailable states produce no stale dispatch work
- Unit tests for triage:
  - all five classifications render with reasons/evidence
  - only `FIX` findings enter the editor task block
  - non-fix findings remain visible in the summary
- Command tests:
  - `kit review-loop --help` fails because the compatibility root command has been removed
  - `kit dispatch --help` includes `--loop`
  - `kit dispatch --loop` routes to the review-loop prompt-prep runner
  - invalid flag combinations fail actionably
  - no actionable feedback skips editor launch
- Read-only safety tests:
  - default execution does not run `git add`, `git commit`, `git push`, PR comment commands, review-thread resolve commands, or CodeRabbit mutation commands
  - default execution does not write `.kit.yaml`, `.kit/`, or project docs
- Resolve mutation tests:
  - `--resolve` requires `--yes`
  - resolution candidates include unresolved, non-outdated review threads
  - `--coderabbit` limits resolution candidates to CodeRabbit-authored threads
- Capability and documentation tests:
  - `kit capabilities review-loop --json` fails because the compatibility root command has been removed
  - `kit capabilities --search review-loop --json` finds the review-loop surface
- End verification:
  - `go test ./...`
  - `go run ./cmd/kit check 0034-review-loop`
  - `go run ./cmd/kit map 0034-review-loop`
  - `git diff --check`
