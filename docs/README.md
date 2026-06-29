# Kit Documentation

Use this directory as the durable documentation map for Kit. The root
[README](../README.md) is intentionally short and routes here for detail.

## Guides

| Guide | Purpose |
| --- | --- |
| 🧭 [Overview](overview.md) | What Kit is, where it applies, and why the harness is broader than software. |
| ⚙️ [Commands](commands.md) | Installation, command groups, prompt behavior, prompt libraries, and scaffold refresh. |
| 🔁 [Workflows](workflows.md) | V2 single-`SPEC.md` workflow, legacy v1 foundations, loops, usage examples, and project structure. |

## Project Contract

| Document | Purpose |
| --- | --- |
| 🧱 [CONSTITUTION.md](CONSTITUTION.md) | Project invariants, workflow definitions, and completion rules. |
| 📈 [PROJECT_PROGRESS_SUMMARY.md](PROJECT_PROGRESS_SUMMARY.md) | Current feature index and progress summary. |

## Agent Runtime Docs

| Document | Purpose |
| --- | --- |
| 🤖 [agents/README.md](agents/README.md) | Agent routing table and loading rules. |
| 🛡️ [agents/GUARDRAILS.md](agents/GUARDRAILS.md) | Completion, safety, validation, and GitHub delivery hard gates. |
| 🔁 [agents/WORKFLOWS.md](agents/WORKFLOWS.md) | Spec-driven versus ad hoc flow semantics. |
| 🧠 [agents/RLM.md](agents/RLM.md) | Just-in-time context loading and progressive disclosure. |
| 🧰 [agents/TOOLING.md](agents/TOOLING.md) | Command capability discovery, dispatch, and tool routing. |

## References

| Document | Purpose |
| --- | --- |
| 📌 [references/README.md](references/README.md) | Durable reference index. |
| 🧪 [references/testing.md](references/testing.md) | Testing guidance. |
| 🛠️ [references/tooling.md](references/tooling.md) | Tooling notes. |
| 🌐 [references/external-systems.md](references/external-systems.md) | External system references. |
| 📜 [references/rules/](references/rules/) | Repo-local rulesets for agents and Kit-managed workflows. |

## Specs And Notes

- `docs/specs/<feature>/SPEC.md` is the v2 durable feature workflow artifact.
- Legacy `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files may exist under feature directories as historical staged artifacts.
- `docs/notes/<feature>/` stores optional source material, screenshots, references, draft responses, design inputs, and gitignored private conversation context.
- `docs/references/rules/feature-notes.md` defines how agents load, reference, promote, or ignore feature notes.
