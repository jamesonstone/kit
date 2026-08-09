# CONSTITUTION

Kit is a provider-neutral, repository-native contract layer for coding agents.
It materializes registry-backed Markdown and resolves a deterministic local
contract without launching, supervising, or reasoning on behalf of an agent.

## PRINCIPLES

### 1. Repository-local truth

- Materialized Markdown is canonical and directly reviewable.
- Resolved JSON is a reproducible machine projection, never a competing source
  of truth.
- Durable product rationale belongs in canonical repository documentation when
  code and tests cannot preserve it.

### 2. Coding-agent-first interfaces

- Agent routing points to `kit contract resolve` and local artifacts.
- Selection uses explicit feature, path, applicability, and workflow hints.
- Workflows declare execution phases and gates but do not manage processes.
- Human commands exist to initialize, inspect, install, and reconcile the
  contract.

### 3. Provider neutrality

- Catalog and project schemas do not encode one coding-agent vendor.
- Rules and workflows use Markdown plus YAML front matter.
- Routing updates are bounded and preserve provider-specific project content.
- Registry source access is isolated behind a provider interface.

### 4. Determinism and explicit state

- Versions, digests, source revisions, paths, section hashes, and states are
  visible in `.kit.yaml`.
- Contract resolution performs no network calls, writes, Git operations, model
  calls, or task-text inference.
- Invalid or unresolved required artifacts block the contract with actionable
  diagnostics.

### 5. Fail-closed reconciliation

- Reconciliation is read-only by default.
- The complete catalog and plan are validated before writes.
- Local-only edits become `local-custom`.
- Disjoint Markdown-section edits merge; same-section divergence conflicts.
- Registry content replaces customization only through an exact targeted
  acceptance.
- Managed routing changes only inside marker-delimited sections.

### 6. Narrow product boundary

- Kit owns artifact distribution, provenance, resolution, and reconciliation.
- Coding agents own planning, implementation, validation, and delivery through
  repository-native tools.
- `git-wt` is a separate worktree convenience and not part of the contract
  runtime.

## CONSTRAINTS

### Kit-managed baseline rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Run `kit contract resolve` before implementation and when scope changes.
- Treat returned repository-local rulesets and workflows as authoritative.
- Keep routing, registry provenance, and local artifacts synchronized through
  `kit init` and `kit reconcile`.
- Preserve intentional local customization and stop on unresolved conflicts.
- Keep every version-control-eligible handwritten implementation and test file
  at 300 physical lines or less.
- Validate affected code and curate material repository rationale after
  implementation.
<!-- END KIT-MANAGED BASELINE RULES -->

### Registry invariants

1. Catalog entries have a supported kind, unique kind/slug key, safe relative
   source and target paths, positive version, digest, read policy, and valid
   dependencies.
2. Catalog dependency graphs are complete and acyclic.
3. Fresh initialization installs every downstream-visible artifact.
4. Schema-v2 is the only runtime project schema; v1 is only a reconcile input.
5. Repeated initialization or reconciliation against unchanged state is
   idempotent.

### Implementation invariants

1. `cmd/kit` is a thin entrypoint; human and machine CLI adapters live in
   `pkg/agentcli`.
2. Registry fetching, validation, provenance, merge, planning, and application
   live in `internal/registry`.
3. Local-only selection and resolved-contract projection live in
   `internal/contract`.
4. Handwritten source and test files do not exceed 300 physical lines.
5. Errors include context and never silently continue after partial state.
6. `cmd/git-wt` and `internal/worktree` remain independent of the Kit contract
   runtime.

### Public CLI allowlist

The `kit` command tree is limited to:

- `init`
- `reconcile`
- `contract resolve`
- `registry add|list|status|view`
- `rules add|list|view`

## NON-GOALS

Kit does not:

- launch, dispatch, supervise, evaluate, or retry coding agents;
- infer applicability from natural-language task text;
- generate prompts or manage feature/spec lifecycles;
- run CI, review, pull-request, cloud, or deployment workflows;
- maintain hidden databases, remote state, credentials, or telemetry;
- overwrite project customization by default;
- retain broad Kit 1.x CLI compatibility.

## DEFINITIONS

- **Artifact**: A typed catalog entry and its Markdown content.
- **Ruleset**: Declarative constraints and verification for applicable work.
- **Workflow**: Declarative phases, gates, dependencies, and completion
  expectations for a coding agent.
- **Resolved contract**: Versioned JSON selecting ordered local artifacts and
  explaining applicability, provenance, state, diagnostics, and next actions.
- **Managed**: Local content matches the tracked installed state.
- **Local-custom**: Valid repository-owned content differs intentionally.
- **Conflict**: Local and registry edits diverge in the same Markdown section.
- **Routing section**: Marker-bounded provider instruction text owned by Kit.
