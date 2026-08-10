# Git Worktrees

## Policy Authority

Native `git worktree` commands and ordinary filesystem operations define this
portable workflow. Kit prepares writable repair lanes through the same native
operations, but worktree lifecycle policy never depends on an external wrapper,
alias, editor integration, or plugin.

## Mental Model

A worktree is another checkout attached to the same Git clone.

Each worktree has separate working files, an index, and `HEAD`. All worktrees
of the clone share commits, objects, refs, remotes, most Git configuration, and
stash entries. A worktree protects one checkout from unrelated file and branch
changes; it is not a second clone or an isolation boundary for shared Git
state.

Keep the primary checkout on the protected default branch. Develop and test in
assigned durable lanes, and never check one branch out in two worktrees at
once.

## Canonical Hierarchy

Keep linked worktrees outside the source clone:

```text
~/worktrees/<owner>/<repository>/<lane>
```

Owner and repository names are lowercase. Durable issue lanes use exact
uppercase `GH-<number>`. Detached pull-request inspection lanes use exact
uppercase `PR-<number>`.

Do not put linked worktrees inside a repository, including under
`.worktrees/`. External placement prevents recursive tooling, watchers,
search, backup rules, builds, and cleanup from treating one checkout as another
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

Detached `PR-<number>` lanes are inspection-only. For repair, resolve the pull
request's same-repository head branch and reuse or attach that durable branch
instead.

## Target-Aware Kit Repair Commands

When a Kit command already identifies a pull request or failed branch, use that
target to resolve the writable lane automatically. The user does not need to
navigate to a worktree before running `kit pr fix`, PR-backed dispatch or
review-loop commands, `kit loop review --pr`, or `kit ci --dispatch`.

Resolution proves the current clone owns the requested repository, uses the
exact same-repository PR head or exact diagnosed branch, and consults
`git worktree list --porcelain` for registered ownership. It may fetch
`origin` and add or attach the canonical writable lane, but it must not choose
by recency, substring, fuzzy matching, or interactive selection.

Before generating or running repair instructions, record the remote target
head, local `HEAD`, exact worktree path, and push target. If the worktree is
dirty, show `git status --porcelain` and ask whether the existing changes
belong in the repair:

- `include` makes the existing diff part of the full repair review and
  validation scope.
- `exclude` requires preserving those paths, avoiding staging or modification,
  and stopping when the requested repair overlaps them.

Neither choice authorizes stash, reset, clean, rebase, force operations, or
discarding user work. Prompt-producing commands remain prompt-producing after
lane preparation; staging, commits, pushes, comments, review-thread resolution,
and PR delivery retain their explicit gates.

## Writable-Lane Environment Links

The clone's primary checkout owns the shared repository-root `.env` and
`.envrc`. Link each stable source into writable lanes by default when it
exists:

```bash
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

ensure_environment_link() {
  name="$1"
  source_path="$PRIMARY_ROOT/$name"
  destination_path="$WORKTREE_PATH/$name"

  if [ -L "$destination_path" ]; then
    if [ ! -e "$destination_path" ]; then
      echo "ABORT: destination $name is a broken link" >&2
      exit 1
    fi
    resolved_target="$(resolve_link_target "$destination_path")" || {
      echo "ABORT: destination $name is unreadable" >&2
      exit 1
    }
    if [ "$resolved_target" != "$source_path" ]; then
      echo "ABORT: destination $name points to an unexpected target" >&2
      exit 1
    fi
  elif [ -e "$destination_path" ]; then
    if [ "$name" = ".envrc" ]; then
      echo "Preserving existing destination .envrc: $destination_path"
      return
    fi
    echo "ABORT: destination $name already exists: $destination_path" >&2
    exit 1
  elif [ -f "$source_path" ]; then
    ln -s "$source_path" "$destination_path"
  else
    echo "No primary-checkout $name exists; no $name link was created."
  fi
}

ensure_environment_link ".env"
ensure_environment_link ".envrc"
```

Reusing a writable lane repeats each exact source and destination validation and
creates missing links. Omit both links intentionally when isolation is
required.

Never copy environment contents or overwrite destination material. A regular
destination `.env` and any broken or unexpected environment symlink are
collisions that stop the operation. Preserve a regular destination `.envrc`,
which may be tracked by Git or owned by the user.

`.envrc` is executable shell configuration. Review the primary source before
sharing it, and retain direnv's separate path-specific approval by running
`direnv allow "$WORKTREE_PATH"` after inspecting a newly linked lane. Detached
PR inspection and migration do not create environment links.

## Inspection, Pruning, And Removal

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

Before removal, prove the target is an exact registered path, is not the current
checkout, has no tracked, untracked, ignored, dirty, or unpublished state, and
has no unsafe environment material. Verified `.env` and `.envrc` symlinks
to matching primary-checkout sources are the sole narrow exceptions:

1. Verify each environment destination is a symlink whose target matches the
   same name beneath `$PRIMARY_ROOT`.
2. Unlink only those verified destination symlinks.
3. Run ordinary non-force `git worktree remove "/exact/registered/path"`.
4. If Git removal fails, restore every removed symlink.

Refuse regular ignored environment files, unexpected symlinks, and every other
dirty, ignored, or unpublished item. A clean tracked `.envrc` remains ordinary
Git-managed content. Never use `--force`, reset, clean, stash, or branch deletion
as part of worktree removal.

## Scope Boundary

Worktree preparation manages checkout paths, branches, native Git operations,
and the narrow writable-lane environment links. Runtime services, databases, ports,
Temporal state, process supervision, application startup, and sibling
repositories remain outside its scope.
