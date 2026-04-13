# PLAN

## SUMMARY

- Add versioned instruction scaffolding with a `v2` thin ToC model by default, preserve the verbose legacy model behind `--version 1`, and make prompt routing plus reconciliation version-aware.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-09][SPEC-15] Add version state to config and teach init plus scaffold commands to load, persist, and present the active instruction-model version.
- [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07][SPEC-08][SPEC-14][SPEC-16] Add `v2` instruction and docs-tree templates, plus help rendering for the versioned scaffold surface.
- [PLAN-03][SPEC-10][SPEC-11][SPEC-12][SPEC-13][SPEC-23] Extend scaffold planning and write behavior to support safe version-aware downgrade flows, including removal planning for Kit-managed `v2` artifacts.
- [PLAN-04][SPEC-17][SPEC-18][SPEC-19][SPEC-20] Update prompt-generation helpers so `v2` repos route agents through `docs/agents/README.md` and use explicit RLM progressive-disclosure guidance.
- [PLAN-05][SPEC-21][SPEC-22] Make instruction drift detection and reconciliation version-aware.
- [PLAN-06][SPEC-16][SPEC-23] Add focused tests for config defaults, versioned scaffolding, downgrade safety, prompt output, and help rendering, then run targeted verification.
- [PLAN-07][SPEC-25][SPEC-26][SPEC-27] Extend command-local inventories and read-order surfaces so `map`, `handoff`, and other prompt builders reflect the active repo instruction model.

## COMPONENTS

- `internal/config/config.go`
  - persisted instruction version state
  - defaults and helpers
- `internal/templates/instruction_templates*.go`
  - `v1` verbose instruction files
  - `v2` thin entrypoint files
  - `docs/agents/*` and `docs/references/*` templates
- `pkg/cli/scaffold_agents.go`
  - `--version` flag
  - help rendering
  - version-aware planning and downgrade messaging
- `pkg/cli/instruction_files.go`
  - version-aware write planning
  - docs-tree artifact planning
  - downgrade preflight behavior
- `pkg/cli/init.go`
  - default `v2` scaffolding on new repos
- `pkg/cli/spec_context.go`
  - version-aware context rows
  - repo-local docs-first routing
- `pkg/cli/spec_template.go`
  - version-aware template prompt routing
- `pkg/cli/spec_output.go`
  - version-aware compiled prompt routing
- `pkg/cli/skills_prompt.go`
  - runtime skill/instruction suffix updates
- `pkg/cli/reconcile_audit.go`
  - version-aware instruction drift expectations
- `internal/feature/map.go`
  - version-aware global document inventory for the project map
- `pkg/cli/handoff_prompt.go`
  - version-aware handoff documentation inventory
- `pkg/cli/implement.go`, `pkg/cli/reflect.go`, `pkg/cli/catchup_prompt.go`, `pkg/cli/skill_prompt.go`, `pkg/cli/summarize.go`
  - command-local docs-tree consistency for `v2`
- tests under `pkg/cli/`, `internal/config/`, and `internal/templates/`

## DATA

- Persist instruction scaffold version in `.kit.yaml`.
- Scaffold `v2` repo-local docs under:
  - `docs/agents/`
  - `docs/references/`
- No new network or external service dependencies.
- No hidden state outside repo files and `.kit.yaml`.

## INTERFACES

- `kit scaffold-agents --version <1|2>`
- `kit scaffold-agent --version <1|2>`
- persisted `.kit.yaml` field for instruction scaffold version
- version-aware prompt output for:
  - `kit spec`
  - shared skill/instruction suffix users
  - reconcile instruction drift audits

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| scaffold-agents command | code | `pkg/cli/scaffold_agents.go` | version flag, confirmation flow, and user-facing messaging | active |
| instruction write planning | code | `pkg/cli/instruction_files.go` | version-aware create, merge, and downgrade behavior | active |
| instruction template registry | code | `internal/templates/instruction_templates.go` | legacy verbose templates and new thin templates | active |
| config persistence | code | `internal/config/config.go` | remembering the active instruction version | active |
| spec prompt builders | code | `pkg/cli/spec_context.go`, `pkg/cli/spec_template.go`, `pkg/cli/spec_output.go` | docs-first routing and RLM guidance | active |
| reconcile audits | code | `pkg/cli/reconcile_audit.go` | version-aware drift detection | active |

## RISKS

- Downgrade behavior can accidentally delete user-authored files if Kit-managed `v2` artifact ownership is not explicit and bounded.
- Prompt changes can become too verbose if `v2` adds a docs tree but still repeats the old “read everything first” wording.
- Version-aware branching can cause drift if one code path updates `v2` and another still assumes the legacy template set.
- Help output can become brittle if the version table is hard to keep aligned between TTY and non-TTY contexts.

## TESTING

- Add unit tests for:
  - config default and persistence of instruction scaffold version
  - `v2` default scaffolding
  - explicit `--version 1`
  - downgrade refusal without `--force`
  - safe downgrade with `--force`
  - ambiguous downgrade failure when extra files exist
  - prompt output differences between `v1` and `v2`
  - version-aware `map` and `handoff` inventories
  - command-local prompt references to `docs/agents/README.md`
  - version-aware reconcile findings
  - scaffold help version table rendering
- Run:
  - `go test ./internal/config ./internal/templates ./pkg/cli`
  - `make build`
  - `./bin/kit scaffold-agents --help`
  - `./bin/kit spec --help`
