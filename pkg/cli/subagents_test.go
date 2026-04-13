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
		"use the repository instruction entrypoints as a map, not the full manual",
		"docs/agents/README.md",
		"read that feature's SPEC.md and the `## SKILLS` table first",
		"open each referenced `SKILL.md` and use those skills during execution",
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
		"drive to understanding first",
		"do RLM-style discovery first",
		"then drive task orchestration coordination",
		"default to subagents when the work spans multiple distinct areas",
		"do not turn broad discovery into parallel execution",
		"predict likely touched files or interfaces",
		"apply the same discovery-first discipline as kit dispatch",
		"git worktree add ~/worktrees/<repo>-<branch> <branch>",
		"keep all worktrees flat under `~/worktrees/`",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("expected augmented prompt to contain %q", check)
		}
	}

	if strings.Count(got, "## Subagent Orchestration") != 1 {
		t.Fatalf("expected one subagent section, got %q", got)
	}
}

func TestSingleAgentFlagRegisteredOnRootCommand(t *testing.T) {
	if flag := rootCmd.PersistentFlags().Lookup("single-agent"); flag == nil {
		t.Fatal("expected root command to register --single-agent")
	}
}

func TestLegacySubagentsFlagRemainsAvailable(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("subagents")
	if flag == nil {
		t.Fatal("expected root command to retain hidden --subagents compatibility flag")
	}
	if !flag.Hidden {
		t.Fatal("expected legacy --subagents flag to be hidden")
	}
}
