package bootstrap

const providerRouter = `# Coding Agent Router

Read ` + "`docs/agents/README.md`" + ` first. Resolve the repository-local contract
with ` + "`kit contract resolve --json`" + ` and progressively load only the returned
workflow, rules, project references, and relevant repository evidence.

Kit resolves instructions; it does not launch or supervise agents.
`

const agentsReadmeStarter = `# Agent Routing

## Entry Point

Run ` + "`kit contract resolve --json`" + ` before implementation and whenever task
scope changes materially. Repository-local Markdown is authoritative.

## Local Guidance

- ` + "`WORKFLOWS.md`" + ` — implementation, bootstrap, and reconciliation flows.
- ` + "`GUARDRAILS.md`" + ` — safety, validation, and completion boundaries.
- ` + "`RLM.md`" + ` — progressive context loading.
- ` + "`TOOLING.md`" + ` — Kit and native-tool ownership.
- ` + "`docs/CONSTITUTION.md`" + ` — demonstrated durable project invariants.
- ` + "`docs/specs/<feature>/SPEC.md`" + ` — mandatory living history for every
  non-trivial feature.
`

const workflowsStarter = `# Agent Workflows

## Repository Bootstrap

Resolve ` + "`repository-bootstrap`" + ` after ` + "`kit init`" + `. Inspect repository
evidence progressively and populate only verified project context and commands.

## Implementation

Resolve ` + "`implementation-delivery`" + ` with ` + "`--feature <feature>`" + `, create or
adopt the complete living V3 spec from the accepted native plan, re-resolve,
then implement, validate, curate repository memory, and deliver.

## Reconciliation

Preview with ` + "`kit reconcile --json --diff`" + ` and apply only reviewed,
conflict-free changes with ` + "`kit reconcile --apply`" + `.
`

const rlmStarter = `# Progressive Context Loading

1. Start with ` + "`docs/agents/README.md`" + ` and a resolved Kit contract.
2. Read selected workflows, mandatory rules, and explicit dependencies.
3. For feature work, inspect the current spec, progress summary, explicitly
   related historical specs, and only relevant repository evidence.
4. Expand context only when a decision or newly discovered scope requires it.
5. Re-resolve with explicit path, applicability, feature, or workflow hints
   when scope changes.

Do not load the entire repository or documentation tree by default.
`

const toolingStarter = `# Agent Tooling

## Kit-Owned Operations

- ` + "`kit init`" + ` validates and materializes the repository bootstrap.
- ` + "`kit contract resolve`" + ` is local-only, deterministic, and read-only.
- ` + "`kit reconcile`" + ` previews drift and explicitly applies safe changes.
- ` + "`kit registry`" + ` and ` + "`kit rules`" + ` administer typed artifacts.

## Agent-Owned Operations

The coding agent owns repository inspection, native planning, edits, testing,
Git/GitHub delivery, review repair, and external-system tools under local rules.
`

const guardrailsStarter = `# Agent Guardrails

- Preserve user-owned work and verify the exact work lane before mutation.
- Follow resolved safety, delivery, testing, and source-size rules.
- Never infer credentials, endpoints, owners, deployment state, or project
  commands from Kit starter files.
- Keep contract resolution local-only and read-only.
- Treat a blocked feature-spec state as permission to author the spec only;
  re-resolve before source implementation.
- Report skipped, blocked, pending, unavailable, and failing validation
  literally.
- Keep the living feature spec current through validation, outcome, delivery,
  and repository-memory curation.
`
