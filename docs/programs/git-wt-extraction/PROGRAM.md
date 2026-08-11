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
| `WS-kit-removal` | `jamesonstone/kit` | `docs/specs/0060-git-wt-removal/SPEC.md` | `jamesonstone/kit#139` | `GH-139` | `jamesonstone/kit#140` | Kit release and command docs |

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
| `WS-kit-removal` | v2 reconciliation complete at ordinary merge commit `34bdc6b65e8ce5c84289089c9e4ddce1bf70723e` from PR head `8df339a08a3541ed76523df42473876bb3d1fa8f` and main `0eb617e80839d3bc9ae326dbe3c63ddc5d0b0591` | ready PR `jamesonstone/kit#140` remains at old remote head `8df339a`; push and fresh hosted checks pending | not applicable before merge; release unobserved | integrated full Go/race/vet, builds, release snapshot, lint, improve, context, project, source-size, vulnerability, and secret checks passed locally; exact-commit improve and release provenance recorded |

## Milestones And Gates

| ID | State | Evidence required to advance |
| --- | --- | --- |
| `M1-kura-ready` | satisfied | Kura PR #2 at `195184cf` is ready, mergeable, and all observed hosted checks passed |
| `M2-kit-decommission` | in progress | Kit implementation reconciled with v2, complete local validation, ready PR, and observed hosted state |
| `GATE-kura-merged` | satisfied | Kura PR #2 is `MERGED` at exact merge commit `2ec9cbe0`; post-merge main CI passed |
| `GATE-kit-merge-ready` | pending revalidation | Updated Kit PR #140 is mergeable and clean with validate and CodeRabbit passing |
| `M3-transition-complete` | blocked | Both PRs merged, release/runtime obligations observed or explicitly accepted, and final repository memory reconciled |

## Ready Frontier And Blockers

- Ready frontier: commit the validated ordinary merge, push `GH-139`, update
  existing Kit PR #140, and observe fresh hosted validation.
- `GATE-kura-merged` is satisfied. No cross-repository blocker remains for Kit
  PR delivery; Kit merge readiness is pending the current reconciliation.
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

- Observed at: `2026-08-11T12:17:01Z`
- Supervisor: `jamesonstone`
- State changes: ordinary merge commit `34bdc6b6` now contains the six resolved
  conflicts, stable feature ID `0060`, the unblocked v2 context contract, and
  the clean integrated implementation. Exact-commit improve runs and the clean
  GoReleaser snapshot passed.
- Ready frontier: commit this evidence checkpoint, push the existing `GH-139`
  lane, update PR #140, and observe fresh hosted state.
- Blockers: none for local delivery. Kit merge readiness is pending fresh
  hosted evidence; merge, releases, and host installation remain separate
  unperformed actions.
- Next safe actions: commit this checkpoint, push without force, refresh PR
  #140's description, then reconcile hosted checks and review.
- Live claims still required: Kit merge state, Kura and Kit release publication,
  and any actual host installation remain unobserved.
