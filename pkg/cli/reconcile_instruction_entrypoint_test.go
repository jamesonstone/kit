package cli

import (
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRootInstructionMaxLinesIncludesCurrentGeneratedContract(t *testing.T) {
	want := countLines(templates.MemoryAgentsMD)
	if want <= rootInstructionMinimumMaxLines {
		t.Fatalf("V3 AGENTS.md has %d lines, want above baseline %d for regression coverage", want, rootInstructionMinimumMaxLines)
	}

	got := rootInstructionMaxLines(
		instructions.AgentsMDPath,
		config.InstructionScaffoldVersionMemory,
	)
	if got != want {
		t.Fatalf("rootInstructionMaxLines() = %d, want generated V3 AGENTS.md lines %d", got, want)
	}
}
