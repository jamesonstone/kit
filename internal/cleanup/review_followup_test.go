package cleanup

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestBranchCleanupAfterPrimaryAdvancesPreservesPrimaryHistory(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "worktree", "remove", f.lane)
	mustTreeWrite(t, filepath.Join(f.repo, "file"), "primary advances beyond the lane")
	f.git(f.repo, "add", "file")
	f.git(f.repo, "commit", "-q", "-m", "primary merge")
	next := f.git(f.repo, "rev-parse", "HEAD")
	f.git(f.repo, "push", "-q", "origin", "main")
	f.git(f.repo, "update-ref", "ORIG_HEAD", next)
	logs := f.git(f.repo, "reflog", "show", "--format=%H:%gs", "HEAD")
	f.receipt = f.record("branch", f.repo, "refs/heads/lane")
	assertCleanupState(t, f.execute(), "PRUNED")
	if f.git(f.repo, "rev-parse", "HEAD") != next || f.git(f.repo, "reflog", "show", "--format=%H:%gs", "HEAD") != logs {
		t.Fatal("primary history changed")
	}
}
func TestPrimaryPreviewMissingRemoteObjectDoesNotClaimDivergenceOrFetch(t *testing.T) {
	f := newCleanupFixture(t)
	f.receipt = f.record("primary", f.repo, "")
	f.git(f.remote, "config", "user.name", "Fixture")
	f.git(f.remote, "config", "user.email", "fixture@example.invalid")
	next := f.git(f.remote, "commit-tree", f.head+"^{tree}", "-p", f.head, "-m", "remote-only advance")
	f.git(f.remote, "update-ref", "refs/heads/main", next)
	f.options.Apply = false
	assertCleanupState(t, f.execute(), "KEEP_NEEDS_FETCH")
	if _, err := treeTestRun(f.ctx, f.repo, "git", "cat-file", "-e", next); err == nil {
		t.Fatal("preview fetched remote objects")
	}
	f.options.Apply = true
	assertCleanupState(t, f.execute(), "UPDATED")
	if got := f.git(f.repo, "rev-parse", "main"); got != next {
		t.Fatal("apply failed to acquire and sync")
	}
}
func TestCheckoutFilterPathsExceedArgumentLimit(t *testing.T) {
	f := newCleanupFixture(t)
	f.git(f.repo, "config", "filter.fixture.smudge", "touch filter-executed")
	var paths strings.Builder
	for i := 0; i < 12000; i++ {
		fmt.Fprintf(&paths, "%s/%05d%c", strings.Repeat("a", 200), i, byte(0))
	}
	original := f.s.run
	f.s.run = func(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
		if name == "git" && strings.Contains(strings.Join(args, " "), "ls-tree") {
			return []byte(paths.String()), nil
		}
		return original(ctx, cwd, name, args...)
	}
	if err := f.s.checkoutFilters(f.ctx, f.repo, f.head); err != nil {
		t.Fatalf("large inactive path list: %v", err)
	}
	mustTreeWrite(t, filepath.Join(f.repo, ".git", "info", "attributes"), "* filter=fixture\n")
	if err := f.s.checkoutFilters(f.ctx, f.repo, f.head); err == nil {
		t.Fatal("large active path list accepted")
	}
	assertAbsent(t, filepath.Join(f.repo, "filter-executed"))
}
