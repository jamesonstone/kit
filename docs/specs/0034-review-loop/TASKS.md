---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0034
  slug: review-loop
  dir: 0034-review-loop
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Extract reusable PR review intake helpers [PLAN-COMPONENTS] | done | agent | |
| T002 | Add review-loop command and dispatch alias routing [PLAN-INTERFACES] | done | agent | T001 |
| T003 | Implement GitHub PR metadata and CodeRabbit watch services [PLAN-COMPONENTS] | done | agent | T001, T002 |
| T004 | Implement conservative review finding triage [PLAN-DATA] | done | agent | T001, T003 |
| T005 | Render summaries and dispatch editor input [PLAN-INTERFACES] | done | agent | T001, T004 |
| T006 | Update capability metadata, help, and docs [PLAN-COMPONENTS] | done | agent | T002, T005 |
| T007 | Add focused command, watch, triage, and safety tests [PLAN-TESTING] | done | agent | T001, T002, T003, T004, T005, T006 |
| T008 | Run final verification and reconcile docs [PLAN-TESTING] | done | agent | T007 |
| T009 | Add explicit review-thread resolution follow-up [PLAN-INTERFACES] | done | agent | T001, T006 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Extract reusable PR review intake helpers [PLAN-COMPONENTS]
- [x] T002: Add review-loop command and dispatch alias routing [PLAN-INTERFACES]
- [x] T003: Implement GitHub PR metadata and CodeRabbit watch services [PLAN-COMPONENTS]
- [x] T004: Implement conservative review finding triage [PLAN-DATA]
- [x] T005: Render summaries and dispatch editor input [PLAN-INTERFACES]
- [x] T006: Update capability metadata, help, and docs [PLAN-COMPONENTS]
- [x] T007: Add focused command, watch, triage, and safety tests [PLAN-TESTING]
- [x] T008: Run final verification and reconcile docs [PLAN-TESTING]
- [x] T009: Add explicit review-thread resolution follow-up [PLAN-INTERFACES]

## TASK DETAILS

### T001
- **GOAL**: Split dispatch PR review intake so review-loop can reuse fetch/extract/dedupe behavior without launching the editor too early.
- **SCOPE**:
  - Reorganize `pkg/cli/dispatch_pr.go` and `pkg/cli/dispatch_pr_extract.go` around reusable PR target resolution, thread fetch, task extraction, and rendered-task helpers.
  - Preserve existing `kit dispatch --pr` and `--coderabbit` behavior.
  - Keep GraphQL pagination, CodeRabbit prompt extraction, shared-instruction stripping, and dedupe semantics unchanged.
- **ACCEPTANCE**:
  - Existing dispatch PR tests still pass.
  - Review-loop can call a non-editor helper that returns extracted review tasks and common instruction data.
  - No user-visible dispatch behavior changes outside intentional helper refactoring.
- **VERIFY**:
  - `go test ./pkg/cli -run 'Test.*Dispatch.*PR|TestBuildDispatchPRInput|TestResolveDispatchPRTarget|TestSplitDispatchPRInput' -count=1`
- **EXPECTED FILES**:
  - `pkg/cli/dispatch_pr.go`
  - `pkg/cli/dispatch_pr_extract.go`
  - `pkg/cli/dispatch_test.go`
- **RISK**: Medium; refactoring existing dispatch intake can regress current PR review workflows.
- **ROLLBACK**: Revert the helper split and restore direct editor launch path.
- **NOTES**: Keep this task behavior-preserving before adding review-loop orchestration.

### T002
- **GOAL**: Add the public dispatch review-loop prompt-prep surface.
- **SCOPE**:
  - Add `pkg/cli/review_loop.go`.
  - Add `--loop` to `dispatch` as the prompt-prep trigger.
  - Reject incompatible dispatch inputs when `--loop` is set.
  - Share output flags, editor flags, CodeRabbit filtering, watch mode, and max-subagent settings where applicable.
