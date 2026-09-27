# Kit Overview

Kit is a small, vendor-neutral harness for coding agents. It keeps the few
project-level invariants that make agents safer and more correct, gives them a
place for durable project memory, and provides deterministic commands where
code is more reliable than model judgment.

## What Kit Does

- Renders one universal agent contract into `AGENTS.md`, `CLAUDE.md`, and
  `.github/copilot-instructions.md` inside a Kit-managed block.
- Ships contextual rules with each release and installs the core set under
  `docs/references/rules/`; agents read a rule only when the contract's trigger
  for it applies.
- Scaffolds project memory: `docs/CONSTITUTION.md`, `docs/specs/<feature>/SPEC.md`
  (with worktree-safe numbering), and `docs/references/testing.md`.
- Validates and converges Kit-managed state (`kit check`, `kit reconcile`,
  `kit health`, `kit registry status`).
- Verifies AWS identity before AWS work (`kit aws verify`).
- Keeps local, minimal usage evidence (`kit usage`).

## What Kit Deliberately Does Not Do

- Plan, route context, or choose which files an agent reads.
- Describe host tools, models, or delegation mechanics; hosts expose those.
- Generate prompts that wrap user intent in orchestration boilerplate.
- Fetch rules or instructions from the network at runtime.
- Launch, supervise, or orchestrate agents, or call a model.

## Core Artifacts

| Artifact | Role |
| --- | --- |
| `.kit.yaml` | Project configuration, installed-rule state, optional `source_file_line_limit` and AWS context |
| `AGENTS.md`, `CLAUDE.md`, Copilot instructions | The universal contract in a Kit-managed block; project guidance goes outside it |
| `docs/references/rules/*.md` | Contextual rules shipped with the Kit binary |
| `docs/CONSTITUTION.md` | Demonstrated project-wide invariants |
| `docs/specs/<feature>/SPEC.md` | Living feature rationale, plan, validation, and outcome |
| `docs/references/testing.md` | The project's validation commands |
