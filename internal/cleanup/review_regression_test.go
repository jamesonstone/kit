package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultBranchProtectedWhilePrimaryOnOtherBranch(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "switch", "-q", "-c", "active")
	f.receipt = f.record("branch", f.repo, "refs/heads/main")
	assertCleanupState(t, f.execute(), "KEEP_DEFAULT")
	if got := f.git(f.repo, "rev-parse", "main"); got != f.head {
		t.Fatal("default branch changed")
	}
}
func TestReferenceControlMetadataRetained(t *testing.T) {
	for _, name := range []string{"refs/heads/lane", "packed-refs", "logs/refs/heads/lane"} {
		t.Run(name, func(t *testing.T) {
			f := newCleanupFixture(t)
			f.git(f.repo, "worktree", "remove", f.lane)
			if name == "packed-refs" {
				f.git(f.repo, "pack-refs", "--all")
			}
			f.receipt = f.record("branch", f.repo, "refs/heads/lane")
			path := filepath.Join(f.repo, ".git", filepath.FromSlash(name))
			if err := os.Chmod(path, 0600); err != nil {
				t.Fatal(err)
			}
			assertCleanupState(t, f.execute(), "KEEP_LOCAL_STATE")
			if got := f.git(f.repo, "rev-parse", "lane"); got != f.head {
				t.Fatal("branch lost")
			}
		})
	}
}
func TestPrimaryControlMetadataRetained(t *testing.T) {
	for _, name := range []string{"ORIG_HEAD", "logs/HEAD", "logs/refs/heads/main", "refs/heads/main"} {
		t.Run(name, func(t *testing.T) {
			f := newCleanupFixture(t)
			f.receipt = f.record("primary", f.repo, "")
			path := filepath.Join(f.repo, ".git", filepath.FromSlash(name))
			if name == "ORIG_HEAD" {
				mustTreeWrite(t, path, f.head+"\n")
			}
			if err := os.Chmod(path, 0600); err != nil {
				t.Fatal(err)
			}
			assertCleanupState(t, f.execute(), "KEEP_LOCAL_STATE")
		})
	}
}
func TestPrimaryFetchDoesNotPruneConfiguredSelections(t *testing.T) {
	f := newCleanupFixture(t)
	f.receipt = f.record("primary", f.repo, "")
	f.git(f.repo, "update-ref", "refs/remotes/origin/stale", f.head)
	f.git(f.repo, "config", "fetch.prune", "true")
	f.git(f.repo, "config", "remote.origin.prune", "true")
	f.git(f.repo, "config", "fetch.pruneTags", "true")
	f.git(f.repo, "config", "fetch.recurseSubmodules", "true")
	mustTreeWrite(t, filepath.Join(f.lane, "file"), "remote advance\n")
	f.git(f.lane, "add", "file")
	f.git(f.lane, "commit", "-q", "-m", "advance")
	f.git(f.lane, "push", "-q", "origin", "HEAD:main")
	assertCleanupState(t, f.execute(), "UPDATED")
	if got := f.git(f.repo, "rev-parse", "refs/remotes/origin/stale"); got != f.head {
		t.Fatal("configured prune ran")
	}
}
func TestLocalGraftsCannotAuthorizeRecovery(t *testing.T) {
	f := newCleanupFixture(t)
	mustTreeWrite(t, filepath.Join(f.repo, ".git", "info", "grafts"), f.head+"\n")
	assertCleanupState(t, f.execute(), "KEEP_LOCAL_STATE")
	assertExists(t, f.lane)
}
func TestPrimaryReceiptUsesTargetMetadata(t *testing.T) {
	f := newCleanupFixture(t)
	receipt, err := f.s.Record(f.ctx, "primary", f.repo, f.lane, "task", "")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.GitDir != filepath.Join(f.repo, ".git") {
		t.Fatal("primary receipt bound to lane")
	}
	mustTreeWrite(t, filepath.Join(f.repo, ".git", "MERGE_HEAD"), f.head+"\n")
	f.receipt = receipt
	assertCleanupState(t, f.execute(), "KEEP_LOCAL_STATE")
	if _, err = f.s.Record(f.ctx, "branch", f.lane, f.repo, "task", "refs/heads/main"); err == nil {
		t.Fatal("unrelated reference scope accepted")
	}
}
func TestRecordedEnvironmentLinksPreserveSource(t *testing.T) {
	f := newCleanupFixture(t)
	source := filepath.Join(f.repo, ".env")
	mustTreeWrite(t, source, "fixture-only environment value")
	if err := os.Symlink(source, filepath.Join(f.lane, ".env")); err != nil {
		t.Fatal(err)
	}
	f.receipt = f.record("worktree", f.lane, "")
	assertCleanupState(t, f.execute(), "REMOVED")
	data, err := os.ReadFile(source)
	if err != nil || string(data) != "fixture-only environment value" {
		t.Fatal("environment source changed")
	}
}
func TestTaskCacheBeforeContainingWorktree(t *testing.T) {
	f := newCleanupFixture(t)
	cache := filepath.Join(f.lane, "task-cache")
	if err := os.Mkdir(cache, 0755); err != nil {
		t.Fatal(err)
	}
	mustTreeWrite(t, filepath.Join(cache, "payload"), "generated")
	lane := f.receipt
	f.receipt = f.record("cache", cache, "")
	assertCleanupState(t, f.execute(), "REMOVED")
	f.receipt = lane
	assertCleanupState(t, f.execute(), "REMOVED")
}

func TestProcessInventoryDistinguishesUnnamedKernelDescriptors(t *testing.T) {
	for _, kind := range []string{"PIPE", "NPOLICY", "NEXUS", "REG", ""} {
		t.Run(kind, func(t *testing.T) {
			f := newCleanupFixture(t)
			original := f.s.run
			f.s.run = func(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
				if name == "lsof" {
					return []byte("p99999\nf4\nt" + kind + "\nn\n"), nil
				}
				return original(ctx, cwd, name, args...)
			}
			if kind == "REG" || kind == "" {
				assertCleanupState(t, f.execute(), "UNAVAILABLE")
				assertExists(t, f.lane)
			} else {
				assertCleanupState(t, f.execute(), "REMOVED")
			}
		})
	}
}
