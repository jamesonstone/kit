package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestMemoryInstructionsRouteSourceFileSizeRule(t *testing.T) {
	routes := []string{
		"docs/references/rules/source-file-size.md",
		"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		"whole-project reconcile and scheduled maintenance",
	}
	for name, content := range map[string]string{
		"V3 AGENTS.md":                       MemoryAgentsMD,
		"V3 CLAUDE.md":                       MemoryClaudeMD,
		"V3 .github/copilot-instructions.md": MemoryCopilotInstructionsMD,
		"V3 GUARDRAILS.md": fileContentByPath(
			InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
			"docs/agents/GUARDRAILS.md",
		),
	} {
		for _, route := range routes {
			if !strings.Contains(content, route) {
				t.Errorf("expected %s to contain source-file-size route %q", name, route)
			}
		}
	}

	references := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/references/README.md",
	)
	if !strings.Contains(references, "rules/source-file-size.md") {
		t.Fatal("generated references index does not route source-file-size")
	}
}
