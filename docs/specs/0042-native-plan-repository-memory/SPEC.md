---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0042
  slug: native-plan-repository-memory
  dir: 0042-native-plan-repository-memory
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0042-native-plan-repository-memory
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Recenter Kit as a repository-memory and specification harness: native agent planning owns research, clarification, design, and implementation planning, while Kit ensures that consequential implementation rationale survives in canonical repository documents.

## CONTEXT

- The former default made `kit spec` emit a large V2 lifecycle-supervisor prompt and let `kit loop workflow` drive confidence, acceptance-ID, task-checklist, and validation-map gates.
- Codex already provides `/plan`; other capable agents expose equivalent native planning behavior. Kit should complement those surfaces instead of duplicating them.
- Kit already has the durable primitives needed for the new role: living `SPEC.md` files, repository instructions, Constitution and reference routing, status, validation, evidence, review, repair, and delivery guardrails.
- The implementation agent keeps the accepted native plan in same-thread context and translates it semantically. Kit does not ingest transcripts or automatically copy a proposed-plan response.
- This feature was dogfooded by creating and populating a V2 feature spec before implementation edits, then semantically curating that artifact into this V3 record after implementation and validation.

## REQUIREMENTS

- New living specs use `workflow_version: 3` with PURPOSE, CONTEXT, REQUIREMENTS, ACCEPTED PLAN, DECISIONS, DISCOVERIES, VALIDATION, OUTCOME, and REPOSITORY MEMORY.
- Requirements carry non-goals and observable acceptance without mandatory identifiers, a durable task checklist, clarification confidence, unresolved-question counters, or a 1:1 validation map.
- Before implementation, V3 requires populated purpose, context, requirements, and accepted plan. Completion also requires resolved decisions and discoveries, validation, actual outcome, repository-memory assessment, and no pending TODO placeholders.
- Plain `kit spec <feature>` is non-interactive and only scaffolds, adopts, or orients durable feature memory. It must not emit the lifecycle-supervisor contract.
- The V2 supervisor remains temporarily available through `--legacy-supervisor`; former supervisor-only flags imply compatibility mode and warn.
- Existing V1 and V2 specs remain readable and valid and are not mechanically rewritten into V3.
- Bare `kit loop` and `kit loop workflow` are deprecated V2 compatibility paths. They warn for V2 work and reject V3 specs with native-planning guidance. Review, validation, evidence, repair, status, delivery, and prompt utilities remain available.
- `kit dispatch` supports post-plan execution topology and does not own feature design.
- New repositories default to `instruction_scaffold_version: 3`. A full refresh atomically migrates only an exact generated V2 instruction set; customized V2 instructions remain untouched and receive reviewed-reconciliation or explicit `kit scaffold agents --version 3 --force` guidance.
- Generated rules enforce semantic memory assessment before code, spec capture when material rationale exists, live decision/discovery updates, post-validation curation, scope-based memory routing, and the required Repository Memory final response.
- Product docs and command metadata teach native planning → implementation → curated repository memory without product-level “Kit v2” branding.
- Observable acceptance: V3 and V2 unit/CLI coverage passes; primary `kit spec` output contains concise native-planning guidance and no supervisor contract; project validation succeeds with legacy migration advisories remaining non-blocking warnings; the prompt-system suite passes all tasks.
- Non-goals: transcript ingestion, automatic documentation-diff heuristics, hidden databases or memory stores, mandatory ADRs, immediate deletion of compatibility implementations, or deterministic V2-to-V3 spec conversion.

## ACCEPTED PLAN

1. Add a compact V3 spec template plus version-aware parsing, summaries, phase detection, checks, mapping, status, and completion gates while preserving V2 behavior.
2. Make plain `kit spec` a concise non-interactive scaffold/adopt/orient command and isolate the V2 supervisor behind deprecated compatibility routing.
3. Add V3 repository-instruction templates and atomic exact-template migration while protecting customized V2 files.
4. Deprecate V2 workflow-loop entry points, retain focused execution/review utilities, and position dispatch after native planning.
5. Reposition README, Constitution, workflow, command, capability, root-help, and generated-agent guidance around curated repository memory.
6. Add V3, V2 compatibility, migration, CLI, completion, loop, capability, and two-outcome workflow fixtures; run the full validation matrix and curate this spec to the actual result.

## DECISIONS

- Accepted semantic, agent-owned memory classification because the significance of rationale cannot be inferred reliably from file diffs.
- Accepted a compact V3 schema and phase-aware visible-content gates; rejected carrying V2 confidence, `AC-###`, checklist, and validation-map bureaucracy into new work.
- Accepted same-thread semantic plan translation; rejected transcript parsing and automatic proposed-plan extraction because they would couple Kit to agent internals and preserve uncurated chatter.
- Accepted staged deprecation for the V2 supervisor and workflow loop so existing projects keep a callable migration path; rejected immediate deletion as an unnecessary breaking change.
- Accepted automatic instruction migration only when every V2 managed artifact exactly matches its generated template. Customized V2 files remain on V2 even under a general forced refresh; the explicit versioned scaffold command is the overwrite boundary.
- Accepted legacy V2 reference vocabulary such as `governs` and `loaded` as visible, non-blocking compatibility advisories. Malformed metadata and non-legacy contract drift remain blocking.
- Accepted `not required` as a valid memory outcome when code and tests contain the complete durable truth; rejected documentation created solely for ceremony.
- Superseded this feature's initial V2 planning artifact with this semantically curated V3 record only after the implementation and validation were complete.

