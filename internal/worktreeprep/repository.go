package worktreeprep

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var safeProjectPart = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type repository struct {
	top         string
	primary     string
	owner       string
	name        string
	projectRoot string
}

func (preparer *Preparer) repository(ctx context.Context, cwd string) (repository, error) {
	top, err := preparer.gitText(ctx, cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return repository{}, fmt.Errorf("not inside a Git worktree: %w", err)
	}
	worktrees, err := preparer.worktrees(ctx, top)
	if err != nil {
		return repository{}, fmt.Errorf("list repository worktrees: %w", err)
	}
	if len(worktrees) == 0 {
		return repository{}, fmt.Errorf("repository has no primary worktree")
	}
	remote, err := preparer.gitText(ctx, top, "remote", "get-url", "origin")
	if err != nil {
		return repository{}, fmt.Errorf("read origin URL: %w", err)
	}
	owner, name, err := parseRemoteIdentity(remote)
	if err != nil {
		return repository{}, fmt.Errorf("derive owner/repository from origin %q: %w", remote, err)
	}
	home, err := preparer.homeDir()
	if err != nil {
		return repository{}, fmt.Errorf("determine home directory: %w", err)
	}
	owner = strings.ToLower(owner)
	name = strings.ToLower(name)
	return repository{
		top:         top,
		primary:     filepath.Clean(worktrees[0].path),
		owner:       owner,
		name:        name,
		projectRoot: filepath.Join(home, "worktrees", owner, name),
	}, nil
}

func parseRemoteIdentity(remote string) (string, string, error) {
	path := ""
	if strings.Contains(remote, "://") {
		parsed, err := url.Parse(remote)
		if err != nil {
			return "", "", err
		}
		path = parsed.Path
	} else if colon := strings.Index(remote, ":"); colon > 0 && !filepath.IsAbs(remote) {
		path = remote[colon+1:]
	} else {
		path = remote
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(path, ".git"), "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("expected an origin path ending in owner/repository")
	}
	owner, name := parts[len(parts)-2], parts[len(parts)-1]
	if !isSafeProjectPart(owner) || !isSafeProjectPart(name) {
		return "", "", fmt.Errorf("unsafe owner or repository path segment")
	}
	return owner, name, nil
}

func isSafeProjectPart(value string) bool {
	return value != "." && value != ".." && safeProjectPart.MatchString(value)
}

func canonicalLanePath(repo repository, lane string) (string, error) {
	if err := validateLane(lane); err != nil {
		return "", err
	}
	path := filepath.Join(repo.projectRoot, filepath.FromSlash(lane))
	relative, err := filepath.Rel(repo.projectRoot, path)
	if err != nil || relative == "." || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("lane %q escapes the project worktree directory", lane)
	}
	return path, nil
}

func validateLane(lane string) error {
	if lane == "" || filepath.IsAbs(lane) || strings.ContainsRune(lane, '\x00') ||
		strings.Contains(lane, "\\") {
		return fmt.Errorf("invalid lane %q", lane)
	}
	for _, value := range strings.Split(lane, "/") {
		if value == "" || value == "." || value == ".." {
			return fmt.Errorf("lane %q contains an empty, dot, or parent component", lane)
		}
	}
	for _, value := range lane {
		if value < 0x20 || value == 0x7f {
			return fmt.Errorf("lane contains a control character")
		}
	}
	return nil
}
