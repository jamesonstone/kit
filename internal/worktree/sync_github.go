package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

func (a *App) resolveSyncPullRequests(
	ctx context.Context,
	cwd string,
	repository string,
	branches []string,
) (map[string][]SyncPullRequest, error) {
	result := make(map[string][]SyncPullRequest, len(branches))
	batch, err := a.querySyncPullRequests(ctx, cwd, repository, "")
	if err != nil {
		return nil, fmt.Errorf("batch pull-request lookup: %w", err)
	}
	wanted := make(map[string]bool, len(branches))
	for _, branch := range branches {
		wanted[branch] = true
	}
	for _, pr := range batch {
		if wanted[pr.HeadRefName] {
			result[pr.HeadRefName] = append(result[pr.HeadRefName], pr)
		}
	}
	for _, branch := range branches {
		targeted, targetErr := a.querySyncPullRequests(ctx, cwd, repository, branch)
		if targetErr != nil {
			return nil, fmt.Errorf("pull-request lookup for %s: %w", branch, targetErr)
		}
		confirmed := make([]SyncPullRequest, 0, len(targeted))
		for _, pr := range targeted {
			if pr.HeadRefName == branch {
				confirmed = append(confirmed, pr)
			}
		}
		result[branch] = confirmed
	}
	for branch := range result {
		sort.Slice(result[branch], func(i, j int) bool {
			return result[branch][i].Number < result[branch][j].Number
		})
	}
	return result, nil
}

func (a *App) querySyncPullRequests(
	ctx context.Context,
	cwd string,
	repository string,
	head string,
) ([]SyncPullRequest, error) {
	args := []string{
		"pr",
		"list",
		"--repo",
		repository,
		"--state",
		"all",
		"--limit",
		"100",
		"--json",
		"number,state,mergedAt,baseRefName,headRefName,headRefOid,isCrossRepository,url",
	}
	if head != "" {
		args = append(args, "--head", head)
	}
	output, err := a.command(ctx, cwd, "gh", args...)
	if err != nil {
		return nil, err
	}
	var prs []SyncPullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("decode GitHub pull requests: %w", err)
	}
	return prs, nil
}
