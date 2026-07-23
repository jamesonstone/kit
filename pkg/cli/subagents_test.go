package cli

import (
	"strings"
	"testing"
)

func TestPrepareAgentPromptWithoutSubagents(t *testing.T) {
	previous := singleAgent
	singleAgent = true
	t.Cleanup(func() {
		singleAgent = previous
	})

	prompt := "Please review the plan.\n"
	got := prepareAgentPrompt(prompt)
	checks := []string{
		"Please review the plan.",
		"## Skills",
		"repository instruction entrypoints as routing maps",
		"docs/agents/README.md",
		"canonical front matter `skills`",
		"open every selected or explicitly provided `SKILL.md`",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("expected augmented prompt to contain %q", check)
		}
	}

	if strings.Contains(got, "## Subagent Orchestration") {
		t.Fatalf("expected prompt without subagents not to contain subagent guidance")
	}
}

func TestPrepareAgentPromptWithSubagentsByDefault(t *testing.T) {
	previous := singleAgent
	singleAgent = false
	t.Cleanup(func() {
		singleAgent = previous
	})

	got := prepareAgentPrompt("Please review the plan.\n")
	checks := []string{
		"Please review the plan.",
		"## Skills",
		"## Subagent Orchestration",
		"agent-team-orchestration.md",
		"The supervisor owns scope",
		"Agent Team Plan",
		"low-overlap areas",
		"In normal operation, run at most 3 independent lanes",
		"fourth lane requires explicit exceptional authorization from the supervisor",
		"never exceed 4 lanes",
		"read-only verification agent",
		"single supervisor lane; no specialist or verification agents spawned",
		"supervisor-prepared, explicitly assigned worktree",
		"may not create, switch, move, or remove worktrees",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("expected augmented prompt to contain %q", check)
		}
	}

	if strings.Count(got, "## Subagent Orchestration") != 1 {
		t.Fatalf("expected one subagent section, got %q", got)
	}
	if strings.Contains(got, "at most 3 independent lanes (hard ceiling 4)") {
		t.Fatalf("normal concurrency guidance should not imply an automatic fourth lane, got %q", got)
	}
}

func TestSingleAgentFlagRegisteredOnRootCommand(t *testing.T) {
	if flag := rootCmd.PersistentFlags().Lookup("single-agent"); flag == nil {
		t.Fatal("expected root command to register --single-agent")
	}
}

func TestLegacySubagentsFlagIsRemoved(t *testing.T) {
	if flag := rootCmd.PersistentFlags().Lookup("subagents"); flag != nil {
		t.Fatalf("expected legacy --subagents flag to be removed, got %#v", flag)
	}
}
