# Standing Merge And Deployment Authority Program

## Identity And Scope

- Program ID: `standing-merge-deploy-authority`
- Coordinator: `jamesonstone/kit`
- Ledger: `docs/programs/standing-merge-deploy-authority/PROGRAM.md`
- Program owner and supervisor: `jamesonstone`
- Current milestone: `M1-kit-policy`
- Scope: define bounded standing merge and standard-deployment authority in
  Kit, then refresh the resulting managed instruction contract into LabCore.
- Non-goals: merge or deploy either governance PR; mutate cloud, Kubernetes,
  infrastructure, IAM, networking, KMS, secrets, database schema, production
  data, repository settings, or branch protection.

## Participants

| Workstream | Repository | Local spec | Issue | Branch | Pull request | Operational reference |
| --- | --- | --- | --- | --- | --- | --- |
| `WS-kit-policy` | `jamesonstone/kit` | `docs/specs/0076-standing-merge-deploy-authority/SPEC.md` | `jamesonstone/kit#198` | `GH-198` | `jamesonstone/kit#199` | Kit rules, templates, init, reconcile, and validation |
| `WS-labcore-refresh` | `lsmc-bio/labcore` | not required; managed refresh is code-sufficient | `lsmc-bio/labcore#507` | `GH-507` | `lsmc-bio/labcore#508` | LabCore managed entrypoints and registry copies |

## Standing Authority And Delivery Boundary

- Authorization source: the user accepted this exact two-repository delivery
  task and explicitly required governed lanes plus ready PRs.
- Authorized delivery set: Kit PR #199 and LabCore PR #508. Merge authority is
  explicitly withheld. Policy source commit is
  `7d08f7e2e5721d5e60e2f2b05d5aaf1c50e20ba3`; current LabCore head is
  `57ad7a446ab5c19bb14a51a82e41d523654b78d4`.
- Merge authority for these governance PRs: explicitly withheld by the user.
- Deployment authority for these governance PRs: explicitly withheld by the
  user.
- Authenticated GitHub actor: `jamesonstone` in both repositories.
- Expected bases: `main` in both repositories.
- Merge method and repository policy: unobserved because merge is out of scope.
- In-place remediation: authorized inside each issue scope before ready-PR
  delivery; ordinary commits only, with fresh validation after every head
  change.
- Replacement PR criteria: material scope change, unsafe or inaccessible head,
  repository policy, or explicit user direction.
- Corrective and rollback owner: `jamesonstone`; before merge, update or close
  the relevant PR. Post-merge rollback is outside this task.

## Dependency Graph

```text
WS-kit-policy implementation and validation
  -> GATE-kit-source-verified
  -> WS-labcore-refresh from exact Kit source
  -> GATE-instruction-consistency
  -> M2-pr-delivery
```

- `WS-labcore-refresh` may preview current drift, but write-capable refresh must
  use the exact source-verified Kit commit from `WS-kit-policy`.
- No merge or deployment node belongs to this program's ready frontier.

## Workstream State

| Workstream | Implementation | GitHub delivery | Deployment/runtime | Validation |
| --- | --- | --- | --- | --- |
| `WS-kit-policy` | PR-review repair complete in the commit containing this checkpoint; policy commit `7d08f7e` is superseded | ready PR #199; resolve exact live head after push; hosted checks PENDING | not applicable | full local and independent verification passed |
| `WS-labcore-refresh` | prior refresh complete at `57ad7a446ab5c19bb14a51a82e41d523654b78d4`; parity now stale after Kit D002/D004 repair | ready PR #508 requires in-place managed rule/workflow refresh after final Kit head | not applicable | prior validation passed; refreshed-head validation PENDING |

## Milestones And Gates

| ID | State | Evidence required to advance |
| --- | --- | --- |
| `M1-kit-policy` | satisfied | canonical rule/template/audit implementation and focused plus full Kit validation |
| `GATE-kit-source-verified` | pending push | D001-D005 repair passed focused/full/race/static/build and independent verification; exact pushed head and hosted checks pending |
| `GATE-instruction-consistency` | stale | LabCore #508 matches `7d08f7e`, but D002/D004 change managed workflows/rules and require an in-place downstream refresh after final Kit head |
| `M2-pr-delivery` | in progress | two ready human-assigned PRs exist; terminal final-head hosted check states remain to observe |

## Ready Frontier And Blockers

- Ready frontier: commit and push Kit D001-D005 repair, verify hosted state and
  resolve addressed threads, then refresh LabCore PR #508 in its owning lane.
- Blockers: none.
- Explicit hold: do not merge or deploy either governance PR.

## Compatibility, Recovery, And Completion

- Compatibility: existing grants remain bounded by their recorded scope. The
  new precedence model does not turn generic tasks, ledgers, checks, or PR
  creation into merge/deployment authority.
- Recovery: before delivery, repair only within the owning issue lane. After
  delivery, keep review repair on the same PR head and rerun current evidence.
- Completion: source verification, downstream refresh, two ready PRs, and a
  final reconciled checkpoint. Merge, runtime, deployment, activation, and
  production acceptance remain `NOT_APPLICABLE` to this delivery task.

## Current Checkpoint

- Observed at: `2026-09-05T11:15:49Z`
- Supervisor: `jamesonstone`
- State changes: repaired all five active CodeRabbit findings on Kit PR #199.
  Fresh independent verification is clean. D002/D004 supersede managed content
  previously propagated to LabCore head `57ad7a44`, returning the downstream
  parity gate to stale until an in-place refresh.
- Ready frontier: push the verified Kit repair, complete reflection and thread
  resolution, then hand off LabCore #508 refresh without merging either PR.
- Blockers: merge and deployment intentionally withheld. LabCore targeted
  registry reconciliation against configured Kit `main` remains PENDING until
  Kit PR #199 lands; exact candidate source commit and installed hashes are
  recorded, and no product/runtime files changed.
- Next safe action: complete Kit PR #199 delivery and review resolution. Then
  refresh LabCore PR #508 from the exact final Kit source in its existing lane.
  Do not merge or deploy.
- Live claims still required: exact pushed Kit head, terminal hosted checks,
  resolved thread count, and refreshed LabCore parity.
