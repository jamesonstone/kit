package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV5RequiresDeletionSafety(t *testing.T) {
	content, err := AgentInstructions("v5")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v5\") error = %v", err)
	}

	for _, want := range []string{
		"# Deletion safety",
		"docs/references/rules/deletion-safety.md",
		"An unqualified delete means soft delete",
		"supported, authorized, and tested restore path",
		"Task-owned ephemeral scratch that never became authoritative state",
		"Treat ambiguous state as covered",
		"retention expiry",
		"separate privileged, auditable, server-enforced action",
		"resolve and present exact targets or a bounded",
		"materialized target IDs or an immutable",
		"why soft delete is",
		"After that outline, obtain a specific manual confirmation from the human",
		"Initial requests, general task or plan approval, automation",
		"compare the current target set or version with the confirmed",
		"One post-outline confirmation may satisfy multiple",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v5 instructions do not contain %q", want)
		}
	}

	before := strings.Index(content, "# Deletion safety")
	after := strings.Index(content, "# Execution\n")
	if before < 0 || after < 0 || before >= after {
		t.Fatal("v5 deletion safety must precede the implementation execution contract")
	}
	if CurrentAgentVersion != "v5" {
		t.Fatalf("CurrentAgentVersion = %q, want v5", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV5PreservesV4OutsideDeletionSection(t *testing.T) {
	v4, err := AgentInstructions("v4")
	if err != nil {
		t.Fatal(err)
	}
	v5, err := AgentInstructions("v5")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(v5, "\n# Deletion safety\n")
	end := strings.Index(v5, "\n# Execution\n")
	if start < 0 || end < 0 || start >= end {
		t.Fatal("v5 deletion section boundaries are missing")
	}
	withoutDeletion := v5[:start] + v5[end:]
	if withoutDeletion != v4 {
		t.Fatal("v5 changed content outside the additive deletion-safety section")
	}
}
