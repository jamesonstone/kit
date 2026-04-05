# PLAN

## SUMMARY

- Add a prompt-only `dispatch` command that captures a raw task block, normalizes top-level tasks, and outputs a deterministic subagent-dispatch planning prompt.
- Reuse Kit's existing clipboard-first prompt-output and editor-input patterns, with a vim-compatible editor as the default interactive capture path.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-16][SPEC-27] Add a new `pkg/cli/dispatch.go` command with standard prompt flags, input-source precedence, default editor-backed capture, pre-editor instructions, any-key launch gating, file/stdin support, and max-subagent validation.
- [PLAN-02][SPEC-14][SPEC-15] Add focused task-normalization helpers that split only top-level paragraphs, bullets, and numbered items into dispatchable units while preserving nested detail under the parent task.
- [PLAN-03][SPEC-17][SPEC-18][SPEC-19][SPEC-20][SPEC-21][SPEC-22][SPEC-23][SPEC-24][SPEC-25][SPEC-26] Build a dedicated prompt builder that embeds the normalized task list and enforces discovery-first clustering, conservative overlap handling, dry-run reporting, and approval gating before subagent launch.
- [PLAN-04][SPEC-28] Register the new command in help ordering and README so the public CLI surface matches the shipped behavior.
- [PLAN-05] Add focused tests for input-source precedence, task normalization, prompt invariants, and flag validation, then run the standard verification commands.
- [PLAN-06] Switch `dispatch` to the shared clipboard-first helper that preserves dispatch's no-subagent-suffix prompt shape.

## COMPONENTS

- `pkg/cli/dispatch.go`
  - command registration
  - flag handling
  - input-source resolution
  - prompt output
- `pkg/cli/dispatch_input.go`
  - file/stdin/editor capture helpers
  - input-source precedence
  - pre-editor instruction screen
  - any-key confirmation before editor launch
- `pkg/cli/dispatch_tasks.go`
  - task normalization
  - list-item detection
- `pkg/cli/dispatch_prompt.go`
  - prompt builder
  - normalized-task rendering
- `pkg/cli/dispatch_test.go`
  - prompt tests
  - task-normalization tests
  - input precedence and validation tests
- `pkg/cli/root.go`
  - help ordering
- `README.md`
  - command docs

## DATA

- Input data comes from one of:
  - `--file`
  - piped stdin
  - default editor-backed capture
  - editor override capture
- The command introduces no new persistent artifact type.
- The normalized task list is transient and exists only to shape the generated prompt.

## INTERFACES

- New command:
  - `kit dispatch`
- Flags:
  - `--copy`
  - `--output-only`
  - `--file`
  - `--vim`
  - `--editor`
  - `--max-subagents`
- Output shape:
  - prompt-only, passed through the shared clipboard-first helper without subagent suffix text
  - workflow footer via `printWorkflowInstructions(...)`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| prompt helper | code | `pkg/cli/root.go`, `pkg/cli/implement.go`, `pkg/cli/reflect.go` | clipboard-first prompt output | active |
| editor input helpers | code | `pkg/cli/editor_input.go` | interactive task capture | active |
| root help ordering | code | `pkg/cli/root.go` | command visibility | active |
| README | doc | `README.md` | public command documentation | active |

## RISKS

- Default interactive capture depends on a vim-compatible editor being available, so missing-editor errors must stay actionable.
- The pre-editor keypress flow can become flaky if it depends on terminal state that is not restored cleanly.
- Task normalization can over-split if nested bullets are mistaken for top-level tasks.
- The prompt can encourage unsafe parallelization if overlap ambiguity is not handled conservatively.
- File/stdin/editor precedence can become surprising if not encoded and tested explicitly.
- Clipboard-first output can accidentally append unrelated orchestration text if dispatch stops using its dedicated no-subagent prompt path.

## TESTING

- Add unit tests for:
  - prompt content invariants
  - top-level task normalization across paragraphs, bullets, and numbered items
  - nested-item attachment to parent tasks
  - input-source precedence
  - `--max-subagents` validation
- Add or reuse unit tests for clipboard-first prompt output semantics.
- Run:
  - `make vet`
  - `make test`
  - `make build`
  - `./bin/kit dispatch --help`
  - `./bin/kit --help`
