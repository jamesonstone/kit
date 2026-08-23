package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV6MapsFullFormNewLaneChoices(t *testing.T) {
	content, err := AgentInstructions("v6")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v6\") error = %v", err)
	}
	normalizedContent := strings.Join(strings.Fields(content), " ")

	for _, want := range []string{
		"`new lane`, `new work lane`, `new worklane`, and `new worktree`",
		"as the new-lane choice",
		"human-assigned GitHub issue",
		"exact `GH-<issue-number>` branch",
		"canonical non-primary worktree",
		"ready pull-request plan",
		"ambiguous or contradictory responses fail closed",
	} {
		if !strings.Contains(normalizedContent, want) {
			t.Fatalf("v6 instructions do not contain %q", want)
		}
	}

	question := strings.Index(content, "Before I make any repository changes")
	shorthand := strings.Index(content, "first standalone token")
	fullForm := strings.Index(content, "full-form answers")
	wait := strings.Index(content, "Wait for the user's explicit choice")
	if question < 0 || shorthand <= question || fullForm <= shorthand || wait <= fullForm {
		t.Fatal("v6 full-form semantics must follow shorthand and precede the wait step")
	}
}

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
}

func TestAgentInstructionsV6PreservesV5OutsideAdditiveSections(t *testing.T) {
	v5, err := AgentInstructions("v5")
	if err != nil {
		t.Fatal(err)
	}
	v6, err := AgentInstructions("v6")
	if err != nil {
		t.Fatal(err)
	}
	aliasStart := strings.Index(v6, "\n   Treat the case-insensitive full-form answers")
	if aliasStart < 0 {
		t.Fatal("v6 full-form mapping start is missing")
	}
	aliasEnd := strings.Index(v6[aliasStart:], "\n4. Wait for the user's explicit choice")
	if aliasEnd < 0 {
		t.Fatal("v6 full-form mapping boundary is missing")
	}
	withoutMapping := v6[:aliasStart] + v6[aliasStart+aliasEnd:]
	completionStart := strings.Index(withoutMapping, "\n# Agent completion output\n")
	if completionStart < 0 {
		t.Fatal("v6 completion contract start is missing")
	}
	if withoutMapping[:completionStart] != v5 {
		t.Fatal("v6 changed content outside the additive full-form and completion sections")
	}
}
