---
kit_metadata_version: 1
artifact: "plan"
feature:
  id: "0011"
  slug: "handoff-document-sync"
  dir: "0011-handoff-document-sync"
references:
  - name: constitution contract
    type: doc
    target: docs/CONSTITUTION.md
    relation: informs
    read_policy: conditional
    used_for: canonical workflow and document rules
    status: active
  - name: init project spec
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: informs
    read_policy: conditional
    used_for: shipped handoff behavior summary
    status: active
  - name: project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: informs
    read_policy: conditional
    used_for: project-wide reconciliation context
    status: active
  - name: handoff command
    type: code
    target: pkg/cli/handoff.go
    relation: implements
    read_policy: conditional
    used_for: selector and prompt wiring
    status: active
  - name: handoff prompt builder
    type: code
    target: pkg/cli/handoff_prompt.go
    relation: implements
    read_policy: conditional
    used_for: document inventory and summary generation
    status: active
  - name: README
    type: doc
    target: README.md
    relation: implements
    read_policy: conditional
    used_for: user-facing command description
    status: active
---
# PLAN

## SUMMARY

- Rebuild `kit handoff` so it prepares the current coding agent session to reconcile docs, then hand off a clean, documented feature or project state.

## APPROACH

- [PLAN-01] Record the handoff prompt contract before changing code.
- [PLAN-02] Add prompt-building helpers that generate document inventory tables with absolute paths and concise usage guidance.
- [PLAN-03] Rewrite feature-scoped handoff prompts to require documentation reconciliation before handoff.
- [PLAN-04] Rewrite project-wide handoff prompts to reconcile rollup and active feature docs before handoff.
- [PLAN-05] Define a final response contract that requires documentation-sync confirmation, reference-inventory verification, a document table, and a recent-context summary.
- [PLAN-06] Add tests for feature and project-wide handoff prompt content, then rerun full verification.
- [PLAN-07] Register the shared `--prompt-only` flag on `handoff` so the command surface matches the rest of Kit's feature-scoped prompt commands.

## COMPONENTS

- `pkg/cli/handoff.go`
  - command wiring
  - selector flow
- `pkg/cli/handoff_prompt.go`
  - project-wide prompt generation
  - feature-scoped prompt generation
  - document inventory table helpers
  - reference-inventory verification guidance
- `pkg/cli/handoff_test.go`
  - feature prompt assertions
  - project-wide prompt assertions
- `README.md`
  - update handoff description
- `docs/specs/0000_INIT_PROJECT.md`
  - align handoff command behavior summary

## DATA

- Input data comes from:
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
  - feature docs under `docs/specs/<feature>/`
  - current repo state and conversation context as described in the prompt
- No new persistent state is introduced.

## INTERFACES

- Command:
  - `kit handoff [feature]`
- Modes:
  - feature scope
  - project-wide scope via selector option `0`
- Output shape:
  - clipboard-first prompt transport remains unchanged
  - prompt content becomes an active doc-sync-and-summary workflow
  - `--prompt-only` is accepted as a no-op consistency flag because the command is already prompt-only

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical workflow and document rules | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | shipped handoff behavior summary | active |
| project progress summary | doc | `docs/PROJECT_PROGRESS_SUMMARY.md` | project-wide reconciliation context | active |
| handoff command | code | `pkg/cli/handoff.go` | selector and prompt wiring | active |
| handoff prompt builder | code | `pkg/cli/handoff_prompt.go` | document inventory and summary generation | active |
| README | doc | `README.md` | user-facing command description | active |

## RISKS

- The prompt can become too verbose if it repeats both inventory data and final-response requirements without structure.
- Project-wide mode can drift into an unbounded repo audit unless it is limited to active features and rollup state.
- The final response contract can be ambiguous unless it clearly separates “update docs first” from “then report”.

## TESTING

- Add unit tests for feature-scoped handoff prompts.
- Add unit tests for project-wide handoff prompts.
- Assert prompt includes:
  - document reconciliation instructions
  - reference-inventory verification instructions
  - documentation inventory table with absolute paths
  - final response contract
  - recent conversation-context summary instructions
- Verify `kit handoff --help` exposes `--prompt-only`.
- Run:
  - `go test ./...`
  - `make vet`
  - `make build`
