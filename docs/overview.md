# Kit Overview

Kit is a repository-memory and specification harness for agent-driven work.

The shipped command surface is packaged around repositories and software
delivery, but the underlying model is broader: constraints, clarification,
native planning, execution, verification, curated rationale, and transfer are
useful in any serious domain.

## Principles

- 🧰 **Harness-first, workflow-second** - Kit coordinates work without locking it to one agent vendor.
- 📄 **Repositories own durable memory** - consequential decisions live in canonical files, not only in chat or transcripts.
- 🧠 **Native planning owns design** - use the host agent's planning capability for research, clarification, design, and implementation planning.
- 📐 **Specifications preserve material why** - create or adopt a living spec before code when future agents need rationale that code and tests cannot recover.
- ⚡ **Ad hoc work stays lightweight** - small changes do not need a full feature workflow.
- 🤝 **Portable by default** - generated prompts are meant for capable coding agents, not one runtime.
- 🔍 **Explicit gates beat hidden automation** - issue, branch, PR, delivery, and validation boundaries stay visible.
- 🔄 **Reflection closes the loop** - correctness, evidence, docs, and handoff matter after implementation.

## Cross-Domain Concepts

| Kit Concept | In Software | In Research | In Strategy / Ops | In Writing / Policy |
| --- | --- | --- | --- | --- |
| `CONSTITUTION.md` | Engineering constraints | Study constraints | Operating principles | Editorial or policy constraints |
| `SPEC.md` | Feature workflow artifact | Research question, study plan, proof | Decision brief, rollout, evidence | Argument, outline, revision evidence |
| Requirements and observable acceptance | Behavior checks | Falsifiable success criteria | Decision or rollout gates | Editorial acceptance standards |
| Validation and outcome | Tests, runtime checks, docs review | Result evidence and audit trail | Operational validation | Source/proof trail and revision notes |
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

New feature memory uses a compact V3 `SPEC.md`: purpose, context, requirements,
accepted plan, decisions, discoveries, validation, outcome, and repository
memory. V1 and V2 artifacts remain readable compatibility inputs and are never
mechanically rewritten into V3 because migration requires semantic curation.

## Positioning

Kit is broader than a spec generator. It supports:

- durable plan and decision capture after native planning
- lightweight ad hoc execution when the change is contained
- recovery tools such as `reconcile`, `resume`, `summarize`, and `handoff`
- review and orchestration tools such as `code-review`, `dispatch`, and `loop review`

The key judgment is semantic: persist the why when it matters, and prefer a
clear `not required` decision when code and tests are sufficient.

## Inspiration

Kit is inspired by GitHub's [spec-kit](https://github.com/github/spec-kit),
which pioneered specification-driven development. Kit keeps that discipline
where it helps most, then broadens it into a lighter, more portable,
general-purpose harness.
