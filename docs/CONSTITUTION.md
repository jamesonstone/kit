# CONSTITUTION

## PRINCIPLES

### Coding-Agent-First, Repository-Native

- Kit is a repository contract and evidence harness for coding agents, with bounded human-facing command adapters.
- Repository-local Markdown is authoritative. Machine-readable output is a deterministic view of that local evidence, not a second source of truth.
- Native agent planning owns research, clarification, design, and implementation planning. Kit supplies evidence and guardrails; it does not infer project truth or launch or supervise agents.
- Coding agents use `kit capabilities <command> --json` to establish command behavior and `kit context resolve --workflow <slug> --json` to select the smallest relevant local evidence set.
- A blocked context contract is an evidence gap, never permission to guess.

### Evidence Before Mutation

- Inspect repository state, durable memory, work-lane ownership, and applicable safety rules before mutation.
- Validate findings against current source and current external state before acting.
- Preserve unrelated and project-owned changes. Fail closed when ownership, target identity, or mutation scope is ambiguous.
- Report validation literally; planning evidence, local checks, hosted checks, deployment, and production proof are distinct claims.
- Pull-request merge is a distinct mutation boundary. Only a direct request or accepted bounded merge plan authorizes the exact PR set; delivery consent, checks, assignments, and ledgers never create authority.
- Only current `MERGE_READY` nodes may merge, and merge success is not deployment, runtime, production, or integrated-system proof.

### Durable Repository Memory

- `docs/specs/<feature>/SPEC.md` preserves material feature rationale, accepted plans, discoveries, validation, outcomes, and superseded decisions.
- `docs/CONSTITUTION.md` contains current project-wide invariants, not feature inventories or transient plans.
- `docs/references/` contains reusable repository-wide practices and evidence indices.
- Historical specifications remain historical evidence even when the commands or implementations they describe are retired.
- Code-and-test-sufficient changes may make a justified `Repository Memory: not required` decision.

### Small, Explicit Implementation

- Prefer the smallest complete, production-ready solution.
- Keep command handlers thin and put reusable policy, parsing, and deterministic behavior in internal packages.
- Keep the separate `git-wt` binary independent from Kit command-surface changes.
- Preserve simple data formats and bounded local state over opaque runtimes.

## CONSTRAINTS

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Use native agent planning for research, clarification, design, and implementation planning.
- Before implementation, inspect code and repository memory; create or adopt `SPEC.md` when material rationale exists.
- After validation, curate feature rationale, project invariants, reusable practices, and domain knowledge into their scope-appropriate canonical documents.
- Allow a justified `not required` repository-memory decision when code and tests preserve the complete durable truth.
- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less.
- Before delivery, audit the complete affected source/test scope; whole-project reconcile and scheduled maintenance audit the entire repository.
- Exclude documentation files, all `docs/**`, all `.kit/**`, `.kit.yaml`, ignored files, vendored dependencies, and proven generated files.
- Split oversized files by semantic responsibility while preserving stable public entry points and behavior; never use minification or arbitrary numbered chunks to claim compliance.
<!-- END KIT-MANAGED BASELINE RULES -->

### Supported v2 Command Surface

- The v2 major release intentionally keeps only these user-facing paths and their parent groups:
  - `kit init`
  - `kit spec`
  - `kit context resolve`
  - `kit usage`, `report`, `status`, `refresh`, `clear`, `enable`, and `disable`
  - `kit status`
  - `kit registry status`
  - `kit health`
  - `kit capabilities`
  - `kit config check`
  - `kit aws verify`
  - `kit check`
  - `kit pr fix`
  - `kit pr orchestrate`
  - `kit improve run`
  - `kit rules add`, `list`, `view`, and `link`
  - `kit reconcile`
  - `kit dispatch`
  - `kit instructions`
  - `kit upgrade`, `version`, and `completion`
- Removed command groups are absent, not hidden compatibility aliases.
- Legacy loop, prompt, feature-state, removed-feature, and project-refresh configuration may parse for one major release but has no runtime effect and is not added to fresh configuration.
- `kit dispatch` remains a prompt-producing adapter. It does not become an agent runtime.

### Context Resolution

