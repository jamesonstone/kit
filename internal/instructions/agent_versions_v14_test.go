package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV14BoundsCompletionDensity(t *testing.T) {
	content, err := AgentInstructions("v14")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v14\") error = %v", err)
	}
	for _, want := range []string{
		"Use only `## What happened`, `## Deviations`, and `## Next steps`, in that\n  order",
		"omit Deviations and Next steps entirely\n  when empty instead of writing a None bullet",
		"on its own\n  line, then one to three plain sentences of prose",
		"reserve bold for the\n  status line, blockers, and required actions",
		"Write a briefing, not a transcript",
		"Target twelve rendered lines or fewer",
		"at most five What happened bullets",
		"A PARTIAL, BLOCKED, or FAIL status always carries Deviations",
		"never manufacture or speculatively offer an action",
		"Include an identifier only when the reader needs it to act",
		"Omit a commit SHA cited as proof that something was\n  checked",
		"Name the validation that ran and its result in a few words",
		"using the shortest phrasing that\n  keeps each fact recoverable",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v14 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"Emit exactly `## What happened`, `## Deviations`, and `## Next steps` in",
		"one `**None.**` bullet when there are no deviations",
		"Use one `**None.**` bullet when no action remains",
		"Use at most\n  one nested evidence layer and state each fact once",
		"in the\n  first What happened bullet",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v14 instructions contain superseded completion guidance %q", forbidden)
		}
	}
}

func TestAgentInstructionsV14PreservesV13OutsideCompletionSection(t *testing.T) {
	v13, err := AgentInstructions("v13")
	if err != nil {
		t.Fatal(err)
	}
	v14, err := AgentInstructions("v14")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "# Agent completion output"
	prior, current := strings.Index(v13, marker), strings.Index(v14, marker)
	if prior < 0 || current < 0 {
		t.Fatal("both versions must define the agent completion output section")
	}
	if v13[:prior] != v14[:current] {
		t.Fatal("v14 changed instructions outside the agent completion output section")
	}
}
