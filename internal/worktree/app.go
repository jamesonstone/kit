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
	"sort"
	"strings"
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

const usage = `Usage: git wt <command> [arguments]

Safe worktrees live at ~/worktrees/<owner>/<repository>/<lane>.

Commands:
  issue <number> [--no-link-env]   Create or reuse durable issue lane GH-<number>
  add <branch> [--no-link-env]     Open an existing local or origin branch
  pr <number>                      Create or refresh detached inspection lane PR-<number>
  repair <number> [--no-link-env]  Open a same-repository PR's writable head branch
  list [flags]                     List this clone's worktrees without pruning
  root                             Print this repository's canonical worktree directory
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
  --reverse                        Reverse the selected sort order

Safety:
  PR-<number> is detached and inspection-only; use repair for edits.
  Writable lanes link the primary checkout's .env by default; use --no-link-env for isolation.
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
	runShell   func(context.Context, string) error
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
	app.runShell = runInteractiveShell
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

func (a *App) enterLane(ctx context.Context, cwd, lane string) error {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	path, err := canonicalLanePath(repo, lane)
	if err != nil {
		return err
	}
	selected, err := a.registeredWorktree(ctx, repo.top, path)
	if err != nil {
		return err
	}
	return a.runShell(ctx, selected.path)
}

func runInteractiveShell(ctx context.Context, dir string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	command := exec.CommandContext(ctx, shell)
	command.Dir = dir
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
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

type listOptions struct {
	sortBy  string
	reverse bool
}

func parseListOptions(args []string) (listOptions, error) {
	options := listOptions{sortBy: "updated"}
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--reverse":
			options.reverse = true
		case args[i] == "--sort":
			if i+1 >= len(args) {
				return listOptions{}, errors.New("usage: git wt list [--sort <updated|state|head|path>] [--reverse]")
			}
			i++
			options.sortBy = strings.ToLower(args[i])
		case strings.HasPrefix(args[i], "--sort="):
			options.sortBy = strings.ToLower(strings.TrimPrefix(args[i], "--sort="))
		default:
			return listOptions{}, fmt.Errorf("unknown list flag %q", args[i])
		}
	}
	if options.sortBy != "updated" && options.sortBy != "state" && options.sortBy != "head" && options.sortBy != "path" {
		return listOptions{}, fmt.Errorf("unsupported list sort %q (want updated, state, head, or path)", options.sortBy)
	}
	return options, nil
}

func (a *App) list(ctx context.Context, cwd string, args []string) error {
	options, err := parseListOptions(args)
	if err != nil {
		return err
	}
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		return err
	}
	for i := range entries {
		entries[i].updatedText = "unknown"
		updated, updateErr := a.gitText(ctx, entries[i].path, "log", "-1", "--format=%cI", "HEAD")
		if updateErr == nil {
			if parsed, parseErr := time.Parse(time.RFC3339, updated); parseErr == nil {
				entries[i].lastUpdated = parsed
				entries[i].updatedText = parsed.Format(time.RFC3339)
			}
		}
	}
	for i := range entries {
		entries[i].state = "clean"
		dirty, statusErr := a.status(ctx, entries[i].path, false)
		if statusErr != nil {
			entries[i].state = "unknown"
		} else if dirty != "" {
			entries[i].state = "dirty"
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]
		var less bool
		var equal bool
		switch options.sortBy {
		case "updated":
			less = left.lastUpdated.After(right.lastUpdated)
			equal = left.lastUpdated.Equal(right.lastUpdated)
			if equal {
				less = left.path < right.path
			}
		case "state":
			less = left.state < right.state
			equal = left.state == right.state
			if equal {
				less = left.path < right.path
			}
		case "head":
			less = left.branch < right.branch
			equal = left.branch == right.branch
			if equal {
				less = left.head < right.head
			}
		case "path":
			less = left.path < right.path
		}
		if options.reverse {
			return !less && !equal
		}
		return less
	})
	if err := a.writef("STATE\tHEAD\tLAST UPDATED\tPATH\n"); err != nil {
		return err
	}
	for _, entry := range entries {
		head := entry.branch
		if head == "" {
			head = "detached@" + shortOID(entry.head)
		}
		if err := a.writef("%s\t%s\t%s\t%s\n", entry.state, head, entry.updatedText, entry.path); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) lanePath(ctx context.Context, cwd, lane string) error {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	path, err := canonicalLanePath(repo, lane)
	if err != nil {
		return err
	}
	selected, err := a.registeredWorktree(ctx, repo.top, path)
	if err != nil {
		return err
	}
	return a.writef("%s\n", selected.path)
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
