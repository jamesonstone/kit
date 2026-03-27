package cli

import (
	"strings"
	"testing"
)

func TestPrepareAgentPromptWithoutSubagents(t *testing.T) {
	previous := subagents
	subagents = false
	t.Cleanup(func() {
		subagents = previous
	})

	prompt := "Please review the plan.\n"
	if got := prepareAgentPrompt(prompt); got != prompt {
		t.Fatalf("prepareAgentPrompt() = %q, want %q", got, prompt)
	}
}

func TestPrepareAgentPromptWithSubagents(t *testing.T) {
	previous := subagents
	subagents = true
	t.Cleanup(func() {
		subagents = previous
	})

	got := prepareAgentPrompt("Please review the plan.\n")
	checks := []string{
		"Please review the plan.",
		"## Subagent Orchestration",
		"drive to understanding first",
		"then drive task orchestration coordination",
		"use intelligent routing to identify the different areas of change or analysis",
		"delegate and dispatch to subagents where possible",
		"apply the same discovery-first discipline as kit dispatch",
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

func TestSubagentsFlagRegisteredOnRootCommand(t *testing.T) {
	if flag := rootCmd.PersistentFlags().Lookup("subagents"); flag == nil {
		t.Fatal("expected root command to register --subagents")
	}
}
