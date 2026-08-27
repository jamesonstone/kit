package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesIncludeGitHubDeliveryHardGate(t *testing.T) {
	defaultChecks := []string{
		"## GitHub Delivery Hard Gate",
		"Issue, branch, staging, commit, push, PR, and merge actions are distinct mutation boundaries",
		"Repo-local Kit rules outrank global GitHub/plugin defaults",
	}
	for name, content := range map[string]string{
		"AGENTS.md":                       AgentsMD,
		"CLAUDE.md":                       ClaudeMD,
		".github/copilot-instructions.md": CopilotInstructionsMD,
		"legacy AGENTS.md":                LegacyAgentsMD,
		"legacy CLAUDE.md":                LegacyClaudeMD,
		"legacy Copilot instructions":     LegacyCopilotInstructionsMD,
	} {
		for _, check := range defaultChecks {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(check)) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}

	guardrails := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionTOC),
		"docs/agents/GUARDRAILS.md",
	)
	for _, check := range []string{
		"A Kit-managed project is any repository containing `.kit.yaml`, `docs/CONSTITUTION.md`, or `docs/agents/README.md`",
		"Delivery Contract:",
		"Branch/status/staleness check:",
		"`<type>(<issue_number>): <gitmoji> <short title message>`",
		"`codex/*` branches",
		"global agent/plugin GitHub workflows are fallback tools only",
	} {
		if !strings.Contains(guardrails, check) {
			t.Fatalf("expected GUARDRAILS.md to contain %q", check)
		}
	}
}

func TestInstructionTemplatesDefaultToNewWorkLaneBeforeMutation(t *testing.T) {
	checks := []string{
		"## Work Lane Mutation Hard Gate",
		"docs/agents/GUARDRAILS.md",
		"work-lane-gating",
		"Default to a new worklane without asking",
		"one human-assigned GitHub issue",
		"exact `GH-<issue-number>` branch",
		"canonical non-primary worktree",
		"ready pull-request plan",
		"Continue an existing lane only when the user explicitly directs",
		"Never offer or ask the user to choose between lanes",
		"Ask only",
		"implementation intent",
		"materially ambiguous",
		"Pull-Request Landing Plan",
		"repository file or delivery mutation",
		"issue, branch, staging, commit, push, worktree, and pull-request mutations",
		"before every mutation",
		"primary/root checkout as read-only",
		"Do not stage, commit, push, stash, reset, clean, discard, or silently transfer",
	}
	orderedChecks := []string{
		"docs/agents/GUARDRAILS.md",
		"work-lane-gating",
		"read-only safety recon",
		"Default to a new worklane without asking",
		"Continue an existing lane only when the user explicitly directs",
		"Pull-Request Landing Plan",
	}
	for name, content := range map[string]string{
		"V1 AGENTS.md":            LegacyAgentsMD,
		"V1 CLAUDE.md":            LegacyClaudeMD,
		"V1 Copilot instructions": LegacyCopilotInstructionsMD,
		"V2 AGENTS.md":            AgentsMD,
		"V2 CLAUDE.md":            ClaudeMD,
		"V2 Copilot instructions": CopilotInstructionsMD,
		"V3 AGENTS.md":            MemoryAgentsMD,
		"V3 CLAUDE.md":            MemoryClaudeMD,
		"V3 Copilot instructions": MemoryCopilotInstructionsMD,
	} {
		normalizedContent := strings.Join(strings.Fields(content), " ")
		for _, check := range checks {
			if !strings.Contains(normalizedContent, check) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
		previousIndex := -1
		for _, check := range orderedChecks {
			index := strings.Index(normalizedContent, check)
			if index <= previousIndex {
				t.Fatalf("expected %s to contain %q after the preceding gate step", name, check)
			}
			previousIndex = index
		}
	}
	for name, content := range map[string]string{
		"V1 AGENTS.md":            LegacyAgentsMD,
		"V1 CLAUDE.md":            LegacyClaudeMD,
		"V1 Copilot instructions": LegacyCopilotInstructionsMD,
		"V2 AGENTS.md":            AgentsMD,
		"V2 CLAUDE.md":            ClaudeMD,
		"V2 Copilot instructions": CopilotInstructionsMD,
		"V3 AGENTS.md":            MemoryAgentsMD,
		"V3 CLAUDE.md":            MemoryClaudeMD,
		"V3 Copilot instructions": MemoryCopilotInstructionsMD,
	} {
		for _, forbidden := range []string{
			"Before I make any repository changes, should I create a new GitHub issue",
			"`c` means continue existing",
			"Wait for the explicit choice",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("expected %s to omit superseded lane-choice guidance %q", name, forbidden)
			}
		}
	}
}
