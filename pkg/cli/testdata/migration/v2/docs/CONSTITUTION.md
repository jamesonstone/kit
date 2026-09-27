# CONSTITUTION

## PRINCIPLES

<!-- TODO: define core principles that guide all decisions -->

## CONSTRAINTS

<!-- TODO: define invariant rules that must never be violated -->

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all `docs/**`, all `.kit/**`, or `.kit.yaml`.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->

## CHANGE CLASSIFICATION

<!-- all work falls into one of two tracks — classify before acting -->

### Spec-Driven (Formal)

<!-- use when: new features, kit brainstorm/kit spec, substantial architectural or behavioral changes -->
<!-- workflow: optional BRAINSTORM.md → SPEC.md → PLAN.md → TASKS.md → implement → reflect -->

### Ad Hoc (Lightweight)

<!-- use when: bug fixes, security reviews, refactors, dependency updates, config changes, small refinements -->
<!-- workflow: understand → implement → verify -->
<!-- docs: update only practical docs (READMEs, inline docs, API docs) -->
<!-- do NOT create SPEC.md / PLAN.md / TASKS.md for ad hoc work -->

### Ad Hoc with Existing Specs

<!-- if change touches code with existing spec docs: default to updating them -->
<!-- skip spec updates only for purely mechanical changes (formatting, typo, dep bump) -->

## NON-GOALS

<!-- TODO: define what this project explicitly will not do -->

## DEFINITIONS

<!-- TODO: define key terms used throughout the project -->
