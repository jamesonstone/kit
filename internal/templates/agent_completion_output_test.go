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
		"Write each response, including terminal completions and handoffs, in the shape its content calls for",
		"Match length to consequence rather than to effort spent",
		"conveys what the user now has, what remains unfinished and why",
		"anything blocking completion and what would clear it",
		"the exact command or prompt when there is one",
		"Say plainly whether the work is finished, partly finished, blocked, or failed",
		"Keep blockers and unfinished scope as prominent as the successes",
		"Report each check as observed",
		"PENDING, UNKNOWN, SKIPPED, and NOT_APPLICABLE are preserved verbatim",
		"Report a check as passing only when it ran and passed",
		"When something could not be validated, say so and say why",
		"Distinguish a verified fact from an inference and from a hypothesis",
		"recoverable from the response, with the identifiers a reader needs to find or undo it",
		"an account of where things stand, not an index of everything checked",
		"smallest evidence set that proves each terminal node",
		"Satisfy them on content; a heading alone satisfies none of them",
	}
	for name, content := range map[string]string{
		"V1 AGENTS.md":            LegacyAgentsMD,
		"V1 CLAUDE.md":            LegacyClaudeMD,
		"V1 Copilot instructions": LegacyCopilotInstructionsMD,
		"V2 AGENTS.md":            AgentsMD,
		"V2 CLAUDE.md":            ClaudeMD,
		"V2 Copilot instructions": CopilotInstructionsMD,
	} {
		for _, check := range required {
			if !strings.Contains(content, check) {
				t.Errorf("expected %s to contain %q", name, check)
			}
		}
		for _, forbidden := range []string{
			"## What happened",
			"## Deviations",
			"## Next steps",
			"**Status: PASS|PARTIAL|BLOCKED|FAIL",
			"`**None.**`",
			"retired envelope",
			"fixed template applied to every task",
			"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
			"prioritized action list ordered Blocker, Incomplete, Next, Optional, then None",
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

func TestConstitutionTemplateRequiresAgentCompletionOutput(t *testing.T) {
	for _, check := range []string{
		"docs/references/rules/agent-completion-output.md",
		"Before a substantial terminal completion or handoff response",
		"which requires no response format",
		"never leave the reader wrong about a blocker, incomplete scope, a required action, or a failing or unobserved check",
	} {
		if !strings.Contains(LegacyConstitutionBaselineSection, check) {
			t.Errorf("expected Constitution template to contain %q", check)
		}
	}
}
