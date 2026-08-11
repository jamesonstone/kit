# Agents Docs

## Purpose

- Route agents from native planning through implementation to curated repository memory
- Load only the guidance and repository context needed for the current decision

## Start Here

1. Use `kit capabilities <command> --json` when command safety is not already established.
2. Run `kit context resolve --workflow <slug> --json` with relevant `--feature` and `--path` hints.
3. Load the required selected evidence in order.
4. Before any repository write, obtain the explicit work-lane choice and
   record its Pull-Request Landing Plan.
5. Rerun resolution after a material scope change.

## Runtime Routing

- `WORKFLOWS.md` — native-plan lifecycle and memory routing
- `GUARDRAILS.md` — safety, completion, validation, and final-response rules
- `RLM.md` — progressive disclosure for broad context
- `TOOLING.md` — skills, execution topology, and secondary inputs
- `docs/specs/<feature>/SPEC.md` — material feature rationale when required
- `docs/references/` — durable reusable knowledge

## System Of Record

- Native agent planning owns research, clarification, design, and plan formation
- The repository owns durable rationale; chat and transcripts do not
- V3 `SPEC.md` records purpose, context, requirements, accepted plan, decisions, discoveries, validation, outcome, and repository-memory curation
- V1 and V2 artifacts remain supported legacy inputs and must not be mechanically rewritten into V3
