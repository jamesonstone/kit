# AGENTS

## Purpose

- This file is a routing table, not the full manual
- Start at `docs/agents/README.md` and load only the guidance needed for the current decision
- Use native agent planning for research, clarification, design, and implementation planning
- Treat repo-local markdown under `docs/` as persistent repository memory

## Repository Memory Gate

- Before implementation, inspect relevant code and existing repository memory
- Decide semantically whether the work contains material rationale that code and tests cannot preserve
- When material rationale exists, create or adopt `docs/specs/<feature>/SPEC.md` before editing implementation files and capture the accepted native plan
- When code and tests are sufficient, do not create documentation solely to satisfy a process; record `not required` in the final Repository Memory report
- During implementation, keep material decisions and discoveries current in the spec
- After implementation and validation, load `docs/references/rules/constitution-curation.md`; curate feature rationale into `SPEC.md`, demonstrated project invariants into `docs/CONSTITUTION.md`, reusable practices into `docs/references/` or `docs/references/rules/`, and domain knowledge into its existing canonical documentation
- Remove transient planning chatter and code-recoverable detail during curation; retain material superseded decisions with rationale

## Final Response Contract

- Every implementation final response must include:
  - `Repository Memory`
  - `Decision: created | updated | refactored | deleted | not required`
  - `Rationale: <why this is the correct persistence decision>`
  - `Artifacts: <paths or none>`

## Runtime Routing

- `docs/agents/README.md` — classify the work and choose the next document
- `docs/agents/WORKFLOWS.md` — native planning, implementation, and repository-memory lifecycle
- `docs/agents/GUARDRAILS.md` — completion, safety, and hard rules
- `docs/agents/RLM.md` — just-in-time context loading
- `docs/agents/TOOLING.md` — skills, post-plan dispatch, and secondary inputs

## GitHub Delivery Hard Gate

- Issue, branch, staging, commit, push, and PR actions are mutation boundaries
- Before a delivery mutation, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*` delivery rules
- Repo-local Kit rules outrank generic GitHub or plugin defaults

## AWS Context Hard Gate

- If `.kit.yaml` defines an enabled AWS context, run `kit aws verify` before the first AWS-dependent command and again immediately before AWS mutation
- Use only the verified configured profile; stop on missing credentials, incomplete configuration, or identity mismatch

## Knowledge Map

- `docs/specs/<feature>/SPEC.md` — material feature rationale and living implementation history
- `docs/CONSTITUTION.md` — project invariants
- `docs/references/` — reusable repo-wide knowledge and practices
- domain documentation — canonical domain behavior and interfaces
- `docs/notes/<feature>/` — optional source material, never canonical truth by itself

## Constraints

- Keep AGENTS short and stable
- Put durable workflow guidance in `docs/agents/*` instead of expanding always-loaded files
- Do not ingest or depend on agent transcripts as repository memory
