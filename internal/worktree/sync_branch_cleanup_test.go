package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncRemovesSquashMergedLaneAndExactLocalBranch(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createSquashMergedLane(t, fixture, "topic/squash-merged")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/squash-merged": {
			mergedSyncPR(71, "topic/squash-merged", "main", headOID),
		},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	decision := findSyncLane(t, report, "topic/squash-merged")
	if decision.Action != "removed" ||
		decision.Reason != "proven-safe-merged-lane" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("removed path still exists or stat failed: %v", err)
	}
	if got := gitText(t, fixture.primary, "show-ref", "--verify", "refs/heads/topic/squash-merged"); got != "" {
		t.Fatalf("local branch still exists: %s", got)
	}
	if got := gitText(t, fixture.primary, "show-ref", "--verify", "refs/remotes/kit-wt-sync-proof/topic/squash-merged"); got != "" {
		t.Fatalf("branch-deletion proof still exists: %s", got)
	}
}

func TestSyncPreservesLocalBranchMovedBeforeExactDeletion(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createSquashMergedLane(t, fixture, "topic/moved-before-delete")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/moved-before-delete": {
			mergedSyncPR(72, "topic/moved-before-delete", "main", headOID),
		},
	})
	run := fixture.app.run
	var movedOID string
	fixture.app.run = func(
		ctx context.Context,
		cwd string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name != "git" ||
			!strings.Contains(
				strings.Join(args, "\x00"),
				"branch\x00-d\x00--\x00topic/moved-before-delete",
			) {
			return run(ctx, cwd, name, args...)
		}
		commit, err := run(
			ctx,
			cwd,
			"git",
			"commit-tree",
			headOID+"^{tree}",
			"-p",
			headOID,
			"-m",
			"concurrent branch movement",
		)
		if err != nil {
			return nil, fmt.Errorf("create moved branch commit: %w", err)
		}
		movedOID = strings.TrimSpace(string(commit))
		if _, err := run(
			ctx,
			cwd,
			"git",
			"update-ref",
			"refs/heads/topic/moved-before-delete",
			movedOID,
			headOID,
		); err != nil {
			return nil, fmt.Errorf("move branch before deletion: %w", err)
		}
		return run(ctx, cwd, name, args...)
	}

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatalf(
			"sync expected exact branch deletion failure; moved OID %q; branch %q; report %#v",
			movedOID,
			gitText(t, fixture.primary, "rev-parse", "refs/heads/topic/moved-before-delete"),
			report,
		)
	}
	decision := findSyncLane(t, report, "topic/moved-before-delete")
	if decision.Action != "worktree-removed" ||
		decision.Reason != "branch-deletion-failed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists or stat failed: %v", err)
	}
	if got := gitText(t, fixture.primary, "rev-parse", "refs/heads/topic/moved-before-delete"); got != movedOID {
		t.Fatalf("moved local branch = %q, want %q; detail %q", got, movedOID, decision.Detail)
	}
	if got := gitText(t, fixture.primary, "show-ref", "--verify", "refs/remotes/kit-wt-sync-proof/topic/moved-before-delete"); got != "" {
		t.Fatalf("branch-deletion proof still exists: %s", got)
	}
}

func TestSyncPreservesLaneWhenBranchDeletionProofCannotBePrepared(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createSquashMergedLane(t, fixture, "topic/missing-merge-config")
	runGit(
		t,
		fixture.primary,
		"config",
		"--unset-all",
		"branch.topic/missing-merge-config.merge",
	)
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/missing-merge-config": {
			mergedSyncPR(73, "topic/missing-merge-config", "main", headOID),
		},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected branch-deletion preparation failure")
	}
	decision := findSyncLane(t, report, "topic/missing-merge-config")
	if decision.Action != "preserved" ||
		decision.Reason != "branch-deletion-preparation-failed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("worktree was not preserved: %v", err)
	}
	if got := gitText(t, fixture.primary, "rev-parse", "refs/heads/topic/missing-merge-config"); got != headOID {
		t.Fatalf("local branch = %q, want %q", got, headOID)
	}
}

func TestSyncPreservesLaneOnBranchDeletionProofCollision(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createSquashMergedLane(t, fixture, "topic/proof-collision")
	proofRef := "refs/remotes/kit-wt-sync-proof/topic/proof-collision"
	runGit(t, fixture.primary, "update-ref", proofRef, headOID)
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/proof-collision": {
			mergedSyncPR(75, "topic/proof-collision", "main", headOID),
		},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected branch-deletion proof collision")
	}
	decision := findSyncLane(t, report, "topic/proof-collision")
	if decision.Action != "preserved" ||
		decision.Reason != "branch-deletion-preparation-failed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("worktree was not preserved: %v", err)
	}
	if got := gitText(t, fixture.primary, "rev-parse", proofRef); got != headOID {
		t.Fatalf("colliding proof ref = %q, want %q", got, headOID)
	}
}

func TestSyncPreservesBranchReattachedBeforeDeletion(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createSquashMergedLane(t, fixture, "topic/reattached")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/reattached": {
			mergedSyncPR(74, "topic/reattached", "main", headOID),
		},
	})
	run := fixture.app.run
	reattachedPath := filepath.Join(t.TempDir(), "reattached")
	fixture.app.run = func(
		ctx context.Context,
		cwd string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name != "git" ||
			!strings.Contains(
				strings.Join(args, "\x00"),
				"branch\x00-d\x00--\x00topic/reattached",
			) {
			return run(ctx, cwd, name, args...)
		}
		if _, err := run(
			ctx,
			cwd,
			"git",
			"worktree",
			"add",
			reattachedPath,
			"topic/reattached",
		); err != nil {
			return nil, fmt.Errorf("reattach branch before deletion: %w", err)
		}
		return run(ctx, cwd, name, args...)
	}

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected checked-out branch deletion failure")
	}
	decision := findSyncLane(t, report, "topic/reattached")
	if decision.Action != "worktree-removed" ||
		decision.Reason != "branch-deletion-failed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("original worktree still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(reattachedPath); err != nil {
		t.Fatalf("reattached worktree was not preserved: %v", err)
	}
	if got := gitText(t, fixture.primary, "rev-parse", "refs/heads/topic/reattached"); got != headOID {
		t.Fatalf("reattached branch = %q, want %q", got, headOID)
	}
	if got := gitText(t, fixture.primary, "show-ref", "--verify", "refs/remotes/kit-wt-sync-proof/topic/reattached"); got != "" {
		t.Fatalf("branch-deletion proof still exists: %s", got)
	}
}
