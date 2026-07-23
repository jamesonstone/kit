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
		"load `docs/references/rules/constitution-curation.md`",
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

func TestMemoryGuardrailsPreserveAutonomousRecovery(t *testing.T) {
	generated := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/agents/GUARDRAILS.md",
	)
	for _, want := range []string{
		"Resolve all in-scope issues autonomously and continue until the goal is fully complete",
		"including authenticated `gh`",
		"Ask permission only before large-scale deletion or deleting sensitive files",
		"not as routine retry-permission requests",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("expected V3 guardrails to contain %q", want)
		}
	}
	if strings.Contains(generated, "Do not run `git add` or `git commit` without explicit approval") {
		t.Fatal("expected V3 guardrails to omit routine git approval requirement")
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents", "GUARDRAILS.md"))
	if err != nil {
		t.Fatalf("read checked-in guardrails: %v", err)
	}
	if string(checkedIn) != generated {
		t.Fatal("checked-in guardrails are not aligned with the V3 generator")
	}
}

func TestMemoryRepositoryInstructionsRouteConstitutionCuration(t *testing.T) {
	for relativePath, generated := range map[string]string{
		"AGENTS.md": MemoryAgentsMD,
		"CLAUDE.md": MemoryClaudeMD,
	} {
		if !strings.Contains(generated, "load `docs/references/rules/constitution-curation.md`") {
			t.Fatalf("expected V3 %s to route Constitution curation", relativePath)
		}
		checkedIn, err := os.ReadFile(filepath.Join("..", "..", relativePath))
		if err != nil {
			t.Fatalf("read checked-in %s: %v", relativePath, err)
		}
		if string(checkedIn) != generated {
			t.Fatalf("checked-in %s is not aligned with the V3 generator", relativePath)
		}
	}
}

func TestMemoryRepositoryInstructionsRouteApplicationArchitecture(t *testing.T) {
	routes := []string{
		"load `docs/references/rules/backend-service-architecture.md`",
		"load `docs/references/rules/frontend-application-architecture.md`",
		"responsibility boundaries rather than mandatory directory names",
	}
	for name, content := range map[string]string{
		"AGENTS.md":                       MemoryAgentsMD,
		"CLAUDE.md":                       MemoryClaudeMD,
		".github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	} {
		for _, route := range routes {
			if !strings.Contains(content, route) {
				t.Errorf("expected V3 %s to contain architecture route %q", name, route)
			}
		}
	}

	generatedRLM := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/agents/RLM.md",
	)
	for _, rulePath := range []string{
		"`docs/references/rules/backend-service-architecture.md`",
		"`docs/references/rules/frontend-application-architecture.md`",
	} {
		if !strings.Contains(generatedRLM, rulePath) {
			t.Errorf("expected generated RLM guidance to contain %q", rulePath)
		}
	}
	checkedInRLM, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents", "RLM.md"))
	if err != nil {
		t.Fatalf("read checked-in RLM guidance: %v", err)
	}
	if string(checkedInRLM) != generatedRLM {
		t.Fatal("checked-in RLM guidance is not aligned with the V3 generator")
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
