# TASKS

## TASKS

## PROGRESS TABLE

| ID   | TASK                                                 | STATUS | OWNER | DEPENDENCIES |
| ---- | ---------------------------------------------------- | ------ | ----- | ------------ |
| T001 | Record skill-mine feature docs                       | done   | agent |              |
| T002 | Implement config and CLI command                     | done   | agent | T001         |
| T003 | Update help and README surfaces                      | done   | agent | T002         |
| T004 | Add tests and run verification                       | done   | agent | T002, T003   |
| T005 | Amend docs for insight and audit                     | done   | agent | T004         |
| T006 | Extend prompt for insight synthesis                  | done   | agent | T005         |
| T007 | Extend tests and rerun verification                  | done   | agent | T006         |
| T008 | Amend docs for cross-agent skill paths               | done   | agent | T007         |
| T009 | Switch prompt to canonical plus mirror skill bundles | done   | agent | T008         |
| T010 | Expand tests and rerun verification for path changes | done   | agent | T009         |
| T011 | Switch `skill mine` to clipboard-first prompt output | done   | agent | T010         |
| T012 | Add `--prompt-only` consistency flag to `skill mine` | done   | agent | T011         |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record skill-mine feature docs [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05]
- [x] T002: Implement config and CLI command [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04]
- [x] T003: Update help and README surfaces [PLAN-05]
- [x] T004: Add tests and run verification [PLAN-05]
- [x] T005: Amend docs for insight and audit [PLAN-05] [PLAN-06] [PLAN-07]
- [x] T006: Extend prompt for insight synthesis [PLAN-05] [PLAN-06]
- [x] T007: Extend tests and rerun verification [PLAN-07]
- [x] T008: Amend docs for cross-agent skill paths [PLAN-01] [PLAN-04] [PLAN-06]
- [x] T009: Switch prompt to canonical plus mirror skill bundles [PLAN-01] [PLAN-04] [PLAN-06]
- [x] T010: Expand tests and rerun verification for path changes [PLAN-07]
- [x] T011: Switch `skill mine` to clipboard-first prompt output [PLAN-08]
- [x] T012: Add `--prompt-only` consistency flag to `skill mine` [PLAN-09]

## TASK DETAILS

### T001

- **GOAL**: Create the formal feature record before implementation begins
- **SCOPE**:
  - add `docs/specs/0006-skill-mine-command/`
  - write `SPEC.md`, `PLAN.md`, and `TASKS.md`
  - update `docs/PROJECT_PROGRESS_SUMMARY.md`
- **ACCEPTANCE**:
  - required artifact files exist with complete sections
  - project summary includes the new feature and current phase
- **NOTES**: complete when the repo reflects the new formal feature

### T002

- **GOAL**: Add the skill-mining CLI behavior and config support
- **SCOPE**:
  - update `internal/config/config.go`
  - create `pkg/cli/skill.go`
  - register `skill` and `skills` with shared `mine` behavior
  - implement selector filtering and prompt generation
- **ACCEPTANCE**:
  - both command forms output the same prompt behavior
  - selector eligibility is based on `TASKS.md`
  - prompt content satisfies the spec requirements
- **NOTES**: use `outputPrompt(prompt, outputOnly, skillCopy)` and `printWorkflowInstructions(...)`

### T003

- **GOAL**: Expose the new command in help ordering and docs
- **SCOPE**:
  - update `pkg/cli/root.go`
  - update `README.md`
- **ACCEPTANCE**:
  - root help includes both `skill` and `skills`
  - README has a Skill Mining section with both command entries
- **NOTES**: keep wording short and aligned with existing command tables

### T004

- **GOAL**: Prevent regression and validate command behavior
- **SCOPE**:
  - add `pkg/cli/skill_test.go`
  - run the required verification commands
- **ACCEPTANCE**:
  - automated tests cover prompt invariants
  - `make vet`, `make test`, and `make build` pass
  - manual help output confirms both command surfaces
- **NOTES**: confirm `kit skill mine` without args reaches the selector path

### T005

- **GOAL**: Record the follow-up amendment before changing implementation
- **SCOPE**:
  - update `SPEC.md`, `PLAN.md`, and `TASKS.md`
  - capture novel-insight and stale-skill-audit requirements explicitly
- **ACCEPTANCE**:
  - the formal docs describe insight derivation, signal priority, and skill audit behavior
  - new implementation work is traceable to updated spec and plan entries
- **NOTES**: complete before prompt-code edits

### T006

