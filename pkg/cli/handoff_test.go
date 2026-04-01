package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestSelectFeatureForHandoffAllowsProjectWideSelection(t *testing.T) {
	specsDir := t.TempDir()
	mustMkdirAll(t, filepath.Join(specsDir, "0001-alpha"))
	mustMkdirAll(t, filepath.Join(specsDir, "0002-beta"))

	feat, projectWide, err := selectFeatureForHandoffWithIO(
		specsDir,
		strings.NewReader("0\n"),
		io.Discard,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if feat != nil {
		t.Fatalf("expected no feature, got %#v", feat)
	}
	if !projectWide {
		t.Fatal("expected project-wide selection to be true")
	}
}

func TestSelectFeatureForHandoffReturnsSelectedFeature(t *testing.T) {
	specsDir := t.TempDir()
	mustMkdirAll(t, filepath.Join(specsDir, "0001-alpha"))
	mustMkdirAll(t, filepath.Join(specsDir, "0002-beta"))

	feat, projectWide, err := selectFeatureForHandoffWithIO(
		specsDir,
		strings.NewReader("2\n"),
		io.Discard,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if projectWide {
		t.Fatal("expected feature selection, got project-wide selection")
	}
	if feat == nil {
		t.Fatal("expected feature selection, got nil")
	}
	if feat.Slug != "beta" {
		t.Fatalf("expected beta, got %s", feat.Slug)
	}
}

func TestProjectHandoffIncludesProgressSummaryAndStatus(t *testing.T) {
	projectRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(projectRoot, ".kit.yaml"), []byte{})
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
		[]byte("# CONSTITUTION\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		[]byte("# PROJECT_PROGRESS_SUMMARY\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"),
		[]byte("# SPEC\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "specs", "0002-beta", "TASKS.md"),
		[]byte("- [ ] T001: first task\n"),
	)

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	output, err := projectHandoffWithConfig(projectRoot, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	checks := []string{
		"You are the current coding agent session preparing this project for handoff.",
		"## Documentation Inventory",
		"| File | Full Path | How To Use |",
		filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"),
		filepath.Join(projectRoot, "docs", "specs", "0002-beta", "TASKS.md"),
		"Update any stale feature docs first.",
		"phase dependency inventories",
		"`## DEPENDENCIES` table lists current `active`, `optional`, and `stale` dependencies with exact locations",
		"## Final Response Contract",
		"`Documentation Files`",
		"`Recent Context`",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func TestFeatureHandoffIncludesDocSyncInstructions(t *testing.T) {
	projectRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(projectRoot, ".kit.yaml"), []byte{})
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
		[]byte("# CONSTITUTION\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		[]byte("# PROJECT_PROGRESS_SUMMARY\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"),
		[]byte("# SPEC\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "PLAN.md"),
		[]byte("# PLAN\n"),
	)
	mustWriteFile(
		t,
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "TASKS.md"),
		[]byte("- [ ] T001: first task\n"),
	)

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("failed to prepare working directory: %v", err)
	}
	t.Cleanup(restoreWD)

	output, err := featureHandoff("alpha")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	checks := []string{
		"You are the current coding agent session preparing this feature for handoff.",
		"## Documentation Inventory",
		"| File | Full Path | How To Use |",
		filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"),
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "PLAN.md"),
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "TASKS.md"),
		"Compare current implementation reality, task status, repository findings, and phase dependency inventories against each feature document",
		"If any feature specification document is stale, update it first",
		"Verify that `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` keep their `## DEPENDENCIES` tables current",
		"all relevant documentation files and dependency inventories have been updated and are up to date",
		"## Final Response Contract",
		"`Documentation Sync`",
		"`Documentation Files`",
		"`Recent Context`",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func TestGenericHandoffIncludesRecentContextAndFinalResponseContract(t *testing.T) {
	output := genericHandoffInstructions()

	checks := []string{
		"You are the current coding agent session preparing this project for handoff.",
		"Summarize that recent context into high-signal facts",
		"`## DEPENDENCIES` tables",
		"`Documentation Files`",
		"`Recent Context`",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
