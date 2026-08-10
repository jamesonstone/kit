# Kit Workflows

## Initialize a repository

Run `kit init`. Kit fetches and validates the entire configured catalog,
downloads every downstream-visible artifact, checks all target paths, and
builds one materialization plan before writing. It then installs rules,
workflows, bounded routing sections, and schema-v2 provenance. Repeating the
equivalent operation through reconciliation is idempotent.

## Resolve an agent contract

Before implementation, the coding agent runs:

```bash
kit contract resolve \
  --workflow implementation-delivery \
  --path internal/example.go \
  --applies-to backend
```

Resolution reads only `.kit.yaml` and repository-local artifacts. Explicit
feature, path, applicability, and workflow hints select conditional artifacts;
mandatory rules and workflow dependencies are included automatically. A
missing, invalid, conflicted, or unresolved required artifact returns blocked
JSON and a nonzero status.

## Implement with native agent tooling

The agent reads routing entrypoints and ordered artifacts returned by the
resolver, then uses its native planning, editing, testing, and delivery tools.
Kit does not manage that process.

## Repair asynchronous PR feedback

```bash
kit contract resolve --workflow pr-feedback-repair
```

The selected local workflow has two intake modes: bounded `await` for expected
CodeRabbit feedback and event-triggered or one-shot `collect` for late provider
or human feedback. The host or coding agent runs the single token-free `gh`
observer, preserves exact head and rate-limit evidence, and performs repair;
Kit does not wait, access GitHub, launch agents, or mutate review threads.

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
