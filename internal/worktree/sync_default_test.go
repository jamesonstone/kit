package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSyncOptions(t *testing.T) {
	t.Parallel()
	options, err := parseSyncOptions([]string{"--json", "--dry-run"})
	if err != nil {
		t.Fatalf("parseSyncOptions() error = %v", err)
	}
	if !options.JSON || !options.DryRun {
		t.Fatalf("parseSyncOptions() = %#v", options)
	}
	for _, args := range [][]string{
		{"--apply"},
		{"--dry-run", "--dry-run"},
		{"--json", "--json"},
	} {
		if _, err := parseSyncOptions(args); err == nil {
			t.Fatalf("parseSyncOptions(%q) expected an error", args)
		}
	}
}

func TestSyncReconcilesDefaultBranchStates(t *testing.T) {
	t.Run("current", func(t *testing.T) {
		fixture := newGitFixture(t)
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if report.DefaultBranch.State != "current" ||
			report.DefaultBranch.Action != "none" {
			t.Fatalf("default decision = %#v", report.DefaultBranch)
		}
	})

	t.Run("checked out behind", func(t *testing.T) {
		fixture := newGitFixture(t)
		remoteOID := advanceRemoteMain(t, fixture, "remote-behind.txt")
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if report.DefaultBranch.State != "behind" ||
			report.DefaultBranch.Action != "fast-forwarded" ||
			report.DefaultBranch.LocalOID != remoteOID {
			t.Fatalf("default decision = %#v", report.DefaultBranch)
		}
		if got := gitText(t, fixture.primary, "rev-parse", "main"); got != remoteOID {
			t.Fatalf("main = %s, want %s", got, remoteOID)
		}
	})

	t.Run("not checked out behind", func(t *testing.T) {
		fixture := newGitFixture(t)
		runGit(t, fixture.primary, "switch", "-c", "primary-topic")
		remoteOID := advanceRemoteMain(t, fixture, "remote-ref.txt")
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if report.DefaultBranch.Path != "" ||
			report.DefaultBranch.State != "behind" ||
			report.DefaultBranch.Action != "fast-forwarded" ||
			report.DefaultBranch.LocalOID != remoteOID {
			t.Fatalf("default decision = %#v", report.DefaultBranch)
		}
		if got := gitText(t, fixture.primary, "rev-parse", "main"); got != remoteOID {
			t.Fatalf("main = %s, want %s", got, remoteOID)
		}
		if got := gitText(t, fixture.primary, "branch", "--show-current"); got != "primary-topic" {
			t.Fatalf("current branch = %q, want primary-topic", got)
		}
	})

	t.Run("dirty behind", func(t *testing.T) {
		fixture := newGitFixture(t)
		localOID := gitText(t, fixture.primary, "rev-parse", "main")
		advanceRemoteMain(t, fixture, "remote-dirty.txt")
		if err := os.WriteFile(
			filepath.Join(fixture.primary, "preserve.txt"),
			[]byte("local\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if report.DefaultBranch.State != "dirty" ||
			report.DefaultBranch.Action != "preserved" {
			t.Fatalf("default decision = %#v", report.DefaultBranch)
		}
		if got := gitText(t, fixture.primary, "rev-parse", "main"); got != localOID {
			t.Fatalf("dirty main changed from %s to %s", localOID, got)
		}
	})

	t.Run("ahead", func(t *testing.T) {
		fixture := newGitFixture(t)
		runGit(t, fixture.primary, "commit", "--allow-empty", "-m", "local ahead")
		localOID := gitText(t, fixture.primary, "rev-parse", "main")
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if report.DefaultBranch.State != "ahead" ||
			report.DefaultBranch.Action != "preserved" {
			t.Fatalf("default decision = %#v", report.DefaultBranch)
		}
		if got := gitText(t, fixture.primary, "rev-parse", "main"); got != localOID {
			t.Fatalf("ahead main changed from %s to %s", localOID, got)
		}
	})

	t.Run("diverged", func(t *testing.T) {
		fixture := newGitFixture(t)
		runGit(t, fixture.primary, "commit", "--allow-empty", "-m", "local divergence")
		localOID := gitText(t, fixture.primary, "rev-parse", "main")
		advanceRemoteMain(t, fixture, "remote-divergence.txt")
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err != nil {
			t.Fatalf("sync error = %v", err)
		}
		if report.DefaultBranch.State != "diverged" ||
			report.DefaultBranch.Action != "preserved" {
			t.Fatalf("default decision = %#v", report.DefaultBranch)
		}
		if got := gitText(t, fixture.primary, "rev-parse", "main"); got != localOID {
			t.Fatalf("diverged main changed from %s to %s", localOID, got)
		}
	})
}
