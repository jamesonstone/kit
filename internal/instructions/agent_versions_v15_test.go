package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV15RemovesCompletionFormat(t *testing.T) {
	content, err := AgentInstructions("v15")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v15\") error = %v", err)
	}
	for _, want := range []string{
		"There is no required response format",
		"in the shape the content calls for",
		"match length to consequence rather than to effort spent",
		"Do not emit the retired envelope",
		"Do not replace it with a different fixed template applied to every task",
		"never leave the reader wrong about a\n  blocker, incomplete scope, a required next action",
		"State plainly whether work is finished, partly finished, blocked, or failed",
		"No token, label, or taxonomy is required",
		"never folded into success",
		"Never report a check as passing unless it ran and passed",
		"When something could not be validated, say so and say why",
		"Include an identifier only when the reader needs it to act",
		"They name facts that must reach the\n  reader, never a layout",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v15 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"Use only `## What happened`, `## Deviations`, and `## Next steps`, in that",
		"Open What happened with",
		"Write a briefing, not a transcript",
		"Target twelve rendered lines or fewer",
		"at most five What happened bullets",
		"A PARTIAL, BLOCKED, or FAIL status always carries Deviations",
		"reserve bold for the\n  status line",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v15 instructions contain superseded completion format %q", forbidden)
		}
	}
	if CurrentAgentVersion != "v15" {
		t.Fatalf("CurrentAgentVersion = %q, want v15", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV15PreservesV14OutsideCompletionSection(t *testing.T) {
	v14, err := AgentInstructions("v14")
	if err != nil {
		t.Fatal(err)
	}
	v15, err := AgentInstructions("v15")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "# Agent completion output"
	prior, current := strings.Index(v14, marker), strings.Index(v15, marker)
	if prior < 0 || current < 0 {
		t.Fatal("both versions must define the agent completion output section")
	}
	if v14[:prior] != v15[:current] {
		t.Fatal("v15 changed instructions outside the agent completion output section")
	}
}
