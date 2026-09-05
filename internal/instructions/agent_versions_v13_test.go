package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV13RejectsExactHeadReauthorization(t *testing.T) {
	content, err := AgentInstructions("v13")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v13\") error = %v", err)
	}
	for _, want := range []string{
		"commit SHA or head OID identifies readiness evidence only",
		"never an\n  authorization identity or approval token",
		"invalidates prior\n  checks and review, not standing authority",
		"Never request exact-head\n  reauthorization",
		"After final-head evidence restores `MERGE_READY`",
		"authorized standard deployment and browser retry",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v13 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"prior merge authority is invalid",
		"fresh exact-head authorization",
		"refreshed head needs exact-head authorization",
	} {
		if strings.Contains(strings.ToLower(content), forbidden) {
			t.Fatalf("v13 instructions contain superseded requirement %q", forbidden)
		}
	}
	if CurrentAgentVersion != "v13" {
		t.Fatalf("CurrentAgentVersion = %q, want v13", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV13PreservesV12OutsideAuthoritySection(t *testing.T) {
	v12, err := AgentInstructions("v12")
	if err != nil {
		t.Fatal(err)
	}
	v13, err := AgentInstructions("v13")
	if err != nil {
		t.Fatal(err)
	}
	const mergeStart = "\n## Standing merge and deployment authority\n"
	const executionStart = "\n# Execution environment\n"
	v12Merge := strings.Index(v12, mergeStart)
	v13Merge := strings.Index(v13, mergeStart)
	v12Execution := strings.Index(v12, executionStart)
	v13Execution := strings.Index(v13, executionStart)
	if v12Merge < 0 || v13Merge < 0 || v12Execution < 0 || v13Execution < 0 {
		t.Fatal("instruction section boundaries are missing")
	}
	if v12[:v12Merge] != v13[:v13Merge] || v12[v12Execution:] != v13[v13Execution:] {
		t.Fatal("v13 changed content outside standing authority section")
	}
}
