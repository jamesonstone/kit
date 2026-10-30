# Kit Commands

Every command below does deterministic work that an agent cannot reproduce as
safely or cheaply by itself. Agents are the primary callers; commands avoid
interactive prompts unless a human runs them in a terminal.

One path per job: `kit init` creates a project, `kit reconcile` is the only
command that migrates or repairs an existing one (and from the primary checkout
it writes only to a linked worktree), and `kit health` diagnoses.

## Bootstrap And Memory

| Command | Why it exists |
| --- | --- |
| `kit init` | Creates a new Kit project: `.kit.yaml`, the universal contract in `AGENTS.md`/`CLAUDE.md`/Copilot instructions, the Constitution starter, `DESIGN.md`, `SECURITY.md`, `docs/references/testing.md`, the core rules shipped with this binary, and the developer-experience starter files. Refuses to run where `.kit.yaml` exists or inside an existing Kit project; use `kit reconcile`. Copies its setup prompt to the clipboard only at a terminal and prints it otherwise. |
| `kit spec <feature>` | Allocates a worktree-safe feature number and scaffolds or adopts `docs/specs/<id>-<feature>/SPEC.md`. |

## Rules

| Command | Why it exists |
| --- | --- |
| `kit rules add` | Installs a rule shipped with this Kit version (for example the optional `readme-header-tagline` or `llms-txt`), or creates a project rule. |
| `kit rules list` / `kit rules view <slug>` | Show installed and available rules. |

Rules ship inside the Kit binary. A released binary installs exactly the rules
it was built and tested with; nothing is fetched from GitHub. Unmodified
installed rules update on refresh; locally edited rules are preserved and
reported unless `--force` is used.

## Validation And Maintenance

| Command | Why it exists |
| --- | --- |
| `kit check [feature]` / `kit check --project` | Validates spec front matter and relationships, duplicate feature numbers, rule documents, the managed contract block, and, when `.kit.yaml` sets `source_file_line_limit`, handwritten source-file length. Exits non-zero on blocking findings. |
| `kit reconcile` | Migrates a project created by any Kit release to the current structure, then audits project documents. Keeps project and edited content, removes only unmodified retired Kit files that Git can restore, and reports what it kept. From the primary checkout it writes to a `kit-reconcile` linked worktree. `--dry-run --diff` previews; `--force` also replaces edited Kit sections and edited shipped rules. Prints any remaining findings as severity, file, issue, and fix. |
| `kit health` | Read-only diagnosis for people and scheduled automation: reports what `kit reconcile` would change (`--diff` shows it), runs the project check, and exits non-zero only when that check fails. Never writes. |
| `kit registry status` | Cheap read-only report of whether Kit-managed files and rules match this binary. |
| `kit status` | Current feature and Kit-managed state. |
| `kit config check` | Validates `.kit.yaml` and, interactively, repairs the AWS context. |
| `kit aws verify` | Verifies the configured AWS profile, account, and Region before AWS work. |

## Local Usage

| Command | Why it exists |
| --- | --- |
| `kit usage report` / `status` / `refresh` / `clear` / `enable` / `disable` | Local, minimal command-usage evidence for deciding what Kit should keep. |

Usage collection is local-only and records no arguments, output, paths,
content, environment values, or secrets. Unversioned development builds, test binaries,
`--help` lookups, and processes with `KIT_USAGE_DISABLED=1` never record.

## Utilities

`kit upgrade`, `kit version`, `kit completion`, and `kit help`.

## Removed In Version 3

These no longer exist: `kit context resolve` and the workflow manifests under
`docs/references/workflows/`, `kit capabilities`, `kit dispatch`, `kit pr fix`,
`kit pr orchestrate`, `kit improve run`, `kit instructions`, `kit rules link`,
the `--profile` and `--single-agent` flags, the reconcile agent prompt and its `--copy`, `--prompt-only`, `--migrate-references`, and `--migrate-verification` flags, and the generated
`docs/PROJECT_PROGRESS_SUMMARY.md`. The universal contract and each rule's
`Applies When` section replace workflow routing; hosts already expose their
own tools, delegation, and command help.

### Design and security document lifecycle

`kit init` scaffolds canonical `DESIGN.md` and `SECURITY.md` without replacing
existing policies. `kit reconcile` adds missing documents, updates unchanged
Kit defaults, and preserves edited documents as project-owned. The existing
`.kit.yaml` artifact registry supplies installed-hash evidence. Use
`kit reconcile --dry-run --diff --file SECURITY.md` to preview, including any
ownership metadata change. Customized documents require explicit
`--force --file SECURITY.md` (or `DESIGN.md`) to accept Kit defaults; review the
diff first. Whole-project `--force` preserves them.

Agent instructions reference design guidance for UI/UX/design and security
guidance for security-relevant engineering. The Constitution remains the source
of project invariants; contextual rules provide procedures. `kit check --project`
checks presence, readability, and a Markdown heading, and reports preserved
customizations as nonblocking advisories. It does not certify design quality,
security, reporting-channel availability, or runtime enforcement.

Existing security policies are discovered in `.github/SECURITY.md`, `SECURITY.md`,
then `docs/SECURITY.md` order. Init and reconcile preserve a non-root policy
without creating or registering a root placeholder. Agents read the first
existing policy in that order. Explicit `--force --file SECURITY.md` can create
or replace the root document; it never edits non-root policies.
