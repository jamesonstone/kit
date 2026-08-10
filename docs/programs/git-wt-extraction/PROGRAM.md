# Git Wt Extraction Program

## Identity And Scope

- Program ID: `git-wt-extraction`
- Coordinator: `jamesonstone/kit`
- Ledger: `docs/programs/git-wt-extraction/PROGRAM.md`
- Program owner and supervisor: `jamesonstone`
- Current milestone: `M2-kit-decommission`
- Scope: move distribution and ownership of the generic `git wt` command from
  Kit to Kura without removing Kit's internal PR-repair worktree preparation.
- Non-goals: merge pull requests, deploy cloud infrastructure, mutate host
  installations, rewrite historical Kit feature records, or make Kit-managed
  rules depend on Kura.

## Participants

| Workstream | Repository | Local spec | Issue | Branch | Pull request | Operational reference |
| --- | --- | --- | --- | --- | --- | --- |
| `WS-kura-distribution` | `jamesonstone/kura` | `docs/specs/0001-interactive-script-installer/SPEC.md` | `jamesonstone/kura#1` | `GH-1` | `jamesonstone/kura#2` | Kura README and catalog |
| `WS-kit-removal` | `jamesonstone/kit` | `docs/specs/0059-git-wt-removal/SPEC.md` | `jamesonstone/kit#139` | `GH-139` | `jamesonstone/kit#140` | Kit release and command docs |

## Dependency Graph

```text
WS-kura-distribution implementation and validation
  -> WS-kit-removal implementation and review
  -> GATE-kura-merged
  -> GATE-kit-merge-ready
  -> M3-transition-complete
```

- `WS-kit-removal` implementation may proceed from the validated Kura commit.
- `WS-kit-removal` must not merge until `GATE-kura-merged` is satisfied.
- Consumers of the old Kit release remain compatible because existing
  `git-wt` installations are not removed from hosts by either pull request.

## Workstream State

| Workstream | Implementation | GitHub delivery | Deployment/runtime | Validation |
| --- | --- | --- | --- | --- |
| `WS-kura-distribution` | complete at `195184cf25143a264cd549f7ab0880ca1cb0999c` | PR `jamesonstone/kura#2` merged as `2ec9cbe058bdf6da8a3a0e1f2b9f6dd717137239` | release and host installation unobserved | four PR checks and post-merge main CI passed |
| `WS-kit-removal` | complete locally in validated commit `3d1d882f6b3c0a404f9667c3dc5217636dc75cee` | ready PR `jamesonstone/kit#140` open and mergeable; hosted checks running | not applicable before merge; release unobserved | complete local matrix passed with documented baseline lint and Go advisories |

## Milestones And Gates

| ID | State | Evidence required to advance |
| --- | --- | --- |
| `M1-kura-ready` | satisfied | Kura PR #2 at `195184cf` is ready, mergeable, and all observed hosted checks passed |
| `M2-kit-decommission` | in progress | Kit implementation, complete local validation, ready PR, and observed hosted state |
| `GATE-kura-merged` | satisfied | Kura PR #2 is `MERGED` at exact merge commit `2ec9cbe0`; post-merge main CI passed |
| `GATE-kit-merge-ready` | in progress | `GATE-kura-merged` plus a ready, mergeable Kit PR with required checks resolved |
| `M3-transition-complete` | blocked | Both PRs merged, release/runtime obligations observed or explicitly accepted, and final repository memory reconciled |

## Ready Frontier And Blockers

- Ready frontier: hosted check observation for ready Kit PR #140 and any
  in-scope repair required by those checks.
- `GATE-kura-merged` is satisfied. No cross-repository blocker remains for Kit
  PR delivery; merge remains outside this task's authority.
- No deployment or infrastructure mutation is authorized or required.

## Compatibility, Rollback, And Completion

- Compatibility: Kura installs the extracted command as `git-wt`, preserving
  Git's external-subcommand discovery. Kit's own repair commands keep their
  internal canonical-worktree preparation.
- Merge order: Kura first, Kit second.
- Rollback: before merge, close or revise the Kit PR; after merge, revert the
  Kit removal if Kura distribution is unavailable. Neither PR deletes an
  existing host installation.
- Completion remains unobserved until both merge states and any release/runtime
  obligations are reconciled live.

## Current Checkpoint

- Observed at: `2026-08-10T21:20:29Z`
- Supervisor: `jamesonstone`
- State changes: Kura PR #2 merged its reviewed head unchanged as `2ec9cbe0`
  and post-merge main CI passed; Kit removal and local validation completed;
  ready Kit PR #140 opened mergeable with implementation commit `3d1d882f`
  and checks running.
- Ready frontier: observe Kit PR #140 checks and repair any in-scope failure.
- Blockers: none for delivery. Kit merge, release, and host installation remain
  separate unperformed actions.
- Next safe actions: reconcile Kit PR #140 hosted state and update this ledger.
- Live claims still required: final Kit PR check and merge state, Kura and Kit
  release publication, and any actual host installation remain unobserved.
