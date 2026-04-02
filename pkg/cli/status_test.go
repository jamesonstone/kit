package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/feature"
)

func TestDetermineNextAction_BrainstormOnly(t *testing.T) {
	status := &feature.FeatureStatus{
		Name: "patient-import",
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: true, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: false, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: false, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: false, Path: "/tmp/TASKS.md"},
		},
	}

	got := determineNextAction(status)
	want := "Create specification from brainstorm: run `kit spec patient-import`"
	if got != want {
		t.Fatalf("determineNextAction() = %q, want %q", got, want)
	}
}

func TestDetermineNextAction_NoSpecSuggestsBrainstormOrSpec(t *testing.T) {
	status := &feature.FeatureStatus{
		Name: "patient-import",
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: false, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: false, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: false, Path: "/tmp/TASKS.md"},
		},
	}

	got := determineNextAction(status)
	if !strings.Contains(got, "kit brainstorm patient-import") {
		t.Fatalf("expected brainstorm guidance, got %q", got)
	}
	if !strings.Contains(got, "kit spec patient-import") {
		t.Fatalf("expected spec guidance, got %q", got)
	}
}

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

	if err := outputStatusText(out, status, t.TempDir(), "v1.2.3"); err != nil {
		t.Fatalf("outputStatusText() error = %v", err)
	}

	if !strings.Contains(out.String(), "Kit version: v1.2.3") {
		t.Fatalf("expected version line in output, got %q", out.String())
	}
}

func TestDetermineNextAction_AllTasksCompleteMentionsReadinessGate(t *testing.T) {
	status := &feature.FeatureStatus{
		Name: "patient-import",
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: true, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: true, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: true, Path: "/tmp/TASKS.md"},
		},
		Progress: &feature.TaskProgress{Total: 3, Complete: 3},
	}

	got := determineNextAction(status)
	if !strings.Contains(got, "kit implement patient-import") {
		t.Fatalf("expected implement guidance, got %q", got)
	}
	if !strings.Contains(got, "implementation readiness gate") {
		t.Fatalf("expected readiness gate guidance, got %q", got)
	}
	if !strings.Contains(got, "review and verify implementation") {
		t.Fatalf("expected reflection guidance, got %q", got)
	}
}
