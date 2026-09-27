package templates

import (
	"strings"
	"testing"
)

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
