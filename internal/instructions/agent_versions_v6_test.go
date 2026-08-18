package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV6RequiresAgentCompletionOutput(t *testing.T) {
	content, err := AgentInstructions("v6")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v6\") error = %v", err)
	}
	for _, want := range []string{
		"# Agent completion output",
		"docs/references/rules/agent-completion-output.md",
		"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
		"Type | Action required | Why | Continue with",
		"Order rows Blocker, Incomplete, Next, Optional, then None",
		"Every PASS includes",
		"a None row stating that no action is required",
		"Make required follow-ups copy-ready",
		"PENDING, UNKNOWN, SKIPPED, and",
		"NOT_APPLICABLE literally",
		"After the action table, use left-aligned headings and CommonMark list or",
		"Do not use another Markdown pipe table unless a",
		"implementation,",
		"research, diagnosis, planning, validation, review, operations, coordination,",
		"or fallback",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v6 instructions do not contain %q", want)
		}
	}
	if CurrentAgentVersion != "v6" {
		t.Fatalf("CurrentAgentVersion = %q, want v6", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV6PreservesV5AsExactPrefix(t *testing.T) {
	v5, err := AgentInstructions("v5")
	if err != nil {
		t.Fatal(err)
	}
	v6, err := AgentInstructions("v6")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(v6, v5) {
		t.Fatal("v6 changed content from immutable v5")
	}
	addition := strings.TrimPrefix(v6, v5)
	if !strings.HasPrefix(addition, "\n# Agent completion output\n") {
		t.Fatal("v6 completion contract is not an additive section after v5")
	}
}
