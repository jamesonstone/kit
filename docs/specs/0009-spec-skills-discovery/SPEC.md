# SPEC

## SUMMARY

- Add a mandatory skills discovery phase to `kit spec` and persist the chosen skills in the current feature's `SPEC.md`.

## PROBLEM

- Kit captures reusable skills after implementation, but the specification workflow does not currently tell coding agents which existing skills to use for a feature.
- Feature-specific skill choices are not recorded in the feature spec, so later execution prompts cannot reliably point agents at the right skill files.

## GOALS

- Make `kit spec` instruct coding agents to perform a skills discovery phase.
- Add a required `## SKILLS` section to `SPEC.md`.
- Make later prompt-output commands tell agents to read the current feature's `SPEC.md` `## SKILLS` table and use those skills during execution.
- Keep skill discovery document-driven and prompt-driven.

## NON-GOALS

- Building a runtime plugin system.
- Resolving or executing skills automatically inside the Kit binary.
- Persisting per-feature skill selections in repository instruction files.

## USERS

- Developers using `kit spec` to define a new feature.
- Coding agents using Kit prompts to plan or execute feature work.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## REQUIREMENTS

- [SPEC-01] `SPEC.md` must gain a required `## SKILLS` section.
- [SPEC-02] The `## SKILLS` section must contain a mandatory table with columns `SKILL`, `SOURCE`, `PATH`, `TRIGGER`, and `REQUIRED`.
- [SPEC-03] The default no-skill row must be `none | n/a | n/a | no additional skills required | no`.
- [SPEC-04] `kit spec` prompts must instruct the coding agent to read repository instruction files first.
- [SPEC-05] `kit spec` prompts must instruct the coding agent to inspect repo-local canonical skills under `.agents/skills/*/SKILL.md`.
- [SPEC-06] `kit spec` prompts must instruct the coding agent to inspect the documented global inputs:
  - `~/.claude/CLAUDE.md`
  - `${CODEX_HOME}/AGENTS.md`
  - `${CODEX_HOME}/instructions.md`
  - `${CODEX_HOME}/skills/*/SKILL.md`
- [SPEC-07] `kit spec` prompts must instruct the coding agent to write the minimal relevant skill set into the current feature's `SPEC.md` `## SKILLS` table before sign-off.
- [SPEC-08] Prompt-output commands must include a standard instruction to consult documented skills and use them during execution.
- [SPEC-09] Repository instruction templates must describe the `kit spec` skills discovery phase and the `SPEC.md` `## SKILLS` table workflow.
- [SPEC-10] `.claude/skills` must remain a mirror path only and must not be described as canonical discovery input.

## ACCEPTANCE

- New `SPEC.md` files include the `## SKILLS` section and default row.
- `kit spec` prompts mention repo instruction files, `.agents/skills/*/SKILL.md`, and the documented global inputs.
- `kit spec` prompts explicitly require populating `SPEC.md` `## SKILLS` before completion.
- Prompt-output commands include a standard skill-use note that tells agents to read documented skills first.
- Repository instruction templates describe the new workflow.

## EDGE-CASES

- No additional skills apply to the feature.
- Repository instruction files are missing, but the prompt still needs to describe the workflow.
- Global Claude or Codex input files are absent on disk.
- A project has repo-local `.claude/skills` mirrors present from skill mining.

## OPEN-QUESTIONS

- none
