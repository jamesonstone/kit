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
