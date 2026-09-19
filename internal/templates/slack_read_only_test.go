package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRequireSlackReadOnly(t *testing.T) {
	required := []string{
		"## Slack: Read-Only by Default, Explicit Approval Required to Send",
		"docs/references/rules/slack-read-only.md",
		"Treat all Slack access as **read-only by default**",
		"Drafting a Slack message is not authorization to send it",
		`"send it," "send this," or "yes, send that message."`,
		"Approval is **single-use and message-specific**",
		"When uncertain whether the human authorized a Slack write, **do not perform it. Ask.**",
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

func TestInstructionSupportRoutesSlackReadOnly(t *testing.T) {
	rlmRoute := "Load `docs/references/rules/slack-read-only.md` before any Slack write, and when a Slack link, channel, thread, or search is part of the task"
	indexRoute := "Use `rules/slack-read-only.md` when Slack is in scope"
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, rlmRoute) {
			t.Errorf("expected version %d RLM to route slack-read-only", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, indexRoute) {
			t.Errorf("expected version %d references index to route slack-read-only", version)
		}
	}
}

func TestImplementationDeliverySelectsSlackReadOnlyOptionally(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if artifact.Slug != "implementation-delivery" {
			continue
		}
		if !strings.Contains(artifact.Content, "slug: slack-read-only\n    required: false") {
			t.Fatal("implementation-delivery does not select slack-read-only as optional evidence")
		}
		return
	}
	t.Fatal("embedded implementation-delivery workflow not found")
}

func TestConstitutionTemplateRoutesSlackReadOnly(t *testing.T) {
	for _, check := range []string{
		"docs/references/rules/slack-read-only.md",
		"Treat Slack as read-only by default",
		"explicit, message-specific human approval",
	} {
		if !strings.Contains(Constitution, check) {
			t.Errorf("expected Constitution template to contain %q", check)
		}
	}
}
