# GitHub Copilot Repository Instructions

## Native Planning

Use native planning for research and design. Before implementation, inspect code and repository documentation, then decide whether material rationale requires a living `SPEC.md`. Capture the accepted plan before code when it does. After validation, load `docs/references/rules/constitution-curation.md` and curate durable decisions into the correct repository document; code-and-test-sufficient work may report that no documentation update was required.

Start with `docs/agents/README.md`. Before Git, GitHub, or AWS mutations, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*`. Repo-local Kit rules outrank generic defaults.

## Final Response

Every implementation final response must include:

- Repository Memory
- Decision: created | updated | refactored | deleted | not required
- Rationale: why this persistence decision is correct
- Artifacts: paths or none
