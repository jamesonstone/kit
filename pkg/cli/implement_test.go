package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/feature"
)

func TestBuildImplementationPrompt_IncludesReadinessGate(t *testing.T) {
	root := t.TempDir()
	brainstormPath := filepath.Join(root, "docs", "specs", "0001-sample", "BRAINSTORM.md")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")

	prompt := buildImplementationPrompt(
		&feature.Feature{Slug: "sample", DirName: "0001-sample"},
		brainstormPath,
		filepath.Join(root, "docs", "specs", "0001-sample", "SPEC.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "PLAN.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "TASKS.md"),
		"Ship safer implementation handoffs.",
		root,
	)

	checks := []string{
		"implementation readiness gate",
		"adversarial preflight",
		"contradictions, ambiguous requirements, hidden assumptions, missing edge cases or failure modes, task gaps, and scope creep",
		"Produce an explicit go/no-go decision before coding",
		"Do NOT write production code yet",
		"Update SPEC.md, PLAN.md, and/or TASKS.md to resolve the exact issue",
		"Re-run the implementation readiness gate after the docs are fixed",
		"Do not begin coding until the implementation readiness gate passes",
		"Start by running the implementation readiness gate against the document set.",
		"Once it passes, read TASKS.md to identify the first incomplete task",
		"BRAINSTORM.md | Upstream research findings, relevant files, strategy options | Recovering problem context before execution |",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
}

func TestBuildImplementationPrompt_WithoutBrainstormSkipsBrainstormReferences(t *testing.T) {
	root := t.TempDir()
	prompt := buildImplementationPrompt(
		&feature.Feature{Slug: "sample", DirName: "0001-sample"},
		filepath.Join(root, "docs", "specs", "0001-sample", "BRAINSTORM.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "SPEC.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "PLAN.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "TASKS.md"),
		"Ship safer implementation handoffs.",
		root,
	)

	if strings.Contains(prompt, "BRAINSTORM →") {
		t.Fatalf("expected prompt to skip brainstorm ordering when file is absent, got %q", prompt)
	}
	if strings.Contains(prompt, "- BRAINSTORM:") {
		t.Fatalf("expected prompt to skip brainstorm file listing when file is absent, got %q", prompt)
	}
}
