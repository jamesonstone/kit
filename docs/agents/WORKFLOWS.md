# Workflows

## Native Planning To Repository Memory

1. Inspect the request, relevant code, and existing repository memory.
2. Use the host agent's native planning capability for research, clarification, design, and implementation planning.
3. Before code, assess whether the work contains material rationale that code and tests cannot preserve.
4. When it does, create or adopt `docs/specs/<feature>/SPEC.md` and translate the accepted native plan into it before implementation.
5. Keep material decisions and discoveries current while implementing.
6. Validate the implementation.
7. Load `docs/references/rules/constitution-curation.md` and curate the spec and broader repository memory to match what was actually built.

`kit spec [feature]` scaffolds or adopts the living spec and provides orientation. It does not replace native planning and does not ingest transcripts. The legacy V2 supervisor is compatibility-only.

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
- Bare `kit loop` and `kit loop workflow` are deprecated V2 compatibility paths. V3 work uses native planning.
- `kit dispatch` supports post-plan execution topology; it does not design the feature.
