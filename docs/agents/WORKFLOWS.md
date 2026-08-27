# Workflows

## Agent-First Contract

1. Establish Kit command safety with `kit capabilities <command> --json` when needed.
2. Resolve the applicable local workflow with `kit context resolve --workflow <slug> --json`.
3. Load required selected rules, specs, strategies, references, and source evidence.
4. Use native planning for research, clarification, design, and the accepted plan.
5. Before any repository write, default to a new worklane without asking,
   record the Pull-Request Landing Plan, and enter its non-primary writable
   worktree. Continue an existing lane only when the user explicitly directs it.
6. Existing-PR review, CI, base-refresh, and ordered-merge work reuses every
   targeted head and pull request; it never creates a coordinator or corrective
   pull request for scope-preserving remediation.
7. Before code, create or adopt `docs/specs/<feature>/SPEC.md` when material rationale must survive.
8. Implement, validate, and keep consequential decisions and discoveries current.
9. Curate repository memory to the actual integrated outcome, then rerun context resolution if scope changed.

`kit spec [feature]` scaffolds or adopts the living V3 spec and is
write-capable, so run it only after the lane gate in the selected worktree. It
does not replace native planning, ingest transcripts, or launch an agent.

## Memory Decision

- Create or update a spec for consequential product behavior, architecture, cross-component contracts, rejected alternatives, or historical decisions future agents need.
- Do not create a spec for mechanical or code-sufficient work when code and tests communicate the complete durable truth.
- Route feature rationale to `SPEC.md`, invariants to `CONSTITUTION.md`, reusable practices to references or rules, and domain knowledge to existing canonical domain docs.
- Treat the exact generated Constitution starter as a valid bootstrap state; promote only demonstrated project-wide truth through the Constitution curation rule.

## V3 Phase Gates

- Before implementation: purpose, context, requirements including non-goals and observable acceptance, and accepted plan must be populated.
- At completion: decisions and discoveries must be resolved, validation and actual outcome recorded, repository memory assessed, and pending placeholders removed.

## Compatibility

- V1 and V2 specs remain readable and valid.
- Never mechanically rewrite a V2 spec into V3; migration requires semantic curation.
- `kit dispatch` supports post-plan execution topology; it does not design the feature.
