package worktree

import (
	"context"
	"fmt"
	"path/filepath"
)

// prepareOrCreateBranch opens an existing branch or creates it from origin's
// default branch when it does not exist locally or on origin.
func (a *App) prepareOrCreateBranch(
	ctx context.Context,
	repo repository,
	branch string,
	linkEnv bool,
) (PreparedWorktree, error) {
	if a.refExists(ctx, repo.top, "refs/heads/"+branch) {
		return a.prepareBranch(ctx, repo, branch, linkEnv)
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return PreparedWorktree{}, err
	}
	if a.refExists(ctx, repo.top, "refs/remotes/origin/"+branch) {
		return a.prepareBranch(ctx, repo, branch, linkEnv)
	}

	base, err := a.remoteDefaultBranch(ctx, repo.top)
	if err != nil {
		return PreparedWorktree{}, err
	}
	destination, err := canonicalLanePath(repo, branch)
	if err != nil {
		return PreparedWorktree{}, err
	}
	if err := a.ensureDestinationAvailable(destination); err != nil {
		return PreparedWorktree{}, err
	}
	if err := a.mkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return PreparedWorktree{}, fmt.Errorf("create project worktree directory: %w", err)
	}
	if _, err := a.git(ctx, repo.top, "worktree", "add", "-b", branch, destination, "refs/remotes/origin/"+base); err != nil {
		return PreparedWorktree{}, err
	}
	if err := a.ensureEnvironmentLinks(repo.primary, destination, linkEnv); err != nil {
		return PreparedWorktree{}, err
	}
	return PreparedWorktree{Path: destination, Branch: branch, Created: true}, nil
}