- **ACCEPTANCE**:
  - `kit dispatch --help` exposes `--loop`.
  - `kit dispatch --loop` calls the review-loop prompt-prep runner path.
  - Missing `--pr` and invalid flag combinations fail with actionable errors.
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestReviewLoop.*Command|TestDispatch.*Loop' -count=1`
  - `go run ./cmd/kit dispatch --help`
- **EXPECTED FILES**:
  - `pkg/cli/review_loop.go`
  - `pkg/cli/dispatch.go`
  - `pkg/cli/root_help.go`
  - `pkg/cli/review_loop_test.go`
- **RISK**: Medium; alias routing can drift or conflict with dispatch input precedence.
- **ROLLBACK**: Remove the `dispatch --loop` flag.
- **NOTES**: Do not implement GitHub mutation flags in this task.

### T003
- **GOAL**: Add testable GitHub metadata and CodeRabbit watch services for current-head review-loop execution.
- **SCOPE**:
  - Add `pkg/cli/review_loop_types.go`, `pkg/cli/review_loop_github.go`, and `pkg/cli/review_loop_watch.go`.
  - Fetch PR metadata including number, URL, title, body, current head SHA, and linked issue hints.
  - Fetch CodeRabbit-related check or review activity for the current head.
  - Implement injectable command-runner and clock/sleeper seams.
  - Enforce 90 second initial wait, 15 second polling, 15 minute timeout, and 60 second quiet window.
- **ACCEPTANCE**:
  - Watch mode never accepts CodeRabbit state from an older PR head SHA.
  - Timeout and unavailable states return clear errors and do not collect stale tasks.
  - Unit tests use fake runner/clock; they do not sleep in real time.
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestReviewLoop.*Watch|TestReviewLoop.*GitHub|TestReviewLoop.*Head' -count=1`
- **EXPECTED FILES**:
  - `pkg/cli/review_loop_types.go`
  - `pkg/cli/review_loop_github.go`
  - `pkg/cli/review_loop_watch.go`
  - `pkg/cli/review_loop_test.go`
- **RISK**: High; CodeRabbit completion signals can vary across repositories.
- **ROLLBACK**: Keep command available without `--watch` until watch service is corrected.
- **NOTES**: Use conservative failure states instead of falling back to stale review data.

### T004
- **GOAL**: Classify current review findings into the five required triage outcomes with evidence.
- **SCOPE**:
  - Add `pkg/cli/review_loop_triage.go`.
  - Classify findings as `FIX`, `VALID_OUT_OF_SCOPE`, `FALSE_POSITIVE`, `STALE`, or `NEEDS_HUMAN`.
  - Use PR title/body, linked issue hints, and discoverable Kit feature docs as available context.
  - Preserve file path, line, author, comment URL, and reason text for each classification.
  - Default ambiguous findings to `NEEDS_HUMAN`.
- **ACCEPTANCE**:
  - All five classifications are covered by unit tests.
  - Missing local file/line evidence routes to `STALE` or `NEEDS_HUMAN` as appropriate.
  - Classification output includes reason and evidence for every finding.
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestReviewLoop.*Triage|TestReviewLoop.*Classification' -count=1`
- **EXPECTED FILES**:
  - `pkg/cli/review_loop_triage.go`
  - `pkg/cli/review_loop_test.go`
- **RISK**: High; automated triage can overstate certainty.
- **ROLLBACK**: Downgrade uncertain classifications to `NEEDS_HUMAN` while preserving fetched evidence.
- **NOTES**: Keep triage conservative and deterministic.

### T005
- **GOAL**: Render review-loop summaries and feed only `FIX` items into the dispatch editor/prompt pipeline.
- **SCOPE**:
  - Add `pkg/cli/review_loop_render.go`.
  - Render concise summary output before editor launch.
  - Include non-fix classifications in summary output.
  - Build editor prefill from `FIX` findings only.
  - Reuse dispatch review-task rendering, common CodeRabbit instruction handling, task normalization, and prompt output helpers.
  - Skip editor launch when no actionable `FIX` findings remain.
