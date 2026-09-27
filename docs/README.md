# Kit Documentation

Kit's durable documentation is organized around a coding-agent contract,
repository evidence, and a reduced human maintenance surface.

## Guides

| Guide | Purpose |
| --- | --- |
| [Overview](overview.md) | What Kit does and deliberately does not do. |
| [Commands](commands.md) | Exact supported command groups and removed surfaces. |
| [Migration to v3](migration-v3.md) | Capability-aware orchestration and Go module migration. |
| [v3.0.0 release notes](releases/v3.0.0.md) | Breaking changes and release boundary. |
| [Historical migration to v2](migration-v2.md) | Prior conservative major-upgrade procedure. |
| [Historical v2.0.0 release notes](releases/v2.0.0.md) | Prior breaking changes and activation sequence. |

## Project Contract

| Document | Purpose |
| --- | --- |
| [CONSTITUTION.md](CONSTITUTION.md) | Current project invariants and architecture. |
| [references/README.md](references/README.md) | Rules index and reusable project references. |

## Feature History

- `docs/specs/<feature>/SPEC.md` is the durable living feature record.
- Historical V1/V2 specs and their staged artifacts remain preserved evidence.
- New Kit behavior does not scaffold, route to, or recommend `docs/notes`.
- Existing downstream project-owned content is not automatically deleted.
