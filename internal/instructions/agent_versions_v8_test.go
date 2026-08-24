package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV8RequiresListFirstCompletionOutput(t *testing.T) {
	content, err := AgentInstructions("v8")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v8\") error = %v", err)
	}
	for _, want := range []string{
		"# Agent completion output",
		"docs/references/rules/agent-completion-output.md",
		"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
		"prioritized action list ordered Blocker,",
		"Every PASS includes a None item",
		"Never leave Why or Continue with blank",
		"After the action list, use left-aligned headings and CommonMark list or",
		"Do not use a Markdown pipe table unless a higher-priority",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v8 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"Type | Action required | Why | Continue with",
		"After the action table, use left-aligned headings",
		"a None row stating that no action is required",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v8 instructions still contain table contract %q", forbidden)
		}
	}
	if CurrentAgentVersion != "v8" {
		t.Fatalf("CurrentAgentVersion = %q, want v8", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV8PreservesV7OutsideCompletionSection(t *testing.T) {
	v7, err := AgentInstructions("v7")
	if err != nil {
		t.Fatal(err)
	}
	v8, err := AgentInstructions("v8")
	if err != nil {
		t.Fatal(err)
	}
	const heading = "\n# Agent completion output\n"
	v7Start := strings.Index(v7, heading)
	v8Start := strings.Index(v8, heading)
	if v7Start < 0 || v8Start < 0 {
		t.Fatal("agent completion output section heading is missing")
	}
	if v7[:v7Start] != v8[:v8Start] {
		t.Fatal("v8 changed content outside the agent completion output section")
	}
}