- **ACCEPTANCE**:
  - Only `FIX` findings enter the dispatch task block.
  - Non-fix findings remain visible in summary output.
  - No-actionable-feedback output is clear and does not open the editor.
  - User-cancelled or cleared editor content fails actionably.
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestReviewLoop.*Render|TestReviewLoop.*Editor|TestReviewLoop.*NoActionable' -count=1`
- **EXPECTED FILES**:
  - `pkg/cli/review_loop_render.go`
  - `pkg/cli/review_loop_test.go`
- **RISK**: Medium; summary/editor rendering can duplicate or hide important review context.
- **ROLLBACK**: Revert to dispatch's existing PR editor rendering for actionable tasks.
- **NOTES**: Do not write project files or `.kit/` artifacts.

### T006
- **GOAL**: Make the dispatch review-loop workflow discoverable in Kit command metadata and user/agent docs.
- **SCOPE**:
  - Update dispatch capability metadata for `--loop`.
  - Update README command docs.
  - Update `docs/agents/TOOLING.md`.
  - Update `docs/specs/0000_INIT_PROJECT.md` if it documents command capabilities or generated guidance.
- **ACCEPTANCE**:
  - `kit capabilities review-loop --json` fails because the compatibility root command has been removed.
  - `kit capabilities --search review-loop --json` finds the review-loop surface.
  - Docs distinguish `dispatch --loop` and `dispatch --pr --coderabbit`.
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestCapabilities|TestRootHelp' -count=1`
  - `go run ./cmd/kit capabilities --search review-loop --json`
  - `go run ./cmd/kit --help`
- **EXPECTED FILES**:
  - `pkg/cli/capabilities_catalog.go`
  - `pkg/cli/capabilities_test.go`
  - `pkg/cli/root_help.go`
  - `pkg/cli/root_help_test.go`
  - `README.md`
  - `docs/agents/TOOLING.md`
  - `docs/specs/0000_INIT_PROJECT.md`
- **RISK**: Medium; docs and capability metadata can drift from the actual command surface.
- **ROLLBACK**: Remove review-loop metadata/docs until command behavior is complete.
- **NOTES**: Update documentation after command behavior is stable enough to describe accurately.

### T007
- **GOAL**: Complete focused regression coverage for command behavior, watch behavior, triage, rendering, and safety.
- **SCOPE**:
  - Add or complete `pkg/cli/review_loop_test.go`.
  - Cover PR target reuse for all accepted forms.
  - Cover resolved/outdated filtering, CodeRabbit filtering, prompt extraction, fallback cleanup, boilerplate dedupe, and duplicate collapse.
  - Cover watch timing and current-head SHA enforcement.
  - Cover triage classifications and summary/editor separation.
  - Cover alias equivalence.
  - Cover read-only safety by asserting default execution does not invoke git mutation, PR comment, review-thread resolution, or CodeRabbit mutation commands.
- **ACCEPTANCE**:
  - Focused tests cover every `ACCEPT-*` item in `SPEC.md` that can be proven without live GitHub.
  - Fake runner/clock/editor paths keep tests deterministic.
  - No tests require real GitHub, real CodeRabbit, real sleeping, or real editor interaction.
- **VERIFY**:
  - `go test ./pkg/cli -run 'TestReviewLoop|TestDispatch|TestCapabilities|TestRootHelp' -count=1`
- **EXPECTED FILES**:
  - `pkg/cli/review_loop_test.go`
  - `pkg/cli/dispatch_test.go`
  - `pkg/cli/capabilities_test.go`
  - `pkg/cli/root_help_test.go`
