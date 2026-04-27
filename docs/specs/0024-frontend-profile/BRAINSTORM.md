# BRAINSTORM

## SUMMARY

Add an explicit `--profile=frontend` prompt profile that applies frontend-specific guidance through Kit's existing prompt-output decorator path without creating a new command or always-loaded instruction file. For frontend-profile brainstorms, create optional design-material folders under `docs/notes/<feature>/design/` and track them as non-binding dependencies unless specific assets are used.

## USER THESIS

**Recommendation**
Yes, a frontend profile makes sense. I would add `--profile=frontend`, not `--frontend`.

Reason: frontend work has materially different success criteria: visual taste, layout density, design-system fit, responsive behavior, screenshots/browser verification, asset handling, and interaction polish. The OpenAI frontend guidance is explicitly about steering UI quality and avoiding common generated-UI defaults, not just changing implementation style. It also calls out domain-specific UI judgment, visual assets, responsive layout, text overlap, palette checks, and running/viewing the app where appropriate. Sources: [frontend prompt instructions](https://developers.openai.com/api/docs/guides/frontend-prompt), [prompt guidance frontend section](https://developers.openai.com/api/docs/guides/prompt-guidance#frontend-engineering-and-visual-taste).

**Why `--profile=frontend`**
- Better than `--frontend` because it leaves room for future profiles: `mobile`, `security`, `docs`, `data`, etc.
- Should be a prompt profile, not a new command.
- Should not be loaded by default.
- Should not be inserted into root `AGENTS.md`, `CLAUDE.md`, or Copilot files.
- Should be appended only to prompt-producing commands through the existing prompt pipeline around [prompt_output.go](/Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli/prompt_output.go:25) and [subagents.go](/Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli/subagents.go:28).

I would not auto-detect frontend work as the primary mechanism. It will be wrong often enough to be annoying. Use explicit `--profile=frontend`, then optionally record that profile into the feature dependency table so later commands can carry it forward.

**How It Should Behave**
Example:

```bash
kit brainstorm dashboard-redesign --profile=frontend
kit spec dashboard-redesign --profile=frontend
kit plan dashboard-redesign --profile=frontend
kit implement dashboard-redesign --profile=frontend
```

The profile should add conditional frontend instructions such as:
- inspect existing design system before inventing new UI patterns
- use screenshots/design materials as inputs only when relevant
- implement actual usable screens, not marketing placeholders
- verify desktop and mobile rendering
- run the dev server when needed
- use browser/screenshot checks before completion
- validate text overflow, overlap, spacing, palette, responsive constraints, and interaction states

**Design Materials**
Integrate design materials into the existing notes pattern. Do not create a separate top-level workflow yet.

Recommended structure:

```text
docs/notes/0001-my-feature/
  .gitkeep
  design/
    .gitkeep
    screenshots/
    references/
```

For `--profile=frontend`, `kit brainstorm` should create:

```text
docs/notes/<feature-dir-name>/design/.gitkeep
```

and add an optional dependency row:

```md
| Design materials | design | docs/notes/0001-my-feature/design | optional frontend design input | optional |
```

The prompt should instruct the agent to list design materials, ignore `.gitkeep`, inspect only relevant screenshots/assets, and record any specific file used as `active` in `## DEPENDENCIES`.

This keeps screenshots and design references organized, feature-specific, and mostly ignored unless the frontend profile or explicit dependency path makes them relevant.

## RELATIONSHIPS

- builds on: 0009-spec-skills-discovery
- builds on: 0022-typed-prompt-ir
- related to: 0019-command-surface-simplification
- related to: 0021-project-validation-and-instruction-registry

## CODEBASE FINDINGS

- `docs/CONSTITUTION.md` classifies this as spec-driven work because it is a new capability and affects prompt-producing command behavior; the artifact pipeline and populated-section requirements apply.
- `docs/CONSTITUTION.md` also favors explicit state, agent-agnostic markdown/YAML, and CLI flags overriding configuration. That supports an explicit `--profile=frontend` flag over heuristic frontend auto-detection.
- `docs/specs/0019-command-surface-simplification/SPEC.md` argues against new top-level command surfaces and nested prompt namespaces. A profile flag fits that constraint better than `kit frontend`, `kit design`, or `kit prompt frontend`.
- `docs/specs/0022-typed-prompt-ir/SPEC.md` preserves `pkg/cli/prompt_output.go`, `pkg/cli/skills_prompt.go`, and `pkg/cli/subagents.go` as post-render decorators. Frontend profile guidance should use this same decorator boundary instead of being manually duplicated into every prompt body.
- `pkg/cli/prompt_output.go` routes prompt-producing commands through `prepareAgentPrompt(prompt)` before clipboard/stdout output, which makes it the narrowest central insertion point for optional profile guidance.
- `pkg/cli/subagents.go` currently appends skills guidance first and subagent orchestration second. A profile decorator can fit between skills and subagents so the subagent block continues to preserve command-specific and profile-specific rules above it.
- `pkg/cli/subagents.go` already owns persistent root prompt behavior via `--single-agent`, so a root persistent `--profile` is plausible. The implementation should avoid silently accepting unsupported values; a custom enum flag or early validation should reject unknown profile names.
- `docs/specs/0009-spec-skills-discovery/SPEC.md` separates execution-time skills from broader dependencies. A frontend profile is broader prompt behavior and should be tracked in `## DEPENDENCIES` rather than as a `SPEC.md` skill unless a concrete frontend skill file exists.
- `pkg/cli/spec_context.go` already requires exact `Location` values for Figma or MCP-driven design dependencies, so design materials fit the existing dependency-table model.
- `pkg/cli/brainstorm_notes.go` currently creates `docs/notes/<feature-dir>/.gitkeep` and appends a feature-notes dependency row. Frontend profile behavior can extend this with `docs/notes/<feature-dir>/design/.gitkeep` plus optional child folders without inventing a second side-channel.
- `pkg/cli/brainstorm.go` currently creates notes before writing or updating `BRAINSTORM.md`, so frontend-specific design-material creation can be conditional in the same path when `--profile=frontend` is active.
- The current feature notes directory only contains `.gitkeep`; no additional pre-brainstorm notes or screenshots were available to read.
- Official OpenAI frontend guidance emphasizes matching existing design systems, domain-appropriate UI, feature-complete controls/states, visual assets for websites/games, responsive fit, text overlap prevention, palette checks, and browser/dev-server verification. That is materially different from backend-oriented implementation prompting.

## AFFECTED FILES

- `pkg/cli/prompt_output.go` — central output wrapper currently applies `prepareAgentPrompt(prompt)` before writing prompts; likely integration point for profile-aware prompt preparation.
- `pkg/cli/subagents.go` — owns `preparePrompt`, skills suffix composition, default subagent suffix, and the existing root persistent `--single-agent` flag; likely home for shared profile suffix sequencing or adjacent profile wiring.
- `pkg/cli/skills_prompt.go` — existing shared prompt suffix precedent; profile guidance should follow a similar shared-suffix pattern without replacing skills guidance.
- `pkg/cli/brainstorm.go` — should accept/use the profile flag during brainstorm creation and conditionally create frontend design-material folders.
- `pkg/cli/brainstorm_prompt.go` — should mention design materials only when the frontend profile is active, with instructions to ignore `.gitkeep` and read only relevant assets.
- `pkg/cli/brainstorm_notes.go` — should grow a small helper for `docs/notes/<feature-dir>/design/.gitkeep` and the `Design materials` dependency row.
- `pkg/cli/spec.go`, `pkg/cli/plan.go`, `pkg/cli/tasks.go`, `pkg/cli/implement.go` — prompt-producing workflow commands that should honor `--profile=frontend` through shared prompt preparation rather than duplicated local prompt text.
- `pkg/cli/brainstorm_test.go`, `pkg/cli/brainstorm_prompt_only_test.go`, `pkg/cli/backlog_test.go`, `pkg/cli/subagents_test.go`, `pkg/cli/output_test.go` — likely focused test sites for profile suffix behavior, notes/design directory creation, dependency rows, and prompt-only non-mutation.
- `internal/templates/templates.go` — may not need changes if profile metadata remains prompt-time and dependency-table driven; avoid changing canonical document templates unless the spec later requires profile fields.
- `docs/specs/0024-frontend-profile/*` — canonical artifacts for this feature.
- `docs/notes/0024-frontend-profile/` — optional supporting input directory; currently placeholder-only.

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Feature notes | notes | docs/notes/0024-frontend-profile | optional pre-brainstorm research input | optional |
| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |
| OpenAI frontend prompt instructions | external doc | https://developers.openai.com/api/docs/guides/frontend-prompt | frontend-specific prompting requirements and design-quality guidance | active |
| OpenAI prompt guidance frontend section | external doc | https://developers.openai.com/api/docs/guides/prompt-guidance#frontend-engineering-and-visual-taste | frontend visual-taste guidance and verification expectations | active |
| Constitution | doc | docs/CONSTITUTION.md | workflow classification, command-surface principles, explicit state, and source-of-truth rules | active |
| Agent workflow docs | doc | docs/agents/WORKFLOWS.md, docs/agents/RLM.md, docs/agents/GUARDRAILS.md | just-in-time context routing and completion constraints for this brainstorm | active |
| Command-surface simplification | prior feature | docs/specs/0019-command-surface-simplification/SPEC.md | avoid new commands and keep canonical CLI surface restrained | active |
| Spec skills discovery | prior feature | docs/specs/0009-spec-skills-discovery/SPEC.md | dependency-table model for skills, design refs, and external inputs | active |
| Typed prompt IR | prior feature | docs/specs/0022-typed-prompt-ir/SPEC.md | shared prompt-output decorator boundary and prompt-producing command inventory | active |
| Project validation and instruction registry | prior feature | docs/specs/0021-project-validation-and-instruction-registry/SPEC.md | avoid instruction drift and preserve thin routing docs | active |
| Prompt output wrapper | code | pkg/cli/prompt_output.go | central profile insertion point before clipboard/stdout output | active |
| Shared prompt suffixes | code | pkg/cli/subagents.go, pkg/cli/skills_prompt.go | existing prompt augmentation precedent and suffix ordering | active |
| Brainstorm notes helpers | code | pkg/cli/brainstorm_notes.go, pkg/cli/brainstorm.go, pkg/cli/brainstorm_prompt.go | extend notes pattern for optional frontend design materials | active |
| Design materials | design | docs/notes/0024-frontend-profile/design | optional frontend design input for this feature | optional |

## QUESTIONS

No unresolved questions remain for the brainstorm phase.

Approved defaults:

1. Register `--profile` as a root persistent flag, but apply it only in prompt-producing output paths.
2. Reject unsupported profile values early with enum-like validation.
3. Persist the selected frontend profile and design-material dependencies into feature dependency tables when profile-aware commands create or touch feature artifacts, while allowing explicit flags to override.
4. For frontend brainstorms, create:
   - `docs/notes/<feature>/design/.gitkeep`
   - `docs/notes/<feature>/design/screenshots/.gitkeep`
   - `docs/notes/<feature>/design/references/.gitkeep`
5. Apply the frontend profile to planning and execution prompts because frontend quality requirements affect requirements, acceptance criteria, planning, tasks, and implementation.
6. Keep v1 flag-only; do not add a config default such as `default_profile: frontend`.
7. Keep profile text in Go constants for v1, with possible later promotion to conditional `docs/agents/` routing if profile text becomes large or customizable.

## OPTIONS

### Option A — Shared `--profile=frontend` prompt decorator plus notes/design inputs

- Add an explicit `--profile` flag with `frontend` as the first supported profile.
- Apply profile guidance in the existing prompt-preparation path so prompt-producing commands inherit it without copy-paste.
- Keep canonical root instruction files thin and avoid always-loaded frontend guidance.
- Create `docs/notes/<feature>/design/.gitkeep` only when the frontend profile is active during brainstorm/backlog capture.
- Track `Design materials` as an optional dependency and specific used assets as active dependencies.
- Tradeoff: requires careful flag validation and tests so non-prompt commands do not silently mislead users.

### Option B — Boolean `--frontend`

- Simpler to explain for the first feature.
- Harder to extend to other profiles and likely creates a family of one-off booleans.
- Less aligned with Kit's "harness" direction and future profile categories.

### Option C — Auto-detect frontend work

- Reduces user typing for obvious UI features.
- Risky because feature text, filenames, or dependencies can misclassify work.
- Better as a future advisory fallback than the primary trigger.

### Option D — Top-level `docs/design/<feature>/`

- Visually clean for design-heavy teams.
- Creates a second feature-scoped side channel parallel to `docs/notes/<feature>/`, which raises map/check/resume semantics before the workflow has proven it needs a separate lifecycle.
- Better deferred until design artifacts become repo-wide or cross-feature first-class state.

### Option E — Repo-local frontend instruction doc under `docs/agents/`

- Keeps frontend guidance repo-local and durable.
- Risks turning frontend guidance into an always-visible routing artifact unless loaded conditionally through profile prompts.
- Better as a later optimization if the profile text becomes large enough to route by link instead of embedding.

## RECOMMENDED STRATEGY

Use Option A for the first implementation.

- Implement `--profile=frontend` as an explicit prompt profile, not a new command and not `--frontend`.
- Prefer a central shared prompt decorator so `brainstorm`, `spec`, `plan`, `tasks`, `implement`, `resume`, `handoff`, `summarize`, `code-review`, `dispatch`, and `skill mine` can share behavior through existing output wrappers.
- Keep frontend guidance conditional and absent from root instruction files.
- Reject unsupported profile values early.
- Record the selected profile and design-material directory in dependency tables when profile-aware commands create or touch feature artifacts.
- Extend `docs/notes/<feature>/` with `design/`, `design/screenshots/`, and `design/references/` for frontend materials rather than introducing `docs/design/` in v1.
- Make prompt-only paths non-mutating: they may mention the expected design-material path but should not create directories or patch docs.
- Add tests that verify profile suffix insertion, no duplicate suffixes, unsupported profile rejection, frontend design-material scaffolding, dependency rows, and `--prompt-only` non-mutation.

## NEXT STEP

Run `kit spec frontend-profile` to write a binding `SPEC.md`.
