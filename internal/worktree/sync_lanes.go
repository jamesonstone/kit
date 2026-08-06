package worktree

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func (a *App) syncLanes(
	ctx context.Context,
	repo repository,
	entries []worktreeEntry,
	defaultBranch string,
	dryRun bool,
	report *SyncReport,
) []SyncLaneDecision {
	decisions := make([]SyncLaneDecision, len(entries))
	candidateIndexes := make([]int, 0)
	branches := make([]string, 0)
	for i, entry := range entries {
		decision, candidate := baseSyncLaneDecision(repo, entry, defaultBranch)
		decisions[i] = decision
		if candidate {
			candidateIndexes = append(candidateIndexes, i)
			branches = append(branches, entry.branch)
		}
	}
	sort.Strings(branches)
	branches = uniqueStrings(branches)

	prsByBranch := make(map[string][]SyncPullRequest)
	if len(branches) != 0 {
		var err error
		prsByBranch, err = a.resolveSyncPRs(
			ctx,
			repo.top,
			repo.owner+"/"+repo.name,
			branches,
		)
		if err != nil {
			report.addFailure("resolve-pull-requests", "", err)
			for _, index := range candidateIndexes {
				decisions[index].Action = "preserved"
				decisions[index].Reason = "github-unavailable"
				decisions[index].Detail = err.Error()
			}
			sortSyncLanes(decisions)
			return decisions
		}
	}

	for _, index := range candidateIndexes {
		a.syncLane(
			ctx,
			repo,
			entries[index],
			defaultBranch,
			dryRun,
			prsByBranch[entries[index].branch],
			&decisions[index],
			report,
		)
	}
	sortSyncLanes(decisions)
	return decisions
}

func (a *App) syncLane(
	ctx context.Context,
	repo repository,
	entry worktreeEntry,
	defaultBranch string,
	dryRun bool,
	prs []SyncPullRequest,
	decision *SyncLaneDecision,
	report *SyncReport,
) {
	if len(prs) == 0 {
		decision.Reason = "pull-request-missing"
		return
	}
	if len(prs) != 1 {
		decision.Reason = "pull-request-ambiguous"
		decision.Detail = fmt.Sprintf("%d pull requests use this head branch", len(prs))
		return
	}
	pr := prs[0]
	decision.PullRequest = &pr
	if refusal := pullRequestRefusal(pr, entry, defaultBranch); refusal != "" {
		decision.Reason = refusal
		if refusal == "head-oid-mismatch" {
			decision.Detail = fmt.Sprintf(
				"local %s; merged PR head %s",
				entry.head,
				detailOr(pr.HeadRefOID, "missing"),
			)
		}
		return
	}

	removal, err := a.inspectWorktreeRemoval(ctx, repo, entry, discardIgnoredBuildOutput)
	if err != nil {
		decision.Reason = removalRefusalReason(err)
		decision.Detail = err.Error()
		if decision.Reason == "inspection-failed" {
			report.addFailure("inspect-worktree", entry.path, err)
		}
		return
	}
	if dryRun {
		decision.Action = "would-remove"
		decision.Reason = "proven-safe-merged-lane"
		return
	}
	proofRef, err := a.prepareMergedLocalBranchDeletion(
		ctx,
		repo,
		entry.branch,
		entry.head,
	)
	if err != nil {
		decision.Action = "preserved"
		decision.Reason = "branch-deletion-preparation-failed"
		decision.Detail = err.Error()
		report.addFailure("prepare-local-branch-deletion", entry.path, err)
		return
	}
	if err := a.executeWorktreeRemoval(ctx, repo, removal); err != nil {
		decision.Action = "preserved"
		decision.Reason = "worktree-removal-failed"
		decision.Detail = err.Error()
		report.addFailure("remove-worktree", entry.path, err)
		if cleanupErr := a.cleanupMergedLocalBranchDeletion(
			context.WithoutCancel(ctx),
			repo,
			proofRef,
			entry.head,
		); cleanupErr != nil {
			report.addFailure("remove-branch-deletion-proof", entry.path, cleanupErr)
		}
		return
	}
	deleteErr := a.deleteMergedLocalBranch(
		ctx,
		repo,
		entry.branch,
		entry.head,
		proofRef,
	)
	cleanupErr := a.cleanupMergedLocalBranchDeletion(
		context.WithoutCancel(ctx),
		repo,
		proofRef,
		entry.head,
	)
	if deleteErr != nil || cleanupErr != nil {
		err := errors.Join(deleteErr, cleanupErr)
		decision.Action = "worktree-removed"
		decision.Reason = "branch-deletion-failed"
		decision.Detail = err.Error()
		report.addFailure("delete-local-branch", entry.path, err)
		return
	}
	decision.Action = "removed"
	decision.Reason = "proven-safe-merged-lane"
}

func pullRequestRefusal(
	pr SyncPullRequest,
	entry worktreeEntry,
	defaultBranch string,
) string {
	switch {
	case pr.HeadRefName != entry.branch:
		return "pull-request-head-mismatch"
	case pr.IsCrossRepository:
		return "pull-request-from-fork"
	case !strings.EqualFold(pr.State, "MERGED"):
		if strings.EqualFold(pr.State, "OPEN") {
			return "pull-request-open"
		}
		return "pull-request-not-merged"
	case pr.MergedAt == nil:
		return "pull-request-missing-merge-time"
	case pr.BaseRefName != defaultBranch:
		return "pull-request-wrong-base"
	case pr.HeadRefOID == "" || pr.HeadRefOID != entry.head:
		return "head-oid-mismatch"
	default:
		return ""
	}
}
