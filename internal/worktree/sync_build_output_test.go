package worktree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncRemovesMergedLaneWithIgnoredRootBuildOutput(t *testing.T) {
	fixture := newGitFixture(t)
	configureIgnoredBuildOutput(t, fixture, "bin/\n")
	path, headOID := createMergedLane(t, fixture, "topic/ignored-bin")
	buildArtifact := writeBuildArtifact(t, path)
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/ignored-bin": {mergedSyncPR(81, "topic/ignored-bin", "main", headOID)},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	decision := findSyncLane(t, report, "topic/ignored-bin")
	if decision.Action != "removed" || decision.Reason != "proven-safe-merged-lane" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(buildArtifact); !os.IsNotExist(err) {
		t.Fatalf("ignored build output still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("removed path still exists or stat failed: %v", err)
	}
}

func TestSyncDryRunPreservesIgnoredRootBuildOutput(t *testing.T) {
	fixture := newGitFixture(t)
	configureIgnoredBuildOutput(t, fixture, "bin/\n")
	path, headOID := createMergedLane(t, fixture, "topic/preview-bin")
	buildArtifact := writeBuildArtifact(t, path)
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/preview-bin": {mergedSyncPR(82, "topic/preview-bin", "main", headOID)},
	})

	report, err := runSyncJSON(t, fixture, fixture.primary, "--dry-run")
	if err != nil {
		t.Fatalf("sync dry-run error = %v", err)
	}
	decision := findSyncLane(t, report, "topic/preview-bin")
	if decision.Action != "would-remove" || decision.Reason != "proven-safe-merged-lane" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(buildArtifact); err != nil {
		t.Fatalf("dry-run changed ignored build output: %v", err)
	}
	assertBranch(t, path, "topic/preview-bin")
}