## DISCOVERIES

- The V3 instruction audit initially reused V2 prose expectations, so a correct V3 scaffold failed `kit check --project`. The audit now selects version-specific guidance.
- Historical V2 specs in this repository use `governs` and `loaded` reference vocabulary. Compatibility handling now keeps those specs readable without rewriting their history, while surfacing migration advisories.
- Metadata reference upserts previously replaced unrelated references when adding the notes or frontend-profile dependency. The upsert path now merges by ID/target so adoption preserves existing context.
- Legacy scaffold, plan, task, and workflow-loop helpers had to call the retained V2 builder explicitly; otherwise changing the default template would silently create V3 artifacts inside V2-only flows.
- The first prompt-system run found that the deprecated bare-loop capability had lost its documented generated `gpt-5.6` default. Restoring that compatibility caveat fixed all three repeated failures.
- The project rollup always advertised legacy `PLAN.md` and `TASKS.md` pointers. V3 summaries now read PURPOSE and ACCEPTED PLAN from the living spec and publish only the V3 artifacts that actually exist.
- PR review found that the atomic refresh path discarded rollback failures. Refresh now attempts every rollback in reverse order and reports each path that could not be restored or removed alongside the original apply error.
- PR review also found that a V2-configured project containing exact legacy V1 instruction files received a customized-V2 advisory even though those files were refreshed. Exact V1 refreshes and genuine customized V2 preservation now have distinct notes.
- Generated Copilot guidance and catch-up prompts retained two incomplete compatibility phrases: mutation routing omitted relevant repository delivery rules, and catch-up still named a V2 phase. The generator, checked-in guidance, prompts, and regression tests now use the repository-local rule precedence and living-spec terminology.

## VALIDATION

- `make fmt` — passed.
- `make vet` — passed (`go vet ./...`).
- `go test ./...` — passed across all packages, including V3 document, feature phase, summary, completion, native `kit spec`, instruction migration/rollback, loop deprecation, capabilities, and V2 compatibility coverage.
- `make build` — passed; built `bin/kit` at version `v1.0.89`.
- `./bin/kit spec native-plan-repository-memory` — emitted only concise native-planning orientation, preserved the existing V2 artifact before semantic curation, and contained no lifecycle-supervisor contract.
- `./bin/kit check native-plan-repository-memory` — passed all V3 document and metadata checks.
- `./bin/kit complete native-plan-repository-memory` — passed the V3 completion gate, preserved `workflow_version: 3`, set `phase: complete`, and refreshed the project rollup. It also reported that the separate five-feature Constitution refresh threshold is now due.
- Final `./bin/kit check --project` — exited successfully; reported 15 non-blocking compatibility advisories from historical V2 specs, the due project-refresh advisory, and no blocking findings.
- Initial `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit` run `20260715T163628.210723000Z-b7f2de` — failed three repeated model-capability assertions and exposed the missing `gpt-5.6` caveat.
- Pre-review corrected rerun `20260715T164401.327451000Z-2e581a` — passed all 45 traces and all 345 assertions with 100% task success, output completeness, and repeated-task determinism.
- Post-review repair rerun `20260715T181108.912339000Z-5f306e` — passed 45/45 traces, 345/345 assertions, and 15/15 repeated-task determinism checks against rebuilt binary SHA-256 `f21b6984540122904d532b9f72cc4c3bda4abbc1fb0861a52f29800840b2c4d8`.

## OUTCOME

- Kit now defaults to compact V3 living specifications and instruction scaffold V3 while retaining supported V1/V2 inputs.
- Native agent planning is the primary research and design surface; plain `kit spec` creates or adopts durable memory and prints concise orientation.
- V3 completion preserves workflow version and enforces the final semantic-memory sections and placeholder gate.
- Exact generated V2 instruction sets migrate atomically; customized sets stay untouched until an explicit reviewed version change.
- Deprecated supervisor and workflow-loop behavior remains callable for V2 work with warnings, while V3 receives native-planning guidance.
- Generated instructions, product docs, command help, capabilities, fixtures, and tests consistently route feature rationale, invariants, reusable practices, domain knowledge, and justified no-update outcomes.
- Project rollups summarize V3 purpose and accepted plans without pointing at nonexistent legacy plan/task files.
- Atomic instruction refreshes now surface incomplete rollback operations, and migration notes distinguish exact legacy V1 refreshes from customized V2 preservation.
- GitHub delivery was opened as issue `#58` on branch `GH-58`; the pull request URL and live CI state remain authoritative on GitHub rather than being duplicated as feature rationale.

## REPOSITORY MEMORY

Decision: refactored

Rationale: The feature changes Kit's durable workflow contract and future implementation strategy, so the rationale cannot live only in code. The original pre-implementation V2 spec was semantically refactored into this V3 history, project invariants and product guidance were updated, and generated agent rules now enforce the same contract in downstream repositories.

Artifacts:

- `docs/specs/0042-native-plan-repository-memory/SPEC.md`
- `docs/CONSTITUTION.md`
- `README.md`, `docs/overview.md`, `docs/workflows.md`, and `docs/commands.md`
- `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, and `docs/agents/*`
- `internal/templates/instruction_templates_v3.go` and V3 workflow fixtures under `internal/templates/testdata/repository-memory/`
- Version-aware document, feature, CLI, capability, migration, compatibility, and completion code and tests under `internal/` and `pkg/cli/`
