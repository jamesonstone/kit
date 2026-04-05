package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func TestBuildResumeCandidatesOrdersPausedActiveBacklog(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")

	createFeatureFile(t, specsDir, "0001-paused-plan", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0002-active-tasks", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0002-active-tasks", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0003-deferred-refactor", "BRAINSTORM.md", "# BRAINSTORM\n")

	cfg.SetFeaturePaused("0001-paused-plan", true)
	cfg.SetFeaturePaused("0003-deferred-refactor", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	candidates, err := buildResumeCandidates(specsDir, cfg)
	if err != nil {
		t.Fatalf("buildResumeCandidates() error = %v", err)
	}

	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(candidates))
	}

	got := []string{candidates[0].Label, candidates[1].Label, candidates[2].Label}
	wantChecks := []string{
		"0001-paused-plan (paused, spec)",
		"0002-active-tasks (active, plan)",
		"0003-deferred-refactor (backlog, brainstorm)",
	}

	for i, want := range wantChecks {
		if got[i] != want {
			t.Fatalf("candidate %d = %q, want %q", i, got[i], want)
		}
	}
}

func TestRunResume_BacklogTargetClearsPausedState(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-legacy-endpoint-refactor", "BRAINSTORM.md", `# BRAINSTORM

## SUMMARY

Need to refactor the legacy endpoint.
`)
	cfg.SetFeaturePaused("0001-legacy-endpoint-refactor", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreStdout := silenceStdout(t)
	defer restoreStdout()

	previousCopy := resumeCopy
	defer func() { resumeCopy = previousCopy }()
	resumeCopy = false

	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", false, "")
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}

	if err := runResume(cmd, []string{"legacy-endpoint-refactor"}); err != nil {
		t.Fatalf("runResume() error = %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.IsFeaturePaused("0001-legacy-endpoint-refactor") {
		t.Fatal("expected resume backlog path to clear paused state")
	}
}

func TestRunResume_NonBacklogTargetUsesCatchupPrompt(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-active-feature", "SPEC.md", `# SPEC

## SUMMARY

Active feature summary.
`)
	createFeatureFile(t, specsDir, "0001-active-feature", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0001-active-feature", "TASKS.md", "- [ ] do work\n")

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	previousCopy := resumeCopy
	defer func() { resumeCopy = previousCopy }()
	resumeCopy = false

	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", false, "")
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runResume(cmd, []string{"active-feature"}); err != nil {
			t.Fatalf("runResume() error = %v", err)
		}
	})

	if !strings.Contains(output, "Catch up on feature: active-feature") {
		t.Fatalf("expected catchup prompt, got %q", output)
	}

	reloaded, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if reloaded.IsFeaturePaused("0001-active-feature") {
		t.Fatal("did not expect non-backlog resume path to change paused state")
	}
}
