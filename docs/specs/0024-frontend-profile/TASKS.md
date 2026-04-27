# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add root prompt profile flag and enum validation. [PLAN-01] | done | agent | |
| T002 | Add shared frontend prompt profile suffix rendering. [PLAN-02] [PLAN-06] | done | agent | T001 |
| T003 | Add feature profile inference and dependency-table helpers. [PLAN-03] [PLAN-04] | done | agent | T001 |
| T004 | Extend brainstorm notes scaffolding for frontend design materials. [PLAN-05] | done | agent | T003 |
| T005 | Wire feature-scoped prompt commands through profile-aware preparation. [PLAN-03] | done | agent | T002, T003, T004 |
| T006 | Wire generic and project prompt commands for explicit profile support. [PLAN-03] | done | agent | T002 |
| T007 | Add unit tests for prompt profile parsing, rendering, and inference. [PLAN-08] | done | agent | T001, T002, T003 |
| T008 | Add command tests for frontend scaffold, prompt-only, and prompt coverage behavior. [PLAN-08] | done | agent | T004, T005, T006, T007 |
| T009 | Run validation and refresh generated project progress artifacts. [PLAN-08] | done | agent | T008 |

## TASK LIST

- [x] T001: Add root prompt profile flag and enum validation. [PLAN-01]
- [x] T002: Add shared frontend prompt profile suffix rendering. [PLAN-02] [PLAN-06]
- [x] T003: Add feature profile inference and dependency-table helpers. [PLAN-03] [PLAN-04]
- [x] T004: Extend brainstorm notes scaffolding for frontend design materials. [PLAN-05]
- [x] T005: Wire feature-scoped prompt commands through profile-aware preparation. [PLAN-03]
- [x] T006: Wire generic and project prompt commands for explicit profile support. [PLAN-03]
- [x] T007: Add unit tests for prompt profile parsing, rendering, and inference. [PLAN-08]
- [x] T008: Add command tests for frontend scaffold, prompt-only, and prompt coverage behavior. [PLAN-08]
- [x] T009: Run validation and refresh generated project progress artifacts. [PLAN-08]

## TASK DETAILS

### T001
- **GOAL**: Add a project-wide prompt profile option that accepts `frontend` and rejects unsupported values before command side effects.
- **SCOPE**:
  - Add a root persistent `--profile` flag that is available to prompt-producing commands.
  - Model supported profile values as a closed enum with `""` and `frontend`.
  - Validate profile values before command mutation or prompt generation.
  - Do not add a `--frontend` alias or any new public command.
- **ACCEPTANCE**:
  - `--profile=frontend` is accepted by covered prompt-producing commands.
  - Unsupported profile values fail with a clear error before files or directories are created.
  - Existing invocations without `--profile` preserve current behavior.
  - No public command is added or removed.
- **NOTES**: Keep parsing and validation reusable so later prompt paths do not each implement their own profile handling.

### T002
- **GOAL**: Render the frontend profile as a shared, Kit-owned prompt suffix that preserves the RLM model.
- **SCOPE**:
  - Add one shared profile rendering path for prompt-output commands.
  - Append frontend guidance after skills and before subagent orchestration text.
  - Keep the guidance prompt-scoped, tool-agnostic, and free of runtime OpenAI documentation fetching.
  - Cover design-system inspection, domain/audience fit, usable UI expectations, familiar controls, responsive stability, assets, browser or screenshot validation, generated-UI anti-patterns, layout/text/palette checks, interaction states, and RLM-compatible context loading.
  - Prevent duplicate frontend suffixes when a command composes multiple prompt decorators.
- **ACCEPTANCE**:
  - Frontend guidance appears exactly once when the effective profile is `frontend`.
  - Frontend guidance includes concrete OpenAI-inspired quality constraints for domain fit, familiar controls, anti-pattern avoidance, responsive stability, and rendered UI inspection.
  - Frontend guidance does not appear when no effective profile is selected.
  - Prompt ordering is command prompt, skills, frontend profile, then subagent orchestration where present.
  - Guidance does not require vendor-specific tools or a specific agent runtime.
- **NOTES**: The text may cite OpenAI inspiration in feature artifacts, but generated runtime prompts should contain Kit-owned instructions.

