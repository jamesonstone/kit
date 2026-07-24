# Git Worktrees

## Mental Model

A worktree is another checkout attached to the same Git clone.

Each worktree has its own:

- working files
- index and staging area
- checked-out `HEAD`

All worktrees of that clone share:

- commits and objects
- local and remote refs
- remotes and most Git configuration
- stash entries

Think “separate desk, shared filing cabinet.” A worktree protects an in-flight
checkout from unrelated file and branch changes, but it is not a second clone
or an isolation boundary for refs, fetches, configuration, or stash state.

## Canonical Hierarchy

Keep linked worktrees outside the source clone:

```text
~/worktrees/<owner>/<repository>/<lane>
```

Examples:

```text
~/worktrees/lsmc-bio/labcore/GH-76
~/worktrees/lsmc-bio/labcore/PR-77
~/worktrees/patient-driven-care/mypa/codex/consent-service-fix
```

Owner and repository directory names are lowercase. Durable issue lanes use
exact uppercase `GH-<number>`. Temporary pull-request views use exact uppercase
`PR-<number>`.

Do not put linked worktrees inside a repository, including under
`.worktrees/`. Keeping them in `~/worktrees` prevents recursive tooling,
watchers, search, backup rules, build scripts, and repository cleanup from
mistaking one checkout for content owned by another.

## Install

Kit owns the `git-wt` executable. Git discovers an executable named `git-wt` on
`PATH` as the subcommand `git wt`.

From a Kit source checkout:

```bash
go build -o ~/.local/bin/git-wt ./cmd/git-wt
git wt help
```

`~/.local/bin` must be on `PATH`. Set `GIT_WT_ROOT` only when an explicit
non-default root is needed; normal use requires no configuration and uses
`~/worktrees`. Git reserves `git <command> --help` for manual pages, so use
`git wt help` or `git-wt --help` for inline help.

## Productive Workflow

Create or reuse a durable issue lane after its GitHub issue exists:

```bash
git wt issue 76
cd "$(git wt root)/GH-76"
```

Open an existing local or remote branch:

```bash
git wt add GH-76
git wt add codex/consent-service-fix
```

Inspect a pull request without checking its branch out for editing:

```bash
git wt pr 77
cd "$(git wt root)/PR-77"
```

`PR-77` is detached and inspection-only. To address review feedback, resolve
the pull request's same-repository head branch and reuse its durable worktree:

```bash
git wt repair 77
```

This is the recommended entrypoint before running `kit pr fix --pr 77`: inspect
with `git wt pr 77` when useful, then perform edits and validation in the path
reported by `git wt repair 77`.

Writable `issue`, `add`, and `repair` lanes link the invoking checkout's
repository-root `.env` by default:

```bash
git wt issue 76
git wt add GH-76
git wt repair 77
```

Use explicit isolation when a lane must not share that environment:

```bash
git wt issue 76 --no-link-env
git wt add GH-76 --no-link-env
git wt repair 77 --no-link-env
```

GitWT creates a symlink; it never copies `.env` contents. If the invoking
checkout has no `.env`, lane creation still succeeds and reports that no
environment file was linked. Reusing a writable lane ensures the expected link
exists when linking is enabled. GitWT never overwrites a destination `.env`:
regular files and symlinks pointing anywhere unexpected are preserved and
reported as errors.

GitWT does not automatically link `.envrc`, because it is executable shell
configuration. Detached `git wt pr` inspection lanes never link `.env`, and
`git wt migrate` preserves existing files and links without creating new ones.

Inspect and maintain worktree state:

```bash
git wt list
git wt prune --dry-run
git wt prune
git wt remove PR-77
```

`list` is read-only and never prunes. `remove` targets one exact registered
path, never forces, never deletes the branch, and refuses tracked, untracked,
ignored, or unpushed state. A matching GitWT-managed `.env` symlink is the only
exception: GitWT verifies its target, removes only that link, invokes ordinary
non-force `git worktree remove`, and restores the link if Git removal fails.
Regular `.env` files and unexpected symlinks always block removal.

## Legacy Migration

Preview direct-child linked worktrees currently accumulated beneath
`~/worktrees`:

```bash
git wt migrate
```

Review every source and destination, then apply:

```bash
git wt migrate --apply
```

Migration uses `git worktree move`; it does not use ordinary `mv`, stash,
reset, clean, or force. Dirty contents move with their worktree. It skips
already hierarchical owner directories and standalone clones, and it stops on
destination collisions or unsupported detached identities.

## Safety Rules

- Reuse an existing registered worktree when its branch already has one.
- Keep one active branch in one worktree; Git enforces this for normal branch
  checkout.
- Fetches and ref changes are shared. Confirm the intended repository and lane
  before a mutation.
- Do not use stash as cross-worktree scratch space; stash entries are shared and
  easy to apply in the wrong lane.
- Let GitWT link `.env` for writable lanes by default, or use `--no-link-env`
  when isolation is required. Do not manually copy credentials.
- Never automatically share `.envrc`, credentials other than the explicit
  `.env` symlink, tokens, private keys, or other machine-local configuration.
- Never use reset, clean, force removal, or branch deletion to make a worktree
  operation succeed.
- Do not remove a worktree until its contents are clean and its branch state is
  published or otherwise deliberately retained.
- Subagents must not independently create, switch, move, or remove worktrees. A
  supervisor may assign an already prepared worktree and exact scope.

## Scope Boundary

GitWT is a thin wrapper around native Git worktree operations plus default
`.env` linking for writable lanes. It manages lane paths and branches only. It
does not start or stop applications, manage databases, allocate ports, manage
Temporal state, supervise processes, orchestrate sibling repositories, or
switch the root checkout away from `main`.
