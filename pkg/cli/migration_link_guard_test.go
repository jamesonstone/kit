package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

// The reconcile worktree links the primary checkout's local-only .env and
// .envrc; no forced, targeted write may go through those links.
func TestForcedTargetedReconcileNeverWritesThroughEnvironmentLinks(t *testing.T) {
	root, _ := primaryReconcileFixture(t)
	writeFile(t, filepath.Join(root, envrcPath), "export SECRET=primary-only\n")
	writeFile(t, filepath.Join(root, envPath), "TOKEN=primary-only\n")
	setWorkingDirectory(t, root)
	runManagedReconcileForWorktreeTest(t)
	for _, file := range []string{envrcPath, envPath} {
		resetReconcileFlags(t)
		reconcileForce = true
		reconcileRefreshFiles = []string{file}
		runManagedReconcileForWorktreeTestKeepingFlags(t)
	}
	if got := readFile(t, filepath.Join(root, envrcPath)); got != "export SECRET=primary-only\n" {
		t.Fatalf("primary .envrc overwritten: %q", got)
	}
	if got := readFile(t, filepath.Join(root, envPath)); got != "TOKEN=primary-only\n" {
		t.Fatalf("primary .env overwritten: %q", got)
	}
}

func TestForcedWriteNeverFollowsLinkOutsideProject(t *testing.T) {
	setupMigrationEnvironment(t)
	shared := filepath.Join(t.TempDir(), "deletion-safety.md")
	root := copyMigrationFixture(t, "phase2", func(root string) {
		rule := filepath.Join(root, rulesetTarget("deletion-safety"))
		writeFile(t, shared, readFile(t, rule)+"\n- Team edit: never purge prod.\n")
		if err := os.Remove(rule); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(shared, rule); err != nil {
			t.Fatal(err)
		}
	})
	setWorkingDirectory(t, root)
	plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{force: true, files: []string{rulesetTarget("deletion-safety")}, outputOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readFile(t, shared), "never purge prod") {
		t.Fatal("forced write followed a link outside the project")
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), "is a symbolic link") {
		t.Fatalf("missing link note: %v", plan.notes)
	}
}

func TestTargetedRunDoesNotRecordVersionWithUnmigratedEntries(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v3-precontract", nil)
	setWorkingDirectory(t, root)
	plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{files: []string{config.ConfigFileName, rulesetTarget("deletion-safety")}, outputOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatal(err)
	}
	if loadMigratedConfig(t, root).InstructionScaffoldVersion == config.CurrentInstructionScaffoldVersion {
		t.Fatal("version 4 recorded while entry files were not migrated")
	}
}
