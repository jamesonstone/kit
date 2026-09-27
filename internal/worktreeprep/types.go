// Package worktreeprep inspects checkout ownership so Kit commands can keep
// file writes out of the primary checkout.
package worktreeprep

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

type commandFunc func(context.Context, string, string, ...string) ([]byte, error)

// Location describes the current checkout and its primary-worktree ownership.
type Location struct {
	Path        string
	PrimaryPath string
	InsideGit   bool
	IsPrimary   bool
}

// Preparer inspects checkouts through local Git.
type Preparer struct {
	run        commandFunc
	pathExists func(string) (bool, error)
}

// New creates an inspector backed by local Git.
func New() *Preparer {
	return &Preparer{run: runCommand, pathExists: pathExists}
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func runCommand(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = cwd
	return command.CombinedOutput()
}
