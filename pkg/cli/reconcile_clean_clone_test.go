package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A clean clone of a current project has no local-only .env or .envrc. That
// is not drift: health reports the project current with nothing pending, and
// reconcile from its primary checkout creates no worktree and writes nothing.
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

	cmd := healthCommandForTest(t, "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v", err)
	}
	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatal(err)
	}
	pending := report.Changes.Created + report.Changes.Updated + report.Changes.Merged + report.Changes.Removed
	if report.State != statusKitManagedStateCurrent || pending != 0 || len(report.Files) != 0 {
		t.Fatalf("health on a current clean clone = %#v", report)
	}
	if status := reconcileGitOutput(t, clone, "status", "--porcelain", "--ignored"); status != "" {
		t.Fatalf("health wrote to the clone:\n%s", status)
	}

	var preview strings.Builder
	if _, err := resolveReconcileTarget(&preview, clone, true, true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(preview.String(), "linked worktree") {
		t.Fatalf("current preview claims a worktree would be used: %q", preview.String())
	}

	runManagedReconcileForWorktreeTest(t)
	if branches := reconcileGitOutput(t, clone, "branch", "--list", reconcileBranch); branches != "" {
		t.Fatalf("current clone got branch %q", branches)
	}
	if status := reconcileGitOutput(t, clone, "status", "--porcelain", "--ignored"); status != "" {
		t.Fatalf("current clone was written:\n%s", status)
	}
}
