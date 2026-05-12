---
kit_metadata_version: 1
artifact: "spec"
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
  - name: "Constitution"
    type: "doc"
    location: "docs/CONSTITUTION.md"
    used_for: "filesystem-backed state, command constraints, required spec sections, no hidden external state"
    status: "active"
  - name: "Repository instruction entrypoints"
    type: "doc"
    location: "AGENTS.md, CLAUDE.md, .github/copilot-instructions.md, docs/agents/README.md"
    used_for: "source-of-truth routing and scoped context loading"
    status: "active"
  - name: "Workflow and guardrail docs"
    type: "doc"
    location: "docs/agents/WORKFLOWS.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md"
    used_for: "spec-driven workflow, completion bar, dependency table rules, skills discovery rules"
    status: "active"
  - name: "RLM guide"
    type: "doc/skill"
    location: "docs/agents/RLM.md"
    used_for: "just-in-time prior-work pass and selected `rlm` skill row"
    status: "active"
  - name: "References index"
    type: "doc"
    location: "docs/references/README.md"
    used_for: "confirmed no broader reference was needed beyond testing/tooling conventions"
    status: "optional"
  - name: "BRAINSTORM"
    type: "doc"
    location: "docs/specs/0025-v0-prompt-library/BRAINSTORM.md"
    used_for: "upstream research findings and resolved prompt-library decisions"
    status: "active"
  - name: "kit map output"
    type: "command output"
    location: "kit map 0025-v0-prompt-library"
    used_for: "current phase, relationships, and dependency baseline"
    status: "active"
  - name: "Project progress summary"
    type: "doc"
    location: "docs/PROJECT_PROGRESS_SUMMARY.md"
    used_for: "prior-feature shortlist and current feature status"
    status: "active"
  - name: "Brainstorm-first workflow spec"
    type: "prior feature doc"
    location: "docs/specs/0004-brainstorm-first-workflow/SPEC.md"
    used_for: "editor-default free-text flow, clipboard-first prompt contract, prompt-only flags"
    status: "active"
  - name: "Dispatch command spec"
    type: "prior feature doc"
    location: "docs/specs/0008-dispatch-command/SPEC.md"
    used_for: "interactive editor-backed capture precedent and prompt utility behavior"
    status: "active"
  - name: "Support command clipboard defaults spec"
    type: "prior feature doc"
    location: "docs/specs/0010-support-command-clipboard-defaults/SPEC.md"
    used_for: "`--output-only` and `--output-only --copy` semantics"
    status: "active"
  - name: "Command surface simplification spec"
    type: "prior feature doc"
    location: "docs/specs/0019-command-surface-simplification/SPEC.md"
    used_for: "root help grouping and prompt utility command placement"
    status: "active"
  - name: "Typed prompt IR spec"
    type: "prior feature doc"
    location: "docs/specs/0022-typed-prompt-ir/SPEC.md"
    used_for: "structured prompt generation and prompt-producing command inventory"
    status: "active"
  - name: "Karabiner prompt script"
    type: "external script"
    location: "/Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh"
    used_for: "source for initial `coding-agent short`, `coding-agent long`, and `coding-agent instructions` prompts"
    status: "active"
  - name: "Project config"
    type: "config"
    location: ".kit.yaml"
    used_for: "local prompt storage target and current config schema extension point"
    status: "active"
  - name: "Global Kit config"
    type: "config"
    location: "/Users/jamesonstone/.config/kit/.kit.yaml"
    used_for: "global prompt storage target; created or populated by `kit init`, and created when needed by explicit global save"
    status: "optional"
  - name: "Repo-local canonical skills"
    type: "skill discovery"
    location: ".agents/skills/*/SKILL.md"
    used_for: "checked for reusable skills; directory is absent in this repo state"
    status: "optional"
  - name: "Secondary global inputs"
    type: "docs/skills"
    location: "/Users/jamesonstone/.claude/CLAUDE.md, /Users/jamesonstone/.codex/AGENTS.md, /Users/jamesonstone/.codex/instructions.md, /Users/jamesonstone/.codex/skills/*/SKILL.md"
    used_for: "checked after repo-local docs; no additional selected feature skill"
    status: "optional"
  - name: "Clipboard helpers"
    type: "code"
    location: "pkg/cli/prompt_output.go, pkg/cli/clipboard.go, pkg/cli/output_test.go"
    used_for: "existing clipboard and output contract that prompt retrieval must align with"
    status: "active"
  - name: "Editor helpers"
    type: "code"
    location: "pkg/cli/editor_input.go, pkg/cli/editor_input_test.go"
    used_for: "vim-compatible prompt entry and missing-context capture workflow"
    status: "active"
  - name: "Human output helpers"
    type: "code"
    location: "pkg/cli/human_output.go, pkg/cli/root_help.go"
    used_for: "command grouping, human-readable acknowledgements, selector/list presentation"
    status: "active"
  - name: "Prompt generation surface"
    type: "code"
    location: "internal/promptdoc/doc.go, pkg/cli/prompt_ir_helpers.go, pkg/cli/*_prompt.go, pkg/cli/prompt_golden_test.go"
    used_for: "current Kit prompt-producing surfaces to expose without stale duplication"
    status: "active"
