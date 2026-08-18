// Package worktreeprep prepares canonical writable worktrees for Kit repair workflows.
package worktreeprep

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

type commandFunc func(context.Context, string, string, ...string) ([]byte, error)

type pullRequest struct {
	HeadRefName       string `json:"headRefName"`
	HeadRefOID        string `json:"headRefOid"`
	IsCrossRepository bool   `json:"isCrossRepository"`
	State             string `json:"state"`
	URL               string `json:"url"`
}

// Prepared describes an exact writable branch worktree.
type Prepared struct {
	Path    string
	Branch  string
	Created bool
}

// Location describes the current checkout and its primary-worktree ownership.
type Location struct {
	Path        string
	PrimaryPath string
	InsideGit   bool
	IsPrimary   bool
}

// PullRequest describes a prepared same-repository pull-request head.
type PullRequest struct {
	Prepared
	Repository string
	Number     int
	URL        string
	HeadRefOID string
}

type resolvePullRequestFunc func(context.Context, string, string, int) (pullRequest, error)

// Preparer resolves and prepares canonical writable worktrees.
type Preparer struct {
	run                commandFunc
	homeDir            func() (string, error)
	mkdirAll           func(string, os.FileMode) error
	pathExists         func(string) (bool, error)
	resolvePullRequest resolvePullRequestFunc
}

// New creates a preparer backed by local Git, GitHub CLI, and filesystem operations.
func New() *Preparer {
	preparer := &Preparer{
		run:      runCommand,
		homeDir:  os.UserHomeDir,
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
	preparer.resolvePullRequest = preparer.resolvePullRequestWithCLI
	return preparer
}

func runCommand(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = cwd
	return command.CombinedOutput()
}
