---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: clarify
clarification:
  status: open
  confidence: 65
  unresolved_questions: 6
feature:
  id: 0032
  slug: v1-skills-layer
  dir: 0032-v1-skills-layer
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0032-v1-skills-layer
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
---
# SPEC

## THESIS

`v1-skills-layer` should extend Kit from one-shot, prompt-only skill mining into a document-backed skill optimization lifecycle that can produce, validate, score, promote, reject, version, resolve, and garbage-collect bounded behavioral skill modules.

The canonical lifecycle ledger should be a master skills document. Generated `.kit/` artifacts remain evidence, cache, export, or staging surfaces rather than the source of truth.

## CONTEXT

This SPEC is in `clarify` phase and carries forward the completed brainstorm in `docs/specs/0032-v1-skills-layer/BRAINSTORM.md`. The brainstorm selected a document-first skill lifecycle that preserves Kit's visible Markdown contract while allowing replay/eval-driven skill scoring and promotion.

### Source Map

| ID | Source | Selector | Claim / Fact | Used For | Maps To | Status |
| -- | ------ | -------- | ------------ | -------- | ------- | ------ |
| SRC-001 | `docs/specs/0032-v1-skills-layer/BRAINSTORM.md` | `## SUMMARY` | The selected direction is a document-backed skill optimization lifecycle with a master skills document as canonical ledger. | thesis and requirements | AC-001, AC-002, TASK-001 | confirmed |
| SRC-002 | `docs/CONSTITUTION.md` | project constraints | Kit is document-first and must not rely on hidden databases or hidden runtime state. | source-of-truth boundary | AC-001, AC-003 | confirmed |
| SRC-003 | `docs/specs/0032-v1-skills-layer/BRAINSTORM.md` | `CONFLICT: requested .kit/skills versus current .agents/skills` | `.agents/skills` remains the exported discovery surface; `.kit/skills/*` may be generated evidence or staging only. | storage model | AC-002, TASK-002 | confirmed |
| SRC-004 | `docs/specs/0032-v1-skills-layer/BRAINSTORM.md` | `Existing verification and replay surfaces` | Existing verify, replay, runstore, trace, and eval surfaces can provide deterministic validation evidence. | validation design | AC-004, VAL-001 | confirmed |
| SRC-005 | user decision captured in brainstorm | approved lifecycle split | `kit reflect` notices and recommends skill candidates; `kit skill` scores, promotes, rejects, and prunes. | command boundary | AC-003, TASK-003 | confirmed |

## CLARIFICATIONS

Clarification is not complete.

- Accepted: use a master skills document as the canonical lifecycle artifact rather than making `.kit/skills/*` authoritative.
- Accepted: candidate authoring remains external or prompt-assisted for v1; Kit should not call an optimizer model directly.
- Accepted: `kit reflect` should notice and propose reusable skill candidates, while `kit skill` owns deterministic scoring, promotion, rejection, history, and pruning commands.
- Still unresolved: exact master skills document path, full command names and flags, scoring rubric, suite configuration, migration behavior, and whether any constitutional wording must change before implementation.

## REQUIREMENTS

- REQ-001: Keep the skill lifecycle document-first, with the master skills document as the durable source of truth.
- REQ-002: Keep generated `.kit/skills/*` content non-authoritative if used at all.
- REQ-003: Preserve `.agents/skills/*/SKILL.md` as the exported discovery surface for promoted project-scope skills.
- REQ-004: Extend skill workflows through deterministic commands for optimize, score, promote, reject, history, and garbage collection, subject to final clarification.
- REQ-005: Reuse existing runstore, verify, replay, trace, and eval evidence patterns instead of introducing hidden state.
- REQ-006: Keep direct model-backed optimization out of v1 unless a future spec changes the Constitution, configuration, and security model.

## ASSUMPTIONS

- Accepted: `docs/references/SKILLS.md` is the best provisional master document location because it is durable repo-wide reference material.
- Accepted: replay/eval validation cost is acceptable if promotion remains explicit and evidence-backed.
- Accepted: score-improving promotion must be deterministic and auditable.
- Blocking: exact score formula, suite configuration format, and command surface are not finalized.
- Removed: `.kit/skills/*` as the canonical lifecycle ledger.

## ACCEPTANCE CRITERIA

- AC-001: The final implementation defines a visible Markdown master skills document as the canonical lifecycle ledger.
- AC-002: Generated `.kit/skills/*` artifacts, if present, are documented and implemented as non-authoritative evidence, cache, export, or staging surfaces.
- AC-003: The command boundary between reflection-driven candidate discovery and `kit skill` scoring/promotion/rejection/pruning is explicit in code, docs, and tests.
- AC-004: Skill promotion and rejection decisions map to deterministic validation evidence from existing or extended verify/replay/eval surfaces.
- AC-005: The implementation includes tests for successful promotion, failed promotion, rejected candidate retry prevention, stale pruning, precedence resolution, malformed metadata, and history output.

## IMPLEMENTATION PLAN

Implementation has not started. The planned direction is to define the master skills document schema, then add parser/mutator helpers and deterministic `kit skill` lifecycle subcommands around existing verification evidence. The implementation must be preceded by a clarification pass that locks the master document path, command names, scoring model, lifecycle metadata, validation suite format, and migration expectations.

Likely touched areas are `pkg/cli/skill*.go`, a new internal skill document/store package, `internal/verify`, `internal/runstore`, `internal/state`, generated templates, README, Constitution, agent docs, and prompt/golden tests.

## TASK CHECKLIST

- [ ] TASK-001: Clarify master skills document path, lifecycle schema, and command surface. Maps to AC-001, AC-003.
- [ ] TASK-002: Design non-authoritative `.kit/skills/*` staging/evidence/export behavior. Maps to AC-002.
- [ ] TASK-003: Specify reflection-to-skill command boundary. Maps to AC-003.
- [ ] TASK-004: Define replay/eval score and promotion gates. Maps to AC-004.
- [ ] TASK-005: Define test matrix and validation evidence requirements. Maps to AC-005.

## VALIDATION MAP

- VAL-001: `go test ./internal/skillstore ./pkg/cli` or final approved package paths once implementation exists.
- VAL-002: `go test ./internal/verify ./internal/runstore ./internal/state` after any integration changes.
- VAL-003: Future CLI fixture checks for optimize, score, promote, reject, history, and gc.
- VAL-004: `go run ./cmd/kit check v1-skills-layer`.
- VAL-005: `go run ./cmd/kit check --project`.
- VAL-006: `go test ./...`.

## REFLECTION NOTES

Not applicable yet. No implementation has occurred for this feature. Reflection will be required after validation evidence exists.

## DOCUMENTATION UPDATES

- `docs/specs/0032-v1-skills-layer/BRAINSTORM.md`: complete brainstorm source, already present.
- `docs/specs/0032-v1-skills-layer/SPEC.md`: initialized for v2 clarify phase and populated from brainstorm context.
- `docs/PROJECT_PROGRESS_SUMMARY.md`: references this SPEC as clarify-phase work.
- Future docs likely include README, Constitution, agent docs, references index, and command capability metadata.

## DELIVERY DECISION

No delivery lane has been selected for this feature. It remains in clarify phase and should not be implemented or delivered until requirements, acceptance criteria, task checklist, validation map, and delivery intent are locked.

## EVIDENCE

- `docs/specs/0032-v1-skills-layer/BRAINSTORM.md` records the source research and clarification state.
- `go run ./cmd/kit check --project` should validate this SPEC once placeholder-only sections are removed.
