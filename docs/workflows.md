# Coding-Agent Workflow

Kit's workflow is capability discovery, deterministic evidence resolution,
native planning, implementation, validation, and repository-memory curation.

## Primary Sequence

```text
kit capabilities
        ↓
kit context resolve
        ↓
load selected repository evidence
        ↓
native agent plan
        ↓
living SPEC when rationale must survive
        ↓
implementation and validation
        ↓
curated outcome and kit reconcile
```

## Workflow Contracts

| Slug | Use |
| --- | --- |
| `repository-bootstrap` | Populate starter repository memory from verified project evidence after `kit init`. |
| `implementation-delivery` | Plan, implement, validate, curate, and deliver feature work. |
| `repository-maintenance` | Analyze and maintain repository health, including usage evidence. |
| `pr-feedback-repair` | Collect, verify, repair, validate, and explicitly resolve current PR feedback. |
| `cross-repository-program-coordination` | Coordinate a ledger-backed ready frontier across repositories. |
| `pull-request-merge` | Reconcile exact merge authority, identity, policy, current-head readiness, dependencies, and effects. |
| `release-orchestration` | Build and execute an authority-aware dependency graph through separate merge, deployment, and production gates. |

Workflow documents are declarative contracts. Kit never runs phases or
supervises agents.

## Living Specifications

For consequential feature work, create or adopt
`docs/specs/<feature>/SPEC.md` before source implementation. The agent keeps
purpose, context, evidence, requirements, non-goals, observable acceptance,
accepted plan, decisions, discoveries, validation, actual outcome, delivery
evidence, and repository-memory disposition current.

Historical V1 and V2 specifications remain discoverable through
`docs/PROJECT_PROGRESS_SUMMARY.md`, explicit relationships, and progressive
disclosure. Do not mechanically rewrite history.

Contained mechanical work may use a justified `not required` repository-memory
decision when code and tests preserve the durable truth.

## Reconciliation

`kit reconcile` keeps its established interface and behavior. It remains the
surface for auditing and optionally refreshing Kit-managed files, rules, and
documents. Preview write-capable managed refreshes first:

```bash
kit reconcile --include-files --dry-run --diff
```

## Dispatch And PR Repair

`kit dispatch` turns an accepted native plan into bounded execution topology.
`kit pr fix` turns current review feedback into a coding-agent repair prompt.
`kit pr orchestrate` turns bounded repository scope into a deterministic,
authority-aware release prompt. It does not enumerate PRs, merge, deploy,
mutate infrastructure, or launch an agent.
One supervisor remains accountable; low-overlap lanes are bounded, shared files
are serialized, and nontrivial work receives read-only verification.

GitHub intake and bounded waiting stay in these explicit adapters. They never
move network behavior into `kit context resolve`.

## Weekly Maintenance

The weekly health task retains its existing repository scan and safe maintenance
behavior. Once per overall run, it also reads:

```bash
kit capabilities usage --json
kit usage status --json
kit usage report --since 90d --json
```

The task reports top commands, zero-use supported commands, failures, coverage,
storage diagnostics, and whether evidence is sufficient for a future removal
decision. It does not treat absence outside the recorded coverage window as
proof of non-use.
