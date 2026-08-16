package cli

import (
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRootInstructionMaxLinesIncludesCurrentGeneratedContract(t *testing.T) {
	generatedLines := countLines(templates.MemoryAgentsMD)
	if generatedLines <= rootInstructionMinimumMaxLines {
		t.Fatalf("V3 AGENTS.md has %d lines, want above baseline %d for regression coverage", generatedLines, rootInstructionMinimumMaxLines)
	}
	want := max(
		generatedLines+rootInstructionCustomizationAllowanceLines,
		rootInstructionMinimumMaxLines,
	)

	got := rootInstructionMaxLines(
		instructions.AgentsMDPath,
		config.InstructionScaffoldVersionMemory,
	)
	if got != want {
		t.Fatalf("rootInstructionMaxLines() = %d, want generated V3 AGENTS.md plus customization allowance %d", got, want)
	}
}
