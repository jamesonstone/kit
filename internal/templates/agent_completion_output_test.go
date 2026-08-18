package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestInstructionTemplatesRequireAgentCompletionOutput(t *testing.T) {
	required := []string{
		"## Agent Completion Output Contract",
		"docs/references/rules/agent-completion-output.md",
		"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
		"Type | Action required | Why | Continue with",
		"Order rows Blocker, Incomplete, Next, Optional, then None",
		"every PASS includes a None row",
		"Make required follow-ups copy-ready",
		"PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE",
		"After the action table, use left-aligned headings and CommonMark list or key/value blocks",
		"Do not use another Markdown pipe table unless a higher-priority schema requires it",
		"implementation, research, diagnosis, planning, validation, review, operations, coordination, or fallback",
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

func TestInstructionSupportRoutesAgentCompletionOutput(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		files := InstructionSupportFiles(version)
		rlm := fileContentByPath(files, "docs/agents/RLM.md")
		if !strings.Contains(rlm, "Load `docs/references/rules/agent-completion-output.md` before a terminal task completion or handoff response") {
			t.Errorf("expected version %d RLM to route completion output", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "Use `rules/agent-completion-output.md` before every terminal task completion or handoff response") {
			t.Errorf("expected version %d references index to route completion output", version)
		}
	}
}

func TestContextWorkflowsRequireAgentCompletionOutput(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if !strings.Contains(artifact.Content, "  - slug: agent-completion-output\n    required: true") {
			t.Errorf("workflow %s does not require agent-completion-output", artifact.Slug)
		}
	}
}

func TestConstitutionTemplateRequiresAgentCompletionOutput(t *testing.T) {
	for _, check := range []string{
		"docs/references/rules/agent-completion-output.md",
		"literal status, immediate action table, and primary task profile",
	} {
		if !strings.Contains(Constitution, check) {
			t.Errorf("expected Constitution template to contain %q", check)
		}
	}
}
