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
		"Before a substantial terminal completion or handoff response",
		"This structured contract does not apply to intermediate progress commentary",
		"Answer ordinary conversational requests naturally",
		"Direct questions, definitions, confirmations, rewrites, brief explanations, small read-only lookups",
		"must not receive status tokens, canonical section headings, synthetic None items, task profiles, or repository-memory reporting",
		"Use the structured contract when omitting it could hide a blocker",
		"Do not classify by word count, token count, elapsed time, or tool-call count",
		"When uncertain, prefer natural prose",
		"emit exactly `## What happened`, `## Deviations`, and `## Next steps` in that order",
		"**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**",
		"Do not add a separate status heading",
		"Use at most one nested evidence layer and state each fact once",
		"Use one `**None.**` bullet when there are no deviations",
		"make every required continuation copy-ready",
		"Use one `**None.**` bullet when no action remains",
		"PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE",
		"Do not use Markdown pipe tables, additional profile headings",
		"three canonical sections without duplication",
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
		for _, forbidden := range []string{
			"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
			"prioritized action list ordered Blocker, Incomplete, Next, Optional, then None",
			"After the action list, use left-aligned headings",
		} {
			if strings.Contains(content, forbidden) {
				t.Errorf("expected %s not to contain superseded contract %q", name, forbidden)
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
		if !strings.Contains(rlm, "Load `docs/references/rules/agent-completion-output.md` before a substantial terminal completion or handoff response") {
			t.Errorf("expected version %d RLM to route completion output", version)
		}
		references := fileContentByPath(files, "docs/references/README.md")
		if !strings.Contains(references, "Use `rules/agent-completion-output.md` before substantial terminal completion or handoff responses") {
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
		"Before a substantial terminal completion or handoff response",
		"answer ordinary conversational requests naturally without that structured envelope",
		"report only What happened, Deviations, and Next steps",
	} {
		if !strings.Contains(Constitution, check) {
			t.Errorf("expected Constitution template to contain %q", check)
		}
	}
}
