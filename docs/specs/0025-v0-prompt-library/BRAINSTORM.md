---
kit_metadata_version: 1
artifact: "brainstorm"
feature:
  id: "0025"
  slug: "v0-prompt-library"
  dir: "0025-v0-prompt-library"
relationships:
  - type: "builds_on"
    target: "0004-brainstorm-first-workflow"
  - type: "builds_on"
    target: "0010-support-command-clipboard-defaults"
  - type: "builds_on"
    target: "0022-typed-prompt-ir"
  - type: "related_to"
    target: "0008-dispatch-command"
  - type: "related_to"
    target: "0019-command-surface-simplification"
dependencies:
  - name: "Feature notes"
    type: "notes"
    location: "docs/notes/0025-v0-prompt-library"
    used_for: "optional pre-brainstorm research input"
    status: "optional"
  - name: "Constitution"
    type: "doc"
    location: "docs/CONSTITUTION.md"
    used_for: "project invariants, config/state constraints, required brainstorm sections"
    status: "active"
  - name: "Repo agent routing docs"
    type: "doc"
    location: "AGENTS.md, CLAUDE.md, .github/copilot-instructions.md, docs/agents/README.md"
    used_for: "scoped loading and repo-local source-of-truth routing"
    status: "active"
  - name: "Workflow, RLM, guardrails, tooling docs"
    type: "doc"
    location: "docs/agents/WORKFLOWS.md, docs/agents/RLM.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md"
    used_for: "spec-driven classification, just-in-time prior-work pass, completion bar"
    status: "active"
  - name: "kit map output"
    type: "command output"
    location: "kit map 0025-v0-prompt-library"
    used_for: "current feature state, relationship/dependency baseline"
    status: "active"
  - name: "Project progress summary"
    type: "doc"
    location: "docs/PROJECT_PROGRESS_SUMMARY.md"
    used_for: "prior-feature index and candidate shortlist"
    status: "active"
  - name: "Brainstorm-first workflow spec"
    type: "prior feature doc"
    location: "docs/specs/0004-brainstorm-first-workflow/SPEC.md"
    used_for: "editor-default brainstorm flow and clipboard-first prompt contract"
    status: "active"
  - name: "Dispatch command spec"
    type: "prior feature doc"
    location: "docs/specs/0008-dispatch-command/SPEC.md"
    used_for: "editor-backed default capture and prompt-only utility precedent"
    status: "active"
  - name: "Support command clipboard defaults spec"
    type: "prior feature doc"
    location: "docs/specs/0010-support-command-clipboard-defaults/SPEC.md"
    used_for: "shared clipboard-first output contract"
    status: "active"
  - name: "Command surface simplification spec"
    type: "prior feature doc"
    location: "docs/specs/0019-command-surface-simplification/SPEC.md"
    used_for: "root help grouping and visible command-surface constraints"
    status: "active"
  - name: "Typed prompt IR spec"
    type: "prior feature doc"
    location: "docs/specs/0022-typed-prompt-ir/SPEC.md"
    used_for: "structured prompt-building boundary and prompt-producing command inventory"
    status: "active"
  - name: "Karabiner prompt script"
    type: "external script"
    location: "/Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh"
    used_for: "source content and behavior for existing one-off prompts"
    status: "active"
  - name: "Project config"
    type: "config"
    location: ".kit.yaml"
    used_for: "local config schema and project prompt storage target"
    status: "active"
  - name: "Global Kit config"
    type: "config"
    location: "/Users/jamesonstone/.config/kit/.kit.yaml"
    used_for: "intended global prompt storage target; currently absent"
    status: "optional"
  - name: "Clipboard helper"
    type: "code"
    location: "pkg/cli/prompt_output.go, pkg/cli/clipboard.go, pkg/cli/output_test.go"
    used_for: "copy/default-output behavior for prompt retrieval"
    status: "active"
  - name: "Editor helper"
    type: "code"
    location: "pkg/cli/editor_input.go, pkg/cli/editor_input_test.go"
    used_for: "editor-backed `kit set prompt` capture"
    status: "active"
  - name: "Prompt IR and prompt-producing commands"
    type: "code"
    location: "internal/promptdoc/doc.go, pkg/cli/prompt_ir_helpers.go, pkg/cli/*_prompt.go, pkg/cli/prompt_golden_test.go"
    used_for: "prompt registry scope and built-in prompt handling"
    status: "active"
  - name: "Root help and human output helpers"
    type: "code"
    location: "pkg/cli/root_help.go, pkg/cli/human_output.go"
    used_for: "command placement and readable terminal feedback"
    status: "active"
