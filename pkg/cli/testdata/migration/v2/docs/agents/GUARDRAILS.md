# Guardrails

## Hard Rules

- `docs/CONSTITUTION.md` is the canonical project contract
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree
- If the user message includes an attached pasted-text file and the visible message is empty or minimal, treat the attachment as the active task instructions unless the user says otherwise
- If the attachment appears Kit-generated, follow it directly without asking what the attachment is for
- Never mix multiple features in one `docs/specs/<feature>/` directory
- Update docs first when reality diverges from documented behavior

## Completion Bar

- Populate all required sections in `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, and `TASKS.md`
- Replace placeholder-only sections with `not applicable`, `not required`, or `no additional information required`
- Always update affected documentation and ensure touched docs are current and properly formatted before calling work complete
- Never claim tests passed unless they ran
- Never claim files were inspected unless they were inspected
- Never guess file contents, APIs, or behavior
- If validation cannot run, state why
- Fix relevant lint and test failures before calling work complete
- Keep canonical front matter references and relationships current when those docs are touched

## Code Hygiene

- Remove dead code, unused exports, and public surfaces that are not strictly necessary
- If a symbol is only used locally, reduce its visibility instead of keeping it exported
- Keep implementation/source code files around 300 lines or less when splitting improves clarity
- Do not apply the 300-line guideline to documentation files, `docs/**`, `.kit/**`, or `.kit.yaml`

## Safety

- Prefer explicit error handling over silent failure
- Keep changes minimal and reversible
- Do not run `git add` or `git commit` without explicit approval
- Do not run `coderabbit --prompt-only` unless explicitly requested or approved
