package templates

import (
	"strings"
	"testing"
)

func TestMemoryAgentsOmitsCodexThreadInitializationGate(t *testing.T) {
	const prefix = `# AGENTS

## Browser policy`
	if !strings.HasPrefix(MemoryAgentsMD, prefix) {
		t.Fatalf("V3 AGENTS.md does not start with the Codex browser policy:\n%s", MemoryAgentsMD)
	}
	for _, forbidden := range []string{
		"## Codex Thread Initialization Hard Gate",
		"set_thread_title",
		"set_thread_pinned",
		"Thread initialization: rename <status>; pin <status>.",
	} {
		if strings.Contains(MemoryAgentsMD, forbidden) {
			t.Fatalf("V3 AGENTS.md still contains retired thread init %q", forbidden)
		}
	}
}

func TestCodexThreadInitializationGateIsProviderSpecific(t *testing.T) {
	for name, content := range map[string]string{
		"V3 AGENTS.md":                       MemoryAgentsMD,
		"V3 CLAUDE.md":                       MemoryClaudeMD,
		"V3 .github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	} {
		for _, forbidden := range []string{
			"Codex Thread Initialization Hard Gate",
			"set_thread_title",
			"set_thread_pinned",
		} {
			if strings.Contains(content, forbidden) {
				t.Errorf("%s unexpectedly contains Codex-only guidance %q", name, forbidden)
			}
		}
	}
}