func TestSyncPreservesMaterialOutsideIgnoredRootBuildOutput(t *testing.T) {
	t.Run("ordinary untracked root bin", func(t *testing.T) {
		fixture := newGitFixture(t)
		path, headOID := createMergedLane(t, fixture, "topic/untracked-bin")
		writeBuildArtifact(t, path)
		assertMergedLanePreserved(t, fixture, "topic/untracked-bin", headOID)
	})

	t.Run("nested ignored bin", func(t *testing.T) {
		fixture := newGitFixture(t)
		configureIgnoredBuildOutput(t, fixture, "tools/bin/\n")
		path, headOID := createMergedLane(t, fixture, "topic/nested-bin")
		if err := os.MkdirAll(filepath.Join(path, "tools", "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "tools", "bin", "tool"), []byte("binary\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		assertMergedLanePreserved(t, fixture, "topic/nested-bin", headOID)
	})

	t.Run("other ignored material", func(t *testing.T) {
		fixture := newGitFixture(t)
		configureIgnoredBuildOutput(t, fixture, "bin/\ncache/\n")
		path, headOID := createMergedLane(t, fixture, "topic/other-ignored")
		writeBuildArtifact(t, path)
		if err := os.MkdirAll(filepath.Join(path, "cache"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "cache", "state"), []byte("preserve\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		assertMergedLanePreserved(t, fixture, "topic/other-ignored", headOID)
	})

	t.Run("ignored root bin symlink", func(t *testing.T) {
		fixture := newGitFixture(t)
		configureIgnoredBuildOutput(t, fixture, "bin\n")
		path, headOID := createMergedLane(t, fixture, "topic/symlink-bin")
		target := t.TempDir()
		if err := os.Symlink(target, filepath.Join(path, "bin")); err != nil {
			t.Fatal(err)
		}
		assertMergedLanePreserved(t, fixture, "topic/symlink-bin", headOID)
		if got, err := os.Readlink(filepath.Join(path, "bin")); err != nil || got != target {
			t.Fatalf("ignored bin symlink changed: target=%q err=%v", got, err)
		}
	})

	t.Run("tracked change beneath bin", func(t *testing.T) {
		fixture := newGitFixture(t)
		configureIgnoredBuildOutput(t, fixture, "bin/\n")
		path, _ := createPublishedLane(t, fixture, "topic/tracked-bin")
		tracked := filepath.Join(path, "bin", "tracked-tool")
		if err := os.MkdirAll(filepath.Dir(tracked), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tracked, []byte("tracked\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		runGit(t, path, "add", "--force", "bin/tracked-tool")
		runGit(t, path, "commit", "-m", "track bin tool")
		runGit(t, path, "push", "origin", "topic/tracked-bin")
		headOID := gitText(t, path, "rev-parse", "HEAD")
		runGit(t, fixture.primary, "merge", "--no-ff", "--no-edit", "topic/tracked-bin")
		runGit(t, fixture.primary, "push", "origin", "main")
		if err := os.WriteFile(tracked, []byte("modified\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		assertMergedLanePreserved(t, fixture, "topic/tracked-bin", headOID)
	})

	t.Run("clean tracked file beside nested ignored bin", func(t *testing.T) {
		fixture := newGitFixture(t)
		configureIgnoredBuildOutput(t, fixture, "bin/\n")
		path, _ := createPublishedLane(t, fixture, "topic/tracked-bin-sibling")
		tracked := filepath.Join(path, "bin", "tracked-tool")
		if err := os.MkdirAll(filepath.Dir(tracked), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tracked, []byte("tracked\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		runGit(t, path, "add", "--force", "bin/tracked-tool")
		runGit(t, path, "commit", "-m", "track bin tool")
		runGit(t, path, "push", "origin", "topic/tracked-bin-sibling")
		headOID := gitText(t, path, "rev-parse", "HEAD")
		runGit(t, fixture.primary, "merge", "--no-ff", "--no-edit", "topic/tracked-bin-sibling")
		runGit(t, fixture.primary, "push", "origin", "main")
		ignored := filepath.Join(path, "bin", "generated", "tool")
		if err := os.MkdirAll(filepath.Dir(ignored), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ignored, []byte("ignored\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		assertMergedLanePreserved(t, fixture, "topic/tracked-bin-sibling", headOID)
		if got, err := os.ReadFile(tracked); err != nil || string(got) != "tracked\n" {
			t.Fatalf("tracked bin sibling changed: content=%q err=%v", got, err)
		}
	})
}

func TestManualRemoveStillRefusesIgnoredRootBuildOutput(t *testing.T) {
	fixture := newGitFixture(t)
	configureIgnoredBuildOutput(t, fixture, "bin/\n")
	runGit(t, fixture.primary, "branch", "--track", "topic/manual-bin", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/manual-bin")
	path := canonicalTestLanePath(fixture, "topic/manual-bin")
	buildArtifact := writeBuildArtifact(t, path)

	err := fixture.app.Run(context.Background(), fixture.primary, []string{"remove", "topic/manual-bin"})
	if err == nil || !strings.Contains(err.Error(), "!! bin/") {
		t.Fatalf("manual removal error = %v", err)
	}
	if _, err := os.Stat(buildArtifact); err != nil {
		t.Fatalf("manual removal changed ignored build output: %v", err)
	}
	assertBranch(t, path, "topic/manual-bin")
}

func TestSyncRestoresEnvironmentLinksWhenBuildOutputRemovalFails(t *testing.T) {
	fixture := newGitFixture(t)
	source := writeEnvironmentSource(t, fixture, "TOKEN=preserve\n")
	configureIgnoredBuildOutput(t, fixture, ".env\nbin/\n")
	path, headOID := createMergedLane(t, fixture, "topic/bin-remove-failure")
	writeBuildArtifact(t, path)
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/bin-remove-failure": {
			mergedSyncPR(83, "topic/bin-remove-failure", "main", headOID),
		},
	})
	fixture.app.removeAll = func(string) error { return errors.New("simulated removal failure") }

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected build-output removal failure")
	}
	decision := findSyncLane(t, report, "topic/bin-remove-failure")
	if decision.Action != "preserved" || decision.Reason != "worktree-removal-failed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	assertEnvironmentSymlink(t, filepath.Join(path, environmentFileName), source)
	assertBranch(t, path, "topic/bin-remove-failure")
}

func TestSyncDoesNotRestoreDisposableBuildOutputAfterNativeRemovalFailure(t *testing.T) {
	fixture := newGitFixture(t)
	configureIgnoredBuildOutput(t, fixture, "bin/\n")
	path, headOID := createMergedLane(t, fixture, "topic/native-remove-failure")
	buildArtifact := writeBuildArtifact(t, path)
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/native-remove-failure": {
			mergedSyncPR(85, "topic/native-remove-failure", "main", headOID),
		},
	})
	run := fixture.app.run
	fixture.app.run = func(
		ctx context.Context,
		cwd string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "remove" {
			return []byte("simulated native removal failure"), fmt.Errorf("simulated failure")
		}
		return run(ctx, cwd, name, args...)
	}

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected native worktree removal failure")
	}
	decision := findSyncLane(t, report, "topic/native-remove-failure")
	if decision.Action != "preserved" || decision.Reason != "worktree-removal-failed" {
		t.Fatalf("lane decision = %#v", decision)
	}
	if _, err := os.Stat(buildArtifact); !os.IsNotExist(err) {
		t.Fatalf("disposable build output was restored or stat failed: %v", err)
	}
	assertBranch(t, path, "topic/native-remove-failure")
}

func configureIgnoredBuildOutput(t *testing.T, fixture gitFixture, ignoreRules string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(fixture.primary, ".gitignore"), []byte(ignoreRules), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, fixture.primary, "add", ".gitignore")
	runGit(t, fixture.primary, "commit", "-m", "ignore build output")
	runGit(t, fixture.primary, "push", "origin", "main")
}

func writeBuildArtifact(t *testing.T, worktreePath string) string {
	t.Helper()
	path := filepath.Join(worktreePath, "bin", "tool")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("binary\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertMergedLanePreserved(t *testing.T, fixture gitFixture, branch, headOID string) {
	t.Helper()
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		branch: {mergedSyncPR(84, branch, "main", headOID)},
	})
	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("sync error = %v", err)
	}
	decision := findSyncLane(t, report, branch)
	if decision.Action != "preserved" || decision.Reason != "worktree-dirty" {
		t.Fatalf("lane decision = %#v", decision)
	}
	assertBranch(t, canonicalTestLanePath(fixture, branch), branch)
}
