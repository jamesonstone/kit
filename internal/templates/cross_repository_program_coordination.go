package templates

const crossRepositoryProgramCoordinationGate = `## Cross-Repository Program Coordination Gate

- Before implementing or resuming an accepted plan that spans multiple repositories and includes dependent deliverables, staged deployment or activation, or expected agent or session handoff, load ` + "`docs/references/rules/cross-repository-program-coordination.md`" + `.
- Designate one coordinator repository and create or adopt one canonical ` + "`docs/programs/<program>/PROGRAM.md`" + ` ledger before implementation; participant repositories remain authoritative for local specs, delivery state, runbooks, and evidence.
- Dispatch only the reconciled ready frontier, checkpoint every material transition and handoff, and reconcile recorded claims against live repositories, GitHub, runtime, and validation evidence before resume or completion.

`
