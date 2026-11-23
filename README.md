```text
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║
██╔═██╗ ██║   ██║
██║  ██╗██║   ██║
╚═╝  ╚═╝╚═╝   ╚═╝

              repository-local contract for coding agents
```

Kit is a provider-neutral, repository-local contract for coding agents. It
installs one universal agent contract and contextual rules, keeps project memory
(Constitution, living specs, testing reference), validates it, and migrates any
existing Kit project to the current structure without touching the primary
checkout. Kit does not infer project truth, choose a model, or launch or
supervise agents.

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/kit)](https://github.com/jamesonstone/kit/commits) [![Open issues](https://img.shields.io/github/issues/jamesonstone/kit)](https://github.com/jamesonstone/kit/issues) [![Pull requests](https://img.shields.io/github/issues-pr/jamesonstone/kit)](https://github.com/jamesonstone/kit/pulls) [![Release](https://img.shields.io/github/v/release/jamesonstone/kit)](https://github.com/jamesonstone/kit/releases)
<!-- END KIT-MANAGED README BADGES -->

## Discover Kit functionality

`kit guide` is the official command guide for coding agents. Start with its
short index of commands and repository artifacts. Retrieve one command's usage,
examples, and flags with `kit guide <command>`, for example
`kit guide rules view`. Use `kit guide --json` for the complete structured catalog,
or `kit guide <command> --json` for a single command. The guide derives its
content from the live CLI and works outside a project without network access,
configuration changes, or usage writes.

The canonical agent contract and init onboarding direct coding agents to the
guide. These discovery routes do not guarantee that an agent reads it. Artifact
guidance includes the Constitution, living specs, contextual rules, references,
and the project-owned `docs/sops/<slug>/SOP.md` convention. SOPs use draft, active,
and retired states; the convention creates no command or execution engine.

## Design and security guidance

`kit init` creates `DESIGN.md` for product design, interface quality, and usability,
and `SECURITY.md` for security engineering and private vulnerability reporting.
The universal contract points agents to them only for relevant work. The
Constitution retains project invariants; contextual rules retain procedures.
These documents guide decisions; code and infrastructure enforce runtime controls.

Customize either document directly, including the project-owned reporting
placeholders in `SECURITY.md`. `kit reconcile` creates missing documents and
refreshes unchanged installed defaults using the existing artifact registry.
Existing documents and any local edits are preserved and reported; generated
agent instructions update independently. Preview with
`kit reconcile --dry-run --diff`. Replacing customized guidance requires explicit
`kit reconcile --force --file DESIGN.md` (or `SECURITY.md`); whole-project `--force`
preserves these policies. The template creates no reporting channel, support
range, response SLA, or security setting.

## How Kit Works

Kit gives every coding agent one compact universal contract (rendered into
`AGENTS.md`, `CLAUDE.md`, and Copilot instructions), contextual rules that are
read only when their trigger applies, and project memory in
`docs/CONSTITUTION.md`, `docs/specs/`, `docs/sops/`, and
`docs/references/testing.md`. The
binary ships the rules it was tested with and adds deterministic commands for
bootstrap, validation, drift maintenance, spec allocation, and AWS identity
checks. Agents plan, explore, and delegate with their own native tools.

New project:

```bash
kit init
kit init --dry-run          # print the file plan and write nothing
kit init --issue-template   # also add .github/ISSUE_TEMPLATE/issue.md
kit spec my-feature
# coding agent plans, implements, validates, and curates repository memory
kit check --project
```

Existing Kit project, created by any Kit release:

```bash
kit reconcile --dry-run --diff   # preview
kit reconcile                    # migrate to the current structure
```

For a Mint v1 deployment project with a team-first schema 2 `.mint.yaml` and
reviewed build, promotion and observation adapters:

```bash
kit init --mint                         # new Kit project
kit reconcile --mint --dry-run --diff    # existing Kit project: preview
kit reconcile --mint                    # create a missing controller
```

Kit runs the latest published v1 renderer with Go and uses that release's
immutable action commit in the generated controller. The renderer uses the
policy's environment names and adapter workflow names. Mint v2 leaves deployment
orchestration to project workflows, so `--mint` reports that the controller
scaffold is unavailable once v2 is the latest release. Existing controllers
remain project owned. Controller scaffolding never activates a deployment.

`kit init` adds a Mint release starter when no workflows exist. Each run
resolves the latest stable GitHub Release and loads its Mint Action. Existing
workflow structure and Mint integration are preserved. To add release
metadata, an enriched changelog and artifact retention to an existing job, select
its placement explicitly:

```bash
kit reconcile --mint-workflow .github/workflows/deploy.yaml \
  --mint-job deploy --mint-before-step publish --dry-run --diff
# Remove --dry-run --diff to apply after reviewing the preview.
```

Without placement, Kit reports what to select; it does not add a diagnostic
workflow. Use `kit reconcile --mint-workflows --dry-run --diff` to preview the
scaffold plan for an existing project. Kit adds files and preparation steps;
Mint and the project own release policy and the use of generated outputs. The
added steps use Bash, GitHub CLI and a checkout at the workspace root.

During init and reconcile, Kit refreshes supported `jamesonstone/mint` Action
references to the latest verified published stable release. Stable version tags
remain tags; full commit SHA pins become the verified release commit SHA. Custom
refs, shared YAML anchors and aliases, and v1 controller commands unsupported
by the latest release are preserved with a diagnostic. Lookup failures stop
init and reconcile before writes. Preview with `kit reconcile --dry-run --diff`.
Kit does not manage an installed Mint executable.

Mint supports explicit Linear links in generated changelogs. Private ticket
details are not fetched, and published release notes keep their subject-based format.

Diagnose without changing anything:

```bash
kit health
```

Reconcile keeps everything the project wrote: guidance outside the managed
contract block, edited Kit sections, edited rules, and project rules. It removes
retired Kit files (`docs/agents/`, `docs/references/workflows/`, the progress
summary, retired rules) only when they are exactly as a Kit release generated
them and committed to Git, and reports every file it keeps. From the primary
checkout it writes to a `kit-reconcile` linked worktree for review.

## Standard Operating Procedures

Kit scaffolds [a project-owned SOP guide](docs/sops/README.md). Agents create
repeatable procedures at `docs/sops/<slug>/SOP.md` and follow the
[SOP rules](docs/references/rules/standard-operating-procedures.md) to find,
validate, apply, and maintain them under existing authorization. Init and
reconcile create the guide only when missing; existing content is preserved,
even with force. Agents perform the work; Kit adds no SOP command or executor.

Keep personal procedures at `docs/sops/local/<slug>/SOP.md` in the primary
checkout. Init adds the ignore entry. Reconcile from the primary checkout
adds it in the `kit-reconcile` worktree. Merge that change and update the
primary checkout to receive the entry. Before writing, run
`git check-ignore -- docs/sops/local/example/SOP.md` in the primary checkout
with your intended path and confirm that Git reports it as ignored.
Agents in linked worktrees find the same local files through the SOP rule.
Local SOPs stay with this clone and need a separate backup.

## Shell and project environments

Kit ships a `shell-environment` rule for commands that depend on shell startup,
PATH, version managers, or project activation. Each project records its command
environment in `docs/references/testing.md` under Environment Preflights, or links
existing environment documentation there. Keep the working directory, toolchain,
activation, safe probes, and local/CI differences project-owned. Each team member
keeps personal host settings in their own global agent instructions.

An omitted `shell` setting uses each user's initialized interactive-login shell
on the execution host. A project can specify an exact invocation in `.kit.yaml`:

```yaml
shell:
  command: [zsh, -lic]
```

This is an executable/options prefix, not shell source. The agent appends its
task command as one argument. `shell: {}` or `command: []` restores the user
environment default. No variable or tilde expansion occurs in the prefix;
keep personal paths and secrets out of shared configuration. Use the shell's
own options, including its command-execution flag, for other shells or wrappers.
Run `kit config check --json` to validate configuration and see `shell: user`
or `shell: command`. These describe the configuration source, not a live shell.
Kit does not launch the shell itself or change its own subprocess environment.
Functions must run inside the initialized shell; only exported variables reach
child processes. Explicit script and CI execution contracts remain authoritative.

New projects receive the rule through `kit init`. Existing projects receive it
through `kit reconcile` after upgrading to a release that includes it; preview
with `kit reconcile --dry-run --diff`. Edited references and customized rules
stay preserved. Shell guidance does not change startup files, execution-host
settings, or network permissions, and cannot guarantee host compliance.

## Issue Providers

GitHub Issues is the default, including noninteractive init. GitHub issue
branches use `GH-<issue number>`. Interactive init also offers Linear. Use
verified non-secret workspace and team identities for scripts:

```bash
kit init --issue-provider linear --linear-workspace acme \
  --linear-team-id <verified-team-uuid> --linear-team-key ENG
kit config set --issue-provider linear --linear-workspace acme \
  --linear-team-id <verified-team-uuid> --linear-team-key ENG --dry-run --json
```

Remove `--dry-run` to apply configuration, then run the returned reconciliation
commands. Add `--linear-project-id <verified-project-uuid>` for a project or
`--linear-project-id ''` to clear it. Switch back with
`kit config set --issue-provider github`. Configuration is repository-local:

```yaml
issues:
  provider: linear
  linear:
    workspace: acme
    team_id: <verified-team-uuid>
    team_key: ENG
    # project_id: <verified-project-uuid>
```

The setter preserves unrelated configuration. Credentials belong in the agent
host's authorized capability, never `.kit.yaml`. Ticket creation and lifecycle
updates follow Kit's agent instructions through that capability. Missing Linear
access stops ticket operations; it never falls back to GitHub.

During setup, the agent reads the verified Linear team's key through its
authorized connection. For a project with multiple teams, the selected issue
team supplies the key. `--linear-team-key` is optional for existing setups.
`kit init` and `kit config set` report the configured branch prefix. Use the exact Linear
issue identifier, such as `ENG-123`, as the branch and conventional commit
scope, with a full URL footer:

```text
feat(ENG-123): :sparkles: add invitations

Refs: https://linear.app/acme/issue/ENG-123
```

Keep the footer when squashing. Existing GitHub references remain supported.
Mint reads explicit links offline; bare Linear identifiers cannot resolve a URL.
Review committed text and link visibility before publication.

## Work ledger

`work_ledger` is optional. When it is omitted, Kit does not sync a work ledger.

[Plane](https://plane.so) is the only supported provider. `linear`, `jira`, and
`notion` are reserved and fail validation for now. Open an issue at
https://github.com/jamesonstone/kit/issues if you want support for another
ledger tool.

Kit reads and writes global configuration at `$XDG_CONFIG_HOME/kit/.kit.yaml`
when `XDG_CONFIG_HOME` is absolute and nonempty. Otherwise it uses
`~/.config/kit/.kit.yaml`. For unset global keys, Kit reads `~/.kit/.kit.yaml`
and then `~/.config/kit/.kit.yaml` if it differs from the canonical path.
Kit writes only to the canonical path and leaves fallback files in place. If
that path aliases a fallback, use a separate file for canonical writes.
Kit warns when it finds a fallback file. To clear the warning, merge any
settings you want to keep into the canonical file, preserving values already
set there. If the canonical config aliases the fallback, create a separate
canonical file first. Then move the fallback file aside. `kit config check --json`
includes the warning in its findings.

The repository `.kit.yaml` uses the same schema and overrides global defaults
key by key. A repository scalar, list, or explicit null replaces the global
value at that key. Issue routing, the issue-template opt-in, and repository
schema and scaffold versions come from the repository. Reconcile writes only
repository-owned configuration. Generated README links use the repository's
configured GitHub identity or its origin. Keep user config private. Kit stores no
credentials and does not call the ledger API.

```yaml
work_ledger:
  provider: plane
  owner_label: <owner-label>
  plane:
    workspace: <workspace>
    project: <project>
```

Omitted state names mean `Active`, `In review`, and `Done`. Set
`work_ledger.provider: none` in a repository to turn off a global ledger there.
`kit config check --json` identifies obsolete flat ledger keys. Update them
to the schema above before running reconcile.

## Command Surface

| Area | Commands |
| --- | --- |
| Coding-agent entry | `kit guide [<command path>] [--json]` |
| Bootstrap and memory | `kit init`, `kit spec` |
| Rules | `kit rules add\|list\|view` |
| Local guidance | `kit inspect [--path <directory>] [--json]` |
| Validation and maintenance | `kit check`, `kit reconcile`, `kit health`, `kit registry status`, `kit status`, `kit config check\|set`, `kit aws verify` |
| Local usage | `kit usage [report\|status\|refresh\|clear\|enable\|disable]` |
| Utilities | `kit upgrade`, `kit version`, `kit completion` |

See the [command guide](docs/commands.md), the [v3 migration
guide](docs/migration-v3.md), and the [release notes](docs/releases/v3.0.0.md).

## Inspect local guidance

Run `kit inspect` from a project or any nested directory. Use
`kit inspect --path ../another-project --json` to select a different project
without changing your working directory. The report lists current local rulesets,
SOPs, agent entries, the Constitution, and design, security, testing, and SOP guides.
It does not list shipped-but-uninstalled rules or historical specs.

JSON schema version 1 contains `project_root`, config findings, ordered `entries`,
and `errors`. Empty `findings` is `[]`. Each entry has a project-relative path, kind, state, and explicit
`active` boolean. Rule and SOP metadata includes descriptions and status; SOPs
also include an owner. Missing configured files are reported as `missing`.
Invalid metadata and read failures preserve other findings and return a nonzero
exit status. Unknown status is never active. Symlink guidance and paths outside
the project are rejected. Inspect does not repair configuration, run instructions,
write files, use the network, or record usage telemetry.

## Local Usage Data

Kit records minimal local command events by default so maintainers can identify
unused surfaces using evidence rather than intuition. It never records command arguments, output, repository names, paths, file content, environment values, or secrets; project identity is local and pseudonymous. Data remains on the
machine, is retained for at most 365 days, and is capped at 16 MiB total with
2 MiB shards. Unversioned development builds, test binaries, and processes with
`KIT_USAGE_DISABLED=1` never record.

```bash
kit usage status
kit usage report --since 90d
kit usage refresh
kit usage disable --global
kit usage clear --all --yes
```

A global disable is absolute. Project-level enable or disable is stored in that
repository's `.kit.yaml`. Usage commands do not record themselves.

## Install

```bash
go install github.com/jamesonstone/kit/v3/cmd/kit@latest
```

Or clone the repository and run `make build`. Enable repository-managed hooks
for the clone with `make install-git-hooks`.

## Documentation

- [Overview](docs/overview.md)
- [Commands](docs/commands.md)
- [Migration to v3](docs/migration-v3.md)
- [Historical v2 migration](docs/migration-v2.md)
- [Rules and references](docs/references/README.md)
- [Project Constitution](docs/CONSTITUTION.md)

## License

MIT

## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).

Existing security policies are discovered in `.github/SECURITY.md`, `SECURITY.md`,
then `docs/SECURITY.md` order. Init and reconcile preserve a non-root policy
without creating or registering a root placeholder. Agents read the first
existing policy in that order. Explicit `--force --file SECURITY.md` can create
or replace the root document; it never edits non-root policies.
