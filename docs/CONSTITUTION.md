# CONSTITUTION

## PRINCIPLES

### Coding-Agent-First, Repository-Native

- Kit is a small, vendor-neutral harness for coding agents: a universal contract, contextual rules, project memory, and deterministic commands.
- Repository-local Markdown is authoritative. Machine-readable output is a deterministic view of that local evidence, not a second source of truth.
- Native agent planning owns research, clarification, design, and implementation planning. Kit supplies evidence and guardrails; it does not infer project truth or launch or supervise agents.
- Kit owns one canonical universal agent contract (`internal/templates/universal_contract.md`). Agent entry files render it deterministically inside a Kit-managed block; contextual rules load only when their trigger applies. Never restate universal rules in other Kit-owned documents.
- Agents use their host's native delegation. One primary agent owns integration and reporting, and reports distinguish actual separate agents from logical lanes and self-review.

### Evidence Before Mutation

- Inspect repository state, durable memory, work-lane ownership, and applicable safety rules before mutation.
- Delivery, merge, deployment, infrastructure, deletion, and authorship boundaries are defined once in the universal contract and its contextual rules.
- Validate findings against current source and current external state before acting.
- Preserve unrelated and project-owned changes. Fail closed when ownership, target identity, or mutation scope is ambiguous.

### Durable Repository Memory

- `docs/specs/<feature>/SPEC.md` preserves material feature rationale, accepted plans, discoveries, validation, outcomes, and superseded decisions.
- `docs/CONSTITUTION.md` contains current project-wide invariants, not feature inventories or transient plans.
- `docs/references/` contains reusable repository-wide practices and evidence indices.
- Historical specifications remain historical evidence even when the commands or implementations they describe are retired.
- Code-and-test-sufficient changes may make a justified `Repository Memory: not required` decision.

### Small, Explicit Implementation

- Prefer the smallest complete, production-ready solution.
- Keep command handlers thin and put reusable policy, parsing, and deterministic behavior in internal packages.
- Keep target-aware worktree preparation internal to Kit; generic script distribution belongs outside Kit.
- Preserve simple data formats and bounded local state over opaque runtimes.

## CONSTRAINTS

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Kit's universal agent rules live in the Kit-managed block of `AGENTS.md` (rendered identically into `CLAUDE.md` and `.github/copilot-instructions.md`), and contextual rules live in `docs/references/rules/`. This Constitution records project-specific invariants and does not restate them.
<!-- END KIT-MANAGED BASELINE RULES -->

### Supported Command Surface

- User-facing paths: `kit init`, `kit spec`, `kit rules add|list|view`, `kit check`, `kit reconcile`, `kit health`, `kit registry status`, `kit status`, `kit config check`, `kit aws verify`, `kit usage report|status|refresh|clear|enable|disable`, `kit upgrade`, `kit version`, and `kit completion`.
- A command exists only when it does deterministic work an agent cannot reproduce as safely or cheaply: bootstrap and managed-file convergence, validation, spec allocation, AWS identity verification, or usage evidence. Kit adds no prompt wrappers, context routers, capability catalogs, or orchestration commands.
- Removed command groups are absent, not hidden compatibility aliases.

### Rules, Initialization, Reconciliation, and Health

- Rules ship inside the released binary (`rules.go` embeds `docs/references/rules/`); Kit never fetches rules or instructions at runtime. Unmodified installed rules update on refresh, locally edited rules are preserved and reported, and optional rules install only through `kit rules add`.
- Retired rules stay listed in `pkg/cli/rules_retired.go` so historical references and legacy installs remain recognizable.
- `kit init` scaffolds the universal contract, the Constitution starter, `docs/references/testing.md`, the core rules, and the developer-experience starter files, preserving existing project-owned content.
- `internal/templates/universal_contract.md` is the single source of the agent contract; checked-in entry files must match it.
- `kit reconcile` retains its drift-detection, preview, inclusion, and primary-checkout deferral semantics until the final migration redesign.
- `kit health` is the scheduled-maintenance entry point: safe managed updates, then the project check.
- `.kit.yaml` `source_file_line_limit` makes a line limit a deterministic project invariant; Kit sets 300 for itself.

