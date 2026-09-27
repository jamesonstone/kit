package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

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
