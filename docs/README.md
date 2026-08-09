# Kit Documentation

## Product

- [Overview](overview.md) — coding-agent-first architecture and boundaries.
- [Workflows](workflows.md) — initialize, resolve, implement, and reconcile.
- [Commands](commands.md) — exact major-release CLI and exit behavior.
- [Constitution](CONSTITUTION.md) — durable project invariants.

## Major release

- [Migration guide](migrations/coding-agent-first-major.md) — schema-v1 preview,
  apply, customization, and conflict handling.
- [Kit 2.0 release notes](releases/coding-agent-first-major.md) — protected,
  added, and removed surfaces.
- [Accepted architecture](specs/0058-coding-agent-first/SPEC.md) — rationale,
  decisions, validation, and outcome.

## Agent contract

- [Agent routing](agents/README.md)
- [Agent workflow](agents/WORKFLOWS.md)
- [Tool boundaries](agents/TOOLING.md)
- [Context loading](agents/RLM.md)
- [Guardrails](agents/GUARDRAILS.md)
- [Registry catalog](../registry/catalog.yaml)
- [Versioned schemas](../schemas)
- [Rules and workflows](references/README.md)

## Project references

- [Testing](references/testing.md)
- [Worktrees and `git-wt`](references/worktrees.md)
- [Tooling](references/tooling.md)
- [External systems](references/external-systems.md)

Historical feature specifications remain under `docs/specs/` and are not
runtime compatibility documentation.
