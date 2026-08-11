package worktreeprep

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// PrepareBranch attaches or reuses the exact writable worktree for branch.
func (preparer *Preparer) PrepareBranch(
	ctx context.Context,
	cwd string,
	branch string,
	linkEnvironment bool,
) (Prepared, error) {
	repo, err := preparer.repository(ctx, cwd)
	if err != nil {
		return Prepared{}, err
	}
	if err := validateLane(branch); err != nil {
		return Prepared{}, err
	}
	if _, err := preparer.git(ctx, repo.top, "check-ref-format", "--branch", branch); err != nil {
		return Prepared{}, fmt.Errorf("invalid branch %q: %w", branch, err)
	}
	if !preparer.refExists(ctx, repo.top, "refs/heads/"+branch) {
		if err := preparer.fetchOrigin(ctx, repo.top); err != nil {
			return Prepared{}, err
		}
	}
	return preparer.prepareBranch(ctx, repo, branch, linkEnvironment)
}

// PreparePullRequest prepares an open same-repository pull request's head branch.
func (preparer *Preparer) PreparePullRequest(
	ctx context.Context,
	cwd string,
	number int,
	linkEnvironment bool,
) (PullRequest, error) {
	if number < 1 {
		return PullRequest{}, fmt.Errorf("pull request number must be positive")
	}
	repo, err := preparer.repository(ctx, cwd)
	if err != nil {
		return PullRequest{}, err
	}
	repositoryName := repo.owner + "/" + repo.name
	pullRequest, err := preparer.resolvePullRequest(ctx, repo.top, repositoryName, number)
	if err != nil {
		return PullRequest{}, err
	}
	if pullRequest.IsCrossRepository {
		return PullRequest{}, fmt.Errorf(
			"PR %d is from a fork; automatic repair supports same-repository head branches only",
			number,
		)
	}
	if !strings.EqualFold(pullRequest.State, "OPEN") {
		return PullRequest{}, fmt.Errorf("PR %d is %s, not open", number, strings.ToLower(pullRequest.State))
	}
	if pullRequest.HeadRefName == "" {
		return PullRequest{}, fmt.Errorf("PR %d has no head branch", number)
	}
	if strings.HasPrefix(strings.ToUpper(pullRequest.HeadRefName), "PR-") {
		return PullRequest{}, fmt.Errorf("PR %d head %q is not a durable branch", number, pullRequest.HeadRefName)
	}
	if err := preparer.fetchOrigin(ctx, repo.top); err != nil {
		return PullRequest{}, err
	}
	prepared, err := preparer.prepareBranch(ctx, repo, pullRequest.HeadRefName, linkEnvironment)
	if err != nil {
		return PullRequest{}, err
	}
	return PullRequest{
		Prepared:   prepared,
		Repository: repositoryName,
		Number:     number,
		URL:        pullRequest.URL,
		HeadRefOID: pullRequest.HeadRefOID,
	}, nil
}

func (preparer *Preparer) prepareBranch(
	ctx context.Context,
	repo repository,
	branch string,
	linkEnvironment bool,
) (Prepared, error) {
	entries, err := preparer.worktrees(ctx, repo.top)
	if err != nil {
		return Prepared{}, err
	}
	for _, entry := range entries {
		if entry.branch != branch {
			continue
		}
		if samePath(entry.path, repo.primary) {
			return Prepared{}, fmt.Errorf(
				"branch %q is checked out in the protected primary worktree: %s",
				branch,
				entry.path,
			)
		}
		if err := ensureEnvironmentLinks(repo.primary, entry.path, linkEnvironment); err != nil {
			return Prepared{}, err
		}
		return Prepared{Path: entry.path, Branch: branch}, nil
	}
	local := preparer.refExists(ctx, repo.top, "refs/heads/"+branch)
	remote := preparer.refExists(ctx, repo.top, "refs/remotes/origin/"+branch)
	if !local && !remote {
		return Prepared{}, fmt.Errorf("branch %q does not exist locally or on origin", branch)
	}
	destination, err := canonicalLanePath(repo, branch)
	if err != nil {
		return Prepared{}, err
	}
	exists, err := preparer.pathExists(destination)
	if err != nil {
		return Prepared{}, fmt.Errorf("inspect destination %s: %w", destination, err)
	}
	if exists {
		return Prepared{}, fmt.Errorf(
			"destination already exists but is not the expected registered worktree: %s",
			destination,
		)
	}
	if err := preparer.mkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return Prepared{}, fmt.Errorf("create project worktree directory: %w", err)
	}
	if local {
		_, err = preparer.git(ctx, repo.top, "worktree", "add", destination, branch)
	} else {
		_, err = preparer.git(
			ctx,
			repo.top,
			"worktree",
			"add",
			"--track",
			"-b",
			branch,
			destination,
			"origin/"+branch,
		)
	}
	if err != nil {
		return Prepared{}, err
	}
	if err := ensureEnvironmentLinks(repo.primary, destination, linkEnvironment); err != nil {
		return Prepared{}, preparer.rollbackNewWorktree(ctx, repo.top, destination, err)
	}
	return Prepared{Path: destination, Branch: branch, Created: true}, nil
}

func (preparer *Preparer) rollbackNewWorktree(
	ctx context.Context,
	repositoryRoot string,
	destination string,
	setupErr error,
) error {
	if _, err := preparer.git(ctx, repositoryRoot, "worktree", "remove", destination); err != nil {
		return fmt.Errorf(
			"%w; additionally failed to remove newly created worktree %s: %v",
			setupErr,
			destination,
			err,
		)
	}
	return setupErr
}

func (preparer *Preparer) resolvePullRequestWithCLI(
	ctx context.Context,
	cwd string,
	repositoryName string,
	number int,
) (pullRequest, error) {
	output, err := preparer.command(
		ctx,
		cwd,
		"gh",
		"pr",
		"view",
		strconv.Itoa(number),
		"--repo",
		repositoryName,
		"--json",
		"headRefName,headRefOid,isCrossRepository,state,url",
	)
	if err != nil {
		return pullRequest{}, err
	}
	var result pullRequest
	if err := json.Unmarshal(output, &result); err != nil {
		return pullRequest{}, fmt.Errorf("decode gh PR response: %w", err)
	}
	return result, nil
}
