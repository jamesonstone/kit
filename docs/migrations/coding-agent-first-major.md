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
   kit contract resolve --json
   ```

7. Restore any missing canonical bootstrap starters and receive the
   evidence-gathering prompt:

   ```bash
   kit init
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
- The retired `feature-notes` ruleset is replaced by mandatory
  `feature-specification`. Historical specs and project-owned content are not
  mechanically rewritten or deleted.

## Feature-history correction

The initial coding-agent-first implementation made feature specs advisory.
That was incomplete. Every feature must now resolve with
`--work-type feature` and an explicit feature directory, then create or adopt a complete living V3
`docs/specs/<feature>/SPEC.md` before source edits. Missing or incomplete specs
allow spec authoring but block source implementation until re-resolution.
Genuinely mechanical maintenance must use `--work-type maintenance`; omitted
classification is blocked rather than treated as an exemption.
Fresh init still creates only the empty `docs/specs/` container and never
invents a feature or restores `0000_INIT_PROJECT.md` downstream.

## Canonical init correction

The initial Kit 2.0 implementation materialized only registry artifacts and
routing. That was incomplete. `kit init` again owns the full safe bootstrap:
environment and Makefile starters, CodeRabbit and GitHub support, bounded
README and Constitution sections, `docs/PROJECT_PROGRESS_SUMMARY.md`, the
complete RLM support tree, project references, rules, workflows, and
provenance. It preserves existing content and emits the
`repository-bootstrap` coding-agent prompt. It does not restore removed
feature-lifecycle commands or files.

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
- `kit pr fix`

Flags, text output, and incidental Kit 1.x behavior are not compatibility
contracts.

## Removed surfaces

Feature/spec lifecycle, general prompt generation, loop and dispatch execution,
the broader review/CI/pull-request family, health, capabilities, config, AWS, status and map,
verification and replay, state, improvement, scaffold, skill, and legacy
commands have been removed. Use the coding agent's native tools and
repository-owned automation for those responsibilities.

`kit pr fix` is the sole restored PR helper. It collects active feedback,
prepares the exact writable PR-head lane, and renders the registry-backed
`pr-feedback-repair` supervisor prompt. It does not launch agents or perform
repairs or delivery. Contract resolution itself still performs no GitHub
access. No dispatch, loop, or broader legacy PR family is restored.

The separate `git-wt` executable is unchanged.
