package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestMemoryCopilotInstructionsPreserveMutationRouting(t *testing.T) {
	for _, want := range []string{
		"Start with `docs/agents/README.md`",
		"Before Git, GitHub, or AWS mutations, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*`",
		"Repo-local Kit rules outrank generic defaults",
	} {
		if !strings.Contains(MemoryCopilotInstructionsMD, want) {
			t.Fatalf("expected V3 Copilot instructions to contain %q", want)
		}
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatalf("read checked-in Copilot instructions: %v", err)
	}
	if string(checkedIn) != MemoryCopilotInstructionsMD {
		t.Fatal("checked-in Copilot instructions are not aligned with the V3 generator")
	}
}

func TestInstructionTemplatesIncludeGitHubDeliveryHardGate(t *testing.T) {
	defaultChecks := []string{
		"## GitHub Delivery Hard Gate",
		"issue, branch, staging, commit, push, and PR actions are mutation boundaries",
		"Repo-local Kit rules outrank global GitHub/plugin defaults",
	}
	for name, content := range map[string]string{
		"AGENTS.md":                       AgentsMD,
		"CLAUDE.md":                       ClaudeMD,
		".github/copilot-instructions.md": CopilotInstructionsMD,
		"legacy AGENTS.md":                LegacyAgentsMD,
		"legacy CLAUDE.md":                LegacyClaudeMD,
		"legacy Copilot instructions":     LegacyCopilotInstructionsMD,
	} {
		for _, check := range defaultChecks {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(check)) {
				t.Fatalf("expected %s to contain %q", name, check)
			}
		}
	}

	guardrails := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionTOC),
		"docs/agents/GUARDRAILS.md",
	)
	for _, check := range []string{
		"A Kit-managed project is any repository containing `.kit.yaml`, `docs/CONSTITUTION.md`, or `docs/agents/README.md`",
		"Delivery Contract:",
		"Branch/status/staleness check:",
		"`<type>(<issue_number>): <gitmoji> <short title message>`",
		"`codex/*` branches",
		"global agent/plugin GitHub workflows are fallback tools only",
	} {
		if !strings.Contains(guardrails, check) {
			t.Fatalf("expected GUARDRAILS.md to contain %q", check)
		}
	}
}
