package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
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
		"Open `TASKS.md` first",
		"Load only the relevant `PLAN.md` section",
		"Load only the relevant `SPEC.md` requirement",
		"Load `CONSTITUTION.md` only when required",
		"Use an RLM-style just-in-time prior-work pass over",
		filepath.Join(root, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
		"Inspect the relevant code before editing",
		"Run relevant validation before completion",
		"Start by opening TASKS.md and selecting the first incomplete unblocked task",
		"Continue the task loop until every non-blocked task in TASKS.md is complete",
		"Select the next incomplete unblocked task in dependency order",
		"Repeat with the next incomplete task",
		"Do not stop after one task",
		"Complete every non-blocked task in dependency order from TASKS.md before reporting implementation complete",
		"until every non-blocked task in TASKS.md is complete",
		"Only for unresolved rationale; non-binding context",
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

func TestBuildImplementationPrompt_IncludesRepoDocsForTOCRepos(t *testing.T) {
	root := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(root, "docs", "agents", "README.md"), "# Agents Docs\n")
	writeFile(t, filepath.Join(root, "docs", "references", "README.md"), "# References\n")

	prompt := buildImplementationPrompt(
		&feature.Feature{Slug: "sample", DirName: "0001-sample"},
		filepath.Join(root, "docs", "specs", "0001-sample", "BRAINSTORM.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "SPEC.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "PLAN.md"),
		filepath.Join(root, "docs", "specs", "0001-sample", "TASKS.md"),
		"Ship safer implementation handoffs.",
		root,
	)

	for _, check := range []string{
		"docs/agents/README.md",
		"docs/references/README.md",
		"Repo-local runtime routing index",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
}
