---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0082"
  slug: "universal-agent-contract"
  dir: "0082-universal-agent-contract"
---
# SPEC

## PURPOSE

Replace Kit's duplicated, largely always-loaded instruction corpus with one small canonical universal contract, deterministically rendered vendor adapters, project memory, and contextual rules loaded only when their trigger applies (GH-217, modernization phase 2).

## CONTEXT

- Before this change, `CLAUDE.md`, `AGENTS.md`, and the Copilot file each carried about 2.8k words of fifteen "hard gates"; `docs/agents/GUARDRAILS.md` restated most of them, and the Constitution baseline, reference index, workflow manifests, and TOOLING/RLM/WORKFLOWS docs restated them again.
- `kit context resolve --workflow implementation-delivery` required seven rules and four routing docs, so a representative implementation task read about 22k words (about 30k tokens, 14 files) before its first edit; a fresh downstream project read about 21k words.
- History showed pointer-only gates were unreliable (#196: self-assignment failed until the rule was stated inline), so critical invariants must stay directly visible.
- About 245 exact-phrase reconcile expectations plus phrase-pinned tests made every wording change a multi-file synchronization task and stalled the earlier cleanup (#194).
- The multi-agent orchestration rule (3.4k words, 20-state lifecycle, manifests, model profiles) was a mandatory pre-plan gate with no code enforcing or consuming it.

## REQUIREMENTS

- One canonical source for the universal contract; every agent entry file is derived from it.
- Critical invariants stay directly visible in the always-loaded contract: human-only authorship and assignee, primary-checkout safety, PR-based delivery and default lane, protected branches, merge boundary and current-head readiness, infrastructure approval, hard-delete confirmation, Slack read-only, other external actions, secrets, AWS identity verification, source size, delegation safety, and truthful reporting.
- Specialized rules load only on their trigger; workflow manifests require only their own domain rules.
- Refresh must never append a second contract to an existing file, must be idempotent, and must preserve project content outside Kit's block.
- Non-goals: CLI surface reduction, `context resolve` removal, reconcile/health redesign, full existing-project migration (phases 3–4), and protected DX/bootstrap files.

## ACCEPTED PLAN

1. Author the universal contract as `internal/templates/universal_contract.md` and render it, wrapped in `BEGIN/END KIT-MANAGED CONTRACT` markers, into all three v3 entry files.
2. Drop the v3 GUARDRAILS, RLM, TOOLING, and WORKFLOWS support docs; reduce `docs/agents/README.md` and the references index to maps that do not restate rules.
3. Make the Constitution baseline a pointer for projects whose entry files carry the contract; keep the legacy baseline for pre-contract and v1/v2 projects.
4. Rewrite workflow manifests to require only domain rules; drop the circular `coding-agent-context-usage` requirement.
5. Reduce `agent-team-orchestration` to invariants and align the dispatch prompt and subagent suffix with it.
6. Replace phrase-sync tests and audits for v3 entry files with derivation, invariant-marker, budget, and idempotency tests.

Single-lane execution: the change is tightly coupled across templates, refresh planning, audits, and tests.

## DECISIONS

- Render the full contract into `CLAUDE.md` and the Copilot file instead of a pointer or `@AGENTS.md` import: pointer-only guidance proved unreliable (#196), and host import semantics were not verified. Duplication is deterministic, not hand-maintained.
- Refresh replaces only the managed block. Entry files without the block are left unchanged and reported (refresh note and reconcile finding) rather than partially appended; migrating them is phase 4.
- The Constitution baseline is chosen by whether every entry file carries the contract after the planned refresh, so pre-contract projects keep their visible legacy invariants and migrations converge in one pass.
- Constitution planning now runs after instruction planning; otherwise a v2→v3 migration picked the stale baseline and needed a second refresh.
- `docs/agents/README.md` keeps a `## Purpose` heading so append-only refresh of existing projects recognizes the file instead of failing.
- v1/v2 scaffolds, their gate constants, and their tests are unchanged until the legacy tiers retire.
- The Codex browser policy and Codex subagent binding were removed from the shared contract; the vendor-neutral browser lifecycle remains in `testing-and-environment-validation`, and personal host preferences belong in user-global instructions.
- Superseded: the mandatory multi-agent orchestration evaluation gate, `single-lane, because <reason>` records, the Pull-Request Landing Plan as a pre-read requirement in always-loaded text, and phrase-pinned reconcile expectations for v3 entry files.

## DISCOVERIES

- The v3 superseded-guidance detector only ran on files that still had phrase expectations; it now iterates its own file list so pre-contract entry files are still checked.
- `docs/agents/README.md` without a shared heading made the append-only merger fail every refresh of an existing project.
- `kit spec` rewrites every progress-summary created date; that churn was reverted and only the 0082 rows were added.

## VALIDATION

- PASS: `gofmt -l .` (no output), `go vet ./...`, `go test ./... -count=1`, `golangci-lint run ./...` (0 issues), `make build`, `git diff --check`, and a 300-line audit of changed Go files.
- PASS: fresh `kit init` in an empty repository renders the contract block and the pointer baseline.
- PASS: `kit init --refresh --dry-run --diff` on an archive copy of an existing v3 project (`lsmc-bio/aquarium`) leaves its pre-contract entry files unchanged, emits the pre-contract note, and refreshes its legacy Constitution baseline.
- Measurement, words (tokens ≈ words × 4/3), representative implementation task = `CLAUDE.md` plus required `implementation-delivery` evidence:

| Metric | Before | After |
| --- | ---: | ---: |
| Always-loaded words | 2,806 | 1,107 |
| Mandatory pre-implementation words, Kit repo | 22,461 (14 files) | 3,223 (4 files) |
| Mandatory pre-implementation words, fresh project | 20,839 (14 files) | 1,888 (4 files) |
| Fresh-init Kit documentation words | 59,590 | 43,451 |

## OUTCOME

Delivered as described in the accepted plan. Remaining risks: existing downstream v3 projects keep their pre-contract entry files and retired support docs until the phase 4 migration, and large procedural delivery rules (`github-pr-delivery`, `work-lane-gating`, `safety-guardrails`, `testing-and-environment-validation`) are now contextual but not yet condensed.

## REPOSITORY MEMORY

- Created this spec.
- Updated `docs/CONSTITUTION.md` to record the single canonical contract and remove restated universal rules.
- Updated `docs/references/README.md`, `docs/agents/README.md`, `docs/README.md`, `README.md`, `docs/commands.md`, and `docs/migration-v3.md`.
- Deleted Kit's generated `docs/agents/GUARDRAILS.md`, `RLM.md`, `TOOLING.md`, and `WORKFLOWS.md`.
