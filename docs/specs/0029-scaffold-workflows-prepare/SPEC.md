---
kit_metadata_version: 1
artifact: "spec"
feature:
  id: "0029"
  slug: "scaffold-workflows-prepare"
  dir: "0029-scaffold-workflows-prepare"
relationships:
  - type: "builds_on"
    target: "0013-scaffold-agents-safe-merge"
  - type: "builds_on"
    target: "0019-command-surface-simplification"
  - type: "related_to"
    target: "0004-brainstorm-first-workflow"
references:
  - name: brainstorm command
    type: code
    target: pkg/cli/brainstorm.go, pkg/cli/brainstorm_notes.go
    relation: implements
    read_policy: conditional
    used_for: prepare-mode feature and notes scaffolding
    status: active
  - name: scaffold command
    type: code
    target: pkg/cli/scaffold.go
    relation: implements
    read_policy: conditional
    used_for: new scaffold namespace and workflow subcommands
    status: active
  - name: scaffold agents command
    type: code
    target: pkg/cli/scaffold_agents.go
    relation: implements
    read_policy: conditional
    used_for: existing repository instruction scaffolding behavior to move under scaffold agents
    status: active
  - name: artifact templates
    type: code
    target: internal/templates/templates.go
    relation: implements
    read_policy: conditional
    used_for: empty workflow document scaffolds
    status: active
  - name: root help
    type: code
    target: pkg/cli/root_help.go
    relation: implements
    read_policy: conditional
    used_for: visible command grouping and removed command behavior
    status: active
  - name: init project spec
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: implements
    read_policy: conditional
    used_for: canonical command behavior
    status: active
---
# SPEC

## SUMMARY

- Add `kit brainstorm --prepare` to create the brainstorm-phase directory/file scaffolding before emitting the brainstorm prompt.
- Redefine visible `kit scaffold` as a namespace for creating workflow document structures and subdirectories, with `kit scaffold agents` replacing the old visible `kit scaffold-agents` command.

## PROBLEM

- Users sometimes need to copy notes, documents, screenshots, examples, or other source material into `docs/notes/<feature>/` before starting the actual `kit brainstorm` workflow.
- Today `kit brainstorm` creates those directories as part of the same flow that emits the planning prompt, which is too late for preloading materials.
- The existing hidden `kit scaffold <feature>` means "create the full pipeline skeleton", while the desired meaning is narrower: create the empty document structure and supporting directories for one workflow.
- `kit scaffold-agents` is a separate root command even though it is conceptually scaffolding.

## GOALS

- Add `kit brainstorm <feature> --prepare`.
- Add visible `kit scaffold` as a root namespace.
- Add `kit scaffold brainstorm <feature>` as an alias-equivalent to `kit brainstorm <feature> --prepare`.
- Add `kit scaffold spec <feature>`, `kit scaffold plan <feature>`, and `kit scaffold tasks <feature>` to create the target phase's empty document scaffolding without emitting the workflow prompt.
- Add `kit scaffold agents` as the canonical agent-instruction scaffold command.
- Remove `kit scaffold-agents` from the public command surface in favor of `kit scaffold agents`.
- Change scaffold completion output to say:
  - `♻️ <doc_type/workflow> directory and files empty scaffolding created. Please prepare your notes, documents, images, and examples for the <doc_type/workflow> phase`
- Keep workflow commands such as `kit brainstorm`, `kit spec`, `kit plan`, and `kit tasks` responsible for prompt generation and phase-specific agent instructions.
- Preserve existing safe write behavior for agent instruction scaffolding under the new namespace.

## NON-GOALS

- Adding scaffold subcommands for implementation or reflection.
- Generating agent prompts from `kit scaffold`.
- Creating a generic arbitrary file-tree scaffolder.
- Automatically importing or parsing user-provided notes.
- Changing feature-number allocation semantics.
- Changing the content of document templates beyond what existing workflow commands already create.

## USERS

