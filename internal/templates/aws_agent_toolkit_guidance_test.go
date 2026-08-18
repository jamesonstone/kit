package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRouteAWSAgentToolkitGuidance(t *testing.T) {
	required := []string{
		"docs/references/rules/aws-agent-toolkit-guidance.md",
		"current AWS skill, official documentation, AWS MCP Server or CLI fallback",
		"repo-local Kit gates remain authoritative",
		"Treat the verified account",
		"ARN, and Region as authoritative",
		"verified configured profile and Region explicitly",
		"never use default, another discovered profile, or ambient credentials",
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
		for _, check := range required {
			if !strings.Contains(content, check) {
				t.Errorf("expected %s to contain %q", name, check)
			}
		}
		routeIndex := strings.Index(content, "Before AWS-dependent work")
		contextIndex := strings.Index(strings.ToLower(content), "enabled aws context")
		if contextIndex >= 0 && (routeIndex < 0 || routeIndex > contextIndex) {
			t.Errorf("expected %s to route AWS guidance before the optional Kit AWS context", name)
		}
	}
}

func TestMemoryInstructionsPreserveAWSIdentityGate(t *testing.T) {
	for _, check := range []string{
		"## AWS Context Hard Gate",
		"If `.kit.yaml` defines an enabled AWS context",
		"run `kit aws verify` before the first AWS-dependent command",
		"Treat the verified account, ARN, and Region as authoritative",
		"Use the verified configured profile and Region explicitly",
		"After verification, never use default, another discovered profile, or ambient credentials",
	} {
		for name, content := range map[string]string{
			"AGENTS.md":               MemoryAgentsMD,
			"CLAUDE.md":               MemoryClaudeMD,
			"copilot-instructions.md": MemoryCopilotInstructionsMD,
		} {
			if !strings.Contains(content, check) {
				t.Errorf("expected V3 %s to contain %q", name, check)
			}
		}
	}
}

func TestInstructionSupportRoutesAWSAgentToolkitGuidance(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, "Load `docs/references/rules/aws-agent-toolkit-guidance.md` before AWS-dependent work") {
			t.Errorf("expected version %d RLM to route AWS Agent Toolkit guidance", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "Use `rules/aws-agent-toolkit-guidance.md` before AWS-dependent work") {
			t.Errorf("expected version %d references index to route AWS Agent Toolkit guidance", version)
		}
		guardrails := fileContentByPath(files, "docs/agents/GUARDRAILS.md")
		for _, check := range []string{
			"Treat the returned account ID, ARN, and Region as authoritative",
			"Use the verified configured profile and Region explicitly",
			"Never fall back to default, another discovered profile, or ambient credentials",
		} {
			if !strings.Contains(guardrails, check) {
				t.Errorf("expected version %d guardrails to contain %q", version, check)
			}
		}
	}
}