---
# BRAINSTORM

## SUMMARY

`v0-prompt-library` should add a first-class prompt registry and retrieval surface around `kit prompt <noun> <verb>` plus an extensible `kit set prompt <noun> <verb>` editor flow. The likely direction is a layered registry with built-in Kit prompts, optional global prompts in `~/.config/kit/.kit.yaml`, and optional project prompts in the project root `.kit.yaml`, with clear origin and precedence shown to the user.

## USER THESIS

build a new feature called "prompt library" and accessible with the command `kit prompt <noun> <verb>`. This command will operate in two ways. It will have a "toolbox" functionality where I can store a series of "one-off prompts" that I use during the normal course of work (right now those are defined in /Users/jamesonstone/.config/karabiner/scripts where I have 'long', 'short', and 'instruction')  so, for example, `kit prompt coding-agent instructions` would, by default, copy to clipboard the "instructions prmopt" command from that file and output a simple message to the command line (including emojis) to make it visible to the user the prompt that was copied. We should also output a copy of that prompt to the command line so the user can ensure they've selected the correct prompt. As a part of this feature, we need to come up with an ordering/heirarchy that makes it easy to comprehend and predict where useful prompts will end up. This feature should house every prompt we currently use for all commands in kit, and make it easy to add additional prompts using the `kit set prompt <noun> <verb>` which will then open the vim editor (as is our current pattern). `kit` should allow for saving prompts at the project-root `.kit.yaml` as well as the system root (i.e. `/Users/jamesonstone/.config/kit/.kit.yaml`) location. `kit set` should be extendable but, for right now, it will only work on prompts and it will always have the option to set `--local` (project root; default if the current directory contains a `.kit.yaml` otherwise it should pre-prompt the user that the current directory isn't a kit project, and to either set one or save with `--global`) and `--global` (system root; parameter). We should include an "interactive functionality" by default so that if `kit set` is used without options, we open instructions, interactive editor open, and an options of where to save the command (system or project). Both of these commands need to be highly intuitive and interactive by default (when no flags or options are present). We should use existing patterns where possible.

## RELATIONSHIPS

- builds on: 0004-brainstorm-first-workflow
- builds on: 0010-support-command-clipboard-defaults
- builds on: 0022-typed-prompt-ir
- related to: 0008-dispatch-command
- related to: 0019-command-surface-simplification

## CODEBASE FINDINGS

