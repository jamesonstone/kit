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
	if artifact, _ := loadMigratedConfig(t, root).RegistryArtifact(rulesetKind, "deletion-safety"); artifact.State != registryArtifactStateLocalCustom {
		t.Fatalf("skipped linked rule recorded as %q, want local-custom", artifact.State)
	}
}

// Linked entry files converge whichever name is the link.
func TestLinkedEntryFilesConvergeInEitherDirection(t *testing.T) {
	for _, link := range []struct{ name, target string }{{"CLAUDE.md", "AGENTS.md"}, {"AGENTS.md", "CLAUDE.md"}} {
		t.Run(link.name, func(t *testing.T) {
			setupMigrationEnvironment(t)
			root := copyMigrationFixture(t, "v3-precontract", func(root string) {
				path := filepath.Join(root, link.name)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(link.target, path); err != nil {
					t.Fatal(err)
				}
			})
			setWorkingDirectory(t, root)
			plan := migrateProject(t, root, false)
			if strings.Contains(strings.Join(plan.notes, "\n"), "edited Kit sections") {
				t.Fatalf("false edited-section diagnostic: %v", plan.notes)
			}
			if info, err := os.Lstat(filepath.Join(root, link.name)); err != nil || info.Mode()&os.ModeSymlink == 0 {
				t.Fatal("link replaced")
			}
			if loadMigratedConfig(t, root).InstructionScaffoldVersion != config.CurrentInstructionScaffoldVersion {
				t.Fatal("linked entry files did not converge")
			}
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second pass not a no-op: %v", changes)
			}
		})
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
