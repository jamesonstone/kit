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
		"Write each response, including terminal completions and handoffs, in the",
		"shape its content calls for",
		"Match length to consequence rather than to effort spent",
		"What the user now has",
		"What remains unfinished, and why",
		"Anything blocking completion, and what would clear it",
		"exact command or prompt when there is one",
		"Whether the work is finished, partly finished, blocked, or failed, said",
		"Keep blockers and unfinished scope as prominent as the successes",
		"Report each check as observed",
		"PENDING, UNKNOWN, SKIPPED, and\n  NOT_APPLICABLE verbatim",
		"Report a check as passing only when it ran and passed",
		"When something could not be validated, say so and say why",
		"Distinguish a verified fact from an inference and from a hypothesis",
		"not an index of everything checked",
		"smallest\n  evidence set that proves each terminal node",
		"Satisfy them on\n  content; a heading alone satisfies none of them",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v15 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"## What happened",
		"## Deviations",
		"## Next steps",
		"**Status: PASS|PARTIAL|BLOCKED|FAIL",
		"`**None.**`",
		"retired envelope",
		"fixed template applied to every task",
		"Write a briefing, not a transcript",
		"Target twelve rendered lines or fewer",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v15 instructions contain superseded completion format %q", forbidden)
		}
	}
}

func TestAgentInstructionsV15RemovesFixedFinalResponseStructureEverywhere(t *testing.T) {
	content, err := AgentInstructions("v15")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v15\") error = %v", err)
	}
	for _, want := range []string{
		"no labels, headings, or section are required",
		"No fixed order or\n  format is required; lead with whatever matters most to the reader",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v15 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"must include `Repository Memory`,",
		"- Lead final responses with:",
		"  1. Outcome",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v15 instructions still impose a fixed final-response structure %q", forbidden)
		}
	}
}

func TestAgentInstructionsV15PreservesV14BeforeReportingSections(t *testing.T) {
	v14, err := AgentInstructions("v14")
	if err != nil {
		t.Fatal(err)
	}
	v15, err := AgentInstructions("v15")
	if err != nil {
		t.Fatal(err)
	}
	// v15 rewrites only the three sections that impose a response shape:
	// repository memory completion, communication, and agent completion output.
	// Everything before the first of them must be untouched.
	const marker = "# Repository memory completion"
	prior, current := strings.Index(v14, marker), strings.Index(v15, marker)
	if prior < 0 || current < 0 {
		t.Fatal("both versions must define the repository memory completion section")
	}
	if v14[:prior] != v15[:current] {
		t.Fatal("v15 changed instructions outside the response-shape sections")
	}

	const pullRequest = "# Pull request"
	if v14[strings.Index(v14, pullRequest):strings.Index(v14, "# Communication")] !=
		v15[strings.Index(v15, pullRequest):strings.Index(v15, "# Communication")] {
		t.Fatal("v15 changed the pull request section")
	}
}
