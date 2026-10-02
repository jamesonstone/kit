---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0087"
  slug: "investigation-evidence"
  dir: "0087-investigation-evidence"
relationships:
  - type: builds_on
    target: 0084-harness-subtraction
    note: Adds an optional source-evidence convention without restoring removed runtime machinery.
---
# SPEC

## PURPOSE

Improve how coding agents use any Kit rule and preserve evidence across investigations, forks, fresh sessions, and provider handoffs with less unnecessary context loading (GH-241). The human broadened the initial handoff request to a generalized ruleset that can improve Kit performance across rules.

## CONTEXT

- Baseline: `docs/CONSTITUTION.md` makes Markdown authoritative and native hosts responsible for planning/delegation; its Supported Command Surface rejects context routers and orchestration commands.
- The universal contract already loads contextual rules by trigger, but does not give a shared method for bounding source reads, retaining interim evidence, or detecting missing original requirements across sessions.
- `internal/templates/workflow_templates.go` has a compact living spec for canonical decisions. `docs/README.md` explicitly does not recommend or scaffold notes; existing project-owned notes remain sources, not canonical truth.
- `internal/feature/feature.go` derives state from spec metadata; the lifecycle compatibility hook is a no-op. There is no persistent inference runtime to extend.
- `cross-repository-program-coordination` requires revision evidence and live reconciliation for programs. General investigations lack an equivalent compact source-preserving convention.
- `0084-harness-subtraction` records low use and high overhead for deleted wrappers, context routing, capability catalogs, and host adapters. Existing optional embedded rules already install through `kit rules add` and preserve local customization during reconcile.
- Inspected baseline: `ff9e043b67665237aff48b0372ce7d3c8ccf7c65`; clean primary checkout; fresh default-branch fetch. No repository-local `.agents` directory exists.

## REQUIREMENTS

- One generalized rule applicable to reading/applying any Kit ruleset, source investigation, review, and context transfer; handoff is one use case.
- Keep source material revisitable with revision identities where useful; distinguish observations, inferences, uncertainty, missing inputs, and current validity.
- Reduce unnecessary reads and repeated explanation without omitting applicable safety rules, inventing facts, or prescribing a host topology.
- Reuse canonical specs/references/program ledgers; a summary or evidence index never replaces them. No mandatory transcript capture, record per tool call, new metadata schema, command, service, model call, or context router.
- External/sensitive input stays outside tracked files by default. Ignoring a file is not a privacy boundary or retention guarantee; secrets are never captured.
- Verify installation and local-edit preservation plus five deterministic handoff cases. Do not claim model understanding, speed, or cross-provider behavior from fixture success.

## ACCEPTED PLAN

1. Ship `context-evidence` as an optional, conditional embedded ruleset using the existing distribution path. Include a compact adaptable evidence/handoff example and read-only Git recipes.
2. Explain installation and triggers in the references index. Keep universal instructions, default scaffold, spec schema, and canonical project memory structure unchanged.
3. Add native Go tests for optional installation, reconciliation, source reopening, dirty/stale/missing evidence, and fresh-process recovery. Test the documented shell recipes themselves rather than inventing a production validator.
4. Run required local checks, inspect the diff, curate measured outcomes here, and deliver a ready human-assigned PR. No merge or deployment.

## DECISIONS

### Weighted comparison

Scores are predictions on a 1–5 scale, not measured model performance. Weighted score is the sum of score times percentage weight. Correctness/traceability (25%), freshness (20%), and usefulness across rule application and handoffs (20%) dominate because a cheap wrong or unrecoverable answer is not an improvement. Portability (10%), maintenance/fit with harness subtraction (10%), friction (10%), and privacy (5%) penalize host coupling, ceremony, and unnecessary source capture. Privacy remains a hard viability constraint despite its smaller ranking weight.

| Alternative | Traceability 25% | Freshness 20% | Usefulness 20% | Portability 10% | Maintenance 10% | Friction 10% | Privacy 5% | Weighted /5 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Existing conventions unchanged | 2 | 1 | 2 | 5 | 5 | 5 | 4 | 2.80 |
| Extend existing spec/notes guidance | 3 | 3 | 3 | 5 | 5 | 4 | 4 | 3.55 |
| Optional general rule with inline template | 4 | 4 | 4 | 5 | 4 | 4 | 5 | 4.15 |
| Same general rule installed by default | 4 | 4 | 4 | 5 | 4 | 3 | 5 | 4.05 |
| Lightweight evidence command | 4 | 5 | 3 | 4 | 2 | 3 | 4 | 3.70 |

Supporting evidence and limits:

