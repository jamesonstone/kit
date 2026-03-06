# SPEC

## SUMMARY

Adds an optional `BRAINSTORM.md` artifact and makes `kit brainstorm` the interactive, planning-only entrypoint before `kit spec`. Removes `kit oneshot` and all git branch automation while preserving the existing spec → plan → tasks → implement → reflect workflow.

## PROBLEM

Kit currently treats brainstorming as an external or standalone activity, while the formal workflow starts at `SPEC.md`. That creates a gap between early research and canonical project documents, duplicates functionality with `kit oneshot`, and couples feature creation to branch automation that is outside Kit's core document-centered purpose.

## GOALS

- create or reuse `docs/specs/<feature>/` from `kit brainstorm`
- create `BRAINSTORM.md` as an optional first-class feature artifact
- make `kit brainstorm` interactive by default with two inputs: feature name and user thesis
- emit a planning-only prompt that begins with `/plan`
- require the coding agent to continue asking questions until `>=95%` understanding of the problem and solution strategy
- require the coding agent to persist findings to `BRAINSTORM.md`
- reference `BRAINSTORM.md` from `spec`, `plan`, `tasks`, `implement`, and `reflect` when present
- remove `kit oneshot` from code, docs, and help output
- remove git branch automation from commands, config, and docs
- show `brainstorm` as an optional pre-spec phase in visible workflow messaging

## NON-GOALS

- make `BRAINSTORM.md` a hard prerequisite for `SPEC.md`
- add implementation or build execution to `kit brainstorm`
- invent a new non-markdown artifact format
- change feature directory naming rules
- expand Kit into git workflow management

## USERS

- engineers starting a new feature who need structured codebase-aware research
- coding agents that need a canonical research document before writing `SPEC.md`
- maintainers who need a simpler workflow without `oneshot` or branch automation

## REQUIREMENTS

- `kit brainstorm` must normalize and validate feature names using existing feature naming rules
- `kit brainstorm` must create or reuse the numbered feature directory under `docs/specs/`
- `kit brainstorm` must create `BRAINSTORM.md` if missing and keep it in the feature directory
- `kit brainstorm` must ask for a multiline issue/feature thesis in interactive mode
- the generated prompt must start with `/plan`
- the generated prompt must explicitly forbid implementation/build work and keep the agent in information-gathering mode
- the generated prompt must require repeated clarification until `>=95%` understanding
- the generated prompt must instruct the agent to update `BRAINSTORM.md` with filepath-specific findings
- downstream command prompts must include `BRAINSTORM.md` when the file exists
- features with `BRAINSTORM.md` but no `SPEC.md` must be represented distinctly from `spec` phase features
- `kit status`, rollup output, and handoff/help messaging must reflect the brainstorm phase accurately
- all `oneshot` command code and references must be removed
- all branch automation code, flags, config fields, and references must be removed

## ACCEPTANCE

- running `kit brainstorm` interactively creates `docs/specs/<n>-<slug>/BRAINSTORM.md`
- the brainstorm prompt begins with `/plan`
- the brainstorm prompt instructs the agent to research the full codebase, avoid implementation, and continue clarifying until `>=95%` understanding
- `kit status` shows brainstorm-only features without mislabeling them as `spec`
- `kit spec`, `kit plan`, `kit tasks`, `kit implement`, and `kit reflect` reference `BRAINSTORM.md` when present
- `kit oneshot` is no longer available from the CLI or help output
- repository config and docs contain no active branch automation guidance
- help text and README show brainstorming as an optional pre-spec phase
- automated tests cover brainstorm prompt generation and brainstorm phase detection

## EDGE-CASES

- feature directory already exists with `BRAINSTORM.md`
- feature directory already exists with `SPEC.md`, `PLAN.md`, or `TASKS.md`
- brainstorm-only features coexist with later-phase features in status and rollup output
- user provides a feature name that needs normalization or fails slug validation
- user enters an empty thesis or interrupts interactive input
- downstream commands run for features that do not have `BRAINSTORM.md`

## OPEN-QUESTIONS

- none
