package worktree

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	issueLanePattern = regexp.MustCompile(`(?i)^(?:GH-)?([1-9][0-9]*)$`)
	prLanePattern    = regexp.MustCompile(`(?i)^(?:PR-)?([1-9][0-9]*)$`)
	safeProjectPart  = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
)

func isSafeProjectPart(value string) bool {
	return value != "." && value != ".." && safeProjectPart.MatchString(value)
}

const usage = `Usage: git wt [command] [arguments]

Safe worktrees live at ~/worktrees/<owner>/<repository>/<lane>.

Commands:
  <branch>                         Open or create a worktree for a branch
  issue <number> [--no-link-env]   Create or reuse durable issue lane GH-<number>
  add <branch> [--no-link-env]     Open an existing local or origin branch
  pr <number>                      Create or refresh detached inspection lane PR-<number>
  repair <number> [--no-link-env]  Open a same-repository PR's writable head branch
  list [flags]                     List this clone's worktrees without pruning (default)
  sync [--dry-run] [--json]       Reconcile origin and proven merged worktree lanes
  home                             Open a shell in this clone's primary worktree
  root                             Print the canonical linked-worktree directory
  path <lane>                      Print an exact registered lane path for shell navigation
  cd <lane>                        Open an interactive shell in an exact registered lane
  remove <lane|path>               Remove one exact clean, fully-pushed worktree
  prune [--dry-run]                Explicitly prune stale worktree metadata
  migrate [--apply]                Preview or apply legacy flat-directory migration
  help                             Show this help

Environment:
  GIT_WT_ROOT          Override ~/worktrees (primarily for testing)

List flags:
  --sort <attribute>               Sort by updated, state, head, or path
  --root-position <top|bottom>     Pin the primary worktree (default: top)
  --reverse                        Reverse the selected sort order
  --plain                          Print the table instead of opening the selector

List PR# markers:
  - no open PR   NG gh unavailable   RL rate limited   TO timed out   ?? other failure

Safety:
  PR-<number> is detached and inspection-only; use repair for edits.
  Writable lanes link the primary checkout's .env by default; use --no-link-env for isolation.
  .envrc is never linked automatically.
  remove never forces, deletes a branch, or discards dirty/unpushed state.
  sync removes only exact clean lanes proven merged into origin's default branch.
  sync --dry-run performs no fetch, ref update, removal, deletion, or pruning.
  migrate previews by default and uses git worktree move when applied.
  No command starts applications or manages databases, ports, or runtime services.
  No command stashes, resets, cleans, or force-removes worktrees.`

type commandFunc func(context.Context, string, string, ...string) ([]byte, error)

// PR identifies the writable head of a pull request.
type PR struct {
	HeadRefName       string `json:"headRefName"`
	HeadRefOID        string `json:"headRefOid"`
	IsCrossRepository bool   `json:"isCrossRepository"`
	State             string `json:"state"`
	URL               string `json:"url"`
}

// PreparedWorktree describes an exact writable branch worktree.
type PreparedWorktree struct {
	Path    string
	Branch  string
	Created bool
}

// PullRequestRepair describes the writable same-repository head of a pull request.
type PullRequestRepair struct {
	PreparedWorktree
	Repository string
	Number     int
	URL        string
	HeadRefOID string
}

type resolvePRFunc func(context.Context, string, string, int) (PR, error)

// App implements the git-wt command.
type App struct {
	out            io.Writer
	errOut         io.Writer
	run            commandFunc
	homeDir        func() (string, error)
	getenv         func(string) string
	readDir        func(string) ([]os.DirEntry, error)
	mkdirAll       func(string, os.FileMode) error
	pathExists     func(string) (bool, error)
	lookPath       func(string) (string, error)
	resolvePR      resolvePRFunc
	resolveListPRs listPRResolverFunc
	resolveSyncPRs syncPRResolverFunc
	listPRTimeout  time.Duration
	runShell       func(context.Context, string) error
	isTerminal     func() bool
	selectList     listSelectorFunc
	stdin          io.Reader
}

