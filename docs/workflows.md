# Kit Workflows

Kit connects native agent planning to implementation and curated repository
memory. It does not replace the host agent's planning capability.

## Native Plan To Repository Memory

```text
request
  → native agent research, clarification, design, and accepted plan
  → semantic repository-memory decision
  → create or adopt SPEC.md before code when material rationale exists
  → implementation with live decisions and discoveries
  → validation
  → curate actual outcome into canonical repository memory
```

For Codex, use `/plan`. Other capable agents may expose an equivalent planning
surface. Kit does not ingest transcripts or automatically copy a
`<proposed_plan>` response. The same-thread implementation agent semantically
translates the accepted plan into repository language.

## Repository Memory Gate

Before implementation:

1. Inspect relevant code and existing repository documentation.
2. Decide whether the work contains consequential rationale that code and tests
   cannot preserve.
3. If yes, run `kit spec <feature>` and capture the accepted plan in the living
   spec before editing implementation files.
4. If no, continue without a documentation change and explain `not required`
   in the final Repository Memory report.

After implementation and validation, curate durable knowledge by scope:

| Knowledge | Canonical home |
| --- | --- |
| Feature rationale, choices, discoveries, and actual outcome | `docs/specs/<feature>/SPEC.md` |
| Project invariants | `docs/CONSTITUTION.md` |
| Reusable practices and rules | `docs/references/` or `docs/references/rules/` |
| Domain behavior and interfaces | Existing canonical domain documentation |
| Transient research input | `docs/notes/<feature>/` until promoted or discarded |

Keep material superseded decisions with rationale. Remove transient planning
chatter and details that are fully recoverable from code.

## `kit spec`

Plain `kit spec <feature>` is non-interactive. It allocates or adopts the
feature, ensures its notes scaffold, updates the project index, preserves an
existing spec, and prints concise native-planning orientation.

New specs use `workflow_version: 3` and these sections:

- `PURPOSE`
- `CONTEXT`
- `REQUIREMENTS` including non-goals and observable acceptance
- `ACCEPTED PLAN`
- `DECISIONS`
- `DISCOVERIES`
- `VALIDATION`
- `OUTCOME`
- `REPOSITORY MEMORY`

The V3 gate is phase-aware. Before implementation, purpose, context,
requirements, and accepted plan must be populated. Completion additionally
requires resolved decisions and discoveries, exact validation, actual outcome,
repository-memory assessment, and no pending TODO placeholders.

## Final Response Contract

Every implementation final response in a V3-instructed repository includes:

```text
Repository Memory
Decision: created | updated | refactored | deleted | not required
Rationale: <why this is the correct persistence decision>
Artifacts: <paths or none>
```

## Compatibility

- V1 and V2 specs remain readable and valid.
- `kit complete` preserves the document's workflow version and applies its
  version-specific completion gate.
- V2 specs are never deterministically migrated to V3.
- `kit spec <feature> --legacy-supervisor` retains the V2 supervisor during the
  compatibility period. Former supervisor-only flags imply that mode and warn.
- Bare `kit loop` and `kit loop workflow` are deprecated V2 compatibility
  paths. They warn on V2 and reject V3 with native-planning guidance.
- `kit loop review`, validation, evidence, repair, status, and delivery
  utilities remain supported.
- `kit dispatch` is post-plan execution-topology support, not feature design.

## Instruction Migration

New projects default to `instruction_scaffold_version: 3`. On full refresh,
Kit atomically upgrades V2 instruction artifacts only when every managed file
exactly matches the generated V2 templates. Customized V2 instructions remain
untouched and on V2; review `kit reconcile --include-files` or explicitly run:

```bash
kit scaffold agents --version 3 --force
```

V1 and V2 instruction scaffolds remain supported legacy inputs, with migration
advisories reported as warnings.

## Project Structure

```text
.kit.yaml
docs/
  CONSTITUTION.md
  PROJECT_PROGRESS_SUMMARY.md
  agents/
  notes/
    0001-my-feature/
  references/
    rules/
  specs/
    0001-my-feature/
      SPEC.md
```

Local generated evidence under `.kit/runs/`, `.kit/loops/`, and `.kit/state.json`
is inspectable but non-authoritative. Markdown remains the durable source.