- **RISK**: Medium; broad command behavior can be under-tested if fake seams are too shallow.
- **ROLLBACK**: Narrow implementation until each behavior has deterministic local coverage.
- **NOTES**: Prefer assertions on command arguments and rendered outputs over brittle full-string snapshots.

### T008
- **GOAL**: Prove the completed implementation and docs satisfy the feature contract.
- **SCOPE**:
  - Run full Go test suite.
  - Run feature map/check commands.
  - Run whitespace checks.
  - Inspect changed docs for stale placeholders.
  - Update `docs/PROJECT_PROGRESS_SUMMARY.md` if implementation changes the feature phase or summary.
- **ACCEPTANCE**:
  - Full verification commands pass or failures are documented with exact blocker state.
  - `TASKS.md` checkboxes reflect completed task status after implementation.
  - Project progress summary remains aligned with the highest completed artifact.
- **VERIFY**:
  - `go test ./...`
  - `go run ./cmd/kit check 0034-review-loop`
  - `go run ./cmd/kit map 0034-review-loop`
  - `git diff --check`
- **EXPECTED FILES**:
  - `docs/specs/0034-review-loop/TASKS.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Low; final verification may uncover earlier incomplete task coverage.
- **ROLLBACK**: Reopen the failed task and leave final task unchecked until evidence passes.
- **NOTES**: Do not claim CI or live GitHub behavior passed unless it was actually observed.

### T009
- **GOAL**: Add an explicit post-fix/no-op review-thread resolution path for `kit dispatch --pr`.
- **SCOPE**:
  - Add `kit dispatch --pr <target> --resolve --yes`.
  - Resolve currently matching unresolved, non-outdated GitHub review threads through GraphQL.
  - Preserve `--coderabbit` filtering for resolution candidates.
  - Require `--yes` so resolution cannot happen accidentally during prompt generation.
  - Keep default `dispatch --pr` and `dispatch --loop` behavior read-only.
- **ACCEPTANCE**:
  - `--resolve` fails without `--yes`.
  - Resolution candidates skip resolved and outdated threads.
  - `--coderabbit` limits resolution candidates to CodeRabbit-authored review threads.
  - Help, capabilities, README, agent docs, and init spec document the mutation boundary.
- **VERIFY**:
  - `go test ./pkg/cli -run 'Test.*Dispatch|TestReviewLoop|TestCapabilities|TestRootHelp' -count=1`
  - `go run ./cmd/kit dispatch --help`
  - `go run ./cmd/kit capabilities dispatch --json`
- **EXPECTED FILES**:
  - `pkg/cli/dispatch.go`
  - `pkg/cli/dispatch_pr.go`
  - `pkg/cli/dispatch_pr_resolve.go`
  - `pkg/cli/dispatch_pr_resolve_test.go`
  - `pkg/cli/capabilities_catalog.go`
- `pkg/cli/capabilities_test.go`
- `README.md`
- `docs/agents/TOOLING.md`
- `docs/specs/0000_INIT_PROJECT.md`
- **RISK**: Medium; review-thread resolution is a GitHub mutation and must not run unless the user explicitly confirms handled feedback.
- **ROLLBACK**: Remove `--resolve`/`--yes`, the resolution helper, and the related docs/capability metadata.
- **NOTES**: This is a follow-up mutation path, not part of default review-loop prompt generation.

## DEPENDENCIES

- Execute tasks in numeric order; each task depends on the reusable review intake and command wiring established before it.
- T003, T004, and T005 may be developed in separate low-overlap passes after T001 and T002, but T007 must integrate and verify them together.
- No external blocker or missing product decision remains.

## NOTES

- Keep all implementation work in the existing project directory; do not create or use git worktrees.
- Preserve the default read-only mutation boundary throughout implementation.
- Do not run CodeRabbit mutation commands, resolve GitHub review threads, post PR comments, stage files, commit, or push as part of this feature implementation unless separately approved through the repo-local delivery workflow.

<!-- REFLECTION_COMPLETE -->
