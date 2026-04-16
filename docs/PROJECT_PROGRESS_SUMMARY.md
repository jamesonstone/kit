# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID | FEATURE | PATH | PHASE | PAUSED | CREATED | SUMMARY |
| -- | ------- | ---- | ----- | ------ | ------- | ------- |
| 0001 | refactor-plan-command | `docs/specs/0001-refactor-plan-command` | reflect | no | 2026-04-05 | - Refactor `kit plan [feature]` into a dedicated implementation-plan step that scaffolds `PLAN.md`, enforces the spec-first workflow, and keeps project state in sync. - Keep the command focused on plan creation and planning guidance, not task generation or implementation execution. |
| 0002 | cicd-goreleaser-releases | `docs/specs/0002-cicd-goreleaser-releases` | reflect | no | 2026-04-05 | - Add an automated release pipeline that versions Kit from semantic Git tags and publishes cross-platform GitHub release artifacts from the main branch workflow. |
| 0003 | inplace-upgrade-update | `docs/specs/0003-inplace-upgrade-update` | reflect | no | 2026-04-05 | - Add a canonical self-update command (`kit upgrade`) and keep `kit update` as a hidden deprecated compatibility entry point so users can move to the latest Kit release from GitHub Releases without manual install steps. - Updates must be safe and predictable: never leave users with a broken binary, and always provide clear outcome and recovery guidance. |
| 0004 | brainstorm-first-workflow | `docs/specs/0004-brainstorm-first-workflow` | reflect | no | 2026-04-05 | Adds an optional `BRAINSTORM.md` artifact and makes `kit brainstorm` the interactive, planning-only entrypoint before `kit spec`. Removes `kit oneshot` and all git branch automation while preserving the existing spec → plan → tasks → implement → reflect workflow. Core workflow prompt commands default to copying generated instructions to the clipboard, require `--output-only` for raw stdout prompt output, and accept `--prompt-only` to regenerate prompts for existing features without mutating workflow docs. |
| 0005 | version-command | `docs/specs/0005-version-command` | reflect | no | 2026-04-05 | - Add an explicit `kit version` subcommand that prints the installed Kit version from the same build metadata already used by `--version`. - The command must be stable, script-friendly, and visible in CLI help so users can inspect their installed release version directly. |
| 0006 | skill-mine-command | `docs/specs/0006-skill-mine-command` | reflect | no | 2026-04-05 | - Add a new canonical `kit skill mine [feature]` command, plus a deprecated hidden `skills` compatibility alias, that outputs a prompt for an active coding agent to mine reusable procedural skills from a completed feature. - The command must follow the same clipboard-first output-prompt contract as `kit implement` and `kit reflect` and write nothing itself except the generated prompt. - The command must also accept `--prompt-only` as a consistency flag for regenerating the selected feature prompt without mutating repo docs. - Mined skills must use a transferable directory bundle layout that can be consumed by multiple coding agent systems. |
| 0007 | catchup-command | `docs/specs/0007-catchup-command` | reflect | no | 2026-04-05 | - Add a new `kit catchup [feature]` command that outputs a prompt for a coding agent to recover the current state of a selected feature before any implementation resumes. After command-surface simplification, `catchup` remains callable as a hidden deprecated compatibility surface while `kit resume` becomes the canonical general resume command. - The command must be prompt-only, feature-scoped, and explicitly keep the agent in plan mode until the user approves moving into implementation. - Default command output must copy the generated prompt to the clipboard and reserve raw stdout prompt output for `--output-only`. |
| 0008 | dispatch-command | `docs/specs/0008-dispatch-command` | reflect | no | 2026-04-05 | - Add a new `kit dispatch` command that outputs a prompt for a coding agent to discover likely file overlap across a pasted task set, cluster overlapping work, and queue subagents safely. - The command must be prompt-only and must force a discovery-and-approval step before any subagent execution begins. - Default command output must copy the generated prompt to the clipboard and reserve raw stdout prompt output for `--output-only`. |
| 0009 | spec-skills-discovery | `docs/specs/0009-spec-skills-discovery` | reflect | no | 2026-04-05 | - Add a mandatory skills discovery phase to `kit spec`, keep the chosen skills in `SPEC.md`, and separately track the broader supporting dependencies that shaped the spec. |
| 0010 | support-command-clipboard-defaults | `docs/specs/0010-support-command-clipboard-defaults` | reflect | no | 2026-04-05 | - Change `kit handoff`, `kit summarize`, and `kit code-review` to copy generated output to the clipboard by default. - Reserve raw stdout prompt output for explicit `--output-only` usage while keeping `--copy` as an explicit override for `--output-only`. |
| 0011 | handoff-document-sync | `docs/specs/0011-handoff-document-sync` | reflect | no | 2026-04-05 | - Change `kit handoff` from a passive “new session context dump” into an active prompt for the current coding agent session to reconcile feature docs with implementation reality before transfer. - Require the generated prompt to produce a concise final response that confirms documentation sync, includes a full-path document table, and summarizes the most recent conversation context. |
| 0012 | default-subagent-orchestration | `docs/specs/0012-default-subagent-orchestration` | reflect | no | 2026-04-05 | - Change Kit's shared prompt-orchestration default from single-agent to subagent-first. - Add `--single-agent` as the explicit opt-out when a user wants one lane only. - Keep `kit dispatch` as the stricter discovery-first queue-planning command with explicit approval before launch. - Clarify that repository-scale RLM discovery narrows context first, while dispatch and subagents handle execution planning after discovery. |
| 0012 | implement-readiness-gate | `docs/specs/0012-implement-readiness-gate` | reflect | no | 2026-04-05 | - Add an implementation-readiness gate inside `kit implement` that requires an adversarial preflight over the feature docs before coding begins. - Keep the workflow as a gate, not a new lifecycle phase, command, or artifact type. |
| 0013 | scaffold-agents-safe-merge | `docs/specs/0013-scaffold-agents-safe-merge` | reflect | no | 2026-04-05 | - Make `kit scaffold-agents` safer when repository instruction files already exist by adding an overwrite confirmation gate and an explicit append-only mode. - Keep append-only deterministic and constrained to known Kit-managed sections; do not attempt fuzzy free-form merges. |
| 0014 | human-readable-terminal-output | `docs/specs/0014-human-readable-terminal-output` | reflect | no | 2026-04-05 | - Improve human-readable terminal output with clearer spacing, semantic emoji markers, and more readable help sections. - Keep generated coding-agent prompts, scaffolded agent instruction files, `--output-only` raw stdout, and `--json` output unchanged. |
| 0015 | pause-remove-commands | `docs/specs/0015-pause-remove-commands` | complete | no | 2026-04-05 | Add `kit pause` and `kit remove` so users can explicitly pause in-flight work or remove a feature cleanly while keeping Kit's generated progress views and selectors consistent. |
| 0016 | document-map-relationships | `docs/specs/0016-document-map-relationships` | complete | no | 2026-04-05 | Add a read-only `kit map` command that renders the canonical document graph and current project state, and add explicit `RELATIONSHIPS` sections to `BRAINSTORM.md` and `SPEC.md` for feature-to-feature lineage. |
| 0017 | reconcile-command | `docs/specs/0017-reconcile-command` | reflect | no | 2026-04-05 | - Add a new `kit reconcile [feature]` command that audits Kit-managed project documents against the current Kit document contract and outputs a prompt for an agent to reconcile stale or missing documentation. - The command must default to whole-project reconciliation, stay prompt-only in v1, and emit exact file targets, update instructions, and codebase search guidance instead of editing docs directly. |
| 0018 | backlog-command | `docs/specs/0018-backlog-command` | reflect | no | 2026-04-05 | Add a first-class `kit backlog` command and matching `kit brainstorm` backlog flags so users can capture out-of-scope follow-up features, list them later, and pick them back up without introducing a second document format. The command remains the backlog-specific surface after command-surface simplification, while `kit resume` becomes the canonical general resume entry point. |
| 0019 | command-surface-simplification | `docs/specs/0019-command-surface-simplification` | reflect | no | 2026-04-05 | Simplify the top-level Kit command surface by introducing a canonical `resume` flow, adding `status --all` as the project overview mode, and deprecating overlapping or duplicate command entry points while keeping them callable for compatibility. |
| 0020 | versioned-instruction-model | `docs/specs/0020-versioned-instruction-model` | reflect | no | 2026-04-12 | - Default new Kit repos to a thin table-of-contents instruction model that points agents into a repo-local docs tree, while preserving the current model for existing repos unless `--version` explicitly switches them. |
| 0021 | project-validation-and-instruction-registry | `docs/specs/0021-project-validation-and-instruction-registry` | reflect | no | 2026-04-12 | - Centralize Kit's instruction-model metadata in one internal registry and add a project-scoped validation mode so repo-level contract drift is checked mechanically instead of being spread across prompt builders. |
| 0022 | typed-prompt-ir | `docs/specs/0022-typed-prompt-ir` | reflect | no | 2026-04-13 | - Replace ad hoc string-built prompt construction with a typed prompt IR across Kit's prompt-producing commands, while keeping the current output wrappers and shared prompt decorators intact. |
| 0023 | worktree-safe-feature-allocation | `docs/specs/0023-worktree-safe-feature-allocation` | reflect | no | 2026-04-16 | Reserve feature numbers from a repo-shared allocator so multiple worktrees from the same clone cannot create duplicate numbered feature directories. |

