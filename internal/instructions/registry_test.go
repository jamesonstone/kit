package instructions

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionRelativePathsIncludesCopilotOnce(t *testing.T) {
	cfg := config.Default()
	cfg.Agents = []string{AgentsMDPath, ClaudeMDPath, CopilotInstructionsPath}

	got := InstructionRelativePaths(cfg)
	want := []string{AgentsMDPath, ClaudeMDPath, CopilotInstructionsPath}

	if len(got) != len(want) {
		t.Fatalf("InstructionRelativePaths() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("InstructionRelativePaths()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
