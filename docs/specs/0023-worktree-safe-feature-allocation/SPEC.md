# SPEC

## SUMMARY

Reserve feature numbers from a repo-shared allocator so multiple worktrees from the same clone cannot create duplicate numbered feature directories.

## PROBLEM

Feature numbers are currently allocated by scanning only the local `docs/specs/` tree, so separate worktrees created from the same commit can reserve the same next numeric prefix and produce conflicting feature directories after merge.

## GOALS

- make new feature allocation worktree-safe within the same Git clone
- keep numeric prefixes monotonic and human-readable instead of switching to hash-based identities
- detect duplicate numeric prefixes during project validation
- let `kit map` present features in dependency order without renumbering directories

## NON-GOALS

- coordinate feature numbering across separate clones or separate machines
- renumber existing feature directories automatically
- replace canonical directory names with hashes, timestamps, or UUIDs
- make `related to` relationships affect dependency ordering

## USERS

- maintainers working in multiple local worktrees from the same repository clone
- coding agents that create new features from parallel branches
- users reading project structure who need dependency order without losing chronological numbering

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: `0004-brainstorm-first-workflow`
- builds on: `0016-document-map-relationships`
- builds on: `0021-project-validation-and-instruction-registry`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| feature creation model | code | `internal/feature/feature.go` | current numbering and directory creation flow | active |
| project validation | code | `pkg/cli/check.go`, `pkg/cli/reconcile_audit.go` | duplicate prefix detection | active |
| map renderer | code | `internal/feature/map.go`, `pkg/cli/map.go` | dependency-based project ordering | active |
| Git common dir | tool | `git rev-parse --git-common-dir` | shared allocator state across worktrees in one clone | active |

## REQUIREMENTS

- new feature creation through `kit brainstorm`, `kit spec`, and other directory-creating flows must reserve the next numeric prefix from a shared allocator when the project is inside a Git repository
- the shared allocator must use the Git common dir so worktrees from the same clone see the same reservation state
- if no Git common dir is available, feature allocation must fall back to the current local scan behavior
- allocator state must be locked so concurrent reservations from separate worktrees in the same clone cannot return the same number
- reserved numbers must remain stable after allocation even if gaps later appear
- numeric prefixes must remain zero-padded and use the existing directory naming rules
- project validation must fail when duplicate numeric prefixes already exist in `docs/specs/`
- status rendering must not rely on numeric ID alone when distinguishing active rows
- `kit map` project-wide rendering must order features by dependency graph when `builds on` or `depends on` relationships provide a usable order
- logical ordering must preserve the existing directory names and fall back to deterministic numeric ordering when dependencies are absent or cyclic
- `related to` relationships must remain visible in the map but must not affect dependency order

## ACCEPTANCE

- two worktrees from the same clone can create new features sequentially without reserving the same numeric prefix
- the allocator stores its shared state outside the checked-out branch content
- feature directories remain human-readable numeric directories such as `0023-example-feature`
- `kit check --project` reports duplicate numeric prefixes as an error
- `kit map` project-wide output places prerequisite features before dependents when `builds on` or `depends on` relationships connect them
- status output still highlights only the true active feature even if duplicate numeric prefixes exist in a legacy repo
- automated tests cover shared allocator reservation, duplicate detection, and dependency ordering

## EDGE-CASES

- the repository is not a Git repo
- the shared allocator state file does not exist yet
- a stale allocator lock is left behind by a crashed process
- duplicate numeric prefixes already exist before the allocator ships
- relationship cycles prevent a full topological ordering
- a project has only `related to` edges and no ordering edges

## OPEN-QUESTIONS

- none
