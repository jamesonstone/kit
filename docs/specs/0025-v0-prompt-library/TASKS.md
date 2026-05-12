---
kit_metadata_version: 1
artifact: "tasks"
feature:
  id: "0025"
  slug: "v0-prompt-library"
  dir: "0025-v0-prompt-library"
---
# TASKS

## PROGRESS TABLE

| ID   | TASK                                                                                     | STATUS | OWNER | DEPENDENCIES                         |
| ---- | ---------------------------------------------------------------------------------------- | ------ | ----- | ------------------------------------ |
| T001 | Run implementation readiness gate [PLAN-RISKS](PLAN.md#risks)                           | done   | agent |                                      |
| T002 | Add prompt config schema and global path helpers [PLAN-DATA](PLAN.md#data)               | done   | agent | T001                                 |
| T003 | Add prompt registry identity and merge package [PLAN-COMPONENTS](PLAN.md#components)     | done   | agent | T001                                 |
| T004 | Integrate prompt config loading and saving [PLAN-INTERFACES](PLAN.md#interfaces)         | done   | agent | T002, T003                           |
| T005 | Add static coding-agent built-in providers [PLAN-COMPONENTS](PLAN.md#components)         | done   | agent | T003                                 |
| T006 | Add Kit command built-in provider adapters [PLAN-COMPONENTS](PLAN.md#components)         | done   | agent | T003                                 |
| T007 | Add dynamic built-in context collection [PLAN-INTERFACES](PLAN.md#interfaces)            | done   | agent | T006                                 |
| T008 | Add prompt-library output helper [PLAN-APPROACH](PLAN.md#approach)                       | done   | agent | T003                                 |
| T009 | Implement `kit prompt` resolution and selectors [PLAN-INTERFACES](PLAN.md#interfaces)    | done   | agent | T004, T005, T006, T007, T008         |
| T010 | Implement `kit prompt list` table output [PLAN-INTERFACES](PLAN.md#interfaces)           | done   | agent | T004, T005, T006                     |
| T011 | Implement `kit set prompt` editing and scope writes [PLAN-INTERFACES](PLAN.md#interfaces) | done   | agent | T004, T008                           |
| T012 | Register commands and root help placement [PLAN-COMPONENTS](PLAN.md#components)          | done   | agent | T009, T010, T011                     |
| T013 | Add config and registry unit tests [PLAN-TESTING](PLAN.md#testing)                       | done   | agent | T004, T005, T006                     |
| T014 | Add prompt retrieval and list CLI tests [PLAN-TESTING](PLAN.md#testing)                  | done   | agent | T009, T010, T012                     |
| T015 | Add set-prompt, editor, and dynamic-provider tests [PLAN-TESTING](PLAN.md#testing)       | done   | agent | T007, T011, T012                     |
| T016 | Update README and configuration references [PLAN-INTERFACES](PLAN.md#interfaces)         | done   | agent | T012, T013, T014, T015               |
| T017 | Run full validation and reconcile task status [PLAN-TESTING](PLAN.md#testing)            | done   | agent | T013, T014, T015, T016               |

## TASK LIST

- [x] T001: Run implementation readiness gate [PLAN-RISKS](PLAN.md#risks)
- [x] T002: Add prompt config schema and global path helpers [PLAN-DATA](PLAN.md#data)
- [x] T003: Add prompt registry identity and merge package [PLAN-COMPONENTS](PLAN.md#components)
- [x] T004: Integrate prompt config loading and saving [PLAN-INTERFACES](PLAN.md#interfaces)
- [x] T005: Add static coding-agent built-in providers [PLAN-COMPONENTS](PLAN.md#components)
- [x] T006: Add Kit command built-in provider adapters [PLAN-COMPONENTS](PLAN.md#components)
- [x] T007: Add dynamic built-in context collection [PLAN-INTERFACES](PLAN.md#interfaces)
- [x] T008: Add prompt-library output helper [PLAN-APPROACH](PLAN.md#approach)
- [x] T009: Implement `kit prompt` resolution and selectors [PLAN-INTERFACES](PLAN.md#interfaces)
- [x] T010: Implement `kit prompt list` table output [PLAN-INTERFACES](PLAN.md#interfaces)
- [x] T011: Implement `kit set prompt` editing and scope writes [PLAN-INTERFACES](PLAN.md#interfaces)
- [x] T012: Register commands and root help placement [PLAN-COMPONENTS](PLAN.md#components)
- [x] T013: Add config and registry unit tests [PLAN-TESTING](PLAN.md#testing)
- [x] T014: Add prompt retrieval and list CLI tests [PLAN-TESTING](PLAN.md#testing)
- [x] T015: Add set-prompt, editor, and dynamic-provider tests [PLAN-TESTING](PLAN.md#testing)
- [x] T016: Update README and configuration references [PLAN-INTERFACES](PLAN.md#interfaces)
- [x] T017: Run full validation and reconcile task status [PLAN-TESTING](PLAN.md#testing)

## TASK DETAILS

### T001

- **GOAL**: Confirm the fixed docs are implementation-ready before writing production code.
- **SCOPE**:
  - Review `SPEC.md`, `PLAN.md`, and `TASKS.md` for contradictions, ambiguity, hidden assumptions, missing failure modes, and scope creep.
  - Keep `BRAINSTORM.md` as rationale only.
  - Update canonical docs first if the gate finds a material issue.
- **ACCEPTANCE**:
  - `kit check v0-prompt-library` exits 0 before implementation begins.
  - No unresolved gate finding remains in `SPEC.md`, `PLAN.md`, or `TASKS.md`.
  - Evidence: command output from `kit check v0-prompt-library` and any doc diff if corrections were required.
- **NOTES**: This task gates all code tasks.

### T002

- **GOAL**: Extend Kit configuration so prompt entries and global config paths are representable without breaking existing `.kit.yaml` files.
- **SCOPE**:
  - Add YAML-backed prompt entry types under `internal/config`.
  - Add optional project-root discovery for prompt retrieval outside Kit projects.
  - Add global config path helpers for `~/.config/kit/.kit.yaml`.
  - Preserve existing default config behavior.
- **ACCEPTANCE**:
  - Existing `.kit.yaml` files without `prompts` still load with defaults.
  - Global config absence is distinguishable from parse or permission errors.
  - Evidence: updated `internal/config` files and focused config tests introduced or updated in T013.
- **NOTES**: Do not create or mutate the real user global config during tests.

### T003

- **GOAL**: Add a testable prompt registry package for identity normalization, source precedence, resolution, and shadow metadata.
- **SCOPE**:
  - Create `internal/promptlib` with focused files for types, normalization, source loading, merging, resolving, sorting, and nearest-match helpers.
  - Normalize noun and verb inputs to lowercase kebab-case.
  - Preserve all lower-precedence matches needed for shadow/override display.
  - Reject invalid identities and normalized collisions with actionable errors.
- **ACCEPTANCE**:
  - Registry APIs can be tested without Cobra, clipboard, editor, or filesystem side effects.
  - Effective prompts are sorted by noun then verb.
  - Evidence: new `internal/promptlib` package and tests introduced or updated in T013.
- **NOTES**: Keep files short and single-purpose.

### T004

- **GOAL**: Connect config prompt storage to the prompt registry and implement safe prompt upserts for local and global scopes.
- **SCOPE**:
  - Load prompt sources from local `.kit.yaml`, global `.kit.yaml`, and built-in providers.
  - Return empty global prompt source when global config is absent.
  - Save prompt entries to local or global scope with object form `content` and optional `description`.
  - Preserve unrelated config fields and unknown future prompt metadata where practical.
- **ACCEPTANCE**:
  - Local, global, and absent-global load paths are covered.
  - Global save creates parent directory and file when needed.
  - Evidence: config integration code and T013 tests for load/save/upsert behavior.
- **NOTES**: Use temporary home/config paths in tests.

### T005

- **GOAL**: Expose the three static coding-agent toolbox prompts as built-in prompt providers.
- **SCOPE**:
  - Add `coding-agent short`.
  - Add `coding-agent long`.
  - Add `coding-agent instructions`.
  - Preserve prompt payload intent and content from the Karabiner script.
  - Exclude auto-paste, clipboard restore, sleeps, and AppleScript behavior.
- **ACCEPTANCE**:
  - The three toolbox prompts appear in the built-in catalog with descriptions.
  - Prompt bodies do not include shell automation behavior.
  - Evidence: built-in provider code and T013/T014 assertions for catalog presence and payload content.
- **NOTES**: Do not modify `/Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh`.

### T006

- **GOAL**: Expose current Kit command prompts through built-in providers without duplicating stale prompt text.
- **SCOPE**:
  - Add workflow providers for `workflow brainstorm`, `workflow spec`, `workflow plan`, `workflow tasks`, `workflow implement`, and `workflow reflect`.
  - Add support providers for `support resume`, `support handoff`, `support summarize`, `support reconcile`, `support dispatch`, and `support code-review`.
  - Add `skill mine` and `project init` providers.
  - Extract pure prompt-builder helpers only where existing code mixes building, output, and mutation.
- **ACCEPTANCE**:
  - Built-in providers delegate to existing prompt builders or pure extracted builders.
  - Existing prompt golden tests still pass or are intentionally updated for builder extraction only.
  - Evidence: provider adapter files, any pure-builder refactors, and prompt golden test output.
- **NOTES**: Providers must not call command `RunE` functions.

### T007

- **GOAL**: Add shared context inference and missing-context capture for dynamic built-in prompts.
- **SCOPE**:
  - Infer feature context for feature-scoped built-ins when possible.
  - Ask whether the user has missing required context when inference fails.
  - Capture user-supplied context through the existing vim-compatible editor flow.
  - Fail with actionable guidance when the user declines or context remains insufficient.
- **ACCEPTANCE**:
  - Dynamic providers declare the context they require.
  - Missing-context flows reuse `readEditorText` semantics, including save/cancel behavior.
  - Evidence: context helper code and T015 tests for infer, accept, decline, and insufficient-context paths.
- **NOTES**: Do not introduce stdin or `--file` input for `kit set prompt` as part of this task.

### T008

- **GOAL**: Add prompt-library output behavior that copies by default and visibly prints selected prompt content.
- **SCOPE**:
  - Copy the exact resolved prompt body by default.
  - Print command name, origin, and shadow metadata in default human-readable mode.
  - Print the prompt body in a clearly delimited block in default mode.
  - Preserve raw `--output-only` and `--output-only --copy` semantics.
  - Fail clearly if clipboard copying fails.
- **ACCEPTANCE**:
  - Default mode copies and prints metadata plus body.
  - `--output-only` prints raw prompt only and skips copy.
  - `--output-only --copy` prints and copies raw prompt text.
  - Evidence: new output helper code and T014 output tests.
- **NOTES**: Do not change existing prompt-output helpers for unrelated commands unless necessary.

### T009

- **GOAL**: Implement `kit prompt [noun] [verb]` retrieval with direct resolution and interactive selectors.
- **SCOPE**:
  - Register `kit prompt` with `--output-only` and `--copy` flags.
  - Implement no-arg noun selector and verb selector.
  - Implement noun-only verb selector.
  - Implement noun/verb direct lookup.
  - Show descriptions in selectors when available.
  - Return nearest noun or verb suggestions on no-match errors.
- **ACCEPTANCE**:
  - `kit prompt coding-agent short` resolves directly and uses the prompt-library output helper.
  - `kit prompt` and `kit prompt coding-agent` use deterministic selector ordering.
  - No-match paths fail fast with actionable suggestions.
  - Evidence: command implementation and T014 CLI tests.
- **NOTES**: `list` remains reserved for `kit prompt list`.

### T010

- **GOAL**: Implement `kit prompt list` as the deterministic discovery surface.
- **SCOPE**:
  - Render effective merged prompts by default.
  - Include command name, description, and shadow/overriding columns.
  - Sort by noun then verb.
  - Keep output readable in TTY and non-TTY contexts.
- **ACCEPTANCE**:
  - Local-over-global-over-built-in entries show correct shadow metadata.
  - Built-in workflow/support providers appear in list output.
  - Evidence: command implementation and T014 table-output tests.
- **NOTES**: Do not add a `--source` selector in v0.

### T011

- **GOAL**: Implement `kit set` and `kit set prompt [noun] [verb]` for editor-backed prompt creation and updates.
- **SCOPE**:
  - Register `kit set` with `prompt` as the only v0 resource.
  - Route naked `kit set` to the prompt-setting wizard.
  - Support `--local`, `--global`, and `--local --global`.
  - Default to local inside a Kit project.
  - Prompt for global save outside a Kit project when no scope flag is present.
  - Use one editor capture for dual-scope writes.
  - Confirm overwrites separately for local and global scopes.
- **ACCEPTANCE**:
  - `kit set prompt custom review` inside a Kit project saves locally through the editor flow.
  - `kit set prompt custom review --global` creates global config when needed.
  - `kit init` can also create or populate the global config defaults.
  - `--local` outside a Kit project fails with actionable guidance.
  - Declined overwrite confirmations skip only the declined scope and cancel when no scopes remain.
  - Evidence: command implementation and T015 tests for scope, overwrite, save, and cancel behavior.
- **NOTES**: Do not support stdin or `--file` for setting prompt content.

### T012

- **GOAL**: Make the new command surfaces discoverable through command registration and root help.
- **SCOPE**:
  - Register `prompt`, `prompt list`, `set`, and `set prompt` under Cobra.
  - Place `prompt` in the Prompt Utilities group.
  - Place or describe `set` so prompt-setting is discoverable without implying broader v0 resource support.
  - Ensure unsupported v0 flags are absent from help.
- **ACCEPTANCE**:
  - `kit --help` shows `prompt` in Prompt Utilities.
  - `kit --help` or relevant command help makes `set prompt` discoverable.
  - `kit prompt --help` and `kit set prompt --help` show supported flags only.
  - Evidence: root help updates and T014/T015 help tests.
- **NOTES**: Keep root help concise.

### T013

- **GOAL**: Cover config and registry behavior with focused unit tests.
- **SCOPE**:
  - Test config defaults and backward-compatible loads.
  - Test prompt YAML schema with missing, empty, and unknown metadata cases.
  - Test global absent, global present, and global save creation paths.
  - Test registry precedence, sorting, collisions, and nearest-match suggestions.
- **ACCEPTANCE**:
  - `go test ./internal/config ./internal/promptlib` passes.
  - Tests cover all PLAN testing bullets for config and registry behavior.
  - Evidence: test files and focused test command output.
- **NOTES**: Use temp directories and isolated HOME/config paths.

### T014

- **GOAL**: Cover prompt retrieval, output flags, selectors, list output, and help behavior with CLI tests.
- **SCOPE**:
  - Test direct retrieval for `coding-agent short`, `long`, and `instructions`.
  - Test default copy plus visible body output.
  - Test `--output-only` and `--output-only --copy`.
  - Test noun and verb selector ordering and descriptions.
  - Test no-match errors and suggestions.
  - Test `kit prompt list` table columns and shadow metadata.
  - Test root and prompt command help placement.
- **ACCEPTANCE**:
  - Focused `pkg/cli` tests pass for prompt retrieval and list behavior.
  - Clipboard is stubbed in tests; no real clipboard dependency is required.
  - Evidence: CLI test files and focused `go test ./pkg/cli` output.
- **NOTES**: Keep non-TTY assertions stable.

### T015

- **GOAL**: Cover set-prompt editing, overwrite confirmation, and dynamic provider context behavior with CLI tests.
- **SCOPE**:
  - Test local, global, dual-scope, and outside-project save paths.
  - Test per-scope overwrite confirmation behavior.
  - Test editor save, unchanged/cancel, missing editor, and empty content behavior.
  - Test dynamic built-in context inference, editor capture, user decline, and insufficient-context failures.
- **ACCEPTANCE**:
  - Focused `pkg/cli` tests pass for `kit set prompt` and dynamic provider paths.
  - Tests do not mutate real project or user global config files.
  - Evidence: CLI test files and focused `go test ./pkg/cli` output.
- **NOTES**: Reuse existing editor stubs where possible.

### T016

- **GOAL**: Update user-facing documentation and configuration references to match the implemented prompt library.
- **SCOPE**:
  - Update `README.md` for `kit prompt`, `kit prompt list`, and `kit set prompt` usage.
  - Update configuration reference docs with the prompt schema and precedence rules.
  - Update generated project/template references when `.kit.yaml` schema examples are duplicated there.
  - Document output semantics and non-goals: no auto-paste, no clipboard restore, no `--source`, no `--no-copy`.
- **ACCEPTANCE**:
  - Docs describe local/global/built-in precedence and YAML shape.
  - Docs include examples for retrieval, listing, and setting prompts.
  - Evidence: documentation diffs and successful `kit check v0-prompt-library` after docs updates.
- **NOTES**: Keep docs concise and avoid adding a separate summary document.

### T017

- **GOAL**: Run final validation and leave feature docs synchronized with implementation reality.
- **SCOPE**:
  - Run full repository tests.
  - Run feature validation.
  - Update `TASKS.md` checkboxes and progress table statuses as tasks complete.
  - Update docs first if validation reveals behavior that differs from `SPEC.md`, `PLAN.md`, or `TASKS.md`.
- **ACCEPTANCE**:
  - `go test ./...` passes.
  - `kit check v0-prompt-library` exits 0 with no unexpected warnings.
  - `kit status` shows task progress consistent with `TASKS.md` checkboxes.
  - Evidence: command outputs and final docs diff.
- **NOTES**: Do not commit changes unless explicitly requested.

## DEPENDENCIES

- No external blockers remain.
- T001 gates all production code changes.
- T002-T004 establish storage and registry foundations before CLI work.
- T005-T008 establish provider and output foundations before retrieval commands.
- T009-T012 implement public command surfaces after foundations exist.
- T013-T015 verify the core behavior before documentation sign-off.
- T016 documents only implemented behavior.
- T017 is the final validation and synchronization gate.

## NOTES

- Use temporary directories and isolated environment variables for global config tests.
- Do not mutate the real `~/.config/kit/.kit.yaml` during tests.
- Do not modify the Karabiner script; treat it as a source reference only.
- Do not add new third-party dependencies unless the plan is updated first.
- Keep implementation files focused and short; split prompt-library package files by concept.
