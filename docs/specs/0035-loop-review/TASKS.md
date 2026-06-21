---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0035
  slug: loop-review
  dir: 0035-loop-review
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Refactor loop command group [PLAN-COMPONENTS-01] | done | agent | |
| T002 | Implement review runner [PLAN-COMPONENTS-02] | done | agent | T001 |
| T003 | Update capabilities and docs [PLAN-COMPONENTS-03][PLAN-COMPONENTS-04] | done | agent | T001,T002 |
| T004 | Add tests and run verification [PLAN-COMPONENTS-05] | done | agent | T001,T002,T003 |
| T005 | Backfill loop agent config through init refresh [PLAN-COMPONENTS-06] | done | agent | T003 |
| T006 | Harden rerun and agent setup behavior [PLAN-COMPONENTS-02][PLAN-COMPONENTS-03] | done | agent | T002,T003 |

## TASK LIST

- [x] T001: Refactor loop command group [PLAN-COMPONENTS-01]
- [x] T002: Implement review runner [PLAN-COMPONENTS-02]
- [x] T003: Update capabilities and docs [PLAN-COMPONENTS-03][PLAN-COMPONENTS-04]
- [x] T004: Add tests and run verification [PLAN-COMPONENTS-05]
- [x] T005: Backfill loop agent config through init refresh [PLAN-COMPONENTS-06]
- [x] T006: Harden rerun and agent setup behavior [PLAN-COMPONENTS-02][PLAN-COMPONENTS-03]

## TASK DETAILS

### T001
- **GOAL**: Make `kit loop` a command group while preserving legacy workflow loop usage.
- **SCOPE**: Add `workflow` and `review` subcommands, retain `kit loop [feature]`, and set default iterations to 10.
- **ACCEPTANCE**: Legacy and new workflow command paths route to the same runner.
- **VERIFY**: `go test ./pkg/cli -run 'TestLoop.*Command|TestExecuteLoop' -count=1`
- **EXPECTED FILES**: `pkg/cli/loop.go`, `internal/config/config.go`
- **RISK**: Medium; command routing can regress existing users.
- **ROLLBACK**: Restore flat `kit loop [feature]` command.
- **NOTES**: Keep old flags available on the legacy path.

### T002
- **GOAL**: Implement the changed-code review loop.
- **SCOPE**: Add review diff discovery, prompts, result parsing, agent execution, artifacts, and PR feedback polling.
- **ACCEPTANCE**: The loop stops only on `done` plus sufficient correctness and handles default PR pending status.
- **VERIFY**: `go test ./pkg/cli -run 'TestLoopReview' -count=1`
- **EXPECTED FILES**: `pkg/cli/loop_review.go`, `pkg/cli/loop_review_test.go`
- **RISK**: High; autonomous repair loops need strict mutation boundaries.
- **ROLLBACK**: Remove `loop review` subcommand.
- **NOTES**: Do not stage, commit, push, or mutate GitHub.

### T003
- **GOAL**: Keep command docs and discovery metadata accurate.
- **SCOPE**: Update capabilities, README, agent tooling docs, and init spec command reference.
- **ACCEPTANCE**: `kit capabilities loop review --json` returns accurate metadata.
- **VERIFY**: `go test ./pkg/cli -run TestCapabilities -count=1`
- **EXPECTED FILES**: `pkg/cli/capabilities_catalog.go`, `README.md`, `docs/agents/TOOLING.md`, `docs/specs/0000_INIT_PROJECT.md`
- **RISK**: Medium; stale command metadata misleads agents.
- **ROLLBACK**: Revert metadata/docs changes with command removal.
- **NOTES**: Mark legacy prompt-prep command separately from the new repair loop.

### T004
- **GOAL**: Prove behavior locally.
- **SCOPE**: Add focused tests and run full suite.
- **ACCEPTANCE**: Tests pass or exact blockers are reported.
- **VERIFY**: `go test ./...`
- **EXPECTED FILES**: `pkg/cli/loop_review_test.go`, `pkg/cli/capabilities_test.go`
- **RISK**: Low; environment may lack Go on PATH.
- **ROLLBACK**: Fix failing tests before completion.
- **NOTES**: Use `/opt/homebrew/bin/go` if `go` is not on PATH.

### T005
- **GOAL**: Make `kit init --refresh` install the config required by `kit loop review`.
- **SCOPE**: Seed default init config with a Codex stdin loop agent and backfill existing `.kit.yaml` files when `loop.agent.command` is blank, `your-agent`, or a known generated default.
- **ACCEPTANCE**: Refresh adds the default loop agent config and preserves custom loop agent commands.
- **VERIFY**: `go test ./pkg/cli -run 'TestRunInit.*Loop|TestCapabilitiesTargetedJSON' -count=1`
- **EXPECTED FILES**: `pkg/cli/init.go`, `pkg/cli/init_refresh.go`, `pkg/cli/init_test.go`, `pkg/cli/capabilities_catalog.go`
- **RISK**: Medium; config refresh must avoid overwriting intentionally custom agent commands.
- **ROLLBACK**: Remove the init refresh loop config backfill.
- **NOTES**: The default command is `codex exec` reading the loop prompt from stdin.

### T006
- **GOAL**: Make review-loop reruns and agent setup failures explicit.
- **SCOPE**: Prompt interactive users before rerunning when prior review-loop evidence exists or max iterations are reached, stop immediately on agent command failures with stderr context, keep review prompts single-agent by default, add opt-in subagent pre-analysis, and stream runner/agent progress to stderr.
- **ACCEPTANCE**: Invalid agent commands stop after one failed iteration, rerun confirmation accepts `y`/`n` shorthand, review prompts omit subagent guidance by default, `--subagents` adds parent pre-analysis and shared subagent guidance, and human-readable runs expose progress plus child-agent stdout/stderr.
- **VERIFY**: `go test ./pkg/cli -run 'TestLoopReview|TestReadLoopReviewConfirmation|TestLatestLoopReviewReport' -count=1`
- **EXPECTED FILES**: `pkg/cli/loop_review.go`, `pkg/cli/loop_review_test.go`, `pkg/cli/capabilities_catalog.go`, `pkg/cli/capabilities_test.go`
- **RISK**: Medium; interactive prompts must not affect JSON, dry-run, or non-terminal automation.
- **ROLLBACK**: Remove rerun prompting and immediate command-failure stop behavior.
- **NOTES**: Keep the required final output contract as the final review prompt section.

## DEPENDENCIES

- T002 depends on T001.
- T003 depends on T001 and T002.
- T004 depends on T001 through T003.
- T005 depends on T003.
- T006 depends on T002 and T003.

## NOTES

Implementation follows the clarification contract from the planning discussion.
