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
		"Implement every remaining non-blocked task",
		"implementation readiness gate",
		"unambiguous, in scope, mapped to acceptance/evidence",
		"material choice remains non-discoverable",
		"do not request routine approval for safe in-scope work",
		"kit legacy verify sample --task <task-id>",
		"record exact validation evidence",
		"Repeat in dependency order until every non-blocked task is complete",
		"Do not stop after one task",
		"Delivery Contract",
		"## Success And Output",
		"mapped acceptance criteria are complete with validation and documentation evidence",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
	assertFinalResponseContractHeadings(t, prompt,
		"Work Done",
		"Files Changed",
		"Validation",
		"How To Test",
		"How To View",
		"Docs/Tasks Updated",
		"Follow-ups",
	)
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
		"Load only when present and relevant",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
}
