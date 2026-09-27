package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/instructions"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestRootInstructionMaxLinesIncludesCurrentGeneratedContract(t *testing.T) {
	generatedLines := countLines(templates.MemoryAgentsMD)
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
