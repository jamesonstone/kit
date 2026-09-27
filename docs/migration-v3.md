# Migrating to Kit v3

Kit v3 makes subagent orchestration capability-aware and changes the public Go
module identity to `github.com/jamesonstone/kit/v3`. Repository data and
historical specifications remain compatible; the breaking boundary is the CLI,
generated guidance, and Go installation/import surface.

## Before Upgrading

1. Commit or otherwise preserve project-owned work.
2. Record the current binary version with `kit version`.
3. Preview managed guidance changes:

   ```bash
   kit reconcile --include-files --dry-run --diff
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

- Remove `--max-subagents` from `kit dispatch` and `kit pr fix` invocations.
  The flag is now unknown rather than deprecated.
- Keep `--single-agent` when one supervisor lane is explicitly required.
- Let the active coding-agent runtime report and own model availability,
  continuation, waiting, parallel scheduling, and capacity.

## Generated Agent Guidance

Kit retains exactly three default instruction targets: `AGENTS.md`,
`CLAUDE.md`, and `.github/copilot-instructions.md`. It does not generate
`WARP.md` or provider agent-definition directories.

All three files render the same Kit-managed universal contract block from one
canonical source, with no vendor-specific bindings. Specialized rules under
`docs/references/rules/` load only when the contract's trigger for them applies.
Instruction files that predate the managed block are left unchanged by refresh
and reported by `kit reconcile` until you preview and apply a replacement.

Apply reviewed managed updates only after the preview is understood:

```bash
kit reconcile --include-files
kit check --project
```

## Compatibility Evidence

A host may degrade to sequential actual children or serialized logical lanes
when it cannot confirm richer controls. That degradation must be reported
literally. Logical roles, plans, task lists, and separate conversations do not
become actual subagents unless the host creates a child execution and returns a
separate result.

Historical v1 and v2 specifications remain evidence and must not be rewritten
mechanically. See [Kit v3.0.0](releases/v3.0.0.md) for the release boundary.
