package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstructionTemplatesIncludeGitHubMergeHardGate(t *testing.T) {
	checks := []string{
		"Standing merge authority exists only when a human explicitly authorizes a bounded task, goal, or program",
		"never invent merge readiness",
		"pull-request-merge",
		"MERGE_READY",
		"one complete preflight snapshot",
		"do not rerun unchanged checks or poll repeatedly",
		"may bind later-created in-scope PRs and refreshed heads",
		"A changed in-scope head invalidates readiness, not standing authority",
		"A commit SHA or head OID identifies readiness evidence only",
		"Never request exact-head reauthorization",
		"already-authorized standard deployment and browser retry",
		"IAM, network, KMS, secrets, database-schema or data-loss changes",
		"Pause, hold, or revocation stops affected actions and dependents",
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
		for _, want := range []string{
			"Only exact current `MERGE_READY` nodes may merge",
			"one complete preflight snapshot",
			"may bind later-created in-scope PRs and refreshed heads",
			"A changed in-scope head invalidates readiness, not standing authority",
			"A commit SHA or head OID identifies readiness evidence only",
			"Never request exact-head reauthorization",
			"already-authorized standard deployment and browser retry",
			"IAM, network, KMS, secrets, database-schema or data-loss changes",
			"Pause, hold, or revocation stops affected actions and dependents",
		} {
			if !strings.Contains(string(content), want) {
				t.Errorf("checked-in %s is missing merge guidance %q", path, want)
			}
		}
	}
}
