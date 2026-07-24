# Git Worktrees

## Policy Authority

Native `git worktree` commands and ordinary filesystem operations define the
portable workflow. Kit rules, reconciled files, and generated agent
instructions must work for teammates who do not have `git-wt`, a shell alias,
an editor integration, or any other wrapper installed.

Optional helpers may make the same workflow more convenient for manual use,
but they do not define policy and must preserve every path, environment, and
safety invariant in this document.

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

## Portable Native Git Workflow

Start every operation from a checkout of the intended repository. Keep the root
checkout on its protected default branch, inspect the registered worktrees, and
reuse the exact branch worktree when one already exists:

```bash
git worktree list --porcelain
```

After the GitHub issue exists, fetch the remote base and create its durable
lane. Substitute the repository's actual owner, repository, issue, and base:

```bash
INVOKING_ROOT="$(cd "$(git rev-parse --show-toplevel)" && pwd -P)"
BASE_BRANCH="main"
BRANCH="GH-123"
WORKTREE_PATH="$HOME/worktrees/example-owner/example-repository/$BRANCH"

git fetch origin "$BASE_BRANCH"
mkdir -p "$(dirname "$WORKTREE_PATH")"
git worktree add -b "$BRANCH" "$WORKTREE_PATH" "origin/$BASE_BRANCH"
```

If the branch already exists but has no registered worktree, attach it without
recreating it:

```bash
git worktree add "$WORKTREE_PATH" "$BRANCH"
```

If only the remote branch exists, create a tracking branch and worktree:

```bash
git fetch origin "$BRANCH"
git worktree add --track -b "$BRANCH" "$WORKTREE_PATH" "origin/$BRANCH"
```

Never bypass Git's protection against checking out the same branch in two
worktrees. Resolve an existing lane only from exact records in
`git worktree list --porcelain`; do not use substring matching.

For temporary pull-request inspection, fetch the pull request head and create a
detached `PR-<number>` lane:

```bash
PR_PATH="$HOME/worktrees/example-owner/example-repository/PR-77"
git fetch origin "pull/77/head"
git worktree add --detach "$PR_PATH" FETCH_HEAD
```

Detached PR lanes are inspection-only. For review repair, resolve the pull
request's same-repository head branch, then reuse its exact registered durable
worktree or attach that branch with `git worktree add`.

### Writable-Lane Environment Link

For a writable lane, link the invoking checkout's repository-root `.env` by
default when the source exists:

```bash
SOURCE_ENV="$INVOKING_ROOT/.env"
DEST_ENV="$WORKTREE_PATH/.env"

resolve_link_target() {
  link_text="$(readlink "$1")" || return 1
  case "$link_text" in
    /*) target_path="$link_text" ;;
    *) target_path="$(dirname "$1")/$link_text" ;;
  esac
  target_dir="$(cd -P "$(dirname "$target_path")" 2>/dev/null && pwd)" ||
    return 1
  printf '%s/%s\n' "$target_dir" "$(basename "$target_path")"
}

if [ -L "$DEST_ENV" ]; then
  if [ ! -e "$DEST_ENV" ]; then
    echo "ABORT: destination .env is a broken link" >&2
    exit 1
  fi
  RESOLVED_TARGET="$(resolve_link_target "$DEST_ENV")" || {
    echo "ABORT: destination .env is unreadable" >&2
    exit 1
  }
  if [ "$RESOLVED_TARGET" != "$SOURCE_ENV" ]; then
    echo "ABORT: destination .env points to an unexpected target" >&2
    exit 1
  fi
elif [ -e "$DEST_ENV" ]; then
  echo "ABORT: destination .env already exists: $DEST_ENV" >&2
  exit 1
elif [ -f "$SOURCE_ENV" ]; then
  ln -s "$SOURCE_ENV" "$DEST_ENV"
else
  echo "No repository-root .env exists; no environment file was linked."
fi
```

Capture `INVOKING_ROOT` before changing directories. Reusing a writable lane
must run the same exact source and destination validation and create the link
when it is missing. When isolation is required, intentionally omit this linking
step.

