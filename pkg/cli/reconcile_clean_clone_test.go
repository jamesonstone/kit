package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// A clean clone of a current project has no local-only .env or .envrc. That
// is not drift: reconcile from its primary checkout creates no worktree and
// writes nothing, matching what `kit health` reports.
func TestCleanCloneOfCurrentProjectReconcilesWithoutWriting(t *testing.T) {
	setupMigrationEnvironment(t)
	source := freshInitProject(t)
	clone := filepath.Join(t.TempDir(), "clone")
	runGitForSourceAuditTest(t, source, "clone", "-q", source, clone)
	for _, file := range []string{envPath, envrcPath} {
		if _, err := os.Lstat(filepath.Join(clone, file)); !os.IsNotExist(err) {
			t.Fatalf("fixture clone unexpectedly has %s", file)
		}
	}
	stubReconcileWorktreeRoot(t, t.TempDir())
	setWorkingDirectory(t, clone)

	runManagedReconcileForWorktreeTest(t)
	if branches := reconcileGitOutput(t, clone, "branch", "--list", reconcileBranch); branches != "" {
		t.Fatalf("current clone got branch %q", branches)
	}
	if status := reconcileGitOutput(t, clone, "status", "--porcelain", "--ignored"); status != "" {
		t.Fatalf("current clone was written:\n%s", status)
	}
}
