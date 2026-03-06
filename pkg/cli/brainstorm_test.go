package cli

import (
	"strings"
	"testing"
)

func TestBuildBrainstormPrompt(t *testing.T) {
	prompt := buildBrainstormPrompt(
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"sample-feature",
		"/tmp/project",
		"Need better import validation for malformed CSV uploads.",
		95,
	)

	checks := []string{
		"/plan",
		"You are in planning mode for feature: **sample-feature**",
		"Do NOT implement code, write production changes, or move into execution",
		"Continue asking clarifying questions until you reach >= 95% understanding",
		"planning only — no implementation",
		"kit spec sample-feature",
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"/tmp/project/docs/CONSTITUTION.md",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if !strings.HasPrefix(prompt, "/plan\n\n") {
		t.Fatalf("expected prompt to start with /plan, got %q", prompt[:8])
	}
}