1. `docs/CONSTITUTION.md` defines Kit as filesystem-backed only: markdown plus `.kit.yaml`, explicit state, no hidden databases, no external state, and actionable failures. A prompt library should therefore stay in YAML and embedded code, not a database or daemon.
2. `internal/config/config.go` currently supports only project-root discovery via `FindProjectRoot()` and project config loading via `Load(projectRoot)`. It has no global config path, no `~/.config/kit` directory creation, and no prompt-related fields.
3. The project `.kit.yaml` currently contains workflow settings only: `goal_percentage`, `specs_dir`, `skills_dir`, `constitution_path`, `allow_out_of_order`, `agents`, `instruction_scaffold_version`, and `feature_naming`. Adding prompts requires an additive schema extension that preserves existing configs.
4. `/Users/jamesonstone/.config/kit/.kit.yaml` does not exist yet, and `/Users/jamesonstone/.config/kit` does not exist. The feature needs a create-if-missing path for global prompt saves.
5. `pkg/cli/prompt_output.go` centralizes output behavior. `writePromptWithClipboardDefault` copies by default, prints a concise acknowledgement, prints raw text only with `--output-only`, and supports `--output-only --copy`.
6. `pkg/cli/clipboard.go` currently shells to `pbcopy`; there is no cross-platform clipboard abstraction beyond macOS and no existing Kit behavior that auto-pastes or restores the previous clipboard.
7. `/Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh` defines the existing external one-off prompts: `short`, `long`, and `coding-agent-instructions`. That script prepends `---`, copies to clipboard, auto-pastes via AppleScript, sleeps, and restores the old clipboard. The user thesis asks Kit to copy and print instead, so auto-paste/restore should be treated as out of scope unless explicitly requested.
8. `pkg/cli/editor_input.go` provides reusable editor-backed free-text capture with `--vim`, `--editor`, default vim-compatible editor resolution (`nvim`, `vim`, `vi`), launch instructions, press-any-key gating, temp-file editing, and save/quit cancellation semantics. `kit set prompt` should reuse this instead of introducing a new editor path.
9. `pkg/cli/dispatch.go` and `pkg/cli/dispatch_input.go` show the nearest interactive no-args pattern for free-form input: file/stdin/editor source precedence and editor-backed default capture. `kit set prompt` has a different target model but can reuse the same editor and instruction style.
10. `pkg/cli/root_help.go` groups visible commands into Setup, Workflow, Inspect & Repair, Prompt Utilities, and Utilities. `kit prompt` most likely belongs in Prompt Utilities; `kit set` needs a placement decision because it is a mutating configuration command but supports prompt utilities only in v0.
11. `internal/promptdoc/doc.go` is the typed prompt IR used by prompt-producing commands. `pkg/cli/prompt_ir_helpers.go` exposes helpers such as `renderPromptDocument`. Any built-in prompt registry that houses current Kit prompts should avoid regressing to unstructured string construction.
12. `pkg/cli/subagents.go`, `pkg/cli/skills_prompt.go`, and `pkg/cli/prompt_profile.go` append shared prompt suffixes after command-specific prompt rendering. A prompt-library design that exposes current Kit prompts must decide whether stored/retrieved prompts are raw base prompts or fully prepared prompts with skills, profile, and subagent suffixes.
13. Prompt-producing command surfaces currently include `brainstorm`, `spec`, `plan`, `tasks`, `implement`, `reflect`, `resume`/`catchup`, `handoff`, `reconcile`, `dispatch`, `skill mine`, `summarize`, `code-review`, and the init-time constitution prompt. These are discoverable through `pkg/cli/*_prompt.go`, command files, and `pkg/cli/prompt_golden_test.go`.
14. Existing tests cover clipboard output (`pkg/cli/output_test.go`), editor behavior (`pkg/cli/editor_input_test.go`), dispatch input/prompt behavior (`pkg/cli/dispatch_test.go`), config defaults (`internal/config/config_test.go`), root help (`pkg/cli/root_help_test.go`), and prompt golden outputs (`pkg/cli/prompt_golden_test.go`). A production implementation should extend those patterns.
15. Feature notes at `docs/notes/0025-v0-prompt-library` currently contain only `.gitkeep`, so they provide no additional requirements beyond this brainstorm.

## AFFECTED FILES

