package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func (a *App) issue(ctx context.Context, cwd, value string, linkEnv bool) error {
	number, err := parseNumber(value, issueLanePattern, "GH")
	if err != nil {
		return err
	}
	branch := fmt.Sprintf("GH-%d", number)
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	if a.refExists(ctx, repo.top, "refs/heads/"+branch) {
		return a.addBranch(ctx, repo, branch, linkEnv)
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return err
	}
	if a.refExists(ctx, repo.top, "refs/remotes/origin/"+branch) {
		return a.addBranch(ctx, repo, branch, linkEnv)
	}
	base, err := a.remoteDefaultBranch(ctx, repo.top)
	if err != nil {
		return err
	}
	destination, err := canonicalLanePath(repo, branch)
	if err != nil {
		return err
	}
	if err := a.ensureDestinationAvailable(destination); err != nil {
		return err
	}
	if err := a.mkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create project worktree directory: %w", err)
	}
	if _, err := a.git(ctx, repo.top, "worktree", "add", "-b", branch, destination, "refs/remotes/origin/"+base); err != nil {
		return err
	}
	if err := a.ensureEnvironmentLinks(repo.primary, destination, linkEnv); err != nil {
		return a.rollbackNewWorktreeSetup(ctx, repo.top, destination, err)
	}
	return a.writef("Created %s from origin/%s\n", destination, base)
}

func (a *App) add(ctx context.Context, cwd, branch string, linkEnv bool) error {
	prepared, err := a.PrepareBranch(ctx, cwd, branch, linkEnv)
	if err != nil {
		return err
	}
	return a.writePreparedBranch(prepared)
}

// PrepareBranch creates, attaches, or reuses the exact writable worktree for branch.
func (a *App) PrepareBranch(
	ctx context.Context,
	cwd string,
	branch string,
	linkEnv bool,
) (PreparedWorktree, error) {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return PreparedWorktree{}, err
	}
	if _, err := validateLane(branch); err != nil {
		return PreparedWorktree{}, err
	}
	if _, err := a.git(ctx, repo.top, "check-ref-format", "--branch", branch); err != nil {
		return PreparedWorktree{}, fmt.Errorf("invalid branch %q: %w", branch, err)
	}
	if a.refExists(ctx, repo.top, "refs/heads/"+branch) {
		return a.prepareBranch(ctx, repo, branch, linkEnv)
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return PreparedWorktree{}, err
	}
	return a.prepareBranch(ctx, repo, branch, linkEnv)
}

func (a *App) addBranch(ctx context.Context, repo repository, branch string, linkEnv bool) error {
	prepared, err := a.prepareBranch(ctx, repo, branch, linkEnv)
	if err != nil {
		return err
	}
	return a.writePreparedBranch(prepared)
}

func (a *App) prepareBranch(
	ctx context.Context,
	repo repository,
	branch string,
	linkEnv bool,
) (PreparedWorktree, error) {
	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		return PreparedWorktree{}, err
	}
	for _, entry := range entries {
		if entry.branch == branch {
			if err := a.ensureEnvironmentLinks(repo.primary, entry.path, linkEnv); err != nil {
				return PreparedWorktree{}, err
			}
			return PreparedWorktree{Path: entry.path, Branch: branch}, nil
		}
	}

	local := a.refExists(ctx, repo.top, "refs/heads/"+branch)
	remote := a.refExists(ctx, repo.top, "refs/remotes/origin/"+branch)
	if !local && !remote {
		return PreparedWorktree{}, fmt.Errorf("branch %q does not exist locally or on origin; use `git wt issue <number>` for a new GH lane", branch)
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
	if local {
		if _, err := a.git(ctx, repo.top, "worktree", "add", destination, branch); err != nil {
			return PreparedWorktree{}, err
		}
	} else {
		if _, err := a.git(ctx, repo.top, "worktree", "add", "--track", "-b", branch, destination, "origin/"+branch); err != nil {
			return PreparedWorktree{}, err
		}
	}
	if err := a.ensureEnvironmentLinks(repo.primary, destination, linkEnv); err != nil {
		return PreparedWorktree{}, a.rollbackNewWorktreeSetup(ctx, repo.top, destination, err)
	}
	return PreparedWorktree{Path: destination, Branch: branch, Created: true}, nil
}

func (a *App) writePreparedBranch(prepared PreparedWorktree) error {
	if prepared.Created {
		return a.writef("Created %s for %s\n", prepared.Path, prepared.Branch)
	}
	return a.writef("Reusing %s for %s\n", prepared.Path, prepared.Branch)
}

