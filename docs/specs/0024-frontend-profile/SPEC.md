# SPEC

## SUMMARY

Add an explicit `--profile=frontend` prompt profile that applies frontend-specific guidance through Kit's existing prompt-output flow without adding commands, root-instruction bloat, or automatic frontend detection. Frontend brainstorms must keep design materials under `docs/notes/<feature-dir>/design/` and expose them through just-in-time dependency routing rather than inlining all assets by default.

## PROBLEM

- Kit's current prompts treat frontend implementation like ordinary code work, but frontend success depends on additional criteria: design-system fit, visual taste, responsive behavior, asset handling, interaction states, browser verification, and avoiding common generated-UI failures.
- Users need a way to opt into those frontend-specific expectations without adding a new command, expanding root instruction files, or making every Kit prompt carry frontend guidance.
- Feature-specific research and design materials need an organized location that stays mostly ignored unless a frontend profile or explicit dependency makes them relevant.

## GOALS

- Add a public `--profile=frontend` CLI flag for frontend-focused prompt generation.
- Keep profile guidance conditional, prompt-scoped, and absent from root instruction files.
- Preserve the existing command surface; do not add `kit frontend`, `--frontend`, or another validation command.
- Make frontend profile prompts tell agents to inspect existing design systems, use relevant design materials, build usable frontend experiences, and verify desktop/mobile rendering when applicable.
- Derive the frontend profile instruction set from the relevant OpenAI prompt-guidance documents while adapting it to Kit's agent-agnostic, RLM-based workflow.
- Store optional frontend design inputs under the existing feature-notes pattern at `docs/notes/<feature-dir>/design/`.
- Keep design-material context just-in-time: list or reference available materials, ignore placeholders, and load only relevant assets at runtime.
- Record active profile and design-material dependencies in feature artifact dependency tables when profile-aware commands create or update those artifacts.
- Preserve existing prompt output when no frontend profile is selected or recorded.

## NON-GOALS

- Adding a `kit frontend`, `kit design`, `kit agent`, `kit instructions`, `kit verify`, `kit validate`, or `kit doctor` command.
- Adding a boolean `--frontend` flag.
- Auto-detecting frontend work from feature names, files, dependencies, or prompt text as the primary behavior.
- Adding a `.kit.yaml` default profile setting in v1.
- Creating a top-level `docs/design/` workflow or a separate design artifact lifecycle.
- Inlining all screenshots, design files, or note contents into generated prompts by default.
- Requiring vendor-specific tools such as Figma, Playwright, Browser Use, Codex, Claude, or Copilot.
- Adding an always-loaded monolithic frontend instruction file or expanding `AGENTS.md`, `CLAUDE.md`, or `.github/copilot-instructions.md` with frontend manuals.
- Implementing additional non-frontend profiles in this feature.

## USERS

- Kit users defining frontend-heavy features such as dashboards, forms, design-system work, websites, games, and UI redesigns.
- Maintainers evolving Kit's prompt-producing commands without duplicating profile text across command bodies.
- Coding agents using Kit prompts that need frontend-specific completion criteria and validation expectations.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| rlm | repo-local agent doc | `docs/agents/RLM.md` | broad or noisy-context feature work; use immediate decision -> smallest artifact -> required facts -> act or recurse | yes |

## RELATIONSHIPS

- builds on: 0009-spec-skills-discovery
- builds on: 0022-typed-prompt-ir
- related to: 0019-command-surface-simplification
- related to: 0021-project-validation-and-instruction-registry

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Feature notes | notes | `docs/notes/0024-frontend-profile` | optional pre-brainstorm research input; currently placeholder-only | optional |
| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |
| Design materials | design | `docs/notes/0024-frontend-profile/design` | optional frontend design input path for this feature and the desired v1 directory shape | optional |
| OpenAI prompt guidance | external doc | `https://developers.openai.com/api/docs/guides/prompt-guidance` | general coding-agent prompt structure, validation, and frontend guidance context | active |
| OpenAI frontend prompt instructions | external doc | `https://developers.openai.com/api/docs/guides/frontend-prompt` | frontend-specific prompting requirements and design-quality guidance | active |
| OpenAI prompt guidance frontend section | external doc | `https://developers.openai.com/api/docs/guides/prompt-guidance#frontend-engineering-and-visual-taste` | frontend visual-taste guidance and verification expectations | active |
| Constitution | doc | `docs/CONSTITUTION.md` | workflow classification, no implementation details in specs, explicit state, and command-surface constraints | active |
| Agent workflow docs | doc | `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/RLM.md`, `docs/agents/TOOLING.md`, `docs/agents/GUARDRAILS.md` | source-of-truth order, just-in-time context loading, skill discovery, and completion rules | active |
| Command-surface simplification | prior feature | `docs/specs/0019-command-surface-simplification/SPEC.md` | preserve public command surface and avoid prompt-specific top-level commands | active |
| Spec skills discovery | prior feature | `docs/specs/0009-spec-skills-discovery/SPEC.md` | keep execution-time skills separate from broader dependency tracking | active |
| Typed prompt IR | prior feature | `docs/specs/0022-typed-prompt-ir/SPEC.md` | preserve shared prompt-output decorators as the profile insertion boundary | active |
| Project validation and instruction registry | prior feature | `docs/specs/0021-project-validation-and-instruction-registry/SPEC.md` | avoid instruction drift and keep root instruction files thin | active |
| Prompt output wrapper | code | `pkg/cli/prompt_output.go` | central prompt output behavior that must compose with profile guidance | active |
| Shared prompt suffixes | code | `pkg/cli/subagents.go`, `pkg/cli/skills_prompt.go` | existing skill and orchestration suffix behavior that frontend profile guidance must not override | active |
| Brainstorm notes helpers | code | `pkg/cli/brainstorm_notes.go`, `pkg/cli/brainstorm.go`, `pkg/cli/brainstorm_prompt.go` | existing feature-notes directory creation and brainstorm dependency-row behavior | active |
| Spec context helpers | code | `pkg/cli/spec_context.go` | dependency and design-reference prompt semantics, including exact design locations | active |