1. `internal/config/config.go` — likely needs prompt registry schema fields, global config path helpers, config merge/loading helpers, and save behavior for local/global prompt edits.
2. `internal/config/config_test.go` — should cover prompt schema defaults, global path resolution, local/global merge precedence, and backward-compatible config loading.
3. `pkg/cli/root.go` — root command registration surface for any new `prompt` and `set` command wiring.
4. `pkg/cli/root_help.go` — command ordering and help grouping must include `prompt` and likely `set` without making root help noisy.
5. `pkg/cli/prompt_output.go` — likely reusable for clipboard-first prompt retrieval, but may need a variant that both copies and prints the selected prompt body in default `kit prompt` mode.
6. `pkg/cli/clipboard.go` — currently macOS-only via `pbcopy`; prompt retrieval should reuse it unless cross-platform support becomes required in scope.
7. `pkg/cli/human_output.go` — likely home for concise emoji-rich acknowledgement/origin display helpers.
8. `pkg/cli/editor_input.go` — should be reused by `kit set prompt` for editor capture; avoid duplicating editor launch behavior.
9. `pkg/cli/dispatch_input.go` — reference for interactive source precedence if `kit set prompt` eventually supports `--file` or stdin; v0 may not need those flags.
10. `internal/promptdoc/doc.go` and `pkg/cli/prompt_ir_helpers.go` — relevant if built-in Kit prompt definitions move into a registry while staying typed/structured.
11. `pkg/cli/brainstorm_prompt.go`, `pkg/cli/spec_output.go`, `pkg/cli/plan.go`, `pkg/cli/tasks.go`, `pkg/cli/implement.go`, `pkg/cli/reflect.go`, `pkg/cli/catchup_prompt.go`, `pkg/cli/handoff_prompt.go`, `pkg/cli/reconcile_prompt.go`, `pkg/cli/dispatch_prompt.go`, `pkg/cli/skill_prompt.go`, `pkg/cli/summarize.go`, `pkg/cli/code_review.go`, and `pkg/cli/init.go` — current prompt-producing surfaces that shape the scope of “house every prompt we currently use for all commands in kit.”
12. `pkg/cli/prompt_golden_test.go` and `pkg/cli/testdata/*.golden` — prompt registry changes may require updated golden coverage or new registry-specific golden tests.
13. `pkg/cli/output_test.go` — should cover default prompt-library copy-and-print behavior, `--output-only`, and `--copy` interactions if supported.
14. `pkg/cli/editor_input_test.go` — should cover `kit set prompt` reuse of editor launch and cancellation semantics if new wrapper logic is added.
15. `README.md` — command docs must describe `kit prompt`, `kit set prompt`, prompt hierarchy, and global/local storage.
16. `docs/CONSTITUTION.md` — may need a small configuration reference update if prompt-library YAML becomes part of the canonical config schema.
17. `internal/templates/templates.go` and `docs/specs/0000_INIT_PROJECT.md` — may need config reference/template updates if new prompt fields should appear in initialized projects.
18. `/Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh` — external source for initial one-off prompt content; should not be modified by Kit, but its three prompt modes are migration inputs.
19. `/Users/jamesonstone/.config/kit/.kit.yaml` — intended global prompt config path; currently absent and should be created only by explicit global save behavior.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Feature notes | notes | docs/notes/0025-v0-prompt-library | optional pre-brainstorm research input | optional |
| Constitution | doc | docs/CONSTITUTION.md | project invariants, config/state constraints, required brainstorm sections | active |
| Repo agent routing docs | doc | AGENTS.md, CLAUDE.md, .github/copilot-instructions.md, docs/agents/README.md | scoped loading and repo-local source-of-truth routing | active |
| Workflow, RLM, guardrails, tooling docs | doc | docs/agents/WORKFLOWS.md, docs/agents/RLM.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md | spec-driven classification, just-in-time prior-work pass, completion bar | active |
| kit map output | command output | `kit map 0025-v0-prompt-library` | current feature state, relationship/dependency baseline | active |
| Project progress summary | doc | docs/PROJECT_PROGRESS_SUMMARY.md | prior-feature index and candidate shortlist | active |
| Brainstorm-first workflow spec | prior feature doc | docs/specs/0004-brainstorm-first-workflow/SPEC.md | editor-default brainstorm flow and clipboard-first prompt contract | active |
| Dispatch command spec | prior feature doc | docs/specs/0008-dispatch-command/SPEC.md | editor-backed default capture and prompt-only utility precedent | active |
| Support command clipboard defaults spec | prior feature doc | docs/specs/0010-support-command-clipboard-defaults/SPEC.md | shared clipboard-first output contract | active |
| Command surface simplification spec | prior feature doc | docs/specs/0019-command-surface-simplification/SPEC.md | root help grouping and visible command-surface constraints | active |
| Typed prompt IR spec | prior feature doc | docs/specs/0022-typed-prompt-ir/SPEC.md | structured prompt-building boundary and prompt-producing command inventory | active |
| Karabiner prompt script | external script | /Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh | source content and behavior for existing one-off prompts | active |
| Project config | config | .kit.yaml | local config schema and project prompt storage target | active |
| Global Kit config | config | /Users/jamesonstone/.config/kit/.kit.yaml | intended global prompt storage target; currently absent | optional |
| Clipboard helper | code | pkg/cli/prompt_output.go, pkg/cli/clipboard.go, pkg/cli/output_test.go | copy/default-output behavior for prompt retrieval | active |
| Editor helper | code | pkg/cli/editor_input.go, pkg/cli/editor_input_test.go | editor-backed `kit set prompt` capture | active |
| Prompt IR and prompt-producing commands | code | internal/promptdoc/doc.go, pkg/cli/prompt_ir_helpers.go, pkg/cli/*_prompt.go, pkg/cli/prompt_golden_test.go | prompt registry scope and built-in prompt handling | active |
| Root help and human output helpers | code | pkg/cli/root_help.go, pkg/cli/human_output.go | command placement and readable terminal feedback | active |

## QUESTIONS

1. Resolved: prompt lookup precedence is local project `.kit.yaml` > global `~/.config/kit/.kit.yaml` > built-in Kit prompts, and prompt output should show the selected prompt's origin.
2. Resolved: `kit prompt <noun> <verb>` copies the prompt and prints the prompt body by default so users can confirm the selected prompt.
3. Resolved: v0 does not auto-paste, restore the old clipboard, or run AppleScript-style OS automation; it copies and prints only.
4. Resolved: the initial built-in toolbox prompts are `kit prompt coding-agent short`, `kit prompt coding-agent long`, and `kit prompt coding-agent instructions`.
5. Resolved: built-in Kit workflow prompts should be retrievable as rendered prepared prompts including the shared Skills/Profile/Subagent suffixes where those suffixes apply today.
6. Resolved: `kit set prompt <noun> <verb>` should require overwrite confirmation when a prompt already exists.
7. Resolved: naked `kit set prompt` should run an interactive wizard for noun, verb, save target, and editor capture.
8. Resolved: `kit set prompt <noun> <verb>` defaults to local only when a project `.kit.yaml` is found; outside a Kit project it prompts for global save instead of silently creating project config.
9. Resolved: prompt nouns and verbs are normalized to lowercase kebab-case.
10. Resolved: v0 does not support stdin or `--file` for setting prompt text, but the design should not block adding those later.
11. Resolved: YAML prompt entries use object form with required `content` and optional `description`, addressed as `prompts.<noun>.<verb>.content`.
12. Resolved: local and global user prompts may shadow built-in prompt identities directly according to precedence.
13. Resolved: v0 does not add `--source=builtin|global|local|auto`; origin display is enough.
14. Resolved: built-in Kit command prompts use taxonomy `workflow brainstorm|spec|plan|tasks|implement|reflect`, `support resume|handoff|summarize|reconcile|dispatch|code-review`, `skill mine`, and `project init`.
15. Resolved: `kit prompt` with no args shows an interactive noun selector, then verb selector.
16. Resolved: `kit prompt <noun>` shows an interactive verb selector under that noun.
17. Resolved: `kit prompt <noun> <verb>` fails fast when no prompt exists and shows nearest available nouns or verbs.
18. Resolved: `kit prompt` supports `--output-only` to print raw prompt text and skip clipboard copying.
19. Resolved: `kit prompt --output-only --copy` prints raw prompt text and copies it.
20. Resolved: v0 does not add `--no-copy`; use `--output-only` for non-copy behavior.
21. Resolved: prompt text entry and editing must use the existing vim-compatible editor path/workflow from `pkg/cli/editor_input.go`, including launch instructions, press-any-key gating, temp-file editing, and save/quit semantics.
22. Resolved: `kit set prompt` must allow overriding an existing prompt locally and/or globally, but must confirm the overwrite before saving.
23. Resolved: `kit set prompt <noun> <verb> --local --global` edits once, then saves the same prompt content to both local and global scopes.
24. Resolved: when dual-scope save overwrites existing prompts in both places, Kit confirms each overwrite separately before saving.
25. Resolved: built-in command prompts that require runtime context should delegate to the existing command prompt-generation flow where possible.
26. Resolved: when a built-in command prompt cannot resolve required context automatically, Kit should ask the user whether they have the missing information and whether they want to enter it through the existing interactive vim-compatible editor flow.
27. Resolved: v0 includes `kit prompt list` for non-interactive discovery.
28. Resolved: `kit prompt list` shows effective merged prompts by default using table output.
29. Resolved: prompt list table output includes command name, description, and shadow/overriding information.
30. Resolved: prompt descriptions are shown in selectors and `kit prompt list` when present.
31. No unresolved questions remain for the brainstorm phase.

## OPTIONS

1. Layered YAML registry with built-in fallback.
   - Store user prompts in `.kit.yaml` under a prompt map keyed by noun and verb.
   - Load local prompts when in a Kit project, load global prompts from `~/.config/kit/.kit.yaml` when present, and merge over embedded built-ins.
   - Pros: matches user thesis, keeps all state explicit, preserves filesystem-only constraint, supports local overrides.
   - Cons: requires new config merge/save helpers and careful YAML compatibility.
2. Embedded-only prompt catalog for current Kit prompts.
   - Expose existing Kit prompts through `kit prompt`, but do not support user-editable YAML in v0.
   - Pros: smaller implementation.
   - Cons: misses `kit set prompt`, global/local storage, and toolbox goals.
3. Separate prompt files under a `.kit/prompts/` directory.
   - Store one prompt per file with directory hierarchy.
   - Pros: easier editing and diffs for large prompts.
   - Cons: conflicts with the explicit `.kit.yaml` storage request and adds another persisted artifact family.
4. Treat prompt library as aliases over existing command invocations.
   - `kit prompt workflow brainstorm` would internally run existing prompt builders.
   - Pros: avoids duplicated prompt definitions.
   - Cons: insufficient for arbitrary one-off prompts and cannot cleanly represent global/local user prompts.

## RECOMMENDED STRATEGY

1. Use a layered prompt registry with precedence `local project` > `global user` > `built-in Kit`.
2. Store user prompts in YAML using object form with required `content` and optional `description`, addressed as `prompts.<noun>.<verb>.content`.
3. Add embedded built-ins for the migrated toolbox prompts and for current Kit command prompts without breaking the existing command-specific prompt builders.
4. Make `kit prompt <noun> <verb>` default to copy plus visible terminal output:
   - concise emoji acknowledgement
   - prompt identity and origin
   - prompt body printed in a clearly delimited block
5. Make `kit prompt` and partial invocations interactive by default:
   - no args: select noun, then verb
   - noun only: select verb under that noun
   - noun plus verb: resolve directly
   - if a built-in prompt needs runtime context that cannot be inferred, ask whether the user has the missing information and wants to enter it via the existing vim-compatible editor flow
   - fail with actionable guidance only when the user declines context entry or required context remains unresolved
6. Make `kit set prompt` extensible as a root command with a `prompt` subcommand in v0 only:
   - no args: wizard for noun, verb, save target, and editor capture
   - noun plus verb: editor capture with target resolution
   - `--local`: save to project `.kit.yaml`; if not in a Kit project, fail with actionable guidance
   - `--global`: save to `~/.config/kit/.kit.yaml`, creating the directory/file if needed
   - `--local --global`: edit once, save the same prompt content to both scopes
   - use the existing vim-compatible editor path/workflow for prompt text entry and editing
   - confirm before overwriting an existing local and/or global prompt
   - confirm local and global overwrites separately when both scopes are targeted
7. Add `kit prompt list` for discovery:
   - render table output
   - include command name, description, and shadow/overriding information
   - show effective merged prompts by default
   - include descriptions in selectors and list output when present
8. Keep auto-paste/clipboard-restore out of v0.
9. Keep implementation modular:
   - prompt registry/resolution logic outside Cobra command files
   - command files focused on CLI input/output
   - config helpers focused on load/save and global/local paths
10. Cover the feature with unit tests for schema compatibility, precedence, command interaction paths, output behavior, editor save/cancel behavior, dynamic context prompting, list table output, and help grouping.

## NEXT STEP

Brainstorm confidence is at least 95%. Next workflow step: run `kit spec v0-prompt-library`.
