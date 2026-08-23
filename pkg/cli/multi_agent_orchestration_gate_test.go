package cli

import (
	"strings"
	"testing"
)

func TestMultiAgentOrchestrationGatePresentAndOrderedBeforeWorkLaneGate(t *testing.T) {
	const gateHeading = "## Multi-Agent Orchestration Evaluation Hard Gate"
	const workLaneHeading = "## Work Lane Mutation Hard Gate"

	for _, path := range []string{
		"CLAUDE.md",
		"AGENTS.md",
		".github/copilot-instructions.md",
		"docs/agents/GUARDRAILS.md",
	} {
		content := readRepositoryFile(t, path)

		gateIndex := strings.Index(content, gateHeading)
		if gateIndex < 0 {
			t.Fatalf("%s missing %q", path, gateHeading)
		}

		workLaneIndex := strings.Index(content, workLaneHeading)
		if workLaneIndex < 0 {
			t.Fatalf("%s missing %q", path, workLaneHeading)
		}

		if gateIndex >= workLaneIndex {
			t.Errorf("%s: %q must precede %q", path, gateHeading, workLaneHeading)
		}
	}
}

func TestMultiAgentOrchestrationGateContent(t *testing.T) {
	instructionChecks := []string{
		"load `docs/references/rules/agent-team-orchestration.md`",
		"single-lane, because <reason>",
		"never skip",
		"skip the evaluation silently",
	}

	for _, path := range []string{
		"CLAUDE.md",
		"AGENTS.md",
		".github/copilot-instructions.md",
		"docs/agents/GUARDRAILS.md",
	} {
		content := strings.ToLower(readRepositoryFile(t, path))
		checks := instructionChecks
		if path == "docs/agents/GUARDRAILS.md" {
			checks = []string{
				"load `docs/references/rules/agent-team-orchestration.md`",
				"single-lane, because <reason>",
				"this gate is mandatory",
				"single mechanical edit, a direct question, or read-only research",
				"never treat this evaluation as permission to force parallel execution",
			}
		}
		for _, check := range checks {
			if !strings.Contains(content, strings.ToLower(check)) {
				t.Errorf("%s missing gate content %q", path, check)
			}
		}
	}
}

func TestMultiAgentOrchestrationGateRLMMandatoryTOOLINGUnchanged(t *testing.T) {
	rlm := readRepositoryFile(t, "docs/agents/RLM.md")
	if !strings.Contains(rlm, "mandatory first-pass evaluation before finalizing any native implementation plan") {
		t.Error("RLM.md agent-team-orchestration pointer is not a mandatory first-pass evaluation")
	}
	if strings.Contains(rlm, "do not load it for trivial single-lane tasks") {
		t.Error("RLM.md still contains the old conditional-only pointer wording")
	}

	tooling := readRepositoryFile(t, "docs/agents/TOOLING.md")
	if !strings.Contains(tooling, "Load `docs/references/rules/agent-team-orchestration.md` when dispatch, direct subagent execution, or read-only verification topology affects the task") {
		t.Error("TOOLING.md dispatch-time agent-team-orchestration pointer must remain conditional and unchanged")
	}
}

func TestMultiAgentOrchestrationRequiredInImplementationDelivery(t *testing.T) {
	for _, path := range []string{
		"docs/references/workflows/implementation-delivery.md",
		"internal/templates/context_workflows/implementation-delivery.md",
	} {
		content := readRepositoryFile(t, path)
		if !strings.Contains(content, "slug: agent-team-orchestration\n    required: true") {
			t.Errorf("%s does not mark agent-team-orchestration required", path)
		}
	}
}
