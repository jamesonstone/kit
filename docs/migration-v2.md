# Migration To Kit 2

Kit 2 is an intentional command-surface reset. It preserves repository files,
historical specifications, `kit reconcile` behavior, dispatch, and the weekly
health maintenance contract while removing duplicative command families.

## Before Upgrading

1. Commit or otherwise account for project-owned changes.
2. Record the current Kit version and read the removed-command list in
   [commands.md](commands.md).
3. Do not delete historical feature documents or downstream `docs/notes`
   content merely because Kit no longer scaffolds or routes to notes.

## Upgrade And Preview

```bash
kit upgrade
kit version
kit capabilities --json
kit reconcile --include-files --dry-run --diff
```

Review every planned managed-file change. Apply only when it preserves local
customization:

```bash
kit reconcile --include-files
kit check --project
kit context resolve --workflow repository-maintenance --json
```

`kit reconcile` has not been redesigned in this release. Existing flags and
safe merge behavior remain the maintenance boundary.

## Replace Removed Commands

| Former need | Version 2 path |
| --- | --- |
| command selection | `kit capabilities <command> --json` |
| agent workflow and evidence | `kit context resolve --workflow <slug> --json` |
| native feature planning | host-agent planning plus `kit spec` when durable rationale is required |
| multi-lane prompt | `kit dispatch` |
| PR feedback prompt | `kit pr fix` |
| managed drift | `kit reconcile` |
| scheduled maintenance | `kit health` plus one overall `kit usage` analysis |

There is no replacement command for notes, project refresh, the prompt library,
or executable loop/runtime commands. Put durable feature rationale in `SPEC.md`,
reusable evidence in references, and project-wide invariants in the
Constitution. Use native agent and repository tools for transient work.

## Usage Data

Collection is local, bounded, and enabled by default. Inspect or opt out before
ordinary use if desired:

```bash
kit usage status --json
kit usage disable --global
kit usage clear --all --yes
```

A global disable is absolute. Project preference lives in `.kit.yaml`. No
network export exists.

## Weekly Health Activation

After v2.0.0 is released and installed, update the weekly task once to retain
its existing repository scan and add these overall-run reads:

```bash
kit capabilities usage --json
kit usage status --json
kit usage report --since 90d --json
```

Do not update the automation before the released v2 binary is installed.
