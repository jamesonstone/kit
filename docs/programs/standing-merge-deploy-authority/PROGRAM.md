# Standing Merge And Deployment Authority Program

## Identity And Scope

- Program ID: `standing-merge-deploy-authority`
- Coordinator: `jamesonstone/kit`
- Ledger: `docs/programs/standing-merge-deploy-authority/PROGRAM.md`
- Program owner and supervisor: `jamesonstone`
- Current milestone: `M3-eliminate-exact-head-reauthorization`
- Scope: define bounded standing merge and standard-deployment authority in
  Kit, then keep LabCore and LabCore UI aligned without exact-head
  reauthorization prompts.
- Non-goals: merge or deploy either governance PR; mutate cloud, Kubernetes,
  infrastructure, IAM, networking, KMS, secrets, database schema, production
  data, repository settings, or branch protection.

## Participants

| Workstream | Repository | Local spec | Issue | Branch | Pull request | Operational reference |
| --- | --- | --- | --- | --- | --- | --- |
| `WS-kit-policy` | `jamesonstone/kit` | `docs/specs/0076-standing-merge-deploy-authority/SPEC.md` | `jamesonstone/kit#198` | `GH-198` | `jamesonstone/kit#199` | Kit rules, templates, init, reconcile, and validation |
| `WS-labcore-refresh` | `lsmc-bio/labcore` | not required; managed refresh is code-sufficient | `lsmc-bio/labcore#507` | `GH-507` | `lsmc-bio/labcore#508` | LabCore managed entrypoints and registry copies |
| `WS-kit-no-head-reauthorize` | `jamesonstone/kit` | `docs/specs/0076-standing-merge-deploy-authority/SPEC.md` | `jamesonstone/kit#200` | `GH-200` | pending | SHA-as-evidence invariant, generators, audits, and tests |
| `WS-labcore-no-head-reauthorize` | `lsmc-bio/labcore` | active program ledger plus managed guidance | `lsmc-bio/labcore#539` | `GH-539` | pending | Final Kit refresh and active Hybrid WGS coordination correction |
| `WS-labcore-ui-no-head-reauthorize` | `lsmc-bio/labcore-ui` | not required; managed refresh is code-sufficient | `lsmc-bio/labcore-ui#323` | `GH-323` | pending | Full managed instruction/rule/workflow refresh |

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

```text
WS-kit-no-head-reauthorize
  -> GATE-kit-no-head-reauthorize-verified
  -> WS-labcore-no-head-reauthorize
  -> WS-labcore-ui-no-head-reauthorize
  -> GATE-three-repository-consistency
  -> M3-pr-delivery
```

## Workstream State

| Workstream | Implementation | GitHub delivery | Deployment/runtime | Validation |
| --- | --- | --- | --- | --- |
| `WS-kit-policy` | PR-review repair complete in the commit containing this checkpoint; policy commit `7d08f7e` is superseded | ready PR #199; resolve exact live head after push; hosted checks PENDING | not applicable | full local and independent verification passed |
| `WS-labcore-refresh` | prior refresh complete at `57ad7a446ab5c19bb14a51a82e41d523654b78d4`; parity now stale after Kit D002/D004 repair | ready PR #508 requires in-place managed rule/workflow refresh after final Kit head | not applicable | prior validation passed; refreshed-head validation PENDING |
| `WS-kit-no-head-reauthorize` | in progress from `aae08e1881cdbdd617918a55b21c9fe45c26100b` | issue #200; branch `GH-200`; PR pending | not applicable | pending |
| `WS-labcore-no-head-reauthorize` | planned from `d89f5d7829d2b34e5bb750a1c21c94ab6a4c6aa0` | issue #539; branch `GH-539`; PR pending | not applicable | pending |
| `WS-labcore-ui-no-head-reauthorize` | planned from `26ff6a0c6c01ec0e61db82b7cdf2f47553e09dea` | issue #323; branch `GH-323`; PR pending | not applicable | pending |

## Milestones And Gates

| ID | State | Evidence required to advance |
| --- | --- | --- |
| `M1-kit-policy` | satisfied | canonical rule/template/audit implementation and focused plus full Kit validation |
| `GATE-kit-source-verified` | pending push | D001-D005 repair passed focused/full/race/static/build and independent verification; exact pushed head and hosted checks pending |
| `GATE-instruction-consistency` | stale | LabCore #508 matches `7d08f7e`, but D002/D004 change managed workflows/rules and require an in-place downstream refresh after final Kit head |
| `M2-pr-delivery` | in progress | two ready human-assigned PRs exist; terminal final-head hosted check states remain to observe |
| `GATE-kit-no-head-reauthorize-verified` | pending | canonical SHA-as-evidence invariant, active-language audit, generation, and full Kit validation |
| `GATE-three-repository-consistency` | pending | Kit, LabCore, and LabCore UI active guidance contains no exact-head reauthorization requirement |
| `M3-pr-delivery` | pending | three ready human-assigned PRs with exact current heads and observed hosted state |

## Ready Frontier And Blockers

- Ready frontier: `WS-kit-no-head-reauthorize`; downstream refreshes wait for
  its exact source-verified commit.
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

- Observed at: `2026-09-05T19:24:00Z`
- Supervisor: `jamesonstone`
- State changes: PR #199 and PR #508 are merged. A later LabCore run still
  requested exact-head reauthorization after PR #538 changed from `70fdaf` to
  `204f692`, proving active downstream and copied task context remains stale.
  Issues #200, #539, and #323 now own the durable three-repository correction.
- Ready frontier: implement and verify Kit issue #200, then refresh LabCore and
  LabCore UI from that exact source.
- Blockers: merge and deployment intentionally withheld. LabCore targeted
  registry reconciliation against configured Kit `main` remains PENDING until
  Kit PR #199 lands; exact candidate source commit and installed hashes are
  recorded, and no product/runtime files changed.
- Next safe action: strengthen Kit's canonical rule/generator/audit contract,
  validate and commit it, then apply downstream refreshes. Do not merge or
  deploy the governance PRs.
- Live claims still required: three implementation heads, ready PR identities,
  hosted checks, and final active-language parity.
