package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

func TestOutputStatusJSONIncludesKitVersion(t *testing.T) {
	status := &feature.FeatureStatus{Name: "patient-import"}
	out := &bytes.Buffer{}

	if err := outputStatusJSON(out, status, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusJSON() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got := payload["kit_version"]; got != "v1.2.3" {
		t.Fatalf("kit_version = %v, want %q", got, "v1.2.3")
	}
}

func TestRunStatusAllIncludesRemovedFeatureTombstones(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	if _, _, err := ensureFeatureNotesDir(projectRoot, "0001-alpha"); err != nil {
		t.Fatalf("ensureFeatureNotesDir() error = %v", err)
	}
	cfg.RecordRemovedFeature(config.RemovedFeature{
		Number:    1,
		Slug:      "alpha",
		DirName:   "0001-alpha",
		CreatedAt: "2026-04-05T00:00:00Z",
		RemovedAt: "2026-05-06T12:00:00Z",
	})
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("all", false, "")
	if err := cmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("Flags().Set(all) error = %v", err)
	}

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runStatus(cmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	content := out.String()
	for _, check := range []string{
		"Active feature: none in progress",
		"0001-alpha",
		"REMOVED",
		"Notes",
		"yes",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got %q", check, content)
		}
	}
}

func TestOutputStatusTextIncludesKitVersion(t *testing.T) {
	status := &feature.FeatureStatus{
		ID:    "0001",
		Name:  "patient-import",
		Phase: feature.PhaseSpec,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: false, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: false, Path: "/tmp/TASKS.md"},
		},
	}
	out := &bytes.Buffer{}

	if err := outputStatusText(out, status, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	if !strings.Contains(out.String(), "Kit version: v1.2.3") {
		t.Fatalf("expected version line in output, got %q", out.String())
	}
}

func TestOutputStatusTextShowsCurrentStepAndRemainingWork(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "SPEC.md")
	writeFile(t, specPath, validV2SpecWithPhase("0001-patient-import", "clarify"))
	status := &feature.FeatureStatus{
		ID:      "0001",
		Name:    "patient-import",
		Phase:   feature.PhaseClarify,
		Summary: "Import patient records from partner feeds.",
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: filepath.Join(dir, "BRAINSTORM.md")},
			"spec":       {Exists: true, Path: specPath},
			"plan":       {Exists: false, Path: filepath.Join(dir, "PLAN.md")},
			"tasks":      {Exists: false, Path: filepath.Join(dir, "TASKS.md")},
		},
	}
	out := &bytes.Buffer{}

	if err := outputStatusText(out, status, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	content := out.String()
	for _, check := range []string{
		"At a glance",
		"State: active (not paused)",
		"Paused: no",
		"Current step: clarify",
		"Left: clarify requirements -> ready gate -> implement -> validate -> reflect -> deliver -> complete",
		"Next: Continue v2 clarification in SPEC.md until unresolved questions are 0 and acceptance criteria are binary-verifiable",
		"V2 phase progress",
		"CLARIFY",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got %q", check, content)
		}
	}
}

func TestOutputStatusTextPausedStartsWithResumeGuidance(t *testing.T) {
	status := &feature.FeatureStatus{
		ID:     "0001",
		Name:   "patient-import",
		Phase:  feature.PhaseImplement,
		Paused: true,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: true, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: true, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: true, Path: "/tmp/TASKS.md"},
		},
		Progress: &feature.TaskProgress{Total: 4, Complete: 1},
	}
	out := &bytes.Buffer{}

	if err := outputStatusText(out, status, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	content := out.String()
	for _, check := range []string{
		"State: paused",
		"Paused: yes",
		"Current step: implementation",
		"Tasks: 1/4 complete (3 left)",
		"Left: complete 3 remaining task(s) -> reflect -> complete",
		"Next: Run `kit resume patient-import` when ready",
		"After resume: Complete 3 remaining task(s) in /tmp/TASKS.md",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got %q", check, content)
		}
	}
}

func TestOutputStatusTextUsesANSIColorWhenTerminalEnabled(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	defer func() { terminalWriterCheck = previousCheck }()

	status := &feature.FeatureStatus{
		ID:    "0001",
		Name:  "patient-import",
		Phase: feature.PhasePlan,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: true, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: true, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: false, Path: "/tmp/TASKS.md"},
		},
	}
	out := &bytes.Buffer{}

	if err := outputStatusText(out, status, "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "\033[") {
		t.Fatalf("expected ANSI color sequences in terminal output, got %q", content)
	}
	if !strings.Contains(content, "Current step:") {
		t.Fatalf("expected current step field in output, got %q", content)
	}
}
