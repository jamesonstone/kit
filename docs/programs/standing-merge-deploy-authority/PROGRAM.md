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
| `WS-kit-no-head-reauthorize` | `jamesonstone/kit` | `docs/specs/0076-standing-merge-deploy-authority/SPEC.md` | `jamesonstone/kit#200` | `GH-200` | `jamesonstone/kit#201` | SHA-as-evidence invariant, generators, audits, and tests |
| `WS-labcore-no-head-reauthorize` | `lsmc-bio/labcore` | active program ledger plus managed guidance | `lsmc-bio/labcore#539` | `GH-539` | `lsmc-bio/labcore#541` | Final Kit refresh and active Hybrid WGS coordination correction |
| `WS-labcore-ui-no-head-reauthorize` | `lsmc-bio/labcore-ui` | not required; managed refresh is code-sufficient | `lsmc-bio/labcore-ui#323` | `GH-323` | `lsmc-bio/labcore-ui#326` | Full managed instruction/rule/workflow refresh |

## Standing Authority And Delivery Boundary

- Authorization source: the user accepted this exact three-repository delivery
  task and explicitly required the exact-head reauthorization correction in
  Kit, LabCore, and LabCore UI through governed lanes and ready PRs.
- Authorized delivery set: Kit PR #201, LabCore PR #541, and LabCore UI PR
  #326. Merge and deployment authority for these governance PRs are explicitly
  withheld. Canonical policy source commit is
  `56b8802fcf810b3113353196868015f389840a85`; downstream heads are
  `819186d59c2a48ca846a293b5ec5424e7846127f` and
  `c3ed1c7cad0e2afd75ea6f0cceb6a56353b21201`.
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
| `WS-kit-no-head-reauthorize` | complete at policy source `56b8802fcf810b3113353196868015f389840a85`; final checkpoint follows | ready PR #201 | not applicable | full local validation passed; hosted checks PENDING |
| `WS-labcore-no-head-reauthorize` | complete at `819186d59c2a48ca846a293b5ec5424e7846127f` from base `d89f5d7829d2b34e5bb750a1c21c94ab6a4c6aa0` | ready PR #541 | not applicable | full Go, vet, lint, docs, parity, and secret-diff validation passed; hosted checks PENDING |
| `WS-labcore-ui-no-head-reauthorize` | complete at `c3ed1c7cad0e2afd75ea6f0cceb6a56353b21201` from base `26ff6a0c6c01ec0e61db82b7cdf2f47553e09dea` | ready PR #326 | not applicable | 826 tests, lint, typecheck, changed-file format, file-size, web build, parity, and secret-diff validation passed; hosted checks PENDING |

## Milestones And Gates

| ID | State | Evidence required to advance |
| --- | --- | --- |
| `M1-kit-policy` | satisfied | canonical rule/template/audit implementation and focused plus full Kit validation |
| `GATE-kit-source-verified` | satisfied | Kit PR #199 merged as `aae08e1` after D001-D005 validation and review repair |
| `GATE-instruction-consistency` | satisfied | superseded M2 drift is resolved by the M3 exact-source refreshes |
| `M2-pr-delivery` | satisfied | Kit PR #199 and LabCore PR #508 merged before the M3 correction |
| `GATE-kit-no-head-reauthorize-verified` | satisfied | canonical SHA-as-evidence invariant, active-language audit, generation, and full Kit validation passed |
| `GATE-three-repository-consistency` | satisfied | active guidance contains no exact-head reauthorization requirement; managed copies and generated entrypoints converge on Kit source `56b8802` |
| `M3-pr-delivery` | in progress | three ready human-assigned PRs exist; exact current hosted states remain to observe |

## Ready Frontier And Blockers

- Ready frontier: hosted validation and in-place review repair for PRs #201,
  #541, and #326 only.
- Blockers: none.
- Explicit hold: do not merge or deploy any of the three M3 governance PRs.

## Compatibility, Recovery, And Completion

- Compatibility: existing grants remain bounded by their recorded scope. The
  new precedence model does not turn generic tasks, ledgers, checks, or PR
  creation into merge/deployment authority.
- Recovery: before delivery, repair only within the owning issue lane. After
  delivery, keep review repair on the same PR head and rerun current evidence.
- Completion: source verification, downstream refresh, three ready PRs, and a
  final reconciled checkpoint. Merge, runtime, deployment, activation, and
  production acceptance remain `NOT_APPLICABLE` to this delivery task.

## Current Checkpoint

- Observed at: `2026-09-05T19:51:45Z`
- Supervisor: `jamesonstone`
- State changes: PR #199 and PR #508 are merged. A later LabCore run still
  requested exact-head reauthorization after PR #538 changed from `70fdaf` to
  `204f692`, proving active downstream and copied task context remains stale.
  Issues #200, #539, and #323 now own the durable three-repository correction.
- Ready frontier: observe hosted validation and repair only in place on the
  three existing PR heads when needed.
- Blockers: none for delivery. Merge and deployment are intentionally outside
  this task, and no product/runtime files changed.
- Next safe action: review PRs #201, #541, and #326. Do not merge or deploy the
  governance PRs without a separate applicable grant.
- Live claims still required: terminal hosted checks and review state.
