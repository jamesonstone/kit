# Migrating to Kit v3

Kit v3 makes subagent orchestration capability-aware and changes the public Go
module identity to `github.com/jamesonstone/kit/v3`. Repository data and
historical specifications remain compatible; the breaking boundary is the CLI,
generated guidance, and Go installation/import surface.

## Before Upgrading

1. Commit or otherwise preserve project-owned work.
2. Record the current binary version with `kit version`.
3. Preview the migration:

   ```bash
   kit reconcile --dry-run --diff
   ```

4. Check scripts and automation for `--max-subagents` and the old unversioned
   Go install path.

## Install Or Upgrade

The preferred path is the checksum-verified GitHub Release flow used by the
existing binary:

```bash
kit upgrade
kit version
```

For a direct Go installation, use the major-version module path:

```bash
go install github.com/jamesonstone/kit/v3/cmd/kit@latest
kit version
```

The repository and GitHub API URLs remain unversioned. Only Go module imports
and install paths gain `/v3`.

## CLI Changes

- Removed commands: `kit context resolve`, `kit capabilities`, `kit dispatch`,
  `kit pr fix`, `kit pr orchestrate`, `kit improve run`, `kit instructions`,
  and `kit rules link`, plus the `--profile`, `--single-agent`, and
  `--max-subagents` flags. Agents use their host's tools, delegation, and
  `kit <command> --help` instead.
- `kit init --refresh` is removed: `kit init` creates new projects only, and
  `kit reconcile` updates existing ones. `kit health` no longer writes; it
  reports what `kit reconcile` would change.
- Rules ship inside the binary; upgrade Kit to pick up rule changes. Nothing is
  fetched from GitHub.
- `docs/agents/`, `docs/references/workflows/`, the other retired support
  documents, `docs/PROJECT_PROGRESS_SUMMARY.md`, and retired rules are no longer
  generated. `kit reconcile` removes copies that are exactly as a Kit release
  generated them and committed to Git, and reports edited copies it keeps.
- The 300-line source limit applies only when `.kit.yaml` sets
  `source_file_line_limit`.

## Generated Agent Guidance

Kit retains exactly three default instruction targets: `AGENTS.md`,
`CLAUDE.md`, and `.github/copilot-instructions.md`. It does not generate
`WARP.md` or provider agent-definition directories.

All three files render the same Kit-managed universal contract block from one
canonical source, with no vendor-specific bindings. Specialized rules under
`docs/references/rules/` load only when the contract's trigger for them applies.
`kit reconcile` migrates instruction files that predate the managed block:
Kit-generated sections are replaced by the block and sections Kit never wrote
are kept below it. When a Kit-generated section was edited, reconcile leaves the
file unchanged and names the section; review it, then run
`kit reconcile --force`, which replaces Kit's sections and still keeps project
sections.

```bash
kit reconcile
kit check --project
```

Run from the primary checkout, reconcile writes to a `kit-reconcile` linked
worktree and prints how to review, commit, and open a pull request.

## Compatibility Evidence

A host may degrade to sequential actual children or serialized logical lanes
when it cannot confirm richer controls. That degradation must be reported
literally. Logical roles, plans, task lists, and separate conversations do not
become actual subagents unless the host creates a child execution and returns a
separate result.

Historical v1 and v2 specifications remain evidence and must not be rewritten
mechanically. See [Kit v3.0.0](releases/v3.0.0.md) for the release boundary.
