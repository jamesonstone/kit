# SPEC

## SUMMARY

Add a read-only `kit map` command that renders the canonical document graph and
current project state, and add explicit `RELATIONSHIPS` sections to
`BRAINSTORM.md` and `SPEC.md` for feature-to-feature lineage.

## PROBLEM

Kit documents have a clear internal hierarchy, but that structure is currently
split across static help text, repository docs, and individual feature files.
Users and coding agents can read the pieces, but there is no single dynamic
surface that shows how global docs, feature docs, lifecycle phases, and
cross-feature dependencies fit together in the current repository state.

## GOALS

- add `kit map` as a read-only project mapping command
- render document hierarchy, lifecycle flow, and feature-to-feature
  relationships from canonical docs and filesystem state
- support both project-wide and feature-scoped views
- keep the first pass terminal-first and avoid creating another persisted
  markdown summary artifact
- add explicit `RELATIONSHIPS` sections to `BRAINSTORM.md` and `SPEC.md`
- treat explicit relationship declarations as the canonical source for
  cross-feature edges

## NON-GOALS

- add a new persisted graph document in this phase
- add Mermaid, Graphviz, or image output in this phase
- infer and persist feature relationships automatically without explicit
  document updates
- surface relationship data in `kit status` or
  `PROJECT_PROGRESS_SUMMARY.md` in this phase
- model non-feature repo files outside the canonical Kit document set

## USERS

- maintainers who need a fast structural view of how Kit documents relate
- coding agents that need a low-token orientation surface before deeper reads
- users resuming work who need to see feature lineage and current artifact
  state together

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| core project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical document model and command contract | active |
| constitution | doc | `docs/CONSTITUTION.md` | invariant rules for explicit state and source-of-truth docs | active |
| document validator | code | `internal/document/document.go` | required section validation and section parsing | active |
| document templates | code | `internal/templates/templates.go` | scaffolded `BRAINSTORM.md` and `SPEC.md` content | active |
| feature model | code | `internal/feature/feature.go`, `internal/feature/status.go` | current feature state, phase, and file existence | active |
| root CLI help | code | `pkg/cli/root.go` | command registration and static hierarchy surface | active |

## REQUIREMENTS

- `kit map` must be a new top-level read-only command
- running `kit map` with no feature argument must render a project-wide
  terminal-friendly graphical text map to stdout
- running `kit map <feature>` must render a feature-scoped terminal-friendly
  graphical text map for the
  resolved feature and include declared incoming or outgoing feature
  relationships that touch that feature
- the project-wide map must show:
  - global canonical docs
  - each feature directory
  - which canonical documents exist for each feature
  - whether a document is required or optional
  - the feature's current phase
  - the feature's paused state
  - declared feature-to-feature relationship edges
  - which Kit command creates or updates each canonical document
- `kit map` must derive state from the filesystem and canonical markdown docs
  rather than adding new hidden state
- `BRAINSTORM.md` must gain a required `## RELATIONSHIPS` section
- `SPEC.md` must gain a required `## RELATIONSHIPS` section
- `## RELATIONSHIPS` content must be either `none` or one bullet per explicit
  relationship
- each relationship bullet must begin with a relationship type label followed
  by a canonical feature directory identifier, for example
  `builds on: 0007-catchup-command`
- relationship targets may be wrapped in inline code for readability, but
  validation and rendering must normalize them back to the canonical feature
  directory identifier
- supported relationship types in this phase must be:
  - `builds on`
  - `depends on`
  - `related to`
- `kit map` must parse relationship edges only from explicit
  `## RELATIONSHIPS` content
- `kit map` must tolerate malformed relationship lines by rendering the valid
  edges, skipping only the invalid lines, and surfacing those skips as
  warnings in the map output
- when writing to a terminal, `kit map` may color labels, doc-state markers,
  and relationship state for scanability without changing non-TTY output
- if a referenced feature identifier does not exist, `kit map` must still show
  the declared edge and label it as unresolved rather than silently dropping it
- newly generated brainstorm and spec templates must seed `## RELATIONSHIPS`
  with `none`
- existing brainstorm and spec docs in this repository must be backfilled with
  a populated `## RELATIONSHIPS` section so validation stays consistent after
  the contract change
- prompt-producing workflow commands for brainstorm and spec creation or update
  must tell agents to refresh `## RELATIONSHIPS` and record concrete feature
  lineage when applicable
- document validation must treat placeholder-only `## RELATIONSHIPS` sections
  as invalid for touched or newly generated brainstorm and spec docs

## ACCEPTANCE

- `kit map` prints a stable project-wide terminal graph that shows the current
  canonical document hierarchy and feature state without mutating repo files
- `kit map <feature>` prints a stable feature-scoped terminal graph for that feature
- features with `RELATIONSHIPS` entries show explicit edges to related features
  in the map output
- features with `RELATIONSHIPS` set to `none` show no cross-feature edges
- harmless inline-code formatting around relationship targets stays valid after
  normalization
- malformed relationship lines remain visible as map warnings without blocking
  the rest of the read-only output
- non-existent relationship targets remain visible in map output as unresolved
- new brainstorm and spec templates include `## RELATIONSHIPS` with `none`
- existing brainstorm and spec docs in the repository gain a populated
  `## RELATIONSHIPS` section
- validation requires `RELATIONSHIPS` in brainstorm and spec docs
- workflow prompts and practical docs mention the new relationship field where
  relevant
- automated tests cover relationship parsing, map rendering, unresolved edges,
  and validator changes

## EDGE-CASES

- projects with no features
- a feature that has only `BRAINSTORM.md`
- duplicate feature numbers with different slugs in the same repo
- a feature declaring the same relationship in both brainstorm and spec docs
- conflicting relationship declarations between brainstorm and spec docs
- relationships that point to deleted or misspelled feature identifiers
- a feature with no declared relationships but missing docs
- project-wide maps with paused latest features

## OPEN-QUESTIONS

- none