- `kit context resolve` emits schema `kit.context/v1`.
- Resolution is deterministic, local-only, and read-only: no network access, writes, Git mutation, model inference, or agent launch.
- Workflows under `docs/references/workflows/` declare ordered dependencies, required rules, evidence, phases, and completion gates.
- The supported workflow set is repository bootstrap, implementation delivery, repository maintenance, PR feedback repair, pull-request merge, release orchestration, and cross-repository program coordination.
- Required missing or invalid evidence blocks resolution with a nonzero exit; optional gaps remain explicit diagnostics.
- Feature and path hints narrow evidence selection without changing canonical documents.

### Initialization, Registry, Reconciliation, and Health

- `kit init` is the canonical project bootstrap. It preserves existing project-owned content and materializes routing, references, registry-backed rules, and local workflow contracts.
- `internal/templates` remains the canonical embedded scaffold architecture and must stay synchronized with checked-in generated artifacts.
- Rules remain registry-backed, provenance-aware, and materialized under `docs/references/rules/`.
- `kit reconcile` retains its established drift-detection, preview, inclusion, merge, and safety semantics in v2.
- `kit health` retains its established maintenance interface, including the existing weekly scheduled-task behavior.
- The weekly health task reads capabilities and bounded usage analysis once per overall run after v2 is installed; its repository set, cadence, maintenance actions, no-merge rule, and existing output remain unchanged.

### Local Usage Telemetry

- Usage telemetry is local-only, best-effort, and enabled by default.
- Events contain only schema version, timestamp, normalized command path, Kit version, exit outcome, elapsed time, anonymized project identity, and interactivity.
- Never record arguments, command output, repository paths or names, file contents, environment values, secrets, or network identifiers.
- Usage commands do not record themselves.
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

- Load `docs/references/rules/testing-and-environment-validation.md` and `docs/references/testing.md` before implementation or validation.
- Preserve language-native unit and integration tests and pull-request checks; end-to-end or live integration supplements rather than replaces them.
- Keep every version-control-eligible handwritten source and test file at 300 physical lines or less.
- Run formatting, vetting, complete Go tests, race tests, linting, binary builds, release packaging, security checks, self-host validation, and affected source-size audits for a major release.

### Delivery and Release

- Issue, branch, staging, commit, push, and pull-request operations follow the repository GitHub delivery rules.
- Work happens on the exact owned writable lane. Subagents may not mutate Git or GitHub delivery state.
- A v2 release is cut only after the major-release pull request merges and the exact merged head is verified.
- Installing v2 and updating the external weekly automation are post-release activation steps, never pre-merge source changes.

## CHANGE CLASSIFICATION

### Feature Work

- Use when product behavior, architecture, public interfaces, or material rationale changes.
- Native plan, create or adopt the living spec, resolve context, implement, validate, curate memory, and deliver.

### Ad Hoc Maintenance

- Use for genuinely small fixes, security reviews, refactors, dependency updates, and mechanical maintenance whose complete durable truth is in code and tests.
- Understand, implement, validate, and report the justified repository-memory disposition.
- If the work changes consequential rationale or an existing feature contract, adopt the relevant spec instead.

## NON-GOALS

- Kit does not choose models, launch coding agents, supervise agent processes, or replace native agent planning.
- Kit does not fetch external evidence during context resolution.
- Kit does not treat generated JSON, telemetry, prompts, or agent transcripts as canonical repository memory.
- Kit does not preserve every historical CLI path after v2.
- Kit does not change `kit reconcile` semantics as part of the coding-agent-first pivot.
- Kit does not alter the separate `git-wt` binary as part of this command-surface reset.
- Kit does not execute pull-request merges or silently overwrite project-owned content; coding agents may merge only under the exact active authorization contract.

## DEFINITIONS

- **Coding-agent contract** — the ordered repository-local workflows, rules, specifications, references, and source evidence selected for a task.
- **Capability metadata** — read-only command behavior and safety information returned by `kit capabilities`.
- **Context resolution** — deterministic projection of applicable local evidence into `kit.context/v1`.
- **Workflow** — a declarative repository-local execution contract containing dependencies, rules, evidence, phases, and completion gates.
- **Ruleset** — a durable, pointer-loaded Markdown policy artifact managed through the rules registry.
- **Living spec** — a V3 `SPEC.md` maintained from accepted planning through actual outcome and repository-memory disposition.
- **Project-owned content** — repository material outside a bounded Kit-managed section or artifact contract.
- **Usage telemetry** — bounded local aggregateable command events with no arguments, content, secrets, or network transport.
- **Weekly health boundary** — the existing scheduled maintenance interface whose behavior remains stable while adding one overall usage analysis.
