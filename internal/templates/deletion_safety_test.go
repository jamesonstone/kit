package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRequireDeletionSafety(t *testing.T) {
	required := []string{
		"## Deletion Safety Hard Gate",
		"docs/references/rules/deletion-safety.md",
		"An unqualified delete means soft delete",
		"Task-owned ephemeral scratch that never became authoritative state",
		"retention expiry",
		"separate privileged, auditable, server-enforced action",
		"bounded selector first resolved to the exact current target set",
		"materialized target IDs or an immutable snapshot/version token",
		"obtain a specific manual confirmation from the human",
		"Initial requests, general task or plan approval, automation",
		"compare the current target set or version with the confirmed snapshot",
		"One post-outline confirmation may satisfy multiple deletion gates",
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

func TestInstructionSupportRoutesDeletionSafety(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, "Load `docs/references/rules/deletion-safety.md` before designing deletion behavior or deleting persistent project, user, business, or external-system state") {
			t.Errorf("expected version %d RLM to route deletion safety", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "Use `rules/deletion-safety.md` before designing deletion behavior or deleting persistent project, user, business, or external-system state") {
			t.Errorf("expected version %d references index to route deletion safety", version)
		}
		testingReference := fileContentByPath(files, "docs/references/testing.md")
		if !strings.Contains(testingReference, "Follow `rules/deletion-safety.md` for cleanup") {
			t.Errorf("expected version %d testing reference to route deletion cleanup", version)
		}
	}
}

func TestContextWorkflowsRequireDeletionSafety(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if !strings.Contains(artifact.Content, "  - slug: deletion-safety\n    required: true") {
			t.Errorf("workflow %s does not require deletion-safety", artifact.Slug)
		}
	}
}

func TestConstitutionTemplateRequiresDeletionSafety(t *testing.T) {
	for _, check := range []string{
		"docs/references/rules/deletion-safety.md",
		"Default unqualified deletion to a recoverable soft delete",
		"post-outline specific manual confirmation",
	} {
		if !strings.Contains(Constitution, check) {
			t.Errorf("expected Constitution template to contain %q", check)
		}
	}
}