## REQUIREMENTS

- [SPEC-01] Kit must expose `--profile` as a root persistent prompt-profile flag.
- [SPEC-02] The first supported non-empty profile value must be `frontend`.
- [SPEC-03] Unsupported non-empty profile values must fail early with an actionable message that names the supported value.
- [SPEC-04] `--profile=frontend` must not create or imply a new public command, alias, command namespace, or boolean `--frontend` flag.
- [SPEC-05] When no profile is selected and no active profile dependency exists for the current feature, existing prompt output must remain semantically unchanged.
- [SPEC-06] Frontend profile guidance must be appended only to generated agent prompts, not to human status output, project validation output, root instruction files, or scaffolded always-loaded docs.
- [SPEC-07] Frontend profile guidance must compose with existing skills and subagent guidance without duplicating either section or weakening command-specific rules.
- [SPEC-08] Frontend profile guidance must be included for feature-scoped prompt-producing commands when `--profile=frontend` is explicitly supplied.
- [SPEC-09] Feature-scoped prompt-producing commands must also include frontend profile guidance when the current feature artifacts contain an active dependency indicating the frontend profile.
- [SPEC-10] Explicit `--profile` input must take precedence over profile information inferred from feature artifacts.
- [SPEC-11] Feature artifacts created or refreshed under the frontend profile must record an active dependency row for the frontend profile.
- [SPEC-12] Feature artifacts created or refreshed under the frontend profile must record the design-material directory as an optional dependency unless specific non-placeholder assets have been used.
- [SPEC-13] Specific design materials that materially shape a feature artifact must be recorded as active dependencies with exact file paths or external design references.
- [SPEC-14] Placeholder files such as `.gitkeep` must be ignored as design or note inputs and must never be recorded as active dependencies.
- [SPEC-15] `kit brainstorm --profile=frontend <feature>` must create the normal feature notes directory plus:
  - `docs/notes/<feature-dir>/design/.gitkeep`
  - `docs/notes/<feature-dir>/design/screenshots/.gitkeep`
  - `docs/notes/<feature-dir>/design/references/.gitkeep`
