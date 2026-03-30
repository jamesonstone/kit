# SPEC

## SUMMARY

- Add a new `kit catchup [feature]` command that outputs a prompt for a coding agent to recover the current state of a selected feature before any implementation resumes.
- The command must be prompt-only, feature-scoped, and explicitly keep the agent in plan mode until the user approves moving into implementation.

## PROBLEM

- Kit has `status`, `handoff`, `summarize`, and `implement`, but no feature-scoped command dedicated to helping a coding agent catch up on a feature's current stage and state without drifting into execution too early.
- Users who return to an in-flight feature often need a lightweight “resume and orient” prompt rather than a full handoff or immediate implementation context.
- Without an explicit catch-up step, agents can skip clarification, miss recent state encoded in repo artifacts, or start implementing before the user confirms the next move.

## GOALS

- Add `kit catchup [feature]` as a root-level prompt-output command.
- Show an interactive feature selector when no feature argument is provided.
- Let the selector include all feature directories under `docs/specs/` and display the phase beside each feature.
- Derive the selected feature's current stage and state from Kit's existing feature/status model.
- Output a prompt that tells the coding agent to catch up on the selected feature, stay in plan mode, ask questions first, and ask explicitly before starting implementation.
- Reuse the standard prompt-output contract already used by Kit prompt commands.

## NON-GOALS

- Replacing `kit handoff` for project-wide or cross-agent session transfer.
- Replacing `kit summarize` for conversation-context summarization.
- Replacing `kit implement` for implementation execution prompts.
- Adding project-wide catch-up mode in this phase.
- Running implementation, tests, or repo mutations directly from `kit catchup`.

## USERS

- Users returning to a feature after a pause who need a coding agent to recover the feature state before taking action.
- Users switching back into an in-flight feature without needing the broader project-wide handoff flow.
- Coding agents that need a narrow, explicit “resume the feature and ask questions first” instruction set.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## REQUIREMENTS

- Expose a new root command `kit catchup [feature]`.
- When no feature argument is provided, show an interactive numbered selector for all features returned by `feature.ListFeatures`.
- The selector must show each feature's directory name and current phase.
- The command must support `--copy` and `--output-only` and must use `outputPrompt(...)`.
- The command must resolve the selected feature using existing feature-resolution logic.
- The generated prompt must begin with `/plan`.
- The generated prompt must identify the selected feature's current stage and current state.
- “Stage” must come from the existing feature phase model.
- “State” must summarize current artifact presence, task progress when available, and the current next-action guidance.
- The prompt must instruct the agent to read `CONSTITUTION.md`, `PROJECT_PROGRESS_SUMMARY.md`, and the feature docs in order: `BRAINSTORM.md` when present, then `SPEC.md`, `PLAN.md`, `TASKS.md`.
- The prompt must instruct the agent to stay in plan mode and not implement yet.
- The prompt must instruct the agent to start by asking clarifying questions.
- The prompt must instruct the agent to ask explicitly before moving from catch-up/planning into implementation.
- The prompt may reference `kit summarize <feature>` only as an optional aid for missing conversation context.
- If the selected feature is already `complete`, the prompt must treat catch-up as review/reopen triage only and must not assume more implementation is needed.
- The command must call `printWorkflowInstructions(...)` after prompt output when not in `--output-only` mode.

## ACCEPTANCE

- Running `kit catchup` with no arguments shows an interactive feature selector with phases.
- Running `kit catchup <feature>` outputs a feature-scoped catch-up prompt.
- The prompt starts with `/plan`.
- The prompt states the selected feature's current stage and current state.
- The prompt instructs the agent to ask questions first and remain in plan mode.
- The prompt tells the agent to request explicit approval before starting implementation.
- The prompt does not duplicate the full project-wide `handoff` content or the full summarization workflow.
- The prompt references `kit summarize <feature>` only as an optional support command.
- A `complete` feature produces a review/reopen-style catch-up prompt rather than an implementation-start prompt.
- Help and README document the new command distinctly from `handoff`, `summarize`, and `implement`.

## EDGE-CASES

- No features exist in `docs/specs/`.
- The selected feature has only `BRAINSTORM.md`.
- The selected feature has `SPEC.md` or `PLAN.md` but no `TASKS.md`.
- The selected feature has `TASKS.md` with no checkbox progress.
- The selected feature is already `complete`.
- The selected feature has missing conversation context, but the repository docs are sufficient to reconstruct current state.

## OPEN-QUESTIONS

- none
