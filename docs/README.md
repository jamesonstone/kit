# Kit Documentation

Kit's durable documentation is organized around a coding-agent contract,
repository evidence, and a reduced human maintenance surface.

## Guides

| Guide | Purpose |
| --- | --- |
| [Overview](overview.md) | Product boundary and evidence model. |
| [Commands](commands.md) | Exact supported command groups and removed surfaces. |
| [Workflows](workflows.md) | Agent evidence, native planning, implementation, and maintenance flow. |
| [Migration to v2](migration-v2.md) | Conservative major-upgrade procedure. |
| [v2.0.0 release notes](releases/v2.0.0.md) | Breaking changes and activation sequence. |

## Project Contract

| Document | Purpose |
| --- | --- |
| [CONSTITUTION.md](CONSTITUTION.md) | Current project invariants and architecture. |
| [PROJECT_PROGRESS_SUMMARY.md](PROJECT_PROGRESS_SUMMARY.md) | Historical feature index. |
| [agents/README.md](agents/README.md) | Coding-agent routing entrypoint. |
| [references/README.md](references/README.md) | Durable rules, workflows, and reusable evidence. |

## Feature History

- `docs/specs/<feature>/SPEC.md` is the durable living feature record.
- Historical V1/V2 specs and their staged artifacts remain preserved evidence.
- New Kit behavior does not scaffold, route to, or recommend `docs/notes`.
- Existing downstream project-owned content is not automatically deleted.