## PROJECT INTENT

<!-- TODO: describe the overall project purpose -->

## GLOBAL CONSTRAINTS

See `docs/CONSTITUTION.md` for project-wide constraints and principles.

## FEATURE SUMMARIES

### refactor-plan-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Implementation strategy needs its own explicit artifact between requirements and tasks. - Without a dedicated `plan` command, strategy details drift into `SPEC.md` or straight into code, which weakens traceability and makes task generation less deterministic. - The workflow also needs one place to enforce `SPEC.md` as the prerequisite for planning and to keep `PROJECT_PROGRESS_SUMMARY.md` aligned with the highest completed artifact.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Keep `pkg/cli/plan.go` as the single entry point for prerequisite enforcement and `PLAN.md` creation so the workflow stays explicit. - [PLAN-02][SPEC-04] Reuse feature discovery and selection helpers to target features that are ready for planning. - [PLAN-03][SPEC-05][SPEC-06] Use the embedded plan template to enforce the required plan sections and dependency inventory. - [PLAN-04][SPEC-07] Regenerate `PROJECT_PROGRESS_SUMMARY.md` after plan creation so project state reflects the highest completed artifact. - [PLAN-05][SPEC-08] Emit a planning prompt that reads the constitution and spec inputs first, then keeps the output focused on implementation strategy rather than tasks or code.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0001-refactor-plan-command/SPEC.md`, `docs/specs/0001-refactor-plan-command/PLAN.md`, `docs/specs/0001-refactor-plan-command/TASKS.md`

