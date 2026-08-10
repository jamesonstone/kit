package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestInstructionTemplatesRequireCrossRepositoryProgramCoordination(t *testing.T) {
	required := []string{
		"## Cross-Repository Program Coordination Gate",
		"docs/references/rules/cross-repository-program-coordination.md",
		"spans multiple repositories",
		"dependent deliverables, staged deployment or activation",
		"docs/programs/<program>/PROGRAM.md",
		"participant repositories remain authoritative",
		"Dispatch only the reconciled ready frontier",
		"reconcile recorded claims against live repositories, GitHub, runtime, and validation evidence",
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
	}
}

func TestInstructionSupportRoutesCrossRepositoryProgramCoordination(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, "Load `docs/references/rules/cross-repository-program-coordination.md` before implementing or resuming an accepted plan that spans multiple repositories with dependent deliverables, staged deployment or activation, or expected agent or session handoff") {
			t.Errorf("expected version %d RLM to route program coordination", version)
		}
		tooling := fileContentByPath(files, "docs/agents/TOOLING.md")
		if !strings.Contains(tooling, "dispatch only the canonical program ledger's reconciled ready frontier") {
			t.Errorf("expected version %d tooling to route the program ready frontier", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "Use `rules/cross-repository-program-coordination.md` before implementing or resuming accepted plans that span multiple repositories with dependent deliverables, staged deployment or activation, or expected handoff") {
			t.Errorf("expected version %d references index to route program coordination", version)
		}
	}
}
