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
| `WS-kit-policy` | `jamesonstone/kit` | `docs/specs/0076-standing-merge-deploy-authority/SPEC.md` | `jamesonstone/kit#198` | `GH-198` | pending | Kit rules, templates, init, reconcile, and validation |
| `WS-labcore-refresh` | `lsmc-bio/labcore` | not required; managed refresh is code-sufficient | `lsmc-bio/labcore#507` | `GH-507` | pending | LabCore managed entrypoints and registry copies |

## Standing Authority And Delivery Boundary

- Authorization source: the user accepted this exact two-repository delivery
  task and explicitly required governed lanes plus ready PRs.
- Authorized PR set: the two PR identities created from `GH-198` and `GH-507`;
  exact numbers and current heads are materialized at delivery checkpoints.
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
| `WS-kit-policy` | source verified in `GH-198`; exact commit pending | issue #198; branch `GH-198`; PR pending | not applicable | full local Kit matrix passed |
| `WS-labcore-refresh` | planned from base `d083a8e212033275ba8e2cab96e82539fe14b0cc` | issue #507; branch `GH-507`; PR pending | not applicable | pending |

## Milestones And Gates

| ID | State | Evidence required to advance |
| --- | --- | --- |
| `M1-kit-policy` | satisfied | canonical rule/template/audit implementation and focused plus full Kit validation |
| `GATE-kit-source-verified` | pending commit | exact Kit commit; source, generation, init/refresh, reconcile, full Go/race, vet, lint, build, project, and source-size checks already pass |
| `GATE-instruction-consistency` | pending | LabCore's three entrypoints and required managed copies agree with exact Kit source while preserving local policy |
| `M2-pr-delivery` | pending | two ready human-assigned PRs at exact heads with observed hosted state |

## Ready Frontier And Blockers

- Ready frontier: commit `WS-kit-policy`, then `WS-labcore-refresh` from that
  exact source.
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

- Observed at: `2026-09-05T02:05:00Z`
- Supervisor: `jamesonstone`
- State changes: implemented the narrowed standing-authority contract across
  canonical rules, generated entrypoints, workflows, release prompt, semantic
  audits, instruction v12, and regression tests. Full Kit source verification
  is complete; exact commit identity is pending.
- Ready frontier: commit `WS-kit-policy`, then refresh `WS-labcore-refresh`
  from that exact commit.
- Blockers: none.
- Next safe action: commit the source-verified Kit change, then refresh LabCore
  from that exact commit without touching its primary checkout or local-custom
  production safeguards.
- Live claims still required: Kit and LabCore implementation heads, ready PR
  identities, hosted checks, and final refresh/reconcile convergence.
