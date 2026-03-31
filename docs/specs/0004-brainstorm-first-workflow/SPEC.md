# SPEC

## SUMMARY

Adds an optional `BRAINSTORM.md` artifact and makes `kit brainstorm` the interactive, planning-only entrypoint before `kit spec`. Removes `kit oneshot` and all git branch automation while preserving the existing spec → plan → tasks → implement → reflect workflow. Core workflow prompt commands default to copying generated instructions to the clipboard and require `--output-only` for raw stdout prompt output.

## PROBLEM

Kit currently treats brainstorming as an external or standalone activity, while the formal workflow starts at `SPEC.md`. That creates a gap between early research and canonical project documents, duplicates functionality with `kit oneshot`, and couples feature creation to branch automation that is outside Kit's core document-centered purpose.

## GOALS

- create or reuse `docs/specs/<feature>/` from `kit brainstorm`
- create `BRAINSTORM.md` as an optional first-class feature artifact
- make `kit brainstorm` interactive by default with two inputs: feature name and user thesis
- emit a planning-only prompt that begins with `/plan`
- require the coding agent to use numbered lists, ask clarifying questions in batches of up to 10, include a recommended default/proposed solution/assumption for every question, accept `yes` / `y` as full-batch approval and `yes 3, 4, 5` / `y 3, 4, 5` as numbered approval, support `no` / `n` overrides, show percentage-understanding progress after each batch, and continue until the specification is precise enough for a production-quality solution
- require the coding agent to persist findings to `BRAINSTORM.md`
- reference `BRAINSTORM.md` from `spec`, `plan`, `tasks`, `implement`, and `reflect` when present
- require downstream prompts that use the `>=95%` clarification loop to preserve the same approval semantics across `spec`, `plan`, and `tasks`
- default `brainstorm`, `spec`, `plan`, `tasks`, `implement`, and `reflect` to copying generated prompts to the clipboard instead of printing the prompt body to stdout
- require `--output-only` to emit the raw prompt to stdout for `brainstorm`, `spec`, `plan`, `tasks`, `implement`, and `reflect`
- keep `--copy` as an explicit compatibility flag, especially for `--output-only` usage
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

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## REQUIREMENTS

- `kit brainstorm` must normalize and validate feature names using existing feature naming rules
- `kit brainstorm` must create or reuse the numbered feature directory under `docs/specs/`
- `kit brainstorm` must create `BRAINSTORM.md` if missing and keep it in the feature directory
- `kit brainstorm` must ask for a multiline issue/feature thesis in interactive mode
- free-text interactive prompts must accept `Shift+Enter` for newline insertion without submitting
- free-text interactive prompts must preserve consecutive blank lines within submitted content
- free-text interactive prompts must support `--vim` and `--editor=vim` to open a vim-compatible editor
- editor-backed free-text prompts must show a short step-specific instruction screen before opening the editor
- editor-backed free-text prompts must wait for any key press before opening the editor
- in editor mode, save+quit must submit and quit-without-save must cancel or skip
- the generated prompt must start with `/plan`
- the generated prompt must explicitly forbid implementation/build work and keep the agent in information-gathering mode
- the generated prompt must require numbered batched clarification and visible percentage-understanding progress until `>=95%` confidence in the problem and desired solution
- the generated prompt must instruct the agent to update `BRAINSTORM.md` with filepath-specific findings
- downstream command prompts must include `BRAINSTORM.md` when the file exists
- `kit brainstorm`, `kit spec`, `kit plan`, `kit tasks`, `kit implement`, and `kit reflect` must copy generated prompts to the clipboard by default when `--output-only` is not set
- in default mode, those commands must acknowledge that the prompt was copied to the clipboard and must not print the prompt body to stdout
- for those commands, `--output-only` must print the raw prompt to stdout and must suppress automatic clipboard copying unless `--copy` is also set
- `kit brainstorm --output <path>` must continue writing the prepared prompt to the requested file and, in default mode, must also copy the prompt to the clipboard
- features with `BRAINSTORM.md` but no `SPEC.md` must be represented distinctly from `spec` phase features
- `kit status`, rollup output, and handoff/help messaging must reflect the brainstorm phase accurately
- `kit status` must include the running Kit version as minor informational metadata without displacing feature guidance
- all `oneshot` command code and references must be removed
- all branch automation code, flags, config fields, and references must be removed

## ACCEPTANCE

- running `kit brainstorm` interactively creates `docs/specs/<n>-<slug>/BRAINSTORM.md`
- `kit brainstorm` and `kit spec --interactive` allow multiline responses with `Shift+Enter` without accidental submit
- `kit brainstorm --vim` and `kit spec --interactive --vim` open free-text responses in a vim-compatible editor
- `kit brainstorm --vim` and `kit spec --interactive --vim` first show step-specific instructions and wait for any key before opening the editor
- the brainstorm prompt begins with `/plan`
- the brainstorm prompt instructs the agent to research the full codebase, avoid implementation, and use numbered batched clarification with recommended defaults, `yes` / `y` whole-batch approval, `yes 3, 4, 5` / `y 3, 4, 5` numbered approval, `no` / `n` overrides, uncertainties, and visible percentage-understanding progress until the specification is precise enough for a production-quality solution
- `kit status` shows brainstorm-only features without mislabeling them as `spec`
- `kit status` includes the running Kit version while preserving brainstorm-aware feature guidance
- `kit spec`, `kit plan`, and `kit tasks` preserve the same clarification-loop approval semantics, and `kit implement` plus `kit reflect` reference `BRAINSTORM.md` when present
- `kit brainstorm`, `kit spec`, `kit plan`, `kit tasks`, `kit implement`, and `kit reflect` copy their generated prompt to the clipboard by default, print an acknowledgement, and do not print the prompt body unless `--output-only` is passed
- `kit brainstorm --output <path>` still writes the prompt file and also copies the prompt to the clipboard in default mode
- `kit brainstorm --output-only`, `kit spec --output-only`, `kit plan --output-only`, `kit tasks --output-only`, `kit implement --output-only`, and `kit reflect --output-only` print the raw prompt to stdout and only copy when `--copy` is also passed
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
- user enters multiple consecutive blank lines in a free-text response
- user exits the editor without saving a required response
- user passes `--output-only --copy` and expects both raw stdout and clipboard output
- user passes `kit brainstorm --output <path>` with and without `--output-only`
- downstream commands run for features that do not have `BRAINSTORM.md`

## OPEN-QUESTIONS

- none