Never copy `.env`, overwrite a destination `.env`, or automatically share
`.envrc`; `.envrc` is executable shell configuration. Reuse accepts only an
existing symlink that resolves to the exact expected source. A regular
destination file, a broken link, or a link to any unexpected target is a
collision and must stop the operation. Detached PR inspection does not link
`.env`, and migration preserves existing files and links without creating new
ones.

### Inspection, Migration, And Removal

Listing is read-only:

```bash
git worktree list --porcelain
```

Review stale administrative metadata before pruning:

```bash
git worktree prune --dry-run --verbose
git worktree prune --verbose
```

Move a registered legacy worktree only after validating the exact source,
destination, and all candidate collisions:

```bash
git worktree move "/exact/registered/source" \
  "$HOME/worktrees/example-owner/example-repository/GH-123"
```

Migration uses `git worktree move`, preserves dirty contents and existing
environment files or links, and never uses ordinary `mv`, stash, reset, clean,
or force.

Before removal, prove the target is an exact registered path, is not the current
checkout, contains no tracked, untracked, ignored, dirty, or unpublished state,
and has no unsafe `.env`. A verified `.env` symlink whose target exactly matches
the expected invoking-checkout source is the sole narrow exception:

1. Verify the destination is a symlink and its target is the exact expected
   repository-root `.env`.
2. Unlink only that destination symlink.
3. Run ordinary non-force `git worktree remove "/exact/registered/path"`.
4. If Git removal fails, restore the same symlink.

Refuse regular `.env` files, unexpected symlinks, and every other tracked,
untracked, ignored, dirty, or unpublished item. Never use `--force`, reset,
clean, stash, or branch deletion.

## Optional GitWT Convenience

Kit also ships `git-wt` for people who want shorter manual commands. Git
discovers the executable as `git wt`; no Kit rule or `kit reconcile` output
requires it.

From a Kit source checkout:

```bash
make build
git wt help
```

`make build` produces `bin/kit` and `bin/git-wt`, then installs or updates
`~/.local/bin/git-wt`. `~/.local/bin` must be on `PATH`. Set `GIT_WT_ROOT` only
when an explicit non-default root is needed; normal use uses `~/worktrees`.

Manual command cheat sheet:

```bash
git wt issue 123                  # create or reuse writable GH-123
git wt add GH-123                 # attach an existing branch
git wt pr 77                      # detached inspection-only PR-77
git wt repair 77                  # writable lane for the PR head branch
git wt path GH-123                # print the exact registered path
cd "$(git wt path GH-123)"        # navigate in the current shell

git wt issue 123 --no-link-env    # writable lane without the default .env link
git wt add GH-123 --no-link-env
git wt repair 77 --no-link-env

git wt list                       # read-only; never prunes
git wt remove PR-77               # exact, conservative, non-force removal
git wt prune --dry-run            # preview stale metadata
git wt prune                      # apply reviewed pruning
git wt migrate                    # preview legacy hierarchy migration
git wt migrate --apply            # apply the reviewed migration
```

`git wt path <lane>` prints only an exact registered path because an external
Git subcommand cannot change its parent shell's working directory. Writable
`issue`, `add`, and `repair` commands implement the default `.env` link and
`--no-link-env` opt-out. Detached `pr` and migration do not create environment
links.

GitWT remains a thin wrapper around native worktree and filesystem operations.
It does not start or stop applications, manage databases, allocate ports,
manage Temporal state, supervise processes, orchestrate sibling repositories,
or switch the root checkout away from its protected default branch.

## Safety Rules

- Reuse an existing registered worktree when its branch already has one.
- Keep one active branch in one worktree.
- Confirm the intended repository, exact lane, and shared ref effects before a
  mutation.
- Do not use stash as cross-worktree scratch space.
- Link only the exact `.env` symlink for writable lanes by default, or omit it
  when isolation is required. Do not manually copy credentials.
- Never automatically share `.envrc`, credentials other than the explicit
  `.env` symlink, tokens, private keys, or other machine-local configuration.
- Never use reset, clean, force removal, or branch deletion to make a worktree
  operation succeed.
- Do not remove a worktree until its contents are clean and its branch state is
  published or otherwise deliberately retained.
- Subagents must not independently create, switch, move, or remove worktrees. A
  supervisor may assign an already prepared worktree and exact scope.
