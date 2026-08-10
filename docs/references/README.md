# Contract References

Repository-local contract artifacts are Markdown with typed YAML front matter.
The canonical catalog at [`registry/catalog.yaml`](../../registry/catalog.yaml)
defines their visibility, source and target paths, versions, digests,
applicability, dependencies, and read policies.

## Rulesets

Rulesets under `docs/references/rules/` state constraints, applicability,
anti-patterns, and verification. `kit rules list|view|add` is the filtered human
interface; `kit contract resolve` selects applicable installed rules for coding
agents.

## Workflows

Workflows under `docs/references/workflows/` declare phases, gates,
dependencies, validation, and completion expectations. They never execute or
supervise an agent.

- [Implementation delivery](workflows/implementation-delivery.md)
- [PR feedback repair](workflows/pr-feedback-repair.md)
- [Registry maintenance](workflows/registry-maintenance.md)

## Other references

Project-maintainer references such as testing, tooling, external-system, and
worktree guidance are not automatically registry artifacts unless listed in
the catalog.
