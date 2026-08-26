package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV9RequiresProportionalCompletionOutput(t *testing.T) {
	content, err := AgentInstructions("v9")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v9\") error = %v", err)
	}
	for _, want := range []string{
		"# Agent completion output",
		"Before a substantial terminal completion or handoff response",
		"specific exception to the earlier general",
		"Answer ordinary conversational requests naturally",
		"Direct questions, definitions, confirmations, rewrites, brief explanations",
		"must not receive a status heading, Next actions section, synthetic",
		"Use the structured contract when omitting it could hide a blocker",
		"Do not classify by word count, token count, elapsed time, or tool-call count",
		"When uncertain, prefer natural prose",
		"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
		"prioritized action list ordered Blocker,",
		"Every structured PASS includes a None",
		"Never leave Why or Continue with blank",
		"After the action list, use left-aligned headings and CommonMark list or",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v9 instructions do not contain %q", want)
		}
	}
}

func TestAgentInstructionsV9PreservesV8OutsideCompletionSection(t *testing.T) {
	v8, err := AgentInstructions("v8")
	if err != nil {
		t.Fatal(err)
	}
	v9, err := AgentInstructions("v9")
	if err != nil {
		t.Fatal(err)
	}
	const heading = "\n# Agent completion output\n"
	v8Start := strings.Index(v8, heading)
	v9Start := strings.Index(v9, heading)
	if v8Start < 0 || v9Start < 0 {
		t.Fatal("agent completion output section heading is missing")
	}
	if v8[:v8Start] != v9[:v9Start] {
		t.Fatal("v9 changed content outside the agent completion output section")
	}
}
