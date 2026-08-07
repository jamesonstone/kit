package templates

import (
	"strings"
	"testing"
)

func TestMemoryAgentsContainsCodexBrowserPolicy(t *testing.T) {
	policy := codexBrowserPolicy("AGENTS")
	if policy == "" {
		t.Fatal("Codex AGENTS browser policy is empty")
	}
	if count := strings.Count(MemoryAgentsMD, policy); count != 1 {
		t.Fatalf("V3 AGENTS.md contains the browser policy %d times, want 1", count)
	}

	threadGate := strings.Index(MemoryAgentsMD, "## Codex Thread Initialization Hard Gate")
	browserPolicy := strings.Index(MemoryAgentsMD, "## Browser policy")
	purpose := strings.Index(MemoryAgentsMD, "## Purpose")
	if threadGate < 0 || browserPolicy <= threadGate || purpose <= browserPolicy {
		t.Fatalf("V3 AGENTS.md browser policy is not between the thread gate and purpose")
	}

	for _, want := range []string{
		"use Codex's built-in browser through `@Browser`",
		"Do not use `@Chrome`, control my active Chrome profile",
		"Playwright, Selenium, Cypress, or browser MCP tools",
		"If `@Browser` is unavailable, report the limitation",
		"terminate and verify all",
		"task-owned browser and automation processes before finishing",
	} {
		if !strings.Contains(policy, want) {
			t.Errorf("Codex browser policy is missing %q", want)
		}
	}
}

func TestCodexBrowserPolicyIsProviderSpecific(t *testing.T) {
	for name, content := range map[string]string{
		"V3 CLAUDE.md":                       MemoryClaudeMD,
		"V3 .github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	} {
		if strings.Contains(content, "## Browser policy") {
			t.Errorf("%s unexpectedly contains the Codex-only Browser policy", name)
		}
	}
}
