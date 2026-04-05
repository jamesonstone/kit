# SPEC

## SUMMARY

Simplify the top-level Kit command surface by introducing a canonical
`resume` flow, adding `status --all` as the project overview mode, and
deprecating overlapping or duplicate command entry points while keeping them
callable for compatibility.

## PROBLEM

Kit's current top-level command list mixes lifecycle commands, prompt-only
support commands, maintenance commands, and duplicate aliases at the same
level. The result is harder onboarding, denser root help, and overlapping
ways to resume work (`catchup`, `backlog --pickup`, `brainstorm --pickup`)
that require users to understand internal distinctions before they can move
forward.

## GOALS

- add `kit resume [feature]` as the canonical resume entry point
- add `kit status --all` as the canonical project overview mode
- keep default `kit status` focused on the active feature only
- keep `kit backlog` as the visible backlog-specific list and pickup surface
- preserve backward compatibility for deprecated commands and flags for at
  least one release cycle
- hide deprecated command entry points from default root help
- group root help output into clearer product-oriented sections
- update shipped docs and help text so they teach only the canonical surface

## NON-GOALS

- redesign the CLI into nested namespaces such as `kit prompt ...`
- remove deprecated entry points immediately
- change backlog classification semantics
- change the existing default `kit status --json` payload shape
- remove `PROJECT_PROGRESS_SUMMARY.md` generation in this phase
- change prompt bodies beyond the routing and wording needed for the new
  canonical command surface

## USERS

- new users discovering Kit from `kit --help` or the README
- maintainers resuming paused or in-flight work without remembering the
  distinction between catch-up, backlog pickup, and brainstorm pickup
- automation and scripts that need compatibility while the surface is cleaned
  up

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: 0003-inplace-upgrade-update
- builds on: 0006-skill-mine-command
- builds on: 0007-catchup-command
- builds on: 0014-human-readable-terminal-output
- builds on: 0015-pause-remove-commands
- builds on: 0018-backlog-command

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| root command wiring | code | `pkg/cli/root.go` | grouped help and canonical command ordering | active |
| human-readable help/output helpers | code | `pkg/cli/human_output.go` | grouped root help rendering | active |
| catchup command | code | `pkg/cli/catchup.go` | non-backlog resume prompt behavior | active |
| backlog command and helpers | code | `pkg/cli/backlog.go`, `pkg/cli/backlog_shared.go` | backlog-specific resume behavior | active |
| backlog classification | code | `internal/feature/backlog.go` | canonical backlog identification | active |
| status command | code | `pkg/cli/status.go`, `pkg/cli/status_output.go` | active-feature and all-features status rendering | active |
| README and init project spec | doc | `README.md`, `docs/specs/0000_INIT_PROJECT.md` | user-facing command documentation | active |

## REQUIREMENTS

- [SPEC-01] Add a new root command `kit resume [feature]`
- [SPEC-02] When the explicit `resume` target is a backlog item, the command
  must clear paused state, refresh generated lifecycle views, and output the
  existing brainstorm planning prompt for that backlog item
- [SPEC-03] When the explicit `resume` target is not a backlog item, the
  command must output the existing catch-up prompt behavior for that feature
- [SPEC-04] When no `resume` feature argument is provided, the selector must
  order candidates as:
  1. paused non-backlog features
  2. active in-flight non-backlog feature, when present
  3. backlog items labeled as backlog
- [SPEC-05] `resume` must remain in the CLI layer and must not introduce a new
  persisted lifecycle concept
- [SPEC-06] Add `--copy` and `--output-only` to `resume` with the shared
  clipboard-first prompt contract
- [SPEC-07] Keep `kit backlog` visible in root help and documentation
- [SPEC-08] `kit backlog --pickup` must remain supported and must point users
  toward `kit resume` as the canonical general resume flow
- [SPEC-09] `kit brainstorm --backlog` must remain supported for deferred
  capture
- [SPEC-10] `kit brainstorm --pickup` must remain callable for compatibility,
  but it must be hidden from default help and documented as deprecated in favor
  of `kit resume <feature>` or `kit backlog --pickup <feature>`
- [SPEC-11] Add `--all` to `kit status`
- [SPEC-12] Default `kit status` text output must remain focused on the active
  feature and must no longer append the all-features table
- [SPEC-13] `kit status --all` text output must render a terminal-friendly
  fixed-width all-features matrix showing feature identity, lifecycle progress
  across the standard phase columns, paused or backlog state, and task progress
  when available
- [SPEC-14] Existing `kit status --json` output must remain backward compatible
- [SPEC-15] `kit status --all --json` must use a distinct payload shape that
  includes:
  - `mode`
  - `kit_version`
  - `active_feature`
  - `backlog_count`
  - `features`
- [SPEC-16] Each feature entry in `status --all --json` must include the normal
  per-feature status plus `is_backlog` and `next_action`
- [SPEC-17] Root help must group visible commands into:
  - Setup
  - Workflow
  - Inspect & Repair
  - Prompt Utilities
  - Utilities
- [SPEC-18] Root help grouping must apply only to the root command surface;
  subcommand help may continue using the generic template
- [SPEC-19] `kit update` must remain callable but hidden from default help and
  deprecated in favor of `kit upgrade`
- [SPEC-20] `kit skills` must remain callable but hidden from default help and
  deprecated in favor of `kit skill`
- [SPEC-21] `kit catchup` must remain callable but hidden from default help and
  deprecated in favor of `kit resume`
- [SPEC-22] `kit scaffold` must remain callable as a hidden compatibility
  command and must no longer be taught as a primary workflow step
- [SPEC-23] `kit rollup` must remain callable as a hidden maintenance command
  and must no longer be taught as part of the primary user workflow
- [SPEC-24] README, root help text, and canonical workflow docs must teach the
  canonical commands and identify deprecated commands only as migration notes

## ACCEPTANCE

- `kit --help` shows grouped sections containing only visible canonical
  commands
- `kit resume <backlog-feature>` performs backlog pickup behavior
- `kit resume <feature>` for a non-backlog feature emits the catch-up prompt
- `kit resume` with no arguments shows a mixed selector in the required order
- `kit status` continues to show the active feature without the appended fleet
  table
- `kit status --all` shows a fleet view with the fixed-width lifecycle matrix
  and paused or backlog state
- `kit status --json` remains backward compatible
- `kit status --all --json` returns the new all-features payload shape
- `kit update`, `kit skills`, `kit catchup`, `kit scaffold`, and `kit rollup`
  remain callable but are hidden from default help
- `kit brainstorm --pickup` remains callable but is hidden from default help
  and described as deprecated
- README and canonical workflow docs teach `resume`, `status --all`,
  `upgrade`, and `skill` as the canonical surfaces

## EDGE-CASES

- no resumable features exist and `kit resume` is invoked without arguments
- only backlog items exist and `kit resume` is invoked without arguments
- the explicit `resume` target is complete
- the explicit `resume` target is paused but not backlog
- `status --all` is requested in a repo with no features
- `status --all --json` is requested when only backlog items exist
- deprecated commands are invoked through scripts or shell history
- hidden deprecated flags are invoked together with `--help`

## OPEN-QUESTIONS

- none
