---
kind: ruleset
slug: constitution-curation
description: Keeps the Constitution aligned with demonstrated project-wide truth through normal post-validation repository-memory curation.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - implementation
  - validation
  - repository-memory
  - constitution
  - project-refresh
read_policy_default: must
---

# Ruleset: constitution-curation

## Purpose

- Keep `docs/CONSTITUTION.md` aligned with the project's current durable contract as implementation evolves.
- Make Constitution maintenance part of normal post-validation repository-memory curation without requiring a complete project definition during `kit init`.
- Preserve feature evolution and historical rationale in `SPEC.md` while keeping the Constitution concise, current, and project-wide.

## Applies When

- A coding agent finishes implementation and validation in a Kit-managed project.
- The implemented work establishes, changes, or disproves a project-wide principle, constraint, non-goal, definition, vocabulary term, or workflow boundary.
- A generated starter Constitution still contains only Kit-managed baseline rules and project-specific placeholders.
- `kit project refresh` reports that a broader semantic Constitution review is due.

## Rules

### Bootstrap State

- Treat the exact generated Constitution starter as a valid bootstrap state.
- Do not ask the user to explain the entire project during initialization.
- Do not populate project-specific Constitution sections from aspiration, guesses, or Kit-generated scaffolding.
- Keep initial product ideas, feature goals, and accepted native plans in the relevant `SPEC.md` until implementation demonstrates that they are durable project-wide truth.

### Post-Validation Curation

- After implementation and validation, compare the implemented outcome, current spec decisions and discoveries, affected canonical docs, and existing Constitution.
- Prefer evidence in this order:
  1. validated implemented behavior and interfaces
  2. accepted decisions and discoveries reconciled in the current `SPEC.md`
  3. recurring conventions demonstrated across completed work
  4. current canonical domain documentation
- Promote only durable project-wide principles, constraints, non-goals, definitions, vocabulary, or workflow boundaries.
- Keep feature-specific rationale, rejected alternatives, superseded decisions, and historical evolution in the relevant `SPEC.md`.
- Exclude transient planning chatter, speculative future intent, changelog entries, and details that current code and tests communicate completely.
- Preserve the Kit-managed baseline section and marker comments.
- When current implementation disproves a constitutional rule, correct or remove the stale rule and retain material historical rationale in the relevant spec.
- When no project-wide truth changed, leave the Constitution unchanged and report `Repository Memory` as `not required` with the evidence-based rationale.

### Periodic Refresh

- Treat project-refresh cadence as a trigger for reviewed semantic analysis, never as permission for automatic edits.
- When refresh is due, use `kit project refresh` to look for missed, stale, or cross-feature patterns after completing the current change's normal curation gate.
- Record a refresh with `kit project refresh --now` only after the reviewed semantic refresh is complete.

## Anti-Patterns

- Turning `kit init` into a long project-definition interview.
- Treating the original product pitch as permanent constitutional truth.
- Copying an accepted plan or completed spec wholesale into the Constitution.
- Adding a rule after one local implementation choice when it is not a project-wide invariant.
- Using the Constitution as a changelog or implementation inventory.
- Updating the Constitution merely because a timer or feature threshold elapsed.
- Leaving a known-stale constitutional rule in place because it was once accurate.

## Verification

- Confirm every Constitution change is supported by current implementation, validation, reconciled specs, recurring evidence, or canonical domain documentation.
- Confirm project-specific additions are durable and project-wide rather than feature-local.
- Confirm superseded historical rationale remains in the relevant spec when its removal from the Constitution would otherwise erase an important decision.
- Confirm the Kit-managed baseline and marker comments remain intact.
- Run `kit check --project` after changing the Constitution.
- Review `git diff -- docs/CONSTITUTION.md docs/specs docs/references` before finalizing.
- State the Constitution curation result in the final `Repository Memory` report, including `not required` when no update was warranted.

## Examples

Promote after implementation proves a durable invariant:

```text
Implemented and validated every persistence adapter behind the same transaction boundary.
The reconciled spec records that callers must not manage transactions directly.
Update CONSTITUTION.md with that project-wide boundary.
```

Keep feature-local history in the spec:

```text
The feature changed from polling to webhooks during implementation.
Record the superseded polling decision and final webhook rationale in SPEC.md.
Update CONSTITUTION.md only if webhook-driven integration is now a project-wide invariant.
```

Leave bootstrap content unchanged:

```text
The repository contains only Kit scaffolding and no implemented product behavior.
Keep the generated Constitution starter unchanged and report that no project-specific constitutional curation was required.
```
