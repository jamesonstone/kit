# SPEC

## SUMMARY

- Default new Kit repos to a thin table-of-contents instruction model that points agents into a repo-local docs tree, while preserving the current model for existing repos unless `--version` explicitly switches them.

## PROBLEM

- Kit currently scaffolds long, policy-dense `AGENTS.md` and `CLAUDE.md` files that act as encyclopedias instead of lightweight entrypoints.
- That verbose model conflicts with the thin `AGENTS.md` plus progressive-disclosure pattern described in OpenAI's February 11, 2026 harness engineering article.
- Kit has started to hint at repository-scale RLM work in prompts, but it does not yet give agents a repo-local knowledge tree or a strong runtime routing model.
- `kit scaffold-agents` has no versioned migration model for moving between the verbose legacy layout and the thinner docs-first layout.

## GOALS

- Make new or uninitialized repos default to a `v2` thin table-of-contents instruction model.
- Preserve the current instruction model for existing repos when `kit scaffold-agents` runs without an explicit `--version`.
- Preserve the current verbose instruction model behind `--version 1`.
- Add a repo-local knowledge tree under `docs/agents/` and `docs/references/` for the `v2` model.
- Update prompt-generation flows so `v2` repos route agents through the docs tree and use RLM-style progressive disclosure for repository-scale work.
- Persist the active instruction-model version in `.kit.yaml`.
- Make reconciliation and instruction drift checks version-aware.

## NON-GOALS

- Executing subagents directly from the Kit binary.
- Replacing `docs/specs/` as the feature system of record.
- Turning Kit into a runtime skill/plugin executor.
- Automatically deleting user-authored content during a `v1` downgrade.

## USERS

- Teams initializing a new Kit repository and wanting a default thin instruction model.
- Teams already using verbose instruction files and needing an explicit compatibility mode.
- Coding agents that need a stable entrypoint plus repo-local runtime context selection.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- builds on: `0009-spec-skills-discovery`
- builds on: `0013-scaffold-agents-safe-merge`
- builds on: `0014-human-readable-terminal-output`
- builds on: `0017-reconcile-command`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical workflow, change classification, and repo-instruction invariants | active |
| scaffold agents command | code | `pkg/cli/scaffold_agents.go` | versioned instruction scaffolding behavior | active |
| instruction templates | code | `internal/templates/instruction_templates.go` | current verbose instruction model and scaffold template wiring | active |
| instruction merge helpers | code | `pkg/cli/instruction_files.go`, `pkg/cli/instruction_file_merge.go` | safe append-only and overwrite behavior | active |
| spec prompt flow | code | `pkg/cli/spec_context.go`, `pkg/cli/spec_template.go`, `pkg/cli/spec_output.go`, `pkg/cli/spec_rlm.go` | routing prompts through repo-local docs and RLM guidance | active |
| skill prompt suffix | code | `pkg/cli/skills_prompt.go` | runtime skill and instruction discovery wording | active |
| reconcile audits | code | `pkg/cli/reconcile_audit.go` | version-aware instruction drift detection | active |
| OpenAI harness engineering article | doc | `https://openai.com/index/harness-engineering/` | thin ToC, progressive disclosure, and repo-local system-of-record model | active |

## REQUIREMENTS

