# SPEC

## SUMMARY

- Replace ad hoc string-built prompt construction with a typed prompt IR across Kit's prompt-producing commands, while keeping the current output wrappers and shared prompt decorators intact.

## PROBLEM

- Kit's core product output is prompts, but many commands still build them with local `strings.Builder` logic spread across multiple files.
- That makes prompt structure implicit, increases formatting drift, and makes broad prompt changes harder to reason about or test mechanically.
- Repeated prompt structures such as headings, tables, numbered steps, response contracts, and doc inventories are currently duplicated as raw strings.

## GOALS

- Introduce a typed internal prompt IR that represents prompt structure before rendering to markdown/plain text.
- Migrate all prompt-producing commands to build prompts through that IR.
- Preserve the current output flow where the rendered prompt is still passed through clipboard/output wrappers and shared decorators.
- Add exact-output golden coverage for representative migrated prompt builders, prioritizing the most structurally dense and cross-cutting prompt surfaces.
- Improve structure reuse without forcing every command into a generic template engine.

## NON-GOALS

- Adding a cached project context in this feature.
- Rewriting command execution or clipboard behavior.
- Changing prompt wording unless it improves structure, clarity, or performance while preserving semantics.
- Launching subagents directly from the Kit binary.

## USERS

- Maintainers evolving Kit's prompt surface without breaking existing command behavior.
- Coding agents that rely on prompts with stable structure and consistent formatting.
- Contributors making cross-cutting prompt updates who need stronger mechanical guarantees.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: `0021-project-validation-and-instruction-registry`
- related to: `0020-versioned-instruction-model`
- related to: `0012-default-subagent-orchestration`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| prompt output flow | code | `pkg/cli/prompt_output.go` | preserve current rendering-plus-decorator boundary | active |
| prompt-producing commands | code | `pkg/cli/implement.go`, `pkg/cli/reflect.go`, `pkg/cli/catchup_prompt.go`, `pkg/cli/handoff_prompt.go`, `pkg/cli/reconcile_prompt.go`, `pkg/cli/dispatch_prompt.go`, `pkg/cli/skill_prompt.go`, `pkg/cli/brainstorm_prompt.go`, `pkg/cli/plan.go`, `pkg/cli/tasks.go`, `pkg/cli/spec_template.go`, `pkg/cli/spec_output.go`, `pkg/cli/summarize.go`, `pkg/cli/code_review.go`, `pkg/cli/init.go` | current string-built prompt surfaces to migrate | active |
| shared prompt decorators | code | `pkg/cli/skills_prompt.go`, `pkg/cli/subagents.go` | keep skills and subagent augmentation outside the IR rendering step | active |

## REQUIREMENTS

- [SPEC-01] Kit must expose a typed internal prompt IR package for building prompt documents before rendering.
- [SPEC-02] The prompt IR must support at least:
  - headings
  - paragraphs
  - bullet lists
  - ordered lists
  - markdown tables
  - code blocks
  - raw text blocks for bounded escape hatches
- [SPEC-03] The prompt IR renderer must return a plain string suitable for the existing clipboard/output helpers.
- [SPEC-04] The prompt IR must be used by all prompt-producing commands in this feature, including:
  - `brainstorm`
  - `spec`
  - `plan`
  - `tasks`
  - `implement`
  - `reflect`
  - `catchup`
  - `handoff`
  - `reconcile`
  - `dispatch`
  - `skill mine`
  - `summarize`
  - `code-review`
  - the init-time constitution prompt
- [SPEC-05] Shared prompt decorators in `prompt_output.go` must remain the boundary after IR rendering:
  - render the base prompt first
  - then apply skills/subagent augmentation where the command currently does so
- [SPEC-06] The migration may change exact formatting or wording when it produces clearer or more concise output, but it must preserve the semantic meaning and command intent of the current prompts.
- [SPEC-07] Repeated prompt structures should move into reusable IR helpers where that improves consistency, but command-specific content must remain readable and local to the command.
- [SPEC-08] The feature must add exact-output golden tests for representative migrated prompt builders, prioritizing the most structurally dense and cross-cutting prompt surfaces.
- [SPEC-09] Golden tests must normalize unstable absolute paths or other environment-specific values so the rendered prompts stay reviewable and deterministic.
- [SPEC-10] Cached project context remains explicitly out of scope for this feature.

## ACCEPTANCE

- A typed prompt IR package exists and is used by all prompt-producing commands in scope.
- Existing prompt output wrappers and decorators continue to work without API changes to users.
- Prompt builders no longer directly depend on `strings.Builder` as their primary construction mechanism.
- Golden tests exist for representative migrated prompt builders and pass.
- Full test and build verification pass after the migration.

## EDGE-CASES

- Ordered-list items that include continuation lines or nested sub-bullets.
- Tables with long cell values or zero data rows.
- Prompts that still need a bounded raw block because the content is highly specialized.
- Prompt builders that currently mix tables, prose, and long multi-step numbered instructions in one section.

## OPEN-QUESTIONS

- none
