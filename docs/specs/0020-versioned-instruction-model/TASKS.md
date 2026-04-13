# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Record versioned instruction-model feature artifacts | done | agent | |
| T002 | Add config state and versioned scaffold templates | done | agent | T001 |
| T003 | Implement version-aware scaffold planning and downgrade behavior | done | agent | T002 |
| T004 | Update prompt routing and RLM docs-first guidance for v2 repos | done | agent | T002 |
| T005 | Make reconcile and validation version-aware | done | agent | T003, T004 |
| T006 | Add help output, tests, and verification | done | agent | T003, T004, T005 |
| T007 | Make map, handoff, and command-local inventories version-aware | done | agent | T004, T005 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Record versioned instruction-model feature artifacts [PLAN-01] [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05] [PLAN-06]
- [x] T002: Add config state and versioned scaffold templates [PLAN-01] [PLAN-02]
- [x] T003: Implement version-aware scaffold planning and downgrade behavior [PLAN-01] [PLAN-03]
- [x] T004: Update prompt routing and RLM docs-first guidance for v2 repos [PLAN-04]
- [x] T005: Make reconcile and validation version-aware [PLAN-05]
- [x] T006: Add help output, tests, and verification [PLAN-02] [PLAN-03] [PLAN-04] [PLAN-05] [PLAN-06]
- [x] T007: Make map, handoff, and command-local inventories version-aware [PLAN-07]

## TASK DETAILS

### T001
- **GOAL**: capture the approved contract before code changes
- **SCOPE**:
  - create `SPEC.md`
  - create `PLAN.md`
  - create `TASKS.md`
- **ACCEPTANCE**:
  - `docs/specs/0020-versioned-instruction-model/` exists
  - the docs define `v1` versus `v2`, downgrade rules, and prompt-routing expectations

### T002
- **GOAL**: add persisted version state and the new `v2` scaffolded artifact set
- **SCOPE**:
  - update config defaults and persistence helpers
  - add `v1` and `v2` instruction template selection
  - add `docs/agents/*` and `docs/references/*` templates
- **ACCEPTANCE**:
  - new repos default to version `2`
  - `v2` scaffolding produces thin top-level files plus the docs tree

### T003
- **GOAL**: implement safe version-aware scaffold writes and downgrade behavior
- **SCOPE**:
  - add `--version`
  - add version-aware planning across skip, append-only, and overwrite modes
  - support bounded `v2` artifact removal during `v1` downgrade
- **ACCEPTANCE**:
  - `--version 1` without `--force` explains that downgrade requires `--force`
  - `--version 1 --force` restores verbose top-level files and removes only known Kit-managed `v2` artifacts
  - append-only mode never downgrades

### T004
- **GOAL**: make `v2` prompts route through repo-local docs and use explicit RLM progressive disclosure
- **SCOPE**:
  - update spec prompt builders
  - update shared skills/instruction suffix wording
  - keep `v1` wording stable where appropriate
- **ACCEPTANCE**:
  - `v2` prompts route to `docs/agents/README.md`
  - `v2` RLM wording describes `index -> filter -> map -> reduce`
  - prompts treat global inputs as secondary context
- **NOTES**:
  - the repo-local `RLM.md` and `GUARDRAILS.md` docs explicitly remind agents to keep affected documentation current and properly formatted
  - the shared code-hygiene guidance now covers dead code, unused exports, and public surfaces that are not strictly necessary
  - `RLM.md`, `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` now define RLM behaviorally on first use so newcomer agents do not have to infer the acronym
  - `.github/copilot-instructions.md` includes a compact fallback read order and inline non-negotiable rules for weaker doc-traversal environments
  - the default `.kit.yaml` contract now lists `.github/copilot-instructions.md` alongside `AGENTS.md` and `CLAUDE.md` so repository instruction files are explicit in config as well as runtime prompt context

### T005
- **GOAL**: enforce the correct instruction contract per repo version
- **SCOPE**:
  - update reconcile instruction drift checks
  - keep `v1` and `v2` expectations separate
- **ACCEPTANCE**:
  - `v2` repos are checked against the thin-ToC contract
  - `v1` repos continue to use the verbose contract

### T006
- **GOAL**: expose the new versioned model clearly and prevent regression
- **SCOPE**:
  - add `scaffold-agents --help` version table
  - add focused tests
  - run targeted verification
- **ACCEPTANCE**:
  - help output explains the two versions clearly
  - tests cover version selection, downgrade safety, prompt routing, and reconcile behavior

### T007
- **GOAL**: keep command-local document inventories aligned with the active instruction model
- **SCOPE**:
  - update `kit map` global docs output
  - update `kit handoff` documentation inventories
  - update command-local read-order or context-doc prompts where `v2` repo docs materially improve routing
- **ACCEPTANCE**:
  - `kit map` shows the active instruction contract for `v1` and `v2`
  - `kit handoff` inventories the thin entrypoints and repo-local docs tree for `v2`
  - prompt builders with local doc inventories mention `docs/agents/README.md` when that docs tree is present

## DEPENDENCIES

- T002 depends on T001 because templates and config changes must follow the recorded contract
- T003 depends on T002 because version-aware scaffold behavior needs the new templates and config state
- T004 depends on T002 because prompt routing depends on the new `v2` docs tree contract
- T005 depends on T003 and T004 because reconcile must reflect the shipped scaffold and prompt behavior
- T006 depends on T003, T004, and T005 because tests and help output must validate the final behavior
- T007 depends on T004 and T005 because command-local inventories must match the shipped prompt and validation contract

## NOTES

- `docs/specs/` remains the feature system of record in both versions
- `v2` is the recommended default model
