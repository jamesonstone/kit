package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestInstructionTemplatesRequireInfrastructureChangeApproval(t *testing.T) {
	required := []string{
		"## Infrastructure Change Approval Hard Gate",
		"docs/references/rules/infrastructure-change-approval.md",
		"present one consolidated outline of the target context",
		"obtain explicit user confirmation",
		"sufficiently detailed initial request may satisfy the gate",
		"execute the exact approved batch to completion",
		"obtain renewed confirmation",
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
		"V3 GUARDRAILS.md": fileContentByPath(
			InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
			"docs/agents/GUARDRAILS.md",
		),
	} {
		for _, check := range required {
			if !strings.Contains(content, check) {
				t.Errorf("expected %s to contain %q", name, check)
			}
		}
	}
}

func TestInstructionSupportRoutesInfrastructureChangeApproval(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, "Load `docs/references/rules/infrastructure-change-approval.md` before planning or performing public-cloud or infrastructure-as-code mutations") {
			t.Errorf("expected version %d RLM to route infrastructure approval", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "rules/infrastructure-change-approval.md") {
			t.Errorf("expected version %d references index to route infrastructure approval", version)
		}
	}
}
