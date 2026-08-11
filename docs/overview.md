# Kit Overview

Kit is a coding-agent-first repository contract and evidence harness. Humans
bootstrap and maintain the repository contract; coding agents are the primary
consumers of the resulting rules, workflows, specifications, strategies,
references, and source evidence.

## Product Boundary

Kit:

- materializes repository-local rules and declarative workflow contracts;
- preserves living feature specifications and project references;
- reports command capabilities and mutation boundaries;
- resolves deterministic, ordered local evidence for an explicit workflow; and
- provides bounded prompt adapters for dispatch, pull-request repair, and
  authority-aware release orchestration.

Kit does not:

- infer project truth;
- call a model;
- launch or supervise agents;
- put network access or writes inside `kit context resolve`; or
- replace native planning, Git, GitHub, test, or delivery authority.

## Evidence Flow

1. Use `kit capabilities <command> --json` when command behavior is uncertain.
2. Run `kit context resolve --workflow <slug> --json` with relevant feature
   and path hints.
3. Load required selected artifacts in order.
4. Use native agent planning and repository evidence.
5. Create or adopt a living `SPEC.md` when material rationale must survive.
6. Implement, validate, and curate the actual integrated outcome.
7. Rerun resolution after material scope changes.

Repository-local Markdown and source remain authoritative. Resolved JSON is a
reproducible projection, not a second source of truth.

## Conservative Major Reset

Version 2 removes command families that duplicated native agent capabilities or
had no durable role. It retains the bootstrap, specification, dispatch,
rules-registry, reconciliation, health, inspection, validation, PR-repair, and
utility surfaces. Historical specifications remain valid evidence; Kit does not
mechanically rewrite them or delete downstream project-owned content.

Local bounded usage telemetry supplies future removal evidence without network
collection. The weekly health task retains its repository-maintenance behavior
and adds one overall usage analysis per run.

## Core Artifacts

| Artifact | Role |
| --- | --- |
| `.kit.yaml` | Project configuration, registry state, and optional project usage preference |
| `docs/references/rules/*.md` | Durable just-in-time agent rules |
| `docs/references/workflows/*.md` | Declarative execution contracts |
| `docs/specs/<feature>/SPEC.md` | Living feature rationale, plan, validation, and outcome |
| `docs/PROJECT_PROGRESS_SUMMARY.md` | Feature-history index |
| `docs/CONSTITUTION.md` | Demonstrated project-wide invariants |
| `docs/references/*.md` | Reusable repository evidence |
| `AGENTS.md` and provider files | Thin routes into the repository contract |
