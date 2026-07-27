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

In a terminal, the optional `git wt` helper opens the same colorized selector
as `git wt list`, with Git's primary worktree pinned at the top and the
remaining lanes ordered by `LAST UPDATED`. Each timestamp uses the running
user's local timezone and shows the calendar day plus hour and minute, with no
seconds. The primary checkout and every `main` branch row stay bright green
across repositories. Use arrow keys or Tab to move, Enter to open a child shell
in the selected worktree, `h` to open the primary worktree immediately, and
`q` to cancel. The child shell cannot change its parent shell's directory.

Piped or redirected output remains a plain table. Use `--plain` to request the
table from a terminal, `--sort updated|state|head|path` to choose another key,
`--reverse` to invert that ordering, or `--root-position bottom` to pin the
primary checkout below the sorted lanes. For example:

```bash
git wt list --plain --sort path
git wt list --plain --sort state --reverse
git wt list --root-position bottom
git wt home
```

The `PR#` column runs one batched `gh` lookup for open same-repository pull
requests and stops that lookup after two seconds. An exact branch match shows
the pull request number; a successful lookup without a match shows `-`.
Multiple matches appear as ascending comma-separated numbers. Failures never
prevent listing: `NG` means `gh` is unavailable, `RL` means GitHub rate
limiting, `TO` means timeout, and `??` means another lookup or decode failure.

For direct branch navigation, use `git wt <branch>`, for example
`git wt GH-93`. Existing registered lanes open immediately. Missing lanes ask
`do you want to create this worktree? (y/n)`. If the requested branch already
exists locally or on origin, answering `y` attaches that branch in the
canonical owner/repository directory. Only when the branch exists in neither
the local repository nor origin does answering `y` create it from the origin
default branch; `n` exits without changes.
`git wt home` opens the same primary checkout in a child shell from any linked
worktree. Use it when returning to the clone's stable home checkout without
looking up a lane name.

Listing never fetches, prunes, changes Git state, or requires GitHub to
succeed. Its bounded pull-request annotation is read-only and fail-soft. Use
the separate maintenance command when live reconciliation is intended:

```bash
git wt sync --dry-run
git wt sync
git wt sync --json
```

The dry run reads live origin and GitHub state but does not fetch or perform any
local ref, worktree, branch, metadata, symlink, or filesystem mutation.
Ordinary sync fetches and prunes only `origin`; fast-forwards the local default
branch only when it is clean and strictly behind; removes only exact canonical
lanes backed by one same-repository PR merged into that default branch whose
head OID exactly equals local `HEAD`; and then uses ordinary `git branch -d`.
Every dirty, ignored, fork-backed, open, closed-unmerged, wrong-base, missing,
ambiguous, OID-mismatched, detached, legacy, primary, or current lane is
preserved with a reason. The command never stashes, resets, cleans,
force-removes, force-deletes, force-pushes, or deletes a remote branch.

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

## Target-Aware Repair Commands

When a Kit command already identifies a pull request or failed branch, use that
target to resolve the writable lane automatically. The user does not need to
navigate to a worktree before running `kit pr fix`, PR-backed dispatch or
review-loop commands, `kit loop review --pr`, or `kit ci --dispatch`.

Resolution must prove the current clone owns the requested repository, use the
exact same-repository PR head or exact diagnosed branch, and consult
`git worktree list --porcelain` for registered ownership. It may fetch `origin`
and add or attach the canonical writable lane, but must not choose by recency,
substring, fuzzy matching, or interactive selection.

Before generating or running repair instructions, record the remote target
head, local `HEAD`, exact worktree path, and push target. If the worktree is
dirty, show `git status --porcelain` and ask whether the existing changes belong
in the repair:

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
`.envrc`. Link each stable source into writable lanes by default when it exists:

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

Reusing a writable lane must repeat each exact source and destination
validation and create missing links. Omit both links intentionally when
isolation is required.

Never copy environment contents or overwrite destination material. A regular
destination `.env` and any broken or unexpected environment symlink are
collisions that must stop the operation. Preserve a regular destination
`.envrc`, which may be tracked by Git or owned by the user.

`.envrc` is executable shell configuration. Review the primary source before
sharing it, and retain direnv's separate path-specific approval by running
`direnv allow "$WORKTREE_PATH"` after inspecting a newly linked lane. Detached
PR inspection and migration do not create environment links.

## Inspection, Synchronization, Migration, and Removal

Listing is read-only:

```bash
git worktree list --porcelain
```

Review stale administrative metadata before pruning:

```bash
git worktree prune --dry-run --verbose
git worktree prune --verbose
```

`git wt sync` is the explicit higher-level maintenance path described above.
GitHub and fetch failures fail closed. A failure for one candidate does not
prevent an independently proven-safe candidate from being processed, but any
operation failure makes the overall command exit nonzero after its complete
human or JSON report.

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
state, and has no unsafe environment material. Verified `.env` and `.envrc`
symlinks to the matching primary-checkout sources are the sole narrow
exceptions:

1. Verify each environment destination is a symlink whose target matches the
   same name beneath `$PRIMARY_ROOT`.
2. Unlink only those verified destination symlinks.
3. Run ordinary non-force `git worktree remove "/exact/registered/path"`.
4. If Git removal fails, restore every removed symlink.

Refuse regular ignored environment files, unexpected symlinks, and every other
dirty, ignored, or unpublished item. A clean tracked `.envrc` remains ordinary
Git-managed content. Manual `git wt remove` never uses `--force`,
reset, clean, stash, or branch deletion. Sync uses its stricter merged-PR and
exact-head proof instead of upstream/ahead proof, and only after successful
worktree removal attempts ordinary local `git branch -d`.

## Scope Boundary

Worktree tooling manages checkout paths, branches, native Git operations, and
the narrow writable-lane environment links. Runtime services, databases, ports,
Temporal state, process supervision, application startup, and sibling
repositories remain outside its scope.
