package worktree

import (
	"context"
	"fmt"
	"strings"
)

func (a *App) runSync(ctx context.Context, cwd string, options SyncOptions) SyncReport {
	report := SyncReport{
		SchemaVersion: syncReportSchemaVersion,
		DryRun:        options.DryRun,
		Fetch:         SyncOperation{Status: "pending"},
		Prune:         SyncOperation{Status: "pending"},
		Lanes:         make([]SyncLaneDecision, 0),
		Worktrees:     make([]SyncWorktree, 0),
		Failures:      make([]SyncFailure, 0),
	}

	repo, err := a.repository(ctx, cwd)
	if err != nil {
		report.Fetch = SyncOperation{Status: "skipped", Detail: "repository discovery failed"}
		report.Prune = SyncOperation{Status: "skipped", Detail: "repository discovery failed"}
		report.DefaultBranch = SyncDefaultDecision{
			State:  "unavailable",
			Action: "failed",
			Detail: err.Error(),
		}
		report.addFailure("repository", "", err)
		return report
	}
	report.Repository = repo.owner + "/" + repo.name

	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		report.Fetch = SyncOperation{Status: "skipped", Detail: "worktree discovery failed"}
		report.Prune = SyncOperation{Status: "skipped", Detail: "worktree discovery failed"}
		report.DefaultBranch = SyncDefaultDecision{
			State:  "unavailable",
			Action: "failed",
			Detail: err.Error(),
		}
		report.addFailure("worktree-list", "", err)
		return report
	}

	defaultBranch, remoteOID, err := a.liveOriginDefault(ctx, repo.top)
	if err != nil {
		report.Fetch = SyncOperation{Status: "skipped", Detail: "origin default discovery failed"}
		report.Prune = SyncOperation{Status: "skipped", Detail: "origin default discovery failed"}
		report.DefaultBranch = SyncDefaultDecision{
			State:  "unavailable",
			Action: "failed",
			Detail: err.Error(),
		}
		report.addFailure("origin-default", "", err)
		report.preserveAllLanes(repo, entries, "sync-precondition-failed", err.Error())
		a.refreshSyncWorktrees(ctx, repo, &report)
		return report
	}
	report.DefaultBranch.Branch = defaultBranch
	report.DefaultBranch.RemoteOID = remoteOID

	if !a.fetchForSync(ctx, repo, defaultBranch, remoteOID, options, &report) {
		report.DefaultBranch.State = "unavailable"
		report.DefaultBranch.Action = "skipped"
		report.DefaultBranch.Detail = "origin fetch failed; no reconciliation attempted"
		report.preserveAllLanes(
			repo,
			entries,
			"sync-precondition-failed",
			"origin fetch failed; no lane mutation attempted",
		)
		report.Prune = SyncOperation{
			Status: "skipped",
			Detail: "origin fetch failed; fail-closed before metadata mutation",
		}
		a.refreshSyncWorktrees(ctx, repo, &report)
		return report
	}

	report.DefaultBranch = a.reconcileDefaultBranch(
		ctx,
		repo,
		entries,
		defaultBranch,
		remoteOID,
		options.DryRun,
		&report,
	)
	report.Lanes = a.syncLanes(
		ctx,
		repo,
		entries,
		defaultBranch,
		options.DryRun,
		&report,
	)
	a.pruneForSync(ctx, repo, options.DryRun, &report)
	a.refreshSyncWorktrees(ctx, repo, &report)
	return report
}

func (a *App) fetchForSync(
	ctx context.Context,
	repo repository,
	defaultBranch string,
	remoteOID string,
	options SyncOptions,
	report *SyncReport,
) bool {
	if options.DryRun {
		report.Fetch = SyncOperation{
			Status: "skipped",
			Detail: "strict dry run; live origin state was read without fetching",
		}
		return true
	}
	if _, err := a.git(ctx, repo.top, "fetch", "--prune", "--no-tags", "origin"); err != nil {
		report.Fetch = SyncOperation{Status: "failed", Detail: err.Error()}
		report.addFailure("fetch-origin", "", err)
		return false
	}
	trackingOID, err := a.gitText(
		ctx,
		repo.top,
		"rev-parse",
		"--verify",
		"refs/remotes/origin/"+defaultBranch,
	)
	if err == nil && trackingOID != remoteOID {
		err = fmt.Errorf(
			"origin/%s is %s after fetch, expected live origin OID %s",
			defaultBranch,
			trackingOID,
			remoteOID,
		)
	}
	if err != nil {
		report.Fetch = SyncOperation{Status: "failed", Detail: err.Error()}
		report.addFailure("verify-fetch", "", err)
		return false
	}
	report.Fetch = SyncOperation{
		Status: "applied",
		Detail: "fetched and pruned origin only",
	}
	return true
}

func (a *App) pruneForSync(
	ctx context.Context,
	repo repository,
	dryRun bool,
	report *SyncReport,
) {
	args := []string{"worktree", "prune", "--verbose"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	output, err := a.git(ctx, repo.top, args...)
	if err != nil {
		report.Prune = SyncOperation{Status: "failed", Detail: err.Error()}
		report.addFailure("worktree-prune", "", err)
		return
	}
	detail := strings.TrimSpace(string(output))
	if dryRun {
		report.Prune = SyncOperation{
			Status: "previewed",
			Detail: detailOr(detail, "no stale worktree metadata"),
		}
		return
	}
	report.Prune = SyncOperation{
		Status: "applied",
		Detail: detailOr(detail, "no stale worktree metadata"),
	}
}

func (report *SyncReport) preserveAllLanes(
	repo repository,
	entries []worktreeEntry,
	reason string,
	detail string,
) {
	for _, entry := range entries {
		decision, candidate := baseSyncLaneDecision(repo, entry, report.DefaultBranch.Branch)
		if candidate {
			decision.Reason = reason
			decision.Detail = detail
		}
		report.Lanes = append(report.Lanes, decision)
	}
	sortSyncLanes(report.Lanes)
}

func (a *App) refreshSyncWorktrees(
	ctx context.Context,
	repo repository,
	report *SyncReport,
) {
	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		report.addFailure("refresh-worktree-list", "", err)
		return
	}
	a.populateListMetadata(ctx, entries)
	sortListEntries(entries, listOptions{sortBy: "updated"})
	report.Worktrees = make([]SyncWorktree, 0, len(entries))
	for _, entry := range entries {
		report.Worktrees = append(report.Worktrees, SyncWorktree{
			State:       entry.state,
			Head:        displayHead(entry),
			LastUpdated: entry.updatedText,
			Path:        entry.path,
		})
	}
}
