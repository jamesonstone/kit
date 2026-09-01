package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRequireInfrastructureChangeApproval(t *testing.T) {
	required := []string{
		"## Infrastructure Change Approval Hard Gate",
		"docs/references/rules/infrastructure-change-approval.md",
		"Kubernetes resources or cluster state",
		"does not alter cloud resources, Kubernetes objects",
		"Proceed autonomously when the graph contains only additive or rollback-preserving effects",
		"Deleting, destroying, or removing infrastructure always requires explicit confirmation",
		"During merge or release orchestration, do not execute infrastructure deletion",
		"isolate it as a separate task with its own exact post-outline authorization",
		"continue the rest of the task to completion in one pass",
		"collect all then-known changes into one follow-up outline",
		"Do not re-confirm actions already included in an approved batch",
		"do not require another prompt",
		"are not infrastructure-approval batches",
		"never authorize deletion",
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
		if !strings.Contains(rlm, "Load `docs/references/rules/infrastructure-change-approval.md` before planning or performing mutations to public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state") {
			t.Errorf("expected version %d RLM to route infrastructure approval", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "Use `rules/infrastructure-change-approval.md` before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state to classify effects, proceed autonomously for additive or rollback-preserving work, and require exact confirmation only for deletion or removal") {
			t.Errorf("expected version %d references index to route infrastructure approval", version)
		}
	}
}
