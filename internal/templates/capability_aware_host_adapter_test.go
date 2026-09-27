package templates

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/instructions"
)

func TestCapabilityAwareHostAdapterIsSharedAndProviderNeutral(t *testing.T) {
	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
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
			"host-specific bindings live in",
			"illustrative, never normative",
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
