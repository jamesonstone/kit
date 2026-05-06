package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
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

func TestRunRemove_RemovesFeatureAndClearsState(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-alpha", "SPEC.md", "# SPEC\n")
	if _, _, err := ensureFeatureNotesDir(projectRoot, "0001-alpha"); err != nil {
		t.Fatalf("ensureFeatureNotesDir() error = %v", err)
	}
	cfg.SetFeaturePaused("0001-alpha", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restore, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restore()

	setRemoveFlagsForTest(t, true, false)

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runRemove(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(specsDir, "0001-alpha")); !os.IsNotExist(err) {
		t.Fatalf("expected feature directory to be removed, got %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.IsFeaturePaused("0001-alpha") {
		t.Fatalf("expected persisted lifecycle state to be cleared")
	}
	if len(updated.RemovedFeatures) != 1 {
		t.Fatalf("expected one removed feature tombstone, got %d", len(updated.RemovedFeatures))
	}
	if updated.RemovedFeatures[0].DirName != "0001-alpha" {
		t.Fatalf("removed feature DirName = %q, want 0001-alpha", updated.RemovedFeatures[0].DirName)
	}
	if _, err := os.Stat(featureNotesPath(projectRoot, "0001-alpha")); err != nil {
		t.Fatalf("expected feature notes to be retained, got %v", err)
	}
	if !strings.Contains(out.String(), "status: removed") {
		t.Fatalf("expected remove output to show removed status, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Retained notes at docs/notes/0001-alpha") {
		t.Fatalf("expected remove output to show retained notes, got %q", out.String())
	}

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", summaryPath, err)
	}
	content := string(data)
	if !strings.Contains(content, "| 0001 | alpha | `docs/specs/0001-alpha` | removed | no |") {
		t.Fatalf("expected removed feature row in rollup, got %q", content)
	}
	if !strings.Contains(content, "- **STATUS**: removed") {
		t.Fatalf("expected removed feature summary in rollup, got %q", content)
	}
}

func TestConfirmFeatureRemoval(t *testing.T) {
	useTestStdin(t, "yes\n")

	confirmed, err := confirmFeatureRemoval(&featureRefAlpha)
	if err != nil {
		t.Fatalf("confirmFeatureRemoval() error = %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmation to accept yes")
	}
}

func TestOutputStatusTextShowsPausedState(t *testing.T) {
	status := pausedStatus()
	out := &bytes.Buffer{}

	if err := outputStatusText(out, status, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "Paused: yes") {
		t.Fatalf("expected paused lifecycle section, got %q", content)
	}
	if !strings.Contains(content, "Active Feature: 0001-alpha") {
		t.Fatalf("expected active feature header, got %q", content)
	}
}

func TestConfirmFeatureNotesRemoval(t *testing.T) {
	useTestStdin(t, "yes\n")

	confirmed, err := confirmFeatureNotesRemoval(&featureRefAlpha, "docs/notes/0001-alpha")
	if err != nil {
		t.Fatalf("confirmFeatureNotesRemoval() error = %v", err)
	}
	if !confirmed {
		t.Fatal("expected notes confirmation to accept yes")
	}
}
func setupLifecycleTestProject(t *testing.T) (string, *config.Config) {
	t.Helper()

	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "docs", "specs"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	return projectRoot, cfg
}

func setRemoveFlagsForTest(t *testing.T, yes bool, notes bool) {
	t.Helper()
	previousYes := removeYes
	previousNotes := removeNotes
	removeYes = yes
	removeNotes = notes
	t.Cleanup(func() {
		removeYes = previousYes
		removeNotes = previousNotes
	})
}

func useTestStdin(t *testing.T, input string) {
	t.Helper()

	originalStdin := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		os.Stdin = originalStdin
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close() error = %v", err)
		}
	})

	if _, err := writer.WriteString(input); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	os.Stdin = reader
}

var featureRefAlpha = feature.Feature{
	DirName: "0001-alpha",
	Path:    "/tmp/0001-alpha",
}

func pausedStatus() *feature.FeatureStatus {
	return &feature.FeatureStatus{
		ID:     "0001",
		Name:   "alpha",
		Phase:  feature.PhaseSpec,
		Paused: true,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: false, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: false, Path: "/tmp/TASKS.md"},
		},
	}
}
