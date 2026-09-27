package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetiredFilesGitCannotRestoreAreKept(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v2", nil)
	// Tracked but locally changed, and an untracked exact copy.
	rlm := filepath.Join(root, "docs", "agents", "RLM.md")
	writeFile(t, rlm, readFile(t, rlm)+"\n")
	runGitForSourceAuditTest(t, root, "rm", "-q", "--cached", "docs/agents/TOOLING.md")
	setWorkingDirectory(t, root)

	plan := migrateProject(t, root, false)
	for _, path := range []string{rlm, filepath.Join(root, "docs", "agents", "TOOLING.md")} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s was removed although Git could not restore it", path)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "agents", "README.md")); !os.IsNotExist(err) {
		t.Fatal("committed unmodified support doc was not removed")
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), "docs/agents/TOOLING.md (unmodified, but not committed to Git") {
		t.Fatalf("missing uncommitted note: %v", plan.notes)
	}
}

func TestNothingIsRemovedOutsideGit(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v2", nil)
	if err := os.RemoveAll(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	setWorkingDirectory(t, root)
	for _, change := range migrateProject(t, root, false).changes {
		if change.result == instructionFileRemoved {
			t.Fatalf("removed %s without Git history to restore it", change.relativePath)
		}
	}
}

func TestProgressSummaryOwnership(t *testing.T) {
	setupMigrationEnvironment(t)
	for _, test := range []struct {
		name    string
		content string
		removed bool
	}{
		{name: "generated rollup", content: validProgressSummary("0001", "alpha"), removed: true},
		{name: "hand-maintained", content: validProgressSummary("0001", "alpha") + "\n## MAINTENANCE RULE\n\nkeep\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := copyMigrationFixture(t, "phase2", func(root string) {
				writeFile(t, filepath.Join(root, "docs", "PROJECT_PROGRESS_SUMMARY.md"), test.content)
			})
			setWorkingDirectory(t, root)
			migrateProject(t, root, false)
			_, err := os.Stat(filepath.Join(root, "docs", "PROJECT_PROGRESS_SUMMARY.md"))
			if removed := os.IsNotExist(err); removed != test.removed {
				t.Fatalf("removed = %v, want %v", removed, test.removed)
			}
		})
	}
}

func TestProtectedDeveloperFilesAreNotRewritten(t *testing.T) {
	setupMigrationEnvironment(t)
	protected := map[string]string{
		envPath:                 "SECRET=local\n",
		envrcPath:               "dotenv\nsource_env .custom\n",
		makefilePath:            ".PHONY: test\ntest:\n\tgo test ./...\n",
		codeRabbitConfigPath:    "reviews:\n  profile: assertive\n",
		pullRequestTemplatePath: "## Why\n",
	}
	root := copyMigrationFixture(t, "v3-precontract", func(root string) {
		for path, content := range protected {
			writeFile(t, filepath.Join(root, path), content)
		}
	})
	setWorkingDirectory(t, root)
	migrateProject(t, root, true)
	for path, content := range protected {
		if got := readFile(t, filepath.Join(root, path)); got != content {
			t.Errorf("%s changed:\n%s", path, got)
		}
	}
}

// A clean clone never has the local-only .env and .envrc; migration and the
// project check must not treat their absence as a problem.
func TestCleanCheckoutWithoutLocalEnvironmentFilesPassesProjectCheck(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v3-precontract", nil)
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	for _, path := range []string{envPath, envrcPath} {
		if err := os.Remove(filepath.Join(root, path)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	cfg := loadMigratedConfig(t, root)
	var out bytes.Buffer
	if err := checkProjectContractTo(&out, root, cfg); err != nil {
		t.Fatalf("project check failed on a clean migrated checkout: %v\n%s", err, out.String())
	}
}
