# SPEC

## SUMMARY

- Make `kit scaffold-agents` safer when repository instruction files already exist by adding an overwrite confirmation gate and an explicit append-only mode.
- Keep append-only deterministic and constrained to known Kit-managed sections; do not attempt fuzzy free-form merges.

## PROBLEM

- `kit scaffold-agents` currently has binary behavior:
  - skip existing files by default
  - overwrite existing files blindly with `--force`
- That makes it easy to destroy customized `AGENTS.md`, `CLAUDE.md`, or `.github/copilot-instructions.md` content by accident.
- The command also has no supported middle path for preserving custom content while adding newly introduced Kit-managed sections.

## GOALS

- Add a confirmation gate for `kit scaffold-agents --force` when selected target files already exist.
- Add a non-interactive escape hatch so scripts can bypass the confirmation gate explicitly.
- Add an explicit `--append-only` mode that preserves existing content while adding only missing Kit-managed sections.
- Keep append-only constrained to known Kit-managed sections and explicit markdown anchors.
- Fail safely in append-only mode when the existing file does not contain recognizable anchors for a deterministic merge.
- Suggest append-only as the safer next step when normal scaffolding skips existing files.
- Support both the confirmation gate and append-only mode in the same feature release.

## NON-GOALS

- Making append-only the default write mode.
- Implementing fuzzy semantic merge of arbitrary prose.
- Overwriting or rewriting user-authored content inside existing matched sections in append-only mode.
- Adding a Git-hook-style extension system.
- Changing the scaffolded instruction template semantics beyond what is already defined elsewhere.

## USERS

- Users who want to refresh instruction files without accidentally losing custom guidance.
- Users who want to adopt new Kit-managed sections incrementally in existing instruction files.
- Teams scripting `kit scaffold-agents --force` in automation and needing an explicit non-interactive bypass.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER                       | REQUIRED |
| ----- | ------ | ---- | ----------------------------- | -------- |
| none  | n/a    | n/a  | no additional skills required | no       |

## REQUIREMENTS

- `kit scaffold-agents --force` must prompt for confirmation when at least one selected target file already exists.
- The overwrite confirmation prompt must list the exact files that will be overwritten.
- The overwrite confirmation prompt must accept `y` / `yes` to proceed and treat all other input as cancellation.
- Cancelling the overwrite confirmation must leave all target files unchanged.
- Add `--yes` / `-y` so `kit scaffold-agents --force --yes` skips the confirmation prompt.
- `--yes` must be documented as an automation escape hatch for overwrite mode.
- Add `--append-only` as an explicit write mode for `kit scaffold-agent` / `kit scaffold-agents`.
- `--append-only` must be mutually exclusive with `--force`.
- `--append-only` must merge only known Kit-managed top-level markdown sections from the scaffold template.
- In append-only mode:
  - existing matched sections must remain byte-for-byte unchanged
  - missing Kit-managed sections must be inserted in template order
  - extra user-defined sections must be preserved
  - non-existent target files must be created from the template
- Append-only mode must fail safely when an existing target file has no recognizable Kit-managed section anchors.
- Append-only mode must fail safely when an existing target file contains duplicate recognized Kit-managed section headings that make the merge ambiguous.
- When append-only mode fails for any selected file, the command must stop before writing partial results to any target file.
- When normal scaffolding skips existing files, the command output must suggest `--append-only` as the non-destructive alternative and `--force` as the destructive alternative.
- Existing targeted-selection behavior (`--agentsmd`, `--claude`, `--copilot`) must continue to work in all write modes.
- The singular alias `kit scaffold-agent` must continue to work with the new flags and semantics.
- `README.md` and `docs/specs/0000_INIT_PROJECT.md` must describe the new flags and behaviors.

## ACCEPTANCE

- Running `kit scaffold-agents --force` against existing instruction files prompts before overwriting.
- Running `kit scaffold-agents --force --yes` overwrites without prompting.
- Running `kit scaffold-agents --append-only` adds missing Kit-managed sections while preserving existing matched section content.
- Append-only mode inserts missing sections in template order rather than appending them blindly to the end.
- Append-only mode preserves extra user-authored sections.
- Append-only mode fails with actionable guidance when no recognizable anchors exist.
- When default scaffolding skips existing files, the output suggests `--append-only` and `--force`.
- Targeted selection and the `scaffold-agent` alias continue to work.
- Automated tests cover overwrite confirmation, append-only merge behavior, failure-safe preflight behavior, and CLI flag validation.

## EDGE-CASES

- One selected file exists and another does not.
- All selected files exist but only some are mergeable in append-only mode.
- A custom `AGENTS.md` contains no recognizable Kit-managed sections.
- A custom instruction file contains duplicate recognized section headings.
- A user passes `--append-only --force`.
- A user passes `--yes` without `--force`.
- A user uses `--append-only` with only `--copilot` selected.

## OPEN-QUESTIONS

- none
