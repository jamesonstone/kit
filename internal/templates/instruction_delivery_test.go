package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestMemoryWorktreeReferenceIsManagedAndNative(t *testing.T) {
	generated := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/references/worktrees.md",
	)
	for _, want := range []string{
		"Native `git worktree` commands and ordinary filesystem operations define this",
		"The clone's primary checkout owns the shared repository-root `.env`",
		"The user does not need to",
		"`include` makes the existing diff part of the full repair review",
		"`git worktree remove",
		"Runtime services, databases, ports",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("expected V3 worktree reference to contain %q", want)
		}
	}
	for _, forbidden := range []string{"git wt issue", "--no-link-env"} {
		if strings.Contains(generated, forbidden) {
			t.Fatalf("V3 worktree reference must not require optional wrapper syntax %q", forbidden)
		}
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "worktrees.md"))
	if err != nil {
		t.Fatalf("read checked-in worktree reference: %v", err)
	}
	if string(checkedIn) != generated {
		t.Fatal("checked-in worktree reference is not aligned with the V3 generator")
	}
}

func TestInstructionTemplatesRouteTestingAndEnvironmentValidation(t *testing.T) {
	routes := []string{
		"Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md` and the project's `docs/references/testing.md`",
		"end-to-end and live-integration suites supplement rather than replace them",
	}
	for name, content := range map[string]string{
		"V2 AGENTS.md":                       AgentsMD,
		"V2 CLAUDE.md":                       ClaudeMD,
		"V2 .github/copilot-instructions.md": CopilotInstructionsMD,
	} {
		for _, route := range routes {
			if !strings.Contains(content, route) {
				t.Errorf("expected %s to contain testing route %q", name, route)
			}
		}
	}

	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
	} {
		generatedRLM := fileContentByPath(
			InstructionSupportFiles(version),
			"docs/agents/RLM.md",
		)
		for _, route := range []string{
			"Load `docs/references/rules/testing-and-environment-validation.md` and `docs/references/testing.md` before implementation or validation, including browser automation and browser testing",
		} {
			if !strings.Contains(generatedRLM, route) {
				t.Errorf("expected version %d RLM guidance to contain %q", version, route)
			}
		}
	}

	generatedTesting := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/references/testing.md",
	)
	for _, check := range []string{
		"rules/testing-and-environment-validation.md",
		"## Code-Level Validation",
		"## High-Level Suites",
		"## Environment Preflights",
		"## Credentials And Test Data",
		"## Evidence And Retention",
		"`tests/RUN_STATUS.md`",
		"## Automation And Fallbacks",
		"## Known Gaps",
	} {
		if !strings.Contains(generatedTesting, check) {
			t.Errorf("expected generated testing reference to contain %q", check)
		}
	}
	checkedInTesting, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "testing.md"))
	if err != nil {
		t.Fatalf("read checked-in testing reference: %v", err)
	}
	// testing.md is a project-owned reference seeded by the generator; Kit's
	// own copy records real commands, so require the scaffold structure only.
	for _, heading := range []string{
		"rules/testing-and-environment-validation.md",
		"## Code-Level Validation",
		"## High-Level Suites",
		"## Environment Preflights",
		"## Credentials And Test Data",
		"## Evidence And Retention",
		"## Automation And Fallbacks",
		"## Known Gaps",
	} {
		if !strings.Contains(string(checkedInTesting), heading) {
			t.Errorf("checked-in testing reference missing %q", heading)
		}
	}
	if strings.Contains(string(checkedInTesting), "Document the canonical command") {
		t.Error("checked-in testing reference still contains unfilled template placeholders")
	}
}
