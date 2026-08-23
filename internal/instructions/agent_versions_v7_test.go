package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV7RequiresMultiAgentOrchestrationGate(t *testing.T) {
	content, err := AgentInstructions("v7")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v7\") error = %v", err)
	}
	for _, want := range []string{
		"# Multi-agent orchestration evaluation gate",
		"docs/references/rules/agent-team-orchestration.md",
		"mandatory even when the recorded decision is one",
		"single-lane, because <reason>",
		"does not trigger this gate",
		"before the first repository mutation",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v7 instructions do not contain %q", want)
		}
	}
	if CurrentAgentVersion != "v7" {
		t.Fatalf("CurrentAgentVersion = %q, want v7", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV7PreservesV6OutsideAdditiveSection(t *testing.T) {
	v6, err := AgentInstructions("v6")
	if err != nil {
		t.Fatal(err)
	}
	v7, err := AgentInstructions("v7")
	if err != nil {
		t.Fatal(err)
	}
	additiveStart := strings.Index(v7, "\n# Multi-agent orchestration evaluation gate\n")
	if additiveStart < 0 {
		t.Fatal("v7 multi-agent orchestration section start is missing")
	}
	additiveEnd := strings.Index(v7[additiveStart:], "\n# Repository context and mutation gate")
	if additiveEnd < 0 {
		t.Fatal("v7 multi-agent orchestration section boundary is missing")
	}
	withoutAdditive := v7[:additiveStart] + v7[additiveStart+additiveEnd:]
	if withoutAdditive != v6 {
		t.Fatal("v7 changed content outside the additive multi-agent orchestration section")
	}
}
