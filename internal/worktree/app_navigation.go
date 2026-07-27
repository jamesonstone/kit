package worktree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

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

func (a *App) enterHome(ctx context.Context, cwd string) error {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	return a.runShell(ctx, repo.primary)
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
