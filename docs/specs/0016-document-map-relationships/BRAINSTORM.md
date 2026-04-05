# BRAINSTORM

## SUMMARY

Add a read-only `kit map` view that makes the canonical document hierarchy and
cross-feature relationships visible without adding another persisted project
summary document.

## USER THESIS

Show the hierarchy of documents and how they relate, potentially as a graph or
new command, without wasting tokens by adding another document unless it is
actually useful for the coding agent.

## RELATIONSHIPS

none

## CODEBASE FINDINGS

- `kit --help` already includes a static artifact-pipeline diagram in
  `pkg/cli/root.go`, but it is conceptual only and does not reflect repo state
- `PROJECT_PROGRESS_SUMMARY.md` captures phase and summary state but does not
  expose document relationships or feature-to-feature lineage explicitly
- canonical document validation currently has no explicit cross-feature
  relationship section in `BRAINSTORM.md` or `SPEC.md`
- the current feature model already knows file existence, phase, and paused
  state, which is enough to drive a dynamic read-only map view

## AFFECTED FILES

- `docs/CONSTITUTION.md`
  - update the canonical document contract and read-only mapping semantics
- `docs/specs/0000_INIT_PROJECT.md`
  - record the `kit map` contract and new `RELATIONSHIPS` sections
- `internal/document/document.go`
  - require and validate the new `RELATIONSHIPS` section
- `internal/templates/templates.go`
  - add `RELATIONSHIPS` to brainstorm and spec templates
- `internal/feature/*`
  - parse or expose relationship metadata for mapping output
- `pkg/cli/root.go`
  - register and order the new `map` command
- `pkg/cli/map.go`
  - implement the read-only project and feature mapping output
- `README.md`
  - document the new command and relationship metadata

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| root help pipeline diagram | code | `pkg/cli/root.go` | current conceptual hierarchy surface | active |
| document validator | code | `internal/document/document.go` | required section enforcement | active |
| document templates | code | `internal/templates/templates.go` | new section scaffolding | active |
| feature listing and status | code | `internal/feature/feature.go`, `internal/feature/status.go` | phase and existence state for map output | active |
| progress summary generator | code | `internal/rollup/rollup.go` | contrast with existing derived state view | active |

## QUESTIONS

- should `kit map` remain terminal-only in the first pass or emit a machine
  format too
- should relationships be strictly explicit or partly inferred from other docs
- how much operational detail belongs in the map before it becomes a duplicate
  of `status` or `rollup`

## OPTIONS

- add another persisted markdown summary
  - pro: easy to inspect in git and agents can read it later
  - con: duplicates existing state and consumes tokens on every session
- add a read-only `kit map` command
  - pro: zero new persistent state and can render current repo reality
  - con: requires an explicit command invocation
- overload `kit status` or `kit --help`
  - pro: no new command
  - con: mixes operational status with structural explanation and bloats common
    surfaces

## RECOMMENDED STRATEGY

- add a new read-only `kit map` command as the primary surface
- keep the first version terminal-first with ASCII output rather than adding a
  new persisted markdown graph
- add explicit `RELATIONSHIPS` sections to `BRAINSTORM.md` and `SPEC.md` so
  feature-to-feature edges are canonical instead of inferred guesswork

## NEXT STEP

Create `SPEC.md`, `PLAN.md`, and `TASKS.md` for `0016-document-map-relationships`.
