# PLAN

## SUMMARY

- Add a new prompt-only CLI surface under `skill mine` and `skills mine`.
- Reuse existing selection, output, and workflow-instruction patterns so the new command behaves like the current prompt-output commands.
- Add config support for a canonical transferable skills directory, extend the prompt to derive higher-order insights, and add a mandatory stale-skill audit section with Claude mirror cleanup.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-32][SPEC-36] Extend config with a configurable canonical skills directory and keep the prompt explicit about the canonical root versus the Claude mirror root.
- [PLAN-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Implement a new `pkg/cli/skill.go` file that registers `skill` and `skills` as separate top-level commands sharing the same `mine` subcommand behavior.
- [PLAN-03][SPEC-07][SPEC-08] Reuse the feature-list pattern from `implement`/`reflect`, but filter on `TASKS.md` existence and include phase labels in the selector.
- [PLAN-04][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16][SPEC-17][SPEC-33] Build a deterministic markdown prompt that instructs the coding agent how to analyze the feature pipeline, compare plan vs implementation, de-duplicate against existing skills, write one canonical skill bundle, and duplicate it into the Claude mirror root.
- [PLAN-05][SPEC-21][SPEC-22][SPEC-23][SPEC-24][SPEC-25] Expand the prompt so it synthesizes across `PROJECT_PROGRESS_SUMMARY.md`, constitution alignment, and emergent workflows, with an explicit signal priority ladder for insight derivation.
- [PLAN-06][SPEC-26][SPEC-27][SPEC-28][SPEC-29][SPEC-30][SPEC-31][SPEC-34][SPEC-35] Add a mandatory stale-skill audit section that evaluates canonical skills under `<skills_dir>/*/SKILL.md`, retains passing canonical bundles unchanged, and removes the Claude mirror whenever a stale canonical skill is deleted.
- [PLAN-07][SPEC-18][SPEC-19][SPEC-20] Keep the existing command surface and workflow instructions intact while rerunning verification.

## COMPONENTS

- `internal/config/config.go`
  - add `SkillsDir`
  - add `SkillsPath`
- `pkg/cli/skill.go`
  - define `skill` and `skills`
  - define shared selector and command execution
- `pkg/cli/skill_prompt.go`
  - define prompt builder and prompt-section helpers
  - define canonical and Claude mirror path text
- `pkg/cli/root.go`
  - add command ordering entries
- `pkg/cli/skill_test.go`
  - cover prompt generation invariants, canonical-plus-mirror paths, insight signals, and audit requirements
- `README.md`
  - add Skill Mining command table section
- `.kit.yaml`
  - reflect the canonical skills root when explicitly configured

## DATA

- Input config: `.kit.yaml` gains `skills_dir`, defaulting to `.agents/skills`.
- Existing feature metadata continues to come from `internal/feature`.
- Project-wide insight input comes from `docs/PROJECT_PROGRESS_SUMMARY.md`.
- Canonical skill root is `.agents/skills` by default.
- Claude discovery mirror root is `.claude/skills`.
- Audit input path is `<skills_dir>/*/SKILL.md`, with Claude mirror cleanup as a side effect when deleting stale skills.
- No new persisted state beyond the config schema and feature docs.

## INTERFACES

- New CLI surfaces:
  - `kit skill mine [feature]`
  - `kit skills mine [feature]`
- Shared flags:
  - `--copy`
  - `--output-only`
- Prompt output remains plain markdown passed through `outputPrompt`.

## RISKS

- Two separate top-level commands can drift if they do not share the same subcommand instance or builder.
- Selector eligibility can diverge from the intended workflow if it filters on phase instead of `TASKS.md`.
- Prompt wording can become inconsistent with existing commands if output status handling does not reuse `outputPrompt` and `printWorkflowInstructions`.
- Help ordering can hide one alias if `commandOrder` is not updated for both names.
- Adding the insight-derivation and audit sections can push `pkg/cli/skill.go` past the repository file-size limit unless prompt-building logic is split into a helper file.
- Canonical and mirror roots can drift conceptually if the prompt does not state that `.agents/skills` is the source of truth and `.claude/skills` is only the mirrored discovery path.
- The prompt must not imply that one directory is auto-discovered by both agents; duplication is the compatibility mechanism.

## TESTING

- Add unit tests for `buildSkillMinePrompt`.
- Verify prompt content includes canonical skills directory, Claude mirror directory, directory-based `SKILL.md` paths, git diff instructions, de-duplication instructions, `PROJECT_PROGRESS_SUMMARY.md`, `CONSTITUTION.md`, signal-priority language, `SKILL AUDIT`, audit criteria, deletion instructions for both roots, and `Skill Audit Summary`.
- Verify prompt does not mention API calls or HTTP.
- Run:
  - `make vet`
  - `make test`
  - `make build`
  - `./bin/kit skill mine --help`
  - `./bin/kit skills mine --help`
  - `./bin/kit --help`