- [SPEC-16] `kit brainstorm --profile=frontend <feature>` must seed or preserve a `Design materials` dependency row pointing at `docs/notes/<feature-dir>/design`.
- [SPEC-17] `kit brainstorm --profile=frontend` prompts must instruct agents to inspect feature notes and design materials just in time, ignore `.gitkeep`, load only relevant files, and copy durable conclusions into `BRAINSTORM.md`.
- [SPEC-18] Prompt-only modes must remain non-mutating: `--prompt-only --profile=frontend` may include frontend guidance and expected design-material paths, but must not create directories or patch dependency tables.
- [SPEC-19] Frontend profile guidance must be a distinct coding-agent instruction set for frontend work, not a one-line reminder or generic backend implementation suffix.
- [SPEC-20] Frontend profile guidance must encode durable patterns from the OpenAI frontend and prompt-guidance documents in Kit-owned wording.
- [SPEC-21] Frontend profile prompts must not require agents to fetch OpenAI documentation at runtime before acting; the external docs are specification dependencies, while generated prompts should contain the actionable guidance needed for the current decision.
- [SPEC-22] Frontend profile guidance must preserve Kit's RLM model by directing agents to load the smallest relevant repo, feature, notes, design, and code artifacts rather than broad frontend context by default.
- [SPEC-23] Frontend profile guidance must tell agents to inspect existing frontend architecture, component libraries, styling systems, tokens, and design conventions before inventing new UI patterns.
- [SPEC-24] Frontend profile guidance must tell agents to build or evaluate the actual usable frontend experience rather than marketing placeholders when the task asks for an app, tool, game, or site.
- [SPEC-25] Frontend profile guidance must tell agents to handle expected UI states and familiar controls for the domain, including loading, empty, error, validation, interaction, responsive states, icon/tool controls, toggles, segmented controls, menus, tabs, and tooltips when relevant.
- [SPEC-26] Frontend profile guidance must tell agents to use relevant visual assets or design materials when provided or materially helpful for inspection, while avoiding broad asset loading when they are not needed for the immediate decision.
- [SPEC-27] Frontend profile guidance must require browser or screenshot-based verification when a frontend app needs a running renderer to validate the change.
- [SPEC-28] Frontend profile guidance must require agents to check for text overflow, overlapping elements, responsive layout failures, unstable fixed-format dimensions, palette problems, spacing issues, clipped controls, and broken interaction states before claiming completion.
- [SPEC-29] Frontend profile guidance must stay tool-agnostic: it may mention browser, screenshot, or design-material verification generically, but must not require a specific vendor tool unless the user's prompt or dependencies do.
- [SPEC-30] Planning prompts under the frontend profile must carry frontend acceptance concerns forward so `SPEC.md`, `PLAN.md`, and `TASKS.md` include visual, responsive, asset, and validation expectations where relevant.
- [SPEC-31] Implementation prompts under the frontend profile must still follow the existing Kit execution order: start from `TASKS.md`, recurse into linked `PLAN.md` and `SPEC.md`, inspect code before editing, and run relevant validation before completion.
- [SPEC-32] The frontend profile must not change scaffolded root instruction files or create an always-loaded `core.md`, frontend manual, or profile manual.
- [SPEC-33] At minimum, frontend profile guidance must cover `brainstorm`, `spec`, `plan`, `tasks`, `implement`, `reflect`, `resume`, `catchup`, `handoff`, `summarize`, `code-review`, `reconcile`, `dispatch`, and `skill mine` when those commands emit agent prompts.
- [SPEC-34] `--profile=frontend` on commands that do not emit agent prompts must have no profile-specific side effects.
- [SPEC-35] Frontend profile guidance must warn against common generated-UI defaults including unnecessary landing pages, generic heroes, nested cards, decorative gradients or blobs, one-note palettes, visible instructional copy, and placeholder-first layouts unless explicitly requested.
- [SPEC-36] Tests must cover profile flag registration, unsupported profile rejection, frontend guidance insertion, no-profile output behavior, no duplicate profile section, OpenAI-inspired frontend quality constraints, and composition with skills/subagent suffixes.
- [SPEC-37] Tests must cover frontend brainstorm design-directory creation, dependency-row seeding, `.gitkeep` ignoring semantics in prompts, and `--prompt-only` non-mutation.

## ACCEPTANCE

- `kit --help` exposes `--profile` as a root persistent flag without adding new public commands.
- `kit <prompt-command> --profile=frontend --output-only` includes one frontend profile guidance section.
- `kit <prompt-command> --output-only` without a profile and without an active feature profile does not include frontend profile guidance.
- An unsupported profile value fails before writing or mutating files and explains that `frontend` is supported.
- `kit brainstorm frontend-example --profile=frontend` creates `docs/notes/<feature-dir>/design/`, `design/screenshots/`, and `design/references/` with placeholder files.
- Frontend brainstorm output points agents at feature notes and design materials, tells them to ignore `.gitkeep`, and does not inline all design files by default.
- Frontend-profile artifact updates record `Frontend profile` as active and `Design materials` as optional unless specific assets are used.
- A later feature-scoped prompt for a feature with an active frontend profile dependency includes frontend guidance even when `--profile` is not repeated.
- `--prompt-only --profile=frontend` produces profile-aware prompt text without creating notes directories or changing markdown files.
- Existing public commands remain callable; no `--frontend` flag or `kit frontend` command exists.
- Root instruction files remain thin routing tables and do not include the frontend profile manual.
- Focused unit tests for profile selection, prompt composition, brainstorm design materials, dependency rows, and prompt-only behavior pass.
- Full repository verification passes with the normal Go test suite.

## EDGE-CASES

- A feature has no `docs/notes/<feature-dir>/design/` directory yet and a prompt-only frontend command is run.
- The design directory exists but contains only `.gitkeep` files.
- The design directory contains screenshots, references, or external design notes that are irrelevant to the immediate frontend decision.
- A feature contains a stale frontend profile dependency row.
- A feature contains an optional `Design materials` row but no active frontend profile row.
- `--profile=frontend` is used with a command that does not emit an agent prompt.
- `--profile=frontend` is combined with `--single-agent`.
- `--profile=frontend` is combined with commands that intentionally suppress subagent guidance, such as dispatch-style prompt output.
- Existing feature docs already contain dependency rows for notes or design materials.
- A user supplies Figma, MCP, or external design references in notes; exact references must be preserved only when those inputs shape the artifact.
- Existing tests assert exact prompt output and need stable expectations when no profile is selected.

## OPEN-QUESTIONS

- none
