package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV10RequiresThreeSectionCompletionOutput(t *testing.T) {
	content, err := AgentInstructions("v10")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v10\") error = %v", err)
	}
	for _, want := range []string{
		"# Agent completion output",
		"Answer ordinary conversational requests naturally",
		"Emit exactly `## What happened`, `## Deviations`, and `## Next steps`",
		"**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**",
		"Do not add a separate status heading",
		"one nested evidence layer and state each fact once",
		"one `**None.**` bullet when there are no deviations",
		"Name the actor and make every required continuation copy-ready",
		"Use one `**None.**` bullet when no action remains",
		"PENDING, UNKNOWN, SKIPPED, and",
		"Completed, Validation, Delivery, Feature State, Residual Notes,",
		"three canonical sections without duplication",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v10 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
		"prioritized action list ordered Blocker",
		"Every structured PASS includes a None",
		"After the action list",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v10 instructions contain superseded output contract %q", forbidden)
		}
	}
	if CurrentAgentVersion != "v10" {
		t.Fatalf("CurrentAgentVersion = %q, want v10", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV10PreservesV9OutsideCompletionSection(t *testing.T) {
	v9, err := AgentInstructions("v9")
	if err != nil {
		t.Fatal(err)
	}
	v10, err := AgentInstructions("v10")
	if err != nil {
		t.Fatal(err)
	}
	const heading = "\n# Agent completion output\n"
	v9Start := strings.Index(v9, heading)
	v10Start := strings.Index(v10, heading)
	if v9Start < 0 || v10Start < 0 {
		t.Fatal("agent completion output section heading is missing")
	}
	if v9[:v9Start] != v10[:v10Start] {
		t.Fatal("v10 changed content outside the agent completion output section")
	}
}