// NewApp creates an App backed by the local Git and GitHub CLIs.
func NewApp(out, errOut io.Writer) *App {
	app := &App{
		out:           out,
		errOut:        errOut,
		stdin:         os.Stdin,
		run:           runCommand,
		homeDir:       os.UserHomeDir,
		getenv:        os.Getenv,
		readDir:       os.ReadDir,
		mkdirAll:      os.MkdirAll,
		lookPath:      exec.LookPath,
		listPRTimeout: defaultListPRTimeout,
		pathExists: func(path string) (bool, error) {
			_, err := os.Lstat(path)
			if err == nil {
				return true, nil
			}
			if errors.Is(err, os.ErrNotExist) {
				return false, nil
			}
			return false, err
		},
	}
	app.resolvePR = app.resolvePullRequest
	app.resolveListPRs = app.resolveListPullRequests
	app.resolveSyncPRs = app.resolveSyncPullRequests
	app.runShell = runInteractiveShell
	app.isTerminal, app.selectList = newListInteraction(out)
	return app
}

// Run executes one command from cwd.
func (a *App) Run(ctx context.Context, cwd string, args []string) error {
	if len(args) == 0 {
		return a.list(ctx, cwd, nil)
	}

	switch args[0] {
	case "help", "-h", "--help":
		if len(args) != 1 {
			return fmt.Errorf("help accepts no arguments")
		}
		return a.writef("%s\n", usage)
	case "root":
		if len(args) != 1 {
			return fmt.Errorf("root accepts no arguments")
		}
		repo, err := a.repository(ctx, cwd)
		if err != nil {
			return err
		}
		return a.writef("%s\n", repo.projectRoot)
	case "home":
		if len(args) != 1 {
			return fmt.Errorf("home accepts no arguments")
		}
		return a.enterHome(ctx, cwd)
	case "path":
		if len(args) != 2 {
			return fmt.Errorf("usage: git wt path <lane>")
		}
		return a.lanePath(ctx, cwd, args[1])
	case "cd", "enter":
		if len(args) != 2 {
			return fmt.Errorf("usage: git wt cd <lane>")
		}
		return a.enterLane(ctx, cwd, args[1])
	case "list":
		return a.list(ctx, cwd, args[1:])
	case "sync":
		return a.sync(ctx, cwd, args[1:])
	case "issue":
		value, linkEnv, err := writableLaneArgs("issue", "number", args[1:])
		if err != nil {
			return err
		}
		return a.issue(ctx, cwd, value, linkEnv)
	case "add":
		value, linkEnv, err := writableLaneArgs("add", "branch", args[1:])
		if err != nil {
			return err
		}
		return a.add(ctx, cwd, value, linkEnv)
	case "pr":
		if len(args) != 2 {
			return fmt.Errorf("usage: git wt pr <number>")
		}
		return a.pr(ctx, cwd, args[1])
	case "repair":
		value, linkEnv, err := writableLaneArgs("repair", "number", args[1:])
		if err != nil {
			return err
		}
		return a.repair(ctx, cwd, value, linkEnv)
	case "remove":
		if len(args) != 2 {
			return fmt.Errorf("usage: git wt remove <lane|path>")
		}
		return a.remove(ctx, cwd, args[1])
	case "prune":
		return a.prune(ctx, cwd, args[1:])
	case "migrate":
		return a.migrate(ctx, cwd, args[1:])
	default:
		if len(args) == 1 {
			return a.openOrCreateLane(ctx, cwd, args[0])
		}
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func (a *App) prune(ctx context.Context, cwd string, args []string) error {
	dryRun := false
	switch {
	case len(args) == 0:
	case len(args) == 1 && args[0] == "--dry-run":
		dryRun = true
	default:
		return fmt.Errorf("usage: git wt prune [--dry-run]")
	}
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	gitArgs := []string{"worktree", "prune", "--verbose"}
	if dryRun {
		gitArgs = append(gitArgs, "--dry-run")
	}
	output, err := a.git(ctx, repo.top, gitArgs...)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(output)) > 0 {
		if err := a.writef("%s", output); err != nil {
			return err
		}
	}
	if dryRun {
		return a.writef("Dry run complete; no worktree metadata was pruned.\n")
	}
	return a.writef("Pruned stale worktree metadata.\n")
}

func (a *App) writef(format string, args ...any) error {
	if _, err := fmt.Fprintf(a.out, format, args...); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
