# Kit Workflows

## Initialize a repository

Run `kit init`. Kit fetches and validates the entire configured catalog,
downloads every downstream-visible artifact, checks all target paths, and
builds one materialization plan before writing. It then installs safe
environment, GitHub, repository-memory, RLM, and reference starters alongside
rules, workflows, bounded sections, and schema-v2 provenance. Repeated init is
idempotent and preserves project-owned files.

After materialization, give the emitted prompt to a coding agent. It begins
with:

```bash
kit contract resolve --workflow repository-bootstrap --json
```

The agent progressively inspects repository evidence and populates only
verified project context and commands. Kit never reads `.env`, infers project
truth, launches an agent, or grants direnv trust.

## Resolve an agent contract

Before implementation, the coding agent runs:

```bash
kit contract resolve \
  --workflow implementation-delivery \
  --feature 0059-example \
  --path internal/example.go \
  --applies-to backend
```

Resolution reads only `.kit.yaml` and repository-local artifacts. Explicit
feature, path, applicability, and workflow hints select conditional artifacts;
mandatory rules and workflow dependencies are included automatically. A
missing, invalid, conflicted, or unresolved required artifact returns blocked
JSON and a nonzero status.

For a non-trivial feature, resolution also reports the canonical living V3
spec path, required and missing sections, related historical specs, RLM
indices, and phase permissions. A missing or incomplete spec permits authoring
that spec but blocks source implementation. The agent writes it from the
accepted native plan and actual evidence, re-resolves before source edits,
keeps decisions and discoveries live, and reconciles validation and outcome
against the integrated diff before delivery.

## Implement with native agent tooling

The agent reads routing entrypoints and ordered artifacts returned by the
resolver, then uses its native planning, editing, testing, and delivery tools.
Kit does not manage that process.

Historical specifications remain under `docs/specs/` and are loaded only when
the progress summary, explicit relationships, or repository evidence makes
them relevant. Removed commands or conventions are not mechanically rewritten
into current behavior.

## Repair asynchronous PR feedback

```bash
kit pr fix --pr owner/repo#123 --output-only
```

The protected adapter resolves the local workflow and offers two intake modes:
bounded `--wait` for expected CodeRabbit feedback and immediate one-shot
collection for late provider or human feedback. It runs a token-free,
rate-bounded `gh` observer, preserves exact head and lane evidence, and emits
the agent prompt. Kit does not launch the agent or perform repairs or delivery.
Thread resolution requires a separate explicit `--resolve --head --thread
--yes` invocation after pushed-head verification.

`SUCCESS / Review completed` and `SUCCESS / Review skipped: <reason>` are
different terminal outcomes. Pending, skipped, timeout, unavailable, provider
failure, head change, and rate limit are never inferred to be clean review.
See the [workflow contract](references/workflows/pr-feedback-repair.md).

## Reconcile drift

```bash
kit reconcile --json --diff       # preview; no writes
kit reconcile --apply             # apply conflict-free changes
```

Reconciliation compares the last installed registry sections, current local
Markdown, and the configured registry revision:

- Remote-only changes update managed content.
- Local-only changes become `local-custom`.
- Changes to different Markdown sections merge while preserving customization.
- Same-section divergence becomes `conflict` and is not written.
- Missing managed content is restored.
- Retired managed artifacts are removed only when unchanged; customized ones
  remain local.
- Routing regenerates only the bounded Kit-managed section.

Replacing customization requires the exact key:

```bash
kit reconcile --accept-registry ruleset/example --apply
```

## Migrate schema-v1 projects

Schema migration is part of the normal reconciliation plan. Preview it first,
review artifact classifications and routing diffs, then apply explicitly.
There is no schema-v1 runtime or broad legacy CLI after migration. See the
[migration guide](migrations/coding-agent-first-major.md).