### T003
- **GOAL**: Infer frontend profile state from active feature dependencies and maintain canonical dependency rows without corrupting existing content.
- **SCOPE**:
  - Add shared dependency-table helpers for `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`.
  - Add or refresh canonical `Frontend profile` and `Design materials` rows while preserving unrelated rows and stale history.
  - Remove placeholder `none` rows only when real dependencies are inserted.
  - Infer `frontend` only from an active `Frontend profile` dependency row in the current feature artifacts.
  - Ensure explicit `--profile` always overrides inferred feature state.
- **ACCEPTANCE**:
  - Active frontend dependency rows produce an effective frontend profile for feature-scoped prompt commands.
  - Optional, stale, missing, or malformed profile rows do not enable frontend profile behavior.
  - Dependency updates are idempotent and do not duplicate rows.
  - Existing dependency rows, including stale rows, remain intact unless the row being refreshed is canonical profile metadata.
- **NOTES**: Keep inference feature-path based; do not infer from prompt prose or global active-project state.

### T004
- **GOAL**: Create frontend design-materials note directories for brainstorm features without changing prompt-only behavior.
- **SCOPE**:
  - Extend non-prompt-only `kit brainstorm --profile=frontend` setup to create `docs/notes/<feature>/design/.gitkeep`.
  - Create `docs/notes/<feature>/design/screenshots/.gitkeep` and `docs/notes/<feature>/design/references/.gitkeep`.
  - Seed or refresh `Feature notes`, `Frontend profile`, and `Design materials` dependency rows in `BRAINSTORM.md`.
  - Teach brainstorm prompt text to ignore `.gitkeep` and placeholder files while allowing relevant design materials to inform the brainstorm.
  - Preserve existing append-only behavior for user-created note files.
- **ACCEPTANCE**:
  - Frontend brainstorm creates the design directory tree for the canonical numbered feature directory.
  - Prompt-only brainstorm output references design materials but performs no filesystem mutation.
  - Existing user note and design files are preserved.
  - `.gitkeep` files are not treated as meaningful note or design inputs.
- **NOTES**: Only brainstorm creates design-material directories; later workflow commands consume dependency metadata and explicit file references.

### T005
- **GOAL**: Apply frontend profile behavior consistently to feature-scoped prompt-producing commands.
- **SCOPE**:
  - Route `brainstorm`, backlog pickup, `spec`, `plan`, `tasks`, `implement`, `reflect`, `resume`, feature `handoff`, feature `summarize`, and `skill mine` through profile-aware prompt preparation.
  - Use explicit `--profile` when provided.
  - Infer the frontend profile from active feature dependency metadata when the command resolves a feature path.
  - Persist canonical profile dependencies when creating or refreshing `BRAINSTORM.md`, `SPEC.md`, or `PLAN.md` under the frontend profile.
  - Do not add profile metadata to `TASKS.md` because its dependency section has task-ordering semantics.
- **ACCEPTANCE**:
  - Each covered feature-scoped command emits frontend guidance when `--profile=frontend` is explicit.
  - Each covered feature-scoped command emits frontend guidance when the current feature has an active frontend profile dependency and no explicit profile overrides it.
  - Non-frontend feature workflows preserve existing prompt output.
  - Feature artifacts continue to satisfy their required section formats.
- **NOTES**: Keep command-specific prompt content intact; add profile behavior through shared preparation rather than command-local prompt rewrites where practical.

### T006
- **GOAL**: Support explicit frontend profiles for generic and project-scoped prompt commands without feature inference.
- **SCOPE**:
  - Route `dispatch`, project `handoff`, project `reconcile`, generic `summarize`, and `code-review` through the shared profile suffix path.
  - Enable frontend profile behavior only when `--profile=frontend` is explicitly passed to these generic or project-scoped commands.
  - Preserve dispatch behavior that suppresses subagent orchestration while still allowing frontend guidance.
  - Avoid adding side effects to read-only or prompt-only command paths.
- **ACCEPTANCE**:
  - Generic and project-scoped prompts can include frontend guidance through explicit `--profile=frontend`.
  - Generic and project-scoped prompts do not infer a profile from unrelated feature docs.
  - Dispatch prompt output remains free of subagent orchestration text.
  - Existing no-profile prompt output remains stable except for intentional shared-helper formatting.
