# PLAN

## SUMMARY

Implement the prompt library as a small layered registry that sits between Kit's YAML configuration and the Cobra command surface. User prompts will be loaded from local and global `.kit.yaml` files, built-in prompts will be exposed through provider adapters, and `kit prompt` will resolve one effective prompt by precedence before copying and rendering it. `kit set prompt` will mutate only the selected YAML scope through the existing editor flow, while existing Kit command prompts remain owned by their current prompt builders to avoid stale duplicate prompt text.

## APPROACH

- [PLAN-01][SPEC-11][SPEC-12][SPEC-17] Extend the config layer additively so existing `.kit.yaml` files continue to load, global config can be absent, and prompt entries are read from `prompts.<noun>.<verb>`.
- [PLAN-02][SPEC-15][SPEC-16][SPEC-61] Normalize prompt identities once at the boundary using lowercase kebab-case, reject empty or colliding normalized identities, and sort all registry views deterministically.
- [PLAN-03][SPEC-17][SPEC-18][SPEC-20] Build an effective prompt registry by merging built-in, global, and local sources in that order while retaining lower-precedence matches for shadow/override reporting.
- [PLAN-04][SPEC-43][SPEC-52] Keep built-in prompt text in provider adapters under `pkg/cli`, not in config, so static toolbox prompts are embedded once and current Kit workflow/support prompts delegate to existing builders.
- [PLAN-05][SPEC-29][SPEC-33] Add a prompt-library output helper instead of changing `writePromptWithClipboardDefault`; the new helper must copy the resolved prompt body by default, print origin and shadow metadata, then print the body in a delimited block.
- [PLAN-06][SPEC-38][SPEC-42] Implement prompt editing through the existing `readEditorText` path and keep stdin and `--file` out of v0.
- [PLAN-07][SPEC-53][SPEC-55] Model built-in prompts that need runtime context as providers with explicit context requirements; infer context where existing command helpers can, otherwise ask whether the user can provide it and capture it through the editor.
- [PLAN-08][SPEC-56][SPEC-63] Integrate `prompt`, `prompt list`, `set`, and `set prompt` into the human-readable CLI surface with table/list output and root-help placement.
- [PLAN-09][SPEC-64][SPEC-67] Update docs and tests alongside the implementation so command behavior, YAML schema, precedence, output flags, editor semantics, help grouping, and dynamic context paths are mechanically covered.
- Planning metadata: `parallelization_mode: "rlm"`.
- Tradeoff decisions: use nested YAML instead of per-prompt files; reserve `kit prompt list` as the discovery command; do not add a source selector, non-copy flag, stdin setter, file setter, auto-paste, or clipboard restore path in v0.

## COMPONENTS

- `internal/config`
  - Add prompt schema fields and validation-compatible load behavior.
  - Add optional project-root discovery so `kit prompt` can run outside a Kit project.
  - Add global config path helpers for `~/.config/kit/.kit.yaml`.
  - Allow `kit init` to populate missing default fields in the global config without replacing prompts.
  - Add prompt upsert helpers that create global config when requested and preserve unrelated YAML fields when practical.
- `internal/promptlib`
  - Own prompt identity normalization, source precedence, effective prompt merging, shadow metadata, deterministic ordering, nearest-match suggestions, and validation.
  - Keep files focused by concept: types, normalization, registry merge, resolver, and matching helpers.
  - Expose small structs and functions that can be tested without Cobra, clipboard, or editor dependencies.
- `pkg/cli/prompt.go`
  - Register `kit prompt [noun] [verb]`.
  - Route no-arg and noun-only invocations through numbered selectors.
  - Resolve noun/verb invocations directly and return actionable no-match errors.
- `pkg/cli/prompt_list.go`
  - Register `kit prompt list`.
  - Render the effective merged registry as a deterministic table with command name, description, and shadow/overriding information.
- `pkg/cli/prompt_builtin_toolbox.go`
  - Define static built-ins for `coding-agent short`, `coding-agent long`, and `coding-agent instructions` from the Karabiner script's prompt payloads only.
- `pkg/cli/prompt_builtin_workflow.go`
  - Register workflow, support, skill, and project built-ins as adapters over existing prompt builders.
  - Refactor existing prompt-building functions only as needed to return prompt text without writing output or mutating docs.
- `pkg/cli/prompt_context.go`
  - Centralize feature selection, project detection, free-text context collection, and editor-backed missing-context prompts used by dynamic built-ins.
