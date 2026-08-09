# Coding-Agent-First Major Migration

This release intentionally removes the broad Kit 1.x product and CLI model.
Projects move to schema v2 through `kit reconcile`; no dual-mode runtime is
retained.

## Before upgrading

1. Commit or otherwise preserve project-owned changes.
2. Install the new major version.
3. Run the migration preview from the project root:

   ```bash
   kit reconcile --json --diff
   ```

4. Review every artifact state, routing-section diff, and diagnostic.
5. Apply the conflict-free migration explicitly:

   ```bash
   kit reconcile --apply
   ```

6. Resolve the local agent contract:

   ```bash
   kit contract resolve --workflow implementation-delivery
   ```

## What migration changes

- `.kit.yaml` becomes schema v2 and records the configured catalog source and
  typed artifact provenance.
- Existing managed rules are adopted with current source and section hashes.
- Modified rules remain `local-custom`.
- New project-visible workflows are installed.
- Agent-routing instructions are added or updated only inside bounded managed
  sections; surrounding text remains project-owned.
- Retired managed artifacts are removed only when their installed content is
  unchanged. Customized retired files are preserved.

## Conflict handling

Edits to different Markdown sections merge. Divergent changes to the same
section are reported without writing that artifact. Resolve the content
manually or explicitly replace one artifact with registry content:

```bash
kit reconcile --accept-registry ruleset/<slug> --apply
```

Acceptance is exact and repeatable; it never implies acceptance of another
artifact.

## Protected command paths

Only these existing names and core purposes are protected:

- `kit init`
- `kit reconcile`
- `kit rules add`
- `kit rules list`
- `kit rules view`
- `kit registry status`

Flags, text output, and incidental Kit 1.x behavior are not compatibility
contracts.

## Removed surfaces

Feature/spec lifecycle, prompt generation, loop and dispatch execution, review,
CI and pull-request helpers, health, capabilities, config, AWS, status and map,
verification and replay, state, improvement, scaffold, skill, and legacy
commands have been removed. Use the coding agent's native tools and
repository-owned automation for those responsibilities.

The separate `git-wt` executable is unchanged.
