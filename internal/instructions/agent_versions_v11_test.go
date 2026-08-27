package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV11DefaultsToNewWorklane(t *testing.T) {
	content, err := AgentInstructions("v11")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v11\") error = %v", err)
	}
	for _, want := range []string{
		"Default to a new worklane without asking",
		"one human-assigned",
		"GitHub issue, exact `GH-<issue-number>` branch",
		"exact `GH-<issue-number>` branch",
		"canonical non-primary",
		"ready pull-request plan",
		"Continue an existing lane only when the user explicitly directs",
		"Never offer or ask the user to choose between lanes",
		"Do not infer existing-lane continuation",
		"Ask only to clarify implementation intent or a user-named target",
		"ask for a new-versus-existing lane preference",
		"Pull-Request Landing Plan",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v11 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"Before I make any repository changes, should I create a new GitHub issue",
		"`c` means continue existing",
		"Wait for the user's explicit choice",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v11 instructions contain superseded lane-choice guidance %q", forbidden)
		}
	}
	if CurrentAgentVersion != "v11" {
		t.Fatalf("CurrentAgentVersion = %q, want v11", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV11PreservesV10OutsideMutationGate(t *testing.T) {
	v10, err := AgentInstructions("v10")
	if err != nil {
		t.Fatal(err)
	}
	v11, err := AgentInstructions("v11")
	if err != nil {
		t.Fatal(err)
	}
	const contextStart = "\n# Repository context and mutation gate\n"
	const mergeStart = "\n## Merge authorization\n"
	const executionStart = "\n# Execution environment\n"
	const deletionStart = "\n# Deletion safety\n"
	v10Context := strings.Index(v10, contextStart)
	v11Context := strings.Index(v11, contextStart)
	v10Merge := strings.Index(v10, mergeStart)
	v11Merge := strings.Index(v11, mergeStart)
	v10Execution := strings.Index(v10, executionStart)
	v11Execution := strings.Index(v11, executionStart)
	v10Deletion := strings.Index(v10, deletionStart)
	v11Deletion := strings.Index(v11, deletionStart)
	if v10Context < 0 || v11Context < 0 || v10Merge < 0 || v11Merge < 0 ||
		v10Execution < 0 || v11Execution < 0 || v10Deletion < 0 || v11Deletion < 0 {
		t.Fatal("instruction section boundaries are missing")
	}
	if v10[:v10Context] != v11[:v11Context] ||
		v10[v10Merge:v10Execution] != v11[v11Merge:v11Execution] ||
		v10[v10Deletion:] != v11[v11Deletion:] {
		t.Fatal("v11 changed content outside lane-routing sections")
	}
}
