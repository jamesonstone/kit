# SPEC

## SUMMARY

- Add a new `kit dispatch` command that outputs a prompt for a coding agent to discover likely file overlap across a pasted task set, cluster overlapping work, and queue subagents safely.
- The command must be prompt-only and must force a discovery-and-approval step before any subagent execution begins.

## PROBLEM

- Kit has prompt generators for planning, catch-up, implementation, reflection, and skill mining, but no prompt-only command specialized for turning a raw task list into a safe subagent dispatch plan.
- When users hand a coding agent a mixed set of bullets, numbered items, and paragraphs, the agent can parallelize too aggressively and create conflicting edits across the same files.
- Users need a deterministic prompt that tells the coding agent to discover first, predict touched files, merge ambiguous overlap conservatively, and wait for approval before launching subagents.

## GOALS

- Add `kit dispatch` as a root-level prompt-output command.
- Make naked `kit dispatch` interactive and default to a vim-compatible editor for multi-line task capture.
- Support `--file` and piped stdin with precedence `--file` > stdin > default editor-backed prompt.
- Support `--vim` and `--editor` for editor-backed task capture.
- Show a short step-specific instruction screen before the default editor opens and wait for any key before launch.
- Normalize top-level paragraphs, bullets, and numbered items into dispatchable tasks while keeping nested items attached to their parent task.
- Output a prompt that requires discovery first, clustering by predicted file overlap, conservative handling of ambiguity, and explicit approval before subagent launch.
- Support `--max-subagents` with default `10` and minimum `1`.
- Reuse the standard prompt-output contract already used by Kit prompt commands.

## NON-GOALS

- Executing subagents directly from the Kit binary.
- Requiring an existing feature directory or `TASKS.md`.
- Writing a new canonical artifact type for dispatch plans.
- Building a full parser for arbitrary markdown constructs beyond top-level paragraphs and list items.

## USERS

- Users who want to hand a coding agent a free-form task set and have it plan safe subagent delegation.
- Coding agents that need deterministic instructions for discovery, overlap clustering, queueing, and approval gating before execution.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## REQUIREMENTS

- [SPEC-01] Expose a new root command `kit dispatch`.
- [SPEC-02] The command must accept no positional arguments.
- [SPEC-03] The command must support `--copy` and `--output-only` and must use `outputPrompt(...)`.
- [SPEC-04] The command must support `--file <path>` for reading the raw task set from disk.
- [SPEC-05] The command must support `--vim` and `--editor` for editor-backed task capture.
- [SPEC-06] The command must support `--max-subagents <n>`, default `10`, minimum `1`.
- [SPEC-07] If `--file` is present, input must come from that file regardless of stdin or editor flags.
- [SPEC-08] If `--file` is absent and stdin is piped, input must come from stdin.
- [SPEC-09] If neither `--file` nor piped stdin is present, the command must enter interactive editor-backed capture mode.
- [SPEC-10] The default interactive capture mode must open a vim-compatible editor for multi-line task input.
- [SPEC-11] `--editor` must allow overriding the default vim-compatible editor when no higher-precedence input source is active.
- [SPEC-12] Before opening the editor, the command must show a short instruction screen describing the step and what should be pasted into the editor.
- [SPEC-13] After the instruction screen, the command must wait for any key before opening the editor.
- [SPEC-14] The command must normalize the raw task set into dispatchable tasks using only:
  - top-level paragraphs
  - top-level bullet items
  - top-level numbered items
- [SPEC-15] Nested bullets or numbered items must remain attached to their parent task as task detail, not become separate tasks.
- [SPEC-16] Empty or whitespace-only task input must fail with an actionable error.
- [SPEC-17] The generated prompt must identify the effective max-subagent cap explicitly.
- [SPEC-18] The generated prompt must instruct the coding agent to start with discovery, not execution.
- [SPEC-19] The generated prompt must instruct the coding agent to predict likely touched files for each normalized task before dispatching work.
- [SPEC-20] The generated prompt must instruct the coding agent to cluster tasks by predicted file overlap.
- [SPEC-21] The generated prompt must instruct the coding agent to assign one subagent per overlap cluster and preserve task order within each cluster.
- [SPEC-22] The generated prompt must instruct the coding agent to parallelize only disjoint clusters.
- [SPEC-23] The generated prompt must instruct the coding agent to merge low-confidence or ambiguous overlap into the same cluster instead of parallelizing it.
- [SPEC-24] The generated prompt must require a dry-run discovery report before any subagent execution.
- [SPEC-25] The dry-run discovery report must include:
  - normalized tasks
  - predicted touched files per task
  - overlap clusters
  - dispatch queue
  - subagent assignments
  - risks and unknowns
- [SPEC-26] The generated prompt must instruct the coding agent to wait for explicit user approval after the dry-run report and before launching any subagents.
- [SPEC-27] The command must call `printWorkflowInstructions(...)` after prompt output when not in `--output-only` mode.
- [SPEC-28] Root help and README must document the new command.

## ACCEPTANCE

- Running `kit dispatch` with no stdin or file opens a vim-compatible editor for task capture.
- Running `kit dispatch` with no stdin or file first shows a short instruction screen and waits for any key before opening the editor.
- Running `kit dispatch --file tasks.txt` reads the task set from `tasks.txt`.
- Running `cat tasks.txt | kit dispatch` reads the task set from stdin.
- Running `kit dispatch --vim` behaves the same as the default interactive editor-backed capture flow.
- The command rejects empty task input.
- The generated prompt states the effective max-subagent cap.
- The generated prompt requires discovery first, conservative overlap clustering, and explicit approval before subagent launch.
- The generated prompt includes the required dry-run sections.
- Help and README expose `kit dispatch` as a prompt-only command.

## EDGE-CASES

- The task set mixes paragraphs, bullets, and numbered items.
- Top-level list items contain nested bullets or numbered sub-steps.
- The input is a single paragraph with no blank lines.
- The input contains multiple blank lines between task groups.
- `--file` is passed together with piped stdin.
- `--vim` or `--editor` is passed together with `--file` or piped stdin.
- `--max-subagents` is `0` or negative.
- The file passed to `--file` does not exist.

## OPEN-QUESTIONS

- none