- `pkg/cli/prompt_library_output.go`
  - Implement default copy-plus-visible-body output and raw `--output-only` behavior for prompt-library retrieval.
- `pkg/cli/set.go` and `pkg/cli/set_prompt.go`
  - Register `kit set` and `kit set prompt [noun] [verb]`.
  - Implement the no-arg wizard, scope resolution, per-scope overwrite confirmation, editor capture, and local/global save behavior.
- `pkg/cli/root_help.go`
  - Add `prompt` to Prompt Utilities and place `set` where prompt-setting is discoverable without expanding unrelated configuration scope.
- Documentation and templates
  - Update README command docs, configuration reference docs, and any generated project references that describe `.kit.yaml`.
- Tests
  - Add focused tests in `internal/config`, `internal/promptlib`, and `pkg/cli`; extend existing prompt, output, editor, and root-help tests instead of creating a separate test harness.

## DATA

- Prompt storage stays in YAML.
  - Local: project-root `.kit.yaml`.
  - Global: `~/.config/kit/.kit.yaml`.
  - Built-in: embedded provider catalog in code.
- YAML prompt shape:
  - `prompts` is a map keyed by normalized noun.
  - Each noun maps to normalized verbs.
  - Each verb maps to an object with required `content` and optional `description`.
  - Unknown future metadata fields must not make reads fail unless they conflict with required fields.
- Prompt identity:
  - `noun` and `verb` are normalized to lowercase kebab-case.
  - Empty normalized values are invalid.
  - Inputs that normalize to the same noun/verb within one source are collisions and must fail with the config path and prompt identity.
  - `list` is reserved for `kit prompt list` in v0.
- Source model:
  - `builtin` has lowest precedence.
  - `global` overrides built-in when no local prompt exists.
  - `local` overrides global and built-in.
  - Effective prompts retain a list of shadowed lower-precedence sources for display.
- Effective prompt record:
  - command name as `<noun> <verb>`.
  - normalized noun and verb.
  - description from the winning source.
  - origin source and config path when applicable.
  - content provider for static or dynamic prompt text.
  - shadow summary such as `local overrides global, builtin`.
- Built-in catalog:
  - Toolbox: `coding-agent short`, `coding-agent long`, `coding-agent instructions`.
  - Workflow: `workflow brainstorm`, `workflow spec`, `workflow plan`, `workflow tasks`, `workflow implement`, `workflow reflect`.
  - Support: `support resume`, `support handoff`, `support summarize`, `support reconcile`, `support dispatch`, `support code-review`.
  - Other: `skill mine`, `project init`.
- Dynamic built-ins produce final prompt text at resolution time.
  - Feature-scoped providers use existing feature selection and prompt builders.
  - Free-text providers reuse the editor-backed input path.
  - Providers must not mutate repository docs when invoked through `kit prompt`.

## INTERFACES

- `kit prompt`
  - No args: load effective registry, select noun, then select verb.
  - One arg: normalize noun, then select an effective verb under that noun.
  - Two args: normalize noun and verb, resolve directly, and fail fast if no effective prompt exists.
  - `list`: render effective prompt table instead of selecting a prompt.
  - `--output-only`: print only raw prompt content and skip default clipboard copy.
  - `--output-only --copy`: print raw prompt content and copy the same raw content.
  - `--copy` without `--output-only` is accepted for consistency but does not change the default copy behavior.
- Default `kit prompt <noun> <verb>` output:
  - copy exactly the resolved prompt body.
  - print concise human-readable acknowledgement.
  - print command name, origin, and shadow/override metadata.
  - print the selected prompt body in a clearly delimited block.
- `kit prompt list`
  - shows only effective merged prompts by default.
  - columns: command name, description, shadow/overriding.
  - sorted by noun then verb.
- `kit set`
  - no subcommand delegates to the prompt-setting wizard because `prompt` is the only v0 resource.
  - future resources can be added as subcommands without changing prompt storage.
- `kit set prompt [noun] [verb]`
  - no noun/verb: wizard prompts for noun, verb, optional description, save scope, and prompt content.
  - noun/verb: normalize arguments and open the editor for prompt content.
  - no scope flags inside a Kit project: save locally.
  - no scope flags outside a Kit project: ask whether to save globally; cancel if declined.
  - `--local`: require a project `.kit.yaml`; fail outside a Kit project.
  - `--global`: save to `~/.config/kit/.kit.yaml`, creating parent directory and file when needed.
  - `--local --global`: capture content once and write identical content to both target scopes.
  - overwrites are confirmed per scope before save; declined scopes are skipped, and the command cancels if no target scope remains.
