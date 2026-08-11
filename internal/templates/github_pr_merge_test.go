package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstructionTemplatesIncludeGitHubMergeHardGate(t *testing.T) {
	checks := []string{
		"Merge is a distinct mutation boundary",
		"never imply merge consent",
		"pull-request-merge",
		"MERGE_READY",
		"Never bypass protection",
	}
	for name, content := range map[string]string{
		"AGENTS.md":                       AgentsMD,
		"CLAUDE.md":                       ClaudeMD,
		".github/copilot-instructions.md": CopilotInstructionsMD,
		"legacy AGENTS.md":                LegacyAgentsMD,
		"legacy CLAUDE.md":                LegacyClaudeMD,
		"legacy Copilot instructions":     LegacyCopilotInstructionsMD,
	} {
		for _, check := range checks {
			if !strings.Contains(content, check) {
				t.Errorf("%s missing merge gate %q", name, check)
			}
		}
	}
}

func TestCheckedInInstructionsIncludeGitHubMergeHardGate(t *testing.T) {
	for _, path := range []string{
		"AGENTS.md",
		"CLAUDE.md",
		".github/copilot-instructions.md",
		"docs/agents/GUARDRAILS.md",
	} {
		content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(content), "Only exact current `MERGE_READY` nodes may merge") {
			t.Errorf("checked-in %s is missing merge readiness guidance", path)
		}
	}
}