- [SPEC-01] `kit scaffold-agents` must accept `--version <n>` where supported values are `1` and `2`.
- [SPEC-02] When a repo has no established instruction model, the default scaffold version must be `2`.
- [SPEC-03] When a repo already has an established instruction model and no explicit `--version` is passed, `kit scaffold-agents` must preserve the current version instead of switching models implicitly.
- [SPEC-04] `.kit.yaml` must persist the active instruction scaffold version so later commands can load the repo contract without guessing.
- [SPEC-05] Version `2` must scaffold thin `AGENTS.md` and `CLAUDE.md` files that act as table-of-contents entrypoints into repo-local docs.
- [SPEC-06] Version `2` must scaffold a compact `.github/copilot-instructions.md` that still points into the same repo-local docs tree.
- [SPEC-07] Version `2` must scaffold `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/RLM.md`, `docs/agents/TOOLING.md`, and `docs/agents/GUARDRAILS.md`.
- [SPEC-08] Version `2` must scaffold `docs/references/README.md` plus starter reference files for durable repo-local reference content.
- [SPEC-09] Version `2` must keep `docs/specs/` unchanged as the feature system of record.
- [SPEC-10] Version `1` must scaffold the current verbose instruction-file model.
- [SPEC-11] Running `kit scaffold-agents --version 1` without `--force` against an existing `v2` repo must not silently do nothing; it must explain that a downgrade requires `--force`.
- [SPEC-12] Running `kit scaffold-agents --version 1 --force` must attempt to remove the known Kit-managed `v2` docs-tree artifacts and restore verbose `AGENTS.md` and `CLAUDE.md`.
- [SPEC-13] The `v1` downgrade path must remove only the known Kit-managed `v2` scaffold set and must fail with an actionable error if extra content makes deletion ambiguous.
- [SPEC-14] `--append-only --version 1` must not attempt a downgrade and must not delete `v2` artifacts.
- [SPEC-15] In version `2`, targeted scaffolding such as `--agentsmd` or `--claude` must still scaffold the supporting `docs/agents/` and `docs/references/` tree because the top-level files depend on it.
- [SPEC-16] `kit init` must default new repositories to instruction scaffold version `2`.
- [SPEC-17] `scaffold-agents --help` must describe both versions in a human-readable fixed-width table and present `v2` as the recommended default.
- [SPEC-18] Prompt-generation flows in version `2` must prioritize repo-local docs first:
  - injected entrypoints: `AGENTS.md`, `CLAUDE.md`
  - repo-local knowledge tree: `docs/agents/*`
  - feature docs: `docs/specs/*`
  - repo-local references: `docs/references/*`
  - global inputs only as secondary context
- [SPEC-19] Version `2` prompt-generation flows must explicitly route agents through `docs/agents/README.md` instead of treating the top-level instruction files as the full contract.
- [SPEC-20] Repository-scale prompt guidance in version `2` must define RLM as a runtime discovery pattern:
  - index candidate docs, files, skills, and references
  - filter to the minimal relevant subset
  - map bounded reads or file-scoped workers
  - reduce into a synthesized result with source attribution
- [SPEC-21] Version `2` prompts must require the selected docs, skills, and references to be recorded in feature dependency tables when they materially shape the feature.
- [SPEC-22] Version-aware validation or reconciliation must enforce the thin-ToC contract only for `v2` repos and preserve the legacy contract for `v1` repos.
- [SPEC-23] Reconciliation guidance for `v2` repos must point users toward restoring the docs tree and thin entrypoints rather than the verbose model.
- [SPEC-24] Existing overwrite confirmation and append-only safety behavior must remain intact across both versions.
- [SPEC-25] `kit map` must include version-aware global instruction artifacts so `v1` repos show the verbose instruction files and `v2` repos show both the thin entrypoints and the repo-local docs tree.
- [SPEC-26] `kit handoff` must include version-aware instruction entrypoints and repo-local docs-tree artifacts in its documentation inventory for `v2` repos.
- [SPEC-27] Commands with command-local documentation inventories or read-order guidance must mention the repo-local docs tree when a repo uses `v2` and the relevant docs exist.

## ACCEPTANCE

- `kit scaffold-agents` without `--version` creates the `v2` thin instruction model for new repos and preserves the established model for existing repos.
- `kit scaffold-agents --version 1 --force` restores the verbose instruction files and removes only the known Kit-managed `v2` docs-tree artifacts.
- `kit scaffold-agents --version 1` without `--force` tells the user that downgrade requires `--force`.
- `scaffold-agents --help` shows a readable version table that explains `v1` versus `v2`.
- `kit init` creates a repo whose persisted config marks the instruction scaffold version as `2`.
- `v2` spec and execution prompt output route through `docs/agents/README.md` and describe RLM as progressive disclosure instead of bulk instruction loading.
- Reconcile or validation behavior uses the configured version when checking instruction-file drift.
- `kit map` shows the active global instruction contract for the repo's instruction model.
- `kit handoff` and other command-local prompt inventories route through the repo-local docs tree when the repo uses `v2`.

## EDGE-CASES

- A repo already contains a custom `docs/agents/` tree before scaffolding `v2`.
- A `v2` repo contains extra files under `docs/agents/` or `docs/references/` when the user requests `--version 1 --force`.
- A repo requests `--append-only --version 1` while still on the `v2` model.
- A repo selects only `--agentsmd` but still needs the supporting docs tree in `v2`.
- A repo is missing `.kit.yaml` instruction version state and needs a safe default.
- Prompt-generation commands run in a `v1` repo and must preserve legacy instruction wording.

## OPEN-QUESTIONS

- none
