# Guardrails

## Hard Rules

- `docs/CONSTITUTION.md` is the canonical project contract
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree
- Never mix multiple features in one `docs/specs/<feature>/` directory
- Update docs first when reality diverges from documented behavior

## Completion Bar

- Populate all required sections in `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md`
- Replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
- Always update affected documentation and ensure touched docs are current and properly formatted before calling work complete
- Fix relevant lint and test failures before calling work complete
- Keep dependency and relationship sections current when those docs are touched

## Code Hygiene

- Remove dead code, unused exports, and public surfaces that are not strictly necessary
- If a symbol is only used locally, reduce its visibility instead of keeping it exported

## Safety

- Prefer explicit error handling over silent failure
- Keep changes minimal and reversible
- Do not run `git add` or `git commit` without explicit approval
- Do not run `coderabbit --prompt-only` unless explicitly requested or approved