func (a *App) pr(ctx context.Context, cwd, value string) error {
	number, err := parseNumber(value, prLanePattern, "PR")
	if err != nil {
		return err
	}
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	lane := fmt.Sprintf("PR-%d", number)
	destination, err := canonicalLanePath(repo, lane)
	if err != nil {
		return err
	}
	ref := fmt.Sprintf("refs/git-wt/pr/%d", number)
	refspec := fmt.Sprintf("+refs/pull/%d/head:%s", number, ref)
	if _, err := a.git(ctx, repo.top, "fetch", "--force", "--no-tags", "origin", refspec); err != nil {
		return fmt.Errorf("fetch pull request %d from origin: %w", number, err)
	}

	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if samePath(entry.path, destination) {
			if entry.branch != "" {
				return fmt.Errorf("%s is registered on branch %s; PR lanes must be detached", destination, entry.branch)
			}
			dirty, err := a.status(ctx, destination, false)
			if err != nil {
				return err
			}
			if dirty != "" {
				return fmt.Errorf("%s has local changes; refusing to refresh detached PR view", destination)
			}
			if _, err := a.git(ctx, destination, "checkout", "--detach", ref); err != nil {
				return err
			}
			return a.writef("Refreshed detached inspection lane %s\n", destination)
		}
	}
	if err := a.ensureDestinationAvailable(destination); err != nil {
		return err
	}
	if err := a.mkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create project worktree directory: %w", err)
	}
	if _, err := a.git(ctx, repo.top, "worktree", "add", "--detach", destination, ref); err != nil {
		return err
	}
	if err := a.writef("Created detached inspection lane %s\n", destination); err != nil {
		return err
	}
	return a.writef("Use `git wt repair %d` for writable PR work.\n", number)
}

func (a *App) repair(ctx context.Context, cwd, value string, linkEnv bool) error {
	number, err := parseNumber(value, prLanePattern, "PR")
	if err != nil {
		return err
	}
	repair, err := a.PreparePullRequestRepair(ctx, cwd, number, linkEnv)
	if err != nil {
		return err
	}
	if err := a.writef("PR %d uses writable head branch %s\n", number, repair.Branch); err != nil {
		return err
	}
	return a.writePreparedBranch(repair.PreparedWorktree)
}

// PreparePullRequestRepair prepares the exact writable same-repository PR head.
func (a *App) PreparePullRequestRepair(
	ctx context.Context,
	cwd string,
	number int,
	linkEnv bool,
) (PullRequestRepair, error) {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return PullRequestRepair{}, err
	}
	repository := repo.owner + "/" + repo.name
	pr, err := a.resolvePR(ctx, repo.top, repository, number)
	if err != nil {
		return PullRequestRepair{}, err
	}
	if pr.IsCrossRepository {
		return PullRequestRepair{}, fmt.Errorf("PR %d is from a fork; automatic repair supports same-repository head branches only", number)
	}
	if !strings.EqualFold(pr.State, "OPEN") {
		return PullRequestRepair{}, fmt.Errorf("PR %d is %s, not open", number, strings.ToLower(pr.State))
	}
	if pr.HeadRefName == "" {
		return PullRequestRepair{}, fmt.Errorf("PR %d has no head branch", number)
	}
	if strings.HasPrefix(strings.ToUpper(pr.HeadRefName), "PR-") {
		return PullRequestRepair{}, fmt.Errorf("PR %d head %q is not a durable branch", number, pr.HeadRefName)
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return PullRequestRepair{}, err
	}
	prepared, err := a.prepareBranch(ctx, repo, pr.HeadRefName, linkEnv)
	if err != nil {
		return PullRequestRepair{}, err
	}
	return PullRequestRepair{
		PreparedWorktree: prepared,
		Repository:       repository,
		Number:           number,
		URL:              pr.URL,
		HeadRefOID:       pr.HeadRefOID,
	}, nil
}

func (a *App) resolvePullRequest(ctx context.Context, cwd, slug string, number int) (PR, error) {
	output, err := a.command(ctx, cwd, "gh", "pr", "view", strconv.Itoa(number), "--repo", slug, "--json", "headRefName,headRefOid,isCrossRepository,state,url")
	if err != nil {
		return PR{}, err
	}
	var pr PR
	if err := json.Unmarshal(output, &pr); err != nil {
		return PR{}, fmt.Errorf("decode gh PR response: %w", err)
	}
	return pr, nil
}
