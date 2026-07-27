package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncFromCandidatePreservesCurrentWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "topic/current", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/current", "--no-link-env")
	path := canonicalTestLanePath(fixture, "topic/current")
	oid := gitText(t, path, "rev-parse", "HEAD")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/current": {mergedSyncPR(22, "topic/current", "main", oid)},
	})

	report, err := runSyncJSON(t, fixture, path)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	if decision := findSyncLane(t, report, "topic/current"); decision.Reason != "current-worktree" {
		t.Fatalf("current decision = %#v", decision)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("current worktree was not preserved: %v", err)
	}
}

func TestSyncManagedEnvironmentLinkAndRestoration(t *testing.T) {
	t.Run("removed", func(t *testing.T) {
		fixture := newGitFixture(t)
		source := writeEnvironmentSource(t, fixture, "TOKEN=source\n")
		rcSource := writeEnvironmentRCSource(t, fixture, "dotenv\n")
		path, headOID := createMergedLane(t, fixture, "topic/env")
		fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
			"topic/env": {mergedSyncPR(30, "topic/env", "main", headOID)},
		})

		if _, err := runSyncJSON(t, fixture, fixture.primary); err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("worktree still exists or stat failed: %v", err)
		}
		if data, err := os.ReadFile(source); err != nil || string(data) != "TOKEN=source\n" {
			t.Fatalf("source environment changed: data=%q err=%v", data, err)
		}
		if data, err := os.ReadFile(rcSource); err != nil || string(data) != "dotenv\n" {
			t.Fatalf("source environment configuration changed: data=%q err=%v", data, err)
		}
	})

	t.Run("restored after removal failure", func(t *testing.T) {
		fixture := newGitFixture(t)
		source := writeEnvironmentSource(t, fixture, "TOKEN=restore\n")
		rcSource := writeEnvironmentRCSource(t, fixture, "dotenv\n")
		path, headOID := createMergedLane(t, fixture, "topic/env-failure")
		destination := filepath.Join(path, environmentFileName)
		fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
			"topic/env-failure": {
				mergedSyncPR(31, "topic/env-failure", "main", headOID),
			},
		})
		run := fixture.app.run
		fixture.app.run = func(
			ctx context.Context,
			cwd string,
			name string,
			args ...string,
		) ([]byte, error) {
			if name == "git" &&
				len(args) >= 2 &&
				args[0] == "worktree" &&
				args[1] == "remove" {
				return []byte("simulated removal failure"), fmt.Errorf("simulated failure")
			}
			return run(ctx, cwd, name, args...)
		}

		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err == nil {
			t.Fatal("sync expected aggregate failure")
		}
		decision := findSyncLane(t, report, "topic/env-failure")
		if decision.Reason != "worktree-removal-failed" ||
			!strings.Contains(decision.Detail, "restored environment symlink") {
			t.Fatalf("decision = %#v", decision)
		}
		assertEnvironmentSymlink(t, destination, source)
		assertEnvironmentSymlink(
			t,
			filepath.Join(path, environmentRCFileName),
			rcSource,
		)
	})
}
