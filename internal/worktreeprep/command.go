package worktreeprep

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type worktreeEntry struct {
	path   string
	branch string
}

func (preparer *Preparer) worktrees(ctx context.Context, cwd string) ([]worktreeEntry, error) {
	output, err := preparer.git(ctx, cwd, "worktree", "list", "--porcelain")
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
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "worktree "):
			current.path = strings.TrimPrefix(line, "worktree ")
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

func (preparer *Preparer) fetchOrigin(ctx context.Context, cwd string) error {
	if _, err := preparer.git(ctx, cwd, "fetch", "--no-tags", "origin"); err != nil {
		return fmt.Errorf("fetch origin: %w", err)
	}
	return nil
}

func (preparer *Preparer) refExists(ctx context.Context, cwd, ref string) bool {
	_, err := preparer.git(ctx, cwd, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

func (preparer *Preparer) gitText(ctx context.Context, cwd string, args ...string) (string, error) {
	output, err := preparer.git(ctx, cwd, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (preparer *Preparer) git(ctx context.Context, cwd string, args ...string) ([]byte, error) {
	return preparer.command(ctx, cwd, "git", args...)
}

func (preparer *Preparer) command(
	ctx context.Context,
	cwd string,
	name string,
	args ...string,
) ([]byte, error) {
	output, err := preparer.run(ctx, cwd, name, args...)
	if err == nil {
		return output, nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil, fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, detail)
}

func samePath(left, right string) bool {
	leftPath := resolvedPath(left)
	rightPath := resolvedPath(right)
	return leftPath == rightPath
}

func resolvedPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return filepath.Clean(resolved)
	}
	parent, parentErr := filepath.EvalSymlinks(filepath.Dir(absolute))
	if parentErr != nil {
		return filepath.Clean(absolute)
	}
	return filepath.Join(parent, filepath.Base(absolute))
}
