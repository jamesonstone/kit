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

## How Kit Works

Kit gives every coding agent one compact universal contract (rendered into
`AGENTS.md`, `CLAUDE.md`, and Copilot instructions), contextual rules that are
read only when their trigger applies, and project memory in
`docs/CONSTITUTION.md`, `docs/specs/`, and `docs/references/testing.md`. The
binary ships the rules it was tested with and adds deterministic commands for
bootstrap, validation, drift maintenance, spec allocation, and AWS identity
checks. Agents plan, explore, and delegate with their own native tools.

New project:

```bash
kit init
kit spec my-feature
# coding agent plans, implements, validates, and curates repository memory
kit check --project
```

Existing Kit project, created by any Kit release:

```bash
kit reconcile --dry-run --diff   # preview
kit reconcile                    # migrate to the current structure
```

Mint deployment project with an existing team-first schema 2 `.mint.yaml` and
reviewed build, promotion and observation adapters:

```bash
kit init --mint                         # new Kit project
kit reconcile --mint --dry-run --diff    # existing Kit project: preview
kit reconcile --mint                    # create a missing controller
```

Kit uses the embedded Mint v0.5.0 renderer to generate repository Actions for
promotion, hotfix, rollback, resume, observation and reconciliation. It uses the
policy's environment names and adapter workflow names. Existing controllers,
policies and adapters stay project owned, even with `--force`. No registry,
provider or production environment is assumed. Package and artifact projects
retain their existing publication workflows. Scaffolding keeps the
`MINT_RELEASE_ENABLED` gate; it does not activate deployments. Review source-build
isolation, adapter manifests, runtime configuration identities and verified
baselines before enabling that gate.

`kit init` automatically plans GitHub Actions integration using published Mint
v0.9.0, pinned to `eac91490f772adbd2ab030cb6eba3910c1bca93c`:

- No workflows: create one canonical release-state pipeline for default-branch
  pushes. It publishes tags/GitHub Releases; it does not invent application build,
  registry, cloud or deployment adapters. Review before merging the scaffold.
- One release/deploy workflow with one ordinary `ubuntu-latest` steps job and
  explicit read-only permissions, without environment/ambient credentials or credential-export steps: append
  Mint installation/version verification after existing steps. Triggers, jobs,
  dependencies, permissions, secrets, environments and deployment steps stay intact.
  Existing release publishers continue to own publication.
- CI-only, ambiguous, privileged, dynamic or reusable jobs: add a separate read-only tooling
  check. No publisher/deployment is inferred. Select a reviewed steps job with
  `--mint-workflow .github/workflows/deploy.yaml --mint-job deploy`. Privileged
  selections explicitly opt into Mint sharing that job's ambient credentials.
- Existing Mint action integration: preserve its pin and behavior. Repeated runs
  add no duplicate Mint steps or publishers. Custom files are not overwritten.

Existing projects can preview the same planner with
`kit reconcile --mint-workflows --dry-run --diff`; apply with
`kit reconcile --mint-workflows`. The selectors also work with reconcile, whose
usual delivery isolation remains in force. `--force` does not replace workflows.
Init plans against the original workflow inventory before creating auto-assignment.
Remote default-branch metadata is used where available; a new repository without
that metadata gets a `main` starter and a runtime actual-default-branch guard.
Inspect the branch filter before adoption in repositories using another branch.
Local `kit mint update`, these action pins, and the older embedded environment
controller renderer are separate version surfaces.

Kit serializes cooperating Mint workflow writes and rollbacks with per-file
`.kit-mint.lock` files and checks snapshots at write boundaries. Existing locks
fail closed and are never reclaimed automatically; inspect their PID/path before
retrying. External editors do not honor this protocol, so their final
check-to-replacement window cannot be eliminated by these locks. Avoid concurrent
external editing during apply. Batch failure uses best-effort rollback: a complete
newly created Mint workflow is retained and its exact path reported rather than
risk deleting an intervening edit.

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

## Command Surface

| Area | Commands |
| --- | --- |
| Bootstrap and memory | `kit init`, `kit spec` |
| Rules | `kit rules add\|list\|view` |
| Validation and maintenance | `kit check`, `kit reconcile`, `kit health`, `kit registry status`, `kit status`, `kit config check`, `kit aws verify` |
| Local usage | `kit usage [report\|status\|refresh\|clear\|enable\|disable]` |
| Utilities | `kit upgrade`, `kit version`, `kit completion` |

See the [command guide](docs/commands.md), the [v3 migration
guide](docs/migration-v3.md), and the [release notes](docs/releases/v3.0.0.md).

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

### Issue provider

GitHub Issues remains the default, including noninteractive `kit init`.
Interactive init offers GitHub Issues or Linear. For scripts, use verified
non-secret Linear target identities:

```bash
kit init --issue-provider linear --linear-workspace acme \
  --linear-team-id 11111111-1111-4111-8111-111111111111
```

The UUID above is a placeholder. Obtain the real team UUID through your authorized
Linear capability. Add `--linear-project-id <project-uuid>` for a verified project;
omitting it means team-only work. Init does not authenticate or create tickets.
Do not store API keys or tokens in `.kit.yaml`.

To switch an existing repository, edit its tracked `.kit.yaml` and run
`kit reconcile` in its isolated delivery checkout (or preview with
`kit reconcile --dry-run --diff`):

```yaml
issues:
  provider: linear
  linear:
    workspace: acme
    team_id: 11111111-1111-4111-8111-111111111111
    # project_id: <verified-project-uuid>  # optional
```

