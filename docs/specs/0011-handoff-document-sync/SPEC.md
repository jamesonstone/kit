# SPEC

## SUMMARY

- Change `kit handoff` from a passive “new session context dump” into an active prompt for the current coding agent session to reconcile feature docs with implementation reality before transfer.
- Require the generated prompt to produce a concise final response that confirms documentation sync, includes a full-path document table, and summarizes the most recent conversation context.

## PROBLEM

- `kit handoff` currently focuses on orienting a fresh agent session, but it does not explicitly require the current session to update feature docs so they match the actual implementation before handoff.
- That leaves too much room for stale `SPEC.md`, `PLAN.md`, `TASKS.md`, and rollup data to survive into the next session.
- The current prompt also does not enforce a standard final response that confirms doc sync, lists the authoritative file paths, and captures recent conversation context in a reusable way.

## GOALS

- Make `kit handoff` tell the current coding agent session to reconcile documentation with implementation reality before handoff.
- Preserve the current command surface:
  - `kit handoff [feature]`
  - interactive selector with `0` for project-wide mode
- Accept `--prompt-only` as a consistency flag so users can explicitly request the selected handoff prompt without changing repo state.
- For feature scope, require the prompt to review and update:
  - `CONSTITUTION.md`
  - optional `BRAINSTORM.md`
  - `SPEC.md`
  - `PLAN.md`
  - `TASKS.md`
  - optional `ANALYSIS.md`
  - `PROJECT_PROGRESS_SUMMARY.md`
- For project-wide scope, require the prompt to review `PROJECT_PROGRESS_SUMMARY.md` plus active feature docs that are out of sync with implementation reality.
- Include a markdown table in the generated prompt that shows:
  - file name
  - full absolute filesystem path
  - concise guidance for how to use the document
- Require the prompt to verify dependency inventories in touched `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` docs before handoff.
- Require the prompted agent's final response to:
  - confirm documentation sync in stdout/chat
  - include a concise documentation table with full paths
  - summarize the most recent conversation context into high-signal facts

## NON-GOALS

- Changing `kit handoff` selector behavior or argument shape.
- Reverting clipboard-first transport behavior.
- Changing `kit summarize` or `kit catchup`.
- Running implementation work unrelated to documentation reconciliation.

## USERS

- Users preparing to hand off an in-flight feature to another session.
- Users who need documentation to become the source of truth before context transfer.
- Coding agents that need an explicit doc-sync-and-summarize workflow before handoff.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## REQUIREMENTS

- `kit handoff` must keep the existing selector and direct feature-argument behavior.
- The generated prompt must address the current coding agent session, not a hypothetical fresh session.
- The generated prompt must explicitly instruct the agent to compare implementation reality against the relevant docs and update the docs first when they diverge.
- For feature scope, the prompt must include a documentation inventory table with columns:
  - `File`
  - `Full Path`
  - `How To Use`
- The feature-scope documentation inventory table must include every relevant existing document among:
  - `CONSTITUTION.md`
  - `BRAINSTORM.md`
  - `SPEC.md`
  - `PLAN.md`
  - `TASKS.md`
  - `ANALYSIS.md`
  - `PROJECT_PROGRESS_SUMMARY.md`
- All file paths shown in the prompt must be absolute filesystem paths.
- The prompt must instruct the agent to update `PROJECT_PROGRESS_SUMMARY.md` when the reconciled feature docs change the current feature state.
- In project-wide mode, the prompt must instruct the agent to reconcile `PROJECT_PROGRESS_SUMMARY.md` plus active feature docs that are inconsistent with repository reality.
- The prompt must instruct the agent to verify or refresh `## DEPENDENCIES` tables in touched `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` docs, including exact locations for design dependencies.
- The prompt must include explicit instructions for summarizing the most recent conversation context into high-signal facts covering:
  - decisions made
  - blockers
  - validation or verification results
  - open questions
  - next steps
- The prompt must define a final response contract that tells the agent to output:
  - a concise confirmation that documentation files and dependency inventories are up to date
  - a markdown table of documentation files with full paths and usage guidance
  - a concise recent-context summary
- If docs are not up to date, the prompt must instruct the agent to update them first and only then produce the final confirmation/table/summary.
- The prompt must remain prompt-only; Kit itself still does not update the docs directly.
- `kit handoff` must accept `--prompt-only`; because the command already only emits prompts, the flag may reuse the normal generation path and must not introduce any new repo mutations.

## ACCEPTANCE

- Running `kit handoff <feature>` outputs a prompt that tells the current agent session to reconcile feature docs with implementation reality before handoff.
- The feature-scoped prompt contains a markdown document table with `File`, `Full Path`, and `How To Use`.
- The feature-scoped prompt references absolute paths for the relevant docs.
- Running `kit handoff` in project-wide mode outputs a prompt that tells the current agent session to reconcile rollup and active feature docs before handoff.
- The prompt includes a final response contract that requires:
  - documentation-sync confirmation
  - documentation table
  - recent conversation-context summary
- The prompt explicitly requires dependency inventory verification for touched feature docs.
- The prompt includes explicit instructions for summarizing recent conversation context into high-signal facts.
- `kit handoff --prompt-only <feature>` is accepted and preserves the existing prompt-only behavior.

## EDGE-CASES

- The selected feature only has `BRAINSTORM.md`.
- The selected feature has implementation evidence that outpaced `TASKS.md` or `PLAN.md`.
- `ANALYSIS.md` does not exist.
- Project-wide mode finds multiple active features with stale docs.
- Recent conversation context includes decisions that were never persisted to repo docs.

## OPEN-QUESTIONS

- none
