package worktree

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
)

var (
	issueLanePattern = regexp.MustCompile(`(?i)^(?:GH-)?([1-9][0-9]*)$`)
	prLanePattern    = regexp.MustCompile(`(?i)^(?:PR-)?([1-9][0-9]*)$`)
	safeProjectPart  = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
)

func isSafeProjectPart(value string) bool {
	return value != "." && value != ".." && safeProjectPart.MatchString(value)
}

const usage = `Usage: git wt <command> [arguments]

Safe worktrees live at ~/worktrees/<owner>/<repository>/<lane>.

Commands:
  issue <number> [--no-link-env]   Create or reuse durable issue lane GH-<number>
  add <branch> [--no-link-env]     Open an existing local or origin branch
  pr <number>                      Create or refresh detached inspection lane PR-<number>
  repair <number> [--no-link-env]  Open a same-repository PR's writable head branch
  list                             List this clone's worktrees without pruning
  root                             Print this repository's canonical worktree directory
  remove <lane|path>               Remove one exact clean, fully-pushed worktree
  prune [--dry-run]                Explicitly prune stale worktree metadata
  migrate [--apply]                Preview or apply legacy flat-directory migration
  help                             Show this help

Environment:
  GIT_WT_ROOT          Override ~/worktrees (primarily for testing)

Safety:
  PR-<number> is detached and inspection-only; use repair for edits.
  Writable lanes link the invoking checkout's .env by default; use --no-link-env for isolation.
  .envrc is never linked automatically.
  remove never forces, deletes a branch, or discards dirty/unpushed state.
  migrate previews by default and uses git worktree move when applied.
  No command starts applications or manages databases, ports, or runtime services.
  No command stashes, resets, cleans, or force-removes worktrees.`

type commandFunc func(context.Context, string, string, ...string) ([]byte, error)

// PR identifies the writable head of a pull request.
type PR struct {
	HeadRefName       string `json:"headRefName"`
	IsCrossRepository bool   `json:"isCrossRepository"`
	State             string `json:"state"`
	URL               string `json:"url"`
}

type resolvePRFunc func(context.Context, string, string, int) (PR, error)

// App implements the git-wt command.
type App struct {
	out        io.Writer
	errOut     io.Writer
	run        commandFunc
	homeDir    func() (string, error)
	getenv     func(string) string
	readDir    func(string) ([]os.DirEntry, error)
	mkdirAll   func(string, os.FileMode) error
	pathExists func(string) (bool, error)
	resolvePR  resolvePRFunc
}

// NewApp creates an App backed by the local Git and GitHub CLIs.
func NewApp(out, errOut io.Writer) *App {
	app := &App{
		out:      out,
		errOut:   errOut,
		run:      runCommand,
		homeDir:  os.UserHomeDir,
		getenv:   os.Getenv,
		readDir:  os.ReadDir,
		mkdirAll: os.MkdirAll,
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
	return app
}

// Run executes one command from cwd.
func (a *App) Run(ctx context.Context, cwd string, args []string) error {
	if len(args) == 0 {
		return a.writef("%s\n", usage)
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
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("list accepts no arguments")
		}
		return a.list(ctx, cwd)
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
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func writableLaneArgs(command, placeholder string, args []string) (string, bool, error) {
	commandUsage := fmt.Sprintf("usage: git wt %s <%s> [--no-link-env]", command, placeholder)
	switch {
	case len(args) == 1:
		return args[0], true, nil
	case len(args) == 2 && args[1] == "--no-link-env":
		return args[0], false, nil
	default:
		return "", false, errors.New(commandUsage)
	}
}

func (a *App) list(ctx context.Context, cwd string) error {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		return err
	}
	if err := a.writef("STATE\tHEAD\tPATH\n"); err != nil {
		return err
	}
	for _, entry := range entries {
		state := "clean"
		dirty, statusErr := a.status(ctx, entry.path, false)
		if statusErr != nil {
			state = "unknown"
		} else if dirty != "" {
			state = "dirty"
		}
		head := entry.branch
		if head == "" {
			head = "detached@" + shortOID(entry.head)
		}
		if err := a.writef("%s\t%s\t%s\n", state, head, entry.path); err != nil {
			return err
		}
	}
	return nil
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
