package worktree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSyncRemovesExactMergedLaneAndLocalBranch(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createMergedLane(t, fixture, "topic/merged")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/merged": {mergedSyncPR(17, "topic/merged", "main", headOID)},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	decision := findSyncLane(t, report, "topic/merged")
	if decision.Action != "removed" ||
		decision.Reason != "proven-safe-merged-lane" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("removed path still exists or stat failed: %v", err)
	}
	if got := gitText(t, fixture.primary, "show-ref", "--verify", "refs/heads/topic/merged"); got != "" {
		t.Fatalf("local branch still exists: %s", got)
	}
}

func TestSyncDeletesMergedBranchWhenDefaultIsNotCheckedOut(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createMergedLane(t, fixture, "topic/delete-off-main")
	runGit(t, fixture.primary, "switch", "-c", "primary-topic")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/delete-off-main": {
			mergedSyncPR(18, "topic/delete-off-main", "main", headOID),
		},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	if decision := findSyncLane(t, report, "topic/delete-off-main"); decision.Action != "removed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("removed path still exists or stat failed: %v", err)
	}
	if got := gitText(t, fixture.primary, "branch", "--show-current"); got != "primary-topic" {
		t.Fatalf("sync changed the current branch to %q", got)
	}
}

func TestSyncPreservesUnsafePullRequestEvidence(t *testing.T) {
	fixture := newGitFixture(t)
	branches := []string{
		"topic/open",
		"topic/closed",
		"topic/wrong-base",
		"topic/fork",
		"topic/missing",
		"topic/ambiguous",
		"topic/oid",
		"topic/missing-merge-time",
	}
	oids := make(map[string]string)
	for _, branch := range branches {
		runGit(t, fixture.primary, "branch", branch, "origin/main")
		runWT(t, fixture.app, fixture.primary, "add", branch, "--no-link-env")
		path := canonicalTestLanePath(fixture, branch)
		oids[branch] = gitText(t, path, "rev-parse", "HEAD")
	}
	mergedAt := time.Now().UTC()
	prs := map[string][]SyncPullRequest{
		"topic/open": {
			{
				Number:      1,
				State:       "OPEN",
				BaseRefName: "main",
				HeadRefName: "topic/open",
				HeadRefOID:  oids["topic/open"],
			},
		},
		"topic/closed": {
			{
				Number:      2,
				State:       "CLOSED",
				BaseRefName: "main",
				HeadRefName: "topic/closed",
				HeadRefOID:  oids["topic/closed"],
			},
		},
		"topic/wrong-base": {
			{
				Number:      3,
				State:       "MERGED",
				MergedAt:    &mergedAt,
				BaseRefName: "release",
				HeadRefName: "topic/wrong-base",
				HeadRefOID:  oids["topic/wrong-base"],
			},
		},
		"topic/fork": {
			{
				Number:            4,
				State:             "MERGED",
				MergedAt:          &mergedAt,
				BaseRefName:       "main",
				HeadRefName:       "topic/fork",
				HeadRefOID:        oids["topic/fork"],
				IsCrossRepository: true,
			},
		},
		"topic/ambiguous": {
			mergedSyncPR(5, "topic/ambiguous", "main", oids["topic/ambiguous"]),
			mergedSyncPR(6, "topic/ambiguous", "main", oids["topic/ambiguous"]),
		},
		"topic/oid": {
			mergedSyncPR(7, "topic/oid", "main", strings.Repeat("a", 40)),
		},
		"topic/missing-merge-time": {
			{
				Number:      8,
				State:       "MERGED",
				BaseRefName: "main",
				HeadRefName: "topic/missing-merge-time",
				HeadRefOID:  oids["topic/missing-merge-time"],
			},
		},
	}
	fixture.app.resolveSyncPRs = staticSyncPRs(prs)

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	wantReasons := map[string]string{
		"topic/open":               "pull-request-open",
		"topic/closed":             "pull-request-not-merged",
		"topic/wrong-base":         "pull-request-wrong-base",
		"topic/fork":               "pull-request-from-fork",
		"topic/missing":            "pull-request-missing",
		"topic/ambiguous":          "pull-request-ambiguous",
		"topic/oid":                "head-oid-mismatch",
		"topic/missing-merge-time": "pull-request-missing-merge-time",
	}
	for branch, wantReason := range wantReasons {
		decision := findSyncLane(t, report, branch)
		if decision.Action != "preserved" || decision.Reason != wantReason {
			t.Errorf("%s decision = %#v, want reason %s", branch, decision, wantReason)
		}
		if _, err := os.Stat(canonicalTestLanePath(fixture, branch)); err != nil {
			t.Errorf("%s path was not preserved: %v", branch, err)
		}
	}
}

func TestSyncPreservesDirtyIgnoredCurrentPrimaryDetachedAndNonCanonical(t *testing.T) {
	fixture := newGitFixture(t)
	prs := make(map[string][]SyncPullRequest)

	for _, branch := range []string{"topic/dirty", "topic/ignored"} {
		runGit(t, fixture.primary, "branch", branch, "origin/main")
		runWT(t, fixture.app, fixture.primary, "add", branch, "--no-link-env")
	}
	if err := os.WriteFile(
		filepath.Join(canonicalTestLanePath(fixture, "topic/dirty"), "dirty.txt"),
		[]byte("preserve\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	ignoredPath := canonicalTestLanePath(fixture, "topic/ignored")
	if err := os.WriteFile(filepath.Join(ignoredPath, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, ignoredPath, "add", ".gitignore")
	runGit(t, ignoredPath, "commit", "-m", "ignore test material")
	if err := os.WriteFile(
		filepath.Join(ignoredPath, "ignored.txt"),
		[]byte("preserve\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	for _, branch := range []string{"topic/dirty", "topic/ignored"} {
		path := canonicalTestLanePath(fixture, branch)
		prs[branch] = []SyncPullRequest{
			mergedSyncPR(20, branch, "main", gitText(t, path, "rev-parse", "HEAD")),
		}
	}

	detached := filepath.Join(fixture.worktreeRoot, "example", "project", "PR-88")
	runGit(t, fixture.primary, "worktree", "add", "--detach", detached, "origin/main")
	legacy := filepath.Join(fixture.worktreeRoot, "legacy-topic")
	runGit(t, fixture.primary, "branch", "topic/legacy", "origin/main")
	runGit(t, fixture.primary, "worktree", "add", legacy, "topic/legacy")
	prs["topic/legacy"] = []SyncPullRequest{
		mergedSyncPR(
			21,
			"topic/legacy",
			"main",
			gitText(t, legacy, "rev-parse", "HEAD"),
		),
	}
	fixture.app.resolveSyncPRs = staticSyncPRs(prs)

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	want := map[string]string{
		"topic/dirty":   "worktree-dirty",
		"topic/ignored": "worktree-dirty",
		"topic/legacy":  "non-canonical-worktree",
	}
	for branch, reason := range want {
		if decision := findSyncLane(t, report, branch); decision.Reason != reason {
			t.Errorf("%s decision = %#v, want %s", branch, decision, reason)
		}
	}
	if reason := findSyncLaneByPath(t, report, fixture.primary).Reason; reason != "primary-and-current-worktree" {
		t.Errorf("primary decision reason = %s", reason)
	}
	if reason := findSyncLaneByPath(t, report, detached).Reason; reason != "detached-worktree" {
		t.Errorf("detached decision reason = %s", reason)
	}
}
