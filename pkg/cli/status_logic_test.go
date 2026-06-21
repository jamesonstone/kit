package cli

import (
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

func TestDetermineNextAction_RemovedFeatureMentionsRetainedNotes(t *testing.T) {
	status := &feature.FeatureStatus{
		Name:    "patient-import",
		Removed: true,
		Notes: &feature.FileStatus{
			Exists: true,
			Path:   "/repo/docs/notes/0001-patient-import",
		},
	}

	got := determineNextAction(status)
	if !strings.Contains(got, "Feature is removed") {
		t.Fatalf("expected removed guidance, got %q", got)
	}
	if !strings.Contains(got, "/repo/docs/notes/0001-patient-import") {
		t.Fatalf("expected retained notes path, got %q", got)
	}
}

func TestDetermineNextAction_NoSpecSuggestsV2SpecAndLegacyBrainstorm(t *testing.T) {
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
	if !strings.Contains(got, "kit spec patient-import") {
		t.Fatalf("expected spec guidance, got %q", got)
	}
	if !strings.Contains(got, "kit legacy brainstorm patient-import") {
		t.Fatalf("expected legacy brainstorm guidance, got %q", got)
	}
}

func TestDetermineNextAction_PausedWrapsResumeGuidance(t *testing.T) {
	status := &feature.FeatureStatus{
		Name:   "patient-import",
		Paused: true,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: true, Path: "/tmp/BRAINSTORM.md"},
			"spec":       {Exists: true, Path: "/tmp/SPEC.md"},
			"plan":       {Exists: true, Path: "/tmp/PLAN.md"},
			"tasks":      {Exists: true, Path: "/tmp/TASKS.md"},
		},
		Progress: &feature.TaskProgress{Total: 4, Complete: 1},
	}

	got := determineNextAction(status)
	if !strings.Contains(got, "Feature is paused") {
		t.Fatalf("expected paused guidance, got %q", got)
	}
	if !strings.Contains(got, "Suggested next step after resume") {
		t.Fatalf("expected resume guidance, got %q", got)
	}
	if !strings.Contains(got, "kit resume patient-import") {
		t.Fatalf("expected explicit resume command, got %q", got)
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
	if !strings.Contains(got, "kit legacy implement patient-import") {
		t.Fatalf("expected implement guidance, got %q", got)
	}
	if !strings.Contains(got, "review and validate implementation") {
		t.Fatalf("expected validation guidance, got %q", got)
	}
}
