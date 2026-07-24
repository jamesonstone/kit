# Git Worktrees

## Policy Authority

Native `git worktree` commands and ordinary filesystem operations define this
portable workflow. No Kit-managed rule depends on `git-wt`, `git wt`, a shell
alias, an editor integration, or another wrapper.

Optional helpers may make the same workflow more convenient for manual use,
but they must preserve every path, environment, and safety invariant here.

## Mental Model

A worktree is another checkout attached to the same Git clone.

Each worktree has separate working files, an index, and `HEAD`. All worktrees
of the clone share commits, objects, refs, remotes, most Git configuration, and
stash entries. A worktree protects one checkout from unrelated file and branch
changes; it is not a second clone or an isolation boundary for shared Git
state.

Keep the primary checkout on the protected default branch. Agents develop and
test inside assigned durable lanes, and the same branch must never be checked
out in two worktrees at once.

## Canonical Hierarchy

Keep linked worktrees outside the source clone:

```text
~/worktrees/<owner>/<repository>/<lane>
```

Owner and repository names are lowercase. Durable issue lanes use exact
uppercase `GH-<number>`. Detached pull-request inspection lanes use exact
uppercase `PR-<number>`.

Do not put linked worktrees inside a repository, including under
`.worktrees/`. External placement prevents recursive tooling, watchers, search,
backup rules, builds, and cleanup from treating one checkout as another
checkout's content.

## Portable Native Git Workflow

Start from a checkout of the intended repository and inspect exact registered
worktrees before creating anything:

```bash
git worktree list --porcelain
```

The first entry is Git's primary worktree. Capture its stable physical path for
environment-link validation:

```bash
PRIMARY_ROOT="$(
  git worktree list --porcelain |
    sed -n '1s/^worktree //p'
)"
PRIMARY_ROOT="$(cd "$PRIMARY_ROOT" && pwd -P)"
```

After the GitHub issue exists, fetch the remote base and create its durable
lane. Substitute the actual owner, repository, issue, and base branch:

```bash
BASE_BRANCH="main"
BRANCH="GH-123"
WORKTREE_PATH="$HOME/worktrees/example-owner/example-repository/$BRANCH"

git fetch origin "$BASE_BRANCH"
mkdir -p "$(dirname "$WORKTREE_PATH")"
git worktree add -b "$BRANCH" "$WORKTREE_PATH" "origin/$BASE_BRANCH"
```

If the branch already exists locally but has no registered worktree:

```bash
git worktree add "$WORKTREE_PATH" "$BRANCH"
```

If only the remote branch exists:

```bash
git fetch origin "$BRANCH"
git worktree add --track -b "$BRANCH" "$WORKTREE_PATH" "origin/$BRANCH"
```

Reuse an exact registered branch worktree when it already exists. Never use
substring matching or bypass Git's one-branch-per-worktree protection.

For detached pull-request inspection:

```bash
PR_PATH="$HOME/worktrees/example-owner/example-repository/PR-77"
git fetch origin "pull/77/head"
git worktree add --detach "$PR_PATH" FETCH_HEAD
```

Detached `PR-<number>` lanes are inspection-only. For repair, resolve the
pull request's same-repository head branch and reuse or attach that durable
branch instead.

## Writable-Lane Environment Link

The clone's primary checkout owns the shared repository-root `.env`. Link that
stable source into writable lanes by default when it exists:

```bash
SOURCE_ENV="$PRIMARY_ROOT/.env"
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
  echo "No primary-checkout .env exists; no environment file was linked."
fi
```

Reusing a writable lane must repeat the exact source and destination
validation and create the link when missing. Omit the link intentionally when
isolation is required.

Never copy `.env`, overwrite a destination `.env`, or automatically share
`.envrc`; `.envrc` is executable shell configuration. A regular destination,
broken link, or unexpected symlink is a collision and must stop the operation.
Detached PR inspection and migration do not create environment links.

## Inspection, Migration, and Removal

Listing is read-only:

```bash
git worktree list --porcelain
```

Review stale administrative metadata before pruning:

```bash
git worktree prune --dry-run --verbose
git worktree prune --verbose
```

Move a registered legacy worktree only after validating its exact source,
destination, and every collision:

```bash
git worktree move "/exact/registered/source" \
  "$HOME/worktrees/example-owner/example-repository/GH-123"
```

Migration preserves dirty contents and existing environment files or links.
Never use ordinary `mv`, stash, reset, clean, or force.

Before removal, prove the target is an exact registered path, is not the
current checkout, has no tracked, untracked, ignored, dirty, or unpublished
state, and has no unsafe `.env`. A verified `.env` symlink to the primary
checkout's exact source is the sole narrow exception:

1. Verify the destination is a symlink whose target matches
   `$PRIMARY_ROOT/.env`.
2. Unlink only that destination symlink.
3. Run ordinary non-force `git worktree remove "/exact/registered/path"`.
4. If Git removal fails, restore the same symlink.

Refuse regular `.env` files, unexpected symlinks, and every other dirty,
ignored, or unpublished item. Never use `--force`, reset, clean, stash, or
branch deletion.

## Scope Boundary

Worktree tooling manages checkout paths, branches, native Git operations, and
the narrow writable-lane `.env` link. Runtime services, databases, ports,
Temporal state, process supervision, application startup, and sibling
repositories remain outside its scope.
