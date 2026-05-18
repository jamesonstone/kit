---
kit_metadata_version: 1
artifact: "plan"
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
  - name: root help
    type: code
    target: pkg/cli/root_help.go
    relation: implements
    read_policy: conditional
    used_for: visible command grouping and removed command behavior
    status: active
  - name: README
    type: doc
    target: README.md
    relation: guides
    read_policy: conditional
    used_for: user-facing command guidance
    status: active
---
# PLAN

## SUMMARY

- Convert `scaffold` into the visible filesystem-preparation namespace and add `brainstorm --prepare`.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Add shared brainstorm prepare behavior used by both `kit brainstorm --prepare` and `kit scaffold brainstorm`.
- [PLAN-02][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13] Rewrite `pkg/cli/scaffold.go` as a namespace with workflow subcommands that create target document scaffolds without prompts.
- [PLAN-03][SPEC-14][SPEC-15][SPEC-16][SPEC-17] Move agent instruction scaffolding under `kit scaffold agents` while preserving existing flags and write-mode behavior.
- [PLAN-04][SPEC-18][SPEC-19][SPEC-20] Update command help, README, core docs, and internal scaffold-agent guidance.
- [PLAN-05][SPEC-21][SPEC-22][SPEC-23] Add focused tests for prepare mode, scaffold subcommands, root help, and command output wording.

## COMPONENTS

- `pkg/cli/brainstorm.go`
  - `--prepare` flag
  - prepare-mode validation
  - shared prepare call
- `pkg/cli/scaffold.go`
  - visible scaffold namespace
  - `brainstorm`, `spec`, `plan`, `tasks`, and `agents` subcommands
  - shared output wording
- `pkg/cli/scaffold_agents.go`
  - register under `scaffold agents`
  - preserve existing flags and behavior
- `pkg/cli/reconcile_audit.go`, `pkg/cli/reconcile_prompt.go`, `pkg/cli/project_refresh_prompt.go`
  - update guidance to `kit scaffold agents`
- `pkg/cli/root_help.go`
  - visible setup command grouping
  - no root `scaffold-agents`
- Tests under `pkg/cli`
  - prepare/scaffold file creation
  - help visibility
  - instruction scaffold command registration
- Docs
  - `README.md`
  - `docs/specs/0000_INIT_PROJECT.md`
  - `docs/CONSTITUTION.md`

## DATA

- No new persistent config fields.
- Scaffolding writes only filesystem-backed markdown and `.gitkeep` placeholders already used by Kit.
- Feature numbering uses the existing feature allocator.

## INTERFACES

- `kit brainstorm [feature] --prepare`
- `kit scaffold brainstorm <feature>`
- `kit scaffold spec <feature>`
- `kit scaffold plan <feature>`
- `kit scaffold tasks <feature>`
- `kit scaffold agents [flags]`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| brainstorm command | code | `pkg/cli/brainstorm.go, pkg/cli/brainstorm_notes.go` | prepare-mode feature and notes scaffolding | active |
| scaffold command | code | `pkg/cli/scaffold.go` | new scaffold namespace and workflow subcommands | active |
| scaffold agents command | code | `pkg/cli/scaffold_agents.go` | existing repository instruction scaffolding behavior to move under scaffold agents | active |
| root help | code | `pkg/cli/root_help.go` | visible command grouping and removed command behavior | active |
| README | doc | `README.md` | user-facing command guidance | active |

## RISKS

- Removing root `scaffold-agents` can break scripts, but the user explicitly asked to replace it with `scaffold agents`.
- Creating `BRAINSTORM.md` during prepare may look like the brainstorm workflow started; output must say prompt generation has not run yet.
- `scaffold plan` and `scaffold tasks` can create incoherent docs if prerequisites are skipped, so they should require the previous artifact.
- Root help tests may still encode the old hidden scaffold compatibility behavior.

## TESTING

- Add tests for:
  - `kit brainstorm --prepare`
  - `kit scaffold brainstorm`
  - `kit scaffold spec`
  - `kit scaffold plan` prerequisite failure
  - `kit scaffold tasks` prerequisite failure
  - `kit scaffold agents` registration and flags
  - root help visibility
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
  - `go run ./cmd/kit check --project`
  - `go run ./cmd/kit check --all`
