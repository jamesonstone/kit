---
kit_metadata_version: 1
artifact: "plan"
feature:
  id: "0028"
  slug: "project-refresh-advisory"
  dir: "0028-project-refresh-advisory"
relationships:
  - type: "builds_on"
    target: "0017-reconcile-command"
  - type: "builds_on"
    target: "0025-v0-prompt-library"
  - type: "related_to"
    target: "0027-implement-readiness-gate"
references:
  - name: prompt library
    type: code
    target: pkg/cli/prompt_builtin_kit.go, pkg/cli/prompt_builtin_render.go
    relation: implements
    read_policy: conditional
    used_for: built-in prompt registration and runtime rendering
    status: active
  - name: project refresh prompt
    type: code
    target: pkg/cli/project_refresh_prompt.go
    relation: implements
    read_policy: conditional
    used_for: new prompt body and advisory text
    status: active
  - name: reflect command
    type: code
    target: pkg/cli/reflect.go
    relation: implements
    read_policy: conditional
    used_for: late workflow advisory gate
    status: active
  - name: complete command
    type: code
    target: pkg/cli/complete.go
    relation: implements
    read_policy: conditional
    used_for: post-completion advisory output
    status: active
  - name: README
    type: doc
    target: README.md
    relation: guides
    read_policy: conditional
    used_for: user-facing command guidance
    status: active
  - name: init project spec
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: informs
    read_policy: conditional
    used_for: canonical product behavior summary
    status: active
---
# PLAN

## SUMMARY

- Add a manual project refresh prompt and wire late workflow commands to recommend it as a soft advisory gate.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13] Add `project refresh` to the built-in prompt catalog and implement a docs-only prompt builder.
- [PLAN-02][SPEC-14][SPEC-15][SPEC-16] Add a project-refresh advisory step to the reflection prompt.
- [PLAN-03][SPEC-17][SPEC-18][SPEC-19] Print a one-time non-blocking project refresh advisory after successful completion.
- [PLAN-04][SPEC-20] Update README and the core project spec with the manual refresh command and soft advisory behavior.
- [PLAN-05][SPEC-21][SPEC-22][SPEC-23][SPEC-24] Add focused tests for prompt catalog registration, prompt content, reflection prompt content, and completion output.

## COMPONENTS

- `pkg/cli/prompt_builtin_kit.go`
  - register `project refresh`
- `pkg/cli/prompt_builtin_render.go`
  - add runtime adapter for project context
- `pkg/cli/project_refresh_prompt.go`
  - prompt body
  - shared advisory text
- `pkg/cli/reflect.go`
  - reflection advisory gate step
- `pkg/cli/complete.go`
  - post-completion advisory output
- `pkg/cli/prompt_builtin_kit_test.go`
  - built-in prompt catalog and render assertions
- `pkg/cli/prompt_golden_test.go` and `pkg/cli/testdata/reflect_feature_prompt.golden`
  - reflection prompt contract
- `pkg/cli/complete_test.go`
  - completion advisory assertion
- `README.md`
  - command guidance
- `docs/specs/0000_INIT_PROJECT.md`
  - canonical command behavior

## DATA

- No new persisted state.
- Inputs are current filesystem-backed project documents and code:
  - `.kit.yaml`
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
  - repository instruction docs
  - `docs/specs/`
  - current git status and diffs

## INTERFACES

- New prompt-library identity:
  - `kit prompt project refresh`
- Existing workflow commands gain advisory wording only:
  - `kit reflect [feature]`
  - `kit complete [feature]`
  - `kit complete --all`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| prompt library | code | `pkg/cli/prompt_builtin_kit.go, pkg/cli/prompt_builtin_render.go` | built-in prompt registration and runtime rendering | active |
| project refresh prompt | code | `pkg/cli/project_refresh_prompt.go` | new prompt body and advisory text | active |
| reflect command | code | `pkg/cli/reflect.go` | late workflow advisory gate | active |
| complete command | code | `pkg/cli/complete.go` | post-completion advisory output | active |
| README | doc | `README.md` | user-facing command guidance | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical product behavior summary | active |

## RISKS

- The refresh prompt can duplicate `kit reconcile --all` if it tries to enumerate structural findings itself.
- The advisory can become noisy if it reads like a mandatory warning after every late workflow step.
- The prompt can invite broad rewrites unless it strongly limits updates to durable project-level truth.
- A top-level `kit project refresh` command may be useful later, but adding it now would expand the visible command surface before the prompt behavior proves itself.

## TESTING

- Assert the built-in prompt catalog includes `project refresh`.
- Assert the rendered prompt includes:
  - `/plan`
  - docs-only scope
  - `docs/CONSTITUTION.md`
  - `kit reconcile --all`
  - `kit rollup`
  - `kit check --project`
- Assert the reflection golden prompt includes the project refresh advisory gate.
- Assert completion output prints the refresh advisory once for a batch.
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
