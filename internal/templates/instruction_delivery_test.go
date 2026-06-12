package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestInstructionTemplatesIncludeGitHubDeliveryHardGate(t *testing.T) {
	defaultChecks := []string{
		"## GitHub Delivery Hard Gate",
		"issue, branch, staging, commit, push, and PR actions are mutation boundaries",
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
		"`codex/*` branches",
		"global agent/plugin GitHub workflows are fallback tools only",
	} {
		if !strings.Contains(guardrails, check) {
			t.Fatalf("expected GUARDRAILS.md to contain %q", check)
		}
	}
}
