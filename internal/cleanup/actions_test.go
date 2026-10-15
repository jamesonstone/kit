package cleanup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEligibleWorktreeRemovalAndRemoteRestoration(t *testing.T) {
	f := newCleanupFixture(t)
	r := f.execute()
	assertCleanupState(t, r, "REMOVED")
	if r.Bytes <= 0 {
		t.Fatal("cleanup reported no reclaimed file bytes")
	}
	assertAbsent(t, f.lane)
	restored := filepath.Join(f.base, "restored")
	f.git(f.repo, "fetch", "-q", "origin", "lane")
	f.git(f.repo, "worktree", "add", "-q", "--detach", restored, "FETCH_HEAD")
	data, err := os.ReadFile(filepath.Join(restored, "file"))
	if err != nil || string(data) != "recoverable content\n" {
		t.Fatalf("remote restoration: %q %v", data, err)
	}
}
func TestEligibleBranchCASDeletion(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "worktree", "remove", f.lane)
	f.receipt = f.record("branch", f.repo, "refs/heads/lane")
	assertCleanupState(t, f.execute(), "PRUNED")
	if _, err := treeTestRun(f.ctx, f.repo, "git", "show-ref", "--verify", "refs/heads/lane"); err == nil {
		t.Fatal("branch survived")
	}
	if got := f.git(f.repo, "ls-remote", "origin", "refs/heads/lane"); !strings.Contains(got, f.head) {
		t.Fatal("recovery ref lost")
	}
}
func TestEligibleStaleTrackingRefPreservesLocalRecovery(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "push", "-q", "origin", ":lane")
	f.git(f.repo, "update-ref", "refs/remotes/origin/stale", f.head)
	f.receipt = f.record("ref", f.repo, "refs/remotes/origin/stale")
	assertCleanupState(t, f.execute(), "PRUNED")
	if got := f.git(f.repo, "rev-parse", "refs/heads/lane"); got != f.head {
		t.Fatal("local recovery changed")
	}
}
func TestEligibleStaleWorktreeRegistration(t *testing.T) {
	f := newCleanupFixture(t)
	if err := os.RemoveAll(f.lane); err != nil {
		t.Fatal(err)
	}
	assertCleanupState(t, f.execute(), "PRUNED")
	if got := f.git(f.repo, "worktree", "list", "--porcelain"); strings.Contains(got, f.lane) {
		t.Fatal("stale registration survived")
	}
}
func TestEligibleGeneratedArtifactsReclaimBytes(t *testing.T) {
	for _, kind := range []string{"scratch", "cache"} {
		t.Run(kind, func(t *testing.T) {
			f := newCleanupFixture(t)
			path := filepath.Join(f.base, kind)
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatal(err)
			}
			payload := strings.Repeat("generated", 8192)
			mustTreeWrite(t, filepath.Join(path, "payload"), payload)
			f.receipt = f.record(kind, path, "")
			r := f.execute()
			assertCleanupState(t, r, "REMOVED")
			if r.Bytes != int64(len(payload)) {
				t.Fatalf("reclaimed %d, expected %d", r.Bytes, len(payload))
			}
			assertAbsent(t, path)
		})
	}
}
func TestEligiblePrimaryFastForward(t *testing.T) {
	for _, other := range []bool{false, true} {
		t.Run(map[bool]string{false: "default-checked-out", true: "other-branch"}[other], func(t *testing.T) {
			f := newCleanupFixture(t)
			f.receipt = f.record("primary", f.repo, "refs/heads/main")
			mustTreeWrite(t, filepath.Join(f.lane, "file"), "new remote content\n")
			f.git(f.lane, "add", "file")
			f.git(f.lane, "commit", "-q", "-m", "remote advance")
			next := f.git(f.lane, "rev-parse", "HEAD")
			f.git(f.lane, "push", "-q", "origin", "HEAD:main")
			if other {
				f.git(f.repo, "switch", "-q", "-c", "active")
				mustTreeWrite(t, filepath.Join(f.repo, "file"), "active unique work\n")
			}
			assertCleanupState(t, f.execute(), "UPDATED")
			if got := f.git(f.repo, "rev-parse", "main"); got != next {
				t.Fatal("default branch not advanced")
			}
			want := "new remote content\n"
			if other {
				want = "active unique work\n"
			}
			data, err := os.ReadFile(filepath.Join(f.repo, "file"))
			if err != nil || string(data) != want {
				t.Fatalf("working files %q %v", data, err)
			}
		})
	}
}

func TestBranchChangedDuringCleanupIsRetained(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "worktree", "remove", f.lane)
	f.receipt = f.record("branch", f.repo, "refs/heads/lane")
	next := f.git(f.repo, "commit-tree", f.head+"^{tree}", "-p", f.head, "-m", "concurrent unique commit")
	f.s.beforeRevalidate = func() { f.git(f.repo, "update-ref", "refs/heads/lane", next, f.head) }
	r := f.execute()
	if r.State == "PRUNED" {
		t.Fatal("concurrently updated branch removed")
	}
	if got := f.git(f.repo, "rev-parse", "refs/heads/lane"); got != next {
		t.Fatal("concurrent branch update lost")
	}
}
func TestPrimarySyncRetainsUniqueWorkingState(t *testing.T) {
	for _, mode := range []string{"ignored", "hidden", "diverged"} {
		t.Run(mode, func(t *testing.T) {
			f := newCleanupFixture(t)
			f.receipt = f.record("primary", f.repo, "refs/heads/main")
			switch mode {
			case "ignored":
				mustTreeWrite(t, filepath.Join(f.repo, "ignored"), "unique primary data")
			case "hidden":
				f.git(f.repo, "update-index", "--assume-unchanged", "file")
				mustTreeWrite(t, filepath.Join(f.repo, "file"), "hidden primary data")
			case "diverged":
				mustTreeWrite(t, filepath.Join(f.repo, "file"), "unique primary commit")
				f.git(f.repo, "add", "file")
				f.git(f.repo, "commit", "-q", "-m", "local unique")
			}
			contents := preservedFiles(t, f.repo)
			r := f.execute()
			if r.State == "UPDATED" {
				t.Fatalf("unsafe primary sync: %+v", r)
			}
			for path, want := range contents {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != want {
					t.Fatalf("primary state changed: %s %v", path, err)
				}
			}
		})
	}
}
