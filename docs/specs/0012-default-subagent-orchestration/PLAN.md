# PLAN

## SUMMARY

- Flip Kit's shared orchestration model to subagents-by-default.
- Add `--single-agent` as the documented opt-out while keeping `kit dispatch` as the stricter queue-planning surface.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-10] Update the shared prompt helper to make subagent guidance default-on, add `--single-agent`, and keep a hidden compatibility alias for `--subagents`.
- [PLAN-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Rewrite the shared orchestration suffix so it defaults to subagents while preserving conservative overlap handling and main-agent ownership.
- [PLAN-03][SPEC-08] Verify that `dispatch` still uses the dedicated no-shared-subagent path.
- [PLAN-04][SPEC-09] Update README and help-facing wording to reflect the new default and opt-out flag.
- [PLAN-05] Add focused tests for default suffix behavior, `--single-agent`, legacy flag registration, and dispatch isolation, then rerun verification.

## COMPONENTS

- `pkg/cli/subagents.go`
  - persistent flag registration
  - default prompt augmentation
  - shared orchestration suffix wording
- `pkg/cli/subagents_test.go`
  - default/opt-out tests
  - flag registration tests
- `pkg/cli/output_test.go`
  - dispatch no-suffix regression coverage
- `README.md`
  - public command behavior docs

## DATA

- No new persistent state.
- Root persistent flag surface changes from documented `--subagents` opt-in to documented `--single-agent` opt-out.
- `--subagents` may remain as a hidden compatibility alias only.

## INTERFACES

- Shared prompt commands:
  - subagent orchestration guidance is on by default
  - `--single-agent` disables the shared orchestration suffix
- `kit dispatch`:
  - still outputs only its dedicated dispatch-planning prompt
  - still requires explicit approval before any subagent launch

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| shared prompt helper | code | /Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli/subagents.go | implement the default orchestration switch | active |
| dispatch output helper | code | /Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli/root.go | preserve dispatch's no-shared-suffix behavior | active |

## RISKS

- A default-on suffix can accidentally alter dispatch output if the dedicated no-suffix path regresses.
- Help text can become misleading if the documented flag does not match the shipped persistent flag surface.
- Over-aggressive wording could encourage unsafe parallelism unless conservative overlap handling stays explicit.

## TESTING

- Verify default prompt augmentation includes `## Subagent Orchestration`.
- Verify `--single-agent` registration and behavior through the shared helper tests.
- Verify the hidden compatibility alias for `--subagents` remains available.
- Verify dispatch output still omits the shared suffix.
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