- **NOTES**: This task should not add new workflow commands or new root instruction files.

### T007
- **GOAL**: Cover profile parsing, suffix composition, and feature inference with focused unit tests.
- **SCOPE**:
  - Test accepted and rejected profile values.
  - Test suffix presence, absence, ordering, and duplicate suppression.
  - Test feature inference for active, optional, stale, missing, malformed, and explicit-override cases.
  - Test dispatch-specific suffix composition without subagent orchestration.
- **ACCEPTANCE**:
  - Unit tests fail if unsupported profile values are accepted.
  - Unit tests fail if frontend guidance appears without an effective frontend profile.
  - Unit tests fail if frontend guidance is duplicated or placed after subagent orchestration text.
  - Unit tests fail if optional or stale dependency rows activate the frontend profile.
- **NOTES**: Keep tests near the shared profile, prompt-output, and dependency helper code they exercise.

### T008
- **GOAL**: Verify command-level frontend behavior for scaffolding, prompt-only safety, and prompt coverage.
- **SCOPE**:
  - Add or update command tests for `kit brainstorm --profile=frontend`.
  - Add tests for design directory creation, dependency row seeding, and `.gitkeep` handling.
  - Add tests proving prompt-only brainstorm emits instructions but does not create or modify files.
  - Add tests for `SPEC.md` and `PLAN.md` dependency persistence under frontend profile.
  - Add prompt-output coverage for representative feature-scoped, generic, project-scoped, and dispatch commands.
  - Add a mutation-order test proving invalid profile values fail before filesystem changes.
- **ACCEPTANCE**:
  - Tests demonstrate frontend brainstorm creates the expected notes and design tree only in non-prompt-only mode.
  - Tests demonstrate prompt-only mode is read-only.
  - Tests demonstrate covered prompt-producing commands include or omit frontend guidance according to effective profile state.
  - Tests demonstrate invalid profiles do not leave partial feature artifacts.
- **NOTES**: Prefer representative command tests plus shared-helper coverage over brittle golden snapshots for every command.

### T009
- **GOAL**: Validate the implementation and refresh project progress artifacts after the task work is complete.
- **SCOPE**:
  - Run `go test ./...`.
  - Run `make build`.
  - Run `git diff --check`.
  - Run `./bin/kit check frontend-profile`.
  - Run `./bin/kit map 0024-frontend-profile`.
  - Run `./bin/kit rollup` if generated project progress artifacts need updating.
  - Update `TASKS.md` statuses as implementation progresses.
- **ACCEPTANCE**:
  - Test and build results are recorded in the final implementation response.
  - Project checks pass or any failures are explained with exact causes.
  - `PROJECT_PROGRESS_SUMMARY.md` reflects the highest completed artifact for `0024-frontend-profile`.
  - No validation success is claimed unless the corresponding command ran.
- **NOTES**: If a validation command cannot run in the local environment, record the exact blocker and the residual risk.

## DEPENDENCIES

- T002 depends on T001 because profile rendering needs the validated effective profile value.
- T003 depends on T001 because dependency inference resolves to the same profile enum used by the root flag.
- T004 depends on T003 because brainstorm dependency rows should use the shared dependency-table helper.
- T005 depends on T002, T003, and T004 because feature commands need suffix rendering, profile inference, and brainstorm-specific setup behavior.
- T006 depends on T002 because generic and project commands only need explicit profile suffix rendering.
- T007 depends on T001, T002, and T003 because unit tests cover profile parsing, suffix composition, and inference helpers.
- T008 depends on T004, T005, T006, and T007 because command behavior tests should exercise the implemented scaffolding and prompt wiring after helper coverage exists.
- T009 depends on T008 because validation should run after implementation and tests are complete.

## NOTES

- SPEC and PLAN are fixed inputs for implementation; do not expand scope beyond `--profile=frontend` prompt behavior, frontend design-material notes, dependency metadata, and validation coverage.
- Do not add `--frontend`, new public commands, root instruction-file expansions, or an always-loaded frontend manual.
- Keep frontend profile instructions prompt-scoped and compatible with Kit's RLM just-in-time context model.