- Baseline memory is durable and cheap but specs are curated, capture is optional, and general evidence has no required revision/availability fields. Program rules cover freshness only in their domain.
- Spec extensions help feature work but miss standalone investigations and applying rules; restoring notes scaffolding conflicts with the current structure. Existing notes can be used if a project already owns them.
- The optional rule serves all rules and investigations with one reusable convention. Existing distribution requires no new runtime. Manual adherence limits freshness and semantic correctness to 4; a template cannot enforce reasoning.
- Default installation increases reach, but adds guidance and reading to simple tasks without measured benefit. Optional scope does not restrict applicability: installed rule triggers cover any ruleset application where selective context/evidence handling helps.
- A command could compare revisions deterministically (freshness 5 for supported source checks), but no deterministic mechanism establishes semantic validity or captures absent input. Schema and storage maintenance, local-path portability, and data collection burden are not justified by current evidence.
- Selected the optional generalized rule (4.15). Rejected a custom RLM runtime, automatic transcript capture, a capability catalog, and new default ceremony. Retain native host search and delegation; optimize how evidence is selected and carried.

## DISCOVERIES

- The only open issue before allocation was unrelated #239. Jameson explicitly approved a new issue/worktree/branch/PR; issue #241 is assigned to `jamesonstone` and branch/worktree is `GH-241`.
- `kit spec investigation-evidence` allocated 0087 through the existing shared allocator; 0086 was already reserved elsewhere. Preserve that allocation and avoid editing another lane.
- The installed baseline binary reports a pre-existing README refresh; do not run broad reconcile or mix unrelated managed-file changes into this feature.

## VALIDATION

- PASS: `gofmt -l .` (no output), `git diff --check`, `make vet`, `make lint` (0 issues), `make build`, and `KIT_USAGE_DISABLED=1 ./bin/kit check --project`. The project check reports one non-blocking compatibility advisory for the project-owned references index.
- PASS: `env -u KIT_USAGE_DISABLED make test` across all packages. Initial run with `KIT_USAGE_DISABLED=1` failed the usage store fixture because that flag intentionally suppresses recording; diagnosed from `internal/usage/isolation.go` and reran without it. Ordinary test binaries already suppress usage; the store suite intentionally records in temporary homes. No code workaround was needed.
- PASS: six focused tests plus four source-freshness subcases. New handwritten Go test files are 124 and 156 lines, under the 300-line cap.

| Representative case | Observed deterministic result | Limit |
| --- | --- | --- |
| Normal code question | Retained code opens by full commit/path; identical working content reports SOURCE_UNCHANGED | Does not evaluate an agent's interpretation |
| Historical decision versus current implementation | Old decision and current superseding source open separately | Semantic comparison remains the agent's job |
| Stale source revision | Changed committed content and dirty content at identical HEAD both report NEEDS_REVALIDATION; old content remains readable | Unchanged content does not establish dependency or claim validity |
| Uncaptured original requirement | Fixture retains UNCAPTURED marker; outline/guidance forbid inventing the missing requirement | Cannot recover missing input or automatically infer its absence |
| Fresh-process handoff | A transferred fixture record supplies coordinates to a fresh shell, which reopens code without old conversation; missing Git object fails | Fresh process is not a fresh model session or a provider trial |

Measured overhead: default agent entry files, universal contract, and CLI surface are byte-for-byte unchanged from the baseline. Fresh-init test confirms no optional rule or rule reference appears in default instructions; optional install and local-custom preservation pass. The new rule is 809 whitespace-separated words, bounded by a 900-word regression limit, read once per session unless changed. No records are created automatically.

Source-recovery and freshness feasibility are measured. Reduced duplicate loading, improved understanding/review, token savings, and live Codex/Claude/Cursor handoff behavior remain UNMEASURED. The baseline already selects by trigger; this change supplies shared operational guidance rather than claiming a new routing improvement. No paid models, credentials, or services were used.

## OUTCOME

Implemented the generalized optional `context-evidence` rule and embedded inline outline/Git recipes through existing rule distribution, with two native Go regression files and index guidance. No default-contract or schema changes, new commands, or production runtime code. User steering broadened scope beyond handoff-only evidence into applying any Kit rule and limiting unnecessary context.

The highest predicted score remains 4.15; fixtures establish source revisiting and fail-closed content checks without mandatory bookkeeping. They do not establish model performance gains. Local sources and Git objects must remain accessible; digest-only working evidence cannot recover original content; human/agent compliance is not enforced by Kit. POSIX recipes use native Git alternatives where no shell is available. Live cross-provider testing remains unrun.

Delivery: issue #241, branch/worktree GH-241; human-author identity verified. PR publication authorized, merge/deployment withheld. Publish a draft PR for review after final checks; hosted checks are not yet observed.

## REPOSITORY MEMORY

Adopted this living spec before source edits; durable user guidance is in `docs/references/rules/context-evidence.md` and the references index. No Constitution change is needed because no project-wide invariant changes.
