# Agent Routing

## Entry point

Run `kit contract resolve --json` before implementation and whenever the task's
feature, paths, applicability, or workflow changes materially. Treat the
returned repository-local Markdown as the authoritative contract.

## Local guidance

- `WORKFLOWS.md` — how this repository moves from resolved contract to
  implementation and reconciliation.
- `GUARDRAILS.md` — safety, delivery, validation, and completion boundaries.
- `RLM.md` — progressive context loading.
- `TOOLING.md` — Kit command and native-tool ownership.
- `docs/specs/<feature>/SPEC.md` — mandatory living history for every
  non-trivial feature.
- `docs/CONSTITUTION.md` — durable project invariants.

## Contract states

- `ready`: read the selected workflow and ordered rules, then proceed.
- `local-custom`: valid project-owned customization; treat it as authoritative.
- `blocked`: stop implementation, follow diagnostics and next actions, and use
  `kit reconcile` to inspect required artifact state.

A blocked feature-spec state still permits authoring the reported `SPEC.md`;
re-resolve it before source implementation.

Kit resolves instructions. It does not plan, implement, test, commit, or
supervise work for the agent.

<!-- BEGIN KIT AGENT CONTRACT -->
## Kit Agent Contract

- Run `kit contract resolve --json` before implementation and whenever task scope materially changes.
- Resolve non-trivial feature work with `--workflow implementation-delivery --feature <feature>`; author the reported living V3 spec and re-resolve before source edits.
- Treat repository-local rulesets and workflows returned by the resolver as the authoritative contract.
- The resolver is local-only and read-only; use `kit registry status` for remote freshness and `kit reconcile` to preview drift.
- Installed contract: 14 ruleset(s), 4 workflow(s).
<!-- END KIT AGENT CONTRACT -->