Set `issues.provider: github` to switch back. Omission also means GitHub.
Use a Kit version containing issue-provider support on every agent host before
selecting Linear; older binaries cannot enforce this selection.
Targets are repository-local and survive reconcile, including a forced config
refresh. Existing lanes retain their original ticket; switching affects new work.
Reconcile switches exact Kit-generated PR templates; custom templates stay intact.
Update a custom Ticket section to the selected provider before new delivery.
Canonical delivery rules verify target membership, human assignment, ticket
payload parity, duplicate searches, provider errors and completion states. A host
with no authorized Linear capability reports that blocker; it never creates a
GitHub issue in place of the selected Linear ticket. Connect Linear separately
through your normal secure setup when needed; Kit does not install credentials.

Linear uses `ENG-123` for branches and conventional commit scope, for example
`feat(ENG-123): :sparkles: add invitations`, with the full canonical URL in a final
`Refs: https://linear.app/acme/issue/ENG-123` footer. Preserve the footer when
squashing. GitHub issues, PR hosting, checks and historical `GH-123` references
retain their behavior. Use the coordinated Mint Linear changelog support before
adopting Linear if release notes must retain its links. Mint parses explicit
links offline; identifiers, URLs and commit text can be public, but private ticket
content is never fetched for enrichment. Review commit text for public visibility.

Kit does not contain a native GitHub or Linear ticket-creation API. Ticket
creation/population and state changes are agent actions through the host's
capabilities; init/reconcile only persist configuration and ship the instructions.
The Linear path matches that existing GitHub mechanism. It does not provision
Linear integrations, map workflow states automatically, install credentials, or
configure server-side closure on merge. An authorized agent must verify and apply
the team's state transitions after observing completion. Contract tests prove
which instructions ship, not that a host executed provider actions successfully.

### Agent-friendly provider changes

Use the same verified non-secret target IDs as init:

```bash
kit config set --issue-provider linear --linear-workspace acme \
  --linear-team-id 11111111-1111-4111-8111-111111111111 --dry-run --json
# Remove --dry-run to apply configuration, then execute the returned next_actions.
```

The UUID is a placeholder. Add `--linear-project-id <verified-uuid>` for a project,
or `--linear-project-id ''` to clear it. Existing target values are retained unless
explicitly supplied. Switch back with `kit config set --issue-provider github`.
Repeat applications are no-ops. Unknown providers, invalid/missing required targets,
newer schemas and unsafe files fail before config writes. Only issue fields change;
unrelated YAML values and comments remain. Root YAML merges and aliased or anchored issue settings fail
without writes and need manual review to avoid changing other alias users.

`--json` emits the non-secret issue settings, target, changed/dry-run state and
`reconciliation_required: true` with explicit next actions. The setter does not
silently migrate other files. Run the reported `kit reconcile --dry-run --diff`,
then `kit reconcile`, in that target and review/deliver the changes. The existing
reconcile target resolver protects primary checkouts by preparing/reusing its
`kit-reconcile` worktree. Dry runs never prepare a worktree: their config preview
uses the invoking checkout; a later primary-checkout write uses the canonical
base/worktree and never copies unrelated working changes into it.

### Mint version checks and explicit local updates

```bash
kit mint check --json
kit mint check --require-version v0.8.0 --json
kit mint check --require-linear --json
kit mint update --version v0.8.0 --dry-run --json
kit mint update --version v0.8.0 --json
```

The first command only inspects the resolved `mint` executable's Go metadata and
reads official published stable release metadata. It never installs or executes
Mint. An outdated stable install is reported with `update_available`; missing or
unsupported local installations and failed requirements exit nonzero. Network or
rate-limit failures are errors, not evidence of compatibility. `--json` includes a
report and error on operational failures. `--require-version` is a numeric local
minimum, not proof of feature support; `--version` selects a published target.

Local executable, repository `go.mod` dependency (including replacements), and
`.github/workflows` Mint action refs are separate report fields. Pin comments are
unverified version hints. SHA pins and repository compatibility are reported as
`not_verified`; this command does not resolve their release ancestry or rewrite
any dependency/workflow. Updating local Mint never updates Kit's embedded Mint
renderer dependency or consumer CI pins. Review those separately in repository PRs.
Symlinked pin files/directories fail rather than being read outside the repository.

Mint currently publishes no binary assets. Explicit update uses its documented
`go install github.com/jamesonstone/mint/cmd/mint@<published-tag>` source mechanism,
with the official Go proxy and checksum database required. It stages under the
verified Go bin, verifies module path/version/host platform, preserves ownership,
group and mode, and replaces atomically only after a successful build. Existing
installations must be verified user-owned Go-bin builds. Development/pseudo versions,
downgrades, package-manager/custom/symlink paths, ambiguous ownership, unsafe
permissions and unsupported systems fail clearly; no sudo, arbitrary installer or
PATH changes occur. Initial install requires an existing safe Go bin already on
PATH. The installed Go toolchain must support the target; no automatic Go upgrade.
A failed build/verification leaves the previous Mint executable unchanged.
`update --dry-run` validates the proposed update without building or installing.
Omitting `--version` chooses the latest published stable release without downgrading.

Linear CHANGELOG link support was released in Mint v0.9.0 from merged PR #33.
The annotated tag resolves to `eac91490f772adbd2ab030cb6eba3910c1bca93c`, which
contains the offline Linear footer parser. `--require-linear` now checks v0.9.0
or newer for the local executable/update target. Existing consumer CI SHA pins
remain separate and need their own verified release mapping. This capability
covers explicit Linear links in `mint changelog`; release-state lightweight notes
retain their existing subject-based format and do not fetch private ticket data.
