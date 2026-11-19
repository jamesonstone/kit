# Plugins

A plugin adds one instruction block to the agent entry files. It is off until the project lists its id. Kit does not run plugin code, read the plugin config file, store credentials, or use the network. Issue tracking stays on `issues.provider` (GitHub by default, or Linear).

## Discovery

Kit loads an enabled plugin only from:

```text
.kit/plugins/<id>/kit-plugin.yaml
```

`<id>` matches `^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$` and must equal the directory name. Copy or submodule the plugin repository so its root is that directory. Kit does not clone it and does not look in any other directory. A symlink is rejected.

## Manifest

`kit-plugin.yaml`:

```yaml
schema_version: 1
id: layer2context
instructions: instructions.md
config: config.yaml
```

- `schema_version` is `1`.
- `id` matches the directory name.
- `instructions` is a UTF-8 LF file inside the plugin directory, at most 64 KiB. Kit trims surrounding whitespace. The text must not contain Kit managed markers (`KIT-MANAGED CONTRACT` or `KIT-PLUGIN`).
- `config` is optional. It is a relative path inside the plugin directory. Kit does not create, open, or parse that file. The file may be absent. Unknown manifest fields are ignored.

Do not put secrets in the config file. Kit will not read them, and the agent that does read the file should not need them there.

## Enable

```yaml
plugins:
  enabled:
    - layer2context
```

Omitted `plugins` means no plugins. Order in the list is the order of the blocks.

```bash
kit init --plugin layer2context
kit config set --plugin layer2context --enable
kit config set --plugin layer2context --disable
kit reconcile
```

`kit init --plugin` requires the directory to exist already. Repeat the flag for more than one plugin. `kit config set` toggles one id and does not change the issue provider. `--disable` removes the id. When the list is empty, Kit removes the `plugins` key. Pass `--dry-run` to preview the config edit. Then run `kit reconcile`.

## Reconciled block

`kit reconcile` writes the same block into `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md`, immediately after the Kit-managed contract and before project text:

```text
<!-- BEGIN KIT-PLUGIN layer2context -->
<!-- Plugin config: .kit/plugins/layer2context/config.yaml -->
instruction text
<!-- END KIT-PLUGIN layer2context -->
```

The config comment is present only when the manifest sets `config`. The path uses forward slashes and is relative to the project root. That comment is how config is passed through: the agent reads the file itself. Kit does not substitute values.

A second reconcile with the same plugin bytes changes nothing. Editing the instruction file and reconciling updates the block. Kit overwrites edits made inside the block. Text outside the block is kept.

Disabling the plugin and reconciling removes the block and the blank line Kit inserted. The entry file matches its previous bytes. A plugin that is present on disk but not listed leaves no block.

## Config that Kit does not interpret

A context plugin can keep a Linear workspace, team, and project in the file named by `config`. These keys are the plugin's, not Kit's:

```yaml
workspace: example
team: ENG
project: whos-doing-what
```

The plugin's instruction text tells the agent to read that file and to use the agent's own Linear connector. Kit stores no Linear credentials and adds no Linear SDK.

## Template

A template repository's root contains `kit-plugin.yaml`, the instruction file it names, and optionally the config file. Kit picks it up when that root is placed at `.kit/plugins/<id>/` and the id is enabled.