skills:
  - name: "rlm"
    source: "repo-local guide"
    path: "docs/agents/RLM.md"
    trigger: "broad/noisy feature discovery; analyze codebase; scan all files; large repository analysis; recursive language model; context routing by explicit relationships and dependencies"
    required: true
---
# SPEC

## SUMMARY

Add a layered prompt library that exposes reusable prompts through `kit prompt <noun> <verb>` and editable local/global prompt storage through `kit set prompt <noun> <verb>`. Prompt lookup must be explicit, filesystem-backed, clipboard-friendly, interactive by default, and compatible with Kit's existing prompt-generation and editor workflows.

## PROBLEM

Kit has many prompt-producing commands, and the user also maintains separate one-off prompts outside Kit. Those prompts are not discoverable through one hierarchy, are not configurable at project and user scope, and do not share Kit's existing clipboard, editor, prompt profile, skills, subagent, and help-output conventions.

## GOALS

- Add `kit prompt <noun> <verb>` as the primary retrieval surface for reusable prompts.
- Add `kit prompt list` as a non-interactive discovery surface.
- Add an extensible `kit set` command with `prompt` as the only v0 configurable resource.
- Allow prompt storage in project-local `.kit.yaml` and global `~/.config/kit/.kit.yaml`.
- Resolve prompts with precedence local project > global user > built-in Kit.
- Display prompt origin and shadow/override state so users can predict which prompt will run.
- Copy selected prompts to the clipboard by default and print the selected prompt body for confirmation.
- Preserve existing `--output-only` and `--output-only --copy` semantics for raw prompt output.
- Use the existing vim-compatible editor workflow for prompt entry and missing-context capture.
- Expose the existing Karabiner one-off prompt set as built-in toolbox prompts under `coding-agent`.
- Expose current Kit command prompts through a predictable built-in taxonomy without duplicating stale prompt text.
- Preserve Kit's filesystem-only state model and avoid hidden registries, databases, or external services.

## NON-GOALS

- Do not auto-paste prompts into other applications.
- Do not restore the previous clipboard after copying.
- Do not run AppleScript or other OS automation for prompt insertion.
- Do not add `--source=builtin|global|local|auto` in v0.
- Do not add `--no-copy` in v0.
- Do not support stdin or `--file` for `kit set prompt` in v0.
- Do not introduce a prompt database, lock file, daemon, or state outside markdown and YAML files.
- Do not replace existing Kit prompt-producing commands.
- Do not add a web UI or non-terminal prompt library interface.
- Do not store secrets or credentials in prompt entries.
- Do not support arbitrary configurable resource types under `kit set` in v0 beyond `prompt`.

## USERS

