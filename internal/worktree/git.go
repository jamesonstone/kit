package worktree

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type worktreeEntry struct {
	path   string
	head   string
	branch string
}

func (a *App) worktrees(ctx context.Context, cwd string) ([]worktreeEntry, error) {
	output, err := a.git(ctx, cwd, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	entries := make([]worktreeEntry, 0)
	var current worktreeEntry
	flush := func() {
		if current.path != "" {
			entries = append(entries, current)
			current = worktreeEntry{}
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse worktree list: %w", err)
	}
	flush()
	return entries, nil
}

func (a *App) remoteDefaultBranch(ctx context.Context, cwd string) (string, error) {
	symbolic, err := a.gitText(ctx, cwd, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	if err == nil && strings.HasPrefix(symbolic, "origin/") {
		return strings.TrimPrefix(symbolic, "origin/"), nil
	}
	output, err := a.git(ctx, cwd, "ls-remote", "--symref", "origin", "HEAD")
	if err != nil {
		return "", fmt.Errorf("discover origin default branch: %w", err)
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "ref: refs/heads/") && strings.HasSuffix(line, "\tHEAD") {
			branch := strings.TrimSuffix(strings.TrimPrefix(line, "ref: refs/heads/"), "\tHEAD")
			if branch == "" {
				break
			}
			if _, err := a.git(ctx, cwd, "fetch", "--no-tags", "origin", "+refs/heads/"+branch+":refs/remotes/origin/"+branch); err != nil {
				return "", err
			}
			return branch, nil
		}
	}
	return "", fmt.Errorf("origin did not advertise a default branch")
}

func (a *App) fetchOrigin(ctx context.Context, cwd string) error {
	if _, err := a.git(ctx, cwd, "fetch", "--no-tags", "origin"); err != nil {
		return fmt.Errorf("fetch origin: %w", err)
	}
	return nil
}

func (a *App) refExists(ctx context.Context, cwd, ref string) bool {
	_, err := a.git(ctx, cwd, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

func (a *App) ensureDestinationAvailable(path string) error {
	exists, err := a.pathExists(path)
	if err != nil {
		return fmt.Errorf("inspect destination %s: %w", path, err)
	}
	if exists {
		return fmt.Errorf("destination already exists but is not the expected registered worktree: %s", path)
	}
	return nil
}

func (a *App) status(ctx context.Context, cwd string, includeIgnored bool) (string, error) {
	args := []string{"status", "--porcelain=v1", "--untracked-files=all"}
	if includeIgnored {
		args = append(args, "--ignored=matching")
	}
	output, err := a.git(ctx, cwd, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (a *App) gitText(ctx context.Context, cwd string, args ...string) (string, error) {
	output, err := a.git(ctx, cwd, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (a *App) git(ctx context.Context, cwd string, args ...string) ([]byte, error) {
	return a.command(ctx, cwd, "git", args...)
}

func (a *App) command(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
	output, err := a.run(ctx, cwd, name, args...)
	if err == nil {
		return output, nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil, fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, detail)
}

func runCommand(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = cwd
	return cmd.CombinedOutput()
}

func samePath(left, right string) bool {
	leftAbs, leftErr := resolvedPath(left)
	rightAbs, rightErr := resolvedPath(right)
	if leftErr != nil || rightErr != nil {
		return filepath.Clean(left) == filepath.Clean(right)
	}
	return leftAbs == rightAbs
}

func resolvedPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return filepath.Clean(resolved), nil
	}
	parent, parentErr := filepath.EvalSymlinks(filepath.Dir(absolute))
	if parentErr != nil {
		return filepath.Clean(absolute), nil
	}
	return filepath.Join(parent, filepath.Base(absolute)), nil
}

func shortOID(oid string) string {
	if len(oid) > 12 {
		return oid[:12]
	}
	return oid
}