### cicd-goreleaser-releases

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit lacks an automated release pipeline for cross-platform binary distribution. - Releases are not consistently versioned and published from `main`.
- **APPROACH**: - Add a `main` branch workflow that computes next `vMAJOR.MINOR.PATCH` tag from existing semantic tags and pushes it. - In the same `main` workflow, run vet/test, build artifacts with GoReleaser, and publish GitHub release with generated notes for that tag. - Keep a tag workflow (`v*`) for manual semantic tag release publication. - Use Git tags as semantic version source of truth.
- **OPEN ITEMS**: - None.
- **POINTERS**: `docs/specs/0002-cicd-goreleaser-releases/SPEC.md`, `docs/specs/0002-cicd-goreleaser-releases/PLAN.md`, `docs/specs/0002-cicd-goreleaser-releases/TASKS.md`

### inplace-upgrade-update

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit users currently need manual update flows (for example, reinstalling with Go tooling), which is slower and inconsistent. - There is no built-in way to check installed version versus latest GitHub release and apply an in-place upgrade. - Friction in the update process increases version drift and delays adoption of fixes and improvements.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Add a new `pkg/cli/upgrade.go` file that registers canonical `upgrade` and hidden deprecated `update` commands sharing the same behavior and `--yes` confirmation bypass flag. - [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Resolve the current version with `currentVersion()`, fetch the latest stable release from the GitHub Releases API, skip prereleases, and compare versions with stdlib-only semver parsing. - [PLAN-03][SPEC-08][SPEC-09][SPEC-10] Derive the expected release artifact from `.goreleaser.yaml` naming rules, download the artifact plus `checksums.txt`, and verify SHA-256 before any filesystem write. - [PLAN-04][SPEC-11][SPEC-12][SPEC-13] Extract the executable from `tar.gz` or `zip` in memory and replace the installed binary using a same-directory temp file plus atomic rename semantics, with Windows-specific best-effort fallback. - [PLAN-05][SPEC-14][SPEC-15][SPEC-16][SPEC-17] Keep user-facing output exact and actionable for already-current, confirmation, success, timeout, rate-limit, missing-asset, checksum, and permission-failure paths. - [PLAN-06][SPEC-18][SPEC-19][SPEC-20] Add command ordering, README utility docs, and focused unit tests for asset naming, checksum parsing, version comparison, and executable path resolution.
- **OPEN ITEMS**: - Should prereleases be ignored by default, with no opt-in in this phase? - For `dev` builds, should the command refuse update or allow update to latest stable? - Is Windows support required in the first release, or can v1 scope be macOS/Linux only? - Is a dry-run mode (for example `--check`) required in v1? - Should fallback guidance explicitly include `go install github.com/jamesonstone/kit/cmd/kit@latest` when in-place replacement is not possible? - Should update behavior be blocked for package-manager detected installs, or attempted with warning? - Is checksum verification against `checksums.txt` sufficient for v1 without signature verification? - What exit code policy should be used for "already up to date" versus "update applied"? - Should the command support pinning a target version (for example `--version vX.Y.Z`) in v1? - What maximum acceptable runtime should be targeted for update checks on normal network conditions?
- **POINTERS**: `docs/specs/0003-inplace-upgrade-update/SPEC.md`, `docs/specs/0003-inplace-upgrade-update/PLAN.md`, `docs/specs/0003-inplace-upgrade-update/TASKS.md`

### brainstorm-first-workflow

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: Kit currently treats brainstorming as an external or standalone activity, while the formal workflow starts at `SPEC.md`. That creates a gap between early research and canonical project documents, duplicates functionality with `kit oneshot`, and couples feature creation to branch automation that is outside Kit's core document-centered purpose.
- **APPROACH**: 1. formalize the workflow contract in repo docs and generated templates 2. add `BRAINSTORM.md` support and a dedicated brainstorm phase in feature/status/rollup logic 3. refactor `kit brainstorm` into the interactive, planning-only feature entrypoint 4. thread `BRAINSTORM.md` through downstream prompts as optional upstream context and phase dependency source 5. keep prompt output behavior command-scoped by adding a clipboard-first helper for the core workflow commands without changing support utilities 6. add a shared `--prompt-only` flag for feature-scoped prompt commands and branch artifact-writing commands into side-effect-free regeneration mode 7. make supported multiline free-text prompts editor-default and add an explicit `--inline` opt-out where inline entry already exists 8. remove `kit oneshot` and git branch automation from code, config, help, and docs 9. add tests for prompt generation, clipboard-first output, prompt-only regeneration, editor-default free-text flows, and phase detection, then run full verification
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0004-brainstorm-first-workflow/SPEC.md`, `docs/specs/0004-brainstorm-first-workflow/PLAN.md`, `docs/specs/0004-brainstorm-first-workflow/TASKS.md`

### version-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit currently exposes version information only through the root `--version` flag. - Users and scripts do not have a first-class `kit version` command path. - Existing update-related documentation already refers to `kit version`, creating a product gap.
- **APPROACH**: - Implement a no-arg Cobra command with `RunE` that writes the resolved version to `cmd.OutOrStdout()`. - Keep formatting intentionally minimal so the command is script-friendly. - Prefer existing release/build injection behavior and fall back to Go build info for `go install` builds.
- **OPEN ITEMS**: - None.
- **POINTERS**: `docs/specs/0005-version-command/SPEC.md`, `docs/specs/0005-version-command/PLAN.md`, `docs/specs/0005-version-command/TASKS.md`

### skill-mine-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit has no built-in command for turning completed feature work into reusable agent skills. - Reusable implementation patterns currently remain trapped in feature docs and diffs instead of being promoted into a skills library. - Teams need a deterministic prompt that tells an active agent how to compare planned work against implemented work and draft a `SKILL.md` only when the pattern is truly reusable.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-32][SPEC-36] Extend config with a configurable canonical skills directory and keep the prompt explicit about the canonical root versus the Claude mirror root. - [PLAN-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Implement `pkg/cli/skill.go` so `skill` is canonical and `skills` remains a hidden deprecated compatibility root sharing the same `mine` behavior. - [PLAN-03][SPEC-07][SPEC-08] Reuse the feature-list pattern from `implement`/`reflect`, but filter on `TASKS.md` existence and include phase labels in the selector. - [PLAN-04][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16][SPEC-17][SPEC-33] Build a deterministic markdown prompt that instructs the coding agent how to analyze the feature pipeline, compare plan vs implementation, de-duplicate against existing skills, write one canonical skill bundle, and duplicate it into the Claude mirror root. - [PLAN-05][SPEC-21][SPEC-22][SPEC-23][SPEC-24][SPEC-25] Expand the prompt so it synthesizes across `PROJECT_PROGRESS_SUMMARY.md`, constitution alignment, and emergent workflows, with an explicit signal priority ladder for insight derivation. - [PLAN-06][SPEC-26][SPEC-27][SPEC-28][SPEC-29][SPEC-30][SPEC-31][SPEC-34][SPEC-35] Add a mandatory stale-skill audit section that evaluates canonical skills under `<skills_dir>/*/SKILL.md`, retains passing canonical bundles unchanged, and switches stale-skill cleanup to an approval-gated flow instead of immediate destructive deletion guidance. - [PLAN-07][SPEC-18][SPEC-19][SPEC-20] Keep the existing command surface and workflow instructions intact while rerunning verification. - [PLAN-08][SPEC-06][SPEC-06a][SPEC-06b] Switch `skill mine` output to the shared clipboard-first helper while keeping `--output-only` and `--copy` behavior explicit. - [PLAN-09][SPEC-06c] Register the shared `--prompt-only` flag on `skill mine` so the command surface stays consistent with the rest of Kit's feature-scoped prompt commands without changing runtime behavior.
- **OPEN ITEMS**: - None.
- **POINTERS**: `docs/specs/0006-skill-mine-command/SPEC.md`, `docs/specs/0006-skill-mine-command/PLAN.md`, `docs/specs/0006-skill-mine-command/TASKS.md`

### catchup-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit has `status`, `handoff`, `summarize`, and `implement`, but no feature-scoped command dedicated to helping a coding agent catch up on a feature's current stage and state without drifting into execution too early. - Users who return to an in-flight feature often need a lightweight “resume and orient” prompt rather than a full handoff or immediate implementation context. - Without an explicit catch-up step, agents can skip clarification, miss recent state encoded in repo artifacts, or start implementing before the user confirms the next move.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04] Create a new `pkg/cli/catchup.go` command with optional feature argument, selector fallback, and standard `--copy` / `--output-only` behavior. - [PLAN-02][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09] Resolve the selected feature with existing feature/status helpers and derive stage plus state from `feature.GetFeatureStatus(...)` and current next-action guidance. - [PLAN-03][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16] Build a feature-scoped `/plan` prompt that tells the coding agent how to catch up on the selected feature, ask questions first, stay in plan mode, and request explicit approval before implementation. - [PLAN-04][SPEC-17] Add complete-phase-specific prompt wording so completed features are treated as review/reopen triage rather than resumed implementation. - [PLAN-05][SPEC-18] Register the command in help ordering and README with wording that clearly distinguishes it from `handoff`, `summarize`, and `implement`, then move it to hidden deprecated compatibility status when `resume` becomes canonical. - [PLAN-06][SPEC-19] Add focused tests for prompt generation and state rendering, then run the normal verification commands. - [PLAN-07] Switch `catchup` to the shared clipboard-first helper while keeping `--output-only` and `--copy` behavior explicit. - [PLAN-08] Register the shared `--prompt-only` flag on `catchup` so the command surface matches the rest of Kit's feature-scoped prompt commands.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0007-catchup-command/SPEC.md`, `docs/specs/0007-catchup-command/PLAN.md`, `docs/specs/0007-catchup-command/TASKS.md`

### dispatch-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit has prompt generators for planning, catch-up, implementation, reflection, and skill mining, but no prompt-only command specialized for turning a raw task list into a safe subagent dispatch plan. - When users hand a coding agent a mixed set of bullets, numbered items, and paragraphs, the agent can parallelize too aggressively and create conflicting edits across the same files. - Users need a deterministic prompt that tells the coding agent to discover first, predict touched files, merge ambiguous overlap conservatively, and wait for approval before launching subagents.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-16][SPEC-27] Add a new `pkg/cli/dispatch.go` command with standard prompt flags, input-source precedence, default editor-backed capture, pre-editor instructions, any-key launch gating, file/stdin support, and max-subagent validation. - [PLAN-02][SPEC-14][SPEC-15] Add focused task-normalization helpers that split only top-level paragraphs, bullets, and numbered items into dispatchable units while preserving nested detail under the parent task. - [PLAN-03][SPEC-17][SPEC-18][SPEC-19][SPEC-20][SPEC-21][SPEC-22][SPEC-23][SPEC-24][SPEC-25][SPEC-26] Build a dedicated prompt builder that embeds the normalized task list and enforces discovery-first clustering, conservative overlap handling, dry-run reporting, and approval gating before subagent launch. - [PLAN-04][SPEC-28] Register the new command in help ordering and README so the public CLI surface matches the shipped behavior. - [PLAN-05] Add focused tests for input-source precedence, task normalization, prompt invariants, and flag validation, then run the standard verification commands. - [PLAN-06] Switch `dispatch` to the shared clipboard-first helper that preserves dispatch's no-subagent-suffix prompt shape.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0008-dispatch-command/SPEC.md`, `docs/specs/0008-dispatch-command/PLAN.md`, `docs/specs/0008-dispatch-command/TASKS.md`

### spec-skills-discovery

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit captures reusable skills after implementation, but the specification workflow does not currently tell coding agents which existing skills to use for a feature. - Feature-specific skill choices are not recorded in the feature spec, so later execution prompts cannot reliably point agents at the right skill files.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Update document validation and templates so `SPEC.md` requires a `## SKILLS` section with the fixed table shape and default row. - [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Update `kit spec` prompt generation to treat skills discovery as a first-class phase and to name repo-local and documented global inputs explicitly. - [PLAN-02A][SPEC-08][SPEC-09][SPEC-10][SPEC-11] Keep `SPEC.md` dependency inventories separate from `## SKILLS` and require exact locations for design dependencies. - [PLAN-03][SPEC-12] Add a shared prompt suffix that tells coding agents to consult documented skills before execution, so every prompt-output command stays aligned. - [PLAN-04][SPEC-13][SPEC-14] Update repository instruction templates and checked-in instruction files to describe the new workflow and to keep `.claude/skills` mirror-only.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0009-spec-skills-discovery/SPEC.md`, `docs/specs/0009-spec-skills-discovery/PLAN.md`, `docs/specs/0009-spec-skills-discovery/TASKS.md`

### support-command-clipboard-defaults

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - These three support commands still print their full prompt or output body by default while the rest of Kit's prompt-producing surfaces are now clipboard-first. - The inconsistent default makes the CLI harder to predict and forces users to remember different copy/paste flows for adjacent commands. - Users explicitly requested that the remaining commands follow the same default output contract.
- **APPROACH**: - [PLAN-01] Update the formal docs for the three commands before changing code. - [PLAN-02] Route all three commands through the shared clipboard-first helper. - [PLAN-03] Update README and help strings to reflect the new default output contract. - [PLAN-04] Reuse existing clipboard-first helper tests and rerun repository verification.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0010-support-command-clipboard-defaults/SPEC.md`, `docs/specs/0010-support-command-clipboard-defaults/PLAN.md`, `docs/specs/0010-support-command-clipboard-defaults/TASKS.md`

### handoff-document-sync

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - `kit handoff` currently focuses on orienting a fresh agent session, but it does not explicitly require the current session to update feature docs so they match the actual implementation before handoff. - That leaves too much room for stale `SPEC.md`, `PLAN.md`, `TASKS.md`, and rollup data to survive into the next session. - The current prompt also does not enforce a standard final response that confirms doc sync, lists the authoritative file paths, and captures recent conversation context in a reusable way.
- **APPROACH**: - [PLAN-01] Record the handoff prompt contract before changing code. - [PLAN-02] Add prompt-building helpers that generate document inventory tables with absolute paths and concise usage guidance. - [PLAN-03] Rewrite feature-scoped handoff prompts to require documentation reconciliation before handoff. - [PLAN-04] Rewrite project-wide handoff prompts to reconcile rollup and active feature docs before handoff. - [PLAN-05] Define a final response contract that requires documentation-sync confirmation, dependency-inventory verification, a document table, and a recent-context summary. - [PLAN-06] Add tests for feature and project-wide handoff prompt content, then rerun full verification. - [PLAN-07] Register the shared `--prompt-only` flag on `handoff` so the command surface matches the rest of Kit's feature-scoped prompt commands.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0011-handoff-document-sync/SPEC.md`, `docs/specs/0011-handoff-document-sync/PLAN.md`, `docs/specs/0011-handoff-document-sync/TASKS.md`

### default-subagent-orchestration

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit currently assumes a single-agent lane unless the user explicitly adds `--subagents`. - That default underuses subagents on work that naturally splits across multiple distinct areas in both research and implementation. - Users now want subagent orchestration to be the standard path, while still preserving conservative overlap handling and a way to force one-lane execution when needed.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-10] Update the shared prompt helper to make subagent guidance default-on, add `--single-agent`, and keep a hidden compatibility alias for `--subagents`. - [PLAN-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Rewrite the shared orchestration suffix so it defaults to subagents while preserving conservative overlap handling and main-agent ownership. - [PLAN-03][SPEC-08] Verify that `dispatch` still uses the dedicated no-shared-subagent path. - [PLAN-04][SPEC-09] Update README and help-facing wording to reflect the new default and opt-out flag. - [PLAN-05] Add focused tests for default suffix behavior, `--single-agent`, legacy flag registration, and dispatch isolation, then rerun verification. - [PLAN-06][SPEC-11] Tighten shipped orchestration wording so RLM discovery and dispatch-style execution planning remain distinct.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0012-default-subagent-orchestration/SPEC.md`, `docs/specs/0012-default-subagent-orchestration/PLAN.md`, `docs/specs/0012-default-subagent-orchestration/TASKS.md`

### implement-readiness-gate

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - `kit implement` currently moves directly from document reading to task execution. - That leaves no explicit semantic challenge step to catch contradictions, ambiguity, weak task coverage, or missing failure cases before code is written. - The existing `kit check` command validates document structure, but it does not serve as a pre-implementation adversarial review.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Record the approved readiness-gate contract in feature docs before changing code. - [PLAN-02][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11] Rewrite the `kit implement` prompt so it begins with an implementation-readiness gate and adversarial preflight instructions. - [PLAN-03][SPEC-12][SPEC-13] Update adjacent workflow wording in `kit status`, `README.md`, `docs/CONSTITUTION.md`, `docs/specs/0000_INIT_PROJECT.md`, and scaffolded instruction templates without introducing a new phase. - [PLAN-04][SPEC-14] Leave `kit check` unchanged for v1. - [PLAN-05][SPEC-15][SPEC-16] Add focused tests for the implement prompt and status wording, then run verification.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0012-implement-readiness-gate/SPEC.md`, `docs/specs/0012-implement-readiness-gate/PLAN.md`, `docs/specs/0012-implement-readiness-gate/TASKS.md`

### scaffold-agents-safe-merge

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - `kit scaffold-agents` currently has binary behavior: - skip existing files by default - overwrite existing files blindly with `--force` - That makes it easy to destroy customized `AGENTS.md`, `CLAUDE.md`, or `.github/copilot-instructions.md` content by accident. - The command also has no supported middle path for preserving custom content while adding newly introduced Kit-managed sections.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Record the scaffold safety contract in a dedicated feature spec before changing code. - [PLAN-02][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14] Extend scaffold write handling to support explicit write modes, overwrite confirmation, and append-only preflight planning. - [PLAN-03][SPEC-15][SPEC-16] Add a deterministic instruction-file section merge helper for append-only mode. - [PLAN-04][SPEC-17][SPEC-18] Update command wiring, help text, and post-run guidance for the new flags and safer suggestions. - [PLAN-05][SPEC-19][SPEC-20] Add tests for overwrite confirmation, append-only success/failure behavior, targeted selection, and flag validation. - [PLAN-06][SPEC-21] Update shipped docs for the new scaffold-agents semantics and rerun verification.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0013-scaffold-agents-safe-merge/SPEC.md`, `docs/specs/0013-scaffold-agents-safe-merge/PLAN.md`, `docs/specs/0013-scaffold-agents-safe-merge/TASKS.md`

### human-readable-terminal-output

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit mixes plain text, sparse emoji usage, dense section layouts, and default Cobra help formatting across its human-facing CLI surfaces. - The inconsistent presentation makes interactive flows slower to scan, especially in selectors, workflow guidance, status output, and help text. - Terminal applications cannot reliably change font size, so readability improvements must come from spacing, grouping, and consistent visual cues.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Record the approved scope and exclusions in a dedicated feature spec before code changes. - [PLAN-02][SPEC-06][SPEC-07][SPEC-08][SPEC-09][SPEC-10] Add shared terminal-formatting helpers for human-readable acknowledgements, headings, selector prompts, and TTY detection. - [PLAN-03][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Apply the shared formatting to grouped root help, workflow guidance, editor-launch instructions, selector screens, and related command follow-up output. - [PLAN-04][SPEC-01][SPEC-05] Improve human-readable status presentation without changing status content, and keep fleet views in fixed-width terminal columns instead of Markdown-style tables, with ANSI color gated on TTY output only. - [PLAN-05][SPEC-06][SPEC-07][SPEC-08][SPEC-09] Add or update tests to prove raw-output stability and new human-readable formatting behavior. - [PLAN-06][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05] Update practical docs and rollup state to match the shipped UX.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0014-human-readable-terminal-output/SPEC.md`, `docs/specs/0014-human-readable-terminal-output/PLAN.md`, `docs/specs/0014-human-readable-terminal-output/TASKS.md`

### pause-remove-commands

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Kit currently has no lifecycle controls for work that should stop without being completed or for feature directories that should be removed entirely. That forces users to manage state by hand across `docs/specs/`, `.kit.yaml`, `PROJECT_PROGRESS_SUMMARY.md`, and `kit status`, which is error-prone and can leave stale active-feature views behind.
- **APPROACH**: - extend config with a small per-feature state map keyed by feature directory name - centralize pause lookups and pause mutations in `internal/feature` - add a shared auto-unpause helper for explicit feature-scoped workflow commands - implement `kit pause` as a non-destructive lifecycle toggle for non-complete features - implement `kit remove` as a destructive lifecycle command with confirmation, directory deletion, state cleanup, and rollup regeneration - update rollup and status rendering to surface paused state separately from phase - keep default `status` active-feature focused and move the fleet view into the explicit `status --all` mode - exclude paused features from active-only multi-feature flows except `status` and `status --all`
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0015-pause-remove-commands/SPEC.md`, `docs/specs/0015-pause-remove-commands/PLAN.md`, `docs/specs/0015-pause-remove-commands/TASKS.md`

### document-map-relationships

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Kit documents have a clear internal hierarchy, but that structure is currently split across static help text, repository docs, and individual feature files. Users and coding agents can read the pieces, but there is no single dynamic surface that shows how global docs, feature docs, lifecycle phases, and cross-feature dependencies fit together in the current repository state.
- **APPROACH**: - record the command and document-contract changes in canonical docs first - extend brainstorm and spec templates plus document validation with a required `RELATIONSHIPS` section seeded with `none` - backfill existing brainstorm and spec docs with explicit `RELATIONSHIPS` sections so the new validation rule does not strand the repo in an invalid intermediate state - add a small relationship parser that reads explicit bullets from `BRAINSTORM.md` and `SPEC.md` - normalize harmless inline-code formatting around relationship targets while still rejecting real prose in validation paths - implement `kit map` as a terminal-first renderer over canonical docs and feature state - keep read-only map rendering resilient by warning on malformed relationship lines instead of failing the entire view - keep the map read-only and avoid introducing persisted derived graph files - update practical docs and command help after the command shape is stable
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0016-document-map-relationships/BRAINSTORM.md`, `docs/specs/0016-document-map-relationships/SPEC.md`, `docs/specs/0016-document-map-relationships/PLAN.md`, `docs/specs/0016-document-map-relationships/TASKS.md`

### reconcile-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit's document contract evolves over time, but older projects can drift away from current expectations for sections, tables, workflow semantics, and instruction-file structure. - Existing commands cover validation (`check`), feature catch-up (`catchup`), and handoff preparation (`handoff`), but none are designed to migrate a project's docs forward to newer Kit semantics. - Users currently need to discover document drift manually, decide which canonical source defines the current contract, and invent search strategies for filling missing content.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Add a new `pkg/cli/reconcile.go` command with project-wide default behavior, optional feature scoping, `--all`, and the shared prompt-output flags. - [PLAN-02][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16][SPEC-17][SPEC-18][SPEC-19] Implement reconciliation audit helpers that inspect Kit-managed docs for missing sections, placeholder-only required content, malformed required tables, safe structural truncation, and bounded semantic drift. - [PLAN-03][SPEC-20][SPEC-21][SPEC-22][SPEC-23] Add cross-document consistency checks for task alignment, relationship targets, rollup presence, and instruction-file drift. - [PLAN-04][SPEC-24][SPEC-25][SPEC-26][SPEC-27][SPEC-28] Build a reconciliation prompt that groups findings, cites canonical contract sources, prescribes update actions, and emits concise deduplicated search guidance plus a compact response contract. - [PLAN-05][SPEC-29][SPEC-30][SPEC-31] Integrate the command into root help and README with wording that distinguishes it from validation, catch-up, handoff, and instruction scaffolding. - [PLAN-06][SPEC-06][SPEC-12][SPEC-17][SPEC-18][SPEC-20][SPEC-21][SPEC-23][SPEC-27][SPEC-28] Add focused tests for audit findings, clean-project behavior, prompt generation, and flag handling, then run normal verification commands. - [PLAN-07] Add a human-readable terminal summary for non-`--output-only` reconcile runs without changing the raw prompt payload. - [PLAN-08] Keep the compact prompt aligned with default orchestration by explicitly telling the coding agent to use subagents and queue work according to overlapping file changes, while omitting that line under `--single-agent`.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0017-reconcile-command/SPEC.md`, `docs/specs/0017-reconcile-command/PLAN.md`, `docs/specs/0017-reconcile-command/TASKS.md`

### backlog-command

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: Users often discover legitimate follow-up work while actively defining or implementing another feature. That follow-up work is real enough to capture, but it is intentionally out of scope for the current implementation. Today the closest durable artifact is a normal feature directory, but Kit has no focused workflow for recording that future work as deferred, listing those deferred items later, or resuming one without it taking over the active feature lane.
- **APPROACH**: - keep persistence unchanged by reusing the existing paused lifecycle flag - define backlog eligibility structurally: paused + `brainstorm` phase - add shared backlog helpers for filtering, description extraction, selection, and pickup validation - make `kit brainstorm --backlog` a capture-only flow - make `kit backlog --pickup`, `kit resume`, and the deprecated `kit brainstorm --pickup` path share the same resume helper and prompt output - update active-feature selection logic so deferred items do not become active
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0018-backlog-command/SPEC.md`, `docs/specs/0018-backlog-command/PLAN.md`, `docs/specs/0018-backlog-command/TASKS.md`

### command-surface-simplification

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: Kit's current top-level command list mixes lifecycle commands, prompt-only support commands, maintenance commands, and duplicate aliases at the same level. The result is harder onboarding, denser root help, and overlapping ways to resume work (`catchup`, `backlog --pickup`, `brainstorm --pickup`) that require users to understand internal distinctions before they can move forward.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Add a new `pkg/cli/resume.go` command that routes backlog items through the existing backlog pickup helper and routes non-backlog features through the existing catch-up prompt behavior - [PLAN-02][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16] Refactor `status` so the default mode stays active-feature focused while `--all` renders the fleet view in dedicated text and JSON output paths, with the human-readable path using a fixed-width lifecycle matrix instead of a Markdown-style table - [PLAN-03][SPEC-17][SPEC-18] Rework root help rendering in `pkg/cli/root.go` and `pkg/cli/human_output.go` so only root help is grouped into product sections - [PLAN-04][SPEC-19][SPEC-20][SPEC-21][SPEC-22][SPEC-23] Convert duplicate or maintenance commands into hidden deprecated compatibility surfaces while preserving invocation behavior - [PLAN-05][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-24] Update backlog, brainstorm, catchup, upgrade, skill, README, and canonical workflow docs so they teach the simplified command surface accurately - [PLAN-06] Add or update focused tests for resume routing, status all-features output, grouped root help, and deprecated command visibility, then rerun the normal verification suite
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0019-command-surface-simplification/SPEC.md`, `docs/specs/0019-command-surface-simplification/PLAN.md`, `docs/specs/0019-command-surface-simplification/TASKS.md`

### versioned-instruction-model

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit currently scaffolds long, policy-dense `AGENTS.md` and `CLAUDE.md` files that act as encyclopedias instead of lightweight entrypoints. - That verbose model conflicts with the thin `AGENTS.md` plus progressive-disclosure pattern described in OpenAI's February 11, 2026 harness engineering article. - Kit has started to hint at repository-scale RLM work in prompts, but it does not yet give agents a repo-local knowledge tree or a strong runtime routing model. - `kit scaffold-agents` has no versioned migration model for moving between the verbose legacy layout and the thinner docs-first layout.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-09][SPEC-15] Add version state to config and teach init plus scaffold commands to load, persist, and present the active instruction-model version. - [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-14][SPEC-16] Add `v2` instruction and docs-tree templates, plus help rendering for the versioned scaffold surface. - [PLAN-03][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-23] Extend scaffold planning and write behavior to support safe version-aware downgrade flows, including removal planning for Kit-managed `v2` artifacts. - [PLAN-04][SPEC-17][SPEC-18][SPEC-19][SPEC-20] Update prompt-generation helpers so `v2` repos route agents through `docs/agents/README.md` and use explicit RLM progressive-disclosure guidance. - [PLAN-05][SPEC-21][SPEC-22] Make instruction drift detection and reconciliation version-aware. - [PLAN-06][SPEC-16][SPEC-23] Add focused tests for config defaults, versioned scaffolding, downgrade safety, prompt output, and help rendering, then run targeted verification. - [PLAN-07][SPEC-25][SPEC-26][SPEC-27] Extend command-local inventories and read-order surfaces so `map`, `handoff`, and other prompt builders reflect the active repo instruction model.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0020-versioned-instruction-model/SPEC.md`, `docs/specs/0020-versioned-instruction-model/PLAN.md`, `docs/specs/0020-versioned-instruction-model/TASKS.md`

### project-validation-and-instruction-registry

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - The current `v1` versus `v2` instruction contract is duplicated across templates, prompt helpers, map output, and version-detection helpers. - That duplication increases correctness risk because one command can learn a new repo doc or routing rule while another still uses stale assumptions. - `kit check` only validates feature-scoped docs today, so repo-level instruction drift and thin-ToC contract breakage are mostly surfaced through `kit reconcile` prompts instead of a direct validator. - Subagent guidance exists, but the shipped contract should distinguish RLM discovery from dispatch-style execution more explicitly.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02] Add a shared internal instruction-contract registry package for version detection and repo-doc metadata. - [PLAN-02][SPEC-02] Refactor current consumers to use the shared registry instead of duplicating hardcoded path sets. - [PLAN-03][SPEC-03][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08] Extend `kit check` with `--project` and reuse the repo-audit engine for project-scoped validation output. - [PLAN-04][SPEC-09][SPEC-10] Tighten the shipped subagent guidance in shared prompt suffixes and repo-local docs so RLM and dispatch stay distinct. - [PLAN-05] Add focused tests and verification for registry use, project validation, and subagent guidance wording.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0021-project-validation-and-instruction-registry/SPEC.md`, `docs/specs/0021-project-validation-and-instruction-registry/PLAN.md`, `docs/specs/0021-project-validation-and-instruction-registry/TASKS.md`

### typed-prompt-ir

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: - Kit's core product output is prompts, but many commands still build them with local `strings.Builder` logic spread across multiple files. - That makes prompt structure implicit, increases formatting drift, and makes broad prompt changes harder to reason about or test mechanically. - Repeated prompt structures such as headings, tables, numbered steps, response contracts, and doc inventories are currently duplicated as raw strings.
- **APPROACH**: - [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Add an internal prompt IR package with a minimal typed block model and a markdown/plain-text renderer. - [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Migrate prompt-producing commands to build through the IR while preserving the current output-wrapper and decorator flow. - [PLAN-03][SPEC-07] Add reusable IR helpers for repeated prompt structures where that improves consistency without hiding command-local wording. - [PLAN-04][SPEC-08][SPEC-09] Add exact-output golden tests for representative migrated prompt builders, with normalization for unstable paths and environment-specific values. - [PLAN-05][SPEC-10] Keep cached project context out of scope and avoid behavior changes outside prompt construction.
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0022-typed-prompt-ir/SPEC.md`, `docs/specs/0022-typed-prompt-ir/PLAN.md`, `docs/specs/0022-typed-prompt-ir/TASKS.md`

### worktree-safe-feature-allocation

- **STATUS**: reflect
- **PAUSED**: no
- **INTENT**: Feature numbers are currently allocated by scanning only the local `docs/specs/` tree, so separate worktrees created from the same commit can reserve the same next numeric prefix and produce conflicting feature directories after merge.
- **APPROACH**: - record the worktree-safe numbering contract and logical-ordering rules first - add a small shared allocator in `internal/feature/` that uses `git rev-parse --git-common-dir` - update feature creation to reserve numbers through the shared allocator before creating directories - add duplicate-number grouping helpers and surface them in project validation - harden status rendering so active-row comparison uses a unique feature identity instead of numeric ID alone - apply dependency ordering only to project-wide map views and only for `builds on` and `depends on`
- **OPEN ITEMS**: - none
- **POINTERS**: `docs/specs/0023-worktree-safe-feature-allocation/SPEC.md`, `docs/specs/0023-worktree-safe-feature-allocation/PLAN.md`, `docs/specs/0023-worktree-safe-feature-allocation/TASKS.md`

## LAST UPDATED

2026-04-16 14:20:32 EDT