- Users who want one predictable command hierarchy for one-off prompts used during everyday agent work.
- Users who want project-specific prompt overrides in the project `.kit.yaml`.
- Users who want user-wide prompt defaults in `~/.config/kit/.kit.yaml`.
- Users who need to inspect, confirm, and copy the exact prompt selected before pasting it into an agent.
- Maintainers who need current Kit command prompts discoverable without duplicating prompt text across the codebase.
- Coding agents that need prompt-library behavior specified before implementation begins.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| rlm | repo-local guide | docs/agents/RLM.md | broad/noisy feature discovery; analyze codebase; scan all files; large repository analysis; recursive language model; context routing by explicit relationships and dependencies | yes |

## RELATIONSHIPS

- builds on: 0004-brainstorm-first-workflow
- builds on: 0010-support-command-clipboard-defaults
- builds on: 0022-typed-prompt-ir
- related to: 0008-dispatch-command
- related to: 0019-command-surface-simplification

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Constitution | doc | docs/CONSTITUTION.md | filesystem-backed state, command constraints, required spec sections, no hidden external state | active |
| Repository instruction entrypoints | doc | AGENTS.md, CLAUDE.md, .github/copilot-instructions.md, docs/agents/README.md | source-of-truth routing and scoped context loading | active |
| Workflow and guardrail docs | doc | docs/agents/WORKFLOWS.md, docs/agents/GUARDRAILS.md, docs/agents/TOOLING.md | spec-driven workflow, completion bar, dependency table rules, skills discovery rules | active |
| RLM guide | doc/skill | docs/agents/RLM.md | just-in-time prior-work pass and selected `rlm` skill row | active |
| References index | doc | docs/references/README.md | confirmed no broader reference was needed beyond testing/tooling conventions | optional |
| BRAINSTORM | doc | docs/specs/0025-v0-prompt-library/BRAINSTORM.md | upstream research findings and resolved prompt-library decisions | active |
| kit map output | command output | `kit map 0025-v0-prompt-library` | current phase, relationships, and dependency baseline | active |
| Project progress summary | doc | docs/PROJECT_PROGRESS_SUMMARY.md | prior-feature shortlist and current feature status | active |
| Brainstorm-first workflow spec | prior feature doc | docs/specs/0004-brainstorm-first-workflow/SPEC.md | editor-default free-text flow, clipboard-first prompt contract, prompt-only flags | active |
| Dispatch command spec | prior feature doc | docs/specs/0008-dispatch-command/SPEC.md | interactive editor-backed capture precedent and prompt utility behavior | active |
| Support command clipboard defaults spec | prior feature doc | docs/specs/0010-support-command-clipboard-defaults/SPEC.md | `--output-only` and `--output-only --copy` semantics | active |
| Command surface simplification spec | prior feature doc | docs/specs/0019-command-surface-simplification/SPEC.md | root help grouping and prompt utility command placement | active |
| Typed prompt IR spec | prior feature doc | docs/specs/0022-typed-prompt-ir/SPEC.md | structured prompt generation and prompt-producing command inventory | active |
| Karabiner prompt script | external script | /Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh | source for initial `coding-agent short`, `coding-agent long`, and `coding-agent instructions` prompts | active |
| Project config | config | .kit.yaml | local prompt storage target and current config schema extension point | active |
| Global Kit config | config | /Users/jamesonstone/.config/kit/.kit.yaml | global prompt storage target; created or populated by `kit init`, and created when needed by explicit global save | optional |
| Repo-local canonical skills | skill discovery | .agents/skills/*/SKILL.md | checked for reusable skills; directory is absent in this repo state | optional |
| Secondary global inputs | docs/skills | /Users/jamesonstone/.claude/CLAUDE.md, /Users/jamesonstone/.codex/AGENTS.md, /Users/jamesonstone/.codex/instructions.md, /Users/jamesonstone/.codex/skills/*/SKILL.md | checked after repo-local docs; no additional selected feature skill | optional |
| Clipboard helpers | code | pkg/cli/prompt_output.go, pkg/cli/clipboard.go, pkg/cli/output_test.go | existing clipboard and output contract that prompt retrieval must align with | active |
| Editor helpers | code | pkg/cli/editor_input.go, pkg/cli/editor_input_test.go | vim-compatible prompt entry and missing-context capture workflow | active |
| Human output helpers | code | pkg/cli/human_output.go, pkg/cli/root_help.go | command grouping, human-readable acknowledgements, selector/list presentation | active |
| Prompt generation surface | code | internal/promptdoc/doc.go, pkg/cli/prompt_ir_helpers.go, pkg/cli/*_prompt.go, pkg/cli/prompt_golden_test.go | current Kit prompt-producing surfaces to expose without stale duplication | active |

## REQUIREMENTS

- [SPEC-01] Add `kit prompt [noun] [verb]` as a visible prompt-library command.
- [SPEC-02] Add `kit prompt list` as a visible prompt discovery subcommand.
- [SPEC-03] Add `kit set` as an extensible configuration command with `prompt` as the only v0 configurable resource.
- [SPEC-04] `kit set` with no subcommand must start the prompt-setting interactive flow because `prompt` is the only v0 resource.
- [SPEC-05] `kit set prompt [noun] [verb]` must support `--local` and `--global`.
- [SPEC-06] `kit set prompt [noun] [verb] --local --global` must edit prompt text once and save the same prompt to both scopes.
- [SPEC-07] `kit set prompt` without `--local` or `--global` must default to local when a project `.kit.yaml` is found.
- [SPEC-08] `kit set prompt` without `--local` or `--global` outside a Kit project must ask whether to save globally instead of silently creating project config.
- [SPEC-09] `--local` outside a Kit project must fail with an actionable message that suggests running from a Kit project or using `--global`.
- [SPEC-10] `--global` must save to `~/.config/kit/.kit.yaml`, creating the directory and file when needed.
- [SPEC-11] Prompt storage must remain filesystem-backed in YAML.
- [SPEC-12] User prompt entries must use object form addressed as `prompts.<noun>.<verb>.content`.
- [SPEC-13] User prompt entries must require `content`.
- [SPEC-14] User prompt entries may include optional `description`.
- [SPEC-15] Prompt noun and verb values must be normalized to lowercase kebab-case.
- [SPEC-16] Invalid prompt noun or verb values must fail with actionable guidance.
- [SPEC-17] Prompt lookup precedence must be local project prompt > global user prompt > built-in prompt.
- [SPEC-18] Local and global prompts may shadow built-in prompt identities directly.
- [SPEC-19] Prompt output must show the selected prompt's origin in human-readable default mode.
- [SPEC-20] Prompt output and list output must show whether an effective prompt overrides a lower-precedence prompt.
- [SPEC-21] v0 must not add a `--source` selector for local/global/built-in lookup.
- [SPEC-22] `kit prompt` with no noun must show an interactive noun selector.
- [SPEC-23] After a noun is selected, `kit prompt` must show an interactive verb selector under that noun.
- [SPEC-24] `kit prompt <noun>` must show an interactive verb selector under the normalized noun.
- [SPEC-25] Selectors must show prompt descriptions when descriptions are available.
- [SPEC-26] `kit prompt <noun> <verb>` must resolve directly without a selector when an effective prompt exists.
- [SPEC-27] `kit prompt <noun> <verb>` must fail fast when no prompt exists.
- [SPEC-28] No-match errors must show nearest available nouns or verbs when possible.
- [SPEC-29] Default `kit prompt` output must copy the selected prompt to the clipboard.
- [SPEC-30] Default `kit prompt` output must print the selected prompt body in a clearly delimited block.
- [SPEC-31] Default `kit prompt` output must include concise human-readable feedback with emoji consistent with current Kit style.
- [SPEC-32] `kit prompt --output-only` must print only the raw prompt text and skip clipboard copying.
- [SPEC-33] `kit prompt --output-only --copy` must print raw prompt text and copy that same raw prompt text.
- [SPEC-34] v0 must not add `--no-copy`.
- [SPEC-35] v0 must not auto-paste the prompt into another application.
- [SPEC-36] v0 must not restore the previous clipboard after copying.
- [SPEC-37] v0 must not run AppleScript or similar OS automation for prompt insertion.
- [SPEC-38] `kit set prompt` must use the existing vim-compatible editor workflow for prompt text entry.
- [SPEC-39] The editor workflow must include launch instructions, press-any-key gating, temp-file editing, save-and-quit submission, and quit-without-save cancellation semantics.
- [SPEC-40] `kit set prompt` must confirm before overwriting an existing prompt in the target scope.
- [SPEC-41] When both local and global target scopes overwrite existing prompts, overwrite confirmation must happen separately per scope.
- [SPEC-42] v0 must not support stdin or `--file` for `kit set prompt`.
- [SPEC-43] Built-in toolbox prompts must include `coding-agent short`.
- [SPEC-44] Built-in toolbox prompts must include `coding-agent long`.
- [SPEC-45] Built-in toolbox prompts must include `coding-agent instructions`.
- [SPEC-46] The built-in toolbox prompts must preserve the prompt intent and content from `/Users/jamesonstone/.config/karabiner/scripts/insert_prompt.sh` while excluding that script's auto-paste and clipboard-restore behavior.
- [SPEC-47] Built-in Kit command prompts must be exposed under the taxonomy `workflow brainstorm`, `workflow spec`, `workflow plan`, `workflow tasks`, `workflow implement`, and `workflow reflect`.
- [SPEC-48] Built-in Kit support prompts must be exposed under the taxonomy `support resume`, `support handoff`, `support summarize`, `support reconcile`, `support dispatch`, and `support code-review`.
- [SPEC-49] Built-in Kit skill and project prompts must be exposed as `skill mine` and `project init`.
- [SPEC-50] Built-in Kit command prompts that require runtime context must delegate to the existing command prompt-generation flow where possible.
- [SPEC-51] Built-in Kit workflow prompts must be rendered as prepared prompts including shared Skills/Profile/Subagent suffixes where those suffixes apply today.
- [SPEC-52] Built-in prompt exposure must avoid maintaining stale duplicate prompt text when an existing command already owns prompt generation.
- [SPEC-53] When required runtime context cannot be inferred automatically, Kit must ask whether the user has that information.
- [SPEC-54] When the user has missing runtime context, Kit must offer context entry through the existing vim-compatible editor flow.
- [SPEC-55] Kit must fail with actionable guidance only when the user declines context entry or required context remains unresolved.
- [SPEC-56] `kit prompt list` must render table output.
- [SPEC-57] `kit prompt list` table output must include command name.
- [SPEC-58] `kit prompt list` table output must include description.
- [SPEC-59] `kit prompt list` table output must include shadow/overriding information.
- [SPEC-60] `kit prompt list` must show effective merged prompts by default.
- [SPEC-61] Prompt list and selector output must be deterministic and sorted for scanability.
- [SPEC-62] Root help must place `kit prompt` in the Prompt Utilities group.
- [SPEC-63] Root help must place or describe `kit set` in a way that makes its prompt-setting purpose discoverable without making the command surface noisy.
- [SPEC-64] README and relevant workflow docs must describe `kit prompt`, `kit prompt list`, `kit set prompt`, prompt precedence, YAML storage, and output semantics.
- [SPEC-65] Configuration reference docs must include the prompt library schema.
- [SPEC-66] The downstream plan must preserve RLM/discovery-first routing and record `parallelization_mode: "rlm"` in planning notes or execution metadata.
- [SPEC-67] The implementation must include tests for YAML schema compatibility, prompt precedence, local/global/global-absent loading, overwrite confirmation, editor save/cancel paths, selectors, no-match errors, output flags, list table output, dynamic context prompting, and help grouping.

## ACCEPTANCE

- Running `kit prompt coding-agent short` copies the short prompt and prints the selected prompt body with origin metadata.
- Running `kit prompt coding-agent long` copies the long prompt and prints the selected prompt body with origin metadata.
- Running `kit prompt coding-agent instructions` copies the instructions prompt and prints the selected prompt body with origin metadata.
- Running `kit prompt --output-only coding-agent short` prints only the raw short prompt and does not copy by default.
- Running `kit prompt --output-only --copy coding-agent short` prints the raw short prompt and copies it.
- Running `kit prompt` with no args shows a noun selector, then a verb selector.
- Running `kit prompt coding-agent` shows a verb selector for the `coding-agent` noun.
- Running `kit prompt missing noun` fails with nearest available prompt guidance.
- Running `kit prompt list` renders a table with command name, description, and shadow/overriding information.
- Local project prompts override global prompts and built-ins with the same noun and verb.
- Global prompts override built-ins with the same noun and verb when no local prompt exists.
- Prompt output identifies whether the selected effective prompt overrides another prompt.
- `kit set prompt custom review` in a Kit project defaults to local save and opens the existing vim-compatible editor workflow.
- `kit set prompt custom review --global` creates `~/.config/kit/.kit.yaml` when needed and saves the prompt globally.
- `kit set prompt custom review --local --global` edits once and saves identical content to both scopes.
- Overwriting an existing local prompt requires confirmation before local save.
- Overwriting existing local and global prompts requires separate confirmations.
- `kit set prompt custom review --local` outside a Kit project fails with actionable guidance.
- `kit set prompt custom review` outside a Kit project asks whether to save globally.
- Quitting the editor without saving cancels prompt creation or update.
- Built-in workflow/support prompt entries are discoverable via `kit prompt list`.
- Built-in prompt entries that can infer context render through the current command prompt-generation behavior.
- Built-in prompt entries that need missing context ask whether the user has that information and can collect it through the vim-compatible editor workflow.
- README and configuration reference docs describe prompt storage, precedence, and command usage.
- `SPEC.md` validation passes with populated required sections.
- Focused unit tests cover the prompt-library behavior listed in [SPEC-67].

## EDGE-CASES

- `~/.config/kit/.kit.yaml` does not exist before `kit init` or when a global save is requested.
- `~/.config/kit/.kit.yaml` exists but has only ordinary Kit config fields and no `prompts` section.
- Project `.kit.yaml` exists but has no `prompts` section.
- Local, global, and built-in prompts all share the same noun and verb.
- A local prompt shadows a global prompt, and the global prompt shadows a built-in prompt.
- A prompt entry exists but has empty `content`.
- A prompt entry has `description` but no `content`.
- A prompt entry has unknown future metadata fields.
- A user enters noun or verb values with uppercase letters, spaces, underscores, or punctuation.
- Normalization produces an empty noun or verb.
- Two different user inputs normalize to the same noun and verb.
- `kit prompt` is run outside a Kit project with only built-in prompts available.
- `kit prompt` is run outside a Kit project with global prompts available.
- `kit prompt list` is run when no local or global config exists.
- Clipboard copying fails.
- The user passes `--output-only --copy`.
- The user passes `--no-copy` and expects it to work.
- The user passes unsupported stdin or `--file` to `kit set prompt`.
- The editor executable cannot be found.
- The user cancels before launching the editor.
- The user quits the editor without saving.
- The user saves an empty prompt body.
- The user declines overwrite confirmation for one scope in a dual-scope save.
- The user approves overwrite for local scope but declines global scope.
- A built-in prompt requires a feature argument and no active or selectable feature exists.
- A built-in prompt requires task text or review context that cannot be inferred.
- The user declines to enter missing context.
- Missing context entered through the editor is still insufficient.
- Prompt list output contains long descriptions.
- Prompt body output is very long.
- Terminal output is non-TTY and should remain readable without color.

## OPEN-QUESTIONS

none
