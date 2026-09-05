package instructions

import (
	"strings"
	"testing"
)

func TestAgentInstructionsV12DefinesStandingAuthority(t *testing.T) {
	content, err := AgentInstructions("v12")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v12\") error = %v", err)
	}
	for _, want := range []string{
		"Standing authority exists only when a human explicitly authorizes a bounded",
		"may bind later-created in-scope PRs and refreshed",
		"number or final OID was unknown at grant",
		"Only exact current `MERGE_READY` nodes may merge",
		"changed in-scope head invalidates readiness, not standing authority",
		"named existing standard workflow",
		"IAM, network,",
		"KMS, secrets, database schema/data-loss",
		"Pause, hold, or revocation",
		"until explicit human resume",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v12 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{
		"Merge authority requires a direct user request or an accepted bounded merge plan naming the exact pull-request set",
		"accepted task or active `/goal` authorizes",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v12 instructions contain superseded authority guidance %q", forbidden)
		}
	}
	if CurrentAgentVersion != "v12" {
		t.Fatalf("CurrentAgentVersion = %q, want v12", CurrentAgentVersion)
	}
}

func TestAgentInstructionsV12PreservesV11OutsideAuthoritySections(t *testing.T) {
	v11, err := AgentInstructions("v11")
	if err != nil {
		t.Fatal(err)
	}
	v12, err := AgentInstructions("v12")
	if err != nil {
		t.Fatal(err)
	}
	const mergeStartV11 = "\n## Merge authorization\n"
	const mergeStartV12 = "\n## Standing merge and deployment authority\n"
	const executionStart = "\n# Execution environment\n"
	v11Merge := strings.Index(v11, mergeStartV11)
	v12Merge := strings.Index(v12, mergeStartV12)
	v11Execution := strings.Index(v11, executionStart)
	v12Execution := strings.Index(v12, executionStart)
	if v11Merge < 0 || v12Merge < 0 || v11Execution < 0 || v12Execution < 0 {
		t.Fatal("instruction section boundaries are missing")
	}
	if v11[:v11Merge] == v12[:v12Merge] {
		t.Fatal("v12 must also update the in-place repair authority wording before the authority section")
	}
	if v11[v11Execution:] != v12[v12Execution:] {
		t.Fatal("v12 changed content after the authority section")
	}
}