- Config integration:
  - Existing `config.FindProjectRoot()` remains strict for commands that require a Kit project.
  - New optional project-root lookup supports prompt retrieval outside projects.
  - Global config load returns an empty prompt set when absent and fails with actionable parse errors when present but invalid.
  - Init-time global config population creates missing global config defaults and preserves existing prompt entries.
- Prompt-builder integration:
  - Existing command-specific prompt builders remain the source of truth.
  - If a builder currently mixes prompt construction with output or document mutation, extract a pure builder and keep the original command behavior as a caller of that builder.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Constitution | doc | docs/CONSTITUTION.md | filesystem-only state, command constraints, config reference, actionable errors | active |
| Agent routing docs | doc | AGENTS.md, docs/agents/README.md, docs/agents/WORKFLOWS.md, docs/agents/GUARDRAILS.md | formal workflow routing, required plan content, completion bar | active |
| RLM guide | doc/skill | docs/agents/RLM.md | scoped prior-work and code-pattern discovery; `parallelization_mode: "rlm"` | active |
| BRAINSTORM | doc | docs/specs/0025-v0-prompt-library/BRAINSTORM.md | resolved hierarchy, CLI behavior, editor, list, and overwrite decisions | active |
| SPEC | doc | docs/specs/0025-v0-prompt-library/SPEC.md | fixed requirements and acceptance contract for implementation | active |
| PLAN template | doc | docs/specs/0025-v0-prompt-library/PLAN.md | target implementation strategy artifact | active |
| kit map output | command output | `kit map 0025-v0-prompt-library` | current phase, relationship verification, dependency baseline | active |
| Project progress summary | doc | docs/PROJECT_PROGRESS_SUMMARY.md | prior-feature index and current feature status | active |
| Brainstorm-first workflow spec | prior feature doc | docs/specs/0004-brainstorm-first-workflow/SPEC.md | editor-default free-text flow and clipboard-first workflow prompt precedent | active |
| Dispatch command spec | prior feature doc | docs/specs/0008-dispatch-command/SPEC.md | editor-backed capture and dispatch prompt provider behavior | active |
| Support command clipboard defaults spec | prior feature doc | docs/specs/0010-support-command-clipboard-defaults/SPEC.md | `--output-only` and `--output-only --copy` semantics | active |
| Command surface simplification spec | prior feature doc | docs/specs/0019-command-surface-simplification/SPEC.md | root help grouping and prompt-utility placement | active |
| Typed prompt IR spec | prior feature doc | docs/specs/0022-typed-prompt-ir/SPEC.md | structured prompt builders and prompt-producing command inventory | active |
| Karabiner prompt script | external script | /Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh | source payloads for `coding-agent` built-ins; shell automation excluded | active |
| Project config | config | .kit.yaml | local prompt storage and schema extension point | active |
| Global Kit config | config | ~/.config/kit/.kit.yaml | global prompt storage target; created or populated by `kit init`, and may still be absent before init or first global save | active |
| Config package | code | internal/config/config.go | config schema, strict and optional project discovery, save/load helpers | active |
| Prompt output helpers | code | pkg/cli/prompt_output.go, pkg/cli/clipboard.go, pkg/cli/output_test.go | clipboard behavior, raw output flags, copy failure behavior | active |
| Editor helpers | code | pkg/cli/editor_input.go, pkg/cli/editor_input_test.go | vim-compatible prompt entry and missing-context capture | active |
| Human output and selectors | code | pkg/cli/human_output.go, pkg/cli/spec_selection.go | numbered selection UI and terminal-readable presentation | active |
| Root help | code | pkg/cli/root_help.go, pkg/cli/root_help_test.go | visible command grouping for `prompt` and `set` | active |
| Prompt IR | code | internal/promptdoc/doc.go, pkg/cli/prompt_ir_helpers.go | structured prompt construction for built-in adapters | active |
| Existing prompt builders | code | pkg/cli/*_prompt.go, pkg/cli/spec_output.go, pkg/cli/plan.go, pkg/cli/tasks.go, pkg/cli/implement.go, pkg/cli/reflect.go, pkg/cli/resume.go, pkg/cli/handoff.go, pkg/cli/summarize.go, pkg/cli/code_review.go, pkg/cli/skill.go, pkg/cli/init.go | built-in prompt providers without stale duplication | active |
| Prompt golden tests | tests | pkg/cli/prompt_golden_test.go, pkg/cli/testdata/*.golden | regression coverage for existing built-in prompt text | active |
| README and config references | docs | README.md, docs/CONSTITUTION.md, docs/specs/0000_INIT_PROJECT.md, internal/templates/templates.go | user-facing command docs and generated configuration guidance | active |
| Feature notes | notes | docs/notes/0025-v0-prompt-library | checked optional input; no usable files beyond `.gitkeep` | optional |

## RISKS

- Built-in prompt duplication can drift from existing commands.
  - Mitigation: use provider adapters over existing prompt builders; extract pure builders where needed; extend golden tests.
- Existing config save behavior can drop unknown YAML fields.
  - Mitigation: use targeted YAML-node prompt upserts for `kit set prompt` and test preservation of unrelated root keys and unknown prompt metadata.
- Commands that currently require a project root could make `kit prompt` fail outside a Kit project.
  - Mitigation: add optional root discovery for prompt retrieval and keep strict discovery for commands that mutate project docs.
- Dynamic built-ins can accidentally mutate documents if they reuse command runners.
  - Mitigation: providers call pure builders and context resolvers, never command `RunE` functions with side effects.
- Missing runtime context can become vague or inconsistent across built-ins.
  - Mitigation: each provider declares required context fields and shares one editor-backed missing-context flow.
- Shadow/override output can be misleading if merge state is discarded too early.
  - Mitigation: carry all lower-precedence source records into the effective prompt model and test local/global/built-in combinations.
- `kit prompt list` conflicts with a potential `list` noun.
  - Mitigation: reserve `list` for the v0 discovery command and reject it as a noun with actionable guidance.
- Default copy plus printed body can produce long terminal output.
  - Mitigation: this is required by the spec; keep raw/script mode available through `--output-only`.
- Clipboard failures can leave users thinking a prompt was copied.
  - Mitigation: copy first, fail clearly on copy errors, and print body only after successful default copy unless raw output mode is selected.

## TESTING

- Config tests:
  - default config remains backward compatible.
  - local prompt schema loads with and without `prompts`.
  - global config absent returns an empty prompt source.
  - init-time global config population creates missing defaults without overwriting prompts.
  - global save creates `~/.config/kit/.kit.yaml`.
  - prompt upsert preserves unrelated config fields.
  - invalid empty content, invalid IDs, and normalized collisions fail with actionable paths.
- Prompt registry tests:
  - built-in-only, global-over-built-in, local-over-global, and local-over-global-over-built-in precedence.
  - effective shadow summaries for all override combinations.
  - deterministic sorting by noun and verb.
  - nearest noun/verb suggestions for no-match errors.
  - unknown future metadata does not break reads.
- `kit prompt` CLI tests:
  - direct lookup copies and prints metadata plus body.
  - no-arg and noun-only selector flows show descriptions.
  - missing noun or verb fails fast with suggestions.
  - `--output-only` prints raw body and skips copy.
  - `--output-only --copy` prints and copies raw body.
  - `kit prompt list` renders the required table columns.
- Built-in provider tests:
  - toolbox prompts match the Karabiner payloads and exclude auto-paste/clipboard-restore behavior.
  - workflow/support/skill/project built-ins are registered under the required taxonomy.
  - dynamic built-ins use existing prompt builders and preserve profile, skills, and subagent suffix behavior where applicable.
  - context-required built-ins infer context when available, ask when missing, accept editor-entered context, and fail when declined.
- `kit set prompt` tests:
  - default local save inside a Kit project.
  - no-scope outside-project prompt for global save.
  - `--local` outside a Kit project fails.
  - `--global` creates missing global config.
  - `--local --global` edits once and saves identical content to both scopes.
  - local and global overwrites require separate confirmations.
  - declined overwrite skips that scope and cancels when no scopes remain.
  - editor unchanged, editor cancelled, missing editor, and empty saved prompt paths fail or cancel according to existing editor semantics.
- Help and docs tests:
  - root help includes `prompt` and discoverable `set` placement.
  - command help exposes supported flags without unsupported v0 flags.
  - README/config reference examples match the implemented schema.
- Verification after implementation:
  - run focused package tests while developing.
  - run `go test ./...`.
  - run `kit check v0-prompt-library`.
  - update golden files only when prompt text changes are intentional and traceable to this plan.
