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
	cfg.SetFeaturePaused("0001-alpha", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restore, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restore()

	removeYes = true
	defer func() { removeYes = false }()

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})

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

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", summaryPath, err)
	}
	if strings.Contains(string(data), "alpha") {
		t.Fatalf("expected removed feature to disappear from rollup, got %q", string(data))
	}
}

func TestConfirmFeatureRemoval(t *testing.T) {
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer reader.Close()

	if _, err := writer.WriteString("yes\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	_ = writer.Close()
	os.Stdin = reader

	confirmed, err := confirmFeatureRemoval(&featureRefAlpha)
	if err != nil {
		t.Fatalf("confirmFeatureRemoval() error = %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmation to accept yes")
	}
}

func TestOutputStatusTextShowsPausedState(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-alpha", "SPEC.md", "# SPEC\n")
	cfg.SetFeaturePaused("0001-alpha", true)

	status := pausedStatus()
	out := &bytes.Buffer{}

	if err := outputStatusText(out, status, specsDir, cfg, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "Paused: yes") {
		t.Fatalf("expected paused lifecycle section, got %q", content)
	}
	if !strings.Contains(content, "PAUSED") || !strings.Contains(content, "| 0001-alpha") {
		t.Fatalf("expected paused all-features table, got %q", content)
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
