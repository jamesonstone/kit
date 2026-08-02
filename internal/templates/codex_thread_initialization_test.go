package templates

import (
	"strings"
	"testing"
)

func TestMemoryAgentsStartsWithCodexThreadInitializationGate(t *testing.T) {
	const prefix = `# AGENTS

## Codex Thread Initialization Hard Gate`
	if !strings.HasPrefix(MemoryAgentsMD, prefix) {
		t.Fatalf("V3 AGENTS.md does not start with the Codex thread gate:\n%s", MemoryAgentsMD)
	}

	ordered := []string{
		"before the first commentary message",
		"First, call the available thread-title operation (`set_thread_title` when available)",
		"Second, call the available thread-pin operation (`set_thread_pinned` when available)",
		"Thread initialization: rename <status>; pin <status>.",
		"## Purpose",
	}
	previous := -1
	for _, snippet := range ordered {
		index := strings.Index(MemoryAgentsMD, snippet)
		if index < 0 {
			t.Fatalf("V3 AGENTS.md is missing %q", snippet)
		}
		if index <= previous {
			t.Fatalf("V3 AGENTS.md has %q out of order", snippet)
		}
		previous = index
	}
	if count := strings.Count(MemoryAgentsMD, "## Codex Thread Initialization Hard Gate"); count != 1 {
		t.Fatalf("V3 AGENTS.md contains the gate %d times, want 1", count)
	}
}

func TestCodexThreadInitializationGateIsProviderSpecific(t *testing.T) {
	for name, content := range map[string]string{
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
