package worktree

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type repository struct {
	top         string
	owner       string
	name        string
	projectRoot string
}

func (a *App) repository(ctx context.Context, cwd string) (repository, error) {
	top, err := a.gitText(ctx, cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return repository{}, fmt.Errorf("not inside a Git worktree: %w", err)
	}
	owner, name, err := a.projectIdentity(ctx, top)
	if err != nil {
		return repository{}, err
	}
	root, err := a.baseRoot()
	if err != nil {
		return repository{}, err
	}
	return repository{
		top:         top,
		owner:       strings.ToLower(owner),
		name:        strings.ToLower(name),
		projectRoot: filepath.Join(root, strings.ToLower(owner), strings.ToLower(name)),
	}, nil
}

func (a *App) baseRoot() (string, error) {
	if configured := strings.TrimSpace(a.getenv("GIT_WT_ROOT")); configured != "" {
		if !filepath.IsAbs(configured) {
			return "", fmt.Errorf("GIT_WT_ROOT must be an absolute path")
		}
		return filepath.Clean(configured), nil
	}
	home, err := a.homeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, "worktrees"), nil
}

func (a *App) projectIdentity(ctx context.Context, cwd string) (string, string, error) {
	configuredOwner, _ := a.gitText(ctx, cwd, "config", "--get", "wt.owner")
	configuredRepo, _ := a.gitText(ctx, cwd, "config", "--get", "wt.repository")
	if configuredOwner != "" || configuredRepo != "" {
		if configuredOwner == "" || configuredRepo == "" {
			return "", "", fmt.Errorf("both wt.owner and wt.repository must be configured together")
		}
		if !isSafeProjectPart(configuredOwner) || !isSafeProjectPart(configuredRepo) {
			return "", "", fmt.Errorf("wt.owner and wt.repository may contain only letters, digits, dot, underscore, and hyphen")
		}
		return configuredOwner, configuredRepo, nil
	}

	remote, err := a.gitText(ctx, cwd, "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("read origin URL (or configure wt.owner and wt.repository): %w", err)
	}
	owner, repo, err := parseRemoteIdentity(remote)
	if err != nil {
		return "", "", fmt.Errorf("derive owner/repository from origin %q (or configure wt.owner and wt.repository): %w", remote, err)
	}
	return owner, repo, nil
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
	owner, repo := parts[len(parts)-2], parts[len(parts)-1]
	if !isSafeProjectPart(owner) || !isSafeProjectPart(repo) {
		return "", "", fmt.Errorf("unsafe owner or repository path segment")
	}
	return owner, repo, nil
}

func canonicalLanePath(repo repository, lane string) (string, error) {
	cleanLane, err := validateLane(lane)
	if err != nil {
		return "", err
	}
	path := filepath.Join(repo.projectRoot, filepath.FromSlash(cleanLane))
	relative, err := filepath.Rel(repo.projectRoot, path)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("lane %q escapes the project worktree directory", lane)
	}
	return path, nil
}

func (a *App) registeredWorktree(
	ctx context.Context,
	repositoryRoot string,
	path string,
) (*worktreeEntry, error) {
	entries, err := a.worktrees(ctx, repositoryRoot)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		if samePath(entries[i].path, path) {
			return &entries[i], nil
		}
	}
	return nil, fmt.Errorf("%s is not an exact registered worktree for this clone", path)
}

func validateLane(lane string) (string, error) {
	if lane == "" || filepath.IsAbs(lane) || strings.ContainsRune(lane, '\x00') || strings.Contains(lane, "\\") {
		return "", fmt.Errorf("invalid lane %q", lane)
	}
	for _, r := range lane {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("lane contains a control character")
		}
	}
	parts := strings.Split(lane, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("lane %q contains an empty, dot, or parent component", lane)
		}
	}
	return strings.Join(parts, "/"), nil
}

func parseNumber(value string, pattern *regexp.Regexp, kind string) (int, error) {
	match := pattern.FindStringSubmatch(value)
	if match == nil {
		return 0, fmt.Errorf("%s must be a positive number or %s-<number>", kind, strings.ToUpper(kind))
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("parse %s number: %w", kind, err)
	}
	return number, nil
}
