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

func TestOutputStatusTextIncludesKitVersion(t *testing.T) {
	status := &feature.FeatureStatus{
		ID:   "0001",
		Name: "patient-import",
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

func TestOutputAllFeaturesStatusText(t *testing.T) {
	active := &feature.FeatureStatus{ID: "0002", Name: "beta", Phase: feature.PhasePlan}
	entries := []allFeatureStatusEntry{
		{
			Status: &feature.FeatureStatus{
				ID:    "0001",
				Name:  "alpha",
				Phase: feature.PhaseSpec,
				Files: map[string]feature.FileStatus{
					"brainstorm": {Exists: false},
					"spec":       {Exists: true},
					"plan":       {Exists: false},
					"tasks":      {Exists: false},
				},
			},
			IsBacklog:  false,
			NextAction: "kit plan alpha",
		},
		{
			Status: &feature.FeatureStatus{
				ID:    "0002",
				Name:  "beta",
				Phase: feature.PhasePlan,
				Files: map[string]feature.FileStatus{
					"brainstorm": {Exists: false},
					"spec":       {Exists: true},
					"plan":       {Exists: true},
					"tasks":      {Exists: false},
				},
			},
			IsBacklog:  false,
			NextAction: "kit tasks beta",
		},
		{
			Status: &feature.FeatureStatus{
				ID:     "0003",
				Name:   "gamma",
				Phase:  feature.PhaseBrainstorm,
				Paused: true,
				Files: map[string]feature.FileStatus{
					"brainstorm": {Exists: true},
					"spec":       {Exists: false},
					"plan":       {Exists: false},
					"tasks":      {Exists: false},
				},
			},
			IsBacklog:  true,
			NextAction: "kit resume gamma",
		},
	}

	out := &bytes.Buffer{}
	if err := outputAllFeaturesStatusText(out, active, entries, 1, "v1.2.3"); err != nil {
		t.Fatalf("outputAllFeaturesStatusText() error = %v", err)
	}

	content := out.String()
	checks := []string{
		"Project Overview",
		"Active feature: 0002-beta",
		"Backlog items: 1",
		"Feature",
		"BRN",
		"SPEC",
		"PLAN",
		"TASK",
		"IMPL",
		"REFL",
		"DONE",
		"State",
		"Prog",
		"0003-gamma",
		"BACKLOG",
		"Legend: ● complete, ◐ current phase, ○ not reached",
		"Kit version: v1.2.3",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got %q", check, content)
		}
	}
	if strings.Contains(content, "| feature |") {
		t.Fatalf("expected fixed-width matrix instead of markdown table, got %q", content)
	}
}

func TestOutputAllFeaturesStatusText_UsesANSIColorWhenTerminalEnabled(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	defer func() { terminalWriterCheck = previousCheck }()

	active := &feature.FeatureStatus{ID: "0002", Name: "beta", Phase: feature.PhasePlan}
	entries := []allFeatureStatusEntry{
		{
			Status: &feature.FeatureStatus{
				ID:    "0002",
				Name:  "beta",
				Phase: feature.PhasePlan,
				Files: map[string]feature.FileStatus{
					"spec": {Exists: true},
					"plan": {Exists: true},
				},
			},
		},
	}

	out := &bytes.Buffer{}
	if err := outputAllFeaturesStatusText(out, active, entries, 0, "v1.2.3"); err != nil {
		t.Fatalf("outputAllFeaturesStatusText() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "\033[") {
		t.Fatalf("expected ANSI color sequences in terminal output, got %q", content)
	}
	if !strings.Contains(content, "◐") {
		t.Fatalf("expected lifecycle marker in matrix output, got %q", content)
	}
}

func TestOutputAllFeaturesStatusText_DoesNotMarkDuplicateNumericIDsAsActive(t *testing.T) {
	active := &feature.FeatureStatus{
		ID:    "0012",
		Name:  "implement-readiness-gate",
		Path:  "/repo/docs/specs/0012-implement-readiness-gate",
		Phase: feature.PhasePlan,
	}
	entries := []allFeatureStatusEntry{
		{
			Status: &feature.FeatureStatus{
				ID:    "0012",
				Name:  "default-subagent-orchestration",
				Path:  "/repo/docs/specs/0012-default-subagent-orchestration",
				Phase: feature.PhaseSpec,
				Files: map[string]feature.FileStatus{},
			},
		},
		{
			Status: &feature.FeatureStatus{
				ID:    "0012",
				Name:  "implement-readiness-gate",
				Path:  "/repo/docs/specs/0012-implement-readiness-gate",
				Phase: feature.PhasePlan,
				Files: map[string]feature.FileStatus{},
			},
		},
	}

	out := &bytes.Buffer{}
	if err := outputAllFeaturesStatusText(out, active, entries, 0, "v1.2.3"); err != nil {
		t.Fatalf("outputAllFeaturesStatusText() error = %v", err)
	}

	content := out.String()
	if strings.Count(content, "ACTIVE") != 1 {
		t.Fatalf("expected exactly one ACTIVE row, got %q", content)
	}
	if !strings.Contains(content, "0012-implement-readiness-gate") {
		t.Fatalf("expected active feature row, got %q", content)
	}
}

func TestRunStatusAllJSONUsesDedicatedShape(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")

	createFeatureFile(t, specsDir, "0001-active", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0002-backlog", "BRAINSTORM.md", "# BRAINSTORM\n")
	cfg.SetFeaturePaused("0002-backlog", true)
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
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("Flags().Set(json) error = %v", err)
	}
	if err := cmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("Flags().Set(all) error = %v", err)
	}

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runStatus(cmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got := payload["mode"]; got != "all" {
		t.Fatalf("mode = %v, want %q", got, "all")
	}
	if got := payload["backlog_count"]; got != float64(1) {
		t.Fatalf("backlog_count = %v, want %v", got, 1)
	}
	features, ok := payload["features"].([]any)
	if !ok || len(features) != 2 {
		t.Fatalf("features = %v, want 2 entries", payload["features"])
	}
}
