# SPEC

## SUMMARY

- Add a new `kit skill mine [feature]` command, plus a `skills` alias, that outputs a prompt for an active coding agent to mine reusable procedural skills from a completed feature.
- The command must follow the same output-prompt contract as existing prompt-only commands and write nothing itself except the generated prompt.
- Mined skills must use a transferable directory bundle layout that can be consumed by multiple coding agent systems.

## PROBLEM

- Kit has no built-in command for turning completed feature work into reusable agent skills.
- Reusable implementation patterns currently remain trapped in feature docs and diffs instead of being promoted into a skills library.
- Teams need a deterministic prompt that tells an active agent how to compare planned work against implemented work and draft a `SKILL.md` only when the pattern is truly reusable.

## GOALS

- Add a top-level `skill` command with a `mine` subcommand.
- Add a top-level `skills` command as an alias surface with the same `mine` behavior.
- Reuse the existing prompt-output contract used by `kit reflect`, `kit implement`, and `kit brainstorm`.
- Let users target a feature directly or pick interactively from features that have reached at least `TASKS.md`.
- Include the feature's spec pipeline, project constraints, git diff instructions, and skill de-duplication instructions in the output prompt.
- Instruct the active coding agent to derive novel insights from the feature pipeline, git diff, and project-wide progression instead of limiting output to simple pattern extraction.
- Instruct the active coding agent to audit and remove stale skills that no longer match the current codebase or workflow reality.
- Make the canonical skills output directory configurable through `.kit.yaml` with a sensible default.
- Make the generated prompt place canonical skill bundles in an agent-neutral repo path and duplicate them into Claude's project-local discovery path.
- Document the new command in CLI help ordering and README command tables.

## NON-GOALS

- Calling any external API or HTTP service.
- Writing `SKILL.md` files directly from the Kit binary.
- Inventing a new output format or workflow distinct from the existing prompt-only commands.
- Mining multiple features in a single command invocation.

## USERS

- Developers who want to capture reusable implementation patterns after finishing a feature.
- Teams building a local skill library under source control.
- Coding agents that need a deterministic prompt for skill extraction work.

## REQUIREMENTS

- [SPEC-01] Add `skills_dir` to `.kit.yaml` config loading with default `.agents/skills`.
- [SPEC-02] Add `Config.SkillsPath(projectRoot string) string` to resolve the absolute skills directory path.
- [SPEC-03] Register two top-level Cobra commands: `skill` and `skills`.
- [SPEC-04] Both top-level commands must expose a `mine` subcommand and route to identical `RunE` behavior.
- [SPEC-05] The `mine` subcommand must accept zero or one positional feature argument.
- [SPEC-06] The `mine` subcommand must support `--copy` and `--output-only` flags with the same semantics as `kit reflect`.
- [SPEC-07] When no feature argument is given, interactive selection must list only features with `TASKS.md` present.
- [SPEC-08] The interactive selector must show each eligible feature with its current phase label.
- [SPEC-09] The output prompt must instruct the active coding agent to read `CONSTITUTION.md`.
- [SPEC-10] The output prompt must instruct the active coding agent to read the feature pipeline in order: optional `BRAINSTORM.md`, then `SPEC.md`, `PLAN.md`, and `TASKS.md`.
- [SPEC-11] The output prompt must instruct the active coding agent to run `git diff main`, falling back to `git diff master` if `main` does not exist.
- [SPEC-12] The output prompt must instruct the active coding agent to read existing `*.SKILL.md` files under the configured skills directory to avoid duplication.
- [SPEC-13] The output prompt must emphasize mining the delta between planned behavior and actual implementation as the highest-signal source of reusable patterns.
- [SPEC-14] The output prompt must instruct the active coding agent to write the canonical skill bundle to `<skills_dir>/<feature-slug>/SKILL.md`.
- [SPEC-15] The output prompt must define the required transferable skill-bundle structure and `SKILL.md` frontmatter and procedural body structure.
- [SPEC-16] The output prompt must explicitly state that the `description` frontmatter describes when the skill should trigger, not what it does.
- [SPEC-17] The output prompt must explicitly forbid writing anything when no genuinely reusable pattern is found.
- [SPEC-18] The command must print post-output workflow instructions for reviewing the generated draft and reusing the command later.
- [SPEC-19] Root help output must include both `skill` and `skills`.
- [SPEC-20] README command documentation must include both command forms under a Skill Mining heading.
- [SPEC-21] The output prompt must instruct the active coding agent to read `PROJECT_PROGRESS_SUMMARY.md` to detect recurring themes across multiple features.
- [SPEC-22] The output prompt must include an explicit novel-insight derivation block covering:
  - spec delta analysis
  - feature progression analysis
  - constitution alignment
  - emergent workflow insights
