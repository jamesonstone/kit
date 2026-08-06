package worktree

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSyncDryRunIsStrictlyMutationFree(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createSquashMergedLane(t, fixture, "topic/dry-run")
	runGit(t, fixture.remote, "update-ref", "-d", "refs/heads/topic/dry-run")
	advanceRemoteMain(t, fixture, "remote-after-merge.txt")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/dry-run": {mergedSyncPR(40, "topic/dry-run", "main", headOID)},
	})

	before := captureSyncState(t, fixture)
	report, err := runSyncJSON(t, fixture, fixture.primary, "--dry-run")
	if err != nil {
		t.Fatalf("dry-run sync error = %v", err)
	}
	after := captureSyncState(t, fixture)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("dry run mutated state:\nbefore=%#v\nafter=%#v", before, after)
	}
	if report.Fetch.Status != "skipped" ||
		report.Prune.Status != "previewed" {
		t.Fatalf("dry-run operations = fetch %#v prune %#v", report.Fetch, report.Prune)
	}
	if decision := findSyncLane(t, report, "topic/dry-run"); decision.Action != "would-remove" {
		t.Fatalf("dry-run lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("dry-run worktree missing: %v", err)
	}
}

func TestSyncDryRunPreviewsStaleMetadataWithoutResolvingPullRequests(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "topic/stale", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/stale", "--no-link-env")
	path := canonicalTestLanePath(fixture, "topic/stale")
	moved := path + ".preserved"
	if err := os.Rename(path, moved); err != nil {
		t.Fatal(err)
	}
	fixture.app.resolveSyncPRs = func(
		context.Context,
		string,
		string,
		[]string,
	) (map[string][]SyncPullRequest, error) {
		return nil, fmt.Errorf("stale metadata must not require GitHub evidence")
	}

	report, err := runSyncJSON(t, fixture, fixture.primary, "--dry-run")
	if err != nil {
		t.Fatalf("dry-run sync error = %v", err)
	}
	decision := findSyncLane(t, report, "topic/stale")
	if decision.Reason != "stale-worktree-metadata" ||
		decision.Action != "preserved" {
		t.Fatalf("stale decision = %#v", decision)
	}
	if report.Prune.Status != "previewed" ||
		!strings.Contains(report.Prune.Detail, "Removing") {
		t.Fatalf("prune preview = %#v", report.Prune)
	}
	if _, err := os.Stat(moved); err != nil {
		t.Fatalf("dry run modified preserved stale contents: %v", err)
	}
}

func TestSyncDeletedUpstreamAfterMergeStillRemovesLane(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createMergedLane(t, fixture, "topic/pruned-upstream")
	runGit(t, fixture.remote, "update-ref", "-d", "refs/heads/topic/pruned-upstream")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/pruned-upstream": {
			mergedSyncPR(41, "topic/pruned-upstream", "main", headOID),
		},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	if decision := findSyncLane(t, report, "topic/pruned-upstream"); decision.Action != "removed" {
		t.Fatalf("decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists or stat failed: %v", err)
	}
}
