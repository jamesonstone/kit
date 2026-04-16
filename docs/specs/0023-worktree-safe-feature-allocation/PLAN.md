# PLAN

## SUMMARY

Move feature-number reservation into a shared Git-common-dir allocator, harden project validation against duplicate numeric prefixes, and keep project-wide map rendering logically ordered by dependency edges instead of raw directory number alone.

## APPROACH

- record the worktree-safe numbering contract and logical-ordering rules first
- add a small shared allocator in `internal/feature/` that uses `git rev-parse --git-common-dir`
- update feature creation to reserve numbers through the shared allocator before creating directories
- add duplicate-number grouping helpers and surface them in project validation
- harden status rendering so active-row comparison uses a unique feature identity instead of numeric ID alone
- apply dependency ordering only to project-wide map views and only for `builds on` and `depends on`

## COMPONENTS

- `docs/specs/0023-worktree-safe-feature-allocation/*`
  - record the allocator, duplicate-detection, and ordering contract
- `internal/feature/feature.go`
  - use deterministic sorting for duplicate numbers and route creation through the shared allocator
- `internal/feature/allocator.go`
  - reserve numbers from the Git common dir with a lock and fallback path
- `internal/feature/map.go`
  - order project-wide feature graphs by dependency edges
- `pkg/cli/reconcile_audit.go`
  - flag duplicate numeric prefixes in project-wide audits
- `pkg/cli/status_matrix.go`
  - compare active rows using path-aware identity

## DATA

- shared allocator state
  - Git common dir path
  - last reserved numeric prefix
  - update timestamp for debugging
- shared allocator lock
  - single lock file guarding reservation updates
- duplicate number group
  - numeric prefix
  - conflicting feature directory names
- logical map order
  - dependency edges from `builds on` and `depends on`
  - fallback order from numeric prefix and directory name

## INTERFACES

- feature allocation
  - use shared reservation when Git common dir is available
  - fall back to local scan when it is not
- project validation
  - emit an error when a numeric prefix is reused by multiple feature directories
- map ordering
  - project-wide map uses dependency order where available
  - feature-scoped map keeps the current focused relationship view

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Git CLI | tool | `git rev-parse --git-common-dir` | shared clone-local allocator location | active |
| feature creation entrypoints | code | `pkg/cli/brainstorm.go`, `pkg/cli/spec.go`, `pkg/cli/scaffold.go` | shared reservation consumers | active |
| reconcile report | code | `pkg/cli/reconcile_audit.go` | duplicate-prefix enforcement | active |
| status matrix | code | `pkg/cli/status_matrix.go` | active-row correctness with legacy duplicates | active |

## RISKS

- shared allocator lock handling can become brittle if stale locks are never cleared
  - mitigate with bounded wait time and stale-lock cleanup
- Git-common-dir lookup can fail in non-Git projects
  - mitigate with explicit fallback to local scan behavior
- dependency ordering can become surprising if `related to` affects sort order
  - mitigate by ignoring `related to` for ordering
- legacy duplicate prefixes can still confuse downstream docs until they are fixed
  - mitigate by failing project validation and hardening active-row comparison immediately

## TESTING

- allocator tests for shared reservation across simulated worktrees
- allocator fallback tests for non-Git repos
- duplicate-group tests for conflicting numeric prefixes
- project validation tests that fail on duplicate numeric prefixes
- map tests that prove dependency ordering
- status tests that prove only one duplicate-ID feature is marked active