- Users who need to preload notes, screenshots, examples, or references before a brainstorm prompt starts.
- Maintainers who want `scaffold` to mean filesystem/document structure preparation.
- Users who expect agent instruction scaffolding to live under the scaffold namespace.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: 0013-scaffold-agents-safe-merge
- builds on: 0019-command-surface-simplification
- related to: 0004-brainstorm-first-workflow

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| brainstorm command | code | `pkg/cli/brainstorm.go, pkg/cli/brainstorm_notes.go` | prepare-mode feature and notes scaffolding | active |
| scaffold command | code | `pkg/cli/scaffold.go` | new scaffold namespace and workflow subcommands | active |
| scaffold agents command | code | `pkg/cli/scaffold_agents.go` | existing repository instruction scaffolding behavior to move under scaffold agents | active |
| artifact templates | code | `internal/templates/templates.go` | empty workflow document scaffolds | active |
| root help | code | `pkg/cli/root_help.go` | visible command grouping and removed command behavior | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical command behavior | active |

## REQUIREMENTS

- `kit brainstorm --prepare <feature>` must create or reuse the feature directory.
- `kit brainstorm --prepare <feature>` must create `docs/notes/<feature>/.gitkeep`.
- `kit brainstorm --prepare <feature>` must create `BRAINSTORM.md` from the normal brainstorm artifact template when missing.
- `kit brainstorm --prepare <feature>` must not ask for a brainstorm thesis.
- `kit brainstorm --prepare <feature>` must not emit or copy a brainstorm prompt.
- `kit brainstorm --prepare <feature>` must reject prompt-output flags and interactive thesis flags that only make sense for the full brainstorm flow.
- When the frontend profile is active, brainstorm prepare must also create the design materials directories and placeholders.
- `kit scaffold brainstorm <feature>` must perform the same filesystem changes as `kit brainstorm --prepare <feature>`.
- `kit scaffold spec <feature>` must create or reuse the feature directory and create `SPEC.md` from the normal spec template when missing.
- `kit scaffold plan <feature>` must require `SPEC.md` and create `PLAN.md` from the normal plan template when missing.
- `kit scaffold tasks <feature>` must require `PLAN.md` and create `TASKS.md` from the normal tasks template when missing.
- Scaffold workflow subcommands must update `PROJECT_PROGRESS_SUMMARY.md` when scaffolding succeeds.
- Scaffold workflow subcommands must not emit phase prompts.
- Scaffold workflow subcommands must be idempotent: existing files are reported and preserved.
- `kit scaffold agents` must expose the existing agent instruction flags and behavior from `kit scaffold-agents`.
- The root `kit scaffold-agents` command must no longer be registered as a visible canonical command.
- Any remaining internal guidance should point to `kit scaffold agents`, not `kit scaffold-agents`.
- The old hidden full-pipeline `kit scaffold <feature>` behavior must be removed from the visible command model.
- Root help must show `scaffold` in setup commands and omit `scaffold-agents`.
- README and the core project spec must document the new `scaffold` namespace and `brainstorm --prepare`.

## ACCEPTANCE

- `kit brainstorm sample --prepare` creates the feature directory, `BRAINSTORM.md`, and `docs/notes/<feature>/.gitkeep`, then exits without prompt output.
- `kit scaffold brainstorm sample` performs the same preparation path.
- `kit scaffold spec sample` creates `SPEC.md` without printing a spec prompt.
- `kit scaffold plan sample` creates `PLAN.md` when `SPEC.md` exists and fails clearly when it does not.
- `kit scaffold tasks sample` creates `TASKS.md` when `PLAN.md` exists and fails clearly when it does not.
- `kit scaffold agents --append-only` uses the existing append-only instruction-file behavior.
- `kit --help` shows `scaffold` and does not show `scaffold-agents`.
- `kit scaffold --help` shows `agents`, `brainstorm`, `spec`, `plan`, and `tasks`.
- Command output includes the requested `♻️ ... empty scaffolding created...` wording.
- Tests cover prepare mode, scaffold workflow subcommands, root help, and agent scaffolding under the new namespace.

## EDGE-CASES

- The feature directory already exists with copied notes but no `BRAINSTORM.md`.
- `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, or `TASKS.md` already exists.
- Frontend profile is active before brainstorm prepare.
- `kit scaffold plan` is run before `SPEC.md`.
- `kit scaffold tasks` is run before `PLAN.md`.
- Existing automation still invokes root `kit scaffold-agents`.

## OPEN-QUESTIONS

- none
