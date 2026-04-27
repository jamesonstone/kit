# PLAN

## SUMMARY

Implement `--profile=frontend` as a small prompt-profile layer in the existing prompt preparation pipeline, with explicit enum validation, feature-artifact dependency detection, and frontend-specific notes/design scaffolding for brainstorms. Keep profile text in Go constants for v1, append it between the existing skills and subagent suffixes, and mutate only standardized feature dependency tables outside prompt-only paths.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04] Add a root persistent `--profile` flag backed by an enum-like flag value that accepts only empty or `frontend`, so invalid values fail during flag parsing before any command `RunE` mutates files.
- [PLAN-02][SPEC-05][SPEC-06][SPEC-07][SPEC-19][SPEC-20][SPEC-21][SPEC-22] Introduce a shared prompt-profile decorator that appends a single `## Frontend Profile` section after `## Skills` and before `## Subagent Orchestration`; keep OpenAI-inspired guidance encoded in Kit-owned text and do not require agents to fetch OpenAI docs at runtime.
- [PLAN-03][SPEC-08][SPEC-09][SPEC-10][SPEC-33][SPEC-34] Route prompt-producing commands through profile-aware preparation with explicit profile taking precedence over feature-artifact inference; use feature path context where commands already resolve a feature, and use explicit profile only for project-wide or generic prompts.
- [PLAN-04][SPEC-11][SPEC-12][SPEC-13][SPEC-14] Add shared dependency-table helpers for `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` that append or preserve `Frontend profile` and `Design materials` rows, remove the default `none` row when real rows exist, avoid duplicates, and leave stale rows intact.
- [PLAN-05][SPEC-15][SPEC-16][SPEC-17][SPEC-18] Extend the brainstorm notes path so `kit brainstorm --profile=frontend` creates the design-material directory tree and seeds dependency rows, while `--prompt-only --profile=frontend` only mentions expected paths in generated guidance and performs no filesystem or markdown mutations.
- [PLAN-06][SPEC-23][SPEC-24][SPEC-25][SPEC-26][SPEC-27][SPEC-28][SPEC-29][SPEC-30][SPEC-31] Keep the frontend profile guidance concise but materially different from backend prompting: RLM context routing first, existing design-system inspection, actual usable UI output, state/control coverage, relevant visual assets, responsive/browser/screenshot validation, and layout/text/palette checks before completion.
- [PLAN-07][SPEC-32] Do not modify `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, scaffolded root instruction templates, or add any always-loaded profile manual.
- [PLAN-08][SPEC-36][SPEC-37] Add focused tests around flag parsing, prompt composition, frontend quality constraints, feature dependency inference, brainstorm design scaffolding, prompt-only non-mutation, and default no-profile behavior; finish with feature-doc validation and the normal Go test suite.

Tradeoff decisions:

- Keep profile guidance in Go constants for v1 rather than `docs/agents/` because the profile is conditional prompt behavior, not repo-local always-loaded routing guidance.
- Use explicit feature path context for dependency inference rather than parsing generated prompt text or relying on global active-feature state.
- Persist profile dependencies only in docs with standardized dependency tables: `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`. `TASKS.md` keeps its existing task-dependency section shape and receives frontend guidance through prompts only.
- Let `dispatch` keep suppressing subagent guidance while still allowing the frontend profile suffix, because profile quality guidance and subagent orchestration are independent decorators.
- Do not create design-material directories from `spec`, `plan`, `tasks`, `implement`, or prompt-only flows; only frontend brainstorm creation owns that filesystem scaffold.

## COMPONENTS

- [PLAN-01] `pkg/cli/prompt_profile.go`
  - Owns supported profile constants, flag value validation, profile suffix text, suffix de-duplication, and effective-profile selection.
  - Provides profile-aware prompt preparation helpers that accept optional feature path context.
  - Keeps profile text tool-agnostic and RLM-compatible.

- [PLAN-02] `pkg/cli/subagents.go` and `pkg/cli/prompt_output.go`
  - Preserve the existing public helper names where possible.
  - Add feature-aware variants for commands that can pass a resolved feature path.
  - Maintain suffix order: command prompt -> skills -> frontend profile when active -> subagents when enabled.
  - Preserve `preparePromptWithoutSubagents` behavior by suppressing only subagent guidance, not profile guidance.

- [PLAN-03] Prompt-producing command call sites in `pkg/cli/`
  - Feature-scoped commands pass resolved feature paths into profile-aware output:
    `brainstorm`, backlog pickup prompt path, `spec`, `plan`, `tasks`, `implement`, `reflect`, `resume`/`catchup`, feature `handoff`, feature `summarize`, and `skill mine`.
  - Generic or project-wide commands use explicit `--profile` only:
    `dispatch`, project `handoff`, project `reconcile`, generic `summarize`, and `code-review`.
  - Existing prompt-only checks remain command-owned and run before any mutation.

- [PLAN-04] `pkg/cli/profile_dependencies.go`
  - Reads dependency rows from `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`.
  - Detects active frontend profile rows by dependency/type/location/status rather than broad prose search.
  - Appends canonical profile and design rows idempotently.
  - Reuses `document.Parse`, `document.Write`, and the existing markdown section replacement pattern.

- [PLAN-05] `pkg/cli/brainstorm_notes.go` and `pkg/cli/brainstorm.go`
  - Add design-material path helpers under `docs/notes/<feature-dir>/design`.
  - Create `.gitkeep` files for `design`, `design/screenshots`, and `design/references` only during non-prompt-only frontend brainstorm creation.
  - Seed `Feature notes`, `Frontend profile`, and `Design materials` dependencies without removing existing user rows.

- [PLAN-06] Tests in `pkg/cli/*_test.go`
  - Add or extend tests near the behavior under test rather than building a broad integration-only suite.
  - Prefer direct helper tests for suffix ordering, dependency parsing, and directory creation, plus a small number of command-level tests for mutation boundaries.

## DATA

- Prompt profile enum:
  - empty string: no explicit profile selected
  - `frontend`: frontend profile selected

- Canonical dependency rows:

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Frontend profile | profile | `--profile=frontend` | apply frontend-specific coding-agent instruction set | active |
| Design materials | design | `docs/notes/<feature-dir>/design` | optional frontend design input | optional |

- Effective profile resolution:
  - explicit valid `--profile=frontend`
  - otherwise active `Frontend profile` dependency in the current feature's `BRAINSTORM.md`, `SPEC.md`, or `PLAN.md`
  - otherwise no profile

- Feature dependency detection:
  - Search only the `## DEPENDENCIES` table in standardized feature docs.
  - Treat status case-insensitively and only `active` as enabling inferred profile behavior.
  - Ignore `stale`, `optional`, malformed, and placeholder `none` rows for profile activation.
  - Strip harmless inline-code backticks around cells before comparison.

- Design-material filesystem layout:

```text
docs/notes/<feature-dir>/
  .gitkeep
  design/
    .gitkeep
    screenshots/
      .gitkeep
    references/
      .gitkeep
```

- Frontend profile suffix data:
  - Heading: `## Frontend Profile`
  - Content groups: context routing, design-system fit, domain/audience fit, usable UI output, familiar controls, state coverage, assets/design materials, generated-UI anti-patterns, responsive stability, validation evidence.
  - The suffix must contain actionable guidance and no runtime instruction to fetch OpenAI documentation.

## INTERFACES

- CLI flag:
  - `kit ... --profile=frontend`
  - Empty profile remains the default.
  - Unsupported values fail before command execution and before file creation or writes.

- Prompt output interfaces:
  - Existing helpers continue to work for no-feature calls.
  - New feature-aware helper variants accept `featurePath` and use it for dependency-based profile inference.
  - `outputPromptWithoutSubagentsWithClipboardDefault` continues to suppress subagent guidance but still allows frontend guidance when effective profile is frontend.

- Command behavior:
  - `brainstorm --profile=frontend <feature>` creates notes/design folders, creates or updates `BRAINSTORM.md`, seeds dependency rows, refreshes rollup, and emits frontend-profile prompt text.
  - `brainstorm --prompt-only --profile=frontend <feature>` emits frontend-profile prompt text and expected design paths without creating directories or changing markdown.
  - `spec`, `plan`, and `tasks` include frontend guidance when explicit or inferred; non-prompt-only `spec` and `plan` persist standardized dependency rows when the effective profile is frontend.
  - `implement`, `reflect`, `resume`, `catchup`, feature `handoff`, feature `summarize`, and `skill mine` include frontend guidance when explicit or inferred and do not mutate profile dependencies themselves.
  - `dispatch`, project `handoff`, project `reconcile`, generic `summarize`, and `code-review` include frontend guidance only when `--profile=frontend` is explicit.
  - Non-prompt commands may accept the root flag but must not perform frontend-specific side effects.

- Files and artifacts touched during implementation:
  - `pkg/cli/prompt_profile.go`
  - `pkg/cli/prompt_output.go`
  - `pkg/cli/subagents.go`
  - `pkg/cli/brainstorm_notes.go`
  - `pkg/cli/brainstorm.go`
  - `pkg/cli/brainstorm_prompt.go`
  - prompt-producing command files that need feature-aware helper calls
  - focused tests under `pkg/cli/`
  - `docs/specs/0024-frontend-profile/TASKS.md` after task generation only

- Files intentionally not touched:
  - `AGENTS.md`
  - `CLAUDE.md`
  - `.github/copilot-instructions.md`
  - `internal/templates/instruction_templates*.go`
  - new `docs/agents/*` profile manuals
  - top-level `docs/design/`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Constitution | doc | `docs/CONSTITUTION.md` | artifact pipeline, explicit-state rule, command-surface constraints, and Go quality rules | active |
| RLM guide | repo-local agent doc | `docs/agents/RLM.md` | preserve just-in-time context routing inside frontend profile guidance and planning | active |
| Agent workflow docs | doc | `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/TOOLING.md`, `docs/agents/GUARDRAILS.md` | source-of-truth order, skills/routing context, and completion rules | active |
| Frontend profile spec | spec | `docs/specs/0024-frontend-profile/SPEC.md` | binding feature contract for profile behavior, design materials, prompts, and tests | active |
| Frontend profile brainstorm | research | `docs/specs/0024-frontend-profile/BRAINSTORM.md` | validated implementation seams and prior design decisions | active |
| OpenAI prompt guidance | external doc | `https://developers.openai.com/api/docs/guides/prompt-guidance` | source inspiration for agent prompt structure and validation expectations | active |
| OpenAI frontend prompt instructions | external doc | `https://developers.openai.com/api/docs/guides/frontend-prompt` | source inspiration for frontend-specific coding-agent guidance | active |
| OpenAI prompt guidance frontend section | external doc | `https://developers.openai.com/api/docs/guides/prompt-guidance#frontend-engineering-and-visual-taste` | source inspiration for visual-taste and frontend verification guidance | active |
| Command-surface simplification | prior feature | `docs/specs/0019-command-surface-simplification/SPEC.md` | avoid new commands and keep deprecated surfaces callable | active |
| Spec skills discovery | prior feature | `docs/specs/0009-spec-skills-discovery/SPEC.md` | keep profile/dependency tracking separate from execution-time skills | active |
| Typed prompt IR | prior feature | `docs/specs/0022-typed-prompt-ir/SPEC.md` | keep shared decorators after prompt rendering rather than duplicating prompt bodies | active |
| Project validation and instruction registry | prior feature | `docs/specs/0021-project-validation-and-instruction-registry/SPEC.md` | preserve thin root instruction files and avoid instruction drift | active |
| Prompt output wrapper | code | `pkg/cli/prompt_output.go` | shared prompt output and profile insertion boundary | active |
| Shared prompt suffixes | code | `pkg/cli/subagents.go`, `pkg/cli/skills_prompt.go` | suffix ordering and composition with skills/subagent guidance | active |
| Brainstorm notes helpers | code | `pkg/cli/brainstorm_notes.go`, `pkg/cli/brainstorm.go`, `pkg/cli/brainstorm_prompt.go` | feature notes and frontend design-material scaffolding | active |
| Spec context helpers | code | `pkg/cli/spec_context.go`, `pkg/cli/plan.go` | dependency inventory prompt semantics and related-feature RLM text | active |
| Dependency map parser | code | `internal/feature/map.go` | reference behavior for dependency table parsing and map output expectations | active |
| Feature notes | notes | `docs/notes/0024-frontend-profile` | optional pre-brainstorm input; placeholder-only for this feature | optional |
| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |
| Design materials | design | `docs/notes/0024-frontend-profile/design` | optional frontend design input path; no assets currently present | optional |

## RISKS

- Profile suffix may leak into prompts where it is not relevant.
  - Mitigation: make explicit command behavior part of tests, use feature-aware helpers only where a feature is known, and keep generic/project commands explicit-profile only.

- Unsupported profile validation could happen after side effects.
  - Mitigation: validate through the flag value or Cobra flag parsing layer before command `RunE`; add a command-level mutation-boundary test.

- Dependency inference could accidentally activate from stale or optional rows.
  - Mitigation: parse only standardized dependency tables and require an active canonical frontend profile row.

- Dependency row mutation could erase user content or duplicate rows.
  - Mitigation: use section-level replacement only for `## DEPENDENCIES`, preserve all non-target rows, remove only the placeholder `none` row, and test idempotency.

- Prompt-only frontend brainstorm could create notes/design folders by sharing the normal creation path.
  - Mitigation: keep prompt-only paths read-only and test absence of notes/design directories after prompt-only runs.

- Frontend guidance could become too large and weaken RLM by encouraging broad context loading.
  - Mitigation: keep the suffix concise, lead with smallest relevant artifact loading, and defer broad asset/design inspection unless needed for the immediate decision.

- Exact-output tests may become brittle after adding a cross-cutting suffix.
  - Mitigation: update tests around stable invariant substrings and add targeted suffix ordering tests instead of overfitting every full prompt.

## TESTING

- Unit tests:
  - Profile flag registration and enum validation on the root command.
  - No-profile prompt output remains unchanged except existing skills/subagent suffixes.
  - Frontend profile suffix appears exactly once.
  - Suffix order is command prompt -> skills -> frontend profile -> subagents.
  - `--single-agent` removes subagent guidance but preserves frontend guidance.
  - Dispatch/no-subagent helper preserves frontend guidance while omitting subagent guidance.
  - Feature dependency inference activates only from active canonical `Frontend profile` rows.
  - Stale, optional, malformed, and placeholder dependency rows do not activate the profile.
  - Dependency row helpers append idempotently, preserve existing rows, and remove default `none` rows when real dependencies exist.

- Command-focused tests:
  - `kit brainstorm --profile=frontend` creates feature notes plus `design`, `design/screenshots`, and `design/references` `.gitkeep` files.
  - Frontend brainstorm seeds `Feature notes`, `Frontend profile`, and `Design materials` rows.
  - `brainstorm --prompt-only --profile=frontend` does not create directories, output files, or markdown changes.
  - `spec` and `plan` persist standardized dependency rows when explicit or inferred frontend profile is active.
  - Feature-scoped prompt commands include frontend guidance from inferred profile dependencies when `--profile` is omitted.
  - Generic/project prompts include frontend guidance only when the profile is explicit.
  - Unsupported profile values fail before file creation in at least one mutating command test.

- Evidence commands:
  - `go test ./...`
  - `make build`
  - `git diff --check`
  - `./bin/kit check frontend-profile`
  - `./bin/kit map 0024-frontend-profile`
