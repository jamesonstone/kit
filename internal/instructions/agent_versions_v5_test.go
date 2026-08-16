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
		"selector first resolved to the exact current target set",
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

func TestAgentInstructionsV5AcceptsWorkLaneShorthand(t *testing.T) {
	content, err := AgentInstructions("v5")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v5\") error = %v", err)
	}
	normalizedContent := strings.Join(strings.Fields(content), " ")

	for _, want := range []string{
		"first standalone token after trimming surrounding",
		"`c` means continue existing",
		"`n` or `y` means new lane",
		"primary lane choice",
		"remaining text is supplemental lane instructions",
		"explicit full-form choices",
		"ambiguous or contradictory responses fail closed",
	} {
		if !strings.Contains(normalizedContent, want) {
			t.Fatalf("v5 instructions do not contain %q", want)
		}
	}

	question := strings.Index(content, "Before I make any repository changes")
	shorthand := strings.Index(content, "first standalone token")
	wait := strings.Index(content, "Wait for the user's explicit choice")
	if question < 0 || shorthand <= question || wait <= shorthand {
		t.Fatal("v5 shorthand semantics must follow the question and precede the wait step")
	}
}

func TestAgentInstructionsV5PreservesV4OutsideAdditiveSections(t *testing.T) {
	v4, err := AgentInstructions("v4")
	if err != nil {
		t.Fatal(err)
	}
	v5, err := AgentInstructions("v5")
	if err != nil {
		t.Fatal(err)
	}
	shorthandStart := strings.Index(v5, "\n   Interpret the response's first standalone token")
	if shorthandStart < 0 {
		t.Fatal("v5 shorthand section start is missing")
	}
	shorthandEnd := strings.Index(v5[shorthandStart:], "\n4. Wait for the user's explicit choice")
	if shorthandEnd < 0 {
		t.Fatal("v5 shorthand section boundaries are missing")
	}
	withoutShorthand := v5[:shorthandStart] + v5[shorthandStart+shorthandEnd:]

	deletionStart := strings.Index(withoutShorthand, "\n# Deletion safety\n")
	deletionEnd := strings.Index(withoutShorthand, "\n# Execution\n")
	if deletionStart < 0 || deletionEnd < 0 || deletionStart >= deletionEnd {
		t.Fatal("v5 deletion section boundaries are missing")
	}
	withoutDeletion := withoutShorthand[:deletionStart] + withoutShorthand[deletionEnd:]
	if withoutDeletion != v4 {
		t.Fatal("v5 changed content outside the additive deletion and shorthand sections")
	}
}
