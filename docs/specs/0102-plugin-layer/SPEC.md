---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0102"
  slug: "plugin-layer"
  dir: "0102-plugin-layer"
---
# SPEC

## PURPOSE

Let a project opt in to a plugin whose only effect is an agent-instruction block Kit can add, update, and remove. With no plugin enabled, generated instruction files stay byte-identical to a project that has none.

## CONTEXT

Issue GH-296. Spec 0092 keeps Kit offline: no credentials, no provider SDKs, and no ticket API client. Native agents own ticket API work through their own connectors. Issue tracking stays Layer 1 (`issues.provider` omitted or `github`, or `linear`). A plugin must not become a second issue provider or a network client.

Two later repositories will follow this contract: `jamesonstone/kit-plugin-template` and `appliedsymbolics/layer2context`. The context plugin's user-edited YAML holds a Linear workspace, team, and project, and it commits no secrets.

## REQUIREMENTS

- Plugins are off by default. An omitted `plugins` key enables nothing.
- Discovery is only `.kit/plugins/<id>/kit-plugin.yaml`. Kit does not scan the home directory, clone a repository, or use the network.
- The manifest names an instruction file and may name a config file. Kit copies the instruction text into each agent entry file and does not read the config file.
- `plugins.enabled` lists ids. `kit init --plugin` and `kit config set --plugin --enable|--disable` toggle that list. `kit reconcile` adds, updates, and removes the blocks idempotently.
- Disabling a plugin removes its block and the blank line Kit added. Project text outside the block stays byte-identical.
- Layer 1 issue configuration is unchanged.

Verified when: a project with no enabled plugin has the same agent-entry bytes as today; enabling a fixture plugin adds its block and config path comment; disabling it removes that block; full Go tests, vet, build, lint, format, and `kit check --project` pass. Out of scope: plugin code execution, credential storage, config-value interpolation, and the two follow-up repositories.

## ACCEPTED PLAN

Add a small `internal/plugin` package that loads an enabled manifest and reconciles a marked region immediately after the universal-contract block. Thread that through the existing instruction-file plan so init and reconcile share it. Patch `.kit.yaml` the same way issue settings are patched, without rewriting unrelated keys. Document the contract in `docs/references/plugins.md`.

## DECISIONS

- Project-local `.kit/plugins/<id>/` is the only discovery path. A home-directory or network search would hide the dependency and break hermetic tests.
- Enablement is `plugins.enabled`, a list of ids. An `enabled: false` map would leave a disabled plugin in config. Removing the last id removes the `plugins` key.
- Kit does not read or interpolate plugin config. The block contains `<!-- Plugin config: .kit/plugins/<id>/<file> -->` only when the manifest sets `config`. The agent reads that file. Copying values into agent files would commit them and change the block on every edit.
- The universal contract text is unchanged. Plugin bytes exist only in the separate region, so a project with no plugin matches today's generator.
- Plugin code, SDKs, and credentials are out of scope. Rejected: executing a plugin, embedding a provider client, and treating a plugin as an issue provider.
- Unknown manifest fields are ignored so a template may carry a human description Kit does not interpret.
- Instruction files are LF text of at most 64 KiB. Symlinks and paths that leave the plugin directory are rejected.

## DISCOVERIES

Instruction entry planning is the single write path for init and reconcile. Applying plugins there keeps health, status, and reconcile on the same bytes. `internal/config/schema.go` is already at the 300-line limit, so plugin YAML edits live in their own file and reuse the existing node helpers.

A file that ends on the contract marker, with no newline, keeps that ending when the plugin block is removed. A newline that already ended the marker line stays. `git worktree add` does not populate a submodule, so reconcile copies a missing plugin directory from the checkout that already has it and does not fetch it. Config set reads the config file once and plans from those bytes. A YAML merge that supplies `plugins.enabled` is overridden with an explicit empty list. An alias or anchor is rejected. Issue-mode JSON omits `plugins`.

## VALIDATION

Observed on this branch before delivery:

- `gofmt -l .` printed nothing.
- `make vet` exited 0.
- `make test` exited 0. Fixture coverage checks byte-identical entry files with no plugin, including project text after the contract; reconcile adds, updates, and removes a fixture block; config set preserves Linear issue settings and a comment; init `--plugin` writes the block.
- `make build` exited 0.
- `make lint` (`golangci-lint` 2.11.0) reported 0 issues.
- `./bin/kit check --project` exited 0. It reported one existing non-blocking advisory that `docs/references/README.md` is a retired Kit file kept because it is edited. That advisory was already non-blocking before this change.

Hosted CI was not run locally. The follow-up repositories were not built.

Review fixes on the same branch were checked the same way: `gofmt -l .` printed nothing, and `make vet`, `make test`, `make build`, `make lint`, and `./bin/kit check --project` exited 0. The README advisory was unchanged.

## OUTCOME

GH-296 adds discovery at `.kit/plugins/<id>/kit-plugin.yaml`, `plugins.enabled`, and reconcile of one instruction block per enabled plugin after the universal contract. No plugin code, credentials, provider SDK, or network call was added. With no plugin enabled, generated entry files match the current contract renderer. Issue provider behavior is unchanged.

## REPOSITORY MEMORY

This spec is the contract's rationale. `docs/references/plugins.md` is the usage contract for later repositories. The Constitution records the offline plugin boundary. The universal contract is intentionally unchanged.