- [SPEC-23] The output prompt must explicitly state that spec-vs-implementation divergence is the highest-priority signal for reusable insight derivation.
- [SPEC-24] The output prompt rules must define a signal priority order covering:
  - spec-vs-implementation divergence
  - recurring themes across 2+ features
  - implicit workflows not yet documented
  - single-feature reusable patterns
  - constitution alignment gaps
- [SPEC-25] The output prompt rules must explicitly state that insights from signal priority 1 or 2 are always worth writing, while signal 4 or 5 requires a stronger reusability case.
- [SPEC-26] The output prompt must include a mandatory `SKILL AUDIT` phase that runs even when no new skill is written.
- [SPEC-27] The `SKILL AUDIT` phase must instruct the active coding agent to read every existing canonical skill at `<skills_dir>/*/SKILL.md`.
- [SPEC-28] The `SKILL AUDIT` phase must define four explicit audit criteria:
  - accuracy
  - relevance
  - coverage
  - trigger condition validity
- [SPEC-29] The `SKILL AUDIT` phase must instruct the active coding agent to log deletion reasons before deleting stale skills.
- [SPEC-30] The output prompt must define a `Skill Audit Summary` output format with created, removed, retained, and no-action sections.
- [SPEC-31] The output prompt rules must state that passing skills must not be modified, incomplete skills are not stale by themselves, and if nothing changes the agent should output `No skill changes - audit complete` and stop.
- [SPEC-32] The canonical skill bundle layout must be directory-based so it can be transferred across coding agent types:
  - `<skills_dir>/<skill-name>/SKILL.md`
  - optional bundled resources under the same directory
- [SPEC-33] The output prompt must instruct the active coding agent to duplicate every newly created canonical skill bundle into `.claude/skills/<skill-name>/` for Claude Code project discovery.
- [SPEC-34] The `SKILL AUDIT` phase must use `<skills_dir>` as the source of truth, not a hardcoded Claude-only root.
- [SPEC-35] When deleting a stale skill, the output prompt must instruct the active coding agent to remove both the canonical bundle directory and the Claude mirror directory:
  - `rm -rf <skills_dir>/<skill-name>/`
  - `rm -rf .claude/skills/<skill-name>/`
- [SPEC-36] The output prompt must identify `<skills_dir>` as the source of truth and `.claude/skills` as the Claude discovery mirror.

## ACCEPTANCE

- Running `kit skill mine <feature>` outputs a prompt that references the feature docs, configured canonical skills directory, Claude mirror path, project root, and deterministic draft output path.
- Running `kit skills mine <feature>` produces the same prompt behavior as `kit skill mine <feature>`.
- Running `kit skill mine` with no feature argument opens an interactive selector containing only features with `TASKS.md`, each annotated with a phase label.
- The prompt includes instructions for `git diff main`, fallback to `master`, skill de-duplication, reusable-pattern filtering, and the required `SKILL.md` format block.
- The prompt includes `PROJECT_PROGRESS_SUMMARY.md` as an input for cross-feature theme detection.
- The prompt includes a spec-vs-implementation divergence analysis step and an explicit signal priority order.
- The prompt includes a mandatory skill-audit section with audit criteria, deletion instructions, and a `Skill Audit Summary` format block.
- The prompt writes canonical skills to `<skills_dir>/<slug>/SKILL.md` and duplicates them to `.claude/skills/<slug>/SKILL.md`.
- The prompt audits canonical skills from `<skills_dir>/*/SKILL.md` and removes the Claude mirror when a canonical skill is deleted.
- The prompt starts with a direct task statement and contains no HTTP or API-call instructions.
- Root help and README both expose the new command surfaces.

## EDGE-CASES

- The feature has `TASKS.md` but no `BRAINSTORM.md`.
- The configured `skills_dir` contains no existing skill directories to audit.
- `.claude/skills/` contains no existing mirror directories yet.
- The local repository uses `master` rather than `main`.
- The git diff is empty or unavailable.
- All existing skills pass audit and no new skill is warranted.
- A skill is incomplete but still accurate and relevant, so it should be retained.
- A canonical skill exists but its Claude mirror directory is missing.
- The selected feature is in `tasks`, `implement`, `reflect`, or `complete` phase.

## OPEN-QUESTIONS

- None.
