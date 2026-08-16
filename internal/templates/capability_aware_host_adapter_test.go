package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/instructions"
)

func TestCapabilityAwareHostAdapterIsSharedAndProviderNeutral(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		tooling := fileContentByPath(
			InstructionSupportFiles(version),
			"docs/agents/TOOLING.md",
		)
		for _, want := range []string{
			"## Capability-Aware Host Adapter",
			"`architect`",
			"`orchestrator`",
			"`mapper`",
			"`specialist`",
			"`precision`",
			"`verifier`",
			"child-launch, per-child model and effort, same-agent follow-up",
			"Record confirmed absence as unavailable and unexposed controls as `unknown`",
			"Let the host govern concurrency",
			"do not invent a numeric cap",
			"equal-or-stronger eligible configuration",
			"exact user model or effort pin remains blocked",
			"If capacity changes or a spawn fails",
			"requested and effective profiles plus model and effort",
			"confirmed or unknown parallelism",
			"Unknown or single-agent host",
			"Never report a role prompt, task list, handoff, or manually opened conversation as a child",
			"https://learn.chatgpt.com/docs/agent-configuration/subagents",
			"https://code.claude.com/docs/en/sub-agents",
			"https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/invoke-custom-agents",
			"https://docs.warp.dev/platform/orchestration/",
			"https://docs.warp.dev/agents/capabilities/rules/",
			"not as pinned capability promises",
		} {
			if !strings.Contains(tooling, want) {
				t.Errorf("version %d TOOLING.md missing %q", version, want)
			}
		}
		for _, host := range []string{"Codex", "Claude Code", "GitHub Copilot", "Warp/Oz"} {
			if !strings.Contains(tooling, "| "+host+" |") {
				t.Errorf("version %d TOOLING.md missing illustrative %s mapping", version, host)
			}
		}
		for _, obsolete := range []string{
			"Default to at most 3 concurrent lanes",
			"never exceed 4",
			"hard ceiling 4",
		} {
			if strings.Contains(tooling, obsolete) {
				t.Errorf("version %d TOOLING.md retains obsolete policy %q", version, obsolete)
			}
		}
	}
}

func TestConditionalCodexBindingAppearsOnlyOnceInMemoryAgents(t *testing.T) {
	const heading = "## Conditional Codex Subagent Binding"
	if count := strings.Count(MemoryAgentsMD, heading); count != 1 {
		t.Fatalf("V3 AGENTS.md contains Codex binding %d times, want 1", count)
	}
	for _, want := range []string{
		"only when the active coding host is Codex",
		"Warp/Oz and every other host",
		"`list_agents`",
		"`spawn_agent`",
		"`followup_task`",
		"`wait_agent`",
		"`model`",
		"`reasoning_effort`",
		"children must not spawn descendants",
	} {
		if !strings.Contains(MemoryAgentsMD, want) {
			t.Errorf("V3 AGENTS.md Codex binding missing %q", want)
		}
	}

	for name, content := range map[string]string{
		"V2 AGENTS.md":                       AgentsMD,
		"V3 CLAUDE.md":                       MemoryClaudeMD,
		"V3 .github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	} {
		if strings.Contains(content, heading) || strings.Contains(content, "`spawn_agent`") {
			t.Errorf("%s unexpectedly contains the Codex-only binding", name)
		}
	}
}

func TestCapabilityAdapterKeepsDefaultInstructionTargets(t *testing.T) {
	got := instructions.InstructionRelativePaths(config.Default())
	want := []string{
		instructions.AgentsMDPath,
		instructions.ClaudeMDPath,
		instructions.CopilotInstructionsPath,
	}
	if len(got) != len(want) {
		t.Fatalf("default instruction targets = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("default instruction target %d = %q, want %q", i, got[i], want[i])
		}
	}
	for _, path := range got {
		if strings.Contains(strings.ToLower(path), "warp") {
			t.Fatalf("default instruction targets unexpectedly include %q", path)
		}
	}
}

func TestCheckedInCapabilityAdapterMatchesGeneratedArtifacts(t *testing.T) {
	generated := map[string]string{
		"AGENTS.md": MemoryAgentsMD,
		"CLAUDE.md": MemoryClaudeMD,
		"docs/agents/TOOLING.md": fileContentByPath(
			InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
			"docs/agents/TOOLING.md",
		),
		".github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	}
	for relativePath, want := range generated {
		got, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(relativePath)))
		if err != nil {
			t.Fatalf("read checked-in %s: %v", relativePath, err)
		}
		if string(got) != want {
			t.Errorf("checked-in %s is not aligned with the V3 generator", relativePath)
		}
	}
}