### Local Usage Telemetry

- Usage telemetry is local-only, best-effort, and enabled by default.
- Events contain only schema version, timestamp, normalized command path, Kit version, exit outcome, elapsed time, anonymized project identity, and interactivity.
- Never record arguments, command output, repository paths or names, file contents, environment values, secrets, or network identifiers.
- Usage commands, `--help` lookups, development builds, test binaries, and processes with `KIT_USAGE_DISABLED=1` do not record.
- A global disable is absolute. A project may opt out but cannot override a global disable.
- Retain at most 365 days, 16 MiB total, and 2 MiB per JSONL shard. Maintenance prunes complete oldest shards rather than partially truncating one.
- `kit usage refresh`, `clear`, `enable`, and `disable` are the only maintenance and control surfaces for usage data.

### Repository Memory Lifecycle

- Create or adopt a living V3 feature spec before source edits when consequential rationale would otherwise be lost.
- Keep accepted decisions and discoveries current while implementing.
- Reconcile validation results and the actual integrated outcome before completion.
- After validation, curate durable project-wide truth through `docs/references/rules/constitution-curation.md`.
- Do not mechanically rewrite historical specifications to match current command names.

### Testing and Source Size

- Kit's validation commands live in `docs/references/testing.md`; pull-request CI runs formatting, vet, tests, build, lint, and `kit check --project`.
- Handwritten Go source and test files stay at 300 physical lines or less (enforced by `kit check --project`).
- Run formatting, vetting, complete Go tests, race tests, linting, binary builds, release packaging, security checks, self-host validation, and affected source-size audits for a major release.

### Delivery and Release

- Release quality gates run before tag creation. Mint owns immutable tag and
  GitHub Release state; Kit retains exact version selection, GoReleaser builds
  and checksums, and idempotent artifact upload.
- Release verification must tie the exact merged source to the tag, hosted
  workflow, GitHub Release, artifacts, checksums, and installed binary; none of
  those claims is inferred from a merge or local build.

## CHANGE CLASSIFICATION

### Feature Work

- Use when product behavior, architecture, public interfaces, or material rationale changes.
- Native plan, create or adopt the living spec, resolve context, implement, validate, curate memory, and deliver.

### Ad Hoc Maintenance

- Use for genuinely small fixes, security reviews, refactors, dependency updates, and mechanical maintenance whose complete durable truth is in code and tests.
- Understand, implement, validate, and report the justified repository-memory disposition.
- If the work changes consequential rationale or an existing feature contract, adopt the relevant spec instead.

## NON-GOALS

- Kit does not choose models, prescribe a planning or delegation lifecycle, launch coding agents, supervise agent processes, or replace native agent planning.
- Kit does not fetch rules, instructions, or evidence from the network at runtime.
- Kit does not treat generated JSON, telemetry, prompts, or agent transcripts as canonical repository memory.
- Kit does not preserve every historical CLI path across major releases.
- Kit does not change `kit reconcile` semantics as part of the coding-agent-first pivot.
- Kit does not execute pull-request merges or silently overwrite project-owned content; coding agents may merge only under the exact active authorization contract.

## DEFINITIONS

- **Universal contract** — Kit's single canonical agent contract, rendered into the Kit-managed block of each agent entry file.
- **Contextual rule** — a ruleset under `docs/references/rules/` that an agent reads only when the contract's trigger for it applies.
- **Ruleset** — a durable Markdown policy artifact managed through the rules registry.
- **Living spec** — a V3 `SPEC.md` maintained from accepted planning through actual outcome and repository-memory disposition.
- **Project-owned content** — repository material outside a bounded Kit-managed section or artifact contract.
- **Usage telemetry** — bounded local aggregateable command events with no arguments, content, secrets, or network transport.
- **Weekly health boundary** — the existing scheduled maintenance interface whose behavior remains stable while adding one overall usage analysis.
