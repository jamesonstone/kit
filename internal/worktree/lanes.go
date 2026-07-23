package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func (a *App) issue(ctx context.Context, cwd, value string) error {
	number, err := parseNumber(value, issueLanePattern, "GH")
	if err != nil {
		return err
	}
	branch := fmt.Sprintf("GH-%d", number)
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return err
	}
	if a.refExists(ctx, repo.top, "refs/heads/"+branch) || a.refExists(ctx, repo.top, "refs/remotes/origin/"+branch) {
		return a.addBranch(ctx, repo, branch)
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
	return a.writef("Created %s from origin/%s\n", destination, base)
}

func (a *App) add(ctx context.Context, cwd, branch string) error {
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	if _, err := validateLane(branch); err != nil {
		return err
	}
	if _, err := a.git(ctx, repo.top, "check-ref-format", "--branch", branch); err != nil {
		return fmt.Errorf("invalid branch %q: %w", branch, err)
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return err
	}
	return a.addBranch(ctx, repo, branch)
}

func (a *App) addBranch(ctx context.Context, repo repository, branch string) error {
	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.branch == branch {
			return a.writef("Reusing %s for %s\n", entry.path, branch)
		}
	}

	local := a.refExists(ctx, repo.top, "refs/heads/"+branch)
	remote := a.refExists(ctx, repo.top, "refs/remotes/origin/"+branch)
	if !local && !remote {
		return fmt.Errorf("branch %q does not exist locally or on origin; use `git wt issue <number>` for a new GH lane", branch)
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
	if local {
		if _, err := a.git(ctx, repo.top, "worktree", "add", destination, branch); err != nil {
			return err
		}
	} else {
		if _, err := a.git(ctx, repo.top, "worktree", "add", "--track", "-b", branch, destination, "origin/"+branch); err != nil {
			return err
		}
	}
	return a.writef("Created %s for %s\n", destination, branch)
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

func (a *App) repair(ctx context.Context, cwd, value string) error {
	number, err := parseNumber(value, prLanePattern, "PR")
	if err != nil {
		return err
	}
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	pr, err := a.resolvePR(ctx, repo.top, repo.owner+"/"+repo.name, number)
	if err != nil {
		return err
	}
	if pr.IsCrossRepository {
		return fmt.Errorf("PR %d is from a fork; automatic repair supports same-repository head branches only", number)
	}
	if !strings.EqualFold(pr.State, "OPEN") {
		return fmt.Errorf("PR %d is %s, not open", number, strings.ToLower(pr.State))
	}
	if pr.HeadRefName == "" {
		return fmt.Errorf("PR %d has no head branch", number)
	}
	if strings.HasPrefix(strings.ToUpper(pr.HeadRefName), "PR-") {
		return fmt.Errorf("PR %d head %q is not a durable branch", number, pr.HeadRefName)
	}
	if err := a.fetchOrigin(ctx, repo.top); err != nil {
		return err
	}
	if err := a.writef("PR %d uses writable head branch %s\n", number, pr.HeadRefName); err != nil {
		return err
	}
	return a.addBranch(ctx, repo, pr.HeadRefName)
}

func (a *App) resolvePullRequest(ctx context.Context, cwd, slug string, number int) (PR, error) {
	output, err := a.command(ctx, cwd, "gh", "pr", "view", strconv.Itoa(number), "--repo", slug, "--json", "headRefName,isCrossRepository,state,url")
	if err != nil {
		return PR{}, err
	}
	var pr PR
	if err := json.Unmarshal(output, &pr); err != nil {
		return PR{}, fmt.Errorf("decode gh PR response: %w", err)
	}
	return pr, nil
}
