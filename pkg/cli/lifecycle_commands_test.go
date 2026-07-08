package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunPause_PersistsStateAndUpdatesRollup(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	createFeatureFile(t, filepath.Join(projectRoot, "docs", "specs"), "0001-alpha", "SPEC.md", "# SPEC\n")

	restore, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restore()

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runPause(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runPause() error = %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if !updated.IsFeaturePaused("0001-alpha") {
		t.Fatalf("expected paused state to persist in .kit.yaml")
	}

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", summaryPath, err)
	}
	content := string(data)
	if !strings.Contains(content, "| 0001 | alpha | `docs/specs/0001-alpha` | spec | yes |") {
		t.Fatalf("expected paused row in rollup, got %q", content)
	}
	if !strings.Contains(content, "- **PAUSED**: yes") {
		t.Fatalf("expected paused feature summary in rollup, got %q", content)
	}
}

func TestRunRemoveWithNotesRemovesFeatureNotes(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-alpha", "SPEC.md", "# SPEC\n")
	if _, _, err := ensureFeatureNotesDir(projectRoot, "0001-alpha"); err != nil {
		t.Fatalf("ensureFeatureNotesDir() error = %v", err)
	}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restore, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restore()

	setRemoveFlagsForTest(t, true, true)

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runRemove(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	if _, err := os.Stat(featureNotesPath(projectRoot, "0001-alpha")); !os.IsNotExist(err) {
		t.Fatalf("expected feature notes to be removed, got %v", err)
	}
	if !strings.Contains(out.String(), "Removed notes at docs/notes/0001-alpha") {
		t.Fatalf("expected remove output to show removed notes, got %q", out.String())
	}
}

func TestRunRemoveInteractiveCanRemoveFeatureNotes(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-alpha", "SPEC.md", "# SPEC\n")
	if _, _, err := ensureFeatureNotesDir(projectRoot, "0001-alpha"); err != nil {
		t.Fatalf("ensureFeatureNotesDir() error = %v", err)
	}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restore, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restore()

	setRemoveFlagsForTest(t, false, false)
	useTestStdin(t, "yes\nyes\n")

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})

	if err := runRemove(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	if _, err := os.Stat(featureNotesPath(projectRoot, "0001-alpha")); !os.IsNotExist(err) {
		t.Fatalf("expected feature notes to be removed, got %v", err)
	}
}

func TestRunImplement_ClearsPausedState(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-alpha", "SPEC.md", "# SPEC\n\n## SUMMARY\n\nalpha\n")
	createFeatureFile(t, specsDir, "0001-alpha", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0001-alpha", "TASKS.md", "# TASKS\n\n- [ ] do work\n")
	cfg.SetFeaturePaused("0001-alpha", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restore, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restore()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", false, "")
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}
	cmd.SetOut(&bytes.Buffer{})

	if err := runImplement(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runImplement() error = %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.IsFeaturePaused("0001-alpha") {
		t.Fatalf("expected paused state to clear after explicit implement")
	}
}
