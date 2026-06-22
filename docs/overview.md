# Kit Overview

Kit v2 is a general-purpose harness for disciplined thought work.

The shipped command surface is packaged around repositories and software
delivery, but the underlying model is broader: constraints, clarification,
planning, execution, verification, reflection, evidence, and transfer are
useful in any serious domain.

## Principles

- 🧰 **Harness-first, workflow-second** - Kit coordinates work without locking it to one agent vendor.
- 📄 **Documents are the source of truth** - durable decisions live in files, not only in chat.
- 🧠 **Spec-driven planning is the strongest engine** - ambiguous or high-risk work starts with explicit structure.
- ⚡ **Ad hoc work stays lightweight** - small changes do not need a full feature workflow.
- 🤝 **Portable by default** - generated prompts are meant for capable coding agents, not one runtime.
- 🔍 **Explicit gates beat hidden automation** - issue, branch, PR, delivery, and validation boundaries stay visible.
- 🔄 **Reflection closes the loop** - correctness, evidence, docs, and handoff matter after implementation.

## Cross-Domain Concepts

| Kit Concept | In Software | In Research | In Strategy / Ops | In Writing / Policy |
| --- | --- | --- | --- | --- |
| `CONSTITUTION.md` | Engineering constraints | Study constraints | Operating principles | Editorial or policy constraints |
| `SPEC.md` | Feature workflow artifact | Research question, study plan, proof | Decision brief, rollout, evidence | Argument, outline, revision evidence |
| Acceptance criteria | Binary behavior checks | Falsifiable success criteria | Decision or rollout gates | Editorial acceptance standards |
| Validation map and evidence | Tests, runtime checks, docs review | Result evidence and audit trail | Operational validation | Source/proof trail and revision notes |
| Legacy `BRAINSTORM.md` / `PLAN.md` / `TASKS.md` | Historical staged artifacts | Historical staged artifacts | Historical staged artifacts | Historical staged artifacts |
| `reconcile` / `resume` / `summarize` / `handoff` | Reconcile, resume, or transfer context | Resume investigation | Transfer project state | Transfer editorial context |

## Artifact Model

Feature artifacts use typed YAML front matter for canonical metadata such as:

- artifact identity
- feature identity
- relationships
- references
- skills
- summary or intent
- workflow phase

New v2 feature work uses `SPEC.md` as the single durable feature artifact.
Deprecated v1 staged artifacts remain readable during migration and are only
binding when a user explicitly chooses `kit legacy`.

## Positioning

Kit is broader than a spec generator. It supports:

- structured planning when scope is unclear
- lightweight ad hoc execution when the change is contained
- recovery tools such as `reconcile`, `resume`, `summarize`, and `handoff`
- review and orchestration tools such as `code-review`, `dispatch`, and `loop review`

Spec-driven development remains a core engine inside the harness, not the whole
identity of the tool.

## Inspiration

Kit is inspired by GitHub's [spec-kit](https://github.com/github/spec-kit),
which pioneered specification-driven development. Kit keeps that discipline
where it helps most, then broadens it into a lighter, more portable,
general-purpose harness.