- **GOAL**: Extend the skill-mine prompt contract for insight synthesis and stale-skill auditing
- **SCOPE**:
  - update prompt context to include `PROJECT_PROGRESS_SUMMARY.md`
  - add novel-insight derivation instructions and signal priority rules
  - add mandatory `SKILL AUDIT` instructions and summary format
  - split prompt code into a helper file if needed to respect file-size limits
- **ACCEPTANCE**:
  - prompt instructs the agent to read `PROJECT_PROGRESS_SUMMARY.md` and `CONSTITUTION.md`
  - prompt includes spec-vs-implementation divergence analysis and signal priority order
  - prompt includes the four audit criteria and `rm -rf .claude/skills/<skill-name>/`
  - prompt includes a `Skill Audit Summary` format block
- **NOTES**: this remains prompt-only; Kit still does not execute deletions itself

### T007

- **GOAL**: Validate the amended prompt contract and keep the feature releasable
- **SCOPE**:
  - expand `pkg/cli/skill_test.go`
  - rerun vet, tests, build, and help checks
- **ACCEPTANCE**:
  - tests cover insight-derivation signals and audit criteria
  - `make vet`, `make test`, and `make build` pass
  - generated help still works for `kit skill mine` and `kit skills mine`
- **NOTES**: use the existing verification command set

### T008

- **GOAL**: Record the cross-agent path amendment before changing code
- **SCOPE**:
  - update `SPEC.md`, `PLAN.md`, and `TASKS.md`
  - redefine canonical and mirrored skill roots
- **ACCEPTANCE**:
  - formal docs state `.agents/skills` is canonical by default
  - formal docs state `.claude/skills` is the mirrored Claude discovery path
- **NOTES**: complete before prompt or config changes

### T009

- **GOAL**: Make mined skill instructions compatible with both Codex and Claude discovery models
- **SCOPE**:
  - change default canonical root to `.agents/skills`
  - update prompt paths to `<skills_dir>/<slug>/SKILL.md`
  - duplicate created skills into `.claude/skills/<slug>/`
  - audit canonical root and delete both canonical and mirror roots for stale skills
- **ACCEPTANCE**:
  - prompt identifies canonical and mirrored roots clearly
  - prompt no longer uses flat `<slug>.SKILL.md` paths
  - prompt deletes both roots when removing stale skills
- **NOTES**: keep `skills_dir` configurable as the canonical source of truth

### T010

- **GOAL**: Validate the cross-agent path amendment
- **SCOPE**:
  - expand `pkg/cli/skill_test.go`
  - rerun vet, tests, build, and prompt checks
- **ACCEPTANCE**:
  - tests assert canonical `.agents/skills/<slug>/SKILL.md` and Claude mirror paths
  - verification still passes cleanly
- **NOTES**: preserve the existing command surface

### T011

- **GOAL**: Align `skill mine` prompt output with the clipboard-first core workflow contract
- **SCOPE**:
  - update `pkg/cli/skill.go`
  - keep raw stdout prompt output behind `--output-only`
  - preserve `--copy` as an explicit override for `--output-only`
- **ACCEPTANCE**:
  - default command output acknowledges clipboard copy and does not print the prompt body
  - `--output-only` prints the raw prompt to stdout
  - `--output-only --copy` both prints and copies
  - verification still passes cleanly
- **NOTES**: reuse the shared clipboard-first helper instead of duplicating output logic

### T012

- **GOAL**: keep `skill mine` aligned with the shared feature-prompt command surface
- **SCOPE**:
  - register `--prompt-only` on `kit skill mine` and `kit skills mine`
  - preserve the existing prompt-only runtime behavior
  - cover the flag in docs/help verification
- **ACCEPTANCE**:
  - `kit skill mine --prompt-only` is accepted
  - the flag does not introduce any repo mutation path
  - help/docs describe the shared command-surface contract accurately
- **NOTES**: this is a consistency flag because the command already only emits prompts

## DEPENDENCIES

- T002 depends on T001 because implementation must follow the approved formal docs.
- T003 depends on T002 because help ordering and README must describe the final command surface.
- T004 depends on T002 and T003 because verification must validate the implemented surfaces.
- T006 depends on T005 because the amended prompt behavior must be specified first.
- T007 depends on T006 because tests and verification validate the amended prompt.
- T009 depends on T008 because the path change must be specified before implementation.
- T010 depends on T009 because tests and verification validate the canonical-plus-mirror change.
- T011 depends on T010 because the clipboard-first amendment must follow the existing prompt contract changes.
- T012 depends on T011 because the shared prompt-only flag is layered onto the existing prompt-output surface.

## NOTES

- The new command remains prompt-only and must not write `SKILL.md` files itself.
