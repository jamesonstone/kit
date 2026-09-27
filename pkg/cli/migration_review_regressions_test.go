package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func primaryReconcileFixture(t *testing.T) (string, string) {
	t.Helper()
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v3-precontract", nil)
	worktrees := t.TempDir()
	stubReconcileWorktreeRoot(t, worktrees)
	return root, worktrees
}

func TestReconcileThroughSymlinkedPathNeverWritesPrimary(t *testing.T) {
	root, _ := primaryReconcileFixture(t)
	link := filepath.Join(t.TempDir(), "via-link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	setWorkingDirectory(t, link)
	runManagedReconcileForWorktreeTest(t)
	if status := reconcileGitOutput(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("primary checkout changed through a symlinked path:\n%s", status)
	}
}

func TestReconcileRefusesPrimaryOnReconcileBranch(t *testing.T) {
	root, _ := primaryReconcileFixture(t)
	runGitForSourceAuditTest(t, root, "checkout", "-q", "-b", reconcileBranch)
	setWorkingDirectory(t, root)
	resetReconcileFlags(t)
	target, err := resolveReconcileTarget(&strings.Builder{}, root, false)
	if err == nil || !strings.Contains(err.Error(), "primary checkout is on branch") {
		t.Fatalf("target = %#v, err = %v", target, err)
	}
	if status := reconcileGitOutput(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("primary changed:\n%s", status)
	}
}

func TestReconcileRefusesStaleWorktree(t *testing.T) {
	root, worktrees := primaryReconcileFixture(t)
	setWorkingDirectory(t, root)
	runManagedReconcileForWorktreeTest(t)
	if err := os.RemoveAll(filepath.Join(worktrees, filepath.Base(root), reconcileBranch)); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveReconcileTarget(&strings.Builder{}, root, false); err == nil || !strings.Contains(err.Error(), "git worktree prune") {
		t.Fatalf("expected stale worktree refusal, got %v", err)
	}
}

func TestForcedWholeProjectReconcileKeepsProjectConfig(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "phase2", func(root string) {
		path := filepath.Join(root, config.ConfigFileName)
		content := strings.Replace(readFile(t, path), "specs_dir: docs/specs", "specs_dir: design/specs", 1)
		writeFile(t, path, content+"source_file_line_limit: 400\nhealth:\n    managed: false\n")
	})
	setWorkingDirectory(t, root)
	migrateProject(t, root, true)
	cfg := loadMigratedConfig(t, root)
	if cfg.SpecsDir != "design/specs" || cfg.SourceFileLineLimit != 400 || cfg.IsHealthManaged() {
		t.Fatalf("forced reconcile discarded project config: %#v", cfg)
	}
}

func TestEntryMigrationKeepsProjectTitleAndFencedExamples(t *testing.T) {
	setupMigrationEnvironment(t)
	example := "```markdown\n## Constraints\n\n- example only\n```\n"
	root := copyMigrationFixture(t, "v2", func(root string) {
		path := filepath.Join(root, "AGENTS.md")
		content := strings.Replace(readFile(t, path), "# AGENTS", "# Acme Payments Agent Guide", 1)
		writeFile(t, path, content+"\n## Examples\n\n"+example)
	})
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	got := readFile(t, filepath.Join(root, "AGENTS.md"))
	want := "# Acme Payments Agent Guide\n\n" + templates.UniversalContractBlock() + "\n## Examples\n\n" + example
	if got != want {
		t.Fatalf("AGENTS.md =\n%s\nwant\n%s", got, want)
	}
}

func TestConfigRewriteReportsDroppedKeysAndComments(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v1", func(root string) {
		path := filepath.Join(root, config.ConfigFileName)
		writeFile(t, path, "# owned by payments\n"+readFile(t, path)+"team_owner: payments\n")
	})
	setWorkingDirectory(t, root)
	notes := strings.Join(migrateProject(t, root, false).notes, "\n")
	for _, want := range []string{"removed keys Kit no longer reads:", "team_owner", "goal_percentage", "comments are not preserved"} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes missing %q:\n%s", want, notes)
		}
	}
}

func TestSymlinkedEntryFilesConvergeInOnePass(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v2", func(root string) {
		claude := filepath.Join(root, "CLAUDE.md")
		if err := os.Remove(claude); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("AGENTS.md", claude); err != nil {
			t.Fatal(err)
		}
	})
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("second pass not a no-op: %v", changes)
	}
	if loadMigratedConfig(t, root).InstructionScaffoldVersion != config.CurrentInstructionScaffoldVersion {
		t.Fatal("symlinked entry files did not converge")
	}
}
